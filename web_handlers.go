package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/flosch/pongo2"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func handleRedirectWelcome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/welcome", http.StatusTemporaryRedirect)
}

func handleCreateRandomDashboard(w http.ResponseWriter, r *http.Request) {
	var urlProposal string
	for {
		urlProposal = generateAPIKey()[0:20]
		if exists, err := store.Exists(urlProposal); err == nil && !exists {
			break
		}
	}
	http.Redirect(w, r, fmt.Sprintf("/%s", urlProposal), http.StatusTemporaryRedirect)
}

func handleDisplayDashboard(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)

	dash, err := loadDashboard(vars["dashid"], store)
	switch err {
	case nil:
		// All fine

	case errDashboardNotFound:
		dash = &dashboard{APIKey: generateAPIKey(), Metrics: []*dashboardMetric{}}

	default:
		log.WithError(err).
			WithField("dashboard_id", vars["dashid"]).
			Error("Unable to load dashboard")
		http.Error(w, "Could not load dashboard", http.StatusInternalServerError)
		return
	}

	// Filter out expired metrics
	metrics := []*dashboardMetric{}
	for _, m := range dash.Metrics {
		if m.Meta.LastUpdate.After(time.Now().Add(time.Duration(m.Expires*-1) * time.Second)) {
			metrics = append(metrics, m)
		}
	}

	sort.Slice(metrics, func(j, i int) bool { return metrics[i].Meta.LastUpdate.Before(metrics[j].Meta.LastUpdate) })

	renderTemplate("dashboard.html", pongo2.Context{
		"dashid":  vars["dashid"],
		"metrics": metrics,
		"apikey":  dash.APIKey,
		"baseurl": cfg.BaseURL,
	}, w)
}

func handleDisplayDashboardJSON(w http.ResponseWriter, r *http.Request) {
	var vars = mux.Vars(r)

	dash, err := loadDashboard(vars["dashid"], store)
	switch err {
	case nil:
		// All fine

	case errDashboardNotFound:
		dash = &dashboard{APIKey: generateAPIKey(), Metrics: []*dashboardMetric{}}

	default:
		log.WithError(err).
			WithField("dashboard_id", vars["dashid"]).
			Error("Unable to load dashboard")
		http.Error(w, "Could not load dashboard", http.StatusInternalServerError)
		return
	}

	response := struct {
		APIKey  string `json:"api_key,omitempty"`
		Metrics []struct {
			ID          string    `json:"id"`
			Title       string    `json:"title"`
			Description string    `json:"description"`
			Status      string    `json:"status"`
			Value       float64   `json:"value"`
			LastUpdate  time.Time `json:"last_update"`
		} `json:"metrics"`
	}{}

	// Filter out expired metrics
	for _, m := range dash.Metrics {
		if m.Meta.LastUpdate.After(time.Now().Add(time.Duration(m.Expires*-1) * time.Second)) {
			response.Metrics = append(response.Metrics, struct {
				ID          string    `json:"id"`
				Title       string    `json:"title"`
				Description string    `json:"description"`
				Status      string    `json:"status"`
				Value       float64   `json:"value"`
				LastUpdate  time.Time `json:"last_update"`
			}{
				ID:          m.MetricID,
				Title:       m.Title,
				Description: m.Description,
				Status:      m.PreferredStatus(),
				Value:       m.Value,
				LastUpdate:  m.Meta.LastUpdate,
			})
		}
	}

	if len(response.Metrics) == 0 {
		response.APIKey = dash.APIKey
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")
	json.NewEncoder(w).Encode(response)
}

func handleDeleteDashboard(w http.ResponseWriter, r *http.Request) {
	var (
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Token ")
		vars  = mux.Vars(r)
	)

	dash, err := loadDashboard(vars["dashid"], store)
	switch err {
	case nil:
		// All fine

	case errDashboardNotFound:
		http.Error(w, "Dasboard not found", http.StatusNotFound)
		return

	default:
		log.WithError(err).
			WithField("dashboard_id", vars["dashid"]).
			Error("Unable to load dashboard")
		http.Error(w, "Could not load dashboard", http.StatusInternalServerError)
		return
	}

	if dash.APIKey != token {
		http.Error(w, "APIKey did not match.", http.StatusUnauthorized)
		return
	}

	store.Delete(vars["dashid"])
	http.Error(w, "OK", http.StatusOK)
}

func handlePutMetric(w http.ResponseWriter, r *http.Request) {
	var (
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Token ")
		vars  = mux.Vars(r)
	)

	metricUpdate := newDashboardMetric()
	if err := json.NewDecoder(r.Body).Decode(metricUpdate); err != nil {
		http.Error(w, "Unable to unmarshal json body", http.StatusBadRequest)
		return
	}

	dash, err := loadDashboard(vars["dashid"], store)
	switch err {
	case nil:
		// All fine

	case errDashboardNotFound:
		// Dashboard may be created with first metrics put
		if len(token) < 10 {
			http.Error(w, "APIKey is too insecure", http.StatusBadRequest)
			return
		}

		dash = &dashboard{
			APIKey:      token,
			Metrics:     []*dashboardMetric{},
			DashboardID: vars["dashid"],
			storage:     store,
		}

	default:
		log.WithError(err).
			WithField("dashboard_id", vars["dashid"]).
			Error("Unable to load dashboard")
		http.Error(w, "Could not load dashboard", http.StatusInternalServerError)
		return
	}

	if dash.APIKey != token {
		http.Error(w, "APIKey did not match.", http.StatusUnauthorized)
		return
	}

	valid, reason := metricUpdate.IsValid()
	if !valid {
		http.Error(w, fmt.Sprintf("Invalid data: %s", reason), http.StatusBadRequest)
		return
	}

	updated := false
	for _, m := range dash.Metrics {
		if m.MetricID == vars["metricid"] {
			m.Update(metricUpdate)
			updated = true
			break
		}
	}

	if !updated {
		tmp := newDashboardMetric()
		tmp.MetricID = vars["metricid"]
		tmp.Update(metricUpdate)
		dash.Metrics = append(dash.Metrics, tmp)
	}

	if err := dash.Save(); err != nil {
		log.WithError(err).Error("Unable to save dashboard")
		http.Error(w, "Was not able to save the dashboard", http.StatusInternalServerError)
		return
	}

	http.Error(w, "OK", http.StatusOK)
}

func handleDeleteMetric(w http.ResponseWriter, r *http.Request) {
	var (
		token = strings.TrimPrefix(r.Header.Get("Authorization"), "Token ")
		vars  = mux.Vars(r)
	)

	dash, err := loadDashboard(vars["dashid"], store)
	switch err {
	case nil:
		// All fine

	case errDashboardNotFound:
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return

	default:
		log.WithError(err).
			WithField("dashboard_id", vars["dashid"]).
			Error("Unable to load dashboard")
		http.Error(w, "Could not load dashboard", http.StatusInternalServerError)
		return
	}

	if dash.APIKey != token {
		http.Error(w, "APIKey did not match.", http.StatusUnauthorized)
		return
	}

	tmp := []*dashboardMetric{}
	for _, m := range dash.Metrics {
		if m.MetricID != vars["metricid"] {
			tmp = append(tmp, m)
		}
	}
	dash.Metrics = tmp

	if err := dash.Save(); err != nil {
		log.WithError(err).Error("Unable to save dashboard")
		http.Error(w, "Was not able to save the dashboard", http.StatusInternalServerError)
		return
	}

	http.Error(w, "OK", http.StatusOK)
}
