# FeCIM SDC Timing Constraints
# Generated: 2026-02-20
# Design:    fecim_crossbar_2x2
# Array:     2x2 passive (sky130)
#
# This SDC is appropriate for a pure FeCIM crossbar array with NO clock.
# - RUN_CTS=0 in config.json (no Clock Tree Synthesis)
# - All write/read paths are combinational (set by external DAC timing)
# - No setup/hold violations expected
#
# If you add a digital control wrapper (FSM, address decoder) UNCOMMENT
# the clock section below and set a realistic period.

# ── Current design: No clock ──────────────────────────────────────────────────
# The FeCIM array is driven by DAC outputs (word lines) and sensed
# by TIA/ADC (bit lines). Timing is governed by the peripheral circuits,
# not by the array itself.

# I/O delay: 0 ns (array pins connect directly to peripheral circuits)
set_input_delay  0.0 [all_inputs]
set_output_delay 0.0 [all_outputs]

# Max transition: FeFET write path constraint
# Reference: Trentzsch et al. IEDM 2016 (28nm FDSOI FeFET, ~50 ns write)
# Using conservative 10 ns here (read-dominated timing requirement)
set_max_transition 10.0 [all_outputs]

# Load capacitance: FeFET input capacitance (for STA buffer sizing)
# Reference: FeFET mid-range input cap ~0.015 pF
set_load 0.0150 [all_outputs]

# ── Optional: Digital control wrapper clock ───────────────────────────────────
# Uncomment and set CLK_PERIOD if a control FSM is added around the array.
# SKY130 max speed grade: ~100 MHz at 1.8V typical corner.
# Suggested values for educational designs:
#   10 ns period → 100 MHz (fast, near sky130 limit)
#   20 ns period →  50 MHz (balanced, good timing margin)
#   40 ns period →  25 MHz (conservative, easy closure)
#
# set CLK_PERIOD 20.0
# create_clock -period $CLK_PERIOD -name clk [get_ports clk]
# set_clock_uncertainty 0.25 [all_clocks]
# set_clock_transition  0.15 [all_clocks]
# set_input_delay  [expr {$CLK_PERIOD * 0.15}] -clock clk [get_ports {WL[*]}]
# set_output_delay [expr {$CLK_PERIOD * 0.15}] -clock clk [get_ports {BL[*]}]
