// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"ironlattice-vis/demo2-crossbar/pkg/crossbar"
	"ironlattice-vis/demo3-mnist/pkg/training"
)

// MNISTApp is the main application for the MNIST demo.
type MNISTApp struct {
	fyneApp fyne.App
	window  fyne.Window

	// Neural network
	network *training.MNISTNetwork

	// GUI components
	digitCanvas        *DigitCanvas
	layerView          *LayerActivationView
	confusionMatrix    *ConfusionMatrix
	metricsPanel       *MetricsPanel
	classStatsPanel    *ClassStatsPanel
	outputChart        *OutputBarChart

	// Labels
	statusLabel     *widget.Label
	predictionLabel *widget.Label
	confidenceLabel *widget.Label

	// Test data for confusion matrix
	testImages [][]float64
	testLabels []int

	// Data directory
	dataDir string
}

// NewMNISTApp creates and initializes the MNIST demo application.
func NewMNISTApp() *MNISTApp {
	ma := &MNISTApp{}

	// Create Fyne app
	ma.fyneApp = app.NewWithID("com.ironlattice.mnist-demo")
	ma.fyneApp.Settings().SetTheme(theme.DarkTheme())

	// Find data directory
	ma.dataDir = findDataDir()

	// Create crossbar arrays for layers
	// Layer 1: hidden x 784 (transposed for MVM)
	layer1Config := &crossbar.Config{
		Rows:       128, // hidden size
		Cols:       784, // input size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer1, _ := crossbar.NewArray(layer1Config)

	// Layer 2: 10 x hidden
	layer2Config := &crossbar.Config{
		Rows:       10,  // output size
		Cols:       128, // hidden size
		NoiseLevel: 0.01,
		ADCBits:    6,
		DACBits:    8,
	}
	layer2, _ := crossbar.NewArray(layer2Config)

	// Create network
	ma.network = training.NewMNISTNetwork(layer1, layer2)

	// Try to load pretrained weights
	weightsPath := filepath.Join(ma.dataDir, "pretrained_weights.json")
	if _, err := os.Stat(weightsPath); err == nil {
		if err := ma.network.LoadWeights(weightsPath); err == nil {
			fmt.Println("Loaded pretrained weights from", weightsPath)
		}
	}

	return ma
}

// findDataDir locates the demo3-mnist/data directory.
func findDataDir() string {
	// Try common locations
	paths := []string{
		"data",
		"demo3-mnist/data",
		"../data",
		"../../demo3-mnist/data",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "data" // Default
}

// Run starts the GUI application.
func (ma *MNISTApp) Run() {
	ma.window = ma.fyneApp.NewWindow("IronLattice Demo 3: MNIST Neural Network")
	ma.window.Resize(fyne.NewSize(1400, 900))

	// Create main layout
	content := ma.createMainLayout()
	ma.window.SetContent(content)

	// Initialize
	ma.updateStatus("Ready. Draw a digit or load test data.")

	ma.window.ShowAndRun()
}

// createMainLayout builds the main application layout.
func (ma *MNISTApp) createMainLayout() fyne.CanvasObject {
	// Create components
	ma.digitCanvas = NewDigitCanvas()
	ma.digitCanvas.OnDigitChanged = ma.onDigitChanged

	ma.layerView = NewLayerActivationView()
	ma.outputChart = NewOutputBarChart()

	ma.confusionMatrix = NewConfusionMatrix()
	ma.confusionMatrix.OnCellTapped = ma.onConfusionCellTapped

	ma.metricsPanel = NewMetricsPanel()
	ma.classStatsPanel = NewClassStatsPanel()

	// Status labels
	ma.statusLabel = widget.NewLabel("Status: Ready")
	ma.statusLabel.TextStyle = fyne.TextStyle{Bold: true}

	ma.predictionLabel = widget.NewLabel("Prediction: -")
	ma.predictionLabel.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}

	ma.confidenceLabel = widget.NewLabel("Confidence: -")

	// Control buttons
	clearBtn := widget.NewButton("Clear Canvas", func() {
		ma.digitCanvas.Clear()
		ma.updateStatus("Canvas cleared")
	})

	randomBtn := widget.NewButton("Random Test", func() {
		ma.loadRandomTestDigit()
	})

	evalBtn := widget.NewButton("Evaluate All", func() {
		ma.evaluateNetwork()
	})
	evalBtn.Importance = widget.HighImportance

	loadTestBtn := widget.NewButton("Load Test Data", func() {
		ma.loadTestData()
	})

	// Buttons container
	buttonBox := container.NewHBox(
		clearBtn,
		randomBtn,
		loadTestBtn,
		evalBtn,
	)

	// Title and header
	titleLabel := widget.NewLabel("IronLattice MNIST Neural Network")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	subtitleLabel := widget.NewLabel("Ferroelectric Compute-in-Memory with 30 Discrete Analog States")
	subtitleLabel.Alignment = fyne.TextAlignCenter

	specsLabel := widget.NewLabel("Architecture: 784 -> 128 -> 10 | Target: 87% | 30 Levels")
	specsLabel.Alignment = fyne.TextAlignCenter

	header := container.NewVBox(
		titleLabel,
		subtitleLabel,
		specsLabel,
		widget.NewSeparator(),
	)

	// Left panel: Drawing canvas + prediction
	canvasLabel := widget.NewLabel("Draw a Digit (0-9)")
	canvasLabel.TextStyle = fyne.TextStyle{Bold: true}
	canvasLabel.Alignment = fyne.TextAlignCenter

	predictionBox := container.NewVBox(
		widget.NewSeparator(),
		ma.predictionLabel,
		ma.confidenceLabel,
	)

	leftPanel := container.NewVBox(
		canvasLabel,
		container.NewCenter(ma.digitCanvas),
		predictionBox,
		widget.NewSeparator(),
		buttonBox,
	)

	// Center panel: Layer activations
	activationLabel := widget.NewLabel("Network Activations")
	activationLabel.TextStyle = fyne.TextStyle{Bold: true}
	activationLabel.Alignment = fyne.TextAlignCenter

	centerPanel := container.NewVBox(
		activationLabel,
		widget.NewSeparator(),
		ma.layerView,
		widget.NewSeparator(),
		ma.outputChart,
	)

	// Right panel: Confusion matrix + metrics
	rightPanel := container.NewVBox(
		ma.confusionMatrix,
		widget.NewSeparator(),
		ma.metricsPanel,
		widget.NewSeparator(),
		ma.classStatsPanel,
	)

	// Tabs for different views
	drawTab := container.NewTabItem("Draw & Predict",
		container.NewHSplit(
			container.NewPadded(leftPanel),
			container.NewPadded(centerPanel),
		),
	)

	metricsTab := container.NewTabItem("Evaluation Metrics",
		container.NewPadded(rightPanel),
	)

	tabs := container.NewAppTabs(drawTab, metricsTab)

	// Footer
	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			ma.statusLabel,
			layout.NewSpacer(),
			widget.NewLabel("IronLattice Ferroelectric CIM | 30 Discrete Levels"),
		),
	)

	// Main content
	mainContent := container.NewBorder(
		header,  // top
		footer,  // bottom
		nil,     // left
		nil,     // right
		tabs,    // center
	)

	return mainContent
}

// onDigitChanged handles canvas drawing updates.
func (ma *MNISTApp) onDigitChanged(pixels []float64) {
	// Run inference
	input, hidden, probs := ma.network.GetLayerActivations(pixels)

	// Update visualization
	ma.layerView.SetActivations(input, hidden, probs)
	ma.outputChart.SetValues(probs)

	// Update prediction
	pred, conf := ma.network.Predict(pixels)
	ma.predictionLabel.SetText(fmt.Sprintf("Prediction: %d", pred))
	ma.confidenceLabel.SetText(fmt.Sprintf("Confidence: %.1f%%", conf*100))
}

// loadRandomTestDigit loads a random digit from test data.
func (ma *MNISTApp) loadRandomTestDigit() {
	if len(ma.testImages) == 0 {
		ma.loadTestData()
		if len(ma.testImages) == 0 {
			ma.updateStatus("No test data available")
			return
		}
	}

	idx := rand.Intn(len(ma.testImages))
	pixels := ma.testImages[idx]
	label := ma.testLabels[idx]

	ma.digitCanvas.SetPixels(pixels)
	ma.onDigitChanged(pixels)

	ma.updateStatus(fmt.Sprintf("Loaded test digit (label: %d)", label))
}

// loadTestData loads MNIST test data.
func (ma *MNISTApp) loadTestData() {
	ma.updateStatus("Loading test data...")

	// Try to load from IDX files
	testImagesPath := filepath.Join(ma.dataDir, "t10k-images-idx3-ubyte")
	testLabelsPath := filepath.Join(ma.dataDir, "t10k-labels-idx1-ubyte")

	images, labels, err := loadMNISTData(testImagesPath, testLabelsPath)
	if err != nil {
		// Fall back to generating synthetic data for demo
		ma.updateStatus("Using synthetic test data")
		ma.testImages, ma.testLabels = generateSyntheticData(100)
		return
	}

	// Limit to 1000 samples for speed
	if len(images) > 1000 {
		ma.testImages = images[:1000]
		ma.testLabels = labels[:1000]
	} else {
		ma.testImages = images
		ma.testLabels = labels
	}

	ma.updateStatus(fmt.Sprintf("Loaded %d test samples", len(ma.testImages)))
}

// evaluateNetwork runs evaluation on test data.
func (ma *MNISTApp) evaluateNetwork() {
	if len(ma.testImages) == 0 {
		ma.loadTestData()
		if len(ma.testImages) == 0 {
			ma.updateStatus("No test data for evaluation")
			return
		}
	}

	ma.updateStatus("Evaluating network...")

	// Compute confusion matrix
	confMatrix := ma.network.ComputeConfusionMatrix(ma.testImages, ma.testLabels)

	// Convert to [10][10]int
	var matrix [10][10]int
	for i := 0; i < 10; i++ {
		for j := 0; j < 10; j++ {
			matrix[i][j] = confMatrix[i][j]
		}
	}
	ma.confusionMatrix.SetMatrix(matrix)

	// Compute metrics
	precision, recall, f1 := ma.network.GetPerClassMetrics(confMatrix)

	var precArr, recArr, f1Arr [10]float64
	for i := 0; i < 10; i++ {
		precArr[i] = precision[i]
		recArr[i] = recall[i]
		f1Arr[i] = f1[i]
	}

	accuracy := ma.confusionMatrix.GetAccuracy()
	ma.metricsPanel.SetMetrics(precArr, recArr, f1Arr, accuracy)

	ma.updateStatus(fmt.Sprintf("Evaluation complete. Accuracy: %.1f%%", accuracy*100))
}

// onConfusionCellTapped handles clicks on the confusion matrix.
func (ma *MNISTApp) onConfusionCellTapped(actual, predicted, count int) {
	precision, recall, f1 := ma.confusionMatrix.GetClassMetrics(actual)

	// Estimate TP based on click location
	tp := 0
	fp := 0
	fn := 0
	if actual == predicted {
		tp = count
	} else {
		fp = count // Misclassification
	}

	ma.classStatsPanel.SetClass(actual, precision, recall, f1, tp, fp, fn, count)

	ma.updateStatus(fmt.Sprintf("Cell [%d,%d]: %d samples (actual=%d, predicted=%d)",
		actual, predicted, count, actual, predicted))
}

// updateStatus updates the status label.
func (ma *MNISTApp) updateStatus(status string) {
	ma.statusLabel.SetText("Status: " + status)
}

// loadMNISTData loads MNIST data from IDX files.
func loadMNISTData(imagesPath, labelsPath string) ([][]float64, []int, error) {
	// Read images
	imagesData, err := os.ReadFile(imagesPath)
	if err != nil {
		return nil, nil, err
	}

	// Read labels
	labelsData, err := os.ReadFile(labelsPath)
	if err != nil {
		return nil, nil, err
	}

	// Parse images (IDX3 format)
	if len(imagesData) < 16 {
		return nil, nil, fmt.Errorf("images file too small")
	}

	// Skip 16-byte header
	numImages := int(imagesData[4])<<24 | int(imagesData[5])<<16 | int(imagesData[6])<<8 | int(imagesData[7])
	numRows := int(imagesData[8])<<24 | int(imagesData[9])<<16 | int(imagesData[10])<<8 | int(imagesData[11])
	numCols := int(imagesData[12])<<24 | int(imagesData[13])<<16 | int(imagesData[14])<<8 | int(imagesData[15])

	if numRows != 28 || numCols != 28 {
		return nil, nil, fmt.Errorf("unexpected image dimensions: %dx%d", numRows, numCols)
	}

	imageSize := 28 * 28
	images := make([][]float64, numImages)
	for i := 0; i < numImages; i++ {
		images[i] = make([]float64, imageSize)
		offset := 16 + i*imageSize
		for j := 0; j < imageSize && offset+j < len(imagesData); j++ {
			images[i][j] = float64(imagesData[offset+j]) / 255.0
		}
	}

	// Parse labels (IDX1 format)
	if len(labelsData) < 8 {
		return nil, nil, fmt.Errorf("labels file too small")
	}

	numLabels := int(labelsData[4])<<24 | int(labelsData[5])<<16 | int(labelsData[6])<<8 | int(labelsData[7])
	labels := make([]int, numLabels)
	for i := 0; i < numLabels && 8+i < len(labelsData); i++ {
		labels[i] = int(labelsData[8+i])
	}

	return images, labels, nil
}

// generateSyntheticData creates simple synthetic digit patterns for demo.
func generateSyntheticData(count int) ([][]float64, []int) {
	images := make([][]float64, count)
	labels := make([]int, count)

	for i := 0; i < count; i++ {
		digit := rand.Intn(10)
		labels[i] = digit
		images[i] = make([]float64, 784)

		// Generate simple digit-like pattern
		switch digit {
		case 0: // Circle
			for y := 8; y < 20; y++ {
				for x := 8; x < 20; x++ {
					dx := float64(x) - 14
					dy := float64(y) - 14
					dist := dx*dx + dy*dy
					if dist > 25 && dist < 49 {
						images[i][y*28+x] = 0.8 + rand.Float64()*0.2
					}
				}
			}
		case 1: // Vertical line
			for y := 6; y < 22; y++ {
				images[i][y*28+14] = 0.8 + rand.Float64()*0.2
				images[i][y*28+13] = 0.5 + rand.Float64()*0.3
			}
		case 7: // Diagonal line from top-left
			for i := 0; i < 16; i++ {
				y := 6 + i
				x := 8 + i/2
				if y < 28 && x < 28 {
					images[labels[0]][y*28+x] = 0.8 + rand.Float64()*0.2
				}
			}
		default: // Random blob for other digits
			cx := 10 + rand.Intn(8)
			cy := 10 + rand.Intn(8)
			for y := 0; y < 28; y++ {
				for x := 0; x < 28; x++ {
					dx := float64(x - cx)
					dy := float64(y - cy)
					dist := dx*dx + dy*dy
					if dist < 36 {
						images[i][y*28+x] = 0.5 + rand.Float64()*0.5
					}
				}
			}
		}

		// Add noise
		for j := 0; j < 784; j++ {
			if rand.Float64() < 0.02 {
				images[i][j] = rand.Float64() * 0.3
			}
		}
	}

	return images, labels
}
