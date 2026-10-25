//go:build display

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// TestFullScreenshotterCapturesEveryModuleSequentially exercises the public
// screenshot command instead of constructing modules one at a time in tests.
// It is intentionally strict: a GUI lifecycle panic or a missing module frame
// must fail the full-app smoke test.
func TestFullScreenshotterCapturesEveryModuleSequentially(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping full GUI smoke test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	outDir := t.TempDir()
	cmd := exec.CommandContext(ctx, "go", "run", ".", "-out", outDir, "-w", "640", "-h", "480", "-tag", "full-app-test")
	cmd.Env = append(cmd.Environ(), "FECIM_DISABLE_CALIBRATION_SAVE=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("full screenshot command failed: %v\n%s", err, output)
	}
	if ctx.Err() != nil {
		t.Fatalf("full screenshot command timed out: %v\n%s", ctx.Err(), output)
	}

	for _, module := range []string{"hysteresis", "crossbar", "mnist", "circuits", "comparison", "eda", "docs"} {
		path := filepath.Join(outDir, module+"_full-app-test.png")
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("full screenshot command did not create %s frame: %v\noutput:\n%s", module, err, output)
		}
		if info.Size() == 0 {
			t.Fatalf("full screenshot command created an empty %s frame\noutput:\n%s", module, output)
		}
	}
}
