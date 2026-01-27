# OpenLane & Open-Source EDA Tools: Comprehensive CLI Reference

**Purpose:** Complete reference for OpenLane flow and its component tools
**Last Updated:** 2026-01-26
**Sources:** Official documentation, GitHub repositories, community resources

---

## Table of Contents

1. [OpenLane Overview](#1-openlane-overview)
2. [Yosys - Synthesis](#2-yosys---synthesis)
3. [OpenROAD - Physical Design](#3-openroad---physical-design)
4. [Magic VLSI - Layout](#4-magic-vlsi---layout)
5. [KLayout - GDSII Viewer/Editor](#5-klayout---gdsii-viewereditor)
6. [Netgen - LVS](#6-netgen---lvs)
7. [OpenSTA - Timing Analysis](#7-opensta---timing-analysis)
8. [PDK Installation](#8-pdk-installation)
9. [Quick Reference Tables](#9-quick-reference-tables)

---

## 1. OpenLane Overview

OpenLane is an automated RTL-to-GDSII flow based on OpenROAD, Yosys, Magic, Netgen, CVC, KLayout, and custom scripts.

**Official Resources:**
- Documentation: https://openlane2.readthedocs.io/
- GitHub (OpenLane 2): https://github.com/efabless/openlane2
- GitHub (OpenLane 1): https://github.com/The-OpenROAD-Project/OpenLane
- PyPI: https://pypi.org/project/openlane/

### 1.1 OpenLane 1 vs OpenLane 2

| Aspect | OpenLane 1 | OpenLane 2 |
|--------|------------|------------|
| Language | Tcl scripts | Python with Tcl compatibility |
| Configuration | config.tcl | JSON, YAML, or Tcl |
| Architecture | Procedural scripts | Modular step-based |
| API | Minimal | Well-documented Python API |
| Type Checking | None | Built-in validation |
| Entry Point | `./flow.tcl` | `openlane` CLI |

### 1.2 Installation Methods

#### Nix-based (Recommended)
```bash
# Install Nix and Cachix first
nix-shell
openlane --smoke-test
```

#### Docker-based
```bash
cd $HOME/OpenLane
make mount
# Inside container:
./flow.tcl -design spm
```

#### PyPI (OpenLane 2)
```bash
pip install openlane
openlane --dockerized --smoke-test
```

### 1.3 OpenLane CLI Commands

#### Basic Flow Execution
```bash
# Run complete flow (OpenLane 2)
openlane --pdk-root /path/to/pdk config.json

# Run with Docker
openlane --dockerized --pdk-root /path/to/pdk config.json

# Smoke test
openlane --log-level ERROR --condensed --show-progress-bar --smoke-test

# Open results in KLayout
openlane --last-run --flow OpenInKLayout config.json

# Open results in Magic
openlane --last-run --flow OpenInMagic config.json
```

#### OpenLane 1 flow.tcl Commands
```bash
# Run autonomous flow
./flow.tcl -design <design_name>

# Interactive mode
./flow.tcl -interactive

# Initialize new design
./flow.tcl -design <design_name> -init_design_config

# Synthesis exploration
./flow.tcl -design <design_name> -synth_explore

# Specify PDK
./flow.tcl -design <design_name> -pdk sky130A
```

### 1.4 Interactive Mode Commands (OpenLane 1)

```tcl
# Must run first
package require openlane 0.9
prep -design <design_name>

# Individual steps
run_synthesis
run_floorplan
run_placement
run_cts
run_routing
run_magic
run_magic_spice_export
run_lvs
run_antenna_check
```

### 1.5 OpenLane 2 Flow Steps (Classic Flow)

The Classic flow executes these steps in sequence:

```
Yosys.JsonHeader → Yosys.Synthesis → Checker.YosysUnmappedCells →
Checker.YosysSynthChecks → OpenROAD.CheckSDCFiles → OpenROAD.Floorplan →
OpenROAD.TapEndcapInsertion → OpenROAD.GeneratePDN → OpenROAD.IOPlacement →
OpenROAD.GlobalPlacement → OpenROAD.DetailedPlacement →
OpenROAD.GlobalRouting → OpenROAD.DetailedRouting → OpenROAD.FillInsertion →
Magic.StreamOut → Magic.DRC → Magic.SpiceExtraction → Netgen.LVS
```

### 1.6 Key Configuration Variables

#### Synthesis
| Variable | Description | Default |
|----------|-------------|---------|
| `SYNTH_STRATEGY` | Optimization strategy (AREA/DELAY) | AREA 0 |
| `SYNTH_NO_FLAT` | Disable hierarchy flattening | 0 |
| `SYNTH_SIZING` | Enable gate sizing | 0 |
| `SYNTH_BUFFERING` | Enable buffering | 1 |
| `MAX_FANOUT_CONSTRAINT` | Maximum fanout | 10 |

#### Timing
| Variable | Description | Default |
|----------|-------------|---------|
| `CLOCK_PERIOD` | Clock period (ns) | 10 |
| `CLOCK_PORT` | Clock pin name | clk |
| `CLOCK_NET` | Clock net name | - |

#### Floorplan
| Variable | Description | Default |
|----------|-------------|---------|
| `FP_CORE_UTIL` | Core utilization (%) | 50 |
| `FP_ASPECT_RATIO` | Aspect ratio | 1 |
| `FP_SIZING` | absolute/relative | relative |
| `FP_PDN_CORE_RING` | Add power ring | 0 |

#### Placement
| Variable | Description | Default |
|----------|-------------|---------|
| `PL_TARGET_DENSITY` | Target density | 0.55 |
| `PL_ROUTABILITY_DRIVEN` | Enable routability | 1 |
| `PL_TIME_DRIVEN` | Enable timing-driven | 1 |

---

## 2. Yosys - Synthesis

Yosys is the open-source RTL synthesis suite that transforms Verilog to gate-level netlists.

**Official Resources:**
- GitHub: https://github.com/YosysHQ/yosys
- Documentation: https://yosyshq.readthedocs.io/projects/yosys/en/latest/
- Man page: https://www.mankier.com/1/yosys

### 2.1 CLI Options

```bash
# Basic usage
yosys [OPTIONS] [INFILES]

# Key options
-b, --backend <backend>      # Output backend (e.g., verilog, json)
-f, --frontend <frontend>    # Input frontend (e.g., verilog, liberty)
-s, --scriptfile <file>      # Execute script file
-p, --commands <cmds>        # Execute commands (semicolon-separated)
-c, --tcl-scriptfile <file>  # Execute TCL script
-r, --top <module>           # Specify top module
-m, --plugin <plugin>        # Load plugin module
-D, --define <name>[=val]    # Set Verilog define
-S, --synth                  # Run default synth command
-o, --outfile <file>         # Write design to file on exit
-q, --quiet                  # Quiet operation
-v, --verbose <level>        # Verbosity level (0-9)
-l, --logfile <file>         # Write log to file
-Q                           # Suppress banner
-T                           # Suppress footer
-V, --version                # Print version
```

### 2.2 Essential Synthesis Commands

#### Reading/Writing Files
```tcl
# Read Verilog
read_verilog design.v
read_verilog -defer design.v         # Defer elaboration
read_verilog -sv design.sv           # SystemVerilog

# Read Liberty library
read_liberty -lib cells.lib

# Write outputs
write_verilog synth.v                # Verilog netlist
write_json design.json               # JSON netlist
write_blif design.blif               # BLIF format
write_edif design.edif               # EDIF format
```

#### Hierarchy & Elaboration
```tcl
hierarchy -check -top <module>       # Elaborate with top module
hierarchy -auto-top                  # Auto-detect top
flatten                              # Flatten hierarchy
```

#### High-Level Transforms
```tcl
proc                                 # Process procedures
opt                                  # General optimization
fsm                                  # FSM extraction/optimization
memory                               # Memory inference
```

#### Technology Mapping
```tcl
techmap                              # Map to internal library
dfflibmap -liberty cells.lib         # Map flip-flops
abc -liberty cells.lib               # Map combinational logic
abc -lut 4                           # Map to 4-LUTs (FPGA)
abc9 -lut 4                          # Enhanced LUT mapping
```

#### Cleanup
```tcl
clean                                # Remove unused
opt_clean                            # Clean + optimize
```

### 2.3 Complete Synthesis Script Example

```tcl
# ASIC synthesis with Liberty library
read_verilog design.v
hierarchy -check -top top_module

# High-level synthesis
proc; opt; fsm; opt; memory; opt

# Technology mapping
techmap; opt
dfflibmap -liberty mycells.lib
abc -liberty mycells.lib

# Cleanup and output
clean
write_verilog synth.v
stat
```

### 2.4 Platform-Specific Synth Commands

```tcl
# Generic synthesis
synth -top <module>

# Specific targets
synth_xilinx -top <module>           # Xilinx FPGAs
synth_ice40 -top <module>            # Lattice iCE40
synth_ecp5 -top <module>             # Lattice ECP5
synth_gowin -top <module>            # Gowin FPGAs
synth_intel -top <module>            # Intel/Altera FPGAs

# Generic (for verification)
prep -top <module>                   # Coarse-grain only
```

### 2.5 ABC Options in Yosys

```tcl
# Basic ABC with Liberty
abc -liberty cells.lib

# With timing constraints
abc -liberty cells.lib -D 1000       # 1000 ps delay target
abc -liberty cells.lib -constr constraints.sdc

# LUT mapping (FPGA)
abc -lut 6                           # 6-input LUTs
abc9 -lut 4:6                        # Variable LUT sizes

# Options
abc -liberty cells.lib -dff          # Include flip-flops
abc -liberty cells.lib -keepff       # Keep FF outputs
abc -liberty cells.lib -nocleanup    # Debug: keep temp files
```

---

## 3. OpenROAD - Physical Design

OpenROAD is the unified application for physical design, handling floorplanning through routing.

**Official Resources:**
- GitHub: https://github.com/The-OpenROAD-Project/OpenROAD
- Documentation: https://openroad.readthedocs.io/en/latest/
- Flow Scripts: https://github.com/The-OpenROAD-Project/OpenROAD-flow-scripts

### 3.1 CLI Options

```bash
# Basic usage
openroad [OPTIONS] [script.tcl]

# Key options
-help                    # Show help
-version                 # Show version
-no_init                 # Skip .openroad init file
-no_splash               # Suppress startup message
-exit                    # Exit after script
-gui                     # Launch GUI
-log <file>              # Log file
-metrics <file>          # Metrics output file
```

### 3.2 Floorplan Commands

```tcl
# Initialize floorplan
initialize_floorplan \
  -die_area {0 0 1000 1000} \
  -core_area {100 100 900 900} \
  -site unithd

# Or by utilization
initialize_floorplan \
  -utilization 50 \
  -aspect_ratio 1.0 \
  -core_space 10 \
  -site unithd

# Add routing tracks
make_tracks metal1 -x_offset 0.17 -x_pitch 0.34 -y_offset 0.17 -y_pitch 0.34
make_tracks metal2 -x_offset 0.23 -x_pitch 0.46 -y_offset 0.23 -y_pitch 0.46

# Tapcell and endcap insertion
tapcell \
  -distance 14 \
  -tapcell_master TAPCELL_X1 \
  -endcap_master ENDCAP_X1

# Power distribution network
pdngen pdn.cfg
```

### 3.3 Placement Commands

```tcl
# Global placement
global_placement \
  -density 0.6 \
  -pad_left 2 \
  -pad_right 2

# IO placement
place_pins \
  -hor_layers metal3 \
  -ver_layers metal2

# Detailed placement
detailed_placement \
  -max_displacement {100 100}

# Filler cell insertion
filler_placement [list FILL1 FILL2 FILL4]

# Remove fillers (if needed)
remove_fillers

# Optimize placement for timing
repair_design
repair_timing
```

### 3.4 Clock Tree Synthesis (CTS)

```tcl
# Configure CTS
set_wire_rc -clock -layer metal3

# Run CTS
clock_tree_synthesis \
  -root_buf CLKBUF_X3 \
  -buf_list {CLKBUF_X1 CLKBUF_X2 CLKBUF_X3} \
  -wire_unit 20

# Repair clock tree
repair_clock_nets
repair_clock_inverters
```

### 3.5 Routing Commands

```tcl
# Set routing layers
set_routing_layers -signal metal1-metal5 -clock metal3-metal5

# Global routing
global_route \
  -guide_file route.guide \
  -congestion_iterations 50 \
  -verbose

# Detailed routing (TritonRoute)
detailed_route \
  -guide route.guide \
  -output_drc drc.rpt \
  -output_maze maze.log \
  -verbose 1

# Antenna repair
repair_antennas -iterations 5

# Fill insertion
density_fill \
  -rules fill_rules.json
```

### 3.6 Analysis & Reports

```tcl
# Timing reports
report_checks -path_delay max
report_checks -path_delay min
report_tns
report_wns

# Power reports
report_power

# Design statistics
report_design_area
report_cell_usage

# DRC
check_placement -verbose
check_routing
```

### 3.7 File I/O

```tcl
# Read files
read_lef tech.lef
read_lef cells.lef
read_def design.def
read_liberty cells.lib
read_sdc constraints.sdc
read_verilog design.v
link_design top_module

# Write files
write_def output.def
write_verilog output.v
write_db design.odb
```

---

## 4. Magic VLSI - Layout

Magic is the open-source VLSI layout editor with DRC, extraction, and LVS capabilities.

**Official Resources:**
- Website: http://opencircuitdesign.com/magic/
- GitHub: https://github.com/RTimothyEdwards/magic
- Command Reference: http://opencircuitdesign.com/magic/userguide.html
- Tutorials: http://opencircuitdesign.com/magic/tutorials/

### 4.1 CLI Options

```bash
# Basic usage
magic [OPTIONS] [cellname]

# Key options
-noconsole               # No console window
-dnull                   # No graphics (batch mode)
-T <techfile>            # Specify technology file
-rcfile <file>           # Use specific startup file
-norcfile                # Skip .magicrc
-d <display>             # Graphics driver (X11, OGL, NULL)
```

### 4.2 File Operations

```tcl
# Load/save cells
load cellname
save cellname
writeall                 # Save all modified cells

# Read GDSII
gds read design.gds
gds readonly true
gds flatten true

# Write GDSII
gds write design.gds

# Read/write CIF
cif read design.cif
cif write design.cif

# Read DEF/LEF
lef read cells.lef
def read design.def
def write design.def
```

### 4.3 Layout Commands

```tcl
# Box operations
box 0 0 100 100           # Set box coordinates
box width 50              # Set box width
box height 50             # Set box height
box move right 10         # Move box

# Paint/erase
paint metal1              # Paint layer in box
paint m1                  # Shorthand
erase metal1              # Erase layer in box
erase *                   # Erase all layers

# Wire tool
wire type metal1
wire width 0.5
wire horizontal
wire vertical

# Polygon
polygon metal1 0 0 10 0 10 10 5 15 0 10

# Labels
label "signal_name" center metal1
port make input
port class input
```

### 4.4 Selection Commands

```tcl
# Select operations
select area              # Select in box
select cell cellname     # Select cell instances
select clear             # Clear selection
select top cell          # Select top cell
select visible           # Select visible layers

# Move/copy
move right 10
move up 5
copy
```

### 4.5 DRC (Design Rule Checking)

```tcl
# Run DRC
drc check                # Check current cell
drc catchup              # Complete DRC on cell
drc find                 # Find next error
drc find [nth]           # Find nth error
drc why                  # Explain error in box
drc count                # Count errors

# DRC settings
drc on                   # Enable continuous DRC
drc off                  # Disable continuous DRC
drc style drc(full)      # Full DRC
drc euclidean on         # Euclidean distance checks
```

### 4.6 Extraction & SPICE

```tcl
# Extract parasitic netlist
extract all              # Extract all cells
extract unique           # Extract with unique names
extract no all           # Clear extraction

# Generate SPICE
ext2spice lvs            # Setup for LVS
ext2spice                # Generate SPICE file

# Options
ext2spice cthresh 0.01   # Capacitance threshold
ext2spice rthresh 1      # Resistance threshold
ext2spice hierarchy on   # Hierarchical extraction
ext2spice subcircuit top on  # Top cell as subckt

# Resistor extraction
extresist all
extresist tolerance 10

# Combined flow
extract all
ext2spice lvs
ext2spice -o design.spice
```

### 4.7 Batch Mode Script Example

```bash
#!/bin/bash
magic -dnull -noconsole << EOF
tech load sky130A
gds read design.gds
load topcell
select top cell
extract all
ext2spice lvs
ext2spice -o design.spice
quit
EOF
```

### 4.8 Common Command Reference

| Command | Description |
|---------|-------------|
| `help` | List all commands |
| `help <cmd>` | Help for specific command |
| `tech load <name>` | Load technology |
| `cellname list children` | List sub-cells |
| `cellname list parents` | List parent cells |
| `property` | View/set cell properties |
| `cif ostyle` | Set CIF output style |
| `gds ordering on` | Preserve GDS ordering |

---

## 5. KLayout - GDSII Viewer/Editor

KLayout is a high-performance layout viewer and editor supporting GDSII, OASIS, and other formats.

**Official Resources:**
- Website: https://www.klayout.de/
- Command Args: https://www.klayout.de/command_args.html
- Documentation: https://www.klayout.de/doc.html

### 5.1 CLI Options

```bash
# Basic usage
klayout [OPTIONS] [files...]

# Key options
-b                       # Batch mode (no GUI)
-zz                      # Non-GUI mode (no display)
-e                       # Edit mode
-ne                      # Non-edit mode (view only)
-r <script>              # Run script and exit
-rm <script>             # Run script then continue
-rd <var>=<value>        # Define variable for script
-l <file>                # Layer properties file
-u <file>                # Session file
-s                       # Sync with other instance
-p <plugin>              # Load plugin
-j <threads>             # Number of threads
-t                       # Enable undo/redo
-nn <tech>               # Technology name
-n <tech>                # Technology file
```

### 5.2 Batch Mode Operations

```bash
# Run DRC in batch mode
klayout -b -r drc_rules.drc

# With variables
klayout -b \
  -rd input=design.gds \
  -rd report=drc_report.lyrdb \
  -r my_drc.drc

# Convert formats
klayout -b -r convert.rb \
  -rd input=design.gds \
  -rd output=design.oas

# Run LVS
klayout -b \
  -rd schematic=design.spice \
  -rd layout=design.gds \
  -rd report=lvs_report.lvsdb \
  -r lvs_rules.lvs
```

### 5.3 DRC Script Example (.drc)

```ruby
# DRC script (Ruby-based DSL)
source($input)  # Variable from -rd input=...
report($report) # Variable from -rd report=...

# Define layers
metal1 = input(68, 20)
metal2 = input(69, 20)
via1 = input(68, 44)

# DRC rules
metal1.width(0.14).output("M1 width < 0.14um")
metal1.space(0.14).output("M1 space < 0.14um")
metal2.width(0.14).output("M2 width < 0.14um")
metal2.space(0.14).output("M2 space < 0.14um")

# Enclosure
metal1.enclosing(via1, 0.03).output("M1 via enclosure")
metal2.enclosing(via1, 0.03).output("M2 via enclosure")
```

### 5.4 LVS Script Example (.lvs)

```ruby
# LVS script (Ruby-based DSL)
deep

# Source layout and schematic
source($layout)
schematic($schematic)

# Define device recognition layers
nwell = input(64, 20)
diff = input(65, 20)
poly = input(66, 20)
nsdm = input(93, 44)
psdm = input(94, 20)

# Define devices
nmos = nsdm & diff
pmos = psdm & diff

# Extract and compare
extract
compare
report($report)
```

### 5.5 Python/Ruby Scripting

```python
# Python script for KLayout
import pya

# Load layout
layout = pya.Layout()
layout.read("design.gds")

# Access top cell
top_cell = layout.top_cell()

# Iterate shapes
layer = layout.layer(68, 20)
for shape in top_cell.shapes(layer).each():
    print(shape.bbox)

# Save
layout.write("output.gds")
```

```ruby
# Ruby script for KLayout
layout = RBA::Layout.new
layout.read("design.gds")

top_cell = layout.top_cell
layer = layout.layer(68, 20)

top_cell.shapes(layer).each do |shape|
  puts shape.bbox.to_s
end

layout.write("output.gds")
```

### 5.6 Environment Variables

| Variable | Description |
|----------|-------------|
| `KLAYOUT_PATH` | Search paths (: separated on Linux) |
| `KLAYOUT_HOME` | Home directory for config |

---

## 6. Netgen - LVS

Netgen is the LVS (Layout vs. Schematic) verification tool.

**Official Resources:**
- Website: http://opencircuitdesign.com/netgen/
- GitHub: https://github.com/RTimothyEdwards/netgen
- Reference: http://opencircuitdesign.com/netgen/reference.html
- Tutorial: http://opencircuitdesign.com/netgen/tutorial/tutorial.html

### 6.1 CLI Options

```bash
# Basic usage
netgen [OPTIONS] [script.tcl]

# Key options
-noconsole               # No console
-batch                   # Batch mode
-log <file>              # Log file
```

### 6.2 LVS Commands

```tcl
# Simple LVS comparison
lvs "layout.spice subckt" "schematic.spice subckt" setup.tcl output.txt

# Full form
lvs layout.spice schematic.spice sky130A_setup.tcl lvs_results.log

# If setup.tcl doesn't exist, uses defaults
lvs layout.spice schematic.spice

# Read netlists separately
readnet spice layout.spice
readnet spice schematic.spice
# Then compare
compare
```

### 6.3 Common Commands

```tcl
# Read netlists
readnet spice design.spice
readnet verilog design.v

# Setup comparison
equate classes                # Equate device classes
equate pins                   # Equate pin names
property                      # Check properties
ignore                        # Ignore specific elements

# Run comparison
compare
run converge                  # Iterate until stable

# Reports
summary                       # Print summary
nodes                         # Print node info
elements                      # Print element info
print                         # Print full report
```

### 6.4 LVS Script Example

```tcl
#!/usr/bin/env netgen -batch source

# Load PDK setup
source /path/to/sky130A_setup.tcl

# Read netlists
readnet spice extracted.spice
readnet spice schematic.spice

# Run LVS
lvs "extracted.spice topcell" "schematic.spice topcell" \
    /path/to/setup.tcl \
    lvs_output.log

# Check results
if {[info exists lvs_result]} {
    if {$lvs_result == 0} {
        puts "LVS CLEAN"
    } else {
        puts "LVS ERRORS: $lvs_result"
    }
}
```

### 6.5 PDK Setup Files

Each PDK provides a setup file (e.g., `sky130A_setup.tcl`):

```tcl
# Example setup file content
permute default
property default
equate class {nfet_01v8 nfet_01v8_lvt}
equate class {pfet_01v8 pfet_01v8_hvt}
```

---

## 7. OpenSTA - Timing Analysis

OpenSTA is the static timing analysis engine used in OpenROAD.

**Official Resources:**
- GitHub: https://github.com/The-OpenROAD-Project/OpenSTA
- Manual: https://github.com/parallaxsw/OpenSTA/blob/master/doc/OpenSTA.pdf

### 7.1 CLI Options

```bash
# Basic usage
sta [OPTIONS] [script.tcl]

# Runs in TCL interpreter mode
```

### 7.2 Setup Commands

```tcl
# Read libraries
read_liberty -corner fast fast.lib
read_liberty -corner slow slow.lib
read_liberty cells.lib

# Read design
read_verilog design.v
link_design top_module

# Read constraints
read_sdc design.sdc

# Read parasitics
read_spef design.spef
# Or SDF
read_sdf design.sdf
```

### 7.3 SDC Constraint Commands

```tcl
# Create clock
create_clock -name clk -period 10.0 [get_ports clk]
create_clock -name clk -period 10 -waveform {0 5} [get_ports clk]

# Generated clocks
create_generated_clock -name clk_div2 \
    -source [get_ports clk] \
    -divide_by 2 \
    [get_pins divider/Q]

# Input/output delays
set_input_delay -clock clk 2.0 [get_ports {data_in[*]}]
set_output_delay -clock clk 2.0 [get_ports {data_out[*]}]

# Input transition
set_input_transition 0.5 [get_ports data_in]

# Output load
set_load 0.1 [get_ports data_out]

# Clock uncertainty
set_clock_uncertainty -setup 0.5 [get_clocks clk]
set_clock_uncertainty -hold 0.3 [get_clocks clk]

# False paths
set_false_path -from [get_clocks clk1] -to [get_clocks clk2]

# Multi-cycle paths
set_multicycle_path 2 -setup -from [get_pins reg1/Q] -to [get_pins reg2/D]

# Max delay
set_max_delay 5.0 -from [get_ports in] -to [get_ports out]
```

### 7.4 Reporting Commands

```tcl
# Timing reports
report_checks                     # All timing checks
report_checks -path_delay max     # Setup (max) paths
report_checks -path_delay min     # Hold (min) paths
report_checks -to [get_pins reg/D]  # To specific pin
report_checks -through [get_nets net1]  # Through net
report_checks -group_path_count 10  # Top 10 paths
report_checks -digits 4           # 4 decimal places

# Slack reports
report_tns                        # Total negative slack
report_wns                        # Worst negative slack

# Power analysis
report_power                      # Power report
report_power -instances           # Per-instance power

# Clock reports
report_clocks                     # Clock summary
report_clock_skew                 # Clock skew

# Design info
report_design                     # Design summary
report_units                      # Unit definitions
```

### 7.5 Multi-Corner Analysis

```tcl
# Define corners
define_corners slow fast

# Read libraries per corner
read_liberty -corner slow slow.lib
read_liberty -corner fast fast.lib

# Apply derating
set_timing_derate -early 0.95 -corner slow
set_timing_derate -late 1.05 -corner slow

# Report per corner
report_checks -corner slow
report_checks -corner fast
```

### 7.6 Complete STA Script Example

```tcl
# Read libraries
read_liberty cells.lib

# Read design
read_verilog synth.v
link_design top

# Read constraints
read_sdc constraints.sdc

# Read parasitics
read_spef design.spef

# Report timing
report_checks -path_delay max -format full_clock_expanded > setup.rpt
report_checks -path_delay min -format full_clock_expanded > hold.rpt
report_tns > tns.rpt
report_wns > wns.rpt

# Exit
exit
```

---

## 8. PDK Installation

### 8.1 SkyWater SKY130 PDK

**Official Resources:**
- GitHub: https://github.com/google/skywater-pdk
- Open_PDKs: https://github.com/RTimothyEdwards/open_pdks
- Documentation: https://skywater-pdk.readthedocs.io/

#### Installation via open_pdks

```bash
# Clone open_pdks
git clone https://github.com/RTimothyEdwards/open_pdks.git
cd open_pdks

# Configure for SKY130
./configure \
    --prefix=/usr \
    --enable-sky130-pdk \
    --enable-sram-sky130

# Build and install
make
sudo make install

# Set environment
export PDK_ROOT=/usr/share/pdk
export PDK=sky130A
```

#### Minimal Installation (Analog Only)

```bash
./configure \
    --enable-sky130-pdk \
    --enable-sram-sky130 \
    --disable-sc-hs-sky130 \
    --disable-sc-ms-sky130 \
    --disable-sc-ls-sky130 \
    --disable-sc-lp-sky130 \
    --disable-sc-hd-sky130 \
    --disable-sc-hdll-sky130 \
    --disable-sc-hvl-sky130
make
sudo make install
```

#### PDK Variants

| Variant | Description |
|---------|-------------|
| sky130A | Standard digital (most common) |
| sky130B | With ReRAM option |

### 8.2 GlobalFoundries GF180MCU PDK

**Official Resources:**
- GitHub: https://github.com/google/gf180mcu-pdk
- Open_PDKs support included

#### Installation

```bash
# Clone open_pdks
git clone https://github.com/RTimothyEdwards/open_pdks.git
cd open_pdks

# Configure for GF180MCU
./configure \
    --prefix=/usr \
    --enable-gf180mcu-pdk

# Build and install
make
sudo make install

# Set environment
export PDK_ROOT=/usr/share/pdk
export PDK=gf180mcuD
```

#### GF180MCU Variants

| Variant | Metal Stack | Description |
|---------|-------------|-------------|
| gf180mcuA | 3 metal | Basic |
| gf180mcuB | 4 metal | Standard |
| gf180mcuC | 5 metal | 0.9um thick top metal |
| gf180mcuD | 5 metal | 1.1um thick top metal (shuttles) |

### 8.3 Conda Installation (Alternative)

```bash
# Sky130
conda install -c litex-hub open_pdks.sky130A

# GF180MCU
conda install -c litex-hub open_pdks.gf180mcuC
```

---

## 9. Quick Reference Tables

### 9.1 File Formats

| Format | Extension | Tool | Purpose |
|--------|-----------|------|---------|
| Verilog | .v | Yosys, OpenROAD | RTL/netlist |
| Liberty | .lib | Yosys, OpenSTA | Timing library |
| LEF | .lef | OpenROAD, Magic | Cell abstracts |
| DEF | .def | OpenROAD, Magic | Placement/routing |
| GDSII | .gds | Magic, KLayout | Physical layout |
| OASIS | .oas | KLayout | Compressed layout |
| SPICE | .spice, .sp | Netgen, Magic | Circuit netlist |
| SDC | .sdc | OpenSTA, OpenROAD | Timing constraints |
| SPEF | .spef | OpenSTA | Parasitics |
| SDF | .sdf | OpenSTA | Delay file |

### 9.2 Common CLI Patterns

| Task | OpenLane 2 | OpenLane 1 |
|------|------------|------------|
| Run flow | `openlane config.json` | `./flow.tcl -design name` |
| Interactive | `openlane -i config.json` | `./flow.tcl -interactive` |
| Smoke test | `openlane --smoke-test` | N/A |
| Docker | `openlane --dockerized` | `make mount` |

### 9.3 OpenLane Directory Structure

```
designs/
└── mydesign/
    ├── config.json          # OpenLane 2 config
    ├── config.tcl           # OpenLane 1 config
    ├── src/
    │   └── design.v         # RTL source
    └── runs/
        └── RUN_*/
            ├── logs/        # Tool logs
            ├── reports/     # Analysis reports
            ├── results/     # Output files
            └── tmp/         # Intermediate files
```

### 9.4 Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `PDK_ROOT` | PDK installation path | `/usr/share/pdk` |
| `PDK` | Active PDK variant | `sky130A` |
| `STD_CELL_LIBRARY` | Standard cell library | `sky130_fd_sc_hd` |
| `OPENLANE_ROOT` | OpenLane installation | `~/OpenLane` |
| `KLAYOUT_PATH` | KLayout search paths | `~/.klayout` |

---

## Sources

- [OpenLane 2 Documentation](https://openlane2.readthedocs.io/)
- [OpenLane GitHub](https://github.com/The-OpenROAD-Project/OpenLane)
- [Yosys Documentation](https://yosyshq.readthedocs.io/projects/yosys/en/latest/)
- [Yosys GitHub](https://github.com/YosysHQ/yosys)
- [OpenROAD Documentation](https://openroad.readthedocs.io/en/latest/)
- [OpenROAD Flow Scripts](https://openroad-flow-scripts.readthedocs.io/en/latest/)
- [Magic VLSI](http://opencircuitdesign.com/magic/)
- [Magic GitHub](https://github.com/RTimothyEdwards/magic)
- [KLayout](https://www.klayout.de/)
- [Netgen](http://opencircuitdesign.com/netgen/)
- [Netgen GitHub](https://github.com/RTimothyEdwards/netgen)
- [OpenSTA GitHub](https://github.com/The-OpenROAD-Project/OpenSTA)
- [SkyWater PDK](https://github.com/google/skywater-pdk)
- [Open_PDKs](https://github.com/RTimothyEdwards/open_pdks)
- [GF180MCU PDK](https://github.com/google/gf180mcu-pdk)
- [OpenLane WOSET Paper](https://woset-workshop.github.io/PDFs/2020/a21.pdf)
- [TritonRoute](https://github.com/The-OpenROAD-Project/TritonRoute)
- [ABC Toolbox](https://yosyshq.readthedocs.io/projects/yosys/en/latest/using_yosys/synthesis/abc.html)

---

*Document generated from web research on 2026-01-26*
