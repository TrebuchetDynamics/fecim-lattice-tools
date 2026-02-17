# FeCIM Lattice Tools - Feature Catalog (PRD)

**Document Status:** Draft v1  
**Generated:** 2026-02-16  
**Purpose:** Consolidated feature reference validated against code

---

## 1. Executive Summary

| Module | Name | Core Function | Status |
|--------|------|---------------|--------|
| M1 | Hysteresis | P-E curve simulation | ✅ Implemented |
| M2 | Crossbar | Analog MVM + non-idealities | ✅ Implemented |
| M3 | MNIST | Neural inference demo | ✅ Implemented |
| M4 | Circuits | Peripheral signal chain | ✅ Implemented |
| M5 | Comparison | Business case viz | ✅ Implemented |
| M6 | EDA | Export (JSON/SPICE/Verilog/LEF) | ✅ Implemented |
| M7 | Docs | Help system | ✅ Implemented |

---

## 2. Feature Matrix

### M1: Hysteresis (P-E Curve Simulation)

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| Preisach hysteresis | `shared/physics/preisach.go` | ✅ | Mayergoyz-style, discrete states |
| Landau-Khalatnikov | `shared/physics/landau.go` | ✅ | Dynamic switching, educational |
| ISPP Write/Verify | `module1-hysteresis/pkg/controller/writer.go` | ✅ | Level-based programming |
| Material presets | `shared/physics/material.go` | ✅ | 9 materials (HZO, FeCIM, AlScN, etc.) |
| Temperature scaling | `shared/physics/material.go` | ✅ | Ec/Pr tempco |
| Discrete level mapping | `shared/physics/material.go` | ✅ | 8-140 levels |
| P-E loop export | `shared/export/` | ✅ | JSON, CSV |
| GUI P-E plot | `module1-hysteresis/pkg/gui/` | ✅ | Real-time animation |
| Headless mode | `module1-hysteresis/pkg/controller/` | ✅ | CLI + regression |
| Polydomain LK | `shared/physics/polydomain.go` | ✅ | Ensemble model |

### M2: Crossbar (Analog MVM)

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| Kirchhoff MVM | `module2-crossbar/pkg/gui/analysis.go` | ✅ | I = G × V |
| 30-level quantization | `module3-mnist/pkg/core/quantize.go` | ✅ | Shared with M3 |
| Conductance mapping | `shared/physics/material.go` | ✅ | Linear, exp, LUT |
| IR drop analysis | `module2-crossbar/pkg/gui/ir_analysis.go` | ✅ | Line resistance |
| Sneak path analysis | `module2-crossbar/pkg/gui/sneak_paths.go` | ✅ | Parasitic currents |
| Drift modeling | `module2-crossbar/pkg/gui/drift.go` | ✅ | Time-dependent G |
| Process variation | `shared/physics/variation.go` | ✅ | Gaussian Ec/Pr |
| GPU acceleration | `shared/gpu/` | ✅ | Optional compute shader |
| Heatmap visualization | `module2-crossbar/pkg/gui/` | ✅ | Per-tab views |

### M3: MNIST (Neural Inference)

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| FP32 inference | `module3-mnist/pkg/core/inference.go` | ✅ | Baseline |
| CIM inference | `module3-mnist/pkg/core/inference.go` | ✅ | Quantized + noise |
| Drawing canvas | `module3-mnist/pkg/gui/canvas.go` | ✅ | 28×28 input |
| Weight quantization | `module3-mnist/pkg/core/quantize.go` | ✅ | Symmetric N-level |
| Read noise model | `module3-mnist/pkg/core/noise.go` | ✅ | Gaussian multiplicative |
| DAC/ADC bit depth | `module3-mnist/pkg/core/` | ✅ | 3-16 bits |
| Confusion matrix | `module3-mnist/pkg/gui/metrics.go` | ✅ | Per-class stats |
| Energy estimation | `module3-mnist/pkg/gui/energy.go` | ✅ | Model-based |

### M4: Circuits (Peripheral Signal Chain)

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| READ mode | `module4-circuits/pkg/arraysim/` | ✅ | Sensing |
| WRITE mode | `module4-circuits/pkg/arraysim/` | ✅ | ISPP programming |
| COMPUTE mode | `module4-circuits/pkg/arraysim/` | ✅ | MVM execution |
| Tier-A solver | `module4-circuits/pkg/arraysim/tier_a*.go` | ✅ | Simple Kirchhoff |
| Tier-B solver | `module4-circuits/pkg/arraysim/tier_b*.go` | ✅ | Full MNA |
| DAC model | `shared/peripherals/dac.go` | ✅ | 5-bit, ±1.5V |
| ADC model | `shared/peripherals/adc.go` | ✅ | 5-bit SAR |
| TIA model | `shared/peripherals/tia.go` | ✅ | 10kΩ transimpedance |
| Charge pump | `shared/peripherals/chargepump.go` | ✅ | Dickson 2-stage |
| Voltage zones | `module4-circuits/pkg/gui/` | ✅ | Safe read/write windows |
| Array heatmap | `module4-circuits/pkg/gui/` | ✅ | Level visualization |
| GPU batch peripherals | `module4-circuits/pkg/gpuperiph/` | ✅ | Experimental |

### M5: Comparison (Business Case)

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| CPU+DRAM model | `module5-comparison/pkg/comparison/architecture.go` | ✅ | 5nm baseline |
| GPU+HBM model | `module5-comparison/pkg/comparison/architecture.go` | ✅ | 4nm baseline |
| FeCIM model | `module5-comparison/pkg/comparison/architecture.go` | ✅ | 45nm estimated |
| Workload library | `module5-comparison/pkg/comparison/workloads.go` | ✅ | MNIST, ResNet, BERT, GPT |
| Energy race viz | `module5-comparison/pkg/gui/` | ✅ | Animated comparison |
| ROI summary | `module5-comparison/pkg/gui/` | ✅ | Cost/power projections |

### M6: EDA (Export Tools)

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| JSON export | `module6-eda/pkg/export/json.go` | ✅ | Array config |
| CSV export | `module6-eda/pkg/export/csv.go` | ✅ | Cell list |
| SPICE netlist | `module6-eda/pkg/export/spice.go` | ✅ | Subcircuit |
| Verilog | `module6-eda/pkg/export/verilog.go` | ✅ | Module definition |
| LEF export | `module6-eda/pkg/export/lef.go` | ✅ | Cell layout |
| Liberty timing | `module6-eda/pkg/export/liberty.go` | ✅ | Placeholder values |
| DEF export | `module6-eda/pkg/export/def.go` | ✅ | Placement |
| PDK presets | `module6-eda/cells/` | ✅ | SKY130, GF180MCU |
| OpenLane helpers | `module6-eda/pkg/gui/openlane.go` | ✅ | Config generation |
| Yosys validation | `module6-eda/pkg/gui/validation.go` | ✅ | Synthesis check |

### M7: Documentation

| Feature | Code Path | Status | Notes |
|---------|-----------|--------|-------|
| Help system | `module7-docs/pkg/gui/` | ✅ | Fyne-based docs |
| Cross-module nav | `module7-docs/pkg/gui/` | ✅ | Tab help |

---

## 3. Shared Components

| Component | Path | Used By |
|-----------|------|----------|
| Material definitions | `shared/physics/material.go` | M1, M2, M3 |
| Preisach model | `shared/physics/preisach.go` | M1 |
| Landau-Khalatnikov | `shared/physics/landau.go` | M1 |
| ISPP engine | `shared/physics/ispp.go` | M1, M4 |
| Peripherals (DAC/ADC/TIA) | `shared/peripherals/` | M4 |
| GPU compute | `shared/gpu/` | M2, M3 |
| Export utilities | `shared/export/` | M1, M6 |
| Validation suite | `validation/` | All modules |

---

## 4. Materials Database

| Material | Levels | Source | Status |
|----------|--------|--------|--------|
| HZO (Si-doped) | 30 | Literature (Park 2015) | ✅ |
| FeCIM HZO | 30 | Demo preset | ✅ |
| Literature Superlattice | 64 | Cheema 2020 | ✅ |
| Cryogenic HZO | 30 | Demo preset | ✅ |
| HZO Standard 32 | 32 | Demo preset | ✅ |
| HZO FTJ 140 | 140 | Demo preset | ✅ |
| AlScN | 8-16 | Literature | ✅ |
| In2Se3 | TBD | Tour lab (unverified) | ⚠️ |

---

## 5. Validation Status

### Regression Suites

| Test Suite | Path | Status |
|------------|------|--------|
| M1 headless ISPP | `scripts/run_headless_ispp_regressions.sh` | ✅ Running |
| M4 headless | `scripts/run_headless_module4_regressions.sh` | ✅ Running |
| Literature validation | `scripts/run_literature_validation.sh` | ✅ Running |
| Full validation | `scripts/run_full_validation.sh` | ✅ Running |

### Known Issues

| Issue | Severity | Status |
|-------|----------|--------|
| Module 2 vet error (liveslide_logic_test) | Low | Known pre-existing |
| VK-1/2/3 Vulkan rendering | Deferred | ⏳ Low priority |
| DOCA-01/11/12 | Closed | ✅ |

---

## 6. Code ↔ Feature Validation

Each feature in Section 2 has been validated against actual code paths. Features marked ✅ have corresponding implementation files.

---

## 7. Gaps & Next Actions

Based on code audit, priority gaps:

| Gap | Priority | Action |
|-----|----------|--------|
| In2Se3 material parameters incomplete | Medium | Add DOI-backed params |
| Liberty timing placeholders | Low | Needs SPICE char |
| Vulkan rendering | Deferred | 16-24hr estimate |

---

## 8. Feature ↔ TODO Validation

Cross-reference with active TODO items:

| Feature | TODO ID | Status |
|---------|---------|--------|
| ISPP Write/Verify | G04, G04b, G04c | ✅ Done |
| Polydomain LK | LK-PD-1 through LK-PD-6 | ✅ Done |
| LK mid-target optimization | LK05, LK07 | ✅ Done |
| Tier-B MNA solver | RG-PHY-OBS (physics) | ⏳ In Progress |
| WRITE boundary integrity | M4-WRITE-P0 | ✅ Done |
| Monte Carlo uncertainty | RG-VAL-M1-04 | ⏳ In Progress |
| Literature P-E validation | RG-PHY-OBS-01 | ⏳ In Progress |
| Golden P-E regression | RG-VAL-M1-02 | ⏳ In Progress |

---

## 9. Documentation Alignment

| Doc File | Features Covered | Gaps |
|----------|-----------------|------|
| `module1-hysteresis/FEATURES.md` | M1 all | Complete |
| `module2-crossbar/FEATURES.md` | M2 all | Complete |
| `module3-mnist/FEATURES.md` | M3 all | Complete |
| `module4-circuits/FEATURES.md` | M4 all | Complete |
| `module5-comparison/FEATURES.md` | M5 all | Complete |
| `module6-eda/FEATURES.md` | M6 all | Complete |
| `docs/FEATURE_CATALOG.md` | Cross-module | This doc |

---

**Document validated against code:** 2026-02-16  
**Next review:** After TODO item completion