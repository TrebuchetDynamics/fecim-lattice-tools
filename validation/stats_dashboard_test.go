package validation

import (
	"encoding/json"
	"testing"
)

func TestGenerateStatsDashboardFromSampleMC(t *testing.T) {
	samples := []MCSample{
		{Corner: "tt", Metric: "accuracy", Value: 0.93, Pass: true, Iteration: 1},
		{Corner: "ss", Metric: "accuracy", Value: 0.88, Pass: false, Iteration: 1},
		{Corner: "ff", Metric: "accuracy", Value: 0.95, Pass: true, Iteration: 2},
		{Corner: "ss", Metric: "latency_ns", Value: 75, Pass: false, Iteration: 2},
		{Corner: "tt", Metric: "latency_ns", Value: 60, Pass: true, Iteration: 3},
	}

	data, raw, err := GenerateStatsDashboard(samples, DashboardConfig{WorstCornerFailureThreshold: 0.5})
	if err != nil {
		t.Fatalf("GenerateStatsDashboard failed: %v", err)
	}
	if len(data.Distributions) == 0 || len(data.WorstCorners) == 0 || len(data.PassRateTrend) == 0 {
		t.Fatalf("dashboard missing sections: %+v", data)
	}

	var decoded DashboardData
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("dashboard JSON invalid: %v", err)
	}
	if len(decoded.Distributions) != len(data.Distributions) {
		t.Fatalf("distribution count mismatch after JSON encode/decode")
	}
}
