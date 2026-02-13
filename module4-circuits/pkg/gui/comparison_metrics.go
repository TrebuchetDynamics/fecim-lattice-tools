package gui

import "fmt"

type comparisonMetricRow struct {
	Label      string
	LatencyNS  float64
	EnergyPJ   float64
	TOPSW      float64
	EnergyOpPJ float64
}

func computeComparisonMetrics(arraySize int) (comparisonMetricRow, comparisonMetricRow, comparisonMetricRow) {
	if arraySize <= 0 {
		arraySize = 8
	}
	macs := float64(arraySize * arraySize)
	scale := macs / 64.0

	cpu := comparisonMetricRow{Label: "CPU", LatencyNS: 500 * scale, EnergyPJ: 64000 * scale}
	gpu := comparisonMetricRow{Label: "GPU", LatencyNS: 50 * scale, EnergyPJ: 6400 * scale}
	fefet := comparisonMetricRow{Label: "FeFET", LatencyNS: 76 * scale, EnergyPJ: 2.9 * scale}

	rows := []*comparisonMetricRow{&cpu, &gpu, &fefet}
	for _, r := range rows {
		if r.LatencyNS > 0 {
			r.TOPSW = (2.0 * macs) / r.LatencyNS / 1e3
		}
		if macs > 0 {
			r.EnergyOpPJ = r.EnergyPJ / macs
		}
	}
	return cpu, gpu, fefet
}

func metricLatency(v float64) string { return fmt.Sprintf("%.0f ns", v) }
func metricEnergy(v float64) string  { return fmt.Sprintf("%.1f pJ", v) }
func metricTOPSW(v float64) string   { return fmt.Sprintf("%.3f", v) }
