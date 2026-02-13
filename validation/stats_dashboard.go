package validation

import (
	"encoding/json"
	"sort"
)

// MCSample is a Monte Carlo sample for one validation run.
type MCSample struct {
	Corner    string  `json:"corner"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Pass      bool    `json:"pass"`
	Iteration int     `json:"iteration"`
}

// DashboardConfig controls alert thresholds.
type DashboardConfig struct {
	WorstCornerFailureThreshold float64 `json:"worst_corner_failure_threshold"`
}

// MetricDistribution summarizes one metric.
type MetricDistribution struct {
	Metric string  `json:"metric"`
	Count  int     `json:"count"`
	Mean   float64 `json:"mean"`
	P05    float64 `json:"p05"`
	P50    float64 `json:"p50"`
	P95    float64 `json:"p95"`
}

// WorstCornerAlert flags corners with excessive fail rate.
type WorstCornerAlert struct {
	Corner   string  `json:"corner"`
	FailRate float64 `json:"fail_rate"`
	Alert    bool    `json:"alert"`
}

// PassRatePoint captures pass-rate evolution through iterations.
type PassRatePoint struct {
	Iteration int     `json:"iteration"`
	PassRate  float64 `json:"pass_rate"`
}

// DashboardData is JSON-serializable statistical dashboard output.
type DashboardData struct {
	Distributions []MetricDistribution `json:"distributions"`
	WorstCorners  []WorstCornerAlert   `json:"worst_corners"`
	PassRateTrend []PassRatePoint      `json:"pass_rate_trend"`
}

func GenerateStatsDashboard(samples []MCSample, cfg DashboardConfig) (DashboardData, []byte, error) {
	if cfg.WorstCornerFailureThreshold <= 0 {
		cfg.WorstCornerFailureThreshold = 0.20
	}

	metricValues := map[string][]float64{}
	cornerCounts := map[string]int{}
	cornerFails := map[string]int{}
	iterCounts := map[int]int{}
	iterPass := map[int]int{}

	for _, s := range samples {
		metricValues[s.Metric] = append(metricValues[s.Metric], s.Value)
		cornerCounts[s.Corner]++
		if !s.Pass {
			cornerFails[s.Corner]++
		}
		iterCounts[s.Iteration]++
		if s.Pass {
			iterPass[s.Iteration]++
		}
	}

	data := DashboardData{
		Distributions: make([]MetricDistribution, 0, len(metricValues)),
		WorstCorners:  make([]WorstCornerAlert, 0, len(cornerCounts)),
	}

	for metric, vals := range metricValues {
		sorted := append([]float64(nil), vals...)
		sort.Float64s(sorted)
		data.Distributions = append(data.Distributions, MetricDistribution{
			Metric: metric,
			Count:  len(sorted),
			Mean:   Mean(sorted),
			P05:    percentile(sorted, 0.05),
			P50:    percentile(sorted, 0.50),
			P95:    percentile(sorted, 0.95),
		})
	}
	sort.Slice(data.Distributions, func(i, j int) bool { return data.Distributions[i].Metric < data.Distributions[j].Metric })

	for corner, count := range cornerCounts {
		failRate := float64(cornerFails[corner]) / float64(count)
		data.WorstCorners = append(data.WorstCorners, WorstCornerAlert{
			Corner:   corner,
			FailRate: failRate,
			Alert:    failRate >= cfg.WorstCornerFailureThreshold,
		})
	}
	sort.Slice(data.WorstCorners, func(i, j int) bool { return data.WorstCorners[i].FailRate > data.WorstCorners[j].FailRate })

	iterations := make([]int, 0, len(iterCounts))
	for it := range iterCounts {
		iterations = append(iterations, it)
	}
	sort.Ints(iterations)
	data.PassRateTrend = make([]PassRatePoint, 0, len(iterations))
	for _, it := range iterations {
		data.PassRateTrend = append(data.PassRateTrend, PassRatePoint{
			Iteration: it,
			PassRate:  float64(iterPass[it]) / float64(iterCounts[it]),
		})
	}

	raw, err := json.MarshalIndent(data, "", "  ")
	return data, raw, err
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if p <= 0 {
		return sorted[0]
	}
	if p >= 1 {
		return sorted[len(sorted)-1]
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}
