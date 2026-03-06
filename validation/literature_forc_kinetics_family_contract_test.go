package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type forcArtifact struct {
	MaterialID      string  `json:"material_id"`
	GeneratedAt     string  `json:"generated_at"`
	IntegralErrPct  float64 `json:"integral_err_pct"`
	HCErrPct        float64 `json:"hc_err_pct"`
	NegativeFrac    float64 `json:"negative_fraction"`
	Pass            bool    `json:"pass"`
}

type kineticsArtifact struct {
	MaterialID    string  `json:"material_id"`
	GeneratedAt   string  `json:"generated_at"`
	Monotonic     bool    `json:"monotonic"`
	LogisticRMSE  float64 `json:"logistic_rmse"`
	FracAtHalfEc  float64 `json:"frac_at_half_ec"`
	FracAtEc      float64 `json:"frac_at_ec"`
	FracAtTwoEc   float64 `json:"frac_at_two_ec"`
	Pass          bool    `json:"pass"`
}

func TestLiteratureFORCAndKineticsFamily_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	forcPaths, err := filepath.Glob(filepath.Join(repoRoot, "output", "validation", "literature", "module1_forc_*.json"))
	if err != nil {
		t.Fatalf("glob forc: %v", err)
	}
	kinPaths, err := filepath.Glob(filepath.Join(repoRoot, "output", "validation", "literature", "module1_switching_kinetics_*.json"))
	if err != nil {
		t.Fatalf("glob kinetics: %v", err)
	}
	if len(forcPaths) != 2 {
		t.Fatalf("expected 2 FORC artifacts, got %d", len(forcPaths))
	}
	if len(kinPaths) != 3 {
		t.Fatalf("expected 3 kinetics artifacts, got %d", len(kinPaths))
	}

	for _, p := range forcPaths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec forcArtifact
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.MaterialID == "" || rec.GeneratedAt != "1970-01-01T00:00:00Z" {
			t.Fatalf("%s invalid identity/timestamp: material_id=%q generated_at=%q", p, rec.MaterialID, rec.GeneratedAt)
		}
		if rec.NegativeFrac < 0 || rec.NegativeFrac > 1 {
			t.Fatalf("%s negative_fraction out of range: %g", p, rec.NegativeFrac)
		}
		if rec.IntegralErrPct < 0 || rec.HCErrPct < 0 {
			t.Fatalf("%s invalid error metrics: integral=%g hc=%g", p, rec.IntegralErrPct, rec.HCErrPct)
		}
		if !rec.Pass {
			t.Fatalf("%s pass=false", p)
		}
	}

	for _, p := range kinPaths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec kineticsArtifact
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.MaterialID == "" || rec.GeneratedAt != "1970-01-01T00:00:00Z" {
			t.Fatalf("%s invalid identity/timestamp: material_id=%q generated_at=%q", p, rec.MaterialID, rec.GeneratedAt)
		}
		if !(rec.FracAtHalfEc <= rec.FracAtEc && rec.FracAtEc <= rec.FracAtTwoEc) {
			t.Fatalf("%s invalid switching fractions ordering: half=%g ec=%g two=%g", p, rec.FracAtHalfEc, rec.FracAtEc, rec.FracAtTwoEc)
		}
		if rec.LogisticRMSE < 0 {
			t.Fatalf("%s negative logistic_rmse=%g", p, rec.LogisticRMSE)
		}
		if !rec.Monotonic || !rec.Pass {
			t.Fatalf("%s expected monotonic/pass true; got monotonic=%v pass=%v", p, rec.Monotonic, rec.Pass)
		}
	}
}
