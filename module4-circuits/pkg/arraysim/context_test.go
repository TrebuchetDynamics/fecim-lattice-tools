package arraysim

import (
	"context"
	"errors"
	"testing"
)

func TestSolveContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := SolveParams{WLVoltages: []float64{1}, BLVoltages: []float64{0}, Conductance: [][]float64{{1e-5}}}
	for _, engine := range []Engine{NewTierASolver(), NewTierBSolver()} {
		if _, err := SolveContext(ctx, engine, p); !errors.Is(err, context.Canceled) {
			t.Fatalf("cancelled solver returned %v", err)
		}
	}
}
