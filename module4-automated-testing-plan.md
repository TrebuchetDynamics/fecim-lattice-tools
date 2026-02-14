# Module 4 Automated Testing Plan

## Purpose
Build a reliable, headless automated test pipeline for Module 4 (`module4-circuits`) that validates `READ`, `WRITE`, and `COMPUTE` behavior from a data-integrity perspective, not just UI behavior.

## Source Docs Used
- `docs/circuits/module4-flow-audit.md`
- `docs/circuits/module4-write-path-proof.md`
- `docs/circuits/signal-flow.md`
- `docs/peripheral-circuits/ARRAY_SIMULATION_FIDELITY.md`
- `docs/development/PHYSICS_ACCEPTANCE_CRITERIA.md`
- `docs/development/CI.md`
- `docs/development/HEADLESS.md`
- `docs/testing/TEST_GUIDE.md`

## Primary Problem To Catch
- Non-physical global or row/column broadcast updates during single-cell write.
- Hidden state mutation during `READ` or `COMPUTE`.
- Drift, non-determinism, and invalid (NaN/Inf) analog/digital outputs.

## Goals
- Detect incorrect data movement in `write -> read -> compute` flows.
- Enforce operation invariants in headless CI.
- Produce machine-readable artifacts for regression comparisons between commits.

## Research-Grade Target
- Move from "internally consistent simulation" to "physically defensible model validation."
- Demonstrate bounded agreement with higher-fidelity reference methods.
- Quantify uncertainty and confidence, not only pass/fail booleans.

## Material-Selected Awareness (Required)
- Every test case must bind an explicit material selection. No implicit/default material runs in regression gates.
- Material choice must be part of the test ID and artifact key:
  - example key: `m4/0T1R/TierA/8x8/fecim_hzo/seed1`
- All voltage targets and safety checks must be material-derived, not hard-coded:
  - use material-aware ranges (`readRange`, `writeRange`) from the active `DeviceState`.
  - use material `Ec`, `Ps`, thickness, and conductance bounds when computing normalized metrics.
- Pass/fail verdicts are per material and then aggregated; no single global verdict may hide one failing material.
- Regression artifacts must include a material snapshot:
  - material name/id
  - key physical params used in that run (at minimum: `Ec`, `Ps`, `Pr`, thickness, `Gmin`, `Gmax`, quant levels).

## Non-Goals
- Visual/layout regression (handled by screenshot/UI crawler tests).
- Full tapeout signoff replacement.

## Fully Headless Requirement (Mandatory)
- Required gates must run with no display stack:
  - `DISPLAY` unset
  - `WAYLAND_DISPLAY` unset
  - no `xvfb-run`
- Required gates must not depend on GUI rendering, window creation, or screenshot capture.
- Any Fyne/UI harness tests are optional/non-gating and tracked separately.

## Doc-Grounded Operational Facts
- Active write dispatch is through shared physics (`StartISPP`/`ISPPIterate` and LK write controller), with legacy `writeReadVerifyLoop` non-dispatched.
- `READ` and `COMPUTE` should not update `arrayWeights`; they consume weights to produce currents/ADC codes.
- Default coupled solve path is Tier-A, with Ideal fallback.
- 0T1R write behavior expects V/2 half-select exposure on selected row and selected column only.
- Typical read region is low-voltage non-destructive (`~0.1-0.3V`) and write path uses higher voltage range.

## GUI/Headless Physics Parity (Mandatory)
- Headless validation must call the same physics path as GUI operations.
- No separate "headless-only" physics equations for `READ`, `WRITE`, or `COMPUTE`.
- Allowed difference:
  - orchestration layer (CLI/test harness) may differ.
- Not allowed difference:
  - solver path, ISPP logic, conductance mapping, or coupling equations.
- Enforcement:
  - add parity tests that run equivalent scenarios through GUI action path and headless harness path, then compare physics outputs (within tolerance).
  - keep parity checklist tied to `docs/circuits/module4-write-path-proof.md`.

## Existing Coverage Baseline (Keep + Extend)

### Operation Invariants Already Present
- `module4-circuits/pkg/arraysim/current_validation_test.go`
  - `READ` and `COMPUTE` do not mutate state.
- `module4-circuits/pkg/gui/device_state_read_coupling_test.go`
  - signed read VI behavior, end-to-end read chain checks.
- `module4-circuits/pkg/gui/device_state_halfselect_dac_arraysim_test.go`
  - DAC-quantized write voltage mapped to target and half-select Vcell values.
- `module4-circuits/pkg/gui/device_state_ispp_coupled_write_test.go`
  - coupled voltage monotonicity and IR-drop-bounded update behavior.
- `module4-circuits/pkg/gui/tab_unified_halfselect_residue_test.go`
  - half-select residue accumulation and row/column disturb pattern.
- `module4-circuits/pkg/gui/unified_buttons_test.go`
  - headless action-flow coverage for READ/WRITE/COMPUTE controls.

### Gaps To Close
- Explicit anti-broadcast guard on single-cell write updates.
- Deterministic end-to-end `WRITE -> READ -> COMPUTE` data-validation artifact runs.
- CI-friendly regression summary output specific to Module 4.
- Cross-validation versus higher-fidelity external references (ngspice/Xyce netlists).
- Statistical coverage and uncertainty reporting across corners.
- Explicit GUI-vs-headless physics parity tests and trace comparison artifacts.

## Validation Pyramid (Research-Grade)

### Tier T0: Invariant and Unit Correctness
- Fast deterministic invariants in `arraysim` and `gui` packages.
- Enforced in normal PR CI.

### Tier T1: Deterministic Workflow Reproducibility
- Headless operation sequences with fixed seeds and artifact snapshots.
- Enforced in normal PR CI.

### Tier T2: Cross-Model Validation
- Compare Module 4 behavioral outputs against reference solves:
  - internal dense nodal baseline
  - external SPICE-derived fixtures (ngspice/Xyce)
- Run daily/nightly.

### Tier T3: Statistical and Corner Robustness
- Monte Carlo and corner sweeps:
  - architecture, coupling tier, material, temperature proxy, array size, load.
- Run nightly/release-gate.

## External Ground Truth and Calibration Track
- Use `docs/opensource-tools/circuit-simulation-tools.md` workflow to build reference fixtures.
- Keep fixture set versioned in repo (small deterministic netlists + expected outputs).
- For each fixture record:
  - source simulator + version
  - model deck hash
  - expected I/V/code outputs
  - tolerance band and rationale
- Acceptance target from Module 4 improvement roadmap:
  - work toward `<5%` error for nominated reference fixtures.

## Test Layers

### 1) Core Deterministic Unit/Integration (Package-Level)
Target: `module4-circuits/pkg/gui`, `module4-circuits/pkg/arraysim`

- `READ` invariants:
  - `arrayWeights` unchanged before/after read.
  - Selected path signals (`current`, `tia`, `adc`) are finite.
  - For fixed setup, higher level yields non-decreasing read code/current.

- `WRITE` invariants:
  - Target cell moves toward target level within pulse budget.
  - 0T1R: non-target disturbance only allowed on selected row/column half-select set.
  - 1T1R/2T1R: non-target change bounded to near-zero threshold.
  - Fail fast if many non-target cells jump by large `Δlevel` (broadcast guard).

- `COMPUTE` invariants:
  - `arrayWeights` unchanged by compute.
  - Output matches expected dot-product behavior (Ideal mode strict, Tier-A bounded error).
  - Results deterministic under fixed seed/input.

### 2) Optional Integration Harness (Non-Gating)
Target: `module4-circuits/pkg/gui` (`test.NewApp` harness when needed)

- Execute actual action handlers:
  - `onUnifiedProgram()`
  - `onUnifiedRead()`
  - `onUnifiedCompute()`
- Validate operation sequence:
  - `WRITE -> READ -> COMPUTE`
  - `READ -> COMPUTE -> READ`
  - repeated cycles (stability)
- Note: keep this lane out of mandatory fully-headless gates.

### 3) Regression Artifact Emission
Write JSON summaries under:
- `output/regression/module4/`

Each run emits:
- metadata (commit, date, seed, architecture, coupling tier, array size)
- changed cell map (`before` vs `after`)
- target error metrics
- max non-target `Δlevel`
- compute residual metrics

## Quantitative Acceptance Rules

### Rules from Cross-Module Physics Criteria
- Read chain correctness and sign behavior must pass exact fixture checks.
- Coupling behavior: default read path remains covered under Tier-A+ tests.

### Module 4 Regression Thresholds (Planned)
- `READ` and `COMPUTE` weight mutation count must be `0`.
- Single-cell `WRITE` target error:
  - strict profile: `|final-target| <= 1 level`
  - bounded profile for difficult corners: explicit partial result with tracked error budget.
- Broadcast guard:
  - if >`K` non-target cells change by >`L` levels in one write iteration, fail.
  - initial defaults: `K=3`, `L=3`, then calibrate from baseline data.
- Determinism:
  - same seed/config run pair must match artifact metrics within fixed tolerance.
- Cross-model agreement:
  - for fixture set, behavioral vs reference residual must remain inside per-fixture bounds.

## Design-of-Experiments Coverage
- Axes:
  - architecture (`0T1R`, `1T1R`, `2T1R`)
  - coupling (`Ideal`, `Tier-A`, `Tier-B`)
  - array size (`2x2`, `8x8`, `16x16`, `32x32`)
  - target levels (`low`, `mid`, `high`, randomized set)
  - input vectors (sparse, dense, signed pattern)
  - material preset (`fecim_hzo`, `literature_superlattice`)
- Required coverage rule:
  - every release candidate must include at least one test point for every axis value.
  - nightly must execute the full matrix or a documented sampled DOE plan.

## Material Sweep Set (Gate)
- Required PR material sweep (fast):
  - `fecim_hzo`
  - `literature_superlattice`
  - `default_hzo`
- Required nightly/release material sweep (extended):
  - `fecim_hzo`
  - `fecim_hzo_target`
  - `default_hzo`
  - `literature_superlattice`
  - `cryogenic_hzo`
  - `hzo_standard_32`
  - `hzo_ftj_140`
  - `hzo_custom_14`
  - `alscn`

## Required Test Matrix

### Architectures
- `0T1R`
- `1T1R`
- `2T1R`

### Coupling
- `Ideal` (strict math baseline)
- `Tier-A` (default coupled path)
- `Tier-B` (optional nightly/extended)

### Sizes
- Fast lane: `2x2`, `8x8`
- Extended lane: `16x16`, `32x32`

### Materials/Quantization
- Material is mandatory axis in all matrices.
- PR gate minimum:
  - `fecim_hzo`, `literature_superlattice`, `default_hzo`
- Nightly/release:
  - full material sweep set above
- Quantization:
  - `quantLevels=30` baseline
  - optional sensitivity check at neighboring level counts if enabled

## Suggested Test Additions by File
- `module4-circuits/pkg/gui/headless_rw_compute_regression_test.go`
  - deterministic workflow tests
  - anti-broadcast assertions
  - regression artifact write
- `module4-circuits/pkg/gui/headless_gui_physics_parity_test.go`
  - run same scenario through GUI dispatch and headless harness
  - compare per-step outputs (`effectiveCellVoltage`, `rowCurrent`, `rowLevel`, target level trajectory)
- `module4-circuits/pkg/gui/headless_rw_compute_artifact.go` (or test-local helpers)
  - snapshot diff + metrics serializer
- `module4-circuits/pkg/arraysim/current_validation_test.go`
  - extend table-driven operation invariants for extra corner cases

## Implementation Plan

## Phase 1: Guardrails (Immediate)
- Add `headless_rw_compute_regression_test.go` in `module4-circuits/pkg/gui`.
- Add broadcast/locality guard tests for write.
- Add immutable-state tests for read/compute.
- Add first GUI/headless parity test for single-cell write and single-cell read.

## Phase 2: Workflow + Artifacts
- Add deterministic sequence tests and JSON artifact output.
- Add helper utilities to compare snapshots and summarize deltas.
- Add small-array analytic cross-checks (Ideal vs expected dot-product).

## Phase 3: CI Integration
- Add script: `scripts/run_headless_module4_regressions.sh`.
- Fast CI command:
  - `go test -tags=ci -count=1 -shuffle=off -trimpath ./module4-circuits/pkg/gui -run HeadlessRWCompute`
- Extended nightly matrix via env flag:
  - `FECIM_M4_EXTENDED=1`

## Phase 4: Research-Grade Cross-Validation
- Add external reference fixture runner:
  - `scripts/run_module4_reference_validation.sh`
- Ingest SPICE/solver outputs and compute residual reports.
- Track trend metrics (mean error, P95 error, worst-case error) in JSON artifacts.
- Add release gate for cross-model agreement.

## Suggested Script
`scripts/run_headless_module4_regressions.sh` should:
- create `output/regression/module4/<timestamp>/`
- run fast suite (headless-only)
- optionally run extended suite
- print paths to JSON summaries

## Execution Commands
- Fast local:
  - `env -u DISPLAY -u WAYLAND_DISPLAY go test ./module4-circuits/pkg/arraysim ./module4-circuits/pkg/gui -run 'CurrentValidation|ReadCoupling|HeadlessRWCompute' -count=1`
- CI-like deterministic:
  - `env -u DISPLAY -u WAYLAND_DISPLAY GO_TEST_TIMEOUT=10m go test -tags=ci -count=1 -shuffle=off -trimpath -timeout 10m ./module4-circuits/...`
- Targeted race lane (existing CI pattern):
  - `env -u DISPLAY -u WAYLAND_DISPLAY GO_TEST_RACE_TIMEOUT=20m make test-race-ci`

## Pass/Fail Criteria
- No NaN/Inf values in captured metrics.
- `READ` and `COMPUTE` do not mutate weights.
- `WRITE` target reaches tolerance or exits with explicit bounded partial result.
- Non-target change constraints respected per architecture.
- Deterministic runs produce stable metrics within defined tolerance.
- Cross-model fixture error remains within agreed research bounds.
- DOE coverage completeness is reported and meets release target.
- Material-specific pass map is complete (no missing material verdicts).
- Mandatory lanes pass fully headless (no display env and no Xvfb).
- GUI and headless parity checks pass for required scenarios.

## Risks and Mitigations
- Risk: flaky timing in GUI-driven headless tests.
  - Mitigation: keep invariant tests at `DeviceState` level; use workflow harness only for integration smoke.
- Risk: thresholds too strict for coupled modes.
  - Mitigation: split strict (Ideal) vs bounded (Tier-A/B) profiles and baseline on saved artifacts.
- Risk: conflating visual regressions with data regressions.
  - Mitigation: keep screenshot/Xvfb suites separate from this plan.

## Deliverables
- New headless regression test file(s) in `module4-circuits/pkg/gui`.
- Optional helper in `module4-circuits/pkg/arraysim` for analytic checks.
- Runner script in `scripts/`.
- Regression output directory and README note in testing docs.
- Traceability appendix mapping each invariant to at least one test function.
- Versioned reference-fixture pack and comparison reports for research-grade validation.
