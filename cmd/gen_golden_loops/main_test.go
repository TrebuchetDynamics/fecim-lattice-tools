package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

func TestSafeFilename(t *testing.T) {
	got := safeFilename("HZO FTJ (140 states)")
	if got == "" {
		t.Fatal("safeFilename returned empty")
	}
}

func TestGeneratePreisachLoopShape(t *testing.T) {
	mat := physics.DefaultHZO()
	m := ferroelectric.NewPreisachModel(mat)
	E, P := generatePreisachLoop(m, 2*mat.Ec, 25)
	if len(E) != 25 || len(P) != 25 {
		t.Fatalf("unexpected loop lengths E=%d P=%d", len(E), len(P))
	}
	if E[0] >= E[len(E)-1] {
		t.Fatalf("expected increasing field sweep, got E0=%g Elast=%g", E[0], E[len(E)-1])
	}
}

func TestRunReportsOutputDirectoryError(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(outPath, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write output placeholder: %v", err)
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := runGenGoldenLoops([]string{"-output", outPath}, &stdout, &stderr)

	if code != 1 {
		t.Fatalf("exit code=%d, want 1; stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stderr.String(), "prepare output directory") {
		t.Fatalf("stderr=%q, want output-directory context", stderr.String())
	}
	if strings.Contains(stderr.String(), "panic") {
		t.Fatalf("stderr=%q, must not include panic output", stderr.String())
	}
}
