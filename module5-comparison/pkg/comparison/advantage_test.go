package comparison

import (
	"math"
	"testing"
)

func almostEqual(t *testing.T, got, want, tol float64, context string) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Fatalf("%s: got %.12f, want %.12f (tol %.12f)", context, got, want, tol)
	}
}

// TestInferenceThroughputFromArrayAndClock verifies throughput/ops calculation from
// a representative crossbar array size and clock frequency.
func TestInferenceThroughputFromArrayAndClock(t *testing.T) {
	// 256x256 array, 1 GHz clock, 2 ops/cell/cycle (MAC-like accounting)
	arraySize := 256.0
	clockHz := 1e9
	opsPerCyclePerCell := 2.0

	expectedOpsPerSec := arraySize * arraySize * opsPerCyclePerCell * clockHz
	expectedTOPS := expectedOpsPerSec / 1e12

	arch := CustomArchitecture("FeCIM-like", expectedTOPS, 5.0, 50.0)
	arch.Technology = "FeFET Crossbar" // avoid memory-latency branch
	arch.MemoryBW = 0

	workloadOps := int(1_048_576) // 2^20 operations per inference
	batchSize := 1
	result := arch.RunInference(workloadOps, batchSize)

	// Expected throughput in inferences/s = ops/s / ops-per-inference
	expectedInferenceThroughput := expectedOpsPerSec / float64(workloadOps)

	almostEqual(t, result.Throughput, expectedInferenceThroughput, expectedInferenceThroughput*1e-12, "inference throughput")

	// Back-calculate ops/s from measured throughput.
	measuredOpsPerSec := result.Throughput * float64(workloadOps)
	almostEqual(t, measuredOpsPerSec, expectedOpsPerSec, expectedOpsPerSec*1e-12, "ops/s")
}

// TestEnergyEfficiencyTOPSPerWatt verifies TOPS/W efficiency calculation.
func TestEnergyEfficiencyTOPSPerWatt(t *testing.T) {
	tops := 50.0
	powerW := 5.0
	arch := CustomArchitecture("FeCIM-eff", tops, powerW, 25.0)

	expectedTOPSPerWatt := tops / powerW
	almostEqual(t, arch.TOPSPerWatt, expectedTOPSPerWatt, 1e-12, "TOPS/W")

	if arch.TOPSPerWatt <= 0 {
		t.Fatalf("TOPS/W must be positive, got %.6f", arch.TOPSPerWatt)
	}
}

// TestDataCenterScalingChipsRequired verifies chips-required scaling math.
func TestDataCenterScalingChipsRequired(t *testing.T) {
	workload := Workload{
		Name:        "Synthetic-DC",
		Description: "Deterministic scaling test",
		TotalOps:    1_000_000,
		Layers:      1,
		Parameters:  1_000_000,
	}

	arch := CustomArchitecture("FeCIM-scale", 10.0, 5.0, 20.0) // 10e12 ops/s
	arch.Technology = "FeFET Crossbar"
	arch.MemoryBW = 0

	singleChip := arch.RunInference(workload.TotalOps, 1)
	targetThroughput := singleChip.Throughput * 10.3 // forces ceil -> 11 chips

	metrics := ScaleToDataCenter(arch, targetThroughput, workload)

	expectedChips := int(math.Ceil(targetThroughput / singleChip.Throughput))
	if metrics.ChipsRequired != expectedChips {
		t.Fatalf("chips required mismatch: got %d, want %d", metrics.ChipsRequired, expectedChips)
	}

	if metrics.ChipsRequired < 1 {
		t.Fatalf("chips required should be >= 1, got %d", metrics.ChipsRequired)
	}
}

// TestFeCIMAdvantageRatiosVsGPUBaseline verifies FeCIM-vs-GPU advantage ratios
// are positive and within a reasonable modeled range.
func TestFeCIMAdvantageRatiosVsGPUBaseline(t *testing.T) {
	comparison := CompareArchitectures(ResNet50Workload(), 1, 100000.0)
	adv := CalculateAdvantages(comparison)

	checks := map[string]float64{
		"energy reduction":  adv.VsGPU.EnergyReduction,
		"latency reduction": adv.VsGPU.LatencyReduction,
		"area reduction":    adv.VsGPU.AreaReduction,
		"power reduction":   adv.VsGPU.PowerReduction,
		"cost reduction":    adv.VsGPU.CostReduction,
	}

	for name, ratio := range checks {
		if ratio <= 0 {
			t.Fatalf("%s should be positive, got %.6f", name, ratio)
		}
		if ratio < 1 {
			t.Fatalf("%s should favor FeCIM (>1), got %.6f", name, ratio)
		}
		if ratio > 1e4 {
			t.Fatalf("%s appears unreasonable (>1e4), got %.6f", name, ratio)
		}
	}
}
