# IronLattice Visualizer

**GPU-Accelerated Ferroelectric Compute-in-Memory Visualization**

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev)
[![Vulkan](https://img.shields.io/badge/Vulkan-1.3-AC162C?logo=vulkan)](https://www.vulkan.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## Overview

This repository contains GPU-accelerated visualizations of **ferroelectric compute-in-memory (CIM)** technology, inspired by the groundbreaking work of **Dr. external research group** and **Dr. Jaeho Shin** at external research institution.

IronLattice represents a paradigm shift in computing: performing computation directly in memory using ferroelectric superlattices, eliminating the Von Neumann bottleneck that wastes 90%+ of energy in traditional AI hardware.

> *"This could lower the requirements in a data center by 80 to 90% of the energy requirements."*
> — Dr. external research group

---

## The Technology

### Core Innovation

| Aspect | Description |
|--------|-------------|
| **Compute-in-Memory** | Same device performs memory AND computation |
| **Ferroelectric Superlattice** | Atomically precise HfO₂/ZrO₂ layered structure |
| **CMOS Compatible** | Works on standard fabrication lines |
| **Analog Computing** | 30+ discrete states, not just 0/1 |

### Performance vs. Existing Technologies

| Metric | vs NAND Flash | vs DRAM |
|--------|---------------|---------|
| Read/Write Energy | **10,000,000× lower** | **1,000× lower** |
| Speed | **1,000,000× faster** | Comparable |
| Voltage | **90% reduction** | Lower |
| Data Retention | Non-volatile | **Zero refresh** |

### Current Status (TRL 4)

| Metric | Value |
|--------|-------|
| Technology Readiness Level | **4** (lab validation) |
| Discrete Analog States | **30** levels |
| MNIST Accuracy | **87%** (near theoretical max) |
| Endurance Target | **10¹² cycles** |

---

## Project Goals

This visualization project aims to:

1. **Simulate** ferroelectric physics (Landau-Khalatnikov, Preisach models)
2. **Visualize** domain switching and hysteresis in real-time
3. **Demonstrate** crossbar array matrix-vector multiplication
4. **Educate** on compute-in-memory principles

---

## Implementation Status

| Demo | Physics | Graphics | Overall |
|------|---------|----------|---------|
| **Demo 1: Hysteresis** | Complete | In Progress | Headless working |
| **Demo 2: Crossbar MVM** | Partial | Not Started | Infrastructure only |
| **Demo 3: Phase-Field** | Designed | Not Started | Specification only |

**What works today:**
```bash
go run demo1-hysteresis/cmd/hysteresis/main.go --headless
```

---

## Repository Structure

```
ironlattice-vis/
├── docs/                        # Comprehensive documentation (3.7 MB)
│   ├── CURRICULUM.md            # 8-area doctoral curriculum
│   ├── CURRICULUM_DETAILED.md   # Expanded learning path
│   ├── IRONLATTICE_PARADIGM.md  # Technology deep-dive
│   ├── PROJECT_ROADMAP.md       # Implementation timeline
│   ├── VULKAN_DEMO_GUIDE.md     # Graphics implementation guide
│   ├── HZO_PARAMETERS.md        # Material constants
│   ├── RESEARCH_LOG.md          # Research journal
│   └── RESEARCH_FINDINGS_*.md   # Weekly research summaries
│
├── papers/                      # Scientific papers collection
│   ├── downloaded/              # 19 PDFs (arXiv, Nature, IEEE, etc.)
│   ├── DOWNLOAD_PLAN.md         # Paper acquisition roadmap
│   ├── paper_metadata.json      # Paper index
│   └── paper_downloader.py      # Automated fetcher
│
├── demo1-hysteresis/            # Single cell P-E curve visualizer
│   ├── cmd/hysteresis/          # Application entry point
│   ├── pkg/
│   │   ├── ferroelectric/       # Preisach model, material params
│   │   ├── simulation/          # Time-stepping engine
│   │   └── render/              # Graphics pipeline (WIP)
│   ├── shaders/                 # GLSL compute/graphics shaders
│   ├── PHYSICS.md               # Physics documentation
│   └── README.md                # Demo-specific docs
│
├── demo2-crossbar/              # Crossbar array MVM visualizer
│   ├── cmd/inference/           # Application entry point
│   ├── pkg/
│   │   ├── crossbar/            # Array modeling
│   │   ├── network/             # Neural network layers
│   │   └── data/                # MNIST loading
│   ├── shaders/                 # MVM compute shaders
│   ├── PHYSICS.md               # Physics documentation
│   └── README.md                # Demo-specific docs
│
├── demo3-phasefield/            # GPU phase-field domain simulator
│   ├── PHYSICS.md               # TDGL equations documentation
│   └── README.md                # Specifications
│
└── go.mod                       # Go module definition
```

---

## Demos

### Demo 1: Ferroelectric Hysteresis Visualizer

**Status:** Physics complete, graphics in progress

Interactive visualization of a single ferroelectric memory cell:

```
┌────────────────┐      ┌──────────────────────┐
│                │      │         P            │
│     CELL       │      │         ↑    +Pᵣ     │
│  (Color = P)   │      │         ┌────╮       │
│                │      │    ─────┼────┼──→ E  │
│                │      │         ╰────┘       │
│                │      │              -Pᵣ     │
└────────────────┘      └──────────────────────┘
```

**Implemented:**
- Preisach hysteresis model with history tracking
- HZO material parameters from literature
- Time-stepping simulation engine
- 30 discrete analog state generation
- Multiple waveforms (sine, triangle, square)
- Headless mode for data output

**In Progress:**
- Vulkan graphics pipeline
- Real-time visualization
- Interactive voltage control

**Run headless mode:**
```bash
go run demo1-hysteresis/cmd/hysteresis/main.go --headless
```

### Demo 2: Crossbar Array MVM

**Status:** Infrastructure complete, computation in progress

Visualize Matrix-Vector Multiplication in memory:

```
V₁ ──→ [G₁₁][G₁₂][G₁₃] ──→ I₁ = Σ(Vⱼ × Gⱼ₁)
V₂ ──→ [G₂₁][G₂₂][G₂₃] ──→ I₂ = Σ(Vⱼ × Gⱼ₂)
V₃ ──→ [G₃₁][G₃₂][G₃₃] ──→ I₃ = Σ(Vⱼ × Gⱼ₃)

Ohm's Law:      I = V × G  (multiplication)
Kirchhoff's Law: Iₜₒₜₐₗ = ΣI (summation)
```

**Implemented:**
- Crossbar array data structures
- Cell conductance modeling with noise
- Weight programming interface
- ADC/DAC quantization support
- Network layer scaffolding

**In Progress:**
- MVM compute shader execution
- MNIST inference pipeline
- Non-ideality modeling (IR drop, sneak paths)

### Demo 3: GPU Phase-Field Domain Simulator

**Status:** Design complete, implementation not started

GPU-accelerated Time-Dependent Ginzburg-Landau (TDGL) simulation for ferroelectric domain dynamics. Will visualize domain nucleation, growth, and switching at the nanoscale.

---

## Tech Stack

| Component | Technology | Purpose | Status |
|-----------|------------|---------|--------|
| Language | Go 1.21+ | Performance + simplicity | Ready |
| Graphics API | Vulkan 1.3 | Cross-platform GPU access | Planned |
| Shaders | GLSL → SPIR-V | Compute + rendering | Defined |
| Physics | Preisach model | Ferroelectric hysteresis | Implemented |
| Simulation | TDGL | Domain dynamics | Planned |

### Planned Dependencies

```go
// Currently in go.mod as comments, to be added:
github.com/bbredesen/go-vk  // Vulkan bindings
github.com/go-gl/glfw       // Window management
gonum.org/v1/gonum          // Math operations
```

---

## Getting Started

### Prerequisites

- Go 1.21+
- Vulkan SDK 1.3+ (for graphics demos)
- GLSL compiler `glslc` (for shader compilation)

### Quick Start (Headless Physics)

The physics simulation runs without any external dependencies:

```bash
# Clone repository
git clone https://github.com/yourusername/ironlattice-vis.git
cd ironlattice-vis

# Run demo 1 in headless mode (no graphics required)
go run demo1-hysteresis/cmd/hysteresis/main.go --headless
```

### Full Installation (Graphics)

```bash
# Install system dependencies (Ubuntu/Debian)
sudo apt install vulkan-tools libvulkan-dev glslc

# Install Go dependencies (when implemented)
go mod tidy

# Compile shaders
cd demo1-hysteresis/shaders && ./compile.sh && cd ../..

# Build
go build -o bin/hysteresis ./demo1-hysteresis/cmd/hysteresis

# Run
./bin/hysteresis
```

> **Note:** Graphics mode is currently in development. Use `--headless` flag for working physics output.

---

## Learning Resources

### Documentation

| Document | Description |
|----------|-------------|
| [CURRICULUM.md](docs/CURRICULUM.md) | 8-area doctoral-level curriculum |
| [CURRICULUM_DETAILED.md](docs/CURRICULUM_DETAILED.md) | Expanded learning path |
| [IRONLATTICE_PARADIGM.md](docs/IRONLATTICE_PARADIGM.md) | Technology paradigm analysis |
| [PROJECT_ROADMAP.md](docs/PROJECT_ROADMAP.md) | Implementation timeline |
| [VULKAN_DEMO_GUIDE.md](docs/VULKAN_DEMO_GUIDE.md) | Graphics implementation guide |
| [HZO_PARAMETERS.md](docs/HZO_PARAMETERS.md) | Material constants reference |
| [papers/](papers/) | 19 scientific papers (arXiv, Nature, IEEE) |

### Key Concepts Covered

1. **Solid-State Physics** — HfO₂ crystallography, phase stabilization
2. **Ferroelectric Devices** — FeFET, FeRAM, domain dynamics
3. **Compute-in-Memory** — Crossbar arrays, Kirchhoff's laws
4. **Neural Networks** — Weight mapping, noise-aware training
5. **Simulation** — TDGL, Preisach, phase-field models
6. **GPU Programming** — Vulkan compute shaders
7. **Scientific Visualization** — Real-time domain rendering
8. **Commercialization** — Manufacturing, IP strategy

---

## The Team Behind IronLattice

| Person | Role |
|--------|------|
| **Dr. external research group** | Principal Investigator, external research institution |
| **Dr. Jaeho Shin** | Device Engineer, Superlattice Inventor |
| **Tawfik Jarjour** | Commercialization Lead |

> *"We haven't raised a penny to date. We've taken no money because we really want to move with the best strategy."*

---

## Market Context

### Go-to-Market Strategy

```
Phase 1: Replace NAND Flash    →  Drop-in replacement
Phase 2: Replace DRAM          →  Non-volatile, lower energy
Phase 3: Full Compute-in-Memory →  Neural network inference on-chip
```

### George Gilder's Prediction

In response to *"The Microchip Era is About to End"* (WSJ, Nov 2024), IronLattice addresses:

1. Memory bottleneck → **Eliminated**
2. Energy constraints → **90% reduction**
3. CMOS compatibility → **Native integration**

---

## External Resources

### Primary Sources
- Dr. Tour's IronLattice Talk (Nov 2024) — Search "external research group IronLattice" on YouTube
- [external research institution News](https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants)

### Technical Papers
- Böscke, T.S., et al. "Ferroelectricity in hafnium oxide thin films." APL (2011)
- Park, M.H., et al. "Ferroelectricity in Doped HfO₂." Advanced Materials (2015)
- Shin, J., et al. "BEOL-Compatible Superlattice FEFET Analog Synapse" IEEE (2022)

### Dr. Tour's Ministry
- [Jesus and Science Foundation](https://jesusandscience.org)

---

## Contributing

Contributions welcome! Current priorities:

- [x] Preisach model implementation
- [ ] Vulkan graphics pipeline for demo 1
- [ ] Landau-Khalatnikov solver
- [ ] MVM compute shader execution (demo 2)
- [ ] Phase-field simulation (demo 3)
- [ ] MNIST inference on crossbar array

---

## License

MIT License

IronLattice is a trademark of its respective owners at external research institution. This is an independent educational project with no affiliation.

---

## Acknowledgments

**Dr. external research group** — For pioneering this technology and being a bold witness for Christ in the scientific community.

**Dr. Jaeho Shin** — For the engineering innovation that makes this possible.

> *"If you do not believe in the physical resurrection of Jesus Christ, send me an email... and we will get together and I will share with you about why I embrace the resurrection of Jesus."*
> — Dr. external research group

---

*Built with Go, Vulkan, and curiosity.*
