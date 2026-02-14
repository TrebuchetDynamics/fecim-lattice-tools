# Module 4 Automation (Circuits + Peripherals)

This document describes the **automated validation suite for Module 4** (peripheral circuits, array-sim electrical consistency, and compute-read reliability metrics).

## What the automation validates

### Thermodynamics / energy accounting (Phase 1 / P0)
Validated in `module4-circuits/pkg/arraysim/`:
- **Power conservation** (budget closure within tolerance)
- **Cell dissipation** (no negative/physically impossible power)
- **Wire loss** (I²R consistency, included in the total budget)
- **Energy monotonicity** (write/ISPP energy accumulates monotonically; no unphysical energy creation)

### Standard access patterns (Phase 1 / P0)
Validated in `module4-circuits/pkg/arraysim/`:
- **Checkerboard** and stress-y mixed states (pattern coverage)
- **Walking ones/zeros** (systematic line activation coverage)
- **March C-** (read/write/stuck-at coverage)
- **Sneak-path worst case** (identifiable and bounded current ratios)

### Retention (Phase 2 / P1)
Validated in `module4-circuits/pkg/gui/`:
- **Half-select retention stress** (drift bounded by model constraints)
- **Read-stress retention** (mid-level conductance remains stable under repeated reads)

### Write disturb (Phase 2 / P1)
Validated in `module4-circuits/pkg/gui/`:
- **N-cycle disturb** (half-selected accumulation stays bounded over cycles)
- **Spatial disturb** (disturb decays away from selected row/column)

### PVT (Phase 2 / P1)
Validated in `module4-circuits/pkg/gui/`:
- **Temperature sweep** (Ec/Pr trends remain monotonic + functional)
- **Process corners / Monte Carlo** (yield above threshold)
- **Peripheral PVT coupling** (DAC write voltage responds to PVT knobs)

### Compute/read quality metrics (Phase 3)
Validated in `module4-circuits/pkg/gui/`:
- **MVM accuracy** vs ideal / Tier-A target
- **BER** under sense/noise model
- **Read margin** (adjacent levels separated by ≥ 3σ)

### Peripheral regression (Phase 3)
Validated in `shared/peripherals/`:
- **DAC/ADC INL/DNL** regression constraints
- **Noise model validation** (thermal/flicker/shot composition; ADC/TIA impact)

## How to run

### Fast gate (Phase 1 / P0 only)
~30s typical.

```bash
bash scripts/module4_automation.sh --fast
```

### Full gate (all phases)
~60s typical.

```bash
bash scripts/module4_automation.sh --full
```

### JSON summary output

```bash
bash scripts/module4_automation.sh --full --json
```

## Interpreting failures (fast triage)

- **KCL residuals > 1e-12**
  - Treat as: **solver/accounting bug** (numerical or topology error)

- **BER > 5%**
  - Treat as: **noise / quantization / sense chain issue** (or unrealistic noise model)

- **Read margin < 3σ**
  - Treat as: **level spacing problem** (insufficient separation, calibration, or drift)

- **Power conservation error > 1%**
  - Treat as: **energy accounting error** (missing term or sign inconsistency)

## Phases and delivering commits

| Phase | Scope | Primary artifacts | Delivering commit |
|---|---|---|---|
| Phase 1 (P0) | Thermodynamics + standard patterns + Kirchhoff consistency | `module4-circuits/pkg/arraysim/*_test.go` | `6931174` (thermo), `92a9676` (patterns) |
| Phase 2 (P1) | Retention + write disturb + PVT | `module4-circuits/pkg/gui/*retention*/*disturb*/*pvt*_test.go` | `b6d507a` |
| Phase 3 | MVM + BER + read margin + peripherals regression | `module4-circuits/pkg/gui/*mvm*/*ber*/*margin*_test.go`, `shared/peripherals/*_test.go` | `42e68dd` |
