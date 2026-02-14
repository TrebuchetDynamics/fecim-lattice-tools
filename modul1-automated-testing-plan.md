# Modul1 Automated Testing Plan

## Purpose
Strengthen headless automated validation for Module 1 (`module1-hysteresis`) so that write/verify physics, readback mapping, and exported data remain numerically stable and regression-safe.

Note: filename intentionally follows requested spelling `modul1`.

## Scope (Module 1 View)
- `WRITE`: ISPP pulse progression and convergence behavior.
- `READ/VERIFY`: field/polarization/conductance-to-level consistency checks.
- `COMPUTE-SUPPORT`: data outputs used by downstream modules remain stable and finite.

## Source Docs Used
- `docs/testing/TEST_GUIDE.md`
- `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`
- `docs/development/CI.md`
- `docs/development/HEADLESS.md`
- `docs/development/evidence/G08-mid-stability-evidence-2026-02-11.md`
- `validation/testdata/ispp_regression/preisach_wrd_ispp_regression.json`
- `validation/testdata/ispp_regression/lk_wrd_ispp_regression.json`

## Goals
- Catch NaN/Inf and unit mistakes early.
- Guarantee deterministic headless behavior under fixed seeds.
- Validate convergence and overshoot handling across materials/engines.
- Emit machine-readable summaries for trend/regression analysis.

## Research-Grade Target
- Move from "headless smoke/regression only" to publishable, traceable validation evidence.
- Report uncertainty and bounded behavior per engine/model class.
- Keep dual acceptance profiles where physics model maturity differs (Preisach vs LK baseline).

## Material-Selected Awareness (Required)
- Every regression run must set `FECIM_MATERIAL` explicitly.
- Test IDs and artifacts must include material:
  - example key: `m1/lk/literature_superlattice/targets_lo_mid_hi/seed1`
- Metrics and thresholds must be material-normalized where possible:
  - field normalized by `Ec`
  - polarization normalized by `Ps`
  - level mapping validated against material-specific conductance bounds.
- Per-material verdict is mandatory; aggregate pass cannot mask a single failing material.
- Each artifact must include a material parameter snapshot used by the run:
  - at minimum: `Ec`, `Ps`, `Pr`, thickness, `Gmin`, `Gmax`, `TargetRangeFrac`.

## Cross-Module Physics Criteria Alignment
- Loop regression target: RMS(E), RMS(P) vs golden baseline <= 2% full-scale.
- Material parameter checks: within documented bounds with 10% engineering tolerance unless a tighter test exists.
- ISPP level-hit target for strict profile: level error <= ±1.

## Non-Goals
- GUI layout rendering checks.
- Foundry/device-level silicon signoff.

## Fully Headless Requirement (Mandatory)
- Required lanes must run without display services:
  - `DISPLAY` unset
  - `WAYLAND_DISPLAY` unset
  - no `xvfb-run`
- Required acceptance must come from headless mode tests and controller tests only.
- GUI visual/lifecycle tests are optional and non-gating for research validation.

## Existing Coverage Baseline (Keep + Extend)

### Current Strong Coverage
- `module1-hysteresis/pkg/controller/headless_regression_test.go`
  - emits JSON summaries
  - includes pulse/overshoot budgets
- `cmd/fecim-lattice-tools/mode_lk_headless_ispp_5targets_test.go`
  - validates finite outputs and unit-consistent columns in CSV logs
- `cmd/fecim-lattice-tools/mode_lk_ispp_test.go`
  - validates ISPP rows and phase emission
- `cmd/fecim-lattice-tools/mode_engine_matrix_test.go`
  - engine/material matrix, no-NaN/no-crash checks
- `cmd/fecim-lattice-tools/mode_lk_ispp_convergence_20targets_test.go`
  - deterministic stress convergence signal checks

### Current Gap
- No explicit unified acceptance model that separates strict target-lock behavior from bounded-completion behavior by engine maturity.

## Test Layers

### 1) Engine-Level Headless Regression
Target: `cmd/fecim-lattice-tools` mode tests + `module1-hysteresis/pkg/controller`

- Run deterministic headless modes for:
  - Preisach
  - LK (Landau-Khalatnikov)
- Assert:
  - no NaN/Inf fields
  - unit consistency (`V/m` vs `MV/cm` columns)
  - valid coercive/polarization ranges
  - expected ISPP phase transitions

### 2) Controller/Algorithm Invariants
Target: `module1-hysteresis/pkg/controller`, `module1-hysteresis/pkg/algo`

- ISPP stepping monotonicity by direction.
- Overshoot reset behavior.
- Pulse budget handling and timeout behavior.
- Deterministic target selection by seed.

### 3) Data Artifact and Baseline Comparison
Output to:
- primary: `output/regression/module1/`
- compatibility with existing script/output:
  - `FECIM_REGRESSION_JSON_DIR` (currently used by controller regression tests)
  - `output/regression/` default from `scripts/run_headless_ispp_regressions.sh`

Per run store:
- summary JSON (pass/fail, convergence stats, attempts, overshoots)
- source metadata (engine, material, seed, env vars)
- extracted CSV-derived metrics (min/max/mean error bands)

### 4) Research-Grade Statistical Validation
- Multi-run seeded sweeps for each material/engine profile.
- Report:
  - mean, std, P95 level error
  - pulse count distribution
  - overshoot/retry distribution
- Archive full run manifest (seed list + env vars + commit hash).

## Required Matrix

### Engines
- `preisach`
- `lk`

### Materials
- `fecim_hzo`
- `literature_superlattice`
- `default_hzo` (or available default set)

### Material Gate Policy
- PR gate minimum:
  - `fecim_hzo`, `literature_superlattice`, `default_hzo`
- Nightly/release gate:
  - `fecim_hzo`
  - `fecim_hzo_target`
  - `default_hzo`
  - `literature_superlattice`
  - `cryogenic_hzo`
  - `hzo_standard_32`
  - `hzo_ftj_140`
  - `hzo_custom_14`
  - `alscn`

### Target Sets
- fixed set: `lo,mid,hi`
- randomized deterministic set with seed
- stress set (e.g., 20 targets)

### Runtime Lanes
- Fast CI lane:
  - `FECIM_HEADLESS_FAST=1`
  - small target sets
- Extended/nightly lane:
  - more targets
  - tighter diagnostics

### Headless Environment Variables To Sweep
- `FECIM_MATERIAL`
- `FECIM_RANGE_FRAC`
- `FECIM_ISPP_STEPS_PER_PULSE`
- `FECIM_HEADLESS_FAST`
- `FECIM_ISPP_TARGETS`
- `FECIM_ISPP_TARGET_SEED`
- `FECIM_ISPP_TARGET_LEVELS`
- `FECIM_ISPP_MAX_PULSES`
- `FECIM_HEADLESS_ALLOW_TIMEOUT`

## Implementation Plan

## Phase 1: Consolidate Existing Headless Coverage
- Standardize around existing tests in:
  - `cmd/fecim-lattice-tools/mode*_test.go`
  - `module1-hysteresis/pkg/controller/headless_regression_test.go`
- Add missing assertions for:
  - deterministic target order
  - explicit convergence window reporting
  - CSV schema checks for key columns used by downstream analysis

## Phase 2: Artifact-Driven Regression
- Extend runner script behavior to emit normalized JSON summaries for both engines.
- Add automatic comparison against previous baseline in `output/regression/module1/`.
- Import and compare against frozen references in:
  - `validation/testdata/ispp_regression/*.json`

## Phase 3: CI Integration
- Fast lane in CI:
  - run deterministic headless suite on both engines.
- Optional extended lane:
  - enabled by env flag (`FECIM_M1_EXTENDED=1`).

## Phase 4: Research-Grade Acceptance and Reporting
- Add dual acceptance profiles:
  - strict target-lock profile (Preisach-like)
  - bounded-completion profile (LK single-domain baseline until upgraded)
- Emit trend report with confidence intervals and per-material verdicts.
- Add release gate requiring no statistically significant regression vs baseline.

## Suggested Runner Commands
- Existing:
  - `scripts/run_headless_ispp_regressions.sh`
- Add/extend:
  - `env -u DISPLAY -u WAYLAND_DISPLAY go test ./cmd/fecim-lattice-tools -run Headless.*ISPP -count=1`
  - `env -u DISPLAY -u WAYLAND_DISPLAY go test ./module1-hysteresis/pkg/controller -run HeadlessRegression_WRD_ISPP -count=1`
  - `env -u DISPLAY -u WAYLAND_DISPLAY GO_TEST_TIMEOUT=10m go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 10m ./cmd/fecim-lattice-tools/...`

## Acceptance Profiles

### Profile A: Strict Target-Lock (Research Baseline)
- Used for Preisach and any model proven target-lock capable.
- Criteria:
  - level error `<= ±1` level (cross-module physics criteria)
  - convergence required
  - pulse and overshoot budgets respected
  - current controller regression defaults:
    - pulses `<= 30` per target (Preisach regression suite)

### Profile B: Bounded Completion (Current LK Baseline)
- Used for current LK single-domain baseline while stabilization work is pending.
- Criteria:
  - must reach terminal done state
  - must satisfy pulse/overshoot/retry bounds
  - current controller regression defaults:
    - pulses `<= 80` per target
    - overshoots `<= 20` per target
  - must remain finite and unit-consistent
  - level error tracked and reported, not hidden

## Engine × Material Profile Map (Required)
- Preisach:
  - default profile is `A` for all gated materials unless evidence says otherwise.
- LK:
  - default profile is `B` for all gated materials until promoted by evidence.
  - promotion to `A` requires recorded evidence for that material with:
    - repeated deterministic runs meeting strict level-hit and budget criteria.
- The active profile map must be versioned with artifacts for each release.

## Pass/Fail Criteria
- Zero NaN/Inf tokens in logs and parsed metrics.
- Unit-consistency checks pass.
- Convergence/partial behavior is explicit and bounded (no silent drift).
- Deterministic seeds produce stable results within tolerance windows.
- Regression comparator reports no statistically significant degradation.
- Acceptance profile for each engine/material is declared and enforced.
- Material coverage is complete for the active gate (PR or nightly).
- Mandatory lanes pass fully headless (no display env and no Xvfb).

## Deliverables
- Updated/added headless tests for Module 1 invariants.
- Extended regression JSON outputs.
- CI hooks for fast and extended lanes.
- Short docs update in `docs/testing/TEST_GUIDE.md` referencing module1 regression workflow.
- Research report artifact containing:
  - profile verdicts
  - trend metrics
  - uncertainty summary
