# Demo 1: Ferroelectric Hysteresis Visualization

**IronLattice Visualizer - Ferroelectric P-E Curve**

> *"It's got 30 discrete states. So it's not 0-1-0-1."* — Dr. external research group

## Overview

Demo 1 provides an interactive visualization of ferroelectric hysteresis in HfO2-ZrO2 (HZO) superlattice materials. This demo illustrates the fundamental physics of ferroelectric memory cells that enable IronLattice's compute-in-memory technology.

### What This Demo Shows

1. **P-E Hysteresis Loop** — The characteristic polarization-electric field curve of ferroelectric materials
2. **30 Discrete States** — How IronLattice achieves multi-level cell (MLC) storage with ~4.9 bits/cell
3. **Preisach Hysteresis Model** — Physics-accurate simulation of domain switching
4. **Real-time Simulation** — Interactive control of electric field and waveforms

## Quick Start

```bash
# Navigate to demo directory
cd demo1-hysteresis

# Build the demo
go build ./cmd/hysteresis

# Run Fyne GUI mode (default - recommended)
./hysteresis

# Run terminal UI mode (for SSH/remote)
./hysteresis --tui

# Run headless mode (static ASCII output)
./hysteresis --headless

# Run with Vulkan graphics (advanced)
./hysteresis --vulkan
```

## Run Modes

### 1. Fyne GUI Mode (Default - Recommended)

A cross-platform native GUI application featuring:
- **Smooth P-E hysteresis curve** with real-time animation
- **30-level state indicator** with color gradients
- **Interactive sliders** for E-field control
- **Material selector** dropdown
- **Waveform selector** (Sine, Triangle, Square, Manual)
- **Live parameter display** with professional styling
- **Dark theme** optimized for IronLattice branding

**GUI Controls:**
- **E-field Slider**: Drag to control electric field (Manual mode)
- **Waveform Dropdown**: Select input waveform type
- **Material Dropdown**: Switch between HZO variants
- **Frequency Slider**: Adjust waveform frequency (0.1-5 Hz)
- **Pause/Resume Button**: Control simulation
- **Reset Button**: Clear history and restart

### 2. Terminal UI Mode (TUI)

Full-screen terminal interface for SSH/remote access:
```bash
./hysteresis --tui
```

**TUI Controls:**
| Key | Action |
|-----|--------|
| `↑/k` | Increase E-field |
| `↓/j` | Decrease E-field |
| `←/h` | Fine decrease |
| `→/l` | Fine increase |
| `Space` | Pause/Resume |
| `Tab` | Cycle waveform |
| `a` | Toggle auto mode |
| `m` | Cycle materials |
| `r` | Reset simulation |
| `?` | Show help |
| `q/Esc` | Quit |

### 3. Headless Mode

Static ASCII output for terminals without interactivity:
```bash
./hysteresis --headless
```

### 4. Vulkan Graphics Mode

GPU-accelerated visualization (advanced):
```bash
./hysteresis --vulkan
```

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

### How Hysteresis Emerges (Verified in Code)

Each hysteron is a bistable switch:
```go
if E >= Alpha { State = +1 }      // Switch UP
else if E <= Beta { State = -1 }  // Switch DOWN
// Between Beta and Alpha: State UNCHANGED (memory effect!)
```

**The loop is EMERGENT**, not drawn. The gap between α and β creates hysteresis.

### 30-Level Discretization (Verified in Code)

Continuous polarization mapped to discrete levels:
```go
discreteLevel = round((P/Ps + 1) / 2 * 29)  // 0 to 29
```

Linear spacing in polarization, not voltage thresholds.

### Key Parameters (HZO Materials)

| Parameter | Default HZO | Optimized | IronLattice |
|-----------|-------------|-----------|-------------|
| Pr (µC/cm²) | 25 | 45 | 30 |
| Ps (µC/cm²) | 30 | 50 | 35 |
| Ec (MV/cm) | 1.2 | 0.8 | 1.0 |
| τ (ns) | 1.0 | 0.5 | 10* |
| Endurance | 10¹⁰ | 10¹² | 10¹¹ |

*τ is defined but NOT used in real-time visualization (quasistatic approximation).

### Temperature Dependence

The model includes temperature-dependent coercive field:
```
Ec(T) = Ec₀ × (1 - T/Tc)^0.5
```

Where Tc ≈ 723K (450°C) for HZO.

### What's Simulated vs. Displayed

| Feature | Status | Notes |
|---------|--------|-------|
| Preisach hysterons | ✅ Active | ~450 hysterons on 30×30 grid |
| Gaussian μ(α,β) | ✅ Active | σ = 20% of Ec |
| Hysteresis loop | ✅ Emergent | Not forced/drawn |
| 30 discrete levels | ✅ Active | Linear in P |
| τ switching dynamics | ⚠️ Defined | Not used in real-time loop |
| KAI domain growth | ⚠️ Defined | Available via `SimulateDomainSwitching()` |
| Fatigue/wake-up | ✅ Active | Very low rate (1e-10) |

---

## Proposed Improvements (From Literature Analysis)

### 1. Stack-Based Voltage History Tracking (Priority: HIGH)

**Reference:** Mayergoyz, I.D. "Mathematical Models of Hysteresis" IEEE Trans. Mag. (1986)

**Current Gap:** Basic Preisach model may not fully track "wiping-out" property.

**Improvement:** Implement stack-based algorithm to track voltage reversal points:
- Record voltage extrema {u₁, u₂, ..., uₙ}
- Dynamically update Preisach integral boundaries
- Track S⁺(t) and S⁻(t) regions on Preisach triangle
- Implement geometric interface updates

```go
// Proposed enhancement
type EnhancedPreisachModel struct {
    grid [100][100]float64  // μ(α,β) distribution
    state [100][100]int     // ±1 for each hysteron
    voltageStack []float64   // Local extrema history (wiping-out)
}
```

### 2. Minor Loop Visualization (Priority: MEDIUM)

**Reference:** Physical Reality Preisach Model (Nature 2018)

**Improvement:** Visualize minor loops when user reverses direction before saturation:
- Show nested loops forming on P-E plot
- Color-code minor vs major loops
- Display Preisach plane state during partial cycles

### 3. Preisach Neural Network Self-Calibration (Priority: LOW)

**Reference:** B-Spline Everett Map Preisach (arXiv:2410.02797)

**Improvement:** Implement Preisach-NN architecture:
- Layer 1: Stop operator neurons (one per hysteron)
- Layer 2: Linear summation with learned weights
- Benefit: Self-calibrating model learns μ(α,β) from device measurements

### 4. Domain Wall Dynamics Visualization (Priority: LOW)

**Reference:** TDGL_Ferroelectric_Domains_arXiv.pdf (FerroX framework)

**Improvement:** Add domain nucleation/propagation animation:
- Show domains switching during field application
- Visualize grain-by-grain switching in polycrystalline HZO

---

## Papers Supporting This Demo

### Currently Available
| Paper | Location | Relevance |
|-------|----------|-----------|
| Preisach_Ferroelectric_Modeling_arXiv.pdf | opensource/papers/01_Core_Materials/ | Hysteresis loop modeling |
| HZO_Ferroelectric_Discovery_arXiv.pdf | opensource/papers/01_Core_Materials/ | HZO polarization switching |
| TDGL_Ferroelectric_Domains_arXiv.pdf | opensource/papers/01_Core_Materials/ | FerroX TDGL framework |
| newton_secant_preisach_control_2024.pdf | papers/downloaded/arxiv/ | B-Spline Everett Preisach |
| physical_reality_preisach_2018.pdf | papers/downloaded/nature/ | Domain physics |

### Recommended for Download
| Paper | Source | Why Needed |
|-------|--------|------------|
| **Mayergoyz IEEE 1986** | IEEE Xplore | Original Preisach mathematics (CORRUPTED - needs re-download) |
| **Böscke et al. APL 2011** | AIP Publishing | HfO₂ ferroelectric discovery - mechanical confinement mechanism |
| **Domain wall dynamics in HZO** | IEEE EDL | For domain-level animation physics |

---

## Architecture

```
demo1-hysteresis/
├── cmd/hysteresis/
│   └── main.go              # Entry point with mode selection
├── pkg/
│   ├── ferroelectric/       # Physics engine
│   │   ├── preisach.go      # Basic Preisach model
│   │   ├── preisach_advanced.go  # Full Mayergoyz model
│   │   ├── material.go      # HZO material parameters
│   │   └── render.go        # ASCII rendering utilities
│   ├── gui/
│   │   └── gui.go           # Fyne GUI (default, recommended)
│   ├── simulation/
│   │   └── engine.go        # Time-stepping simulation
│   ├── render/
│   │   ├── vulkan.go        # Vulkan graphics backend
│   │   └── plot.go          # Plot data structures
│   └── tui/
│       └── tui.go           # Terminal UI (for SSH/remote)
└── shaders/                 # SPIR-V shaders for Vulkan
```

## Go Package Dependencies

### GUI Framework (Primary)

| Package | Purpose | Why Chosen |
|---------|---------|------------|
| [fyne-io/fyne/v2](https://github.com/fyne-io/fyne) | Cross-platform GUI | Native look, canvas drawing, widgets, dark theme support |

### TUI Libraries (Terminal Fallback)

| Package | Purpose | Why Chosen |
|---------|---------|------------|
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework | Elm-inspired MVU architecture, SSH-friendly |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Terminal styling | CSS-like styling API |
| [charmbracelet/bubbles](https://github.com/charmbracelet/bubbles) | TUI components | Help system, keyboard bindings |

### Graphics Libraries (Vulkan Backend)

| Package | Purpose | Why Chosen |
|---------|---------|------------|
| [vulkan-go/vulkan](https://github.com/vulkan-go/vulkan) | GPU rendering | Modern Vulkan API for shader-based graphics |
| [go-gl/glfw](https://github.com/go-gl/glfw) | Windowing | Cross-platform window creation |

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

# Run ferroelectric package tests
go test ./pkg/ferroelectric -v

# Run simulation tests
go test ./pkg/simulation -v
```

Test coverage:
- Hysteresis loop generation
- Coercive field switching
- 30 discrete states verification
- Material parameter validation
- Preisach model reset
- Normalized polarization bounds

---

## Troubleshooting

### GUI (Fyne) fails to start

If the Fyne GUI doesn't start:

**Linux:** Install required dependencies:
```bash
# Debian/Ubuntu
sudo apt-get install libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel

# Arch
sudo pacman -S mesa libxcursor libxrandr libxinerama libxi
```

**Fallback to TUI:**
```bash
./hysteresis --tui
```

### TUI mode fails

If the terminal UI doesn't work:
```bash
# Try with different TERM setting
TERM=xterm-256color ./hysteresis --tui

# Or use headless mode
./hysteresis --headless
```

### Vulkan mode fails

Ensure you have:
1. Vulkan-capable GPU drivers installed
2. GLFW with Vulkan support
3. Compiled SPIR-V shaders in `shaders/` directory

Recompile shaders:
```bash
cd shaders && ./compile.sh
```

---

## References

1. Mayergoyz, I.D. "Mathematical Models of Hysteresis" IEEE Trans. Mag. (1986) - **CRITICAL**
2. Böscke et al. "Ferroelectricity in HfO₂ Thin Films" APL (2011) - **FOUNDATIONAL**
3. Park et al. "Ferroelectricity in Doped Hafnium Oxide" Adv. Mater. (2015)
4. Dr. external research group, "IronLattice Presentation" (Nov 2024)
5. Bartic et al. "Preisach Model for Ferroelectric Capacitors" J. Appl. Phys. (2001)

---

## License

Part of the IronLattice Visualizer project.
