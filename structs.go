package main

import (
	"encoding/json"
	"errors"
	"log"
	"time"

	"launchpad.net/goamz/s3"
)

type Dashboard struct {
	DashboardID string           `json:"-"`
	APIKey      string           `json:"api_key"`
	Metrics     DashboardMetrics `json:"metrics"`
}

func LoadDashboard(dashid string) (*Dashboard, error) {
	data, err := s3Storage.Get(dashid)
	if err != nil {
		return &Dashboard{}, errors.New("Dashboard not found")
	}

	tmp := &Dashboard{DashboardID: dashid}
	json.Unmarshal(data, tmp)

	return tmp, nil
}

func (d *Dashboard) Save() {
	data, err := json.Marshal(d)
	if err != nil {
		log.Printf("Error while marshalling dashboard: %s", err)
		return
	}
	err = s3Storage.Put(d.DashboardID, data, "application/json", s3.Private)
	if err != nil {
		log.Printf("Error while storing dashboard: %s", err)
	}
}

type DashboardMetrics []*DashboardMetric

func (a DashboardMetrics) Len() int      { return len(a) }
func (a DashboardMetrics) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a DashboardMetrics) Less(i, j int) bool {
	return a[i].HistoricalData[0].Time.Before(a[j].HistoricalData[0].Time)
}

type DashboardMetric struct {
	MetricID       string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Status         string                 `json:"status"`
	Expires        int64                  `json:"expires,omitifempty"`
	Freshness      int64                  `json:"freshness,omitifempty"`
	HistoricalData DashboardMetricHistory `json:"history,omitifempty"`
	Meta           DashboardMetricMeta    `json:"meta,omitifempty"`
}

type DashboardMetricStatus struct {
	Time   time.Time `json:"time"`
	Status string    `json:"status"`
}

type DashboardMetricMeta struct {
	LastUpdate time.Time
	LastOK     time.Time
	PercOK     float64
	PercWarn   float64
	PercCrit   float64
}

type DashboardMetricHistory []DashboardMetricStatus

func NewDashboardMetric() *DashboardMetric {
	return &DashboardMetric{
		Status:         "Unknown",
		Expires:        604800,
		Freshness:      3600,
		HistoricalData: DashboardMetricHistory{},
		Meta:           DashboardMetricMeta{},
	}
}

func (dm *DashboardMetric) Update(m *DashboardMetric) {
	dm.Title = m.Title
	dm.Description = m.Description
	dm.Status = m.Status
	if m.Expires != 0 {
		dm.Expires = m.Expires
	}
	if m.Freshness != 0 {
		dm.Freshness = m.Freshness
	}
	dm.HistoricalData = append(DashboardMetricHistory{DashboardMetricStatus{
		Time:   time.Now(),
		Status: m.Status,
	}}, dm.HistoricalData...)

	countStatus := make(map[string]float64)

	expired := time.Now().Add(time.Duration(dm.Expires*-1) * time.Second)
	tmp := DashboardMetricHistory{}
	for _, s := range dm.HistoricalData {
		if s.Time.After(expired) {
			tmp = append(tmp, s)
			countStatus[s.Status] = countStatus[s.Status] + 1
			countStatus["Total"] = countStatus["Total"] + 1
			if dm.Meta.LastOK.Before(s.Time) && s.Status == "OK" {
				dm.Meta.LastOK = s.Time
			}
		}
	}
	dm.HistoricalData = tmp

	dm.Meta.LastUpdate = time.Now()
	if countStatus["Total"] > 0 {
		dm.Meta.PercCrit = countStatus["Critical"] / countStatus["Total"] * 100
		dm.Meta.PercWarn = countStatus["Warning"] / countStatus["Total"] * 100
		dm.Meta.PercOK = countStatus["OK"] / countStatus["Total"] * 100
	}
}

func (dm *DashboardMetric) IsValid() (bool, string) {
	if dm.Expires > 604800 || dm.Expires < 0 {
		return false, "Expires not in range 0 < x < 640800"
	}

	if dm.Freshness > 604800 || dm.Freshness < 0 {
		return false, "Freshness not in range 0 < x < 640800"
	}

	if !stringInSlice(dm.Status, []string{"OK", "Warning", "Critical", "Unknowm"}) {
		return false, "Status not allowed"
	}

	if len(dm.Title) > 512 || len(dm.Description) > 1024 {
		return false, "Title or Description too long"
	}

	return true, ""
}
