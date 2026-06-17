package main

import (
	"flag"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()

	// Isolate hysteresis log output to a temp dir so concurrent package runs
	// (e.g. alongside cmd/fecim-lattice-tools-fyne) cannot cross-contaminate
	// timestamp-based log discovery in newestHysteresisCSVAfter / newestHysteresisLogAfter.
	logDir, err := os.MkdirTemp("", "fecim-cmd-logs-*")
	if err == nil {
		_ = os.Setenv("FECIM_LOGS_DIR", logDir)
		defer os.RemoveAll(logDir)
	}

	_ = os.Setenv("FECIM_DISABLE_CALIBRATION_SAVE", "1")
	_ = os.Setenv("FECIM_DISABLE_STARTUP_CALIBRATION", "1")

	os.Exit(m.Run())
}
