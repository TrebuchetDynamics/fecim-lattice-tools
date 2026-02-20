// pkg/export/pyspice.go
// PySpice and OpenVAF simulation script generators for FeCIM arrays.
//
// Two simulation pathways:
//
//  1. PySpice + Ngspice:
//     PySpice (https://github.com/FabriceSalvaire/PySpice) provides a Python
//     interface to Ngspice for circuit-level simulation of crossbar arrays.
//     The generated script builds a crossbar netlist programmatically,
//     sweeps the input voltage, and extracts per-cell read currents.
//
//  2. OpenVAF Verilog-A compiler:
//     OpenVAF (https://github.com/OpenVAF/OpenVAF) compiles Verilog-A models
//     to OSDI shared objects for use with OSDI-compatible simulators.
//     The generated Verilog-A model captures the Landau-Khalatnikov FeCIM
//     device physics (polarization hysteresis, remnant states).
//     Compile: openvaf fecim_lk.va --output fecim_lk.so
//
// References:
//   PySpice:  https://github.com/FabriceSalvaire/PySpice
//   OpenVAF:  https://github.com/OpenVAF/OpenVAF
//   Ngspice:  https://ngspice.sourceforge.io/
//   OSDI:     https://github.com/OpenVAF/OSDI
package export

import (
	"fmt"
	"strings"
	"time"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

// GeneratePySpiceScript returns a Python script that uses PySpice to build
// a FeCIM crossbar netlist and simulate it with Ngspice.
//
// The script models each crossbar cell as a conductance (Ron/Roff based on
// its programmed state). A sweep of the input DAC voltage reads out the
// column current (analog MVM result).
//
// Usage:
//
//	pip install PySpice pyyaml numpy matplotlib
//	pip install ngspice (or install system ngspice)
//	python3 run_pyspice.py
func GeneratePySpiceScript(cfg config.ArrayConfig) string {
	designName := fmt.Sprintf("fecim_crossbar_%dx%d", cfg.Rows, cfg.Cols)

	// Conductance range
	var gMaxUS, gMinUS float64
	switch strings.ToLower(cfg.Architecture) {
	case "1t1r", "2t1r":
		gMaxUS = 100.0
		gMinUS = 0.01
	default:
		gMaxUS = 10.0
		gMinUS = 0.001
	}
	// Wire resistance per word line (1 Ω/cell from metal sheet resistance).
	// WL is horizontal, spanning cfg.Cols cells, so total WL resistance = cols × 1 Ω/cell.
	wireResOhm := float64(cfg.Cols) * 1.0

	return fmt.Sprintf(`#!/usr/bin/env python3
# FeCIM PySpice Crossbar Simulation
# Generated: %s
# Design: %s
# Array: %dx%d, architecture=%s, technology=%s
#
# Simulates the FeCIM crossbar as a resistive MVM (matrix-vector multiply).
# Each cell is modeled as Ron (LRS) or Roff (HRS) based on its programmed level.
# Column-sum currents represent the MVM output (weighted sum of row inputs).
#
# Usage:
#   pip install PySpice pyyaml numpy matplotlib
#   python3 run_pyspice.py
#
# Ngspice must be installed:
#   Ubuntu: sudo apt install ngspice
#   macOS:  brew install ngspice
#   PyPI:   pip install PySpice[ngspice]  (bundled ngspice)

import sys
import os
import numpy as np

try:
    from PySpice.Spice.Netlist import Circuit
    from PySpice.Unit import *
    from PySpice.Spice.NgSpice.Shared import NgSpiceShared
except ImportError:
    print("PySpice not installed. Install with:")
    print("  pip install PySpice")
    sys.exit(1)

# ── Array Configuration ─────────────────────────────────────────────────────
ROWS = %d
COLS = %d
ARCHITECTURE = "%s"

# Conductance range (Siemens)
G_MAX = %.6fe-6          # µS → S (low-resistance state, LRS)
G_MIN = %.9fe-6          # µS → S (high-resistance state, HRS)
G_RANGE = G_MAX - G_MIN

# Wire resistance (parasitic line resistance per row)
R_WIRE_OHM = %.2f         # Ω total word-line resistance

# Supply voltage
V_READ = 0.1              # V read voltage (small to avoid disturb)
V_SUPPLY = 1.8            # V supply (for TIA bias)

# ── Build Weight Matrix ─────────────────────────────────────────────────────
# Random conductance levels (0 to 30 discrete states → linear mapping to G)
np.random.seed(42)
N_LEVELS = 30
level_matrix = np.random.randint(0, N_LEVELS + 1, size=(ROWS, COLS))
G_matrix = G_MIN + level_matrix / N_LEVELS * G_RANGE  # Conductance per cell (S)
R_matrix = 1.0 / G_matrix                              # Resistance per cell (Ω)

print(f"FeCIM PySpice Simulation: %s")
print(f"Array: {{ROWS}}x{{COLS}} = {{ROWS*COLS}} cells")
print(f"G range: [{{G_MIN*1e6:.4f}}, {{G_MAX*1e6:.4f}}] µS")
print(f"R range: [{{(1/G_MAX):.0f}}, {{(1/G_MIN):.0f}}] Ω")
print("")

# ── Build SPICE Netlist ─────────────────────────────────────────────────────
circuit = Circuit('FeCIM_Crossbar_%dx%d')

# Ground node
circuit.raw_spice += '.global GND\\n'

# Row voltage sources (DAC outputs — uniform input for test)
input_vector = np.ones(ROWS)  # Uniform input (test vector)

for row in range(ROWS):
    v_in = V_READ * input_vector[row]
    circuit.V(f'row{{row}}', f'WL{{row}}', circuit.gnd, v_in)

# FeCIM cells as resistors (R = 1/G per programmed state)
# Row wire resistance modeled as series resistors
for row in range(ROWS):
    for col in range(COLS):
        r_cell = R_matrix[row, col]
        r_wire_seg = R_WIRE_OHM / COLS  # Distribute wire R across WL (COLS segments per word line)

        # Word-line wire segment (parasitic)
        circuit.R(f'Rwire_{{row}}_{{col}}', f'WL{{row}}_{{col}}',
                  f'WL{{row}}_{{col+1}}' if col < COLS-1 else f'WL{{row}}',
                  r_wire_seg)

        # Ferroelectric cell (programmed conductance)
        circuit.R(f'Rcell_{{row}}_{{col}}', f'WL{{row}}_{{col+1}}' if col < COLS-1 else f'WL{{row}}',
                  f'BL{{col}}', r_cell)

# Column virtual ground (TIA sense amplifier holds BL at virtual GND)
for col in range(COLS):
    circuit.V(f'Vtia_{{col}}', f'BL{{col}}', circuit.gnd, 0)  # V=0 (TIA virtual ground)

# ── Run DC Analysis ─────────────────────────────────────────────────────────
print("Running Ngspice DC analysis...")
print(f"Netlist: {{len(list(circuit.elements))}} elements")

try:
    simulator = circuit.simulator(temperature=25, nominal_temperature=25)
    analysis = simulator.operating_point()

    # Column currents (MVM output)
    col_currents = np.array([
        float(analysis[f'vtia{{col}}'])
        for col in range(COLS)
    ])

    # Ideal MVM (no parasitics)
    ideal_output = G_matrix.T @ input_vector  # cols × rows × rows = cols

    print("")
    print("=== FeCIM MVM Result ===")
    for col in range(COLS):
        i_sim = col_currents[col] * 1e6  # A → µA
        i_ideal = ideal_output[col] * V_READ * 1e6  # µA
        print(f"  BL{{col}}: I_sim={{i_sim:+.3f}} µA  I_ideal={{i_ideal:+.3f}} µA")

    # Error analysis
    rmse = np.sqrt(np.mean((col_currents - ideal_output * V_READ) ** 2))
    print(f"")
    print(f"RMSE vs ideal: {{rmse*1e9:.2f}} nA")
    print(f"Relative RMSE: {{rmse / np.mean(np.abs(ideal_output * V_READ)) * 100:.2f}}%%")

except Exception as e:
    print(f"Ngspice simulation error: {{e}}")
    print("Make sure ngspice is installed: sudo apt install ngspice")
    sys.exit(1)

print("")
print("Simulation complete.")
print(f"Crossbar: %dx%d cells, {{ARCHITECTURE}} architecture")
`, time.Now().Format("2006-01-02"),
		designName,
		cfg.Rows, cfg.Cols, cfg.Architecture, cfg.Technology,
		cfg.Rows, cfg.Cols,
		cfg.Architecture,
		gMaxUS, gMinUS,
		wireResOhm,
		designName,
		cfg.Rows, cfg.Cols,
		cfg.Rows, cfg.Cols,
	)
}

// GenerateOpenVAFVerilogA returns a Verilog-A compact model for the FeCIM
// ferroelectric cell using a simplified Landau-Khalatnikov (L-K) equation.
//
// The model is suitable for compilation with OpenVAF to produce an
// OSDI-compatible shared object for Ngspice or Melange simulation.
//
// Compile:
//
//	openvaf fecim_lk.va --output fecim_lk.so
//	# Then simulate with Ngspice (requires OSDI-enabled Ngspice):
//	ngspice circuit_with_fecim.sp
//
// References:
//   OpenVAF:          https://github.com/OpenVAF/OpenVAF
//   L-K equation:     Landau & Khalatnikov, Dokl. Akad. Nauk SSSR 96, 469 (1954)
//   FeCIM parameters: Mikolajick et al., Adv. Electron. Mater. 2020
func GenerateOpenVAFVerilogA(cfg config.CellConfig) string {
	tech := strings.ToLower(cfg.Technology)
	var ecMVcm, prUCcm2, psSvg, tFE float64
	switch {
	case strings.Contains(tech, "ihp") || strings.Contains(tech, "sg13"):
		ecMVcm = 1.0  // MV/cm coercive field (typical HZO for 130nm node)
		prUCcm2 = 15.0 // µC/cm² remnant polarization
		psSvg = 30.0  // µC/cm² saturation polarization
		tFE = 10.0    // nm FE layer thickness
	case strings.Contains(tech, "gf180"):
		ecMVcm = 1.0
		prUCcm2 = 15.0
		psSvg = 30.0
		tFE = 10.0
	default: // sky130 / generic HZO
		ecMVcm = 1.0
		prUCcm2 = 15.0
		psSvg = 30.0
		tFE = 10.0
	}

	// Note: Verilog-A uses backtick directives (`include) and $vt operator.
	// Go raw string literals can't contain backticks, so we build the string
	// by concatenation.
	tick := "`"
	header := fmt.Sprintf(
		"// FeCIM Ferroelectric Cell - Verilog-A Compact Model\n"+
			"// Generated: %s\n"+
			"// Technology: %s\n"+
			"//\n"+
			"// Simplified Landau-Khalatnikov (L-K) model for HZO ferroelectric capacitor.\n"+
			"// Compile with OpenVAF:\n"+
			"//   openvaf fecim_lk.va --output fecim_lk.so\n"+
			"//\n"+
			"// Terminals:\n"+
			"//   plus  (T): top electrode (applied voltage)\n"+
			"//   minus (B): bottom electrode (BL or GND)\n"+
			"//\n"+
			"// Parameters: Mikolajick et al., Adv. Electron. Mater. 6, 1900078 (2020)\n"+
			"// WARNING: Educational model only — not silicon-validated.\n\n",
		time.Now().Format("2006-01-02"), cfg.Technology,
	)
	body := fmt.Sprintf(
		tick+"include \"constants.vams\"\n"+
			tick+"include \"disciplines.vams\"\n\n"+
			"module fecim_lk(plus, minus);\n"+
			"    inout plus, minus;\n"+
			"    electrical plus, minus;\n\n"+
			"    // Physical parameters\n"+
			"    parameter real T_FE    = %.1f  from (0:inf); // nm FE layer thickness\n"+
			"    parameter real EC      = %.2f  from (0:inf); // MV/cm coercive field\n"+
			"    parameter real PR      = %.1f  from (0:inf); // uC/cm2 remnant polarization\n"+
			"    parameter real PS      = %.1f  from (0:inf); // uC/cm2 saturation polarization\n"+
			"    parameter real AREA    = 1.0   from (0:inf); // um2 cell area\n"+
			"    parameter real RHO     = 1.0e5 from (0:inf); // Ohm*cm2 L-K damping coeff\n"+
			"    parameter real EPS_INF = 25.0  from (0:inf); // Background dielectric\n\n"+
			"    // State variables\n"+
			"    real P, dPdt, Vfe, Efe, C_par, alpha, beta;\n\n"+
			"    initial begin\n"+
			"        P = -PR;  // Start in negative remnant state\n"+
			"    end\n\n"+
			"    analog begin\n"+
			"        Vfe   = V(plus, minus);\n"+
			"        Efe   = Vfe / (T_FE * 1e-7);     // V/cm\n"+
			"        alpha = -EC / (2.0 * PS);\n"+
			"        beta  =  EC / (4.0 * PS * PS * PS);\n"+
			"        // L-K: rho * dP/dt = E - 2*alpha*P - 4*beta*P^3\n"+
			"        dPdt = (Efe - 2.0*alpha*P - 4.0*beta*P*P*P) / (RHO * 1e-4);\n"+
			"        P     = idt(dPdt, -PR);\n"+
			"        C_par = 8.854e-14 * EPS_INF * (AREA * 1e-8) / (T_FE * 1e-7);\n"+
			"        I(plus, minus) <+ (AREA * 1e-8) * dPdt * 1e-6;\n"+
			"        I(plus, minus) <+ C_par * ddt(Vfe);\n"+
			"    end\n\n"+
			"endmodule  // fecim_lk\n",
		tFE, ecMVcm, prUCcm2, psSvg,
	)
	return header + body
}
