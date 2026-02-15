# About the Science (Learn More)

This section is a **single, unified entry point** for learning the scientific background behind FeCIM Lattice Tools.

**Honesty policy:** unless a number is explicitly tied to a dataset/test artifact, treat it as **model behavior** (not validated physics).

---

## 1) What is FeCIM?

**Ferroelectric Compute-in-Memory (FeCIM)** uses the same physical device state that stores memory (ferroelectric polarization **P**) to represent **analog conductance** for compute.

Core chain used across modules:
- Apply voltages (**DAC**) → devices source currents (**Ohm’s law**) → currents sum on wires (**Kirchhoff’s law**) → analog voltage sensed (**TIA**) → digitized (**ADC**).

See also:
- Module 1 (Hysteresis): `docs/hysteresis/hysteresis.physics.md`
- Module 2 (Crossbar): `docs/crossbar/reference/PHYSICS.md`

---

## 2) Physics building blocks used in this app

### Polarization switching (P–E hysteresis)
Module 1 models hysteresis using two families:
- **Preisach** (rate-independent hysteresis; minor loops / history dependence)
- **Landau–Khalatnikov (LK)** (energy-based dynamics; switching kinetics)

Key observables:
- **Pr** (remanent polarization) [µC/cm²]
- **Ec** (coercive field) [MV/cm]

### Thickness dependence (field vs voltage)
The device physics is set by **electric field** E [V/m].
UI voltages relate via film thickness **t**:
- **V ≈ E × t**
- **Vc ≈ Ec × t** (coercive voltage)

Therefore “safe read voltage” and “write voltage” are **thickness-dependent**.

---

## 3) Literature anchors (examples used as calibrated references)

These are included as *calibrated presets* and/or validation anchors in the repository.

- Park et al. (2015) HZO 10 nm P–E loop example (used for calibrated reference)
  - DOI: **10.1002/adma.201404531**
- Cheema et al. (2020) HZO superlattice 5 nm P–E loop example (used for calibrated reference)
  - DOI: **10.1038/s41586-020-2208-x**

Important: “calibrated to DOI X” means we fit preset parameters to match that reference curve. It is **not** a first‑principles prediction.

---

## 4) Transparency & reproducibility

- Honesty / caveats and claim hygiene:
  - `HONESTY_AUDIT.md`
- Full validation guide:
  - `docs/testing/TEST_GUIDE.md`
- Proof artifacts (regression JSON outputs):
  - `output/` (generated)

---

## 5) Where to go next (per module)

- Module 1 (Hysteresis): P–E loops, ISPP, FORC workflow
- Module 2 (Crossbar): IR drop, sneak paths, solver convergence
- Module 3 (MNIST): end-to-end inference + quantization/noise
- Module 4 (Circuits): DAC→device→TIA→ADC chain and voltage rules
- Module 5 (Comparison): scenario modeling with TRL caveats
- Module 6 (EDA): educational exports (not sign-off)

Use the Docs browser tree (left) to open module docs and references.
