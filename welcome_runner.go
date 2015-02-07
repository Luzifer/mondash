package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func RunWelcomePage() {
	baseURL := os.Getenv("BASE_URL")
	welcomeAPIToken := os.Getenv("API_TOKEN")
	generateTicker := time.NewTicker(time.Minute)

	for {
		select {
		case <-generateTicker.C:
			beers := rand.Intn(24)
			status := "OK"
			switch {
			case beers < 6:
				status = "Critical"
				break
			case beers < 12:
				status = "Warning"
				break
			}

			beer := DashboardMetric{
				Title:       "Amount of beer in the fridge",
				Description: fmt.Sprintf("Currently there are %d bottles of beer in the fridge", beers),
				Status:      status,
				Expires:     86400,
				Freshness:   120,
			}

			body, err := json.Marshal(beer)
			if err != nil {
				log.Println(err)
			}
			url := fmt.Sprintf("%s/welcome/beer_available", baseURL)
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			req.Header.Add("Authorization", welcomeAPIToken)
			http.DefaultClient.Do(req)
		}
	}

}
