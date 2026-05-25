//go:build legacy_fyne

package hysteresiscli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestLegacyFyneCommandHelpDeclaresDeprecation(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Run([]string{"--help"}); err != nil {
			t.Fatalf("Run(--help): %v", err)
		}
	})

	assertLegacyFyneDeprecationNotice(t, output)
	for _, stale := range []string{
		"recommended",
		"GPU accelerated",
	} {
		if strings.Contains(output, stale) {
			t.Fatalf("legacy Fyne help still markets deprecated UI with %q in output:\n%s", stale, output)
		}
	}
}

func TestLegacyFyneCommandListMaterialsDeclaresDeprecation(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Run([]string{"--list-materials"}); err != nil {
			t.Fatalf("Run(--list-materials): %v", err)
		}
	})

	assertLegacyFyneDeprecationNotice(t, output)
	if !strings.Contains(output, "Available materials") {
		t.Fatalf("list materials output missing material listing:\n%s", output)
	}
}

func assertLegacyFyneDeprecationNotice(t *testing.T, output string) {
	t.Helper()
	for _, want := range []string{
		"DEPRECATED",
		"legacy Fyne",
		"gogpu/ui",
		"CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools --module hysteresis",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("legacy Fyne output missing %q in output:\n%s", want, output)
		}
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = old }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe: %v", err)
	}
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("read stdout pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return buf.String()
}
