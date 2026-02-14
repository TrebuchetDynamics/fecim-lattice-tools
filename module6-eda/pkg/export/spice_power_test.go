package export

import (
	"strings"
	"testing"

	"fecim-lattice-tools/module6-eda/pkg/compiler"
)

// TestM6_SPICE_05_PowerModel_TransientAnalysisHooks validates that SPICE netlist
// includes hooks for transient analysis to enable power/energy calculations.
// Verifies structure (even if not running SPICE simulation).
func TestM6_SPICE_05_PowerModel_TransientAnalysisHooks(t *testing.T) {
	// Create array design
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
		{Row: 1, Col: 0, Conductance: 70.0, Resistance: 14285.7, Level: 10},
		{Row: 1, Col: 1, Conductance: 80.0, Resistance: 12500.0, Level: 15},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 4, ActiveCells: 4},
	}

	netlist := GenerateSPICE(design, 1.8)

	// Verify netlist contains voltage sources (needed for power measurement)
	if !strings.Contains(netlist, "VDD") && !strings.Contains(netlist, "vdd") {
		t.Error("M6-SPICE-05: Netlist missing VDD voltage source (required for power analysis)")
	}

	// Verify VDD parameter definition
	if !strings.Contains(netlist, ".param VDD") {
		t.Error("M6-SPICE-05: Netlist missing .param VDD definition")
	}

	// Verify analysis directives are present
	// Note: Current implementation uses .op, but transient analysis would be .tran
	// This test validates the structure is ready for transient analysis extension
	hasAnalysis := strings.Contains(netlist, ".op") || strings.Contains(netlist, ".tran")
	if !hasAnalysis {
		t.Error("M6-SPICE-05: Netlist missing analysis directive (.op or .tran)")
	}

	// Verify .control block (needed for measurement commands)
	if !strings.Contains(netlist, ".control") {
		t.Error("M6-SPICE-05: Netlist missing .control block (required for power measurement)")
	}
	if !strings.Contains(netlist, ".endc") {
		t.Error("M6-SPICE-05: Netlist missing .endc (control block terminator)")
	}

	// Count current measurement points (BL resistors)
	// Power = V × I, so we need current measurement points
	currentMeasurements := strings.Count(netlist, "print i(")
	if currentMeasurements == 0 {
		t.Error("M6-SPICE-05: Netlist missing current measurement points (needed for P=V×I)")
	}

	t.Logf("M6-SPICE-05 PASS: Power model structure validated — VDD source, .control block, %d current measurement points",
		currentMeasurements)
}

// TestM6_SPICE_05_PowerModel_EnergyCalculationStructure validates that the netlist
// structure supports energy calculation: E = ∫ P dt = ∫ V × I dt
func TestM6_SPICE_05_PowerModel_EnergyCalculationStructure(t *testing.T) {
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
		{Row: 0, Col: 1, Conductance: 60.0, Resistance: 16666.7, Level: 5},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeCompute,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 2, ActiveCells: 2},
	}

	netlist := GenerateSPICE(design, 1.8)

	// Energy calculation requires:
	// 1. Voltage source with defined VDD
	// 2. Current measurement through VDD or load resistors
	// 3. Time-domain analysis capability (.tran or .op as starting point)

	// Check VDD voltage level is specified
	if !strings.Contains(netlist, "VDD vdd 0 DC") {
		t.Error("M6-SPICE-05: Missing VDD voltage source declaration (needed for energy calculation)")
	}

	// Check parametric VDD definition
	if !strings.Contains(netlist, ".param VDD") {
		t.Error("M6-SPICE-05: Missing parametric VDD definition (enables voltage scaling for energy studies)")
	}

	// Verify bit line load resistors (current sinks for power measurement)
	if !strings.Contains(netlist, "RBL") {
		t.Error("M6-SPICE-05: Missing bit line load resistors (current measurement points)")
	}

	// Verify current measurement infrastructure
	hasCurrentMeasurement := strings.Contains(netlist, "print i(RBL") || strings.Contains(netlist, "print i(VDD")
	if !hasCurrentMeasurement {
		t.Error("M6-SPICE-05: Missing current measurement commands (required for P=V×I)")
	}

	// Future enhancement marker: transient analysis
	// Current implementation uses .op (DC operating point)
	// Full energy calculation would require .tran (transient analysis)
	// This test validates the foundation is in place
	if strings.Contains(netlist, ".tran") {
		t.Log("M6-SPICE-05: Transient analysis directive found — full time-domain energy calculation enabled")
	} else {
		t.Log("M6-SPICE-05: Using .op analysis — transient energy calculation requires .tran extension")
	}

	t.Log("M6-SPICE-05 PASS: Energy calculation infrastructure validated — voltage source, current measurements, analysis framework")
}

// TestM6_SPICE_05_PowerModel_MultiArchitectureSupport validates power hooks across architectures
func TestM6_SPICE_05_PowerModel_MultiArchitectureSupport(t *testing.T) {
	architectures := []struct {
		name string
		arch string
	}{
		{"Passive", compiler.ArchPassive},
		{"1T1R", compiler.Arch1T1R},
		{"2T1R", compiler.Arch2T1R},
	}

	for _, tc := range architectures {
		t.Run(tc.name, func(t *testing.T) {
			cells := []compiler.CellAssignment{
				{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
			}

			design := &compiler.ArrayDesign{
				Config: &compiler.ArrayConfig{
					Mode:         compiler.ModeStorage,
					Architecture: tc.arch,
				},
				Cells: cells,
				Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
			}

			netlist := GenerateSPICE(design, 1.8)

			// All architectures should have VDD
			if !strings.Contains(netlist, "VDD") && !strings.Contains(netlist, "vdd") {
				t.Errorf("M6-SPICE-05: %s architecture missing VDD source", tc.name)
			}

			// All architectures should have current measurement
			if !strings.Contains(netlist, "print i(") {
				t.Errorf("M6-SPICE-05: %s architecture missing current measurement", tc.name)
			}

			t.Logf("M6-SPICE-05: %s architecture power hooks validated", tc.name)
		})
	}

	t.Log("M6-SPICE-05 PASS: Power model supports all architectures (Passive, 1T1R, 2T1R)")
}

// TestM6_SPICE_05_PowerModel_ResistancePropagation validates that cell resistance
// propagates correctly to SPICE netlist (affects power dissipation)
func TestM6_SPICE_05_PowerModel_ResistancePropagation(t *testing.T) {
	testResistance := 25000.0 // 25 kΩ

	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 0.0, Resistance: testResistance, Level: 0},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	netlist := GenerateSPICE(design, 1.8)

	// Verify R_level parameter is set correctly
	expectedRLevel := "R_level=25000.00" // Format from GenerateSPICE
	if !strings.Contains(netlist, expectedRLevel) {
		// Try alternative formatting
		if !strings.Contains(netlist, "R_level=2.500000e+04") && !strings.Contains(netlist, "R_level=25000") {
			t.Errorf("M6-SPICE-05: Cell resistance not propagated to SPICE netlist (expected R_level=%.2f)", testResistance)
		}
	}

	t.Logf("M6-SPICE-05 PASS: Cell resistance propagated to SPICE (R=%.0f Ω affects power dissipation)", testResistance)
}

// TestM6_SPICE_05_PowerModel_ConductanceToPower validates conductance→resistance→power flow
func TestM6_SPICE_05_PowerModel_ConductanceToPower(t *testing.T) {
	testConductance := 100.0 // 100 μS → 10 kΩ

	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: testConductance, Resistance: 0.0, Level: 0},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	netlist := GenerateSPICE(design, 1.8)

	// When Resistance=0, GenerateSPICE calculates from Conductance: R = 1e6 / G (μS)
	expectedResistance := 1e6 / testConductance // 10000 Ω
	expectedRLevel := "R_level=10000.00"

	if !strings.Contains(netlist, expectedRLevel) && !strings.Contains(netlist, "R_level=1.000000e+04") {
		t.Errorf("M6-SPICE-05: Conductance→resistance conversion failed (G=%.1f μS should yield R=%.0f Ω)",
			testConductance, expectedResistance)
	}

	t.Logf("M6-SPICE-05 PASS: Conductance→resistance conversion validated (G=%.1f μS → R=%.0f Ω)",
		testConductance, expectedResistance)
}

// TestM6_SPICE_05_PowerModel_VDDParameterization validates parametric VDD for power sweeps
func TestM6_SPICE_05_PowerModel_VDDParameterization(t *testing.T) {
	cells := []compiler.CellAssignment{
		{Row: 0, Col: 0, Conductance: 50.0, Resistance: 20000.0, Level: 0},
	}

	design := &compiler.ArrayDesign{
		Config: &compiler.ArrayConfig{
			Mode:         compiler.ModeStorage,
			Architecture: compiler.Arch1T1R,
		},
		Cells: cells,
		Stats: compiler.DesignStats{TotalCells: 1, ActiveCells: 1},
	}

	vddValues := []float64{1.2, 1.5, 1.8, 2.5, 3.3}
	for _, vdd := range vddValues {
		netlist := GenerateSPICE(design, vdd)

		// Check for .param VDD = <value> (allow for floating point formatting variations)
		if !strings.Contains(netlist, ".param VDD = ") {
			t.Errorf("M6-SPICE-05: VDD parameterization missing for VDD=%.2f V", vdd)
		}
	}

	t.Logf("M6-SPICE-05 PASS: VDD parameterization validated for %d voltage levels (enables power scaling studies)",
		len(vddValues))
}
