//go:build legacy_fyne

package comparison

import "testing"

func TestComputeMetrics_DefaultAndFormats(t *testing.T) {
	cpu, gpu, fefet := ComputeMetrics(0)
	if cpu.Label != "CPU" || gpu.Label != "GPU" || fefet.Label != "FeFET" {
		t.Fatalf("unexpected labels: %#v %#v %#v", cpu, gpu, fefet)
	}
	if MetricLatency(cpu.LatencyNS) != "500 ns" {
		t.Fatalf("latency format got %q", MetricLatency(cpu.LatencyNS))
	}
	if MetricEnergy(fefet.EnergyPJ) != "2.9 pJ" {
		t.Fatalf("energy format got %q", MetricEnergy(fefet.EnergyPJ))
	}
	if MetricGOPS(fefet.GOPS) != "0.002 GOPS" {
		t.Fatalf("gops format got %q", MetricGOPS(fefet.GOPS))
	}
}

func TestBuildDesignSpaceSweep_CountAndFields(t *testing.T) {
	points := BuildDesignSpaceSweep([]int{8, 16}, []int{4, 6}, []string{"FeFET", "RRAM"})
	if got, want := len(points), 8; got != want {
		t.Fatalf("len(points) = %d, want %d", got, want)
	}
	for _, p := range points {
		if p.ArraySize <= 0 || p.ADCBits <= 0 {
			t.Fatalf("invalid sweep point: %+v", p)
		}
		if p.LatencyNS <= 0 || p.EnergyPJ <= 0 {
			t.Fatalf("non-positive metrics: %+v", p)
		}
	}
}

func TestRunProcessVariationMonteCarlo_Basic(t *testing.T) {
	stats := RunProcessVariationMonteCarlo(1e-6, 0.1, 1000, 42)
	if stats.Mean <= 0 {
		t.Fatalf("mean should be > 0, got %g", stats.Mean)
	}
	if stats.StdDev <= 0 {
		t.Fatalf("stddev should be > 0, got %g", stats.StdDev)
	}
	if stats.Min < 0 {
		t.Fatalf("min should be >= 0, got %g", stats.Min)
	}
	if stats.Max < stats.Min {
		t.Fatalf("max should be >= min, got min=%g max=%g", stats.Min, stats.Max)
	}
}
