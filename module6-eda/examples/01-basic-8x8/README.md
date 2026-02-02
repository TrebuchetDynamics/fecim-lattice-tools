# Example 01: Basic 8x8 Crossbar

A minimal example demonstrating FeCIM weight compilation for an 8x8 crossbar array.

## Overview

This example compiles a simple 8x8 weight matrix with mixed positive and negative values, showcasing the full range of 30-level quantization.

## Files

| File | Description |
|------|-------------|
| `weights.json` | 8x8 weight matrix with values from -0.9 to +0.9 |
| `run.sh` | Script to compile and export all formats |
| `expected_output/` | Reference output for validation |

## Running the Example

```bash
# From repository root
cd module6-eda

# Method 1: Use the provided script
./examples/01-basic-8x8/run.sh

# Method 2: Run CLI directly
go run ./cmd/eda-cli \
  -input examples/01-basic-8x8/weights.json \
  -output examples/01-basic-8x8/output \
  -rows 8 -cols 8 -levels 30
```

## Expected Output

After running, you should see:

```
FeCIM Macro Compiler v1.0
Loading weights from: examples/01-basic-8x8/weights.json
Compiling 8x8 array with 30 levels...

Compilation Statistics:
  Total Cells:    64
  Utilized:       64 (100.0%)
  PSNR:           42.3 dB
  Level Range:    0 - 29

Exporting to: examples/01-basic-8x8/output/
  ✓ mapping.json
  ✓ cells.csv
  ✓ crossbar.sp
  ✓ crossbar.v
  ✓ crossbar.def

Done!
```

## Output Files

### mapping.json

Complete compilation result with statistics:

```json
{
  "config": {
    "array_rows": 8,
    "array_cols": 8,
    "levels": 30,
    "g_min": 1.0,
    "g_max": 100.0
  },
  "cells": [
    {"row": 0, "col": 0, "weight": 0.1, "level": 16, "conductance": 55.17},
    ...
  ],
  "stats": {
    "total_cells": 64,
    "utilized_cells": 64,
    "utilization": 1.0,
    "psnr_db": 42.3
  }
}
```

### cells.csv

Spreadsheet-compatible format:

```csv
row,col,weight,level,conductance_us,resistance_ohm
0,0,0.100,16,55.17,18125
0,1,-0.200,13,44.83,22319
...
```

### crossbar.v

Structural Verilog for simulation:

```verilog
module fecim_crossbar_8x8 (
    input  wire [7:0] WL,
    output wire [7:0] BL,
    inout  wire VPWR,
    inout  wire VGND
);
    fecim_bit cell_0_0 (.WL(WL[0]), .BL(BL[0]), .VPWR(VPWR), .VGND(VGND));
    // ... 64 cells total
endmodule
```

### crossbar.def

Physical placement for OpenLane:

```def
COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( 5000 5000 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 5460 5000 ) N ;
  ...
END COMPONENTS
```

## Validation

### Check SPICE Netlist

```bash
# Verify syntax with ngspice
ngspice -b -c 'source output/crossbar.sp; listing'
```

### Check Verilog

```bash
# Compile with iverilog
iverilog -o /dev/null output/crossbar.v
echo "Verilog syntax OK"
```

### Visual Inspection

Open `mapping.json` and verify:
- All 64 cells have valid levels (0-29)
- Conductance values are within [1, 100] μS
- Level distribution spans the full range

## Weight Matrix

The test weights are designed to exercise the full quantization range:

```
     Col 0   Col 1   Col 2   Col 3   Col 4   Col 5   Col 6   Col 7
Row 0:  0.1   -0.2    0.3   -0.4    0.5   -0.6    0.7   -0.8
Row 1: -0.1    0.2   -0.3    0.4   -0.5    0.6   -0.7    0.8
Row 2:  0.15  -0.25   0.35  -0.45   0.55  -0.65   0.75  -0.85
Row 3: -0.15   0.25  -0.35   0.45  -0.55   0.65  -0.75   0.85
Row 4:  0.05  -0.15   0.25  -0.35   0.45  -0.55   0.65  -0.75
Row 5: -0.05   0.15  -0.25   0.35  -0.45   0.55  -0.65   0.75
Row 6:  0.2   -0.3    0.4   -0.5    0.6   -0.7    0.8   -0.9
Row 7: -0.2    0.3   -0.4    0.5   -0.6    0.7   -0.8    0.9
```

## Next Steps

After validating this example:

1. Try with different array sizes (16x16, 32x32)
2. Modify weights.json with your own values
3. Run ngspice simulation (see `02-mnist-layer` for testbench)
4. Integrate with OpenLane (see `03-openlane-integration`)
