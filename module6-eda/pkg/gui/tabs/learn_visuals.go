// pkg/gui/tabs/learn_visuals.go
// Visual components for the Learn tab - OpenLane flow diagram and crossbar visualization

package tabs

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Colors for the diagrams
var (
	colorBgDark      = color.RGBA{0, 30, 60, 255}
	colorBoxStandard = color.RGBA{0, 80, 160, 255}
	colorBoxOurs     = color.RGBA{0, 180, 220, 255} // Cyan for our contribution
	colorArrow       = color.RGBA{100, 150, 200, 255}
	colorText        = color.RGBA{255, 255, 255, 255}
	colorTextMuted   = color.RGBA{180, 200, 220, 255}
	colorWL          = color.RGBA{255, 100, 100, 255} // Red for word lines
	colorBL          = color.RGBA{100, 200, 255, 255} // Blue for bit lines
	colorFeFET       = color.RGBA{255, 200, 50, 255}  // Gold for FeFET devices
	colorHighlight   = color.RGBA{0, 255, 150, 255}   // Green highlight
)

// =============================================================================
// OPENLANE FLOW DIAGRAM
// =============================================================================

// OpenLaneFlowDiagram creates a visual pipeline diagram
func OpenLaneFlowDiagram(showOurContribution bool) fyne.CanvasObject {
	// Diagram dimensions
	boxW := float32(100)
	boxH := float32(50)
	spacing := float32(30)
	startX := float32(20)
	startY := float32(60)

	objects := []fyne.CanvasObject{}

	// Title
	title := canvas.NewText("OpenLane RTL-to-GDSII Flow", colorText)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(startX, 10))
	objects = append(objects, title)

	// Subtitle
	if showOurContribution {
		subtitle := canvas.NewText("Cyan = Our Array Builder contribution", colorBoxOurs)
		subtitle.TextSize = 12
		subtitle.Move(fyne.NewPos(startX, 32))
		objects = append(objects, subtitle)
	}

	// Pipeline stages - row 1
	stages := []struct {
		name    string
		tool    string
		isOurs  bool
		x, y    float32
	}{
		{"Verilog", "Input", false, startX, startY},
		{"Synthesis", "Yosys", false, startX + boxW + spacing, startY},
		{"Floorplan", "OpenROAD", false, startX + 2*(boxW+spacing), startY},
		{"Placement", "RePlAce", false, startX + 3*(boxW+spacing), startY},
	}

	// Row 2 (reversed flow)
	row2Y := startY + boxH + spacing + 20
	stages = append(stages, []struct {
		name    string
		tool    string
		isOurs  bool
		x, y    float32
	}{
		{"CTS", "TritonCTS", false, startX + 3*(boxW+spacing), row2Y},
		{"Routing", "TritonRoute", false, startX + 2*(boxW+spacing), row2Y},
		{"Signoff", "Magic/LVS", false, startX + boxW + spacing, row2Y},
		{"GDSII", "Output", false, startX, row2Y},
	}...)

	// Mark which stages we contribute to
	if showOurContribution {
		stages[0].isOurs = true  // Verilog - we provide
		stages[2].isOurs = true  // Floorplan - our LEF
		stages[3].isOurs = true  // Placement - our DEF (FIXED)
	}

	// Draw stages
	for i, stage := range stages {
		boxColor := colorBoxStandard
		if stage.isOurs && showOurContribution {
			boxColor = colorBoxOurs
		}

		// Box
		box := canvas.NewRectangle(boxColor)
		box.Resize(fyne.NewSize(boxW, boxH))
		box.Move(fyne.NewPos(stage.x, stage.y))
		box.CornerRadius = 6
		objects = append(objects, box)

		// Stage name
		nameText := canvas.NewText(stage.name, colorText)
		nameText.TextSize = 12
		nameText.TextStyle = fyne.TextStyle{Bold: true}
		nameText.Move(fyne.NewPos(stage.x+8, stage.y+8))
		objects = append(objects, nameText)

		// Tool name
		toolText := canvas.NewText(stage.tool, colorTextMuted)
		toolText.TextSize = 10
		toolText.Move(fyne.NewPos(stage.x+8, stage.y+28))
		objects = append(objects, toolText)

		// Draw arrows between stages
		if i > 0 && i < 4 {
			// Row 1 arrows (left to right)
			arrow := createArrow(
				stages[i-1].x+boxW, stages[i-1].y+boxH/2,
				stage.x, stage.y+boxH/2,
			)
			objects = append(objects, arrow...)
		} else if i == 4 {
			// Down arrow from Placement to CTS
			arrow := createArrowDown(
				stages[3].x+boxW/2, stages[3].y+boxH,
				stage.x+boxW/2, stage.y,
			)
			objects = append(objects, arrow...)
		} else if i > 4 {
			// Row 2 arrows (right to left)
			arrow := createArrow(
				stages[i-1].x, stages[i-1].y+boxH/2,
				stage.x+boxW, stage.y+boxH/2,
			)
			objects = append(objects, arrow...)
		}
	}

	// Add "Our Files" labels if showing contribution
	if showOurContribution {
		// LEF label
		lefLabel := canvas.NewText("Our LEF", colorHighlight)
		lefLabel.TextSize = 10
		lefLabel.Move(fyne.NewPos(stages[2].x+10, stages[2].y-15))
		objects = append(objects, lefLabel)

		// DEF label
		defLabel := canvas.NewText("Our DEF (FIXED)", colorHighlight)
		defLabel.TextSize = 10
		defLabel.Move(fyne.NewPos(stages[3].x+5, stages[3].y-15))
		objects = append(objects, defLabel)

		// Verilog label
		vLabel := canvas.NewText("Our Verilog", colorHighlight)
		vLabel.TextSize = 10
		vLabel.Move(fyne.NewPos(stages[0].x+5, stages[0].y-15))
		objects = append(objects, vLabel)
	}

	// Container with fixed size
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(500, 220))

	return cont
}

// createArrow creates a horizontal arrow
func createArrow(x1, y1, x2, y2 float32) []fyne.CanvasObject {
	// Line
	line := canvas.NewLine(colorArrow)
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2-8, y2)

	// Arrowhead
	head := canvas.NewLine(colorArrow)
	head.StrokeWidth = 2
	head.Position1 = fyne.NewPos(x2-12, y2-5)
	head.Position2 = fyne.NewPos(x2-4, y2)

	head2 := canvas.NewLine(colorArrow)
	head2.StrokeWidth = 2
	head2.Position1 = fyne.NewPos(x2-12, y2+5)
	head2.Position2 = fyne.NewPos(x2-4, y2)

	return []fyne.CanvasObject{line, head, head2}
}

// createArrowDown creates a vertical down arrow
func createArrowDown(x1, y1, x2, y2 float32) []fyne.CanvasObject {
	line := canvas.NewLine(colorArrow)
	line.StrokeWidth = 2
	line.Position1 = fyne.NewPos(x1, y1)
	line.Position2 = fyne.NewPos(x2, y2-8)

	head := canvas.NewLine(colorArrow)
	head.StrokeWidth = 2
	head.Position1 = fyne.NewPos(x2-5, y2-12)
	head.Position2 = fyne.NewPos(x2, y2-4)

	head2 := canvas.NewLine(colorArrow)
	head2.StrokeWidth = 2
	head2.Position1 = fyne.NewPos(x2+5, y2-12)
	head2.Position2 = fyne.NewPos(x2, y2-4)

	return []fyne.CanvasObject{line, head, head2}
}

// =============================================================================
// ISOMETRIC CROSSBAR DIAGRAM
// =============================================================================

// IsometricCrossbar creates an isometric view of a crossbar array
func IsometricCrossbar(rows, cols int, showLabels bool) fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Isometric projection parameters
	cellSize := float32(30)
	isoAngle := float32(30 * math.Pi / 180) // 30 degrees
	cosA := float32(math.Cos(float64(isoAngle)))
	sinA := float32(math.Sin(float64(isoAngle)))

	// Starting position
	startX := float32(150)
	startY := float32(50)

	// Layer separation (Z height)
	layerGap := float32(40)

	// Convert grid coordinates to isometric
	toIso := func(gridX, gridY, z float32) (float32, float32) {
		x := startX + (gridX-gridY)*cellSize*cosA
		y := startY + (gridX+gridY)*cellSize*sinA - z
		return x, y
	}

	// Title
	title := canvas.NewText("Crossbar Array Structure (Isometric View)", colorText)
	title.TextSize = 14
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(20, 10))
	objects = append(objects, title)

	// Draw bottom layer (Word Lines - horizontal in real space)
	for i := 0; i <= rows; i++ {
		x1, y1 := toIso(0, float32(i), 0)
		x2, y2 := toIso(float32(cols), float32(i), 0)

		line := canvas.NewLine(colorWL)
		line.StrokeWidth = 3
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		// WL label
		if showLabels && i < rows {
			label := canvas.NewText("WL"+string(rune('0'+i)), colorWL)
			label.TextSize = 10
			label.Move(fyne.NewPos(x1-25, y1-5))
			objects = append(objects, label)
		}
	}

	// Draw top layer (Bit Lines - vertical in real space)
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, layerGap)
		x2, y2 := toIso(float32(j), float32(rows), layerGap)

		line := canvas.NewLine(colorBL)
		line.StrokeWidth = 3
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		// BL label
		if showLabels && j < cols {
			label := canvas.NewText("BL"+string(rune('0'+j)), colorBL)
			label.TextSize = 10
			label.Move(fyne.NewPos(x1-5, y1-20))
			objects = append(objects, label)
		}
	}

	// Draw FeFET devices at intersections (vertical pillars)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Bottom connection point (on WL layer)
			x1, y1 := toIso(float32(j)+0.5, float32(i)+0.5, 0)
			// Top connection point (on BL layer)
			x2, y2 := toIso(float32(j)+0.5, float32(i)+0.5, layerGap)

			// Vertical connector (the FeFET)
			pillar := canvas.NewLine(colorFeFET)
			pillar.StrokeWidth = 4
			pillar.Position1 = fyne.NewPos(x1, y1)
			pillar.Position2 = fyne.NewPos(x2, y2)
			objects = append(objects, pillar)

			// Small circle at center to represent the device
			midX := (x1 + x2) / 2
			midY := (y1 + y2) / 2
			device := canvas.NewCircle(colorFeFET)
			device.Resize(fyne.NewSize(8, 8))
			device.Move(fyne.NewPos(midX-4, midY-4))
			objects = append(objects, device)
		}
	}

	// Legend
	legendY := float32(180)
	legendX := float32(20)

	// WL legend
	wlBox := canvas.NewRectangle(colorWL)
	wlBox.Resize(fyne.NewSize(15, 10))
	wlBox.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, wlBox)
	wlText := canvas.NewText("Word Lines (WL) - Row Select", colorTextMuted)
	wlText.TextSize = 10
	wlText.Move(fyne.NewPos(legendX+20, legendY-2))
	objects = append(objects, wlText)

	// BL legend
	blBox := canvas.NewRectangle(colorBL)
	blBox.Resize(fyne.NewSize(15, 10))
	blBox.Move(fyne.NewPos(legendX, legendY+15))
	objects = append(objects, blBox)
	blText := canvas.NewText("Bit Lines (BL) - Data/Output", colorTextMuted)
	blText.TextSize = 10
	blText.Move(fyne.NewPos(legendX+20, legendY+13))
	objects = append(objects, blText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(10, 10))
	feBox.Move(fyne.NewPos(legendX+2, legendY+30))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight/data)", colorTextMuted)
	feText.TextSize = 10
	feText.Move(fyne.NewPos(legendX+20, legendY+28))
	objects = append(objects, feText)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(350, 250))

	return cont
}

// =============================================================================
// FILE FORMAT PREVIEW CARDS
// =============================================================================

// FileFormatCard creates a styled card showing file format examples
func FileFormatCard(title, format, content string) fyne.CanvasObject {
	// Header
	headerBg := canvas.NewRectangle(colorBoxOurs)
	headerBg.Resize(fyne.NewSize(280, 25))
	headerBg.CornerRadius = 4

	titleText := canvas.NewText(title+" (."+format+")", colorBgDark)
	titleText.TextSize = 12
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Move(fyne.NewPos(8, 4))

	// Content area
	contentBg := canvas.NewRectangle(color.RGBA{0, 20, 40, 255})
	contentBg.Resize(fyne.NewSize(280, 100))
	contentBg.Move(fyne.NewPos(0, 25))

	// Code content
	codeText := widget.NewLabel(content)
	codeText.Wrapping = fyne.TextWrapOff
	codeText.Move(fyne.NewPos(8, 30))

	card := container.NewWithoutLayout(headerBg, titleText, contentBg, codeText)
	card.Resize(fyne.NewSize(280, 125))

	return card
}

// LEFPreviewCard shows a LEF file example
func LEFPreviewCard() fyne.CanvasObject {
	content := `MACRO fecim_bitcell
  SIZE 0.460 BY 2.720 ;
  PIN WL
    DIRECTION INPUT ;
  END WL
END fecim_bitcell`
	return FileFormatCard("LEF", "lef", content)
}

// DEFPreviewCard shows a DEF file example
func DEFPreviewCard() fyne.CanvasObject {
	content := `COMPONENTS 16 ;
  - cell_0_0 fecim_bitcell
    + FIXED ( 10000 10000 ) N ;
  - cell_0_1 fecim_bitcell
    + FIXED ( 10460 10000 ) N ;
END COMPONENTS`
	return FileFormatCard("DEF", "def", content)
}

// VerilogPreviewCard shows a Verilog file example
func VerilogPreviewCard() fyne.CanvasObject {
	content := `module fecim_array_4x4 (
  input  [3:0] WL,
  inout  [3:0] BL
);
  fecim_bitcell c00(.WL(WL[0]),.BL(BL[0]));
  // ... 16 cells
endmodule`
	return FileFormatCard("Verilog", "v", content)
}
