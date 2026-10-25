package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type literaturePELoopArtifact struct {
	SchemaVersion          string `json:"schema_version"`
	TimestampUTC           string `json:"timestamp_utc"`
	Commit                 string `json:"commit"`
	Gate                   string `json:"gate"`
	TestID                 string `json:"test_id"`
	Verdict                string `json:"verdict"`
	GeneratedAt            string `json:"generated_at"`
	Dataset                string `json:"dataset"`
	DOI                    string `json:"doi"`
	Pass                   bool   `json:"pass"`
	ComparisonKind         string `json:"comparison_kind"`
	Evidence               string `json:"evidence"`
	ExperimentalValidation *bool  `json:"experimental_validation"`
}

func TestLiteraturePELoopArtifactFamily_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	glob := validationArtifactGlob(repoRoot, "literature", "module1_pe_loop_*.json")
	paths, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("glob %s: %v", glob, err)
	}
	if len(paths) != 8 {
		t.Fatalf("expected exactly 8 reference-regression artifacts after BTO exclusion, got %d", len(paths))
	}

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec literaturePELoopArtifact
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.SchemaVersion != "v1" || rec.TimestampUTC != "1970-01-01T00:00:00Z" || rec.Commit != "reproducible-build" {
			t.Fatalf("%s envelope mismatch schema=%q ts=%q commit=%q", p, rec.SchemaVersion, rec.TimestampUTC, rec.Commit)
		}
		if rec.GeneratedAt != "1970-01-01T00:00:00Z" {
			t.Fatalf("%s generated_at=%q want deterministic", p, rec.GeneratedAt)
		}
		if rec.Gate == "" || rec.TestID == "" || rec.Verdict == "" || rec.Dataset == "" {
			t.Fatalf("%s missing required fields gate/test_id/verdict/dataset", p)
		}
		if rec.DOI == "" || !strings.HasPrefix(strings.ToLower(rec.DOI), "10.") {
			t.Fatalf("%s invalid doi=%q", p, rec.DOI)
		}
		if strings.Contains(filepath.Base(p), "bto2021_cryst11101192_hysteresis_digitized") {
			t.Fatalf("excluded BTO extraction still counted: %s", p)
		}
		if strings.Contains(filepath.Base(p), "park2015") {
			if !strings.Contains(filepath.Base(p), "synthetic_shape") || rec.ComparisonKind != "synthetic-regression" || rec.Evidence != "simulation-default" || rec.ExperimentalValidation == nil || *rec.ExperimentalValidation {
				t.Fatalf("Park synthetic regression promoted to experimental evidence: %+v", rec)
			}
		}
		if !rec.Pass {
			t.Fatalf("%s pass=false", p)
		}
	}
}
