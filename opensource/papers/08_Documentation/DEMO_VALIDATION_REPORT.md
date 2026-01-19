# IronLattice Demo Validation Report

**Comparison of Dr. Tour's Transcript Claims vs. Demo Implementations**

**Date:** 2026-01-19
**Source:** ironlattice-transcript.md (Dr. Tour's Nov 2024 Presentation)

---

## Executive Summary

This report compares the technical claims from Dr. Tour's IronLattice presentation against our demo implementations to identify accuracy, gaps, and discrepancies.

| Area | IronLattice Claim | Demo Implementation | Status |
|------|-------------------|---------------------|--------|
| **30 Discrete States** | "30 discrete states" | `IronLatticeLevels = 30` | **CORRECT** |
| **MNIST Accuracy** | 87% (88% theoretical max) | Claims 95.8% | **DISCREPANCY** |
| **Endurance** | 10^9 demonstrated, 10^12 target | 10^10 default, 10^12 optimized | **REASONABLE** |
| **Energy vs NAND** | 10,000,000x lower | Not explicitly modeled | **MISSING** |
| **Switching Speed** | 10ns demonstrated | 1ns in material params | **CLOSE** |
| **P-E Curve Shape** | Square loop, stable over 10^7s | Tanh model + Preisach | **SIMPLIFIED** |

---

## Detailed Analysis

### 1. 30 Discrete Analog States

**IronLattice Claim (Transcript):**
> "It's got 30 discrete states. So we have all these intermediate states. So it's not 0-1-0-1."

**Slide 10 Data:**
- "30 Intermediate States" graph showing PSC (State) from 0-60
- Shows long-term potentiation and depression (LTP and LTD)
- Vg = 0.1V measurement condition

**Demo Implementation:**

| Location | Implementation |
|----------|----------------|
| `demo2-crossbar/pkg/crossbar/array.go:12` | `const IronLatticeLevels = 30` |
| `demo2-crossbar/pkg/crossbar/array.go:88-96` | `QuantizeTo30Levels()` function |
| `demo1-hysteresis/pkg/ferroelectric/preisach.go:203-213` | `DiscreteStates(N int)` function |
| `demo1-hysteresis/pkg/ferroelectric/material.go:201-209` | `DiscreteLevel()` function |

**Assessment:** **CORRECT** - The 30-level quantization is correctly implemented throughout.

---

### 2. MNIST Handwritten Digit Recognition

**IronLattice Claim (Transcript):**
> "We've done the compute in memory. We've put this on the MNIST system. So we can read these handwritten numbers. **We're at 87% validation here**."
>
> "Estimated pattern recognition accuracy is ~87% (**88% theoretical maximum**)."

**Slide 10 Data:**
- Graph shows accuracy vs. epoch reaching ~60% (experimental curve)
- "Estimated pattern recognition accuracy is ~87%"
- Black dashed line shows "Ideal case"

**Demo Implementation:**
- `demo2-crossbar/README.md` and `demo2-crossbar/demo2.README.md` claim **95.8%**
- `demo3-mnist` neural network achieves this through simulation

**Assessment:** **MAJOR DISCREPANCY**

| Metric | IronLattice Hardware | Our Simulation |
|--------|---------------------|----------------|
| Accuracy | 87% | 95.8% |
| Theoretical Max | 88% | ~98% (software) |
| Test Conditions | Physical FeFET array | Software simulation |

**Why the Discrepancy:**
1. IronLattice's 87% is on **real hardware** with:
   - Device-to-device variation
   - Cycle-to-cycle noise
   - IR drop effects
   - Sneak path currents
   - ADC/DAC quantization noise
   - Temperature drift

2. Our simulation achieves 95.8% because:
   - Noise levels are configurable (default may be optimistic)
   - Perfect voltage control (no IR drop in simple MVM)
   - Idealized quantization

3. The "88% theoretical maximum" appears to be specific to their architecture constraints, not a universal limit.

**Recommendation:**
- Add a "realistic mode" with IronLattice-calibrated non-idealities
- Show both "ideal simulation" and "IronLattice-matched" accuracy
- Clarify in documentation that 87% is the **hardware-demonstrated** value

---

### 3. P-E Hysteresis Curve Parameters

**IronLattice Claims (Slides 8-9):**
- Stable P-E curves for retention times: 10^4, 10^5, 10^6, 10^7 seconds
- ON-OFF ratio > 10^5 maintained across 10^9 cycles
- Pulse conditions: 100, 10, 1 μs, 100, and 10 ns
- "Wake-up → Stable operation → Fatigue" lifecycle

**Demo Implementation:**

| Parameter | IronLattice | Demo (`IronLatticeMaterial()`) |
|-----------|-------------|-------------------------------|
| Pr | ~20-30 μC/cm² (estimated) | 30 μC/cm² |
| Ps | ~25-35 μC/cm² (estimated) | 35 μC/cm² |
| Ec | ~1 MV/cm (typical HZO) | 1.0 MV/cm |
| Switching time | 10-100 ns | 1 ns |
| Endurance | 10^9 demonstrated | 10^11 |
| Retention | >10 years | 100 years |

**Assessment:** **REASONABLE** but values are estimates since exact numbers weren't disclosed.

**Improvements Needed:**
1. The transcript doesn't provide exact Pr/Ps values - we're using literature values
2. Switching time of 1ns may be optimistic (10ns demonstrated in slides)
3. Should add "pulse condition" simulation showing different switching speeds

---

### 4. Energy Comparisons

**IronLattice Claims (Transcript):**

| Comparison | Metric | Improvement |
|------------|--------|-------------|
| vs NAND Flash | Read/Write Energy | 10,000,000x lower |
| vs NAND Flash | Speed | 1,000,000x faster |
| vs NAND Flash | Voltage | 90% reduction (2-3V vs 10-20V) |
| vs DRAM | Read/Write Energy | 1,000x lower |
| vs DRAM | Refresh | Zero (non-volatile) |
| Data Center | Total Energy | 80-90% reduction |

**Demo Implementation:**
- Demo 4 (peripherals) models DAC/ADC energy but not the full comparison
- These comparisons are mentioned in documentation but not visualized

**Assessment:** **PARTIALLY MISSING**

**Recommendation:**
- Add energy comparison visualization to demos
- Show the "80-90% data center reduction" claim with supporting calculations

---

### 5. Endurance and Retention

**IronLattice Claims:**
- **Demonstrated:** 10^9 cycles with stable operation
- **Target:** 10^12 cycles (mentioned as goal, not achieved)
- **Retention:** Demonstrated over 10^7 seconds (~116 days)
- Slides show ON-OFF ratio maintained throughout

**Demo Implementation:**
```go
// IronLatticeMaterial() in material.go
EnduranceCycles: 1e11,  // 10^11 cycles
RetentionTime:   3.15e9, // 100 years at 85°C
```

**Assessment:** **OPTIMISTIC**

Our demo uses 10^11 endurance (between demonstrated 10^9 and target 10^12). This is reasonable as an intermediate value but should be documented.

---

### 6. Device Architecture

**IronLattice Slide 3 (Two Device Structures):**

1. **FTJ (Ferroelectric Tunnel Junction)** - 2-terminal
   - Ferroelectric layer with proprietary superlattice
   - Compact design

2. **FeFET (Ferroelectric FET)** - 3-terminal
   - Ferroelectric layer with proprietary superlattice
   - Higher functionality

**Demo Implementation:**
- Currently models FeFET behavior (conductance modulation)
- Does not explicitly distinguish FTJ vs FeFET physics

**Recommendation:**
- Add FTJ device model for comparison
- Document which device type each demo assumes

---

### 7. Pulse Timing Specifications

**IronLattice Slide 8:**
> "Pulse condition: 100, 10, 1 μs, 100, and 10 ns"

This indicates switching was demonstrated across 4 orders of magnitude of pulse widths.

**Demo Implementation:**
- Demo 1 uses `Tau: 1e-9` (1 ns) as characteristic switching time
- Does not visualize different pulse width effects

**Recommendation:**
- Add pulse width parameter to Demo 1
- Show how different pulse widths affect the P-E loop (as in Oh et al. Scheme A/B/C)

---

## Data Extraction from Slides

### Slide 8: Endurance Test Data Points (Estimated from Graph)

```
Cycle Number (log scale) | Polarization State
------------------------|-------------------
10^0                    | Clear dual states
10^1                    | Clear dual states
10^2                    | Clear dual states (end of shown data)
```

### Slide 9: Retention Data Points (Estimated)

```
Retention Time | ON-OFF Ratio
--------------|-------------
10^4 s        | >10^5
10^5 s        | >10^5
10^6 s        | >10^5
10^7 s        | >10^5
```

### Slide 10: 30-State PSC Data (Estimated from Graph)

```
State Number | PSC (arbitrary units)
------------|---------------------
0           | ~0
5           | ~10
10          | ~20
15          | ~30
20          | ~40
25          | ~50
30          | ~60
```

The relationship appears **linear**, confirming IronLattice achieves good linearity.

---

## Recommendations Summary

### Immediate Fixes

1. **MNIST Accuracy Claim**
   - Change "95.8%" claims to clarify this is simulation
   - Add "IronLattice demonstrated: 87% on hardware"
   - Add note that 88% is stated theoretical maximum for their system

2. **Documentation Updates**
   - Clearly distinguish "simulation results" from "IronLattice hardware results"
   - Add source citations to transcript for all claims

### Demo Enhancements

1. **Demo 1 (Hysteresis)**
   - Add pulse width parameter (10ns to 100μs range)
   - Show retention degradation over 10^7 seconds
   - Visualize endurance cycling effects

2. **Demo 2 (Crossbar)**
   - Add "IronLattice mode" with calibrated noise levels
   - Target 87% accuracy to match hardware
   - Show IR drop effects more prominently

3. **Demo 3 (MNIST)**
   - Add mode that targets 87% accuracy (realistic noise)
   - Document why simulation exceeds hardware
   - Add "theoretical maximum" reference

4. **New: Energy Comparison Demo**
   - Visualize 10,000,000x NAND improvement
   - Show 80-90% data center energy reduction
   - Compare to DRAM refresh power

---

## Validation Checklist

| Claim | Verified | Notes |
|-------|----------|-------|
| 30 discrete states | **YES** | Correctly implemented |
| MNIST 87% accuracy | **PARTIAL** | Simulation exceeds claim |
| 10^9 endurance | YES | Model supports this |
| Non-volatile | YES | Implicit in ferroelectric model |
| CMOS compatible | N/A | Not modeled (manufacturing) |
| 80-90% energy reduction | **NO** | Not directly visualized |
| 10ns switching | YES | 1ns in model (optimistic) |
| FTJ/FeFET architectures | **PARTIAL** | FeFET only |

---

## Conclusion

Our demos **correctly model the 30-level analog states** which is the core IronLattice innovation. However, there are discrepancies in accuracy claims and missing energy comparison visualizations. The main action items are:

1. **Fix the 95.8% → 87% accuracy discrepancy** in documentation
2. **Add IronLattice-calibrated non-ideality mode** for realistic simulation
3. **Create energy comparison visualizations** showing the claimed improvements

The demos successfully communicate IronLattice's core value proposition but need calibration to match the actual hardware demonstration results.
