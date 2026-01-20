# Demo 2: Crossbar Array MVM Visualization

**Ferroelectric CIM Visualizer - Compute-in-Memory Crossbar**

> *"Compute in memory where the same device does the memory and the computation."* — Dr. external research group

## Overview

Demo 2 provides an interactive visualization of Matrix-Vector Multiplication (MVM) in a ferroelectric crossbar array. This demo illustrates how Ferroelectric CIM performs analog neural network inference using physical Ohm's law and Kirchhoff's current law, achieving massive parallelism.

### What This Demo Shows

1. **Matrix-Vector Multiplication (MVM)** — Parallel analog computation using conductance × voltage = current
2. **30 Discrete Conductance Levels** — Each cell stores ~4.9 bits (30 states) of synaptic weight
3. **Non-Idealities Modeling** — IR drop, sneak paths, device variation, ADC quantization
4. **Real-time Crossbar Visualization** — Interactive heatmap with cell-level inspection

## Quick Start

```bash
# Navigate to demo directory
cd demo2-crossbar

# Build the GUI application
go build -o crossbar-gui ./cmd/crossbar-gui

# Run Fyne GUI mode (recommended)
./crossbar-gui

# Or run terminal version with specific analysis
go run ./cmd/inference --show-mvm
go run ./cmd/inference --show-irdrop
go run ./cmd/inference --show-sneak
go run ./cmd/inference --show-nonidealities
```

## Run Modes

### 1. Fyne GUI Mode (Recommended)

Cross-platform native GUI featuring:
- **Interactive heatmap visualization** with click-to-select cells
- **Three tabbed views**: Conductance, IR Drop, Sneak Paths
- **Real-time control panel** with sliders:
  - Array size (8×8 to 128×128)
  - Noise level (0-20%)
  - ADC resolution (4-10 bits)
- **Custom "Ferroelectric CIM" colormap** matching 30 discrete levels
- **30-level discrete indicator widget**
- **Vector bar charts** for input/output visualization
- **One-click MVM, IR Drop, and Sneak Path analysis**
- **RMSE comparison charts** (ideal vs actual)
- **Live statistics panel**

**GUI Controls:**
| Control | Function |
|---------|----------|
| Array Size Slider | Resize crossbar (8×8 to 128×128) |
| Noise Slider | Device-to-device variation (0-20%) |
| ADC Bits Slider | ADC resolution (4-10 bits) |
| Colormap Dropdown | ferroelectric-cim, viridis, plasma, coolwarm |
| Run MVM | Execute matrix-vector multiplication |
| Analyze IR Drop | Show voltage drop heatmap |
| Analyze Sneak Paths | Show sneak current map |
| Reset Array | Reprogram random weights |

**Heatmap Interaction:**
- Click any cell to see its conductance level (0-29)
- Right-click to clear selection
- Tabs switch between Conductance, IR Drop, and Sneak Path views
- Yellow border highlights selected/worst-case cells

### 2. Terminal Mode

Command-line analysis with ASCII visualization:
```bash
# Show MVM computation
go run ./cmd/inference --show-mvm

# Show IR drop analysis
go run ./cmd/inference --show-irdrop

# Show sneak path analysis
go run ./cmd/inference --show-sneak

# Show all non-idealities combined
go run ./cmd/inference --show-nonidealities
```

## Physics Model

### Matrix-Vector Multiplication (MVM)

The crossbar array performs parallel MVM using Ohm's law:

```
Input Vector (Voltages)
    V₀  V₁  V₂  V₃  V₄  V₅  V₆  V₇
    ↓   ↓   ↓   ↓   ↓   ↓   ↓   ↓
   ┌───┬───┬───┬───┬───┬───┬───┬───┐
I₀ │G₀₀│G₀₁│G₀₂│G₀₃│G₀₄│G₀₅│G₀₆│G₀₇│→ I₀ = Σⱼ Gᵢⱼ × Vⱼ
I₁ │G₁₀│G₁₁│G₁₂│G₁₃│G₁₄│G₁₅│G₁₆│G₁₇│→ I₁ = Σⱼ Gᵢⱼ × Vⱼ
I₂ │G₂₀│G₂₁│G₂₂│G₂₃│G₂₄│G₂₅│G₂₆│G₂₇│→ I₂ = Σⱼ Gᵢⱼ × Vⱼ
   └───┴───┴───┴───┴───┴───┴───┴───┘

Output Current = Weight Matrix × Input Vector
     I        =        G      ×      V
```

**Formula:**
```
I[i] = Σⱼ G[i,j] × V[j]
```

Where:
- `G[i,j]` = Conductance of cell (i,j) — represents synaptic weight
- `V[j]` = Input voltage on column j
- `I[i]` = Output current on row i

### Non-Idealities Modeled

| Effect | Description | Impact | Mitigation |
|--------|-------------|--------|------------|
| **IR Drop** | Voltage attenuation along wires due to resistance | Cells far from driver see lower voltage | Add driver amplifiers, limit array size |
| **Sneak Paths** | Parasitic currents through unselected cells | Corrupts output currents | Use 1T1R (transistor) or selector devices |
| **Device Variation** | Cell-to-cell conductance spread (σ/μ) | Reduces effective precision | Scheme C programming, calibration |
| **ADC Quantization** | Limited output bit precision | Quantization noise | Higher resolution ADC (6+ bits) |

---

## Proposed Improvements (From Literature Analysis)

### 1. Scheme C Incremental Amplitude Programming (Priority: CRITICAL)

**Reference:** Oh et al. "HfZrOₓ-based Ferroelectric Synapse Device with 32 levels" IEEE EDL (2017)

**Current Gap:** May be using constant-amplitude pulses (Scheme A) which causes state bunching.

**The Bug:** Three pulse schemes compared:
- **Scheme A** (Identical pulses) — ❌ FAILS due to domain screening
- **Scheme B** (Variable width) — ⚠️ Works but complex timing
- **Scheme C** (Incremental amplitude) — ✅ **SOLUTION for 30 linear states**

**Implementation:**
```go
// Scheme C: Incremental Amplitude Pulses (ISPP)
func ProgramToLevel(device *FeFET, targetLevel int) error {
    const (
        V_start = 1.0   // Starting voltage (V)
        V_step  = 0.05  // 50mV increment per level
        pulseWidth = 100 // nanoseconds (fixed)
    )

    for i := 0; i < targetLevel; i++ {
        voltage := V_start + float64(i) * V_step
        device.ApplyPulse(voltage, pulseWidth)
        time.Sleep(10 * time.Microsecond)  // Recovery time
    }
    return nil
}
```

**Key Parameters:**
| Parameter | Value | Notes |
|-----------|-------|-------|
| V_start | 1.0V | Initial programming voltage |
| V_end | 3.0V | Maximum programming voltage |
| ΔV | 50mV | Voltage step per level |
| Pulse width | 100ns | Fixed duration |
| Levels | 30-32 | Usable distinct states |

**Physics Explanation:**
- Each voltage increment switches a specific grain population
- Overcomes varying coercive fields in polycrystalline HZO
- Prevents screening effects that cause Scheme A failure
- Domain nucleation occurs at E_c ~ 1 MV/cm per grain

### 2. Enhanced IR Drop Model (Priority: HIGH)

**Reference:** Crossbar_Sneak_Path_Analysis_arXiv.pdf (Variability-aware Crossbars Tutorial)

**Improvement:** Model wire resistance more accurately:
```go
// Enhanced IR drop model
type WireModel struct {
    R_word      float64 // Word line resistance (Ω/cell)
    R_bit       float64 // Bit line resistance (Ω/cell)
    R_via       float64 // Via resistance (Ω)
    R_access    float64 // Access transistor on-resistance
}

func (w *WireModel) EffectiveVoltage(row, col int, V_applied float64) float64 {
    // Account for cumulative IR drop
    IR_drop := float64(col) * w.R_word * I_avg
    return V_applied - IR_drop
}
```

### 3. Sneak Path Suppression Visualization (Priority: MEDIUM)

**Reference:** sneak_path_self_rectifying_arrays_2022.pdf

**Improvement:** Visualize selector device benefits:
- Show sneak currents with/without selectors
- Model 1T1R (1 Transistor 1 Resistor) configuration
- Compare passive vs active array architectures

### 4. Conductance Linearity Verification (Priority: MEDIUM)

**Improvement:** Add verification plot showing:
- Target level vs actual conductance
- ±3σ state separation margins
- Non-linearity quantification (INL/DNL for weights)

### 5. Temperature-Dependent Conductance (Priority: LOW)

**Reference:** Temperature_Resilient_FeFET_CIM_2024.pdf

**Improvement:** Model conductance drift with temperature:
- G(T) = G₀ × (1 + α × (T - T₀))
- Show impact on accuracy at elevated temperatures

---

## Papers Supporting This Demo

### Currently Available
| Paper | Location | Relevance |
|-------|----------|-----------|
| Crossbar_Sneak_Path_Analysis_arXiv.pdf | opensource/papers/04_CIM_Architectures/ | Variability-aware crossbar tutorial |
| Analog_CIM_Energy_Efficiency_arXiv.pdf | opensource/papers/04_CIM_Architectures/ | CIM energy analysis |
| sneak_path_self_rectifying_arrays_2022.pdf | papers/downloaded/frontiers/ | Sneak path solutions |
| multilevel_fefet_crossbar_2023.pdf | papers/downloaded/nature/ | Multi-level FeFET crossbar |
| memory_tech_crossbar_dnn_accuracy_2024.pdf | papers/downloaded/arxiv/ | Memory technology comparison |
| IR drop analysis papers | various | Wire resistance effects |

### Recommended for Download
| Paper | Source | Why Needed |
|-------|--------|------------|
| **Oh et al. IEEE EDL 2017** | IEEE Xplore | "32 levels of Conductance States" - Scheme C details |
| **FeFET crossbar design guidelines** | IEEE IEDM | Physical layout considerations |
| **1T1R FeFET array papers** | Various | Sneak path suppression architecture |

---

## Architecture

```
demo2-crossbar/
├── cmd/
│   ├── crossbar-gui/
│   │   └── main.go           # Fyne GUI entry point
│   └── inference/
│       └── main.go           # Terminal analysis entry point
├── pkg/
│   ├── crossbar/             # Array modeling
│   │   ├── array.go          # Crossbar structure & MVM
│   │   ├── cell.go           # FeFET/FTJ cell model
│   │   ├── wire.go           # Wire resistance model
│   │   └── nonidealities.go  # IR drop, sneak paths
│   ├── gui/
│   │   ├── app.go            # Main application, layout
│   │   ├── heatmap.go        # CrossbarHeatmap widget
│   │   ├── controls.go       # ControlPanel, StatsPanel
│   │   └── vectors.go        # VectorBarChart widgets
│   └── compute/              # Future: Vulkan compute
│       ├── mvm.go            # MVM kernel
│       └── nonideal.go       # Non-ideality injection
└── shaders/                  # SPIR-V shaders (future)
    ├── mvm.comp              # MVM compute shader
    ├── crossbar.vert         # Grid vertex shader
    └── crossbar.frag         # Cell color shader
```

## Go Package Dependencies

### GUI Framework

| Package | Purpose |
|---------|---------|
| [fyne-io/fyne/v2](https://github.com/fyne-io/fyne) | Cross-platform GUI toolkit |
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Terminal UI (CLI version) |

### Numerical

| Package | Purpose |
|---------|---------|
| [gonum/gonum](https://github.com/gonum/gonum) | Matrix operations, linear algebra |

---

## The Story This Demo Tells

This demo answers the question: **"How do we compute in memory?"**

1. **Ohm's Law Computing** — Current = Conductance × Voltage happens in physics
2. **Massively Parallel** — All 64×64 = 4096 multiplications happen simultaneously
3. **Energy Efficient** — No data movement between memory and processor
4. **30-Level Precision** — Each weight has ~5 bits of analog precision
5. **Real Non-Idealities** — Shows practical challenges (IR drop, sneak paths)

---

## Tests

```bash
# Run all tests
cd demo2-crossbar
go test ./...

# Run crossbar package tests
go test ./pkg/crossbar -v

# Run with verbose non-idealities tests
go test ./pkg/crossbar -v -run TestNonidealities
```

Test coverage:
- MVM correctness verification
- IR drop calculation
- Sneak path current analysis
- 30-level conductance quantization
- Non-ideality impact on accuracy

---

## Benchmarks (from Literature)

| Architecture | MNIST Accuracy | Source |
|--------------|----------------|--------|
| 24×24 FE Memristor (sim) | 98.78% | ScienceDirect 2025 |
| Multi-Level FeFET 28nm (sim) | 96.6% | Nature Comms 2023 |
| FTJ Crossbar (sim) | 92% | SemiEngineering 2024 |
| **Ferroelectric CIM Hardware** | **87%** | Dr. Tour presentation |
| **Theoretical Max (Ferroelectric CIM)** | **88%** | Dr. Tour presentation |

**Important Note:** Dr. Tour stated that Ferroelectric CIM achieves **87% accuracy on physical hardware** with a theoretical maximum of **88%** for their architecture. Our simulation (Demo 3) may achieve higher accuracy because it doesn't capture all hardware non-idealities. The 87% figure represents actual measured performance on ferroelectric crossbar arrays.

---

## Troubleshooting

### GUI fails to start

**Linux:** Install required dependencies:
```bash
# Debian/Ubuntu
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel \
    libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel
```

### Array computation is slow

For large arrays (>64×64):
- Use GPU acceleration (future Vulkan compute shader)
- Reduce visualization update frequency
- Consider chunked MVM for memory

---

## References

1. Oh et al. "HfZrOₓ FeFET Synapse with 32 levels" IEEE EDL (2017) - **CRITICAL for Scheme C**
2. Crossbar Array Tutorial, arXiv:2204.09543 - **Non-idealities modeling**
3. Dr. external research group, "Ferroelectric CIM Presentation" (Nov 2024)
4. IBM AIHWKit documentation - CIM simulation methodology

---

## License

Part of the Ferroelectric CIM Visualizer project.
