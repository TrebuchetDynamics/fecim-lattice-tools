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

Displays:
- P-E hysteresis loop as ASCII art
- Preisach plane domain states
- 30 discrete states table
- Switching dynamics
- Temperature dependence
- Material comparison

### 4. Vulkan Graphics Mode

GPU-accelerated visualization (advanced):
```bash
./hysteresis --vulkan
```

Features shader-based rendering with 60 FPS animation.

## Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| (none) | GUI | Run Fyne GUI mode (recommended) |
| `--tui` | `false` | Run terminal UI mode |
| `--headless` | `false` | Run headless ASCII mode |
| `--vulkan` | `false` | Run Vulkan graphics mode |
| `--optimized` | `false` | Use optimized superlattice parameters |
| `--freq` | `1e6` | Waveform frequency in Hz |

## Physics Model

### Preisach Hysteresis Model

The demo implements the **Mayergoyz Preisach model** following classical hysteresis theory:

```
P(E) = ∫∫ μ(α, β) γ_αβ dα dβ
```

Where:
- `μ(α, β)` — Preisach distribution function (Gaussian)
- `γ_αβ` — Hysteron state (+1 or -1)
- `α` — Up-switching field
- `β` — Down-switching field

### Key Parameters (HZO Materials)

| Parameter | Default HZO | Optimized | IronLattice |
|-----------|-------------|-----------|-------------|
| Pr (µC/cm²) | 25 | 45 | 30 |
| Ps (µC/cm²) | 30 | 50 | 35 |
| Ec (MV/cm) | 1.2 | 0.8 | 1.0 |
| τ (ns) | 1.0 | 0.5 | 1.0 |
| Endurance | 10¹⁰ | 10¹² | 10¹¹ |

### Temperature Dependence

The model includes temperature-dependent coercive field:
```
Ec(T) = Ec₀ × (1 - T/Tc)^0.5
```

Where Tc ≈ 723K (450°C) for HZO.

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

Fyne provides:
- **Canvas API** for custom P-E curve rendering
- **Widget system** for sliders, dropdowns, buttons
- **Theme support** with IronLattice custom dark theme
- **Cross-platform** builds for Windows, macOS, Linux

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

### Recommended Additional Packages

For future enhancements:

| Package | Purpose |
|---------|---------|
| [gonum/gonum](https://github.com/gonum/gonum) | Numerical computing (matrix ops, FFT) |
| [guptarohit/asciigraph](https://github.com/guptarohit/asciigraph) | ASCII line graphs |

## The Story This Demo Tells

This demo answers the question: **"How does the memory cell work?"**

1. **Ferroelectric Effect** — Applying an electric field switches the polarization state
2. **Hysteresis** — The P-E curve shows memory effect (polarization depends on history)
3. **30 Discrete States** — By controlling voltage precisely, we can store 30 distinct levels
4. **Non-Volatile** — Polarization persists without power (shown by remanent polarization Pr)
5. **Fast Switching** — ~1 ns switching time enables high-speed operation

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

### Building on Windows

Fyne requires CGO on Windows. Install:
1. [MSYS2](https://www.msys2.org/)
2. Run: `pacman -S mingw-w64-x86_64-gcc`
3. Add to PATH: `C:\msys64\mingw64\bin`

### Performance issues

For smoother animation, try:
```bash
# Reduce simulation frequency
./hysteresis --freq 1e5
```

## References

1. Mayergoyz, I.D. "Mathematical Models of Hysteresis" (1991)
2. Park et al. "Ferroelectricity in Doped Hafnium Oxide" Adv. Mater. (2015)
3. Dr. external research group, "IronLattice Presentation" (Nov 2024)
4. Bartic et al. "Preisach Model for Ferroelectric Capacitors" J. Appl. Phys. (2001)

## License

Part of the IronLattice Visualizer project.
