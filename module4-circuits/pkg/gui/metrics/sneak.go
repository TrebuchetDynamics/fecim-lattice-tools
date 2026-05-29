//go:build legacy_fyne

package metrics

import (
	"fmt"
	"math"
	"sort"
)

const sneakCurrentEpsilonA = 1e-12

type scaledUnit struct {
	unit  string
	scale float64
}

// SneakCellImpact records the sneak-path current magnitude for one cell.
type SneakCellImpact struct {
	Row      int     `json:"row"`
	Col      int     `json:"col"`
	CurrentA float64 `json:"current_a"`
}

// SneakPathMetrics summarizes non-target sneak-path current in an array snapshot.
type SneakPathMetrics struct {
	TotalSneakCurrentA float64           `json:"total_sneak_current_a"`
	MaxSneakCurrentA   float64           `json:"max_sneak_current_a"`
	AffectedCells      int               `json:"affected_cells"`
	HalfSelectCells    int               `json:"half_select_cells"`
	SneakOnlyCells     int               `json:"sneak_only_cells"`
	TopAffectedCells   []SneakCellImpact `json:"top_affected_cells"`
}

// ComputeSneakPath calculates aggregate sneak-path current metrics for all
// non-target cells in a current matrix.
func ComputeSneakPath(currents [][]float64, selectedRow, selectedCol int) SneakPathMetrics {
	metrics := SneakPathMetrics{}
	if len(currents) == 0 {
		return metrics
	}

	top := make([]SneakCellImpact, 0, 8)
	for r := range currents {
		for c, current := range currents[r] {
			if r == selectedRow && c == selectedCol {
				continue
			}
			mag := math.Abs(current)
			if mag <= sneakCurrentEpsilonA {
				continue
			}
			metrics.TotalSneakCurrentA += mag
			metrics.AffectedCells++
			if mag > metrics.MaxSneakCurrentA {
				metrics.MaxSneakCurrentA = mag
			}
			if r == selectedRow || c == selectedCol {
				metrics.HalfSelectCells++
			} else {
				metrics.SneakOnlyCells++
			}
			top = append(top, SneakCellImpact{Row: r, Col: c, CurrentA: mag})
		}
	}

	sort.Slice(top, func(i, j int) bool {
		if top[i].CurrentA == top[j].CurrentA {
			if top[i].Row == top[j].Row {
				return top[i].Col < top[j].Col
			}
			return top[i].Row < top[j].Row
		}
		return top[i].CurrentA > top[j].CurrentA
	})
	if len(top) > 3 {
		top = top[:3]
	}
	metrics.TopAffectedCells = top
	return metrics
}

// FormatSneakPathSummary formats a compact human-readable sneak-path summary.
func FormatSneakPathSummary(metrics SneakPathMetrics) string {
	if metrics.AffectedCells == 0 {
		return "0T1R: sneak current 0 A (0 affected cells)"
	}
	summary := fmt.Sprintf(
		"0T1R: sneak=%s, affected=%d (half-select=%d, sneak-only=%d)",
		formatCurrentA(metrics.TotalSneakCurrentA),
		metrics.AffectedCells,
		metrics.HalfSelectCells,
		metrics.SneakOnlyCells,
	)
	if len(metrics.TopAffectedCells) == 0 {
		return summary
	}
	top := metrics.TopAffectedCells[0]
	return fmt.Sprintf("%s, top=[%d,%d] %s", summary, top.Row, top.Col, formatCurrentA(top.CurrentA))
}

func formatCurrentA(currentA float64) string {
	return formatSignedScaled(currentA, []scaledUnit{
		{unit: "A", scale: 1.0},
		{unit: "mA", scale: 1e-3},
		{unit: "uA", scale: 1e-6},
		{unit: "nA", scale: 1e-9},
		{unit: "pA", scale: 1e-12},
	})
}

func formatSignedScaled(value float64, units []scaledUnit) string {
	absValue := math.Abs(value)
	if absValue < 1e-12 {
		return fmt.Sprintf("0 %s", units[0].unit)
	}

	chosen := units[len(units)-1]
	for _, candidate := range units {
		if absValue >= candidate.scale {
			chosen = candidate
			break
		}
	}
	scaled := value / chosen.scale
	absScaled := math.Abs(scaled)
	format := "%+.3f"
	switch {
	case absScaled >= 100:
		format = "%+.0f"
	case absScaled >= 10:
		format = "%+.1f"
	case absScaled >= 1:
		format = "%+.2f"
	}
	return fmt.Sprintf(format+" %s", scaled, chosen.unit)
}
