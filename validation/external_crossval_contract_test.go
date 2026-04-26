package validation

import (
	"encoding/json"
	"os"
	"testing"
)

type externalCrossvalRecord struct {
	N       int     `json:"n"`
	RWL     float64 `json:"RWL_ohm"`
	RBL     float64 `json:"RBL_ohm"`
	MaxIErr float64 `json:"maxIErr_A"`
	MaxVErr float64 `json:"maxVErr_V"`
	PassI   bool    `json:"pass_I"`
	PassV   bool    `json:"pass_V"`
}

func TestExternalMVMCrossvalArtifacts_Contract(t *testing.T) {
	paths := []string{
		releaseArtifactPath("validation", "output", "validation", "external", "mvm_numpy_crossval_4x4.json"),
		releaseArtifactPath("validation", "output", "validation", "external", "mvm_numpy_crossval_8x8.json"),
		releaseArtifactPath("validation", "output", "validation", "external", "mvm_numpy_crossval_16x16.json"),
	}
	seen := map[int]bool{}
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec externalCrossvalRecord
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.N <= 0 || rec.RWL <= 0 || rec.RBL <= 0 {
			t.Fatalf("%s invalid geometry/resistance values", p)
		}
		if rec.MaxIErr < 0 || rec.MaxVErr < 0 {
			t.Fatalf("%s invalid negative errors", p)
		}
		if !rec.PassI || !rec.PassV {
			t.Fatalf("%s pass flags false: pass_I=%v pass_V=%v", p, rec.PassI, rec.PassV)
		}
		if rec.MaxVErr > 1e-12 {
			t.Fatalf("%s maxVErr_V=%g exceeds contract 1e-12", p, rec.MaxVErr)
		}
		seen[rec.N] = true
	}
	for _, n := range []int{4, 8, 16} {
		if !seen[n] {
			t.Fatalf("missing crossval artifact for n=%d", n)
		}
	}
}
