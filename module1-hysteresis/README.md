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

## UI-free interactive session

[`workbench/hysteresis`](../workbench/hysteresis/session.go) provides the in-memory
session used by future UI/transport adapters. It imports no GUI package, starts no
background loop, and does not save presets, calibration or results. The existing
CLI/Fyne launch paths are unchanged. This API does not itself deliver a browser
application or qualify browser responsiveness/accessibility.

```go
// Import "fecim-lattice-tools/workbench/hysteresis".
cfg := hysteresis.DefaultConfig()
cfg.Mode = hysteresis.Manual
session, err := hysteresis.NewSession(cfg)
if err != nil { return err }
if _, err = session.Apply(hysteresis.Command{Kind: hysteresis.SetField, Field: -1e8}); err != nil { return err }
if _, err = session.Apply(hysteresis.Command{Kind: hysteresis.Run}); err != nil { return err }
snapshot, err := session.Advance(hysteresis.Budget{Steps: 1, DT: 0.01})
if err != nil { return err }
page, err := session.Page(hysteresis.PageQuery{Revision: snapshot.Revision, Offset: 0})
if err != nil { return err }
// Read page.Rows; use offsets 32, 64, ... through page.Total.
// E.SI is V/m; P.SI is C/m². Presentation contains Go display/speech text.
_ = page
```

`NewSession(Config{})` selects the same defaults as `DefaultConfig()`: native
literature-superlattice preference (then first available preset/DefaultHZO),
Preisach, Sine, **paused**, 0.5 Hz, playback scale 1, 300 K, 1 GPa, 30 educational
levels and 50000 retained trail samples. The target range comes from the selected
preset, with the native 0.98 fallback. Other configurations must be complete:
start from `DefaultConfig()` and replace fields explicitly. `Catalogue()` returns
owned preset copies with IDs/evidence and Go-computed `DefaultLevels` /
`DefaultRangeFraction` for material selection (native level limits applied).
To reproduce the native picker, copy those defaults along with `MaterialID` into
Configure; explicit level/range overrides are otherwise preserved. Initial
sessions still use the native global 30 levels. Snapshots add controls, supported
models/modes, assumptions and simulation-only evidence. These presets and levels
are **simulation defaults, not validated device measurements**.

- `Apply(Command)` supports Configure, SetField, SetView, Run, Pause, Cancel and
  Reset. Configure replaces the complete config atomically. SetField is accepted
  only in Manual and within ±2.5 effective Ec; it does not step polarization.
  Manual signed reversals use the same persistent Preisach/L-K solvers, not a
  freshly recomputed major loop.
- Frequency is positive with representable period/phase; the native 0.01–1 Hz
  slider is only a suggestion, not the entry's domain. Playback scale is
  1e-9–1. Temperature 200–700 K and stress 0–5e9 Pa apply to L-K; Preisach is fixed
  at 300 K/1e9 Pa. Levels are 2–64, target range 0.5–1, retention 1000–100000.
- `Advance(Budget)` accepts 1–256 solver calls and physical `DT` in (0, 1] seconds.
  Playback never rescales this timestep. Pause/Cancel retain science without
  advancing; Run resumes. An optional context is checked before each step;
  cancellation retains completed work and its revisions. Interactive L-K keeps
  native single-domain dynamics, NLS/noise disabled; Seed is explicit identity,
  not a global/time-based RNG seed.
- Reset clears E/time/history and resets the active solver (Preisach P=0; L-K
  starts at negative material Pr, with native fallbacks). It preserves settings
  and run/pause state. Material changes reset E/time/history and reconfigure the
  native solvers. Engine changes clear history, carrying P into L-K (negative
  default near zero) or resetting the Preisach stack. Mode/frequency changes clear
  history/time but retain model memory. Temperature changes greater than 25 K
  clear history; stress/levels/range/playback changes do not. Retention changes
  keep the newest samples. `ConfigRevision` advances on settings changes for
  downstream invalidation; view changes never change scientific state.
- Snapshots include at most 2000 supplementary plot points. Full-history or
  last-complete-cycle pages expose every retained row, including SI E/P, step
  index, physical time, branch and Go display/speech. L-K trail sampling preserves
  the native 5e-4 s interval; `SolverSteps` counts even unretained steps. `Count`
  and `Dropped` make ring eviction explicit. `Page` returns up to 32 rows and
  position/count; an old revision returns `ErrStaleRevision`, never mixed rows.
- Inputs with NaN, infinity, overflow, unsupported settings or malformed command
  payloads are rejected before mutation or input-sized allocation. Solver work
  is capped at 100000 steps between Reset/material changes, also bounding hidden
  Preisach reversal storage. `ErrWorkLimit` requires an explicit reset; settings
  and view changes do not silently renew the allowance. These are safety limits,
  **not measured responsiveness claims**.

[`workbench/presentation`](../workbench/presentation/value.go) converts canonical
SI E to MV/cm with `1e-8` and P to µC/cm² with `100`. Canonical floats stay intact;
display/speech preserve sign, zero and exponents. Unavailable presentation values
retain a reason instead of fabricating zero. Unsupported `FormatMetric` units
still report unavailable; SI callers use the explicit conversion functions.

## Testing

```bash
# Run all module 1 tests
go test ./module1-hysteresis/...

# With race detector
go test -race ./module1-hysteresis/...

# Verbose (see individual test names)
go test -v ./module1-hysteresis/...
```

Focused session/presentation and independent native characterization checks:

```bash
go test ./workbench/hysteresis ./workbench/presentation -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/gui -run 'TestWorkbenchSessionCharacterization|TestResetAllStateLocked_IsDeterministic|TestWaveformSwitch|TestLastCompleteCycle' -count=1
env CGO_ENABLED=1 go test -race ./workbench/hysteresis ./workbench/presentation -count=1
```

`TestSession` checks reversal memory, reset, invalid-input atomicity, retention,
revision paging and work/allocation bounds. `TestWorkbenchSessionCharacterization`
compares both engines and Manual/Sine/Triangle against native `UpdatePhysics(dt)`;
`TestHysteresisValues` checks dimensional conversion and speech. These automated
checks do not establish hardware validation or Orca qualification.

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
