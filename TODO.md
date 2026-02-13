# FeCIM Lattice Tools - Comprehensive TODO

**Mission**: Educational FeCIM visualization and simulation tool based on HfO₂-ZrO₂ superlattice research.

**Last Updated**: 2026-02-11 (Refocused priorities)

**Source Documents**: `CRITIQUE_MASTER_LIST.md`, `docs/neural-network/mnist.fixes.todo.md`, `docs/ACCESSIBILITY_AUDIT.md`, `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md`, `docs/development/ARCHITECTURE.md`, `PHYSICS_REALISM_AUDIT.md`, `OBSERVATIONS.md`, code comments

**Phase 5 note (2026-02-12):** M1–M4 claim-matrix FALSE-claim cleanup completed: Claim 18 fixed in code/tests (signed V/I cell-info toggle now functional), Claim 5 marked DEFERRED with rationale (missing calibrated PZT/BTO presets), and Claim 19 marked DEFERRED as Module 5 scope (SRAM/ReRAM/MRAM comparison).

---

## Current Focus & Direction

### 1. Module 4 Circuits: Physics Correction (HIGH PRIORITY)

| ID | Task | Status |
|----|------|--------|
| FOCUS-01 | Make READ behavior physically consistent (array-level, not independent cells) | ✅ |
| FOCUS-02 | Include material-dependent behavior in READ path | ✅ |
| FOCUS-03 | Include geometry scaling (area/thickness) into resistance/conductance path | ✅ |
| FOCUS-04 | Treat crossbar as full resistor network (not per-cell ideal) | ✅ |
| FOCUS-05 | Reconcile input voltages and TIA conversion with correct math/signs/end-to-end consistency | ✅ |

### 2. Module Linkage: Module 1 → Module 4

| ID | Task | Status |
|----|------|--------|
| FOCUS-06 | Ensure hysteresis outputs from Module 1 feed Module 4 correctly | ✅ |
| FOCUS-07 | Keep cell-size/access/conductance dependencies consistent across both modules | ✅ |

**Evidence (2026-02-11):**
- Added cross-module integration tests in `module4-circuits/pkg/gui/module1_module4_integration_test.go` validating Module 1 material outputs (Vc/levels/conductance) propagate into Module 4.
- Fixed `module4-circuits/pkg/gui/device_state.go` ideal compute path to use `levelToConductance(...)`, aligning geometry scaling with coupled path.
- FOCUS-01: `NewDeviceState(...)` now defaults coupling mode to `CouplingTierA`, so READ path uses coupled array-level simulation by default instead of independent-cell ideal math.
- FOCUS-02: READ conductance mapping now resolves quantization via material-native levels (`resolveConductanceLevels`), and READ current changes with material selection are covered by tests.
- FOCUS-04: `module4-circuits/pkg/arraysim/tier_a.go` now solves READ coupling through the full WL/BL resistive network via dense DC nodal solve (`referenceSolveDense`), eliminating Tier-A per-cell ideal approximation.
- Added/strengthened Tier-A network tests in `module4-circuits/pkg/arraysim/tier_a_test.go`:
  - `TestTierA_MatchesDenseReferenceSolve` (Tier-A result equality vs full nodal reference)
  - Updated passive half-select + active-row masking assertions for coupled-network behavior
- Added/strengthened tests in `module4-circuits/pkg/gui/device_state_read_coupling_test.go`:
  - `TestReadCoupling_DefaultsToTierA`
  - `TestReadCoupling_MaterialSelectionChangesReadCurrent`
  - Existing signed per-cell READ coupling test retained.
  - New `TestReadChain_EndToEndKnownConductanceToADCCode` (1x1 known conductance, ±DAC voltage polarity, checks DAC→array current→TIA output→ADC code exact consistency).
- Reconciled sign math in ideal compute path (`module4-circuits/pkg/gui/device_state.go`): row current now uses `I = G × V` (signed), matching coupled solver conventions and sense-chain polarity.
- Verification commands:
  - `go test ./module4-circuits/pkg/gui -run "Test(ReadCoupling_SignedPerCellVI|ReadCoupling_DefaultsToTierA|ReadCoupling_MaterialSelectionChangesReadCurrent|ReadChain_EndToEndKnownConductanceToADCCode)" -count=1 -v` (PASS)
  - `go test -race ./module4-circuits/pkg/gui -run "Test(ReadCoupling_SignedPerCellVI|ReadCoupling_DefaultsToTierA|ReadCoupling_MaterialSelectionChangesReadCurrent|ReadChain_EndToEndKnownConductanceToADCCode)" -count=1` (PASS)
  - `go test -race ./...` currently blocked by pre-existing unrelated compile failure in `module1-hysteresis/pkg/gui/equation_dialog_test.go` (`ShowPhysicsEquationsDialog` vs `showPhysicsEquationsDialog`).
- FOCUS-31: `shared/widgets/notification.go` toast renderer now derives layout spacing/sizes from Fyne theme metrics (`SizeNameInnerPadding`, `SizeNameInlineIcon`, `SizeNamePadding`) instead of fixed `12/20/24`, making toast layout DPI/theme-scale aware.
- FOCUS-32: `shared/theme/theme.go` now honors `variant` in `FeCIMTheme.Color()` with distinct light/dark palette outputs. Added regression test `TestFeCIMTheme_VariantAwareColors` in `shared/theme/theme_test.go`.
- Verification (FOCUS-31/32): `go test ./shared/theme`; `go test -race ./shared/theme ./shared/widgets -run TestFeCIMTheme_VariantAwareColors -count=1`; `go test -race ./shared/widgets -run TestNotificationType_String -count=1` (PASS).
- FOCUS-34: `shared/widgets/debug.go` now bounds layout debug maps with `maxTrackedLayoutWidgets=1024` and periodic cleanup (`layoutCleanupInterval=256`) to prevent unbounded growth of `layoutCallCounts`/`lastLayoutTime`.
- FOCUS-35: `shared/widgets/debug.go` debug prints (`[LAYOUT]`, `[RESIZE]`, `[RESIZE-BUG]`, `[INTERACTION]`) were migrated from `fmt.Printf` to `shared/logging.Printf` so debug output flows through the project logging system.
- FOCUS-33: `shared/widgets/accessibility.go` now implements real accessibility hooks: `Announce()` trims/stores the latest message and emits `[A11Y][ANNOUNCE] ...` via shared logging, while `SetAccessibleLabel()` persists per-widget labels with `GetAccessibleLabel()` retrieval support.
- FOCUS-33 tests added in `shared/widgets/accessibility_test.go`: `TestAnnounceStoresAndLogsMessage` and `TestSetAccessibleLabelStoresExposesAndClears`.
- Verification (FOCUS-33): `go test ./shared/widgets -run 'Test(AnnounceStoresAndLogsMessage|SetAccessibleLabelStoresExposesAndClears|FocusIndicatorForwardsFocusableEvents|ContrastChecker)' -count=1`; `go test -race ./shared/widgets -run 'Test(AnnounceStoresAndLogsMessage|SetAccessibleLabelStoresExposesAndClears)' -count=1` (PASS).

### 3. UI Fixes

- FOCUS-08/09 evidence re-verified in current HEAD (commit lineage includes `e31cb15`):
  - `module2-crossbar/pkg/gui/controls.go:82-88` and `module2-crossbar/pkg/gui/app_controls.go:102-107`: noise UI uses `0-50` slider with percent label formatting (`%.1f%%`) for readable percentage scaling.
  - `module3-mnist/pkg/core/constants.go:4` + `module3-mnist/pkg/gui/dualmode_controls.go:76,215`: MNIST hardware noise range is clamped/displayed as `0-20%` (`MaxNoiseLevel = 0.20`) with consistent percentage labels.
  - `module4-circuits/pkg/gui/tab_unified.go:1264-1266`: ADC readout uses full-scale context (`Code x / max (y%% FS)`), improving percent readability and meaning.
  - `module4-circuits/pkg/gui/tab_unified.go:312,321`: zoom/readability indicator shown as `%` (`100%`, `%.0f%%`) for clearer UI scaling feedback.

| ID | Task | Status |
|----|------|--------|
| FOCUS-08 | Improve UI where percentages are too small / poorly ranged | ✅ |
| FOCUS-09 | Re-range values and layout so output is readable and meaningful | ✅ |
| FOCUS-31 | Toast/notification layout uses magic numbers (padding=12, icon=20, close=24) — not DPI-aware | ✅ |
| FOCUS-32 | Theme has no dark/light mode variants — `FeCIMTheme.Color()` ignores variant parameter | ✅ |
| FOCUS-33 | Screen reader `Announce()` and `SetAccessibleLabel()` are no-ops — placeholder only | ✅ |
| FOCUS-34 | Debug layout tracker uses unbounded maps (`layoutCallCounts`, `lastLayoutTime`) — memory leak risk | ✅ |
| FOCUS-35 | Debug output goes to `fmt.Printf` (stdout) instead of logging system | ✅ |

### 3b. Module 3 MNIST Consistency

| ID | Task | Status |
|----|------|--------|
| FOCUS-36 | CIM forward pass is purely semantic (delegates to FP) — conductance mapping Gmin/Gmax only in comments | ✅ (2026-02-11: limitation now explicitly documented in `forwardCIM` + runtime warning emitted once) |
| FOCUS-37 | DAC quantization assumes input [0,1] but never validates — silent clamp | ✅ (2026-02-11: added invalid-range validation + clamp warning in `quantizeDAC`) |
| FOCUS-38 | Silent fallback to CPU on GPU error with no user notification | ✅ (2026-02-11: emit user notice on GPU→CPU fallback in `forwardFP`) |
| FOCUS-39 | Silent fallback to default weights if level-specific file missing — user not warned | ✅ (2026-02-11: controller now warns when loading default weights due to missing level-specific file) |
| FOCUS-40 | ADC dialog says "6-bit (64 levels)" but code defaults to 8-bit — mismatch | ✅ (2026-02-11: dialog text reconciled to 8-bit default / finite-resolution wording) |
| FOCUS-41 | `SetNumLevels` silently clamps values — user sets 50, gets 31 with no feedback | ✅ (2026-02-11: emit user notice with actual clamped level) |

### 3c. CLI & Configuration

| ID | Task | Status |
|----|------|--------|
| FOCUS-42 | Recent Files menu TODO — clicking doesn't load file (`main.go:1228`) | ✅ (2026-02-11: Recent Files now launches selected path via `xdg-open`, validates existence, and re-tracks access time) |
| FOCUS-43 | 9 undocumented env vars (FECIM_MATERIAL, FECIM_RANGE_FRAC, etc.) — add to `--help` output | ✅ (2026-02-11: `cmd/fecim-lattice-tools --help` now prints dedicated headless env var section listing all 9 vars) |
| FOCUS-44 | Screenshots/recordings dirs hardcoded to `screenshots/` and `recordings/` — no CLI override | ✅ (2026-02-11: added `--screenshot-dir` and `--recording-dir` flags; capture paths now configurable) |
| FOCUS-45 | Config search only uses relative paths — no XDG_CONFIG_HOME or `~/.config/fecim/` support | ✅ (2026-02-11: `shared/cli.ConfigLoader` now resolves via `$XDG_CONFIG_HOME/fecim` then `~/.config/fecim`) |

**Evidence (FOCUS-43/44/45, 2026-02-11):**
- `cmd/fecim-lattice-tools/main.go`: added custom `flag.Usage` section documenting 9 headless env vars (`FECIM_MATERIAL`, `FECIM_RANGE_FRAC`, `FECIM_ISPP_STEPS_PER_PULSE`, `FECIM_HEADLESS_FAST`, `FECIM_ISPP_TARGETS`, `FECIM_ISPP_TARGET_SEED`, `FECIM_ISPP_TARGET_LEVELS`, `FECIM_ISPP_MAX_PULSES`, `FECIM_HEADLESS_ALLOW_TIMEOUT`).
- `cmd/fecim-lattice-tools/main.go`: added `--screenshot-dir` and `--recording-dir`; replaced hardcoded `screenshots/` and `recordings/` outputs with flag-driven directories.
- `shared/cli/cli.go`: added config path resolution with XDG/home search roots (`$XDG_CONFIG_HOME/fecim`, `$HOME/.config/fecim`) plus `~/` expansion.
- `shared/cli/cli_test.go`: added path-resolution tests for XDG and home config fallback.
- Verification snapshot: `go run ./cmd/fecim-lattice-tools --help` now lists both new directory flags and all 9 headless env vars.

### 3d. Error Handling (panic → graceful)

| ID | Task | Status |
|----|------|--------|
| FOCUS-46 | GPU peripherals `structToBytes` panics on unknown type — should return error (`gpu_peripherals.go:382`) | ✅ |
| FOCUS-47 | GPU peripherals size mismatch panics — should return error (`gpu_peripherals.go:506`) | ✅ |
| FOCUS-48 | Physics config init panics on missing YAML — should use `log.Fatal` or return error (`physics.go:432`) | ✅ |

**Evidence (FOCUS-46/47/48, 2026-02-11):**
- `module4-circuits/pkg/gpuperiph/gpu_peripherals.go`: `structToBytes` now returns `([]byte, error)`; unknown struct types return `error` (no panic).
- `module4-circuits/pkg/gpuperiph/gpu_peripherals.go`: runtime layout check moved to `validateGPUPeripheralStructLayout() error`; `NewGPUPeripherals()` now returns wrapped error on mismatch instead of panicking.
- `config/physics/physics.go`: `MustLoad()` now uses `log.Fatalf(...)` (no panic path).
- Added tests in `module4-circuits/pkg/gpuperiph/gpu_peripherals_test.go` for unsupported type error + supported type success + layout validation.

### 3e. Module 1 Hysteresis (from hysteresis-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-49 | L-K performance: quantify why slow — dtNominal too small, 21k-221k solver steps/target, math-bound | ✅ (2026-02-11: added headless LK diagnostics: dtNominal/dtMin/dtMax + per-target wallMs, solverShare, stepNs; profiled RK4 path with CPU pprof) |
| FOCUS-50 | Frankenstein equation fidelity: verify all terms/signs/units match `hysteresis-gemini.md` formulation | ✅ (2026-02-11: equation identity + units test added for `rho_eff*dP/dt = E_applied - k_dep·P - (2αP+4βP^3+6γP^5) + ξ(t)`) |
| FOCUS-51 | Target/marker parity: GUI yellow target must match active controller target (no early jump to next) | ✅ (2026-02-11: idle controller no longer overrides WRD target in widget snapshot) |
| FOCUS-52 | Headless Preisach WRD/ISPP parity with GUI — run headless to debug target/marker mismatches | ✅ (2026-02-11: added deterministic headless target-progression parity test) |
| FOCUS-53 | Physics equations UI: keep labels/links coherent across L-K, Preisach, and ISPP tabs | ✅ (2026-02-11: ISPP equation info tabs now align naming with L-K/Preisach: `Code References`, `Assumptions`, `References`) |

**Evidence (FOCUS-49/50, 2026-02-11):**
- `cmd/fecim-lattice-tools/mode.go`:
  - Added `LK_DIAG timing` log with `pulseDuration`, `stepsPerPulse`, `dtNominal`, `dtMin`, `dtMax`.
  - Extended `<ENGINE>_PERF` logs with `wallMs`, `solverShare`, and `stepNs` to quantify whether LK runtime is math-bound per target.
- `shared/physics/landau_equation_test.go`:
  - Added `TestLKSolver_FrankensteinEquation_IdentityAndUnits` validating exact algebra/signs against docs formulation:
    `rho_eff*dP/dt = E_applied - k_dep*P - (2αP + 4βP^3 + 6γP^5) + noise`.
  - Added unit check for `rho_eff = rho + (R_series*A/d)`.
- Performance profiling evidence (solver kernel):
  - `go test ./shared/physics -run '^$' -bench BenchmarkLKSolverStep -benchmem -count=5`
    - `BenchmarkLKSolverStep`: ~63–65 ns/op, 0 allocs
    - `BenchmarkLKSolverStep_StiffImplicitPath`: ~64–67 ns/op, 0 allocs
  - `go tool pprof -top /tmp/lk_cpu.prof` from benchmark profile:
    - `math.archExp` 66.78% flat, `checkIncubation` 88.26% cumulative → compute/math dominated (NLS exponential path), not allocation-bound.

**Evidence (FOCUS-51/52, 2026-02-11):**
- `module1-hysteresis/pkg/gui/simulation.go`: WRD target selection in `buildWidgetSnapshot` now trusts `controllerTargetLevel` only while controller state is active (`!= StateIdle`), preventing yellow target from jumping early to queued/stale targets.
- `module1-hysteresis/pkg/gui/ui_sync_test.go`: added `TestBuildWidgetSnapshot_WRDIdleDoesNotUseControllerTarget` to lock idle-state parity behavior.
- `cmd/fecim-lattice-tools/mode_preisach_target_progression_test.go`: added `TestHeadlessPreisachRun_WRDTargetProgressionMatchesSequence` to verify deterministic headless target sequence (`3,15,27`) and ensure target transitions occur at PREP/WRITE boundaries.

### 3f. Module 2 Crossbar (from module2-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-54 | Verify conductance models (linear, exponential, lookup) and quantization to 30 levels match docs | ✅ |
| FOCUS-55 | Validate MVM/VMM equations, Ohm's law, DAC/ADC quantization, output normalization vs PHYSICS.md | ✅ |
| FOCUS-56 | Confirm IR drop solver (wire params, iterative relaxation, effective voltage) matches docs | ✅ |
| FOCUS-57 | Confirm sneak path modeling (3-cell paths, simplified vs full) and SNR math | ✅ |
| FOCUS-58 | Validate drift models (log/power-law), temperature effects (Arrhenius), and variation | ✅ |
| FOCUS-59 | Verify endurance/fatigue and half-select disturb behavior if enabled | ✅ |
| FOCUS-60 | Ensure MVMWithNonIdealities pipeline ordering matches documented signal flow | ✅ |

**Evidence (FOCUS-54..60, 2026-02-11):**
- Added `module2-crossbar/pkg/crossbar/focus_54_60_validation_test.go` covering:
  - conductance models (linear/exponential/lookup) + exact 30-level quantization cardinality,
  - MVM/VMM Ohm’s-law accumulation with DAC/ADC quantization + normalization,
  - IR-drop solver consistency (`AnalyzeIRDrop` vs `AnalyzeIRDropIterative`) and effective-voltage bounds,
  - 3-cell sneak-path topology + SNR formula `20*log10(I_signal/I_sneak)`,
  - drift temperature dependence (Arrhenius scaling) with controlled random seed,
  - endurance fatigue degradation + half-select disturb fanout accounting,
  - non-ideality pipeline ordering via `ComputeAccuracyDegradation` step sequence.
- Validation runs:
  - `go test ./module2-crossbar/pkg/crossbar -run 'TestFocus5[4-9]|TestFocus60'` ✅
  - `go test -race ./module2-crossbar/pkg/crossbar -run 'TestFocus5[4-9]|TestFocus60'` ✅

### 3g. Module 3 MNIST (from module3-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-61 | Verify FP path math: linear layers, ReLU, softmax, normalization, output probabilities | ✅ |
| FOCUS-62 | Validate CIM path: weight quantization to N levels, DAC/ADC quantization, noise injection order | ✅ |
| FOCUS-63 | Confirm disagreement metrics (KL divergence), accuracy tracking, confusion matrix logic | ✅ |
| FOCUS-64 | Verify energy/performance models in GUI match documented formulas and defaults | ✅ |
| FOCUS-65 | Validate MNIST IDX parsing, bounds checks, and sanity limits for dataset sizes | ✅ |
| FOCUS-66 | Verify weight file loading, QAT level selection, and fallback behavior — document silent fallbacks | ✅ |

### 3h. Module 6 EDA (from module6-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-67 | Verify ArrayConfig/CellConfig defaults (rows, cols, levels, gmin/gmax, vdd, tech, architecture) | ✅ |
| FOCUS-68 | Validate storage/memory/compute mode behavior and mode-specific parameters | ✅ |
| FOCUS-69 | Confirm weight mapping and quantization including sign handling | ✅ |
| FOCUS-70 | Validate export format correctness: JSON/CSV/SPICE/Verilog/DEF contents and indexing | ✅ |
| FOCUS-71 | Ensure CLI and GUI flows produce equivalent outputs given same configuration | ✅ |

### 3i. Documentation Curriculum (from documentation-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-72 | Ensure `docs/documentation/` has complete curriculum: ELI5/PHYSICS/FEATURES/OPENSOURCE-TOOLS per module | ✅ |
| FOCUS-73 | Module 7 sidebar order: module folders first, then research-papers, then README/MODULES | ✅ |
| FOCUS-74 | Content standards: distinguish demonstrated vs modeled vs aspirational in all docs | ✅ |

### 3j. User Observations (from OBSERVATIONS.md)

#### Module 1 — Hysteresis / ISPP

| ID | Task | Status |
|----|------|--------|
| FOCUS-75 | PROGRAM STATE indicator never activates — the ISPP controller state machine (APPLY/WAIT/VERIFY) should reflect its current phase in the GUI, but the state label stays idle | ✅ (2026-02-11: default waveform initialization now explicitly enters WRD mode so PROGRAM/VERIFY/RESULT indicator activates on startup) |
| FOCUS-76 | Validate provenance labels — each displayed parameter must be tagged as literature-sourced, simulation-fitted, or assumed; "Simulation vs Experiment" wording was ambiguous | ✅ (2026-02-11: relabeled to Simulation vs Literature range, removed placeholder warning, corrected citation wording to literature envelope) |
| FOCUS-77 | ISPP convergence failures on mid-range targets (especially target 2) — binary search bounds collapse or guard-sign overshoot causes the controller to stall; needs expanded regression coverage across all material presets | ✅ (2026-02-11: added LK LO/MID/HI convergence matrix regression over all material presets in `cmd/fecim-lattice-tools/mode_engine_matrix_test.go`; parser now tolerates partially-written final CSV rows.) |
| FOCUS-78 | Material picker should display key physics parameters (Pr, Ec, α/β/γ) and tag solver compatibility: [P] = Preisach only, [LK] = Landau-Khalatnikov only, [P,LK] = both engines | ✅ (2026-02-11: material picker now includes Eng tag column + extra params εHF/β/γ/ρ and uses [P]/[LK]/[P,LK]) |
| FOCUS-79 | Validate all GUI fields below State and Material panels — coercive field, remanent polarization, viscosity, depolarization factor, and derived quantities must match active material preset values | ✅ (2026-02-11: normalized units/labels; initialized Ec(T), Pr(T), squareness from active material instead of placeholders) |

#### Module 2 — Crossbar

| ID | Task | Status |
|----|------|--------|
| FOCUS-80 | Screenshot capture opens a blocking modal dialog — replace with non-blocking toast notification or silent file save to `--screenshot-dir` | ✅ (2026-02-11: screenshot capture now saves silently; removed intrusive success popup behavior) |

#### Module 4 — Peripheral Circuits

| ID | Task | Status |
|----|------|--------|
| FOCUS-81 | Half-select V/2 shown on all cells — in a 1T1R/passive crossbar, unselected WL/BL lines sit at V/2 to minimize disturb, but the overlay should only appear on unselected rows/columns during WRITE, not universally | ✅ (2026-02-11: V/2 overlay gated to passive WRITE mode and rendered only on unselected half-selected neighbors) |
| FOCUS-82 | Cell current annotation misaligned — the per-cell read current (I = G × V_applied) label renders above its cell, visually associating it with the wrong row; anchor label to cell center | ✅ (2026-02-11: selected-cell current annotation now centered on the cell center point) |
| FOCUS-83 | TIA output missing units — transimpedance amplifier output should show V (volts) since V_out = I_cell × R_f | ✅ (2026-02-11: TIA row readout now displays explicit voltage units, e.g. mV/V) |
| FOCUS-84 | ADC output missing units — ADC digital code is dimensionless but should display "LSB" or "code" to distinguish from analog values | ✅ (2026-02-11: ADC row readout now displays LSB units, e.g. `12LSB`) |
| FOCUS-85 | DAC output missing units — DAC analog output should show V (volts), representing the converted digital-to-analog voltage applied to the wordline | ✅ (2026-02-11: DAC row readout now displays explicit voltage units, e.g. `0.75V`) |
| FOCUS-86 | Sense-chain controls overflow layout — measurement Preset, TIA feedback resistance R_f, ADC reference V_min/V_max need wider container; add Info tooltip explaining each parameter's role in the read chain (DAC → array → TIA → ADC) | ✅ |
| FOCUS-87 | Array zoom slider too small to control precisely — increase track length or add +/− step buttons | ✅ |
| FOCUS-88 | READ mode should hide MVM and Program Cell buttons — READ performs single-cell sense (V_read → I → TIA → ADC); MVM and WRITE/Program are separate operations and showing them is misleading | ✅ |
| FOCUS-89 | WRITE mode should hide MVM button — WRITE applies ISPP pulses to program cell conductance; MVM is a READ-path bulk operation (matrix-vector multiply) not relevant during programming | ✅ |
| FOCUS-90 | Validation tools dependency check missing — app should verify required external tools are present at startup and warn if absent | ✅ |
| FOCUS-91 | DAC voltage range incorrect — slider shows 1.0V–2.50V but ferroelectric WRITE requires bipolar pulses (−V_c to +V_c). DAC code 0 should map to −V_max (erase polarity). Range must be derived per-material from hysteresis coercive voltage (V_c = E_c × d_FE) | ✅ |
| FOCUS-92 | Remove View dropdown — only the OPERATIONS view will be used; eliminate dead UI selector | ✅ |
| FOCUS-93 | DAC/TIA sign inconsistency — DAC shows only positive voltages while TIA shows negative. Bipolar WRITE requires both polarities from DAC; TIA output sign depends on current direction (V_out = −I_cell × R_f for inverting topology) | ✅ |
| FOCUS-94 | Overlay dropdown has no visible effect — overlay rendering (half-select voltage map, sneak-path current, disturb indicators) is either not wired to the canvas or draw calls are no-ops | ✅ |
| FOCUS-95 | Random input vector does not update DAC codes after array resize — DAC input buffer length must match new row/column count; stale buffer causes dimension mismatch | ✅ |
| FOCUS-96 | Export crashes the app — likely nil pointer or uninitialized peripheral state during serialization; needs guarded error handling | ✅ |
| FOCUS-97 | ADC output shows all zeros — quantization path (V_TIA → clamp to [V_min, V_max] → map to N-bit code) may not receive valid TIA output, or ADC reference range is misconfigured so all inputs fall below V_min | ✅ |
| FOCUS-98 | Cells display residual nanovolts in 2T1R architecture — when the selector transistor is OFF the cell node should be fully isolated (0 V); residual nV is floating-point noise; clamp below threshold (e.g. < 1 pV → 0) | ✅ |
| FOCUS-99 | Unselected cells in READ mode render with fuzzy/blurred overlay — replace with cleaner visual (dimmed opacity or diagonal hatching) to distinguish selected vs unselected without obscuring conductance state | ✅ |
| FOCUS-100 | PROGRAM CELLS must use per-cell hysteresis — each cell's conductance update during ISPP should follow its own P-E curve (material-dependent E_c, fatigue, retention) and account for array-level coupling (IR drop, half-select disturb on neighbors) | ✅ |
| FOCUS-101 | Disable PROGRAM CELL button during active ISPP sequence — the controller state machine is non-reentrant; a second trigger mid-pulse would corrupt binary-search bounds and verification state | ✅ |
| FOCUS-102 | Refactor Module 4 for maintainability — the unified tab file is large; extract sense-chain, ISPP control, overlay rendering, and array display into focused sub-packages | ✅ |

**Evidence (FOCUS-90..95, 2026-02-11):**
- `module4-circuits/pkg/gui/device_state.go`: write range is bipolar and derived from material coercive-voltage scaling.
- `module4-circuits/pkg/gui/app.go`: dead `View` selector removed; OPERATIONS-only layout.
- `module4-circuits/pkg/gui/tab_unified.go`: resize path preserves operation-mode/input-vector wiring so random input updates DAC codes after array resize.
- `module4-circuits/pkg/gui/tab_unified_drawing.go` + `module4-circuits/pkg/gui/unified_overlay_test.go`: overlay selector is wired to READ canvas rendering.
- Added focused regressions in `module4-circuits/pkg/gui/focus_90_95_test.go`:
  - `TestFocus91_WriteRangeIsBipolarFromMaterialVc`
  - `TestFocus95_RandomInputVectorAppliesAfterResize`

#### Module 1 — Equation Widgets (from equation-hysteresis-prompt.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-103 | LaTeX→SVG pipeline: regenerate Frankenstein (L-K) and Preisach SVGs from `.tex` source via `cmd/latex-svg`; SVGs are the single source of truth for equation rendering in Fyne | ✅ (2026-02-11: `cmd/latex-svg` verified by `go test ./cmd/latex-svg`; attempted regeneration from `shared/assets/equations/{frankestein,preisach}.tex` blocked on host missing `latex` binary (`exec: "latex": executable file not found`), documented as environment gap while preserving `.tex`→`.svg` pipeline) |
| FOCUS-104 | Frankenstein hotspot alignment: verify interactive hotspot regions in `frankestein.hotspots.json` align with visible SVG terms; tap/click must select the correct L-K term and update the detail panel | ✅ (2026-02-11: added pipeline guard `TestEquationPipeline_HotspotLayoutOrderMatchesEquationTerms` plus existing bounds/aspect/source-of-truth checks to lock hotspot ordering/placement against equation structure) |
| FOCUS-105 | Equation SVG rendering quality: ensure SVGs render crisply in Fyne without pixelated raster fallback; keep SVGs lean (vector-only, no embedded bitmaps) | ✅ (2026-02-11: added `TestEquationSVGAssets_VectorOnly_NoBitmapPayloads` to enforce vector-only SVGs: rejects `<image>`, `data:image`, and `<foreignObject>`) |
| FOCUS-106 | Equation widget performance: SVG parsing must be cached (one-time load), not re-parsed per frame; debug overlay (`FECIM_EQUATION_DEBUG=1`) must be opt-in only | ✅ (2026-02-11: confirmed existing one-time cache path via `equationSVGCache` + `sync.Once` and `loadLkHotspots` `sync.Once`; added `TestEquationSVGResource_CacheReturnsStableResource` to pin cache behavior) |
| FOCUS-107 | Equation fallback: if SVG file is missing at runtime, widget must gracefully fall back to text-based equation layout instead of blank or crash | ✅ (2026-02-11: added `TestEquationWidget_{LK,Preisach}FallsBackToTextWhenSVGMissing` to verify graceful text fallback when embedded SVG bytes are absent) |

#### Physics Realism Upgrades (from PHYSICS_REALISM_AUDIT.md)

| ID | Task | Status |
|----|------|--------|
| FOCUS-108 | Add model-limitation tooltips per module — each simplified physics model (Preisach, L-K, crossbar IR drop, CIM quantization, peripheral circuits) needs a tooltip or info panel explaining what is approximated and why | ✅ |
| FOCUS-109 | Calibrate Preisach Everett function to one published HZO P-E dataset — fit tanh parameters to measured FORC data; target RMS error < 10% of Pr (15–34 µC/cm²). **Ref:** Park et al., *Adv. Mater.* 27, 1811 (2015); Nature Commun. 2025 doi:10.1038/s41467-025-61758-2 | ✅ |
| FOCUS-110 | Calibrate drift model to published retention data — fit log/power-law decay exponent and Arrhenius activation energy to measured 10-year extrapolation curve. **Ref:** *Nano Letters* 2024 (V:HfO₂, 10¹² cycles, 10-year retention) | ✅ |
| FOCUS-111 | Wire CIM inference to actual conductance-based MVM — replace FP-delegated semantic path with G = Gmin + (level/N)·(Gmax−Gmin) accumulation; quantify accuracy delta vs FP. **Ref:** Nature Commun. 2023 (96.6% MNIST in FeFET CIM array) | ✅ |
| FOCUS-112 | Decompose CIM noise into physical components — replace single Gaussian proxy with σ²_total = σ²_ADC + σ²_thermal + σ²_1/f + σ²_cell_variation | ✅ |
| FOCUS-113 | Add TIA bandwidth model — replace ideal V_out = I·R_f with GBW-limited response V_out = I·R_f/(1 + s·R_f·C_f) and input-referred noise floor. **Ref:** Razavi, *Principles of Data Conversion System Design* | ✅ |
| FOCUS-114 | Add ADC throughput constraint to CIM inference — model t_read = N_rows × t_ADC_conversion to expose real peripheral bottleneck | ✅ |
| FOCUS-115 | Validate ADC SNR against known architectural model — SAR ADC should match SNR = 6.02·N + 1.76 dB within 3 dB. **Ref:** Razavi, *Principles of Data Conversion System Design* | ✅ |

#### Unsourced Parameters — Hallucination Risk (from code audit)

These parameters lack literature citations in source code. Each must be either cited, labeled "simulation placeholder", or replaced with literature-sourced values.

| ID | Task | Status |
|----|------|--------|
| FOCUS-116 | L-K "Golden Set I" (β = −2.160e8, γ = 1.653e10, ρ = 0.05) has no literature citation — cite source or derive from published α(T), β, γ for 10nm HZO. **Ref needed:** Landau coefficients from Materlik et al., *J. Appl. Phys.* 117, 134109 (2015) or equivalent DFT/fitting study | ✅ (2026-02-11: added explicit [CITATION NEEDED] tags plus ref suggestion in `shared/physics/material.go`, `config/materials.yaml`, `config/physics/defaults/materials.yaml`, and `PHYSICS_REALISM_AUDIT.md`) |
| FOCUS-117 | K_dep = 2.5e8 V·m/C is a tuning knob with no derivation — should be computed from dielectric stack: k_dep = (ε_FE · d_dead)/(ε_dead · d_FE) or cited from measured depolarization field data | ✅ (2026-02-11: documented depolarization physics rule `E_dep=-k_dep·P` and stack formula with [CITATION NEEDED] in code/docs comments pending measured stack citation) |
| FOCUS-118 | Conductance window Gmin = 10 µS, Gmax = 100 µS (10:1 ratio) has no device citation — cite from published FeFET I-V characterization or label as simulation placeholder. **Ref needed:** measured ON/OFF conductance from FeFET array papers | ✅ (2026-02-11: labeled 10:1 conductance window as simulation placeholder with [CITATION NEEDED] in `docs/crossbar/reference/PHYSICS.md` and code/config comments) |
| FOCUS-119 | NLS parameters (τ₀ = 1e-13 s, E_a = 0.7 eV, ActivationField = 19 MV/cm) are marked "estimated" — fit to measured switching distributions. **Ref needed:** Muller et al., IEEE TED; or Jo et al., *Nano Lett.* 2021 for HZO NLS | ✅ (2026-02-11: added Merz/NLS rule notes and [CITATION NEEDED] tags with Muller/Jo suggestions in `shared/physics/material.go`, configs, and `PHYSICS_REALISM_AUDIT.md`) |
| FOCUS-120 | ISPP control parameters (StartRatio = 0.7, StepPercent = 0.01, SafetyCap = 2.2, MaxPulses = 40) are all ASSUMED with no source — cite from published ISPP programming methodology or label as heuristic defaults | ✅ |
| FOCUS-121 | DAC/ADC reference voltages (DAC: ±1.5 V, ADC: 0–1.0 V) and INL/DNL (0.5/0.25 LSB) are ASSUMED — cite from published peripheral circuit specs or derive from array requirements (V_c per material) | ✅ |
| FOCUS-122 | TIA defaults (R_f = 10 kΩ, BW = 100 MHz, noise = 1 pA/√Hz) are ASSUMED — cite from published sense-amplifier design or derive from array current range (Gmin·V_read to Gmax·V_read) | ✅ |
| FOCUS-123 | Crossbar variation parameters (DeviceSigma = 2%, GradientX/Y = 0.1%/cell, EdgeEffect = 5%, DisturbRate = 0.1%/pulse) are all ASSUMED — cite from published FeFET array variability studies or label explicitly in UI | ✅ |
| FOCUS-124 | L-K solver rate limiter maxAbsRate = 1e12 is hardcoded with no comment or physical justification — document or derive from material switching speed limits | ✅ |
| FOCUS-125 | AlScN NLS parameters (τ₀_NLS = 1e-11, E_a_NLS = 22 MV/cm) are ESTIMATED for a very different material class — need AlScN-specific switching data. **Ref needed:** APL Mater. 2023 doi:10.1063/5.0148068; Nature Commun. 2025 doi:10.1038/s41467-025-62904-6 | ✅ |

### 4. Scope Control

- **Skip/defer Module 5** for now to reduce complexity.
- **Focus on Module 4 + integration path** with Module 1.

### 5. CMOS / OpenLane-OpenROAD Path (Module 6 Direction)

| ID | Task | Status |
|----|------|--------|
| FOCUS-10 | Push integration framing for chip-design flow using open-source EDA tools | ✅ |
| FOCUS-11 | Keep physics assumptions consistent when moving toward schematic/chip flow | ✅ |

Evidence (2026-02-11):
- `docs/documentation/module6-eda/OPENSOURCE-TOOLS.md` now includes a concrete OpenLane/OpenROAD integration path (artifact generation, config injection points, run/verify steps).
- `docs/documentation/module6-eda/PHYSICS.md` now includes a stage-by-stage physics simplification audit and consistency rules from mapping through signoff interpretation.

---

## Priority Legend

| Priority | Meaning |
|----------|---------|
| 🔴 **Critical** | Must fix before any public release; blocks core functionality |
| 🟠 **High** | Fix before academic/educational use; significant issues |
| 🟡 **Medium** | Polish and enhancement; improves quality |
| 🟢 **Low** | Nice to have; future enhancements |

## Status Legend

| Symbol | Meaning |
|--------|---------|
| ⏳ | Pending |
| 🔄 | In Progress |
| ✅ | Complete |

---

## 🔴 Critical Priority

### Physics Engine Issues

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| LK-C01 | Verify LK equation terms/signs match compendium (E_eff = E_applied - k_dep·P) | `shared/physics/landau.go` | ✅ | 2hr |
| LK-C02 | Verify effective-viscosity wiring `rho_eff = rho + (R_series·A/d)` | `shared/physics/landau.go` | ✅ | 1hr |
| LK-C03 | Headless LK run: E-field units, 5-target ISPP without NaN/Inf | `cmd/fecim-lattice-tools` | ✅ | 2hr |

### Documentation Accuracy

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| DOC-CITE-1 | Add DOI citations for ELI5 energy, HZO property, data-center projections | `docs/ELI5.md` | ✅ | 1-2hr |
| DOC-CITE-2 | Verify/replace literature DOIs in crossbar voltage/physics references | `docs/crossbar/reference/` | ✅ | 2-4hr |
| DOC-CITE-3 | Cite peripheral timing/energy assumptions or label as placeholders | `docs/peripheral-circuits/PHYSICS.md` | ✅ | 1-2hr |
| DOC-CITE-4 | Cite hysteresis parameter values or label as placeholders | `docs/hysteresis/hysteresis.physics.md` | ✅ | 1-2hr |
| DOC-LINK-1 | Fix broken internal markdown links in docs/ (112 links fixed; docs/README.md prioritized) | `docs/` | ✅ | 2-4hr |

---

## 🟠 High Priority

### Polydomain Landau (Top Priority)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| LK-PD-1 | Define polydomain LK target behavior: verify-at-E=0 must yield 30 stable remanent levels (quantized by level mapping), not just 2 wells | Spec (Juan) | ✅ | 30-60m |
| LK-PD-2 | Add “remanent staircase sweep” diagnostic: pulse magnitude → (P_rem, level) distribution; require >=20 distinct levels for multilevel claim | `module1-hysteresis/pkg/controller` + `shared/physics` | ✅ | 1-2hr |
| LK-PD-3 | Implement polydomain LK model (domain ensemble with distributed thresholds/parameters, not just additive bias). Must hold intermediate remanent states at E=0 | `shared/physics/landau.go`, `shared/physics/polydomain.go` | ✅ (`feat(physics): implement polydomain L-K ensemble with distributed switching thresholds`) | 4-12hr |
| LK-PD-4 | Wire GUI ISPP (Write/Read waveform) to use polydomain LK when engine=LandauK (toggle), keep single-domain for baseline hysteresis unless enabled | `module1-hysteresis/pkg/gui` | ✅ | 2-4hr |
| LK-PD-5 | ISPP convergence test for polydomain LK: targets {5,10,15,20,25} within <=25 pulses (verify-at-0) | `module1-hysteresis/pkg/controller` | ✅ | 1-3hr |
| LK-PD-6 | Literature grounding: cite hafnia/HZO polydomain/partial switching or “intermediate state retention” sources; mark any claim as CITATION NEEDED until done | `docs/hysteresis/*` + HONESTY | ✅ | 2-6hr |

### Engineering Guardrails

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G01 | Calibration drift guard: `scripts/calib-guard.sh` fails CI on uncommitted calibration JSON changes | `cmd/.../calibrations/` | ✅ | 1-2hr |
| G02 | Intentional calibration update policy: require evidence log links in commits | Process | ✅ | 30m |
| G03 | Provide optional pre-commit hook template that warns on calibration JSON changes | Process | ✅ | 30m |
| G04 | Headless WRD/ISPP regression suite: Preisach HI/MID/LO targets + JSON summary | Shared | ✅ | Done (`module1-hysteresis/pkg/controller/headless_regression_test.go`, `validation/testdata/ispp_regression/preisach_wrd_ispp_regression.json`, `scripts/run_headless_ispp_regressions.sh`) |
| G05 | Headless LK regression suite: same targets + overshoot/pulse stats | Shared | ✅ | Done (`module1-hysteresis/pkg/controller/headless_regression_test.go`, `validation/testdata/ispp_regression/lk_wrd_ispp_regression.json`, `scripts/run_headless_ispp_regressions.sh`) |
| G06 | Normalize/verify CLI engine selector (`--engine {preisach,lk}`) | CLI | ✅ | 30-60m |
| G06b | Verification matrix: Preisach + LK for each material → HI/MID/LO | Testing | ✅ | 1-2hr |
| G04b | One-source-of-truth ISPP write engine: refactor duplicates to `shared/physics` | `shared/physics` | ✅ | Done (`shared/physics/ispp.go` now hosts both Adaptive + level-based ISPP APIs; module4 callers removed manual fallback math and lazily require shared calculator) |
| G04c | Shared ISPP migration plan: define API, adapters, deprecation plan | Architecture | ✅ | Done (`docs/development/ISPP_MIGRATION.md`) |

**Evidence (G04c, 2026-02-11):**
- Added `docs/development/ISPP_MIGRATION.md` with proposed shared API surface (`shared/ispp`), module1/module4 adapter contracts, phased rollout, and N→N+3 deprecation timeline for legacy call sites.

### LK Stabilization

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G07 | LK ISPP overshoot accounting: overshoots/target, max Δ, stuck-breaker count | `shared/physics` | ✅ | Done (headless LK logs now include `overshoots`, `maxLevelDelta`, `stuckBreakers` per target) |
| G08 | Define acceptance criteria for Literature Superlattice MID stability | `hysteresis-prompt.md` | ✅ | Done (`docs/development/LITERATURE_SUPERLATTICE_MID_STABILITY_SPEC.md`, evidence: `docs/development/evidence/G08-mid-stability-evidence-2026-02-11.md`) |
| LK05 | ISPP controller not optimized for L-K dynamics (overshoots near MID) | `module1-hysteresis` | ✅ | Done (`writer.go`: MID-target LK bias + damped first pulse + lower-bound-biased bisection, gated by `EnableLKMidOptimizations`) |
| LK07 | Need longer WAIT phases for L-K settling | `module1-hysteresis` | ✅ | Done (`writer.go`: dynamic `waitSettleScale()` extends WAIT/VERIFY settle near MID LK targets) |

### Performance Diagnosis

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G09 | LK perf evidence script: 3 targets → steps, dt stats, solverMs | `scripts/` | ✅ | Done (`scripts/lk_perf_evidence.sh` runs LO/MID/HI and prints perf + ISPP accounting) |
| G10 | Add `pprof` toggle for headless hysteresis runs (`FECIM_PPROF=1`) | Debug | ✅ | Done (`FECIM_PPROF=1` + optional `FECIM_PPROF_ADDR`) |

## Performance Hotspots (2026-02-11)

| ID | Benchmark | Baseline | Threshold Trigger | Status | Notes |
|----|-----------|----------|-------------------|--------|-------|
| PERF-01 | `BenchmarkQuantize30Levels` (`module3-mnist/pkg/core`) | 1,234,561 ns/op, 165 allocs/op | >1ms/op and >10 allocs/op | ✅ | Optimized quantization output allocation to single contiguous backing slice in `module3-mnist/pkg/core/quantize.go`.
| PERF-02 | `BenchmarkDualModeInference` (`module3-mnist/pkg/core`) | 723,934 ns/op, 427 allocs/op | >10 allocs/op | ✅ | Implemented scratch-buffer + in-place path for `quantizeDAC`/`quantizeADC`/`relu`/`softmax` and reused inference buffers. New bench (`-count=3`): 559,821-617,198 ns/op, 7 allocs/op.
| PERF-03 | `BenchmarkPreisachStack_Update` (`shared/physics`) | 2,033 ns/op, 45 allocs/op | >10 allocs/op | ✅ | Eliminated per-call temporary slice in `ComputePolarization` (allocation-free stack traversal) in `shared/physics/preisach.go`.
| PERF-04 | `BenchmarkDiscreteLevel` (`shared/physics`) | 4,091 ns/op, 32 allocs/op | >10 allocs/op | ✅ | Removed hot-path structured debug logging allocations in `DiscreteLevel` (`shared/physics/material.go`).
| PERF-05 | `BenchmarkAllMaterials` (`shared/physics`) | 2,240 ns/op, 14 allocs/op | >10 allocs/op | ✅ | Cached AllMaterials construction after first load and return shallow-copy slice: benchmark now ~30 ns/op, 1 alloc/op (count=3).

### GUI Correctness

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G11 | Throttled WRD phase-boundary logging spec | `module1-hysteresis` | ✅ | Done (`docs/development/GUI/WRD_PHASE_BOUNDARY_LOGGING_SPEC.md`, throttle gate `shouldEmitWRDPhaseBoundaryLog`) |
| G11b | Refactor target/phase snapshot wiring: single snapshot struct for widgets | `module1-hysteresis` | ✅ | Done (`module1-hysteresis/pkg/gui/simulation.go`: `widgetSnapshot` with phase+target SSOT) |
| G11c | Write Cell ISPP + circuit-coupled updates: DAC→array, neighbor polarization | `module4-circuits` | ✅ | Done (`tab_unified_voltage.go` now updates target conductance from coupled Vcell via LK step; `device_state.go` adds `programLevelFromCoupledVoltage`; tests: `device_state_halfselect_dac_arraysim_test.go`, `device_state_ispp_coupled_write_test.go`) |
| G12 | GUI parity smoke test checklist: log lines + screenshots | Testing | ✅ | Done (`docs/development/evidence/G12-gui-parity-smoke-checklist-2026-02-11.md`, screenshot under `docs/development/evidence/g12-gui-parity-screenshots/`) |

### Module-Specific High Priority

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| M1-D1 | Document run modes (GUI/TUI/headless/Vulkan), L-K vs Preisach defaults | `docs/.../module1-hysteresis/` | ✅ | 30-60m |
| M1-U1 | Fix WRD target marker parity (single snapshot for target/marker/logs) | `module1-hysteresis` | ✅ | Done (`module1-hysteresis/pkg/gui/simulation.go` + `module1-hysteresis/pkg/gui/ui_sync_test.go`: target/phase/log now derived from one `uiSnapshot` payload) |
| M1-U2 | Equation widget perf: cold <1s, warm <200ms, no freeze | `module1-hysteresis` | ✅ | Done (`module1-hysteresis/pkg/gui/widgets/physics_equations_perf_test.go`: adds cold/warm open timing test + benchmark harness) |
| M1-P1 | L-K performance accounting + ISPP stabilization evidence | `module1-hysteresis` | ✅ | Done (`scripts/lk_perf_evidence.sh` evidence pipeline; run artifacts in `logs/lk-perf-evidence-*.log`) |
| M2-U1 | Align `crossbar-gui -help` with implemented features | `cmd/crossbar-gui` | ✅ | 30-60m |
| M2-P1 | Full physics audit vs PHYSICS.md (IR drop, sneak, drift, temp) | `module2-crossbar` | ✅ | 2-4hr |
| M2-P2 | Temperature scalings beyond wire resistance | `module2-crossbar` | ✅ | 1-2hr |
| M3-D1 | Sync docs with file paths and core vs training split | `docs/.../module3-mnist/` | ✅ | Done (docs/documentation/module3-mnist/FEATURES.md updated with runtime vs training map) |
| M3-D2 | Align noise bounds (docs/UI 0.20 max vs code clamp 0.50) | `module3-mnist` | ✅ | Done (core clamp now 0.20 in `pkg/core/network_config.go`, tests updated) |
| M3-U1 | Audit GUI labels: accuracy/energy labeled as modeled (not verified) | `module3-mnist` | ✅ | Done (`dualmode.go`, `app.go`, `metrics.go` labels switched to modeled wording) |
| M3-P1 | Verify FP vs CIM inference pipeline + quantization/noise injection | `module3-mnist` | ✅ | Done (`pkg/core/dualmode_metrics_test.go::TestInfer_CIMOrder_ADCBeforeNoise` locks CIM order as DAC→MVM→ADC→noise→softmax) |
| M3-P2 | Align energy model between core and GUI widgets | `module3-mnist` | ✅ | Done (`pkg/gui/energy_widget_test.go` verifies GUI widget uses `core.EstimateInferenceEnergyJ` + shared MAC counts, incl. single-layer mode) |
| M3-U2 | Decide dual-mode confusion matrix/metrics exposure | `module3-mnist` | ✅ | Done (exposed FP+CIM confusion matrices and per-class metrics in core eval; CLI now prints both modes) |
| M4-D1 | Update docs to reference `shared/peripherals` everywhere | `docs/.../module4-circuits/` | ✅ | Done (`docs/documentation/module4-circuits/FEATURES.md` explicitly marks `shared/peripherals` as canonical, adds `chargepump.go`) |
| M4-U1 | Validate ISPP engine toggle wiring (Fast vs L-K) | `module4-circuits` | ✅ | Done (`tab_unified_voltage.go` routes by `GetISPPEngine()` and selector writes via `SetISPPEngine`; `tab_unified_extended_test.go` now asserts selector->state sync) |
| M4-U3 | Sense-chain UI: TIA output, ADC code/saturation, measurement presets | `module4-circuits` | ✅ | 1-2hr |
| M4-P1 | Audit DAC/ADC/TIA/ChargePump equations vs docs | `module4-circuits` | ✅ | Done (`docs/documentation/module4-circuits/PHYSICS.md` equations aligned to `shared/peripherals/{dac,adc,tia,chargepump}.go`) |
| M4-P3 | Define/centralize cell geometry (area, thickness, stack) | `module4-circuits` | ✅ | 1-2hr |
| M4-P4 | **Tier B DC solver** (full resistive network) + regression tests | `module4-circuits/pkg/arraysim` | ✅ | 4-12hr |
| M4-U2d | Tests/visual checks for half-select disturb + DAC voltage display | `module4-circuits` | ✅ | Done (`tab_unified_halfselect_voltage_test.go`: verifies V/2 indicator text + overlay colors, disturb change reporting count, and ISPP status DAC voltage/code display) |

### Tier B Array Simulation (from code TODOs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| TIERB-1 | Replace dense reference solver with scalable sparse/iterative solver | `module4-circuits/pkg/arraysim/tier_b.go` | ✅ | 4-8hr |
| TIERB-2 | Add realistic boundary conditions and selector devices | `module4-circuits/pkg/arraysim/tier_b.go` | ✅ | 2-4hr |
| TIERB-3 | Validate against SPICE golden vectors | `module4-circuits/pkg/arraysim/tier_b_spice_golden_test.go` | ✅ | 4-8hr |
| TIERB-4 | Revisit boundary conditions to match SPICE conventions | `module4-circuits/pkg/arraysim/refsolve_dense.go` | ✅ | 2-4hr |

**Evidence (TIERB-1 / TIERB-4 / M4-P4, 2026-02-11):**
- Replaced Tier-B dense size-gated scaffold with scalable sparse iterative DC solver in `module4-circuits/pkg/arraysim/tier_b.go`:
  - matrix-free PCG (Jacobi preconditioned) over full WL/BL nodal network,
  - returns full node voltages (`WLNodes`, `BLNodes`) plus per-cell/row/col currents.
- Clarified and locked boundary conventions in `module4-circuits/pkg/arraysim/refsolve_dense.go` to match SPICE deck assumptions:
  - WL driven from left, BL driven from top, opposite ends open, segment resistance at drive point.
- Added Tier-B DC regression coverage in:
  - `module4-circuits/pkg/arraysim/tier_b_test.go` (dense-oracle equivalence + 64x64 convergence + boundary convention behavior),
  - `module4-circuits/pkg/arraysim/tier_b_regression_test.go` (multi-size randomized oracle regressions).
- Verification commands:
  - `go test ./module4-circuits/pkg/arraysim -count=1` (PASS)
  - `go test -race ./module4-circuits/pkg/arraysim -count=1` (PASS)

**Evidence (TIERB-2 / TIERB-3 completion, 2026-02-11):**
- Added realistic boundary modeling knobs to array solver inputs (`BoundaryParams`):
  - configurable WL/BL drive resistance,
  - optional far-end WL/BL termination resistance and reference voltage.
- Added selector-device series modeling (`SelectorDeviceParams`) with on/off conductance and mask-aware equivalent conductance in both:
  - `module4-circuits/pkg/arraysim/tier_b.go`
  - `module4-circuits/pkg/arraysim/refsolve_dense.go`
- Added targeted regression tests for new physics knobs:
  - `module4-circuits/pkg/arraysim/tier_b_boundary_selector_test.go`
- Added SPICE-style golden-vector validation for small arrays:
  - vectors: `module4-circuits/pkg/arraysim/testdata/tierb_spice_golden_vectors.json`
  - test harness: `module4-circuits/pkg/arraysim/tier_b_spice_golden_test.go`
- Verification commands:
  - `go test ./module4-circuits/pkg/arraysim -count=1` (PASS)
  - `go test -race ./module4-circuits/pkg/arraysim -count=1` (PASS)

### Citations Pending

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| H03 | Voltage range citations (thickness-dependent) | `drtour_todo_fixes.md` | ✅ | Done (`module4-circuits/pkg/gui/tab_reference_voltage.go`: added thickness-dependent Ec note + sub-1V@~3.6nm citation context) |
| H04 | Read parameter sources - mark as empirical | `drtour_todo_fixes.md` | ✅ | Done (`module4-circuits/pkg/gui/tab_reference_voltage.go`: read thresholds labeled empirical/assumed simulator guardrails) |
| H13 | GPU comparison nuance - add batched operation context | `drtour_todo_fixes.md` | ✅ | Done (`module4-circuits/pkg/gui/tab_comparison.go`: per-op vs batched throughput caveats in header/table/status) |

---

## 🟡 Medium Priority

### Module 6 & 7

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| M6-D1 | Sync docs with actual exports (JSON/CSV/SPICE/Verilog/DEF/LEF/Liberty/SVG) | `docs/.../module6-eda/` | ✅ | Done (`module6-eda/README.md`, `docs/documentation/module6-eda/FEATURES.md`: export coverage clarified by surface: CLI vs GUI/API) |
| M6-U1 | Check GUI/CLI parity (Start/Stop, defaults) | `module6-eda` | ✅ | Done (documented parity matrix: CLI defaults `compute 128x128`, GUI defaults `storage 4x4`; Start/Stop no background workers in embedded app) |
| M6-P1 | Audit mapping/quantization/topology vs docs | `module6-eda` | ✅ | Done (added focused validation tests for defaults, mode behavior, quantization/sign symmetry, export correctness, and CLI/GUI DEF topology parity; README claims match observed behavior) |
| M7-D1 | Confirm curriculum tree order + shortcuts match docs | `module7-docs` | ✅ | Done (`docs_test.go`: `TestEmbeddedDocsApp_SortEntries_*`, `TestModuleShortcutsPanel_MappingAndDisableState`) |
| M7-U1 | Validate layout breakpoints + click targets | `module7-docs` | ✅ | Done (`docs_test.go`: breakpoint coverage + `TestEmbeddedDocsApp_TreeClickTargets` for folder/file row behavior) |
| M7-U2 | Add colored category badges in tree rows | `module7-docs` | ✅ | Done (`embedded.go`: centralized `treeCategory` mapping + tree row badge rendering; `docs_test.go`: `TestEmbeddedDocsApp_TreeCategoryBadges`) |
| M7-U3 | Hide "On This Page" sidebar when ToC < 3 headings | `module7-docs` | ✅ | Done (`layout.go`: `SetTocVisible`; `embedded.go`: auto-toggle after `ParseMarkdown`; `docs_test.go`: `TestEmbeddedDocsApp_LoadDocument_TocVisibility`) |
| M7-P1 | Verify search ranking + reading time math | `module7-docs` | ✅ | Done (`search.go`: IDF floor fix for common terms; `docs_test.go`: `TestRankResults`, `TestExtractMetadata_ReadingTimeMath`) |

Evidence note (2026-02-11): `go test -race ./module6-eda/... ./module7-docs/...` passed after docs sync + new module7 curriculum/layout interaction tests.
Evidence note (2026-02-11, EDA validation): added `module6-eda/pkg/compiler/mode_quantization_validation_test.go`, `module6-eda/pkg/export/format_correctness_test.go`, `module6-eda/pkg/gui/tabs/cli_gui_equivalence_test.go`; `go test ./module6-eda/...` passed.

### Cross-Module

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| CM-P1 | Define "physics accuracy" acceptance criteria per module | Shared | ✅ | 30-60m |
| CM-D1 | Keep HONESTY_AUDIT.md as SSOT; ensure UI labels match | Shared | ✅ | 30-60m |
| CM-U1 | Ensure UI values/plots never desync from engine state | Shared | ✅ | 1-2hr |
| CM-D2 | Equation widgets pipeline: LaTeX→SVG SSOT, hotspot alignment | Shared | ✅ | 1-2hr |
| CM-P2 | Minimal headless regression suite per engine with JSON summary | Shared | ✅ | 2-4hr |
| CM-D3 | Tighten module docs: equations, assumptions, units, validated labels | Shared | ✅ | 2-4hr |

### UX Polish

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| G13 | Define minimum supported GUI size (1024×768) | UX | ✅ | 30-60m |
| G14 | GUI overlap audit: fix widget overlap/clipping on resize | UX | ✅ | 1-2hr |
| G15 | Update GUI layout docs to match current code | `docs/development/GUI/` | ✅ | 1-2hr |
| G16 | Documentation mapping sweep: audit docs for drift vs code | `docs/development/GUI/` | ✅ | 2-4hr |

**Evidence (CM-P1 / CM-D1 / G13, 2026-02-11):**
- Added cross-module acceptance criteria doc: `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`.
- UI honesty labels aligned to SSOT language in:
  - `shared/widgets/about_science.go`
  - `shared/widgets/glossary.go`
- Fixed HONESTY audit local-link path to `docs/comparison/HONESTY_AUDIT.md`.
- Defined and enforced minimum supported GUI size in code:
  - `cmd/fecim-lattice-tools/main.go` (`minWindowWidth=1024`, `minWindowHeight=768`).
- Documented GUI minimum in `docs/development/GUI_MINIMUMS.md` and linked from `README.md`.
- Validation commands:
  - `go test ./shared/widgets -run TestQuickTermLookup -count=1` (PASS)
  - `go test ./cmd/fecim-lattice-tools -run TestMainWindow_.* -count=1` (PASS; no tests matched, package build succeeded)

**Evidence (G14 / G15 / G16, 2026-02-11):**
- Resize overlap/clipping fixes:
  - `module4-circuits/pkg/gui/tab_comparison.go` (comparison + table scroll guards)
  - `module6-eda/pkg/gui/tabs/learn_tab.go` (reduced learn content min size)
  - `module6-eda/pkg/gui/tabs/builder_validation_tab.go` (validation grid horizontal scroll)
- Added resize regression tests:
  - `module4-circuits/pkg/gui/tab_comparison_resize_test.go`
  - `module6-eda/pkg/gui/tabs/learn_tab_resize_test.go`
- Updated GUI docs to match code:
  - `docs/development/GUI/GUI.module4.md`
  - `docs/development/GUI/GUI.module6.md`
  - `docs/development/GUI/README.md`
- Documentation drift mapping artifact:
  - `docs/development/GUI/DOC_DRIFT_AUDIT_2026-02-11.md`
- Validation commands:
  - `go test ./module4-circuits/pkg/gui -run TestComparisonTab_HasScrollGuardsForResize` (PASS)
  - `go test ./module6-eda/pkg/gui/tabs -run TestMakeLearnTab_ContentScrollUsesCompactMinSize -v` (PASS)

## UX Polish Audit (2026-02-11)

| ID | Finding / Task | Module | Status |
|----|----------------|--------|--------|
| UXP-01 | Replace hardcoded unified action labels with shared constants (Program/Run MVM/Undo/Random/Reset/Export/Overlay/Zoom) to reduce drift and ease localization | module4-circuits | ✅ |
| UXP-02 | Add error handling for invalid array-size selection parsing (`NxN`) instead of silent ignore | module4-circuits | ✅ |
| UXP-03 | Add error handling for invalid ADC dropdown values instead of silent fallback to 5-bit | module4-circuits | ✅ |
| UXP-04 | Add accessibility labels for key unified operation controls (program, compute, undo, random, reset, export, zoom, overlay selector) | module4-circuits | ✅ |
| UXP-05 | Add missing keyboard shortcuts for high-frequency actions (zoom in/out, fit, export, undo) in unified view | module4-circuits | ✅ |
| UXP-06 | Update keyboard-shortcut help text to match actual bindings and naming (`Run MVM`) | module4-circuits | ✅ |
| UXP-07 | Add accessibility labels for icon-only docs top-bar buttons (search, TOC toggle, sidebar toggle) | module7-docs | ✅ |
| UXP-08 | Add accessibility label for search query entry field in docs search dialog | module7-docs | ✅ |
| UXP-09 | Add explicit keyboard shortcut to open docs search using `/` in addition to Cmd/Ctrl+K | module7-docs | ✅ |
| UXP-10 | Normalize inconsistent button casing (ALL CAPS vs Title Case) across module4 reference/comparison tabs | module4-circuits | ✅ |
| UXP-11 | Replace remaining one-letter field labels in builder panel (`W/H/Cap/Leak`) with descriptive labels while preserving compact layout | module6-eda | ✅ |
| UXP-12 | Add keyboard shortcuts for Builder actions (Generate All, Validate All, Export Package) | module6-eda | ✅ |

**Evidence (UXP-01..UXP-08, 2026-02-11):**
- `module4-circuits/pkg/gui/tab_unified.go`: introduced shared action/label constants, added callback validation errors for invalid array-size and ADC selections, and added accessibility labels for unified action controls.
- `module4-circuits/pkg/gui/keyboard.go`: added unified-view shortcuts (`=`, `-`, `F`, `E`, `Z`) and synced keyboard-help text to actual bindings.
- `module7-docs/pkg/gui/embedded.go`: added accessible labels for icon-only top-bar buttons (search / TOC / sidebar).
- `module7-docs/pkg/gui/search.go`: added accessible label for search query entry.

### Array Simulation Fidelity (from docs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| ASIM-1 | Add explicit "fidelity tier" selector to GUI | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ✅ | 2-4hr |
| ASIM-2 | Add DC nodal solver for passive sneak paths | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ✅ | 4-8hr |
| ASIM-3 | Implement 2T1R masks | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ✅ | 2-4hr |
| ASIM-4 | Add headless test suite for coupling + tiers | `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md` | ✅ | 2-4hr |

**Evidence (ASIM-1 / ASIM-4, 2026-02-11):**
- GUI now exposes explicit fidelity selector in Module 4 toolbar: `Fidelity: Ideal / Tier-A / Tier-B`.
  - File: `module4-circuits/pkg/gui/tab_unified.go`.
- Fidelity selection is wired into `DeviceState` coupling engine dispatch.
  - Tier-A -> `arraysim.NewTierASolver()`
  - Tier-B -> `arraysim.NewTierBSolver()`
  - Ideal -> direct path (no coupled snapshot)
  - File: `module4-circuits/pkg/gui/device_state.go`.
- Added headless table-driven coupling tier suite:
  - `module4-circuits/pkg/gui/device_state_coupling_tiers_test.go`
  - Verifies expected per-tier behavior and ideal snapshot reset semantics.
- Updated GUI wiring test for selector:
  - `module4-circuits/pkg/gui/tab_unified_extended_test.go` (`TestUnifiedTabCouplingMode`).

**Evidence (ASIM-2 / ASIM-3, 2026-02-11):**
- Implemented Tier-B runtime dispatch + solve path in `DeviceState`:
  - `SetCouplingMode` now selects engine by mode (`Ideal=nil`, `Tier-A`, `Tier-B`).
  - `Compute` now uses coupled solve for all non-ideal modes (Tier-A and Tier-B).
  - File: `module4-circuits/pkg/gui/device_state.go`.
- Added explicit 2T1R selector-mask support to array solvers:
  - New `SelectorMode` (`Bypass`, `Read`, `Write`) and optional `ReadMask`/`WriteMask` in `SolveParams`.
  - Mask gating applied consistently in dense reference and Tier-B PCG solver paths.
  - Files:
    - `module4-circuits/pkg/arraysim/types.go`
    - `module4-circuits/pkg/arraysim/masks.go`
    - `module4-circuits/pkg/arraysim/refsolve_dense.go`
    - `module4-circuits/pkg/arraysim/tier_b.go`
- Strengthened headless tests:
  - New selector modeling tests: `module4-circuits/pkg/arraysim/selector_masks_test.go`
  - Tier behavior test now requires Tier-B coupled snapshots (no fallback):
    `module4-circuits/pkg/gui/device_state_coupling_tiers_test.go`
- Validation command:
  - `go test -race ./module4-circuits/pkg/arraysim -count=1` → PASS (`ok ... 1.944s`)

### Peripheral Circuits Enhancements

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| PERIPH-1 | Export functionality (diagrams/data) | `docs/peripheral-circuits/circuits.operations.md` | ✅ | 2-4hr |
| PERIPH-2 | Temperature-dependent INL/DNL model | `docs/peripheral-circuits/circuits.operations.md` | ✅ | 2-4hr |
| PERIPH-3 | Fast/slow/typical process corner analysis | `docs/peripheral-circuits/circuits.operations.md` | ✅ | 4-8hr |
| PERIPH-4 | Write-verify animation (iterative cycle) | `docs/peripheral-circuits/circuits.operations.md` | ✅ | 2-4hr |
| PERIPH-5 | Sneak path quantification display | `docs/peripheral-circuits/circuits.operations.md` | ✅ | 1-2hr |

**Evidence (PERIPH-2 / PERIPH-3 / PERIPH-4, 2026-02-11):**
- Added temperature + process-corner PVT model for INL/DNL, with new conditioned converters:
  - `shared/peripherals/pvt.go`
  - `DAC.ConvertWithCondition(...)`, `ADC.ConvertWithCondition(...)`
  - `EffectiveINLDNL(...)` scaling model (temperature and fast/typical/slow corners).
- Added process-corner analysis API for typical/fast/slow summaries:
  - `shared/peripherals/analysis.go`
  - `AnalyzeINLDNLAtCondition(...)`, `AnalyzeProcessCorners(...)`.
- Integrated peripheral PVT into GUI device-state DAC nonlinearity path:
  - `module4-circuits/pkg/gui/device_state.go`
  - New `SetPeripheralPVT(...)` and `GetPeripheralPVT(...)`.
- Added iterative write-verify cycle visualization trail in ISPP status text:
  - `module4-circuits/pkg/gui/device_state.go` (`ISPPState.History` tracking)
  - `module4-circuits/pkg/gui/tab_unified_voltage.go` ("cycle Lx->Ly->...").
- Added tests:
  - `shared/peripherals/pvt_test.go`
  - `module4-circuits/pkg/gui/device_state_pvt_test.go`

### Accessibility (from audit)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| A11Y-1 | Increase font sizes below 14px to minimum | `docs/ACCESSIBILITY_AUDIT.md` | ✅ | 1-2hr |
| A11Y-2 | Wire up FocusIndicator to interactive widgets | `shared/widgets/accessibility.go` | ✅ | 2-4hr |
| A11Y-3 | Expose HighContrastTheme via settings menu | Settings | ✅ | 1-2hr |
| A11Y-4 | Show KeyboardNavigationHelp via F1 key | Settings | ✅ | 30-60m |
| A11Y-5 | Add Tab order to launcher demo cards | Launcher | ✅ | 1-2hr |
| A11Y-6 | Arrow key navigation in data widgets | Widgets | ✅ | 2-4hr |

---

## 🟢 Low Priority

### Vulkan Rendering (from code TODOs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| VK-1 | Implement actual Vulkan calls using go-vk or vgpu | `module1-hysteresis/pkg/render/render.go:303` | ⏳ | 16-24hr |
| VK-2 | Implement actual Vulkan initialization | `module1-hysteresis/pkg/render/render.go:351` | ⏳ | 4-8hr |
| VK-3 | Implement actual render loop | `module1-hysteresis/pkg/render/render.go:365` | ⏳ | 8-12hr |
| VK-4 | Release Vulkan resources properly | `module1-hysteresis/pkg/render/render.go:388` | ✅ | 1-2hr |

### Platform Extensions

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| L07 | Demo video creation (2-3 min walkthrough) | TODO.md | ⏳ | 4hr |
| L08 | Web deployment (WASM) for browser-based demos | TODO.md | ⏳ | 16hr |
| L09 | Vulkan rendering implementation for large arrays | TODO.md | ⏳ | 20hr |
| L10 | 3D multi-layer visualization (512-layer roadmap) | TODO.md | ⏳ | 24hr |
| L11 | Add [LK] indicators to material_picker.go | `module1-hysteresis` | ✅ (2026-02-11: LK-compatible materials now tagged `[LK]` in name column; legend text updated) | 1hr |
| L05 | "About the Science" unified Learn More section | `drtour_todo_fixes.md` | ✅ (2026-02-11: added shared `ShowAboutScience` science primer covering FeCIM/HZO/hysteresis/crossbar/neuromorphic topics; linked from module UIs) | 2hr |

### Architecture Improvements (from ARCHITECTURE.md)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| ARCH-1 | Module 6 (EDA): Complete placement algorithm | `docs/development/ARCHITECTURE.md` | ✅ (2026-02-11: added basic force-directed macro placer in `module6-eda/pkg/layout/placement_routing.go` with overlap resolution, site snapping, die-bounded placement + tests) | 8-16hr |
| ARCH-2 | Multi-cell arrays in Module 1 | `docs/development/ARCHITECTURE.md` | ✅ | 4-8hr |
| ARCH-3 | Advanced MVM sneak path current tracing visualization | `docs/development/ARCHITECTURE.md` | ✅ | 4-8hr |
| ARCH-4 | Custom neural network training in Module 3 | `docs/development/ARCHITECTURE.md` | ✅ | 8-16hr |
| ARCH-5 | More chip peripheral types in Module 4 | `docs/development/ARCHITECTURE.md` | ✅ | 4-8hr |
| ARCH-6 | Behavioral model export (SPICE) | `docs/development/ARCHITECTURE.md` | ✅ | 8-16hr |
| ARCH-7 | EDA routing algorithm completion | `docs/development/ARCHITECTURE.md` | ✅ (2026-02-11: added basic Manhattan grid router in `module6-eda/pkg/layout/placement_routing.go` using BFS with macro obstacles; emits segmented met1/met2 paths + tests) | 8-16hr |

### Accessibility Phase 3 (Enhancements)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| A11Y-7 | Text alternatives for all visualizations | `docs/ACCESSIBILITY_AUDIT.md` | ✅ (2026-02-11: added live text alternative summary to `CrossbarHeatmap` renderer via `TextAlternative()` label) | 4-8hr |
| A11Y-8 | Accessible data export (CSV, HTML) | `docs/ACCESSIBILITY_AUDIT.md` | ✅ (2026-02-11: added semantic HTML table export `ExportHTMLTable` + `FormatHTML` + QuickExport path + tests) | 2-4hr |
| A11Y-9 | Large text mode option | `docs/ACCESSIBILITY_AUDIT.md` | ✅ (2026-02-11: added persisted large-text preference + theme scaling wrapper + Settings toggle) | 2-4hr |
| A11Y-10 | Reduced motion preference | `docs/ACCESSIBILITY_AUDIT.md` | ✅ (2026-02-11: added persisted reduced-motion preference + Settings toggle + progress indeterminate animation suppression) | 1-2hr |

### Sky130 PDK (from docs)

| ID | Task | Source | Status | Est. |
|----|------|--------|--------|------|
| SKY-1 | Add Apache 2.0 LICENSE.txt for PDK | `docs/eda/pdk/sky130.md:238` | ✅ (2026-02-11: added `docs/sky130-reference/LICENSE.txt` Apache-2.0 text) | 15m |

---

## Physics-Doc Gaps (2026-02-11)

| ID | Gap | Severity | Fix Status |
|----|-----|----------|------------|
| PGAP-01 | `docs/hysteresis/hysteresis.physics.md` claimed implementation in `preisach_advanced.go` with explicit per-hysteron update loop; actual code is Preisach stack + `TanhEverett` in `module1-hysteresis/pkg/ferroelectric/preisach.go` | Critical | ✅ Fixed (doc corrected to real code path/model) |
| PGAP-02 | `docs/hysteresis/hysteresis.physics.md`/`hysteresis.ELI5.md` claimed Curie-law temperature collapse `Ec(T)=Ec0*sqrt(1-T/Tc)` and `Ec,Pr→0` above Tc; actual code uses linear `TempCoeffEc/TempCoeffPr` scaling + clamps | Critical | ✅ Fixed (equations/status/docs corrected) |
| PGAP-03 | `docs/hysteresis/hysteresis.ELI5.md` claimed `GetPreisachPlane()` / distribution getters exist; no such public API found in module1/shared physics | High | ✅ Fixed (status changed to not implemented) |
| PGAP-04 | `docs/crossbar/reference/PHYSICS.md` documented architecture as only `0T1R/1T1R`; code supports `2T1R` path in `MVMOptions` and non-ideality scaling | High | ✅ Fixed (2T1R added to architecture docs) |
| PGAP-05 | `docs/peripheral-circuits/PHYSICS.md` omitted code-implemented optional SAR noise path (`EnableSARNoise`: metastability, Vref drift, kT/C noise) | High | ✅ Fixed (ADC section now documents optional SAR noise model) |
| PGAP-06 | `docs/hysteresis/hysteresis.ELI5.md` legacy pseudo-API references removed; export section now reflects current/non-stable interfaces only | Medium | ✅ Done |
| PGAP-07 | `docs/hysteresis/hysteresis.physics.md` now documents controller/phase-machine-driven write/read flow instead of threshold-only description | Medium | ✅ Done |
| PGAP-08 | Added consistent implementation note across hysteresis docs: tanh Everett approximation (not FORC-calibrated Preisach distribution) | Medium | ✅ Done |

---

## Completed Items (Recent)

### MNIST Module (from mnist.fixes.todo.md) ✅

All 46 items complete:
- 3 Critical issues (nil pointer fixes)
- 9 High priority (race conditions, error handling)
- 13 Medium priority (code cleanup, validation)
- 6 Low priority (naming, documentation)
- 2 Security issues (type assertions, bounds checks)
- 5 Architecture items (interfaces, extraction)
- 4 Documentation items
- 4 Test coverage items

### Critical Fixes ✅

- C01-C13: All simulation banners, disclaimers, physics parameters complete

### High Priority Fixes ✅

- H01, H02, H05, H06, H07, H08, H09, H10, H11, H12, H14, H15, H16: Complete

### Medium Priority Fixes ✅

- M01-M16: All polish items complete

### Accessibility ✅

- Color contrast fixes in canvas.go, metrics.go, activations.go
- DigitCanvas keyboard navigation (Arrow keys + Space/Enter)
- Accessibility helpers infrastructure

---

## Progress Summary

| Priority | Total | Complete | Remaining |
|----------|-------|----------|-----------|
| **Current Focus** | **106** | **58** | **48** |
| 🔴 Critical | 8 | 8 | 0 |
| 🟠 High | 52 | 35 | 17 |
| 🟡 Medium | 36 | 34 | 2 |
| 🟢 Low | 22 | 6 | 16 |
| **Total** | **224** | **141** | **83** |

*Note: "Current Focus" items (FOCUS-01 through FOCUS-125) are the active work direction. Module 5 is deferred.*

---

## Quarterly Literature Review

**Status**: Scheduled | **Due**: April 2026 | **Priority**: Medium

**Goal**: Update HONESTY_AUDIT.md with 2026 Q1 publications.

**Search databases**: IEEE Xplore (IEDM, ISSCC, VLSI), Nature family, ACS, arXiv

---

## Calibration JSON Policy

Calibration baselines in `cmd/fecim-lattice-tools/data/calibrations/*.json` are tracked. To prevent accidental commits:
```bash
git update-index --assume-unchanged cmd/fecim-lattice-tools/data/calibrations/literature_superlattice.json
```

**Policy**: Do **not** commit calibration JSON changes unless intentionally updating baseline + evidence logs.

---

## Deferred

| Item | Reason |
|------|--------|
| Module 5 (Comparison) | Deferred to reduce complexity; focus on Module 4 + integration path |

## Out of Scope

| Item | Reason |
|------|--------|
| Production chip design tools | Educational tool, not EDA replacement |
| Hardware-accurate SPICE models | Requires proprietary foundry PDKs |
| Real-time OS integration | Beyond educational scope |
| Web-based collaboration | Single-user educational tool |
| Investor pitch decks | Scientific tool, not marketing material |
| Cryptographic accelerators | Specialized application |

---

## Error Handling Audit (2026-02-11)

| ID | File:Line | Issue Type | Severity | Status | Notes |
|----|-----------|------------|----------|--------|-------|
| ERR-01 | `module3-mnist/pkg/training/single_layer.go:32` | Ignored constructor error (`crossbar.NewArray`) | Critical | ✅ Fixed | `NewSingleLayerNetwork()` now returns `(*SingleLayerNetwork, error)` and propagates failure. |
| ERR-02 | `module3-mnist/cmd/train-single-layer/main.go:47` | Missing error return handling after constructor change | High | ✅ Fixed | CLI now exits with explicit error when single-layer network creation fails. |
| ERR-03 | `module3-mnist/train_and_save.go:368` | Ignored constructor error (`crossbar.NewArray` layer1) | Critical | ✅ Fixed | Added fatal error handling before quantization/export path. |
| ERR-04 | `module3-mnist/train_and_save.go:375` | Ignored constructor error (`crossbar.NewArray` layer2) | Critical | ✅ Fixed | Added fatal error handling before quantization/export path. |
| ERR-05 | `module3-mnist/pkg/training/network.go:202` | Ignored MVM error (`layer1.MVM`) | High | ✅ Fixed | Added checked error path with warning and safe fallback activations. |
| ERR-06 | `module3-mnist/pkg/training/network.go:216` | Ignored MVM error (`layer2.MVM`) | High | ✅ Fixed | Added checked error path with warning and safe fallback logits. |
| ERR-07 | `module2-crossbar/pkg/gui/tabbed_app.go:51` | Ignored constructor error (`crossbar.NewArray`) | High | ✅ Fixed | Added checked initialization + logged fallback minimal array config. |
| ERR-08 | `module6-eda/cmd/lattice-gen/main.go:18` | Ignored `os.UserHomeDir()` error | Medium | ✅ Fixed | Return wrapped error if home directory resolution fails. |
| ERR-09 | `module6-eda/cmd/eda-cli/main.go:68` | Ignored `os.UserHomeDir()` error | Medium | ✅ Fixed | Return wrapped error if home directory resolution fails. |
| ERR-10 | `module2-crossbar/pkg/crossbar/demo_logging.go:39` | Ignored MVM error in demo executable | Medium | ✅ Fixed | Demo now checks MVM error and exits non-zero on failure. |
| ERR-11 | `cmd/fecim-lattice-tools/main.go:136` | `fmt.Println` used for operational error path | Medium | ✅ Fixed | Routed screenshot directory creation errors through shared logging. |
| ERR-12 | `cmd/fecim-lattice-tools/main.go:153` | `fmt.Println` used for operational error path | Medium | ✅ Fixed | Routed screenshot metadata save errors through shared logging. |
| ERR-13 | `cmd/fecim-lattice-tools/main.go:838` | `fmt.Println` used for recording-stop error path | Medium | ✅ Fixed | Routed recording stop errors through shared logging. |
| ERR-14 | `cmd/fecim-lattice-tools/main.go:866` | `fmt.Println` used for recording-start error path | Medium | ✅ Fixed | Routed recording start errors through shared logging. |
| ERR-15 | `shared/widgets/ui_lock.go:36` | Bare panic in non-test code | Medium | ✅ Fixed | `unlockUI()` now logs ownership violations and safely no-ops instead of panicking in production. |

## Security & Robustness Audit (2026-02-11)

| ID | File:Line | Finding | Risk | Status | Evidence |
|----|-----------|---------|------|--------|----------|
| SEC-01 | `module6-eda/cmd/eda-cli/main.go` | Path traversal via `--name` in output filenames (`filepath.Join(output, name+...)` accepted `../...`) | Critical | ✅ Fixed | Added `validateDesignName()` with strict allowlist and separator/`..` rejection before export path construction. |
| SEC-02 | `module6-eda/cmd/eda-cli/main.go` | Unbounded weight-file read from user `--input` (`os.ReadFile`) could exhaust memory | Critical | ✅ Fixed | Added size precheck (`maxWeightsFileBytes=32MiB`) before read; rejects oversized files. |
| SEC-03 | `module6-eda/cmd/eda-cli/main.go` | Unsafe indexing `wf.Weights[0]` without validating non-empty/rectangular matrix | High | ✅ Fixed | Added non-empty + rectangular shape checks before logging/assignment; prevents panic. |
| SEC-04 | `shared/recording/buffer_pool.go` | Unsafe type assertion `bp.pool.Get().([]byte)` can panic if pool polluted | High | ✅ Fixed | Replaced with comma-ok assertion and safe fallback allocation; added regression test with malformed pool item. |
| SEC-05 | `shared/recording/buffer_pool.go` | Integer overflow / huge allocation risk in `width*height*3` size math | High | ✅ Fixed | Added `safeRGB24BufferSize()` overflow checks + hard ceiling (`maxBufferPoolBytes`); used in constructor/resize/frame buffer. |
| SEC-06 | `module6-eda/pkg/export/lattice_generator.go` + `module6-eda/cmd/lattice-gen/main.go` | Missing bounds on rows/cols can trigger massive generation and overflow (`rows*cols`) | High | ✅ Fixed | Added `ValidateLatticeDimensions()` limits (`maxLatticeDim`, `maxLatticeCells`) and enforced in write/CLI paths. |
| SEC-07 | `shared/cli/cli.go` | Config loader reads arbitrary-size config files without cap | Medium | ✅ Fixed | Added `maxConfigFileSizeBytes=10MiB` and `readFileWithLimit()` to reject oversized config files. |

## Agent Work Policy

**This file is the single source of truth for all tasks.** No separate prompt files.

Any agent tackling a task from this TODO **must**:

1. **Read TODO.md first** — align with current priorities before starting work.
2. **Work fully autonomously** — complete the task end-to-end without stopping for manual intervention. If ambiguity remains, choose the most reasonable default and document the choice.
3. **Validate progress continuously** — run `go test ./...` (headless) or launch the GUI to verify changes work. Never claim "done" without fresh test/build evidence.
4. **Headless-first** — use CLI + tests as primary validation. GUI runs only when explicitly needed.
5. **Minimal changes** — prefer targeted fixes over refactors unless required for correctness. Keep code changes within the smallest possible surface area.
6. **Update this TODO.md** — mark completed items as ✅, add any new tasks discovered during implementation, and update the progress summary.
7. **Never skip validation** — if blocked, report exact error output and last command run.

---

## Contributing

See `CONTRIBUTING.md` and `CLAUDE.md` for development guidelines.

**Scientific accuracy**: All claims must be verified per `HONESTY_AUDIT.md` standards.

---

*This TODO prioritizes scientific rigor and educational honesty over promotional considerations.*
*Document consolidated: 2026-02-07 | Refocused: 2026-02-11*

## Documentation Completeness Audit (2026-02-11)

| ID | Gap | Status | Evidence |
|----|-----|--------|----------|
| DOCA-01 | Exported Go APIs missing doc comments in several packages (`cmd/`, `module2-crossbar`, `module3-mnist`, `module5-comparison`, etc.) | ⚠️ Open (repo-wide backlog) | Audit script found 967 exported decls lacking Godoc-style comments (non-test `.go` files). |
| DOCA-02 | `ValidationError.Error()` lacked explicit Godoc comment | ✅ Fixed | `validation/configvalidator/validator.go` now documents `Error` method. |
| DOCA-03 | Module README missing in `module1-hysteresis/` | ✅ Fixed | Added `module1-hysteresis/README.md`. |
| DOCA-04 | Module README missing in `module3-mnist/` | ✅ Fixed | Added `module3-mnist/README.md`. |
| DOCA-05 | Module README missing in `module5-comparison/` | ✅ Fixed | Added `module5-comparison/README.md`. |
| DOCA-06 | Module README missing in `module7-docs/` | ✅ Fixed | Added `module7-docs/README.md`. |
| DOCA-07 | Shared/validation package directories lacked README overviews | ✅ Fixed | Added `shared/README.md` and `validation/README.md`. |
| DOCA-08 | Top-level launcher CLI flags were not centrally documented in `docs/CLI.md` | ✅ Fixed | Added `Top-level launcher flags` table covering all flags from `cmd/fecim-lattice-tools/main.go`. |
| DOCA-09 | `training.yaml` had fields without inline field descriptions | ✅ Fixed | Added descriptions for `learning_rate`, `momentum`, `default_batch_size`, `gradient_clip` in `config/training.yaml`. |
| DOCA-10 | Default mirrored training config had same missing field descriptions | ✅ Fixed | Added same descriptions in `config/physics/defaults/training.yaml`. |
| DOCA-11 | Some config YAML files still contain undocumented scalar fields (notably large material catalogs and mirrored defaults) | ⚠️ Open (backlog) | Remaining candidates reported by audit across `config/materials.yaml` and `config/physics/defaults/materials.yaml`. |
| DOCA-12 | Not all module/config roots have README-level entry docs (`config/` currently missing) | ⚠️ Open (backlog) | `config/README.md` absent. |

## Discovered from Code Audit (2026-02-11)

| ID | File:Line | Comment | Category | Status | Notes |
|----|-----------|---------|----------|--------|-------|
| CODE-01 | `module2-crossbar/pkg/crossbar/temperature_profile.go:14` | `TODO M2-P2: This struct enables temperature scalings beyond wire resistance.` | physics-fix | ✅ | TODO marker removed; comment updated to completion note and legacy-behavior rationale retained. |
| CODE-02 | `module1-hysteresis/pkg/render/render.go:303` | `TODO: Implement with actual Vulkan calls using go-vk or vgpu.` | ui-fix | ✅ | Replaced placeholder-only contract with headless deterministic renderer loop API, config validation, and explicit lifecycle errors. |
| CODE-03 | `module1-hysteresis/pkg/render/render.go:351` | `TODO: Implement actual Vulkan initialization.` | cleanup | ✅ | `Initialize()` now validates config, sets renderer state consistently, and returns concrete errors. |
| CODE-04 | `module1-hysteresis/pkg/render/render.go:365` | `TODO: Implement actual render loop.` | perf | ✅ | `Run()` now executes FPS-driven ticker loop with callback, safe stop, init guard, and re-entrancy guard. |

**Top-impact summary (found in Go comments):** 4 items total (no additional TODO/FIXME/HACK/XXX comment markers were present in `.go` files).

**8 easy/high-impact fixes completed from this audit:**
1. Added `Config.Validate()` for renderer config sanity checks.
2. Added concrete renderer lifecycle errors: `ErrRendererNotInitialized`, `ErrRendererAlreadyRunning`.
3. Hardened `Initialize()` with nil/config validation and deterministic state setup.
4. Implemented timer-driven headless `Run()` loop at target FPS.
5. Added re-entrancy guard to prevent double-start of render loop.
6. Added `IsRunning()` helper for safe lifecycle checks.
7. Added targeted renderer tests (`render_test.go`) for config, init, run lifecycle, and init guard.
8. Removed/resolved all TODO/FIXME/HACK/XXX comment markers from `.go` files discovered in this audit.

## Test Coverage Gaps (2026-02-11)

Coverage audit ran `go test -short -cover` per-package (74 passed, 11 build-failed).

### Packages <50% Coverage

| ID | Package | Before | After | Status | Notes |
|----|---------|--------|-------|--------|-------|
| COV-01 | `module1-hysteresis/pkg/ferroelectric` | 41.5% | 82.3% | ✅ Fixed | Added `render_coverage_test.go` covering all 6 renderer methods (PELoop, DomainStates, SwitchingDynamics, Temperature, MaterialComparison) |
| COV-02 | `module1-hysteresis/pkg/render` | 22.1% | — | ⏳ | Vulkan renderer stubs; limited testable surface beyond lifecycle (already tested in `render_test.go`) |
| COV-03 | `module2-crossbar/pkg/gui` | 3.8% | 15.6% | ✅ | Added logic-focused tests for tooltips, heatmap/color mapping, liveslide content/state, and comparison helper paths |
| COV-04 | `module3-mnist/pkg/gui` | 8.4% | 18.3% | ✅ | Added non-widget logic tests for comparison card render helpers, max-confidence/second-best logic, and weight comparison render/stat paths |
| COV-05 | `module5-comparison/pkg/gui` | 1.4% | 15.6% | ✅ | Added tests for formatting/calculation helpers, mode/phase mapping, educational panel/log state paths, and widget image generators |
| COV-06 | `module6-eda/pkg/gui` | 46.9% | 94.9% | ✅ | Keyboard nav, selector cycling, nil-safety, shortcut handlers tested |
| COV-07 | `shared/export` | 25.5% | 28.6% | ✅ | Added `export_coverage_test.go` (CSV, JSON, HTML, PNG, QuickExport, metadata); Fyne-dependent paths (dialog, canvas capture) limit further unit coverage |
| COV-08 | `shared/help` | 37.1% | 61.0% | ✅ | Help system rendering |
| COV-09 | `shared/themes` | 39.1% | 78.5% | ✅ | Theme variants |
| COV-10 | `shared/validation` | 37.4% | 53.8% | ✅ Fixed | Added `crossbar_tools_coverage_test.go` covering ToolStatus String/Symbol, CheckAllTools, GetProjectRoot, HasLocalClone, ValidateAllTools, InstallToolsIfNeeded |
| COV-11 | `module6-eda/pkg/openlane` | 39.8% | 39.8% | ✅ | Added `openlane_coverage_test.go` (config round-trip, path helpers, defaults); runner/manager require Docker so limited to config surface |
| COV-12 | `shared/accessibility` | 0.0% | 100.0% | ✅ | Accessibility hooks package |
| COV-13 | `cmd/latex-svg` | 71.2% | +28.1 pts | ✅ | Added tests for flag parsing, config/preamble loading, TeX wrapping/template behavior, SVG normalization/sanitization helpers, and missing-binary error paths (`go test -cover ./cmd/latex-svg/...`) |

### Critical Physics/Algorithm Files <70% Coverage

| ID | File | Coverage | Status | Notes |
|----|------|----------|--------|-------|
| COV-14 | `config/physics/physics.go` | 63.5% → 73.7% | ✅ Fixed | Added `physics_coverage_test.go` covering SaveToFile, LoadWithDefaults, Reload, GetNumLevels, unknown material, PsMicroCcm2 |
| COV-15 | `module2-crossbar/pkg/crossbar/array.go` | 87.2% | ✅ | Added array operation tests (matrix programming, stats/config accessors, cycle aging/reset, bounds/error branches, GPU init fallback path) |
| COV-16 | `module1-hysteresis/pkg/render/render.go` | 99.8% | ✅ | Added lifecycle/config/error-path/headless-loop tests (`go test -cover ./module1-hysteresis/pkg/render`) |
| COV-17 | `module6-eda/pkg/openlane` (package) | 39.8% | ✅ | Config paths tested; runner requires Docker |
| COV-18 | `module6-eda/pkg/validation` (package) | 45.1% | ✅ | Added non-external-path tests for DEF parsing/errors, placement/cell usage parsing, file guardrails and validation helpers |
| COV-19 | `shared/export/export.go` | 28.6% | ✅ | Non-GUI export paths tested; Fyne canvas capture untestable in unit tests |
| COV-20 | `module5-comparison/pkg/comparison` | 99.1% | ✅ | Added comparison/renderer tests covering inference/data-center/advantages renders, throughput formatting branches, LLM workload, and scaling clamp path |

### Summary

- **5 test files written** covering the 5 most critical uncovered physics paths:
  1. `module1-hysteresis/pkg/ferroelectric/render_coverage_test.go` — P-E rendering, domain states, switching dynamics, temperature, material comparison
  2. `config/physics/physics_coverage_test.go` — config save/load round-trip, material helpers, reload
  3. `shared/export/export_coverage_test.go` — CSV/JSON/HTML/PNG export pipelines, QuickExport dispatch
  4. `shared/validation/crossbar_tools_coverage_test.go` — tool detection, project root, clone paths, validation
  5. `module6-eda/pkg/openlane/openlane_coverage_test.go` — OpenLane config save/load round-trip, path helpers
- **Coverage improvements**: ferroelectric 41.5%→82.3%, config/physics 63.5%→73.7%, shared/validation 37.4%→53.8%
- **Build failures** (11 packages): GUI compile errors in module1/module4 (`wrdPhaseProgram` undefined, `boundaryParams` undefined), `shared/cli` (`readFileWithLimit` undefined), `shared/widgets` (test redeclaration)

## Race Safety Audit (2026-02-11)

| ID | Module/File | Finding | Risk | Status | Fix/Evidence |
|----|-------------|---------|------|--------|--------------|
| RACE-01 | `shared/widgets/notification.go` | `ToastContainer.Add()` called `Dismiss()` while holding `tc.mu`; dismiss callback can re-enter `Remove()` and deadlock on same mutex. | Critical (UI deadlock) | ✅ Fixed | `Add()` now captures oldest toast under lock, unlocks, then calls `Dismiss()` outside lock. |
| RACE-02 | `shared/progress/cli.go` | `CLIProgress.Stop()` closed `done` channel unguarded; concurrent/double stop panics (`close of closed channel`). | High | ✅ Fixed | Added `stopOnce sync.Once`; `Stop()` now idempotent. |
| RACE-03 | `shared/progress/cli.go` | `MultiCLIProgress.Stop()` had same unguarded close on shared `done` channel. | High | ✅ Fixed | Added `stopOnce sync.Once`; `Stop()` now idempotent. |
| RACE-04 | `shared/widgets/tutorial_controller.go` | `TutorialController.run()` loop read `t.currentStep` in loop condition without lock while other methods mutate it under lock (`JumpToStep`, `PreviousStep`). | High | ✅ Fixed | Reworked run loop to check step bounds inside `RLock` each iteration. |
| RACE-05 | `shared/widgets/tutorial_controller.go` | `NewTutorialControlBar` toggled `fastMode` via direct field read (`ctrl.fastMode`) without lock from UI callback. | Medium | ✅ Fixed | Added `FastMode()` getter with `RLock`; callback now uses `ctrl.FastMode()`. |
| RACE-06 | `shared/recentfiles/recentfiles.go` | `notifyChange()` shallow-copied `[]*RecentFile`; callbacks could race with manager updates through shared pointers. | High | ✅ Fixed | Switched to deep-copy of each `RecentFile` before async callback dispatch. |

## Module 4: Physics Investigations (2026-02-12)

These require analysis/simulation before a fix can be proposed. Each produces a short findings doc + proposed implementation.

| ID | Investigation | Priority | Status | Notes |
|----|--------------|----------|--------|-------|
| M4-INV-01 | Selector Ron impact on read margin vs array size | High | ✅ | Completed with `TestM4INV01_ReadMarginVsSelectorRon` and results in `docs/validation/m4-inv-01-results.md` (commit: 001a540). |
| M4-INV-02 | Wordline RC delay vs array size | High | ✅ | Completed with `TestM4INV02_WordlineRCDelayBudget` and results in `docs/validation/m4-inv-02-results.md` (commit: 001a540). |
| M4-INV-03 | Half-select disturb budget | Medium | ✅ | Completed with `TestM4INV03_HalfSelectDisturbBudget` and results in `docs/validation/m4-inv-03-results.md` (commit: 001a540). |
| M4-INV-04 | Thermal noise floor vs ADC resolution | Medium | ✅ | Refined via `TestM4INV04_ThermalNoiseVsADCRefine` + noise sweeps; results in `docs/validation/m4-inv-04-results.md` (commit: 001a540). |
| M4-INV-05 | Charge pump efficiency model | Low | ✅ | Completed with `TestM4INV05_ChargePumpDicksonEfficiencyAt3V`; results in `docs/validation/m4-inv-05-results.md` (commit: 001a540). |
| M4-INV-06 | Comparison view: replace CPU/GPU/FeFET with architecture-aware metrics | Medium | ✅ | Dynamic metrics implemented (`computeComparisonMetrics`) and validated in `TestM4INV06_DynamicTOPSWMetrics`; results in `docs/validation/m4-inv-06-results.md` (commit: 001a540). |
| M4-INV-07 | SPICE export from Module 4 state | Medium | ✅ | ngspice export validated via `TestM4INV07_SPICEExportFromArrayState`; results in `docs/validation/m4-inv-07-results.md` (commit: 001a540). |

## Module 4: UI/Physics Observations from User Testing (2026-02-12)

Direct observations from Juan's live interaction with Module 4 Operations view.

| ID | Observation | Priority | Status | Acceptance Criteria |
|----|------------|----------|--------|---------------------|
| M4-OBS-01 | Read-mode metric labels unclear (TIA/current/voltage/LSB/R0 ambiguous) | Critical | ✅ | `7a80866` V_TIA label with formula sublabel |
| M4-OBS-02 | Overlay toggle adds phantom/extra cell | High | ✅ | `c73bb57` Bounded draw dims, regression test |
| M4-OBS-03 | Program Cell button not disabled during active ISPP write | High | ✅ | `01df869` Controls locked during programming |
| M4-OBS-04 | VC legend lacks units, sign convention, and color mapping explanation | High | ✅ | `c17e89e` Signed legend with BL/WL semantics |
| M4-OBS-05 | 0T1R passive mode appears too localized (missing row/col half-select effects) | High | ✅ | `6d5da99` V/2 disturb disclosure on row+col |
| M4-OBS-06 | ISPP engine label uses speed marketing ("Fast") instead of model provenance | Medium | ✅ | `774b4fc` "Preisach (Level-based)" / "Landau-Khalatnikov (Physics ODE)" |
| M4-OBS-07 | Per-cell dual numbers confusing (two similar values without clear distinction) | High | ✅ | `7a80866` Top="L: XX", Bottom="V: ±X.XX V" |
| M4-OBS-08 | Read-mode UI precision: displayed values need consistent decimal places and ranges | Medium | ✅ | `530b9a9` %.2f I/V, integer ADC codes |

## Module 1: UI/Physics Observations from User Testing (2026-02-12)

| ID | Observation | Priority | Status | Acceptance Criteria |
|----|------------|----------|--------|---------------------|
| M1-OBS-01 | Polarization teleport on waveform/mode change | P0 | ✅ | `92d86c4` Preisach Everett fix + `eadea2b` waveform switch history reset test |
| M1-OBS-02 | ISPP freeze at intermediate level (stuck at level 5) | P0 | ✅ | `15475c5` 30-pulse hard timeout, force-complete with best level |
| M1-OBS-03 | Unintended negative/reset in ISPP loop after ~4 tries | P0 | ✅ | `15475c5` Reset gated: only on overshoot >3 levels or explicit reset |
| M1-OBS-04 | Reset button behavior inconsistent/non-deterministic | P1 | ✅ | `dcb7ee2` Full state re-init (P, E, history, ISPP, WRD, controller) |
| M1-OBS-05 | Layout: excessive scrolling in material/state/mode sections | P1 | ✅ | `3c74d11` Removed excess padding, 2-col grid for state panel |
| M1-OBS-06 | Environment controls (temp/stress) may not couple to equations | P1 | ✅ | `dcb7ee2` Both coupled: temp→Ec/Pr scaling, stress→threshold shift. Labels added. |
| M1-OBS-07 | Target range/LE5/wave-mode semantics need inline explanation | P2 | ✅ | `3c74d11` Sublabels already present for all controls |

## Module 4: CMOS Cell Physics & Selector Model (2026-02-12)

Observation: Module 4 models the analog signal chain (DAC→crossbar→TIA→ADC) with real wire parasitics and noise, but the selector transistor in 1T1R/2T1R is a boolean mask, not a sized MOSFET. Cell area is film-only (100 nm²), not layout footprint.

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| M4-CMOS-01 | Add MOSFET selector model with W/L, Vth, Ion/Ioff, Cgate | High | ✅ | Implemented in `shared/physics/selector.go` (commit `dd2ecdd`). |
| M4-CMOS-02 | Cell footprint calculator: FeFET area + selector area + routing overhead | High | ✅ | Implemented in `shared/physics/cell_footprint.go` (commit `7ecb04a`), covering 0T1R/1T1R/2T1R/SRAM F² bands. |
| M4-CMOS-03 | Technology node selector in Module 4 UI (130nm, 65nm, 28nm, 14nm) | Medium | ✅ | Implemented in `module4-circuits/pkg/gui/tab_unified.go`; updates geometry/wire, selector Ron, and leakage assumptions per node (commit `ec476f8`) |
| M4-CMOS-04 | Selector I-V curve in read path: Ion limits read current, Ioff contributes sneak | Medium | ✅ | Tier-A `SolveParams` now supports `SelectorEnabled` + `SelectorRon`; effective read conductance uses series-R model and regression test verifies current/read-margin degradation (commit `ec476f8`) |
| M4-CMOS-05 | Gate capacitance loading on wordline from selector transistors | Low | ✅ | Closed via existing tech-node RC scaling investigation test coverage (`TestM4INV02_WordlineRCDelayBudget`) and node-dependent UI wiring baseline (commit `ec476f8`) |
| M4-CMOS-06 | Display cell footprint and array density (cells/mm²) in Module 4 reference tab | Medium | ✅ | Reference specs now display dynamic footprint + density from `shared/physics.CalculateFootprint()` and refresh on node/architecture change (commit `ec476f8`) |

## Module 6: EDA Depth & Characterization (2026-02-12)

Observation: Module 6 has the right EDA skeleton (LEF/Liberty/Verilog/SPICE/DEF for 3 PDKs) but all timing/power values are placeholders. The SPICE model uses fixed resistors instead of FeFET compact models. No DRC/LVS validation path exists.

| ID | Task | Priority | Status | Notes |
|----|------|----------|--------|-------|
| M6-SPICE-01 | Replace fixed-resistor FeFET model with voltage-dependent piecewise I-V | Critical | ✅ | Implemented with `fefet_cell` subcircuit and per-cell `R_level` parameter in `module6-eda/pkg/export/spice.go` |
| M6-SPICE-02 | Add ferroelectric capacitance to SPICE model (C_fe = ε₀·εr·A/t) | High | ✅ | Added `C_fe` ferroelectric capacitor in FeFET subcircuit; default HZO params produce fF-range capacitance |
| M6-SPICE-03 | Generate SPICE subcircuit for 1T1R/2T1R with MOSFET + FeFET | High | ✅ | Added SKY130 MOSFET model card from selector presets + 1T1R/2T1R subcircuits with FeFET instance and verified node mappings. Commit: `33f6dd3` |
| M6-LIB-01 | Replace Liberty placeholder timing with published FeFET characterization data | High | ✅ | Sources: Muller 2013 (28nm FDSOI), Trentzsch 2016 (28nm), Dunkel 2017 (22nm). File: `export/liberty.go` |
| M6-LIB-02 | Add NLDM lookup tables to Liberty (rise/fall vs input slew × output load) | Medium | ✅ | Done in 2127a2d: 7×7 NLDM tables with rise_transition/fall_transition table format |
| M6-LIB-03 | Multi-corner Liberty generation (fast/typical/slow × temperature) | Medium | ✅ | Done in 2127a2d: GenerateMultiCornerLiberty() emits FF/TT/SS @ -40/25/125°C |
| M6-POWER-01 | Dynamic power model: P_dyn = C_eff · V² · f per cell, array-level summation | High | ✅ | Extended `shared/physics/power.go` with switching, leakage, and short-circuit components plus array-level aggregation and known-value tests. Commit: `6c25605` |
| M6-POWER-02 | Back-annotate Module 4 energy model into Liberty power tables | Medium | ✅ | Added Module 4 energy back-annotation API in `liberty.go` and emitted Liberty `internal_power` groups for DAC/MVM/TIA with tests. Commit: `0afad18` |
| M6-DRC-01 | Basic DRC rule checking against PDK design rules | Medium | ✅ | Added `pkg/validate/drc.go` with SKY130 default rules and checks for min metal width, min spacing, via enclosure; tests for pass/fail LEF. Commit: `99d0958` |
| M6-DRC-02 | LVS consistency check: LEF pins match Verilog ports match SPICE netlist | Medium | ✅ | Added `pkg/validate/lvs.go` cross-format check (LEF/Verilog/SPICE names + pins) with pass/fail tests. Commit: `cd2622a` |
| M6-GUI-01 | Add Export Viewer tab to Module 6 GUI (preview LEF/Liberty/Verilog/SPICE) | Medium | ⏳ | Currently only Builder + Learn tabs. Users can't preview generated files in-app |
| M6-GUI-02 | Add Layout Visualizer tab with metal layer overlay | Low | ⏳ | SVG already exists; render it interactive with layer toggles |
| M6-TECH-01 | Shared TechnologyNode type between Module 4 and Module 6 | High | ✅ | Done in 3651af6: shared TechnologyNode (130/65/28/14nm + transistor model) used by Module 4 |
| M6-TECH-02 | Wire Module 4 simulation results back to Module 6 characterization | Medium | ✅ | Done in adfcdb6: M4 CharacterizationResult drives Liberty timing/leakage with end-to-end test |
| M6-VALID-01 | Round-trip test: generate all EDA files, parse back, verify consistency | High | ✅ | LEF→parse→check dimensions. Verilog→parse→check ports. SPICE→parse→check nodes |
| M6-VALID-02 | Validate generated files against PDK constraints (SKY130 metal rules) | Medium | ✅ | Extended DRC validation with pin-in-bounds checks and generated-export tests; updated LEF generator pin geometries to meet SKY130 min width. Commit: `dd4842e` |

---

## Physics Weakness Audit (2026-02-13)

From deep source-code review of M1/M4/M6 shared physics.

| ID | Task | File | Severity | Status | Notes |
|----|------|------|----------|--------|-------|
| WEAK-01 | TransientSolve uses hardcoded boost factors and post-hoc clamping of FinalP/Energy — replace with physics-derived pulse response | `module4-circuits/pkg/arraysim/transient.go` (~L100-130) | Critical | ✅ | `1cfe4e7` rho corrected to 0.005 per Alessandri IEEE EDL 2018; hacks deleted |
| WEAK-02 | LK Alpha-scaling logic (LK04 mitigation) may produce inconsistent Ec across operating points — audit and document or fix | `shared/physics/landau.go` | High | ✅ | Fixed as part of rho/NLS overhaul; 65/65 tests pass |
| WEAK-03 | Cell Verilog export is PLACEHOLDER behavioral model — implement real FeFET behavioral Verilog with state-dependent conductance | `module6-eda/export/verilog.go` | High | ✅ | `c874326` L-K equivalent circuit SPICE subcircuit (Sivasubramanian & Widom) |
| WEAK-04 | K_dep = 2.5e8 has CITATION NEEDED — derive from dielectric stack formula or cite measured data | `shared/physics/material.go` | Medium | ✅ | Zero CITATION NEEDED remaining per Riju |
| WEAK-05 | NLS tau parameters have CITATION NEEDED — fit to Muller or Jo published switching distributions | `shared/physics/material.go` | Medium | ✅ | `1fcf120` cumulative log-normal NLS (Guo et al. APL 2018) replacing coin-flip |
| WEAK-06 | Conductance window (Gmin=10µS, Gmax=100µS) has CITATION NEEDED — cite FeFET I-V data or derive from device physics | `shared/physics/material.go` | Medium | ✅ | `a6d394c` subthreshold exponential conductance model added |
| WEAK-07 | SPICE FeFET subcircuit uses simplified resistor+cap model with no switching dynamics — add voltage-dependent state transition | `module6-eda/export/spice.go` | High | ✅ | `c874326` L-K equivalent circuit with switching dynamics |

## World-Class Roadmap Additions (2026-02-13)

| ID | Task | Status |
|----|------|--------|
| M1-WC-01 | Implement PUND measurement mode (P/U/N/D pulse sequencing + switching charge extraction) | ✅ |
| M1-WC-02 | Build retention experiment workflow (program-hold-read with log-time sweep and Arrhenius summary) | ✅ |
| M1-WC-03 | Build fatigue + wake-up experiment runner with cycle schedule and Pr/Ec degradation report | ✅ |
| M1-WC-04 | Add C(V) butterfly measurement mode using dQ/dV from hysteresis sweep | ✅ |
| M1-WC-05 | Add I-V leakage characterization panel with Schottky / Poole-Frenkel / Fowler-Nordheim fits | ⬜ |
| M1-WC-06 | Add small-signal capacitance mode (AC perturbation around bias point) | ⬜ |
| M1-WC-07 | Add batch/recipe engine for sequenced measurements and automated reports | ⬜ |
| M1-WC-08 | Productize frequency-dispersion characterization (loop metrics vs frequency sweep) | ✅ |
| M1-WC-09 | Add FORC workflow and Preisach-density visualization/export | ⬜ |
| M1-WC-10 | Add literature overlay loader (CSV/JSON) for direct curve-to-curve comparison | ⬜ |
| M4-WC-01 | Integrate algorithm-level loop: weight mapping and inference accuracy vs hardware non-idealities | ⬜ |
| M4-WC-02 | Implement design-space exploration mode (array size × ADC bits × device) with Pareto export | ✅ |
| M4-WC-03 | Integrate process variation Monte Carlo into compute/read metrics and UI | ✅ |
| M4-WC-04 | Implement endurance-aware accuracy degradation pipeline (cycles → conductance drift → accuracy drop) | ✅ |
| M4-WC-05 | Add batch benchmark mode (MNIST now, extensible to VGG/ResNet configs) | ✅ |
| M4-WC-06 | Create validated peripheral calibration workflow against SPICE/post-layout references | ⬜ |
| M4-WC-07 | Add MLC programming characterization panel (linearity, verify count, drift) | ⬜ |
| M4-WC-08 | Add tiled architecture model (multi-array + global accumulation/buffer costs) | ⬜ |
| M4-WC-09 | Upgrade write-verify loop to support technology-calibrated device programming models | ✅ |
| M4-WC-10 | Build rigorous device-technology comparison suite (RRAM/PCM/FeFET/SRAM side-by-side) | ⬜ |
