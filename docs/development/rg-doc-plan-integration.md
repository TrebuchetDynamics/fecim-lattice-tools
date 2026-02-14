# RG-DOC Plan Integration Notes (Module 1 + Module 4)

This document closes the RG-DOC checklist items by recording how the historic automated-testing plans map to the current **implemented** test lanes, scripts, and TODO trackers.

## Module 4 plan → current reality

The large research-grade testing plan in `module4-automated-testing-plan.md` has been substantially implemented via:

- Deterministic, headless-required lanes (no DISPLAY / no xvfb) in CI scripts.
- Material-aware headless regression runners.
- Kirchhoff/current-conservation validation and ngspice cross-check harnesses.
- Standard patterns (checkerboard, walking ones/zeros, March C/C-).
- MVM accuracy + BER + read-margin sweeps + peripheral chain validation.
- Retention / disturb / PVT validation lanes.
- Unified automation runner: `scripts/module4_automation.sh --fast/--full/--json`.

Remaining plan items that are **still open** are already tracked as explicit TODO IDs in `TODO.md` (e.g., RG-VAL-03..05 and RG-PAR-03..05). No duplicate tickets were created.

## Module 1 plan → current reality

The research-grade plan in `modul1-automated-testing-plan.md` is reflected by the current physics test pyramid and headless regression lanes:

- Literature-aligned validation (Pr/Ec bands, material snapshots).
- Cross-engine consistency tests (Preisach vs LK where applicable).
- Property-based and fuzz stability layers.
- Versioned golden/curve regression in `validation/`.
- Deterministic headless ISPP/WRD regressions with explicit per-material verdicts.

Remaining plan gaps are already tracked as TODO IDs (not duplicated), notably parity trace artifacts and additional material deep-regression expansion items.
