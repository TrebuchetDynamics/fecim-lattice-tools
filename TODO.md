# IronLattice-vis TODO

> Based on Dr. external research group's November 2024 presentation on IronLattice technology.
> Source: ironlattice-transcript.md

---

## IronLattice Key Specs (From Dr. Tour)

| Metric | Target | Current Status |
|--------|--------|----------------|
| Discrete analog states | **30 levels** | ✅ Implemented |
| MNIST accuracy | **87%** (88% theoretical max) | ✅ **95.8%** achieved |
| Energy vs NAND | 10,000,000× lower | N/A (simulation) |
| Energy vs DRAM | 1,000× lower | N/A (simulation) |
| P-E hysteresis | Square loop characteristic | Simplified tanh model |
| CMOS compatible | Standard fab | N/A |
| TRL | 4 (lab validation) | Demo/educational |

---

## 8-Demo Roadmap

```
Demo 1 ──→ Demo 2 ──→ Demo 3 ──→ Demo 4 ──→ Demo 5 ──→ Demo 6 ──→ Demo 7 ──→ Demo 8
(cell)    (array)    (app)     (system)  (thermal)  (3D)      (real)    (compare)

Physics ──→ Computation ──→ Application ──→ Engineering ──→ Production Reality
```

| Demo | Name | Purpose | Audience | Status |
|------|------|---------|----------|--------|
| 1 | Hysteresis | Single cell physics | Everyone | ✅ Complete |
| 2 | Crossbar MVM | Compute-in-memory | Engineers | ✅ Complete |
| 3 | MNIST | AI application | Investors | ✅ Complete (95.8%) |
| 4 | Peripherals | Full system | Foundries | ✅ Complete |
| 5 | Thermal | Heat analysis | Engineers | ✅ Complete |
| 6 | Multi-Layer 3D | Architecture | Designers | 🔲 Planned |
| 7 | Non-Idealities | Real-world issues | Engineers | 🔲 Planned |
| 8 | Comparison | Why IronLattice wins | Investors | 🔲 Planned |

---

## Phase 1: Core Demos ✅ COMPLETE

### Demo 1: Ferroelectric Hysteresis ✅
- [x] P-E hysteresis curve visible
- [x] 30 discrete levels clearly shown (LevelIndicator)
- [x] Interactive E-field control
- [x] Preisach model with HZO parameters
- [x] Thread-safe simulation engine
- [ ] Square loop characteristic (IronLattice advantage) - enhancement

### Demo 2: Crossbar MVM ✅
- [x] Matrix-vector multiplication works
- [x] 30-level conductance states
- [x] Input/output visualization
- [x] Terminal display with color coding
- [ ] Animated voltage/current flow - enhancement

### Demo 3: MNIST Classification ✅
- [x] Can classify handwritten digits
- [x] **Achieves 87% accuracy** → **95.8%!**
- [x] Uses 30 discrete weight levels
- [x] Interactive drawing/testing
- [x] Pretrained weights saved to data/pretrained_weights.json

---

## Phase 2: System Integration ✅ COMPLETE

### Demo 4: Peripheral Circuits ✅

**Purpose:** Show investors/foundries the full chip system

```
WRITE PATH                 READ PATH

Level: [22]               Current: [67 μA]
    │                          ↑
    ▼                          │
┌───────┐                  ┌───────┐
│  DAC  │                  │  TIA  │
│ 5-bit │                  │       │
└───┬───┘                  └───┬───┘
    │                          ↑
    ▼                          │
┌───────┐                  ┌───────┐
│ Charge│                  │  ADC  │
│ Pump  │                  │ 5-bit │
└───┬───┘                  └───────┘
    │
    ▼
┌─────────────────┐
│    CROSSBAR     │
└─────────────────┘
```

**Features implemented:**
- [x] DAC visualization: Digital → Write voltage (5-bit, 30 levels)
- [x] Charge pump: 1V → ±1.5V for write operations
- [x] TIA (Transimpedance Amplifier): Current → Voltage conversion
- [x] ADC visualization: Analog → Digital level (5-bit)
- [x] Noise injection visualization
- [x] Show CMOS compatibility (standard process)
- [x] Energy consumption display per operation

**Technical approach:**
```go
// pkg/peripherals/dac.go
type DAC struct {
    Bits       int     // 5 bits for 30 levels
    VrefHigh   float64 // +1.5V
    VrefLow    float64 // -1.5V
}

func (d *DAC) Convert(level int) float64 {
    return d.VrefLow + float64(level)/float64((1<<d.Bits)-1)*(d.VrefHigh-d.VrefLow)
}
```

---

### Demo 5: Thermal Simulation ✅

**Purpose:** Show engineers heat management is solved

```
Top View (Heat Map)        Side View

░░░▒▒▓▓████▓▓▒▒░░░        ███ Layer 3
░░▒▒▓██████████▓▒▒░░       ↕ heat
░▒▓████████████████▓▒░     ███ Layer 2
░░▒▒▓██████████▓▒▒░░       ↕ heat
░░░▒▒▓▓████▓▓▒▒░░░         ███ Layer 1
                           ░░░ Heat Sink

25°C ░▒▓█ 85°C
```

**Features implemented:**
- [x] 2D heat map visualization (terminal with color)
- [x] Real-time heat diffusion simulation
- [x] Multi-layer heat coupling
- [x] Hotspot identification and highlighting
- [x] Thermal throttling warning system
- [x] Workload-dependent heat generation
- [x] Show IronLattice's low-power advantage (1000x cooler!)

**Technical approach:**
```go
// pkg/thermal/simulation.go
type ThermalSim struct {
    Grid        [][]float64 // Temperature grid
    Conductivity float64    // Thermal conductivity
    AmbientTemp  float64    // 25°C
    MaxTemp      float64    // 85°C threshold
}

func (t *ThermalSim) Step(powerMap [][]float64, dt float64) {
    // Heat diffusion equation: dT/dt = α∇²T + Q
}
```

---

## Phase 3: Full Vision

### Demo 6: Multi-Layer 3D Architecture 🔲

**Purpose:** Show designers the scalable architecture

```
         ╔════════════════════╗
        ╱ Layer 3: 64×10     ╱│
       ╔════════════════════╗ │
      ╱ Layer 2: 128×64    ╱│ │
     ╔════════════════════╗ │ │
     ║ Layer 1: 784×128   ║ │╱
     ║  ●  ●  ●  ●  ●  ● ║╱
     ╚════════════════════╝
              ↑
          Input (784)
```

**Features to implement:**
- [ ] 3D rendered multi-layer stack (Vulkan)
- [ ] Via connections between layers
- [ ] Heat overlay integration (from Demo 5)
- [ ] Exploded view mode for inspection
- [ ] Design space exploration (layer sizes)
- [ ] Data flow animation through layers
- [ ] Memory density calculations

---

### Demo 7: Non-Idealities 🔲

**Purpose:** Show engineers we understand real-world challenges

```
IR Drop:           1.0V → 0.95V → 0.90V → 0.85V
Sneak Paths:       Current shortcuts through array
Conductance Drift: Level 15 → Level 14.8 (1 week)
Variation:         Write 15: [14, 15, 15, 16, 15, 14]
```

**Features to implement:**
- [ ] IR drop visualization across array
- [ ] Sneak path current animation
- [ ] Conductance drift over simulated time
- [ ] Cycle-to-cycle variation (write noise)
- [ ] Device-to-device variation
- [ ] Impact on accuracy (real-time display)
- [ ] Mitigation strategies visualization

**Technical approach:**
```go
// pkg/nonidealities/irdrop.go
func ComputeIRDrop(array *crossbar.Array, wireResistance float64) [][]float64 {
    // Model voltage drop: V(x) = V0 - I*R*x
}

// pkg/nonidealities/sneakpath.go
func AnalyzeSneakPaths(array *crossbar.Array, targetCell [2]int) [][]float64 {
    // Find parasitic current paths
}
```

---

### Demo 8: Technology Comparison 🔲

**Purpose:** Investor pitch — why IronLattice wins

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│    DRAM     │  │    GPU      │  │ IronLattice │
│    +CPU     │  │   (CUDA)    │  │    (CIM)    │
├─────────────┤  ├─────────────┤  ├─────────────┤
│ Time: 100μs │  │ Time: 10μs  │  │ Time: 0.1μs │
│ Energy: 100 │  │ Energy: 50  │  │ Energy: 0.1 │
│ Steps: 1000 │  │ Steps: 100  │  │ Steps: 1    │
└─────────────┘  └─────────────┘  └─────────────┘
```

**Features to implement:**
- [ ] Side-by-side comparison animation
- [ ] DRAM+CPU vs GPU vs IronLattice
- [ ] Time metric comparison
- [ ] Energy metric comparison
- [ ] Operation count comparison
- [ ] Scalable matrix size demonstration
- [ ] Data center savings projection

**Comparison metrics:**
| Metric | DRAM+CPU | GPU | IronLattice |
|--------|----------|-----|-------------|
| Memory bandwidth | 100 GB/s | 1 TB/s | ∞ (in-situ) |
| Energy per MAC | 10 pJ | 1 pJ | 0.001 pJ |
| Latency | 100 ns | 10 ns | 1 ns |
| Data movement | O(n²) | O(n²) | 0 |

---

## Code Quality Tasks

### Critical Bugs ✅ COMPLETED
- [x] **Race conditions** (engine.go) - Added sync.RWMutex
- [x] **O(n³) weight updates** (network.go) - Fetch matrix once
- [x] **Panics in production** (network/network.go:117) - Replaced with error returns

### Test Coverage ✅ 54 TESTS
- [x] 30-level quantization tests (7 tests)
- [x] MVM output verification (included in array_test.go)
- [x] Engine thread-safety tests (5 tests)
- [x] Network forward/backward tests (7 tests)
- [x] Weight save/load roundtrip (included in network_test.go)
- [x] MNIST accuracy >= 85% on test set (95.8% achieved)
- [x] P-E curve hysteresis verification (7 tests)
- [x] Peripheral circuits tests (9 tests)
- [x] Thermal simulation tests (17 tests)

---

## Educational Enhancements

### "Why CIM?" Panel
> "This could lower the requirements in a data center by 80 to 90%." — Dr. Tour

- [ ] Traditional architecture diagram (memory ↔ CPU bottleneck)
- [ ] CIM architecture diagram (compute happens in memory)
- [ ] Energy comparison visualization
- [ ] "30 states vs binary: ~5 bits per cell vs 1 bit" explanation

### Market Context
- [ ] Comparison table from Dr. Tour's slides
- [ ] TRL progression: "We are here (TRL 4) → Production (TRL 9)"

---

## Repository Structure

```
ironlattice-vis/
├── demo1-hysteresis/     ✅ Single cell P-E curve
├── demo2-crossbar/       ✅ Crossbar MVM visualization
├── demo3-mnist/          ✅ MNIST classifier (95.8%)
├── demo4-circuits/       ✅ Peripheral circuits (DAC, ADC, TIA, Charge Pump)
├── demo5-thermal/        ✅ Thermal simulation (1000x cooler operation)
├── demo6-multilayer/     🔲 3D multi-layer
├── demo7-nonidealities/  🔲 Real-world issues
├── demo8-comparison/     🔲 Technology comparison
├── docs/                 Documentation
├── papers/               Scientific papers
└── go.mod
```

---

## Success Criteria

### From Dr. Tour's Presentation
- [x] 30 discrete levels (not binary)
- [x] 87% MNIST accuracy (achieved 95.8%)
- [x] Compute-in-memory demonstration
- [ ] Square hysteresis loops (IronLattice advantage)
- [ ] Full system with peripherals
- [ ] Comparison showing 1000× energy advantage

---

## References

- **Primary Source**: Dr. external research group, IronLattice presentation (Nov 2024)
- **Key Paper**: Shin, J., et al. "BEOL-Compatible Superlattice FEFET Analog Synapse" IEEE (2022)
- **MNIST Benchmark**: 88% theoretical maximum, 87% achieved by IronLattice, **95.8% by this demo**
- **30 States**: Post-synaptic current with 30 discrete levels (LTP/LTD demonstration)

---

## Notes from Dr. Tour's Presentation

> "It's got **30 discrete states**. So it's not 0-1-0-1."

> "We're at **87% validation** here... theoretical is 88% is the theoretical maximum."

> "**Compute in memory** where the same device does the memory and the computation."

> "This could lower the requirements in a data center by **80 to 90%** of the energy requirements."

> "Works on a **standard CMOS line** and can translate just like that."

> "There's **no exotic materials** in here. There's no graphene."
