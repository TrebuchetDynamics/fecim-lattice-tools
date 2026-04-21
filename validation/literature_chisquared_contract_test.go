package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type chiSquaredResult struct {
	Material string  `json:"material"`
	DOI      string  `json:"doi"`
	Chi2r    float64 `json:"chi2_r"`
	Pass     bool    `json:"pass"`
}

type chiSquaredArtifact struct {
	Description string             `json:"description"`
	SigmaRule   string             `json:"sigma_rule"`
	Results     []chiSquaredResult `json:"results"`
}

func TestLiteraturePEChiSquared_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	p := validationArtifactPath(repoRoot, "literature", "pe_chisquared_fit.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec chiSquaredArtifact
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if rec.Description == "" || rec.SigmaRule == "" {
		t.Fatalf("%s missing description/sigma_rule", p)
	}
	if len(rec.Results) < 6 {
		t.Fatalf("%s expected >=6 results, got %d", p, len(rec.Results))
	}

	failSet := map[string]bool{}
	for i, r := range rec.Results {
		if r.Material == "" || r.DOI == "" || r.Chi2r <= 0 {
			t.Fatalf("%s invalid result[%d]: %+v", p, i, r)
		}
		if !r.Pass {
			failSet[r.Material] = true
		}
	}

	// Current validated baseline keeps exactly two difficult datasets flagged false.
	if len(failSet) != 2 || !failSet["pzt2024_nano14050432_fig2_thinfilm"] || !failSet["bto2021_cryst11101192_hysteresis"] {
		t.Fatalf("%s unexpected fail-set baseline: %+v", p, failSet)
	}
}
