# ADR 0001: Level Calibration Workflow — UI-Neutral Service Architecture


Update: the authoritative Level Calibration Engine now lives in `shared/physics/level_calibration.go` so downstream modules can reuse it without importing Module 1 internals. The older `module1-hysteresis/pkg/algo/CalibrationManager` remains an adaptive write-controller calibration helper, not the Level Calibration Workflow engine.

## Context

- `module1-hysteresis/pkg/algo/CalibrationManager` is write-controller adaptive tuning state and should not be treated as the user-facing Level Calibration Engine.

## Decision

- **Level Calibration Engine** lives in `shared/physics/` as a UI-neutral service with a clean `Calibrate(Inputs) -> Summary` interface.
- **Level Calibration State** (not-calibrated/stale/fresh) is derived from the engine output, not from UI timestamps.
- **Level Calibration Export** is an explicit user action that writes artifacts (JSON/CSV) from the engine summary; it does not depend on the UI framework.

## Considered Options

3. **Extract UI-neutral service** — chosen. The computation is straightforward math; the UI layer is thin and replaceable.

## Consequences

- `shared/physics/` now carries the Level Calibration Engine contract (`shared/physics/level_calibration.go`), which downstream modules (Module 4 ISPP, Module 6 EDA export) can also consume.
- New user-facing Level Calibration Workflow computation should be added to `shared/physics/`, not to GUI code. Adaptive write-controller tuning may remain in Module 1 internals when it is not a cross-module workflow.
