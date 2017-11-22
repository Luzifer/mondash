package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/Luzifer/mondash/config"
)

func runWelcomePage(cfg *config.Config) {
	baseURL := cfg.BaseURL
	welcomeAPIToken := cfg.APIToken
	generateTicker := time.NewTicker(time.Minute)

	// Do one initial push on start
	postWelcomeMetric(baseURL, welcomeAPIToken)

	for range generateTicker.C {
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
		Freshness:   120,
		Value:       float64(beers),
	}

	body, err := json.Marshal(beer)
	if err != nil {
		log.Println(err)
		return
	}
	url := fmt.Sprintf("%s/welcome/beer_available", baseURL)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	req.Header.Add("Authorization", welcomeAPIToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("[WelcomeRunner] %s", err)
		return
	}
	resp.Body.Close()
}
