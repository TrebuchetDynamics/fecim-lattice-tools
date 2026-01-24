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
	// Diagram dimensions - INCREASED 40%
	boxW := float32(120)
	boxH := float32(55)
	spacing := float32(25)
	startX := float32(30)
	startY := float32(50)

	objects := []fyne.CanvasObject{}

	// Title
	title := canvas.NewText("OpenLane RTL-to-GDSII Flow", colorText)
	title.TextSize = 18
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(startX, 8))
	objects = append(objects, title)

	// Subtitle
	if showOurContribution {
		subtitle := canvas.NewText("CYAN = Our Array Builder files inject here", colorBoxOurs)
		subtitle.TextSize = 13
		subtitle.TextStyle = fyne.TextStyle{Bold: true}
		subtitle.Move(fyne.NewPos(startX, 30))
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

	// Row 2 (reversed flow) - more vertical space
	row2Y := startY + boxH + spacing + 30
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
		nameText.TextSize = 13
		nameText.TextStyle = fyne.TextStyle{Bold: true}
		nameText.Move(fyne.NewPos(stage.x+10, stage.y+10))
		objects = append(objects, nameText)

		// Tool name
		toolText := canvas.NewText(stage.tool, colorTextMuted)
		toolText.TextSize = 11
		toolText.Move(fyne.NewPos(stage.x+10, stage.y+30))
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

	// Container with fixed size - INCREASED
	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(620, 250))

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

	// Isometric projection parameters - INCREASED 40%
	cellSize := float32(42)
	isoAngle := float32(30 * math.Pi / 180) // 30 degrees
	cosA := float32(math.Cos(float64(isoAngle)))
	sinA := float32(math.Sin(float64(isoAngle)))

	// Starting position
	startX := float32(180)
	startY := float32(60)

	// Layer separation (Z height) - INCREASED
	layerGap := float32(55)

	// Convert grid coordinates to isometric
	toIso := func(gridX, gridY, z float32) (float32, float32) {
		x := startX + (gridX-gridY)*cellSize*cosA
		y := startY + (gridX+gridY)*cellSize*sinA - z
		return x, y
	}

	// Title
	title := canvas.NewText("PASSIVE Crossbar Structure", colorText)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(20, 10))
	objects = append(objects, title)

	// Draw bottom layer (Word Lines - horizontal in real space)
	for i := 0; i <= rows; i++ {
		x1, y1 := toIso(0, float32(i), 0)
		x2, y2 := toIso(float32(cols), float32(i), 0)

		line := canvas.NewLine(colorWL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		// WL label
		if showLabels && i < rows {
			label := canvas.NewText("WL"+string(rune('0'+i)), colorWL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-35, y1-5))
			objects = append(objects, label)
		}
	}

	// Draw top layer (Bit Lines - vertical in real space)
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, layerGap)
		x2, y2 := toIso(float32(j), float32(rows), layerGap)

		line := canvas.NewLine(colorBL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		// BL label
		if showLabels && j < cols {
			label := canvas.NewText("BL"+string(rune('0'+j)), colorBL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-5, y1-25))
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
			pillar.StrokeWidth = 5
			pillar.Position1 = fyne.NewPos(x1, y1)
			pillar.Position2 = fyne.NewPos(x2, y2)
			objects = append(objects, pillar)

			// Larger circle at center to represent the device
			midX := (x1 + x2) / 2
			midY := (y1 + y2) / 2
			device := canvas.NewCircle(colorFeFET)
			device.Resize(fyne.NewSize(12, 12))
			device.Move(fyne.NewPos(midX-6, midY-6))
			objects = append(objects, device)
		}
	}

	// Legend - moved down
	legendY := float32(240)
	legendX := float32(20)

	// WL legend
	wlBox := canvas.NewRectangle(colorWL)
	wlBox.Resize(fyne.NewSize(20, 12))
	wlBox.Move(fyne.NewPos(legendX, legendY))
	objects = append(objects, wlBox)
	wlText := canvas.NewText("Word Lines (WL) - Row Select", colorTextMuted)
	wlText.TextSize = 12
	wlText.Move(fyne.NewPos(legendX+25, legendY-2))
	objects = append(objects, wlText)

	// BL legend
	blBox := canvas.NewRectangle(colorBL)
	blBox.Resize(fyne.NewSize(20, 12))
	blBox.Move(fyne.NewPos(legendX, legendY+18))
	objects = append(objects, blBox)
	blText := canvas.NewText("Bit Lines (BL) - Data/Output", colorTextMuted)
	blText.TextSize = 12
	blText.Move(fyne.NewPos(legendX+25, legendY+16))
	objects = append(objects, blText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(14, 14))
	feBox.Move(fyne.NewPos(legendX+3, legendY+36))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight/data)", colorTextMuted)
	feText.TextSize = 12
	feText.Move(fyne.NewPos(legendX+25, legendY+36))
	objects = append(objects, feText)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(420, 310))

	return cont
}

// Isometric1T1RCrossbar creates an isometric view of a 1T1R crossbar array
func Isometric1T1RCrossbar(rows, cols int) fyne.CanvasObject {
	objects := []fyne.CanvasObject{}

	// Isometric projection parameters
	cellSize := float32(42)
	isoAngle := float32(30 * math.Pi / 180)
	cosA := float32(math.Cos(float64(isoAngle)))
	sinA := float32(math.Sin(float64(isoAngle)))

	startX := float32(180)
	startY := float32(60)
	layerGap := float32(55)

	toIso := func(gridX, gridY, z float32) (float32, float32) {
		x := startX + (gridX-gridY)*cellSize*cosA
		y := startY + (gridX+gridY)*cellSize*sinA - z
		return x, y
	}

	// Title
	title := canvas.NewText("1T1R Crossbar Structure", colorText)
	title.TextSize = 16
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Move(fyne.NewPos(20, 10))
	objects = append(objects, title)

	subtitle := canvas.NewText("(Transistor isolates each cell - NO sneak paths)", colorHighlight)
	subtitle.TextSize = 12
	subtitle.Move(fyne.NewPos(20, 30))
	objects = append(objects, subtitle)

	// Draw WL layer (bottom)
	for i := 0; i <= rows; i++ {
		x1, y1 := toIso(0, float32(i), 0)
		x2, y2 := toIso(float32(cols), float32(i), 0)

		line := canvas.NewLine(colorWL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		if i < rows {
			label := canvas.NewText("WL"+string(rune('0'+i)), colorWL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-35, y1-5))
			objects = append(objects, label)
		}
	}

	// Draw BL layer (top)
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, layerGap)
		x2, y2 := toIso(float32(j), float32(rows), layerGap)

		line := canvas.NewLine(colorBL)
		line.StrokeWidth = 4
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		if j < cols {
			label := canvas.NewText("BL"+string(rune('0'+j)), colorBL)
			label.TextSize = 12
			label.TextStyle = fyne.TextStyle{Bold: true}
			label.Move(fyne.NewPos(x1-5, y1-25))
			objects = append(objects, label)
		}
	}

	// Source Line color (green)
	colorSL := color.RGBA{100, 255, 100, 255}

	// Draw SL layer (middle - unique to 1T1R)
	slZ := layerGap / 2
	for j := 0; j <= cols; j++ {
		x1, y1 := toIso(float32(j), 0, slZ)
		x2, y2 := toIso(float32(j), float32(rows), slZ)

		line := canvas.NewLine(colorSL)
		line.StrokeWidth = 3
		line.Position1 = fyne.NewPos(x1, y1)
		line.Position2 = fyne.NewPos(x2, y2)
		objects = append(objects, line)

		if j < cols {
			label := canvas.NewText("SL"+string(rune('0'+j)), colorSL)
			label.TextSize = 10
			label.Move(fyne.NewPos(x2+5, y2-5))
			objects = append(objects, label)
		}
	}

	// Draw 1T1R cells (transistor symbol + FeFET)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			x1, y1 := toIso(float32(j)+0.5, float32(i)+0.5, 0)
			xMid, yMid := toIso(float32(j)+0.5, float32(i)+0.5, slZ)
			x2, y2 := toIso(float32(j)+0.5, float32(i)+0.5, layerGap)

			// Bottom part (transistor) - square symbol
			transistor := canvas.NewRectangle(colorHighlight)
			transistor.Resize(fyne.NewSize(10, 10))
			transistor.Move(fyne.NewPos(x1-5, y1-5))
			objects = append(objects, transistor)

			// Line from transistor to FeFET
			conn := canvas.NewLine(colorFeFET)
			conn.StrokeWidth = 3
			conn.Position1 = fyne.NewPos(x1, y1)
			conn.Position2 = fyne.NewPos(x2, y2)
			objects = append(objects, conn)

			// FeFET device (circle at top)
			device := canvas.NewCircle(colorFeFET)
			device.Resize(fyne.NewSize(12, 12))
			device.Move(fyne.NewPos(xMid-6, yMid-6))
			objects = append(objects, device)
		}
	}

	// Legend
	legendY := float32(240)
	legendX := float32(20)

	// Transistor legend
	tBox := canvas.NewRectangle(colorHighlight)
	tBox.Resize(fyne.NewSize(14, 14))
	tBox.Move(fyne.NewPos(legendX+3, legendY))
	objects = append(objects, tBox)
	tText := canvas.NewText("Select Transistor (gate on WL)", colorTextMuted)
	tText.TextSize = 12
	tText.Move(fyne.NewPos(legendX+25, legendY))
	objects = append(objects, tText)

	// SL legend
	slBox := canvas.NewRectangle(colorSL)
	slBox.Resize(fyne.NewSize(20, 12))
	slBox.Move(fyne.NewPos(legendX, legendY+18))
	objects = append(objects, slBox)
	slText := canvas.NewText("Source Lines (SL) - Current path", colorTextMuted)
	slText.TextSize = 12
	slText.Move(fyne.NewPos(legendX+25, legendY+18))
	objects = append(objects, slText)

	// FeFET legend
	feBox := canvas.NewCircle(colorFeFET)
	feBox.Resize(fyne.NewSize(14, 14))
	feBox.Move(fyne.NewPos(legendX+3, legendY+36))
	objects = append(objects, feBox)
	feText := canvas.NewText("FeFET Device (stores weight)", colorTextMuted)
	feText.TextSize = 12
	feText.Move(fyne.NewPos(legendX+25, legendY+36))
	objects = append(objects, feText)

	cont := container.NewWithoutLayout(objects...)
	cont.Resize(fyne.NewSize(420, 310))

	return cont
}

// =============================================================================
// FILE FORMAT PREVIEW CARDS
// =============================================================================

// FileFormatCard creates a styled card showing file format examples
func FileFormatCard(title, format, content string) fyne.CanvasObject {
	// Header - LARGER
	headerBg := canvas.NewRectangle(colorBoxOurs)
	headerBg.Resize(fyne.NewSize(340, 32))
	headerBg.CornerRadius = 6

	titleText := canvas.NewText(title+" (."+format+")", colorBgDark)
	titleText.TextSize = 14
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	titleText.Move(fyne.NewPos(12, 6))

	// Content area - LARGER
	contentBg := canvas.NewRectangle(color.RGBA{0, 25, 50, 255})
	contentBg.Resize(fyne.NewSize(340, 140))
	contentBg.Move(fyne.NewPos(0, 32))
	contentBg.CornerRadius = 4

	// Code content with monospace style
	codeLabel := widget.NewLabel(content)
	codeLabel.Wrapping = fyne.TextWrapOff
	codeLabel.TextStyle = fyne.TextStyle{Monospace: true}

	codeContainer := container.NewPadded(codeLabel)
	codeContainer.Move(fyne.NewPos(4, 36))
	codeContainer.Resize(fyne.NewSize(332, 132))

	card := container.NewWithoutLayout(headerBg, titleText, contentBg, codeContainer)
	card.Resize(fyne.NewSize(340, 175))

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
