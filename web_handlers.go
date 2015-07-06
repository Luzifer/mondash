package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/flosch/pongo2"
	"github.com/go-martini/martini"
)

func handleRedirectWelcome(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/welcome", 302)
}

func handleCreateRandomDashboard(res http.ResponseWriter, req *http.Request) {
	urlProposal := generateAPIKey()[0:20]
	_, err := s3Storage.Get(urlProposal)
	for err == nil {
		urlProposal = generateAPIKey()[0:20]
		_, err = s3Storage.Get(urlProposal)
	}
	http.Redirect(res, req, fmt.Sprintf("/%s", urlProposal), http.StatusTemporaryRedirect)
}

func handleDisplayDashboard(params martini.Params, res http.ResponseWriter) {
	dash, err := loadDashboard(params["dashid"])
	if err != nil {
		dash = &dashboard{APIKey: generateAPIKey(), Metrics: dashboardMetrics{}}
	}

	// Filter out expired metrics
	metrics := dashboardMetrics{}
	for _, m := range dash.Metrics {
		if m.Meta.LastUpdate.After(time.Now().Add(time.Duration(m.Expires*-1) * time.Second)) {
			metrics = append(metrics, m)
		}
	}

	sort.Sort(sort.Reverse(dashboardMetrics(metrics)))
	renderTemplate("dashboard.html", pongo2.Context{
		"dashid":  params["dashid"],
		"metrics": metrics,
		"apikey":  dash.APIKey,
		"baseurl": os.Getenv("BASE_URL"),
	}, res)
}

func handleDeleteDashboard(params martini.Params, req *http.Request, res http.ResponseWriter) {
	dash, err := loadDashboard(params["dashid"])
	if err != nil {
		http.Error(res, "This dashboard does not exist.", http.StatusInternalServerError)
		return
	}

	if dash.APIKey != req.Header.Get("Authorization") {
		http.Error(res, "APIKey did not match.", http.StatusUnauthorized)
		return
	}

	_ = s3Storage.Del(params["dashid"])
	http.Error(res, "OK", http.StatusOK)
}

func handlePutMetric(params martini.Params, req *http.Request, res http.ResponseWriter) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	metricUpdate := newDashboardMetric()
	err = json.Unmarshal(body, metricUpdate)
	if err != nil {
		http.Error(res, "Unable to unmarshal json", http.StatusInternalServerError)
		return
	}

	dash, err := loadDashboard(params["dashid"])
	if err != nil {
		if len(req.Header.Get("Authorization")) < 10 {
			http.Error(res, "APIKey is too insecure", http.StatusUnauthorized)
			return
		}
		dash = &dashboard{APIKey: req.Header.Get("Authorization"), Metrics: dashboardMetrics{}, DashboardID: params["dashid"]}
	}

	if dash.APIKey != req.Header.Get("Authorization") {
		http.Error(res, "APIKey did not match.", http.StatusUnauthorized)
		return
	}

	valid, reason := metricUpdate.IsValid()
	if !valid {
		http.Error(res, fmt.Sprintf("Invalid data: %s", reason), http.StatusInternalServerError)
		return
	}

	updated := false
	for _, m := range dash.Metrics {
		if m.MetricID == params["metricid"] {
			m.Update(metricUpdate)
			updated = true
			break
		}
	}

	if !updated {
		tmp := newDashboardMetric()
		tmp.MetricID = params["metricid"]
		tmp.Update(metricUpdate)
		dash.Metrics = append(dash.Metrics, tmp)
	}

	dash.Save()

	http.Error(res, "OK", http.StatusOK)
}

func handleDeleteMetric(params martini.Params, req *http.Request, res http.ResponseWriter) {
	dash, err := loadDashboard(params["dashid"])
	if err != nil {
		dash = &dashboard{APIKey: req.Header.Get("Authorization"), Metrics: dashboardMetrics{}, DashboardID: params["dashid"]}
	}

	if dash.APIKey != req.Header.Get("Authorization") {
		http.Error(res, "APIKey did not match.", http.StatusUnauthorized)
		return
	}

	tmp := dashboardMetrics{}
	for _, m := range dash.Metrics {
		if m.MetricID != params["metricid"] {
			tmp = append(tmp, m)
		}
	}
	dash.Metrics = tmp
	dash.Save()

	http.Error(res, "OK", http.StatusOK)
}
