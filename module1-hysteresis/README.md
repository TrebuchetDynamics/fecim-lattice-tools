# Module 1: Hysteresis

Ferroelectric hysteresis modeling for FeCIM materials. Implements Preisach and Landau-Khalatnikov (L-K) engines with ISPP/WRD target progression, calibration workflows, and interactive visualization.

## Overview

Module 1 is the physics core of fecim-lattice-tools. It models polarization–electric field (P-E) hysteresis loops for ferroelectric materials (HZO, AlScN, etc.), supporting both the classical Preisach model and the Landau-Khalatnikov time-domain approach. The module drives ISPP (Incremental Step Pulse Programming) and WRD (Write-Read-Disturb) workflows used by downstream modules for multi-level cell programming.

## Package Structure

### `pkg/ferroelectric/` — Core Physics Models

- **preisach.go** — Preisach model: hysteron distribution, P-E loop generation
- **material.go** — Material parameter definitions (P_r, E_c, thickness, etc.)
- **level_bins.go** — Discrete conductance level binning for multi-level operation
- **render.go** — Data rendering helpers for loop visualization

### `pkg/simulation/` — Simulation Engines

- **engine.go** — Main simulation engine orchestrating Preisach/L-K solvers
- **multicell.go** — Multi-cell ensemble simulation for statistical variation studies

### `pkg/controller/` — Write/Verify Workflows

- **writer.go** — ISPP writer: pulse-train generation, target convergence, stress protocols

### `pkg/algo/` — Calibration

- **calibration.go** — Automated calibration routines fitting model params to experimental data

### `pkg/gui/` — Fyne GUI

- **gui.go** — Main module GUI layout and lifecycle
- **embedded.go** — Embeddable app for the unified launcher
- **simulation.go** — GUI-driven simulation control
- **physics_engine.go** — Physics engine bridge for GUI
- **plot_view.go** — P-E loop and waveform plotting
- **controls.go** — Parameter sliders, mode selectors
- **widgets/** — Custom widgets: P-E plot, ISPP visualization, phase diagram, stability indicator, cell state, physics equations display

### `pkg/render/` — Rendering

- **plot.go**, **render.go** — Plot generation and export
- **vulkan.go** — Vulkan compute path (optional GPU acceleration)

### `pkg/tui/` — Terminal UI

- **tui.go** — Text-based interactive interface for headless environments

### `cmd/hysteresis/` — Standalone Entry Point

- **main.go** — CLI launcher for module 1

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `PreisachModel` | `pkg/ferroelectric` | Preisach hysteron ensemble, generates P-E loops |
| `Material` | `pkg/ferroelectric` | Material parameters (P_r, E_c, ε, thickness) |
| `LevelBins` | `pkg/ferroelectric` | Maps continuous polarization to discrete levels |
| `Engine` | `pkg/simulation` | Top-level simulation driver |
| `MultiCell` | `pkg/simulation` | Ensemble of cells with variation |
| `Writer` | `pkg/controller` | ISPP write-verify loop with convergence tracking |
| `Calibration` | `pkg/algo` | Fit model to experimental P-E data |

## Testing

```bash
# Run all module 1 tests
go test ./module1-hysteresis/...

# With race detector
go test -race ./module1-hysteresis/...

# Verbose (see individual test names)
go test -v ./module1-hysteresis/...
```

Key test suites:
- `pkg/controller/` — ISPP convergence, L-K tuning, stress protocols, writer regression
- `pkg/algo/` — Calibration accuracy
- `pkg/gui/` — Plot view rendering, widget layout

## Physics Context

**Preisach Model:** Decomposes the macroscopic hysteresis loop into a distribution of elementary rectangular hysterons on the (α, β) half-plane. The integral over the Preisach density function ρ(α, β) yields the polarization P for a given field history.

**Landau-Khalatnikov:** Time-domain ODE approach: γ(dP/dt) = -∂G/∂P + E(t), where G is the Landau free energy with coefficients α, β, γ. Captures switching dynamics and transient behavior.

**ISPP:** Incremental Step Pulse Programming applies voltage pulses of increasing amplitude, verifying the cell state after each pulse until the target conductance level is reached within tolerance.

**Key metrics:** P_r (remnant polarization, µC/cm²), E_c (coercive field, kV/cm), 2P_r (memory window), endurance (cycles), retention (seconds).

## Related Documentation

- `docs/2-learn/module1-hysteresis/` — ELI5, features, physics, open-source tools
- `docs/2-learn/module1-hysteresis/physics.md` — Preisach theory, L-K derivations, materials reference
- `docs/1-getting-started/cli-reference.md` — Command-line interface reference
