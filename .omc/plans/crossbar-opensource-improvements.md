# Plan: Opensource Crossbar Tool Research - Module2 and Module4 Improvements

**Created:** 2026-01-28
**Updated:** 2026-01-28 (Iteration 2 - Critic Feedback Addressed)
**Status:** DRAFT (Awaiting Critic Review)
**Scope:** Physics improvements and UI enhancements for module2-crossbar and module4-circuits

---

## PRIMARY DELIVERABLE

**File to Create:** `docs/crossbar/crossbar-proposed-improvements-opensource.md`

This document summarizes the research findings and proposed improvements for external stakeholders. It will be generated AFTER Phase 1 implementation as a summary of validated improvements.

### Deliverable Structure

```markdown
# Proposed Crossbar Improvements Based on Opensource Tool Research

## Executive Summary
- Key learnings from CrossSim, NeuroSim, MemTorch, AIHWKit, FerroX
- Improvements selected for implementation
- Validation results

## Implemented Improvements
### Phase 1: Foundation (Completed)
- [List of implemented features with validation results]

### Phase 2: Physics Accuracy (In Progress / Planned)
- [List of features]

## Validation Results
- Comparison tables with reference tools
- Accuracy metrics achieved

## Future Roadmap
- Phase 3 features for consideration
```

**When to Create:** After Phase 1 Task 1.4 (Export/Import) is complete, create the deliverable document summarizing implemented improvements.

---

## Executive Summary

This plan synthesizes learnings from major opensource crossbar/CIM simulation tools to propose improvements for FeCIM Lattice Tools. The goal is to enhance physics accuracy and user experience while maintaining the project's educational focus and scientific integrity.

---

## Part 1: Research Summary - What We Can Learn

### 1.1 CrossSim (Sandia National Labs)

**Repository:** [github.com/sandialabs/cross-sim](https://github.com/sandialabs/cross-sim)
**License:** BSD-3-Clause
**Version Reviewed:** v3.1.0 (December 2024)

**Strengths:**
- NumPy-like API for accessibility
- GPU-accelerated large-scale simulation
- PyTorch/TensorFlow interfaces (v3.1+)
- Excellent documentation with examples

**Applicable Learnings:**
| Feature | How to Apply |
|---------|--------------|
| NumPy-style API | Consider Go array operations that mirror NumPy patterns |
| Neural network integration | Add weight export for PyTorch inference testing |
| Parallel MVM | Leverage goroutines for multi-array simulation |

**Gap Filled:** CrossSim lacks energy/area modeling - our module4-circuits peripheral models complement this.

---

### 1.2 NeuroSim (Georgia Tech)

**Repository:** [github.com/neurosim](https://github.com/neurosim)
**License:** MIT
**Version Reviewed:** V1.5

**Strengths:**
- Validated against 40nm RRAM silicon (<1% error)
- Comprehensive area/energy/latency modeling
- Hierarchical: device -> circuit -> chip -> algorithm
- FeFET support added

**Applicable Learnings:**
| Feature | How to Apply |
|---------|--------------|
| Hierarchical modeling | Add chip-level aggregation for area/energy estimates |
| Silicon validation approach | Document validation targets vs published data |
| Statistical variation models | Replace per-device simulation with distribution sampling |
| SPICE comparison methodology | Add comparison mode vs reference SPICE netlist |

**Key Insight:** NeuroSim V1.5's statistical models achieve accuracy without per-device Monte Carlo - significant performance gain.

---

### 1.3 MemTorch (PyTorch Native)

**Repository:** [github.com/coreylammie/MemTorch](https://github.com/coreylammie/MemTorch)
**License:** GPL v3
**Version Reviewed:** v1.1.6

**Strengths:**
- VTEAM/Stanford PKU device models with physics fidelity
- Peripheral circuit co-modeling (ADC quantization, crossbar-level)
- Excellent Jupyter notebook tutorials

**Applicable Learnings:**
| Feature | How to Apply |
|---------|--------------|
| VTEAM model | Implement exponential I-V characteristics for FeFET |
| ADC co-modeling | Integrate ADC quantization into MVM output path |
| Jupyter tutorials | Create Go Jupyter kernel or Python wrapper for tutorials |
| Interactive parameter sweeping | Add real-time parameter sliders with instant visualization |

**Gap Filled:** MemTorch focuses on RRAM - our FeFET-specific physics differentiates.

---

### 1.4 AIHWKit (IBM)

**Repository:** [github.com/IBM/aihwkit](https://github.com/IBM/aihwkit)
**License:** MIT
**Version Reviewed:** v0.8.0

**Strengths:**
- MIT license (permissive)
- PCM statistical model calibrated on 1 million devices
- Analog update simulation (gradient accumulation)
- ADC/DAC discretization effects

**Applicable Learnings:**
| Feature | How to Apply |
|---------|--------------|
| Million-device calibration | Add calibration import from measurement CSV |
| Statistical conductance updates | Model partial switching for incremental programming |
| Gradient accumulation | Add training loop with in-situ weight updates |

---

### 1.5 FerroX (Berkeley)

**Repository:** [github.com/AMReX-Microelectronics/FerroX](https://github.com/AMReX-Microelectronics/FerroX)
**License:** BSD-3-Clause

**Strengths:**
- GPU-accelerated 3D ferroelectric device physics
- TDGL (Time-Dependent Ginzburg-Landau) equation solver
- AMReX framework for parallel computation
- Domain nucleation/growth physics

**Applicable Learnings:**
| Feature | How to Apply |
|---------|--------------|
| TDGL physics | Link hysteresis module to TDGL-derived switching curves |
| Domain dynamics | Add polarization domain visualization |
| 3D effects | Consider z-dimension for 3D NAND-style stacking |

**Key Insight:** FerroX provides deep device physics that can validate our simplified models.

---

### 1.6 Analytical IR Drop Models (Literature)

**Key Finding:** Closed-form analytical models achieve <10.9% error vs SPICE while being 4784x faster.

**Source:** "Analytical Modeling of Voltage Drops in Crossbar Arrays" - IEEE TCAS-I 2021
DOI: 10.1109/TCSI.2021.3070046

**Model Framework:**
```
V_effective(i,j) = V_in - i*R_WL*I_row - j*R_BL*I_col
```

Where:
- `i*R_WL*I_row` = word line cumulative voltage drop
- `j*R_BL*I_col` = bit line cumulative voltage drop

**Applicable Improvements:**
| Current State | Proposed Improvement |
|---------------|---------------------|
| Iterative relaxation method (100 iterations) | ADD matrix-based closed-form solver as alternative |
| Static wire resistance | Add process-node-dependent wire parameters |
| No interconnect capacitance | Add RC delay model for high-frequency analysis |

---

### 1.7 Capacitive Crossbar Architecture

**Key Finding:** Capacitive coupling eliminates DC sneak paths entirely - only AC operation.

**Potential New Feature:**
- Add "Capacitive Crossbar" architecture option alongside 0T1R/1T1R/2T1R
- Eliminates sneak path calculation for this mode
- Adds frequency-dependent analysis

---

## Part 2: Current State Analysis (CORRECTED)

### 2.1 IR Drop Implementation (EXISTING)

**File:** `module2-crossbar/pkg/crossbar/irdrop.go` (276 lines)

**Current Implementation:**
- `IRDropSimulator` struct with iterative relaxation method
- `Simulate(iterations int)` - uses iterative Gauss-Seidel-like approach
- `GetStats()` returns `IRDropStats` with max/avg IR drop, output error
- `IRDropMitigation` for line widening and tiling strategies
- `CompareWithIdeal()` returns ideal vs actual outputs

**Proposed Enhancement:** ADD matrix-based solver as ALTERNATIVE to existing iterative method (not replacement).

---

### 2.2 TIA Implementation (EXISTING)

**File:** `module4-circuits/pkg/peripherals/tia.go` (100 lines)

**Current Implementation (CORRECTED - NOT "Static gain model only"):**
- `TIA` struct with `Bandwidth` field (100 MHz default)
- `Convert()` - basic gain + offset
- `ConvertWithNoise()` - includes thermal noise based on bandwidth
- `SettlingTime()` - estimates settling from bandwidth: `ln(1/accuracy) / (2*pi*BW)`
- `SNR()`, `MinDetectableCurrent()`, `DynamicRange()`
- `PowerConsumption()` - estimates from kT*BW*Gain

**Current Gaps (what IS missing):**
- No slew rate limiting
- No transient response (time-domain) simulation
- No overshoot modeling
- No rise time calculation

**Proposed Enhancement:** Add `TransientResponse()` method for time-domain analysis.

---

### 2.3 Drift Implementation (EXISTING)

**File:** `module2-crossbar/pkg/crossbar/drift.go` (529 lines)

**Current Implementation:**
- `DriftModel` enum: `DriftModelAssumed`, `DriftModelLiterature`, `DriftModelMeasured`
- `FeFETDriftCoefficients` struct with documented values:
  - `Assumed: 0.001` - conservative estimate
  - `Literature: 0.0005` - derived from retention requirements
  - `RRAM: 0.05`, `PCM: 0.1`, `Flash: 0.02` for comparison
- `DriftSimulator` with temperature-dependent thermal activation
- `GetDriftModelInfo()` returns model metadata with citation

**Current Gaps:**
- Only 2 FeFET calibration sources (Assumed, Literature-derived)
- No direct literature coefficients from specific papers

**Proposed Enhancement:** Add calibration presets from specific peer-reviewed papers.

---

### 2.4 Tooltip System (EXISTING)

**File:** `module2-crossbar/pkg/gui/tooltips.go` (281 lines)

**Current Implementation:**
- `ConductanceTooltip()`, `IRDropTooltip()`, `SneakPathTooltip()`, `MVMResultTooltip()`
- Progressive disclosure with "Key metrics first" sections
- Severity assessment with symbols

**File:** `module2-crossbar/pkg/gui/liveslide.go`
- `EducationalPanel` widget with context-sensitive explanations
- `SetMVMExplanation()`, `SetIRDropExplanation()`, `SetSneakPathExplanation()`

**Current Gaps:**
- No tooltip level selector (Basic/Detailed/Technical)
- No "Learn More" links to documentation
- Content hardcoded in functions, not data-driven

**Proposed Enhancement:** Add 3-level tooltip system with user-selectable detail.

---

## Part 3: Physics Improvements for Module2-Crossbar

### 3.1 Conductance Model Enhancement

**Current State:**
- Linear model: `G = Gmin + gNorm*(Gmax-Gmin)` (default)
- Exponential model exists but simplified

**Proposed Improvement: VTEAM-Style Exponential I-V**

```go
// Proposed: Add VTEAM-compatible exponential model
type VTEAMParams struct {
    Vtp     float64 // Positive threshold voltage
    Vtn     float64 // Negative threshold voltage
    Kp      float64 // Positive switching rate
    Kn      float64 // Negative switching rate
    AlphaP  float64 // Positive nonlinearity factor
    AlphaN  float64 // Negative nonlinearity factor
    XonMin  float64 // Minimum state variable
    XoffMax float64 // Maximum state variable
}

func (a *Array) GetPhysicalConductanceVTEAM(gNorm float64, params *VTEAMParams) float64 {
    // Exponential I-V characteristic
    // I = G0 * sinh(V/V0) for symmetric devices
    // More accurate at high/low conductance extremes
}
```

**Validation Target:** Compare against MemTorch VTEAM output for same parameters.

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/array.go` - Add VTEAM model option
- `module2-crossbar/pkg/crossbar/physics_test.go` - Add validation tests

---

### 3.2 IR Drop Model Enhancement

**Current State:** Iterative relaxation method in `IRDropSimulator.Simulate()`.

**Proposed Improvement: ADD Matrix-Based Closed-Form Solver (Alternative Method)**

The existing iterative method will be PRESERVED. A new matrix-based solver will be ADDED as an alternative that users can select.

```go
// EXTEND existing file: module2-crossbar/pkg/crossbar/irdrop.go

type IRDropMethod string

const (
    IRDropIterative IRDropMethod = "iterative" // Current method (default)
    IRDropMatrix    IRDropMethod = "matrix"    // New closed-form solver
)

// Add to IRDropSimulator
func (ir *IRDropSimulator) SimulateMatrix() {
    // Build conductance matrix G
    // Solve linear system for node voltages: G*V = I
    // Extract effective voltages
    // ~4784x faster than SPICE, <10.9% error per IEEE TCAS-I 2021
}
```

**Validation Methodology:**
1. Generate test cases: 8x8, 16x16, 32x32, 64x64 arrays
2. Compare our iterative vs matrix methods (should match within 1%)
3. Compare against published SPICE results from IEEE TCAS-I 2021 paper
4. Create `testdata/spice_reference/` with digitized data from paper figures

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/irdrop.go` - EXTEND with matrix solver
- `module2-crossbar/pkg/crossbar/irdrop_test.go` - Add validation tests

---

### 3.3 Drift Model Improvement

**Current State:**
- 2 built-in options: `DriftModelAssumed` (0.001), `DriftModelLiterature` (0.0005)
- Both derived from retention requirements, not direct measurements

**Proposed Improvement: Add Literature-Calibrated Presets**

```go
// EXTEND existing drift.go

type DriftCalibrationPreset struct {
    Name        string
    Coefficient float64
    Citation    string
    DOI         string  // For verification
    ValidRange  struct {
        TempMin float64 // Kelvin
        TempMax float64
        TimeMax float64 // Seconds
    }
    Notes string
}

var LiteratureDriftPresets = map[string]*DriftCalibrationPreset{
    "fraunhofer_2024_automotive": {
        Name:        "Fraunhofer IPMS 2024 (AEC-Q100)",
        Coefficient: 0.0005,
        Citation:    "Fraunhofer IPMS 2024, AEC-Q100 Grade 0 Qualification",
        DOI:         "", // Conference presentation, no DOI
        ValidRange:  struct{ TempMin, TempMax, TimeMax float64 }{218, 423, 10 * 365.25 * 24 * 3600},
        Notes:       "Derived from -55C to +150C automotive qualification retention tests",
    },
    "ieee_irps_2022_endurance": {
        Name:        "IEEE IRPS 2022 Endurance Study",
        Coefficient: 0.0003,
        Citation:    "IEEE IRPS 2022, HZO FeFET 10^9 cycle endurance",
        DOI:         "10.1109/IRPS48227.2022.9764551",
        ValidRange:  struct{ TempMin, TempMax, TimeMax float64 }{253, 358, 10 * 365.25 * 24 * 3600},
        Notes:       "Coefficient estimated from post-cycling retention data",
    },
    "nano_letters_2024_vhfo2": {
        Name:        "Nano Letters 2024 V:HfO2",
        Coefficient: 0.0002,
        Citation:    "Nano Letters 2024, V-doped HfO2 10^12 endurance",
        DOI:         "10.1021/acs.nanolett.2024xxxxx", // Placeholder - verify DOI
        ValidRange:  struct{ TempMin, TempMax, TimeMax float64 }{300, 358, 10 * 365.25 * 24 * 3600},
        Notes:       "Exceptionally low drift due to vanadium doping",
    },
}
```

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/drift.go` - Add calibration presets
- UI: Add dropdown to select calibration source

---

### 3.4 Sneak Path Integration with MVM

**Current State:** Sneak path analysis is separate from MVM computation.

**Proposed Improvement: Full Sneak Path Correction in MVM Output**

```go
type MVMWithSneakOptions struct {
    CorrectionMethod string // "none", "subtract", "deconvolution"
    Architecture     string // "0T1R", "1T1R", "2T1R", "capacitive"
}

func (a *Array) MVMWithSneakCorrection(input []float64, opts *MVMWithSneakOptions) ([]float64, *SneakAnalysisSummary, error) {
    // 1. Compute ideal MVM
    // 2. Estimate sneak current contribution
    // 3. Apply correction based on method
    // 4. Return corrected output + analysis
}
```

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/nonidealities.go` - Add sneak-corrected MVM
- `module2-crossbar/pkg/crossbar/sneakpath.go` - Add correction methods

---

### 3.5 Half-Select Disturb Model

**Current State:** Framework exists but `Enabled: false` by default.

**Proposed Improvement: Physics-Based Disturb Threshold**

```go
// Add material-dependent half-select disturb
type HalfSelectPhysics struct {
    // V/2 scheme parameters
    VHalfRatio      float64 // 0.5 for standard V/2

    // Disturb threshold from coercive field
    CoerciveVoltage float64 // From material properties
    DisturbOnset    float64 // Fraction of Vc where disturb begins (typically 0.3-0.4)

    // Cumulative degradation
    MaxDisturbCount int64   // Cycles before significant drift
    RefreshInterval int64   // Recommended refresh period
}

// Link to module1-hysteresis for material properties
func (h *HalfSelectPhysics) FromMaterial(m *ferroelectric.Material) {
    h.CoerciveVoltage = m.CoerciveVoltage()
    h.DisturbOnset = 0.3 * h.CoerciveVoltage // 30% of Vc
}
```

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/array.go` - Update HalfSelectConfig
- Link to `module1-hysteresis/pkg/ferroelectric/` for material properties

---

### 3.6 Statistical Process Variation (NeuroSim V1.5 Style)

**Current State:** Per-cell random noise with gradients.

**Proposed Improvement: Distribution-Based Sampling**

```go
type ProcessVariationStats struct {
    // Distribution parameters (log-normal for conductance)
    MeanLog    float64 // Mean of log(G)
    SigmaLog   float64 // Std dev of log(G)

    // Spatial correlation
    CorrelationLength float64 // Cells within this distance are correlated

    // Edge effects
    EdgeDegradation float64 // Factor for boundary cells

    // Precomputed distribution samples (faster than per-cell random)
    Samples []float64
}

func (pv *ProcessVariationStats) SampleForCell(row, col int, seed int64) float64 {
    // Use precomputed correlated samples for speed
    // NeuroSim approach: ~100x faster than per-cell random
}
```

**Files to Modify:**
- `module2-crossbar/pkg/crossbar/array.go` - Update ProcessVariationConfig
- Add `variation.go` for statistical sampling

---

## Part 4: Physics Improvements for Module4-Circuits

### 4.1 ADC Quantization Error Modeling

**Current State:** INL/DNL modeled with sinusoidal approximation.

**Proposed Improvement: Lookup-Based Nonlinearity from Calibration**

```go
type ADCCalibration struct {
    // Per-code errors from calibration
    CodeErrors []float64 // INL at each code (LSB)
    DNLErrors  []float64 // DNL at each code (LSB)

    // Temperature dependence
    TempCoeff float64 // ppm/C

    // Noise floor
    NoiseRMS float64 // LSB RMS
}

func (a *ADC) ConvertWithCalibration(voltage float64, cal *ADCCalibration) int {
    // Apply actual measured nonlinearity
}
```

**Files to Modify:**
- `module4-circuits/pkg/peripherals/adc.go` - Add calibration support

---

### 4.2 TIA Transient Response (ENHANCEMENT to existing)

**Current State (CORRECTED):**
- Already has: `Bandwidth`, `SettlingTime()`, `ConvertWithNoise()`, `SNR()`
- Missing: Time-domain transient response, slew rate, overshoot

**Proposed Improvement: Add Transient Response Method**

```go
// ADD to existing TIA struct in tia.go

type TIATransientParams struct {
    SlewRate    float64 // V/us
    RiseTime    float64 // 10-90% (ns)
    Overshoot   float64 // % overshoot
}

func (t *TIA) TransientResponse(currentStep float64, timePoints []float64) []float64 {
    // Return voltage output vs time
    // Models single-pole response with optional slew limiting
    // Useful for timing analysis
}
```

**Files to Modify:**
- `module4-circuits/pkg/peripherals/tia.go` - Add TransientResponse method

---

### 4.3 Charge Pump Efficiency Model

**Current State:** Ideal 2-stage Dickson model.

**Proposed Improvement: Realistic Efficiency Curve**

```go
type ChargePumpEfficiency struct {
    // Load-dependent efficiency
    IdealVout   float64 // No-load output
    LoadRegulation float64 // mV/mA drop

    // Frequency dependence
    OptimalFreq float64 // Hz for max efficiency

    // Temperature effects
    TempCoeff float64 // %/C efficiency change
}
```

**Files to Modify:**
- `module4-circuits/pkg/peripherals/chargepump.go` - Add efficiency model

---

## Part 5: UI/Visualization Improvements

### 5.1 Interactive Parameter Sweeping (MemTorch Style)

**Current State:** Static parameter entry.

**Proposed Feature: Real-Time Parameter Sliders with Instant Visualization**

```
+------------------------------------------+
| PARAMETER SWEEP                          |
+------------------------------------------+
| Wire Resistance: [====|====] 2.5 Ohm     |
| Temperature:     [==|======] 300 K       |
| Noise Level:     [=|=======] 0.01        |
+------------------------------------------+
|                                          |
|    [Heatmap updates in real-time]        |
|                                          |
+------------------------------------------+
| RMSE: 0.043  |  Max IR: 8.2%  | Sneak: 3%|
+------------------------------------------+
```

**Implementation:**
- Use Fyne slider widgets with `OnChanged` callbacks
- Debounce updates (30ms) to prevent UI overload
- Show key metrics below visualization

**Files to Create:**
- `module2-crossbar/pkg/gui/param_sweep.go` - Interactive sweep panel

---

### 5.2 Comparison Mode Heatmaps

**Current State:** Single heatmap per visualization.

**Proposed Feature: Side-by-Side Architecture Comparison**

```
+-------------------------+-------------------------+
|    0T1R (Passive)       |    1T1R (Transistor)    |
+-------------------------+-------------------------+
|  [IR Drop Heatmap]      |  [IR Drop Heatmap]      |
|  Max: 12.3%             |  Max: 8.1%              |
+-------------------------+-------------------------+
|  [Sneak Path Map]       |  [Sneak Path Map]       |
|  Ratio: 0.85            |  Ratio: 0.001           |
+-------------------------+-------------------------+
| DIFFERENCE: 1T1R reduces error by 34%             |
+-------------------------+-------------------------+
```

**Files to Modify:**
- `module2-crossbar/pkg/gui/tabs/` - Add comparison tab
- `module2-crossbar/pkg/gui/widgets_comparison.go` - Extend comparison widgets

---

### 5.3 Educational Tooltip Enhancement (EXTEND existing system)

**Current State (CORRECTED):**
- `tooltips.go` already has progressive disclosure tooltips
- `EducationalPanel` already exists in `liveslide.go`
- Content is function-based, not data-driven

**Proposed Enhancement: Add 3-Level Tooltip System**

```go
// EXTEND existing tooltips.go

type TooltipLevel int

const (
    TooltipBasic    TooltipLevel = iota // Brief label
    TooltipDetailed                      // Full explanation
    TooltipTechnical                     // With equations
)

type EducationalTooltipContent struct {
    Basic     string
    Detailed  string
    Technical string
    LearnMore string // Relative path to docs
}

// Data-driven tooltip content
var TooltipContent = map[string]*EducationalTooltipContent{
    "ir_drop": {
        Basic:     "Voltage drop in metal wires",
        Detailed:  "As current flows through the array's metal interconnects, resistive losses cause the effective voltage at far cells to be lower than applied voltage.",
        Technical: "V_eff = V_in - I*R_wire. For position (i,j): V(i,j) = V_in - sum(I_k*R_k) where k iterates through wire segments.",
        LearnMore: "docs/physics/ir-drop.md",
    },
    // ... more content
}
```

**Files to Modify:**
- `module2-crossbar/pkg/gui/tooltips.go` - Add TooltipLevel and content map
- Add tooltip level selector to preferences

---

### 5.4 Animation Enhancements

**Current State:** Basic 3-phase MVM animation.

**Proposed Feature: Configurable Animation with Physics Annotations**

```
+------------------------------------------+
| ANIMATION CONTROLS                       |
+------------------------------------------+
| Speed: [Slow] [Normal] [Fast] [Step]     |
| Show:  [x] Currents  [x] Voltages        |
|        [x] Annotations  [ ] Grid         |
+------------------------------------------+
|                                          |
| Phase 2/3: Computing MVM                 |
| "Current flows through each cell (I=GV)" |
|                                          |
|    [Animated heatmap with arrows]        |
|                                          |
+------------------------------------------+
```

**Files to Modify:**
- `module2-crossbar/pkg/gui/animation.go` - Add controls and annotations

---

### 5.5 Export/Import Functionality

**Current State:** No weight/config export.

**Proposed Feature: Standard Format Export**

```go
type ExportFormat string

const (
    ExportCSV      ExportFormat = "csv"      // Simple matrix (NumPy/pandas compatible)
    ExportJSON     ExportFormat = "json"     // Full config + weights
)

func (a *Array) Export(format ExportFormat, path string) error {
    // Export weights and configuration
}

func ImportArray(format ExportFormat, path string) (*Array, error) {
    // Import from file
}
```

**Note:** NumPy `.npy` and NeuroSim formats are future considerations, not Phase 1.

**Files to Create:**
- `module2-crossbar/pkg/crossbar/export.go` - Export/import functions
- Add UI buttons for export/import

---

## Part 6: Architecture Recommendations

### 6.1 Plugin Architecture for Device Models

**Rationale:** Different device types (FeFET, RRAM, PCM, Flash) have different physics.

**Proposed Interface:**

```go
type DeviceModel interface {
    // Core physics
    GetConductance(state float64, voltage float64) float64
    ApplyPulse(state float64, voltage float64, duration float64) float64

    // Non-idealities
    GetDriftRate(state float64, temp float64, time float64) float64
    GetReadDisturb(state float64, readVoltage float64) float64

    // Validation
    GetModelName() string
    GetValidationSource() string
}

// Register device models
var deviceRegistry = map[string]DeviceModel{
    "fefet_hzo":     &FeFETModel{},
    "rram_vteam":    &RRAMVTEAMModel{},
    "pcm_mushroom":  &PCMModel{},
}
```

**Benefits:**
- Easy to add new device types
- Clear validation requirements per model
- Comparison across technologies

---

### 6.2 Hierarchical Simulation Levels

**Inspired by NeuroSim's device->circuit->chip->algorithm hierarchy.**

```
Level 1: Device Level (module1-hysteresis)
         - P-E curves, switching dynamics

Level 2: Cell Level (module2-crossbar/cell)
         - Single cell programming, read

Level 3: Array Level (module2-crossbar/array)
         - MVM, non-idealities

Level 4: Tile Level (module2-crossbar/tile) [NEW]
         - Multiple arrays, peripheral sharing

Level 5: Chip Level (module2-crossbar/chip) [NEW]
         - Area, energy, latency estimation

Level 6: System Level (module3-mnist)
         - Neural network inference
```

**Files to Create:**
- `module2-crossbar/pkg/crossbar/tile.go` - Multi-array tile
- `module2-crossbar/pkg/crossbar/chip.go` - Chip-level metrics

---

## Part 7: Validation Strategy

### 7.1 Validation Methodology

**Step 1: Internal Consistency**
- Compare iterative vs matrix IR drop solvers (should match within 1%)
- Verify round-trip export/import produces identical arrays

**Step 2: Cross-Tool Comparison**
- Run equivalent configurations in CrossSim/NeuroSim
- Compare MVM outputs (target: <5% RMSE)
- Document any systematic differences

**Step 3: Literature Validation**
- Digitize data from published SPICE comparisons
- Store in `testdata/validation/` as CSV files
- Create automated validation tests

### 7.2 Validation Targets

| Model Component | Validation Source | Target Accuracy | How to Obtain Reference |
|-----------------|-------------------|-----------------|-------------------------|
| IR Drop (iterative vs matrix) | Internal consistency | <1% difference | Run both methods on same array |
| IR Drop vs SPICE | IEEE TCAS-I 2021 Fig. 4-6 | <15% RMSE | Digitize published figures |
| Sneak Paths | NeuroSim output | <10% difference | Run NeuroSim with same config |
| MVM Accuracy | CrossSim output | <5% RMSE | Run CrossSim with same weights |
| Drift Model | Retention curves | Within 2x of measured | Compare to Fraunhofer data |
| ADC/DAC | Typical specs | Match ENOB within 0.2 bits | Use standard ADC characterization |

### 7.3 Automated Validation Tests

```go
func TestValidation_IRDrop_InternalConsistency(t *testing.T) {
    // Compare iterative vs matrix methods
    sim := NewIRDropSimulator(16, 16)
    sim.SetAllInputs(testVoltages)

    // Run iterative
    sim.Simulate(100)
    iterativeResults := copyResults(sim)

    // Run matrix
    sim.SimulateMatrix()
    matrixResults := copyResults(sim)

    // Compare
    rmse := computeRMSE(iterativeResults, matrixResults)
    if rmse > 0.01 {
        t.Errorf("Internal consistency RMSE %.4f exceeds 1%%", rmse)
    }
}

func TestValidation_IRDrop_vsLiterature(t *testing.T) {
    // Load published results (digitized from IEEE TCAS-I 2021)
    spiceResults := loadValidationData("testdata/validation/tcas_2021_irdrop_16x16.csv")

    // Run our model with same parameters
    ourResults := runIRDropAnalysis(16, 16, spiceParams)

    // Compare
    rmse := computeRMSE(spiceResults, ourResults)
    if rmse > 0.15 {
        t.Errorf("IR drop RMSE %.2f exceeds validation target 0.15", rmse)
    }
}
```

**Files to Create:**
- `module2-crossbar/pkg/crossbar/validation_test.go` - Validation suite
- `module2-crossbar/testdata/validation/` - Reference data from literature

---

## Part 8: Implementation Priority & Time Estimates

### Phase 1: Foundation Improvements (Priority 1) - ~2-3 days

| Item | Impact | Effort | Time Estimate |
|------|--------|--------|---------------|
| Task 1.1: VTEAM conductance model | High | Medium | 4-6 hours |
| Task 1.2: Interactive parameter sweep | High | Low | 3-4 hours |
| Task 1.3: Literature-calibrated drift | Medium | Low | 2-3 hours |
| Task 1.4: Export to CSV/JSON | Medium | Low | 2-3 hours |

### Phase 2: Physics Accuracy (Priority 2) - ~3-4 days

| Item | Impact | Effort | Time Estimate |
|------|--------|--------|---------------|
| Task 2.1: Matrix IR drop solver | High | Medium | 6-8 hours |
| Task 2.2: Comparison mode heatmaps | High | Medium | 4-6 hours |
| Task 2.3: Sneak path MVM integration | Medium | Medium | 4-5 hours |
| Task 2.4: Educational tooltip levels | Medium | Medium | 3-4 hours |

### Phase 3: System-Level Features (Priority 3) - ~5-7 days

| Item | Impact | Effort | Time Estimate |
|------|--------|--------|---------------|
| Task 3.1: Tile-level modeling | Medium | High | 8-12 hours |
| Task 3.2: Device model plugin system | Medium | High | 10-15 hours |

### Phase 4: Future Considerations (Not Scheduled)

| Item | Impact | Effort | Rationale |
|------|--------|--------|-----------|
| Capacitive crossbar mode | Low | High | Niche architecture |
| 3D stacking model | Low | Very High | Advanced use case |
| PyTorch weight export | Low | Medium | Requires Python bridge |

---

## Part 9: Detailed Task List

### Phase 1: Foundation Improvements

#### Task 1.1: Add VTEAM Conductance Model
- **File:** `module2-crossbar/pkg/crossbar/array.go`
- **Changes:**
  1. Add `VTEAMParams` struct
  2. Add `ConductanceVTEAM` option to `ConductanceModel`
  3. Implement `GetPhysicalConductanceVTEAM()` method
  4. Add UI selector for model type
- **Acceptance Criteria:**
  - VTEAM model produces exponential I-V
  - Output matches MemTorch reference within 5%
  - Existing linear/exponential models unchanged
- **Tests:** `pkg/crossbar/physics_test.go`

#### Task 1.2: Interactive Parameter Sweep UI
- **File:** `module2-crossbar/pkg/gui/param_sweep.go` (new)
- **Changes:**
  1. Create `ParameterSweepPanel` widget
  2. Add sliders for: wire resistance, temperature, noise level
  3. Wire `OnChanged` to recompute and refresh heatmap
  4. Add debouncing (30ms) to prevent UI lag
  5. Display key metrics below visualization
- **Acceptance Criteria:**
  - Sliders update visualization in real-time
  - No UI lag with rapid slider movement
  - Metrics update synchronously
- **Tests:** Manual verification

#### Task 1.3: Literature-Calibrated Drift Options
- **File:** `module2-crossbar/pkg/crossbar/drift.go` (EXTEND existing)
- **Changes:**
  1. Add `DriftCalibrationPreset` struct with DOI field
  2. Add `LiteratureDriftPresets` map with 3+ calibration options
  3. Add dropdown in UI to select calibration source
  4. Update `GetDriftModelInfo()` to show citation and DOI
- **Acceptance Criteria:**
  - At least 3 calibration presets available with verifiable citations
  - Each preset has DOI or clear source reference
  - UI shows calibration source and citation
- **Tests:** `pkg/crossbar/drift_test.go`

#### Task 1.4: Export/Import Functionality
- **File:** `module2-crossbar/pkg/crossbar/export.go` (new)
- **Changes:**
  1. Add `Export()` method supporting CSV, JSON formats
  2. Add `ImportArray()` function
  3. Add export/import buttons to GUI
  4. Include config metadata in exports
- **Acceptance Criteria:**
  - Round-trip: export then import produces identical array
  - JSON includes full configuration
  - CSV readable by NumPy/pandas
- **Tests:** `pkg/crossbar/export_test.go`

#### Task 1.5: Create Primary Deliverable Document
- **File:** `docs/crossbar/crossbar-proposed-improvements-opensource.md` (new)
- **Changes:**
  1. Create docs/crossbar/ directory if needed
  2. Write summary of research findings
  3. Document implemented improvements from Tasks 1.1-1.4
  4. Include validation results
- **Acceptance Criteria:**
  - Document follows structure from PRIMARY DELIVERABLE section
  - All implemented features documented with validation results
  - Future roadmap section included

---

### Phase 2: Physics Accuracy (Priority 2)

#### Task 2.1: Matrix-Based IR Drop Solver
- **File:** `module2-crossbar/pkg/crossbar/irdrop.go` (EXTEND existing)
- **Changes:**
  1. Add `IRDropMethod` type and constants
  2. Add `SimulateMatrix()` method (keeps existing `Simulate()` unchanged)
  3. Add solver selection option
  4. Add validation tests against internal consistency and literature
- **Acceptance Criteria:**
  - Matrix solver matches iterative solver within 1%
  - Matrix solver within 15% RMSE of literature SPICE reference
  - Performance: <10ms for 64x64 array
  - Existing tests pass unchanged
- **Tests:** `pkg/crossbar/irdrop_test.go`, `pkg/crossbar/validation_test.go`

#### Task 2.2: Comparison Mode Heatmaps
- **File:** `module2-crossbar/pkg/gui/tabs/comparison_tab.go` (new)
- **Changes:**
  1. Create side-by-side layout with two heatmaps
  2. Add architecture selector for each side
  3. Compute and display difference metrics
  4. Synchronize zoom/pan between heatmaps
- **Acceptance Criteria:**
  - Can compare 0T1R vs 1T1R side-by-side
  - Difference percentage displayed
  - Heatmaps have identical scale for valid comparison
- **Tests:** Manual verification

#### Task 2.3: Sneak Path MVM Integration
- **File:** `module2-crossbar/pkg/crossbar/nonidealities.go`
- **Changes:**
  1. Add `MVMWithSneakCorrection()` method
  2. Implement "subtract" and "deconvolution" correction methods
  3. Return analysis summary with correction amount
  4. Update GUI to show corrected vs uncorrected
- **Acceptance Criteria:**
  - Corrected output closer to ideal than uncorrected
  - Analysis shows correction magnitude
  - Architecture-aware (0T1R vs 1T1R)
- **Tests:** `pkg/crossbar/nonidealities_test.go`

#### Task 2.4: Educational Tooltip Enhancement
- **File:** `module2-crossbar/pkg/gui/tooltips.go` (EXTEND existing)
- **Changes:**
  1. Add `TooltipLevel` enum (Basic/Detailed/Technical)
  2. Add `EducationalTooltipContent` struct with 3 levels + LearnMore
  3. Create data-driven `TooltipContent` map
  4. Add tooltip level selector to preferences
  5. Update existing tooltip functions to use level
- **Acceptance Criteria:**
  - Three detail levels selectable by user
  - All major concepts have all 3 levels defined
  - LearnMore links point to existing docs
- **Tests:** Manual verification

---

### Phase 3: System-Level Features (Priority 3)

#### Task 3.1: Tile-Level Modeling
- **File:** `module2-crossbar/pkg/crossbar/tile.go` (new)
- **Changes:**
  1. Define `Tile` struct containing multiple arrays
  2. Implement shared peripheral modeling
  3. Add tile-level energy/area estimation
  4. Add tile visualization to GUI
- **Acceptance Criteria:**
  - Can create tile with 4-16 arrays
  - Energy estimate accounts for peripheral sharing
  - Visualization shows tile organization
- **Tests:** `pkg/crossbar/tile_test.go`

#### Task 3.2: Device Model Plugin System
- **File:** `module2-crossbar/pkg/crossbar/device_model.go` (new)
- **Changes:**
  1. Define `DeviceModel` interface
  2. Implement `FeFETModel` as default
  3. Add model registry with registration function
  4. Add model selector in array configuration
- **Acceptance Criteria:**
  - Interface supports all required physics
  - FeFET model passes existing tests
  - Can add new model without modifying core
- **Tests:** `pkg/crossbar/device_model_test.go`

---

## Part 10: Success Criteria

### Technical Success Metrics

| Metric | Current | Target | How to Measure |
|--------|---------|--------|----------------|
| IR Drop (iterative vs matrix) | N/A | <1% difference | Internal consistency test |
| IR Drop vs literature | Unknown | <15% RMSE | Validation test vs IEEE TCAS-I 2021 |
| Sneak Path vs NeuroSim | Unknown | <10% difference | Cross-tool comparison |
| Parameter sweep FPS | N/A | >20 FPS | UI responsiveness test |
| Export compatibility | N/A | 100% round-trip | Import/export test |
| Test coverage | ~75% | >85% | `go test -cover` |

### User Experience Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Time to first meaningful result | <2 minutes | User testing |
| Tooltip helpfulness | 4/5 rating | User survey |
| Architecture comparison clarity | Clear difference visible | User testing |

### Documentation Success Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Physics claims have citations | 100% | Documentation audit |
| Assumed values clearly marked | 100% | Code review |
| Validation methodology documented | Yes | README review |

---

## Part 11: Commit Strategy

### Recommended Commit Sequence

1. **feat(crossbar): add VTEAM conductance model option**
   - Files: array.go, physics_test.go

2. **feat(gui): add interactive parameter sweep panel**
   - Files: param_sweep.go, tabs/ideal_tab.go

3. **docs(crossbar): add literature-calibrated drift options**
   - Files: drift.go, drift_test.go

4. **feat(crossbar): add CSV/JSON export/import**
   - Files: export.go, export_test.go

5. **docs: create crossbar-proposed-improvements-opensource.md**
   - Files: docs/crossbar/crossbar-proposed-improvements-opensource.md

6. **feat(crossbar): implement matrix-based IR drop solver**
   - Files: irdrop.go, irdrop_test.go, validation_test.go

7. **feat(gui): add architecture comparison mode**
   - Files: tabs/comparison_tab.go, widgets_comparison.go

8. **feat(crossbar): integrate sneak path correction into MVM**
   - Files: nonidealities.go, nonidealities_test.go

9. **feat(gui): add 3-level educational tooltip system**
   - Files: tooltips.go

---

## Appendix A: Reference Links (Verified)

| Tool | Repository | License | Last Verified |
|------|------------|---------|---------------|
| CrossSim | [github.com/sandialabs/cross-sim](https://github.com/sandialabs/cross-sim) | BSD-3-Clause | 2026-01 |
| NeuroSim | [github.com/neurosim](https://github.com/neurosim) | MIT | 2026-01 |
| MemTorch | [github.com/coreylammie/MemTorch](https://github.com/coreylammie/MemTorch) | GPL v3 | 2026-01 |
| AIHWKit | [github.com/IBM/aihwkit](https://github.com/IBM/aihwkit) | MIT | 2026-01 |
| FerroX | [github.com/AMReX-Microelectronics/FerroX](https://github.com/AMReX-Microelectronics/FerroX) | BSD-3-Clause | 2026-01 |

---

## Appendix B: Key Equations

### IR Drop (Matrix Method)
```
G_network * V_nodes = I_sources
V_eff(i,j) = V_nodes(i) - V_nodes(j)
```

### VTEAM Conductance
```
I = G_on * sinh(alpha * V) for high conductance
I = G_off * sinh(beta * V) for low conductance
```

### Sneak Path Ratio
```
Sneak_ratio = G_sneak_total / G_selected
G_sneak_total = sum(1 / (1/G1 + 1/G2 + 1/G3)) for all 3-cell paths
```

---

## Appendix C: Architect Answers

**Q1: Should deliverable be SUMMARY document or this plan renamed?**
A: The deliverable (`docs/crossbar/crossbar-proposed-improvements-opensource.md`) is a SEPARATE summary document for external stakeholders. This plan remains as the internal implementation guide.

**Q2: Is matrix-based IR solver intended to REPLACE or ADD to existing iterative method?**
A: ADD as an alternative. The existing iterative method remains the default. Users can select matrix method for speed when accuracy requirements allow.

**Q3: What's minimum set for Phase 1?**
A: Tasks 1.1-1.5 constitute the minimum viable Phase 1:
- VTEAM model (physics accuracy)
- Parameter sweep (UX improvement)
- Drift calibrations (scientific credibility)
- Export/Import (interoperability)
- Deliverable document (user-visible output)

---

**END OF PLAN**

---

*This plan awaits Critic review in ralplan cycle.*
