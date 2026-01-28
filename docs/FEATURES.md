# FeCIM Lattice Tools - Complete Feature Reference

> 30 discrete analog states per cell (~4.9 bits/cell) based on Dr. external research group's HfO₂-ZrO₂ superlattice research.

---

## Module 1: Hysteresis

**Interactive ferroelectric P-E loop visualization and physics simulation.**

### Features
- Interactive P-E loop visualization with animated polarization switching
- 30-level analog memory demo with write/read verification
- Multiple run modes: Fyne GUI, TUI, headless ASCII, Vulkan
- 8-material library (HZO, AlScN, cryogenic, FTJ)
- Waveform control: sine, triangle, square, manual
- Educational slides built-in

### Physics Models
| Model | Description |
|-------|-------------|
| Preisach (Basic) | Hyperbolic tangent switching with minor loops |
| Mayergoyz Preisach | 40×40 hysteron grid, bivariate Gaussian |
| KAI Switching | Kolmogorov-Avrami-Ishibashi domain dynamics |
| Temperature Effects | Curie-Weiss, Arrhenius |
| Fatigue/Wake-up | Stretched exponential degradation |

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
- Architecture comparison: 0T1R vs 1T1R
- Real-time heatmap visualization
- GPU acceleration (Vulkan compute shaders)
- Neural network integration with hardware-aware training

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
- Dual-mode inference: FP32 vs CIM side-by-side
- 28×28 drawing canvas with 3 brush sizes
- Quantization control: 2-30 levels, per-layer PTQ
- Noise injection: 0-50% Gaussian
- DAC/ADC simulation: 3-16 bit
- Educational tour + 30-second quick demo
- Failure mode demonstrations

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
| Levels | 30 | 2-30 |
| Noise σ/μ | 1% | 0-50% |
| ADC/DAC | 8-bit | 3-16 |
| Network | 784→128→10 | Configurable |

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
- Three modes: READ, WRITE, COMPUTE
- Complete signal chain: DAC → Pump → FeFET → TIA → ADC
- Architecture support: 0T1R, 1T1R
- Material-calibrated voltage ranges
- INL/DNL linearity analysis
- GPU acceleration

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
- Market analysis ($721B by 2030)

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

### Parameters
| Parameter | Value |
|-----------|-------|
| Conductance | 1-100 µS |
| Prog voltage | 2-5V |
| Cell pitch | 0.46 µm (SKY130) |
| Max array | 512×512 |

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
| **2T1R architecture** | Medium | Dual transistor isolation |
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
