// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// LayerActivationView displays neural network layer activations.
type LayerActivationView struct {
	widget.BaseWidget

	// Layer data
	inputLayer  []float64 // 784 values (28x28)
	hiddenLayer []float64 // variable size
	outputLayer []float64 // 10 values

	// Rasters for each layer
	inputRaster  *canvas.Raster
	hiddenRaster *canvas.Raster
	outputRaster *canvas.Raster

	// Labels
	predictionLabel *widget.Label
	confidenceLabel *widget.Label
}

// NewLayerActivationView creates a new layer activation visualization.
func NewLayerActivationView() *LayerActivationView {
	lav := &LayerActivationView{
		inputLayer:  make([]float64, 784),
		hiddenLayer: make([]float64, 128),
		outputLayer: make([]float64, 10),
	}
	lav.ExtendBaseWidget(lav)
	return lav
}

// SetInput sets the input layer (28x28 = 784 values).
func (lav *LayerActivationView) SetInput(input []float64) {
	lav.inputLayer = input
	lav.Refresh()
}

// SetHidden sets the hidden layer activations.
func (lav *LayerActivationView) SetHidden(hidden []float64) {
	lav.hiddenLayer = hidden
	lav.Refresh()
}

// SetOutput sets the output layer (10 class probabilities).
func (lav *LayerActivationView) SetOutput(output []float64) {
	lav.outputLayer = output
	lav.Refresh()
}

// SetActivations sets all layer activations at once.
func (lav *LayerActivationView) SetActivations(input, hidden, output []float64) {
	lav.inputLayer = input
	lav.hiddenLayer = hidden
	lav.outputLayer = output
	lav.Refresh()
}

// GetPrediction returns the predicted class and confidence.
func (lav *LayerActivationView) GetPrediction() (int, float64) {
	if len(lav.outputLayer) == 0 {
		return -1, 0
	}

	maxIdx := 0
	maxVal := lav.outputLayer[0]
	for i, v := range lav.outputLayer {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx, maxVal
}

// CreateRenderer implements fyne.Widget.
func (lav *LayerActivationView) CreateRenderer() fyne.WidgetRenderer {
	// Create rasters for each layer
	lav.inputRaster = canvas.NewRaster(lav.generateInputImage)
	lav.hiddenRaster = canvas.NewRaster(lav.generateHiddenImage)
	lav.outputRaster = canvas.NewRaster(lav.generateOutputImage)

	// Labels
	inputLabel := widget.NewLabel("Input (28×28)")
	inputLabel.TextStyle = fyne.TextStyle{Bold: true}
	inputLabel.Alignment = fyne.TextAlignCenter

	hiddenLabel := widget.NewLabel("Hidden Layer")
	hiddenLabel.TextStyle = fyne.TextStyle{Bold: true}
	hiddenLabel.Alignment = fyne.TextAlignCenter

	outputLabel := widget.NewLabel("Output (10 Classes)")
	outputLabel.TextStyle = fyne.TextStyle{Bold: true}
	outputLabel.Alignment = fyne.TextAlignCenter

	// Prediction labels
	lav.predictionLabel = widget.NewLabel("Prediction: -")
	lav.predictionLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	lav.predictionLabel.Alignment = fyne.TextAlignCenter

	lav.confidenceLabel = widget.NewLabel("Confidence: -")
	lav.confidenceLabel.Alignment = fyne.TextAlignCenter

	// Layout each layer vertically with its label
	inputBox := container.NewVBox(
		inputLabel,
		container.NewCenter(lav.inputRaster),
	)

	hiddenBox := container.NewVBox(
		hiddenLabel,
		container.NewCenter(lav.hiddenRaster),
	)

	outputBox := container.NewVBox(
		outputLabel,
		container.NewCenter(lav.outputRaster),
		widget.NewSeparator(),
		lav.predictionLabel,
		lav.confidenceLabel,
	)

	// Horizontal layout: input -> hidden -> output
	content := container.NewHBox(
		inputBox,
		widget.NewSeparator(),
		hiddenBox,
		widget.NewSeparator(),
		outputBox,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (lav *LayerActivationView) MinSize() fyne.Size {
	return fyne.NewSize(600, 200)
}

// generateInputImage creates the 28x28 input visualization.
func (lav *LayerActivationView) generateInputImage(w, h int) image.Image {
	// Fixed size for input
	size := 140
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(lav.inputLayer) < 784 {
		return img
	}

	// Draw 28x28 pixels
	cellSize := size / 28
	for py := 0; py < 28; py++ {
		for px := 0; px < 28; px++ {
			value := lav.inputLayer[py*28+px]
			if value > 0 {
				intensity := uint8(clamp(value, 0, 1) * 255)
				c := color.RGBA{
					R: uint8(float64(intensity) * 0.7),
					G: intensity,
					B: intensity,
					A: 255,
				}

				x0 := px * cellSize
				y0 := py * cellSize
				for y := y0; y < y0+cellSize; y++ {
					for x := x0; x < x0+cellSize; x++ {
						if x < size && y < size {
							img.Set(x, y, c)
						}
					}
				}
			}
		}
	}

	return img
}

// generateHiddenImage creates the hidden layer activation visualization.
func (lav *LayerActivationView) generateHiddenImage(w, h int) image.Image {
	width := 160
	height := 140
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(lav.hiddenLayer) == 0 {
		return img
	}

	// Arrange neurons in a grid
	neurons := len(lav.hiddenLayer)
	cols := int(math.Ceil(math.Sqrt(float64(neurons))))
	rows := (neurons + cols - 1) / cols

	cellW := (width - 10) / cols
	cellH := (height - 10) / rows
	if cellW < 2 {
		cellW = 2
	}
	if cellH < 2 {
		cellH = 2
	}

	// Find max for normalization
	maxVal := 0.0
	for _, v := range lav.hiddenLayer {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal <= 0 {
		maxVal = 1
	}

	// Draw neurons
	for i, value := range lav.hiddenLayer {
		col := i % cols
		row := i / cols

		normVal := value / maxVal
		intensity := uint8(clamp(normVal, 0, 1) * 255)

		// Use orange-yellow gradient for hidden layer
		c := color.RGBA{
			R: intensity,
			G: uint8(float64(intensity) * 0.7),
			B: uint8(float64(intensity) * 0.2),
			A: 255,
		}

		x0 := 5 + col*cellW
		y0 := 5 + row*cellH
		for y := y0; y < y0+cellH-1 && y < height; y++ {
			for x := x0; x < x0+cellW-1 && x < width; x++ {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

// generateOutputImage creates the output layer bar chart.
func (lav *LayerActivationView) generateOutputImage(w, h int) image.Image {
	width := 200
	height := 140
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Background
	bgColor := color.RGBA{25, 25, 35, 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(lav.outputLayer) != 10 {
		return img
	}

	// Find prediction
	maxIdx := 0
	maxVal := lav.outputLayer[0]
	for i, v := range lav.outputLayer {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	// Update prediction labels
	if lav.predictionLabel != nil {
		lav.predictionLabel.SetText(fmt.Sprintf("Prediction: %d", maxIdx))
	}
	if lav.confidenceLabel != nil {
		lav.confidenceLabel.SetText(fmt.Sprintf("Confidence: %.1f%%", maxVal*100))
	}

	// Draw bar chart
	padding := 15
	barWidth := (width - 2*padding) / 10
	chartHeight := height - 2*padding

	// Draw bars for each class
	for i := 0; i < 10; i++ {
		value := lav.outputLayer[i]
		barHeight := int(clamp(value, 0, 1) * float64(chartHeight))

		x0 := padding + i*barWidth + 1
		x1 := padding + (i+1)*barWidth - 1
		y0 := height - padding - barHeight
		y1 := height - padding

		// Color: green for prediction, coral for others
		var c color.RGBA
		if i == maxIdx {
			c = color.RGBA{100, 255, 150, 255} // Green for predicted
		} else {
			c = color.RGBA{255, 127, 80, 200} // Coral for others
		}

		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if x >= 0 && x < width && y >= 0 && y < height {
					img.Set(x, y, c)
				}
			}
		}

		// Draw class number below (as simple pixels)
		// Skip for simplicity - label is in predictionLabel
	}

	// Draw axis
	axisColor := color.RGBA{80, 80, 90, 255}
	for x := padding; x < width-padding; x++ {
		img.Set(x, height-padding, axisColor)
	}
	for y := padding; y < height-padding; y++ {
		img.Set(padding, y, axisColor)
	}

	return img
}

// OutputBarChart provides a standalone output visualization.
type OutputBarChart struct {
	widget.BaseWidget

	values    []float64
	labels    []string
	predicted int
	raster    *canvas.Raster
}

// NewOutputBarChart creates a new output bar chart.
func NewOutputBarChart() *OutputBarChart {
	obc := &OutputBarChart{
		values:    make([]float64, 10),
		labels:    []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
		predicted: -1,
	}
	obc.ExtendBaseWidget(obc)
	return obc
}

// SetValues updates the chart with new probabilities.
func (obc *OutputBarChart) SetValues(values []float64) {
	obc.values = values

	// Find prediction
	obc.predicted = 0
	maxVal := 0.0
	for i, v := range values {
		if v > maxVal {
			maxVal = v
			obc.predicted = i
		}
	}

	obc.Refresh()
}

// GetPrediction returns the predicted class.
func (obc *OutputBarChart) GetPrediction() int {
	return obc.predicted
}

// CreateRenderer implements fyne.Widget.
func (obc *OutputBarChart) CreateRenderer() fyne.WidgetRenderer {
	obc.raster = canvas.NewRaster(obc.generateImage)

	titleLabel := widget.NewLabel("Class Probabilities")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(
		titleLabel,
		nil,
		nil,
		nil,
		obc.raster,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (obc *OutputBarChart) MinSize() fyne.Size {
	return fyne.NewSize(300, 150)
}

// generateImage creates the bar chart image.
func (obc *OutputBarChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{30, 30, 40, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if len(obc.values) == 0 {
		return img
	}

	// Calculate dimensions
	padding := 30
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding
	barWidth := chartWidth / len(obc.values)

	// Draw bars
	for i, val := range obc.values {
		barHeight := int(clamp(val, 0, 1) * float64(chartHeight))

		x0 := padding + i*barWidth + 2
		x1 := padding + (i+1)*barWidth - 2
		y0 := h - padding - barHeight
		y1 := h - padding

		// Color based on prediction
		var c color.RGBA
		if i == obc.predicted {
			c = color.RGBA{0, 230, 180, 255} // Cyan for predicted
		} else {
			c = color.RGBA{100, 100, 120, 255} // Gray for others
		}

		for y := y0; y < y1; y++ {
			for x := x0; x < x1; x++ {
				if x >= 0 && x < w && y >= 0 && y < h {
					img.Set(x, y, c)
				}
			}
		}
	}

	// Draw axis
	axisColor := color.RGBA{80, 80, 90, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}

	return img
}
