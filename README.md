# IronLattice Visualizer

**GPU-Accelerated Ferroelectric Compute-in-Memory Visualization**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Vulkan](https://img.shields.io/badge/Vulkan-1.3-AC162C?logo=vulkan)](https://www.vulkan.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Demos](https://img.shields.io/badge/Demos-3%2F8-blue.svg)]()

---

> ⚠️ **IMPORTANT DISCLAIMER**: IronLattice is at **TRL 4** (lab validation only). Performance claims in this visualization project include both **verified hardware results** (87% MNIST) and **simulation results** that may exceed real hardware capabilities. Energy claims (10M× vs NAND) are from Dr. Tour's presentation and have not been independently verified. See [HONESTY_AUDIT.md](opensource/papers/08_Documentation/HONESTY_AUDIT.md) for details.

---

## Vision: 8 Demos, Complete Story

```
┌─────────────────────────────────────────────────────────────┐
│                    IRONLATTICE-VIS                          │
│         GPU-Accelerated Ferroelectric CIM Visualization     │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  Demo 1 ──→ Demo 2 ──→ Demo 3 ──→ Demo 4 ──→ Demo 5 ──→ ...│
│  (cell)    (array)    (app)     (system)  (thermal)        │
│                                                             │
│  Physics ──→ Computation ──→ Application ──→ Engineering    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

| Demo | Purpose | Audience | Status |
|------|---------|----------|--------|
| **1. Hysteresis** | Single cell physics | Everyone | ✅ Complete |
| **2. Crossbar MVM** | Compute-in-memory | Engineers | ✅ Complete |
| **3. MNIST** | AI application | Investors | ✅ Complete (sim) |
| **4. Peripherals** | Full system | Foundries | 🔲 Planned |
| **5. Thermal** | Heat analysis | Engineers | 🔲 Planned |
| **6. Multi-Layer 3D** | Architecture | Designers | 🔲 Planned |
| **7. Non-Idealities** | Real-world issues | Engineers | 🔲 Planned |
| **8. Comparison** | Why IronLattice wins | Investors | 🔲 Planned |

---

## Quick Start

```bash
# Demo 1: Vulkan hysteresis visualization
cd demo1-hysteresis && go build -o hysteresis ./cmd/hysteresis && ./hysteresis

# Demo 2: Crossbar MVM visualization (terminal)
cd demo2-crossbar && go build -o inference ./cmd/inference && ./inference --show-mvm

# Demo 3: MNIST digit classifier (simulation)
cd demo3-mnist && go build -o mnist ./cmd/mnist && ./mnist --interactive
```

---

## The Technology

IronLattice represents a paradigm shift: **computation directly in memory** using ferroelectric superlattices, eliminating the Von Neumann bottleneck.

> *"This could lower the requirements in a data center by 80 to 90% of the energy requirements."*
> — Dr. external research group

| Spec | IronLattice Hardware | Our Simulation |
|------|---------------------|----------------|
| Analog states | 30 levels | ✅ 30 levels |
| MNIST accuracy | **87%** (88% max) | Variable* |
| Energy vs NAND | 10M× (claimed) | N/A |
| Energy vs DRAM | 1000× (claimed) | N/A |

*\*Simulation accuracy varies; idealized conditions may exceed hardware reality.*

---

## Demo Details

### Demo 1: Ferroelectric Hysteresis ✅

**Purpose:** Understand single cell physics

```
┌─────────┐      P                    ┌───────────┐
│         │      ↑     ╭────╮         │ ████ 30   │
│  CELL   │   +Pr├─────╯    │         │ ████ 29   │
│ (color) │      │          │         │ ▓▓▓▓ ...  │
│         │   ───┼──────────┼───→ E   │ ░░░░ 1    │
└─────────┘   -Pr├──────────╯         │      0    │
                 ↓                    │ 30 LEVELS │
                                      └───────────┘
```

**Features:**
- Real-time P-E hysteresis curve
- 30 discrete levels visualized
- Preisach model (statistical switches)
- Interactive E-field control
- HZO material parameters

**Run:** `cd demo1-hysteresis && go build -o hysteresis ./cmd/hysteresis && ./hysteresis`

---

### Demo 2: Crossbar Array MVM ✅

**Purpose:** Understand compute-in-memory

```
     V₀   V₁   V₂   V₃  (input voltages)
      │    │    │    │
 ─────●────●────●────●───→ I₀
      │    │    │    │
 ─────●────●────●────●───→ I₁  (output currents)
      │    │    │    │
 ─────●────●────●────●───→ I₂

 ●=conductance (30 levels, color coded)
```

**Physics:**
```
Ohm's Law:      I = V × G (per cell)
Kirchhoff:      I_col = Σ(V_row × G_cell)
Matrix form:    I = G × V (one clock cycle!)
```

**Run:** `cd demo2-crossbar && go build -o inference ./cmd/inference && ./inference --show-mvm`

---

### Demo 3: MNIST Neural Network ✅

**Purpose:** See real AI application

> **Note:** IronLattice hardware achieved **87%** with **88% theoretical max** (Dr. Tour). Our simulation may exceed this due to idealized conditions (no real IR drop, sneak paths, or process variation).

```
┌─────────┐    ┌─────────┐    ┌─────────┐
│ 28 × 28 │    │ 784×128 │    │ 128×10  │
│  INPUT  │ ─→ │ Layer 1 │ ─→ │ Layer 2 │ ─→ Prediction
│  DIGIT  │    │ Crossbar│    │ Crossbar│
└─────────┘    └─────────┘    └─────────┘
```

**Features:**
- 28×28 drawing canvas
- Two crossbar layers visualized
- Softmax probability bars
- Weight quantization to 30 levels
- Pretrained weights included

**Run:** `cd demo3-mnist && go build -o mnist ./cmd/mnist && ./mnist --interactive`

**Train:** `cd demo3-mnist && go run train_and_save.go`

---

### Demo 4: Peripheral Circuits 🔲

**Purpose:** Understand full system

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

**Planned Features:**
- DAC: Digital → Write voltage
- Charge pump: 1V → ±1.5V
- TIA: Current → Voltage
- ADC: Analog → Digital level
- Noise injection visualization
- CMOS compatibility demonstration

---

### Demo 5: Thermal Simulation 🔲

**Purpose:** Engineering analysis

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

**Planned Features:**
- 2D heat map visualization
- Real-time heat diffusion
- Multi-layer heat coupling
- Hotspot identification
- Thermal throttling warning

---

### Demo 6: Multi-Layer 3D Architecture 🔲

**Purpose:** Full system design

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

**Planned Features:**
- 3D rendered multi-layer stack
- Via connections between layers
- Heat overlay integration
- Exploded view mode
- Design space exploration

---

### Demo 7: Non-Idealities 🔲

**Purpose:** Real-world engineering challenges

```
IR Drop:           1.0V → 0.95V → 0.90V → 0.85V
Sneak Paths:       Current shortcuts through array
Conductance Drift: Level 15 → Level 14.8 (1 week)
Variation:         Write 15: [14, 15, 15, 16, 15, 14]
```

**Planned Features:**
- IR drop visualization
- Sneak path current animation
- Conductance drift over time
- Cycle-to-cycle variation
- Impact on accuracy (real-time)

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

**Planned Features:**
- Side-by-side comparison animation
- DRAM+CPU vs GPU vs IronLattice
- Time, energy, operations metrics
- Scalable matrix size

---

## Repository Structure

```
ironlattice-vis/
├── demo1-hysteresis/     ✅ Single cell P-E curve
├── demo2-crossbar/       ✅ Crossbar MVM visualization
├── demo3-mnist/          ✅ MNIST classifier (simulation)
├── demo4-circuits/       🔲 Peripheral circuits
├── demo5-thermal/        🔲 Thermal simulation
├── demo6-multilayer/     🔲 3D multi-layer
├── demo7-nonidealities/  🔲 Real-world issues
├── demo8-comparison/     🔲 Technology comparison
├── docs/                 Documentation
├── papers/               Scientific papers
└── go.mod
```

---

## The Story

```
Demo 1: "This is how the memory cell works"
Demo 2: "This is how we compute in memory"
Demo 3: "This is what we can build with it"
Demo 4: "This is how it fits in a real chip"
Demo 5: "This is how we manage heat"
Demo 6: "This is how we scale to 3D"
Demo 7: "This is what can go wrong (and how we fix it)"
Demo 8: "This is why it beats everything else"
```

---

## Build Timeline

### Phase 1: Core Demos ✅ Complete
- Demo 1: Hysteresis ✅
- Demo 2: Crossbar MVM ✅
- Demo 3: MNIST (simulation) ✅

### Phase 2: System Integration
- Demo 4: Peripheral Circuits
- Demo 5: Thermal Simulation

### Phase 3: Full Vision
- Demo 6: Multi-Layer 3D
- Demo 7: Non-Idealities
- Demo 8: Technology Comparison

---

## Technical Stack

| Component | Technology | Status |
|-----------|------------|--------|
| Language | Go 1.21+ | Ready |
| Graphics | Vulkan 1.3 | Working |
| Shaders | GLSL → SPIR-V | Working |
| Physics | Preisach model | Complete |
| Neural Network | Crossbar MVM | Complete |
| Tests | 19 passing | ✅ |

---

## The Team Behind IronLattice

| Person | Role |
|--------|------|
| **Dr. external research group** | Principal Investigator, external research institution |
| **Dr. Jaeho Shin** | Device Engineer, Superlattice Inventor |
| **Tawfik Jarjour** | Commercialization Lead |

---

## Key Quotes from Dr. Tour

> *"It's got 30 discrete states. So it's not 0-1-0-1."*

> *"We're at 87% validation here... theoretical is 88%."*

> *"Compute in memory where the same device does the memory and the computation."*

> *"This could lower the requirements in a data center by 80 to 90%."*

---

## License

MIT License

IronLattice is a trademark of its respective owners at external research institution. This is an independent educational project with no affiliation.

---

## Acknowledgments

**Dr. external research group** — For pioneering this technology and being a bold witness for Christ in the scientific community.

**Dr. Jaeho Shin** — For the engineering innovation that makes this possible.

---

*8 demos. Complete vision. World-class.*

*Built with Go, Vulkan, and curiosity.*
