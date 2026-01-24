// pkg/compiler/compiler.go
package compiler

import (
	"fmt"
	"math"
)

// Compile transforms a weight matrix into crossbar cell assignments
// If weights is nil, generates a blank array based on Config dimensions
func Compile(weights [][]float64, config CompileConfig) (*CrossbarMapping, error) {
	if weights == nil {
		return GenerateBlank(config), nil
	}

	// Validate weights dimensions
	if len(weights) == 0 || len(weights[0]) == 0 {
		return nil, fmt.Errorf("empty weight matrix")
	}

	rows := len(weights)
	cols := len(weights[0])

	if rows > config.ArrayRows || cols > config.ArrayCols {
		return nil, fmt.Errorf("weights %dx%d exceed array %dx%d",
			rows, cols, config.ArrayRows, config.ArrayCols)
	}

	// Find weight range for symmetric quantization
	wMin, wMax := weights[0][0], weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < wMin {
				wMin = weights[i][j]
			}
			if weights[i][j] > wMax {
				wMax = weights[i][j]
			}
		}
	}
	wAbsMax := math.Max(math.Abs(wMin), math.Abs(wMax))
	if wAbsMax == 0 {
		wAbsMax = 1.0 // Avoid divide by zero
	}

	// Compile each weight
	var cells []CellAssignment
	var mseSum float64
	levelsUsed := make(map[int]bool)

	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			// Default values for unused cells
			var w float64 = 0.0
			var level int = 0
			var conductance float64 = config.GMin
			var progV float64 = config.VProgMin

			// If within weight matrix bounds
			if i < rows && j < cols {
				w = weights[i][j]

				// Quantize: map [-wAbsMax, +wAbsMax] to [0, Levels-1]
				normalized := (w + wAbsMax) / (2 * wAbsMax)
				level = int(math.Round(normalized * float64(config.Levels-1)))
				level = clamp(level, 0, config.Levels-1)
				levelsUsed[level] = true

				// Dequantize for stats
				qNorm := float64(level) / float64(config.Levels-1)
				qValue := -wAbsMax + qNorm*(2*wAbsMax)
				mseSum += (w - qValue) * (w - qValue)

				// specific mapping
				conductance = config.GMin + qNorm*(config.GMax-config.GMin)
				progV = config.VProgMin + qNorm*(config.VProgMax-config.VProgMin)
			}

			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				WeightValue: w,
				QuantLevel:  level,
				Conductance: conductance,
				ProgramV:    progV,
			})
		}
	}

	// Calculate statistics
	numWeights := rows * cols
	mse := mseSum / float64(numWeights)
	psnr := 100.0
	if mse > 1e-9 {
		psnr = 10 * math.Log10((wAbsMax*wAbsMax)/mse)
	}

	return &CrossbarMapping{
		Config: config,
		Cells:  cells,
		Stats: Stats{
			TotalCells:   config.ArrayRows * config.ArrayCols,
			UsedCells:    numWeights,
			Utilization:  float64(numWeights) / float64(config.ArrayRows*config.ArrayCols),
			WeightMin:    wMin,
			WeightMax:    wMax,
			QuantMSE:     mse,
			QuantPSNR:    psnr,
			UniqueLevels: len(levelsUsed),
		},
	}, nil
}

// GenerateBlank creates an initialized array without weights
func GenerateBlank(config CompileConfig) *CrossbarMapping {
	var cells []CellAssignment
	
	// Pre-allocate for performance
	cells = make([]CellAssignment, 0, config.ArrayRows*config.ArrayCols)

	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			// Initialize to GMin / Level 0 (Reset State)
			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				WeightValue: 0.0,
				QuantLevel:  0,
				Conductance: config.GMin,
				ProgramV:    config.VProgMin,
			})
		}
	}

	return &CrossbarMapping{
		Config: config,
		Cells:  cells,
		Stats: Stats{
			TotalCells:   config.ArrayRows * config.ArrayCols,
			UsedCells:    0,
			Utilization:  0.0,
			WeightMin:    0.0,
			WeightMax:    0.0,
			QuantMSE:     0.0,
			QuantPSNR:    0.0,
			UniqueLevels: 1, // Only level 0 used
		},
	}
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
