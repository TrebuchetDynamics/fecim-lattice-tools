package external_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
)

type wireResistance struct {
	Wordline float64 `json:"wordline"`
	Bitline  float64 `json:"bitline"`
}

type crossvalInput struct {
	Weights        [][]float64    `json:"weights"`
	InputVector    []float64      `json:"input_vector"`
	ArraySize      [2]int         `json:"array_size"`
	WireResistance wireResistance `json:"wire_resistance"`
}

type crossvalResult struct {
	IdealOutput          []float64 `json:"ideal_output"`
	IRDropOutput         []float64 `json:"ir_drop_output,omitempty"`
	BadcrossbarAvailable bool      `json:"badcrossbar_available"`
}

func runCrossvalScript(t *testing.T, input crossvalInput) crossvalResult {
	t.Helper()

	payload, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("marshal cross-validation input: %v", err)
	}

	const script = `
import json
import sys
import numpy as np

data = json.load(sys.stdin)
weights = np.array(data["weights"], dtype=float)
input_vector = np.array(data["input_vector"], dtype=float)
ideal = (weights @ input_vector).tolist()

wire = data.get("wire_resistance", {})
wordline = float(wire.get("wordline", 0.0))
bitline = float(wire.get("bitline", 0.0))
effective_r = max(wordline, 0.0) + max(bitline, 0.0)
ir_drop_output = None
if effective_r > 0.0:
    scale = 1.0 / (1.0 + 0.01 * effective_r)
    ir_drop_output = (weights @ input_vector * scale).tolist()

print(json.dumps({
    "ideal_output": ideal,
    "ir_drop_output": ir_drop_output,
    "badcrossbar_available": False
}))
`

	cmd := exec.Command("python3", "-c", script)
	cmd.Stdin = bytes.NewReader(payload)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run numpy oracle: %v\n%s", err, output)
	}

	var result crossvalResult
	if err := json.Unmarshal(output, &result); err != nil {
		t.Fatalf("decode numpy oracle output: %v\n%s", err, output)
	}
	return result
}

func goMVMRaw(weights [][]float64, input []float64) []float64 {
	output := make([]float64, len(weights))
	for row := range weights {
		sum := 0.0
		for col, weight := range weights[row] {
			if col >= len(input) {
				continue
			}
			sum += crossbar.QuantizeToLevels(weight) * input[col]
		}
		output[row] = sum
	}
	return output
}
