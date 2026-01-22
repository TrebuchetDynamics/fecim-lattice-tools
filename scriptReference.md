# FeCIM Visualizer Script Reference

Quick reference for file structure and key functions. Use this for fast lookups.

## Directory Structure

```
multilayer-ferroelectric-cim-visualizer/
├── cmd/
│   ├── fecim-visualizer/          # Unified GUI application entry point
│   │   ├── main.go                # Main entry, creates tabbed app with all 8 demos
│   │   └── launcher.go            # Home tab with demo cards
│   └── launcher/                  # Legacy launcher
│
├── shared/                        # Shared utilities across all demos
│   ├── theme/theme.go             # FeCIM color theme (ColorPrimary, ColorBackground, etc.)
│   └── logging/logging.go         # Shared logger with file + stdout output
│
├── module1-hysteresis/              # Demo 1: P-E Hysteresis Curve
│   ├── cmd/hysteresis/main.go     # Standalone entry point
│   ├── pkg/ferroelectric/
│   │   └── preisach.go            # Preisach hysteresis model
│   ├── pkg/gui/
│   │   └── overlay.go             # Lab bench overlay rendering
│   ├── pkg/render/
│   │   ├── plot.go                # P-E curve plotting
│   │   └── vulkan.go              # Vulkan renderer
│   ├── pkg/simulation/            # Simulation engine
│   └── pkg/tui/                   # Terminal UI mode
│   └── shaders/                   # SPIR-V compute/vertex/fragment shaders
│
├── module2-crossbar/                # Demo 2: Crossbar Array MVM
│   ├── cmd/crossbar-gui/main.go   # Standalone entry point
│   ├── pkg/crossbar/
│   │   ├── array.go               # Core crossbar array implementation
│   │   ├── nonidealities.go       # IR drop, sneak path analysis
│   │   └── reference.go           # Reference data
│   ├── pkg/gui/
│   │   ├── app.go                 # Main CrossbarApp
│   │   ├── heatmap.go             # Conductance heatmap widget
│   │   ├── controls.go            # Control panel
│   │   ├── vectors.go             # MVM visualization
│   │   ├── embedded.go            # EmbeddedCrossbarApp for unified GUI
│   │   └── liveslide.go           # Live slide components
│   └── shaders/                   # MVM compute shaders
│
├── module3-mnist/                   # Demo 3: MNIST Neural Network
│   ├── pkg/mnist/loader.go        # MNIST data loading
│   ├── pkg/training/network.go    # Neural network implementation
│   ├── pkg/gui/
│   │   ├── app.go                 # MNISTApp
│   │   ├── canvas.go              # Drawing canvas
│   │   ├── activations.go         # Layer activation visualization
│   │   ├── metrics.go             # Accuracy metrics display
│   │   └── embedded.go            # EmbeddedMNISTApp
│   └── data/                      # MNIST dataset (gzipped)
│
├── module4-circuits/                # Demo 4: Peripheral Circuits
│   ├── pkg/peripherals/
│   │   ├── adc.go                 # ADC circuit model
│   │   ├── dac.go                 # DAC circuit model
│   │   ├── tia.go                 # Transimpedance amplifier
│   │   ├── chargepump.go          # Charge pump for programming
│   │   └── analysis.go            # Circuit analysis utilities
│   ├── pkg/gui/
│   │   ├── app.go                 # CircuitsApp
│   │   ├── signalflow.go          # Signal flow visualization
│   │   └── embedded.go            # EmbeddedCircuitsApp
│
├── demo5-thermal/                 # Demo 5: Thermal Analysis
│   ├── pkg/thermal/
│   │   ├── simulation.go          # Thermal simulation engine
│   │   ├── heatmap.go             # Thermal heatmap generation
│   │   └── multilayer.go          # Multi-layer thermal model
│   ├── pkg/gui/
│   │   ├── app.go                 # ThermalApp
│   │   └── embedded.go            # EmbeddedThermalApp
│
├── demo6-multilayer/              # Demo 6: 3D Stack Visualization
│   ├── pkg/multilayer/
│   │   ├── stack.go               # Multi-layer stack model
│   │   ├── via.go                 # Via interconnect model
│   │   └── render.go              # 3D rendering utilities
│   ├── pkg/gui/
│   │   ├── app.go                 # MultilayerApp
│   │   └── embedded.go            # EmbeddedMultilayerApp
│
├── demo7-nonidealities/           # Demo 7: Non-Idealities Analysis
│   ├── pkg/nonidealities/
│   │   ├── drift.go               # Conductance drift model
│   │   ├── irdrop.go              # IR drop analysis
│   │   ├── sneakpath.go           # Sneak path analysis
│   │   └── render.go              # Visualization rendering
│   ├── pkg/gui/
│   │   ├── app.go                 # NonIdealitiesApp
│   │   └── embedded.go            # EmbeddedNonIdealitiesApp
│
├── module5-comparison/              # Demo 8: Technology Comparison
│   ├── pkg/comparison/
│   │   ├── architecture.go        # Memory architecture comparison
│   │   └── render.go              # Comparison charts
│   ├── pkg/gui/
│   │   ├── app.go                 # ComparisonApp
│   │   ├── widgets.go             # Custom comparison widgets
│   │   └── embedded.go            # EmbeddedComparisonApp
│
├── docs/papers/                   # Research papers (organized by topic)
└── logs/                          # Runtime logs (datetime-stamped)
```

## Key Constants

```go
// module2-crossbar/pkg/crossbar/array.go:11
const FeCIMLevels = 30  // "It's got 30 discrete states" - Dr. Tour
```

## Core Types & Functions

### Unified Application (cmd/fecim-visualizer/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| main.go:66 | `DemoApp` | Holds all 8 embedded demo instances |
| main.go:77 | `main()` | Creates tabbed Fyne app, manages demo start/stop |
| main.go:32 | `feCIMTheme` | Implements fyne.Theme for FeCIM branding |
| launcher.go:23 | `GetDemos()` | Returns DemoInfo slice for all 8 demos |
| launcher.go:275 | `CreateLauncherContent()` | Creates home tab with clickable demo cards |

### Demo 1: Hysteresis (module1-hysteresis/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/ferroelectric/preisach.go:11 | `PreisachModel` | Preisach hysteresis model with memory |
| pkg/ferroelectric/preisach.go:30 | `NewPreisachModel()` | Constructor with material parameters |
| pkg/ferroelectric/preisach.go:51 | `Update(E float64)` | Apply electric field, return polarization |
| pkg/ferroelectric/preisach.go:168 | `GetHysteresisLoop()` | Generate full P-E curve data |
| pkg/ferroelectric/preisach.go:203 | `DiscreteStates(N int)` | Get N discrete polarization states |
| pkg/gui/overlay.go:51 | `RenderText()` | Render lab bench status overlay |
| pkg/render/vulkan.go | `VulkanRenderer` | GPU-accelerated rendering |

### Demo 2: Crossbar (module2-crossbar/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/crossbar/array.go:14 | `Config` | Array configuration (rows, cols, noise, ADC/DAC bits) |
| pkg/crossbar/array.go:30 | `Array` | Crossbar array with cells matrix |
| pkg/crossbar/array.go:45 | `NewArray(cfg)` | Create new crossbar array |
| pkg/crossbar/array.go:73 | `ProgramWeight()` | Program weight to cell (quantizes to 30 levels) |
| pkg/crossbar/array.go:90 | `QuantizeTo30Levels()` | Quantize value to FeCIM 30 discrete levels |
| pkg/crossbar/array.go:123 | `MVM(input)` | Matrix-vector multiply: y = W * x |
| pkg/crossbar/array.go:155 | `VMM(input)` | Vector-matrix multiply: y = x * W |
| pkg/crossbar/nonidealities.go | `AnalyzeIRDrop()` | Compute IR drop across array |
| pkg/crossbar/nonidealities.go | `AnalyzeSneakPaths()` | Compute sneak path currents |
| pkg/gui/app.go:92 | `CrossbarApp` | Main application with heatmaps, controls |
| pkg/gui/app.go:159 | `NewCrossbarApp()` | Constructor |
| pkg/gui/app.go:190 | `Run()` | Start GUI |
| pkg/gui/app.go:663 | `runMVM()` | Execute animated MVM operation |
| pkg/gui/heatmap.go | `CrossbarHeatmap` | Interactive conductance heatmap widget |
| pkg/gui/embedded.go | `EmbeddedCrossbarApp` | Embeddable version for unified GUI |

### Demo 3: MNIST (module3-mnist/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/mnist/loader.go | `LoadMNIST()` | Load MNIST dataset from gzipped files |
| pkg/training/network.go | `Network` | Simple MLP neural network |
| pkg/training/network.go | `Forward()` | Forward pass through network |
| pkg/gui/canvas.go | `DrawingCanvas` | Custom drawing widget for digit input |
| pkg/gui/app.go | `MNISTApp` | Main MNIST demo application |

### Demo 4: Circuits (module4-circuits/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/peripherals/adc.go | `ADC` | Analog-to-digital converter model |
| pkg/peripherals/dac.go | `DAC` | Digital-to-analog converter model |
| pkg/peripherals/tia.go | `TIA` | Transimpedance amplifier model |
| pkg/peripherals/chargepump.go | `ChargePump` | Programming voltage generator |
| pkg/gui/signalflow.go | `SignalFlow` | Signal flow visualization |

### Demo 5: Thermal (demo5-thermal/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/thermal/simulation.go | `ThermalSimulation` | Heat dissipation simulation |
| pkg/thermal/heatmap.go | `GenerateHeatmap()` | Create thermal heatmap data |
| pkg/thermal/multilayer.go | `MultilayerThermal` | 3D thermal model |

### Demo 6: Multilayer (demo6-multilayer/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/multilayer/stack.go | `Stack` | Multi-layer crossbar stack |
| pkg/multilayer/via.go | `Via` | Inter-layer via connection |
| pkg/multilayer/render.go | `RenderStack()` | 3D visualization of stack |

### Demo 7: Non-Idealities (demo7-nonidealities/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/nonidealities/drift.go | `DriftModel` | Conductance drift over time |
| pkg/nonidealities/irdrop.go | `IRDropModel` | Wire resistance voltage drop |
| pkg/nonidealities/sneakpath.go | `SneakPathModel` | Parasitic current paths |

### Demo 8: Comparison (module5-comparison/)

| File | Type/Function | Purpose |
|------|---------------|---------|
| pkg/comparison/architecture.go | `Architecture` | Memory technology specs |
| pkg/comparison/architecture.go | `CompareEnergy()` | Energy efficiency comparison |
| pkg/comparison/render.go | `RenderComparison()` | Bar chart visualization |

## Shared Utilities

### Theme (shared/theme/theme.go)

```go
ColorPrimary    = color.RGBA{0, 212, 255, 255}   // Cyan
ColorSecondary  = color.RGBA{255, 107, 107, 255} // Coral red
ColorBackground = color.RGBA{0, 50, 100, 255}    // FeCIM blue #003264
```

### Logging (shared/logging/logging.go)

```go
logger := logging.NewLogger("demo-name")  // Creates timestamped log file
logger.Printf("message: %v", value)
defer logger.Close()
```

## Build & Run

```bash
# Build unified visualizer
go build -o fecim-visualizer ./cmd/fecim-visualizer

# Run unified app
./fecim-visualizer

# Or use launch script
./launch.sh
```

## Embedded App Pattern

Each demo follows this pattern for embedding in the unified GUI:

```go
// pkg/gui/embedded.go
type EmbeddedXxxApp struct {
    // internal state
}

func NewEmbeddedXxxApp() *EmbeddedXxxApp { ... }
func (app *EmbeddedXxxApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject { ... }
func (app *EmbeddedXxxApp) Start() { ... }  // Called when tab selected
func (app *EmbeddedXxxApp) Stop() { ... }   // Called when tab deselected
```

## Quick Function Lookups

| Need | File:Line | Function |
|------|-----------|----------|
| Quantize to 30 levels | module2-crossbar/pkg/crossbar/array.go:90 | `QuantizeTo30Levels()` |
| Create crossbar array | module2-crossbar/pkg/crossbar/array.go:45 | `NewArray()` |
| Run MVM | module2-crossbar/pkg/crossbar/array.go:123 | `Array.MVM()` |
| Create Preisach model | module1-hysteresis/pkg/ferroelectric/preisach.go:30 | `NewPreisachModel()` |
| Get P-E loop data | module1-hysteresis/pkg/ferroelectric/preisach.go:168 | `GetHysteresisLoop()` |
| IR drop analysis | module2-crossbar/pkg/crossbar/nonidealities.go | `AnalyzeIRDrop()` |
| Sneak path analysis | module2-crossbar/pkg/crossbar/nonidealities.go | `AnalyzeSneakPaths()` |
| FeCIM theme colors | shared/theme/theme.go:12 | `ColorPrimary`, `ColorBackground` |
| Create logger | shared/logging/logging.go:19 | `NewLogger()` |
