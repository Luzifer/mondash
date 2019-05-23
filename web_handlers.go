package main

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type output struct {
	APIKey  string         `json:"api_key,omitempty"`
	Metrics []outputMetric `json:"metrics"`
}

type outputMetricConfig struct {
	HideMAD   bool `json:"hide_mad"`
	HideValue bool `json:"hide_value"`
}

type outputMetric struct {
	ID            string              `json:"id"`
	Config        outputMetricConfig  `json:"config"`
	Description   string              `json:"description"`
	HistoryBar    []historyBarSegment `json:"history_bar,omitempty"`
	LastOK        time.Time           `json:"last_ok"`
	LastUpdate    time.Time           `json:"last_update"`
	Median        float64             `json:"median"`
	MADMultiplier float64             `json:"mad_multiplier"`
	Status        string              `json:"status"`
	Title         string              `json:"title"`
	Value         float64             `json:"value"`
	ValueHistory  map[int64]float64   `json:"value_history,omitempty"`
}

type outputMetricFromMetricOpts struct {
	Metric          *dashboardMetric
	AddHistoryBar   bool
	AddValueHistory bool
}

func outputMetricFromMetric(opts outputMetricFromMetricOpts) outputMetric {
	out := outputMetric{
		ID:            opts.Metric.MetricID,
		Description:   opts.Metric.Description,
		LastOK:        opts.Metric.Meta.LastOK,
		LastUpdate:    opts.Metric.Meta.LastUpdate,
		Median:        opts.Metric.Median(),
		MADMultiplier: opts.Metric.MadMultiplier(),
		Status:        opts.Metric.PreferredStatus(),
		Title:         opts.Metric.Title,
		Value:         opts.Metric.Value,

		Config: outputMetricConfig{
			HideMAD:   opts.Metric.HideMAD,
			HideValue: opts.Metric.HideValue,
		},
	}

	if opts.AddHistoryBar {
		out.HistoryBar = opts.Metric.GetHistoryBar()
	}

	if opts.AddValueHistory {
		out.ValueHistory = opts.Metric.HistoricalValueMap()
	}

	return out
}

func handleStaticFile(w http.ResponseWriter, r *http.Request, filename string) error {
	if _, err := os.Stat(path.Join(cfg.FrontendDir, filename)); err == nil {
		http.ServeFile(w, r, path.Join(cfg.FrontendDir, filename))
		return nil
	}

	// File was not found in filesystem, serve from packed assets
	body, err := Asset(path.Join("frontend", filename))
	if err != nil {
		return errors.Wrapf(err, "File %q was neither found in FrontendDir nor in assets", filename)
	}

	log.WithField("filename", filename).Debug("Static file loaded from assets")

	w.Header().Set("Content-Type", mime.TypeByExtension(filename[strings.LastIndexByte(filename, '.'):]))
	_, err = w.Write(body)
	return errors.Wrap(err, "Unable to write body")
}

func handleAppJS(w http.ResponseWriter, r *http.Request) { handleStaticFile(w, r, "app.js") }

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
	handleStaticFile(w, r, "index.html")
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

	var (
		addHistoryBar   = r.URL.Query().Get("history_bar") == "true"
		addValueHistory = r.URL.Query().Get("value_history") == "true"
		response        = output{}
	)

	// Filter out expired metrics
	for _, m := range dash.Metrics {
		if m.Meta.LastUpdate.After(time.Now().Add(time.Duration(m.Expires*-1) * time.Second)) {
			response.Metrics = append(response.Metrics, outputMetricFromMetric(outputMetricFromMetricOpts{
				AddHistoryBar:   addHistoryBar,
				AddValueHistory: addValueHistory,
				Metric:          m,
			}))
		}
	}

	sort.Slice(response.Metrics, func(j, i int) bool { return response.Metrics[i].LastUpdate.Before(response.Metrics[j].LastUpdate) })

	if len(response.Metrics) == 0 {
		response.APIKey = dash.APIKey
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache")

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.WithError(err).Error("Unable to encode API JSON")
		http.Error(w, "Unable to encode JSON", http.StatusInternalServerError)
	}
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

	if err = store.Delete(vars["dashid"]); err != nil {
		log.WithError(err).WithField("dashboard_id", vars["dashid"]).Error("Unable to delete dashboard")
		http.Error(w, "Failed to delete dashboard", http.StatusInternalServerError)
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

func handleRedirectWelcome(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/welcome", http.StatusTemporaryRedirect)
}
