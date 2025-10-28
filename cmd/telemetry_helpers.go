package cmd

import (
	"github.com/raucheacho/rosia-cli/internal/telemetry"
	"github.com/raucheacho/rosia-cli/pkg/logger"
)

// getTelemetryStatsPath returns the path to the telemetry stats file
func getTelemetryStatsPath() (string, error) {
	return telemetry.GetDefaultStatsPath()
}

// initTelemetryStore initializes a telemetry store at the given path
func initTelemetryStore(statsPath string) (telemetry.TelemetryStore, error) {
	store, err := telemetry.NewFileStore(statsPath)
	if err != nil {
		logger.Warn("Failed to initialize telemetry store: %v", err)
		return nil, err
	}
	return store, nil
}
