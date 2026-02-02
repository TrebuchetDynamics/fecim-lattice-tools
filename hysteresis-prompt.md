Role

- You are an expert software engineer and ferroelectrics scientist.
- Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
- If an ambiguity remains, choose the most reasonable default and proceed; document the choice.
- Keep scope tight: only change files required to satisfy the objectives.
- Default to **headless-only work** unless a GUI change is required for correctness of WRD/ISPP.

Objective

- Make the **Write/Read ISPP demo** hit its target levels reliably (no convergence to Ec=0, no infinite loops).
- Ensure the **Frankestein equation** in `docs/hysteresis/hysteresis-gemini.md` is correctly understood and implemented
  (terms, signs, units, and effective viscosity).
- Keep calibration autonomous during runtime so WRD converges quickly without manual intervention.
- Update docs only when behavior or equations change.

Primary Focus (ranked)

1) WRD/ISPP target accuracy (highest priority)
- Read/Write demo must **hit targets** with strict equality (level match).
- No "stuck at E=0" convergence or endless binary search loops.
- Direction logic must use **current vs target** (not stale initial state).
- Overshoot reset only when overshoot truly occurs; no unnecessary saturations between targets.

2) Frankestein equation fidelity
- Implement exactly the unified L-K + depolarization + series-resistance formulation
  from `docs/hysteresis/hysteresis-gemini.md`.
- Verify all terms, signs, and units in logs.

3) Autonomous calibration
- Runtime recalibration should trigger only when convergence is poor and run between targets.
- Calibrations must persist and update calibration manager state coherently.

4) Documentation sync
- Keep `docs/hysteresis/hysteresis.demo.md` aligned with WRD/ISPP behavior.
- Update `docs/hysteresis/hysteresis-gemini.md` only if equation handling changes.

Tasks

1) Frankestein equation (no missing terms)
- Verify: `dP/dt = (E_applied - k_dep*P - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + xi) / rho_eff`.
- Ensure `rho_eff = rho + (R_series * A / d)` only when `UseEffectiveViscosity=true`.
- Confirm `E_eff = E_applied - k_dep*P` is what the solver actually uses.
- Log: `E_applied`, `E_dep` or `k_dep*P`, `E_eff`, `dG_dP`, `rho_eff`, `Alpha`, `Beta`, `Gamma`, `K_dep`.

2) WRD/ISPP target hit guarantee
- Fix direction inference for `target == initial` (use current vs target).
- If `currentLevel == targetLevel`, **exit immediately** with success (no pulses).
- Prevent binary search from collapsing to `VMax=0` when the direction is wrong.
- Keep pre-biasing (+/-Ec) but avoid full saturation unless overshoot is detected.
- Ensure retry logic does not spin indefinitely; failures should be explicit and rare.

3) Autonomous recalibration
- Trigger on repeated overshoots or too many pulses.
- Run recalibration **between targets** to avoid corrupting active state.
- Persist calibration file and sync into `CalibrationManager`.

4) Docs
- Update WRD/ISPP sequencing and calibration behavior in `docs/hysteresis/hysteresis.demo.md`.
- If equation handling changes, update `docs/hysteresis/hysteresis-gemini.md` accordingly.

Validation

- Headless physics: `./launch.sh --logger --verbosity debug --mode hysteresis`.
  - Use logs to confirm Frankestein equation terms appear and match signs/units.
- WRD demo: use the **latest WRD log** to verify target hits.
  - Evidence must include "TARGET HIT" lines and no "Unexpected state ... VMax=0" loops.

Frankestein Equation Checklist (must satisfy each run)

- Uses: `E_eff = E_applied - k_dep*P`.
- Uses: `dP/dt = (E_eff - (2*alpha*P + 4*beta*P^3 + 6*gamma*P^5) + xi) / rho_eff`.
- Uses: `rho_eff = rho + (R_series * A / d)` only if enabled.
- Logs show all terms at debug verbosity.

WRD/ISPP Correctness Checklist

- Target hit with strict equality for each WRD cycle.
- If `current == target`, success without pulses.
- No convergence to `E~0` caused by wrong direction inference.
- Overshoot reset only on true overshoot.
- No forced saturation between targets unless overshoot recovery requires it.
- Auto-recalibration occurs between targets and is logged.

Regression Guardrails

- If WRD success rate drops or failures appear, treat as regression and fix immediately.
- If binary search collapses to zero or loops > MaxRetries, fix direction/bounds logic.
- Keep a **baseline** with latest WRD log path + key success/failure stats.

Execution Rules (Autonomous)

- Always inspect the newest WRD log file under `logs/`.
- Prefer minimal, targeted changes; avoid unrelated files.
- If validation fails, report exact error output and last command run.
- GUI changes are allowed only to fix WRD/ISPP correctness.

Deliverable

- Concise report:
  - Frankestein equation verification (what terms/logs confirmed).
  - WRD/ISPP target-hit evidence (log lines).
  - Documentation updates (file paths + summary).
  - Gaps/issues and next iteration target.
- Include validation command and log path.

Baseline (update each run)

- Latest WRD log path:
- <local-path>
- WRD status:
  - target=15 stalled, VMax collapsed to 0 (needs fix)
