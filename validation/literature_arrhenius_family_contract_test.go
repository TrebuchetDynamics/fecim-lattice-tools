package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type arrheniusSwitchPoint struct {
	EOverEc  float64 `json:"E_over_Ec"`
	TauSim   float64 `json:"tau_simulated_s"`
	TauMerz  float64 `json:"tau_merz_law_s"`
	ErrPct   float64 `json:"error_pct"`
}

type arrheniusSwitchArtifact struct {
	Material string              `json:"material"`
	Tau0     float64             `json:"tau0_s"`
	Ec       float64             `json:"ec_V_m"`
	Results  []arrheniusSwitchPoint `json:"results"`
}

type arrheniusMaterialPoint struct {
	Name    string  `json:"name"`
	Tau0    float64 `json:"tau0_nls_s"`
	Ec      float64 `json:"ec_V_m"`
	TauSim  float64 `json:"tau_1Ec_sim_s"`
	TauMerz float64 `json:"tau_1Ec_merz_s"`
	ErrPct  float64 `json:"error_pct"`
}

type arrheniusMaterialArtifact struct {
	Description string                  `json:"description"`
	Results     []arrheniusMaterialPoint `json:"results"`
}

func TestLiteratureArrheniusFamily_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	switchPath := filepath.Join(repoRoot, "output", "validation", "literature", "module1_arrhenius_switching.json")
	multiPath := filepath.Join(repoRoot, "output", "validation", "literature", "module1_arrhenius_multimaterial.json")

	b1, err := os.ReadFile(switchPath)
	if err != nil {
		t.Fatalf("read %s: %v", switchPath, err)
	}
	var sw arrheniusSwitchArtifact
	if err := json.Unmarshal(b1, &sw); err != nil {
		t.Fatalf("decode %s: %v", switchPath, err)
	}
	if sw.Material == "" || sw.Tau0 <= 0 || sw.Ec <= 0 {
		t.Fatalf("%s invalid material/tau0/ec", switchPath)
	}
	if len(sw.Results) < 8 {
		t.Fatalf("%s expected >=8 points, got %d", switchPath, len(sw.Results))
	}
	prevE := 0.0
	prevTau := 0.0
	for i, r := range sw.Results {
		if r.EOverEc <= 0 || r.TauSim <= 0 || r.TauMerz <= 0 {
			t.Fatalf("%s[%d] invalid positive values", switchPath, i)
		}
		if i > 0 {
			if r.EOverEc <= prevE {
				t.Fatalf("%s[%d] non-increasing E_over_Ec: prev=%g cur=%g", switchPath, i, prevE, r.EOverEc)
			}
			if r.TauSim >= prevTau {
				t.Fatalf("%s[%d] tau_simulated not decreasing with field: prev=%g cur=%g", switchPath, i, prevTau, r.TauSim)
			}
		}
		if r.ErrPct != 0 {
			t.Fatalf("%s[%d] error_pct=%g want 0", switchPath, i, r.ErrPct)
		}
		prevE = r.EOverEc
		prevTau = r.TauSim
	}

	b2, err := os.ReadFile(multiPath)
	if err != nil {
		t.Fatalf("read %s: %v", multiPath, err)
	}
	var mm arrheniusMaterialArtifact
	if err := json.Unmarshal(b2, &mm); err != nil {
		t.Fatalf("decode %s: %v", multiPath, err)
	}
	if mm.Description == "" || len(mm.Results) < 9 {
		t.Fatalf("%s invalid description/result size (%d)", multiPath, len(mm.Results))
	}
	for i, r := range mm.Results {
		if r.Name == "" || r.Tau0 <= 0 || r.Ec <= 0 || r.TauSim <= 0 || r.TauMerz <= 0 {
			t.Fatalf("%s[%d] invalid non-positive values/name", multiPath, i)
		}
		if r.ErrPct != 0 {
			t.Fatalf("%s[%d] error_pct=%g want 0", multiPath, i, r.ErrPct)
		}
	}
}
