package validation

import (
	"encoding/json"
	"os"
	"testing"
)

type module1FamilyCase struct {
	Name        string `json:"name"`
	TargetLevel int    `json:"target_level"`
}

type module1FamilyRecord struct {
	Suite     string              `json:"suite"`
	Material  string              `json:"material"`
	Model     string              `json:"model"`
	Timestamp string              `json:"timestamp"`
	Cases     []module1FamilyCase `json:"cases"`
}

func readModule1FamilyRecord(t *testing.T, p string) module1FamilyRecord {
	t.Helper()
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec module1FamilyRecord
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
	if len(rec.Cases) != 3 {
		t.Fatalf("%s expected 3 cases, got %d", p, len(rec.Cases))
	}
	wantTargets := map[int]bool{5: false, 15: false, 27: false}
	for i, c := range rec.Cases {
		if c.Name == "" || c.TargetLevel <= 0 {
			t.Fatalf("%s invalid case[%d]: %+v", p, i, c)
		}
		if _, ok := wantTargets[c.TargetLevel]; ok {
			wantTargets[c.TargetLevel] = true
		}
	}
	for lvl, seen := range wantTargets {
		if !seen {
			t.Fatalf("%s missing target_level=%d", p, lvl)
		}
	}
	return rec
}

func TestModule1RegressionFamily_CompletenessAndBoundaryConsistency(t *testing.T) {
	families := []struct {
		model    string
		material string
	}{
		{"lk", "default_hzo"},
		{"lk", "fecim_hzo"},
		{"lk", "literature_superlattice"},
		{"preisach", "default_hzo"},
		{"preisach", "fecim_hzo"},
		{"preisach", "literature_superlattice"},
	}

	for _, f := range families {
		name := f.model + "_" + f.material
		t.Run(name, func(t *testing.T) {
			file := f.model + "_wrd_ispp_regression_" + f.material + ".json"
			rootPath := releaseArtifactPath("output", "regression", "module1", file)
			ctlPath := releaseArtifactPath("module1-hysteresis", "pkg", "controller", "output", "regression", "module1", file)

			rootRec := readModule1FamilyRecord(t, rootPath)
			ctlRec := readModule1FamilyRecord(t, ctlPath)

			if rootRec.Model != ctlRec.Model || rootRec.Material != ctlRec.Material {
				t.Fatalf("boundary mismatch model/material root=(%s,%s) ctl=(%s,%s)", rootRec.Model, rootRec.Material, ctlRec.Model, ctlRec.Material)
			}
			for i := range rootRec.Cases {
				rc, cc := rootRec.Cases[i], ctlRec.Cases[i]
				if rc.Name != cc.Name || rc.TargetLevel != cc.TargetLevel {
					t.Fatalf("case mismatch idx=%d root=(%s,%d) ctl=(%s,%d)", i, rc.Name, rc.TargetLevel, cc.Name, cc.TargetLevel)
				}
			}
		})
	}
}
