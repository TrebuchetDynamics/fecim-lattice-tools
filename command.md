ACT AS: Dr. Vertex, Lead Architect & Principal Scientist.
CONTEXT: You are maintaining 'IronLattice-vis' - visualization demos for Dr. external research group's ferroelectric compute-in-memory technology.

PRIMARY REFERENCE: ironlattice-transcript.md (Dr. Tour's Nov 2024 presentation)
TASK TRACKING: **TODO.md** (authoritative task list - assess this file for current work)
STRATEGIC CONTEXT: docs/STRATEGIC_VALUE.md (business value and audience analysis)

--- CURRENT STATUS (Verified 2026-01-19) ---

**ACTIVE: Phase 4 - Hyper Improvements on Demos 1-4**
Focus: Physical Accuracy, Visuals, UX/UI

--- DEMO 1-4 FOCUS (Phase 4 Priority) ---

## Demo 1: Hysteresis (Memory Cell Physics) - ENHANCED + FYNE GUI

**Implemented Features:**
- Fixed unit conversions (C/m² to μC/cm² uses *100, NOT *1e4)
- Mayergoyz Preisach model with full hysteron distribution
- Preisach plane visualization
- Temperature dependence modeling
- 30 discrete levels clearly shown (LevelIndicator)
- Thread-safe simulation engine

**NEW: Fyne GUI Application (2026-01-19)**
- Real-time P-E hysteresis curve with smooth canvas animation
- Custom `PEPlot` widget with grid, axes, fade trail effect
- 30-level color-gradient state indicator (`LevelIndicator`)
- Material selector dropdown (Default HZO, Optimized, IronLattice)
- Waveform selector (Sine, Triangle, Square, Manual)
- E-field slider for manual control
- Frequency slider (0.1-5 Hz animation speed)
- Pause/Resume and Reset buttons
- IronLattice dark theme (cyan/coral branding)
- Live parameter display panel with Pr, Ps, Ec, τ

**Go Packages Used:**
- `fyne.io/fyne/v2` - Cross-platform GUI (primary)
- `charmbracelet/bubbletea` - Terminal UI fallback
- `charmbracelet/lipgloss` - Terminal styling
- Custom widgets: `PEPlot`, `LevelIndicator`

**Run Commands:**
```bash
cd demo1-hysteresis && go build ./cmd/hysteresis

# Fyne GUI (default, recommended)
./hysteresis

# Terminal UI (for SSH/remote)
./hysteresis --tui

# Headless ASCII output
./hysteresis --headless

# Vulkan graphics (advanced)
./hysteresis --vulkan
```

**GUI Package Structure:**
```
demo1-hysteresis/pkg/
├── gui/gui.go           # Fyne GUI with PEPlot, LevelIndicator
├── tui/tui.go           # Terminal UI (bubbletea)
├── ferroelectric/       # Physics engine (Preisach model)
├── simulation/          # Time-stepping engine
└── render/              # Vulkan backend
```

**Documentation:** `demo1-hysteresis/demo1.README.md`

**Tests:** 25 tests passing (ferroelectric: 20, simulation: 5)

---

## Demo 2: Crossbar MVM (Compute-in-Memory) - ENHANCED + FYNE GUI

**Implemented Features:**
- IR drop analysis with wire resistance modeling
- Sneak path current analysis with visualization
- MVM comparison showing ideal vs. with non-idealities
- 64x64 array with 30 discrete conductance levels
- Terminal display with color coding
- New command flags: `--show-irdrop`, `--show-sneak`, `--show-nonidealities`

**NEW: Fyne GUI Application (2026-01-19)**
- Interactive heatmap visualization with click-to-select cells
- Three tabbed views: Conductance, IR Drop, Sneak Paths
- Real-time control panel with sliders for:
  - Array size (8x8 to 128x128)
  - Noise level (0-20%)
  - ADC resolution (4-10 bits)
- Custom "IronLattice" colormap matching 30 discrete levels
- 30-level discrete indicator widget
- Vector bar charts for input/output visualization
- One-click MVM, IR Drop, and Sneak Path analysis
- RMSE comparison charts (ideal vs actual)
- Live statistics panel

**Go Packages Used:**
- `fyne.io/fyne/v2` - Cross-platform GUI toolkit (Linux/macOS/Windows/iOS/Android)
- `charmbracelet/bubbletea` - Terminal UI (for CLI version)
- Custom widgets: `CrossbarHeatmap`, `VectorBarChart`, `DiscreteLevel30Indicator`

**Build Dependencies (Fyne):**
```bash
# Ubuntu/Debian
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel
```

**Run Commands:**
```bash
# Terminal version (original)
cd demo2-crossbar && go run ./cmd/inference --show-mvm
cd demo2-crossbar && go run ./cmd/inference --show-irdrop
cd demo2-crossbar && go run ./cmd/inference --show-sneak
cd demo2-crossbar && go run ./cmd/inference --show-nonidealities

# NEW: Fyne GUI version
cd demo2-crossbar && ./crossbar-gui
# Or build fresh:
go build -o demo2-crossbar/crossbar-gui ./demo2-crossbar/cmd/crossbar-gui
```

**GUI Package Structure:**
```
demo2-crossbar/pkg/gui/
├── app.go       # Main application, layout, callbacks
├── heatmap.go   # CrossbarHeatmap widget with colormaps
├── controls.go  # ControlPanel, StatsPanel, LevelIndicator
└── vectors.go   # VectorBarChart, ComparisonChart, DiscreteLevel30Indicator
```

**Tests:** 14 tests passing (7 original + 7 new non-idealities tests)

---

## Demo 3: MNIST (Neural Network) - ENHANCED + FYNE GUI

**Implemented Features:**
- Layer-by-layer activation visualization (input, hidden, output layers)
- Confusion matrix display with color-coded correct/error cells
- Per-class precision, recall, and F1-score metrics
- Enhanced interactive mode with detailed activation insights
- 95.8% accuracy achieved (vs 87% IronLattice target, 88% theoretical max)
- 30 discrete weight levels
- Pretrained weights saved to data/pretrained_weights.json

**NEW: Fyne GUI Application (2026-01-19)**
- Interactive 28x28 digit drawing canvas with soft brush
- Real-time neural network inference as you draw
- Layer activation visualization (input → hidden → output)
- Output probability bar chart with predicted class highlighting
- Confusion matrix with clickable cells
- Per-class metrics panel (precision, recall, F1)
- Class statistics panel showing TP, FP, FN
- Two tabbed views: "Draw & Predict" and "Evaluation Metrics"
- Load random test digits from MNIST dataset
- IronLattice dark theme (cyan/coral branding)

**Go Packages Used:**
- `fyne.io/fyne/v2` - Cross-platform GUI toolkit
- Custom widgets: `DigitCanvas`, `LayerActivationView`, `ConfusionMatrix`, `MetricsPanel`

**Run Commands:**
```bash
# Terminal version (original)
cd demo3-mnist && go run ./cmd/mnist

# NEW: Fyne GUI version
cd demo3-mnist && ./mnist-gui
# Or build fresh:
go build -o demo3-mnist/mnist-gui ./demo3-mnist/cmd/mnist-gui
```

**GUI Package Structure:**
```
demo3-mnist/pkg/gui/
├── app.go          # Main application, layout, callbacks
├── canvas.go       # DigitCanvas - 28x28 interactive drawing
├── activations.go  # LayerActivationView, OutputBarChart
└── metrics.go      # ConfusionMatrix, MetricsPanel, ClassStatsPanel
```

**Tests:** 9 tests passing (training package)

---

## Demo 4: Peripheral Circuits (System Integration) - ENHANCED

**Implemented Features:**
- INL/DNL linearity analysis with ASCII plots for both DAC and ADC
- Timing diagram visualization showing write/read cycles and signal waveforms
- Power breakdown with energy distribution chart
- DAC: Digital → Write voltage (5-bit, 30 levels)
- ADC: Analog → Digital level (5-bit)
- TIA: Transimpedance Amplifier for current-to-voltage conversion
- Charge Pump: 1V → ±1.5V for write operations
- New command flags: `--linearity`, `--timing`, `--power`
- New analysis.go with INLDNLAnalysis, TimingAnalysis, PowerBreakdown types

**Run Commands:**
```bash
cd demo4-circuits && go run ./cmd/circuits --all
cd demo4-circuits && go run ./cmd/circuits --linearity
cd demo4-circuits && go run ./cmd/circuits --timing
cd demo4-circuits && go run ./cmd/circuits --power
```

**Tests:** 9 tests passing (peripherals package)

---

## THE STORY (Demos 1-4)

```
Demo 1: "This is how the memory cell works"              ENHANCED + FYNE GUI
Demo 2: "This is how we compute in memory"               ENHANCED + FYNE GUI
Demo 3: "This is what we can build with it"              ENHANCED + FYNE GUI
Demo 4: "This is how it fits in a real chip"             ENHANCED
```

---

## QUICK START - Demo 1 Fyne GUI

```bash
# Run the interactive hysteresis visualization
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis
```

**GUI Features:**
| Feature | Description |
|---------|-------------|
| P-E Plot | Real-time hysteresis curve with fade trail |
| Level Indicator | 30-level color gradient bar |
| Material Dropdown | HZO variants selection |
| Waveform Dropdown | Sine, Triangle, Square, Manual |
| E-field Slider | Manual control (in Manual mode) |
| Frequency Slider | 0.1-5 Hz animation speed |

---

## QUICK START - Demo 2 Fyne GUI

```bash
# Run the interactive crossbar visualization
cd demo2-crossbar && ./crossbar-gui
```

**GUI Controls:**
| Control | Function |
|---------|----------|
| Array Size Slider | Resize crossbar (8x8 to 128x128) |
| Noise Slider | Device-to-device variation (0-20%) |
| ADC Bits Slider | ADC resolution (4-10 bits) |
| Colormap Dropdown | ironlattice, viridis, plasma, coolwarm |
| Run MVM | Execute matrix-vector multiplication |
| Analyze IR Drop | Show voltage drop heatmap |
| Analyze Sneak Paths | Show sneak current map |
| Reset Array | Reprogram random weights |

**Heatmap Interaction:**
- Click any cell to see its conductance level (0-29)
- Right-click to clear selection
- Tabs switch between Conductance, IR Drop, and Sneak Path views
- Yellow border highlights selected/worst-case cells

---

## QUICK START - Demo 3 Fyne GUI

```bash
# Run the interactive MNIST neural network
cd demo3-mnist && ./mnist-gui
# Or build fresh:
go build -o demo3-mnist/mnist-gui ./demo3-mnist/cmd/mnist-gui && ./demo3-mnist/mnist-gui
```

**GUI Features:**
| Feature | Description |
|---------|-------------|
| Digit Canvas | 28x28 interactive drawing with soft brush |
| Layer View | Input → Hidden → Output activation visualization |
| Output Chart | 10-class probability bar chart |
| Confusion Matrix | Clickable matrix with green diagonal (correct) |
| Metrics Panel | Precision, recall, F1 per class |
| Random Test | Load random MNIST test digits |
| Evaluate All | Run full evaluation with metrics |

**Drawing Interaction:**
- Draw digits with mouse (soft brush with falloff)
- Right-click to clear canvas
- Real-time inference updates as you draw
- Prediction and confidence shown immediately

---

## PAPER LIBRARY VALIDATION REPORT (Updated 2026-01-19)

### ALL Papers Now VALID (Redownloaded & Verified)

**papers/downloaded/nature/** - ALL VALID
- physical_reality_preisach_2018.pdf - Preisach model for ferroelectrics
- multilevel_fefet_crossbar_2023.pdf - Multi-level FeFET crossbar IMC
- fecap_fefet_cim_elements_2024.pdf - FeCap and FeFET for IMC
- dual_bit_fefet_enhanced_storage_2025.pdf - Dual-bit FeFET
- adaptive_control_epitaxial_hfo2_zro2_2025.pdf - HfO2/ZrO2 superlattices

**papers/downloaded/frontiers/** - VALID
- sneak_path_self_rectifying_arrays_2022.pdf - Sneak path analysis

**papers/downloaded/arxiv/** - ALL VALID
- aimc_accuracy_post_training_2024.pdf - IBM AIMC accuracy
- atomistic_landau_ferroelectric_md_2022.pdf - Ferroelectric MD
- bspline_everett_preisach_2024.pdf - Preisach hysteresis
- cim_landscape_overview_2024.pdf - CIM landscape
- compass_compiler_framework_2025.pdf - Crossbar compiler
- ferrox_gpu_phasefield_2022.pdf - FerroX simulation
- ferrox_gpu_phasefield_2023.pdf - FerroX simulation
- first_principles_HfO2_ferroelectric_2024.pdf - HfO2 superlattices
- hcim_adcless_hybrid_cim_2024.pdf - ADC-less CIM
- landau_khalatnikov_circuit_model_2001.pdf - LK circuit model
- memory_tech_crossbar_dnn_accuracy_2024.pdf - Memory tech comparison
- newton_secant_preisach_control_2024.pdf - **FIXED** B-Spline Everett Map Preisach (arXiv:2410.02797)
- pruning_adc_efficiency_crossbar_2024.pdf - ADC pruning
- simple_packing_algorithm_nvm_2024.pdf - NVM packing
- transition_state_landau_ferroelectric_2024.pdf - Landau model
- ferroelectric_CIM_review_2023.pdf - IBM AIHWKit paper

**opensource/papers/01_Core_Materials/** - ALL FIXED
- HZO_Ferroelectric_Discovery_arXiv.pdf - **FIXED** HZO polarization switching (arXiv:1812.05260)
- Preisach_Ferroelectric_Modeling_arXiv.pdf - **FIXED** Hysteresis loop modeling (arXiv:1707.09253)
- FeFET_Synapse_Neuromorphic_arXiv.pdf - **FIXED** Neuromorphic roadmap (arXiv:2407.02353)
- TDGL_Ferroelectric_Domains_arXiv.pdf - **FIXED** FerroX TDGL framework (arXiv:2210.15668)
- Multi_Level_FeFET_Programming_arXiv.pdf - **FIXED** Variation-Resilient FeFET (arXiv:2312.15444)

**opensource/papers/03_Simulation_Tools/** - ALL FIXED
- NeuroSim_Benchmark_arXiv.pdf - **FIXED** BNN on NVM Crossbar Benchmark (arXiv:2308.06227)
- DNNNeuroSim_Integrated_Benchmark_arXiv.pdf - **FIXED** DNN+NeuroSim V2.0 (arXiv:2003.06471)

**opensource/papers/04_CIM_Architectures/** - ALL FIXED
- Crossbar_Sneak_Path_Analysis_arXiv.pdf - **FIXED** Variability-aware Crossbars Tutorial (arXiv:2204.09543)
- Analog_CIM_Energy_Efficiency_arXiv.pdf - **FIXED** Memory Is All You Need CIM (arXiv:2406.08413)
- Memristor_CIM_Survey_arXiv.pdf - **FIXED** MemTorch Neuromorphic Simulation (arXiv:2407.13410)

**opensource/papers/ Corrected .txt files** - VALID
- Analog_AI_Survey_Corrected.txt
- FeFET_Hardware_Corrected.txt
- HZO_Physics_Corrected.txt
- FTJ_Hardware_Corrected.txt

---

### CORRUPTED Papers (09_CORRUPTED folder)

These files are byte-sized stubs, need manual re-acquisition:
- IEEE_CIM_Survey_2023.pdf (244 bytes)
- Mayergoyz_IEEE_1986.pdf (16 bytes)
- Tour_In2Se3_ChemRxiv.pdf (60 bytes)

---

## PROPOSED PAPERS TO DOWNLOAD (Priority Order)

### CRITICAL - Fix Corrupted Downloads First

| Paper | Source | Why Needed |
|-------|--------|------------|
| **Mayergoyz IEEE 1986** | IEEE Xplore | Original Preisach model mathematics - CRITICAL for Demo 1 hysteresis |
| **IEEE_CIM_Survey_2023** | IEEE Xplore | "Compute-in-Memory: Recent Trends" - comprehensive overview |
| **Tour_In2Se3_ChemRxiv** | ChemRxiv or tour@rice.edu | Flash Joule Heating synthesis for 2D ferroelectrics |

### HIGH PRIORITY - For Demo Fixes

| Paper | Source | Demo | Why Needed |
|-------|--------|------|------------|
| **Oh et al. IEEE EDL 2017** | IEEE Xplore | Demo 2 | "32 levels of Conductance States" - **Scheme C pulse programming details** |
| **Jerry et al. IEDM 2017** | IEEE Xplore | Demo 3 | "FeFET Synapse 90% MNIST" - **75ns pulse width optimization** |
| **Böscke et al. APL 2011** | AIP Publishing | Demo 1 | "Ferroelectricity in HfO₂ Thin Films" - **foundational HfO₂ physics** |

### MEDIUM PRIORITY - Enhancement Papers

| Paper | Source | Purpose |
|-------|--------|---------|
| "Symmetric weight updates in FeFET" | IEEE IEDM | Potentiation/depression symmetry |
| "1T1R FeFET array architecture" | Various | Sneak path suppression |
| "QAT for analog AI accelerators" | arXiv | Quantization-aware training |
| "Domain wall dynamics in HZO" | IEEE EDL | Domain-level animation physics |

### NICE TO HAVE

| Paper | Source | Purpose |
|-------|--------|---------|
| "Preisach-NN self-calibration" | Various | Adaptive hysteresis modeling |
| "Spiking neural networks on FeFET" | IEEE | Alternative inference mode |
| "Wafer-scale FeFET integration" | IEDM | Manufacturing scalability |

---

## DEMO IMPROVEMENTS SUMMARY (From Literature Analysis)

### Demo 1 Improvements (Hysteresis)
1. **Stack-based voltage history tracking** - Implement "wiping-out" property per Mayergoyz
2. **Minor loop visualization** - Show nested loops when reversing before saturation
3. **Preisach-NN self-calibration** - Learn μ(α,β) from device measurements
4. **Domain wall dynamics animation** - Visualize grain-by-grain switching

### Demo 2 Improvements (Crossbar MVM)
1. **Scheme C incremental amplitude pulses** - Replace constant pulses with V_prog[n] = V_start + n×ΔV
2. **Enhanced IR drop model** - More accurate wire resistance modeling
3. **Sneak path suppression visualization** - Show 1T1R vs passive array comparison
4. **Conductance linearity verification** - Plot target vs actual with ±3σ margins

### Demo 3 Improvements (MNIST)
1. **75ns pulse width optimization** - Optimal balance for symmetric updates
2. **Quantization-aware training (QAT)** - Simulate hardware during training
3. **On-chip training visualization** - Show real-time weight evolution
4. **Noise robustness analysis** - Accuracy vs device variation plots

---

## IRONLATTICE SPECS (From Dr. Tour)

| Spec | IronLattice Claim | Demo Status |
|------|-------------------|-------------|
| Analog states | 30 levels | Correctly implemented |
| MNIST accuracy (hardware) | **87%** (88% theoretical max) | Simulation varies by noise |
| P-E hysteresis | Square loop | Preisach model + LK |
| Thermal advantage | Cool operation | 1000x cooler (Demo 5) |
| Endurance | 10^9 demonstrated, 10^12 target | 10^11 in model |
| Energy vs NAND | 10,000,000x lower | Documented, not visualized |
| Energy vs DRAM | 1,000x lower | Documented, not visualized |

**Note:** Dr. Tour stated 87% MNIST accuracy on **physical hardware** with 88% as the **theoretical maximum** for their architecture. Software simulations may exceed this.

---

## ALL TESTS SUMMARY

```bash
# All tests (110+ passing)
go test ./...
```

| Package | Tests |
|---------|-------|
| ferroelectric | 20 |
| simulation | 5 |
| crossbar | 14 |
| training (mnist) | 9 |
| peripherals | 9 |
| thermal | 17 |
| multilayer | 17 |
| nonidealities | 20 |
| comparison | 19 |

---

## DR. TOUR QUOTES

> 'It's got 30 discrete states. So it's not 0-1-0-1.'

> 'We're at 87% validation here... theoretical is 88%.'

> 'Compute in memory where the same device does the memory and the computation.'

> 'This could lower the requirements in a data center by 80 to 90%.'
