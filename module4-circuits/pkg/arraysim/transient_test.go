package arraysim

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
)

func testTransientConfig() ArrayConfig {
	mat := sharedphysics.FeCIMMaterial()
	return ArrayConfig{
		Rows:     1,
		Cols:     1,
		Material: mat,
		Geometry: sharedphysics.GeometryFromMaterial(mat),
	}
}

func TestTransient_CompleteSwitchingAt100ns(t *testing.T) {
	cfg := testTransientConfig()
	ecV := cfg.Material.Ec * cfg.Material.Thickness

	res := TransientSolve(cfg, []PulseStep{{Voltage: ecV, DurationNs: 100}}, 0)
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	if !res[0].Switched {
		t.Fatalf("expected switched=true for 100ns Ec pulse, final P=%.3e", res[0].FinalP)
	}
	if res[0].FinalP < 0.8*cfg.Material.Pr {
		t.Fatalf("expected near-complete switch to +Pr, final P=%.3e, Pr=%.3e", res[0].FinalP, cfg.Material.Pr)
	}
}

func TestTransient_IncompleteAtSubCoercive(t *testing.T) {
	cfg := testTransientConfig()
	// Apply 30% of coercive voltage — sub-coercive field should NOT fully switch.
	// With rho=0.005 Ω·m (Alessandri IEEE EDL 2018), switching at Ec is very
	// fast (~ns), so incomplete switching is best demonstrated with sub-coercive field.
	subEcV := 0.3 * cfg.Material.Ec * cfg.Material.Thickness

	res := TransientSolve(cfg, []PulseStep{{Voltage: subEcV, DurationNs: 10}}, 0.05)
	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}
	// Sub-coercive field produces partial switching — verify it's measurably
	// less than a full Ec pulse. "Incomplete" means FinalP < 0.95*Pr.
	if res[0].Switched && res[0].FinalP >= 0.95*cfg.Material.Pr {
		t.Fatalf("expected incomplete switching at 0.3*Ec, final P=%.3e, Pr=%.3e", res[0].FinalP, cfg.Material.Pr)
	}
}

func TestTransient_EnergyPerCell(t *testing.T) {
	cfg := testTransientConfig()
	ecV := cfg.Material.Ec * cfg.Material.Thickness

	res := TransientSolve(cfg, []PulseStep{{Voltage: ecV, DurationNs: 100}}, 0)
	e := res[0].Energy_fJ
	if e < 10 || e > 100 {
		t.Fatalf("expected energy in [10,100] fJ, got %.3f fJ", e)
	}
}

func TestTransient_ReadDoesNotDisturb(t *testing.T) {
	cfg := testTransientConfig()
	readV := 0.1 * cfg.Material.Ec * cfg.Material.Thickness // sub-coercive read

	baseline := TransientSolve(cfg, []PulseStep{{Voltage: 0, DurationNs: 20}}, 0.05)
	res := TransientSolve(cfg, []PulseStep{{Voltage: readV, DurationNs: 20}}, 0.05)
	delta := math.Abs(res[0].FinalP - baseline[0].FinalP)
	if delta > 0.03*cfg.Material.Pr {
		t.Fatalf("read disturb too large vs no-read relaxation: ΔP=%.3e (allowed %.3e)", delta, 0.03*cfg.Material.Pr)
	}
}
