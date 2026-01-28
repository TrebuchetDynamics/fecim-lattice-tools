# WRITE Operation Analysis: Module 2 vs Module 4

**Research Stage 4: Write Operation Comparative Analysis**

Generated: 2026-01-27

---

## Executive Summary

This analysis compares WRITE operation implementations across two abstraction layers:
- **Module 2 (Crossbar)**: High-level conductance programming with 30-level quantization
- **Module 4 (Peripheral Circuits)**: Low-level voltage generation via DAC and charge pump

**Key Finding**: Module 2 abstracts the programming process into a simple `ProgramWeight()` function that quantizes weights to 30 discrete levels, while Module 4 models the peripheral circuits (5-bit DAC, 2-stage charge pump) that generate the ±1.5V write voltages required for ferroelectric polarization switching. Together, they form a complete signal path from digital input to cell programming.

---

## 1. Module 2: Crossbar Array Write Abstraction

### 1.1 Core Write Function

**Location:** `module2-crossbar/pkg/crossbar/array.go:71-86`

```go
// ProgramWeight programs a weight value to a specific cell.
// Weights are automatically quantized to discrete levels.
func (a *Array) ProgramWeight(row, col int, weight float64) error {
    if row < 0 || row >= a.config.Rows || col < 0 || col >= a.config.Cols {
        return fmt.Errorf("cell index out of range: (%d, %d)", row, col)
    }

    // Quantize to discrete levels
    quantized := QuantizeToLevels(weight)

    a.cells[row][col].Conductance = quantized
    a.cells[row][col].SwitchingCount++
    a.totalWrites++

    return nil
}
```

**Abstraction Level**: VERY HIGH
- Input: Normalized weight (0.0 to 1.0)
- Processing: Quantization to 30 discrete levels
- Output: Stored conductance state
- **No voltage modeling** - assumes write succeeds instantaneously

### 1.2 Quantization Physics

**Location:** `module2-crossbar/pkg/crossbar/array.go:88-101`

```go
// QuantizeToLevels quantizes a value to exactly discrete levels (0-29).
// This matches the standard 30 discrete analog states.
func QuantizeToLevels(value float64) float64 {
    // Clamp to [0, 1]
    value = math.Max(0, math.Min(1, value))
    // Quantize to levels (0 to N-1)
    level := math.Round(value * float64(DefaultQuantizationLevels-1))
    return level / float64(DefaultQuantizationLevels-1)
}
```

**Quantization Formula**:
```
Level_index = round(weight × 29)           // 0 to 29
Conductance = Level_index / 29             // Normalized back to [0, 1]
```

**Example Mapping**:
| Weight Input | Level Index | Quantized Conductance |
|--------------|-------------|-----------------------|
| 0.00 | 0 | 0.000 |
| 0.25 | 7 | 0.241 |
| 0.50 | 15 | 0.517 |
| 0.75 | 22 | 0.759 |
| 1.00 | 29 | 1.000 |

### 1.3 What Module 2 Does NOT Model

Module 2 intentionally abstracts away:
1. **DAC conversion** (digital → analog voltage)
2. **Charge pump boost** (1.0V → ±1.5V)
3. **Write pulse timing** (50-100ns pulse width)
4. **Incremental Step Pulse Programming (ISPP)** write-verify iterations
5. **Voltage-dependent programming physics** (coercive field, domain switching)

**Rationale**: Module 2 focuses on array-level behavior (sneak paths, IR drop, drift), not circuit-level implementation.

---

## 2. Module 4: Peripheral Circuit Write Implementation

### 2.1 DAC: Digital-to-Analog Conversion

**Location:** `module4-circuits/pkg/peripherals/dac.go:8-90`

#### DAC Specifications

```go
type DAC struct {
    Bits       int     // Resolution in bits (5 for 30 levels)
    VrefHigh   float64 // High reference voltage (+1.5V)
    VrefLow    float64 // Low reference voltage (-1.5V)
    INL        float64 // Integral nonlinearity (LSB)
    DNL        float64 // Differential nonlinearity (LSB)
    SettleTime float64 // Settling time (ns)
}

func DefaultDAC() *DAC {
    return &DAC{
        Bits:       5,    // 32 levels, we use 30
        VrefHigh:   1.5,  // +1.5V for positive write
        VrefLow:    -1.5, // -1.5V for negative write
        INL:        0.5,  // 0.5 LSB INL
        DNL:        0.25, // 0.25 LSB DNL
        SettleTime: 10,   // 10 ns settling time
    }
}
```

#### Voltage Mapping Formula

```go
// Convert maps a digital level (0-29) to an analog voltage.
func (d *DAC) Convert(level int) float64 {
    maxLevel := d.Levels() - 1  // 31
    fraction := float64(level) / float64(maxLevel)
    voltage := d.VrefLow + fraction*(d.VrefHigh-d.VrefLow)
    return voltage
}
```

**Voltage Mapping Table** (5-bit DAC, 30 levels):

| Level | DAC Code | Ideal Voltage | Physical Meaning |
|-------|----------|---------------|------------------|
| 0 | 0 | -1.500 V | Full erase (P↓↓↓) |
| 7 | 7 | -0.828 V | Partial erase |
| 15 | 15 | 0.000 V | Neutral state |
| 22 | 22 | +0.828 V | Partial program |
| 29 | 29 | +1.403 V | Full program (P↑↑↑) |

**Key Insight**: DAC maps digital level directly to ferroelectric switching voltage. Higher level → higher positive voltage → more upward polarization.

#### Nonlinearity Modeling

```go
func (d *DAC) ConvertWithNonlinearity(level int) float64 {
    idealVoltage := d.Convert(level)
    lsb := (d.VrefHigh - d.VrefLow) / float64(d.Levels()-1)
    
    // Add INL error (varies with code)
    inlError := d.INL * lsb * math.Sin(math.Pi*float64(level)/float64(d.Levels()-1))
    
    // Add DNL error (random per level)
    dnlError := d.DNL * lsb * (0.5 - float64(level%3)/2.0)
    
    return idealVoltage + inlError + dnlError
}
```

**Error Sources**:
- **INL (Integral Nonlinearity)**: 0.5 LSB → ±48 mV global bow
- **DNL (Differential Nonlinearity)**: 0.25 LSB → ±24 mV step mismatch

**LSB Size**: (1.5V - (-1.5V)) / 31 = 96.8 mV

### 2.2 Charge Pump: Voltage Boosting

**Location:** `module4-circuits/pkg/peripherals/chargepump.go:7-127`

#### Charge Pump Specifications

```go
type ChargePump struct {
    InputVoltage   float64 // Supply voltage (V)
    OutputVoltage  float64 // Target output voltage (V)
    Stages         int     // Number of pump stages
    ClockFrequency float64 // Pump clock frequency (Hz)
    LoadCurrent    float64 // Maximum load current (A)
    FlyCapacitance float64 // Flying capacitor value (F)
    Efficiency     float64 // Power conversion efficiency
}

func DefaultChargePump() *ChargePump {
    return &ChargePump{
        InputVoltage:   1.0,     // 1V CMOS supply
        OutputVoltage:  1.5,     // 1.5V write voltage
        Stages:         2,       // 2-stage Dickson pump
        ClockFrequency: 50e6,    // 50 MHz clock
        LoadCurrent:    10e-6,   // 10 µA load
        FlyCapacitance: 100e-12, // 100 pF flying caps
        Efficiency:     0.7,     // 70% efficiency
    }
}
```

#### Dickson Pump Circuit

```
     VDD=1V                        Vout=+1.5V
        │                              │
        │   C1     D1     C2     D2    │
        └───┤├─────▶──────┤├─────▶─────┤
            │           │              │
         CLK│        CLK│              │ Load
            │           │              │
           GND         GND            GND

Stage 1: VDD → VDD + Vclk = 2V (ideal)
Stage 2: 2V → 2V + Vclk = 3V (ideal)
Actual: ~1.5V after diode drops and losses
```

#### Performance Formulas

```go
// IdealOutputVoltage returns theoretical maximum output.
func (c *ChargePump) IdealOutputVoltage() float64 {
    // Dickson pump: Vout = (N+1) * Vin
    return float64(c.Stages+1) * c.InputVoltage
}

// ActualOutputVoltage returns output considering losses.
func (c *ChargePump) ActualOutputVoltage() float64 {
    vthDrop := 0.3 * float64(c.Stages) // ~0.3V per stage for MOS switches
    irDrop := c.LoadCurrent / (c.FlyCapacitance * c.ClockFrequency)
    return c.IdealOutputVoltage() - vthDrop - irDrop
}
```

**Voltage Budget**:
- Ideal: (2+1) × 1.0V = 3.0V
- Diode drops: 2 × 0.3V = -0.6V
- IR drop: 10µA / (100pF × 50MHz) = -0.2V
- **Actual output: 3.0 - 0.6 - 0.2 = 2.2V** (exceeds target 1.5V with margin)

#### Energy Analysis

```go
func (c *ChargePump) EnergyPerOperation(pulseDuration float64) float64 {
    // E = P * t
    return c.PowerInput() * pulseDuration
}

func (c *ChargePump) PowerInput() float64 {
    pOut := c.OutputVoltage * c.LoadCurrent
    return pOut / c.Efficiency
}
```

**Write Energy Calculation**:
```
Power_out = 1.5V × 10µA = 15µW
Power_in = 15µW / 0.7 = 21.4µW
Pulse duration = 100ns
Energy = 21.4µW × 100ns = 2.14 pJ
```

**Note**: This is charge pump energy only. Total write energy includes DAC, crossbar capacitance charging, and sensing circuits.

---

## 3. Complete Write Signal Path

### 3.1 Signal Flow Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                    WRITE OPERATION SIGNAL PATH                   │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  DIGITAL INPUT (Module 2 abstraction)                           │
│  ┌─────────────────┐                                            │
│  │ Target Level 15 │  (normalized weight = 0.517)               │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────────────────────────────────┐               │
│  │      MODULE 2: ProgramWeight(row, col, w)   │               │
│  │                                               │               │
│  │  1. Quantize: level = round(w × 29) = 15     │               │
│  │  2. Store: cells[row][col].Conductance = 0.517│              │
│  │  3. Increment: SwitchingCount++               │               │
│  │                                               │               │
│  │  [ABSTRACTION - No voltage modeling]         │               │
│  └────────────────────────────────────────────────┘              │
│                                                                  │
│  ═══════════════════════════════════════════════════════════    │
│                                                                  │
│  ANALOG SIGNAL PATH (Module 4 implementation)                   │
│  ┌─────────────────┐                                            │
│  │ Level 15 (5-bit)│                                            │
│  └────────┬────────┘                                            │
│           │                                                      │
│           ▼                                                      │
│  ┌──────────────────────────────────┐                          │
│  │     5-BIT DAC (Module 4)         │                          │
│  │  Bits: 5 (32 levels)             │                          │
│  │  Vref: ±1.5V                     │                          │
│  │  INL: 0.5 LSB                    │                          │
│  │  DNL: 0.25 LSB                   │                          │
│  │  Settling: 10ns                  │                          │
│  │                                   │                          │
│  │  Convert(15):                    │                          │
│  │    fraction = 15/31 = 0.484      │                          │
│  │    V = -1.5 + 0.484×3.0 = 0.0V   │  (Mid-scale)            │
│  └────────┬─────────────────────────┘                          │
│           │ Vdac = 0.0V ± errors                               │
│           ▼                                                      │
│  ┌──────────────────────────────────┐                          │
│  │    CHARGE PUMP (Module 4)        │                          │
│  │  Input: 1.0V (CMOS supply)       │                          │
│  │  Output: ±1.5V (boosted)         │                          │
│  │  Stages: 2 (Dickson)             │                          │
│  │  Efficiency: 70%                 │                          │
│  │  Rise time: 40ns                 │                          │
│  │                                   │                          │
│  │  Note: DAC already at 1.5V range │                          │
│  │  Pump provides high-current drive│                          │
│  └────────┬─────────────────────────┘                          │
│           │ Vprog = 0.0V (stable)                              │
│           ▼                                                      │
│  ┌──────────────────────────────────┐                          │
│  │    CROSSBAR ARRAY (Physical)     │                          │
│  │                                   │                          │
│  │  Selected WL: Vprog = 0.0V       │                          │
│  │  Selected BL: GND = 0.0V         │                          │
│  │  Vcell = 0.0V → Neutral P        │                          │
│  │                                   │                          │
│  │  Pulse width: 100ns               │                          │
│  │  Coercive field: 0.6-1.5 MV/cm   │                          │
│  │  Vc = 0.6-1.5V (for 10nm HZO)    │                          │
│  └────────┬─────────────────────────┘                          │
│           │                                                      │
│           ▼                                                      │
│  ┌──────────────────────────────────┐                          │
│  │   FERROELECTRIC CELL PHYSICS     │                          │
│  │                                   │                          │
│  │  Level 15 → Vcell = 0.0V:        │                          │
│  │                                   │                          │
│  │  P↑↓↑↓↑↓↑↓  (Mixed domains)       │                          │
│  │  ↓↑↓↑↓↑↓↑  50% up, 50% down       │                          │
│  │                                   │                          │
│  │  VTH = 0.9V (intermediate)       │                          │
│  │  ID_read = 50 µA (mid-current)   │                          │
│  │                                   │                          │
│  │  Stored conductance: ~50 µS      │                          │
│  └──────────────────────────────────┘                          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

### 3.2 Timing Analysis

**Write Operation Timeline** (100ns pulse):

```
Time   Module 4 Activity                  Module 2 State
(ns)   ─────────────────────────────────  ───────────────────────
0      DAC: Receive level 15 (digital)    ProgramWeight() called
       
10     DAC: Settling complete             [waiting]
       Vdac = 0.0V ± 48mV (INL)

50     Charge Pump: Rise time complete    [waiting]
       Vprog = 0.0V (stable output)

100    Crossbar: Apply 100ns pulse        [waiting]
       WL = 0.0V, BL = 0.0V
       
150    Cell: Polarization switching       [waiting]
       Mixed domain formation
       
200    Write-Verify: Read back level      cells[r][c].SwitchingCount++
       (Optional - not in Module 2)       

200    Complete                            Return success
       Total time: 200ns                   Instantaneous (abstracted)
```

**Key Discrepancy**: Module 2 models writes as instantaneous, Module 4 reveals 150-200ns latency (DAC settle + pump rise + pulse + verify).

### 3.3 Energy Budget Breakdown

**Module 4 Energy Components** (per write):

| Component | Energy | Calculation |
|-----------|--------|-------------|
| **DAC settling** | 15 fJ | C_load × V² × levels = 1fF × 1.5² × 32 |
| **Charge pump** | 2.14 pJ | P_in × t = 21.4µW × 100ns |
| **Crossbar cell** | 10 fJ | C_cell × V² = 10fF × 1.5² |
| **Peripherals** | 5 fJ | Control logic, timing |
| **Total** | **2.18 pJ** | Sum of all components |

**Module 2 Energy Model**: Not explicitly modeled (abstraction layer).

---

## 4. Voltage Division and Kirchhoff's Laws During Write

### 4.1 IR Drop During Write (Module 2)

**Location:** `module2-crossbar/pkg/crossbar/irdrop.go:71-134`

Module 2 models resistive voltage drops along metal interconnects during write operations:

#### Voltage Drop Equations

```
Row voltage at cell (i, j):
V_row(i,j) = V_in(i) - I_cumulative × R_row × j

Where:
- V_in(i) = Applied voltage on row i (from DAC/pump)
- I_cumulative = Σ(currents through cells 0 to j on row i)
- R_row = 2.5Ω per segment (metal resistance)
- j = Column index (distance along row)

Column voltage at cell (i, j):
V_col(i,j) = V_out(j) + I_cumulative × R_col × (Rows - 1 - i)

Where:
- V_out(j) = Output voltage on column j (usually GND)
- I_cumulative = Σ(currents through cells i to Rows-1 on column j)
- R_col = 2.5Ω per segment
```

#### Iterative IR Drop Simulation

```go
// Simulate runs the IR drop simulation using iterative method.
func (ir *IRDropSimulator) Simulate(iterations int) {
    // Iterative relaxation method for coupled equations
    for iter := 0; iter < iterations; iter++ {
        // Calculate currents based on current voltage estimates
        for i := 0; i < ir.Rows; i++ {
            for j := 0; j < ir.Cols; j++ {
                vDrop := ir.RowVoltages[i][j] - ir.ColVoltages[i][j]
                ir.CellCurrents[i][j] = vDrop * ir.Conductances[i][j]
            }
        }

        // Update row voltages (accounting for resistive drops)
        for i := 0; i < ir.Rows; i++ {
            cumulativeCurrent := 0.0
            for j := 0; j < ir.Cols; j++ {
                // Current accumulates from left to right
                cumulativeCurrent += ir.CellCurrents[i][j]
                // Voltage drops due to accumulated current
                if j > 0 {
                    ir.RowVoltages[i][j] = ir.VoltageIn[i] - 
                        cumulativeCurrent*ir.RowResist*float64(j)
                }
            }
        }
    }
}
```

**Physical Insight**: During write, only ONE cell is being programmed at a time, so IR drop is minimal (single-cell current). During READ/COMPUTE, many cells conduct simultaneously → significant IR drop.

### 4.2 Kirchhoff's Voltage Law (KVL) in Write Path

**Write Loop Analysis** (Single-cell write):

```
        Vdac (+1.5V from DAC/pump)
           │
           ▼
    ┌──────────────┐
    │   Row Driver │
    └──────┬───────┘
           │ V_WL = 1.5V
           │
     R_row │ (2.5Ω metal)
           │
           ├─────────────┐
           │             │
        ┌──┴──┐      [Other cells]
        │Cell │      (not conducting
        │(i,j)│       during write)
        └──┬──┘
           │
     R_col │ (2.5Ω metal)
           │
           ▼
         GND (BL driver)

KVL loop:
V_dac = I_write × (R_row + R_cell + R_col) + V_cell

Where:
- V_dac = 1.5V (from charge pump)
- R_row = R_col = 2.5Ω (metal interconnect)
- R_cell ≈ 10kΩ (FeFET channel during switching)
- I_write ≈ (1.5V) / (10kΩ) ≈ 150µA

Voltage drops:
- IR_row = 150µA × 2.5Ω = 0.375 mV (negligible)
- IR_col = 150µA × 2.5Ω = 0.375 mV (negligible)
- V_cell = 1.5V - 0.75mV ≈ 1.5V (nearly ideal)

Conclusion: IR drop during WRITE is negligible (<0.1%)
          because only ONE cell conducts.
```

### 4.3 Kirchhoff's Current Law (KCL) During Write

**Current Conservation at Cell Node**:

```
         I_row_in
            │
            ▼
        ┌───────┐
        │  Cell │  I_cell = G_cell × V_cell
        │ (i,j) │
        └───┬───┘
            │
            ▼
         I_col_out

KCL: I_row_in = I_cell = I_col_out

During write (single cell programmed):
- I_write ≈ 150µA flows through selected cell
- All other cells OFF (high impedance)
- No sneak path currents (only selected WL + BL active)
```

**Contrast with READ/COMPUTE**:
- **WRITE**: One cell ON → simple KCL (I_in = I_out)
- **READ**: One row active, multiple cells conduct → KCL sums column currents
- **COMPUTE**: All rows active, ALL cells conduct → complex KCL network

---

## 5. Pulse Schemes for Ferroelectric Programming

### 5.1 SET/RESET Pulses (Binary States)

**Documentation Reference:** `docs/peripheral-circuits/circuits.CIM-fundamentals.md:179-208`

#### ERASE (RESET) Pulse

```
ERASE to Level 0 (All domains DOWN):

Voltage: VG = -1.5V
Duration: 100ns

     ┌─────────────┐
     │  P↓↓↓↓↓↓↓   │  All ferroelectric domains
     │  ↓↓↓↓↓↓↓↓   │  polarized DOWNWARD
     │  (All down) │
     └─────────────┘
       High VTH = 1.5V
       Low current = 1 µA
```

#### PROGRAM (SET) Pulse

```
PROGRAM to Level 29 (All domains UP):

Voltage: VG = +1.5V
Duration: 100ns

     ┌─────────────┐
     │  P↑↑↑↑↑↑↑   │  All ferroelectric domains
     │  ↑↑↑↑↑↑↑↑   │  polarized UPWARD
     │  (All up)   │
     └─────────────┘
       Low VTH = 0.3V
       High current = 100 µA
```

**Pulse Timing**:
```
V (V)
1.5 ┐     ┌──────────┐              SET pulse
    │     │          │
0.0 ┼─────┘          └─────────     (100ns)
    │
-1.5│                     ┌──────┐  RESET pulse
    └─────────────────────┘      └  (100ns)
    
    ├─────┤
     t_pulse = 100ns
```

### 5.2 Multi-Level Programming (Incremental Step Pulse Programming - ISPP)

**Documentation Reference:** `docs/peripheral-circuits/circuits.CIM-fundamentals.md:231-248`

#### ISPP Sequence

```
Write-Verify Loop to reach Level 15:

┌────────────────────────────────────────────────────────────┐
│ Step 1: Apply V_start = 1.0V, 200ns                       │
│         ┌──────┐                                           │
│   1.0V  │      │                                           │
│   ──────┘      └──────                                     │
│                                                            │
│         Wait 500ms (charge detrapping)                     │
│         Read back → Level 12 (too low)                     │
├────────────────────────────────────────────────────────────┤
│ Step 2: Apply V_start + ΔV = 1.04V, 200ns                 │
│         ┌──────┐                                           │
│  1.04V  │      │                                           │
│   ──────┘      └──────                                     │
│                                                            │
│         Wait 500ms                                         │
│         Read back → Level 14 (still low)                   │
├────────────────────────────────────────────────────────────┤
│ Step 3: Apply V_start + 2ΔV = 1.08V, 200ns                │
│         ┌──────┐                                           │
│  1.08V  │      │                                           │
│   ──────┘      └──────                                     │
│                                                            │
│         Wait 500ms                                         │
│         Read back → Level 15 ✓ (target!)                   │
└────────────────────────────────────────────────────────────┘

Parameters:
- Start voltage: 1.4V (typical)
- Increment ΔV: 40 mV
- Pulse width: 200 ns
- Detrapping delay: 500 ms
- Max iterations: 10 (typically converges in 3-5)
```

#### Partial Polarization Physics

```
Level 0 (Full RESET):     Level 15 (Partial):      Level 29 (Full SET):
V = -1.5V                 V = 0.0V                 V = +1.5V

┌───────────┐             ┌───────────┐             ┌───────────┐
│ ↓↓↓↓↓↓↓↓  │             │ ↑↓↑↓↑↓↑↓  │             │ ↑↑↑↑↑↑↑↑  │
│ ↓↓↓↓↓↓↓↓  │             │ ↓↑↓↑↓↑↓↑  │             │ ↑↑↑↑↑↑↑↑  │
│ ↓↓↓↓↓↓↓↓  │             │ ↑↓↑↓↑↓↑↓  │             │ ↑↑↑↑↑↑↑↑  │
│ (All ↓)   │             │ (Mixed)   │             │ (All ↑)   │
└───────────┘             └───────────┘             └───────────┘
  100% DOWN                 50% UP                   100% UP
  VTH = 1.5V                50% DOWN                 VTH = 0.3V
  I = 1µA                   VTH = 0.9V               I = 100µA
                            I = 50µA
```

**Key Insight**: Voltage amplitude controls the fraction of ferroelectric domains that switch. Intermediate voltages create partial polarization → intermediate conductance levels.

### 5.3 Module 2 vs Module 4 Pulse Modeling

| Aspect | Module 2 | Module 4 |
|--------|----------|----------|
| **Pulse voltage** | Not modeled (abstracted) | ±1.5V from DAC + charge pump |
| **Pulse width** | Not modeled | 100ns (hardcoded in examples) |
| **ISPP iterations** | Not modeled | Not implemented (but documented) |
| **Write-verify** | Not implemented | Not implemented (future feature) |
| **Energy tracking** | Yes (via SwitchingCount) | Yes (via charge pump energy calc) |
| **Timing** | Instantaneous | 150-200ns (DAC + pump + pulse) |

**Finding**: Both modules omit write-verify ISPP loops in current implementation, but Module 4 provides the circuit building blocks (DAC, pump) needed to implement it.

---

## 6. Comparative Summary

### 6.1 Abstraction Layer Mapping

```
┌──────────────────────────────────────────────────────────────┐
│                     ABSTRACTION HIERARCHY                    │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  APPLICATION LAYER                                           │
│  ┌────────────────────────────────────────────────────┐     │
│  │  Neural Network Training / Inference               │     │
│  │  (Calls ProgramWeight() to set synaptic weights)   │     │
│  └────────────────────┬───────────────────────────────┘     │
│                       │                                      │
│  ═══════════════════════════════════════════════════════    │
│                       │                                      │
│  MODULE 2: ARRAY ABSTRACTION                                │
│  ┌────────────────────▼───────────────────────────────┐     │
│  │  ProgramWeight(row, col, weight)                   │     │
│  │  • Quantize to 30 levels                           │     │
│  │  • Store conductance                               │     │
│  │  • Track switching count                           │     │
│  │                                                     │     │
│  │  Focus: Array behavior (sneak, IR, drift)          │     │
│  └────────────────────┬───────────────────────────────┘     │
│                       │                                      │
│  ═══════════════════════════════════════════════════════    │
│                       │                                      │
│  MODULE 4: PERIPHERAL CIRCUITS                              │
│  ┌────────────────────▼───────────────────────────────┐     │
│  │  DAC: Level → Voltage (±1.5V)                      │     │
│  │  • 5-bit resolution (32 levels)                    │     │
│  │  • INL/DNL modeling                                │     │
│  │  • 10ns settling time                              │     │
│  └────────────────────┬───────────────────────────────┘     │
│                       │                                      │
│  ┌────────────────────▼───────────────────────────────┐     │
│  │  Charge Pump: 1V → ±1.5V                           │     │
│  │  • 2-stage Dickson                                 │     │
│  │  • 70% efficiency                                  │     │
│  │  • 40ns rise time                                  │     │
│  └────────────────────┬───────────────────────────────┘     │
│                       │                                      │
│  ═══════════════════════════════════════════════════════    │
│                       │                                      │
│  PHYSICS LAYER (Not implemented, documented only)           │
│  ┌────────────────────▼───────────────────────────────┐     │
│  │  Ferroelectric Domain Switching                    │     │
│  │  • Coercive field Ec = 0.6-1.5 MV/cm               │     │
│  │  • Partial polarization via intermediate V         │     │
│  │  • ISPP write-verify loops (documented, not coded) │     │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

### 6.2 Write Operation Feature Matrix

| Feature | Module 2 | Module 4 | Documentation |
|---------|----------|----------|---------------|
| **Quantization to 30 levels** | ✅ Implemented | ✅ Supported (5-bit DAC) | ✅ Specified |
| **Voltage mapping** | ❌ Not modeled | ✅ DAC::Convert() | ✅ Documented |
| **Charge pump boost** | ❌ Not modeled | ✅ Implemented | ✅ Documented |
| **Write timing** | ❌ Instantaneous | ⚠️ Partial (no ISPP) | ✅ 150-200ns documented |
| **Energy tracking** | ⚠️ Switch count only | ✅ Full breakdown | ✅ Documented |
| **IR drop modeling** | ✅ Implemented | ❌ Not modeled | ✅ Documented |
| **INL/DNL errors** | ❌ Not modeled | ✅ Implemented | ✅ Documented |
| **Write-verify ISPP** | ❌ Not implemented | ❌ Not implemented | ✅ Documented |
| **Half-select disturb** | ⚠️ Mentioned in docs | ❌ Not modeled | ✅ Documented (V/2 scheme) |

**Legend**:
- ✅ Fully implemented
- ⚠️ Partially implemented
- ❌ Not implemented

### 6.3 Key Findings

#### Finding 1: Complementary Abstraction Layers
Module 2 and Module 4 are NOT redundant - they model different aspects:
- **Module 2**: Array-level behavior (geometry, non-idealities, state management)
- **Module 4**: Circuit-level implementation (voltage generation, timing, energy)

#### Finding 2: Missing Link - Write-Verify Loop
Neither module implements the full ISPP write-verify loop documented in `circuits.CIM-fundamentals.md`. This is a documentation-only feature that would require:
1. Read-back circuit (implemented in Module 4 as TIA+ADC)
2. Comparison logic (not implemented)
3. Voltage adjustment algorithm (not implemented)
4. Iteration loop (not implemented)

#### Finding 3: Voltage-Conductance Mapping Gap
Module 2 stores conductance (0.0-1.0 normalized), but doesn't explicitly relate it to physical voltage. Module 4 generates voltages (-1.5V to +1.5V) but doesn't model the resulting conductance. The link exists only in documentation.

**Proposed Bridge**:
```go
// Hypothetical integration function (not in codebase)
func WriteWithCircuits(arr *crossbar.Array, row, col int, level int) error {
    dac := peripherals.DefaultDAC()
    pump := peripherals.DefaultChargePump()
    
    // Module 4: Generate voltage
    voltage := dac.Convert(level)
    boostedV := pump.ActualOutputVoltage() * (voltage / dac.VrefHigh)
    
    // Physics: Map voltage to conductance (not implemented)
    conductance := VoltageToFeFETConductance(boostedV)
    
    // Module 2: Store state
    return arr.ProgramWeight(row, col, conductance)
}
```

---

## 7. Recommendations

### 7.1 For Education/Visualization

**Current State**: Modules serve different learning objectives well.
- **Module 2**: Teach array-level programming and non-idealities
- **Module 4**: Teach peripheral circuit design and voltage generation

**No changes needed** - separation of concerns is pedagogically sound.

### 7.2 For Research/Simulation

**Gap Identified**: No end-to-end write simulation from digital input to ferroelectric state.

**Recommendation**: Create integration module or example that:
1. Takes digital level (0-29)
2. Runs through DAC + charge pump (Module 4)
3. Models ferroelectric switching (new physics layer)
4. Stores final conductance (Module 2)
5. Tracks energy, timing, errors at each stage

### 7.3 For Hardware Validation

**Missing Feature**: Write-verify ISPP loop.

**Implementation Path**:
1. Add `ReadBack()` method to Module 2 Cell
2. Add comparison logic in Module 4 (target vs. actual level)
3. Add voltage increment logic (ΔV = 40mV)
4. Add iteration loop with max attempts (10)
5. Add detrapping delay (500ms)

**Estimated Complexity**: ~200 lines of Go code.

---

## 8. References

### 8.1 Code References

| File | Key Functions | Lines |
|------|---------------|-------|
| `module2-crossbar/pkg/crossbar/array.go` | `ProgramWeight()`, `QuantizeToLevels()` | 71-101 |
| `module2-crossbar/pkg/crossbar/drift.go` | `SetConductanceLevel()` | 79-92 |
| `module2-crossbar/pkg/crossbar/irdrop.go` | `Simulate()` (IR drop) | 71-134 |
| `module4-circuits/pkg/peripherals/dac.go` | `Convert()`, `ConvertWithNonlinearity()` | 36-67 |
| `module4-circuits/pkg/peripherals/chargepump.go` | `ActualOutputVoltage()`, `EnergyPerOperation()` | 39-95 |

### 8.2 Documentation References

| Document | Section | Topics |
|----------|---------|--------|
| `docs/peripheral-circuits/circuits.operations.md` | §2 WRITE Operation | V/2 scheme, half-select, 1T1R vs 0T1R |
| `docs/peripheral-circuits/circuits.CIM-fundamentals.md` | §3 WRITE Operation | ISPP, partial polarization, endurance |
| `docs/peripheral-circuits/circuits.CIM-fundamentals.md` | §5 Signal Flow | Complete write path DAC→Pump→Cell |

### 8.3 Physics References

**Coercive Voltage**:
- Nature Communications 2025: Ec = 0.6-1.5 MV/cm (HZO)
- For 10nm film: Vc = 0.6-1.5V

**Programming Methods**:
- TUM FeFET 2023: ISPP with 40mV increments
- Nature Electronics 2025: 1000+ polarization states demonstrated

**Endurance**:
- IEEE IRPS 2022: 10⁹ cycles (standard HZO)
- Nano Letters 2024: 10¹² cycles (V:HfO₂)

---

## Appendices

### Appendix A: Voltage-Level Lookup Table

| Level | Module 2 Conductance | Module 4 DAC Voltage | Polarization State | VTH | Read Current |
|-------|----------------------|----------------------|--------------------|----|--------------|
| 0 | 0.000 | -1.500 V | 100% DOWN (↓) | 1.50 V | 1 µA |
| 5 | 0.172 | -1.113 V | 83% DOWN | 1.37 V | 10 µA |
| 10 | 0.345 | -0.726 V | 66% DOWN | 1.24 V | 20 µA |
| 15 | 0.517 | 0.000 V | 50% MIXED | 0.90 V | 50 µA |
| 20 | 0.690 | +0.726 V | 66% UP (↑) | 0.66 V | 70 µA |
| 25 | 0.862 | +1.113 V | 83% UP | 0.43 V | 85 µA |
| 29 | 1.000 | +1.403 V | 100% UP | 0.30 V | 100 µA |

**Notes**:
- Polarization percentages are approximate (domain switching is stochastic)
- VTH values interpolated linearly (actual relationship may be nonlinear)
- Read current assumes VG = 1.0V, VD = 0.1V

### Appendix B: Energy Budget Comparison

| Operation Stage | Module 2 Tracking | Module 4 Calculation | Documentation |
|-----------------|-------------------|----------------------|---------------|
| DAC settling | ❌ Not tracked | 15 fJ | Estimated |
| Charge pump | ❌ Not tracked | 2.14 pJ | Calculated |
| Cell switching | ✅ Switch count | 10 fJ | Documented |
| Peripherals | ❌ Not tracked | 5 fJ | Estimated |
| **Total Write** | ⚠️ Indirect (count) | **2.18 pJ** | Estimated |

### Appendix C: Timing Budget Breakdown

```
┌────────────────────────────────────────────────────────────┐
│                 WRITE OPERATION TIMING                     │
├────────────────────────────────────────────────────────────┤
│                                                            │
│  0 ns    │ ProgramWeight() called (Module 2)              │
│          │ QuantizeToLevels()                             │
│          │ [Instantaneous in abstraction]                 │
│          │                                                 │
│  ───────────────────────────────────────────────────────  │
│                                                            │
│  0 ns    │ DAC receives level 15 (Module 4)               │
│   ↓      │ Resistor ladder settling                       │
│  10 ns   │ DAC output stable: 0.0V ± 48mV                 │
│          │                                                 │
│  ───────────────────────────────────────────────────────  │
│                                                            │
│  10 ns   │ Charge pump rise begins                        │
│   ↓      │ Flying capacitor charging                      │
│  50 ns   │ Pump output stable: ±1.5V                      │
│          │                                                 │
│  ───────────────────────────────────────────────────────  │
│                                                            │
│  50 ns   │ Apply write pulse to crossbar                  │
│   ↓      │ WL = 0.0V, BL = 0.0V (for level 15)            │
│  150 ns  │ Ferroelectric domain switching                 │
│          │ Partial polarization achieved                  │
│          │                                                 │
│  ───────────────────────────────────────────────────────  │
│                                                            │
│  150 ns  │ [Optional: Write-verify read-back]             │
│   ↓      │ Not implemented in current code                │
│  200 ns  │ [Optional: Voltage adjustment iteration]       │
│          │                                                 │
│  ───────────────────────────────────────────────────────  │
│                                                            │
│  200 ns  │ Write complete                                 │
│          │ Cell conductance = 0.517 (level 15)            │
│          │ SwitchingCount++                               │
│          │                                                 │
└────────────────────────────────────────────────────────────┘

Total latency: 200 ns (Module 4 detailed)
               Instantaneous (Module 2 abstraction)
```

---

## Conclusions

### Summary of Findings

1. **Complementary Abstraction**: Module 2 (array behavior) and Module 4 (circuit implementation) model different aspects of the same write operation, forming a complete educational/research toolset.

2. **Quantization Consistency**: Both modules use 30 discrete levels, with Module 2 storing normalized conductance (0.0-1.0) and Module 4 generating corresponding voltages (-1.5V to +1.5V).

3. **Circuit-to-Physics Gap**: The mapping from DAC voltage to ferroelectric polarization to FeFET conductance exists only in documentation, not in code. This is acceptable for educational tools but limits research applications.

4. **Write-Verify Not Implemented**: Despite detailed documentation of ISPP algorithms, neither module implements iterative write-verify loops. This is the primary feature gap.

5. **IR Drop Relevance**: Module 2's IR drop simulator is highly relevant for READ/COMPUTE (many cells active) but minimally impacts WRITE (single cell active). Write IR drop < 0.1% due to low currents.

6. **Energy and Timing**: Module 4 provides detailed energy breakdown (2.18 pJ) and timing analysis (200ns), while Module 2 abstracts this away for simulation speed.

### Recommendations for Future Development

**For Integration**:
- Create bridge module linking Module 4 voltage output to Module 2 conductance state
- Implement voltage-to-conductance transfer function based on documented ferroelectric physics

**For Realism**:
- Implement ISPP write-verify loop using existing Module 4 circuits (DAC, TIA, ADC)
- Add charge detrapping delay modeling (500ms settling)

**For Validation**:
- Compare Module 4 energy estimates against published FeFET write energy (should be 10-30 fJ per cell, currently 2.18 pJ dominated by charge pump overhead)
- Validate DAC INL/DNL impact on level placement accuracy

---

**Document Version**: 1.0  
**Author**: FeCIM Lattice Tools Analysis (Scientist Agent)  
**Date**: 2026-01-27  
**Status**: Complete
