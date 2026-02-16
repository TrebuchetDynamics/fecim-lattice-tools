# Module 1 Automated Testing Plan (Execution-Grade)

Scope: `module1-hysteresis` physics/controller validation in headless CI.

## Operating Contract

- Headless only: `DISPLAY` and `WAYLAND_DISPLAY` must be unset.
- No aggregate pass if any required material/dataset fails.
- Required lanes must emit machine-readable artifacts.
- Thresholds and runtime budgets below are hard gates.

---

## P0 / P1 / P2 Waves

### P0 — CI Safety Baseline (required first)
**Deliverables**
- Deterministic build/test lane for Module 1.
- Artifact emission for required validation tests.
- Explicit material matrix (no implicit defaults).

**Exit criteria**
- PR and nightly lanes pass with complete artifacts.
- Artifact schema validates for every required run.

### P1 — Physics Falsification Core (primary gate)
**Deliverables**
- `RG-PHY-OBS-01`: DOI-backed major-loop falsification.
- `RG-VAL-M1-01`, `RG-VAL-M1-02`: 9-material regression + golden drift.
- `RG-VAL-M1-03`: Write/Verify stats export + schema validation.

**Hard thresholds (all required)**
- `|Pr_error| <= 10%`
- `|Ec_error| <= 10%`
- `RMSE(P(E))/Ps <= 0.05`
- `LoopArea_error <= 25%`
- Golden normalized RMS drift `<= 1e-3`

### P2 — Extended Falsification + Uncertainty
**Deliverables**
- `RG-PHY-OBS-02`: switching kinetics falsification.
- `RG-PHY-OBS-03`: FORC/minor-loop falsification.
- `RG-VAL-M1-04`: Monte Carlo uncertainty propagation.

**Hard thresholds (all required)**
- Kinetics: `R^2 >= 0.95`, parameter CI width `<= 30%` of estimate.
- FORC/minor-loop: normalized shape error `<= 0.10`, return-point error `<= 1% Ps`.
- UQ: literature target lies inside 95% CI for `Pr` and `Ec`.

---

## Command Lanes (PR / Nightly / Release)

Run from repo root:

```bash
cd <local-path>
export DISPLAY=
export WAYLAND_DISPLAY=
```

### PR lane (P0 + minimal P1)
**Runtime budget:** target `<= 12 min`, hard cap `15 min`.

```bash
go build ./... && go vet ./...
go test -short -count=1 ./...
go test -v -count=1 ./validation/literature/... -run TestModule1_PELoop_LiteratureBacked
```

**Pass requires**
- Exit code 0 for every command.
- P1 thresholds pass for each dataset/material exercised.
- Required artifacts emitted.

### Nightly lane (full P1)
**Runtime budget:** target `<= 45 min`, hard cap `60 min`.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./module1-hysteresis/... ./shared/physics/...
```

**Pass requires**
- PR lane pass conditions.
- Full 9-material matrix complete.
- Race lane clean.

### Release lane (P0 + P1 + P2)
**Runtime budget:** target `<= 90 min`, hard cap `120 min`.

```bash
go build ./... && go vet ./...
go test -count=1 ./...
go test -v -count=1 ./validation/...
go test -v -count=1 ./validation/literature/...
bash scripts/run_literature_validation.sh
go test -race ./...
```

**Pass requires**
- Nightly lane pass conditions.
- All P2 thresholds satisfied.
- Immutable release artifact bundle keyed by commit SHA.

---

## Statistical Policy (enforced)

### Minimum sample sizes
- Seeded scalar metrics: `n >= 30` nightly, `n >= 100` release.
- Distribution metrics (KS): `n >= 200` per distribution.
- Proportion metrics: `n >= 200` writes per material.

If minima are unmet: mark `insufficient_n` and fail gate.

### CI / hypothesis rules
1. Shapiro-Wilk (`alpha=0.05`) when `8 <= n <= 5000`.
2. If normality not rejected (`p >= 0.05`): two-sided 95% t-interval.
3. Else: BCa bootstrap 95% CI (`2000` nightly, `10000` release; fixed seed).
4. Proportions: Wilson 95% CI.
5. KS for continuous distributions only; report `(D, p)`.
6. KS gate: `p <= 0.01` fail; `0.01 < p < 0.05` warning; `p >= 0.05` pass.

---

## Falsification Matrix

| ID | Observable | Required metrics | Hard fail condition |
|---|---|---|---|
| RG-PHY-OBS-01 | Major P–E loop vs DOI data | Pr error, Ec error, RMSE/Ps, loop area error | Any metric above threshold |
| RG-PHY-OBS-02 | Switching kinetics vs DOI data | R^2, parameter CI width, residual diagnostics | `R^2 < 0.95` or CI width too large |
| RG-PHY-OBS-03 | FORC/minor loops vs DOI data | Shape error, return-point error | Any metric above threshold |
| RG-VAL-M1-01 | 9-material regression | Per-material pass | Any missing/failing material |
| RG-VAL-M1-02 | Golden regression | Normalized RMS drift | Drift `> 1e-3` |
| RG-VAL-M1-03 | WriteVerifyStats export | Schema + finite values | Missing/invalid field |
| RG-VAL-M1-04 | Monte Carlo UQ | 95% CI coverage | Target outside CI |

---

## Artifact Contract

**Path**
- `output/validation/module1/<gate>/<test_id>/<material>/<dataset>.json`

**Required keys**
- `schema_version`, `timestamp_utc`, `commit`, `gate`, `test_id`
- `material{...}`, `dataset{doi,source_ref,units}`
- `metrics{...}`, `thresholds{...}`, `verdict`
- `uncertainty{method,confidence,sample_size,ci,...}` where applicable

**Enforcement**
- Missing required key, NaN/Inf, or unit mismatch => fail.
- Sample-size minima must be met.
- `bootstrap_bca` must use `>=2000` nightly, `>=10000` release.
- KS entries require `d_stat`, `p_value`, and baseline reference.
- `verdict` must be derivable from metrics/thresholds (no manual override).

---

## Execution Order (authoritative)

1. PR lane on every PR.
2. Nightly lane once/day on default branch.
3. Release lane before tagging release.
4. Publish immutable artifact bundle keyed by commit SHA.

## Definition of Done

Done when all are true:
- P0, P1, P2 implemented and passing in release lane.
- All listed IDs have artifacts + explicit pass/fail verdicts.
- Runtime budgets met (or exception documented/approved).
- No unresolved per-material failures.
