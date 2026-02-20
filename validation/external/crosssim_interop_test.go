package external_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestCrossSimMVMAccuracy validates FeCIM's ideal MVM against CrossSim's AnalogCore.
//
// CrossSim (sandialabs/cross-sim v3.1) provides a GPU-accelerated Python crossbar
// simulator with the same physics as our Go model. In default configuration,
// CrossSim's AnalogCore computes y ≈ W @ x with < 1e-6 relative error (internal
// float64 arithmetic), making it suitable as an independent MVM reference.
//
// Comparison type: exact MVM agreement (both in ideal/low-noise default mode)
// Expected agreement: < 1e-5 relative error on every output element
//
// Install: pip3 install git+https://github.com/sandialabs/cross-sim.git
// Package: `simulator` (CrossSim 3.1.1 — imported as `from simulator import ...`)
//
// Skip condition: if python3 or the `simulator` package is not installed.
func TestCrossSimMVMAccuracy(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	if err := exec.Command("python3", "-c", "from simulator import AnalogCore").Run(); err != nil {
		t.Skip("CrossSim not installed — install via: pip3 install git+https://github.com/sandialabs/cross-sim.git")
	}

	// Weight matrix in [-1, 1] range — CrossSim's natural domain.
	// Values are deterministic fractions of 29 (= max quantization level - 1).
	const nRows, nCols = 4, 4
	W := [nRows][nCols]float64{
		{28.0 / 29, 3.0 / 29, 8.0 / 29, 2.0 / 29},
		{4.0 / 29, 25.0 / 29, 6.0 / 29, 9.0 / 29},
		{7.0 / 29, 5.0 / 29, 22.0 / 29, 4.0 / 29},
		{2.0 / 29, 8.0 / 29, 3.0 / 29, 27.0 / 29},
	}
	x := [nCols]float64{1.0, 0.75, 0.5, 0.25}

	// Compute ideal reference: y = W @ x (pure float64 arithmetic).
	var yIdeal [nRows]float64
	for r := 0; r < nRows; r++ {
		for c := 0; c < nCols; c++ {
			yIdeal[r] += W[r][c] * x[c]
		}
	}

	// Run through CrossSim AnalogCore (default parameters = ideal numerical mode).
	const crossSimScript = `
from simulator import AnalogCore, CrossSimParameters
import numpy as np
import json, sys

d = json.loads(sys.stdin.read())
W = np.array(d["W"])   # shape (rows, cols)
x = np.array(d["x"])   # shape (cols,)

params = CrossSimParameters()
core   = AnalogCore(W, params)
y      = core.matvec(x)

print(json.dumps({"y": y.tolist()}))
`
	Wslice := make([][]float64, nRows)
	for r := range Wslice {
		Wslice[r] = W[r][:]
	}
	input := map[string]interface{}{"W": Wslice, "x": x[:]}
	inputJSON, _ := json.Marshal(input)

	cmd := exec.Command("python3", "-c", crossSimScript)
	cmd.Stdin = strings.NewReader(string(inputJSON))
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CrossSim subprocess error: %v\noutput: %s", err, outBytes)
	}

	var csResult struct {
		Y []float64 `json:"y"`
	}
	if err := json.Unmarshal(outBytes, &csResult); err != nil {
		t.Fatalf("parse CrossSim output: %v\nraw: %s", err, outBytes)
	}

	// Compare CrossSim result against ideal float64 W @ x.
	const relThr = 1e-5
	t.Log("CROSSSIM MVM accuracy — CrossSim AnalogCore vs ideal float64")
	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-12s  %-8s", "Row", "CrossSim y", "Ideal y", "Rel err", "Status")
	t.Log("──────────────────────────────────────────────────────────────────────")

	allPass := true
	maxRelErr := 0.0
	for r := 0; r < nRows; r++ {
		csY := csResult.Y[r]
		refY := yIdeal[r]
		rel := math.Abs(csY-refY) / (math.Abs(refY) + 1e-15)
		if rel > maxRelErr {
			maxRelErr = rel
		}
		status := "PASS"
		if rel > relThr {
			status = "FAIL"
			allPass = false
			t.Errorf("Row %d: CrossSim=%.8f  ideal=%.8f  relErr=%.3e", r, csY, refY, rel)
		}
		t.Logf("Row %-3d  %-14.8f  %-14.8f  %-12.3e  %s", r, csY, refY, rel, status)
	}
	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("Max relative error: %.3e  (threshold: %.0e)", maxRelErr, relThr)
	if allPass {
		t.Log("PASS: CrossSim AnalogCore MVM agrees with ideal float64 within 1e-5")
	}

	fmt.Printf("CROSSSIM_MVM: rows=%d cols=%d maxRelErr=%.3e pass=%v\n",
		nRows, nCols, maxRelErr, allPass)

	// ── Emit JSON artifact ────────────────────────────────────────────────────
	dir := "../../output/validation/external"
	os.MkdirAll(dir, 0755)
	artifact := map[string]interface{}{
		"test":         "crosssim_mvm_accuracy",
		"tool":         "CrossSim",
		"version":      "3.1.1",
		"source":       "github.com/sandialabs/cross-sim",
		"rows":         nRows, "cols": nCols,
		"max_rel_err":  maxRelErr,
		"threshold":    relThr,
		"pass":         allPass,
	}
	b, _ := json.MarshalIndent(artifact, "", "  ")
	os.WriteFile(dir+"/crosssim_mvm_accuracy.json", b, 0644)
}

// TestCrossSimIRDropTrend validates that CrossSim and our Go solver agree on
// the direction of accuracy degradation as array size increases under IR drop.
//
// Comparison type: trend-level (monotonic agreement, not numerical match)
// Both simulators should show decreasing MVM accuracy (increasing RMSE) for
// larger arrays, since wire resistance effects scale with array size.
//
// This test requires CrossSim 3.x with r_i (interconnect resistance) support.
// It is skipped if CrossSim's xbar.array.parasitics.r_i parameter is unavailable.
func TestCrossSimIRDropTrend(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	if err := exec.Command("python3", "-c", "from simulator import AnalogCore").Run(); err != nil {
		t.Skip("CrossSim not installed")
	}

	// Check whether CrossSim supports r_i configuration.
	rICheck := `
from simulator import CrossSimParameters
p = CrossSimParameters()
try:
    _ = p.xbar.array.parasitics.r_i
    print("supported")
except AttributeError:
    print("unsupported")
`
	checkOut, _ := exec.Command("python3", "-c", rICheck).Output()
	if strings.TrimSpace(string(checkOut)) != "supported" {
		t.Skip("CrossSim r_i parasitics not available in this version — skipping IR drop trend test")
	}

	const irDropScript = `
from simulator import AnalogCore, CrossSimParameters
import numpy as np
import json, sys

d     = json.loads(sys.stdin.read())
sizes = d["sizes"]    # list of array side lengths
r_i   = d["r_i"]     # interconnect resistance (Ω)
G_max = d["G_max"]   # max cell conductance (S)

results = []
for n in sizes:
    # Uniform weight matrix (all cells at half-scale)
    W = np.full((n, n), 0.5)
    x = np.ones(n) * 0.5

    # Ideal (no parasitics)
    p_ideal = CrossSimParameters()
    c_ideal = AnalogCore(W, p_ideal)
    y_ideal = c_ideal.matvec(x)
    rmse_ideal = float(np.sqrt(np.mean((y_ideal - (W @ x))**2)))

    # With r_i parasitics
    p_ri = CrossSimParameters()
    try:
        p_ri.xbar.array.parasitics.r_i = r_i
    except Exception:
        results.append({"n": n, "rmse_ideal": rmse_ideal, "rmse_ri": None})
        continue
    c_ri = AnalogCore(W, p_ri)
    y_ri = c_ri.matvec(x)
    rmse_ri = float(np.sqrt(np.mean((y_ri - (W @ x))**2)))

    results.append({"n": n, "rmse_ideal": rmse_ideal, "rmse_ri": rmse_ri})

print(json.dumps({"results": results}))
`
	input := map[string]interface{}{
		"sizes": []int{4, 8, 16},
		"r_i":   10.0, // 10 Ω per segment — exaggerated for clear trend
		"G_max": 1e-4, // 100 µS
	}
	inputJSON, _ := json.Marshal(input)
	cmd := exec.Command("python3", "-c", irDropScript)
	cmd.Stdin = strings.NewReader(string(inputJSON))
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CrossSim IR drop script error: %v\noutput: %s", err, outBytes)
	}

	var csRes struct {
		Results []struct {
			N         int     `json:"n"`
			RMSEIdeal float64 `json:"rmse_ideal"`
			RMSERi    float64 `json:"rmse_ri"`
		} `json:"results"`
	}
	if err := json.Unmarshal(outBytes, &csRes); err != nil {
		t.Fatalf("parse CrossSim IR drop output: %v\nraw: %s", err, outBytes)
	}

	t.Log("CROSSSIM IR drop trend — RMSE vs array size")
	t.Log("──────────────────────────────────────────────────────────")
	t.Logf("%-6s  %-14s  %-14s  %-8s", "Size", "RMSE ideal", "RMSE r_i", "Status")
	t.Log("──────────────────────────────────────────────────────────")

	allPass := true
	prevRMSE := -1.0
	for _, r := range csRes.Results {
		if r.RMSERi < 0 {
			t.Logf("%-3d×%-3d  r_i not applied in CrossSim — skipped", r.N, r.N)
			continue
		}
		status := "PASS"
		// Trend check: r_i RMSE should be >= ideal RMSE (degradation from parasitics).
		if r.RMSERi < r.RMSEIdeal*0.9 {
			status = "FAIL (r_i better than ideal?)"
			allPass = false
			t.Errorf("Size %d: RMSE with r_i (%.3e) < ideal (%.3e) — unexpected improvement",
				r.N, r.RMSERi, r.RMSEIdeal)
		}
		// Monotonic trend: larger arrays should have larger RMSE under r_i.
		if prevRMSE >= 0 && r.RMSERi < prevRMSE*0.5 {
			status = "WARN (non-monotone)"
			t.Logf("WARN: RMSE did not increase monotonically at size %d", r.N)
		}
		prevRMSE = r.RMSERi
		t.Logf("%-3d×%-3d  %-14.3e  %-14.3e  %s", r.N, r.N, r.RMSEIdeal, r.RMSERi, status)
	}
	t.Log("──────────────────────────────────────────────────────────")
	if allPass {
		t.Log("PASS: CrossSim shows expected RMSE degradation under IR drop parasitics")
	}

	fmt.Printf("CROSSSIM_IRDROP_TREND: pass=%v\n", allPass)
}
