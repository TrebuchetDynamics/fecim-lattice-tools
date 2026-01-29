---
Module: module2-crossbar
Name: "Crossbar Array MVM Visualization"
Entry: cmd/crossbar-gui/main.go
Package: module2-crossbar/pkg/gui
Last Updated: 2026-01-29
Description: "Interactive visualization of FeCIM crossbar array matrix-vector multiplication with non-ideality analysis"
---

Bugs:
  - [x] BUG-M2-001: Potential race condition on lastInput/lastOutput - FIXED (2026-01-26): Added mutex protection
  - [x] BUG-M2-002: Missing fyne.Do wrapper in updateConductanceDisplay - FIXED (2026-01-26): Added fyne.Do wrappers
  - [x] BUG-M2-003: Heatmap refresh during startup can cause layout oscillation (partially mitigated)
  - [x] BUG-M2-004: Educational content wrapping disabled but can still trigger MinSize changes
  - [x] BUG-M2-005: Auto-demo context cancellation may leak if Stop() not called - FIXED (2026-01-26): Added window close cleanup
  - [x] BUG-M2-006: Missing Run MVM button in enhanced mode - OBSOLETE (2026-01-28): Button removed, auto-run on all changes
  - [x] BUG-M2-007: Missing 2T1R architecture button in enhanced mode - FIXED (2026-01-28): Added arch2T1RBtn for feature parity
  - [x] BUG-M2-008: Inconsistent ADC label format between modes - FIXED (2026-01-28): Standardized to "X bits" format
  - [x] BUG-M2-009: Colormap selector falls through on non-heatmap tabs - FIXED (2026-01-28): Added explicit return for non-heatmap tabs
  - [x] BUG-M2-010: Array size slider missing min/max labels - FIXED (2026-01-28): Added 8/128 labels flanking slider
  - [x] BUG-M2-011: Hover info fixed width causes layout issues - FIXED (2026-01-28): Reduced to 300px for better responsiveness
  - [x] BUG-M2-012: Embedded mode missing onboarding content - FIXED (2026-01-28): Added setEducationalContent call
  - [x] BUG-M2-013: IR Drop/Sneak legends showing fixed 100% - FIXED (2026-01-28): Now shows passive worst-case baseline
  - [x] BUG-M2-014: Heatmap colors not scaled to baseline - FIXED (2026-01-28): GetIRDropMapWithScale/GetSneakMapWithScale
  - [x] BUG-M2-015: Baseline not reset on array resize - FIXED (2026-01-28): Reset baselines in recreateArray()
  - [x] BUG-M2-016: Ideal vs Actual showing random approximations - FIXED (2026-01-28): Uses GetEffectiveConductanceMatrix()

# UI Structure

## Main Layout
```
+-----------------------------------------------------------------------------------+
|                         FeCIM Crossbar Array Visualization                         |
+-----------------------------------------------------------------------------------+
|  LEFT PANEL (15%)  |           CENTER PANEL (60%)              |  RIGHT PANEL (25%)|
|                    |                                            |                   |
| +----------------+ | +----------------------------------------+ | +---------------+ |
| | Architecture   | | |  Conductance | IR Drop | Sneak Paths  | | | Array: [slider]| |
| | Info Card      | | |  Input/Output | Ideal vs Actual       | | | 8 [====] 128  | |
| |                | | |  Accuracy Analysis                     | | | Size: 64x64   | |
| | Title          | | +----------------------------------------+ | +---------------+ |
| | Description    | | |                                        | | | Architecture  | |
| | Advantages     | | |         HEATMAP / VISUALIZATION        | | | [PASSIVE]     | |
| | Tradeoffs      | | |                                        | | | [1T1R GATE]   | |
| |                | | |         (with color legend)            | | | [2T1R]        | |
| +----------------+ | |                                        | | +---------------+ |
| | Best for:      | | |                                        | | | Noise: [===]2%| |
| | N² Operations  | | |                                        | | | ADC: [====]6b | |
| | 256 MACs       | | |                                        | | +---------------+ |
| +----------------+ | +----------------------------------------+ | | Color: [fecim]| |
|                    |                                            | +---------------+ |
|                    |                                            | | [Reset][Export]| |
|                    |                                            | +---------------+ |
|                    |                                            | | Live Metrics  | |
|                    |                                            | | Accuracy: 90% | |
|                    |                                            | | Energy: 9.6pJ | |
|                    |                                            | +---------------+ |
|                    |                                            | | Cell Details  | |
|                    |                                            | | [15,6] L12... | |
+-----------------------------------------------------------------------------------+
| IDLE | Status: Ready | Hover info: [row,col] | Crossbar: 64x64 | Levels: 30      |
+-----------------------------------------------------------------------------------+
```

## Component Hierarchy

### Header
- **Type**: container.VBox
- **File**: app.go
- **Children**:
  - titleLabel: "FeCIM Crossbar Array Visualization" (bold, centered)
  - widget.Separator

### Left Panel (Educational)
- **Type**: container.VScroll > container.VBox
- **File**: app_controls.go:createLeftPanel()
- **Components**:
  - **eduTitleLabel**: Architecture name (e.g., "1T1R Architecture")
  - **eduContentLabel**: Multi-line description with advantages/tradeoffs
  - **keyStatLabel**: "N² Operations" (subtitle)
  - **keyStatValue**: "256 MACs" (dynamic, bold)

### Center Panel (Tabs)
- **Type**: container.AppTabs
- **File**: app_enhanced.go
- **Tabs**:

  1. **Conductance Tab**
     - Content: container.Border(right=condLegend, center=conductanceHeatmap)
     - condLegend: ColorLegend (0-29 levels, "fecim" colormap)
     - conductanceHeatmap: CrossbarHeatmap (rows×cols, tap/hover callbacks)

  2. **IR Drop Tab**
     - Content: container.Border(right=irLegend, center=irDropHeatmap)
     - irLegend: ColorLegend (0 to baselineMaxIRDrop%, "viridis" colormap)
     - irDropHeatmap: CrossbarHeatmap (scaled to passive baseline)

  3. **Sneak Paths Tab**
     - Content: container.Border(right=sneakLegend, center=sneakPathHeatmap)
     - sneakLegend: ColorLegend (0 to baselineMaxSneak%, "plasma" colormap)
     - sneakPathHeatmap: CrossbarHeatmap (scaled to passive baseline)

  4. **Input/Output Tab**
     - Content: container.Max(mvmVis)
     - mvmVis: MVMVisualization (input chart, output chart, mini matrix)

  5. **Ideal vs Actual Tab**
     - Content: container.Border(top=title, center=beforeAfterToggle)
     - beforeAfterToggle: BeforeAfterToggle (split/before/after/diff modes)
     - Uses GetConductanceMatrix() vs GetEffectiveConductanceMatrix()

  6. **Accuracy Analysis Tab**
     - Content: container.Border(top=title, center=accuracyWaterfall)
     - accuracyWaterfall: AccuracyWaterfall (step-by-step degradation chart)

### Right Panel (Controls + Metrics)
- **Type**: container.Border(top=controls, center=metricsScroll)
- **File**: app_controls.go:createRightPanel()

#### Controls Section (Fixed Height)
```
+---------------------------+
| Array: [8|====slider====|128]
|         Size: 64×64       |
+---------------------------+
|      Architecture         |
| [PASSIVE] [1T1R] [2T1R]   |
+---------------------------+
| Noise: [========] 2%      |
| ADC:   [========] 6 bits  |
+---------------------------+
| Color: [fecim ▼]          |
+---------------------------+
| [  Reset  ] [  Export  ]  |
+---------------------------+
```

**Components**:
- **arraySizeSlider**: Slider(8-128, step=8) with min/max labels
- **arraySizeLabel**: "Size: 64×64"
- **archToggle**: 3 buttons (PASSIVE, 1T1R GATE, 2T1R)
- **noiseSlider**: Slider(0-10%, step=1%)
- **noiseLabel**: "2%"
- **adcBitsSlider**: Slider(4-10 bits, step=1)
- **adcBitsLabel**: "6 bits"
- **colormapSelect**: Select(fecim, viridis, plasma, coolwarm)
- **resetButton**: "Reset" - programs new random weights
- **exportButton**: "Export" - saves weights CSV + analysis JSON

#### Metrics Section (Scrollable)
- **metricsPanel**: Live accuracy, energy, performance metrics
- **comparisonBadge**: FeCIM vs GPU energy comparison
- **statsLabel**: Detailed cell analysis (monospace)

### Footer (Status Bar)
- **Type**: container.HBox
- **File**: app_controls.go:createStatusFooter()
- **Components**:
  - **modeIndicator**: ModeIndicatorBox (IDLE/COMPUTE/WRITE/READ)
  - **statusLabel**: Current operation status
  - **hoverInfoLabel**: Cell hover details
  - **infoLabel**: Array config info (size, levels, noise, ADC)

## Data Flow

### Auto-Run MVM (No Manual Button)
All parameter changes automatically trigger MVM recalculation:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│ Array Size      │────▶│ recreateArray() │────▶│ runEnhancedMVM  │
│ Slider Change   │     │ + reset baseline│     │ Instant()       │
└─────────────────┘     └─────────────────┘     └─────────────────┘

┌─────────────────┐     ┌─────────────────┐
│ Noise/ADC       │────▶│ runEnhancedMVM  │
│ Slider Change   │     │ Instant()       │
└─────────────────┘     └─────────────────┘

┌─────────────────┐     ┌─────────────────┐
│ Architecture    │────▶│ runEnhancedMVM  │ (preserves input vector)
│ Toggle Change   │     │ WithCurrentInput│
└─────────────────┘     └─────────────────┘
```

### Baseline Scaling System
IR Drop and Sneak Path heatmaps use passive (0T1R) worst-case as baseline:

```
┌─────────────────────────────────────────────────────────────────┐
│                    BASELINE COMPUTATION                          │
├─────────────────────────────────────────────────────────────────┤
│ 1. On first MVM (or after array resize):                        │
│    - baselineMaxSneak = AnalyzeSneakPathsWithIsolation(1.0)     │
│    - baselineMaxIRDrop = AnalyzeIRDrop(passiveParams)           │
│                                                                  │
│ 2. Baselines stay FIXED regardless of architecture              │
│                                                                  │
│ 3. Legend shows: 0% to baseline% (e.g., 0% to 87.5%)            │
│                                                                  │
│ 4. Heatmap data scaled to baseline:                             │
│    - irMap = GetIRDropMapWithScale(baselineIR / 100)            │
│    - sneakMap = GetSneakMapWithScale(baselineSneak / 100)       │
└─────────────────────────────────────────────────────────────────┘

Result: 1T1R shows ~0.1% of legend, 2T1R shows ~0.01% of legend
```

### Cell Selection Sync
```
User clicks cell on ANY heatmap
        │
        ▼
┌─────────────────┐
│ syncSelection() │ → Updates selectedRow/selectedCol
└────────┬────────┘
         │
         ├──▶ conductanceHeatmap.SetSelection(row, col)
         ├──▶ irDropHeatmap.SetSelection(row, col)
         ├──▶ sneakPathHeatmap.SetSelection(row, col)
         ├──▶ beforeAfterToggle heatmaps (if exists)
         │
         ▼
┌─────────────────┐
│ Generate tooltip│ → Updates statsLabel with detailed cell info
└─────────────────┘
```

## File Structure

```
module2-crossbar/pkg/gui/
├── app.go              # Main app struct, window setup, standard mode
├── app_controls.go     # Control widgets, right panel, footer
├── app_enhanced.go     # Enhanced mode tabs, legends, widgets
├── app_analysis.go     # MVM result processing, baseline computation
├── analysis.go         # IR drop, sneak path analysis functions
├── callbacks.go        # Cell tap/hover callbacks, selection sync
├── tooltips.go         # Tooltip generation (Conductance/IR/Sneak)
├── heatmap.go          # CrossbarHeatmap widget
├── vectors.go          # MVMVisualization, input/output charts
├── liveslide.go        # ModeIndicatorBox, EducationalPanel
├── widgets.go          # MetricsPanel, ComparisonBadge, etc.
└── embedded.go         # EmbeddedApp interface for unified launcher
```

## Key State Variables

```go
type CrossbarApp struct {
    // Core
    array  *crossbar.Array
    config *crossbar.Config

    // Heatmaps
    conductanceHeatmap *CrossbarHeatmap
    irDropHeatmap      *CrossbarHeatmap
    sneakPathHeatmap   *CrossbarHeatmap

    // Legends (show passive baseline as max)
    condLegend  *ColorLegend  // 0-29 levels
    irLegend    *ColorLegend  // 0-baselineMaxIRDrop%
    sneakLegend *ColorLegend  // 0-baselineMaxSneak%

    // Baseline values (computed once from passive, reset on array resize)
    baselineMaxIRDrop float64  // e.g., 2.96%
    baselineMaxSneak  float64  // e.g., 87.5%

    // Selection (synced across all heatmaps)
    selectedRow int
    selectedCol int

    // Architecture
    architecture string  // "0T1R (Passive)", "1T1R (Transistor)", "2T1R (Dual Transistor)"

    // Controls
    resetButton     *widget.Button
    arraySizeSlider *widget.Slider
    noiseSlider     *widget.Slider
    adcBitsSlider   *widget.Slider
    colormapSelect  *widget.Select

    // Protected state (stateMu)
    lastInput     []float64
    lastOutput    []float64
    lastMVMResult *crossbar.MVMResult
}
```

## Architecture Toggle Behavior

| Architecture | Isolation Factor | Sneak Path | IR Drop | Use Case |
|--------------|------------------|------------|---------|----------|
| PASSIVE (0T1R) | 1.0 | ~87.5% | ~3% | Simplest, highest density |
| 1T1R GATE | 0.001 | ~0.09% | ~1% | Industry standard, balanced |
| 2T1R | 0.0001 | ~0.01% | ~0.5% | Best isolation, most complex |

All changes auto-trigger MVM with `runEnhancedMVMWithCurrentInput()` to compare same computation across architectures.
