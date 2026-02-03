# Compute-in-Memory (CIM) Fundamentals

**How to Read, Write, and Compute in Ferroelectric Crossbar Arrays**

*Last Updated: January 2026*

---

## Executive Summary

Compute-in-Memory (CIM) eliminates the von Neumann bottleneck by performing computation directly where data is stored. This document explains the fundamental physics and operations of CIM in ferroelectric (FeFET) crossbar arrays, covering:

1. **READ** - Non-destructive sensing of stored analog states
2. **WRITE** - Programming multi-level polarization states
3. **COMPUTE** - In-situ matrix-vector multiplication (MVM)

**Key Insight:** CIM leverages Ohm's Law (V = I × R for multiplication) and Kirchhoff's Current Law (current summation for accumulation) to perform MAC operations in O(1) time, achieving 10-1000× energy efficiency over digital processors.

**Note:** References to 30 levels refer to the demo baseline (conference claim; pending peer review). Peer‑reviewed devices report 32–140 states.

---

## 1. The von Neumann Problem and CIM Solution

### 1.1 The Memory Wall

Traditional computing separates memory from processing:

```
┌─────────────┐         Data Bus          ┌─────────────┐
│   MEMORY    │ ←───────────────────────→ │  PROCESSOR  │
│  (Storage)  │    ↑ Bottleneck!          │  (Compute)  │
└─────────────┘    Limited bandwidth      └─────────────┘
                   High energy cost

Problem: Moving data consumes 100-1000× more energy than computation
Neural networks: >90% time spent moving weights, not computing
```

### 1.2 CIM Solution: Compute Where Data Lives

```
        ┌────────────────────────────────────┐
        │     CROSSBAR ARRAY                 │
        │  (Weights stored in cells)         │
        │                                    │
Input → │   [W₀₀] [W₀₁] [W₀₂] [W₀₃]         │ → Output
Voltage │   [W₁₀] [W₁₁] [W₁₂] [W₁₃]         │   Current
        │   [W₂₀] [W₂₁] [W₂₂] [W₂₃]         │   = MVM Result
        │   [W₃₀] [W₃₁] [W₃₂] [W₃₃]         │
        │                                    │
        │   Computation happens IN-PLACE     │
        │   No data movement needed!         │
        └────────────────────────────────────┘

Result: 10-1000× energy savings, O(1) compute time
```

---

## 2. READ Operation: Sensing Stored States

### 2.1 Physical Principle

FeFET read measures drain current, which depends on the ferroelectric polarization state:

```
                    Gate (VG)
                       │
           ┌───────────┴───────────┐
           │                       │
           │   FERROELECTRIC       │ ← Polarization P determines
           │   (HfO₂-ZrO₂)         │   threshold voltage VTH
           │                       │
           └───────────┬───────────┘
                       │
    Source ───────[Channel]─────── Drain
           (VS)                   (VD = 0.1V)

Read current: ID = μCox(W/L)(VG - VTH)² (saturation)

Where VTH depends on polarization:
- Positive P (up) → Low VTH → High current → Level 29
- Negative P (down) → High VTH → Low current → Level 0
- Partial P → Intermediate VTH → Intermediate current → Levels 1-28
```

### 2.2 Read Voltage Requirements

| Parameter | Value | Rationale |
|-----------|-------|-----------|
| **Gate Read Voltage (VG)** | 0.5-1.0 V | Must be above VTH but below coercive voltage |
| **Drain Bias (VD)** | 0.1 V | Low voltage prevents disturb |
| **Safe Read Zone** | |V| < 0.5 V| Well below Ec (~0.6-1.5 MV/cm) |
| **Read Latency** | 1-5 ns | Current-domain sensing |

**Critical Rule:** Read voltage must be below the coercive voltage (Vc ≈ 0.6-1.5 V) to prevent polarization switching.

### 2.3 Current-Domain Sensing

```
              FeFET States and Read Current:

Level 0 (Erased)    Level 15 (Mid)    Level 29 (Programmed)
    │                    │                    │
    │  VTH = 1.5V        │  VTH = 0.9V        │  VTH = 0.3V
    │  ID = 1 µA         │  ID = 50 µA        │  ID = 100 µA
    ▼                    ▼                    ▼

Read: VG = 1.0V, VD = 0.1V

Current Range: 1 µA to 100 µA (100:1 ratio)
└──→ Excellent sensing margin for 30 levels
```

**Sensing Chain:**
```
    FeFET        TIA           ADC        Output
   (current)  (amplifier)  (digitizer)   (level)

   1-100 µA  →  10-1000 mV  →  5-bit  →  0-29
              ×10kΩ gain     32 levels
```

### 2.4 Non-Destructive Read

Unlike conventional FeRAM (which destroys data on read), FeFET enables non-destructive read:

| Feature | Conventional FeRAM (1T-1C) | FeFET (1T) |
|---------|---------------------------|-----------|
| Read Mechanism | Charge sensing (destructive) | Current sensing |
| Read Disturb | Write-back required | None with proper VG |
| Endurance | Limited by read-write cycles | >10¹¹ read cycles |
| Sensing | Capacitance difference | VTH difference |

**Non-Destructive Read Principle:**
```
Read field direction (Gate bias):

          Applied Field
             │
             ▼
    ┌─────────────────┐
    │    P↑ (up)      │  Field same direction as P
    │    ↑↑↑↑↑↑       │  → No switching, stable read
    │    ↑↑↑↑↑↑       │
    └─────────────────┘

Key: Read voltage produces field ALIGNED with existing polarization
     → No switching → Non-destructive
```

### 2.5 Read Disturb Prevention

**Problem:** Even small read voltages can cause gradual polarization drift over many cycles.

**Solutions:**

1. **Dual-Port FeFET Architecture:**
   - Separate write gate (ferroelectric) and read gate (non-ferroelectric)
   - Read operation completely isolated from ferroelectric layer
   - Eliminates read disturb entirely

2. **Optimized Read Voltage:**
   ```
   Safe read zone: |VG| < 0.5 × Vc

   For HZO (Vc ≈ 0.6-1.5 V):
   - Conservative: VG < 0.3 V
   - Typical: VG = 0.5 V
   - Aggressive: VG = 0.8 V (monitor retention)
   ```

3. **Periodic Refresh:**
   - Monitor threshold voltage drift
   - Refresh (re-program) cells exceeding drift threshold
   - Typically needed every 10⁶-10⁸ reads for aggressive read voltages

---

## 3. WRITE Operation: Programming Analog States

### 3.1 Physical Principle: Polarization Switching

FeFET write switches ferroelectric domain polarization:

```
ERASE (Level 0):               PROGRAM (Level 29):
Apply VG = -1.5V               Apply VG = +1.5V

     ┌─────────────┐               ┌─────────────┐
     │  P↓↓↓↓↓↓↓   │               │  P↑↑↑↑↑↑↑   │
     │  ↓↓↓↓↓↓↓↓   │  ─────→       │  ↑↑↑↑↑↑↑↑   │
     │  (All down) │   Pulse       │  (All up)   │
     └─────────────┘               └─────────────┘
       High VTH                      Low VTH
       Low current                   High current


PARTIAL POLARIZATION (Level 15):
Apply VG = +1.0V (below full switching)

     ┌─────────────┐
     │  P↑↓↑↓↑↓↑   │
     │  ↓↑↓↑↓↑↓↑   │  Mixed domains
     │  (Partial)  │
     └─────────────┘
       Medium VTH
       Medium current
```

### 3.2 Multi-Level Programming Methods

#### Method 1: Voltage Amplitude Modulation

```
Programming Voltage vs. Stored Level:

VG (V)  │                              ●── Level 29
  1.5   │                          ●
        │                      ●
  1.2   │                  ●
        │              ●
  0.9   │          ●
        │      ●
  0.6   │  ●
        │──●────────────────────────────→ Level
       0   5   10   15   20   25   29

Each voltage produces different partial polarization
```

#### Method 2: Incremental Step Pulse Programming (ISPP)

```
Write-Verify Sequence (reaching Level 15):

Step 1: Apply V_start = 1.0V, 200ns
        Read back → Level 12 (too low)

Step 2: Apply V_start + ΔV = 1.04V, 200ns
        Read back → Level 14 (still low)

Step 3: Apply V_start + 2ΔV = 1.08V, 200ns
        Read back → Level 15 ✓ (target!)

        Typical ΔV = 40-100 mV
        Typical iterations = 3-5
```

**ISPP Advantages:**
- Compensates device-to-device variation
- Achieves tighter level distributions
- Essential for >5-bit precision

#### Method 3: Pulse Width Modulation

```
Fixed Voltage, Variable Pulse Width:

Pulse Width  │                              ●── Level 29
  500 ns     │                          ●
             │                      ●
  350 ns     │                  ●
             │              ●
  200 ns     │          ●
             │      ●
   50 ns     │  ●
             │──●────────────────────────────→ Level
            0   5   10   15   20   25   29

Longer pulse → More domain switching → Higher level
```

### 3.3 Write Specifications

| Parameter | Value | Source |
|-----------|-------|--------|
| **Write Voltage** | ±1.2 to ±1.5 V | Nature Commun. 2023 |
| **Full Program/Erase** | ±4.5-5.0 V, 500 ns | TUM FeFET 2023 |
| **ISPP Start Voltage** | 1.4 V | Multi-level studies |
| **ISPP Increment** | 40 mV | TUM FeFET 2023 |
| **Pulse Width** | 50-500 ns | Varies by method |
| **Write Energy** | <1 fJ per write | FeFET literature |
| **Fast Switching** | <10 ns, sub-ns achieved | Nano Letters 2024 |
| **Ultra-low Energy** | 8 aJ (attojoules) | Boolean logic ops |

### 3.4 Write-Verify Implementation

```go
// Pseudocode for multi-level write-verify
func WriteLevel(cell *FeFET, targetLevel int) error {
    const (
        startVoltage = 1.4    // V
        deltaV       = 0.04   // 40 mV increment
        pulseWidth   = 200e-9 // 200 ns
        maxAttempts  = 10
        detrapping   = 500e-3 // 500 ms settling
    )

    voltage := startVoltage + float64(targetLevel) * 0.02 // Estimate

    for attempt := 0; attempt < maxAttempts; attempt++ {
        // Apply programming pulse
        cell.ApplyPulse(voltage, pulseWidth)

        // Wait for charge detrapping
        time.Sleep(detrapping)

        // Read back current level
        currentLevel := cell.Read()

        if currentLevel == targetLevel {
            return nil // Success!
        } else if currentLevel < targetLevel {
            voltage += deltaV // Increase voltage
        } else {
            // Overshoot - must erase and retry
            cell.Erase()
            voltage -= deltaV * 2
        }
    }
    return errors.New("write failed after max attempts")
}
```

### 3.5 Charge Pump Requirement

Modern CMOS operates at ~1 V, but FeFET switching requires ~1.5 V:

```
CMOS Supply: 1.0V
FeFET Vc:    ~0.6-1.5V
Required:    1.2-1.5V (with margin)

            ┌────────────────────────────┐
            │     CHARGE PUMP            │
 1.0V ────→ │  Dickson 2-stage          │ ────→ ±1.5V
 (CMOS)     │  Efficiency: 70%          │   (Write voltage)
            │  Rise time: 40 ns         │
            └────────────────────────────┘
```

### 3.6 Endurance Considerations

| Material | Endurance | Conditions |
|----------|-----------|------------|
| **HfO₂-ZrO₂ Superlattice** | 5×10¹² cycles | Optimized structure |
| **Standard HZO** | 10⁹ cycles | Typical operation |
| **V:HfO₂ (Vanadium doped)** | 10¹² cycles | IEEE 2024 |
| **General HfO₂** | 10⁸ cycles | Sub-5V operation |

**Endurance Optimization:**
- Lower write voltage extends lifetime exponentially
- Avoid over-programming (stop at target level)
- Balance write completeness vs. cycling stress

---

## 4. COMPUTE Operation: In-Memory Matrix-Vector Multiplication

### 4.1 The Physics of Analog MAC

CIM exploits fundamental circuit laws:

```
OHMS LAW (Multiplication):
────────────────────────────────────
V = I × R  →  I = G × V  (G = 1/R)

Apply voltage V to cell with conductance G:
Current = Weight × Input

    V (input)
    │
    ▼
┌─────────┐
│ G = W   │  Cell stores weight as conductance
└────┬────┘
     │
     ▼
   I = G×V  (current = weight × input = multiplication!)


KIRCHHOFF'S CURRENT LAW (Accumulation):
────────────────────────────────────
Currents sum at a node.

        V₀    V₁    V₂    V₃  (inputs)
         │     │     │     │
         ▼     ▼     ▼     ▼
    ┌───┐ ┌───┐ ┌───┐ ┌───┐
    │G₀₀│ │G₀₁│ │G₀₂│ │G₀₃│  (weights)
    └─┬─┘ └─┬─┘ └─┬─┘ └─┬─┘
      │     │     │     │
      └─────┴──┬──┴─────┘
               │
               ▼
    I_sum = G₀₀×V₀ + G₀₁×V₁ + G₀₂×V₂ + G₀₃×V₃

This IS the dot product: I = G · V (accumulation!)
```

### 4.2 Complete MVM in One Cycle

```
Matrix-Vector Multiplication: y = W × x

Input vector x applied as voltages on columns:
    x = [V₀, V₁, V₂, V₃]

Weight matrix W stored as conductances:
         BL₀   BL₁   BL₂   BL₃
          │     │     │     │
WL₀ ──────●─────●─────●─────●────→ I₀ = Σ G₀ⱼ×Vⱼ
          │     │     │     │
WL₁ ──────●─────●─────●─────●────→ I₁ = Σ G₁ⱼ×Vⱼ
          │     │     │     │
WL₂ ──────●─────●─────●─────●────→ I₂ = Σ G₂ⱼ×Vⱼ
          │     │     │     │
WL₃ ──────●─────●─────●─────●────→ I₃ = Σ G₃ⱼ×Vⱼ

Output vector y read as currents on rows:
    y = [I₀, I₁, I₂, I₃]

ALL computations happen SIMULTANEOUSLY!
N² multiplications + N×(N-1) additions in O(1) time
```

### 4.3 Time-Domain Encoding (Advanced)

Instead of voltage-domain, currents can encode in time:

```
Traditional:        Time-Domain:
V × G = I          t_activation × G ∝ accumulated charge

Input:  Voltage    Input:  Pulse timing/duration
Weight: Conductance Weight: Conductance (same)
Output: Current    Output: Charge (time-integrated)

Advantage: Time-to-digital converter (TDC) replaces ADC
           → Lower power, simpler circuits
```

### 4.4 FeFET-Specific MVM Implementation

```
1FeFET-1R Cell Architecture (TUM):

    Gate Voltage (time-encoded input)
           │
           ▼
    ┌─────────────┐
    │   FeFET     │ ← VTH encodes weight
    └──────┬──────┘
           │
         ─┴─ 1 MΩ resistor (current limiting)
           │
           ▼
    Accumulated current → Capacitor (64 fF) → Comparator

Equations:
- Cell current: ID = f(VG, VTH)
- Accumulation: V_cap = (1/C) ∫ Σ ID,n(t) dt
- Output: 2-bit via StrongArm comparators
```

### 4.5 MVM Accuracy Metrics

| Implementation | MNIST Accuracy | CIFAR-10 Accuracy | Notes |
|----------------|----------------|-------------------|-------|
| **Ideal (floating point)** | 99.11% | 93.22% | Baseline |
| **FeFET CIM (4-state)** | 96.64% | 91.55% | 40 mV variation |
| **2D Ferroelectric** | ~99% | N/A | <0.5% variation |
| **Passive 0T1R** | 85-92% | 75-85% | Sneak path error |
| **1T1R Architecture** | 96-98% | 88-91% | Clean signals |

### 4.6 Energy and Performance

| Metric | FeFET CIM | Digital GPU | Improvement |
|--------|-----------|-------------|-------------|
| **Energy/MAC** | ~100 fJ | ~1000 fJ | 10× |
| **Energy Efficiency** | 885 TOPS/W | 1-10 TOPS/W | 100-1000× |
| **Latency/MVM** | ~76 ns | µs-ms | 1000× |
| **Power (typical)** | 153.6 µW | 100+ W | 10⁶× |

### 4.7 Handling Signed Weights

Neural networks require both positive and negative weights. Solutions:

#### Method 1: Differential Pair

```
Positive weight: G⁺     Negative weight: G⁻
      │                       │
      ▼                       ▼
  ┌───────┐               ┌───────┐
  │ Cell+ │               │ Cell- │
  └───┬───┘               └───┬───┘
      │                       │
      └───────────┬───────────┘
                  │
                  ▼
        I_diff = G⁺×V - G⁻×V = (G⁺ - G⁻)×V

Weight W = G⁺ - G⁻ (can be negative!)
```

#### Method 2: 2's Complement with Dual ADCs

```
Signed weight decomposition:
W = W_sign × |W_magnitude|

Use:
- 1pFeFET for sign bits (W_sign)
- 1nFeFET for magnitude bits (|W_magnitude|)
- Dual 2CM/N2CM ADCs for signed/unsigned segments
```

#### Method 3: Bit-Slicing

```
8-bit weight decomposed into 4× 2-bit slices:

W[7:0] = W[7:6]×64 + W[5:4]×16 + W[3:2]×4 + W[1:0]×1

Each slice stored in separate crossbar row.
Digital shift-add combines partial products.

Advantage: Handles signed weights without differential cells
Disadvantage: 4× more rows, digital accumulation needed
```

---

## 5. Putting It All Together: Signal Flow

### 5.1 Complete Write Path

```
WRITE OPERATION: Store level 15 to cell (2,3)

Step 1: Digital to Analog Conversion
┌─────────────┐
│  Level 15   │  Digital input
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    DAC      │  5-bit DAC
│  15 → 0.5V  │  Maps level to base voltage
└──────┬──────┘
       │
       ▼
┌─────────────┐
│ Charge Pump │  Boost voltage
│ 0.5V → 1.5V │  (1V CMOS → ±1.5V)
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Crossbar   │  Apply to selected cell
│  Cell (2,3) │  Row 2, Column 3
│  WL₂=1.5V   │  Word line selected
│  BL₃=GND    │  Bit line grounded
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Verify    │  Read back and check
│ Level = 15? │  If not, increment voltage
└─────────────┘

Timing: DAC(10ns) + Pump(88ns) + Write(100ns) + Array(5ns) ≈ 203ns
Energy: ~2.15 pJ total (pump-dominated)
```

### 5.2 Complete Read Path

```
READ OPERATION: Read level from cell (2,3)

Step 1: Select Cell
┌─────────────┐
│  Crossbar   │  Apply read bias
│  Cell (2,3) │  WL₂ = 1.0V (read voltage)
│  VG = 1.0V  │  BL₃ = 0.1V (drain bias)
│  VD = 0.1V  │
└──────┬──────┘
       │ ID = 50 µA (corresponds to level 15)
       ▼
┌─────────────┐
│    TIA      │  Current to voltage
│  50µA → 0.5V│  Gain = 10 kΩ
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    ADC      │  Analog to digital
│  0.5V → 15  │  5-bit SAR ADC
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Level 15   │  Digital output
└─────────────┘

Timing: DAC(10ns) + Array(5ns) + TIA(11ns) + ADC(50ns) ≈ 76ns
Energy: ~46 fJ total
```

### 5.3 Complete MVM Path

```
COMPUTE OPERATION: y = W × x (4×4 example)

Step 1: Convert Input Vector to Voltages
┌─────────────────────────────────────────┐
│  x = [x₀, x₁, x₂, x₃]                   │  Digital inputs
│        │    │    │    │                 │
│        ▼    ▼    ▼    ▼                 │
│      DAC  DAC  DAC  DAC                 │  4 parallel DACs
│        │    │    │    │                 │
│      V₀   V₁   V₂   V₃                  │  Voltage inputs
└────────┼────┼────┼────┼─────────────────┘
         │    │    │    │
         ▼    ▼    ▼    ▼
Step 2: Apply to Crossbar (ALL columns simultaneously)
┌─────────────────────────────────────────┐
│      V₀   V₁   V₂   V₃                  │
│       │    │    │    │                  │
│  WL₀ ─●────●────●────●─────→ I₀         │
│       │    │    │    │                  │
│  WL₁ ─●────●────●────●─────→ I₁         │  MVM in O(1)!
│       │    │    │    │                  │
│  WL₂ ─●────●────●────●─────→ I₂         │
│       │    │    │    │                  │
│  WL₃ ─●────●────●────●─────→ I₃         │
│                                         │
└──────────────────────────┬──────────────┘
                           │
Step 3: Sense Output Currents
┌──────────────────────────┼──────────────┐
│                  I₀  I₁  I₂  I₃         │
│                   │   │   │   │         │
│                   ▼   ▼   ▼   ▼         │
│                 TIA TIA TIA TIA         │  4 parallel TIAs
│                   │   │   │   │         │
│                   ▼   ▼   ▼   ▼         │
│                 ADC ADC ADC ADC         │  4 parallel ADCs
│                   │   │   │   │         │
│                   ▼   ▼   ▼   ▼         │
│           y = [y₀, y₁, y₂, y₃]          │  Digital outputs
└─────────────────────────────────────────┘

Timing: DAC(10ns) + Array(5ns) + TIA(11ns) + ADC(50ns) ≈ 76ns
Energy: ~184 fJ for 4×4 MVM (16 MACs)
Energy per MAC: ~11.5 fJ (100× better than digital!)
```

---

## 6. Comparison: Voltage, Current, Time, and Charge Domains

### 6.1 Computing Domain Comparison

| Domain | Input Encoding | Output Encoding | Key Component | Pros | Cons |
|--------|---------------|-----------------|---------------|------|------|
| **Voltage** | Voltage level | Current (via TIA → ADC) | SAR ADC | Mature, accurate | ADC power hungry |
| **Current** | Current pulses | Accumulated charge | Charge amplifier | Direct MAC | Sensitive to noise |
| **Time** | Pulse timing/width | Activation time | TDC | ADC-less, scalable | Timing precision critical |
| **Charge** | Pre-charged caps | Charge sharing | Capacitors | Best variation tolerance | Slower, needs reset |

### 6.2 ADC-Less Architectures

```
Traditional (Voltage-Domain):
─────────────────────────────────────
Input → Crossbar → TIA → ADC → Output
                         ↑
                    Power hog (50-80%)

ADC-Less (Time-Domain):
─────────────────────────────────────
Input (timing) → Crossbar → Comparator → TDC → Output
                                         ↑
                               Simple counter, low power

Power Savings: 50% or more vs. ADC-based
Tradeoff: Slower (integration time), precision limited
```

---

## 7. Key Design Parameters for 30-Level FeCIM

### 7.1 Recommended Specifications

| Component | Parameter | Value | Rationale |
|-----------|-----------|-------|-----------|
| **ADC** | Bits | 5 | 32 levels ≥ 30 FeCIM levels |
| | Type | SAR | Best energy efficiency |
| | Conversion Time | 50 ns | Matches crossbar speed |
| | INL/DNL | <0.5 LSB | Preserves level accuracy |
| **DAC** | Bits | 5 | Match ADC resolution |
| | Output Range | ±1.5 V | FeFET switching voltage |
| | Settling | 10 ns | Fast write initiation |
| **TIA** | Gain | 10 kΩ | 1 µA → 10 mV mapping |
| | Bandwidth | 100 MHz | Fast settling |
| | Noise | 1 pA/√Hz | SNR > 30 dB |
| **Charge Pump** | Input | 1.0 V | CMOS supply |
| | Output | ±1.5 V | Write voltage |
| | Efficiency | 70% | Minimize power loss |

### 7.2 Operation Voltage Summary

| Operation | Voltage | Duration | Energy | Notes |
|-----------|---------|----------|--------|-------|
| **READ** | 0.5-1.0 V | 76 ns | ~46 fJ | Non-destructive |
| **WRITE** | 1.2-1.5 V | 203 ns | ~2.15 pJ | Per cell (pump-dominated) |
| **COMPUTE** | 0-1.0 V (input) | 76 ns | ~46 fJ/row read | Array-wide |

---

## 8. Sources and References

### 8.1 READ Operations
- [Imec Non-Destructive Readout (2024)](https://www.imec-int.com/en/articles/non-destructive-readout-mechanism-ferroelectric-capacitors-0)
- [TUM Multi-level FeFET - Nature Communications 2023](https://pmc.ncbi.nlm.nih.gov/articles/PMC10564859/)
- [3D Ferroelectric Memory Architectures (2025)](https://arxiv.org/html/2504.09713v1)
- [Dual-Port FeFET for Disturb-Free Operation (2023)](https://arxiv.org/abs/2305.01484)

### 8.2 WRITE Operations
- [HfO₂ FeFET Review - Journal of Applied Physics 2024](https://pubs.aip.org/aip/jap/article/138/1/010701/3351745/)
- [Sub-Nanosecond Switching - Nano Letters 2024](https://pubs.acs.org/doi/abs/10.1021/acs.nanolett.2c04706)
- [aJ-Level Boolean Logic - Nano Letters 2024](https://pubs.acs.org/doi/10.1021/acs.nanolett.4c02873)
- [1000+ Polarization States - Nature Electronics 2025](https://www.nature.com/articles/s41928-025-01551-7)

### 8.3 COMPUTE Operations
- [Analog Matrix Solving - Science Advances 2025](https://pmc.ncbi.nlm.nih.gov/articles/PMC11817932/)
- [FeFET CIM Energy Efficiency - arXiv 2024](https://arxiv.org/html/2410.19593v1)
- [Bit-Slicing Techniques - arXiv 2024](https://arxiv.org/html/2512.18459v1)
- [Hybrid Digital-Analog CIM - Science Advances 2024](https://www.science.org/doi/10.1126/sciadv.adp0174)

### 8.4 Emerging Architectures
- [ADC-Less Time-Domain CIM - ACM GLSVLSI 2024](https://dl.acm.org/doi/10.1145/3649476.3658773)
- [Ferroelectric Capacitive Memories - Nano Convergence 2024](https://link.springer.com/article/10.1186/s40580-024-00463-0)
- [FeCIM Annealer - Nature Communications 2024](https://www.nature.com/articles/s41467-024-46640-x)

---

## Related Documentation

- **[circuits.operations.md](circuits.operations.md)** — Detailed 0T1R vs 1T1R architecture comparison
- **[circuits.research.md](circuits.research.md)** — Peripheral circuits meta-study (ADC/DAC/TIA/Pump)
- **[circuits.ELI5.md](circuits.ELI5.md)** — Simple explanations for beginners
- **[circuits.opensource.md](circuits.opensource.md)** — Open-source simulation tools
- **[../crossbar/crossbar.research.md](../crossbar/crossbar.research.md)** — Crossbar array physics

---

**Part of:** FeCIM Lattice Tools - Ferroelectric Compute-in-Memory Visualization Suite
