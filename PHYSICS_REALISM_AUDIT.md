# Physics Realism Audit

**Date:** 2026-02-03
**Scope:** Physics and physics-adjacent models in modules 1–6 and shared physics/peripherals
**Status:** Living document - update as improvements are made

---

## Executive Summary

This simulator prioritizes **educational clarity** over predictive accuracy. Most models are **Medium to Low realism** - sufficient for teaching concepts but not for hardware design or device validation.

| Realism Level | Meaning | Use Case |
|---------------|---------|----------|
| **High** | Calibrated to measured data, validated against literature | Device design, prediction |
| **Medium** | Physically-motivated but uses heuristics or uncalibrated parameters | Concept education, qualitative exploration |
| **Low** | Parametric/analytic approximations, not tied to device physics | Visualization, intuition building |

**Key takeaway:** All simplified models should have explicit disclaimers in the UI and docs.

---

## Realism Summary by Module

| Area | Realism | Status | Main Simplifications |
|------|---------|--------|---------------------|
| Hysteresis (Preisach) | Medium | Active | Quasi-static, tanh Everett, heuristic scaling |
| Hysteresis (Landau) | Medium | Active | 1D ODE, lumped depolarization, empirical NLS |
| Crossbar Array | Medium | Active | Linear conductance, iterative IR drop, Elmore RC |
| MNIST CIM | Low | Active | Linear quantization, software noise proxy |
| Circuits (DAC/ADC/TIA) | Low | Active | Parametric formulas, no transistor-level |
| Comparison/EDA | Low | Active | Estimated metrics, no P&R or parasitics |

---

## Detailed Findings

### 1. Hysteresis: Preisach Model

**Realism:** Medium
**Files:** `module1-hysteresis/pkg/ferroelectric/preisach.go`, `shared/physics/material.go`

| Simplification | Impact | Upgrade Path |
|----------------|--------|--------------|
| Quasi-static hysteresis (no switching kinetics) | Captures memory, not time-dependent switching | Add KAI model or measured switching times |
| tanh Everett function | Smooth but uncalibrated | Fit to measured first-order reversal curves (FORC) |
| Linear stress/temperature scaling | Qualitative only | Derive from Landau free energy expansion |
| Evenly-spaced discrete states | Not physics-based | Use measured state distributions |

**Validation target:** Match published P-E loop (e.g., HZO from Nature Commun. 2025).

---

### 2. Hysteresis: Landau-Khalatnikov Solver

**Realism:** Medium
**Files:** `shared/physics/landau.go`

| Simplification | Impact | Upgrade Path |
|----------------|--------|--------------|
| Single-domain 1D ODE | No spatial effects | Add multi-domain or phase-field |
| K_dep as tuning knob | Creates analog slope but not physical | Model interface depolarization field |
| Empirical NLS time constants | Uncalibrated | Fit to measured switching distributions |

**Validation target:** Reproduce switching transient from Muller et al. or similar.

---

### 3. Crossbar Array + Non-Idealities

**Realism:** Medium
**Files:** `module2-crossbar/pkg/crossbar/*.go`

| Component | Simplification | Upgrade Path |
|-----------|----------------|--------------|
| Conductance mapping | Linear/exponential, no compact model | Add FeFET I-V model |
| IR drop solver | Iterative approximation | Validate vs SPICE for 8x8 |
| RC delay | Elmore with assumed C | Extract from layout or PDK |
| Drift model | Assumed coefficients | Calibrate to retention data |
| Half-select disturb | Linear per-pulse | Add cumulative/threshold model |

**Validation target:** Cross-validate IR drop vs SPICE for small array.

---

### 4. MNIST CIM Inference

**Realism:** Low
**Files:** `module3-mnist/pkg/core/quantize.go`, `module3-mnist/pkg/core/network.go`

| Simplification | Impact | Upgrade Path |
|----------------|--------|--------------|
| Linear binning quantization | Not a device model | Add write/read cycle model |
| Software noise injection | Proxy only | Model ADC noise, cell variation |
| No peripheral constraints | Missing real bottlenecks | Add ADC/DAC timing limits |

**Validation target:** Compare quantization error to measured device variation.

---

### 5. Circuits (DAC/ADC/TIA/Charge Pump)

**Realism:** Low
**Files:** `shared/peripherals/*.go`

| Simplification | Impact | Upgrade Path |
|----------------|--------|--------------|
| Parametric INL/DNL formulas | Heuristic only | Use measured ADC data |
| No transistor-level modeling | Missing noise, PVT | Add SPICE macromodels |
| Heuristic energy estimates | Not calibrated | Tie to measured power |

**Validation target:** Validate ADC SNR against known model.

---

### 6. Comparison + EDA

**Realism:** Low
**Files:** `module5-comparison/pkg/comparison/architecture.go`, `module6-eda/pkg/compiler/compiler.go`

| Simplification | Impact | Upgrade Path |
|----------------|--------|--------------|
| Estimated FeCIM metrics | Not experimentally validated | Replace with measured data when available |
| Analytic latency/energy | No model-specific bottlenecks | Add IO and memory bandwidth limits |
| No placement/routing | Missing parasitics | Add basic floorplan model |

---

## Priority Recommendations

### P0: Correctness & Disclosure (Required)

| Task | Status | Owner |
|------|--------|-------|
| Add UI warning on all simplified models | [ ] | - |
| Label "conference claim" values distinctly | [x] | - |
| Document model limitations in tooltips | [ ] | - |

### P1: Physics Quality (Low Risk)

| Task | Status | Validation |
|------|--------|------------|
| Calibrate Preisach to one measured P-E dataset | [ ] | Match loop shape within 10% |
| Add measured retention curve for drift | [ ] | Reproduce published decay |
| Replace linear quantization with device model | [ ] | Compare to measured cell distribution |

### P2: Higher Fidelity (Future)

| Task | Status | Notes |
|------|--------|-------|
| SPICE-validated DAC/ADC macromodels | [ ] | Requires PDK or published model |
| Compact FeFET conductance model | [ ] | V, T, history dependence |
| Multi-domain Landau (phase-field lite) | [ ] | Significant complexity increase |

---

## Validation Test Plan

| Test | Target | Pass Criteria |
|------|--------|---------------|
| P-E loop matching | Published HZO data | RMS error < 10% |
| Retention curve | Published 10-year extrapolation | Same decay exponent |
| IR drop vs SPICE | 8x8 array | Max error < 5% |
| ADC quantization | Known ADC model | SNR within 3 dB |

---

## References

Key literature for calibration and validation:

1. **HZO P-E characteristics:** Nature Communications 2025 (Pr: 15-34 µC/cm²)
2. **Switching dynamics:** Muller et al., various IEEE publications
3. **Retention/endurance:** Nano Letters 2024 (V:HfO₂, 10¹² cycles)
4. **CIM accuracy benchmarks:** Nature Communications 2023 (96.6% MNIST)

---

## Changelog

| Date | Change |
|------|--------|
| 2026-02-03 | Initial audit created |
| 2026-02-03 | Added executive summary, actionable task tables, validation plan |
