// pkg/gui/tabs/learn_tab.go
// Learning Center tab for FeCIM Design Suite
// Explains OpenLane flow and where the FeCIM Array Builder fits in

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
		"1. Overview",
		"2. OpenLane Flow",
		"3. Where We Fit In",
		"4. What We Generate",
		"5. Cell Types",
		"6. References",
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
			content = makeOpenLaneFlowContent()
		case 2:
			content = makeWhereWeFitContent()
		case 3:
			content = makeWhatWeGenerateContent()
		case 4:
			content = makeCellTypesContent()
		case 5:
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
		widget.NewLabelWithStyle("Topics", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil,
		topicSelector,
	)
	sidebar.Resize(fyne.NewSize(180, 500))

	// Main layout
	split := container.NewHSplit(sidebar, contentScroll)
	split.SetOffset(0.22)

	// Header
	header := container.NewVBox(
		widget.NewLabelWithStyle("FeCIM Array Builder - Learning Center", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Understanding OpenLane and where our tool fits in"),
		widget.NewSeparator(),
	)

	return container.NewBorder(header, nil, nil, nil, split)
}

func makeOverviewContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("What This Tool Does", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	intro := widget.NewLabel(`Module 6 is an ARRAY BUILDER that generates EDA files
for integrating FeCIM crossbar arrays into the OpenLane flow.

WHAT WE DO:
-----------
  * Generate LEF files (cell abstracts)
  * Generate Liberty files (timing - placeholder values)
  * Generate Verilog netlists (behavioral models)
  * Generate DEF files (physical placement)
  * Export OpenLane configuration

WHAT WE DON'T DO:
-----------------
  * We do NOT provide validated FeFET device models
  * We do NOT generate production-ready layouts
  * We do NOT characterize real timing values
  * We do NOT fabricate chips

PURPOSE:
--------
This is an EDUCATIONAL tool that demonstrates how
FeCIM arrays could integrate with open-source EDA.
All timing values are placeholders - real values
require SPICE characterization with validated models.

DISCLAIMER:
-----------
This project is not affiliated with or endorsed by
external research institution, Dr. external research group, or any foundry.`)
	intro.Wrapping = fyne.TextWrapWord

	return container.NewVBox(title, widget.NewSeparator(), intro)
}

func makeOpenLaneFlowContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("The OpenLane RTL-to-GDSII Flow", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Visual flow diagram
	flowDiagram := OpenLaneFlowDiagram(false)

	description := widget.NewLabel(`OpenLane automates the journey from Verilog to GDSII.

THE STAGES EXPLAINED:
---------------------
1. SYNTHESIS (Yosys)
   Converts behavioral Verilog to gate-level netlist
   Example: "a & b" -> sky130_fd_sc_hd__and2_1

2. FLOORPLAN
   Defines die area and I/O pin locations

3. PLACEMENT (RePlAce + OpenDP)
   Assigns X,Y coordinates to every cell

4. CTS (Clock Tree Synthesis)
   Distributes clock signal evenly
   Note: FeCIM arrays often skip this

5. ROUTING (TritonRoute)
   Draws metal wire connections

6. SIGNOFF & GDSII
   DRC/LVS verification, final output

REFERENCES:
  * openlane.readthedocs.io
  * OpenLane Paper: WOSET 2020`)
	description.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		flowDiagram,
		widget.NewSeparator(),
		description,
	)
}

func makeWhereWeFitContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Where the FeCIM Array Builder Fits In", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Visual flow diagram with our contribution highlighted
	flowDiagram := OpenLaneFlowDiagram(true)

	// Isometric crossbar visualization
	crossbarTitle := widget.NewLabelWithStyle("Why We Pre-Place: The Crossbar Structure", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	crossbarDiagram := IsometricCrossbar(4, 4, true)

	description := widget.NewLabel(`OUR FILES AND WHERE THEY GO:
-----------------------------
  Our Verilog -> Input to Synthesis
  Our LEF     -> Defines cell geometry for Floorplan
  Our DEF     -> REPLACES Placement (FIXED positions)
  Our LIB     -> Timing info (placeholder values!)

THE KEY INSIGHT:
----------------
Standard auto-placement would scatter our cells randomly.
We provide a DEF with FIXED positions to maintain the
regular grid structure that enables:

  * Matrix-vector multiply (I = G x V)
  * Predictable IR-drop modeling
  * Uniform sneak path analysis

WHAT OPENLANE STILL DOES:
-------------------------
  * Routing (connecting our pre-placed cells)
  * DRC checking (design rule verification)
  * Final GDSII generation`)
	description.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		flowDiagram,
		widget.NewSeparator(),
		crossbarTitle,
		crossbarDiagram,
		widget.NewSeparator(),
		description,
	)
}

func makeWhatWeGenerateContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("What Files We Generate", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	intro := widget.NewLabel("The Array Builder generates EDA files for OpenLane integration:")
	intro.Wrapping = fyne.TextWrapWord

	// File format preview cards in a grid
	lefCard := LEFPreviewCard()
	defCard := DEFPreviewCard()
	verilogCard := VerilogPreviewCard()

	cardsRow1 := container.NewHBox(lefCard, defCard)
	cardsRow2 := container.NewHBox(verilogCard)

	description := widget.NewLabel(`FILE PURPOSES:
--------------
LEF (Library Exchange Format)
  Defines cell GEOMETRY - size and pin locations
  This is an ABSTRACT view, no transistors

DEF (Design Exchange Format)
  Physical PLACEMENT with X,Y coordinates
  FIXED keyword prevents auto-placement

Verilog Netlist
  Structural description of the array
  Cells are black boxes (behavioral only)

Liberty (.lib)
  Timing information for synthesis
  WARNING: All values are PLACEHOLDERS!

OpenLane Config (JSON)
  Points OpenLane to our custom files

IMPORTANT DISCLAIMERS:
----------------------
* LEF is abstract - no real layout
* Liberty timing values need SPICE characterization
* Verilog doesn't model FeFET physics
* Real fabrication requires validated cells`)
	description.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		intro,
		cardsRow1,
		cardsRow2,
		widget.NewSeparator(),
		description,
	)
}

func makeCellTypesContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Cell Types: Passive vs 1T1R", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	// Visual crossbar diagrams side by side
	passiveDiagram := IsometricCrossbar(3, 3, true)
	oneToneRDiagram := Isometric1T1RCrossbar(3, 3)

	// Put both diagrams in a horizontal box
	diagramsRow := container.NewHBox(passiveDiagram, oneToneRDiagram)

	passiveContent := widget.NewLabel(`PASSIVE CROSSBAR
----------------
  Ports: WL[], BL[], VDD, VSS
  Cell Size: 0.46 x 2.72 um (SKY130 site)

  + Simple, dense packing
  + Lower fabrication complexity
  - SNEAK PATH CURRENTS
  - Limited to small arrays (~32x32)`)
	passiveContent.Wrapping = fyne.TextWrapWord

	oneToneRContent := widget.NewLabel(`1T1R (1 Transistor + 1 Resistor)
--------------------------------
  Ports: WL[], BL[], SL[], VDD, VSS
  Cell Size: 0.92 x 2.72 um (2x width)

  + No sneak paths (transistor isolates)
  + Scales to 128x128+ arrays
  - Larger cell area (2x)
  - More complex routing`)
	oneToneRContent.Wrapping = fyne.TextWrapWord

	sneakPath := widget.NewLabel(`THE SNEAK PATH PROBLEM
----------------------
In passive arrays, reading cell (0,0):

  INTENDED: WL[0] -> Cell(0,0) -> BL[0]

  SNEAK:    WL[0] -> Cell(0,1) -> BL[1]
                  -> Cell(1,1) -> Cell(1,0) -> BL[0]

Error grows as N^2 for NxN arrays!


RECOMMENDATION
--------------
  <= 16x16   -> Passive
  32x32      -> Either (depends on accuracy needs)
  >= 64x64   -> 1T1R required

REFERENCES: RSC Nanoscale Advances 2020, IEEE JSSC`)
	sneakPath.Wrapping = fyne.TextWrapWord

	return container.NewVBox(
		title,
		widget.NewSeparator(),
		diagramsRow,
		widget.NewSeparator(),
		passiveContent,
		oneToneRContent,
		widget.NewSeparator(),
		sneakPath,
	)
}

func makeReferencesContent() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("References", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	content := widget.NewLabel(`All claims in this tool are backed by published research.

OPENLANE / EDA
==============
[1] "OpenLANE: The Open-Source Digital ASIC Flow"
    WOSET 2020
    https://woset-workshop.github.io/PDFs/2020/a21.pdf
    -> Our DEF/Verilog formats follow this spec

[2] OpenLane Documentation
    https://openlane.readthedocs.io/
    -> Configuration variables we generate

[3] LEF/DEF 5.8 Specification
    Si2/OpenAccess Coalition
    -> File format standards we follow


SKYWATER PDK
============
[4] SkyWater SKY130 Process Design Kit
    https://skywater-pdk.readthedocs.io/
    -> Cell dimensions: 0.46um site, 2.72um height
    -> Metal layer specs for pin placement

[5] SKY130 Standard Cells
    https://github.com/google/skywater-pdk
    -> LEF/Liberty format examples


FECIM DEVICE PHYSICS
====================
[6] Shin et al. "Flash In2Se3 for Neuromorphic Computing"
    Advanced Electronic Materials, 2025
    -> 30 discrete analog states

[7] Tour Group HfO2-ZrO2 Superlattice Research
    external research institution
    -> Ferroelectric switching dynamics

[8] "Roadmap on Ferroelectric Hafnia/Zirconia"
    APL Materials, 2023
    -> HfO2/ZrO2 material properties


CIM ARCHITECTURE
================
[9] "Sneak Path Solutions for Crossbar Arrays"
    RSC Nanoscale Advances, 2020
    -> Passive vs 1T1R trade-offs

[10] "Optimizing Hardware-Software Co-Design"
     Science China, 2025
     -> IR-drop and non-ideality analysis

[11] NeuroSim (Georgia Tech)
     -> Circuit-level CIM modeling methodology


DISCLAIMER
==========
This project is NOT affiliated with or endorsed by:
  * external research institution
  * Dr. external research group or Tour Lab
  * SkyWater Technology
  * Google
  * Any foundry or research institution

All references are to publicly available
published research. We implement SIMULATIONS
based on this published data, not validated
against actual hardware.

For the full reference list with DOIs and
paper links, see: docs/eda/REFERENCES.md`)
	content.Wrapping = fyne.TextWrapWord

	// Add a visual separator
	line := canvas.NewLine(theme.ForegroundColor())
	line.StrokeWidth = 1

	return container.NewVBox(title, widget.NewSeparator(), content)
}
