# Crossbar Voltage Rules and Operation Voltages

**FeCIM Lattice Tools - Module 2: Voltage Reference**

> Comprehensive voltage specifications for ferroelectric crossbar operations across 0T1R (Passive), 1T1R, and 2T1R architectures.

**Scope:** Voltage values for 300K nominal operation. Timing parameters and pulse widths are documented separately in `config/physics.yaml`.

---

## Table of Contents

1. [Overview](#overview)
2. [Voltage Constants Summary](#voltage-constants-summary)
3. [Passive (0T1R) Mode](#passive-0t1r-mode)
4. [1T1R Mode](#1t1r-mode)
5. [2T1R Mode](#2t1r-mode)
6. [Voltage Biasing Schemes](#voltage-biasing-schemes)
7. [Code Mappings](#code-mappings)
8. [ASCII Diagrams](#ascii-diagrams)
9. [References](#references)
10. [Quick Reference Card](#quick-reference-card)

---

## Overview

### Purpose

This document provides the authoritative reference for all voltage values used in crossbar array operations. It covers:

- **Peripheral circuit voltages** (DAC, ADC, TIA, charge pump)
- **Operation voltages** (read, write, compute/MVM)
- **Architecture-specific biasing** (V/2 half-select, transistor control)
- **Material physics limits** (coercive field, threshold voltages)

### Voltage Hierarchy

```
┌─────────────────────────────────────────────────┐
│ Charge Pump                                     │
│   Input:  1.0V (CMOS supply)                    │
│   Output: ±1.5V (write voltage generation)      │
└────────────────┬────────────────────────────────┘
                 │
    ┌────────────┴──────────────┐
    │                           │
┌───▼────────────┐     ┌────────▼──────────┐
│ DAC            │     │ ADC                │
│  Vref: ±1.5V   │     │  Vref: 0-1.0V      │
│  Output: MVM   │     │  Input: TIA output │
└───┬────────────┘     └────────▲──────────┘
    │                           │
    │ Compute                   │ Sense
    ▼                           │
┌───────────────────────────────┴───────┐
│ Crossbar Array                        │
│  Operation voltages:                  │
│    READ:  0.1-0.5V (non-destructive)  │
│    WRITE: ±1.2-1.5V (above Vc)        │
│    MVM:   0-1.0V (DAC input range)    │
└───────────────────────────────────────┘
```

---

## Voltage Constants Summary

### Master Voltage Table

| Parameter | Value | Tolerance | Source | Verification Status |
|-----------|-------|-----------|--------|---------------------|
| **Peripheral Circuits** | | | | |
| DAC Vref High | +1.5 V | ±50 mV | `shared/peripherals/defaults.go:19` | ✅ Verified |
| DAC Vref Low | -1.5 V | ±50 mV | `shared/peripherals/defaults.go:22` | ✅ Verified |
| ADC Vref High | +1.0 V | ±20 mV | `shared/peripherals/defaults.go:31` | ✅ Verified |
| ADC Vref Low | 0.0 V | ±5 mV | `shared/peripherals/defaults.go:34` | ✅ Verified |
| Charge Pump Input | 1.0 V | ±50 mV | `module4-circuits/pkg/peripherals/chargepump.go:22` | ✅ Verified |
| Charge Pump Output | 1.5 V | ±100 mV | `module4-circuits/pkg/peripherals/chargepump.go:23` | ✅ Verified |
| TIA Max Output | 1.0 V | ±50 mV | `module4-circuits/pkg/peripherals/tia.go:26` | ✅ Verified |
| **Physics Parameters** | | | | |
| Coercive Field (Ec) | 0.6-1.5 MV/cm | Material-dependent | `config/physics.yaml` (Nature Commun. 2025) | ✅ Peer-reviewed |
| Film Thickness | 10 nm | ±1 nm | `config/physics.yaml` | ✅ Standard |
| Coercive Voltage (Vc) | 0.6-1.5 V | Derived: Vc = Ec × thickness | Calculated from Ec | ⚠️ Estimated |
| **Operation Voltages** | | | | |
| Read Voltage | 0.1-0.5 V | <Vc | `module4-circuits/pkg/peripherals/analysis.go:249` | ✅ Verified |
| Write Voltage (Set) | +1.2-1.5 V | >Vc with margin | Derived from DAC range | ⚠️ Estimated |
| Write Voltage (Erase) | -1.2-1.5 V | Negative polarity | Derived from DAC range | ⚠️ Estimated |
| MVM Input Range | 0.0-1.0 V | DAC output → array | ADC Vref range | ✅ Verified |
| Half-Select (V/2) | 0.75 V | Vwrite/2 (0T1R only) | Calculated: 1.5V / 2 | ⚠️ Estimated |
| **Transistor Control (1T1R/2T1R)** | | | | |
| WL HIGH (ON) | 1.0 V | VDD (logic high) | Standard CMOS | ✅ Standard |
| WL LOW (OFF) | 0.0 V | VSS (logic low) | Standard CMOS | ✅ Standard |
| Source Line (SL) | 0.0 V | Typically grounded | Standard practice | ✅ Standard |

### Key Observations

**Verified Values:**
- All peripheral circuit voltages are hard-coded in source files.
- Physics parameters (Ec) are peer-reviewed and material-specific.
- Transistor control voltages follow standard CMOS logic levels.

**Estimated Values:**
- Write voltages are **inferred** from DAC range (±1.5V) but not explicitly set in code.
- Coercive voltage (Vc) is **calculated** from Ec and thickness, not directly measured.
- Half-select voltage is **derived** as Vwrite/2 for passive arrays.

**Recommended Values (300K Operation):**
- Read: **0.2V** (safe margin below Vc)
- Write: **±1.5V** (maximum DAC range for full switching)
- MVM: **0.0-1.0V** (matches ADC input range)

---

## Passive (0T1R) Mode

### 3.1 Read Operation

**Voltage Configuration:**

```
           BL (Bit Line)
            │ Sense
            ↓
WL ─────────●────────────
(Vread)     │
           ┌┴┐
           │R│  ← FeFET (no transistor)
           └┬┘
            │
           GND
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL Voltage** | 0.1-0.5 V | Apply to word line | Below Vc (non-destructive) |
| **BL Voltage** | Floating → TIA | Sense current | Voltage develops from I×R_wire |
| **Unselected WLs** | 0 V | Ground | Minimize sneak paths |
| **Unselected BLs** | 0 V | Ground | Current sink |

**Read Current:**
```
I_read = G_cell × V_read
       = (10-100 µS) × 0.2V
       = 2-20 µA
```

**Constraints:**
- V_read < Vc (0.6-1.5V) → Non-destructive read
- Recommended: **0.2V** for safe margin
- Higher V_read → better SNR but risk of disturb

### 3.2 Write Operation

**Voltage Configuration (V/2 Half-Select Scheme):**

```
Target cell: (Row 2, Col 2)

          BL0    BL1    BL2    BL3
           │      │      │      │
           ↓      ↓      ↓      ↓
WL0 ── 0V ─●──────●──────●──────●──
           │      │      │      │
WL1 ── 0V ─●──────●──────●──────●──
           │      │      │      │
WL2 ─+0.75V●──────●──────●══════●── Vwrite = +1.5V
           │      │      ║      │
WL3 ── 0V ─●──────●──────●──────●──
           │      │      ║      │
          0V     0V   +0.75V   0V
                        ║
                     TARGET
                   ΔV = +1.5V

Half-selected cells experience V/2 = 0.75V
```

| Parameter | Value | Target | Half-Selected |
|-----------|-------|--------|---------------|
| **Selected WL** | +1.5 V (Set) / -1.5 V (Erase) | Full voltage | Applied to entire row |
| **Selected BL** | -0.75 V (Set) / +0.75 V (Erase) | Creates voltage difference | Applied to entire column |
| **Unselected WLs** | 0 V | No bias | Grounded |
| **Unselected BLs** | 0 V | No bias | Grounded |
| **Effective V (target)** | ±1.5 V | WL - BL | Above Vc → switching |
| **Effective V (half-select)** | ±0.75 V | V/2 | Below Vc → minimal disturb |

**Write Disturb:**
- Half-selected cells see V/2 = 0.75V
- If Vc = 1.2V, then V/2 = 0.625 × Vc (safe margin)
- Repeated half-selects cause cumulative drift (modeled in `HalfSelectConfig`)

### 3.2.1 Multi-Level Write Voltage (30 Levels)

**Critical Insight:** Write voltage is NOT a single fixed value - it varies per target analog level and requires iterative program-verify loops.

#### Per-Level Voltage Calibration

Each of the 30 analog levels requires a different E-field to achieve:

**Calibration Arrays:**
```
calibrationUp[30]   : E-field values for ascending polarization path
calibrationDown[30] : E-field values for descending polarization path
```

**Binary Search Calibration (15 iterations per level):**
```
File: module1-hysteresis/pkg/gui/simulation.go:1564-1750

For each level 0-29:
    1. Initial estimate: Linear interpolation between E_min and E_max
    2. Apply field, read Preisach model polarization
    3. If overshoot → reduce field, retry
    4. If undershoot → increase field, retry
    5. Converge to ±0.5% tolerance
```

**Why Not a Lookup Table?**
- Hysteresis path-dependence: Ascending ≠ Descending branches
- Temperature variation: Field requirements change with T
- Aging/drift: Cells evolve over 10¹² cycles
- Solution: **Adaptive runtime calibration**

#### Hysteresis Path-Dependence

**Preisach Model Governs Switching:**

```
File: module1-hysteresis/pkg/ferroelectric/preisach_advanced.go

           P (Polarization)
           ↑
      Psat │     ╱╲  Ascending branch
           │    ╱  ╲
           │   ╱    ╲
      Pr   ├──●      ╲
           │           ╲
           │            ╲
   ───────┼─────────────●─────→ E (Field)
          │            ╱
     -Pr  ├──────────●╱
          │        ╱
          │       ╱  Descending branch
    -Psat │      ╱
          │
```

**Voltage Implications:**
- **Ascending path** (0 → 29): Requires higher field per level
- **Descending path** (29 → 0): Requires lower field per level
- **Overshoot handling**: If target level exceeded:
  1. RESET cell to opposite saturation (-Psat)
  2. Return to known state (Preisach hysteron reset)
  3. Retry with adjusted voltage

**Code Reference:**
```go
// Preisach model tracks every hysteron's state
// Overshooting requires full RESET to clear hysteretic memory
if overshoot {
    applyReset()  // Saturate to -Psat
    clearHysterons()
    retryFromKnownState()
}
```

#### Program-Verify Loop (ISPP)

**Incremental Step Pulse Programming** - Industry-standard approach for multi-level cells:

```
File: module4-circuits/pkg/gui/tab_unified.go:1188-1279

Write Sequence:
┌─────────────────────────────────────────┐
│ 1. WRITE: Apply calibrated voltage     │
│ 2. READ:  Sense actual conductance     │
│ 3. VERIFY: Compare to target level     │
│ 4. ADJUST: Δ = target - actual         │
│ 5. RETRY:  If |Δ| > tolerance          │
└─────────────────────────────────────────┘

Max 5 iterations per cell
Tolerance: ±0.5 levels (±1.67% of full range)
```

**Voltage Adjustment Strategy:**
```
Initial V = calibrationUp[level]  // or calibrationDown[level]

Loop up to 5 times:
    Apply V → Read actual_level

    If actual_level < target:
        V += step_size  (typically 0.05V increments)
    Else if actual_level > target:
        V -= step_size
    Else:
        SUCCESS (within tolerance)

    If iteration > 5:
        WARNING: Cell may be defective or drifted
```

**Why ISPP is Essential:**
- Cell-to-cell variation: ±10% variation in Ec across array
- Cycle-dependent drift: Switching field evolves over 10⁹-10¹² cycles
- IR drop effects: Cells at array edges see different voltages
- Temperature gradients: Local heating changes Ec

#### Voltage-Level Relationship

**Non-Linear Mapping Due to Hysteresis:**

```
Voltage (V)  Level  Notes
   ↑         29     Psat (saturation)
   │         28
  1.5V ─┐    27     Near saturation (steep slope)
        │    26
  1.4V ─┤    25
        │    24
  1.3V ─┤    23     Linear region (easier to hit)
        │    ...
  1.2V ─┤    15     Vc threshold (steepest slope)
        │    14
  1.1V ─┤    13     Sub-Vc region (minimal switching)
        │    ...
  1.0V ─┤     5
        │     ...
  0.8V ─┘     0     Near zero polarization
   ↓
```

**ASCII Diagram: Voltage vs Level (Ascending Path)**

```
Level vs Required Voltage (HZO, 10nm, 300K)

Level
 29 ────────────────────────────────────────────●  1.50V  Saturation
 28 ────────────────────────────────────────●     1.48V
 27 ──────────────────────────────────────●       1.46V
 26 ────────────────────────────────────●         1.43V
 25 ──────────────────────────────────●           1.40V
 24 ────────────────────────────────●             1.38V  } Steep
 23 ──────────────────────────────●               1.35V  } slope
 22 ────────────────────────────●                 1.32V  } (hard
 21 ──────────────────────────●                   1.29V  } to hit)
 20 ────────────────────────●                     1.26V
 19 ──────────────────────●                       1.24V
 18 ────────────────────●                         1.22V  } Near Ec
 17 ──────────────────●                           1.20V  } (easiest)
 16 ────────────────●                             1.18V
 15 ──────────────●                               1.16V  } Linear
 14 ────────────●                                 1.14V  } region
 13 ──────────●                                   1.12V
 12 ────────●                                     1.10V
 11 ──────●                                       1.08V
 10 ────●                                         1.06V
  9 ──●                                           1.04V  } Sub-Vc
  8 ●                                             1.02V  } (slow)
  7                                               1.00V
  ...
  0 ●                                             0.80V  Zero P
     └────┴────┴────┴────┴────┴────┴────┴────┘
      0.8  1.0  1.2  1.4  1.6  1.8  2.0V

Voltage →

Key observations:
  • Levels 15-18 (near Vc): LINEAR, easiest to target
  • Levels 0-8 (sub-Vc): FLAT, requires fine voltage control
  • Levels 24-29 (saturation): STEEP, prone to overshoot
  • Descending path: Different curve (lower voltages)
```

**Refinement Sources:**
1. **Preisach calibration**: Accounts for hysteresis path
2. **Runtime feedback**: ISPP loop measures actual response
3. **Temperature interpolation**: Vc(T) from physics.yaml
4. **Drift compensation**: Tracks cumulative cycle count

#### 4-Phase Write Sequence

**Full Write Operation (one cell, one level):**

```
Phase 0: RESET (100ns pulse)
┌──────────────────────────────────┐
│ Apply -V_sat (opposite polarity) │
│ Purpose: Saturate to -Psat       │
│ Result: Known starting state     │
└──────────────────────────────────┘
         ↓
Phase 1: HOLD_RESET (50ns)
┌──────────────────────────────────┐
│ Return WL/BL to 0V               │
│ Purpose: Zero field, P persists  │
│ Result: Cell at -Psat, stable    │
└──────────────────────────────────┘
         ↓
Phase 2: WRITE (Program-Verify Loop)
┌──────────────────────────────────┐
│ Apply calibrated V for target    │
│ Purpose: Switch to +P_target     │
│ Result: Cell at desired level    │
│ Iterations: 1-5 (ISPP)           │
└──────────────────────────────────┘
         ↓
Phase 3: HOLD_WRITE (50ns)
┌──────────────────────────────────┐
│ Return WL/BL to 0V               │
│ Purpose: Zero field, P persists  │
│ Result: Non-volatile storage     │
└──────────────────────────────────┘
```

**Timing (from config/physics.yaml):**
```yaml
pulse_widths:
    reset_ns: 100       # Phase 0: RESET
    hold_reset_ns: 50   # Phase 1: HOLD_RESET
    write_ns: 200       # Phase 2: WRITE (single pulse)
    hold_write_ns: 50   # Phase 3: HOLD_WRITE

Total per cell: 400ns + ISPP overhead (5× worst case = 2µs)
```

**Why 4 Phases?**
- **Phase 0 (RESET)**: Erases hysteretic memory (Preisach hysteron reset)
- **Phase 1 (HOLD_RESET)**: Allows domain walls to stabilize
- **Phase 2 (WRITE)**: Applies calibrated field for target level
- **Phase 3 (HOLD_WRITE)**: Ensures polarization persists after field removal

**Code Flow:**
```go
// Simplified from tab_unified.go:1188-1279
func writeCellToLevel(row, col, level int) error {
    // Phase 0: RESET
    applyVoltage(row, col, -V_sat, RESET_PULSE)

    // Phase 1: HOLD_RESET
    applyVoltage(row, col, 0, HOLD_RESET)

    // Phase 2: WRITE (ISPP loop)
    V := calibrationUp[level]
    for iter := 0; iter < 5; iter++ {
        applyVoltage(row, col, V, WRITE_PULSE)
        actual := readCell(row, col)

        if abs(actual - level) < 0.5 {
            break  // SUCCESS
        }
        V += stepSize * sign(level - actual)
    }

    // Phase 3: HOLD_WRITE
    applyVoltage(row, col, 0, HOLD_WRITE)

    return verify(row, col, level)
}
```

**Key Takeaway:** Write voltage is a **dynamic, adaptive parameter**, not a static constant. The program-verify loop compensates for physics complexity that cannot be captured in a simple lookup table.

### 3.3 Compute (MVM)

**Voltage Configuration:**

```
Input vector: [x0, x1, x2, x3]
Applied to BLs via DAC

          BL0    BL1    BL2    BL3
          x0     x1     x2     x3
           │      │      │      │
           ↓      ↓      ↓      ↓
WL0 ── 1V ─●──────●──────●──────●── I0 → ADC
           │      │      │      │
WL1 ── 1V ─●──────●──────●──────●── I1 → ADC
           │      │      │      │
WL2 ── 1V ─●──────●──────●──────●── I2 → ADC
           │      │      │      │
WL3 ── 1V ─●──────●──────●──────●── I3 → ADC

All WLs active (passive mode: no transistor gating)
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL Voltage** | 1.0 V | Bias for current flow | All WLs always active (0T1R) |
| **BL Voltage (DAC)** | 0.0-1.0 V | Input vector encoding | DAC quantized (5-8 bits typical) |
| **Output Current** | I = Σ(G_ij × V_j) | Column sum | ADC quantizes to digital |

**MVM Equation:**
```
I_i = Σ_j (W_ij × x_j)
    = Σ_j (G_ij × V_j)
    = G_i0×V_0 + G_i1×V_1 + ... + G_in×V_n
```

**Voltage Ranges:**
- Input (DAC): 0.0-1.0V (matches ADC Vref High)
- Output (TIA): 0.0-1.0V (TIA max output)
- WL bias: 1.0V (constant, all rows active)

---

## 1T1R Mode

### 4.1 Read Operation

**Voltage Configuration:**

```
           BL (Bit Line)
            │ Sense
            ↓
WL ─────────┬────────── Gate = HIGH (1.0V)
            │
       ┌────┴────┐
  SL ──┤  NMOS   ├── Drain
       │         │
       └────┬────┘
            │ Source
           ┌┴┐
           │R│  ← FeFET
           └┬┘
            │
           GND
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL Voltage (selected)** | 1.0 V | Turn ON transistor | Logic HIGH |
| **WL Voltage (unselected)** | 0.0 V | Turn OFF transistor | R_off ~1 GΩ isolation |
| **BL Voltage** | 0.2 V | Read voltage | Applied through transistor |
| **SL Voltage** | 0.0 V | Source line (ground) | Transistor source terminal |

**Transistor States:**
```
WL = HIGH (1.0V): R_on ≈ 1 kΩ   → Cell accessible
WL = LOW  (0.0V): R_off ≈ 1 GΩ  → Cell isolated (1000× sneak reduction)
```

**Read Current Path:**
```
BL (0.2V) → Transistor (ON) → FeFET → GND
I_read = G_cell × V_read / (1 + R_on/R_FeFET)
       ≈ G_cell × V_read (R_on << R_FeFET typically)
```

### 4.2 Write Operation

**Voltage Configuration:**

```
Selected cell: WL HIGH, full voltage applied
Unselected cells: WL LOW, isolated

          BL0    BL1    BL2
           │      │      │
WL0 ─ 0V ─┬──────┬──────┬──── Transistors OFF
          │      │      │
WL1 ─ 1V ─┬──────┬══════┬──── Transistor ON (target row)
          │      ║      │
          0V   +1.5V    0V
                ║
             TARGET
```

| Parameter | Value | Target Cell | Unselected Cells |
|-----------|-------|-------------|------------------|
| **Selected WL** | 1.0 V | Transistor ON | N/A |
| **Selected BL** | ±1.5 V | Write voltage | Applied to column |
| **Unselected WLs** | 0.0 V | Transistors OFF | Isolated |
| **Unselected BLs** | 0.0 V | Ground | No voltage |
| **SL** | 0.0 V | Ground | All cells |

**No V/2 Scheme Required:**
- Transistor isolation eliminates need for half-select biasing
- Only selected cell sees full write voltage
- Unselected cells isolated by OFF transistors (R_off ~1 GΩ)

**Write Disturb:**
- Negligible (<0.01% vs 0T1R)
- Transistor OFF-state blocks voltage stress

### 4.3 Compute (MVM)

**Voltage Configuration:**

```
User-controlled WL activation (unlike 0T1R)

          BL0    BL1    BL2
          x0     x1     x2
           │      │      │
WL0 ─ 1V ─┬──────┬──────┬──── Active (ON)
          │      │      │
WL1 ─ 0V ─┬──────┬──────┬──── Inactive (OFF)
          │      │      │
WL2 ─ 1V ─┬──────┬──────┬──── Active (ON)
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL Voltage (active rows)** | 1.0 V | Turn ON transistors | User-selectable rows |
| **WL Voltage (inactive rows)** | 0.0 V | Turn OFF transistors | Isolated from computation |
| **BL Voltage (DAC)** | 0.0-1.0 V | Input vector | DAC quantized |
| **SL Voltage** | 0.0 V | Ground | All cells |

**Key Difference from 0T1R:**
- In 0T1R: **All WLs always active** (no gating)
- In 1T1R: **User controls WL activation** (row-selective MVM)

**Selective MVM:**
```
Output only from active rows:
I_i = Σ_j (G_ij × V_j)  if WL_i = HIGH
    = 0                 if WL_i = LOW (transistor blocks current)
```

---

## 2T1R Mode

### 5.1 Read Operation

**Voltage Configuration:**

```
Separate read and write paths

           BL (Bit Line)
            │ Sense
            ↓
WL_write ───┬────────── Gate = LOW (0.0V - isolated)
            │
       ┌────┴────┐
       │  Write  │
       │  NMOS   │
       └────┬────┘
            │
           ┌┴┐
           │R│  ← FeFET
           └┬┘
            │
       ┌────┴────┐
       │  Read   │
       │  NMOS   │
       └────┬────┘
            │
WL_read ────┴────────── Gate = HIGH (1.0V - active)
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL_read (selected)** | 1.0 V | Turn ON read transistor | Read path active |
| **WL_write (selected)** | 0.0 V | Turn OFF write transistor | Write path isolated |
| **BL Voltage** | 0.2 V | Read voltage | Non-destructive |
| **SL Voltage** | 0.0 V | Ground | Common to both transistors |

**Read Path Isolation:**
- Write transistor OFF → No voltage stress on write circuitry
- Complete path isolation → Ultra-low disturb

### 5.2 Write Operation

**Voltage Configuration:**

```
Write path active, read path isolated

WL_write ───┬────────── Gate = HIGH (1.0V - active)
            │
       ┌────┴────┐
       │  Write  │
       │  NMOS   │   ← Write path
       └────┬────┘
            │
           ┌┴┐
           │R│  ← FeFET
           └┬┘
            │
       ┌────┴────┐
       │  Read   │
       │  NMOS   │   ← Read path
       └────┬────┘
            │
WL_read ────┴────────── Gate = LOW (0.0V - isolated)
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL_write (selected)** | 1.0 V | Turn ON write transistor | Write path active |
| **WL_read (selected)** | 0.0 V | Turn OFF read transistor | Read path isolated |
| **BL Voltage** | ±1.5 V | Write voltage | Full voltage to FeFET |
| **SL Voltage** | 0.0 V | Ground | Common terminal |

**Write Path Isolation:**
- Read transistor OFF → Read circuitry protected
- Independent voltage optimization for read/write

### 5.3 Compute (MVM)

**Voltage Configuration:**

```
MVM uses read path (non-destructive)

          BL0    BL1    BL2
          x0     x1     x2
           │      │      │
WL_read0 ─┬──────┬──────┬──── 1.0V (read path ON)
          │      │      │
WL_write0─┬──────┬──────┬──── 0.0V (write path OFF)
          │      │      │
WL_read1 ─┬──────┬──────┬──── 0.0V (inactive)
          │      │      │
WL_write1─┬──────┬──────┬──── 0.0V (inactive)
```

| Parameter | Value | Purpose | Notes |
|-----------|-------|---------|-------|
| **WL_read (active rows)** | 1.0 V | Enable read path | Row-selective |
| **WL_write (all rows)** | 0.0 V | Disable write path | Isolated during MVM |
| **BL Voltage (DAC)** | 0.0-1.0 V | Input vector | Standard MVM range |
| **SL Voltage** | 0.0 V | Ground | Common terminal |

**Ultra-Low Disturb MVM:**
- Write path completely isolated (WL_write = LOW)
- No write stress during compute operations
- Ideal for high-precision analog computing

---

## Voltage Biasing Schemes

### 6.1 V/2 Half-Select (0T1R Only)

**Purpose:** Minimize write disturb in passive crossbar arrays.

**Principle:**
```
Target cell receives:   ΔV = V_WL - V_BL = (+V) - (-V/2) = 1.5V
Half-selected cells:    ΔV = V_WL - 0    = (+V/2)       = 0.75V
                        ΔV = 0 - (-V/2)  = (+V/2)       = 0.75V
```

**Voltage Allocation:**

| Cell Type | WL Voltage | BL Voltage | Effective ΔV | Result |
|-----------|------------|------------|--------------|--------|
| **Target** | +1.5 V (Set) / -1.5 V (Erase) | -0.75 V (Set) / +0.75 V (Erase) | ±1.5 V | Full switching |
| **Same row** | +1.5 V | 0 V | +0.75 V | Half-select disturb |
| **Same column** | 0 V | -0.75 V | +0.75 V | Half-select disturb |
| **Diagonal** | 0 V | 0 V | 0 V | No disturb |

**Half-Select Disturb:**
- V/2 = 0.75V < Vc (1.2V typical)
- Cumulative effect after many writes
- Modeled in code: `HalfSelectConfig` tracks exposure count

### 6.2 V/3 Scheme (Advanced)

**Purpose:** Further reduce disturb in very large passive arrays.

**Principle:**
```
Divide write voltage into thirds:
WL options: {+2V/3, +V/3, 0, -V/3, -2V/3}
BL options: {+2V/3, +V/3, 0, -V/3, -2V/3}

Target cell: ΔV = (+2V/3) - (-V/3) = V (full switching)
```

**Not Currently Implemented:**
- More complex driver circuitry
- Trade-off: reduced disturb vs. increased hardware complexity
- Possible future enhancement

### 6.3 Grounding Schemes

**Standard Grounding (0T1R):**
```
Unselected WLs: 0V
Unselected BLs: 0V

Simple but allows sneak paths.
```

**1T1R Grounding:**
```
All SLs: 0V (source line grounded)
Unselected WLs: 0V (transistors OFF)

Transistor isolation eliminates need for complex biasing.
```

**2T1R Grounding:**
```
SL: 0V (common ground)
Inactive path: WL_read or WL_write = 0V

Complete path isolation.
```

---

## Code Mappings

### 7.1 Peripheral Circuit Constants

**File:** `shared/peripherals/defaults.go`

```go
// DAC reference voltage constants
const (
    DACVrefHigh = 1.5   // Line 19: +1.5V for write operations
    DACVrefLow = -1.5   // Line 22: -1.5V for write operations
)

// ADC reference voltage constants
const (
    ADCVrefHigh = 1.0   // Line 31: 1.0V for read operations
    ADCVrefLow = 0.0    // Line 34: 0.0V (ground reference)
)
```

**Usage:**
```go
dac := DefaultDAC()
voltage := dac.Convert(level)  // Maps level 0-29 to -1.5V to +1.5V
```

### 7.2 Charge Pump Configuration

**File:** `module4-circuits/pkg/peripherals/chargepump.go`

```go
func DefaultChargePump() *ChargePump {
    return &ChargePump{
        InputVoltage:  1.0,  // Line 22: 1V CMOS supply
        OutputVoltage: 1.5,  // Line 23: 1.5V write voltage
        Stages:        2,    // 2-stage Dickson pump
        // ...
    }
}
```

**Boost Factor:**
```go
// Vout = (N+1) × Vin for ideal Dickson pump
// Actual: (N+1) × Vin - N × Vth
// For N=2: 3 × 1.0V - 2 × 0.3V = 2.4V (ideal)
// After losses: ~1.5V (actual)
```

### 7.3 TIA Output Range

**File:** `module4-circuits/pkg/peripherals/tia.go`

```go
func DefaultTIA() *TIA {
    return &TIA{
        Gain:             10e3,   // 10 kΩ transimpedance
        MaxOutputVoltage: 1.0,    // Line 26: 1V max output
        // ...
    }
}
```

**Current-to-Voltage Conversion:**
```go
V_out = I_in × Gain
      = (2-20 µA) × 10 kΩ
      = 0.02-0.2V (typical read current range)
```

### 7.4 Read Voltage (Code Reference)

**File:** `module4-circuits/pkg/peripherals/analysis.go`

```go
// Line 248-249: Read voltage explicitly set
Vread := 0.1  // 0.1V for non-destructive read
```

**Note:** This is the only **explicit** read voltage in code. Other operations derive voltages from DAC/ADC ranges.

### 7.5 Physics Parameters

**File:** `config/physics.yaml`

```yaml
default_hzo:
    ec_v_m: 1.2e8               # Coercive field = 1.2 MV/cm
    thickness_m: 10.0e-9        # Film thickness = 10 nm

# Derived coercive voltage:
# Vc = Ec × thickness
#    = 1.2e8 V/m × 10e-9 m
#    = 1.2 V
```

### 7.6 Half-Select Voltage (Derived)

**Not explicitly coded, but calculated:**

```python
# From DAC Vref High
V_write = 1.5  # V
V_half_select = V_write / 2.0
              = 0.75  # V

# Safety margin check
V_half_select / Vc = 0.75 / 1.2 = 0.625 (62.5% of Vc)
# Safe: <100% means no unintended switching
```

### 7.7 Architecture Mode Detection

**File:** `module4-circuits/pkg/gui/device_state.go`

```go
// Passive mode enforcement (0T1R)
func (ds *DeviceState) SetPassiveMode(passive bool) {
    ds.isPassive = passive
    if passive {
        // Force all WLs on (no transistor gating)
        ds.wlMode = WLAll
        for i := range ds.activeRows {
            ds.activeRows[i] = true  // All WLs = HIGH
        }
    }
}
```

**1T1R Mode:**
```go
// User can control individual WLs
func (ds *DeviceState) SetWLSingle(row int) {
    if ds.isPassive {
        return  // Ignored in passive mode
    }
    // ... enable only selected row
}
```

---

## ASCII Diagrams

### 8.1 Voltage Rails Overview

```
Voltage Levels in FeCIM Crossbar System

    +1.5V ────────────────────────────── DAC Vref High, Write voltage (Set)
                                         Charge Pump Output (positive)

    +1.0V ────────────────────────────── ADC Vref High, TIA max output
            │                            WL HIGH (transistor ON)
            │                            Charge Pump Input (CMOS VDD)
            │
    +0.75V ────────────────────────────── Half-select voltage (V/2, 0T1R)
            │
            │
    +0.5V ────────────────────────────── Read voltage (upper limit)
            │
            │
    +0.2V ────────────────────────────── Read voltage (typical)
            │
            │
     0.0V ══════════════════════════════ Ground (GND)
            │                            ADC Vref Low
            │                            WL LOW (transistor OFF)
            │                            SL (source line)
            │
            │
    -0.75V ────────────────────────────── Half-select voltage (V/2, 0T1R)
            │
            │
    -1.5V ────────────────────────────── DAC Vref Low, Write voltage (Erase)
                                         Charge Pump Output (negative)

Legend:
  ──── Available voltage level
  ════ Reference ground
```

### 8.2 0T1R Half-Select Biasing (3×3 Array)

```
WRITE Operation: Target cell (1,1) - SET to level 29

Voltage Assignment (V/2 scheme):

         BL0      BL1      BL2
          │        │        │
     0V   ↓   -0.75V   ↓   0V   ↓
          │        │        │
  WL0  ───●────────●────────●───  0V
  0V      │        │        │
         ┌┴┐      ┌┴┐      ┌┴┐
         │ │      │ │      │ │   ΔV = 0V (diagonal)
         └┬┘      └┬┘      └┬┘
          │        │        │
  WL1  ───●────────●════════●───  +1.5V
+1.5V     │        ║        │
         ┌┴┐      ┌║┐      ┌┴┐
         │ │      ║ ║      │ │   ΔV = +1.5V (target)
         └┬┘      ║ ║      └┬┘   ΔV = +0.75V (half-select)
          │        ║        │
  WL2  ───●────────║────────●───  0V
  0V      │        ║        │
         ┌┴┐      ┌║┐      ┌┴┐
         │ │      ║ ║      │ │   ΔV = +0.75V (half-select)
         └┬┘      ║ ║      └┬┘
          │        ║        │
         GND      ═╩═      GND
                TARGET

Cell Voltages:
┌─────────┬────────┬──────────┬──────────┐
│ Cell    │ WL     │ BL       │ ΔV       │
├─────────┼────────┼──────────┼──────────┤
│ (1,1)   │ +1.5V  │ -0.75V   │ +1.5V ✓  │  Full switching
│ (1,0)   │ +1.5V  │   0V     │ +0.75V   │  Half-select (same row)
│ (1,2)   │ +1.5V  │   0V     │ +0.75V   │  Half-select (same row)
│ (0,1)   │   0V   │ -0.75V   │ +0.75V   │  Half-select (same col)
│ (2,1)   │   0V   │ -0.75V   │ +0.75V   │  Half-select (same col)
│ (0,0)   │   0V   │   0V     │   0V     │  No disturb (diagonal)
│ (2,2)   │   0V   │   0V     │   0V     │  No disturb (diagonal)
└─────────┴────────┴──────────┴──────────┘

Key:
  ● = FeFET cell (passive, no transistor)
  ═ = Target cell with full switching voltage
  ║ = Current path to target
```

### 8.3 1T1R Transistor Isolation (Single Cell)

```
1T1R Cell Structure and Voltage Control

WRITE Operation (SET):

   WL ──────────┬────────────── Gate = 1.0V (HIGH)
                │                      ↓
                │              Turn transistor ON
           ┌────┴────┐                 ↓
      SL ──┤  NMOS   ├── Drain        │
    (0V)   │  W/L    │                │
           └────┬────┘         R_on ≈ 1 kΩ
                │ Source              │
               ┌┴┐                    ↓
               │R│  ← FeFET     Charge flow
               │ │    (HZO)           │
               └┬┘                    ↓
                │                     │
   BL ──────────┴─────────── +1.5V (Set) or -1.5V (Erase)


READ Operation:

   WL ──────────┬────────────── Gate = 1.0V (HIGH)
                │
           ┌────┴────┐
      SL ──┤  NMOS   ├── Drain
    (0V)   │   ON    │
           └────┬────┘
                │
               ┌┴┐
               │R│  ← I_read = G × V_read
               │ │              │
               └┬┘              │
                │               ↓
   BL ──────────┴─────────── 0.2V (Read)
                │
                ↓
              TIA/ADC


UNSELECTED Cell (isolated):

   WL ──────────┬────────────── Gate = 0.0V (LOW)
                │                      ↓
                │              Transistor OFF
           ┌────┴────┐                 ↓
      SL ──┤  NMOS   ├── Drain        │
    (0V)   │  OFF    │                │
           └────┬────┘         R_off ≈ 1 GΩ
                │                     │
               ┌┴┐              ~1,000,000×
               │R│              isolation!
               │ │                    │
               └┬┘                    ↓
                │              I_leak ≈ 0
   BL ──────────┴─────────── Any voltage (blocked)


Transistor States:
┌──────────────┬─────────┬──────────┬────────────────┐
│ WL Voltage   │ State   │ R_ch     │ Cell Access    │
├──────────────┼─────────┼──────────┼────────────────┤
│ 1.0V (HIGH)  │ ON      │ ~1 kΩ    │ Accessible     │
│ 0.0V (LOW)   │ OFF     │ ~1 GΩ    │ Isolated       │
└──────────────┴─────────┴──────────┴────────────────┘

Sneak Path Suppression:
  0T1R: Sneak/Signal ≈ 2:1  (200%)
  1T1R: Sneak/Signal ≈ 0.002:1 (0.2%)  → 1000× improvement
```

---

## References

### 9.1 Cross-Links to Existing Documentation

| Document | Section | Relevant Content |
|----------|---------|------------------|
| **ARCHITECTURES.md** | §3-5 | 0T1R, 1T1R, 2T1R voltage configurations |
| **PHYSICS.md** | §2-3 | Conductance models, MVM operation |
| **circuits.operations.md** | §2 WRITE | V/2 scheme, half-select disturb |
| **circuits.peripherals.md** | §1-4 | DAC, ADC, TIA, charge pump specifications |

### 9.2 Code File References

| File | Lines | Content |
|------|-------|---------|
| `shared/peripherals/defaults.go` | 19, 22, 31, 34 | DAC/ADC Vref constants |
| `module4-circuits/pkg/peripherals/dac.go` | 23-24 | DAC voltage range |
| `module4-circuits/pkg/peripherals/adc.go` | 32-33 | ADC voltage range |
| `module4-circuits/pkg/peripherals/chargepump.go` | 22-23 | Charge pump I/O voltages |
| `module4-circuits/pkg/peripherals/tia.go` | 26 | TIA max output voltage |
| `module4-circuits/pkg/peripherals/analysis.go` | 249 | Explicit read voltage (0.1V) |
| `config/physics.yaml` | 87, 95 | Ec and thickness (Vc derivation) |
| `module2-crossbar/pkg/crossbar/sneakpath.go` | 219-220 | Half-select voltage struct |
| `module4-circuits/pkg/gui/device_state.go` | 273-282 | Passive mode WL enforcement |

### 9.3 Peer-Reviewed Physics References

| Parameter | Source | DOI/Reference |
|-----------|--------|---------------|
| Ec: 0.6-1.5 MV/cm | Nature Commun. 2025 | doi:10.1038/s41467-025-61758-2 |
| Ec: 0.6-1.5 MV/cm | Nano Letters 2024 | Various FeFET papers |
| Pr: 15-34 µC/cm² | Nature Commun. 2025 | doi:10.1038/s41467-025-61758-2 |
| Pr: 75 µC/cm² @ 4K | Adv. Elec. Mat. 2024 | doi:10.1002/aelm.202300879 |

---

## Quick Reference Card

```
╔════════════════════════════════════════════════════════════════════════════╗
║                     FeCIM CROSSBAR VOLTAGE QUICK REFERENCE                 ║
╠════════════════════════════════════════════════════════════════════════════╣
║                                                                            ║
║  PERIPHERAL VOLTAGES                                                       ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   DAC Vref:      ±1.5V          ADC Vref:      0-1.0V                     ║
║   Charge Pump:   1.0V → 1.5V    TIA Output:    0-1.0V                     ║
║                                                                            ║
║  PHYSICS LIMITS (300K, 10nm HZO)                                           ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   Coercive Field (Ec):    0.6-1.5 MV/cm  (peer-reviewed)                  ║
║   Coercive Voltage (Vc):  0.6-1.5 V      (Ec × 10nm)                      ║
║                                                                            ║
║  OPERATION VOLTAGES                                                        ║
║  ───────────────────────────────────────────────────────────────────────   ║
║                           0T1R          1T1R          2T1R                ║
║   READ:                  0.1-0.5V      0.1-0.5V      0.1-0.5V             ║
║   WRITE (Set):           +1.2-1.5V     +1.2-1.5V     +1.2-1.5V            ║
║   WRITE (Erase):         -1.2-1.5V     -1.2-1.5V     -1.2-1.5V            ║
║   MVM Input:             0.0-1.0V      0.0-1.0V      0.0-1.0V             ║
║                                                                            ║
║   NOTE: Write voltages are LEVEL-DEPENDENT (30 levels)                    ║
║         - Each level requires unique voltage (calibrationUp/Down arrays)  ║
║         - Program-Verify loop (ISPP): 1-5 iterations per cell            ║
║         - See Section 3.2.1 for multi-level complexity                    ║
║                                                                            ║
║  ARCHITECTURE-SPECIFIC VOLTAGES                                            ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   Half-Select (0T1R):    ±0.75V        N/A           N/A                  ║
║   WL HIGH (1T1R/2T1R):   N/A           1.0V          1.0V                 ║
║   WL LOW (1T1R/2T1R):    N/A           0.0V          0.0V                 ║
║   SL (1T1R/2T1R):        N/A           0.0V          0.0V                 ║
║                                                                            ║
║  RECOMMENDED VALUES (Safety Margins)                                       ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   READ:   0.2V    (well below Vc, non-destructive)                        ║
║   WRITE:  ±1.5V   (maximum DAC range, ensures switching)                  ║
║           NOTE: Actual write V varies by target level (30 states)         ║
║                 Requires program-verify loop (ISPP, §3.2.1)               ║
║   MVM:    0-1.0V  (matches ADC input range)                               ║
║                                                                            ║
║  CODE CONSTANTS (source file : line number)                                ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   shared/peripherals/defaults.go : 19    DACVrefHigh = 1.5                ║
║   shared/peripherals/defaults.go : 22    DACVrefLow = -1.5                ║
║   shared/peripherals/defaults.go : 31    ADCVrefHigh = 1.0                ║
║   shared/peripherals/defaults.go : 34    ADCVrefLow = 0.0                 ║
║   module4-circuits/.../chargepump.go : 22-23  1.0V → 1.5V                 ║
║   module4-circuits/.../tia.go : 26       MaxOutputVoltage = 1.0           ║
║   module4-circuits/.../analysis.go : 249  Vread = 0.1                     ║
║                                                                            ║
║  ARCHITECTURE SELECTION GUIDE                                              ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   Array Size    Architecture    Notes                                     ║
║   ≤32×32        0T1R            V/2 scheme, maximum density               ║
║   64-256×256    1T1R            Standard production, 1000× isolation      ║
║   >256×256      2T1R            Ultra-precision, dual-path isolation      ║
║                                                                            ║
║  SAFETY CHECKS                                                             ║
║  ───────────────────────────────────────────────────────────────────────   ║
║   ✓ Vread < Vc                  (non-destructive read)                    ║
║   ✓ Vwrite > Vc                 (ensure switching)                        ║
║   ✓ V_half_select < Vc          (minimize disturb in 0T1R)                ║
║   ✓ MVM range = ADC range       (avoid clipping)                          ║
║                                                                            ║
╚════════════════════════════════════════════════════════════════════════════╝

NOTES:
  - All voltages at 300K (room temperature)
  - Timing parameters in config/physics.yaml
  - Vc varies with material (HZO: 1.2V, AlScN: 5-10V)
  - Half-select voltage = Vwrite/2 (0T1R only)
  - Transistor ON/OFF voltages are standard CMOS logic levels
  - Write voltages are level-dependent (30 unique values per cell)
  - Multi-level writes use ISPP (Incremental Step Pulse Programming)
  - See §3.2.1 for full multi-level write voltage complexity
```

---

**Document Status:**
- ✅ All peripheral voltages verified from source code
- ✅ Physics parameters cross-referenced with `config/physics.yaml`
- ⚠️ Write/half-select voltages are derived (not explicit in code)
- ✅ Architecture modes verified from device_state.go implementation

**Version:** 1.0
**Last Updated:** January 2026
**Part of:** FeCIM Lattice Tools
**License:** See project root
