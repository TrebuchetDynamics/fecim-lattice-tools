package comparison

import (
	"strings"
	"testing"
)

func TestModule5ComparisonE2EWideWorkloadArchitectureMatrix(t *testing.T) {
	workloads := []Workload{MNISTWorkload(), ResNet50Workload(), BERTBaseWorkload(), GPT2Workload(), LLMWorkload()}
	throughputs := []float64{1, 1000, 10000}
	for _, workload := range workloads {
		for _, target := range throughputs {
			t.Run(workload.Name, func(t *testing.T) {
				comp := CompareArchitectures(workload, 2, target)
				if comp.Workload.Name != workload.Name || len(comp.Architectures) != 3 || len(comp.Results) != 3 || len(comp.DataCenter) != 3 {
					t.Fatalf("comparison shape invalid: %+v", comp)
				}
				adv := CalculateAdvantages(comp)
				if adv.VsCPU.PowerReduction <= 0 || adv.VsGPU.AreaReduction <= 0 {
					t.Fatalf("advantages invalid: %+v", adv)
				}
				for i, arch := range comp.Architectures {
					res := comp.Results[i]
					dc := comp.DataCenter[i]
					if arch.Name == "" || res.Architecture != arch.Name || res.Latency <= 0 || res.Throughput <= 0 || res.Energy <= 0 || dc.ChipsRequired < 1 || dc.InferencesPerSec+1e-9 < target || dc.TCO <= 0 {
						t.Fatalf("architecture result invalid arch=%+v res=%+v dc=%+v target=%g", arch, res, dc, target)
					}
				}
				r := NewRenderer()
				r.UseColor = false
				for _, rendered := range []string{
					r.RenderArchitectureSpecs(comp.Architectures),
					r.RenderInferenceComparison(comp.Results, workload),
					r.RenderDataCenterComparison(comp.DataCenter, target),
					r.RenderAdvantages(adv),
				} {
					if !strings.Contains(rendered, "FeCIM") || len(rendered) < 40 {
						t.Fatalf("rendered comparison missing FeCIM/context: %q", rendered)
					}
				}
			})
		}
	}
}

func TestModule5ComparisonE2ECACTIBaselineAndCustomArchitectureWorkflow(t *testing.T) {
	baselines := DefaultCACTIBaselines()
	if len(baselines) < 4 {
		t.Fatalf("expected multiple CACTI baselines, got %d", len(baselines))
	}
	comparisons := CompareFeCIMvsCACTI(0.25, baselines)
	if len(comparisons) == 0 {
		t.Fatal("CACTI comparisons empty")
	}
	for _, c := range comparisons {
		if c.Baseline.Technology == "" || c.SRAM_pJperMAC <= 0 || c.EnergyRatio_SRAM <= 0 || c.AreaRatio_SRAM <= 0 || c.Source == "" {
			t.Fatalf("CACTI comparison invalid: %+v", c)
		}
	}
	table := FormatComparisonTable(comparisons)
	if !strings.Contains(table, "FeCIM vs CACTI") || !strings.Contains(table, "Ratio SRAM") {
		t.Fatalf("CACTI table invalid: %s", table)
	}
	custom := CustomArchitecture("Tiny ASIC", 12, 3, 6)
	if custom.TOPSPerWatt != 4 || custom.TOPSPerMM2 != 2 {
		t.Fatalf("custom architecture derived metrics invalid: %+v", custom)
	}
	zero := CustomArchitecture("Zero", 12, 0, 0)
	if zero.TOPSPerWatt != 0 || zero.TOPSPerMM2 != 0 {
		t.Fatalf("zero denominator custom architecture invalid: %+v", zero)
	}
}
