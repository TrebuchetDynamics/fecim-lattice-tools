# Ferroelectric CIM Lattice Tools

**Educational visualization suite for Ferroelectric Compute-in-Memory (FeCIM) concepts (simulation-only)**

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://go.dev)
[![Fyne](https://img.shields.io/badge/Fyne-2.7.2-blue?logo=go)](https://fyne.io)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()
[![Modules](https://img.shields.io/badge/Modules-7-brightgreen.svg)]()
[![CI](https://github.com/your-org/fecim-lattice-tools/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/fecim-lattice-tools/actions/workflows/ci.yml)

> **Status**: Education phase (simulation-only). See `docs/project/STATUS.md`.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Modules](#modules)
- [Claims Policy](#claims-policy)
- [Technical Stack](#technical-stack)
- [Repository Structure](#repository-structure)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

FeCIM Lattice Tools is a Go/Fyne application that visualizes ferroelectric hysteresis, crossbar MVM behavior, quantization/noise effects, and a small MNIST demo. It is a **simulator** meant for learning and exploration. Values shown in the UI are **model parameters**, not device measurements.

**Simulation defaults:**
- Default conductance quantization: **30 levels** (configurable).
- Material presets and temperature controls are provided for exploration. See `module1-hysteresis/pkg/ferroelectric/material.go` for defaults.

---

## Quick Start

```bash
# Clone and run
git clone https://github.com/your-org/fecim-lattice-tools.git
cd fecim-lattice-tools
./launch.sh
```

Or build manually:

```bash
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools
```

---

## Installation

See `INSTALLATION.md` for prerequisites, optional dependencies, and platform-specific setup.

### Running Tests

```bash
go test ./...                              # See CI for latest status
go test -v ./module2-crossbar/pkg/crossbar # Crossbar tests only
go test -race ./...                        # Race detector (optional)
```

### Command Line Options

```bash
./launch.sh [options]
# Or: ./fecim-lattice-tools [options]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--logger` | off | Enable file logging to `logs/<timestamp>-fecim.log` |
| `--verbosity` | info | Log level: `off`, `info`, `debug`, `trace` |

**Examples:**
```bash
./launch.sh --logger --verbosity debug  # Enable logging with debug output
./launch.sh --verbosity trace           # Console-only trace output (no file)
```

---

## Modules

The visualizer includes 7 interconnected modules:

```
PHYSICS → COMPUTE → APPLICATION → SYSTEM → BUSINESS → TOOLING → REFERENCE

┌────────────┐    ┌────────────┐    ┌────────────┐    ┌────────────┐
│  Module 1  │───▶│  Module 2  │───▶│  Module 3  │    │  Module 7  │
│ Hysteresis │    │  Crossbar  │    │   MNIST    │    │    Docs    │
└────────────┘    └────────────┘    └────────────┘    └────────────┘
      │                                    │                 ▲
      ▼                                    ▼                 │
┌────────────┐    ┌────────────┐    ┌────────────┐          │
│  Module 4  │◀───│  Module 5  │◀───│  Module 6  │──────────┘
│  Circuits  │    │ Comparison │    │    EDA     │
└────────────┘    └────────────┘    └────────────┘
```

| Module | Focus | Description |
|--------|-------|-------------|
| **1. Hysteresis** | Physics | Preisach + Landau-Khalatnikov engines, multi-level state visualization |
| **2. Crossbar** | Compute | MVM operations with IR drop, sneak paths, drift, and noise |
| **3. MNIST** | Application | Dual-mode inference (FP32 vs CIM) with adjustable quantization/noise |
| **4. Circuits** | System | DAC/ADC/TIA behavior and peripheral timing/precision tradeoffs |
| **5. Comparison** | Business | Model-based, clearly labeled comparisons and projections |
| **6. EDA Suite** | Tooling | Generate illustrative JSON/CSV/SPICE/Verilog/DEF outputs (not signoff-ready) |
| **7. Docs** | Reference | In-app documentation browser and glossary |

---

## Claims Policy

This repository does **not** present hardware performance claims. External scientific claims (if any) are tracked in `docs/comparison/HONESTY_AUDIT.md`. If a claim is not listed there, treat it as **unverified**.

---

## Technical Stack

| Component | Technology |
|-----------|------------|
| Language | Go 1.24+ |
| GUI Framework | Fyne 2.7.2 |
| Physics Model | Preisach/Mayergoyz + Landau-Khalatnikov (educational) |
| Compute | Crossbar MVM simulation |
| Non-Idealities | IR drop, sneak paths, drift |
| Tests | See CI (`go test ./...`) |
| Documentation | Markdown + in-app viewer |

---

## Repository Structure

```
fecim-lattice-tools/
├── cmd/fecim-lattice-tools/    # Unified GUI entry point
├── module1-hysteresis/         # P-E curve physics
├── module2-crossbar/           # MVM + non-idealities
├── module3-mnist/              # Neural network demo
├── module4-circuits/           # Peripheral circuits
├── module5-comparison/         # Technology comparisons (model-based)
├── module6-eda/                # Design suite (educational)
├── module7-docs/               # Documentation browser
├── shared/                     # Common theme, logging, widgets
├── docs/                       # Markdown documentation files
├── data/                       # Calibration data
└── go.mod
```

---

## Contributing

Contributions are welcome. Please:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/your-feature`)
3. Follow existing code patterns (see `CLAUDE.md` for conventions)
4. Run tests (`go test ./...`)
5. Submit a pull request

---

## License

MIT License

This is an independent educational visualization project. Ferroelectric CIM research originates from the broader research community. No official affiliation.
