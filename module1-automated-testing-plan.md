# Module 1 Automated Testing Plan (Execution-Ready, Enforceable)

Scope: `module1-hysteresis` physics + controller validation in headless CI.

## 1) Objective and Operating Rules

**Objective:** falsify (not just regress) Module 1 behavior against DOI-backed observables, with deterministic artifacts and explicit pass/fail thresholds.

**Hard rules**
- Required lanes are headless only (`DISPLAY` and `WAYLAND_DISPLAY` unset).
- Every required test emits machine-readable artifacts.
- No aggregate pass if any material/dataset fails.
- Commands and runtime budgets in this plan are binding for CI gates.

## 2) Phased Delivery (P0/P1/P2)

## P0 — CI Safety Baseline (must exist before any claim)

Deliverables
- Deterministic build/test baseline for Module 1 lanes.
- Artifact emission wired into required tests.
- Material-explicit execution matrix (no implicit defaults).

Acceptance
- Gate commands (Section 3) run green in PR and nightly.
- Artifact schema (Section 5) validates for every required run.

## P1 — Physics Falsification Core (primary scientific gate)

Deliverables
- DOI-backed major-loop falsification (`RG-PHY-OBS-01`).
- 9-material deep regression + golden loop drift checks (`RG-VAL-M1-01`, `RG-VAL-M1-02`).
- Write/Verify stats exported and validated (`RG-VAL-M1-03`).

Acceptance thresholds (must all pass)
- `|Pr_error| <= 10%`
- `|Ec_error| <= 10%`
- `RMSE(P(E))/Ps <= 0.05`
- `LoopArea_error <= 25%`
- Golden drift: normalized RMS drift `<= 1e-3` vs approved baseline artifact.

## P2 — Extended Falsification + Uncertainty

Deliverables
- Switching kinetics falsification (`RG-PHY-OBS-02`) with explicit fit quality.
- FORC/minor-loop falsification (`RG-PHY-OBS-03`).
- Monte Carlo uncertainty propagation (`RG-VAL-M1-04`).

Acceptance thresholds (must all pass)
- Kinetics fit: `R^2 >= 0.95` and parameter CI width <= 30% of estimate.
- FORC/minor-loop: normalized shape error `<= 0.10`, return-point error `<= 1% Ps`.
- Uncertainty: literature target metric lies inside 95% CI for `Pr` and `Ec`.

### Statistical Method Policy (enforceable)

Sample-size minima
- Seeded scalar metrics (pulses, overshoots, drift): `n >= 30` runs per material/engine in nightly, `n >= 100` in release.
- Distribution-comparison metrics (KS): `n >= 200` samples per distribution; otherwise mark test `insufficient_n` and fail gate.
- Proportion metrics (success/failure rates): `n >= 200` writes per material.

Decision rules
1. Run Shapiro-Wilk normality test at `alpha = 0.05` when `8 <= n <= 5000`.
2. If normality not rejected (`p >= 0.05`), report two-sided 95% t-interval.
3. Otherwise report BCa bootstrap 95% CI (`2000` resamples nightly, `10000` resamples release, fixed seed).
4. For proportions, always use Wilson 95% CI (not Wald).
5. Use two-sample KS only for continuous distributions and valid `n`; report `(D, p)`.
6. KS gate rule: `p <= 0.01` fail, `0.01 < p < 0.05` warning, `p >= 0.05` pass.

## 3) CI Gates with Exact Commands + Runtime Budgets

All commands executed from repo root:

```bash
cd <local-path>
export DISPLAY=
export WAYLAND_DISPLAY=
```

## PR Gate (target <= 12 min, hard cap 15 min)

Purpose: fail fast on regressions; run P0 + minimal P1.

```bash
go build ./... && go vet ./...
go test -short -count=1 ./...
go test -v -count=1 ./validation/literature/... -run TestModule1_PELoop_LiteratureBacked
```

Pass criteria
- Exit code 0 for every command.
- Required falsification thresholds pass for each dataset/material in run.
- Artifacts generated for each falsification test invocation.

## Nightly Gate (target <= 45 min, hard cap 60 min)

Purpose: full P1 + broad stability checks.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./module1-hysteresis/... ./shared/physics/...
```

Pass criteria
- All PR criteria, plus race-free execution.
- 9-material matrix complete (no missing material verdicts).

## Release Gate (target <= 90 min, hard cap 120 min)

Purpose: P0 + P1 + P2 publication-grade evidence.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./...
```

Pass criteria
- All nightly criteria, plus P2 thresholds met.
- Release artifact bundle produced (Section 5) with immutable commit hash.

### Runtime Impact Estimates for Added Statistical Controls

Estimated incremental cost vs current non-statistical baseline (same hardware class):
- Shapiro-Wilk + CI selection metadata: `< 30s` per nightly run.
- BCa bootstrap (`2000` resamples, nightly): `+4 to +8 min`.
- BCa bootstrap (`10000` resamples, release): `+15 to +30 min`.
- KS drift checks across required artifacts: `+1 to +3 min`.
- JSON uncertainty-schema validation: `< 1 min`.

These estimates are already included in the PR/Nightly/Release runtime budgets above.

## 4) Falsification Matrix (enforceable)

| ID | Observable | Required metric(s) | Fail condition |
|---|---|---|---|
| RG-PHY-OBS-01 | Major P–E loop vs DOI data | Pr error, Ec error, RMSE/Ps, loop area error | Any metric over threshold |
| RG-PHY-OBS-02 | Switching kinetics vs DOI data | R^2, parameter CI width, residual diagnostics | R^2 < 0.95 or CI too wide |
| RG-PHY-OBS-03 | FORC/minor loops vs DOI data | Shape error, return-point error | Any metric over threshold |
| RG-VAL-M1-01 | 9-material regression | per-material pass count | Any material missing/failing |
| RG-VAL-M1-02 | Golden regression | normalized RMS drift | Drift > 1e-3 |
| RG-VAL-M1-03 | WriteVerifyStats export | schema compliance + finite values | Missing/invalid field |
| RG-VAL-M1-04 | Monte Carlo UQ | 95% CI coverage | Target outside CI |

## 5) Artifact Contract (required JSON schema)

Each required test writes one JSON artifact under:
- `output/validation/module1/<gate>/<test_id>/<material>/<dataset>.json`

Minimal schema (required keys)

```json
{
  "schema_version": "m1.validation.v1",
  "timestamp_utc": "RFC3339",
  "commit": "<git sha>",
  "gate": "pr|nightly|release",
  "test_id": "RG-PHY-OBS-01",
  "material": {
    "name": "string",
    "Ec_Vm": 0,
    "Ps_Cm2": 0,
    "Pr_Cm2": 0,
    "thickness_m": 0,
    "Gmin_S": 0,
    "Gmax_S": 0
  },
  "dataset": {
    "doi": "string",
    "source_ref": "figure/table identifier",
    "units": {"E": "MV/cm", "P": "uC/cm2"}
  },
  "metrics": {
    "pr_error_pct": 0,
    "ec_error_pct": 0,
    "rmse_over_ps": 0,
    "loop_area_error_pct": 0,
    "r2": 0,
    "return_point_error_over_ps": 0
  },
  "uncertainty": {
    "method": "t|bootstrap_bca|wilson",
    "confidence": 0.95,
    "sample_size": 0,
    "normality": {"test": "shapiro_wilk", "p_value": 0},
    "bootstrap": {"resamples": 0, "seed": 0},
    "ks": {"d_stat": 0, "p_value": 0, "baseline_ref": "optional"},
    "ci": {
      "Pr_Cm2": {"low": 0, "high": 0},
      "Ec_Vm": {"low": 0, "high": 0}
    }
  },
  "thresholds": {
    "pr_error_pct_max": 10,
    "ec_error_pct_max": 10,
    "rmse_over_ps_max": 0.05,
    "loop_area_error_pct_max": 25
  },
  "verdict": "pass|fail",
  "notes": "optional"
}
```

Schema enforcement
- Missing required key, NaN/Inf metric, or unit mismatch => automatic fail.
- `uncertainty.sample_size` must satisfy policy minima for the metric class; otherwise fail.
- If `uncertainty.method=bootstrap_bca`, `bootstrap.resamples` must be `>=2000` nightly and `>=10000` release.
- If KS is used, `ks.p_value` and `ks.d_stat` are required and must reference a baseline artifact.
- `verdict` must be derivable from metrics + thresholds (no manual override).

## 6) Execution Order (single source of truth)

1. Run PR gate on every PR.
2. Run nightly gate once per day on default branch.
3. Run release gate before tagging release.
4. Publish artifact bundle and keep immutable by commit hash.

## 7) Definition of Done

Plan execution is complete when:
- P0, P1, P2 all implemented and passing at release gate.
- All falsification IDs above have artifacts + explicit pass/fail verdicts.
- Runtime budgets are met (or documented exception approved with reason).
- No unresolved per-material failures.
