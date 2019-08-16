package main

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/Luzifer/go_helpers/str"

	"github.com/Luzifer/mondash/storage"
)

const defaultStalenessStatus = metricStatusUnknown

var errDashboardNotFound = errors.New("Dashboard not found")

// --- Metric Status ---

type metricStatus uint

var metricStatusStringMapping = []string{
	"OK",
	"Warning",
	"Critical",
	"Unknown",
	"Total",
}

const (
	// Nagios status mappings
	metricStatusOK metricStatus = iota
	metricStatusWarning
	metricStatusCritical
	metricStatusUnknown
	metricStatusTotal // Only internally used
)

func metricStatusFromString(in string) metricStatus {
	for i, v := range metricStatusStringMapping {
		if v == in {
			return metricStatus(i)
		}
	}

	return metricStatusUnknown
}

func (m metricStatus) String() string {
	return metricStatusStringMapping[m]
}

// --- Dashboard ---

type dashboard struct {
	DashboardID string             `json:"-"`
	APIKey      string             `json:"api_key"`
	Metrics     []*dashboardMetric `json:"metrics"`

	storage storage.Storage
}

func loadDashboard(dashid string, store storage.Storage) (*dashboard, error) {
	data, err := store.Get(dashid)
	if err != nil {
		return nil, errDashboardNotFound
	}

	tmp := &dashboard{
		DashboardID: dashid,
		storage:     store,
	}

	if err := json.Unmarshal(data, tmp); err != nil {
		return nil, errors.Wrap(err, "Unable to unmarshal dashboard")
	}

	tmp.migrate() // Do a load-migration, it will be applied on save

	return tmp, nil
}

func (d *dashboard) Save() error {
	data, err := json.Marshal(d)
	if err != nil {
		return errors.Wrap(err, "Unable to marshal dashboard")
	}

	return errors.Wrap(d.storage.Put(d.DashboardID, data), "Unable to store dashboard")
}

func (d *dashboard) migrate() {
	// Migrate metadata
	for _, m := range d.Metrics {
		if m.Meta.LastUpdate.IsZero() && !m.Meta.MIGLastUpdate.IsZero() {
			m.Meta.LastUpdate = m.Meta.MIGLastUpdate
			m.Meta.MIGLastUpdate = time.Time{}
		}

		if m.Meta.LastOK.IsZero() && !m.Meta.MIGLastOK.IsZero() {
			m.Meta.LastOK = m.Meta.MIGLastOK
			m.Meta.MIGLastOK = time.Time{}
		}
	}
}

// --- Dashboard Metric ---

type dashboardMetric struct {
	MetricID        string                  `json:"id"`
	Title           string                  `json:"title"`
	Description     string                  `json:"description"`
	DetailURL       string                  `json:"detail_url"`
	Status          string                  `json:"status"`
	Value           float64                 `json:"value,omitempty"`
	Expires         int64                   `json:"expires,omitempty"`
	Freshness       int64                   `json:"freshness,omitempty"`
	IgnoreMAD       bool                    `json:"ignore_mad"`
	HideMAD         bool                    `json:"hide_mad"`
	HideValue       bool                    `json:"hide_value"`
	HistoricalData  []dashboardMetricStatus `json:"history,omitempty"`
	Meta            dashboardMetricMeta     `json:"meta,omitempty"`
	StalenessStatus string                  `json:"staleness_status,omitempty"`
}

type dashboardMetricStatus struct {
	Time   time.Time `json:"time"`
	Status string    `json:"status"`
	Value  float64   `json:"value"`
}

type dashboardMetricMeta struct {
	LastUpdate time.Time `json:"last_update"`
	LastOK     time.Time `json:"last_ok"`
	PercOK     float64   `json:"perc_ok"`
	PercWarn   float64   `json:"perc_warn"`
	PercCrit   float64   `json:"perc_crit"`

	MIGLastUpdate time.Time `json:"LastUpdate,omitempty"`
	MIGLastOK     time.Time `json:"LastOK,omitempty"`
}

func newDashboardMetric() *dashboardMetric {
	return &dashboardMetric{
		Status:         defaultStalenessStatus.String(),
		Expires:        604800,
		Freshness:      3600,
		HistoricalData: []dashboardMetricStatus{},
		Meta:           dashboardMetricMeta{},
	}
}

func (dm dashboardMetric) getValueArray() []float64 {
	values := []float64{}

	for _, v := range dm.HistoricalData {
		values = append(values, v.Value)
	}

	return values
}

func (dm dashboardMetric) Median() float64 {
	return median(dm.getValueArray())
}

func (dm dashboardMetric) MedianAbsoluteDeviation() (float64, float64) {
	values := dm.getValueArray()
	medianValue := dm.Median()

	return medianValue, median(absoluteDeviation(values))
}

func (dm dashboardMetric) MadMultiplier() float64 {
	medianValue, MAD := dm.MedianAbsoluteDeviation()

	if MAD == 0 {
		// Edge-case, causes div-by-zero
		return 1
	}

	return math.Abs(dm.Value-medianValue) / MAD
}

func (dm dashboardMetric) StatisticalStatus() string {
	mult := dm.MadMultiplier()

	switch {
	case mult > 4:
		return "Critical"

	case mult > 3:
		return "Warning"

	default:
		return "OK"
	}
}

func (dm dashboardMetric) PreferredStatus() string {
	// Metric might be stale, return stale status
	if dm.Meta.LastUpdate.Add(time.Duration(dm.Freshness) * time.Second).Before(time.Now()) {
		if dm.StalenessStatus == "" {
			return defaultStalenessStatus.String()
		}
		return dm.StalenessStatus
	}

	// If MAD is ignored use given status
	if dm.IgnoreMAD {
		return dm.Status
	}

	// By default use MAD for status
	return dm.StatisticalStatus()
}

func (dm dashboardMetric) HistoricalValueMap() map[int64]float64 {
	out := map[int64]float64{}

	start := int(math.Max(0, float64(len(dm.HistoricalData)-30)))
	for _, v := range dm.HistoricalData[start:] {
		out[v.Time.Unix()] = v.Value
	}

	return out
}

func (dm dashboardMetric) LabelHistory() []string {
	s := []string{}

	labelStart := len(dm.HistoricalData) - 60
	if labelStart < 0 {
		labelStart = 0
	}

	for _, v := range dm.HistoricalData[labelStart:] {
		s = append(s, strconv.Itoa(int(v.Time.Unix())))
	}

	return s
}

func (dm dashboardMetric) DataHistory() []string {
	s := []string{}

	dataStart := len(dm.HistoricalData) - 60
	if dataStart < 0 {
		dataStart = 0
	}

	for _, v := range dm.HistoricalData[dataStart:] {
		s = append(s, strconv.FormatFloat(v.Value, 'g', 4, 64))
	}

	return s
}

func (dm *dashboardMetric) Update(m *dashboardMetric) {
	dm.Title = m.Title
	dm.Description = m.Description
	dm.Status = m.Status
	dm.Value = m.Value
	dm.IgnoreMAD = m.IgnoreMAD
	dm.HideMAD = m.HideMAD
	dm.HideValue = m.HideValue
	dm.StalenessStatus = m.StalenessStatus

	if m.DetailURL != "" {
		dm.DetailURL = m.DetailURL
	}

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

	countStatus := make(map[metricStatus]float64)

	expired := time.Now().Add(time.Duration(dm.Expires*-1) * time.Second)
	tmp := []dashboardMetricStatus{}

	for _, s := range dm.HistoricalData {
		if s.Time.After(expired) {
			statusVal := metricStatusFromString(s.Status)

			tmp = append(tmp, s)

			countStatus[statusVal] = countStatus[statusVal] + 1
			countStatus[metricStatusTotal] = countStatus[metricStatusTotal] + 1

			if dm.Meta.LastOK.Before(s.Time) && statusVal == metricStatusOK {
				dm.Meta.LastOK = s.Time
			}
		}
	}

	dm.HistoricalData = tmp

	dm.Meta.LastUpdate = time.Now()
	if countStatus[metricStatusTotal] > 0 {
		dm.Meta.PercCrit = countStatus[metricStatusCritical] / countStatus[metricStatusTotal] * 100
		dm.Meta.PercWarn = countStatus[metricStatusWarning] / countStatus[metricStatusTotal] * 100
		dm.Meta.PercOK = countStatus[metricStatusOK] / countStatus[metricStatusTotal] * 100
	}
}

func (dm dashboardMetric) IsValid() (bool, string) {
	if dm.Expires > 604800 || dm.Expires < 0 {
		return false, "Expires not in range 0 < x < 640800"
	}

	if dm.Freshness > 604800 || dm.Freshness < 0 {
		return false, "Freshness not in range 0 < x < 640800"
	}

	if !str.StringInSlice(dm.Status, []string{"OK", "Warning", "Critical", "Unknowm"}) {
		return false, "Status not allowed"
	}

	if len(dm.Title) > 512 || len(dm.Description) > 1024 {
		return false, "Title or Description too long"
	}

	return true, ""
}

type historyBarSegment struct {
	Duration   time.Duration `json:"duration"`
	End        time.Time     `json:"end"`
	Percentage float64       `json:"percentage"`
	Start      time.Time     `json:"start"`
	Status     string        `json:"status"`
}

func (dm dashboardMetric) GetHistoryBar() []historyBarSegment {
	var (
		point     dashboardMetricStatus
		segLength int
		segments  = []historyBarSegment{}
		segStart  time.Time
		status    = defaultStalenessStatus
	)

	for _, point = range dm.HistoricalData {
		if metricStatusFromString(point.Status) == status {
			segLength++
			continue
		}

		// Store the old segment
		if segLength > 0 {
			segments = append(segments, historyBarSegment{
				Duration:   point.Time.Sub(segStart),
				End:        point.Time,
				Percentage: float64(segLength) / float64(len(dm.HistoricalData)),
				Start:      segStart,
				Status:     status.String(),
			})
		}

		// Start a new segment
		segLength = 1
		segStart = point.Time
		status = metricStatusFromString(point.Status)
	}

	segments = append(segments, historyBarSegment{
		Duration:   point.Time.Sub(segStart),
		End:        point.Time,
		Percentage: float64(segLength) / float64(len(dm.HistoricalData)),
		Start:      segStart,
		Status:     status.String(),
	})

	return segments
}
