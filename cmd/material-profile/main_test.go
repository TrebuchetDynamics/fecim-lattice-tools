package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(&out, &errOut, "pr", "version", "\n")
	if code != 0 {
		t.Fatalf("code=%d, want 0", code)
	}
	if strings.TrimSpace(out.String()) == "" {
		t.Fatal("expected non-empty version output")
	}
	if errOut.Len() != 0 {
		t.Fatalf("unexpected stderr: %q", errOut.String())
	}
}

func TestRunUnknownMode(t *testing.T) {
	var out, errOut bytes.Buffer
	code := run(&out, &errOut, "pr", "bad", "\n")
	if code != 2 {
		t.Fatalf("code=%d, want 2", code)
	}
	if !strings.Contains(errOut.String(), "unknown mode") {
		t.Fatalf("stderr missing unknown mode: %q", errOut.String())
	}
}

func TestRunMaterialProfileReportsFlagErrorWithoutExiting(t *testing.T) {
	var out, errOut bytes.Buffer

	code := runMaterialProfile([]string{"-definitely-not-a-flag"}, &out, &errOut)

	if code != 2 {
		t.Fatalf("code=%d, want 2; stderr=%q", code, errOut.String())
	}
	if out.Len() != 0 {
		t.Fatalf("stdout=%q, want empty output", out.String())
	}
	if !strings.Contains(errOut.String(), "flag provided but not defined") {
		t.Fatalf("stderr=%q, want flag parse context", errOut.String())
	}
}
