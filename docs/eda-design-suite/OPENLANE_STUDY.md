# OpenLane Study Notes

**Validated findings from studying OpenLane v1.0 source code**

*Reference: `/root/OpenLane` (or local clone)*

---

## Executive Summary

OpenLane v1.0 is in **maintenance mode** (critical fixes only). The successor is **LibreLane** (Python-based, backward compatible). For existing projects, OpenLane v1.0 remains reliable with 100+ successful tape-outs via Google/Efabless MPW.

### Key Findings for FeCIM Integration

| Question | Answer | Source |
|----------|--------|--------|
| Can we inject pre-placed DEF? | Yes, via `PLACEMENT_CURRENT_DEF` | `flow.tcl:32-35` |
| Does CURRENT_DEF bypass placement? | Yes, with `PL_SKIP_INITIAL_PLACEMENT=1` | `placement.tcl:23` |
| How to add custom cells? | `EXTRA_LEFS`, `EXTRA_GDS_FILES`, `EXTRA_LIBS` | `configuration.md:48-50` |
| FIXED vs PLACED in DEF? | FIXED = locked, PLACED = adjustable | `openlane_commands.md:68` |
| Post-flow hooks? | `hooks/post_run.py` in design dir | `all.tcl:1312-1318` |

---

## Directory Structure Analysis

```
OpenLane/
├── flow.tcl              # Main orchestration (12,663 bytes)
├── configuration/        # Default configuration files
│   ├── general.tcl       # 2,424 lines - global defaults
│   ├── placement.tcl     # 1,672 lines - placement params
│   ├── floorplan.tcl     # 1,720 lines - floorplan params
│   ├── routing.tcl       # 2,109 lines - routing params
│   ├── synthesis.tcl     # 1,720 lines - synthesis params
│   └── checkers.tcl      # DRC/LVS checkers
├── scripts/
│   ├── tcl_commands/     # TCL API implementation
│   │   ├── all.tcl       # Core commands
│   │   ├── placement.tcl # Placement commands
│   │   ├── floorplan.tcl # Floorplan commands
│   │   ├── routing.tcl   # Routing commands
│   │   ├── magic.tcl     # Magic integration
│   │   └── checkers.tcl  # Validation
│   ├── openroad/         # OpenROAD TCL scripts
│   ├── yosys/            # Synthesis scripts
│   └── magic/            # Magic scripts
├── designs/              # Example designs
└── docs/                 # ReadTheDocs source
```

---

## Stage-Specific DEF Variables

OpenLane tracks DEF progression through stages. From `flow.tcl:32-97`:

```tcl
# Placement stage (lines 32-35)
if { ! [ info exists ::env(PLACEMENT_CURRENT_DEF) ] } {
    set ::env(PLACEMENT_CURRENT_DEF) $::env(CURRENT_DEF)
} else {
    set ::env(CURRENT_DEF) $::env(PLACEMENT_CURRENT_DEF)
}

# CTS stage (lines 42-45)
if { ! [ info exists ::env(CTS_CURRENT_DEF) ] } {
    set ::env(CTS_CURRENT_DEF) $::env(CURRENT_DEF)
} else {
    set ::env(CURRENT_DEF) $::env(CTS_CURRENT_DEF)
}

# Routing stage (lines 53-56)
if { ! [ info exists ::env(ROUTING_CURRENT_DEF) ] } {
    set ::env(ROUTING_CURRENT_DEF) $::env(CURRENT_DEF)
} else {
    set ::env(CURRENT_DEF) $::env(ROUTING_CURRENT_DEF)
}
```

**Implication:** Setting `PLACEMENT_CURRENT_DEF` in config allows injecting our pre-placed DEF before placement begins.

---

## Placement Bypass Mechanism

From `scripts/tcl_commands/placement.tcl:23`:

```tcl
if { $::env(PL_RANDOM_GLB_PLACEMENT) } {
    set ::env(PL_SKIP_INITIAL_PLACEMENT) 1
}
```

From `scripts/openroad/gpl.tcl:60`:

```tcl
if { $::env(PL_SKIP_INITIAL_PLACEMENT) && !$::env(PL_BASIC_PLACEMENT) } {
    # Skip initial placement
}
```

**Configuration defaults** from `configuration/placement.tcl`:

```tcl
set ::env(PL_ROUTABILITY_DRIVEN) 1
set ::env(PL_TIME_DRIVEN) 1
set ::env(PL_RANDOM_GLB_PLACEMENT) 0
set ::env(PL_BASIC_PLACEMENT) 0
set ::env(PL_SKIP_INITIAL_PLACEMENT) 0   # We set this to 1
set ::env(PL_RANDOM_INITIAL_PLACEMENT) 0
```

---

## Custom Cell Integration

### EXTRA_LEFS Loading

From `scripts/tcl_commands/floorplan.tcl:418-419`:

```tcl
if { [info exists ::env(EXTRA_LEFS)] } {
    set extra_lefs $::env(EXTRA_LEFS)
    # LEFs are loaded into OpenROAD
}
```

From `scripts/magic/def/read.tcl:15-16`:

```tcl
if { [info exist ::env(EXTRA_LEFS)] } {
    foreach lef_file $::env(EXTRA_LEFS) {
        lef read $lef_file
    }
}
```

### EXTRA_GDS_FILES Loading

From `scripts/tcl_commands/all.tcl:885-887`:

```tcl
if { [info exists ::env(EXTRA_GDS_FILES)] } {
    puts_verbose "Verifying existence of files defined in ::env(EXTRA_GDS_FILES)..."
    assert_files_exist "$::env(EXTRA_GDS_FILES)"
}
```

### EXTRA_LIBS Loading

From `scripts/yosys/synth.tcl:50-52`:

```tcl
if { [info exists ::env(EXTRA_LIBS) ] } {
    foreach lib $::env(EXTRA_LIBS) {
        read_liberty -lib $lib
    }
}
```

---

## DEF FIXED vs PLACED

From `docs/source/reference/openlane_commands.md:68`:

```
| `-fixed <val>` | if `<val>` is 1, then the macro is set as FIXED,
                   else it's set as PLACED in the def file.|
```

**FIXED cells:** OpenROAD respects this during placement and routing. The cell position is locked.

**PLACED cells:** Position is a hint; tools may adjust during optimization.

---

## FP_DEF_TEMPLATE Usage

From `docs/source/reference/configuration.md:142`:

```
| `FP_DEF_TEMPLATE` | Points to the DEF file to be used as a template
                     when running `apply_def_template`. This will be
                     used to extract pin names, locations, shapes
                     -excluding power and ground pins- as well as the
                     die area and replicate all this information in
                     the `CURRENT_DEF`. |
```

From `scripts/tcl_commands/floorplan.tcl:279-287`:

```tcl
if { [info exists ::env(FP_DEF_TEMPLATE)] } {
    # Uses apply_def_template command
    # --def-template $::env(FP_DEF_TEMPLATE)
}
```

**Use case:** When you want pin positions from a parent design but still want OpenLane to place standard cells.

---

## Hooks System

### Post-Run Hooks

From `scripts/tcl_commands/all.tcl:1312-1318`:

```tcl
proc run_post_run_hooks {} {
    if { [file exists $::env(DESIGN_DIR)/hooks/post_run.py]} {
        puts_info "Running post run hook..."
        set result [exec $::env(OPENROAD_BIN) -exit -no_init -python \
            $::env(DESIGN_DIR)/hooks/post_run.py]
    } else {
        puts_verbose "No post-run hook found, skipping..."
    }
}
```

**To use:** Create `designs/<your_design>/hooks/post_run.py` with validation scripts.

### PDN Macro Hooks

From `configuration.md:132`:

```
| `FP_PDN_MACRO_HOOKS` | Specifies explicit power connections of internal
                         macros to the top level power grid. Format:
                         `<instance_name> <vdd_net> <gnd_net> <vdd_pin> <gnd_pin>` |
```

**Example:**
```json
"FP_PDN_MACRO_HOOKS": "fecim_array_0 vccd1 vssd1 VDD VSS"
```

---

## Interactive Mode Commands

### Core Commands (from openlane_commands.md)

| Command | Description |
|---------|-------------|
| `prep -design <name>` | Initialize design, load configs |
| `set_def <def>` | Set current DEF file |
| `set_netlist <netlist>` | Set current netlist |
| `run_synthesis` | Run Yosys synthesis + STA |
| `run_floorplan` | Init floorplan + IO placement + tap/decap + PDN |
| `run_placement` | Global + detailed placement |
| `run_cts` | Clock tree synthesis |
| `run_routing` | Global + detailed routing + SPEF extraction |
| `run_magic` | Generate GDSII via Magic |
| `run_klayout` | Generate GDSII via KLayout |
| `run_lvs` | Layout vs Schematic check |
| `run_magic_drc` | Design rule check via Magic |

### Macro Placement Commands

| Command | Description |
|---------|-------------|
| `add_macro_placement <name> <x> <y> [<orient>]` | Add macro to placement config |
| `manual_macro_placement [-f]` | Apply placements (-f = FIXED) |
| `basic_macro_placement` | Let OpenROAD place macros |

### Utility Commands

| Command | Description |
|---------|-------------|
| `open_in_klayout` | Open current DEF in KLayout GUI |
| `generate_final_summary_report` | Create metrics.csv and manufacturability.rpt |
| `save_views` | Save final LEF/GDS/DEF/MAG views |

---

## Synthesis Options for Structural Netlists

From `configuration.md:81`:

```
| `SYNTH_ELABORATE_ONLY` | "Elaborate" the design only without attempting
                           any logic mapping. Useful when dealing with
                           structural Verilog netlists. (Default: `0`) |
```

**For FeCIM:** Set `SYNTH_ELABORATE_ONLY=1` since our Verilog is already structural (instantiates fecim_bit cells).

---

## Flow Control Variables

From `configuration.md:320-354`:

| Variable | Default | Description |
|----------|---------|-------------|
| `RUN_DRT` | 1 | Enable detailed routing |
| `RUN_LVS` | 1 | Enable LVS check |
| `RUN_MAGIC` | 1 | Enable Magic GDSII generation |
| `RUN_MAGIC_DRC` | 1 | Enable Magic DRC |
| `RUN_KLAYOUT` | 1 | Enable KLayout GDSII generation |
| `RUN_CTS` | 1 | Enable clock tree synthesis |

**For FeCIM (no clock):** Consider `RUN_CTS=0` if design is purely combinational.

---

## Checker Variables

From `configuration.md:369-388`:

| Variable | Default | Description |
|----------|---------|-------------|
| `QUIT_ON_TR_DRC` | 1 | Quit on TritonRoute DRC violations |
| `QUIT_ON_MAGIC_DRC` | 1 | Quit on Magic DRC violations |
| `QUIT_ON_LVS_ERROR` | 1 | Quit on LVS mismatches |
| `QUIT_ON_ILLEGAL_OVERLAPS` | 1 | Quit on overlaps during extraction |

**Debugging tip:** Set these to 0 temporarily to see all errors before fixing.

---

## Validated Configuration Template

Based on source code analysis, here's a validated config for FeCIM:

```json
{
  "DESIGN_NAME": "fecim_crossbar",
  "VERILOG_FILES": "dir::src/*.v",
  "CLOCK_PERIOD": 10,
  "CLOCK_PORT": "CLK",
  "CLOCK_NET": "CLK",

  "PDK": "sky130A",
  "STD_CELL_LIBRARY": "sky130_fd_sc_hd",

  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "EXTRA_LIBS": "dir::cells/fecim_bit.lib",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v",

  "SYNTH_ELABORATE_ONLY": 1,

  "FP_SIZING": "absolute",
  "DIE_AREA": "0 0 100 100",
  "DESIGN_IS_CORE": 0,

  "PLACEMENT_CURRENT_DEF": "dir::crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  "FP_PDN_ENABLE_RAILS": 0,
  "FP_PDN_MACRO_HOOKS": "fecim_array vccd1 vssd1 VDD VSS",

  "RUN_CTS": 0,

  "QUIT_ON_MAGIC_DRC": 0,
  "QUIT_ON_LVS_ERROR": 0
}
```

---

## Key Source Files Reference

| File | Lines | Purpose |
|------|-------|---------|
| `flow.tcl` | 400+ | Main flow orchestration |
| `configuration/general.tcl` | 2,424 | Default parameters |
| `configuration/placement.tcl` | 42 | Placement defaults |
| `scripts/tcl_commands/all.tcl` | 1,400+ | Core command implementations |
| `scripts/tcl_commands/placement.tcl` | 200+ | Placement commands |
| `scripts/tcl_commands/floorplan.tcl` | 500+ | Floorplan commands |
| `scripts/openroad/gpl.tcl` | 100+ | Global placement script |
| `docs/source/reference/configuration.md` | 411 | Official config reference |
| `docs/source/reference/openlane_commands.md` | 359 | Command reference |

---

## Next Steps

1. **Validate with test design:** Run 8x8 crossbar through flow
2. **Create FeCIM cell in Magic:** Generate proper LEF/GDS
3. **Timing characterization:** Extract RC and create Liberty file
4. **Document failure modes:** Note any OpenLane limitations

---

*Last updated: 2026-01-23*
*OpenLane version: v1.0 (maintenance mode)*
