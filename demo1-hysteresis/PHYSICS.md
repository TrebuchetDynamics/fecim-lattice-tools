# Ferroelectric Physics: From Absolute Basics

Start here if you've never studied ferroelectrics before.

---

## Part 1: What Are We Even Talking About?

### Atoms and Charges

Everything is made of atoms. Atoms have:
- **Protons (+)** in the center (positive charge)
- **Electrons (-)** orbiting around (negative charge)

When positive and negative charges are separated, we call this a **dipole**:

```
     Before                After applying force
   
     ⊕⊖                      ⊕─────⊖
   (neutral)              (dipole - charges separated)
```

### What is Polarization?

**Polarization (P)** = how much the charges are separated, on average, in a material.

```
Unpolarized crystal:       Polarized crystal:
┌────────────────┐         ┌────────────────┐
│ ⊕⊖  ⊕⊖  ⊕⊖  │         │ ⊕→⊖ ⊕→⊖ ⊕→⊖ │
│ ⊕⊖  ⊕⊖  ⊕⊖  │         │ ⊕→⊖ ⊕→⊖ ⊕→⊖ │  →→→ Net P
│ ⊕⊖  ⊕⊖  ⊕⊖  │         │ ⊕→⊖ ⊕→⊖ ⊕→⊖ │
└────────────────┘         └────────────────┘
   P = 0                      P > 0 (pointing right)
```

**Units of P:** microcoulombs per square centimeter (μC/cm²)
- 25 μC/cm² means: 25 microcoulombs of charge separation per cm² of material

### What is an Electric Field?

An **Electric Field (E)** is the "push" felt by charges in a region.

```
                  Electric Field E →
         ─────────────────────────────→
         ─────────────────────────────→
         ─────────────────────────────→

Positive charges feel pushed RIGHT →
Negative charges feel pushed LEFT ←
```

**Units of E:** megavolts per centimeter (MV/cm)
- 1 MV/cm = 1,000,000 volts across 1 centimeter of material

**Relationship:** If you apply 1V across a 10nm film:
```
E = Voltage / Thickness = 1V / 10nm = 1V / 10⁻⁶cm = 1 MV/cm
```

---

## Part 2: What Makes Ferroelectrics Special?

### Normal Materials (Dielectrics)

In most materials:
1. Apply electric field → charges separate (polarize)
2. Remove electric field → charges return to original position
3. **No memory!**

```
Normal material response:

P (polarization)
↑
│       /
│      /
│     /
│    /
│   /
├──/─────→ E (field)
│ 
Same path up and down!
```

### Ferroelectric Materials: They REMEMBER!

In ferroelectric materials:
1. Apply field → charges separate AND crystal structure shifts
2. Remove field → **charges stay separated!**
3. **MEMORY!** (like a light switch that stays on)

```
Crystal structure shift (simplified):

    Before field           After field (stays!)
    ┌───┬───┬───┐         ┌───┬───┬───┐
    │   │ ⊕ │   │         │   │   │   │
    │ ⊕ │   │ ⊕ │ ──E→→   │ ⊕ │ ⊕ │ ⊕ │ ← Center atom
    │   │ ⊕ │   │         │   │   │   │    moved UP
    └───┴───┴───┘         └───┴───┴───┘
    
    P = 0                  P > 0 (permanent!)
```

The center atom literally moves to a new stable position in the crystal lattice!

---

## Part 3: Hysteresis - The Loop

### What is Hysteresis?

**Hysteresis** = Greek for "lagging behind"

The output (P) doesn't just depend on the input (E)—it depends on the **history** of what happened before.

**Real-world examples of hysteresis:**
- Thermostat: Turns on at 68°F, off at 72°F (not same point!)
- Mechanical switch: Clicks on, clicks off at different positions
- Rubber band: Stretching path ≠ releasing path

### The P-E Hysteresis Loop

```
                  ③ Saturated (all dipoles aligned)
                     ╭───────╮
         P          ╱         ╲
         ↑         │           │
         │    ②   │           │   ④
     Ps ─┼────────●           ●───────  ← SATURATION
         │       ╱             ╲         (maximum possible P)
     Pr ─┼─────●               │        ← REMANENT
         │     │               │         (P when E=0, THE MEMORY!)
         ├─────┼───────────────┼────→ E
         │     │    ①          │
    -Pr ─┼─────┼───────────────●        
         │     ╲               ╱
    -Ps ─┼─────────●           ●───────
         │          ╲         ╱
         │           ╰───────╯
                         ⑤
              
              -Ec    0    +Ec
                     ↑
              COERCIVE FIELD
        (field needed to SWITCH direction)
```

### Walking Through the Loop

| Step | What happens |
|------|--------------|
| ① | Start: P = +Pr (positive remanent state, no field applied) |
| ② | Apply +E: P increases toward saturation |
| ③ | At high +E: Saturated, P = Ps (all dipoles aligned up) |
| ④ | Reduce E to 0: P drops to Pr (STILL POSITIVE! Memory!) |
| ⑤ | Apply -E: P decreases, crosses through 0 at -Ec |
| ⑥ | At high -E: Saturated negative, P = -Ps |
| ⑦ | Reduce E to 0: P = -Pr (negative remanent state) |
| ⑧ | Apply +E: Must reach +Ec to flip back positive |

### Key Parameters Explained

| Parameter | Name | Meaning | HZO Value |
|-----------|------|---------|-----------|
| **Ps** | Saturation Polarization | Maximum possible separation of charges | 25 μC/cm² |
| **Pr** | Remanent Polarization | Polarization remaining at zero field (THE MEMORY) | ~20 μC/cm² |
| **Ec** | Coercive Field | Field required to flip the polarization | 1.0 MV/cm |

**In plain terms:**
- **Ps = 25 μC/cm²** → "When all dipoles align, we get this much charge separation"
- **Ec = 1.0 MV/cm** → "Need 1 million volts per centimeter to flip the switch"
  - For 10nm film: 1.0 MV/cm × 10nm = 1.0V needed to switch!

---

## Part 4: Why 30 States, Not Just 2?

### Binary Memory (Traditional)

Normal flash memory: ON or OFF (1 or 0)
```
     P
     ↑
 +Pr ├────● State 1 ("ON")
     │
   0 ├────
     │
 -Pr ├────● State 0 ("OFF")
```

### Analog Memory (IronLattice)

Ferroelectrics can be set to IN-BETWEEN values!

```
     P
     ↑
 +Ps ├────● State 30
     ├────● State 29
     ├────● State 28
     ├    ⋮
     ├────● State 16
   0 ├────● State 15
     ├    ⋮
     ├────● State 2
     ├────● State 1
 -Ps ├────● State 0
```

**How?** By stopping at different points on the hysteresis curve using precisely controlled voltage pulses.

**Why useful?** 
- Each cell stores 5 bits instead of 1 bit (log₂(30) ≈ 5)
- For AI: Can represent neural network weights directly (analog compute)

---

## Part 5: The Hysteron Concept

### What is a Hysteron?

A **hysteron** is the simplest possible element with hysteresis: a switch that turns ON and OFF at **different** thresholds.

```
Think of it like a sticky light switch:

         Output (on/off)
           ↑
         1 ├────────────────╮
           │                │
           │    α (ON)      │
         0 ├────────────────┼───────→ Input
           │    β (OFF)     │
           │                │
        -1 ├────────────────╯

α = 2V to turn ON
β = 1V to turn OFF

So: ON at 2V, OFF at 1V, NOT THE SAME!
```

### Material = Many Hysterons

Real ferroelectric = millions of tiny domains, each acting like a hysteron with slightly different (α, β):

```
   One big loop = sum of many small hysterons

   ╭──╮   =   [╭╮] + [╭╮] + [╭╮] + ... millions
  ╱    ╲       α₁β₁   α₂β₂   α₃β₃
 │      │
  ╲    ╱
   ╰──╯
```

This is the **Preisach model**: The macroscopic hysteresis loop emerges from the sum of microscopic hysterons.

---

## Part 6: Minor Loops

### What if We Don't Complete the Full Cycle?

If you only go partway around the loop and reverse, you get a **minor loop**:

```
Full major loop:              Minor loop:
      ╭───────╮                  ╭───╮
     ╱         ╲                ╱ ╭←╯
    │           │              │  │ Turned back
    │     ●     │              │  ↓ early!
     ╲         ╱                ╲   
      ╰───────╯                  ╰─
```

**The Preisach model handles this** by tracking "turning points" (where you reversed direction).

**Why it matters:** In real memory operation, you might do partial writes, and the physics must correctly predict what happens.

---

## Summary Table

| Term | Plain English | Unit |
|------|---------------|------|
| **Polarization (P)** | How much positive/negative charges are separated | μC/cm² |
| **Electric Field (E)** | The "push" on charges from applied voltage | MV/cm |
| **Saturation (Ps)** | Maximum possible polarization | μC/cm² |
| **Remanent (Pr)** | Polarization that remains when E=0 (the memory!) | μC/cm² |
| **Coercive Field (Ec)** | Field needed to flip the polarization direction | MV/cm |
| **Hysteresis** | Output depends on history, path up ≠ path down | - |
| **Hysteron** | One tiny switch element with ON/OFF thresholds | - |
| **Preisach Model** | Many hysterons with distributed thresholds = one loop | - |

---

## What Demo 1 Visualizes

With this understanding, Demo 1 shows:

1. **The P-E Loop** - As you drag voltage left/right, watch P trace the hysteresis curve
2. **The 30 States** - See which analog level you're at based on P value
3. **Minor Loops** - Reverse direction partway and see the inner loops form
4. **Material Comparison** - Different Ec, Ps values → different loop shapes

---

## Part 7: How Demo 1 Actually Implements the Physics

This section documents exactly what the code does — verified by source analysis.

### Core Model: Mayergoyz Preisach

The demo uses the **classical Preisach model** (not tanh approximation). The implementation is in `pkg/ferroelectric/preisach_advanced.go`.

**Key insight:** The macroscopic P-E loop EMERGES from many microscopic hysterons, each with its own switching thresholds.

### Hysteron Definition

```go
type Hysteron struct {
    Alpha float64 // Field where hysteron switches UP (+1)
    Beta  float64 // Field where hysteron switches DOWN (-1)
    State int     // Current state: +1 or -1 (persists between thresholds)
}
```

### How P is Calculated from E

The core physics happens in `Update()` (lines 166-192):

```go
func (m *MayergoyzPreisach) Update(E float64) float64 {
    // Step 1: Update each hysteron's state
    for i := range m.hysterons {
        if E >= m.hysterons[i].Alpha {
            m.hysterons[i].State = +1  // Switch UP
        } else if E <= m.hysterons[i].Beta {
            m.hysterons[i].State = -1  // Switch DOWN
        }
        // Between Beta and Alpha: state UNCHANGED (memory effect!)
    }

    // Step 2: Sum contributions: P = Σ μ(αᵢ, βᵢ) × γᵢ
    m.polarization = 0
    for i, h := range m.hysterons {
        m.polarization += m.distribution[i][0] * float64(h.State)
    }

    return m.polarization
}
```

### Where Hysteresis Comes From

**The hysteresis is EMERGENT, not forced.** Here's why:

1. Each hysteron has Alpha > Beta (e.g., α = +1.1 Ec, β = -0.9 Ec)
2. When E increases past Alpha → hysteron switches to +1
3. When E decreases past Beta → hysteron switches to -1
4. **Between Beta and Alpha: the state PERSISTS** — this is the memory

```
For one hysteron with α = 1.2, β = -1.0:

          E increasing →
State: -1 ─────────────┬───────── +1
                       │α=1.2
                       │
          ← E decreasing
State: +1 ─────────────────────┬─ -1
                              │β=-1.0

The gap between α and β is where hysteresis lives!
```

### Hysteron Distribution (Why the Loop is Square-ish)

Hysterons are distributed on the Preisach plane with a 2D Gaussian:

```go
AlphaMean:   material.Ec,        // Centers positive thresholds at +Ec
BetaMean:    -material.Ec,       // Centers negative thresholds at -Ec
AlphaSigma:  material.Ec * 0.2,  // 20% spread
BetaSigma:   material.Ec * 0.2,  // 20% spread
```

**Narrow σ (20%) = sharp switching = square loop.**
A wider σ would give a more slanted/soft loop.

### How 30 Levels Are Discretized

The continuous polarization P is mapped to discrete levels in the GUI loop (`gui.go:416`):

```go
a.discreteLevel = int(math.Round((a.normalizedP + 1) / 2 * 29))
```

Where `normalizedP = P / Ps` ranges from -1 to +1.

| Normalized P | Level |
|--------------|-------|
| -1.0 (−Ps)   | 0     |
| 0.0          | 15    |
| +1.0 (+Ps)   | 29    |

**Formula:** `Level = round((P/Ps + 1) × 14.5)` = 0 to 29

### Does τ (Switching Time) Affect the Visualization?

**No — the real-time loop uses instantaneous switching.**

The simulation runs at 60 FPS (`dt ≈ 16ms`) and calls:
```go
a.polarization = a.preisach.Update(a.electricField)
```

This `Update()` switches hysterons instantaneously when E crosses their thresholds.

The τ = 10 ns switching time IS defined in the material and there IS a `SimulateDomainSwitching()` function using KAI (Kolmogorov-Avrami-Ishibashi) dynamics:

```go
// KAI model: progress = 1 - exp(-(t/τ)^n)
// n = 2.0 (Avrami exponent for 2D domain growth)
```

But this function is **not called** during the interactive visualization loop. This is physically reasonable: at 1 Hz cycling, τ = 10 ns is negligible (the system is always in equilibrium).

### Temperature Dependence

The coercive field scales with temperature:

```go
Ec(T) = Ec₀ × (1 - T/Tc)^0.5
```

Where Tc = 723 K (~450°C) is the Curie temperature. Above Tc, the material loses ferroelectricity (Ec → 0).

---

## Summary: What's Real vs. Simplified

| Aspect | Implementation | Status |
|--------|---------------|--------|
| P from E | Preisach model (hysteron sum) | ✅ Physics-accurate |
| Hysteresis | Emergent from hysteron memory | ✅ Physics-accurate |
| Loop shape | From Gaussian distribution (σ=20%) | ✅ Emergent, not forced |
| 30 levels | Linear discretization of P | ✅ Simple & correct |
| Minor loops | Implicit via hysteron states | ✅ Works correctly |
| τ switching | Defined but not used in viz | ⚠️ Quasistatic approx |
| Temperature | Ec(T) scaling implemented | ✅ Physics-accurate |
