package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runWorkbenchCommand(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := runMain(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func TestProjectCommandInitAndValidate(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "study")
	stdout, stderr, code := runWorkbenchCommand(t, "project", "init", dir)
	if code != 0 || stderr != "" || !strings.Contains(stdout, "initialized") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, name := range []string{"project.yaml", "design.yaml", "sweep.yaml"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("missing %s: %v", name, err)
		}
	}
	stdout, stderr, code = runWorkbenchCommand(t, "project", "validate", dir)
	if code != 0 || stderr != "" || !strings.Contains(stdout, "valid: hzo-study") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	_, _, code = runWorkbenchCommand(t, "project", "init", dir)
	if code == 0 {
		t.Fatal("second init unexpectedly succeeded")
	}
}

func TestProjectValidateRejectsUnsupportedSweepPath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "study")
	if _, stderr, code := runWorkbenchCommand(t, "project", "init", dir); code != 0 {
		t.Fatalf("init code=%d stderr=%q", code, stderr)
	}
	sweep := `schema_version: 1
seed: 17
max_points: 32
parameters:
  - path: device.unsupported
    values: [1]
`
	if err := os.WriteFile(filepath.Join(dir, "sweep.yaml"), []byte(sweep), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code := runWorkbenchCommand(t, "project", "validate", dir)
	if code == 0 || strings.Contains(stdout, "valid:") || !strings.Contains(stderr, "unsupported sweep path") {
		t.Fatalf("code=%d stdout=%q stderr=%q; unsupported sweep must fail validation", code, stdout, stderr)
	}
}

func TestExperimentSubcommandHonorsCancelledContext(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "study")
	if _, stderr, code := runWorkbenchCommand(t, "project", "init", dir); code != 0 {
		t.Fatalf("init code=%d stderr=%q", code, stderr)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	err := runExperimentSubcommandContext(ctx, []string{"run", dir}, &stdout, &stderr)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v, want context.Canceled", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "runs")); !os.IsNotExist(err) {
		t.Fatalf("cancelled experiment created run storage: %v", err)
	}
}

func TestExperimentAndReportCommandsEndToEnd(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "study")
	if _, stderr, code := runWorkbenchCommand(t, "project", "init", dir); code != 0 {
		t.Fatalf("init code=%d stderr=%q", code, stderr)
	}
	stdout, stderr, code := runWorkbenchCommand(t, "experiment", "run", dir, "-workers", "2")
	if code != 0 || stderr != "" {
		t.Fatalf("run code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, phrase := range []string{"total=8", "feasible=", "pareto="} {
		if !strings.Contains(stdout, phrase) {
			t.Fatalf("stdout=%q missing %q", stdout, phrase)
		}
	}
	entries, err := os.ReadDir(filepath.Join(dir, "runs"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 8 {
		t.Fatalf("run directories=%d want 8", len(entries))
	}
	stdout, stderr, code = runWorkbenchCommand(t, "experiment", "run", dir, "-workers", "4")
	if code != 0 || stderr != "" || !strings.Contains(stdout, "reused=8") {
		t.Fatalf("reuse code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	stdout, stderr, code = runWorkbenchCommand(t, "report", "generate", dir)
	if code != 0 || stderr != "" || !strings.Contains(stdout, "generated") {
		t.Fatalf("report code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, name := range []string{"results.json", "results.csv", "report.md"} {
		if _, err := os.Stat(filepath.Join(dir, "reports", name)); err != nil {
			t.Fatalf("missing report %s: %v", name, err)
		}
	}
	runs, err := os.ReadDir(filepath.Join(dir, "runs"))
	if err != nil || len(runs) == 0 {
		t.Fatalf("run records=%d err=%v", len(runs), err)
	}
	stdout, stderr, code = runWorkbenchCommand(t, "export", dir, "-run", runs[0].Name())
	if code != 0 || stderr != "" || !strings.Contains(stdout, "exported:") {
		t.Fatalf("export code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	for _, name := range []string{"design.spice", "design.v", "design.def", "cell.lef", "cell.lib", "manifest.json"} {
		if _, err := os.Stat(filepath.Join(dir, "exports", runs[0].Name(), name)); err != nil {
			t.Fatalf("missing selected-run export %s: %v", name, err)
		}
	}
}

func TestWorkbenchUnknownNestedCommandDoesNotWriteScientificOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	handled, code := runSubcommandDispatch([]string{"experiment", "unknown"}, &stdout, &stderr)
	if !handled || code == 0 || stdout.Len() != 0 {
		t.Fatalf("handled=%v code=%d stdout=%q stderr=%q", handled, code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "unknown experiment subcommand") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}

func TestWorkbenchDispatchAcceptsInjectedWriters(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := dispatchSubcommandWithWriters([]string{"project", "init"}, &stdout, &stderr); err == nil {
		t.Fatal("missing directory unexpectedly succeeded")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout=%q", stdout.String())
	}
}
