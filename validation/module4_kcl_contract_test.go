package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type kclConvergencePoint struct {
	Size        int     `json:"size"`
	MaxKCL      float64 `json:"max_kcl_A"`
	ExpectedN2E float64 `json:"expected_O_N2_eps"`
}

type kclKvlRecord struct {
	Name      string  `json:"name"`
	Size      int     `json:"size"`
	MaxKCL    float64 `json:"max_kcl_error_A"`
	MaxKVL    float64 `json:"max_kvl_error_V"`
	PassKCL   bool    `json:"pass_kcl"`
	PassKVL   bool    `json:"pass_kvl"`
}

type kclKvlArtifact struct {
	Description string         `json:"description"`
	Results     []kclKvlRecord `json:"results"`
	TolA        float64        `json:"tolerance_A"`
	TolV        float64        `json:"tolerance_V"`
}

func TestModule4KCLArtifacts_Contract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	convPath := filepath.Join(repoRoot, "validation", "output", "validation", "module4", "kcl_convergence.json")
	exhPath := filepath.Join(repoRoot, "validation", "output", "validation", "module4", "kcl_kvl_exhaustive.json")

	convBytes, err := os.ReadFile(convPath)
	if err != nil {
		t.Fatalf("read %s: %v", convPath, err)
	}
	var conv []kclConvergencePoint
	if err := json.Unmarshal(convBytes, &conv); err != nil {
		t.Fatalf("decode %s: %v", convPath, err)
	}
	if len(conv) < 5 {
		t.Fatalf("%s expected >=5 convergence points, got %d", convPath, len(conv))
	}
	for i, p := range conv {
		if p.Size <= 0 {
			t.Fatalf("%s[%d] invalid size=%d", convPath, i, p.Size)
		}
		if p.MaxKCL < 0 || p.ExpectedN2E <= 0 {
			t.Fatalf("%s[%d] invalid kcl metrics: max=%g expected=%g", convPath, i, p.MaxKCL, p.ExpectedN2E)
		}
		if p.MaxKCL > p.ExpectedN2E {
			t.Fatalf("%s[%d] max_kcl_A=%g exceeds expected_O_N2_eps=%g", convPath, i, p.MaxKCL, p.ExpectedN2E)
		}
	}

	exhBytes, err := os.ReadFile(exhPath)
	if err != nil {
		t.Fatalf("read %s: %v", exhPath, err)
	}
	var exh kclKvlArtifact
	if err := json.Unmarshal(exhBytes, &exh); err != nil {
		t.Fatalf("decode %s: %v", exhPath, err)
	}
	if exh.Description == "" {
		t.Fatalf("%s empty description", exhPath)
	}
	if exh.TolA <= 0 || exh.TolV <= 0 {
		t.Fatalf("%s invalid tolerances A=%g V=%g", exhPath, exh.TolA, exh.TolV)
	}
	if len(exh.Results) == 0 {
		t.Fatalf("%s empty results", exhPath)
	}
	for i, r := range exh.Results {
		if r.Name == "" || r.Size <= 0 {
			t.Fatalf("%s[%d] invalid name/size", exhPath, i)
		}
		if !r.PassKCL || !r.PassKVL {
			t.Fatalf("%s[%d] pass flags false: kcl=%v kvl=%v", exhPath, i, r.PassKCL, r.PassKVL)
		}
		if r.MaxKCL > exh.TolA || r.MaxKVL > exh.TolV {
			t.Fatalf("%s[%d] exceeded tolerance: kcl=%g/%g kvl=%g/%g", exhPath, i, r.MaxKCL, exh.TolA, r.MaxKVL, exh.TolV)
		}
	}
}
