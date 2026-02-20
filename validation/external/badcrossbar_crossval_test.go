package external_test

// badcrossbar_crossval_test.go
//
// Two-tier cross-validation connecting badcrossbar (Python passive crossbar solver)
// to the FeCIM Go nodal solver.
//
// ── TIER 1: exact match (badcrossbar vs scipy) ───────────────────────────────
// Compare badcrossbar.compute() against scipy.linalg.solve on the IDENTICAL MNA
// matrix. The scipy script stamps exactly the same system as badcrossbar (verified
// by inspecting badcrossbar/computing/kcl.py). Residuals must be < 1 pA — limited
// only by float64 precision (~1e-16 × cond(A)).
//
// ── TIER 2: Go solver cross-validation ────────────────────────────────────────
// Compare Go TierASolver (referenceSolveDense) against badcrossbar for the same
// passive crossbar. The Go solver is configured with BoundaryParams that exactly
// mirror badcrossbar's circuit:
//   - WLDriveResistance = r_wl  (source at c=0 through 1 segment)
//   - BLTerminationResistance = r_bl, BLTerminationVoltage = 0  (ground at r=rows-1)
//   - BLDriveResistance = 1e9 Ω  (BL top r=0 is effectively open; leakage < 1 pA)
// Expected agreement: < 10 nA (limited by ~2e-7 relative error from gBLDrive=1e-9S
// in a system with max conductance ~1.27 S → cond ≈ 1e9 → eps×cond ≈ 2e-7).
//
// ── badcrossbar circuit topology (from kcl.py inspection) ────────────────────
//   WL row r:  source V[r] through R_wl at c=0 (left); open at c=cols-1 (right)
//   BL col c:  open at r=0 (top); grounded through R_bl at r=rows-1 (bottom)
//   Wire:      one R_wl per WL segment; one R_bl per BL segment
//
// Literature reference:
//   Joksas & Mehonic, SoftwareX 14 (2020) 100617 — doi:10.1016/j.softx.2020.100617
//
// Install: pip3 install badcrossbar scipy
// Skip condition: python3, badcrossbar, or scipy not installed.

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
)

// badcrossbarDualScript runs both badcrossbar AND scipy with identical MNA stamping.
// The scipy stamping is derived from badcrossbar/computing/kcl.py:
//
//	WL c=0 node: diagonal = 2*gWL + Gcell  (gWL source + gWL wire to c=1)
//	WL c=cols-1: diagonal =   gWL + Gcell  (gWL wire to c-2 only, open right end)
//	BL r=0 node: diagonal =   gBL + Gcell  (gBL wire to r=1 only, open top)
//	BL r=rows-1: diagonal = 2*gBL + Gcell  (gBL wire to r-2 + gBL ground at bottom)
const badcrossbarDualScript = `
import numpy as np
import badcrossbar
from scipy import linalg
import logging, json, sys

logging.disable(logging.CRITICAL)

d    = json.loads(sys.stdin.read())
rows = d["rows"]; cols = d["cols"]
V_in = np.array(d["V_in"])     # (rows,)
R    = np.array(d["R"])         # (rows, cols)
G    = 1.0 / R
r_wl = d["r_wl"]; gWL = 1.0 / r_wl
r_bl = d["r_bl"]; gBL = 1.0 / r_bl

# ── badcrossbar ──────────────────────────────────────────────────────────────
bc = badcrossbar.compute(
    V_in.reshape(-1, 1), R,
    r_i_word_line=r_wl,
    r_i_bit_line=r_bl,
)

# ── scipy — stamping mirrors badcrossbar/computing/kcl.py ────────────────────
# Unknowns: x = [WL_nodes (rows*cols) | BL_nodes (rows*cols)]
n = 2 * rows * cols
A = np.zeros((n, n))
b = np.zeros(n)

def iW(r, c): return r * cols + c
def iB(r, c): return rows * cols + r * cols + c

def sc(i, j, g):
    A[i,i] += g; A[j,j] += g; A[i,j] -= g; A[j,i] -= g

def ss(i, g, v):
    A[i,i] += g; b[i] += g * v

for r in range(rows):
    for c in range(cols):
        w  = iW(r, c)
        bl = iB(r, c)

        # WL wire + source
        if c > 0:
            sc(w, iW(r, c-1), gWL)   # wire to left node
        if c == 0:
            ss(w, gWL, float(V_in[r]))  # source through 1 segment (matches badcrossbar)
        # no termination at c=cols-1 (open right end)

        # BL wire + ground
        if r > 0:
            sc(bl, iB(r-1, c), gBL)   # wire to node above
        if r == rows - 1:
            ss(bl, gBL, 0.0)           # ground through 1 segment (matches badcrossbar)
        # no source at r=0 (open top)

        # Cell conductance
        sc(w, bl, float(G[r, c]))

x = linalg.solve(A, b)

sp_device = [[0.0]*cols for _ in range(rows)]
sp_output  = [0.0] * cols
for r in range(rows):
    for c in range(cols):
        vw = x[iW(r,c)]; vb = x[iB(r,c)]
        i  = float(G[r,c]) * (vw - vb)
        sp_device[r][c] = i
        sp_output[c] += i

print(json.dumps({
    "bc_device_I": bc.currents.device.tolist(),  # (rows, cols)
    "bc_output_I": bc.currents.output.tolist(),  # (1, cols)
    "sp_device_I": sp_device,                    # (rows, cols)
    "sp_output_I": sp_output,                    # (cols,)
}))
`

func TestBadcrossbarCrossValidation(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 not installed")
	}
	if err := exec.Command("python3", "-c", "import badcrossbar").Run(); err != nil {
		t.Skip("badcrossbar not installed — run: pip3 install badcrossbar")
	}
	if err := exec.Command("python3", "-c", "import scipy.linalg").Run(); err != nil {
		t.Skip("scipy not installed")
	}

	// 5×5 passive crossbar — graded conductances, realistic SKY130 geometry.
	const rows, cols = 5, 5
	const Gbase = 20e-6 // 20 µS baseline

	G := make([][]float64, rows)
	R := make([][]float64, rows)
	for r := range G {
		G[r] = make([]float64, cols)
		R[r] = make([]float64, cols)
		for c := range G[r] {
			g := Gbase * (1.0 + 0.1*float64(r*cols+c)/float64(rows*cols))
			G[r][c] = g
			R[r][c] = 1.0 / g
		}
	}
	VWL := []float64{1.0, 0.8, 0.6, 0.4, 0.2}

	// SKY130-scale wire parameters (same as other external validation tests).
	geom := arraysim.DefaultCellGeometry()
	xsec := geom.WireWidth * geom.WireThickness
	rPerM := geom.MetalResistivity / xsec
	rWL := rPerM * geom.PitchX
	rBL := rPerM * geom.PitchY
	t.Logf("Wire: r_wl=%.4f Ω  r_bl=%.4f Ω  G_cell_base=%.2e S", rWL, rBL, Gbase)

	// ── Python subprocess: run both badcrossbar and scipy ────────────────────
	input := map[string]interface{}{
		"rows": rows, "cols": cols,
		"V_in": VWL, "R": R,
		"r_wl": rWL, "r_bl": rBL,
	}
	inputJSON, _ := json.Marshal(input)
	cmd := exec.Command("python3", "-c", badcrossbarDualScript)
	cmd.Stdin = strings.NewReader(string(inputJSON))
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Python subprocess error: %v\noutput: %s", err, outBytes)
	}
	var py struct {
		BCDeviceI [][]float64 `json:"bc_device_I"`
		BCOutputI [][]float64 `json:"bc_output_I"`
		SpDeviceI [][]float64 `json:"sp_device_I"`
		SpOutputI []float64   `json:"sp_output_I"`
	}
	if err := json.Unmarshal(outBytes, &py); err != nil {
		t.Fatalf("parse Python output: %v\nraw: %s", err, outBytes)
	}

	// ── TIER 1: badcrossbar vs scipy (identical MNA stamping) ────────────────
	t.Log("TIER 1 — badcrossbar vs scipy (identical MNA, float64 precision limit)")
	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("%-8s  %-14s  %-14s  %-14s  %-8s", "Cell", "BC I (µA)", "Sp I (µA)", "Δ (fA)", "Status")
	t.Log("──────────────────────────────────────────────────────────────────────")

	// For this 5×5 system: cond(A) ≈ gWL/G_cell ≈ 1.27/20e-6 ≈ 6.3e4
	// Float64 backward error ≈ eps × cond(A) ≈ 2.2e-16 × 6.3e4 ≈ 1.4e-11
	// Current error ≈ I × 1.4e-11 ≈ 20µA × 1.4e-11 ≈ 0.3 fA → use 10 pA threshold
	const tier1Thr = 10e-12 // 10 pA — generous float64 precision
	maxTier1 := 0.0
	tier1Pass := true

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			bcI := py.BCDeviceI[r][c]
			spI := py.SpDeviceI[r][c]
			diff := math.Abs(bcI - spI)
			if diff > maxTier1 {
				maxTier1 = diff
			}
			status := "PASS"
			if diff > tier1Thr {
				status = "FAIL"
				tier1Pass = false
				t.Errorf("Tier1 Device[%d][%d]: BC=%.9e  scipy=%.9e  Δ=%.3e A", r, c, bcI, spI, diff)
			}
			t.Logf("(%d,%d)    %-14.6f  %-14.6f  %-14.6f  %s",
				r, c, bcI*1e6, spI*1e6, diff*1e15, status)
		}
	}
	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("Tier 1 max error: %.3e A  (threshold: %.0e A)", maxTier1, tier1Thr)
	if tier1Pass {
		t.Log("PASS Tier 1: badcrossbar == scipy on identical MNA (float64 precision)")
	}

	// ── TIER 2: Go solver vs badcrossbar (same circuit, BLDriveR=1GΩ residual) ─
	//
	// BoundaryParams configured to exactly mirror badcrossbar's circuit:
	//   WLDriveResistance  = r_wl          → WL source through 1 segment (exact match)
	//   BLTerminationResistance = r_bl, V=0 → BL grounded through 1 segment (exact match)
	//   BLDriveResistance  = 1e9 Ω         → BL top effectively open (gBLDrive=1e-9 S)
	//
	// Residual from BLDriveResistance=1e9:
	//   gBLDrive=1e-9 S adds to BL-top diagonal where gBL=0.21 S → fraction 5e-9
	//   Voltage error at BL top: ~1mV × 5e-9 = 5 pV
	//   Current error per cell: ~20µA × 5e-9 = 100 fA → well within 10 nA threshold
	goParams := arraysim.SolveParams{
		WLVoltages:  VWL,
		BLVoltages:  make([]float64, cols), // unused (BLDriveResistance >> any node Z)
		Conductance: G,
		Wire:        arraysim.WireParams{RWordLine: rWL, RBitLine: rBL},
		Boundary: arraysim.BoundaryParams{
			WLDriveResistance:       rWL, // exact badcrossbar WL source resistance
			BLDriveResistance:       1e9, // BL top: effectively open (gBLDrive=1e-9 S)
			BLTerminationResistance: rBL, // exact badcrossbar BL ground resistance
			BLTerminationVoltage:    0,
		},
	}
	solver := arraysim.NewTierASolver()
	goResult, err := solver.Solve(goParams)
	if err != nil {
		t.Fatalf("Go solver: %v", err)
	}

	t.Log("")
	t.Log("TIER 2 — Go MNA solver vs badcrossbar (exact circuit match, BLDriveR=1GΩ)")
	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("%-8s  %-14s  %-14s  %-12s  %-8s", "Cell", "Go I (µA)", "BC I (µA)", "Δ (nA)", "Status")
	t.Log("──────────────────────────────────────────────────────────────────────")

	const tier2Thr = 10e-9 // 10 nA — BLDriveR=1GΩ residual well within this
	maxTier2 := 0.0
	tier2Pass := true

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			goI := goResult.CellCurrents[r][c]
			bcI := py.BCDeviceI[r][c]
			diff := math.Abs(goI - bcI)
			if diff > maxTier2 {
				maxTier2 = diff
			}
			status := "PASS"
			if diff > tier2Thr {
				status = "FAIL"
				tier2Pass = false
				t.Errorf("Tier2 Device[%d][%d]: Go=%.8e A  BC=%.8e A  Δ=%.3e A", r, c, goI, bcI, diff)
			}
			t.Logf("(%d,%d)    %-14.6f  %-14.6f  %-12.4f  %s",
				r, c, goI*1e6, bcI*1e6, diff*1e9, status)
		}
	}

	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("%-8s  %-14s  %-14s  %-12s  %-8s", "Column", "Go Icol (µA)", "BC Iout (µA)", "Δ (nA)", "Status")
	t.Log("──────────────────────────────────────────────────────────────────────")

	for c := 0; c < cols; c++ {
		goIcol := goResult.ColCurrents[c]
		bcIout := py.BCOutputI[0][c]
		diff := math.Abs(goIcol - bcIout)
		if diff > maxTier2 {
			maxTier2 = diff
		}
		status := "PASS"
		if diff > tier2Thr {
			status = "FAIL"
			tier2Pass = false
			t.Errorf("Tier2 Col[%d]: Go=%.8e A  BC=%.8e A  Δ=%.3e A", c, goIcol, bcIout, diff)
		}
		t.Logf("Col %-3d   %-14.6f  %-14.6f  %-12.4f  %s",
			c, goIcol*1e6, bcIout*1e6, diff*1e9, status)
	}

	t.Log("──────────────────────────────────────────────────────────────────────")
	t.Logf("Tier 2 max error: %.3e A  (threshold: %.0e A)", maxTier2, tier2Thr)
	if tier2Pass {
		t.Log("PASS Tier 2: Go MNA solver agrees with badcrossbar within 10 nA")
	}

	allPass := tier1Pass && tier2Pass
	fmt.Printf("BADCROSSBAR_CROSSVAL: rows=%d cols=%d tier1=%.3e/%v tier2=%.3e/%v\n",
		rows, cols, maxTier1, tier1Pass, maxTier2, tier2Pass)

	// ── Emit JSON artifact ────────────────────────────────────────────────────
	dir := "../../output/validation/external"
	os.MkdirAll(dir, 0755)
	artifact := map[string]interface{}{
		"test":        "badcrossbar_crossval",
		"tool":        "badcrossbar",
		"version":     "1.1.0",
		"doi":         "10.1016/j.softx.2020.100617",
		"rows":        rows, "cols": cols,
		"r_wl_ohm":    rWL, "r_bl_ohm": rBL,
		"G_base_S":    Gbase,
		"tier1_max_A": maxTier1, "tier1_thr_A": tier1Thr, "tier1_pass": tier1Pass,
		"tier2_max_A": maxTier2, "tier2_thr_A": tier2Thr, "tier2_pass": tier2Pass,
		"pass":        allPass,
	}
	b, _ := json.MarshalIndent(artifact, "", "  ")
	os.WriteFile(dir+"/badcrossbar_crossval.json", b, 0644)
}
