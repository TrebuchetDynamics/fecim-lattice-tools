package ferroelectric

import (
	"strings"
	"testing"
)

func TestPERenderer_RenderPELoop_IncludesPhysicsMarkers(t *testing.T) {
	r := &PERenderer{Width: 40, Height: 15}
	m := DefaultHZO()
	E := []float64{-m.Ec, -0.5 * m.Ec, 0, 0.5 * m.Ec, m.Ec}
	P := []float64{-m.Pr, -0.5 * m.Pr, 0, 0.5 * m.Pr, m.Pr}

	out := r.RenderPELoop(E, P, m)
	for _, token := range []string{"P-E Hysteresis Loop", "Legend", "Ec", "Pr"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected token %q in render output", token)
		}
	}
}

func TestPERenderer_RenderDomainStates_ReportsSwitchedFraction(t *testing.T) {
	r := NewPERenderer()
	alphas := []float64{1.0, 2.0, 3.0, 4.0}
	betas := []float64{-4.0, -3.0, -2.0, -1.0}
	states := []int{+1, -1, +1, -1}

	out := r.RenderDomainStates(alphas, betas, states)
	for _, token := range []string{"Preisach Plane", "Switched fraction", "█ Up", "░ Down"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected token %q in domain state render", token)
		}
	}
}

func TestPERenderer_RenderSwitchingDynamics_IncludesTauAndFinalP(t *testing.T) {
	r := NewPERenderer()
	m := DefaultHZO()
	times := []float64{0, 1e-9, 2e-9, 3e-9, 4e-9}
	pols := []float64{0, 0.2 * m.Ps, 0.5 * m.Ps, 0.8 * m.Ps, m.Ps}
	switched := []int{0, 20, 50, 80, 100}

	out := r.RenderSwitchingDynamics(times, pols, switched, m)
	for _, token := range []string{"KAI Model", "Switching time (τ)", "Final polarization", "Domains switched"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected token %q in switching dynamics render", token)
		}
	}
}

func TestPERenderer_RenderTemperatureDependence_IncludesCurieSummary(t *testing.T) {
	r := NewPERenderer()
	m := DefaultHZO()

	out := r.RenderTemperatureDependence(m)
	for _, token := range []string{"Temperature Dependence", "Curie Temperature", "Ec (MV/cm)", "Pr (µC/cm²)"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected token %q in temperature render", token)
		}
	}
}

func TestPERenderer_RenderMaterialComparison_ListsKnownMaterial(t *testing.T) {
	r := NewPERenderer()
	out := r.RenderMaterialComparison()

	for _, token := range []string{"HZO Material Comparison", "Material", "Endurance", "FeCIM"} {
		if !strings.Contains(out, token) {
			t.Fatalf("expected token %q in material comparison render", token)
		}
	}
}
