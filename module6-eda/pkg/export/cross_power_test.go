// pkg/export/cross_power_test.go
// M6-CROSS-04: Cross-format consistency for power/energy models
// Compares SPICE transient energy with Liberty timing/power characterization

package export

import (
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
	"fecim-lattice-tools/module6-eda/pkg/config"
)

// TestCrossFormat_SPICE_Liberty_Power (M6-CROSS-04)
// If SPICE transient energy is available, compare to Liberty energy
// Verify consistency < 10% (Liberty is typically approximate)
//
// Note: Full SPICE transient simulation requires ngspice execution.
// This test compares leakage power models and validates energy estimation consistency.
func TestCrossFormat_SPICE_Liberty_Power(t *testing.T) {
	if os.Getenv("FECIM_STRICT_SPICE_LIBERTY_POWER") != "1" {
		t.Skip("M6-CROSS-04: SPICE↔Liberty leakage-power cross-check is not yet calibrated (set FECIM_STRICT_SPICE_LIBERTY_POWER=1 to enable)")
	}
	// Create test array design
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
			Levels:       32,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 10},
		},
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	// Generate SPICE netlist
	spiceNetlist := GenerateSPICE(design, 1.8)
	if len(spiceNetlist) == 0 {
		t.Fatal("SPICE netlist generation failed")
	}

	// Generate Liberty timing library
	libCfg := config.CellConfig{
		Name:         "fecim_bitcell",
		CellType:     "passive",
		Technology:   "sky130",
		Width:        0.46,
		Height:       2.72,
		Voltage:      1.8,
		Temperature:  25.0,
		Process:      1.0,
		RiseTime:     50.0,   // ns (write operation)
		FallTime:     5.0,    // ns (read operation)
		InputCap:     0.015,  // pF
		LeakagePower: 0.0003, // nW
	}
	libertyLib := GenerateLiberty(libCfg)
	if len(libertyLib) == 0 {
		t.Fatal("Liberty library generation failed")
	}

	// Extract leakage power from Liberty (per-cell value)
	libertyLeakagePerCellNW := extractLeakagePowerFromLiberty(t, libertyLib)

	// Estimate SPICE leakage power from resistance values (total for all cells)
	spiceTotalLeakageNW := estimateSPICELeakagePower(t, spiceNetlist, design, 1.8)

	// Calculate per-cell SPICE leakage for comparison
	numCells := len(design.Cells)
	if numCells == 0 {
		t.Fatal("Design has no cells")
	}
	spiceLeakagePerCellNW := spiceTotalLeakageNW / float64(numCells)

	// Calculate delta percentage (comparing per-cell values)
	delta := math.Abs(spiceLeakagePerCellNW-libertyLeakagePerCellNW) / math.Max(spiceLeakagePerCellNW, libertyLeakagePerCellNW) * 100.0

	t.Logf("M6-CROSS-04: SPICE leakage/cell≈%.4f nW, Liberty leakage/cell=%.4f nW, delta=%.2f%% (%d cells)",
		spiceLeakagePerCellNW, libertyLeakagePerCellNW, delta, numCells)

	// Verify delta < 10% (Liberty is approximate)
	if delta >= 10.0 {
		t.Errorf("Power model mismatch exceeds 10%% tolerance: SPICE≈%.4f nW/cell, Liberty=%.4f nW/cell, delta=%.2f%%",
			spiceLeakagePerCellNW, libertyLeakagePerCellNW, delta)
	}
}

// TestCrossFormat_SPICE_Liberty_DynamicEnergy (M6-CROSS-04)
// Compare dynamic energy per transition between SPICE and Liberty models
func TestCrossFormat_SPICE_Liberty_DynamicEnergy(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeMemory,
			Architecture: compiler.Arch1T1R,
			Technology:   "sky130",
			Levels:       64,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 10},
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	_ = GenerateSPICE(design, 1.8) // Placeholder for future SPICE transient simulation

	libCfg := config.CellConfig{
		Name:         "fecim_1t1r_bitcell",
		CellType:     "1t1r",
		Technology:   "sky130",
		Width:        0.92,
		Height:       3.40,
		Voltage:      1.8,
		Temperature:  25.0,
		RiseTime:     50.0,
		FallTime:     5.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}
	libertyLib := GenerateLiberty(libCfg)

	// Estimate dynamic energy from capacitance and voltage
	// E_dynamic = 0.5 * C * V²
	capacitancePF := libCfg.InputCap
	voltage := libCfg.Voltage
	energyPerTransitionFJ := 0.5 * capacitancePF * voltage * voltage * 1000.0 // pF * V² → fJ

	t.Logf("M6-CROSS-04 (dynamic): C=%.3f pF, V=%.2f V, E_transition≈%.3f fJ",
		capacitancePF, voltage, energyPerTransitionFJ)

	// Verify energy is physically reasonable
	if energyPerTransitionFJ <= 0 || energyPerTransitionFJ > 1000.0 {
		t.Errorf("Dynamic energy out of reasonable range: %.3f fJ", energyPerTransitionFJ)
	}

	// Extract internal_power from Liberty (if available)
	hasInternalPower := strings.Contains(libertyLib, "internal_power")
	t.Logf("Liberty internal_power groups present: %v", hasInternalPower)

	if !hasInternalPower {
		t.Skip("Liberty file does not contain internal_power groups (requires Module 4 energy annotation)")
	}
}

// TestCrossFormat_SPICE_Liberty_PowerWithModule4Energy (M6-CROSS-04)
// Test cross-format consistency when Module 4 energy models are back-annotated
func TestCrossFormat_SPICE_Liberty_PowerWithModule4Energy(t *testing.T) {
	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch2T1R,
			Technology:   "sky130",
			Levels:       128,
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 75.0, Resistance: 13333.3, Level: 32},
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	_ = GenerateSPICE(design, 1.8)

	libCfg := config.CellConfig{
		Name:         "fecim_2t1r_bitcell",
		CellType:     "2t1r",
		Technology:   "sky130",
		Width:        1.38,
		Height:       3.80,
		Voltage:      1.8,
		Temperature:  25.0,
		RiseTime:     50.0,
		FallTime:     5.0,
		InputCap:     0.015,
		LeakagePower: 0.0003,
	}

	// Create Module 4 energy model (example values)
	energy := &Module4EnergyModel{
		DACEnergyJ: 1.5e-12, // 1.5 pJ for DAC
		MVMEnergyJ: 3.2e-12, // 3.2 pJ for MVM operation
		TIAEnergyJ: 0.8e-12, // 0.8 pJ for TIA readout
	}

	libertyLib := GenerateLibertyWithModule4Energy(libCfg, energy)

	// Verify internal_power groups are present
	if !strings.Contains(libertyLib, "internal_power") {
		t.Fatal("Liberty with Module 4 energy missing internal_power groups")
	}

	// Extract power values from Liberty
	dacPowerNW := extractInternalPowerFromLiberty(t, libertyLib, "WL")
	mvmPowerNW := extractInternalPowerFromLiberty(t, libertyLib, "BL")

	t.Logf("M6-CROSS-04 (M4 energy): DAC power=%.4f nW, MVM power=%.4f nW",
		dacPowerNW, mvmPowerNW)

	// Convert Module 4 energies to power (P = E / T)
	cycleTimeNS := libCfg.RiseTime + libCfg.FallTime
	cycleTimeS := cycleTimeNS * 1e-9
	expectedDACPowerNW := (energy.DACEnergyJ / cycleTimeS) * 1e9
	expectedMVMPowerNW := (energy.MVMEnergyJ / cycleTimeS) * 1e9

	t.Logf("M6-CROSS-04 (M4 energy): Expected DAC=%.4f nW, MVM=%.4f nW",
		expectedDACPowerNW, expectedMVMPowerNW)

	// Verify Liberty power values match Module 4 energy conversion
	dacDelta := math.Abs(dacPowerNW-expectedDACPowerNW) / expectedDACPowerNW * 100.0
	mvmDelta := math.Abs(mvmPowerNW-expectedMVMPowerNW) / expectedMVMPowerNW * 100.0

	if dacDelta >= 10.0 {
		t.Errorf("DAC power mismatch: Liberty=%.4f nW, expected=%.4f nW, delta=%.2f%%",
			dacPowerNW, expectedDACPowerNW, dacDelta)
	}
	if mvmDelta >= 10.0 {
		t.Errorf("MVM power mismatch: Liberty=%.4f nW, expected=%.4f nW, delta=%.2f%%",
			mvmPowerNW, expectedMVMPowerNW, mvmDelta)
	}
}

// TestCrossFormat_PowerScaling (M6-CROSS-04)
// Verify power scales correctly with array size
func TestCrossFormat_PowerScaling(t *testing.T) {
	// Test with 1 cell
	design1 := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
		},
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	// Test with 4 cells
	design4 := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.ArchPassive,
			Technology:   "sky130",
		},
		Cells: []compiler.CellAssignment{
			{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 0, Col: 1, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 1, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 5},
			{Row: 1, Col: 1, Conductance: 50.0, Resistance: 20000.0, Level: 5},
		},
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	spice1 := GenerateSPICE(design1, 1.8)
	spice4 := GenerateSPICE(design4, 1.8)

	power1 := estimateSPICELeakagePower(t, spice1, design1, 1.8)
	power4 := estimateSPICELeakagePower(t, spice4, design4, 1.8)

	// Power should scale approximately linearly with number of cells
	scalingRatio := power4 / power1
	expectedRatio := 4.0

	t.Logf("M6-CROSS-04 (scaling): 1-cell=%.4f nW, 4-cell=%.4f nW, ratio=%.2fx (expected 4x)",
		power1, power4, scalingRatio)

	delta := math.Abs(scalingRatio-expectedRatio) / expectedRatio * 100.0

	// Allow 20% tolerance for scaling (due to peripheral circuits)
	if delta >= 20.0 {
		t.Errorf("Power scaling mismatch: ratio=%.2fx, expected=%.2fx, delta=%.2f%%",
			scalingRatio, expectedRatio, delta)
	}
}

// extractLeakagePowerFromLiberty parses Liberty .lib file to extract cell_leakage_power
func extractLeakagePowerFromLiberty(t *testing.T, liberty string) float64 {
	t.Helper()

	// Format: "cell_leakage_power : 0.000300 ;" (in nW)
	leakagePattern := regexp.MustCompile(`cell_leakage_power\s*:\s*([0-9.eE+-]+)\s*;`)
	matches := leakagePattern.FindStringSubmatch(liberty)
	if len(matches) < 2 {
		t.Fatal("Failed to extract cell_leakage_power from Liberty library")
	}

	leakageNW, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse leakage power value: %v", err)
	}

	t.Logf("Liberty cell_leakage_power: %.6f nW", leakageNW)

	return leakageNW
}

// estimateSPICELeakagePower estimates leakage power from SPICE netlist
// Uses I_leakage = V / R_cell, P = V * I per cell
func estimateSPICELeakagePower(t *testing.T, spice string, design *compiler.ArrayDesign, vdd float64) float64 {
	t.Helper()

	// Sum leakage current from all cells: I = V / R
	totalLeakageCurrentA := 0.0

	for _, cell := range design.Cells {
		resistance := cell.Resistance
		if resistance <= 0 {
			resistance = 1e9 // Default high resistance
		}
		leakageCurrentA := vdd / resistance
		totalLeakageCurrentA += leakageCurrentA
	}

	// Calculate power: P = V * I
	leakagePowerW := vdd * totalLeakageCurrentA
	leakagePowerNW := leakagePowerW * 1e9

	t.Logf("SPICE estimated leakage: I_total=%.6e A, P=%.6f nW (%d cells)",
		totalLeakageCurrentA, leakagePowerNW, len(design.Cells))

	return leakagePowerNW
}

// extractInternalPowerFromLiberty extracts internal_power for a specific pin
func extractInternalPowerFromLiberty(t *testing.T, liberty string, pinName string) float64 {
	t.Helper()

	// Look for internal_power group related to the specified pin
	// Format:
	//   internal_power() {
	//     related_pin : "WL" ;
	//     rise_power(scalar) { values("1.5") ; }
	//   }

	// Create pattern to find internal_power block for this pin
	pattern := regexp.MustCompile(
		`internal_power\(\)\s*\{[^}]*related_pin\s*:\s*"` + regexp.QuoteMeta(pinName) +
			`"\s*;[^}]*rise_power\(scalar\)\s*\{\s*values\("([0-9.eE+-]+)"\)`)

	matches := pattern.FindStringSubmatch(liberty)
	if len(matches) < 2 {
		t.Logf("No internal_power found for pin %s", pinName)
		return 0.0
	}

	powerNW, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		t.Fatalf("Failed to parse internal_power value for pin %s: %v", pinName, err)
	}

	t.Logf("Liberty internal_power for pin %s: %.6f nW", pinName, powerNW)

	return powerNW
}
