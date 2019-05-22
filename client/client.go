package mondash

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const defaultHost = "https://mondash.org"

// Status type represents a collection of available status strings
type Status string

// Collection of available status strings
const (
	StatusOK       Status = "OK"
	StatusWarning         = "Warning"
	StatusCritical        = "Critical"
	StatusUnknown         = "Unknown"
)

// Client represents an accessor to the MonDash API
type Client struct {
	host         string
	context      context.Context
	board, token string
}

// New creates a new Client pre-filled with board-ID and token
func New(boardID, token string) *Client {
	return &Client{
		board:   boardID,
		context: context.Background(),
		host:    defaultHost,
		token:   token,
	}
}

func (c *Client) do(method, path string, body io.Reader) error {
	req, err := http.NewRequest(method, c.host+path, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Token "+c.token)

	req = req.WithContext(c.context)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("Received unexpected status code %d", res.StatusCode)
	}

	return nil
}

// WithHost creates a copy of the Client with replaced hostname for own instances
func (c *Client) WithHost(host string) *Client {
	c2 := new(Client)
	*c2 = *c
	c2.host = host
	return c2
}

// WithContext craetes a copy of the Client using the passed context instead of context.Background()
func (c *Client) WithContext(ctx context.Context) *Client {
	if ctx == nil {
		panic("nil context")
	}
	c2 := new(Client)
	*c2 = *c
	c2.context = ctx
	return c2
}

// DeleteDashboard will delete all your monitoring results available on your dashboard and release the dashboard URL to the public
func (c *Client) DeleteDashboard() error {
	return c.do(http.MethodDelete, "/"+c.board, nil)
}

// PostMetricInput contains parameters for the API request
type PostMetricInput struct {
	// The unique name for your metric Example: `beer_available`.
	MetricID string `json:"-"`

	// The title of the metric to display on the dashboard
	Title string `json:"title"`

	// A descriptive text for the current state of the metric
	Description string `json:"description"`

	// One of: OK, Warning, Critical, Unknown
	Status Status `json:"status"`

	// The metric value to store with the status
	Value float64 `json:"value"`

	// Time in seconds when to remove the metric if there is no update (Valid: `0 < x < 604800`)
	// Default: `604800`
	Expires int64 `json:"expires,omitempty"`

	// Time in seconds when to switch to stale state of there is no update (Valid: `0 < x < 604800`)
	// Default: 3600
	Freshness int64 `json:"freshness,omitempty"`

	// If set to true the status passed in the update will be used instead of the median absolute deviation
	// Default: false
	IgnoreMAD bool `json:"ignore_mad,omitempty"`

	// If set to true the median absolute deviation is hidden on the dashboard for this metric
	// Default: false
	HideMAD bool `json:"hide_mad,omitempty"`

	// If set to true the value of the metric is not shown on the dashboard
	// Default: false
	HideValue bool `json:"hide_value,omitempty"`

	// If set this status will be set when the metric gets stale (no updates within freshness time range
	// Default: "Unknown"
	StalenessStatus string `json:"staleness_status,omitifempty"`
}

func (p *PostMetricInput) validate() error {
	if p.MetricID == "" {
		return errors.New("Field 'MetricID' is required.")
	}
	if p.Title == "" {
		return errors.New("Field 'Title' is required.")
	}
	if p.Description == "" {
		return errors.New("Field 'Description' is required.")
	}
	if p.Status == "" {
		return errors.New("Field 'Status' is required.")
	}
	return nil
}

// PostMetric submits a new monitoring result
func (c *Client) PostMetric(input *PostMetricInput) error {
	if err := input.validate(); err != nil {
		return err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(buf).Encode(input); err != nil {
		return err
	}

	return c.do(http.MethodPut, "/"+c.board+"/"+input.MetricID, buf)
}

// DeleteMetricInput contains parameters for the API request
type DeleteMetricInput struct {
	// The unique name for your metric Example: `beer_available`.
	MetricID string
}

// DeleteMetric deletes a metric from your dashboard
func (c *Client) DeleteMetric(input *DeleteMetricInput) error {
	return c.do(http.MethodDelete, "/"+c.board+"/"+input.MetricID, nil)
}
