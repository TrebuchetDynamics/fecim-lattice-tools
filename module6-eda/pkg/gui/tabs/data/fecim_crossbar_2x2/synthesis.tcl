# FeCIM Yosys Synthesis Script
# Generated: 2026-02-20
# Array: 2x2, Architecture: passive, Technology: sky130
#
# Usage (native):
#   yosys synth.tcl
#
# Usage (Docker/LibreLane):
#   docker run --rm -v $PWD:/design -w /design \
#     ghcr.io/the-openroad-project/openlane:latest yosys synth.tcl
#
# What this script does:
#   Step 1 — Blackbox the FeCIM bitcell (no logic mapping needed for analog cells)
#   Step 2 — Read structural array netlist
#   Step 3 — Hierarchy + DRC check (validates connectivity)
#   Step 4 — Optional: synthesize any attached control logic with synth_sky130
#
# Note: SYNTH_ELABORATE_ONLY=1 is set in config.json because the array is
# pre-placed structural Verilog. Only run synth_sky130 if you add a digital
# control wrapper module around the array.

# ── Step 1: Blackbox bitcell ─────────────────────────────────────────────────
# read_verilog -lib: registers cell interface without synthesizing internals
read_verilog -lib cells/fecim_bitcell/fecim_bitcell.v

# ── Step 2: Read structural array netlist ────────────────────────────────────
read_verilog output/fecim_crossbar_2x2.v

# ── Step 3: Hierarchy check ──────────────────────────────────────────────────
# -check: assert all modules are defined
# -top: specify top-level module
hierarchy -check -top fecim_crossbar

# Run DRC check on loaded design
check

# Print cell statistics
stat

# ── Step 4 (Optional): Synthesize control logic ──────────────────────────────
# Uncomment the block below ONLY if you have added a digital control wrapper.
# The pure FeCIM array is structural Verilog and does not need logic synthesis.
#
# # Map to sky130 standard cells:
# synth_sky130 -top control_wrapper -json output/control_wrapper.json
#
# # Alternative: write structural netlist without optimization:
# # write_verilog -noattr output/fecim_crossbar_2x2_synth.v
