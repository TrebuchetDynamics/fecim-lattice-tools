//go:build display

package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestScreenshotterFailureStopsAutomaticDisplay catches error exits that bypass
// the deferred cleanup of the real Xvfb process owned by the command.
func TestScreenshotterFailureStopsAutomaticDisplay(t *testing.T) {
	xvfb, err := exec.LookPath("Xvfb")
	if err != nil {
		t.Skip("Xvfb is required")
	}
	binDir := t.TempDir()
	binary := filepath.Join(binDir, "screenshotter")
	if output, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, output)
	}
	// Record the actual server PID without replacing its lifecycle behavior.
	wrapper := "#!/bin/sh\nprintf '%s' \"$$\" > \"$FECIM_TEST_XVFB_PID\"\nexec \"$FECIM_TEST_REAL_XVFB\" \"$@\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "Xvfb"), []byte(wrapper), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"unknown-module", "invalid-output-directory", "single-capture-failure", "child-capture-failure"} {
		t.Run(mode, func(t *testing.T) {
			pidFile := filepath.Join(t.TempDir(), "server.pid")
			outDir := t.TempDir()
			only := "not-a-module"
			wantExit := 2
			if mode == "invalid-output-directory" {
				outDir = filepath.Join(outDir, "file")
				if err := os.WriteFile(outDir, nil, 0o600); err != nil {
					t.Fatal(err)
				}
				only = "docs"
			}
			if mode == "single-capture-failure" || mode == "child-capture-failure" {
				wantExit = 1
				only = "docs"
				module := "docs"
				if mode == "child-capture-failure" {
					only = ""
					module = "hysteresis"
				}
				if err := os.Mkdir(filepath.Join(outDir, module+"_initial.png"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "-out", outDir, "-only", only)
			cmd.Env = append(os.Environ(), "DISPLAY=", "WAYLAND_DISPLAY=", "PATH="+binDir+":"+os.Getenv("PATH"), "FECIM_TEST_XVFB_PID="+pidFile, "FECIM_TEST_REAL_XVFB="+xvfb, "FECIM_DISABLE_CALIBRATION_SAVE=1")
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("command timed out: %s", output)
			}
			if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != wantExit {
				t.Fatalf("want exit status %d, got %v: %s", wantExit, err, output)
			}
			data, err := os.ReadFile(pidFile)
			if err != nil {
				t.Fatalf("automatic display did not start: %v: %s", err, output)
			}
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			if err != nil || pid <= 0 {
				t.Fatalf("invalid server PID %q", data)
			}
			if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
				_ = syscall.Kill(pid, syscall.SIGKILL)
				t.Fatalf("automatic Xvfb process %d survived command failure (probe: %v)", pid, err)
			}
		})
	}
}

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
