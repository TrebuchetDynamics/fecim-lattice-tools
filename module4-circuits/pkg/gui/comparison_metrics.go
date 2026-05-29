//go:build legacy_fyne

package gui

import (
	"fecim-lattice-tools/module4-circuits/pkg/gui/comparison"
)

type comparisonMetricRow = comparison.MetricRow
type DesignSweepPoint = comparison.DesignSweepPoint
type MonteCarloStats = comparison.MonteCarloStats

func computeComparisonMetrics(arraySize int) (comparisonMetricRow, comparisonMetricRow, comparisonMetricRow) {
	return comparison.ComputeMetrics(arraySize)
}

func metricLatency(v float64) string { return comparison.MetricLatency(v) }
func metricEnergy(v float64) string  { return comparison.MetricEnergy(v) }
func metricGOPS(v float64) string    { return comparison.MetricGOPS(v) }

// BuildDesignSpaceSweep returns a lightweight design-space sweep for array size x ADC bits x device.
func BuildDesignSpaceSweep(arraySizes, adcBits []int, devices []string) []DesignSweepPoint {
	return comparison.BuildDesignSpaceSweep(arraySizes, adcBits, devices)
}

// RunProcessVariationMonteCarlo performs a simple Gaussian variation sampling around a base value.
func RunProcessVariationMonteCarlo(baseValue, sigmaFraction float64, samples int, seed int64) MonteCarloStats {
	return comparison.RunProcessVariationMonteCarlo(baseValue, sigmaFraction, samples, seed)
}
