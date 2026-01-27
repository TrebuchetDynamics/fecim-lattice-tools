// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"errors"
	"math"
)

func init() {
	ErrInputSize = errors.New("input size exceeds array columns")
}

// WireParams contains wire resistance parameters for IR drop modeling.
type WireParams struct {
	RwordLine float64 // Word line resistance per unit (Ohm)
	RbitLine  float64 // Bit line resistance per unit (Ohm)
	Rcontact  float64 // Contact resistance (Ohm)
}

// DefaultWireParams returns typical wire parameters for a 45nm technology node.
func DefaultWireParams() *WireParams {
	return &WireParams{
		RwordLine: 2.5, // 2.5 Ohm per cell pitch
		RbitLine:  2.5, // 2.5 Ohm per cell pitch
		Rcontact:  50,  // 50 Ohm contact resistance
	}
}

// IRDropAnalysis contains the results of IR drop analysis.
type IRDropAnalysis struct {
	// Voltage drop matrices
	WordLineVoltages [][]float64 // Voltage at each word line position
	BitLineVoltages  [][]float64 // Voltage at each bit line position
	EffectiveVoltage [][]float64 // Effective voltage across each cell

	// Summary statistics
	MaxIRDrop      float64 // Maximum IR drop in array
	AvgIRDrop      float64 // Average IR drop
	IRDropVariance float64 // Variance in IR drop
	WorstCaseCell  [2]int  // Location of worst-case cell
}

// AnalyzeIRDrop performs IR drop analysis for the array.
// Uses iterative relaxation to solve the resistive network.
func (a *Array) AnalyzeIRDrop(input []float64, params *WireParams) *IRDropAnalysis {
	if params == nil {
		params = DefaultWireParams()
	}

	rows := a.config.Rows
	cols := a.config.Cols

	// Initialize voltage matrices
	wlVoltage := make([][]float64, rows)
	blVoltage := make([][]float64, rows)
	effVoltage := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		wlVoltage[i] = make([]float64, cols)
		blVoltage[i] = make([]float64, cols)
		effVoltage[i] = make([]float64, cols)
	}

	// Simple analytical model for IR drop
	// Driver topology: WL drivers on LEFT, BL sense amps/ground at TOP
	// This matches display convention where row 0 is at top
	for i := 0; i < rows; i++ {
		rowCurrent := a.estimateCurrent(i, input)
		for j := 0; j < cols; j++ {
			// Word line voltage drop (cumulative from left driver)
			// Drop increases with column index j (farther from driver)
			wlDrop := float64(j) * params.RwordLine * rowCurrent
			wlVoltage[i][j] = 1.0 - wlDrop // Assuming 1V input normalized

			// Bit line voltage (sense amp/ground at top, row 0)
			// Drop increases with row index i (farther from sense amp)
			blDrop := float64(i) * params.RbitLine * a.estimateColumnCurrent(j, input)
			blVoltage[i][j] = blDrop // Voltage above ground

			// Effective voltage across cell = WL voltage - BL voltage
			// Worst case is bottom-right (max i, max j): both drops are maximum
			effVoltage[i][j] = wlVoltage[i][j] - blVoltage[i][j]
			if effVoltage[i][j] < 0 {
				effVoltage[i][j] = 0
			}
		}
	}

	// Calculate statistics
	var maxDrop, sumDrop float64
	var worstCell [2]int
	count := 0

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			drop := 1.0 - effVoltage[i][j]
			if drop > maxDrop {
				maxDrop = drop
				worstCell = [2]int{i, j}
			}
			sumDrop += drop
			count++
		}
	}

	avgDrop := sumDrop / float64(count)

	// Calculate variance
	var variance float64
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			drop := 1.0 - effVoltage[i][j]
			variance += (drop - avgDrop) * (drop - avgDrop)
		}
	}
	variance /= float64(count)

	return &IRDropAnalysis{
		WordLineVoltages: wlVoltage,
		BitLineVoltages:  blVoltage,
		EffectiveVoltage: effVoltage,
		MaxIRDrop:        maxDrop,
		AvgIRDrop:        avgDrop,
		IRDropVariance:   variance,
		WorstCaseCell:    worstCell,
	}
}

// estimateCurrent estimates current draw for a word line.
func (a *Array) estimateCurrent(row int, input []float64) float64 {
	var current float64
	for j := 0; j < len(input) && j < a.config.Cols; j++ {
		gNorm := a.cells[row][j].Conductance // Normalized [0,1]
		// Convert normalized conductance to physical units (Siemens)
		// Range: 10 µS (OFF) to 100 µS (ON), linear mapping
		gPhys := (10e-6 + gNorm*90e-6) // 10-100 µS range
		// I = G * V
		current += gPhys * input[j]
	}
	return current
}

// estimateColumnCurrent estimates current through a bit line column.
func (a *Array) estimateColumnCurrent(col int, input []float64) float64 {
	var current float64
	for i := 0; i < a.config.Rows; i++ {
		gNorm := a.cells[i][col].Conductance // Normalized [0,1]
		// Convert normalized conductance to physical units (Siemens)
		// Range: 10 µS (OFF) to 100 µS (ON), linear mapping
		gPhys := (10e-6 + gNorm*90e-6) // 10-100 µS range
		// I = G * V
		if col < len(input) {
			current += gPhys * input[col]
		}
	}
	return current
}

// SneakPathAnalysis contains sneak path current analysis results.
type SneakPathAnalysis struct {
	// Sneak current map (normalized)
	SneakCurrents [][]float64

	// Statistics
	MaxSneakRatio float64 // Maximum sneak/signal ratio
	AvgSneakRatio float64 // Average sneak/signal ratio
	TotalSneak    float64 // Total sneak current
	TotalSignal   float64 // Total signal current
}

// AnalyzeSneakPaths analyzes sneak path currents in the array.
// Sneak paths occur when unselected cells create parallel current paths.
//
// Physical model for passive crossbar when reading cell (selectedRow, selectedCol):
// - Same row: current leaks from selected WL through cell to unselected BL
// - Same column: current from unselected WLs leaks through cell to selected BL
// - Off-diagonal: three-cell sneak path forms a complete loop
//
// For architecture-aware analysis, use AnalyzeSneakPathsWithArch.
func (a *Array) AnalyzeSneakPaths(selectedRow, selectedCol int) *SneakPathAnalysis {
	return a.AnalyzeSneakPathsWithArch(selectedRow, selectedCol, false)
}

// AnalyzeSneakPathsWithArch analyzes sneak paths with architecture consideration.
// If is1T1R is true, transistor isolation reduces sneak by ~1000x.
func (a *Array) AnalyzeSneakPathsWithArch(selectedRow, selectedCol int, is1T1R bool) *SneakPathAnalysis {
	rows := a.config.Rows
	cols := a.config.Cols

	sneakMap := make([][]float64, rows)
	for i := range sneakMap {
		sneakMap[i] = make([]float64, cols)
	}

	// Calculate signal current (selected cell)
	signalG := a.cells[selectedRow][selectedCol].Conductance
	if signalG < 1e-10 {
		signalG = 1e-10 // Avoid division by zero
	}

	// Architecture-dependent isolation factor
	// 1T1R: Transistor provides ~10^6 ON/OFF ratio, use 10^-3 for conservative estimate
	// 0T1R: Full sneak path (factor = 1.0)
	isolationFactor := 1.0
	if is1T1R {
		isolationFactor = 0.001 // 1000x isolation from transistor
	}

	var totalSneak float64

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i == selectedRow && j == selectedCol {
				continue // Skip selected cell
			}

			var sneakG float64

			if i == selectedRow {
				// Same row as selected cell (but different column)
				// Sneak path: selected WL → this cell → BL j → return path
				// The cell's conductance directly contributes to sneak current
				// Higher conductance = more current leaks to this BL
				cellG := a.cells[i][j].Conductance

				// Sum return paths through all cells in column j (except this row)
				var returnPathG float64
				for k := 0; k < rows; k++ {
					if k != selectedRow {
						returnPathG += a.cells[k][j].Conductance
					}
				}

				// Series combination: this cell in series with parallel return paths
				if cellG > 0 && returnPathG > 0 {
					sneakG = (cellG * returnPathG) / (cellG + returnPathG)
				}

			} else if j == selectedCol {
				// Same column as selected cell (but different row)
				// Sneak path: WL i → cells in row i → return to this cell → selected BL
				// Current from other word lines can leak through this cell
				cellG := a.cells[i][j].Conductance

				// Sum paths from other columns in this row
				var feedPathG float64
				for k := 0; k < cols; k++ {
					if k != selectedCol {
						feedPathG += a.cells[i][k].Conductance
					}
				}

				// Series combination: feed paths in series with this cell
				if cellG > 0 && feedPathG > 0 {
					sneakG = (cellG * feedPathG) / (cellG + feedPathG)
				}

			} else {
				// Off-diagonal cell: three-cell sneak path
				// Path: selected WL → cell(sr,j) → BL j → cell(i,j) → WL i → cell(i,sc) → selected BL
				g1 := a.cells[selectedRow][j].Conductance // Entry point on selected row
				g2 := a.cells[i][j].Conductance           // This cell (corner of path)
				g3 := a.cells[i][selectedCol].Conductance // Exit point on selected column

				if g1 > 0 && g2 > 0 && g3 > 0 {
					// Three conductances in series: G_total = 1/(1/g1 + 1/g2 + 1/g3)
					sneakG = 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
				}
			}

			// Apply architecture isolation factor
			sneakG *= isolationFactor

			sneakMap[i][j] = sneakG
			totalSneak += sneakG
		}
	}

	// Calculate statistics
	var maxRatio, sumRatio float64
	count := 0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i == selectedRow && j == selectedCol {
				continue
			}
			ratio := sneakMap[i][j] / signalG
			if ratio > maxRatio {
				maxRatio = ratio
			}
			sumRatio += ratio
			count++
		}
	}

	avgRatio := 0.0
	if count > 0 {
		avgRatio = sumRatio / float64(count)
	}

	return &SneakPathAnalysis{
		SneakCurrents: sneakMap,
		MaxSneakRatio: maxRatio,
		AvgSneakRatio: avgRatio,
		TotalSneak:    totalSneak,
		TotalSignal:   signalG,
	}
}

// MVMWithIRDrop performs MVM with IR drop effects.
func (a *Array) MVMWithIRDrop(input []float64, params *WireParams) ([]float64, *IRDropAnalysis, error) {
	if len(input) > a.config.Cols {
		return nil, nil, ErrInputSize
	}

	if params == nil {
		params = DefaultWireParams()
	}

	// First, analyze IR drop
	irAnalysis := a.AnalyzeIRDrop(input, params)

	output := make([]float64, a.config.Rows)

	for i := 0; i < a.config.Rows; i++ {
		var sum float64
		for j := 0; j < len(input); j++ {
			// Apply IR drop effect to voltage
			effectiveV := input[j] * irAnalysis.EffectiveVoltage[i][j]
			quantizedInput := a.quantizeDAC(effectiveV)

			g := a.cells[i][j].Conductance * a.cells[i][j].NoiseFactor
			sum += g * quantizedInput
		}

		output[i] = a.quantizeADC(sum / float64(len(input)))
		a.totalReads++
	}

	return output, irAnalysis, nil
}

// ErrInputSize indicates input size mismatch.
var ErrInputSize error

// GetIRDropMap returns an IR drop heatmap for visualization.
// Uses a FIXED scale (0-15% drop) so architecture differences are visible.
func (a *IRDropAnalysis) GetIRDropMap() [][]float64 {
	rows := len(a.EffectiveVoltage)
	cols := len(a.EffectiveVoltage[0])

	// Fixed scale: 15% max drop for better color differentiation
	// Typical values: 0T1R ~12%, 1T1R ~8% - this range shows clear differences
	fixedMaxDrop := 0.15

	normalized := make([][]float64, rows)
	for i := range normalized {
		normalized[i] = make([]float64, cols)
		for j := range normalized[i] {
			// IR drop = 1 - effective voltage (assuming 1V input)
			drop := 1.0 - a.EffectiveVoltage[i][j]
			// Normalize to fixed scale, cap at 1.0
			normalized[i][j] = drop / fixedMaxDrop
			if normalized[i][j] > 1.0 {
				normalized[i][j] = 1.0
			}
		}
	}

	return normalized
}

// GetSneakMap returns the sneak current map normalized for visualization.
// Uses a fixed scale of 2.0 (200% sneak ratio) to allow consistent comparison
// between architectures: 0T1R (~1-2 ratio) vs 1T1R (~0.001 ratio).
func (s *SneakPathAnalysis) GetSneakMap() [][]float64 {
	rows := len(s.SneakCurrents)
	cols := len(s.SneakCurrents[0])

	// Fixed scale: 2.0 max sneak ratio for consistent architecture comparison
	// 0T1R typically shows ~1-2 ratio (significant sneak)
	// 1T1R typically shows ~0.001 ratio (transistor isolation)
	fixedMaxSneak := 2.0

	normalized := make([][]float64, rows)
	for i := range normalized {
		normalized[i] = make([]float64, cols)
		for j := range normalized[i] {
			// Normalize to fixed scale, cap at 1.0
			normalized[i][j] = s.SneakCurrents[i][j] / fixedMaxSneak
			if normalized[i][j] > 1.0 {
				normalized[i][j] = 1.0
			}
		}
	}

	return normalized
}

// ComputeError calculates the MVM output error due to non-idealities.
func ComputeError(ideal, actual []float64) float64 {
	if len(ideal) != len(actual) {
		return math.Inf(1)
	}

	var sumSqError float64
	for i := range ideal {
		diff := ideal[i] - actual[i]
		sumSqError += diff * diff
	}

	return math.Sqrt(sumSqError / float64(len(ideal)))
}
