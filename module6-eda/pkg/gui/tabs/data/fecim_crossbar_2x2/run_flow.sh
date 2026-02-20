#!/usr/bin/env bash
# FeCIM RTL-to-GDS Flow Runner
# Generated: 2026-02-20
# Array: 2x2, Architecture: passive
#
# This script orchestrates the complete open-source RTL-to-GDS flow:
#   Step 1 — Validate Verilog with Yosys
#   Step 2 — Generate GDS stub with KLayout (DEF+LEF → GDS II)
#   Step 3 — Check placement with OpenROAD
#   Step 4 — Run LibreLane/OpenLane flow (if installed)
#
# Tools used (install via package manager or Docker):
#   yosys    — RTL synthesis (apt: yosys)
#   klayout  — Layout and GDS generation (apt: klayout)
#   openroad — Place & Route (apt: openroad OR via Docker)
#   librelane — RTL-to-GDS flow (pip: librelane, successor to OpenLane)
#   OpenLane  — Legacy RTL-to-GDS flow (Docker: efabless/openlane2)
#
# LibreLane (recommended for new designs):
#   pip install librelane
#   python -m librelane --config-file config.json
#
# OpenLane v1 (legacy, maintenance mode):
#   cd $OPENLANE_ROOT && ./flow.tcl -design fecim_array
#
# Docker alternative (includes all tools):
#   docker pull ghcr.io/the-openroad-project/openlane:latest

set -e
DESIGN="fecim_crossbar_2x2"
CELL="fecim_bitcell"
OUTPUT="output"
CELLS_DIR="cells/${CELL}"

echo "==================================================="
echo "FeCIM RTL-to-GDS Flow: ${DESIGN}"
echo "==================================================="
echo ""

# ── Step 1: Yosys hierarchy check ────────────────────────────────────────────
echo "Step 1: Yosys hierarchy check..."
if command -v yosys &>/dev/null; then
    yosys -p "read_verilog -lib ${CELLS_DIR}/${CELL}.v; read_verilog ${OUTPUT}/${DESIGN}.v; hierarchy -check -top fecim_crossbar; check; stat" \
        2>&1 | tee output/yosys_check.log
    echo "  ✓ Yosys check passed — see output/yosys_check.log"
else
    echo "  ⚠ Yosys not found. Skipping synthesis check."
    echo "    Install: sudo apt install yosys"
    echo "    Or run inside Docker image."
fi
echo ""

# ── Step 2: KLayout DEF+LEF → GDS ────────────────────────────────────────────
echo "Step 2: KLayout GDS generation..."
mkdir -p "${CELLS_DIR}"
if command -v klayout &>/dev/null; then
    klayout -z -r gen_gds.py \
        -rd lef_file="${CELLS_DIR}/${CELL}.lef" \
        -rd def_file="${OUTPUT}/${DESIGN}.def" \
        -rd out_file="${CELLS_DIR}/${CELL}.gds" \
        2>&1 | tee output/klayout_gds.log
    echo "  ✓ GDS written: ${CELLS_DIR}/${CELL}.gds"
else
    echo "  ⚠ KLayout not found. EXTRA_GDS_FILES will be missing."
    echo "    Install: sudo apt install klayout"
    echo "    Or run inside Docker: docker run ... klayout -z -r gen_gds.py ..."
fi
echo ""

# ── Step 3: OpenROAD placement check ─────────────────────────────────────────
echo "Step 3: OpenROAD placement check..."
mkdir -p output/openroad
if command -v openroad &>/dev/null; then
    CELL_LEF="${CELLS_DIR}/${CELL}.lef" \
    DEF_FILE="${OUTPUT}/${DESIGN}.def" \
    OUT_DIR="output/openroad" \
    openroad -no_splash -exit openroad_flow.tcl \
        2>&1 | tee output/openroad_check.log
    echo "  ✓ Placement check passed — see output/openroad_check.log"
else
    echo "  ⚠ OpenROAD not found. Skipping placement check."
    echo "    Install: sudo apt install openroad"
fi
echo ""

# ── Step 4: LibreLane / OpenLane flow ────────────────────────────────────────
echo "Step 4: Full RTL-to-GDS flow..."

if command -v librelane &>/dev/null || python3 -m librelane --version &>/dev/null 2>&1; then
    echo "  Using LibreLane (successor to OpenLane)..."
    python3 -m librelane --config-file config.json 2>&1 | tee output/librelane.log
    echo "  ✓ LibreLane flow complete — see output/librelane.log"
elif [[ -n "${OPENLANE_ROOT}" ]] && [[ -f "${OPENLANE_ROOT}/flow.tcl" ]]; then
    echo "  Using OpenLane v1 (legacy, maintenance mode)..."
    echo "  Note: For new projects, use LibreLane (pip install librelane)"
    mkdir -p "${OPENLANE_ROOT}/designs/fecim_array/src"
    mkdir -p "${OPENLANE_ROOT}/designs/fecim_array/cells"
    cp "${OUTPUT}/${DESIGN}.v" "${OPENLANE_ROOT}/designs/fecim_array/src/"
    cp "config.json" "${OPENLANE_ROOT}/designs/fecim_array/"
    cp -r "${CELLS_DIR}" "${OPENLANE_ROOT}/designs/fecim_array/cells/" 2>/dev/null || true
    cd "${OPENLANE_ROOT}"
    ./flow.tcl -design fecim_array 2>&1 | tee "${OLDPWD}/output/openlane.log"
    echo "  ✓ OpenLane flow complete — see output/openlane.log"
else
    echo "  ⚠ Neither LibreLane nor OpenLane found."
    echo "    Install LibreLane: pip install librelane"
    echo "    Docs: https://librelane.readthedocs.io/"
    echo ""
    echo "    Or use Docker:"
    echo "    docker run --rm -v \$PWD:/design -w /design \\"
    echo "      ghcr.io/the-openroad-project/openlane:latest \\"
    echo "      python3 -m librelane --config-file /design/config.json"
fi

echo ""
echo "==================================================="
echo "Flow complete for ${DESIGN}"
echo "Outputs:"
echo "  Verilog: ${OUTPUT}/${DESIGN}.v"
echo "  DEF:     ${OUTPUT}/${DESIGN}.def"
echo "  GDS:     ${CELLS_DIR}/${CELL}.gds"
echo "  Config:  config.json"
echo "==================================================="
