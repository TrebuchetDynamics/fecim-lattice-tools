// Package ferroelectric provides physics models for ferroelectric materials.
// This file contains comprehensive spike detection tests for P-E data continuity.
package ferroelectric

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"testing"
)

// =============================================================================
// SPIKE DETECTOR UTILITY
// =============================================================================

// SpikeDetector checks for discontinuities in P-E data.
// Uses the same logic as peplot.go but returns all violations.
type SpikeDetector struct {
	eMax, pMax float64
}

// NewSpikeDetector creates a spike detector with normalization bounds.
func NewSpikeDetector(eMax, pMax float64) *SpikeDetector {
	return &SpikeDetector{eMax: eMax, pMax: pMax}
}

// SpikeInfo contains details about a detected spike.
type SpikeInfo struct {
	Index  int
	DeltaE float64
	DeltaP float64
	NormE  float64
	NormP  float64
	Type   string // "vertical", "horizontal", "large_jump"
}

// DetectSpikes returns detailed information about all spikes.
// Uses same logic as peplot.go:396-398
func (s *SpikeDetector) DetectSpikes(E, P []float64) []SpikeInfo {
	var spikes []SpikeInfo
	for i := 1; i < len(E); i++ {
		eDiff := math.Abs(E[i] - E[i-1])
		pDiff := math.Abs(P[i] - P[i-1])

		normE := eDiff / s.eMax
		normP := pDiff / s.pMax

		var spikeType string
		isSpike := false

		if normE < 0.05 && normP > 0.30 {
			spikeType = "vertical"
			isSpike = true
		} else if normE > 0.30 && normP < 0.05 {
			spikeType = "horizontal"
			isSpike = true
		} else if normP > 0.50 {
			spikeType = "large_jump"
			isSpike = true
		}

		if isSpike {
			spikes = append(spikes, SpikeInfo{
				Index:  i,
				DeltaE: eDiff,
				DeltaP: pDiff,
				NormE:  normE,
				NormP:  normP,
				Type:   spikeType,
			})
		}
	}
	return spikes
}

// HasSpikes returns true if any spikes are detected.
func (s *SpikeDetector) HasSpikes(E, P []float64) bool {
	return len(s.DetectSpikes(E, P)) > 0
}

// AssertNoSpikes fails the test if spikes are detected.
func (s *SpikeDetector) AssertNoSpikes(t *testing.T, E, P []float64) {
	t.Helper()
	spikes := s.DetectSpikes(E, P)
	if len(spikes) > 0 {
		maxShow := 5
		if len(spikes) < maxShow {
			maxShow = len(spikes)
		}
		t.Errorf("Detected %d spikes, first %d:", len(spikes), maxShow)
		for _, spike := range spikes[:maxShow] {
			t.Errorf("  Index %d [%s]: E[%d-1]=%.4e -> E[%d]=%.4e (deltaE=%.4e, normE=%.4f)",
				spike.Index, spike.Type, spike.Index, E[spike.Index-1], spike.Index, E[spike.Index], spike.DeltaE, spike.NormE)
			t.Errorf("              P[%d-1]=%.4e -> P[%d]=%.4e (deltaP=%.4e, normP=%.4f)",
				spike.Index, P[spike.Index-1], spike.Index, P[spike.Index], spike.DeltaP, spike.NormP)
		}
	}
}

// =============================================================================
// CONSECUTIVE POINT TESTS
// =============================================================================

// TestConsecutivePointDeltaE verifies no horizontal spikes in generated data.
// SPIKE: |deltaE| should scale with step size, not jump unexpectedly.
func TestConsecutivePointDeltaE(t *testing.T) {
	materials := []struct {
		name     string
		material *HZOMaterial
	}{
		{"DefaultHZO", DefaultHZO()},
		{"FeCIMMaterial", FeCIMMaterial()},
		{"LiteratureSuperlattice", LiteratureSuperlattice()},
	}

	for _, m := range materials {
		t.Run(m.name, func(t *testing.T) {
			model := NewMayergoyzPreisach(m.material, 50)
			Emax := m.material.Ec * 2.0
			points := 500

			E, _ := model.GetHysteresisLoop(Emax, points)

			// Expected step size: 4*Emax / (4*points) = Emax/points
			expectedStep := Emax / float64(points)
			maxDeltaE := expectedStep * 2.5 // Allow 2.5x tolerance

			for i := 1; i < len(E); i++ {
				deltaE := math.Abs(E[i] - E[i-1])
				if deltaE > maxDeltaE {
					t.Errorf("Horizontal spike at index %d: deltaE=%.4e > max %.4e (%.1fx expected)",
						i, deltaE, maxDeltaE, deltaE/expectedStep)
				}
			}
		})
	}
}

// TestConsecutivePointDeltaP verifies no vertical spikes in generated data.
// SPIKE: Large deltaP with small deltaE indicates data corruption.
func TestConsecutivePointDeltaP(t *testing.T) {
	materials := []struct {
		name     string
		material *HZOMaterial
	}{
		{"DefaultHZO", DefaultHZO()},
		{"FeCIMMaterial", FeCIMMaterial()},
	}

	for _, m := range materials {
		t.Run(m.name, func(t *testing.T) {
			model := NewMayergoyzPreisach(m.material, 50)
			Emax := m.material.Ec * 2.0
			points := 500

			E, P := model.GetHysteresisLoop(Emax, points)
			Ps := m.material.Ps

			// Max susceptibility: dP/dE at steepest point should not exceed 2*Ps/Ec
			maxSusceptibility := 2 * Ps / m.material.Ec

			for i := 1; i < len(E); i++ {
				deltaE := math.Abs(E[i] - E[i-1])
				deltaP := math.Abs(P[i] - P[i-1])

				// Allow for deltaE~0 with small deltaP (stationary point)
				if deltaE < 1e-12 {
					if deltaP > Ps*0.01 { // More than 1% jump at stationary point
						t.Errorf("Vertical spike at index %d: deltaE=%.4e~0 but deltaP=%.4e (%.1f%% of Ps)",
							i, deltaE, deltaP, deltaP/Ps*100)
					}
					continue
				}

				susceptibility := deltaP / deltaE
				if susceptibility > maxSusceptibility*3 { // 3x margin
					t.Errorf("Excessive susceptibility at index %d: dP/dE=%.4e > max %.4e",
						i, susceptibility, maxSusceptibility*3)
				}
			}
		})
	}
}

// =============================================================================
// SPIKE PATTERN TESTS
// =============================================================================

// TestVerticalSpikeDetection reproduces the specific vertical spike pattern.
// CONDITIONS: normE < 0.05 && normP > 0.30 (from peplot.go:396)
func TestVerticalSpikeDetection(t *testing.T) {
	material := DefaultHZO()

	gridSizes := []int{30, 50, 80, 100}
	pointCounts := []int{100, 500, 1000}

	for _, gridSize := range gridSizes {
		for _, points := range pointCounts {
			t.Run(fmt.Sprintf("grid%d_points%d", gridSize, points), func(t *testing.T) {
				model := NewMayergoyzPreisach(material, gridSize)
				Emax := material.Ec * 2.0

				E, P := model.GetHysteresisLoop(Emax, points)

				detector := NewSpikeDetector(Emax, material.Ps)
				spikes := detector.DetectSpikes(E, P)

				verticalCount := 0
				for _, s := range spikes {
					if s.Type == "vertical" {
						verticalCount++
					}
				}

				if verticalCount > 0 {
					t.Errorf("Found %d vertical spikes (normE<0.05 && normP>0.30)", verticalCount)
				}
			})
		}
	}
}

// TestHorizontalSpikeDetection tests for horizontal spikes.
// CONDITIONS: normE > 0.30 && normP < 0.05 (from peplot.go:397)
func TestHorizontalSpikeDetection(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	Emax := material.Ec * 2.0

	E, P := model.GetHysteresisLoop(Emax, 500)

	detector := NewSpikeDetector(Emax, material.Ps)
	spikes := detector.DetectSpikes(E, P)

	horizontalCount := 0
	for _, s := range spikes {
		if s.Type == "horizontal" {
			horizontalCount++
		}
	}

	if horizontalCount > 0 {
		t.Errorf("Found %d horizontal spikes (normE>0.30 && normP<0.05)", horizontalCount)
	}
}

// TestLargeJumpDetection tests for any large P jump.
// CONDITIONS: normP > 0.50 (from peplot.go:398)
func TestLargeJumpDetection(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	Emax := material.Ec * 2.0

	E, P := model.GetHysteresisLoop(Emax, 500)

	detector := NewSpikeDetector(Emax, material.Ps)
	spikes := detector.DetectSpikes(E, P)

	largeJumpCount := 0
	for _, s := range spikes {
		if s.Type == "large_jump" {
			largeJumpCount++
		}
	}

	if largeJumpCount > 0 {
		t.Errorf("Found %d large jumps (normP>0.50)", largeJumpCount)
	}
}

// =============================================================================
// RAPID FIELD REVERSAL TESTS
// =============================================================================

// TestRapidFieldReversalSpikes verifies stability under many direction changes.
// TRIGGER: Frequent direction changes stress the history tracking.
func TestRapidFieldReversalSpikes(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0
	reversals := 100

	var E, P []float64
	currentE := 0.0

	for i := 0; i < reversals; i++ {
		sign := 1.0
		if i%2 == 0 {
			sign = -1.0
		}
		amplitude := (rand.Float64()*0.5 + 0.5) * Emax

		targetE := sign * amplitude
		steps := 50
		for s := 0; s <= steps; s++ {
			newE := currentE + (targetE-currentE)*float64(s)/float64(steps)
			p := model.Update(newE)

			E = append(E, newE)
			P = append(P, p)

			currentE = newE
		}
	}

	for i, p := range P {
		if math.IsNaN(p) {
			t.Fatalf("NaN at index %d after %d reversals", i, i/51)
		}
		if math.IsInf(p, 0) {
			t.Fatalf("Inf at index %d after %d reversals", i, i/51)
		}
		if math.Abs(p) > material.Ps*1.001 {
			t.Errorf("P out of bounds at index %d: |P|=%.4e > Ps=%.4e", i, math.Abs(p), material.Ps)
		}
	}
}

// TestMicroReversals tests very small amplitude reversals.
// These can stress the turning point memory especially near Ec.
func TestMicroReversals(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Ec := material.Ec
	microAmplitude := Ec * 0.01 // 1% of Ec

	var E, P []float64

	// Approach Ec with micro-oscillations
	for level := 0.5; level <= 1.5; level += 0.1 {
		baseE := Ec * level
		for i := 0; i < 50; i++ {
			oscillation := microAmplitude * math.Sin(float64(i)*0.5)
			e := baseE + oscillation
			p := model.Update(e)

			E = append(E, e)
			P = append(P, p)
		}
	}

	// Check for anomalies
	for i, p := range P {
		if math.IsNaN(p) {
			t.Errorf("NaN at index %d during micro-reversals", i)
		}
		if math.Abs(p) > material.Ps*1.001 { // Small tolerance for floating point
			t.Errorf("P out of bounds at index %d: |P|=%.4e > Ps=%.4e", i, math.Abs(p), material.Ps)
		}
	}
}

// =============================================================================
// SIMPLE MODEL SPIKE TESTS
// =============================================================================

// TestSimplePreisachNoSpikes tests the simpler PreisachModel.
func TestSimplePreisachNoSpikes(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)
	Emax := material.Ec * 2.0

	E, P := model.GetHysteresisLoop(Emax, 500)

	detector := NewSpikeDetector(Emax, material.Ps)
	detector.AssertNoSpikes(t, E, P)
}

// TestSimplePreisachRapidReversals tests rapid reversals with simple model.
func TestSimplePreisachRapidReversals(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	Emax := material.Ec * 2.0
	reversals := 500

	var E, P []float64
	currentE := 0.0

	for i := 0; i < reversals; i++ {
		sign := 1.0
		if i%2 == 0 {
			sign = -1.0
		}
		targetE := sign * Emax * 0.8

		steps := 5
		for s := 0; s <= steps; s++ {
			newE := currentE + (targetE-currentE)*float64(s)/float64(steps)
			p := model.Update(newE)
			E = append(E, newE)
			P = append(P, p)
			currentE = newE
		}
	}

	// Check bounds
	for i, p := range P {
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Errorf("Invalid P at index %d: %v", i, p)
		}
		if math.Abs(p) > material.Ps*1.001 {
			t.Errorf("P out of bounds at index %d: %.4e > Ps", i, p)
		}
	}
}

// =============================================================================
// BOUNDARY CONDITION TESTS
// =============================================================================

// TestNearSaturationSpikes verifies smooth approach to +/-Ps.
// SPIKE CAUSE: Hard clamping can create discontinuity.
func TestNearSaturationSpikes(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 3.0
	E, P := model.GetHysteresisLoop(Emax, 1000)

	Ps := material.Ps

	nearSaturationCount := 0
	saturationSpikes := 0

	for i := 1; i < len(P); i++ {
		if math.Abs(P[i]) > 0.98*Ps || math.Abs(P[i-1]) > 0.98*Ps {
			nearSaturationCount++
			deltaP := math.Abs(P[i] - P[i-1])
			deltaE := math.Abs(E[i] - E[i-1])

			if deltaE > 1e-12 && deltaP/deltaE > 10*Ps/material.Ec {
				saturationSpikes++
			}
		}
	}

	if nearSaturationCount == 0 {
		t.Error("Test did not reach saturation region - increase Emax")
	}

	if saturationSpikes > 0 {
		t.Errorf("Found %d spikes near saturation out of %d points",
			saturationSpikes, nearSaturationCount)
	}
}

// TestZeroCrossingSpikes verifies smooth E=0 crossing.
func TestZeroCrossingSpikes(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)

	Emax := material.Ec * 2.0
	E, P := model.GetHysteresisLoop(Emax, 500)

	// Find zero crossings
	zeroCrossingSpikes := 0
	for i := 1; i < len(E); i++ {
		// Check if this segment crosses E=0
		if (E[i-1] < 0 && E[i] > 0) || (E[i-1] > 0 && E[i] < 0) || math.Abs(E[i]) < material.Ec*0.1 {
			deltaP := math.Abs(P[i] - P[i-1])
			// Should not have >10% P jump at zero crossing
			if deltaP > material.Ps*0.10 {
				zeroCrossingSpikes++
			}
		}
	}

	if zeroCrossingSpikes > 0 {
		t.Errorf("Found %d large P jumps at or near E=0 crossing", zeroCrossingSpikes)
	}
}

// =============================================================================
// STATISTICAL PROPERTY TESTS
// =============================================================================

// TestPolarizationAlwaysBounded verifies P is always within [-Ps, +Ps].
func TestPolarizationAlwaysBounded(t *testing.T) {
	materials := AllMaterials()

	for _, material := range materials {
		t.Run(material.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(material, 50)
			Emax := material.Ec * 3.0

			// Generate extensive data
			E, P := model.GetHysteresisLoop(Emax, 1000)

			Ps := material.Ps
			tolerance := Ps * 0.001 // 0.1% tolerance for floating point

			for i, p := range P {
				if p > Ps+tolerance {
					t.Errorf("P[%d]=%.6e exceeds +Ps=%.6e (at E=%.6e)", i, p, Ps, E[i])
				}
				if p < -Ps-tolerance {
					t.Errorf("P[%d]=%.6e below -Ps=%.6e (at E=%.6e)", i, p, -Ps, E[i])
				}
			}
		})
	}
}

// TestNoNaNOrInfInLoop verifies no invalid floating point values.
func TestNoNaNOrInfInLoop(t *testing.T) {
	materials := AllMaterials()

	for _, material := range materials {
		t.Run(material.Name, func(t *testing.T) {
			model := NewMayergoyzPreisach(material, 50)
			Emax := material.Ec * 3.0

			E, P := model.GetHysteresisLoop(Emax, 1000)

			for i := range E {
				if math.IsNaN(E[i]) {
					t.Errorf("E[%d] is NaN", i)
				}
				if math.IsInf(E[i], 0) {
					t.Errorf("E[%d] is Inf", i)
				}
				if math.IsNaN(P[i]) {
					t.Errorf("P[%d] is NaN", i)
				}
				if math.IsInf(P[i], 0) {
					t.Errorf("P[%d] is Inf", i)
				}
			}
		})
	}
}

// TestDeterministicOutput verifies same input produces same output.
func TestDeterministicOutput(t *testing.T) {
	material := DefaultHZO()

	// Generate random but reproducible field sequence
	rng := rand.New(rand.NewSource(42))
	sequence := make([]float64, 1000)
	for i := range sequence {
		sequence[i] = (rng.Float64()*2 - 1) * material.Ec * 2
	}

	// Run twice with fresh models
	model1 := NewMayergoyzPreisach(material, 50)
	model2 := NewMayergoyzPreisach(material, 50)

	for i, e := range sequence {
		p1 := model1.Update(e)
		p2 := model2.Update(e)

		if p1 != p2 {
			t.Errorf("Non-deterministic at step %d: P1=%.10e, P2=%.10e (diff=%.4e)",
				i, p1, p2, math.Abs(p1-p2))
		}
	}
}

// =============================================================================
// CONCURRENCY STRESS TESTS
// =============================================================================

// TestConcurrentModelUpdate verifies Update() does not panic under concurrent calls.
func TestConcurrentModelUpdate(t *testing.T) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 30)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// 10 goroutines, 100 updates each
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func(seed int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("panic in goroutine %d: %v", seed, r)
				}
			}()

			for i := 0; i < 100; i++ {
				E := float64(seed*1000+i) * material.Ec / 5000
				P := model.Update(E)

				// Check for NaN/Inf (data corruption indicators)
				if math.IsNaN(P) {
					errors <- fmt.Errorf("NaN from goroutine %d, iteration %d", seed, i)
				}
				if math.IsInf(P, 0) {
					errors <- fmt.Errorf("Inf from goroutine %d, iteration %d", seed, i)
				}
			}
		}(g)
	}

	wg.Wait()
	close(errors)

	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
		if errorCount > 10 {
			t.Fatal("Too many concurrent errors, stopping")
		}
	}
}

// =============================================================================
// BENCHMARKS
// =============================================================================

// BenchmarkSpikeDetection measures spike detection performance.
func BenchmarkSpikeDetection(b *testing.B) {
	material := DefaultHZO()
	model := NewMayergoyzPreisach(material, 50)
	Emax := material.Ec * 2.0
	E, P := model.GetHysteresisLoop(Emax, 1000)

	detector := NewSpikeDetector(Emax, material.Ps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectSpikes(E, P)
	}
}

// BenchmarkRapidReversals measures reversal handling.
func BenchmarkRapidReversals(b *testing.B) {
	material := DefaultHZO()
	Emax := material.Ec * 2.0

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model := NewMayergoyzPreisach(material, 30)
		for r := 0; r < 100; r++ {
			sign := 1.0
			if r%2 == 0 {
				sign = -1.0
			}
			for s := 0; s < 10; s++ {
				model.Update(sign * Emax * float64(s) / 10)
			}
		}
	}
}
