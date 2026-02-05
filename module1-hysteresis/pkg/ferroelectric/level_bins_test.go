package ferroelectric

import (
	"math"
	"testing"
)

func TestLevelBinsRangeFracAffectsStep(t *testing.T) {
	bins := NewLevelBins(1.0, 3, 0.8, 0)
	want := 0.8
	if got := bins.Step(); math.Abs(got-want) > 1e-9 {
		t.Fatalf("Step() mismatch: got %.6f want %.6f", got, want)
	}
}

func TestLevelBinsLevelForPUsesEffectivePs(t *testing.T) {
	bins := NewLevelBins(1.0, 3, 0.8, 0)

	level, inError, delta := bins.LevelForP(0.8)
	if level != 3 {
		t.Fatalf("LevelForP(0.8) level=%d want=3", level)
	}
	if inError {
		t.Fatalf("LevelForP(0.8) inError=true want=false")
	}
	if math.Abs(delta) > 1e-9 {
		t.Fatalf("LevelForP(0.8) delta=%.6f want=0", delta)
	}

	level, inError, delta = bins.LevelForP(1.0)
	if level != 3 {
		t.Fatalf("LevelForP(1.0) level=%d want=3", level)
	}
	if inError {
		t.Fatalf("LevelForP(1.0) inError=true want=false")
	}
	if math.Abs(delta) > 1e-9 {
		t.Fatalf("LevelForP(1.0) delta=%.6f want=0", delta)
	}
}

func TestLevelBinsGuardUsesEffectiveStep(t *testing.T) {
	bins := NewLevelBins(1.0, 3, 0.8, 0.25)
	_, inError, _ := bins.LevelForP(0.3)
	if !inError {
		t.Fatalf("LevelForP(0.3) inError=false want=true")
	}
}
