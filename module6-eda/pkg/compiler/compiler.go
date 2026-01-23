// Compiler implementation for FeCIM array design generation.
//
// This package supports three operation modes:
//   - Storage Mode: High-density non-volatile storage (NAND replacement)
//   - Memory Mode: High-speed zero-refresh memory (DRAM replacement)
//   - Compute Mode: Analog compute-in-memory for AI inference
//
// For Storage and Memory modes, arrays are generated without weights.
// For Compute mode, weights are optional - arrays can be generated
// for later programming or pre-initialized with trained weights.
package compiler

import (
	"fmt"
	"math"
)

// Compile transforms a weight matrix into FeCIM crossbar cell assignments.
//
// Deprecated: Use GenerateDesign with NewComputeConfig instead:
//
//	config := NewComputeConfig(rows, cols)
//	config.ComputeConfig.InitialWeights = weights
//	design, err := GenerateDesign(config)
//
// The compilation algorithm:
//  1. Validates that weights fit within the configured array dimensions
//  2. Finds the symmetric weight range: [-max(|wmin|,|wmax|), +max(|wmin|,|wmax|)]
//  3. Quantizes each weight to one of 30 discrete levels using symmetric quantization
//  4. Maps each level to a physical conductance value in the configured range
//  5. Calculates programming voltage for each cell
//  6. Computes quality metrics (MSE, PSNR, utilization)
//
// Parameters:
//   - weights: 2D slice of float64 values (neural network weights)
//   - config: Compilation configuration (array size, levels, conductance range)
//
// Returns:
//   - *CrossbarMapping: Complete cell assignments with statistics
//   - error: Non-nil if weights exceed array dimensions or are empty
//
// Quantization Formula:
//
//	level = round((weight + wAbsMax) / (2 * wAbsMax) * (Levels - 1))
//
// where wAbsMax = max(|wmin|, |wmax|) ensures symmetric zero-centering.
//
// Conductance Mapping:
//
//	G = GMin + (level / (Levels - 1)) * (GMax - GMin)
//
// Example:
//
//	weights := [][]float64{{0.5, -0.3}, {0.8, -0.1}}
//	config := DefaultConfig()
//	mapping, err := Compile(weights, config)
//	// mapping.Stats.QuantPSNR indicates quantization quality (>30 dB is good)
func Compile(weights [][]float64, config CompileConfig) (*CrossbarMapping, error) {
	// Validate
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
		wAbsMax = 1.0 // Prevent division by zero
	}

	// Compile each weight
	var cells []CellAssignment
	var mseSum float64
	levelsUsed := make(map[int]bool)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := weights[i][j]

			// Quantize: map [-wAbsMax, +wAbsMax] to [0, Levels-1]
			normalized := (w + wAbsMax) / (2 * wAbsMax) // 0 to 1
			level := int(math.Round(normalized * float64(config.Levels-1)))
			level = clamp(level, 0, config.Levels-1)
			levelsUsed[level] = true

			// Dequantize to get quantized value
			qNorm := float64(level) / float64(config.Levels-1)
			qValue := -wAbsMax + qNorm*(2*wAbsMax)
			mseSum += (w - qValue) * (w - qValue)

			// Map to physical parameters
			gNorm := float64(level) / float64(config.Levels-1)
			conductance := config.GMin + gNorm*(config.GMax-config.GMin)
			progV := config.VProgMin + gNorm*(config.VProgMax-config.VProgMin)

			cells = append(cells, CellAssignment{
				// New field names
				Row:           i,
				Col:           j,
				Level:         level,
				Conductance:   conductance,
				Resistance:    1e6 / conductance,
				ProgramV:      progV,
				InitialWeight: w,
				// Legacy field names for backward compatibility
				WeightValue: w,
				QuantLevel:  level,
			})
		}
	}

	// Calculate statistics
	numCells := rows * cols
	mse := mseSum / float64(numCells)
	psnr := 100.0
	if mse > 0 {
		psnr = 10 * math.Log10((wAbsMax*wAbsMax)/mse)
	}

	totalCells := config.ArrayRows * config.ArrayCols
	areaMM2 := float64(config.ArrayRows) * config.RowHeight * float64(config.ArrayCols) * config.CellPitch / 1e6

	// Copy config and set up for compute mode with weights
	configCopy := config
	configCopy.Mode = ModeCompute
	if configCopy.ComputeConfig == nil {
		configCopy.ComputeConfig = &ComputeArrayConfig{}
	}
	configCopy.ComputeConfig.InitialWeights = weights

	return &CrossbarMapping{
		Config: &configCopy,
		Cells:  cells,
		Stats: Stats{
			// New field names
			TotalCells:  totalCells,
			ActiveCells: numCells,
			AreaMM2:     areaMM2,
			QuantMSE:    mse,
			QuantPSNR:   psnr,
			WeightMin:   wMin,
			WeightMax:   wMax,
			// Legacy field names for backward compatibility
			UsedCells:    numCells,
			Utilization:  float64(numCells) / float64(totalCells),
			UniqueLevels: len(levelsUsed),
		},
	}, nil
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

// GenerateDesign creates a complete FeCIM array design from configuration.
//
// This is the primary entry point for the new three-mode architecture.
// Unlike Compile(), this function:
//   - Works with all three operation modes (Storage, Memory, Compute)
//   - Does not require weights (weights are optional for Compute mode)
//   - Returns ArrayDesign with mode-appropriate cell configurations
//
// For Storage and Memory modes:
//   - Generates array structure without initial programming
//   - Cells are created with default conductance values
//   - Programming happens during device operation
//
// For Compute mode:
//   - If config.ComputeConfig.InitialWeights is set, compiles weights to cells
//   - If no weights provided, generates unprogrammed array structure
//
// Example (Storage - no weights needed):
//
//	config := NewStorageConfig(256, 256)
//	design, err := GenerateDesign(config)
//
// Example (Compute with weights):
//
//	config := NewComputeConfig(64, 64)
//	config.ComputeConfig.InitialWeights = trainedWeights
//	design, err := GenerateDesign(config)
//
// Example (Compute without weights - for later programming):
//
//	config := NewComputeConfig(64, 64)
//	design, err := GenerateDesign(config) // No initial weights
func GenerateDesign(config *ArrayConfig) (*ArrayDesign, error) {
	if config == nil {
		return nil, fmt.Errorf("nil configuration")
	}

	if config.ArrayRows <= 0 || config.ArrayCols <= 0 {
		return nil, fmt.Errorf("invalid array dimensions: %dx%d", config.ArrayRows, config.ArrayCols)
	}

	switch config.Mode {
	case ModeStorage:
		return generateStorageArray(config)
	case ModeMemory:
		return generateMemoryArray(config)
	case ModeCompute:
		return generateComputeArray(config)
	default:
		return nil, fmt.Errorf("unknown operation mode: %d", config.Mode)
	}
}

// generateStorageArray creates a storage-optimized FeCIM array.
// Storage arrays don't have initial weights - they're programmed during use.
func generateStorageArray(config *ArrayConfig) (*ArrayDesign, error) {
	cells := make([]CellAssignment, 0, config.ArrayRows*config.ArrayCols)

	// Generate cells with middle-level conductance (unprogrammed state)
	midLevel := config.Levels / 2
	midConductance := config.GMin + (float64(midLevel)/float64(config.Levels-1))*(config.GMax-config.GMin)

	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				Level:       midLevel,
				Conductance: midConductance,
				Resistance:  1e6 / midConductance, // R = 1/G (G in μS, R in Ω)
				ProgramV:    config.VProgMin,      // Unprogrammed
			})
		}
	}

	totalCells := config.ArrayRows * config.ArrayCols
	areaMM2 := float64(config.ArrayRows) * config.RowHeight * float64(config.ArrayCols) * config.CellPitch / 1e6

	return &ArrayDesign{
		Config: config,
		Cells:  cells,
		Stats: DesignStats{
			TotalCells:  totalCells,
			ActiveCells: totalCells,
			AreaMM2:     areaMM2,
			PowerMW:     estimateStoragePower(config),
		},
	}, nil
}

// generateMemoryArray creates a memory-optimized FeCIM array.
// Memory arrays don't have initial weights - they're programmed during use.
func generateMemoryArray(config *ArrayConfig) (*ArrayDesign, error) {
	cells := make([]CellAssignment, 0, config.ArrayRows*config.ArrayCols)

	// Generate cells with reset state (low conductance)
	resetLevel := 0
	resetConductance := config.GMin

	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				Level:       resetLevel,
				Conductance: resetConductance,
				Resistance:  1e6 / resetConductance,
				ProgramV:    config.VProgMin,
			})
		}
	}

	totalCells := config.ArrayRows * config.ArrayCols
	areaMM2 := float64(config.ArrayRows) * config.RowHeight * float64(config.ArrayCols) * config.CellPitch / 1e6

	// Estimate bandwidth based on configuration
	var bandwidthGBps float64
	if config.MemoryConfig != nil {
		bandwidthGBps = config.MemoryConfig.BandwidthGBps
	}

	return &ArrayDesign{
		Config: config,
		Cells:  cells,
		Stats: DesignStats{
			TotalCells:  totalCells,
			ActiveCells: totalCells,
			AreaMM2:     areaMM2,
			PowerMW:     estimateMemoryPower(config),
			ThroughputGOPS: bandwidthGBps * 8, // Approximate GOPS from bandwidth
		},
	}, nil
}

// generateComputeArray creates a compute-optimized FeCIM array.
// If InitialWeights are provided in ComputeConfig, they are compiled to cells.
// Otherwise, an unprogrammed array structure is generated.
func generateComputeArray(config *ArrayConfig) (*ArrayDesign, error) {
	// Check if weights are provided
	if config.ComputeConfig != nil && config.ComputeConfig.InitialWeights != nil {
		return compileWeightsToDesign(config)
	}

	// No weights - generate unprogrammed compute array
	cells := make([]CellAssignment, 0, config.ArrayRows*config.ArrayCols)

	// Initialize with zero-weight equivalent (middle conductance for signed weights)
	midLevel := config.Levels / 2
	midConductance := config.GMin + (float64(midLevel)/float64(config.Levels-1))*(config.GMax-config.GMin)

	for i := 0; i < config.ArrayRows; i++ {
		for j := 0; j < config.ArrayCols; j++ {
			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				Level:       midLevel,
				Conductance: midConductance,
				Resistance:  1e6 / midConductance,
				ProgramV:    config.VProgMin,
			})
		}
	}

	totalCells := config.ArrayRows * config.ArrayCols
	areaMM2 := float64(config.ArrayRows) * config.RowHeight * float64(config.ArrayCols) * config.CellPitch / 1e6

	return &ArrayDesign{
		Config: config,
		Cells:  cells,
		Stats: DesignStats{
			TotalCells:     totalCells,
			ActiveCells:    totalCells,
			AreaMM2:        areaMM2,
			PowerMW:        estimateComputePower(config),
			ThroughputGOPS: estimateThroughput(config),
		},
	}, nil
}

// compileWeightsToDesign compiles neural network weights to cell assignments.
// This is the internal implementation used when ComputeConfig.InitialWeights is set.
func compileWeightsToDesign(config *ArrayConfig) (*ArrayDesign, error) {
	weights := config.ComputeConfig.InitialWeights

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
		wAbsMax = 1.0 // Prevent division by zero for all-zero weights
	}

	// Compile each weight
	cells := make([]CellAssignment, 0, rows*cols)
	var mseSum float64

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := weights[i][j]

			// Quantize: map [-wAbsMax, +wAbsMax] to [0, Levels-1]
			normalized := (w + wAbsMax) / (2 * wAbsMax)
			level := int(math.Round(normalized * float64(config.Levels-1)))
			level = clamp(level, 0, config.Levels-1)

			// Dequantize to get quantized value for MSE calculation
			qNorm := float64(level) / float64(config.Levels-1)
			qValue := -wAbsMax + qNorm*(2*wAbsMax)
			mseSum += (w - qValue) * (w - qValue)

			// Map to physical parameters
			gNorm := float64(level) / float64(config.Levels-1)
			conductance := config.GMin + gNorm*(config.GMax-config.GMin)
			progV := config.VProgMin + gNorm*(config.VProgMax-config.VProgMin)

			cells = append(cells, CellAssignment{
				Row:           i,
				Col:           j,
				Level:         level,
				Conductance:   conductance,
				Resistance:    1e6 / conductance,
				ProgramV:      progV,
				InitialWeight: w,
			})
		}
	}

	// Calculate statistics
	numCells := rows * cols
	mse := mseSum / float64(numCells)
	psnr := 100.0
	if mse > 0 {
		psnr = 10 * math.Log10((wAbsMax*wAbsMax)/mse)
	}

	totalCells := config.ArrayRows * config.ArrayCols
	areaMM2 := float64(config.ArrayRows) * config.RowHeight * float64(config.ArrayCols) * config.CellPitch / 1e6

	return &ArrayDesign{
		Config: config,
		Cells:  cells,
		Stats: DesignStats{
			TotalCells:     totalCells,
			ActiveCells:    numCells,
			AreaMM2:        areaMM2,
			PowerMW:        estimateComputePower(config),
			ThroughputGOPS: estimateThroughput(config),
			QuantMSE:       mse,
			QuantPSNR:      psnr,
			WeightMin:      wMin,
			WeightMax:      wMax,
		},
	}, nil
}

// Power estimation helpers
func estimateStoragePower(config *ArrayConfig) float64 {
	// Rough estimate: 0.1 mW per 1000 cells for storage
	return float64(config.ArrayRows*config.ArrayCols) * 0.0001
}

func estimateMemoryPower(config *ArrayConfig) float64 {
	// Rough estimate: 0.5 mW per 1000 cells for high-speed memory
	return float64(config.ArrayRows*config.ArrayCols) * 0.0005
}

func estimateComputePower(config *ArrayConfig) float64 {
	// Rough estimate: 1.0 mW per 1000 cells for active compute
	return float64(config.ArrayRows*config.ArrayCols) * 0.001
}

func estimateThroughput(config *ArrayConfig) float64 {
	// GOPS estimate based on array size and clock frequency
	// Each cycle: rows × cols MACs
	clockMHz := config.Peripherals.ClockFreq
	if clockMHz == 0 {
		clockMHz = 100 // Default
	}
	macs := float64(config.ArrayRows * config.ArrayCols)
	return macs * clockMHz / 1000 // GOPS
}
