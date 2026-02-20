#!/usr/bin/env python3
# FeCIM KLayout DEF+LEF → GDS II Script
# Generated: 2026-02-20
# Array: 2x2, Architecture: passive
#
# This script converts the abstract FeCIM bitcell LEF + DEF placement into
# a GDS II layout file for use with OpenLane/LibreLane stream-out.
#
# The output GDS is a simplified rectangular representation suitable for:
#   - OpenLane EXTRA_GDS_FILES injection
#   - KLayout XOR verification (design intent vs. routed GDS)
#   - Educational layout visualization
#
# NOTE: This produces an abstract/educational GDS. Production tapeout requires
#       full transistor-level layout and foundry DRC sign-off (Magic + sky130A PDK).
#
# Usage:
#   klayout -z -r gen_gds.py \
#     -rd lef_file=cells/fecim_bitcell/fecim_bitcell.lef \
#     -rd def_file=fecim_crossbar_2x2.def \
#     -rd out_file=cells/fecim_bitcell/fecim_bitcell.gds
#
#   Or with Docker (OpenLane image contains klayout + pya):
#   docker run --rm -v $PWD:/design -w /design \
#     ghcr.io/the-openroad-project/openlane:latest \
#     klayout -z -r gen_gds.py \
#       -rd lef_file=cells/fecim_bitcell/fecim_bitcell.lef \
#       -rd def_file=fecim_crossbar_2x2.def \
#       -rd out_file=cells/fecim_bitcell/fecim_bitcell.gds

import os
import sys

# ── pya import ────────────────────────────────────────────────────────────────
# pya is the KLayout Python API. Available in:
#   - Native KLayout (https://www.klayout.de/build.html)
#   - OpenLane/LibreLane Docker image (bundled klayout)
#
# NOTE: "pip install klayout" installs KLayout's Python module, but the PyPI
# package has LIMITED DEF/LEF support. For reliable DEF→GDS conversion,
# use native KLayout (installed via package manager or downloaded from klayout.de)
# or the OpenLane Docker container.
try:
    import pya
except ImportError:
    print("ERROR: pya not found. Run this script inside KLayout:")
    print("  klayout -z -r gen_gds.py -rd lef_file=... -rd def_file=... -rd out_file=...")
    print("  NOTE: 'pip install klayout' has limited DEF/LEF support.")
    print("  Use native KLayout: https://www.klayout.de/build.html")
    sys.exit(1)

# ── Variable injection (from -rd flags) ───────────────────────────────────────
# KLayout passes -rd NAME=VALUE as module-level variables.
# Default values are provided for standalone testing.
_lef_file = getattr(pya, 'lef_file', None) or lef_file if 'lef_file' in dir() else "cells/fecim_bitcell/fecim_bitcell.lef"
_def_file = getattr(pya, 'def_file', None) or def_file if 'def_file' in dir() else "fecim_crossbar_2x2.def"
_out_file = getattr(pya, 'out_file', None) or out_file if 'out_file' in dir() else "cells/fecim_bitcell/fecim_bitcell.gds"

print(f"KLayout GDS generator")
print(f"  LEF: {_lef_file}")
print(f"  DEF: {_def_file}")
print(f"  GDS: {_out_file}")

# ── Load layout ───────────────────────────────────────────────────────────────
layout = pya.Layout()

# Use lefdef_config approach: register LEF files as configuration for DEF reading.
# This is the correct KLayout pya API for combined LEF+DEF workflows — it lets
# KLayout resolve cell references in the DEF against the LEF geometry.
layout_options = pya.LoadLayoutOptions()

if os.path.exists(_lef_file):
    # Register LEF as cell-definition source for the DEF reader
    layout_options.lefdef_config.lef_files = [_lef_file]
    print(f"  LEF registered: {_lef_file}")
else:
    print(f"WARNING: LEF file not found: {_lef_file}")
    print("  Generating minimal GDS from DEF placement only")

# Load DEF (defines array placement); LEF geometry is resolved via lefdef_config
if os.path.exists(_def_file):
    layout.read(_def_file, layout_options)
    print(f"  Loaded DEF: {_def_file}")
else:
    print(f"ERROR: DEF file not found: {_def_file}")
    print(f"  Generate DEF first: fecim-lattice-tools eda cli --def")
    sys.exit(1)

# ── Stream out GDS ────────────────────────────────────────────────────────────
os.makedirs(os.path.dirname(os.path.abspath(_out_file)), exist_ok=True)

writer_opts = pya.SaveLayoutOptions()
writer_opts.format = "GDS2"

layout.write(_out_file, writer_opts)
print(f"  Written GDS: {_out_file}")
print()
print("GDS generation complete.")
print("Next steps:")
print("  1. Verify with: klayout " + _out_file)
print("  2. Reference in config.json EXTRA_GDS_FILES field")
print("  3. Run OpenLane/LibreLane flow: ./run_flow.sh")
