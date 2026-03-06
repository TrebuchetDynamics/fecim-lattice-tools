package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type peLoopThresholds struct {
	PrErrPctMax       float64 `json:"pr_err_pct_max"`
	EcErrPctMax       float64 `json:"ec_err_pct_max"`
	RMSEOverPsMax     float64 `json:"rmse_over_ps_max"`
	LoopAreaErrPctMax float64 `json:"loop_area_err_pct_max"`
}

type peLoopMetrics struct {
	PrErrPct       float64 `json:"pr_err_pct"`
	EcErrPct       float64 `json:"ec_err_pct"`
	RMSEOverPs     float64 `json:"rmse_over_ps"`
	LoopAreaErrPct float64 `json:"loop_area_err_pct"`
}

type peLoopArtifactContract struct {
	Dataset    string           `json:"dataset"`
	Generated  string           `json:"generated_at"`
	Pass       bool             `json:"pass"`
	Thresholds peLoopThresholds `json:"thresholds"`
	Metrics    peLoopMetrics    `json:"metrics"`
}

func TestLiteraturePELoop_ThresholdConsistencyContract(t *testing.T) {
	repoRoot := filepath.Clean("..")
	paths, err := filepath.Glob(filepath.Join(repoRoot, "output", "validation", "literature", "module1_pe_loop_*.json"))
	if err != nil {
		t.Fatalf("glob pe loop artifacts: %v", err)
	}
	if len(paths) != 9 {
		t.Fatalf("expected 9 pe-loop artifacts, got %d", len(paths))
	}

	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("read %s: %v", p, err)
		}
		var rec peLoopArtifactContract
		if err := json.Unmarshal(b, &rec); err != nil {
			t.Fatalf("decode %s: %v", p, err)
		}
		if rec.Dataset == "" {
			t.Fatalf("%s missing dataset", p)
		}
		if rec.Generated != "1970-01-01T00:00:00Z" {
			t.Fatalf("%s generated_at=%q want deterministic", p, rec.Generated)
		}
		th := rec.Thresholds
		if th.PrErrPctMax <= 0 || th.EcErrPctMax <= 0 || th.RMSEOverPsMax <= 0 || th.LoopAreaErrPctMax <= 0 {
			t.Fatalf("%s invalid thresholds: %+v", p, th)
		}
		m := rec.Metrics
		within := m.PrErrPct <= th.PrErrPctMax &&
			m.EcErrPct <= th.EcErrPctMax &&
			m.RMSEOverPs <= th.RMSEOverPsMax &&
			m.LoopAreaErrPct <= th.LoopAreaErrPctMax
		if rec.Pass != within {
			t.Fatalf("%s pass/threshold mismatch: pass=%v within=%v metrics=%+v thresholds=%+v", p, rec.Pass, within, m, th)
		}
	}
}
