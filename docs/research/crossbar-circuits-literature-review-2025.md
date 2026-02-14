# Deep Research: How Crossbar & Circuits Should Be Done

**Date:** 2026-02-14
**Scope:** Literature review of FeCIM crossbar arrays and peripheral circuits, ignoring current Module 4 code.
**Purpose:** Inform future architecture decisions for Module 2 (Crossbar) and Module 4 (Circuits).

---

## Executive Summary

The literature has shifted significantly since the project's original architecture was designed. Three major trends emerge:

1. **Capacitive crossbars (FeCAP) eliminate the hardest problems** -- sneak paths, IR drop, and static leakage vanish when computation uses displacement current rather than conductive paths.
2. **ADC is the bottleneck, not the array** -- 50-80% of energy and >50% of area goes to ADCs. The entire peripheral circuit architecture should be designed around minimizing ADC overhead.
3. **4-bit converters are the sweet spot** -- not 5-bit, not 8-bit. Recent hardware-aware quantization studies converge on 4-bit DAC/ADC as optimal cost-performance.

---

## PART 1: CROSSBAR ARRAY -- What the Literature Says

### Architecture Selection: The Capacitive Paradigm Shift

The project currently models three resistive architectures (0T1R, 1T1R, 2T1R). The literature is moving toward a fourth:

| Architecture | Sneak Paths | IR Drop | Static Leakage | Density | Demonstrated Size |
|---|---|---|---|---|---|
| 0T1R resistive | Severe (5-20%) | Severe | Yes | 4F^2 | 128x128 max practical |
| 1T1R resistive | Minimal (<0.1%) | Moderate | Transistor leakage | 8-12F^2 | >1024x1024 |
| 2T1R resistive | Negligible | Low | Low | 10-12F^2 | >1024x1024 |
| **0T1C capacitive (FeCAP)** | **None** | **None** | **None** | **4F^2** | **128x128 demonstrated** |

**Key papers:**
- Capacitive crossbar at 128x128: 3.8 pJ/MVM, 14-57x lower energy than resistive (Adv. Intell. Syst. 2022, DOI: `10.1002/aisy.202100258`)
- FeCAP intrinsic immunity to sneak paths via displacement current (Nano Convergence 2024, DOI: `10.1186/s40580-024-00463-0`)
- FeCaps + FeFETs eliminate selectors entirely (Scientific Reports 2024, DOI: `10.1038/s41598-024-59298-8`)

**Key physics difference:** FeCAPs act as ideal capacitors during the hold phase. No steady-state conductive path exists, thereby eliminating leakage currents and static power dissipation even in large crossbar arrays. Computation is driven by displacement currents rather than conductive paths, which inherently eliminates issues like static leakage, sneak-path effects, and IR drop.

**Recommendation for the simulator:** Add FeCAP as a fourth architecture. The physics is fundamentally different -- charge-domain computation rather than current-domain. This would require:
- Capacitance matrix instead of conductance matrix
- Charge integration readout instead of current summation
- Transient pulse-based MVM (not steady-state)

### Array Sizing: What's Realistic

| Architecture | Practical Max | Limiting Factor |
|---|---|---|
| 0T1R resistive | 32x32 to 128x128 | Sneak paths + IR drop |
| 1T1R resistive | 256x256 to 1024x1024 | Transistor series resistance, IR drop |
| 2T1R resistive | >1024x1024 | Area cost |
| 0T1C capacitive | 128x128 demonstrated | Charge sharing, parasitic capacitance |

Recent demonstrations:
- 24x24 and 48x48 FTJ crossbars (Adv. Intell. Syst. 2025, DOI: `10.1002/aisy.202500817`)
- 2-kilobyte AlScN crossbar operating at 600C (ScienceDirect 2025)
- 256x256 FeFET array (Jerry et al., IEEE IEDM 2017, DOI: `10.1109/IEDM.2017.8268338`)

**1T1R limitation:** The series resistance of the transistor is a major problem in 1T1R crossbar arrays, limiting the maximum current available for inducing resistive switching and degrading array performance. Additionally, the switching time for the 1T1R configuration increases as the crossbar size increases.

### Non-Ideality Modeling: What CrossSim/NeuroSim/MNSIM Include

#### CrossSim V3.1 (Sandia, Jan 2025)
- Arbitrary programming errors (5 distribution models)
- Conductance drift
- Cycle-to-cycle read noise
- Parasitic metal resistance (SOR solver)
- ADC precision loss
- NEW in V3.1: PyTorch/TensorFlow integration, GPU acceleration

#### NeuroSim V1.5 (Georgia Tech, 2025)
- Full device -> circuit -> algorithm stack
- <1% chip-level error after calibration
- 6.5x faster than V1.4
- Now supports non-volatile capacitive memories (FeCAP)
- Includes peripheral circuit area/power/latency estimation

#### MNSIM 2.0 (Tsinghua)
- 7000x faster than SPICE
- Behavior-level interconnect resistance model
- Trade-off optimization across multiple metrics

#### Comparison: Our Project vs. Literature Tools

| Non-Ideality | Our Project | CrossSim | NeuroSim | Priority to Add |
|---|---|---|---|---|
| IR drop (SOR solver) | Yes | Yes | Yes | Done |
| Sneak paths | Yes (3-cell) | Limited | Yes | Extend to multi-hop |
| Programming errors | Yes (5 models) | Yes | Yes | Done |
| Conductance drift | Yes | Yes | Yes | Done |
| Process variation (D2D) | Yes | Yes | Yes | Done |
| Cycle-to-cycle variation | Partial | Yes | Yes | **HIGH -- add state-dependent C2C** |
| Temperature effects | Yes | No | Partial | Done |
| Endurance/fatigue | Yes (basic) | No | Yes | Enhance wake-up model |
| Write disturb (half-select) | Yes | No | Partial | Done |
| **Non-linear I-V** | **No** | Yes | Yes | **HIGH** |
| **Capacitive mode (FeCAP)** | **No** | No | **V1.5 Yes** | **MEDIUM** |
| Read disturb (FE specific) | No | No | No | **MEDIUM** |

### State-Dependent Cycle-to-Cycle Variation (Missing)

Recent measurement data shows C2C variation is **not constant** -- it depends on the conductance state:

> "An effective conductance variation model derived from experimental measurements of C2C and D2D variations performed on FeFET devices fabricated using 28nm HKMG technology. The variations were found to be a **function of different conductance states** within the given programming range." (arXiv 2023, `2312.15444`)

Current project model uses uniform noise. Should be replaced with state-dependent sigma:
```
sigma(G) = sigma_base * f(G)  where f(G) varies across the conductance range
```

Best-in-class devices achieve ~0.3% C2C / ~0.5% D2D (Science Advances 2024, DOI: `10.1126/sciadv.adp0174`).

### Multi-Level Cell (MLC) Programming: State of the Art

| Levels | Bits/Cell | Who | How | Reference |
|---|---|---|---|---|
| 8 | 3 | Various | Standard ISPP | PMC 2023 |
| 16 (QLC) | 4 | KAIST | Gate stack engineering, >10V MW | Science Advances 2024, `10.1126/sciadv.adn1345` |
| 32 | 5 | Samsung | Production FeFET NAND | Nature 2025, `10.1038/s41586-025-09793-3` |
| >256 | 8+ | Various | Optimized write circuits | Nature Comms 2023, `10.1038/s41467-023-36270-0` |
| 3,024 | ~11.5 | Nature Electronics | Gate voltage + source-drain pulse superposition | Nature Electronics 2025, `10.1038/s41928-025-01551-7` |

**Key insight for ISPP:** Standard ISPP has a fundamental limitation -- the threshold voltage shows a **nonlinear relationship** with increasing pulse amplitudes because it's not optimized for partial polarization switching. Two alternatives:

1. **Displacement Current Control (DCC):** One-shot programming tailored to FE polarization switching -- more efficient than iterative ISPP (PMC 2024)
2. **Adaptive ISPP (A-ISPP):** Adaptive step voltage -- small steps near target, large steps far away (Seoul National University)

Our project's ISPP controller already uses binary search bisection which is similar to A-ISPP. The DCC approach would be a different paradigm worth adding as an alternative engine.

### Conductance Model: Linear is Wrong

The literature is clear: **linear conductance mapping is inaccurate for ferroelectrics**:

> "Synaptic weight update behavior must be linear and symmetric for MVM computation... A **mapping function is used to scale the nonlinear device voltage to a linear drive voltage**" (Adv. Intell. Syst. 2024, `10.1002/aisy.202400211`)

Our project already has exponential and lookup models. The Preisach model integration (which we've already done) is the correct approach per literature:

> "The Preisach model successfully describes hysteretic switching in ferroelectrics, connecting the Preisach distribution to measured microscopic switching kinetics... reproductions of polarization-voltage characteristics, the history-dependence and minor loops" (Nature Comms 2018, `10.1038/s41467-018-06717-w`)

**Recommendation:** Make exponential/Preisach the default, not linear. Linear should be explicitly labeled as "simplified/educational mode."

---

## PART 2: PERIPHERAL CIRCUITS -- What the Literature Says

### The ADC Problem: Central Design Challenge

This is the single most important finding:

> "Energy consumption of analog CIMs is **dominated by full-precision ADCs**" (IEEE ISSCC 2023)

> "ADCs in FeFET CiM arrays incur **large area, power, and latency overheads**" (Nature Comms 2024)

> "Architecture-level ADC decisions such as ADC resolution or number of ADCs **significantly impact overall CIM accelerator energy and area**" (arXiv 2024)

Our project currently: 5-bit SAR ADC, 50ns conversion, 25 fJ/conversion. This is reasonable but the architecture should be designed around the ADC, not treat it as just another component.

### ADC Architecture Recommendations

| ADC Type | Resolution Sweet Spot | Energy | Latency | Best For |
|---|---|---|---|---|
| **SAR** | 4-6 bit | Low | Moderate (50ns @ 5-bit) | Per-column, low throughput |
| Flash | 3-4 bit | High | Very fast (1-2ns) | Ultra-low latency |
| Sigma-Delta | 8-12 bit | Medium | Slow (us) | High precision |
| Ramp/Slope | 6-8 bit | Very low | Slow (100ns+) | Column-shared, area-efficient |
| **Comparator-only** | 1 bit | **28x lower than 7-bit ADC** | Fast | Binary/ternary networks |

**Key finding -- 4-bit is optimal:**

> "Experiments demonstrate improvements on CIFAR-10 and ImageNet... identifying **4-bit data converters as the optimal balance** between cost and performance" (arXiv 2024)

> "128x128 non-volatile capacitive crossbar array is compatible with **3-bit ADC quantization**" (Adv. Intell. Syst. 2022)

**Recommendation:** Change default from 5-bit to 4-bit. Add multiple ADC architectures (SAR, Flash, Ramp) as selectable options with different energy/latency/area tradeoffs.

### Column-Shared vs Per-Column ADC

This is a critical architectural decision the project doesn't currently model:

| Strategy | ADC Count | Area | Energy | Latency | Best When |
|---|---|---|---|---|---|
| Per-column | N_cols | Very high | Low per-op | Low | High throughput needed |
| Shared (1 ADC) | 1 | Minimal | High per-op | High (N_cols x t_conv) | Area-constrained |
| Shared (K ADCs) | K | Medium | Medium | Medium | Balanced |

> "Using 1, 2, 4, 8, and 16 ADCs in parallel... the choice of number of ADCs can influence overall accelerator **energy-area product (EAP) by a factor of three**" (arXiv 2024)

**Recommendation:** Add configurable ADC sharing ratio (per-column, shared-4, shared-8, shared-all) and model the latency/area/energy implications.

### DAC Architecture Recommendations

| Topology | Pros | Cons | CIM Use |
|---|---|---|---|
| **Capacitive (switched-cap)** | Low power, good matching | Slower settling | Most common in CIM |
| R-2R ladder | Equal switch currents, good matching | Area | Common alternative |
| Current-steering | Fastest | Highest power | High-speed CIM |
| Hybrid (cap + current) | Balanced | Complex | Emerging |

**Input encoding matters:**

| Encoding | Bits Required | Glitch-Free | Linearity |
|---|---|---|---|
| Binary | N | No (glitches at MSB transitions) | Requires calibration |
| **Thermometer** | 2^N - 1 | Yes | Inherently monotonic |
| Segmented (MSB thermo + LSB binary) | Reduced | Mostly | Good compromise |

Our project uses simple binary quantization. For accuracy, thermometer or segmented encoding should be modeled, especially for >4-bit operation.

**Recommendation:** Change default DAC to 4-bit. Add thermometer encoding option.

### Sense Chain: Charge-Domain is the Future

The project uses current-domain TIA sensing. The literature shows charge-domain is superior for ferroelectric:

> "Charge-based sensing schemes achieving at least an **order of magnitude reduction in power consumption** compared to current-based methods" (Scientific Reports 2024)

> "Ferroelectric capacitive memories transfer memory reading and in-memory computing to **charge domain**" (Nano Convergence 2024)

For resistive crossbars, TIA is correct. But for FeCAP crossbars, the sensing should use charge integration:
- Charge amplifier instead of TIA
- Integration time window instead of steady-state current
- Lower noise (no shot noise from DC current)

**Recommendation:** Keep TIA for resistive modes, add charge amplifier for FeCAP mode.

### Write Drivers: Voltage Requirements

| Device Type | Program Voltage | Erase Voltage | Pulse Width | Energy/bit |
|---|---|---|---|---|
| HZO FeFET (standard) | 3-5V | -(3-5V) | 10-100ns | ~0.1-50 fJ |
| Samsung FeFET NAND | Near-zero pass voltage | -- | 1us | 96% power savings |
| Dual-Bit FeFET | +/-3.3V | +/-3.3V | 1us | -- |
| HZO FeCAP | 1.5-3V | -(1.5-3V) | 10-100ns | <1 fJ |

Our project's charge pump model (1.0V -> 1.5V, Dickson) is appropriate for FeCAP. For FeFET, 3-5V generation is needed -- likely requires a 3-stage or 4-stage pump from 1.0V CMOS supply.

**Recommendation:** Add configurable charge pump staging (2-stage for FeCAP/low-Vc, 4-stage for FeFET/high-Vc).

### DAC/ADC-Less Architectures (Emerging)

Several recent papers eliminate converters entirely:

> "Near-CIM analog memory and nonlinear activation units bring **76.0% energy reduction** compared with DAC/ADC solutions" (ResearchGate 2024)

> "Elimination of high-precision ADCs via **comparator-only digitization** reduces energy up to **28x**" (arXiv 2024)

This is forward-looking but worth modeling as an option: binary-weight networks with comparator readout instead of full ADC.

### System-Level Energy Budget: What Papers Actually Report

| Component | % of Total Energy (Resistive) | % of Total Energy (Capacitive) |
|---|---|---|
| Array (MVM) | 10-30% | 5-15% |
| ADC | **40-60%** | **30-50%** |
| DAC | 5-15% | 5-15% |
| TIA/Sense | 5-10% | 5-10% |
| Write drivers | 10-20% | 5-10% |
| Other (control, routing) | 5-10% | 5-10% |

Our project's current energy breakdown (ADC 55%, DAC 31%, TIA 14%) is in the right ballpark but overweights DAC. Literature suggests ADC dominance is even stronger at higher resolutions.

### Latency Budget

| Component | Our Project | Literature Typical | Notes |
|---|---|---|---|
| DAC settling | 10ns | 5-20ns | Reasonable |
| Array physics | 1-5ns | 1-10ns | OK for resistive; FeCAP may need pulse width |
| TIA settling | 11ns | 5-20ns | Reasonable |
| ADC conversion | 50ns (SAR) | 10-100ns depending on type | OK |
| **Total read** | **76ns** | **30-150ns** | Reasonable |

FeFET search operations have been demonstrated at **100 picoseconds** (Nano Letters 2022), but that's the device physics -- the peripherals still dominate total latency.

---

## PART 3: SYNTHESIS -- How To Build It Right

### Recommended Architecture Overhaul

#### Crossbar Module (Module 2)

1. **Add FeCAP architecture** as a first-class mode alongside 0T1R/1T1R/2T1R
   - Capacitance matrix (not conductance)
   - Charge-domain MVM: Q = C x V, output = sum of charges
   - No sneak paths, no IR drop, no static leakage
   - Transient pulse-based operation

2. **Make non-linear conductance the default**
   - Exponential model as default for resistive
   - Preisach-based for physics-accurate mode (already implemented)
   - Linear only for "educational/simplified" mode

3. **Add state-dependent C2C variation**
   - sigma(G) varies with conductance level
   - Calibrate from published 28nm FeFET data

4. **Add non-linear I-V curves**
   - Currently all cells assumed ohmic (I = G x V)
   - Real FeFETs have non-linear I-V in subthreshold
   - Matters most for low-conductance states

5. **Extend sneak path model**
   - Multi-hop paths beyond 3-cell for large passive arrays
   - Sparse matrix support for >128x128

#### Circuits Module (Module 4)

1. **ADC-centric redesign**
   - Multiple ADC architectures: SAR (default), Flash, Ramp, Comparator-only
   - Default to 4-bit (not 5-bit) per literature consensus
   - Configurable ADC sharing ratio (per-column, shared-K, shared-all)
   - Model area/energy/latency tradeoffs for each configuration

2. **DAC improvements**
   - Default to 4-bit
   - Add thermometer/segmented encoding option
   - Model glitch energy for binary encoding

3. **Dual sensing modes**
   - Current-domain (TIA) for resistive crossbars
   - Charge-domain (charge amplifier) for FeCAP crossbars
   - Configurable gain/bandwidth/noise per architecture

4. **Configurable charge pump**
   - 2-stage for low-Vc (FeCAP, ~1.5V)
   - 3-4 stage for high-Vc (FeFET, 3-5V)
   - Model efficiency vs. number of stages

5. **ISPP enhancements**
   - Add DCC (Displacement Current Control) as alternative to ISPP
   - Keep adaptive bisection (current approach is good)
   - Add one-shot programming option for FeCAP

6. **Peripheral area/overhead model**
   - Literature shows peripherals are ~55% of total chip area
   - Model ADC area as function of resolution and count
   - Show area breakdown pie chart in GUI

### Priority Order for Implementation

| Priority | Change | Impact | Effort |
|---|---|---|---|
| **P0** | Change default to 4-bit DAC/ADC | Accuracy to literature consensus | Low |
| **P0** | Make exponential conductance the default | Physics accuracy | Low (config change) |
| **P1** | Add multiple ADC architectures (SAR, Flash, Ramp) | Educational value, accuracy | Medium |
| **P1** | Add ADC sharing ratio model | Realistic system-level metrics | Medium |
| **P1** | State-dependent C2C variation | Physics accuracy | Medium |
| **P2** | Add FeCAP mode (capacitive crossbar) | Major architecture addition | High |
| **P2** | Charge-domain sensing for FeCAP | Pairs with FeCAP mode | Medium |
| **P2** | Non-linear I-V curves | Physics accuracy | Medium |
| **P3** | DCC programming alternative | Forward-looking | Medium |
| **P3** | Multi-hop sneak paths | Large passive arrays | Medium |
| **P3** | DAC/ADC-less mode (comparator-only) | Emerging architecture | Low-Medium |
| **P3** | Configurable charge pump staging | Completeness | Low |

---

## PART 4: COMPLETE REFERENCE -- Papers and DOIs

### Peer-Reviewed Publications (Cited in This Report)

| # | Topic | DOI | Year |
|---|---|---|---|
| 1 | Capacitive crossbar MVM (128x128) | `10.1002/aisy.202100258` | 2022 |
| 2 | FeCAP memory comprehensive review | `10.1186/s40580-024-00463-0` | 2024 |
| 3 | FeCaps/FeFETs as IMC elements | `10.1038/s41598-024-59298-8` | 2024 |
| 4 | Multi-level FeFET crossbar (885 TOPS/W) | `10.1038/s41467-023-42110-y` | 2023 |
| 5 | 16-level QLC FeFET (>10V MW) | `10.1126/sciadv.adn1345` | 2024 |
| 6 | Samsung FeFET NAND (32 levels) | `10.1038/s41586-025-09793-3` | 2025 |
| 7 | 3,024 states transistor | `10.1038/s41928-025-01551-7` | 2025 |
| 8 | >256 analog states | `10.1038/s41467-023-36270-0` | 2023 |
| 9 | Recent advances in ferroelectrics review | `10.1186/s40580-025-00520-2` | 2025 |
| 10 | Preisach model for ferroelectrics | `10.1038/s41467-018-06717-w` | 2018 |
| 11 | HfO2 FeFET comprehensive review | `10.1063/5.0206599` | 2024 |
| 12 | 2D ferroelectric hybrid CIM | `10.1126/sciadv.adp0174` | 2024 |
| 13 | FeFET CiM annealer | `10.1038/s41467-024-46640-x` | 2024 |
| 14 | 2T2R for differential IMC | `10.1007/s11432-023-3887-0` | 2024 |
| 15 | FTJ crossbar half-bias | `10.1002/aisy.202500817` | 2025 |
| 16 | Dual-Bit FeFET | `10.1038/s44335-025-00030-8` | 2025 |
| 17 | Ferroelectric materials review (China) | `10.1007/s11432-025-4432-x` | 2025 |
| 18 | 2D FE for in-sensor computing | `10.1002/adma.202400332` | 2024 |
| 19 | FeFET FeRAM review | `10.1063/5.0086328` | 2023 |
| 20 | FeFET temperature effects | `10.1016/S0038-1101(24)00103-5` | 2024 |
| 21 | 256x256 FeFET array | `10.1109/IEDM.2017.8268338` | 2017 |
| 22 | Delta-Sigma CIM (21.38 TOPS/W) | `10.1109/ISSCC42615.2023.10067289` | 2023 |
| 23 | Linear conductance modulation | `10.1002/aisy.202400211` | 2024 |
| 24 | CSCDAC hybrid DAC | `10.3390/jlpea15010009` | 2025 |
| 25 | Preisach parameter automation | `10.1038/s41598-021-91492-w` | 2021 |
| 26 | Memristor crossbar sensing review | `10.1021/acs.chemrev.4c00845` | 2024 |

### Conference and arXiv References

| Topic | Reference | Year |
|---|---|---|
| State-dependent C2C FeFET variation | arXiv `2312.15444` | 2023 |
| ADC optimization for CIM | arXiv `2404.06553` | 2024 |
| 4-bit optimal converters (HW-aware quant) | arXiv `2508.21524` | 2024 |
| DAC/ADC-less CIM | arXiv `2412.19869` | 2024 |
| NeuroSim V1.5 | arXiv `2505.02314` | 2025 |
| DCC one-shot programming | PMC `PMC11160465` | 2024 |
| FeFET 3D NAND training accelerator | ResearchGate `349145591` | 2021 |

### Simulation Tool References

| Tool | Version | Source | URL |
|---|---|---|---|
| CrossSim | V3.1 (Jan 2025) | Sandia National Labs | `cross-sim.sandia.gov` |
| NeuroSim | V1.5 (2025) | Georgia Tech | `github.com/neurosim` |
| MNSIM | V2.0 | Tsinghua University | ResearchGate `344931525` |
| badcrossbar | -- | UCL | DOI: `10.1016/j.softx.2020.100617` |

### Textbook / Standard References

- J. F. Dickson, "On-chip high-voltage generation in MNOS integrated circuits using an improved voltage multiplier technique," IEEE JSSC, 1976. (charge pumps)
- B. Razavi, Data Conversion System Design, IEEE Press/Wiley, 1995. (ADC architecture, noise/ENOB)
- IEEE Std 1241-2010, IEEE Standard for Terminology and Test Methods for ADCs. (performance metrics)
- D. M. Young, Iterative Solution of Large Linear Systems, 1971. (SOR algorithm)

---

## PART 5: EXISTING PROJECT STATE (For Context)

### What Module 2 Already Has
- SOR solver ported from CrossSim (solver.go)
- 5 programming error models from CrossSim (device_errors.go)
- 3-cell sneak path model (sneakpath.go)
- Power-law and logarithmic drift (drift.go)
- Temperature effects with 5 presets (temperature.go, temperature_profile.go)
- Process variation with spatial gradients (array.go)
- Endurance/fatigue model (array.go)
- Half-select write disturb (write_disturb.go)
- Differential arrays for signed weights (enhanced.go)
- Write-verify programming (enhanced.go)
- Three conductance models: linear, exponential, lookup (array.go + shared/physics)
- Vulkan GPU-accelerated MVM (gpu_mvm.go)

### What Module 4 Already Has (Shared Peripherals)
- 5-bit DAC with INL/DNL (shared/peripherals/dac.go)
- 5-bit SAR ADC with ENOB/SNR (shared/peripherals/adc.go)
- TIA with noise model (shared/peripherals/tia.go)
- 2-stage Dickson charge pump (shared/peripherals/chargepump.go)
- System-level timing/power analysis (shared/peripherals/analysis.go)
- Tier-A DC nodal solver (module4-circuits/pkg/arraysim/tier_a.go)
- Sense chain model (module4-circuits/pkg/arraysim/sensechain.go)

### Known Gaps (from project's own MODULE4-PHYSICS-IMPROVEMENTS.md)
1. Linear conductance model (need exponential/Preisach) -- HIGH
2. No sneak paths in Module 4 -- HIGH
3. No IR drop in Module 4 -- HIGH
4. No write disturb tracking -- HIGH
5. Temperature effects (only 300K) -- MEDIUM
6. No switching statistics -- MEDIUM
7. No endurance model -- MEDIUM
8. TIA frequency response -- MEDIUM
9. Charge pump transient response -- LOW
10. ADC comparator kickback noise -- LOW
11. SET/RESET asymmetry -- LOW
12. Retention loss model -- LOW
