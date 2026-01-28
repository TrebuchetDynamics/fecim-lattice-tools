# Executive Summary: Write Operation Analysis

**Research Stage 4: Module 2 vs Module 4 Write Operations**  
**Date**: 2026-01-27  
**Status**: Complete

---

## Quick Answer

**Q: How do WRITE operations work in Module 2 vs Module 4?**

**A**: Module 2 provides a high-level abstraction (`ProgramWeight()`) that quantizes weights to 30 discrete conductance levels and stores them. Module 4 models the low-level peripheral circuits (5-bit DAC, 2-stage charge pump) that generate the ±1.5V write voltages required for ferroelectric polarization switching. Together, they form a complete educational toolset covering array behavior and circuit implementation.

---

## Key Statistics

| Metric | Module 2 | Module 4 |
|--------|----------|----------|
| **Abstraction Level** | Array (high) | Circuit (low) |
| **Quantization Levels** | 30 discrete states | 32 DAC levels (use 30) |
| **Voltage Range** | Not modeled | -1.5V to +1.5V |
| **Write Timing** | Instantaneous | 200 ns (DAC + pump + pulse) |
| **Write Energy** | Not modeled | 2.18 pJ (full breakdown) |
| **IR Drop Modeling** | ✅ Implemented | ❌ Not modeled |
| **Nonlinearity** | ❌ Not modeled | ✅ INL/DNL (±48mV) |
| **ISPP Write-Verify** | ❌ Not implemented | ❌ Not implemented |

---

## Signal Path Summary

### Module 2: Array Abstraction

```
Weight (0.517) → QuantizeToLevels() → Conductance (0.517) → Cell Storage
                  [Level 15/30]
```

**What it models**:
- Discrete level quantization (30 states)
- Conductance storage and tracking
- Switching cycle counting
- Array geometry effects (IR drop, drift, sneak paths)

**What it doesn't model**:
- Voltage generation
- Programming pulse timing
- Energy consumption per write
- Circuit-level nonidealities

---

### Module 4: Peripheral Circuits

```
Level (15) → DAC (0.0V) → Charge Pump (±1.5V) → Crossbar (100ns pulse) → FeFET
             5-bit         2-stage Dickson       Physical switching
             ±1.5V range   70% efficiency        Mixed polarization
```

**What it models**:
- Digital-to-analog conversion (5-bit DAC)
- Voltage boosting (1V → ±1.5V charge pump)
- INL/DNL errors (±48mV)
- Energy breakdown (2.18 pJ total)
- Timing analysis (200ns latency)

**What it doesn't model**:
- Array-level behavior (sneak paths, IR drop)
- Cell state storage
- Ferroelectric physics (polarization dynamics)

---

## Complete Write Signal Path

```
┌────────────────────────────────────────────────────────────────┐
│                      FULL WRITE OPERATION                      │
├────────────────────────────────────────────────────────────────┤
│                                                                │
│  [APPLICATION]                                                 │
│       │                                                         │
│       ▼                                                         │
│  Module 2: ProgramWeight(row, col, weight)                     │
│       │                                                         │
│       ├─ Quantize: level = round(weight × 29)                  │
│       ├─ Store: cells[row][col].Conductance = level/29         │
│       └─ Track: SwitchingCount++                               │
│                                                                │
│  ═══════════════════════════════════════════════════════════   │
│                                                                │
│  [CIRCUIT LAYER - Module 4]                                    │
│       │                                                         │
│       ├─ DAC: level → voltage (±1.5V range)                    │
│       │   • 5-bit resolution                                   │
│       │   • INL/DNL errors                                     │
│       │   • 10ns settling                                      │
│       │                                                         │
│       ├─ Charge Pump: 1V → ±1.5V boost                         │
│       │   • 2-stage Dickson                                    │
│       │   • 70% efficiency                                     │
│       │   • 40ns rise time                                     │
│       │                                                         │
│       └─ Crossbar: Apply 100ns pulse                           │
│                                                                │
│  ═══════════════════════════════════════════════════════════   │
│                                                                │
│  [PHYSICS LAYER - Documented only]                            │
│       │                                                         │
│       └─ FeFET: Voltage → Polarization → Conductance           │
│           • Coercive voltage Vc = 0.6-1.5V                     │
│           • Partial polarization for intermediate levels       │
│           • ISPP write-verify (not implemented)                │
│                                                                │
└────────────────────────────────────────────────────────────────┘
```

---

## Voltage-Level-Conductance Mapping

| Level | Module 2 G | Module 4 V | Polarization | Read I |
|-------|------------|------------|--------------|--------|
| 0 | 0.000 | -1.500 V | 100% DOWN (↓) | 1 µA |
| 15 | 0.517 | 0.000 V | 50% MIXED (↑↓) | 50 µA |
| 29 | 1.000 | +1.403 V | 100% UP (↑) | 100 µA |

**Current Range**: 100:1 ratio (1 µA to 100 µA)  
**Sensing Margin**: Excellent for 30-level discrimination

---

## Kirchhoff's Laws During Write

### Voltage Division (KVL)

```
V_dac = I_write × (R_row + R_FeFET + R_col) + V_cell

1.5V = 150µA × (2.5Ω + 10kΩ + 2.5Ω)
     ≈ 150µA × 10kΩ = 1.5V

IR drop: 150µA × 5Ω = 0.75mV (<0.05% error)
```

**Finding**: IR drop negligible during WRITE because only one cell conducts.

### Current Conservation (KCL)

```
I_row_in = I_cell = I_col_out = 150 µA

All other cells: I = 0 (high impedance)
```

**Contrast with COMPUTE**: During MVM, ALL cells conduct → significant IR drop.

---

## Pulse Schemes

### Binary States (SET/RESET)

```
ERASE (Level 0):     PROGRAM (Level 29):
V = -1.5V, 100ns     V = +1.5V, 100ns

↓↓↓↓↓↓↓↓             ↑↑↑↑↑↑↑↑
(All DOWN)           (All UP)
VTH = 1.5V           VTH = 0.3V
I = 1 µA             I = 100 µA
```

### Multi-Level (Partial Polarization)

```
PARTIAL (Level 15):
V = 0.0V, 100ns

↑↓↑↓↑↓↑↓
(50% UP, 50% DOWN)
VTH = 0.9V
I = 50 µA
```

**Key Physics**: Voltage amplitude controls polarization fraction.

---

## Energy Budget (Module 4)

| Component | Energy | % of Total |
|-----------|--------|------------|
| DAC settling | 15 fJ | 0.7% |
| **Charge pump** | **2.14 pJ** | **98.2%** |
| Cell switching | 10 fJ | 0.5% |
| Peripherals | 5 fJ | 0.2% |
| **TOTAL** | **2.18 pJ** | **100%** |

**Note**: Charge pump dominates energy. Published FeFET write energy is 10-30 fJ per cell. The 2.18 pJ includes peripheral overhead that would be amortized across many writes in real systems.

---

## Timing Budget

```
0 ns    │ Module 2: ProgramWeight() called [instantaneous]
        │
0 ns    │ Module 4: DAC receives level 15
10 ns   │ DAC settled (0.0V ± 48mV)
50 ns   │ Charge pump stable (±1.5V)
150 ns  │ Write pulse complete (ferroelectric switched)
200 ns  │ [Optional: Write-verify] NOT IMPLEMENTED
        │
200 ns  │ COMPLETE
```

**Module 2**: Instantaneous (abstraction)  
**Module 4**: 200 ns (detailed circuit timing)

---

## Major Findings

### 1. Complementary Abstraction Layers ✅

Module 2 and Module 4 are **NOT redundant** - they model different aspects:
- **Module 2**: Array behavior (geometry, non-idealities, state management)
- **Module 4**: Circuit implementation (voltage generation, timing, energy)

**Educational Value**: Students learn both high-level programming abstraction AND low-level circuit design.

### 2. Consistent 30-Level Quantization ✅

Both modules use 30 discrete levels:
- **Module 2**: `DefaultQuantizationLevels = 30` → conductance 0.0-1.0
- **Module 4**: 5-bit DAC (32 levels) → voltage -1.5V to +1.5V

**Alignment**: Perfect consistency across abstraction layers.

### 3. Missing Physics Layer ⚠️

Voltage → Polarization → Conductance mapping exists only in **documentation**, not code:
- Ferroelectric switching physics documented in `circuits.CIM-fundamentals.md`
- No code implementation linking Module 4 voltage to Module 2 conductance
- Acceptable for educational tools, limits research applications

**Future Work**: Implement physics-based transfer function.

### 4. No Write-Verify ISPP Implementation ⚠️

Incremental Step Pulse Programming (ISPP):
- **Documentation**: Fully detailed in `circuits.CIM-fundamentals.md:231-248`
- **Circuits**: All building blocks exist (DAC, TIA, ADC, comparator)
- **Integration**: NOT IMPLEMENTED in either module

**Impact**: Current implementation does single-pulse writes without accuracy feedback.

**Estimated Implementation**: ~200 lines of Go code (read-back + comparison + iteration loop).

### 5. IR Drop Irrelevant During WRITE ✅

Module 2's `IRDropSimulator` is highly relevant for READ/COMPUTE (many cells active) but minimally impacts WRITE:
- **WRITE**: Single cell conducts → IR drop <0.1%
- **READ**: One row active → 5-20% error (passive arrays)
- **COMPUTE**: All cells active → IR drop significant

**Conclusion**: IR drop simulator correctly models array-level effects, but write accuracy is dominated by DAC INL/DNL (±48mV), not IR drop.

---

## Recommendations

### For Immediate Understanding

**Use Case 1**: "How do I program a weight to the crossbar?"
→ **Answer**: Use Module 2's `ProgramWeight()` - it handles quantization and storage automatically.

**Use Case 2**: "What voltage does the DAC generate for level 15?"
→ **Answer**: Module 4's `DAC.Convert(15)` returns 0.0V (mid-scale of ±1.5V range).

**Use Case 3**: "How much energy does a write consume?"
→ **Answer**: Module 4's charge pump analysis shows 2.18 pJ total, dominated by voltage boost circuitry.

### For Future Development

**Priority 1**: Implement ISPP write-verify loop
- Add read-back logic in Module 2
- Add voltage adjustment in Module 4
- Add convergence loop (max 10 iterations)

**Priority 2**: Create voltage-to-conductance transfer function
- Model ferroelectric polarization vs. voltage
- Link Module 4 DAC output to Module 2 conductance state
- Enable end-to-end simulation

**Priority 3**: Validate energy estimates
- Compare Module 4 estimates against published data
- Current 2.18 pJ includes peripheral overhead
- Cell-only energy should be 10-30 fJ

---

## References

### Code Locations

- **Module 2 Write**: `module2-crossbar/pkg/crossbar/array.go:71-86`
- **Module 2 Quantization**: `module2-crossbar/pkg/crossbar/array.go:88-101`
- **Module 2 IR Drop**: `module2-crossbar/pkg/crossbar/irdrop.go:71-134`
- **Module 4 DAC**: `module4-circuits/pkg/peripherals/dac.go:36-67`
- **Module 4 Charge Pump**: `module4-circuits/pkg/peripherals/chargepump.go:39-95`

### Documentation

- **Operations Guide**: `docs/peripheral-circuits/circuits.operations.md` (sections 2-3)
- **CIM Fundamentals**: `docs/peripheral-circuits/circuits.CIM-fundamentals.md` (section 3)
- **Write Physics**: `docs/peripheral-circuits/circuits.CIM-fundamentals.md:179-356`

### Physics Sources

- **Coercive Voltage**: Nature Communications 2025 (Ec = 0.6-1.5 MV/cm)
- **ISPP Programming**: TUM FeFET 2023 (40mV increments)
- **Endurance**: Nano Letters 2024 (10¹² cycles for V:HfO₂)

---

## Conclusion

Module 2 and Module 4 provide **complementary views** of the write operation:

- **Module 2** abstracts write as an instantaneous state change, focusing on array-level behavior (quantization, tracking, non-idealities like IR drop and drift).

- **Module 4** exposes the circuit-level implementation, showing how peripheral circuits (DAC, charge pump) generate the ±1.5V write voltages with realistic timing (200ns) and energy (2.18 pJ).

- **Together**, they form a complete educational toolset that teaches both high-level programming abstractions and low-level circuit design.

- **Missing link**: Ferroelectric physics (voltage → polarization → conductance) is documented but not coded, which is acceptable for visualization tools but limits research simulation.

**Bottom Line**: Both modules are needed for complete understanding. Module 2 answers "what state is stored?", Module 4 answers "how is that voltage generated?", and the documentation bridges the gap with "how does voltage create that state?".

---

**Analysis Complete**  
**Total Files Analyzed**: 8 (5 Go sources, 3 documentation files)  
**Report Files Generated**: 3 (analysis, diagram, summary)  
**Stage Duration**: ~7 minutes  
**Status**: ✅ SUCCESS
