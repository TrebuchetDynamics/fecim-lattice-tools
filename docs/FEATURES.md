# FeCIM Lattice Tools - Features

End-to-end toolchain for Ferroelectric Compute-in-Memory (FeCIM) based on Dr. external research group's HfO₂-ZrO₂ superlattice research.

**Core Concept:** Multi-level analog states per cell (8-140 levels depending on material)

---

## Module Summary

| Module | Purpose |
|--------|---------|
| 1. Hysteresis | P-E Curve Simulator |
| 2. Crossbar | MVM Simulator with Non-Idealities |
| 3. MNIST | Neural Network Demo |
| 4. Circuits | Peripheral Circuit Design |
| 5. Comparison | Business Case Analysis |
| 6. EDA | Chip Design Suite |

---

## Module 1: Hysteresis

P-E curve simulator for ferroelectric memory physics.

- 5 physics models (Preisach, Mayergoyz, KAI, Temperature, Fatigue)
- 8+ material library (HZO variants, AlScN, cryogenic, FTJ)
- Interactive P-E loop visualization
- Multi-level analog memory demo (8-140 levels per material)
- Calibration system with temperature awareness
- Multiple run modes (GUI, TUI, headless, Vulkan)

---

## Module 2: Crossbar Array

Matrix-vector multiply simulator with realistic non-idealities.

- Analog MVM computation (I = G × V)
- Non-ideality simulation (IR drop, sneak paths, drift, temperature, variation)
- Architecture comparison (0T1R, 1T1R, 2T1R)
- Real-time heatmaps with 8 colormaps
- Before/after comparison toggle
- GPU acceleration (Vulkan)

---

## Module 3: MNIST

Neural network digit recognition demonstrating FeCIM inference.

- Dual-mode inference (FP32 vs CIM side-by-side)
- 28×28 drawing canvas
- Configurable quantization and noise
- Layer activation visualization
- Educational tour and auto demo modes

---

## Module 4: Peripheral Circuits

DAC/ADC/TIA signal chain for FeCIM arrays.

- Three operation modes (READ, WRITE, COMPUTE)
- Complete signal chain (DAC → Pump → FeFET → TIA → ADC)
- Architecture support (0T1R, 1T1R, 2T1R)
- Voltage zone visualization
- INL/DNL analysis

---

## Module 5: Technology Comparison

Business case analysis for FeCIM vs CPU/GPU.

- Architecture comparison (CPU+DRAM, GPU+HBM, FeCIM)
- 5 workloads (MNIST to LLM-70B)
- Data center calculator (power, cost, CO2)
- Animated visualizations
- 4 presentation modes (Manual, Auto, Investor, Engineer)

---

## Module 6: EDA Tools

Chip design suite for FeCIM arrays.

- 8 export formats (Verilog, SPICE, DEF, LEF, Liberty, SVG, JSON, CSV)
- 3 PDK support (SKY130, GF180MCU, IHP_SG13G2)
- OpenLane RTL-to-GDSII integration
- 3 operation modes (Storage, Memory, Compute)
- Validation tools (Yosys, DEF checker)

---

## Shared Infrastructure

- Unified FeCIM theme
- 1755+ tests
- Centralized logging
- Physics library with material database

---

## Cross-Module Workflow

```
Module 1 (Calibrate) → Module 2 (Simulate) → Module 3 (Train)
                                ↓
                        Module 4 (Design Circuits)
                                ↓
                        Module 5 (Business Case)
                                ↓
                        Module 6 (Export for Fab)
```

---

## See Also

- [Module 1 FEATURES.md](../module1-hysteresis/FEATURES.md)
- [Module 2 FEATURES.md](../module2-crossbar/FEATURES.md)
- [Module 3 FEATURES.md](../module3-mnist/FEATURES.md)
- [Module 4 FEATURES.md](../module4-circuits/FEATURES.md)
- [Module 5 FEATURES.md](../module5-comparison/FEATURES.md)
- [Module 6 FEATURES.md](../module6-eda/FEATURES.md)
