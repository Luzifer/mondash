package main

import (
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/Luzifer/mondash/storage"
)

type dashboard struct {
	DashboardID string           `json:"-"`
	APIKey      string           `json:"api_key"`
	Metrics     dashboardMetrics `json:"metrics"`
	storage     storage.Storage
}

func loadDashboard(dashid string, store storage.Storage) (*dashboard, error) {
	data, err := store.Get(dashid)
	if err != nil {
		return &dashboard{}, errors.New("Dashboard not found")
	}

	tmp := &dashboard{
		DashboardID: dashid,
		storage:     store,
	}
	_ = json.Unmarshal(data, tmp)

	return tmp, nil
}

func (d *dashboard) Save() {
	data, err := json.Marshal(d)
	if err != nil {
		log.Printf("Error while marshalling dashboard: %s", err)
		return
	}
	err = d.storage.Put(d.DashboardID, data)
	if err != nil {
		log.Printf("Error while storing dashboard: %s", err)
	}
}

type dashboardMetrics []*dashboardMetric

func (a dashboardMetrics) Len() int      { return len(a) }
func (a dashboardMetrics) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a dashboardMetrics) Less(i, j int) bool {
	return a[i].HistoricalData[0].Time.Before(a[j].HistoricalData[0].Time)
}

type dashboardMetric struct {
	MetricID       string                 `json:"id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Status         string                 `json:"status"`
	Value          float64                `json:"value,omitifempty"`
	Expires        int64                  `json:"expires,omitifempty"`
	Freshness      int64                  `json:"freshness,omitifempty"`
	HistoricalData dashboardMetricHistory `json:"history,omitifempty"`
	Meta           dashboardMetricMeta    `json:"meta,omitifempty"`
}

type dashboardMetricStatus struct {
	Time   time.Time `json:"time"`
	Status string    `json:"status"`
	Value  float64   `json:"value"`
}

type dashboardMetricMeta struct {
	LastUpdate time.Time
	LastOK     time.Time
	PercOK     float64
	PercWarn   float64
	PercCrit   float64
}

type dashboardMetricHistory []dashboardMetricStatus

func newDashboardMetric() *dashboardMetric {
	return &dashboardMetric{
		Status:         "Unknown",
		Expires:        604800,
		Freshness:      3600,
		HistoricalData: dashboardMetricHistory{},
		Meta:           dashboardMetricMeta{},
	}
}

func median(values []float64) float64 {
	sort.Float64s(values)

	if len(values) == 1 {
		return values[0]
	}

	// If even, take an average
	if len(values)%2 == 0 {
		return 0.5*values[len(values)/2] + 0.5*values[len(values)/2-1]
	}

	log.Printf("len(values)=%v, len(values)/2=%v\n", len(values), len(values)/2)

	return values[len(values)/2-1]
}

func absoluteValue(value float64) float64 {
	if value < 0 {
		value = -value
	}
	return value
}

func absoluteDeviation(values []float64) []float64 {
	medianValue := median(values)

	deviation := make([]float64, len(values))

	for i := range values {
		deviation[i] = absoluteValue(values[i] - medianValue)
	}

	return deviation
}

func (dm *dashboardMetric) getValueArray() []float64 {
	values := []float64{}

	for _, v := range dm.HistoricalData {
		values = append(values, v.Value)
	}

	return values
}

func (dm *dashboardMetric) Median() float64 {
	return median(dm.getValueArray())
}

func (dm *dashboardMetric) MedianAbsoluteDeviation() (float64, float64) {
	values := dm.getValueArray()
	medianValue := dm.Median()

	return medianValue, median(absoluteDeviation(values))
}

func (dm *dashboardMetric) MadMultiplier() float64 {
	medianValue, MAD := dm.MedianAbsoluteDeviation()

	return absoluteValue(dm.Value-medianValue) / MAD
}

func (dm *dashboardMetric) StatisticalStatus() string {
	mult := dm.MadMultiplier()

	if mult > 4 {
		return "Critical"
	} else if mult > 3 {
		return "Warning"
	}

	return "OK"
}

func (dm *dashboardMetric) LabelHistory() string {
	s := "["
	for i, v := range dm.HistoricalData {
		if i != 0 {
			s = s + ", "
		}
		s = s + "" + strconv.Itoa(int(v.Time.Unix())) + ""
	}
	s = s + "]"
	return s
}

func (dm *dashboardMetric) DataHistory() string {
	s := "["
	for i, v := range dm.HistoricalData {
		if i != 0 {
			s = s + ", "
		}
		s = s + strconv.FormatFloat(v.Value, 'g', 4, 64)
	}
	s = s + "]"
	return s
}

func (dm *dashboardMetric) Update(m *dashboardMetric) {
	dm.Title = m.Title
	dm.Description = m.Description
	dm.Status = m.Status
	dm.Value = m.Value
	if m.Expires != 0 {
		dm.Expires = m.Expires
	}
	if m.Freshness != 0 {
		dm.Freshness = m.Freshness
	}
	dm.HistoricalData = append(dm.HistoricalData, dashboardMetricStatus{
		Time:   time.Now(),
		Status: m.Status,
		Value:  m.Value,
	})

	countStatus := make(map[string]float64)

	expired := time.Now().Add(time.Duration(dm.Expires*-1) * time.Second)
	tmp := dashboardMetricHistory{}
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

func (dm *dashboardMetric) IsValid() (bool, string) {
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
