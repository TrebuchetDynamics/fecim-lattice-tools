package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type legacyBoundaryContract struct {
	Suite     string `json:"suite"`
	Material  string `json:"material"`
	Model     string `json:"model"`
	Timestamp string `json:"timestamp"`
	AllPass   bool   `json:"all_pass"`
}

func readLegacyBoundaryContract(t *testing.T, p string) legacyBoundaryContract {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec legacyBoundaryContract
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if rec.Suite != "headless-wrd-ispp-regression" {
		t.Fatalf("%s suite=%q", p, rec.Suite)
	}
	if rec.Model == "" || rec.Material == "" {
		t.Fatalf("%s missing model/material", p)
	}
	if rec.Timestamp != "1970-01-01T00:00:00Z" {
		t.Fatalf("%s timestamp=%q want deterministic", p, rec.Timestamp)
	}
	return rec
}

func TestLegacyRegressionBoundaryConsistency_RootVsController(t *testing.T) {
	repoRoot := filepath.Clean("..")
	cases := []struct {
		name string
		root string
		ctl  string
	}{
		{
			name: "lk_legacy",
			root: filepath.Join(repoRoot, "output", "regression", "lk_wrd_ispp_regression.json"),
			ctl:  filepath.Join(repoRoot, "module1-hysteresis", "pkg", "controller", "output", "regression", "lk_wrd_ispp_regression.json"),
		},
		{
			name: "preisach_legacy",
			root: filepath.Join(repoRoot, "output", "regression", "preisach_wrd_ispp_regression.json"),
			ctl:  filepath.Join(repoRoot, "module1-hysteresis", "pkg", "controller", "output", "regression", "preisach_wrd_ispp_regression.json"),
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			rootRec := readLegacyBoundaryContract(t, tc.root)
			ctlRec := readLegacyBoundaryContract(t, tc.ctl)

			if rootRec.Model != ctlRec.Model {
				t.Fatalf("model mismatch root=%q controller=%q", rootRec.Model, ctlRec.Model)
			}
			if rootRec.Material != ctlRec.Material {
				t.Fatalf("material mismatch root=%q controller=%q", rootRec.Material, ctlRec.Material)
			}
			if rootRec.AllPass != ctlRec.AllPass {
				t.Fatalf("all_pass mismatch root=%v controller=%v", rootRec.AllPass, ctlRec.AllPass)
			}
		})
	}
}
