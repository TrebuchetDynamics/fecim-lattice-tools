package validation

import (
	"encoding/json"
	"os"
	"testing"
)

type module4RootParityContract struct {
	Version       string `json:"version"`
	Profile       string `json:"profile"`
	GeneratedUnix int64  `json:"generated_unix"`
}

func TestModule4RootRegressionParity_MetadataContract(t *testing.T) {
	p := releaseArtifactPath("output", "regression", "module4", "gui_vs_headless_parity.json")
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	var rec module4RootParityContract
	if err := json.Unmarshal(b, &rec); err != nil {
		t.Fatalf("decode %s: %v", p, err)
	}
	if rec.Version != "v1" {
		t.Fatalf("%s version=%q want v1", p, rec.Version)
	}
	if rec.Profile == "" {
		t.Fatalf("%s profile is empty", p)
	}
	if rec.GeneratedUnix != 0 {
		t.Fatalf("%s generated_unix=%d want 0", p, rec.GeneratedUnix)
	}
}
