package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func runWelcomePage() {
	var (
		baseURL         = cfg.BaseURL
		welcomeAPIToken = cfg.APIToken
	)

	for tick := time.NewTicker(10 * time.Minute); ; <-tick.C {
		postWelcomeMetric(baseURL, welcomeAPIToken)
	}
}

func postWelcomeMetric(baseURL, welcomeAPIToken string) {
	beers := rand.Intn(24)
	status := "OK"
	switch {
	case beers < 6:
		status = "Critical"
	case beers < 12:
		status = "Warning"
	}

	beer := dashboardMetric{
		Title:       "Amount of beer in the fridge",
		Description: fmt.Sprintf("Currently there are %d bottles of beer in the fridge", beers),
		IgnoreMAD:   true,
		Status:      status,
		Expires:     86400,
		Freshness:   900,
		Value:       float64(beers),
	}

	body := new(bytes.Buffer)
	if err := json.NewEncoder(body).Encode(beer); err != nil {
		log.WithError(err).Error("Unable to marshal dashboard metric")
		return
	}

	url := fmt.Sprintf("%s/welcome/beer_available", baseURL)
	req, _ := http.NewRequest(http.MethodPut, url, body)
	req.Header.Add("Authorization", welcomeAPIToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.WithError(err).Error("Unable to put welcome-runner metric")
		return
	}
	resp.Body.Close()

	log.Debug("Successfully put welcome-runner metric")
}
