// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// EnergyBarChart displays energy per MAC comparison.
type EnergyBarChart struct {
	widget.BaseWidget

	cpuSpec   EnergySpec
	gpuSpec   EnergySpec
	fecimSpec EnergySpec

	raster *canvas.Raster
}

// NewEnergyBarChart creates a new energy bar chart.
func NewEnergyBarChart() *EnergyBarChart {
	e := &EnergyBarChart{}
	e.ExtendBaseWidget(e)
	return e
}

// SetValues updates the energy specifications.
func (e *EnergyBarChart) SetValues(cpu, gpu, fecim EnergySpec) {
	e.cpuSpec = cpu
	e.gpuSpec = gpu
	e.fecimSpec = fecim
	e.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (e *EnergyBarChart) CreateRenderer() fyne.WidgetRenderer {
	e.raster = canvas.NewRaster(e.generateImage)

	// Create source labels
	cpuSourceLabel := widget.NewLabel(fmt.Sprintf("[1] %s", e.cpuSpec.Source))
	cpuSourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	gpuSourceLabel := widget.NewLabel(fmt.Sprintf("[2] %s", e.gpuSpec.Source))
	gpuSourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	fecimSourceLabel := widget.NewLabel(fmt.Sprintf("[3] %s", e.fecimSpec.Source))
	fecimSourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	sources := container.NewVBox(
		cpuSourceLabel,
		gpuSourceLabel,
		fecimSourceLabel,
	)

	content := container.NewBorder(
		nil,
		sources,
		nil, nil,
		container.NewMax(e.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (e *EnergyBarChart) MinSize() fyne.Size {
	return fyne.NewSize(400, 200)
}

// generateImage creates the bar chart.
func (e *EnergyBarChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 100 || h < 100 {
		return img
	}

	// Chart dimensions
	padding := 60
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding - 40 // Leave room for labels

	// Max value for scaling (CPU is largest)
	maxVal := e.cpuSpec.EnergyFJ
	if maxVal == 0 {
		maxVal = 1000
	}

	// Bar dimensions
	barHeight := chartHeight / 4
	barSpacing := chartHeight / 4

	// Colors
	cpuColor := color.RGBA{200, 100, 100, 255}   // Red
	gpuColor := color.RGBA{200, 180, 100, 255}   // Yellow
	fecimColor := color.RGBA{100, 200, 150, 255} // Green

	// Draw CPU bar
	cpuWidth := int(float64(chartWidth) * e.cpuSpec.EnergyFJ / maxVal)
	drawBar(img, padding, padding, cpuWidth, barHeight, cpuColor)

	// Draw GPU bar
	gpuWidth := int(float64(chartWidth) * e.gpuSpec.EnergyFJ / maxVal)
	drawBar(img, padding, padding+barSpacing, gpuWidth, barHeight, gpuColor)

	// Draw FeCIM bar
	fecimWidth := int(float64(chartWidth) * e.fecimSpec.EnergyFJ / maxVal)
	if fecimWidth < 5 {
		fecimWidth = 5 // Minimum visible width
	}
	drawBar(img, padding, padding+2*barSpacing, fecimWidth, barHeight, fecimColor)

	// Draw axis
	axisColor := color.RGBA{100, 120, 150, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}
	for y := padding; y < h-padding; y++ {
		img.Set(padding, y, axisColor)
	}

	return img
}

// drawBar draws a filled rectangle.
func drawBar(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			img.Set(x+dx, y+dy, c)
		}
	}
}

// ArchitectureDiagram shows Von Neumann vs CIM architecture.
type ArchitectureDiagram struct {
	widget.BaseWidget
	raster *canvas.Raster
}

// NewArchitectureDiagram creates a new architecture diagram.
func NewArchitectureDiagram() *ArchitectureDiagram {
	a := &ArchitectureDiagram{}
	a.ExtendBaseWidget(a)
	return a
}

// CreateRenderer implements fyne.Widget.
func (a *ArchitectureDiagram) CreateRenderer() fyne.WidgetRenderer {
	a.raster = canvas.NewRaster(a.generateImage)

	titleLabel := widget.NewLabel("Why the Difference?")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(
		titleLabel,
		nil, nil, nil,
		container.NewMax(a.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (a *ArchitectureDiagram) MinSize() fyne.Size {
	return fyne.NewSize(400, 150)
}

// generateImage creates the architecture comparison diagram.
func (a *ArchitectureDiagram) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 200 || h < 80 {
		return img
	}

	// Left side: Von Neumann (CPU/GPU)
	// Draw CPU box
	cpuColor := color.RGBA{200, 100, 100, 255}
	drawBox(img, 30, 30, 60, 40, cpuColor)

	// Draw DRAM box
	dramColor := color.RGBA{100, 100, 200, 255}
	drawBox(img, 130, 30, 60, 40, dramColor)

	// Draw arrow between them (bidirectional)
	arrowColor := color.RGBA{255, 200, 100, 255}
	for x := 95; x < 125; x++ {
		img.Set(x, 50, arrowColor)
		img.Set(x, 51, arrowColor)
	}

	// Right side: CIM (FeCIM)
	// Draw combined box
	cimColor := color.RGBA{100, 200, 150, 255}
	midX := w/2 + 50
	drawBox(img, midX, 30, 100, 40, cimColor)

	// Draw "NO DATA MOVEMENT" text indicator (simple line)
	noMoveColor := color.RGBA{100, 255, 150, 255}
	for y := 75; y < 85; y++ {
		for x := midX; x < midX+100; x++ {
			img.Set(x, y, noMoveColor)
		}
	}

	return img
}

// drawBox draws a rectangle outline.
func drawBox(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	// Fill
	fillColor := color.RGBA{c.R / 2, c.G / 2, c.B / 2, 255}
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			img.Set(x+dx, y+dy, fillColor)
		}
	}
	// Border
	for dx := 0; dx < width; dx++ {
		img.Set(x+dx, y, c)
		img.Set(x+dx, y+height-1, c)
	}
	for dy := 0; dy < height; dy++ {
		img.Set(x, y+dy, c)
		img.Set(x+width-1, y+dy, c)
	}
}

// DataCenterCalculator shows power and cost calculations.
type DataCenterCalculator struct {
	widget.BaseWidget

	workloadLabel   *widget.Label
	macsLabel       *widget.Label
	inferencesLabel *widget.Label

	cpuEnergyLabel   *widget.Label
	gpuEnergyLabel   *widget.Label
	fecimEnergyLabel *widget.Label

	cpuPowerLabel   *widget.Label
	gpuPowerLabel   *widget.Label
	fecimPowerLabel *widget.Label

	cpuCostLabel   *widget.Label
	gpuCostLabel   *widget.Label
	fecimCostLabel *widget.Label

	savingsLabel *widget.Label
}

// NewDataCenterCalculator creates a new calculator widget.
func NewDataCenterCalculator() *DataCenterCalculator {
	d := &DataCenterCalculator{}
	d.workloadLabel = widget.NewLabel("Workload: -")
	d.macsLabel = widget.NewLabel("MACs/inference: -")
	d.inferencesLabel = widget.NewLabel("Inferences/sec: -")

	d.cpuEnergyLabel = widget.NewLabel("CPU: - µJ/inf")
	d.gpuEnergyLabel = widget.NewLabel("GPU: - µJ/inf")
	d.fecimEnergyLabel = widget.NewLabel("FeCIM: - µJ/inf")

	d.cpuPowerLabel = widget.NewLabel("CPU: - W")
	d.gpuPowerLabel = widget.NewLabel("GPU: - W")
	d.fecimPowerLabel = widget.NewLabel("FeCIM: - W")

	d.cpuCostLabel = widget.NewLabel("CPU: $-/month")
	d.gpuCostLabel = widget.NewLabel("GPU: $-/month")
	d.fecimCostLabel = widget.NewLabel("FeCIM: $-/month")

	d.savingsLabel = widget.NewLabel("Savings: -")
	d.savingsLabel.TextStyle = fyne.TextStyle{Bold: true}

	d.ExtendBaseWidget(d)
	return d
}

// SetResults updates the calculator with new results.
func (d *DataCenterCalculator) SetResults(
	workload string,
	macs int,
	inferences float64,
	cpuEnergy, gpuEnergy, fecimEnergy float64,
	cpuPower, gpuPower, fecimPower float64,
	cpuCost, gpuCost, fecimCost float64,
) {
	d.workloadLabel.SetText(fmt.Sprintf("Workload: %s", workload))
	d.macsLabel.SetText(fmt.Sprintf("MACs/inference: %s", formatNumberWithSuffix(float64(macs))))
	d.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", inferences))

	d.cpuEnergyLabel.SetText(fmt.Sprintf("CPU: %.2f µJ/inf [1]", cpuEnergy))
	d.gpuEnergyLabel.SetText(fmt.Sprintf("GPU: %.2f µJ/inf [2]", gpuEnergy))
	d.fecimEnergyLabel.SetText(fmt.Sprintf("FeCIM: %.2f µJ/inf [3]*", fecimEnergy))

	d.cpuPowerLabel.SetText(fmt.Sprintf("CPU: %.1f W", cpuPower))
	d.gpuPowerLabel.SetText(fmt.Sprintf("GPU: %.1f W", gpuPower))
	d.fecimPowerLabel.SetText(fmt.Sprintf("FeCIM: %.2f W*", fecimPower))

	d.cpuCostLabel.SetText(fmt.Sprintf("CPU: $%.0f/month", cpuCost))
	d.gpuCostLabel.SetText(fmt.Sprintf("GPU: $%.0f/month", gpuCost))
	d.fecimCostLabel.SetText(fmt.Sprintf("FeCIM: $%.0f/month*", fecimCost))

	savingsVsGPU := (gpuCost - fecimCost) / gpuCost * 100
	d.savingsLabel.SetText(fmt.Sprintf("Savings vs GPU: %.0f%% (if claims hold)*", savingsVsGPU))

	d.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (d *DataCenterCalculator) CreateRenderer() fyne.WidgetRenderer {
	// Configuration section
	configBox := container.NewVBox(
		d.workloadLabel,
		d.macsLabel,
		d.inferencesLabel,
	)

	// Energy section
	energyLabel := widget.NewLabel("Energy per Inference:")
	energyLabel.TextStyle = fyne.TextStyle{Bold: true}
	energyBox := container.NewVBox(
		energyLabel,
		d.cpuEnergyLabel,
		d.gpuEnergyLabel,
		d.fecimEnergyLabel,
	)

	// Power section
	powerLabel := widget.NewLabel("Total Power:")
	powerLabel.TextStyle = fyne.TextStyle{Bold: true}
	powerBox := container.NewVBox(
		powerLabel,
		d.cpuPowerLabel,
		d.gpuPowerLabel,
		d.fecimPowerLabel,
	)

	// Cost section
	costLabel := widget.NewLabel("Monthly Cost (@$0.10/kWh):")
	costLabel.TextStyle = fyne.TextStyle{Bold: true}
	costBox := container.NewVBox(
		costLabel,
		d.cpuCostLabel,
		d.gpuCostLabel,
		d.fecimCostLabel,
		widget.NewSeparator(),
		d.savingsLabel,
	)

	// Disclaimer
	disclaimer := widget.NewLabel("* FeCIM estimates based on Dr. Tour's claims. NOT verified.")
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		configBox,
		widget.NewSeparator(),
		container.NewGridWithColumns(3, energyBox, powerBox, costBox),
		widget.NewSeparator(),
		disclaimer,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (d *DataCenterCalculator) MinSize() fyne.Size {
	return fyne.NewSize(500, 250)
}

// VerifiedClaimsTable shows what's verified vs claimed.
type VerifiedClaimsTable struct {
	widget.BaseWidget
}

// NewVerifiedClaimsTable creates a new table.
func NewVerifiedClaimsTable() *VerifiedClaimsTable {
	v := &VerifiedClaimsTable{}
	v.ExtendBaseWidget(v)
	return v
}

// CreateRenderer implements fyne.Widget.
func (v *VerifiedClaimsTable) CreateRenderer() fyne.WidgetRenderer {
	titleLabel := widget.NewLabel("Verified vs Claimed")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	verifiedLabel := widget.NewLabel("VERIFIED (from Dr. Tour):")
	verifiedLabel.TextStyle = fyne.TextStyle{Bold: true}

	verified := container.NewVBox(
		widget.NewLabel("  30 discrete analog levels"),
		widget.NewLabel("  87% MNIST accuracy"),
		widget.NewLabel("  CMOS compatible fab"),
		widget.NewLabel("  Non-volatile (no refresh)"),
	)

	claimedLabel := widget.NewLabel("CLAIMED (not verified):")
	claimedLabel.TextStyle = fyne.TextStyle{Bold: true}

	claimed := container.NewVBox(
		widget.NewLabel("  10M× lower than NAND"),
		widget.NewLabel("  1000× lower than DRAM"),
		widget.NewLabel("  80-90% DC energy savings"),
	)

	statusLabel := widget.NewLabel("Status: TRL 4 (Lab only)")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true, Italic: true}

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		verifiedLabel,
		verified,
		widget.NewSeparator(),
		claimedLabel,
		claimed,
		widget.NewSeparator(),
		statusLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (v *VerifiedClaimsTable) MinSize() fyne.Size {
	return fyne.NewSize(200, 280)
}

// formatNumberWithSuffix formats large numbers with K, M, B, T suffixes.
func formatNumberWithSuffix(n float64) string {
	switch {
	case n >= 1e12:
		return fmt.Sprintf("%.1fT", n/1e12)
	case n >= 1e9:
		return fmt.Sprintf("%.1fB", n/1e9)
	case n >= 1e6:
		return fmt.Sprintf("%.1fM", n/1e6)
	case n >= 1e3:
		return fmt.Sprintf("%.1fK", n/1e3)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}

// DataCenterTransformation shows before/after data center comparison.
type DataCenterTransformation struct {
	widget.BaseWidget

	mu             sync.RWMutex
	beforePowerW   float64
	afterPowerW    float64
	savingsPercent float64
	animProgress   float64
	raster         *canvas.Raster
	minSize        fyne.Size
}

// NewDataCenterTransformation creates a new transformation visual.
func NewDataCenterTransformation() *DataCenterTransformation {
	d := &DataCenterTransformation{
		beforePowerW:   1000,
		afterPowerW:    100,
		savingsPercent: 90,
		minSize:        fyne.NewSize(450, 120),
	}
	d.ExtendBaseWidget(d)
	return d
}

// SetValues updates the before/after values.
func (d *DataCenterTransformation) SetValues(beforeW, afterW float64) {
	d.mu.Lock()
	d.beforePowerW = beforeW
	d.afterPowerW = afterW
	if beforeW > 0 {
		d.savingsPercent = (beforeW - afterW) / beforeW * 100
	}
	d.mu.Unlock()
	d.Refresh()
}

// UpdateAnimation advances the animation.
func (d *DataCenterTransformation) UpdateAnimation(dt float64) {
	d.mu.Lock()
	d.animProgress += dt
	d.mu.Unlock()
}

// MinSize returns minimum size.
func (d *DataCenterTransformation) MinSize() fyne.Size {
	return d.minSize
}

// CreateRenderer implements fyne.Widget.
func (d *DataCenterTransformation) CreateRenderer() fyne.WidgetRenderer {
	d.raster = canvas.NewRaster(d.generateImage)
	return widget.NewSimpleRenderer(d.raster)
}

// generateImage creates the transformation visual.
func (d *DataCenterTransformation) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 300 || h < 80 {
		return img
	}

	d.mu.RLock()
	beforePower := d.beforePowerW
	afterPower := d.afterPowerW
	savings := d.savingsPercent
	animProgress := d.animProgress
	d.mu.RUnlock()

	midX := w / 2

	// Title
	d.drawTextOnImage(img, "DATA CENTER TRANSFORMATION", midX-90, 8, color.RGBA{0, 212, 255, 255}, 11)

	// LEFT: Before (GPU)
	d.drawTextOnImage(img, "BEFORE (GPU)", 20, 25, color.RGBA{200, 100, 100, 255}, 10)

	// Server rack icons (simplified as rectangles)
	rackY := 45
	rackH := 40
	rackW := 15
	beforeRacks := 10

	for i := 0; i < beforeRacks; i++ {
		rackX := 20 + i*(rackW+3)
		rackColor := color.RGBA{150, 80, 80, 255}
		// Draw rack
		for dy := 0; dy < rackH; dy++ {
			for dx := 0; dx < rackW; dx++ {
				img.Set(rackX+dx, rackY+dy, rackColor)
			}
		}
		// Highlight lines (server slots)
		for slot := 0; slot < 4; slot++ {
			slotY := rackY + 5 + slot*10
			for dx := 2; dx < rackW-2; dx++ {
				img.Set(rackX+dx, slotY, color.RGBA{200, 100, 100, 255})
			}
		}
	}

	// Power label
	powerText := fmt.Sprintf("%.0f W", beforePower)
	d.drawTextOnImage(img, powerText, 20, rackY+rackH+8, color.RGBA{255, 200, 100, 255}, 10)

	// ARROW
	arrowX := midX - 30
	arrowY := h / 2
	arrowColor := color.RGBA{100, 200, 150, 255}

	// Animated arrow
	arrowAlpha := uint8(150 + int(50*((animProgress*2)-float64(int(animProgress*2)))))
	arrowColor.A = arrowAlpha

	for ax := 0; ax < 40; ax++ {
		img.Set(arrowX+ax, arrowY, arrowColor)
		img.Set(arrowX+ax, arrowY+1, arrowColor)
	}
	// Arrow head
	for ay := -5; ay <= 5; ay++ {
		headX := arrowX + 40 - absIntWidget(ay)
		img.Set(headX, arrowY+ay, arrowColor)
	}

	// RIGHT: After (FeCIM)
	d.drawTextOnImage(img, "AFTER (FeCIM)", midX+30, 25, color.RGBA{100, 200, 150, 255}, 10)

	// Fewer server racks
	afterRacks := max(1, int(float64(beforeRacks)*(1-savings/100)))
	for i := 0; i < afterRacks; i++ {
		rackX := midX + 30 + i*(rackW+3)
		rackColor := color.RGBA{80, 150, 100, 255}
		// Draw rack
		for dy := 0; dy < rackH; dy++ {
			for dx := 0; dx < rackW; dx++ {
				img.Set(rackX+dx, rackY+dy, rackColor)
			}
		}
		// Highlight lines
		for slot := 0; slot < 4; slot++ {
			slotY := rackY + 5 + slot*10
			for dx := 2; dx < rackW-2; dx++ {
				img.Set(rackX+dx, slotY, color.RGBA{100, 200, 150, 255})
			}
		}
	}

	// Power label
	afterPowerText := fmt.Sprintf("%.0f W", afterPower)
	d.drawTextOnImage(img, afterPowerText, midX+30, rackY+rackH+8, color.RGBA{100, 255, 150, 255}, 10)

	// Savings headline
	savingsText := fmt.Sprintf("%.0f%% LESS INFRASTRUCTURE", savings)
	d.drawTextOnImage(img, savingsText, midX+afterRacks*(rackW+3)+50, h/2-5, color.RGBA{0, 212, 255, 255}, 11)

	return img
}

// drawTextOnImage draws simple text on an image (placeholder for font rendering).
func (d *DataCenterTransformation) drawTextOnImage(img *image.RGBA, text string, x, y int, c color.RGBA, fontSize int) {
	charWidth := fontSize / 2
	for i, ch := range text {
		if ch == ' ' {
			continue
		}
		cx := x + i*charWidth
		for dy := 0; dy < fontSize; dy++ {
			for dx := 0; dx < charWidth-1; dx++ {
				if cy := y + dy; cy >= 0 && cy < img.Bounds().Dy() {
					if ccx := cx + dx; ccx >= 0 && ccx < img.Bounds().Dx() {
						if dy > 1 && dy < fontSize-2 && dx > 0 && dx < charWidth-2 {
							img.Set(ccx, cy, c)
						}
					}
				}
			}
		}
	}
}

func absIntWidget(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
