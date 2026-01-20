# Demo 1: Ferroelectric Hysteresis Visualization

**Ferroelectric CIM Visualizer - Ferroelectric P-E Curve**

> *"It's got 30 discrete states. So it's not 0-1-0-1."* — Dr. external research group

**Complexity:** Beginner (Graphics only)
**Status:** Active Development

---

## Overview

Demo 1 provides an interactive visualization of ferroelectric hysteresis in HfO2-ZrO2 (HZO) superlattice materials. This demo illustrates the fundamental physics of ferroelectric memory cells that enable Ferroelectric CIM's compute-in-memory technology.

### What This Demo Shows

1. **P-E Hysteresis Loop** — The characteristic polarization-electric field curve of ferroelectric materials
2. **30 Discrete States** — How Ferroelectric CIM achieves multi-level cell (MLC) storage with ~4.9 bits/cell
3. **Preisach Hysteresis Model** — Physics-accurate simulation of domain switching
4. **Real-time Simulation** — Interactive control of electric field and waveforms
5. **Write/Read Operations** — Demonstrates non-volatile memory behavior

---

## Quick Start

```bash
# Navigate to demo directory
cd demo1-hysteresis

# Build and run (recommended)
go run ./cmd/demo

# Or build executable
go build -o hysteresis ./cmd/demo
./hysteresis
```

---

## Visualization

```
┌───────────────────────────────────────────────────────────────────────────────────────────┐
│  Ferroelectric CIM Ferroelectric Hysteresis Visualization                                       │
│  "It's got 30 discrete states. So it's not 0-1-0-1." — Dr. external research group                     │
├───────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                           │
│ ┌────────┐ ┌──────────────────────────┐ ┌───┐ ┌───────────────────┬───────────────────┐  │
│ │ Memory │ │   P-E Hysteresis Loop    │ │30 │ │ Controls          │ What You're       │  │
│ │  Cell  │ │                          │ │ L │ │                   │ Seeing            │  │
│ │ ┌────┐ │ │  P (µC/cm²)              │ │ E │ │ Material: [HZO v] │                   │  │
│ │ │ 24 │ │ │   40 ┼    ╭──────╮       │ │ V │ │ Waveform: [Demo v]│ WRITE/READ DEMO   │  │
│ │ └────┘ │ │  +Pr ┼────╯      │       │ │ E │ │ E-field: ███░░░░  │                   │  │
│ │        │ │   20 ┼           │       │ │ L │ │ Frequency: 0.5 Hz │ 1. WRITE: E>Ec    │  │
│ │ Level  │ │    0 ┼───────────┼─→ E   │ │ S │ │ Trail: 500 pts    │    sets state     │  │
│ │ 24/30  │ │  -20 ┼           │       │ │   │ │ [Pause] [Reset]   │ 2. HOLD: E=0      │  │
│ │        │ │  -Pr ┼────╮      │       │ │ ▓ │ ├───────────────────┤    P persists!    │  │
│ │Positive│ │  -40 ┼    ╰──────╯       │ │ ▓ │ │ Current State     │ 3. READ: E<Ec     │  │
│ │   P    │ │      -1  -Ec 0 +Ec  1    │ │ ▓ │ │ E: 0.85 MV/cm     │    no change      │  │
│ └────────┘ └──────────────────────────┘ │ ░ │ │ P: 25.3 µC/cm²    ├───────────────────┤  │
│                                         │ ░ │ │ Level: 24/30      │ Memory Log        │  │
│ This is the cell                        └───┘ │ Mode: [WRITE]     │                   │  │
│                                               │                   │ >> WRITE(28)      │  │
│                                               │                   │    HOLD @ 27      │  │
│                                               │                   │ << READ...        │  │
│                                               │                   │    Got: 27 [OK]   │  │
│                                               │                   │ >> WRITE(5)       │  │
│                                               └───────────────────┴───────────────────┘  │
│  ● Write/Read Demo | WRITING 5...                                                        │
└───────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Features

### Waveform Modes

| Mode | Description |
|------|-------------|
| **Manual** | Drag slider to control E-field directly |
| **Sine Wave** | Continuous sweep traces full hysteresis loop |
| **Triangle Wave** | Linear ramps show Ec switching thresholds |
| **Square Wave** | Instant jumps show rapid state flipping |
| **Random Walk** | Picks random target levels, demonstrates multi-level storage |
| **Write/Read Demo** | Full memory operation cycle: WRITE → HOLD → READ |

### GUI Controls

- **E-field Slider**: Drag to control electric field (Manual mode)
- **Waveform Dropdown**: Select input waveform type
- **Material Dropdown**: Switch between HZO variants
- **Frequency Slider**: Adjust speed (affects all auto modes)
- **Trail Slider**: Adjust plot history length
- **Pause/Resume Button**: Control simulation
- **Reset Button**: Clear history and restart

### Visual Indicators

- **Memory Cell**: Color-coded square showing current level (1-30)
- **P-E Plot**: Real-time hysteresis curve with Ec/Pr markers
- **Level Bar**: 30-segment vertical indicator
- **Mode Indicator**: Shows WRITE (|E|>Ec) or READ (|E|<Ec)
- **Educational Slide**: Context-sensitive explanations
- **Memory Log**: Real-time read/write operation log

---

## Physics Model

### Preisach Hysteresis Model (Mayergoyz Framework)

The demo implements the **Mayergoyz Preisach model** following classical hysteresis theory:

```
P(E) = ∫∫ μ(α, β) γ_αβ dα dβ  →  Discretized: P = Σ μᵢ × γᵢ
```

Where:
- `μ(α, β)` — Preisach distribution function (2D Gaussian centered at ±Ec)
- `γ_αβ` — Hysteron state (+1 or -1)
- `α` — Up-switching threshold (hysteron switches to +1 when E ≥ α)
- `β` — Down-switching threshold (hysteron switches to -1 when E ≤ β)

### How Hysteresis Emerges

Each hysteron is a bistable switch:
```go
if E >= Alpha { State = +1 }      // Switch UP
else if E <= Beta { State = -1 }  // Switch DOWN
// Between Beta and Alpha: State UNCHANGED (memory effect!)
```

**The loop is EMERGENT**, not drawn. The gap between α and β creates hysteresis.

### 30-Level Discretization

Continuous polarization mapped to discrete levels:
```go
discreteLevel = round((P/Ps + 1) / 2 * 29)  // 0 to 29
```

Linear spacing in polarization, not voltage thresholds.

### Key Parameters (HZO Materials)

| Parameter | Default HZO | Optimized | Ferroelectric CIM |
|-----------|-------------|-----------|-------------|
| Pr (µC/cm²) | 25 | 45 | 30 |
| Ps (µC/cm²) | 30 | 50 | 35 |
| Ec (MV/cm) | 1.2 | 0.8 | 1.0 |
| τ (ns) | 1.0 | 0.5 | 10* |
| Endurance | 10¹⁰ | 10¹² | 10¹¹ |

*τ is defined but NOT used in real-time visualization (quasistatic approximation).

### Write vs Read Operations

```
WRITE: |E| > Ec  → Polarization changes (crosses coercive field)
READ:  |E| < Ec  → Polarization unchanged, state sensed non-destructively
```

This is the fundamental principle of ferroelectric non-volatile memory.

---

## Architecture

```
demo1-hysteresis/
├── cmd/demo/
│   └── main.go              # Entry point
├── pkg/
│   ├── ferroelectric/       # Physics engine
│   │   ├── preisach.go      # Basic Preisach model
│   │   ├── preisach_advanced.go  # Full Mayergoyz model
│   │   ├── material.go      # HZO material parameters
│   │   └── render.go        # ASCII rendering utilities
│   └── gui/
│       └── gui.go           # Fyne GUI application
└── shaders/                 # (Reserved for future Vulkan mode)
```

## Dependencies

### GUI Framework

| Package | Purpose |
|---------|---------|
| [fyne-io/fyne/v2](https://github.com/fyne-io/fyne) | Cross-platform native GUI |

---

## The Story This Demo Tells

This demo answers the question: **"How does the memory cell work?"**

1. **Ferroelectric Effect** — Applying an electric field switches the polarization state
2. **Hysteresis** — The P-E curve shows memory effect (polarization depends on history)
3. **30 Discrete States** — By controlling voltage precisely, we can store 30 distinct levels
4. **Non-Volatile** — Polarization persists without power (shown by remanent polarization Pr)
5. **Fast Switching** — ~1 ns switching time enables high-speed operation

---

## Tests

```bash
# Run all tests
cd demo1-hysteresis
go test ./...

# Run ferroelectric package tests with verbose output
go test ./pkg/ferroelectric -v
```

---

## Troubleshooting

### GUI (Fyne) fails to start

**Linux:** Install required dependencies:
```bash
# Debian/Ubuntu
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel

# Arch
sudo pacman -S mesa libxcursor libxrandr libxinerama libxi
```

---

## References

1. Mayergoyz, I.D. "Mathematical Models of Hysteresis" IEEE Trans. Mag. (1986)
2. Böscke et al. "Ferroelectricity in HfO₂ Thin Films" APL (2011)
3. Park et al. "Ferroelectricity in Doped Hafnium Oxide" Adv. Mater. (2015)
4. Dr. external research group, "Ferroelectric CIM Presentation" (Nov 2024)
5. Bartic et al. "Preisach Model for Ferroelectric Capacitors" J. Appl. Phys. (2001)

---

## License

Part of the Ferroelectric CIM Visualizer project.
