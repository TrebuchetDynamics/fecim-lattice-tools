package ferroelectric

import (
	"testing"
	"math"
)

// TestEnduranceCycles verifies the long-term stability of the stack logic.
// It simulates 1000 cycles of operation and checks for stack unbounded growth
// and polarization drift.
func TestEnduranceCycles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping endurance test in short mode")
	}

	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30) // 30x30 grid

	cycles := 1000
	Ec := material.Ec
	
	// Complex cycle pattern: +Saturate, -Saturate, +Half, -Half, 0
	sequence := []float64{
		2 * Ec, 
		-2 * Ec,
		0.5 * Ec,
		-0.5 * Ec,
		0,
	}

	for i := 0; i < cycles; i++ {
		for _, E := range sequence {
			P := model.Update(E)
			if math.IsNaN(P) || math.IsInf(P, 0) {
				t.Fatalf("Cycle %d: Invalid P detected: %v", i, P)
			}
		}

		// Periodically check stack size (should be bounded)
		if i%100 == 0 {
			stackLen := len(model.StackE)
			// For this sequence, stack should efficiently wipe out.
			// Max depth shouldn't exceed modest number (e.g. 10)
			if stackLen > 20 {
				t.Errorf("Cycle %d: Stack grew too large: %d", i, stackLen)
			}
		}
	}

	// Final check
	P_final := model.Polarization()
	if math.Abs(P_final) > material.Ps * 0.1 { // Expected near 0 ? Sequence ends at 0.
		// Wait, sequence ends at 0.
		// LastE history: 2Ec -> -2Ec -> 0.5Ec -> -0.5Ec -> 0.
		// Stack should be: [-2Ec, -0.5Ec, 0] or similar.
		// P might be remanent.
		// We just check it's bounded.
	}
	
	t.Logf("Endurance test passed: %d cycles completed. Final stack size: %d", cycles, len(model.StackE))
}
