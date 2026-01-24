// Package gui provides Fyne-based GUI components for MNIST visualization.
// comparison_card.go implements P1.2: Enhanced FP vs CIM Comparison Card
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ComparisonResult holds the results of FP vs CIM comparison.
type ComparisonResult struct {
	FPPrediction    int
	FPConfidence    float64
	FPProbabilities []float64

	CIMPrediction    int
	CIMConfidence    float64
	CIMProbabilities []float64

	Match           bool
	ConfidenceDelta float64
	EnergyFeCIM     float64 // nanojoules
	EnergyGPU       float64 // nanojoules
	EnergyRatio     float64 // GPU/FeCIM
}

// ComparisonCard provides enhanced FP vs CIM comparison visualization.
// This is the hero widget showing why FeCIM's accuracy-energy tradeoff matters.
type ComparisonCard struct {
	widget.BaseWidget

	mu     sync.RWMutex
	result *ComparisonResult

	// Visual components
	titleLabel  *widget.Label
	statusLabel *widget.Label
	raster      *canvas.Raster
}

// NewComparisonCard creates a new comparison card widget.
func NewComparisonCard() *ComparisonCard {
	cc := &ComparisonCard{}
	cc.titleLabel = widget.NewLabelWithStyle("FP vs CIM Comparison", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	cc.statusLabel = widget.NewLabel("Draw a digit to compare predictions")
	cc.ExtendBaseWidget(cc)
	return cc
}

// SetResult updates the comparison with new inference results.
func (cc *ComparisonCard) SetResult(result *ComparisonResult) {
	cc.mu.Lock()
	cc.result = result
	cc.mu.Unlock()

	// Update status
	if result != nil {
		if result.Match {
			cc.statusLabel.SetText(fmt.Sprintf("MATCH | Confidence Δ: %.1f%% | %.0fx energy savings",
				result.ConfidenceDelta*100, result.EnergyRatio))
		} else {
			cc.statusLabel.SetText(fmt.Sprintf("MISMATCH | FP: %d vs CIM: %d | Check hardware config!",
				result.FPPrediction, result.CIMPrediction))
		}
	}

	fyne.Do(func() {
		cc.Refresh()
	})
}

// Clear resets the card to idle state.
func (cc *ComparisonCard) Clear() {
	cc.mu.Lock()
	cc.result = nil
	cc.mu.Unlock()
	cc.statusLabel.SetText("Draw a digit to compare predictions")
	fyne.Do(func() {
		cc.Refresh()
	})
}

// MinSize returns the minimum size for the widget.
func (cc *ComparisonCard) MinSize() fyne.Size {
	return fyne.NewSize(500, 280)
}

// CreateRenderer implements fyne.Widget.
func (cc *ComparisonCard) CreateRenderer() fyne.WidgetRenderer {
	cc.raster = canvas.NewRaster(cc.generateImage)

	content := container.NewBorder(
		container.NewVBox(
			cc.titleLabel,
			widget.NewSeparator(),
		),
		cc.statusLabel,
		nil, nil,
		container.NewMax(cc.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the comparison visualization.
func (cc *ComparisonCard) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 500
	}
	if h < 10 {
		h = 220
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 30, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	cc.mu.RLock()
	result := cc.result
	cc.mu.RUnlock()

	if result == nil {
		// Draw idle state
		drawSimpleText(img, "Waiting for inference...", w/2-80, h/2-8, color.RGBA{100, 100, 120, 255})
		return img
	}

	// Layout constants
	padding := 15
	cardWidth := (w - 3*padding) / 2
	cardHeight := h - 80 // Leave room for bottom info

	// === FP Card (Left) ===
	fpX := padding
	fpY := padding
	cc.drawPredictionCard(img, fpX, fpY, cardWidth, cardHeight,
		"FP (Float32)", "Ideal AI",
		result.FPPrediction, result.FPConfidence,
		color.RGBA{100, 150, 255, 255}, // Blue
		result.FPProbabilities)

	// === CIM Card (Right) ===
	cimX := padding*2 + cardWidth
	cimY := padding
	cc.drawPredictionCard(img, cimX, cimY, cardWidth, cardHeight,
		"CIM (30 Levels)", "Hardware",
		result.CIMPrediction, result.CIMConfidence,
		color.RGBA{100, 255, 180, 255}, // Green
		result.CIMProbabilities)

	// === Match/Mismatch indicator (center, between cards) ===
	centerX := w / 2
	indicatorY := cardHeight/2 + padding

	if result.Match {
		// Green checkmark circle
		cc.drawCircle(img, centerX, indicatorY, 15, color.RGBA{50, 180, 100, 255})
		drawSimpleText(img, "=", centerX-3, indicatorY-4, color.RGBA{255, 255, 255, 255})
	} else {
		// Red X circle
		cc.drawCircle(img, centerX, indicatorY, 15, color.RGBA{200, 80, 80, 255})
		drawSimpleText(img, "x", centerX-3, indicatorY-4, color.RGBA{255, 255, 255, 255})
	}

	// === Bottom info section ===
	infoY := cardHeight + padding + 10

	// Energy comparison
	energyText := fmt.Sprintf("Energy: FP=%.0f nJ | CIM=%.1f nJ | %.0fx savings",
		result.EnergyGPU, result.EnergyFeCIM, result.EnergyRatio)
	drawSimpleText(img, energyText, padding, infoY, color.RGBA{180, 180, 200, 255})

	// Second-best predictions
	infoY += 15
	fpSecond, fpSecondConf := cc.getSecondBest(result.FPProbabilities)
	cimSecond, cimSecondConf := cc.getSecondBest(result.CIMProbabilities)

	secondText := fmt.Sprintf("2nd best: FP=%d (%.1f%%) | CIM=%d (%.1f%%)",
		fpSecond, fpSecondConf*100, cimSecond, cimSecondConf*100)
	drawSimpleText(img, secondText, padding, infoY, color.RGBA{140, 140, 160, 255})

	// Confidence delta indicator
	if result.ConfidenceDelta > 0.05 {
		infoY += 15
		deltaText := fmt.Sprintf("Confidence gap: %.1f%% - quantization impact visible",
			result.ConfidenceDelta*100)
		drawSimpleText(img, deltaText, padding, infoY, color.RGBA{255, 200, 100, 255})
	}

	return img
}

// drawPredictionCard draws a single prediction card.
func (cc *ComparisonCard) drawPredictionCard(img *image.RGBA, x, y, w, h int,
	title, subtitle string, prediction int, confidence float64, accentColor color.RGBA, probs []float64) {

	// Card background
	cardBg := color.RGBA{35, 40, 55, 255}
	for cx := x; cx < x+w; cx++ {
		for cy := y; cy < y+h; cy++ {
			img.Set(cx, cy, cardBg)
		}
	}

	// Border (accent color)
	for cx := x; cx < x+w; cx++ {
		img.Set(cx, y, accentColor)
		img.Set(cx, y+h-1, accentColor)
	}
	for cy := y; cy < y+h; cy++ {
		img.Set(x, cy, accentColor)
		img.Set(x+w-1, cy, accentColor)
	}

	// Title
	titleY := y + 8
	drawSimpleText(img, title, x+10, titleY, accentColor)

	// Subtitle
	subtitleY := titleY + 12
	drawSimpleText(img, subtitle, x+10, subtitleY, color.RGBA{120, 120, 140, 255})

	// Large prediction digit
	digitY := subtitleY + 20
	digitText := fmt.Sprintf("%d", prediction)
	if prediction < 0 {
		digitText = "?"
	}
	// Draw large digit (scaled up)
	cc.drawLargeDigit(img, x+w/2-20, digitY, digitText, accentColor)

	// Confidence bar
	barY := digitY + 55
	barX := x + 10
	barWidth := w - 20
	barHeight := 12

	// Background
	for bx := barX; bx < barX+barWidth; bx++ {
		for by := barY; by < barY+barHeight; by++ {
			img.Set(bx, by, color.RGBA{50, 50, 70, 255})
		}
	}

	// Fill
	fillWidth := int(float64(barWidth) * confidence)
	for bx := barX; bx < barX+fillWidth; bx++ {
		for by := barY; by < barY+barHeight; by++ {
			img.Set(bx, by, accentColor)
		}
	}

	// Confidence text
	confY := barY + barHeight + 5
	confText := fmt.Sprintf("%.1f%%", confidence*100)
	drawSimpleText(img, confText, x+w/2-20, confY, color.RGBA{200, 200, 220, 255})

	// Mini probability distribution (bottom of card)
	if len(probs) == 10 {
		probY := confY + 20
		probBarWidth := (w - 30) / 10
		probBarMaxH := h - (probY - y) - 15

		for i, p := range probs {
			probBarX := x + 15 + i*probBarWidth
			probBarH := int(float64(probBarMaxH) * p)
			if probBarH < 1 {
				probBarH = 1
			}

			probBarY := y + h - 10 - probBarH

			barColor := color.RGBA{80, 80, 100, 255}
			if i == prediction {
				barColor = accentColor
			}

			for bx := probBarX; bx < probBarX+probBarWidth-2; bx++ {
				for by := probBarY; by < y+h-10; by++ {
					img.Set(bx, by, barColor)
				}
			}
		}
	}
}

// drawLargeDigit draws a scaled-up digit.
func (cc *ComparisonCard) drawLargeDigit(img *image.RGBA, x, y int, digit string, c color.RGBA) {
	// 3x scale for the digit
	scale := 3

	patterns := map[rune][]string{
		'0': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
		'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
		'2': {"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
		'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
		'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
		'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
		'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
		'?': {"01110", "10001", "00001", "00110", "00100", "00000", "00100"},
	}

	for _, ch := range digit {
		pattern, ok := patterns[ch]
		if !ok {
			continue
		}

		for dy, row := range pattern {
			for dx, pixel := range row {
				if pixel == '1' {
					// Draw scaled pixel
					for sy := 0; sy < scale; sy++ {
						for sx := 0; sx < scale; sx++ {
							px := x + dx*scale + sx
							py := y + dy*scale + sy
							if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
								img.Set(px, py, c)
							}
						}
					}
				}
			}
		}
	}
}

// drawCircle draws a filled circle.
func (cc *ComparisonCard) drawCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for dy := -r; dy <= r; dy++ {
		for dx := -r; dx <= r; dx++ {
			if dx*dx+dy*dy <= r*r {
				px := cx + dx
				py := cy + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}

// getSecondBest returns the second-highest prediction and its confidence.
func (cc *ComparisonCard) getSecondBest(probs []float64) (int, float64) {
	if len(probs) < 2 {
		return -1, 0
	}

	bestIdx, secondIdx := 0, 1
	bestVal, secondVal := probs[0], probs[1]

	if secondVal > bestVal {
		bestIdx, secondIdx = secondIdx, bestIdx
		bestVal, secondVal = secondVal, bestVal
	}

	for i := 2; i < len(probs); i++ {
		if probs[i] > bestVal {
			secondIdx, secondVal = bestIdx, bestVal
			bestIdx, bestVal = i, probs[i]
		} else if probs[i] > secondVal {
			secondIdx, secondVal = i, probs[i]
		}
	}

	return secondIdx, secondVal
}

// DualProbabilityChart shows FP vs CIM probability comparison with divergence highlighting.
type DualProbabilityChart struct {
	widget.BaseWidget

	mu          sync.RWMutex
	fpProbs     []float64
	cimProbs    []float64
	divergences []float64
	fpPred      int
	cimPred     int

	raster *canvas.Raster
}

// NewDualProbabilityChart creates a new dual probability chart.
func NewDualProbabilityChart() *DualProbabilityChart {
	dpc := &DualProbabilityChart{
		fpProbs:     make([]float64, 10),
		cimProbs:    make([]float64, 10),
		divergences: make([]float64, 10),
		fpPred:      -1,
		cimPred:     -1,
	}
	dpc.ExtendBaseWidget(dpc)
	return dpc
}

// SetProbabilities updates both FP and CIM probabilities.
func (dpc *DualProbabilityChart) SetProbabilities(fpProbs, cimProbs []float64, fpPred, cimPred int) {
	dpc.mu.Lock()
	defer dpc.mu.Unlock()

	dpc.fpProbs = fpProbs
	dpc.cimProbs = cimProbs
	dpc.fpPred = fpPred
	dpc.cimPred = cimPred

	// Calculate divergences
	dpc.divergences = make([]float64, len(fpProbs))
	for i := range fpProbs {
		if i < len(cimProbs) {
			dpc.divergences[i] = math.Abs(fpProbs[i] - cimProbs[i])
		}
	}

	fyne.Do(func() {
		dpc.Refresh()
	})
}

// Clear resets the chart.
func (dpc *DualProbabilityChart) Clear() {
	dpc.mu.Lock()
	dpc.fpProbs = make([]float64, 10)
	dpc.cimProbs = make([]float64, 10)
	dpc.divergences = make([]float64, 10)
	dpc.fpPred = -1
	dpc.cimPred = -1
	dpc.mu.Unlock()
	fyne.Do(func() {
		dpc.Refresh()
	})
}

// MinSize returns the minimum size.
func (dpc *DualProbabilityChart) MinSize() fyne.Size {
	return fyne.NewSize(400, 150)
}

// CreateRenderer implements fyne.Widget.
func (dpc *DualProbabilityChart) CreateRenderer() fyne.WidgetRenderer {
	dpc.raster = canvas.NewRaster(dpc.generateImage)

	title := widget.NewLabelWithStyle("Probability Distribution (FP vs CIM)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewBorder(
		title,
		nil, nil, nil,
		container.NewMax(dpc.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the dual probability bar chart.
func (dpc *DualProbabilityChart) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 400
	}
	if h < 10 {
		h = 130
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 30, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	dpc.mu.RLock()
	fpProbs := dpc.fpProbs
	cimProbs := dpc.cimProbs
	divergences := dpc.divergences
	fpPred := dpc.fpPred
	cimPred := dpc.cimPred
	dpc.mu.RUnlock()

	if len(fpProbs) < 10 {
		return img
	}

	padding := 20
	labelHeight := 15
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding - labelHeight

	groupWidth := chartWidth / 10
	barWidth := (groupWidth - 4) / 2

	fpColor := color.RGBA{100, 150, 255, 255}   // Blue
	cimColor := color.RGBA{100, 255, 180, 255}  // Green
	warnColor := color.RGBA{255, 200, 100, 255} // Yellow for divergence

	for i := 0; i < 10; i++ {
		groupX := padding + i*groupWidth

		// FP bar
		fpHeight := int(float64(chartHeight) * fpProbs[i])
		if fpHeight < 1 && fpProbs[i] > 0 {
			fpHeight = 1
		}
		fpBarX := groupX + 1
		fpBarY := padding + chartHeight - fpHeight

		barCol := fpColor
		if i == fpPred {
			barCol = color.RGBA{150, 200, 255, 255} // Brighter for prediction
		}

		for bx := fpBarX; bx < fpBarX+barWidth; bx++ {
			for by := fpBarY; by < padding+chartHeight; by++ {
				img.Set(bx, by, barCol)
			}
		}

		// CIM bar
		cimHeight := int(float64(chartHeight) * cimProbs[i])
		if cimHeight < 1 && cimProbs[i] > 0 {
			cimHeight = 1
		}
		cimBarX := groupX + barWidth + 2
		cimBarY := padding + chartHeight - cimHeight

		barCol = cimColor
		if i == cimPred {
			barCol = color.RGBA{150, 255, 200, 255} // Brighter for prediction
		}

		for bx := cimBarX; bx < cimBarX+barWidth; bx++ {
			for by := cimBarY; by < padding+chartHeight; by++ {
				img.Set(bx, by, barCol)
			}
		}

		// Divergence warning marker (if > 2%)
		if divergences[i] > 0.02 {
			warnY := padding + chartHeight + 2
			warnX := groupX + groupWidth/2 - 2
			for wx := warnX; wx < warnX+4; wx++ {
				for wy := warnY; wy < warnY+4; wy++ {
					img.Set(wx, wy, warnColor)
				}
			}
		}

		// Digit label
		labelX := groupX + groupWidth/2 - 6 // Adjusted for scale 2
		labelY := h - 15
		drawScaledChar(img, rune('0'+i), labelX, labelY, 2, color.RGBA{150, 150, 170, 255})
	}

	// Legend
	legendY := 5
	drawScaledText(img, "FP", padding, legendY, 2, fpColor)
	drawScaledText(img, "CIM", padding+60, legendY, 2, cimColor)

	return img
}
