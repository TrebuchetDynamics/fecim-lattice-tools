# Physics Realism Audit Addendum -- February 2026

**Date:** 2026-02-18
**Session:** sciomc-20260218 (7-stage parallel research synthesis)
**Relation to:** `PHYSICS_REALISM_AUDIT.md` (2026-02-03, updated 2026-02-11)

---

## Purpose

This addendum updates the Physics Realism Audit based on new literature evidence gathered during the February 18, 2026 research session. It covers updated realism ratings, specific parameter corrections, and newly identified gaps.

---

## 1. Updated Realism Ratings

### 1.1 Hysteresis: Preisach Model

**Previous rating:** Medium
**Updated rating:** Medium (unchanged, with increased confidence)

**Rationale:** The product-form tanh Everett function `[1+tanh((alpha-Ec)/Delta)][1-tanh((beta+Ec)/Delta)]*Ps/4` is confirmed as the integral of a sech^2 Preisach density, which is physically reasonable for HZO. The difference between tanh Everett and the Gaussian (erf) alternative is <2% in remanent polarisation for symmetric, fresh HZO (Stage 2, Finding 1.4). The FORC density computation algorithm (`ComputeFORCDensity`) matches the literature-standard central finite-difference approximation. The PolydomainEnsemble distributed-Ec approach is validated as the correct model for device-level variability (Stage 2, Finding 5.2: process-induced stochasticity dominates over thermal noise).

**New simplifications identified:**

| Simplification | Physical Impact | Source |
|----------------|----------------|--------|
| Symmetric tanh Everett (single Delta for both tails) | Cannot capture imprint-induced asymmetry in Preisach density | S2 Finding 1.3: imprint shifts density peak along interaction-field axis |
| FORC output not used for calibration | Delta is set by bisection on Pr, not by FORC-measured sigma_Ec | S2 Finding 1.2: gap between existing FORC engine and calibration |
| No wake-up evolution of Preisach density | Peak splitting and intensity changes during cycling not modelled | S2 Finding 1.1: FORC reveals density evolution during wake-up |

**New validation target:** Match published FORC density peak position and width for 10 nm HZO: Ec_peak ~ 1.0-1.5 MV/cm, FWHM ~ 0.5-1.0 MV/cm (Stage 2, inferred from multiple sources).

---

### 1.2 Hysteresis: Landau-Khalatnikov Solver

**Previous rating:** Medium
**Updated rating:** Medium (unchanged, with clarified limitations)

**Rationale:** The LK04 self-calibration via `ConfigureFromMaterial` (which rescales alpha, beta, gamma to match the material's stated Ec) is confirmed as a sound engineering choice that reduces sensitivity to absolute Landau coefficient values (Stage 1, Finding on LK04). The 6th-order Landau polynomial with depolarisation is consistent with state-of-the-art formulations. The depolarisation coefficient K_dep = 2.5e8 V*m/C is within the documented recommended range of 1-5e8 V*m/C (Stage 2, Finding 7.1).

**Critical parameter correction:** See Section 2.1 below (NLS tau_0).

**New fundamental limitation identified:**

| Limitation | Physical Impact | Source |
|-----------|----------------|--------|
| Single-order-parameter model cannot capture o-t-o switching pathway | Switching in HZO proceeds via orthorhombic -> tetragonal -> orthorhombic transition; the tetragonal intermediate state is not representable in single-P Landau | S5 Section 5.1: Nature Comms 2025 + Advanced Materials 2025 |
| Complex domain walls require coupled P and eta | Hafnia DWs involve simultaneous reversal of polarisation AND tetragonality; standard 180-degree wall model is incorrect | S5 Section 3.2: PNAS 2024 |
| No wake-up model (90 -> 180 degree DW transition) | Pristine HZO has 90-degree UCDWs; cycling converts to 180-degree DWs with 2.3x Pr increase | S5 Section 3.3: PMC11789571 (2025) |

**Upgrade path (Tier 1, minimal):** Add NLS-envelope kinetic gating as an optional mode flag. For each time step, compute switched_fraction via NLS integral with Lorentz distribution and gate the polarisation update accordingly. This adds time-dependent switching physics without changing the core LK solver architecture. At ISPP operating conditions (>100 ns pulses, 1.5-3 MV/cm), the kinetic error is bounded at <5% (Stage 5, key finding), so this is low priority for the educational mission.

**Upgrade path (Tier 3, full):** Coupled P-eta phase-field model with two Landau order parameters. Required to capture FE-AFE competition and the o-t-o switching pathway. Computationally expensive; consider GNN surrogates (PMC11059552, 4.24% error, million-fold speedup) as a future research direction.

---

### 1.3 Crossbar Array

**Previous rating:** Medium
**Updated rating:** Medium for resistive mode; LOW for FeCAP mode (downgraded)

**Rationale for FeCAP downgrade:** Three correctness issues identified:

1. **MAC domain error:** The simulator computes I = G*V (current domain) for all modes including FeCAP. The correct FeCAP model is Q = C(P)*V (charge domain). Bitline accumulates charge, not current. This is a physics model error, not a simplification (Stage 4, GAP-M2-01).

2. **Sneak path error:** DC sneak-path and IR-drop calculations are applied to FeCAP mode. FeCAP has zero DC current path -- charge flows only during voltage transitions. Applying DC non-ideality calculations produces physically meaningless results (Stage 4, GAP-M2-05).

3. **Sensing chain error:** TIA-based sensing is modelled for FeCAP. FeCAP requires charge-sensing amplifier (CSA) with switched-capacitor integration, not a transimpedance amplifier (Stage 4, GAP-M4-01).

These are not parameter calibration issues but structural model errors that produce incorrect physics for FeCAP mode. Until corrected, FeCAP mode output should be treated as unreliable.

**Resistive mode update:** The existing model is adequate for education. New literature confirms state-dependent C2C variation (sigma_G/G proportional to 1/G, with 2-3x ratio between low-G and high-G states) as a first-order missing effect (Stage 6, Finding in Section 3).

---

### 1.4 MNIST CIM Inference

**Previous rating:** Low
**Updated rating:** Low (unchanged)

**New context from literature:** The CIM field has moved substantially beyond MNIST. Current hardware demonstrations include CIFAR-10 at ~92%, speech recognition, and dynamic tracking at 99.8% accuracy. The project's MNIST-only benchmark is educationally adequate but should not be presented as representative of CIM capability. Reference tools: CrossSim V3.1 (Sandia) and NeuroSim V1.5 (Georgia Tech) are the standard benchmarking frameworks. The simulator lacks their key feature: hardware noise injection during inference (noise-aware evaluation).

**Key missing feature:** Accuracy-vs-levels sweep. The simulator uses a fixed 30-level quantisation. Literature shows accuracy plateaus above ~32 levels and drops sharply below ~8 levels (Stage 6). Exposing this as a configurable parameter with sweep capability would significantly increase educational value.

---

### 1.5 Circuits (DAC/ADC/TIA/ChargePump)

**Previous rating:** Low
**Updated rating:** Low (unchanged)

**New calibration data from literature:**

| ADC Type | Literature Energy/Column | Current Simulator | Gap |
|----------|------------------------|-------------------|-----|
| 7-bit SAR | ~26 pJ | Unknown (heuristic) | Needs validation |
| 4-bit Flash | 1.86 pJ | Unknown (heuristic) | Needs validation |
| ADC-less (ternary) | 0.22 pJ | Not modelled | Missing option |
| Nonlinear SAR (FeCAP) | ~4 pJ (estimated) | Not modelled | Missing option |

**New ADC architecture identified:** Power-of-two nonlinear SAR ADC (Yeo 2024, IEEE SSL) matches FeCAP capacitance distribution. This is the correct ADC for FeCAP crossbar CIM and should be added alongside existing SAR/Flash/Ramp options.

**ADC-less alternative:** Binary/ternary comparators replacing ADC entirely achieve 12-28x energy reduction at 1.8-3.5% accuracy cost on CIFAR-10 (HCiM, arXiv 2403.13577). This represents a fundamentally different architecture worth modelling.

---

## 2. Specific Parameter Corrections

### 2.1 NLS tau_0: 10^-13 s -> 10^-10 s (CRITICAL)

**Current values:**
- `shared/physics/landau.go`, `NewLKSolver()`: `TauInf = 1.0e-13`
- `shared/physics/nls.go`, NLSKinetics default: `Tau0 = 1e-13`

**Corrected value:** `1e-10` s

**Evidence:**
- Guo et al., APL 112, 262903 (2018): tau_inf = 1e-10 s (referenced in codebase as NLSSigma source)
- PMC9740545 (2022): tau_inf = 1e-10 s (classical NLS), 4e-10 s (general multi-mechanism)
- Stage 2, Finding 3.2: "multiple independent sources agree on 10^-10 s scale"
- Stage 5, Section 2.1: t_inf = 10^-12 to 10^-11 s for limiting nucleation event (DFT); macroscopic NLS uses 10^-10 s

**Physical impact:** The current 10^-13 s value causes the NLS kinetics to equilibrate approximately 1000x faster than experimentally measured for HZO. For a 100 ns ISPP pulse at E = 2 MV/cm:
- Current (tau_0 = 1e-13): tau(E) ~ 1.1e-9 s, switched fraction ~ 95%
- Literature (tau_0 = 1e-10): tau(E) ~ 7e-8 s, switched fraction ~ 40%

The 60 percentage point difference in switched fraction at short pulses means the current simulator significantly overestimates programming speed for nanosecond-scale operations.

**Note:** The codebase comment references "phonon frequency limit" as justification for 10^-13. This is the theoretical phonon attempt frequency for a single atom, not the macroscopic NLS pre-exponential measured for HZO device switching. The macroscopic tau_0 includes nucleation geometry and domain formation factors that increase it by approximately 3 orders of magnitude.

**Migration consideration:** Changing tau_0 will affect NLS-gated switching behaviour in the LK solver. Golden regression tests that exercise NLS kinetics will need regeneration. The ISPP controller tests should be re-validated after this change.

---

### 2.2 LiteratureSuperlattice Pr: 50 -> ~22 uC/cm^2

**Current value:** `shared/physics/material.go`, LiteratureSuperlattice preset: `Pr: 50e-2` (50 uC/cm^2 in code units)

**Corrected value:** `Pr: 22e-2` (22 uC/cm^2), corresponding to 2Pr ~ 44 uC/cm^2

**Evidence:**
- IEEE 10787441 (2024): Best 2025 nanolaminate 2Pr = 43.32 uC/cm^2, i.e., Pr ~ 21.66 uC/cm^2
- PMC 12254504 (2025): Epitaxial superlattice at 100 nm, Ec ~ 0.85 MV/cm
- Stage 1, Finding on superlattice Pr: "Simulator's Pr=50 uC/cm^2 is overstated by ~2.3x relative to best 2025 nanolaminate measurements"

**Context:** The 50 uC/cm^2 Pr value appears to originate from Cheema 2020 projections for theoretical superlattice maximum. The best 2025 experimental measurements yield Pr = 21-27 uC/cm^2 for actual fabricated superlattice devices.

**Recommended update:**
```
Pr: 22e-2   // ~22 uC/cm^2 (IEEE 10787441, 2024: 2Pr=43.32 uC/cm^2)
Ps: 27e-2   // ~27 uC/cm^2 (estimated Ps for superlattice)
```

---

### 2.3 LiteratureSuperlattice Endurance Label Update

**Current value:** `EnduranceCycles: 1e10` -- no explanatory comment

**Correction:** Keep 1e10 as the default value but add documentation noting that 10^12 is now demonstrated for La-doped 3D-trench devices (ACS AMI 2024), making the 10^10 value conservative.

---

## 3. New Gaps Identified by This Research

### 3.1 Missing Measurement Workflow Simulations

The Physics Realism Audit focused on model equations and did not assess the simulator's coverage of standard ferroelectric characterisation protocols. Stage 7 identified that the following measurement workflows are standard in ferroelectric labs but absent from Module 1:

| Workflow | Status | Literature Standard | Priority |
|----------|--------|-------------------|----------|
| PUND (switching charge separation) | Missing | Most-cited HZO measurement; >50 publications use it | P1 (HIGH) |
| C(V) butterfly curves | Missing | Standard LCR measurement; dP/dV computable from existing Preisach | P1 (HIGH) |
| Retention decay simulation | Missing (model exists, no workflow) | Power-law + Arrhenius extrapolation; industry standard | P1 (HIGH) |
| FORC density visualisation | Partial (computation exists, no UI) | Pike 2003 algorithm; standard in ferroelectric characterisation | P1 (MEDIUM) |
| I-V leakage models | Missing | Schottky/PF/FN; 3 analytical equations | P2 (MEDIUM) |
| Endurance cycling experiment | Partial (model exists, no experiment runner) | Bipolar cycling with PUND read at log intervals | P2 (MEDIUM) |
| Batch/recipe measurement mode | Missing | Multi-step measurement automation | P3 (LOW) |

**Significance:** Stage 7 confirmed that no comprehensive open-source ferroelectric measurement framework exists as of 2026. Implementing these workflows would be genuinely novel educational value.

---

### 3.2 Missing Wake-Up Pr(N) Non-Monotonic Model

**Current model:** `EnduranceAtCycles(N)` applies a monotonic stretched-exponential degradation from cycle 1.

**Literature reality:** Remanent polarisation follows a non-monotonic trajectory:
1. **Wake-up phase** (cycles 1 to ~10^3): Pr increases as tetragonal phase converts to ferroelectric orthorhombic phase. For HZO, Pr rises from ~12 to ~28 uC/cm^2 (2.3x increase, PMC11789571).
2. **Stable plateau** (cycles ~10^3 to ~10^6): Pr relatively constant.
3. **Fatigue phase** (cycles >10^6): Pr decreases as orthorhombic converts to monoclinic phase.

The current monotonic model misses the educational physics of the wake-up hump entirely.

**Recommended model:**
```
Pr(N) = Pr_0 * [1 - exp(-N/N_wakeup)]^alpha_w  *  [1 - beta_f * (N/N_fatigue)^gamma_f]
```
Where the first term captures the rising wake-up phase and the second term captures the declining fatigue phase. Default parameters: N_wakeup ~ 10^3, N_fatigue ~ 10^8, alpha_w ~ 0.3, beta_f ~ 0.5, gamma_f ~ 0.2.

---

### 3.3 Missing FE-AFE Phase Competition

**Current model:** Single-phase Landau double-well or single-peaked Preisach distribution.

**Literature reality:** The ferroelectric (Pca21) and antiferroelectric (Pbca) phases coexist in HZO. Their relative fractions determine device behaviour and are the primary source of device-to-device variability (Stage 1, Finding on phase competition). The o-t-o switching pathway (Stage 5, Section 5) is not representable in a single-order-parameter framework.

**Recommended minimal extension:** Add `AFEFraction float64` field to `HZOMaterial` (range 0.0 to 1.0). When > 0, scale the effective Pr by `(1 - AFEFraction)` and add a qualitative pinched-loop contribution to the P-E output. This captures the educational concept without requiring a full two-component Landau model.

---

### 3.4 Missing State-Dependent C2C Variation

**Current model:** Uniform variation applied across all conductance states.

**Literature reality:** Low-conductance states have 2-3x larger relative variation (sigma_G/G) than high-conductance states (Stage 6, Section 3). This asymmetry biases MVM outputs and is a first-order effect for inference accuracy.

**Recommended model:**
```
sigma_G(G) = sigma_base * (G_max / G)^0.5
```
Where sigma_base is the variation at G_max. This produces 2x larger variation at G_max/4 (low-G states), matching literature observations.

---

### 3.5 FeCAP Charge-Domain Physics (Structural Gap)

The existing crossbar model is built around current-domain physics (I = G*V). FeCAP operation is fundamentally charge-domain (Q = C*V). This is not a calibration issue but a structural model architecture gap. The correct FeCAP MAC is:

```
Q_total = sum_i C_i(w_i) * V_i(x_i)
```

Where C_i is the polarisation-state-dependent capacitance of cell i. The bitline is a charge integrator (capacitive node), not a current summing node. The sensing circuit must resolve accumulated charge, not current.

Key architectural differences:
- No DC current path -> no sneak paths, no IR drop (Stage 4, confirmed by 5 papers)
- Sensing via charge-redistribution SAR ADC or switched-capacitor CSA, not TIA
- Zero static power during computation (dynamic power only)
- Non-destructive readout possible with asymmetric electrodes (Imec 2025)

---

### 3.6 Missing Leakage Current Physics

The simulator does not model leakage current through the ferroelectric capacitor. Literature identifies three dominant mechanisms for HZO:

| Mechanism | Regime | Model Equation | Key Parameter |
|-----------|--------|---------------|---------------|
| Schottky emission | Mid-high field | J = A* T^2 exp[-(phi_B - sqrt(qE/4pi*eps_opt*eps_0)) / kT] | phi_B = 0.48-0.58 eV |
| Poole-Frenkel | Low-mid field | J = C_PF E exp[-(phi_T - sqrt(qE/pi*eps_opt*eps_0)) / kT] | phi_T = 0.3-1.0 eV |
| Fowler-Nordheim tunneling | Very high field | J = (q^3 E^2)/(8pi*h*phi_B) exp[-8pi*sqrt(2m*)*phi_B^1.5 / (3qhE)] | m* ~ 0.15 m_e |

Discriminator: PF slope is exactly 2x Schottky slope in ln(J) vs sqrt(E) plot. Optical dielectric constant from slope must match n^2 ~ 4-5 for HZO.

---

### 3.7 Missing Frequency-Dependent Hysteresis

The LK solver and Preisach model are quasi-static. Frequency-dependent effects arise from:
1. NLS switching time distribution (not all domains switch at high frequency)
2. RC time constant (intrinsic capacitance + series resistance)
3. Grain-size-mediated dielectric relaxation

At ISPP operating conditions (100 ns - 10 us pulses, 1.5-3 MV/cm), the quasi-static approximation is valid to within <5% (Stage 5, key finding). The error grows to >50% for sub-nanosecond pulses. A Jiles-Atherton compact model (Paasio 2025, Adv. Electron. Mater.) is an alternative that natively handles frequency dependence.

---

## 4. Summary: Updated Realism Ratings Table

| Area | Previous Rating | Updated Rating | Change Reason |
|------|----------------|---------------|---------------|
| Hysteresis (Preisach) | Medium | Medium | Confirmed by literature; no change |
| Hysteresis (LK) | Medium | Medium | Confirmed; NLS tau_0 correction needed |
| Crossbar (Resistive) | Medium | Medium | State-dependent C2C gap identified |
| Crossbar (FeCAP) | (not rated) | **Low** | Three structural physics errors in FeCAP mode |
| MNIST CIM Inference | Low | Low | Field moved beyond MNIST; no change |
| Circuits (DAC/ADC/TIA) | Low | Low | ADC energy data now available for calibration |
| NLS Kinetics | (implicit in LK) | **Medium, pending tau_0 fix** | tau_0 off by 3 OOM; physically correct after fix |
| Measurement Workflows | (not rated) | **Not Implemented** | PUND, C(V), retention, FORC, I-V all missing |

---

## 5. Priority Actions from This Addendum

| Priority | Action | Effort | Impact |
|----------|--------|--------|--------|
| P0 | Fix NLS tau_0: 1e-13 -> 1e-10 in landau.go and nls.go | S | H |
| P0 | Fix LiteratureSuperlattice Pr: 50 -> 22 uC/cm^2 | S | M |
| P0 | Fix FeCAP mode: charge-domain MAC, suppress DC non-idealities, CSA sensing | L | H |
| P1 | Add PUND measurement simulation | M | H |
| P1 | Add wake-up Pr(N) non-monotonic model | M | H |
| P1 | Add state-dependent C2C variation | S | H |
| P1 | Add retention decay model (power-law) | M | H |
| P1 | Add C(V) butterfly curve computation | M | M |
| P2 | Add I-V leakage model (Schottky/PF/FN) | M | M |
| P2 | Add AFE phase fraction parameter | S | M |
| P2 | Calibrate ADC energy model to 2025 data | S | M |
| P3 | Add NLS kinetic mode flag (Tier 1 envelope) | M | L |
| P3 | Add Gaussian (erf) Everett alternative | S | L |

---

*This addendum should be read alongside `PHYSICS_REALISM_AUDIT.md` and the master synthesis report at `.omc/research/sciomc-20260218/findings/master-synthesis.md`.*
