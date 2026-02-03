# Physics Realism Audit

Date: 2026-02-03
Scope: physics and physics-adjacent models in modules 1–6 and shared physics/peripherals.

## Purpose
This audit documents where models are physically grounded versus simplified, heuristic, or purely illustrative. The goal is not to remove all simplifications, but to make them explicit and prioritize upgrades that improve scientific fidelity without breaking the educational UI.

## Method
- Read code paths that implement physics or physics-adjacent behavior.
- Tag each model as **High**, **Medium**, or **Low** realism.
- Note the main non-physics shortcuts and their consequences.
- Provide targeted upgrade suggestions.

## Summary (by module)

| Area | Realism | Main Non-Physics Behaviors | Primary Evidence |
|---|---|---|---|
| Hysteresis (Preisach) | Medium | Quasi-static hysteresis, tanh Everett, heuristic stress/temp scaling | `module1-hysteresis/pkg/ferroelectric/preisach.go` |
| Hysteresis (Landau) | Medium | 1D ODE, lumped depolarization term, empirical NLS | `shared/physics/landau.go` |
| Crossbar (array + non-idealities) | Medium | Linear conductance mapping, iterative IR drop solver, Elmore RC, assumed drift | `module2-crossbar/pkg/crossbar/*.go` |
| MNIST (CIM inference) | Low | Linear quantization, software-only inference, noise as proxy | `module3-mnist/pkg/core/*.go` |
| Circuits (DAC/ADC/TIA/chargepump) | Low | Parametric formulas, estimated energy, no transistor-level behavior | `shared/peripherals/*.go` |
| Comparison/EDA | Low | Estimated chip metrics, analytic latency/energy, no P&R or parasitic extraction | `module5-comparison/pkg/comparison/architecture.go`, `module6-eda/pkg/compiler/compiler.go` |

## Findings by Area

### 1) Hysteresis: Preisach Model
Realism: **Medium**

Key simplifications:
1. Quasi-static hysteresis (no switching kinetics or domain dynamics). This captures memory but not time-dependent switching.
2. Everett function is a **tanh approximation**, not a calibrated distribution of hysterons.
3. Stress and temperature scaling are **linear/heuristic**, not derived from a full Landau free energy model.
4. Discrete states are evenly spaced between ±Ps (not derived from device physics).

Evidence:
- `module1-hysteresis/pkg/ferroelectric/preisach.go`
- `shared/physics/material.go`

Impact:
- Correct qualitative hysteresis behavior, but limited quantitative accuracy for switching speed, minor-loop curvature, and temperature-stress coupling.

Recent fix already applied:
- Added reversible dielectric polarization so P relaxes when E is reduced.

### 2) Hysteresis: Landau‑Khalatnikov Solver
Realism: **Medium**

Key simplifications:
1. Single-domain 1D ODE. No spatial domain structure or phase-field dynamics.
2. Depolarization term `K_dep` is a tuning knob to create analog slope, not a first-principles interface model.
3. Nucleation-limited switching is empirical; time constants are not calibrated to device data.

Evidence:
- `shared/physics/landau.go`

Impact:
- Produces plausible dynamic switching curves but cannot predict domain wall behavior, stochastic switching distributions, or size scaling.

### 3) Crossbar Array + Non-Idealities
Realism: **Medium**

Key simplifications:
1. Conductance mapping is linear/exponential/lookup and not a compact device model.
2. IR-drop solver is iterative and approximate, not a full circuit simulator.
3. RC delay uses Elmore approximation with assumed wire capacitances.
4. Drift uses assumed coefficients; not tied to measured retention curves.
5. Half-select disturb is linear and per-pulse.

Evidence:
- `module2-crossbar/pkg/crossbar/array.go`
- `module2-crossbar/pkg/crossbar/nonidealities.go`
- `module2-crossbar/pkg/crossbar/drift.go`
- `module2-crossbar/pkg/crossbar/irdrop.go`
- `module2-crossbar/pkg/crossbar/sneakpath.go`

Impact:
- Good for qualitative non-ideality education, not for predictive circuit accuracy or process-node scaling.

### 4) MNIST CIM Inference
Realism: **Low**

Key simplifications:
1. Quantization is purely mathematical (linear binning), not a device write/read process.
2. Noise injection is a software proxy, not a modeled circuit readout or variability distribution.
3. No peripheral circuit constraints (ADC/DAC timing, settling, bandwidth).

Evidence:
- `module3-mnist/pkg/core/quantize.go`
- `module3-mnist/pkg/core/network.go`

Impact:
- Useful for exploring quantization sensitivity, not for predicting hardware accuracy or energy.

### 5) Circuits (DAC/ADC/TIA/Charge Pump)
Realism: **Low**

Key simplifications:
1. Uses parametric formulas for INL/DNL, energy, and settling.
2. No transistor-level modeling, noise spectra, or PVT variation.
3. Energy estimates are heuristic and calibrated to “typical” values, not measured data.

Evidence:
- `shared/peripherals/dac.go`
- `shared/peripherals/adc.go`
- `shared/peripherals/tia.go`
- `shared/peripherals/chargepump.go`

Impact:
- Good for intuition but not for circuit design or signoff.

### 6) Comparison + EDA
Realism: **Low**

Key simplifications:
1. FeCIM architecture metrics are explicitly estimated and not experimentally validated.
2. Latency/energy computed analytically without model-specific bottlenecks or IO constraints.
3. EDA compiler does mapping and export, not placement, routing, or parasitic extraction.

Evidence:
- `module5-comparison/pkg/comparison/architecture.go`
- `module6-eda/pkg/compiler/compiler.go`

Impact:
- Suitable for visualization and early trade studies, not for hardware planning.

### 7) Materials and Calibration Inputs
Realism: **Mixed**

Key issues:
1. FeCIM target parameters include estimated values and conference claims.
2. Some parameters are literature-based but not tied to a specific measurement setup.

Evidence:
- `shared/physics/material.go`
- `data/calibrations/*.json`

Impact:
- Material parameters are reasonable for education but should not be treated as device-validated.

## Priority Recommendations

P0 (Correctness and disclosure)
1. Ensure every simplified model has a clear in-UI warning or docs statement that it is non-physical or heuristic.
2. Keep “conference claim” and “target” values clearly labeled in UI and docs.

P1 (Physics quality with minimal risk)
1. Calibrate Preisach/everett distribution to at least one measured P–E dataset.
2. Add measured retention/relaxation curves for drift and update drift coefficients.
3. Replace linear quantization in MNIST with a device-level write/read model (even a simplified one).

P2 (Higher fidelity)
1. Add SPICE-validated peripheral models for DAC/ADC/TIA (even coarse macromodels).
2. Add a compact FeFET conductance model for crossbar cells (V, T, history dependence).
3. Add spatial domain effects for Landau (phase-field or simplified multi-domain).

## Suggested Validation Tests

1. Match a published P–E loop with calibrated Preisach and Landau models.
2. Reproduce a published retention curve to verify drift coefficients.
3. Cross-validate IR drop vs. SPICE for a small array (e.g., 8×8).
4. Use a known ADC model to validate quantization error and SNR.

## Notes
- This audit focuses on physics realism, not UI correctness or performance.
- The project already includes honesty disclaimers and a related audit in `docs/comparison/HONESTY_AUDIT.md`.
