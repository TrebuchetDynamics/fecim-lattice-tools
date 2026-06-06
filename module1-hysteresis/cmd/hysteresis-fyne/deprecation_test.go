package hysteresiscli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFyneCompatibilityCommandHelpPointsToCanonicalDesktopShell(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Run([]string{"--help"}); err != nil {
			t.Fatalf("Run(--help): %v", err)
		}
	})

	assertFyneCompatibilityNotice(t, output)
	for _, stale := range []string{
		"go" + "gpu/ui",
		"CGO_ENABLED=0",
		"-tags legacy_fyne",
	} {
		if strings.Contains(output, stale) {
			t.Fatalf("Fyne compatibility help still points to stale migration text %q in output:\n%s", stale, output)
		}
	}
}

func TestFyneCompatibilityCommandListMaterialsPointsToCanonicalDesktopShell(t *testing.T) {
	output := captureStdout(t, func() {
		if err := Run([]string{"--list-materials"}); err != nil {
			t.Fatalf("Run(--list-materials): %v", err)
		}
	})

	assertFyneCompatibilityNotice(t, output)
	if !strings.Contains(output, "Available materials") {
		t.Fatalf("list materials output missing material listing:\n%s", output)
	}
}

func TestFyneCompatibilityCommandDoesNotDuplicateEmbeddedModuleGUI(t *testing.T) {
	body, err := os.ReadFile(filepath.Join("main.go"))
	if err != nil {
		t.Fatalf("read compatibility command source: %v", err)
	}
	if strings.Contains(string(body), "module1-hysteresis/pkg/gui") {
		t.Fatalf("hysteresis-fyne compatibility command must point at the canonical shell, not duplicate the embedded Fyne GUI package")
	}
}

func TestFyneCompatibilityCommandDefaultInvocationFailsFastToCanonicalShell(t *testing.T) {
	output := captureStdout(t, func() {
		err := Run(nil)
		if err == nil {
			t.Fatal("Run(nil) succeeded; compatibility invocation must fail fast to the canonical desktop shell")
		}
		if !strings.Contains(err.Error(), "compatibility shim") ||
			!strings.Contains(err.Error(), "go run ./cmd/fecim-lattice-tools --module hysteresis") {
			t.Fatalf("default compatibility error did not point to canonical Fyne path: %v", err)
		}
	})

	assertFyneCompatibilityNotice(t, output)
	if strings.Contains(output, "go"+"gpu/ui") || strings.Contains(output, "CGO_ENABLED=0") {
		t.Fatalf("compatibility default invocation still points at retired UI migration text:\n%s", output)
	}
}

func assertFyneCompatibilityNotice(t *testing.T, output string) {
	t.Helper()
	for _, want := range []string{
		"Fyne compatibility command",
		"canonical desktop path",
		"go run ./cmd/fecim-lattice-tools --module hysteresis",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("Fyne compatibility output missing %q in output:\n%s", want, output)
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
