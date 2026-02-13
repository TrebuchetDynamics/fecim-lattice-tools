package arraysim

import "testing"

func TestRunBatchMNISTBenchmark(t *testing.T) {
	results := RunBatchMNISTBenchmark([]BenchmarkCase{
		{Name: "baseline", ArraySize: 64, ADCBits: 6, Device: "FeFET"},
		{Name: "lowpower", ArraySize: 32, ADCBits: 5, Device: "RRAM"},
	})
	if len(results) != 2 {
		t.Fatalf("len=%d want 2", len(results))
	}
	if results[0].Name != "baseline" {
		t.Fatalf("results should be name-sorted, got first=%q", results[0].Name)
	}
	for _, r := range results {
		if r.Accuracy <= 0 || r.LatencyMS <= 0 || r.Throughput <= 0 {
			t.Fatalf("invalid benchmark result: %+v", r)
		}
	}
}
