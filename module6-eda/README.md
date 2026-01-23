# Module 6: FeCIM Design Suite

**EDA Bridge Tool for Ferroelectric Compute-in-Memory Crossbar Arrays**

Transform neural network weights into physical chip layouts ready for OpenLane/OpenROAD fabrication flow.

## Overview

The FeCIM Design Suite bridges the gap between neural network training and physical chip implementation. It compiles floating-point weights into 30-level ferroelectric memory cells, generates industry-standard output files (Verilog, DEF, SPICE), and provides an educational interface to understand the OpenLane digital flow.

```
Neural Network Weights          Physical Crossbar Array

┌─────────────────┐             ┌─────────────────┐
│  0.5  -0.3  0.8 │   Compile   │ G₁₅  G₈   G₂₂  │
│ -0.2   0.6  0.1 │ ─────────→  │ G₁₁  G₁₈  G₅   │
│  0.9  -0.7  0.4 │             │ G₂₆  G₃   G₁₄  │
└─────────────────┘             └─────────────────┘
   Floating Point                 30-Level Cells
```

## Quick Start

```bash
# Build all binaries
go build ./...

# Run tests
go test ./... -v

# Launch GUI
go run ./cmd/eda-gui

# Or use CLI for automation
go run ./cmd/eda-cli -input data/sample_weights_8x8.json -rows 8 -cols 8
```

## Architecture: 7-Tab Interface

| Tab | Name | Status | Purpose |
|-----|------|--------|---------|
| 1 | **Compiler** | Implemented | NN weights → 30-level conductance cells |
| 2 | **Layout** | Implemented | Visual crossbar grid (color-coded by conductance) |
| 3 | **HDL** | Implemented | Verilog netlist + DEF placement preview |
| 4 | **Explorer** | Placeholder | Design space "what-if" analysis |
| 5 | **Simulate** | Placeholder | ngspice simulation bridge |
| 6 | **Export** | Implemented | Multi-format output (JSON, CSV, SPICE, Verilog, DEF) |
| 7 | **Learn** | Implemented | Interactive OpenLane/OpenROAD documentation |

---

## Tab Details

### Tab 1: Compiler

Converts floating-point neural network weights to physical FeCIM cell assignments.

**Inputs:**
- JSON weight matrix file
- Array dimensions (rows × cols)
- Quantization levels (default: 30)
- Conductance range (G_min, G_max in μS)

**Outputs:**
- Cell assignments with level, conductance, and resistance
- Compilation statistics (utilization, PSNR, level histogram)

**Key Formulas:**
```
Quantization:  level = round((weight + maxWeight) / (2 * maxWeight) × (Levels-1))
Conductance:   G = G_min + (level / (Levels-1)) × (G_max - G_min)  [μS]
Resistance:    R = 1e6 / G  [Ω]
```

### Tab 2: Layout

Interactive crossbar grid visualization.

- **Color coding:** Blue (low G) → Red (high G)
- **Click any cell** to view: row, col, weight, level, conductance, voltage
- **Zoom/pan** for large arrays (128×128+)

### Tab 3: HDL (Verilog + DEF)

Generates hardware description files for OpenLane integration.

**Verilog Output:**
- Structural netlist instantiating FeCIM cells
- Module ports for wordlines (WL), bitlines (BL), and sense lines (SL)
- Compatible with Yosys synthesis (elaborate-only mode)

**DEF Output:**
- Cell placement with FIXED or PLACED keywords
- Row-major ordering with configurable pitch
- Ready for OpenLane's `PLACEMENT_CURRENT_DEF` injection

**Architecture Support:**
- **Passive crossbar:** Simple resistive network
- **1T1R:** Transistor-gated cells for sneak path mitigation

### Tab 6: Export

Multi-format export for different toolchains:

| Format | Extension | Use Case |
|--------|-----------|----------|
| JSON | `.json` | Full mapping with statistics, version control |
| CSV | `.csv` | Spreadsheet analysis, data science |
| SPICE | `.sp` | ngspice/HSPICE simulation |
| Verilog | `.v` | OpenLane synthesis/elaboration |
| DEF | `.def` | OpenLane placement injection |

### Tab 7: Learn

Interactive OpenLane documentation covering:

- **Digital flow stages:** Synthesis → Floorplan → Placement → CTS → Routing → Signoff
- **Tool descriptions:** Yosys, OpenROAD, Magic, KLayout, netgen
- **Configuration variables:** EXTRA_LEFS, EXTRA_GDS_FILES, CURRENT_DEF
- **Custom cell integration:** How to add FeCIM cells to SKY130 PDK

---

## OpenLane Integration

The FeCIM Design Suite generates files compatible with OpenLane v1.0+ flow.

### Integration Strategy

```
┌─────────────────────────────────────────────────────────────┐
│                    OpenLane Flow                            │
├─────────────────────────────────────────────────────────────┤
│  1. Synthesis (Yosys)                                       │
│     └─ SYNTH_ELABORATE_ONLY=1 for structural netlists       │
│                                                             │
│  2. Floorplan                                               │
│     └─ FP_DEF_TEMPLATE: Use our DEF for die area/pins       │
│                                                             │
│  3. Placement                                               │
│     └─ PLACEMENT_CURRENT_DEF: Inject pre-placed DEF ─────┐  │
│     └─ PL_SKIP_INITIAL_PLACEMENT=1                       │  │
│                                              ┌───────────┘  │
│  4. CTS → 5. Routing → 6. Signoff            │              │
│                                              │              │
└──────────────────────────────────────────────│──────────────┘
                                               │
                          ┌────────────────────┘
                          │
              ┌───────────▼───────────┐
              │  FeCIM Design Suite   │
              │  ┌─────────────────┐  │
              │  │ DEF Generator   │  │
              │  │ - FIXED cells   │  │
              │  │ - 1T1R layout   │  │
              │  └─────────────────┘  │
              └───────────────────────┘
```

### Key Configuration Variables

```tcl
# In OpenLane config.json or config.tcl:

# Custom cell definitions
"EXTRA_LEFS": "/path/to/fecim_cell.lef",
"EXTRA_GDS_FILES": "/path/to/fecim_cell.gds",
"EXTRA_LIBS": "/path/to/fecim_cell.lib",

# Use FeCIM DEF as template
"FP_DEF_TEMPLATE": "/path/to/fecim_crossbar.def",

# Or inject at placement stage
"PLACEMENT_CURRENT_DEF": "/path/to/fecim_crossbar.def",
"PL_SKIP_INITIAL_PLACEMENT": 1,

# For structural netlists
"SYNTH_ELABORATE_ONLY": 1,
"VERILOG_FILES_BLACKBOX": "/path/to/fecim_cell.v"
```

See [docs/INTEGRATION.md](docs/INTEGRATION.md) for detailed OpenLane integration guide.

---

## CLI Tool

For automated/headless compilation:

```bash
# Basic usage
go run ./cmd/eda-cli -input weights.json -rows 64 -cols 64

# Full options
go run ./cmd/eda-cli \
  -input data/sample_weights_8x8.json \
  -output ./output \
  -rows 8 \
  -cols 8 \
  -levels 30 \
  -vdd 1.8 \
  -json=true \
  -csv=true \
  -spice=true \
  -verilog=true \
  -def=true
```

**CLI Options:**

| Flag | Default | Description |
|------|---------|-------------|
| `-input` | (required) | Input weights JSON file |
| `-output` | `.` | Output directory |
| `-rows` | 128 | Array rows |
| `-cols` | 128 | Array columns |
| `-levels` | 30 | Quantization levels (FeCIM standard: 30) |
| `-vdd` | 1.8 | Supply voltage for SPICE |
| `-json` | true | Export JSON mapping |
| `-csv` | true | Export CSV table |
| `-spice` | true | Export SPICE netlist |
| `-verilog` | true | Export Verilog netlist |
| `-def` | true | Export DEF placement |

---

## Project Structure

```
module6-eda/
├── cmd/
│   ├── eda-gui/main.go        # GUI application entry point
│   ├── eda-cli/main.go        # CLI tool for automation
│   └── lattice-gen/main.go    # Lattice generator CLI
├── pkg/
│   ├── compiler/
│   │   ├── types.go           # Data structures (CompileConfig, CellAssignment)
│   │   ├── compiler.go        # Compile() - weight-to-level conversion
│   │   └── compiler_test.go   # Unit tests
│   ├── export/
│   │   ├── verilog.go         # Verilog netlist generation
│   │   ├── def.go             # DEF placement file generation
│   │   ├── spice.go           # SPICE netlist generation
│   │   ├── csv.go             # CSV export
│   │   ├── json.go            # JSON export
│   │   └── lattice_generator.go # Fractal cell placement algorithm
│   └── gui/
│       ├── app.go             # Main window
│       ├── embedded.go        # Embedded version for unified GUI
│       └── tabs/
│           ├── compiler_tab.go
│           ├── layout_tab.go
│           ├── hdl_tab.go
│           ├── export_tab.go
│           └── learn_tab.go
├── cells/
│   ├── fecim_bit.stub.lef     # LEF stub (abstract cell view)
│   └── fecim_1t1r.stub.lef    # 1T1R variant
├── data/
│   ├── sample_weights_8x8.json
│   └── sample_weights_16x16.json
├── docs/
│   └── INTEGRATION.md         # OpenLane integration guide
└── Makefile
```

---

## Sample Data

Test with provided sample weights:

```bash
# 8x8 array
go run ./cmd/eda-cli -input data/sample_weights_8x8.json -rows 8 -cols 8

# 16x16 array
go run ./cmd/eda-cli -input data/sample_weights_16x16.json -rows 16 -cols 16
```

**Sample weights format:**
```json
{
  "weights": [
    [0.5, -0.3, 0.8, ...],
    [-0.2, 0.6, 0.1, ...],
    ...
  ]
}
```

---

## Key Concepts

### 30-Level Quantization

FeCIM cells support 30 discrete conductance states (not binary), enabling ~4.9 bits/cell:

```
Level 0  → G_min (lowest conductance, highest resistance)
Level 15 → G_mid (middle state)
Level 29 → G_max (highest conductance, lowest resistance)
```

### DEF File Format

The DEF (Design Exchange Format) output uses:

- **FIXED:** Cells that placement tools must not move
- **PLACED:** Cells that may be adjusted during optimization

```def
COMPONENTS 64 ;
  - cell_0_0 fecim_bit + FIXED ( 0 0 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 460 0 ) N ;
  ...
END COMPONENTS
```

### Verilog Netlist

Structural netlist instantiating FeCIM cells:

```verilog
module fecim_crossbar_8x8 (
    input  [7:0] WL,    // Wordlines
    output [7:0] BL,    // Bitlines
    inout  VDD,
    inout  VSS
);
    fecim_bit cell_0_0 (.WL(WL[0]), .BL(BL[0]), .VDD(VDD), .VSS(VSS));
    fecim_bit cell_0_1 (.WL(WL[0]), .BL(BL[1]), .VDD(VDD), .VSS(VSS));
    // ...
endmodule
```

---

## Documentation

| Document | Description |
|----------|-------------|
| [INTEGRATION.md](docs/INTEGRATION.md) | OpenLane integration guide |
| [plan-demo6.md](../docs/eda/plan-demo6.md) | Implementation plan with code templates |
| [FeCIM-EDA-Strategy.md](../docs/eda/FeCIM-EDA-Strategy.md) | Project strategy |
| [eda.opensource.md](../docs/eda/eda.opensource.md) | Open-source EDA ecosystem analysis |
| [eda.eli5.md](../docs/eda/eda.eli5.md) | Beginner-friendly EDA explanation |

---

## Roadmap

### Implemented
- [x] Weight-to-conductance compiler
- [x] Visual crossbar layout
- [x] Verilog/DEF generation
- [x] Multi-format export (JSON, CSV, SPICE)
- [x] OpenLane documentation (Learn tab)
- [x] CLI tool

### In Progress
- [ ] OpenLane flow integration testing
- [ ] Custom FeCIM cell LEF/GDS (Magic layout)
- [ ] Liberty timing model generation

### Planned
- [ ] Design space explorer (area/power/throughput estimation)
- [ ] ngspice simulation bridge
- [ ] Automated DRC/LVS validation
- [ ] Multi-layer stacked crossbar support

---

## Contributing

This module is part of the FeCIM Visualizer educational suite. See the root [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

MIT License - See [LICENSE](../LICENSE)
