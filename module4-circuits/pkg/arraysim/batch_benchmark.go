package arraysim

import "sort"

// BenchmarkCase is one MNIST benchmark configuration.
type BenchmarkCase struct {
	Name      string
	ArraySize int
	ADCBits   int
	Device    string
}

// BenchmarkResult stores estimated metrics for one benchmark run.
type BenchmarkResult struct {
	BenchmarkCase
	Accuracy   float64
	LatencyMS  float64
	EnergyUj   float64
	Throughput float64
}

// RunBatchMNISTBenchmark executes configurable benchmark cases and returns results.
func RunBatchMNISTBenchmark(cases []BenchmarkCase) []BenchmarkResult {
	out := make([]BenchmarkResult, 0, len(cases))
	for _, c := range cases {
		lat, en, acc := estimatePoint(c.ArraySize, c.ADCBits, c.Device)
		latMS := lat / 1e6 // ns -> ms
		if latMS <= 0 {
			latMS = 1e-6
		}
		throughput := 1000.0 / latMS
		out = append(out, BenchmarkResult{
			BenchmarkCase: c,
			Accuracy:      acc,
			LatencyMS:     latMS,
			EnergyUj:      en / 1e6,
			Throughput:    throughput,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}
