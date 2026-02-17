package main

import (
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
