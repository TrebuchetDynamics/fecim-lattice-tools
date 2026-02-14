<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# Module 6: EDA Design Suite

## Purpose

Module 6 is an educational EDA (Electronic Design Automation) pipeline that bridges FeCIM simulation to open-source layout and verification tools. It provides lattice design generation, circuit compilation, layout visualization, and integration with OpenLane for standard cell design.

Key capabilities:
- Lattice compiler: transforms array configurations into physical designs
- Export formats: Verilog, DEF, LEF, SPICE, Liberty (timing), CSV
- Layout visualization with interactive canvas
- OpenLane integration for automated place-and-route
- Validation: DRC, LVS, Yosys synthesis verification
- Educational GUI with learn tab and visual demonstrations

## Key Files

### Core Compiler
- `pkg/compiler/compiler.go` - Main entry point (GenerateDesign, GenerateBlank, mapWeights)
- `pkg/compiler/types.go` - Data structures (ArrayConfig, ArrayDesign, CellAssignment, DesignStats)
- `pkg/compiler/compiler_test.go` - Unit tests for compiler logic
- `pkg/compiler/compiler_extended_test.go` - Integration tests
- `pkg/compiler/mode_quantization_validation_test.go` - Quantization mode validation

### Export Generators
- `pkg/export/lattice_generator.go` - Main lattice generation logic
- `pkg/export/verilog.go` - Verilog HDL generation (array and cell definitions)
- `pkg/export/array_verilog.go` - Array-level Verilog with port mappings
- `pkg/export/cell_verilog.go` - Cell library Verilog definitions
- `pkg/export/def.go` - DEF (Design Exchange Format) for layout
- `pkg/export/lef.go` - LEF (Library Exchange Format) for cell abstractions
- `pkg/export/spice.go` - SPICE netlist generation
- `pkg/export/liberty.go` - Liberty timing and power characterization
- `pkg/export/json.go`, `csv.go` - Data export formats
- `pkg/export/openlane_config.go` - OpenLane configuration generation
- `pkg/export/provenance.go` - Design metadata and audit trail
- Export tests: `generators_test.go`, `roundtrip_test.go`, format verification tests

### Layout & Placement
- `pkg/layout/placement_routing.go` - Cell placement and routing algorithms
- `pkg/layout/def_generator.go` - DEF file generation from placement
- `pkg/layout/verilog_generator.go` - Verilog from layout
- `pkg/layout/layout_test.go` - Layout algorithm tests

### Validation
- `pkg/validation/def_validator.go` - DEF syntax and semantic validation
- `pkg/validation/openlane.go` - OpenLane flow validation
- `pkg/validation/yosys.go` - Yosys synthesis validation
- `pkg/validation/circuit_image.go`, `layout_image.go` - Visual validation outputs
- `pkg/validation/cross_check.go` - Cross-file consistency checks
- Extended validation tests: `architecture_test.go`, `cross_check_error_paths_test.go`, `openlane_def_validation_test.go`

### OpenLane Integration
- `pkg/openlane/manager.go` - OpenLane project management
- `pkg/openlane/runner.go` - OpenLane flow execution
- `pkg/openlane/config.go` - OpenLane configuration
- `pkg/openlane/openlane_test.go` - OpenLane integration tests

### Validation CLI
- `pkg/validate/drc.go` - Design Rule Check
- `pkg/validate/lvs.go` - Layout vs Schematic verification
- `pkg/validate/yosys.go` - Synthesis validation
- `pkg/validate/pdk_bridge.go` - PDK (Process Design Kit) bridge to tools
- `pkg/validate/validate_test.go` - Validation tests

### Configuration
- `pkg/config/types.go` - Configuration data structures
- `pkg/config/types_test.go` - Config validation tests

### GUI
- `pkg/gui/app.go` - Main application window setup
- `pkg/gui/tabs/builder_validation_tab.go` - Tab 1: Design configuration and validation
- `pkg/gui/tabs/export_viewer_tab.go` - Tab 2: Export format viewer
- `pkg/gui/tabs/layout_visualizer_tab.go` - Tab 3: Interactive layout visualization
- `pkg/gui/tabs/learn_tab.go` - Tab 4: Educational content (visuals for array, cell, transistor)
- `pkg/gui/tabs/learn_visuals.go`, `learn_visuals_array.go`, `learn_visuals_cell.go`, `learn_visuals_transistor.go` - Learn tab visuals
- `pkg/gui/widgets/layout_canvas.go` - Canvas-based layout rendering
- `pkg/gui/keyboard.go` - Keyboard shortcuts and navigation
- GUI tests: `gui_test.go`, `tabs_test.go`, `cli_gui_equivalence_test.go`, keyboard/builder validation tests

### CLI Entry Points
- `cmd/eda-cli/` - Command-line interface for batch design generation
- `cmd/eda-gui/` - GUI launcher
- `cmd/hello/` - Hello world example
- `cmd/lattice-gen/` - Standalone lattice generator

## Subdirectories

```
module6-eda/
├── cmd/
│   ├── eda-cli/                 # CLI tool
│   ├── eda-gui/                 # GUI launcher
│   ├── hello/                   # Example
│   └── lattice-gen/             # Lattice generator
├── pkg/
│   ├── compiler/                # Core design compilation
│   ├── config/                  # Configuration types
│   ├── export/                  # Export generators (Verilog, DEF, etc)
│   ├── gui/                     # Fyne GUI application
│   ├── layout/                  # Placement and routing
│   ├── openlane/                # OpenLane integration
│   ├── validate/                # CLI validation tools
│   └── validation/              # Validation algorithms
├── examples/                    # Example designs and workflows
├── data/                        # Sample data and weights
├── cells/                       # Cell definitions and layouts
├── Makefile                     # Build commands
├── README.md
└── AGENTS.md                   # This file
```

## For AI Agents

### Working

**Current State:**
- Compiler fully functional: transforms ArrayConfig → ArrayDesign with cell assignments
- Export generators produce valid Verilog, DEF, LEF, SPICE, Liberty
- Layout visualization with interactive canvas widget
- Validation pipeline: DRC, LVS, Yosys synthesis checks
- OpenLane integration for automated place-and-route
- GUI with 4 tabs: builder/validation, export viewer, layout visualizer, learn
- Headless CLI for batch operations

**Task Pattern:**
1. Read ArrayConfig to understand design parameters (rows, cols, levels, mode)
2. Compiler generates ArrayDesign with CellAssignment array
3. Export generators traverse design and output format-specific files
4. Validation checks for correctness (syntax, DRC, LVS)
5. GUI tabs display results with interactive controls

**Key Patterns:**
- Design is immutable after GenerateDesign() - changes require regeneration
- Export is format-agnostic: same design → multiple output formats
- Validation is composable: can run DRC alone or full flow
- GUI uses Stack for tab switching with currentView tracking
- All file I/O is atomic (write-then-verify)

### Testing

**Test Files:**
- `pkg/compiler/*_test.go` - Compiler logic and quantization validation
- `pkg/export/*_test.go` - Export format generation and round-trip validation
- `pkg/layout/layout_test.go` - Placement and routing
- `pkg/validation/*_test.go` - Validation algorithms and file generation
- `pkg/openlane/*_test.go` - OpenLane integration
- `pkg/gui/tabs/*_test.go` - GUI components
- `pkg/validate/validate_test.go` - CLI validation tools

**Run Tests:**
```bash
go test ./module6-eda/...                    # All tests
go test -v ./module6-eda/pkg/compiler        # Compiler only
go test -v ./module6-eda/pkg/export          # Export only
go test -v ./module6-eda/pkg/validation      # Validation only
```

**Test Coverage Notes:**
- Compiler: >85% coverage (array generation, weight mapping, quantization)
- Export: >90% coverage (all formats tested with round-trip validation)
- Validation: >80% coverage (DRC, LVS, cross-checks)
- GUI: smoke tests (components build and respond to input)

**Build Commands (Makefile):**
```bash
make -C module6-eda build              # Build all binaries
make -C module6-eda test               # Run tests
make -C module6-eda run                # Launch GUI
make -C module6-eda cli                # Run CLI demo
make -C module6-eda clean              # Remove artifacts
```

### Patterns

**Compiler Design:**
- ArrayConfig specifies target design (rows, cols, mode, architecture, technology)
- GenerateDesign() produces ArrayDesign with stats (TotalCells, AreaMM2, PowerMW)
- Two modes: Blank (initialized at Level 0) or Compute (with weights)
- Cell assignment includes: Row, Col, Level, Conductance, Resistance, ProgramV

**Export Pipeline:**
- Each exporter reads ArrayDesign and produces format-specific output
- Verilog: HDL module definitions for array and cells
- DEF: physical layout with cell placement coordinates
- LEF: cell library abstractions for place-and-route
- SPICE: netlist for circuit simulation
- Liberty: timing and power characterization

**Validation Strategy:**
- DRC: design rule checks (spacing, minimum width)
- LVS: layout vs schematic verification (connectivity check)
- Yosys: Verilog synthesis validation
- Cross-check: consistency across DEF, Verilog, SPICE
- File generation: produced files are validated before use

**GUI Architecture:**
- Tabs share ArrayConfig state
- Builder tab generates design and shows stats
- Export viewer displays file contents (can switch formats)
- Layout visualizer renders cells on canvas with zoom/pan
- Learn tab shows educational content (no design required)

**OpenLane Integration:**
- Manager creates project structure and configuration
- Runner executes flow stages (synthesis, place, route)
- Config generates OpenLane JSON from ArrayConfig
- Output: GDS, DEF, timing reports

## Dependencies

**Internal:**
- `shared/logging` - Logging infrastructure
- `shared/theme` - Fyne theme and styling
- `shared/widgets` - Reusable Fyne widgets
- `shared/export` - Shared export utilities
- `shared/utils` - Utility functions

**External:**
- `fyne.io/fyne/v2` - GUI framework
- Standard Go packages (fmt, math, json, encoding/csv, os, path/filepath)

**External Tools (Optional):**
- `yosys` - Open-source synthesis (for validation)
- `openLane` - Automated place-and-route (for layout generation)
- `magic` - Layout editor (for DRC/LVS)
- `ngspice` - Circuit simulator (for SPICE validation)

## MANUAL

### Adding a New Export Format

1. **Create Generator** in `pkg/export/myformat.go`:
   ```go
   package export

   func GenerateMyFormat(design *ArrayDesign) (string, error) {
       // Traverse design.Cells and produce format
       return output, nil
   }
   ```

2. **Add Round-Trip Test** in `pkg/export/roundtrip_test.go`:
   ```go
   func TestMyFormat_RoundTrip(t *testing.T) {
       // Generate → Parse → Validate
   }
   ```

3. **Register in Export Viewer** GUI tab
4. **Document** output structure in README.md

### Configuring OpenLane Flow

Edit `pkg/openlane/config.go` to adjust:
- Tool versions (yosys, openroad, magic)
- Timing constraints and clocking
- Power and area targets
- PDK (sky130, sky180, etc)

### Running Validation Pipeline

```bash
# Full validation
go run ./cmd/eda-cli -mode validate -def design.def

# DRC only
go run ./cmd/eda-cli -mode validate -drc-only -def design.def

# LVS only
go run ./cmd/eda-cli -mode validate -lvs-only -verilog design.v -spice design.sp
```

### Interpreting Validation Reports

- **DRC Report**: Lists violations with cell coordinates and rule ID
- **LVS Report**: Shows connectivity mismatches between schematic and layout
- **Yosys Report**: Synthesis warnings and inferred logic
- **Cross-Check Report**: File consistency and metadata validation

### Debugging Compiler Quantization

Enable debug output via `FECIM_DEBUG=compiler`:
```bash
FECIM_DEBUG=compiler go run ./cmd/eda-cli -mode design
```

Logs will show:
- Weight mapping details
- Quantization level assignments
- Cell conductance calculations
- Stats computation

### Educational Learn Tab

The Learn tab demonstrates:
- **Array View**: Grid layout of cells with conductance visualization
- **Cell View**: Single cell structure (FTJ, access transistor, bit/word lines)
- **Transistor View**: MOSFET characteristics and operation

No design required - educational content is self-contained.

### Extending the Compiler

To add a new design mode:
1. Define mode constant in `pkg/config/types.go`
2. Add handling in `GenerateDesign()` (compiler.go)
3. Implement mode-specific logic (mapWeights, GenerateBlank, etc)
4. Add tests in `compiler_test.go`
5. Update GUI builder tab to expose mode selection

---

**Last Updated:** 2026-02-13
