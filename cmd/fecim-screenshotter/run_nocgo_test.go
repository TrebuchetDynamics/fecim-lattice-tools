//go:build !cgo

package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunScreenshotterReportsInvalidDimensionsWithoutExiting(t *testing.T) {
	var stderr bytes.Buffer

	code := runScreenshotter([]string{"-w", "0"}, &stderr)

	if code != 1 {
		t.Fatalf("exit code=%d, want 1; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "screenshot dimensions must be positive") {
		t.Fatalf("stderr=%q, want dimensions context", stderr.String())
	}
}
