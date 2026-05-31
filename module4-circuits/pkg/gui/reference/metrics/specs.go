//go:build legacy_fyne

package metrics

import "fmt"

// ComponentRow contains one row in the reference component summary table.
type ComponentRow struct {
	Component string
	Count     string
	Power     string
	Area      string
	Latency   string
}

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

// ComponentRows returns the canonical reference component summary rows for an array size.
func ComponentRows(size int) []ComponentRow {
	cells := size * size
	return []ComponentRow{
		{Component: "FeFET Array", Count: fmt.Sprintf("%d", cells), Power: "0.1 mW", Area: "0.01 mm²", Latency: "5 ns"},
		{Component: "DACs", Count: fmt.Sprintf("%d", size), Power: "3.2 mW", Area: "0.02 mm²", Latency: "10 ns"},
		{Component: "TIAs", Count: fmt.Sprintf("%d", size), Power: "1.6 mW", Area: "0.01 mm²", Latency: "11 ns"},
		{Component: "ADCs", Count: fmt.Sprintf("%d", size), Power: "16 mW", Area: "0.04 mm²", Latency: "50 ns"},
		{Component: "Control", Count: "1", Power: "0.5 mW", Area: "0.01 mm²", Latency: "2 ns"},
	}
}

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
