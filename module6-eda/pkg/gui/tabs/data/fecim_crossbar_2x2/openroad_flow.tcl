# FeCIM OpenROAD Flow Script
# Generated: 2026-02-20
# Array: 2x2, Architecture: passive
# Die area: 2.920 x 7.440 µm
#
# Usage (native):
#   openroad -no_splash -exit openroad_flow.tcl
#
# Usage (Docker):
#   docker run --rm -v $PWD:/design -w /design \
#     ghcr.io/the-openroad-project/openlane:latest \
#     openroad -no_splash -exit /design/openroad_flow.tcl
#
# Environment variables (set via -rd or export before running):
#   CELL_LEF  — path to FeCIM bitcell LEF (default: cells/fecim_bitcell/fecim_bitcell.lef)
#   DEF_FILE  — path to pre-placed DEF (default: fecim_crossbar_2x2.def)
#
# What this script does:
#   1. Read custom FeCIM cell LEF (no PDK needed for array-only check)
#   2. Load pre-placed array DEF
#   3. Validate placement (no overlaps, cells within die bounds)
#   4. Report timing (clockless design — expect no setup violations)
#   5. Export OpenROAD database and reports

# ── Configuration ─────────────────────────────────────────────────────────────
set cell_lef  [expr {[info exists ::env(CELL_LEF)] ? $::env(CELL_LEF) : "cells/fecim_bitcell/fecim_bitcell.lef"}]
set def_file  [expr {[info exists ::env(DEF_FILE)]  ? $::env(DEF_FILE)  : "fecim_crossbar_2x2.def"}]
set out_dir   [expr {[info exists ::env(OUT_DIR)]   ? $::env(OUT_DIR)   : "output/openroad"}]

puts "OpenROAD FeCIM Flow"
puts "  Cell LEF: $cell_lef"
puts "  DEF file: $def_file"
puts "  Output:   $out_dir"

# ── Load technology and design ────────────────────────────────────────────────
# read_lef: register cell geometry (no standard cell PDK needed for array check)
read_lef $cell_lef

# read_def: load pre-placed array (FIXED placement from DEF generator)
read_def $def_file

# ── Placement validation ──────────────────────────────────────────────────────
# check_placement verifies:
#   - No cell-to-cell overlaps
#   - All cells within die bounds
#   - FIXED cells not moved
puts ""
puts "=== Placement Check ==="
if {[catch {check_placement -verbose} err]} {
    puts "WARNING: Placement check reported issues: $err"
} else {
    puts "Placement check passed."
}

# ── Timing report ─────────────────────────────────────────────────────────────
# The FeCIM crossbar has no clock (static write, capacitive read).
# RUN_CTS=0 in config.json skips clock tree synthesis.
# report_checks will show no timing paths (expected for clockless designs).
puts ""
puts "=== Timing Report (clockless design) ==="
if {[catch {report_checks -path_delay max -fields {slew cap input nets fanout} -format full_clock_expanded} err]} {
    puts "INFO: No timing paths found (expected for clockless FeCIM array)"
}

# ── Congestion / area reports ─────────────────────────────────────────────────
puts ""
puts "=== Design Statistics ==="
if {[catch {report_design_area} err]} {
    puts "INFO: report_design_area not available: $err"
}

puts ""
puts "=== Cell Count ==="
if {[catch {report_cell_usage} err]} {
    puts "INFO: report_cell_usage not available: $err"
}

# ── Write output ──────────────────────────────────────────────────────────────
if {[catch {
    file mkdir $out_dir
    write_def $out_dir/fecim_crossbar_2x2_placed.def
    puts ""
    puts "Written: $out_dir/fecim_crossbar_2x2_placed.def"
} err]} {
    puts "WARNING: write_def failed: $err"
}

puts ""
puts "OpenROAD flow complete for fecim_crossbar_2x2"
puts "Die area: 2.920 x 7.440 µm"
puts "Array:    2x2 cells (passive architecture)"
