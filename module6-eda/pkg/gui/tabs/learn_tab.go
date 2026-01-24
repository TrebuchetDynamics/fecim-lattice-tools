// pkg/gui/tabs/learn_tab.go
// Learning Center tab for FeCIM Design Suite
// Explains OpenLane flow, FeCIM design workflow, and provides educational resources

package tabs

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// MakeLearnTab creates the learning center tab with educational content
func MakeLearnTab(state interface{}, w fyne.Window) fyne.CanvasObject {
	// Topic selector
	topics := []string{
		"🏠 Overview",
		"🔬 What is FeCIM?",
		"🛠️ OpenLane Flow",
		"⚡ Our Workflow",
		"📐 HDL Generation",
		"🏗️ Architecture",
		"📚 References",
	}

	topicSelector := widget.NewList(
		func() int { return len(topics) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(topics[id])
		},
	)
	topicSelector.OnSelected = func(id widget.ListItemID) {
		// Will be connected to content display
	}

	// Content area
	contentScroll := container.NewScroll(makeOverviewContent())
	contentScroll.SetMinSize(fyne.NewSize(600, 500))

	// Connect topic selector to content
	topicSelector.OnSelected = func(id widget.ListItemID) {
		var content fyne.CanvasObject
		switch id {
		case 0:
			content = makeOverviewContent()
		case 1:
			content = makeFeCIMContent()
		case 2:
			content = makeOpenLaneContent()
		case 3:
			content = makeWorkflowContent()
		case 4:
			content = makeHDLContent()
		case 5:
			content = makeArchitectureContent()
		case 6:
			content = makeReferencesContent()
		default:
			content = makeOverviewContent()
		}
		contentScroll.Content = content
		fyne.Do(func() {
			contentScroll.Refresh()
		})
	}

	// Select first topic by default
	topicSelector.Select(0)

	// Layout with sidebar
	sidebar := container.NewBorder(
		widget.NewLabelWithStyle("📚 Topics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		topicSelector,
	)
	sidebar.Resize(fyne.NewSize(180, 500))

	// Main layout
	split := container.NewHSplit(sidebar, contentScroll)
	split.SetOffset(0.2)

	// Header
	header := container.NewVBox(
		widget.NewLabelWithStyle("FeCIM Design Suite - Learning Center", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Learn about EDA, OpenLane, and FeCIM chip design"),
		widget.NewSeparator(),
	)

	return container.NewBorder(header, nil, nil, nil, split)
}

func makeOverviewContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Welcome to the FeCIM Design Suite", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	intro := widget.NewLabel(`This suite bridges the gap between AI models and physical silicon.

What This Tool Does:
━━━━━━━━━━━━━━━━━━━━
• Compiles neural network weights to FeCIM hardware
• Generates industry-standard HDL (Verilog/DEF)
• Visualizes crossbar array layouts
• Exports to open-source EDA tools

The Problem We Solve:
━━━━━━━━━━━━━━━━━━━━
"There is no 'OpenROAD for Analog.' You cannot click a button
and get a routed FeFET crossbar array."

Until now.

This tool provides the missing link between:
  📊 AI Models (weights, tensors)
     ↓
  💾 Physical Memory (FeFET conductances)
     ↓
  🔲 Silicon Layout (GDSII masks)

Navigation:
━━━━━━━━━━━
Select a topic from the sidebar to learn about:
• FeCIM technology and 30-level quantization
• The OpenLane ASIC design flow
• Our specific workflow for FeCIM arrays
• How to generate and validate HDL
• Architecture choices (Passive vs 1T1R)`)
	intro.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), intro)
}

func makeFeCIMContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("What is Ferroelectric Compute-in-Memory?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`FeCIM = Ferroelectric Compute-in-Memory

The Key Innovation:
━━━━━━━━━━━━━━━━━━
"Compute in memory where the same device does the memory
AND the computation." — Dr. external research group

Traditional Computing:
  Memory ←→ CPU ←→ Memory  (data shuttling = slow, power-hungry)

FeCIM Computing:
  Memory = CPU  (computation happens IN the memory array)

The 30-Level Advantage:
━━━━━━━━━━━━━━━━━━━━━━
Unlike binary memory (0/1), FeCIM supports 30 analog states:

  Level 0  ████░░░░░░░░░░░░░░░░  Low conductance (1 μS)
  Level 15 ████████████░░░░░░░░  Medium (50 μS)
  Level 29 ████████████████████  High conductance (100 μS)

This enables:
  • 4.91 bits per cell (log₂(30) = 4.91)
  • Efficient neural network weight storage
  • In-memory matrix-vector multiplication

The Crossbar Array:
━━━━━━━━━━━━━━━━━━━
       Bit Lines (BL)
         ↓   ↓   ↓   ↓
    WL→ [●] [●] [●] [●]  ← Each cell stores one weight
    WL→ [●] [●] [●] [●]     as a conductance value
    WL→ [●] [●] [●] [●]
    WL→ [●] [●] [●] [●]
         ↓   ↓   ↓   ↓
        Output Currents (I = G × V)

Matrix-Vector Multiply in O(1):
  I_out[j] = Σ G[i,j] × V_in[i]

  All multiplications happen simultaneously by Ohm's law!

Dr. Tour's Key Specifications:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  • Discrete Levels: 30 (not 2!)
  • Energy vs NAND: 10,000,000× better
  • Energy vs DRAM: 1,000× better
  • Technology Readiness: TRL 4 (lab validated)`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeOpenLaneContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("The OpenLane ASIC Design Flow", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`OpenLane: Open-Source RTL-to-GDSII Flow

What is OpenLane?
━━━━━━━━━━━━━━━━━
OpenLane automates the journey from code to manufacturable chip:

  Verilog Code → Netlist → Placed Cells → Routed Wires → GDSII

The 5 Main Stages:
━━━━━━━━━━━━━━━━━━

┌─────────────────────────────────────────────────────────┐
│  STAGE A: SYNTHESIS (Yosys)                             │
│  ───────────────────────────────                        │
│  Input:  Verilog HDL (your design description)          │
│  Output: Netlist of standard cells                      │
│  Tool:   Yosys (open-source synthesis)                  │
│                                                         │
│  Example: "assign out = ~in;" → sky130_fd_sc_hd__inv_2  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  STAGE B: FLOORPLANNING                                 │
│  ──────────────────────                                 │
│  Input:  Netlist + die size constraints                 │
│  Output: Die area, pin locations, power grid plan       │
│                                                         │
│  Key Decision: How big is the chip? Where are the IOs?  │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  STAGE C: PLACEMENT (RePlAce + OpenDP)                  │
│  ─────────────────────────────────────                  │
│  Input:  Floorplan + netlist                            │
│  Output: Each cell has X,Y coordinates                  │
│                                                         │
│  Goal: Minimize wire length, meet timing                │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  STAGE D: CLOCK TREE SYNTHESIS (TritonCTS)              │
│  ─────────────────────────────────────────              │
│  Input:  Placed design + clock constraints              │
│  Output: Balanced clock distribution network            │
│                                                         │
│  Note: FeCIM arrays may skip this (no global clock)     │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  STAGE E: ROUTING (TritonRoute)                         │
│  ──────────────────────────────                         │
│  Input:  Placed cells + netlist                         │
│  Output: Metal interconnect paths                       │
│                                                         │
│  Handles: DRC (design rules), antenna effects           │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│  OUTPUT: GDSII (Factory-Ready Layout)                   │
│  ────────────────────────────────────                   │
│  The "PDF" of chip design - sent to the foundry         │
└─────────────────────────────────────────────────────────┘

Our Integration Point:
━━━━━━━━━━━━━━━━━━━━━━
We generate the PLACEMENT (DEF file) directly because:
• FeCIM arrays are regular structures (not random logic)
• Standard place-and-route doesn't understand crossbars
• We need precise control over cell positions

Our Output → OpenLane:
  lattice.v     (Verilog netlist)
  placement.def (Pre-placed cells)
  config.tcl    (OpenLane configuration)`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeWorkflowContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("FeCIM Design Workflow", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`The Complete FeCIM Design Flow

Our 4-Step Process:
━━━━━━━━━━━━━━━━━━━

STEP 1: COMPILE (Tab 1 - Compiler)
─────────────────────────────────
  Input:  Neural network weights (JSON/numpy)

  Process:
    1. Load weight matrix
    2. Find weight range [min, max]
    3. Quantize to 30 levels (symmetric)
    4. Map levels to conductance (1-100 μS)
    5. Calculate programming voltages

  Output: CrossbarMapping structure
    • Cell assignments (row, col, weight, level, G)
    • Compilation statistics (PSNR, utilization)

STEP 2: VISUALIZE (Tab 2 - Layout)
──────────────────────────────────
  Input:  CrossbarMapping from Step 1

  Display:
    • Color-coded conductance grid
    • Click cells for details
    • Verify quantization quality

  Purpose: Sanity check before HDL generation

STEP 3: GENERATE HDL (Tab: HDL)
───────────────────────────────
  Input:  CrossbarMapping + Architecture choice

  Outputs:
    ┌─────────────────────────────────────────┐
    │ lattice.v (Verilog Netlist)             │
    │ ─────────────────────────               │
    │ module fecim_crossbar (                 │
    │     input  wire [N-1:0] WL,             │
    │     inout  wire [M-1:0] BL,             │
    │     input  wire [M-1:0] SL,  // 1T1R    │
    │     ...                                 │
    │ );                                      │
    │     fecim_bit #(.LEVEL(15)) R_0_0 (...);│
    │     ...                                 │
    │ endmodule                               │
    └─────────────────────────────────────────┘

    ┌─────────────────────────────────────────┐
    │ placement.def (Physical Layout)         │
    │ ──────────────────────────              │
    │ VERSION 5.8 ;                           │
    │ DESIGN fecim_crossbar ;                 │
    │ UNITS DISTANCE MICRONS 1000 ;           │
    │ DIEAREA ( 0 0 ) ( 31840 40880 ) ;       │
    │ COMPONENTS 16 ;                         │
    │   - R_0_0 fecim_bit + FIXED (x y) N ;   │
    │   ...                                   │
    │ END COMPONENTS                          │
    └─────────────────────────────────────────┘

STEP 4: EXPORT & VALIDATE (Tab 5 - Export)
──────────────────────────────────────────
  Validation:
    $ yosys -p 'read_verilog lattice.v; hierarchy -check'
    → "Successfully finished Verilog frontend"

  Export Formats:
    • JSON   (full mapping data)
    • CSV    (cell assignments table)
    • SPICE  (ngspice netlist for simulation)
    • DEF    (physical placement)
    • TCL    (OpenLane configuration)

Continuing to OpenLane:
━━━━━━━━━━━━━━━━━━━━━━━
After export, run OpenLane with our pre-placed design:

  $ cd OpenLane
  $ ./flow.tcl -design fecim_crossbar \
               -init_design_config ../generated/config.tcl

OpenLane will:
  1. Skip synthesis (we provide netlist)
  2. Skip placement (we provide DEF)
  3. Run routing (connect the cells)
  4. Generate final GDSII`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeHDLContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("HDL Generation Details", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`Understanding Our HDL Output

Verilog Netlist Structure:
━━━━━━━━━━━━━━━━━━━━━━━━━━

// Module declaration with array dimensions
module fecim_crossbar (
    input  wire [ROWS-1:0] WL,   // Word Lines (row select)
    inout  wire [COLS-1:0] BL,   // Bit Lines (data)
    input  wire [COLS-1:0] SL,   // Source Lines (1T1R only)
    input  wire VDD,             // Power
    input  wire VSS              // Ground
);
    // Parameters embedded in module
    parameter ROWS = 4;
    parameter COLS = 4;
    parameter LEVELS = 30;
    parameter ARCHITECTURE = "1T1R";

    // Cell instances with quantization level
    fecim_bit #(.LEVEL(15)) R_0_0 (
        .WL  (WL[0]),
        .BL  (BL[0]),
        .SL  (SL[0]),   // 1T1R architecture
        .VDD (VDD),
        .VSS (VSS)
    );
    // ... more cells ...
endmodule

Cell Naming Convention:
━━━━━━━━━━━━━━━━━━━━━━
  R_{row}_{col}

  Examples:
    R_0_0  → Cell at row 0, column 0
    R_3_7  → Cell at row 3, column 7

  This matches across: Verilog, DEF, and SPICE

DEF File Structure:
━━━━━━━━━━━━━━━━━━━
VERSION 5.8 ;
DESIGN fecim_crossbar ;
UNITS DISTANCE MICRONS 1000 ;     // 1000 DBU = 1 μm

DIEAREA ( 0 0 ) ( Xmax Ymax ) ;   // Die boundary

COMPONENTS N ;                     // N cells
    - R_0_0 fecim_bit + FIXED ( X Y ) N ;
    // X, Y in database units (multiply by 1000)
    // N = North orientation
END COMPONENTS

PINS M ;                          // M I/O pins
    - WL[0] + NET WL[0] + DIRECTION INPUT ...
    - BL[0] + NET BL[0] + DIRECTION INOUT ...
    - SL[0] + NET SL[0] + DIRECTION INPUT ...  // 1T1R
END PINS

NETS K ;                          // K signal nets
    - WL[0] ( PIN WL[0] ) ( R_0_0 WL ) ( R_0_1 WL ) ...
    - BL[0] ( PIN BL[0] ) ( R_0_0 BL ) ( R_1_0 BL ) ...
END NETS

END DESIGN

Coordinate System:
━━━━━━━━━━━━━━━━━━
  Cell pitch (passive): 0.46 μm = 460 DBU
  Cell pitch (1T1R):    0.92 μm = 920 DBU
  Row height:           2.72 μm = 2720 DBU

  Origin offset: (10 μm, 10 μm) from die corner

  Cell R_i_j placement:
    X = 10000 + j × CellPitch
    Y = 10000 + i × RowHeight

Validation Commands:
━━━━━━━━━━━━━━━━━━━━
  # Check Verilog syntax
  $ yosys -p 'read_verilog lattice.v; hierarchy -check'

  # View in KLayout
  $ klayout placement.def

  # Simulate in ngspice (after SPICE export)
  $ ngspice crossbar.sp`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeArchitectureContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Crossbar Architecture: Passive vs 1T1R", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`Choosing the Right Architecture

Two Options:
━━━━━━━━━━━━

┌─────────────────────────────────────────────────────────┐
│  PASSIVE CROSSBAR                                       │
│  ────────────────                                       │
│       BL[0]  BL[1]                                      │
│         │      │                                        │
│  WL[0]──●──────●──                                      │
│         │      │                                        │
│  WL[1]──●──────●──                                      │
│         │      │                                        │
│                                                         │
│  Ports: WL[], BL[], VDD, VSS                           │
│  Cell:  fecim_bit (0.46 μm × 2.72 μm)                  │
│                                                         │
│  ✓ Simple structure                                     │
│  ✓ Dense packing                                        │
│  ✓ Good for small arrays (≤32×32)                      │
│  ✗ Sneak path currents                                  │
│  ✗ Read accuracy degrades with size                     │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  1T1R (1 Transistor 1 Resistor)                         │
│  ──────────────────────────────                         │
│       BL[0]  BL[1]  SL[0]  SL[1]                        │
│         │      │      │      │                          │
│  WL[0]──┼──────┼──────┼──────┼──                        │
│         │      │      │      │                          │
│  WL[1]──┼──────┼──────┼──────┼──                        │
│         │      │      │      │                          │
│                                                         │
│  Ports: WL[], BL[], SL[], VDD, VSS                     │
│  Cell:  fecim_1t1r (0.92 μm × 2.72 μm)                 │
│                                                         │
│  ✓ Sneak path mitigation                                │
│  ✓ Scales to large arrays (128×128+)                   │
│  ✓ Better read accuracy                                 │
│  ✗ Larger cell size (2× pitch)                          │
│  ✗ Additional routing (SL lines)                        │
└─────────────────────────────────────────────────────────┘

The Sneak Path Problem:
━━━━━━━━━━━━━━━━━━━━━━━
In passive arrays, current can flow through unintended paths:

  Reading cell (0,0):
    Apply V to WL[0], ground BL[0]

    Intended: WL[0] → Cell(0,0) → BL[0]

    Sneak:    WL[0] → Cell(0,1) → BL[1] → Cell(1,1) → ...

  Result: Parasitic current corrupts the read signal
  Impact: Error ∝ N² for N×N arrays

1T1R Solution:
━━━━━━━━━━━━━━
  • Each cell has a select transistor
  • Transistor gate connected to WL
  • When WL is LOW → transistor OFF → cell isolated
  • Source Line (SL) provides controlled current path

  Reading cell (0,0):
    WL[0] = HIGH, others LOW
    → Only row 0 transistors are ON
    → All other cells are isolated
    → No sneak paths!

Recommendation:
━━━━━━━━━━━━━━━
  ┌──────────────┬─────────────────────────┐
  │ Array Size   │ Recommended             │
  ├──────────────┼─────────────────────────┤
  │ ≤ 32×32      │ Passive (simpler)       │
  │ 64×64        │ Either (app dependent)  │
  │ ≥ 128×128    │ 1T1R (accuracy needed)  │
  └──────────────┴─────────────────────────┘

Physical Files:
━━━━━━━━━━━━━━━
  Passive:
    cells/fecim_bit.stub.lef
    generated/fecim_bit.v

  1T1R:
    cells/fecim_1t1r.stub.lef
    generated/fecim_1t1r.v`)
	content.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), content)
}

func makeReferencesContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("References & Resources", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`Scientific Foundation & Learning Resources

Core Physics Papers:
━━━━━━━━━━━━━━━━━━━━
• Shin et al. (2025) "Flash In2Se3 for Neuromorphic Computing"
  - Validates 30-state analog memory
  - Published by Tour Group, external research institution

• Tour Group publications on HfO2-ZrO2 superlattices
  - Ferroelectric switching dynamics
  - Endurance and retention characteristics

EDA Tool Documentation:
━━━━━━━━━━━━━━━━━━━━━━━
• OpenLane Documentation
  https://openlane.readthedocs.io/

• Yosys Manual
  https://yosyshq.readthedocs.io/

• Magic VLSI
  http://opencircuitdesign.com/magic/

• ngspice Manual
  https://ngspice.sourceforge.io/docs.html

• KLayout Documentation
  https://www.klayout.de/doc.html

Open PDKs:
━━━━━━━━━━
• SkyWater SKY130 (130nm)
  https://github.com/google/skywater-pdk
  - Free, extensive documentation
  - No FeFET models (we add our own)

• GlobalFoundries GF180MCU (180nm)
  https://github.com/google/gf180mcu-pdk
  - High voltage support (good for FeFET)

• IHP SG13G2 (130nm)
  https://github.com/IHP-GmbH/IHP-Open-PDK
  - Has RRAM/memristor support
  - Current Tiny Tapeout target

Learning Paths:
━━━━━━━━━━━━━━━
Beginner:
  1. Read docs/eda/eda.eli5.md (EDA Explained Like I'm 5)
  2. Complete Tab 1 tutorial (compile sample weights)
  3. Visualize in Tab 2 (understand the grid)
  4. Generate HDL (see the output)

Intermediate:
  1. Read docs/eda/eda.guide.zero_to_asic.md
  2. Understand the OpenLane flow
  3. Install ngspice, run simulations
  4. Explore passive vs 1T1R trade-offs

Advanced:
  1. Read docs/eda/eda.research.meta-study.md
  2. Understand NeuroSim vs CiMLoop comparison
  3. Design custom FeFET models
  4. Contribute to open-source PDK extensions

Community:
━━━━━━━━━━
• Zero to ASIC Course (Matt Venn)
  https://www.zerotoasiccourse.com/

• Tiny Tapeout
  https://tinytapeout.com/

• Open Silicon Discord
  (Central hub for open-source chip design)

Project Documentation:
━━━━━━━━━━━━━━━━━━━━━━
All documentation is in docs/eda/:
  • plan-demo6.md        - Implementation roadmap
  • eda.eli5.md          - Beginner-friendly EDA guide
  • eda.opensource.ecosystem.md - Tool analysis
  • architecture-1t1r-review.md - Dr. Shin review
  • cell-geometry-decision.md   - Design decisions
  • REFERENCES.md        - Full bibliography`)
	content.Wrapping = fyne.TextWrapWord

	// Add a visual separator
	line := canvas.NewLine(theme.ForegroundColor())
	line.StrokeWidth = 1

	return container.NewVBox(title, widget.NewSeparator(), content)
}
