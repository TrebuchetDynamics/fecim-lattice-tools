package scripts_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLaunchPreservesDesktopAndHeadlessModes(t *testing.T) {
	// Execute the real launcher, replacing only the compiler and built application.
	// Forcing CGO off must fail the default cases; overriding callers must fail
	// the explicit headless case. No display or repository data is used.
	launcher := readFile(t, filepath.Join(repoRoot(t), "launch.sh"))
	for _, tc := range []struct {
		name        string
		cgo         string // "unset" removes the variable entirely.
		wantCGO     string
		args        []string
		wantArgs    []string
		clear       bool
		buildStatus string
		appStatus   string
		wantExit    int
	}{
		{name: "unset desktop", cgo: "unset", wantCGO: "1", args: []string{"--logger", "space value", "", "*.json"}, wantArgs: []string{"--logger", "space value", "", "*.json"}},
		{name: "empty desktop", cgo: "", wantCGO: "1"},
		{name: "explicit desktop", cgo: "1", wantCGO: "1"},
		{name: "explicit headless", cgo: "0", wantCGO: "0", args: []string{"experiment", "run", "project with spaces.json"}, wantArgs: []string{"experiment", "run", "project with spaces.json"}},
		{name: "clear only documented paths", cgo: "0", wantCGO: "0", clear: true, args: []string{"--clear", "report", "--clear", "space value", "--clear=keep"}, wantArgs: []string{"report", "space value", "--clear=keep"}},
		{name: "build failure never launches", cgo: "1", wantCGO: "1", buildStatus: "37", wantExit: 1},
		{name: "application exit propagates", cgo: "1", wantCGO: "1", appStatus: "23", wantExit: 23},
	} {
		t.Run(tc.name, func(t *testing.T) {
			base := t.TempDir()
			root := filepath.Join(base, "checkout with spaces")
			mustWrite(t, filepath.Join(root, "launch.sh"), launcher)
			application := `#!/bin/bash
: > "$FIXTURE/app-args"
for arg in "$@"; do printf '%s\0' "$arg" >> "$FIXTURE/app-args"; done
exit "${APP_STATUS:-0}"
`
			mustWrite(t, filepath.Join(root, "application"), application)
			// A stale executable must not be launched after a failed rebuild.
			mustWrite(t, filepath.Join(root, "fecim-lattice-tools"), application)
			mustWrite(t, filepath.Join(root, "bin", "go"), `#!/bin/bash
printf '%s' "$CGO_ENABLED" > "$FIXTURE/build-cgo"
printf '%s\0' "$@" > "$FIXTURE/build-args"
if [[ "${BUILD_STATUS:-0}" != 0 ]]; then exit "$BUILD_STATUS"; fi
cp "$FIXTURE/application" fecim-lattice-tools
`)
			removed := []string{"logs/sentinel", "docs/assets/screenshots/sentinel"}
			kept := []string{"docs/assets/reference-screenshots/sentinel", "docs/assets/keep", "screenshots/sentinel", "recordings/sentinel"}
			for _, path := range append(append([]string{}, removed...), kept...) {
				mustWrite(t, filepath.Join(root, path), "keep")
			}
			mustWrite(t, filepath.Join(base, "logs", "sentinel"), "outside checkout")

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "bash", append([]string{filepath.Join(root, "launch.sh")}, tc.args...)...)
			cmd.Dir = base // Cleanup/build must be relative to the script, not the caller.
			for _, env := range os.Environ() {
				if !strings.HasPrefix(env, "CGO_ENABLED=") && !strings.HasPrefix(env, "PATH=") {
					cmd.Env = append(cmd.Env, env)
				}
			}
			cmd.Env = append(cmd.Env, "PATH="+filepath.Join(root, "bin")+string(os.PathListSeparator)+os.Getenv("PATH"), "FIXTURE="+root, "BUILD_STATUS="+tc.buildStatus, "APP_STATUS="+tc.appStatus)
			if tc.cgo != "unset" {
				cmd.Env = append(cmd.Env, "CGO_ENABLED="+tc.cgo)
			}
			out, err := cmd.CombinedOutput()
			if ctx.Err() != nil || cmd.ProcessState == nil {
				t.Fatalf("launcher did not finish: %v; %s", err, out)
			}
			if got := cmd.ProcessState.ExitCode(); got != tc.wantExit {
				t.Errorf("exit = %d, want %d; %s", got, tc.wantExit, out)
			}
			if got := readFile(t, filepath.Join(root, "build-cgo")); got != tc.wantCGO {
				t.Errorf("build CGO_ENABLED = %q, want %q", got, tc.wantCGO)
			}
			if got := readFile(t, filepath.Join(root, "build-args")); got != "build\x00-v\x00-o\x00fecim-lattice-tools\x00./cmd/fecim-lattice-tools\x00" {
				t.Errorf("unexpected build arguments: %q", got)
			}
			if tc.buildStatus != "" {
				if _, err := os.Stat(filepath.Join(root, "app-args")); !os.IsNotExist(err) {
					t.Fatalf("failed build launched application: %v", err)
				}
			} else {
				want := ""
				if len(tc.wantArgs) > 0 {
					want = strings.Join(tc.wantArgs, "\x00") + "\x00"
				}
				if got := readFile(t, filepath.Join(root, "app-args")); got != want {
					t.Errorf("application arguments = %q, want %q", got, want)
				}
			}
			for _, path := range removed {
				_, err := os.Stat(filepath.Join(root, path))
				if tc.clear {
					if !os.IsNotExist(err) {
						t.Errorf("--clear did not remove %s: %v", path, err)
					}
				} else if err != nil {
					t.Errorf("removed %s without --clear: %v", path, err)
				}
			}
			for _, path := range kept {
				if got := readFile(t, filepath.Join(root, path)); got != "keep" {
					t.Errorf("changed %s: %q", path, got)
				}
			}
			if got := readFile(t, filepath.Join(base, "logs", "sentinel")); got != "outside checkout" {
				t.Errorf("changed caller's logs: %q", got)
			}
		})
	}
}
