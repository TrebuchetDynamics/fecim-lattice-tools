package ferroelectric

import (
	"math"
	"strings"
	"testing"
)

func TestModule1FerroelectricE2EWidePhysicsRenderMatrix(t *testing.T) {
	materials := []*HZOMaterial{
		DefaultHZO(),
		FeCIMMaterial(),
		LiteratureSuperlattice(),
		CryogenicHZO(),
		HZOStandard32(),
		HZOFJT140(),
		AlScN(),
	}
	renderer := NewPERenderer()
	renderer.Color = false

	for _, material := range materials {
		t.Run(material.Name, func(t *testing.T) {
			model := NewPreisachModel(material)
			if model == nil {
				t.Fatal("NewPreisachModel returned nil")
			}
			fields, pol := model.GetHysteresisLoop(material.Ec*2, 100)
			if len(fields) != 401 || len(pol) != 401 {
				t.Fatalf("loop length = %d/%d, want 401/401", len(fields), len(pol))
			}
			minP, maxP, zeroCrossings := module1E2ELoopStats(fields, pol)
			if minP >= 0 || maxP <= 0 || zeroCrossings < 2 {
				t.Fatalf("loop stats min=%.3e max=%.3e zeroCrossings=%d, want bipolar hysteresis", minP, maxP, zeroCrossings)
			}
			if math.Abs(maxP) > material.Ps*1.5 || math.Abs(minP) > material.Ps*1.5 {
				t.Fatalf("loop polarization exceeds expected material envelope: min=%.3e max=%.3e Ps=%.3e", minP, maxP, material.Ps)
			}

			levels := material.GetNumLevels()
			if levels < 2 || levels > 64 {
				levels = 30
			}
			states := model.DiscreteStates(levels)
			if len(states) != levels {
				t.Fatalf("DiscreteStates(%d) length = %d", levels, len(states))
			}
			if states[0].Level != 1 || states[len(states)-1].Level != levels {
				t.Fatalf("state endpoints levels = %d/%d, want 1/%d", states[0].Level, states[len(states)-1].Level, levels)
			}
			if math.Abs(states[0].NormalizedP+1) > 1e-12 || math.Abs(states[len(states)-1].NormalizedP-1) > 1e-12 {
				t.Fatalf("state normalized endpoints = %.6f/%.6f, want -1/+1", states[0].NormalizedP, states[len(states)-1].NormalizedP)
			}
			for i := 1; i < len(states); i++ {
				if states[i].Polarization <= states[i-1].Polarization {
					t.Fatalf("state %d polarization %.6e <= previous %.6e", i, states[i].Polarization, states[i-1].Polarization)
				}
			}

			pe := renderer.RenderPELoop(fields, pol, material)
			discrete := renderer.RenderDiscreteStates(states)
			temp := renderer.RenderTemperatureDependence(material)
			comparison := renderer.RenderMaterialComparison()
			for _, check := range []struct {
				name string
				text string
				want []string
			}{
				{name: "pe", text: pe, want: []string{"P-E Hysteresis Loop", "Legend", "Ec", "Pr"}},
				{name: "discrete", text: discrete, want: []string{"Discrete Analog Levels", "simulation baseline", "Total modeled levels"}},
				{name: "temperature", text: temp, want: []string{"Temperature Dependence", "Curie Temperature", "T (K)"}},
				{name: "comparison", text: comparison, want: []string{"HZO Material Comparison", "FeCIM HZO", "AlScN"}},
			} {
				if strings.TrimSpace(check.text) == "" {
					t.Fatalf("%s render output empty", check.name)
				}
				for _, want := range check.want {
					if !strings.Contains(check.text, want) {
						t.Fatalf("%s render missing %q\n%s", check.name, want, check.text)
					}
				}
			}
		})
	}
}

func TestModule1FerroelectricE2EMinorLoopResetAndInvalidDiscreteCounts(t *testing.T) {
	material := FeCIMMaterial()
	model := NewPreisachModel(material)
	if model == nil {
		t.Fatal("NewPreisachModel returned nil")
	}
	p0 := model.Update(-material.Ec * 2)
	p1 := model.Update(material.Ec * 0.5)
	p2 := model.Update(-material.Ec * 0.25)
	if !isFiniteModule1FerroelectricE2E(p0) || !isFiniteModule1FerroelectricE2E(p1) || !isFiniteModule1FerroelectricE2E(p2) {
		t.Fatalf("minor-loop updates produced non-finite values: %.3e %.3e %.3e", p0, p1, p2)
	}
	if p1 == p2 {
		t.Fatalf("minor-loop polarization did not respond to reversal: p1=%.3e p2=%.3e", p1, p2)
	}
	model.Reset()
	fieldsAfterReset, polAfterReset := model.GetHysteresisLoop(material.Ec*2, 25)
	fresh := NewPreisachModel(material)
	freshFields, freshPol := fresh.GetHysteresisLoop(material.Ec*2, 25)
	if len(fieldsAfterReset) != len(freshFields) || len(polAfterReset) != len(freshPol) {
		t.Fatalf("reset loop lengths = %d/%d fresh %d/%d", len(fieldsAfterReset), len(polAfterReset), len(freshFields), len(freshPol))
	}
	for i := range polAfterReset {
		if math.Abs(polAfterReset[i]-freshPol[i]) > 1e-12 || fieldsAfterReset[i] != freshFields[i] {
			t.Fatalf("reset loop diverged at %d: field %.3e/%.3e pol %.3e/%.3e", i, fieldsAfterReset[i], freshFields[i], polAfterReset[i], freshPol[i])
		}
	}

	for _, invalid := range []int{0, 1, maxDiscreteStateCount + 1, -4} {
		if got := model.DiscreteStates(invalid); got != nil {
			t.Fatalf("DiscreteStates(%d) = %v, want nil", invalid, got)
		}
	}
}

func module1E2ELoopStats(fields, pol []float64) (minP, maxP float64, zeroCrossings int) {
	minP = math.Inf(1)
	maxP = math.Inf(-1)
	for i, p := range pol {
		if p < minP {
			minP = p
		}
		if p > maxP {
			maxP = p
		}
		if i > 0 && ((pol[i-1] < 0 && p >= 0) || (pol[i-1] > 0 && p <= 0)) {
			zeroCrossings++
		}
		_ = fields
	}
	return minP, maxP, zeroCrossings
}

func isFiniteModule1FerroelectricE2E(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}
