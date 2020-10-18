package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"

	mondash "github.com/Luzifer/mondash/client"
	"github.com/Luzifer/rconfig/v2"
)

var (
	cfg = struct {
		MondashHost    string        `flag:"host,h" env:"HOST" default:"https://mondash.org" description:"Change to use an on-premise instance of MonDash"`
		MondashBoard   string        `flag:"board,b" env:"BOARD" default:"" description:"ID of the board to submit the metric to" validate:"nonzero"`
		MondashToken   string        `flag:"token,t" env:"TOKEN" default:"" description:"Token associated with the specified board" validate:"nonzero"`
		MetricID       string        `flag:"metric-id,m" env:"METRIC_ID" default:"" description:"ID of the metric, if not specified a generated ID from the title will be used"`
		MetricTitle    string        `flag:"metric-title" env:"METRIC_TITLE" default:"" description:"Title of the metric, if not specified the command line will be used"`
		Freshness      time.Duration `flag:"freshness" env:"FRESHNESS" default:"1h" description:"Freshness of the metric (will turn to unknown if no new result was submitted)"`
		StaleStatus    string        `flag:"stale-status" default:"Unknown" description:"Status to set when metric is stale (One of Unknown, OK, Warning, Critical)"`
		Timeout        time.Duration `flag:"timeout" env:"TIMEOUT" default:"1m" description:"Timeout for the script command to be killed"`
		VersionAndExit bool          `flag:"version" default:"false" description:"Prints current version and exits"`
	}{}

	statusMapping = map[int]mondash.Status{
		0: mondash.StatusOK,
		1: mondash.StatusWarning,
		2: mondash.StatusCritical,
		3: mondash.StatusUnknown,
	}

	version = "dev"
)

func init() {
	rconfig.AutoEnv(true)
	if err := rconfig.ParseAndValidate(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	if cfg.VersionAndExit {
		fmt.Printf("mondash-nagios %s\n", version)
		os.Exit(0)
	}
}

func main() {
	command := rconfig.Args()[1:]
	if len(command) == 0 {
		log.Fatal("Please specify a command to execute")
	}

	if cfg.MetricTitle == "" {
		cfg.MetricTitle = strings.Join(command, " ")
	}

	if cfg.MetricID == "" {
		cfg.MetricID = slug.Make(cfg.MetricTitle)
	}

	var (
		ctx    = context.Background()
		cancel context.CancelFunc
	)
	if cfg.Timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), cfg.Timeout)
		defer cancel()
	}

	outputBuffer := new(bytes.Buffer)

	cmd := exec.CommandContext(ctx, command[0], command[1:]...)
	cmd.Stdout = outputBuffer
	cmd.Stderr = os.Stderr

	exitCode := 0
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			log.Warn("Could not get exit code for failed program, using UNKNOWN for reporting")
			exitCode = 3 // Unknown
		}
	}

	if _, ok := statusMapping[exitCode]; !ok {
		log.WithFields(log.Fields{"exit-code": exitCode}).Warn("Undefined exit code, using UNKNOWN for reporting")
		exitCode = 3 // Unknown
	}

	output, value := parseStdout(outputBuffer.String())
	if output == "" {
		output = fmt.Sprintf("exit %d", exitCode)
	}

	client := mondash.New(cfg.MondashBoard, cfg.MondashToken).WithHost(cfg.MondashHost)
	if err := client.PostMetric(&mondash.PostMetricInput{
		MetricID:        cfg.MetricID,
		Title:           cfg.MetricTitle,
		Description:     output,
		Status:          statusMapping[exitCode],
		Value:           value,
		Freshness:       int64(cfg.Freshness / time.Second),
		IgnoreMAD:       true,
		HideMAD:         true,
		StalenessStatus: cfg.StaleStatus,
	}); err != nil {
		log.WithError(err).Fatal("Could not submit metric")
	}
}

func parseStdout(stdout string) (string, float64) {
	// Drop everything after first line
	stdout = strings.SplitN(stdout, "\n", 2)[0]

	// Split output from perf data
	parts := strings.SplitN(stdout, "|", 2)

	// Return in case there is no perf data
	if len(parts) == 1 {
		return parts[0], 0
	}

	// Save the output
	output := strings.TrimSpace(parts[0])

	var (
		perfData = map[string]float64{}
		fallback string
	)

	// Parse perf data included in the output
	perfParts := strings.Split(strings.TrimSpace(parts[1]), ", ")
	for i, part := range perfParts {
		tmp := strings.SplitN(part, "=", 2)
		if len(tmp) != 2 {
			continue
		}

		name := tmp[0]
		value, err := strconv.ParseFloat(tmp[1], 64)
		if err != nil {
			log.WithFields(log.Fields{"perf-part": part}).WithError(err).Error("Unable to parse perf data")
			continue
		}

		// In case there is a data part "value" use this one
		if name == "value" {
			return output, value
		}

		// Store first seen metric as "fallback"
		if i == 0 {
			fallback = name
		}
		perfData[name] = value
	}

	// Return output with first seen metric
	return output, perfData[fallback]
}
