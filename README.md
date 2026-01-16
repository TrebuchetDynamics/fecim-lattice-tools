# IronLattice Visualizer

**GPU-Accelerated Ferroelectric Compute-in-Memory Visualization**

---

## About

This repository contains my exploration and visualization of **ferroelectric compute-in-memory (CIM)** technology, inspired by the work of **Dr. external research group** and **Dr. Jaeho Shin** at external research institution.

IronLattice is a new hardware architecture that performs computation directly in memory using a proprietary ferroelectric superlattice. This eliminates the "Von Neumann bottleneck" — the constant busing of data between logic and memory that consumes most of the energy in traditional computing.

> *"This could lower the requirements in a data center by 80 to 90% of the energy requirements."*
> — Dr. external research group

---

## The Technology

### What It Is

- **Compute-in-Memory (CIM):** The same device does memory AND computation
- **Ferroelectric Superlattice:** Proprietary HfO₂-ZrO₂ layer structure
- **CMOS Compatible:** Works on standard fabrication lines, no exotic materials
- **Capital Light:** No special tools required

### Performance Claims

| Metric | vs NAND Flash | vs DRAM |
|--------|---------------|---------|
| Read/Write Energy | **10,000,000× lower** | **1,000× lower** |
| Speed | **1,000,000× faster** | — |
| Voltage | **90% reduction** | — |
| Refresh Cycles | — | **Zero** (non-volatile) |

### Current Status

| Metric | Value |
|--------|-------|
| TRL (Technology Readiness Level) | **4** (lab validation) |
| Discrete States | **30** (not just 0/1) |
| MNIST Accuracy | **87%** (theoretical max: 88%) |
| Endurance Target | **10¹² cycles** |

---

## Go-to-Market Strategy

IronLattice plans a staged market entry:

```
Phase 1: Replace NAND Flash    →  Drop-in replacement, no software changes
Phase 2: Replace DRAM          →  Non-volatile, lower energy
Phase 3: Full Compute-in-Memory →  Neural network inference on-chip
```

This applies to both data centers and consumer devices (smartphones).

---

## Repository Structure

```
ironlattice-vis/
│
├── docs/
│   ├── PHYSICS.md              # Ferroelectric physics deep dive
│   ├── CURRICULUM.md           # Learning path
│   └── DEMO-SPECS.md           # Demo specifications
│
├── demo1-hysteresis/           # Single cell P-E curve visualizer
│   ├── cmd/
│   ├── pkg/
│   │   ├── ferroelectric/      # Physics models (Preisach, Landau)
│   │   ├── simulation/         # Simulation engine
│   │   └── vulkan/             # GPU rendering
│   └── shaders/
│
├── demo2-crossbar/             # Crossbar array MVM (planned)
│
├── demo3-mnist/                # Neural network on CIM (planned)
│
├── shared/                     # Common utilities
├── assets/                     # Images, fonts
└── README.md
```

---

## Demos

### Demo 1: Ferroelectric Hysteresis Visualizer

Interactive visualization of a single ferroelectric memory cell showing:

- **P-E Hysteresis Curve** — The characteristic ferroelectric response
- **Polarization Switching** — Watch domains flip in real-time
- **30 Discrete States** — Not just 0/1, but analog levels

```
┌────────────────┐      ┌──────────────────────┐
│                │      │         P            │
│     CELL       │      │         ↑    +Pr     │
│  (Color = P)   │      │         ┌────╮       │
│                │      │    ─────┼────┼──→ E  │
│                │      │         ╰────┘       │
│                │      │              -Pr     │
└────────────────┘      └──────────────────────┘
```

### Demo 2: Crossbar Array (Planned)

Visualize Matrix-Vector Multiplication in memory:

```
V₁ ──→ [G₁₁][G₁₂][G₁₃] ──→ I₁ = Σ(Vⱼ × Gⱼ₁)
V₂ ──→ [G₂₁][G₂₂][G₂₃] ──→ I₂ = Σ(Vⱼ × Gⱼ₂)
V₃ ──→ [G₃₁][G₃₂][G₃₃] ──→ I₃ = Σ(Vⱼ × Gⱼ₃)

Ohm's Law: I = V × G
Kirchhoff's Law: I_total = ΣI
```

### Demo 3: MNIST on CIM (Planned)

Handwritten digit recognition running on simulated IronLattice hardware, matching their 87% accuracy benchmark.

---

## Tech Stack

| Component | Technology |
|-----------|------------|
| Language | Go |
| Graphics | Vulkan |
| Shaders | GLSL (SPIR-V) |
| Physics | Preisach model, Landau-Khalatnikov |

---

## Getting Started

```bash
# Clone
git clone https://github.com/yourusername/ironlattice-vis.git
cd ironlattice-vis

# Dependencies (Ubuntu)
sudo apt install vulkan-tools libvulkan-dev glslc

# Build
go mod download
cd demo1-hysteresis/shaders && ./compile.sh && cd ../..
go build -o bin/hysteresis ./demo1-hysteresis/cmd/hysteresis

# Run
./bin/hysteresis
```

---

## The Team Behind IronLattice

From Dr. Tour's November 2024 talk:

- **Dr. Jaeho Shin** — Device engineer, inventor of the superlattice architecture
- **Tawfik Jarjour** — Former Accenture (13 years), leading commercialization
- **Advisors** — IBM patent veteran, semiconductor foundry experts

> *"We haven't raised a penny to date. We've taken no money because we really want to move with the best strategy."*

They are currently in restricted access discussions with major companies and talking to foundries about scaling down from university lab to commercial production.

---

## Context: George Gilder's Prediction

In response to Gilder's Wall Street Journal article (*"The Microchip Era is About to End"*, Nov 3, 2024), Dr. Tour argues that IronLattice could actually **enable** the next era of wafer-scale integrated circuits by solving:

1. Memory bottleneck
2. Energy constraints
3. CMOS compatibility

---

## Why I Built This

I'm a computational physicist interested in the hardware that will power the next generation of AI. When I learned about IronLattice, I wanted to:

1. **Understand** the physics deeply
2. **Visualize** it so others can understand
3. **Learn** GPU programming through a meaningful project

This is an educational project. I have no affiliation with external research institution or IronLattice.

---

## Resources

### Primary Source
- [Dr. Tour's IronLattice Talk (November 2024)](https://www.youtube.com/watch?v=...) — First public presentation

### Dr. Tour's Ministry
- [Jesus and Science Foundation](https://jesusandscience.org)

### Technical Background
- Böscke, T.S., et al. "Ferroelectricity in hafnium oxide thin films." APL (2011)
- Park, M.H., et al. "Ferroelectricity in Doped HfO₂." Advanced Materials (2015)

---

## License

MIT License

IronLattice is a trademark of its respective owners at external research institution. This is an independent educational project.

---

## Acknowledgments

**Dr. external research group** — For pioneering this technology and being a bold witness for Christ in the scientific community.

**Dr. Jaeho Shin** — For the engineering innovation.

> *"If you do not believe in the physical resurrection of Jesus Christ, send me an email... and we will get together and I will share with you about why I embrace the resurrection of Jesus."*
> — Dr. external research group

---

*Built with Go, Vulkan, and curiosity.*
