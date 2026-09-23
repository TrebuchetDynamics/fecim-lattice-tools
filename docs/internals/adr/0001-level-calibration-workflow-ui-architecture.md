# ADR 0001: Level Calibration Workflow — UI-Neutral Service Architecture

Update: `shared/physics/level_calibration.go` exposes a UI-neutral uniform field-spacing estimate. It does not currently simulate LK switching; callers must not present it as a physical calibration. The older `module1-hysteresis/pkg/algo/CalibrationManager` remains an adaptive write-controller calibration helper.

## Context

- `module1-hysteresis/pkg/algo/CalibrationManager` is write-controller adaptive tuning state and should not be treated as the user-facing level-field estimate.

## Decision

- The **uniform field-spacing estimate** lives in `shared/physics/` so callers share one clearly labeled heuristic; it is not an LK simulation.
- **Level Calibration State** (not-calibrated/stale/fresh) is derived from the estimate output, not from UI timestamps.
- **Level Calibration Export** is an explicit user action that writes artifacts (JSON/CSV) from the estimate summary; it does not depend on the UI framework.

## Considered Options

3. **Extract UI-neutral service** — chosen. The computation is straightforward math; the UI layer is thin and replaceable.

## Consequences

- `shared/physics/` carries the uniform field-spacing estimate (`shared/physics/level_calibration.go`). Downstream modules may reuse the estimate, but must preserve its heuristic label; a physical calibration requires a separately validated LK-derived implementation.
- New user-facing Level Calibration Workflow computation should be added to `shared/physics/`, not to GUI code. Adaptive write-controller tuning may remain in Module 1 internals when it is not a cross-module workflow.
