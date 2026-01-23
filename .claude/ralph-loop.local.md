---
active: true
iteration: 1
max_iterations: 1000
completion_promise: "PHASE 1 COMPLETE: Verilog/DEF generators integrated with pkg/export/, validated by Yosys, tests passing, consumes existing CrossbarMapping"
started_at: "2026-01-23T16:24:09Z"
---



PERSONAS:
- Dr. external research group: Ferroelectric materials expert, FeFET device physics,
  commercialization strategy. CONSULT FOR: Device parameter validation,
  ensuring 30-level quantization matches published specs.
- Dr. Sungsik Shin: FeFET array architecture, 1T1R vs passive design,
  sneak path mitigation, IR drop compensation. CONSULT FOR: Phase 2
  architecture selection (1T1R requires SL[] pins, passive does not).
- Senior EDA Engineer: OpenLane flow, DEF/LEF/Verilog generation,
  SKY130 PDK constraints. CONSULT FOR: DEF syntax validation, OpenLane
  config.tcl variables (CURRENT_DEF, PL_TARGET_DENSITY, FP_SIZING).

TASK: Implement FeCIM Lattice Generator integrated with Demo 6 Architecture

CONTEXT: This extends the existing module6-eda infrastructure defined in
docs/eda/plan-demo6.md. The generator MUST consume CrossbarMapping from
pkg/compiler/compiler.go and output to pkg/export/.

PHASE 0 - MOTHER CELL DEFINITION (Foundation):
□ Document cell geometry decision:
  - Option A: Placeholder cell (0.46μm × 2.72μm) for simulation-only
  - Option B: Design actual FeCIM bit in Magic using SKY130 primitives
□ Create cells/fecim_bit.stub.lef with abstract definition
□ Output: docs/eda/cell-geometry-decision.md documenting choice

PHASE 1 - CORE GENERATOR (Priority):
□ Extend pkg/export/ with new files:
  - pkg/export/verilog.go: GenerateVerilog(mapping *compiler.CrossbarMapping)
  - pkg/export/def.go: GenerateDEF(mapping *compiler.CrossbarMapping)
□ Verilog generator:
  - Consume CrossbarMapping.Cells for instance generation
  - Cell naming: R_{row}_{col} (matches existing spice.go convention)
  - Module ports: WL[0:rows-1], BL[0:cols-1]
□ DEF generator:
  - UNITS DATABASE MICRONS 1000
  - COMPONENTS section with FIXED placement
  - Coordinates from CrossbarMapping cell positions
□ Validate: Yosys reads Verilog with 0 errors
□ Test: pkg/export/verilog_test.go, pkg/export/def_test.go

PHASE 2 - ARCHITECTURE CONFIGURATION:
□ Extend compiler.CompileConfig with:
  - Architecture string: 'passive' | '1T1R'
  - CellPitch, RowHeight float64
□ Conditional pin generation:
  - Passive: WL[], BL[]
  - 1T1R: WL[], BL[], SL[] (source lines for transistor)
□ Update GenerateVerilog/GenerateDEF to respect architecture choice
□ Dr. Shin review checkpoint: Verify 1T1R sneak path considerations

PHASE 3 - GUI INTEGRATION:
□ Add to pkg/gui/tabs/layout_tab.go:
  - 'Generate HDL' button consuming current CrossbarMapping
  - Split view: Verilog (left), DEF (right)
  - Visual grid showing placed cells with coordinates
□ Wire into Demo 6 tab architecture per plan-demo6.md
□ Export button writes to generated/ directory

CONSTRAINTS:
- SKY130 PDK compatibility (130nm, 6 metal layers)
- Database units: 1000 per micron (DEF standard)
- Cell dimensions: Configurable, default 0.46μm × 2.72μm
- Output directory: generated/ (local, not /mnt/)
- Must integrate with existing CrossbarMapping struct

FILE STRUCTURE (extends module6-eda/):
pkg/export/
├── verilog.go          # NEW: Verilog netlist generator
├── verilog_test.go     # NEW: Verilog tests
├── def.go              # NEW: DEF placement generator
├── def_test.go         # NEW: DEF tests
├── spice.go            # EXISTING: Update R_{row}_{col} naming
├── json.go             # EXISTING
└── csv.go              # EXISTING

cells/
└── fecim_bit.stub.lef  # NEW: Placeholder LEF for OpenLane

generated/
├── lattice.v           # Output: Verilog netlist
├── placement.def       # Output: Cell placement
└── config.tcl          # Output: OpenLane config snippet

SUCCESS CRITERIA:
□ Phase 0: Cell geometry documented with rationale
□ Phase 1:
  - Generate 4×4 array from sample CrossbarMapping
  - Yosys reads lattice.v with 0 errors: yosys -p 'read_verilog lattice.v'
  - Instance names R_{row}_{col} match between .v and .def
  - go test ./pkg/export -v passes
□ Phase 2:
  - Both passive and 1T1R configs generate valid output
  - 1T1R includes SL[] ports
□ Phase 3:
  - GUI displays generated HDL
  - Export writes to generated/ directory
  - Tab integrates with Demo 6 unified app

DEPENDENCIES:
- pkg/compiler/compiler.go (CrossbarMapping, CompileConfig)
- pkg/compiler/types.go (CellAssignment struct)
- docs/eda/plan-demo6.md (architecture reference)


