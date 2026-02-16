package mnistcli

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"fecim-lattice-tools/module3-mnist/pkg/core"
)

func runExportQuantizedWeights(levels []int, loadFile string, outDirs []string, hiddenSize int) error {
	fmt.Println("\n=== Export Quantized Weights ===")

	if len(levels) == 0 {
		return fmt.Errorf("no levels specified")
	}

	if len(outDirs) == 0 {
		outDirs = []string{"module3-mnist/data"}
	}
	for _, outDir := range outDirs {
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return fmt.Errorf("create output dir %s: %w", outDir, err)
		}
	}

	net := core.NewDualModeNetwork(784, hiddenSize, 10)

	weightsPath, err := resolveWeightsPath(loadFile)
	if err != nil {
		return err
	}
	baseBytes, err := os.ReadFile(weightsPath)
	if err != nil {
		return fmt.Errorf("read base weights: %w", err)
	}
	if err := net.LoadWeights(weightsPath); err != nil {
		return fmt.Errorf("load weights: %w", err)
	}

	for _, outDir := range outDirs {
		baseOut := filepath.Join(outDir, "pretrained_weights.json")
		if err := os.WriteFile(baseOut, baseBytes, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", baseOut, err)
		}
		fmt.Printf("Wrote %s\n", baseOut)
	}

	w1Min, w1Max := weightRange(net.FPWeights1)
	w2Min, w2Max := weightRange(net.FPWeights2)

	for _, level := range levels {
		if level < 2 {
			return fmt.Errorf("levels must be >= 2, got %d", level)
		}

		qW1, l1Scale, l1Offset := quantizeNormalized(net.FPWeights1, w1Min, w1Max, level)
		qW2, l2Scale, l2Offset := quantizeNormalized(net.FPWeights2, w2Min, w2Max, level)

		data := core.WeightsFile{
			Layer1Weights:     qW1,
			Layer2Weights:     qW2,
			Biases1:           net.FPBias1,
			Biases2:           net.FPBias2,
			L1Scale:           l1Scale,
			L1Offset:          l1Offset,
			L2Scale:           l2Scale,
			L2Offset:          l2Offset,
			QuantLevels:       level,
			Layer1QuantLevels: level,
			Layer2QuantLevels: level,
		}

		jsonData, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal weights for level %d: %w", level, err)
		}

		for _, outDir := range outDirs {
			outPath := filepath.Join(outDir, fmt.Sprintf("pretrained_weights_%d.json", level))
			if err := os.WriteFile(outPath, jsonData, 0o644); err != nil {
				return fmt.Errorf("write %s: %w", outPath, err)
			}
			fmt.Printf("Wrote %s\n", outPath)
		}
	}

	fmt.Println("Export complete.")
	return nil
}

func weightRange(weights [][]float64) (float64, float64) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return 0, 0
	}
	minVal := weights[0][0]
	maxVal := weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < minVal {
				minVal = weights[i][j]
			}
			if weights[i][j] > maxVal {
				maxVal = weights[i][j]
			}
		}
	}
	return minVal, maxVal
}

func quantizeNormalized(weights [][]float64, minVal, maxVal float64, levels int) ([][]float64, float64, float64) {
	scale := maxVal - minVal
	offset := minVal
	rows := len(weights)
	if rows == 0 {
		return weights, 0, 0
	}
	cols := len(weights[0])
	q := make([][]float64, rows)

	if scale == 0 {
		scale = 1
		for i := range weights {
			q[i] = make([]float64, cols)
		}
		return q, scale, offset
	}

	for i := range weights {
		q[i] = make([]float64, cols)
		for j := range weights[i] {
			norm := (weights[i][j] - minVal) / scale
			if norm < 0 {
				norm = 0
			}
			if norm > 1 {
				norm = 1
			}
			if levels > 1 {
				q[i][j] = math.Round(norm*float64(levels-1)) / float64(levels-1)
			} else {
				q[i][j] = norm
			}
		}
	}
	return q, scale, offset
}

// parseLevelList and parseDirList moved to utils.go
