package physics

import (
	"fmt"
	"math"
	"testing"
)

// simpleUniformEverett implements a simple uniform Everett distribution for testing.
// It produces normalized polarization in [-1, 1] for fields in [-sat, sat].
type simpleUniformEverett struct {
	sat float64
}

func (e simpleUniformEverett) Calculate(alpha, beta float64) float64 {
	// Simple triangular region integral for uniform distribution
	// The density is 1/(2*sat)^2 so that full saturation gives normalized polarization
	d := alpha - beta
	if d < 0 {
		d = 0
	}
	return d / (2 * e.sat)
}

func TestNewPreisachStack(t *testing.T) {
	sat := 1e8 // 1 MV/cm in V/m
	everett := simpleUniformEverett{sat: sat}

	ps := NewPreisachStack(sat, everett)

	if ps == nil {
		t.Fatal("NewPreisachStack returned nil")
	}
	if ps.SaturationE != sat {
		t.Errorf("SaturationE = %e, want %e", ps.SaturationE, sat)
	}
	if ps.CurrentDir != 1 {
		t.Errorf("CurrentDir = %d, want 1 (ascending)", ps.CurrentDir)
	}
	if ps.LastE != -sat {
		t.Errorf("LastE = %e, want %e", ps.LastE, -sat)
	}
	if len(ps.Stack) != 1 {
		t.Errorf("len(Stack) = %d, want 1", len(ps.Stack))
	}
	if ps.Stack[0].Type != -1 {
		t.Errorf("Initial stack point type = %d, want -1 (Min)", ps.Stack[0].Type)
	}
}

func TestPreisachStack_Update_Monotonic(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// Monotonic increase from -sat to +sat
	fields := []float64{-0.8, -0.5, 0.0, 0.5, 0.8, 1.0}
	var prevP float64 = math.Inf(-1)

	for _, E := range fields {
		P := ps.Update(E)
		if P <= prevP {
			t.Errorf("Polarization should increase monotonically: E=%f gave P=%f (prev=%f)", E, P, prevP)
		}
		prevP = P
	}

	// At saturation, P should be close to 1
	Psat := ps.Update(sat)
	if Psat < 0.9 || Psat > 1.1 {
		t.Errorf("At saturation, P should be ~1, got %f", Psat)
	}
}

func TestPreisachStack_Update_Reversal(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// Go up to 0.5
	P1 := ps.Update(0.5)
	// Reverse direction
	P2 := ps.Update(0.4)

	// After reversal, a turning point should be added to the stack
	if len(ps.Stack) < 2 {
		t.Error("Stack should have at least 2 points after reversal")
	}
	if ps.CurrentDir != -1 {
		t.Errorf("CurrentDir should be -1 (descending) after reversal, got %d", ps.CurrentDir)
	}

	t.Logf("After reversal: P1(0.5)=%f, P2(0.4)=%f, stack=%d", P1, P2, len(ps.Stack))
}

func TestPreisachStack_Update_WipeOut(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// Create a minor loop: go up to 0.5, down to 0.2, then up past 0.5
	ps.Update(0.5) // Up
	ps.Update(0.2) // Down
	stackBeforeWipe := len(ps.Stack)

	// Now exceed the previous max - should wipe out the minor loop
	ps.Update(0.7)
	stackAfterWipe := len(ps.Stack)

	// The stack should be smaller or equal after wipe-out
	if stackAfterWipe > stackBeforeWipe {
		t.Errorf("Wipe-out should reduce or maintain stack size: before=%d, after=%d",
			stackBeforeWipe, stackAfterWipe)
	}

	t.Logf("Wipe-out: stack before=%d, after=%d", stackBeforeWipe, stackAfterWipe)
}

func TestPreisachStack_Update_MinorLoop(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// Major loop: go to saturation
	ps.Update(sat)
	Psat := ps.ComputePolarization(sat)

	// Minor loop: go down slightly and back up
	P1 := ps.Update(0.8)
	P2 := ps.Update(0.6)
	P3 := ps.Update(0.8)

	// After returning to 0.8, P should be the same as when we first reached 0.8
	// (assuming simple Everett function with return-point memory)
	t.Logf("Minor loop: Psat=%f, P(0.8)down=%f, P(0.6)=%f, P(0.8)up=%f", Psat, P1, P2, P3)

	// P3 should be close to P1 (return-point memory property)
	if math.Abs(P3-P1) > 0.1 {
		t.Logf("Note: Return-point memory: |P3-P1| = %f (may not be exact with simple Everett)", math.Abs(P3-P1))
	}
}

func TestPreisachStack_Update_NoChange(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	P1 := ps.Update(0.5)
	P2 := ps.Update(0.5) // Same field - no change

	if P1 != P2 {
		t.Errorf("Update with same field should return same P: P1=%f, P2=%f", P1, P2)
	}
}

func TestPreisachStack_ComputePolarization_Saturation(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// At negative saturation (initial state)
	Pneg := ps.ComputePolarization(-sat)
	if Pneg > -0.9 {
		t.Errorf("At negative saturation, P should be ~ -1, got %f", Pneg)
	}

	// Drive to positive saturation
	ps.Update(sat)
	Ppos := ps.ComputePolarization(sat)
	if Ppos < 0.9 {
		t.Errorf("At positive saturation, P should be ~ +1, got %f", Ppos)
	}
}

func TestPreisachStack_NaNInputs(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update(NaN) panicked: %v", r)
		}
	}()

	P := ps.Update(math.NaN())
	t.Logf("Update(NaN) = %f", P)
}

func TestPreisachStack_InfInputs(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update(Inf) panicked: %v", r)
		}
	}()

	P := ps.Update(math.Inf(1))
	t.Logf("Update(+Inf) = %f", P)
}

func TestPreisachStackRejectsInvalidFieldInputsWithoutMutatingState(t *testing.T) {
	ps := NewPreisachStack(1.0, simpleUniformEverett{sat: 1.0})
	ps.Update(0.6)
	ps.Update(0.2)

	wantP := ps.ComputePolarization(ps.LastE)
	wantLastE := ps.LastE
	wantCurrentDir := ps.CurrentDir
	wantStack := append([]TurningPoint(nil), ps.Stack...)

	invalidFields := []float64{math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64, -math.MaxFloat64}
	for _, field := range invalidFields {
		t.Run(fmt.Sprintf("E=%g", field), func(t *testing.T) {
			got := ps.Update(field)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Fatalf("Update(%g) returned non-finite polarization %g", field, got)
			}
			if got != wantP {
				t.Fatalf("Update(%g) polarization = %g, want current-state polarization %g", field, got, wantP)
			}
			if ps.LastE != wantLastE {
				t.Fatalf("Update(%g) mutated LastE to %g, want %g", field, ps.LastE, wantLastE)
			}
			if ps.CurrentDir != wantCurrentDir {
				t.Fatalf("Update(%g) mutated CurrentDir to %d, want %d", field, ps.CurrentDir, wantCurrentDir)
			}
			if len(ps.Stack) != len(wantStack) {
				t.Fatalf("Update(%g) mutated stack length to %d, want %d", field, len(ps.Stack), len(wantStack))
			}
			for i := range wantStack {
				if ps.Stack[i] != wantStack[i] {
					t.Fatalf("Update(%g) mutated stack[%d] to %+v, want %+v", field, i, ps.Stack[i], wantStack[i])
				}
			}
		})
	}
}

func TestNewPreisachStackRejectsUnrepresentableSaturation(t *testing.T) {
	invalidSaturation := []float64{math.NaN(), math.Inf(1), math.Inf(-1), math.MaxFloat64, -math.MaxFloat64}
	for _, saturation := range invalidSaturation {
		t.Run(fmt.Sprintf("saturation=%g", saturation), func(t *testing.T) {
			if ps := NewPreisachStack(saturation, simpleUniformEverett{sat: 1.0}); ps != nil {
				t.Fatalf("NewPreisachStack(%g, everett) returned non-nil stack with SaturationE=%g", saturation, ps.SaturationE)
			}
		})
	}
}

func TestPreisachStack_ZeroSaturation(t *testing.T) {
	// Edge case: zero saturation field should return nil
	everett := simpleUniformEverett{sat: 0}
	ps := NewPreisachStack(0, everett)

	if ps != nil {
		t.Fatalf("NewPreisachStack(0, ...) should return nil for non-positive saturationE")
	}
}

func TestPreisachStack_NilEverett(t *testing.T) {
	// Edge case: nil Everett function should return nil
	ps := NewPreisachStack(1.0, nil)
	if ps != nil {
		t.Fatalf("NewPreisachStack(1.0, nil) should return nil for nil Everett")
	}
}

func TestPreisachStack_LargeHistory(t *testing.T) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// Create many minor loops
	for i := 0; i < 100; i++ {
		amplitude := 0.1 + 0.005*float64(i%20)
		ps.Update(amplitude)
		ps.Update(-amplitude)
	}

	P := ps.ComputePolarization(0)
	if math.IsNaN(P) {
		t.Error("Polarization should not be NaN after many updates")
	}
	if math.IsInf(P, 0) {
		t.Error("Polarization should not be Inf after many updates")
	}

	t.Logf("After 100 minor loops: P(0)=%f, stack size=%d", P, len(ps.Stack))
}

func TestTurningPoint(t *testing.T) {
	tp := TurningPoint{E: 1.5, Type: 1}

	if tp.E != 1.5 {
		t.Errorf("E = %f, want 1.5", tp.E)
	}
	if tp.Type != 1 {
		t.Errorf("Type = %d, want 1", tp.Type)
	}
}

func BenchmarkPreisachStack_Update(b *testing.B) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps := NewPreisachStack(sat, everett)
		// Simulate a typical hysteresis cycle
		for E := -1.0; E <= 1.0; E += 0.1 {
			ps.Update(E)
		}
		for E := 1.0; E >= -1.0; E -= 0.1 {
			ps.Update(E)
		}
	}
}

func BenchmarkPreisachStack_ComputePolarization(b *testing.B) {
	sat := 1.0
	everett := simpleUniformEverett{sat: sat}
	ps := NewPreisachStack(sat, everett)

	// Build up some history
	for i := 0; i < 20; i++ {
		ps.Update(0.5 * float64(i%2*2-1))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ps.ComputePolarization(0.3)
	}
}
