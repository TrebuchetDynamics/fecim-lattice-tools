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

- **calibration.go** — Runtime field-table feedback and native monotonic level-table repair


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

## Bounded programming and calibration

[`NewProgrammer(*Session, ProgramConfig)`](../workbench/hysteresis/programming.go)
shares the session drive and solvers, reusing `controller.WriteController` and
`LevelBins`. It is UI-free and in-memory; it does not change CLI/Fyne launch defaults.

```go
cfg := hysteresis.DefaultConfig()
cfg.Mode, cfg.Seed = hysteresis.Manual, 42
session, err := hysteresis.NewSession(cfg)
if err != nil { return err }
programmer, err := hysteresis.NewProgrammer(session, hysteresis.ProgramConfig{})
if err != nil { return err }
state, err := programmer.Apply(hysteresis.ProgramCommand{
    Kind: hysteresis.WriteTarget, Target: 22,
})
if err != nil { return err }
// Call again at explicit work boundaries; DT is physical seconds, not playback.
state, err = programmer.Advance(hysteresis.Budget{Steps: 64, DT: 0.01})
if err != nil { return err }
_ = state // Core contains canonical SI values and Go display/speech text.
```

- `TargetWrite` (default) follows native manual PREP/WRITE/DISPLAY; Preisach skips
  PREP on its first step, while L-K pre-biases. Targets use native **1-based**
  indexing; `Core.Level` stays zero-based. Clicking the current manual level is a
  no-op; another click during a manual write returns an ownership error.
  `WriteRead` queues the next click, applying it at the next PREP/WRITE boundary
  (including replacement at WRITE entry after PREP). Without a queued click,
  same-branch target selection uses a private RNG seeded by `Config.Seed`.
- Start from `DefaultProgramConfig()` to select `LinearISPP` or
  `LogarithmicISPP`, guard fraction, pulse/work limits and domain count. DCC is
  unimplemented and rejected. Domains 2–128 select seeded native polydomain L-K
  **only for WriteRead**; other workflows use one domain. NLS/noise remain off.
  The programmer releases ensemble mode when the job ends.
- While programming/calibration owns the drive, direct Session mutation and
  `Advance` return `ErrDriveOwned` (read-only snapshots/pages and SetView remain
  available). Use `CancelProgram` or `ResetProgram` through the programmer, then
  reconfigure. Reset clears target/queue/controller/history through the native
  session reset; Preisach retains its native physical polarization memory across
  turning-point reset, rather than becoming a freshly constructed cell.
- Each `Advance` accepts 1–256 work steps. Programming uses explicit physical
  `Budget.DT`; calibration uses the extracted native sweep timing, not playback
  or `DT`. Defaults cap a programming job at 100000 steps and 30 native
  total-plus-current pulses, and calibration at 128000000 solver calls (enough
  for the native worst-case 126 target writes at 64 levels). Configured
  limits may be smaller. Session lifetime work limits still apply to live writes.
  Cancellation (command or context), pulse/work caps and controller failure have
  distinct terminal statuses and never label partial work complete. Completed
  steps remain readable; a terminal job releases drive ownership and pauses.
  These are safety bounds, **not measured browser responsiveness guarantees**.
- `CalibrateProgram` explicitly starts an in-memory job with scratch solvers.
  [`controller.LevelCalibration`](pkg/controller/level_calibration.go) extracts
  the native Preisach zero-field binary search and single-domain L-K controller
  sweep, including monotonic repair. It is not a measured-data fit or the shared
  uniform field-spacing estimate. Native L-K ±Ec fallback estimates are counted
  in `FallbackTargets`; **fresh does not mean every target converged**.
  No partial table is exposed as a completed calibration. Calibration does not
  drive/reset the interactive cell. Legacy GUI calibration uses the same sweep
  but retains its adapter-owned cache/reset behavior.
- Calibration is `unavailable` before completion, `fresh` for the completed
  settings, or `stale` after material, engine, level count, target range,
  temperature, stress or seed changes. Reverting settings does not resurrect a
  stale table; explicit recalibration is required. Playback/view changes do not
  invalidate it. Only fresh tables seed controller feedback or pass
  `ExportCalibration()`, which returns owned **in-memory data**, not a file.
  The range remains 2–64 levels and 0.5–1.0 usable polarization fraction.
- The legacy GUI's 500 ms recalibration timer and automatic save are **not**
  reproduced in the workbench. There is no automatic timer, file logger, save,
  download or persistence; future local file operations need separate consent.
- `TimeResolved` returns `legacy_protocol_unavailable` with a Go explanation,
  never a switching trace or completion claim. Existing session history is not
  presented as switching-protocol samples. The legacy mode was a stub, not a
  protocol to restore in this migration.

The default **30 levels are educational simulation bins**, not demonstrated
hardware states. Material presets and calibration tables are simulation-only;
none of these workflows establish device validation or browser/Orca qualification.

## Independent diagnostics and read-only reference context

[`workbench/hysteresis/analysis.go`](../workbench/hysteresis/analysis.go) exposes
`NewAnalysis(AnalysisRequest)`, `AnalysisJob.Advance`, `Page` and `Result`. Supply a
session snapshot in `Input` and `AnalysisPUND` or `AnalysisFORC` in `Kind`. The job
copies the selected configuration and input revisions, resolves material values
from the Go catalogue, and never advances or mutates the loop. It starts no
background work: computation occurs only on explicit `Advance` calls.

Both diagnostics identify **independent Preisach simulations**, even when the
selected loop uses L-K. `Context` records the loop model, material, temperature
and stress; these are **context only, not applied diagnostic conditions**.
`Diagnostic` separately records the actual versioned profile, material and exact
applied SI parameters. Evidence remains `simulation_only`, with experimental
validation false.

- PUND uses `legacy-pund-demo` version 1, material `hzo-demo`: Ec=3e7 V/m,
  Ps=0.25 C/m², TanhEverett Delta=0.3×Ec, amplitude=5×Ec, width=100 ns,
  sample interval=5 ns and area=1e-12 m². These are the fixed native GUI demo
  parameters, **not** the selected material or the `DefaultHZO` preset. Results
  retain P/U/N/D current traces and integrated/switching charges in coulombs.
- FORC uses `gui-forc` version 1 and selected-material Ps, with defaults
  **Emax=1 V/m**, 21 reversals and resolution 21. The extracted
  `physics.GUIFORCEverett` retains the native weighted expression and clamps;
  this is not the headless TanhEverett profile or an L-K FORC experiment. Results
  retain descending sweep branches, reversal fields/pairs, the mixed-derivative
  density, and the GUI nearest-index resampled matrix (rows=Eb, columns=Ea).
  Reversals are limited to 3–64 (at most 64 recorded samples/curve), resolution
  to 3–256; nonfinite/unrepresentable field or density scales are rejected.
- `Advance(AnalysisBudget)` accepts 1–256 work units: one fixed PUND protocol
  (105 polarization updates), one FORC curve (at most 65 updates), the bounded
  density calculation (at most 64² cells), or one matrix row (at most 256 cells).
  Shared loops check context cancellation internally. Cancellation is terminal;
  create a new job to retry. Partial pages retain their running/cancelled state;
  `Result` and artifacts return `ErrAnalysisIncomplete` until fully complete.
  Pages require the current job revision, expose at most 32 rows and include
  diagnostic identity. Results/pages are owned copies. These limits are safety
  bounds, **not measured browser responsiveness or accessibility claims**.

[`ParseReference`](../workbench/hysteresis/reference.go) accepts bytes, not paths:
CSV columns `E_MV_cm,P_uC_cm2`, or JSON `source`, `material`, `E_V_m`, `P_C_m2`
and optional `branch`. Each input must contain one strictly monotonic branch;
input order, ascending/descending identity, source units and canonical SI values
are retained. Split full loops into explicit branches rather than sorting them.
Inputs are limited to 1 MiB total reference bytes/metadata per job, 32 branches
and **4096 points total across all branches** (also the standalone parser limit).
Aggregate admission is checked before expanding points into presentation records.
Malformed, nonfinite, duplicate-key (including case variants), conflicting or
unsupported inputs are rejected. Imported context is always `imported_unverified`, including
cited or synthetic inputs; opening bytes never proves experimental validation.

Go reference context includes a scoped model-background citation and labels the
native Pr ±20% / Ec ±30% bands `simulation_default`: illustrative preset ranges,
**not experimental uncertainty or measurement evidence**. Comparisons without
verified branch-matched data are unavailable with a reason, never fabricated
zero-valued fit metrics. The analysis API performs no fitting.

[`Artifact`](../workbench/hysteresis/artifacts.go) serializes only complete results
into in-memory bytes: `AnalysisJSON`, `PUNDCSV`, `FORCSweepCSV`, `FORCDensityCSV`.
JSON includes reference provenance and unavailable reasons; diagnostic CSVs
include profile/version/model, evidence, exact applied parameters and explicitly
context-only loop settings/revisions. Serialized output is capped at 8 MiB;
diagnostic-only CSVs do not copy reference points. These bounded admission/output
limits do not qualify production Worker latency. This performs no file I/O, upload, download,
persistence or stored-run verification; it does not authorize hosted exports.
Legacy headless diagnostics retain their distinct profiles and compatibility guards.

Focused diagnostic verification:

```bash
go test ./workbench/hysteresis -run TestAnalysis -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/gui -run TestAnalysisGUIParity -count=1
go test ./shared/physics -run 'TestPUND|TestFORC' -count=1
go test ./shared/viewmodel/hysteresis ./validation/peloop -count=1
```

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

Programming/calibration contract and native characterization checks:

```bash
go test ./workbench/hysteresis -run 'TestProgramming|TestTimeResolved' -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/gui -run TestWorkbenchSessionCharacterization -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module1-hysteresis/pkg/controller ./module1-hysteresis/pkg/algo ./module1-hysteresis/tests -count=1
env CGO_ENABLED=1 go test -race ./workbench/hysteresis -count=1
```

`TestProgrammingTrace`, `TestProgrammingCalibration`, `TestProgrammingCancellation`,
`TestProgrammingInputs` and `TestTimeResolvedUnavailable` exercise deterministic
writes/reset/ownership, calibration freshness and terminal/unavailable states.
Native characterization compares both ISPP methods/engines (including seeded
polydomain writes), queued targets and captured two-level calibration fields.

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
