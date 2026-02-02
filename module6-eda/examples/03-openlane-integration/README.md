# Example 03: OpenLane Integration

Complete workflow for integrating a FeCIM crossbar macro with OpenLane RTL-to-GDSII flow.

## Overview

This example demonstrates the full path from neural network weights to GDSII-ready design files using OpenLane v1.0.

**Prerequisites:**
- OpenLane v1.0 installed at `~/OpenLane`
- SKY130 PDK configured
- Docker (for containerized OpenLane) or native installation

## Directory Structure

```
03-openlane-integration/
├── README.md              # This file
├── weights.json           # 16x16 test weights
├── cells/
│   ├── fecim_bit.lef      # Cell abstract (stub)
│   ├── fecim_bit.lib      # Timing model (stub)
│   └── fecim_bit.v        # Behavioral model
├── config.json            # OpenLane configuration
├── hooks/
│   └── post_run.py        # Validation script
├── run_compile.sh         # Step 1: Generate files
└── run_openlane.sh        # Step 2: Run OpenLane
```

## Step-by-Step Workflow

### Step 1: Compile Weights

```bash
cd module6-eda

# Generate Verilog and DEF from weights
./examples/03-openlane-integration/run_compile.sh
```

This creates:
- `output/crossbar.v` - Structural Verilog
- `output/crossbar.def` - Pre-placed cells (FIXED)

### Step 2: Prepare OpenLane Design

```bash
# Create design directory in OpenLane
mkdir -p ~/OpenLane/designs/fecim_crossbar

# Copy generated files
cp examples/03-openlane-integration/output/* ~/OpenLane/designs/fecim_crossbar/src/
cp examples/03-openlane-integration/cells/* ~/OpenLane/designs/fecim_crossbar/cells/
cp examples/03-openlane-integration/config.json ~/OpenLane/designs/fecim_crossbar/
cp -r examples/03-openlane-integration/hooks ~/OpenLane/designs/fecim_crossbar/
```

### Step 3: Run OpenLane

```bash
cd ~/OpenLane

# Docker-based execution
make mount
./flow.tcl -design fecim_crossbar -tag v1

# Or native execution
./flow.tcl -design fecim_crossbar -tag v1
```

### Step 4: View Results

```bash
# Check reports
cat designs/fecim_crossbar/runs/v1/reports/metrics.csv

# View layout in KLayout
klayout designs/fecim_crossbar/runs/v1/results/final/gds/fecim_crossbar.gds
```

## Configuration Explained

### config.json

```json
{
  "DESIGN_NAME": "fecim_crossbar_16x16",

  // Input files
  "VERILOG_FILES": "dir::src/crossbar.v",

  // Clock (required but unused for crossbar)
  "CLOCK_PERIOD": 10,
  "CLOCK_PORT": "CLK",

  // Custom FeCIM cell
  "EXTRA_LEFS": "dir::cells/fecim_bit.lef",
  "EXTRA_GDS_FILES": "dir::cells/fecim_bit.gds",
  "EXTRA_LIBS": "dir::cells/fecim_bit.lib",
  "VERILOG_FILES_BLACKBOX": "dir::cells/fecim_bit.v",

  // Structural netlist - skip logic synthesis
  "SYNTH_ELABORATE_ONLY": 1,

  // Fixed die area (adjust based on array size)
  "FP_SIZING": "absolute",
  "DIE_AREA": "0 0 100 100",

  // Use pre-placed DEF, skip OpenLane placement
  "PLACEMENT_CURRENT_DEF": "dir::src/crossbar.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,

  // This is a macro, not a core
  "DESIGN_IS_CORE": 0,

  // Power grid adjustments
  "FP_PDN_ENABLE_RAILS": 0,

  // Skip CTS (no clock in crossbar)
  "RUN_CTS": 0,

  // Relaxed DRC for development (tighten for tapeout)
  "QUIT_ON_MAGIC_DRC": 0,
  "QUIT_ON_LVS_ERROR": 0
}
```

## Custom Cell Files

### fecim_bit.lef (Abstract)

```lef
MACRO fecim_bit
  CLASS CORE ;
  SIZE 0.46 BY 2.72 ;
  SYMMETRY X Y ;
  SITE unithd ;

  PIN WL
    DIRECTION INPUT ;
    PORT LAYER met1 ; RECT 0.0 0.0 0.1 2.72 ; END
  END WL

  PIN BL
    DIRECTION OUTPUT ;
    PORT LAYER met2 ; RECT 0.36 0.0 0.46 2.72 ; END
  END BL

  PIN VPWR
    DIRECTION INOUT ; USE POWER ;
    PORT LAYER met1 ; RECT 0.0 2.62 0.46 2.72 ; END
  END VPWR

  PIN VGND
    DIRECTION INOUT ; USE GROUND ;
    PORT LAYER met1 ; RECT 0.0 0.0 0.46 0.1 ; END
  END VGND
END fecim_bit
```

### fecim_bit.v (Behavioral)

```verilog
module fecim_bit (
    input  WL,
    output BL,
    inout  VPWR,
    inout  VGND
);
    // Behavioral model: BL follows WL
    // Actual conductance determined by programmed state
    assign BL = WL;
endmodule
```

### fecim_bit.lib (Timing)

```liberty
library(fecim_bit) {
  cell(fecim_bit) {
    area : 1.2512;
    pin(WL) { direction : input; capacitance : 0.001; }
    pin(BL) { direction : output; function : "WL";
      timing() { related_pin : "WL"; cell_rise(scalar) { values("0.1"); } }
    }
    pin(VPWR) { direction : inout; pg_type : primary_power; }
    pin(VGND) { direction : inout; pg_type : primary_ground; }
  }
}
```

## Validation Hook

### hooks/post_run.py

```python
#!/usr/bin/env python3
"""Post-run validation for FeCIM crossbar."""

import os
import json

def main():
    run_path = os.environ.get("RUN_DIR", ".")

    # Check metrics
    metrics_file = f"{run_path}/reports/metrics.csv"
    if os.path.exists(metrics_file):
        with open(metrics_file) as f:
            print("=== FeCIM Validation ===")
            for line in f:
                if "wire_length" in line or "cell_count" in line:
                    print(line.strip())

    # Check DRC
    drc_file = f"{run_path}/reports/signoff/drc.rpt"
    if os.path.exists(drc_file):
        with open(drc_file) as f:
            content = f.read()
            violations = content.count("violation")
            print(f"DRC violations: {violations}")

    print("=== Validation Complete ===")

if __name__ == "__main__":
    main()
```

## Expected Results

### Successful Flow Output

```
[STEP 1/13] Running Synthesis...
[INFO]: Elaborating design...
[STEP 2/13] Running Floorplan...
[INFO]: Using absolute die area: 0 0 100 100
[STEP 3/13] Running Placement...
[INFO]: Skipping initial placement (using PLACEMENT_CURRENT_DEF)
[STEP 4/13] Skipping CTS (RUN_CTS=0)...
[STEP 5/13] Running Routing...
[INFO]: Running FastRoute...
[INFO]: Running TritonRoute...
[STEP 6/13] Running Signoff...
[INFO]: Running Magic GDS...
[INFO]: Running DRC...
[STEP 7/13] Generating Reports...
[SUCCESS]: Flow completed!
```

### Output Files

```
results/final/
├── gds/fecim_crossbar.gds    # GDSII for fabrication
├── lef/fecim_crossbar.lef    # Abstract for integration
├── def/fecim_crossbar.def    # Final placement
├── verilog/fecim_crossbar.v  # Gate-level netlist
└── sdc/fecim_crossbar.sdc    # Timing constraints
```

## Troubleshooting

### "EXTRA_LEFS not found"

Verify paths are correct:
```bash
ls -la ~/OpenLane/designs/fecim_crossbar/cells/
```

### "Unplaced cells remain"

Check that DEF file has all cells with FIXED keyword:
```bash
grep "FIXED" output/crossbar.def | wc -l
# Should match number of cells
```

### "DRC violations"

For development, violations are expected with stub cells. For tape-out, design actual cells in Magic with correct DRC rules.

## Next Steps

1. **Design real FeCIM cell:** Use Magic VLSI to create DRC-clean layout
2. **Characterize timing:** Extract RC parasitics, generate accurate Liberty
3. **Scale up:** Test with 64x64, 128x128 arrays
4. **Top-level integration:** Instantiate crossbar in larger SoC design
