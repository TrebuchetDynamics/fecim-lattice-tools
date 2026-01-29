# FeCIM Lattice Tools - Complete Feature Reference

> 30 discrete analog states per cell (~4.9 bits/cell) based on Dr. external research group's HfO₂-ZrO₂ superlattice research.

---

## Module 1: Hysteresis

**Interactive ferroelectric P-E loop visualization and physics simulation.**

### Features
- Interactive P-E loop visualization with animated polarization switching
- 30-level analog memory demo with write/read verification
- Multiple run modes: Fyne GUI, TUI, headless ASCII
- 8-material library with Material Picker dialog (HZO, AlScN, cryogenic, FTJ variants)
- Waveform control: sine, triangle, square, manual, time-resolved switching
- Educational slides built-in with context-specific explanations

### GUI Components
| Component | Description |
|-----------|-------------|
| P-E Hysteresis Plot | Real-time polarization vs field with temperature correction |
| Cell Visualizer | Visual representation of ferroelectric state |
| Level Indicator | Visual gauge for 30 discrete levels (1-30), click-to-level in Manual mode |
| Phase Indicator | State machine: RESET → SETTLE → WRITE → READ → VERIFY |
| Material Picker | Dialog with searchable list, property tables, and detail panel |
| Information Panel | Polarization, level, state, wake-up/fatigue, temperature metrics |

### Calibration System
- Temperature-aware multi-level calibration (233-423 K)
- Key temperatures: -40°C, 0°C, 27°C (room), 100°C, 150°C (automotive)
- Binary search with oscillation detection and relaxation compensation
- CLI calibration tool (`--calibrate`, `--list-materials`, `--material`, `--force`, `--verify`)
- Persistent JSON storage in `data/calibrations/`

### Export
- JSON export with metadata (material, temperature, parameters)
- CSV export for data analysis
- Debug log export for Write/Read Demo (cycle-by-cycle data, energy tracking)

### Keyboard Shortcuts
- Material switching (Q/W keys)
- Waveform selection (1-5 number keys)
- Export/calibration shortcuts (Ctrl+E, Ctrl+L)
- Level adjustment with arrow keys in Manual mode

### Physics Models
| Model | Description |
|-------|-------------|
| Preisach (Basic) | Hyperbolic tangent switching with minor loops |
| Mayergoyz Preisach | 40×40 hysteron grid, bivariate Gaussian |
| KAI Switching | Kolmogorov-Avrami-Ishibashi domain dynamics |
| Temperature Effects | Curie-Weiss, Arrhenius |
| Fatigue/Wake-up | Stretched exponential degradation |
| Substrate Strain | Strain effects on Pr and Ec |

### Parameters
| Parameter | Value |
|-----------|-------|
| Levels | 30 (4.91 bits/cell) |
| Pr (RT) | 15-34 µC/cm² |
| Pr (4K) | 75 µC/cm² |
| Ec | 0.6-5.0 MV/cm |
| Switching τ | 1-20 ns |
| Endurance | 10⁸-10¹² cycles |

---

## Module 2: Crossbar

**Analog compute-in-memory array simulation with non-idealities.**

### Features
- Analog matrix-vector multiply (MVM): I = G × V
- 30-level quantization [0/29 ... 29/29]
- Non-ideality simulation: IR drop, sneak paths, drift, temperature
- Architecture comparison: 0T1R vs 1T1R vs 2T1R
- Real-time heatmap visualization with multiple colormaps
- Neural network integration with hardware-aware training

### GUI Components
| Component | Description |
|-----------|-------------|
| Conductance Heatmap | Real-time cell conductance (0-30 normalized), FeCIM colormap |
| IR Drop Heatmap | Percentage voltage drop (0-100%), Viridis colormap |
| Sneak Path Heatmap | Sneak current ratio (0-200%), Plasma colormap |
| MVM Visualization | Input vector, weight matrix, output computation |
| Before/After Toggle | Side-by-side non-ideality comparison |
| Accuracy Waterfall | Shows accuracy degradation from non-idealities |
| Metrics Panel | Energy per MAC, inference latency |

### Array Configuration
- Array size: 4×4 to 128×128
- Noise level: 0-5% standard deviation
- DAC/ADC bit resolution adjusters
- Colormap selector per tab

### Tabbed Interface
- **Conductance Tab** - Main weight matrix view
- **IR Drop Tab** - Voltage drop effects
- **Sneak Paths Tab** - Current leakage visualization
- **MVM Vectors Tab** - Input/output visualization
- **Analysis Tab** - Detailed metrics and comparisons

### Physics Models
| Non-Ideality | Model | Impact |
|--------------|-------|--------|
| IR Drop | WL/BL resistance 2.5Ω/cell | 10-20% voltage loss |
| Sneak Paths | 3-cell parasitic loops | 5-20% (0T1R), ~0.001% (1T1R) |
| Drift | Power-law + log + Arrhenius | <0.5 level/10yr |
| Temperature | Arrhenius (4K-500K) | Cryo: 1.5× window |
| Variation | Gaussian + spatial | 2% D2D |

### Parameters
| Parameter | Value |
|-----------|-------|
| Gmin | 10 µS |
| Gmax | 100 µS |
| R_wire | 2.5 Ω/cell (45nm) |
| Drift coeff | 0.0005-0.001 |

---

## Module 3: MNIST

**Neural network inference demonstrating FeCIM analog compute.**

### Features
- Dual-mode inference: FP32 vs CIM side-by-side comparison
- 28×28 drawing canvas with stroke rendering
- Quantization control: 8, 14, 30 levels
- Noise injection: 0-5% Gaussian
- DAC/ADC simulation: configurable bit resolution
- Batch testing with MNIST test set
- Educational tour and auto demo modes

### GUI Components
| Component | Description |
|-----------|-------------|
| Digit Canvas | 28×28 pixel drawing surface with preprocessing |
| Dual Prediction Panel | FP and CIM predictions side-by-side |
| Dual Probability Chart | Per-digit confidence (0-9) overlaid comparison |
| Hoverable Weight Heatmap | Layer 1 (784→128) or Layer 2 (128→10) visualization |
| Energy Widget | Per-inference energy comparison (25-100× typical) |
| Layer Activation View | Hidden layer activation heatmap |

### Test Modes
- **Quick Test** - Single inference with results display
- **Batch Test** - Full MNIST test set accuracy calculation
- **Auto Demo** - Continuous inference loop with metric accumulation

### Physics Models
| Model | Description |
|-------|-------------|
| Weight Quantization | Symmetric linear to N levels |
| Read Noise | Gaussian multiplicative (Johnson) |
| DAC/ADC | Voltage/current quantization |
| Energy | 10 fJ/bit per MAC |

### Parameters
| Parameter | Default | Range |
|-----------|---------|-------|
| Levels | 30 | 8, 14, 30 |
| Noise σ/μ | 1% | 0-5% |
| ADC/DAC | 8-bit | Configurable |
| Network | 784→128→10 | Fixed |

### Accuracy
| Config | Accuracy |
|--------|----------|
| FP32 ideal | ~98% |
| 30 levels, low noise | 92-96% |
| Peer-reviewed | 96.6-98.24% |

---

## Module 4: Circuits

**Peripheral circuit simulation: DAC, ADC, TIA, charge pump.**

### Features
- Three modes: READ (default), WRITE, COMPUTE
- Complete signal chain: DAC → Pump → FeFET → TIA → ADC
- Architecture support: 0T1R, 1T1R
- Material-calibrated voltage ranges via Material Picker
- INL/DNL linearity analysis

### GUI Components
| Component | Description |
|-----------|-------------|
| Operation Mode Selector | Write/Read/Compute mode toggle |
| Write Panel | Row/col selectors, level slider, write pulse waveform |
| Read Panel | Safe read voltage, conductance readout, level extraction |
| Compute Panel | Input vector entry, mini heatmap, per-row results |
| Compute Log | Detailed trace of MVM operations |
| Signal Chain Header | Visual data path display |

### Architecture Modes
| Mode | Description |
|------|-------------|
| 1T1R (Active) | Transistor-based row selection, reduced sneak paths |
| 0T1R (Passive) | No transistors, all rows always conductive |

### Tabbed Interface
- **Device Tab** - Main operation interface
- **Comparison Tab** - Side-by-side 1T1R vs 0T1R specs
- **Timing Tab** - Write/read/compute timing diagrams
- **Specifications Tab** - Array config, DAC/ADC settings, TIA parameters

### Circuit Specs
| Circuit | Specs |
|---------|-------|
| DAC | 5-bit, ±1.5V, 10ns, 15 fJ |
| ADC | 5-bit SAR, 0-1V, 50ns, 25 fJ |
| TIA | 10 kΩ, 100 MHz, 1 pA/√Hz |
| Charge Pump | 2-stage Dickson, 70% eff |

### Timing & Energy
| Metric | Value |
|--------|-------|
| Write cycle | ~170 ns |
| Read cycle | ~65 ns |
| Total energy | ~50 fJ/op |
| Throughput | ~4.7 GOPS |

---

## Module 5: Comparison

**Technology comparison: CPU vs GPU vs FeCIM.**

### Features
- Architecture comparison: CPU+DRAM, GPU+HBM, FeCIM CIM
- 5 workloads: MNIST to LLM-70B
- Data center scaling projections
- Animated visualizations
- Calculator interface for custom workloads

### GUI Components
| Component | Description |
|-----------|-------------|
| Animated Energy Race | CPU vs GPU vs FeCIM bar chart animation |
| Market Opportunity Chart | Application workload categorization |
| Competitive Matrix | Technology comparison grid with color-coded metrics |
| Phased Strategy Diagram | Technology deployment roadmap (Phase 1-4) |
| Analog States Comparison | 30 discrete levels vs 2-3 bit (NAND) |
| Data Center Transformation | GPU farm vs FeCIM equivalent visualization |

### Calculator Interface
- Workload selector: GPU, CPU, TPU, Inference
- Inference volume slider: 1K to 10M+ per second
- Dynamic calculation: energy, power, cost, infrastructure savings

### Comparison
| Metric | CPU | GPU | FeCIM (Est.) |
|--------|-----|-----|--------------|
| Energy/MAC | 1000 pJ | 100 pJ | ~1 pJ |
| TDP | 125W | 400W | 5W |
| TOPS/W | 0.008 | 0.25 | 10 |

### Workloads
| Network | MACs |
|---------|------|
| MNIST | 101K |
| ResNet-50 | 4B |
| BERT-Base | 11B |
| GPT-2 | 35B |
| LLM-70B | 140T |

---

## Module 6: EDA Tools

**RTL-to-GDSII design flow for FeCIM arrays.**

### Features
- Three modes: Storage (NAND), Memory (DRAM), Compute (AI)
- 8 export formats: JSON, CSV, SPICE, Verilog, DEF, LEF, Liberty, SVG
- Architecture: Passive, 1T1R, 2T1R
- PDK support: SKY130, GF180MCU, IHP_SG13G2
- OpenLane integration
- Weight mapping to 30 conductance levels
- Validation: Yosys, DEF checker

### Tabbed Interface
- **Builder & Validation Tab** - Array configuration and layout validation
- **Learn Tab** - Educational content on chip design and technology nodes

### Parameters
| Parameter | Value |
|-----------|-------|
| Conductance | 1-100 µS |
| Prog voltage | 2-5V |
| Cell pitch | 0.46 µm (SKY130) |
| Max array | 512×512 |

---

## Module 7: Documentation Viewer

**Interactive documentation browser with search and glossary.**

### Features
- Hierarchical document tree navigation
- Full-text search across all documentation
- Markdown rendering with syntax highlighting
- Glossary integration with clickable glossary:// links
- Table of contents auto-generated from headings
- Favorites and view history
- Persistent layout state

### GUI Components
| Component | Description |
|-----------|-------------|
| Document Tree | Hierarchical folder structure with icons |
| Search Dialog | Full-text search with results |
| Markdown Renderer | Rich text with code blocks and tables |
| Table of Contents | Auto-generated, clickable section navigation |
| Document Metadata | Title, category badges, reading time |
| Glossary Widget | Pop-up term definitions |

---

## Shared Components

**Reusable widgets and infrastructure across all modules.**

### Theme System
- FeCIM Theme with consistent color palette
- Custom background, grid, axis, positive/negative colors
- Warning color for edge cases

### Widgets
| Widget | Description |
|--------|-------------|
| Material Picker | Dialog with searchable list, property tables, detail panel |
| Material Card | Compact material property display |
| Material Detail Panel | Full property table organized by category |
| Color Legend | Scalable color bar with dual-scale support |
| Adaptive Layout | Responsive design with desktop/tablet/mobile breakpoints |
| Architecture Selector | 0T1R/1T1R/2T1R toggle widget |

### Material Picker Features
- Search filtering by name, description, reference, analog states
- 8 property categories: Polarization, Field, Dielectric, Geometry, Dynamics, Temperature, Reliability, Special
- Scientific unit formatting (µC/cm², MV/cm, nm, ns, etc.)
- Split-pane layout (35% list, 65% detail)

### Logging System
- Structured logging with file output
- Verbosity levels: off, info, debug, trace
- Per-module loggers
- Safe goroutine logging

### Recording
- FFmpeg screen recording with optional audio
- Microphone level indicator
- Recording state management

---

## Launcher

**Home screen with module navigation.**

### Features
- 6-module grid interface with cards
- Title, subtitle, description, status indicators
- "Ready" and "Work in Progress" badges
- Screenshot and video recording controls
- Home and documentation quick access buttons

---

## To Implement Later

### Module 1: Hysteresis

| Feature | Priority | Notes |
|---------|----------|-------|
| **Landau-Ginzburg-Devonshire model** | High | Thermodynamic free energy approach |
| **Phase-field simulation** | Low | Very complex, domain wall dynamics |
| **FORC analysis** | Medium | First-Order Reversal Curves for distribution |
| **Frequency dependence** | Medium | RC time constant effects |
| **Vulkan renderer** | Low | GPU-accelerated P-E visualization |

### Module 2: Crossbar

| Feature | Priority | Notes |
|---------|----------|-------|
| **Self-rectifying devices** | Low | Planned for v2.0 |
| **ISPP write-verify** | High | Incremental step pulse programming |
| **Separate R/W paths** | Medium | 2T1R read/write isolation |

### Module 3: MNIST

| Feature | Priority | Notes |
|---------|----------|-------|
| **IR drop in inference** | Medium | Integrate crossbar non-idealities |
| **Sneak path in inference** | Medium | Architecture-aware inference |
| **Larger networks** | Low | ResNet, transformer support |
| **QAT training in-tool** | Medium | Quantization-aware training GUI |

### Module 4: Circuits

| Feature | Priority | Notes |
|---------|----------|-------|
| **SPICE export** | Medium | Full signal chain netlist |
| **Timing characterization** | High | Replace placeholder Liberty values |
| **Corner analysis** | Medium | TT/FF/SS process corners |
| **Noise analysis** | Low | Full noise simulation |

### Module 5: Comparison

| Feature | Priority | Notes |
|---------|----------|-------|
| **Real benchmark data** | High | Replace estimated FeCIM values |
| **FPGA/ASIC comparison** | Medium | More architecture options |
| **TCO calculator** | Low | Detailed cost modeling |

### Module 6: EDA

| Feature | Priority | Notes |
|---------|----------|-------|
| **Magic layout (.mag)** | High | Real tape-out ready cells |
| **SPICE characterization** | High | Replace placeholder timing |
| **PDN design** | Medium | Power distribution network |
| **IO pads** | Medium | Chip-level integration |
| **Redundancy/ECC** | Low | Fault tolerance |
| **ONNX import** | Medium | Neural network weight import |

### Cross-Module

| Feature | Priority | Notes |
|---------|----------|-------|
| **Unified physics.yaml** | High | Single source of truth for all params |
| **End-to-end flow** | Medium | Train → Quantize → Map → Simulate → Export |
| **Glossary DOI completion** | Low | Replace placeholder DOIs in references |
| **TRL progression tracking** | Medium | Technology readiness visualization |

---

## See Also

- [Module 1 FEATURES.md](../module1-hysteresis/FEATURES.md)
- [Module 2 FEATURES.md](../module2-crossbar/FEATURES.md)
- [Module 3 FEATURES.md](../module3-mnist/FEATURES.md)
- [Module 4 FEATURES.md](../module4-circuits/FEATURES.md)
- [Module 5 FEATURES.md](../module5-comparison/FEATURES.md)
- [Module 6 FEATURES.md](../module6-eda/FEATURES.md)
