package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type controllerRootRegressionContract struct {
	Suite     string `json:"suite"`
	Material  string `json:"material"`
	Model     string `json:"model"`
	Timestamp string `json:"timestamp"`
}

func TestControllerRootRegressionArtifacts_MetadataContract(t *testing.T) {
	paths := []string{
		releaseArtifactPath("module1-hysteresis", "pkg", "controller", "output", "regression", "lk_wrd_ispp_regression.json"),
		releaseArtifactPath("module1-hysteresis", "pkg", "controller", "output", "regression", "preisach_wrd_ispp_regression.json"),
	}

	for _, p := range paths {
		p := p
		t.Run(filepath.Base(p), func(t *testing.T) {
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatalf("read %s: %v", p, err)
			}
			var rec controllerRootRegressionContract
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
		})
	}
}
