//go:build legacy_fyne

package reference

import "fmt"

// SpecSummary contains derived values shown in the reference specification summary.
type SpecSummary struct {
	Size             int
	Cells            int
	ThroughputGOPS   float64
	EfficiencyGOPSW  int
	TotalPowerMW     float64
	TotalAreaMM2     float64
	TotalLatencyNS   int
	ThroughputText   string
	EfficiencyText   string
	TotalPowerText   string
	TotalAreaText    string
	TotalLatencyText string
}

// ParseArraySize parses a selected array size and falls back to 32 when invalid.
func ParseArraySize(selected string) int {
	var size int
	fmt.Sscanf(selected, "%d", &size)
	if size == 0 {
		return 32
	}
	return size
}

// NewSpecSummary returns the derived summary values for an array size.
func NewSpecSummary(size int, latencyNS float64) SpecSummary {
	if size <= 0 {
		size = 32
	}
	if latencyNS <= 0 {
		latencyNS = 76
	}
	cells := size * size
	throughput := float64(cells) / latencyNS
	efficiency := int(throughput * 1000 / 21.4)
	return SpecSummary{
		Size:             size,
		Cells:            cells,
		ThroughputGOPS:   throughput,
		EfficiencyGOPSW:  efficiency,
		TotalPowerMW:     21.4,
		TotalAreaMM2:     0.09,
		TotalLatencyNS:   76,
		ThroughputText:   fmt.Sprintf("%d MACs (Ops) / %.0fns = %.1f GOPS", cells, latencyNS, throughput),
		EfficiencyText:   fmt.Sprintf("%.1f GOPS / 21.4 mW = %d GOPS/W", throughput, efficiency),
		TotalPowerText:   "21.4 mW",
		TotalAreaText:    "0.09 mm²",
		TotalLatencyText: "76 ns",
	}
}
