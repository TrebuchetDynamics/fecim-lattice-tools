# FeCIM Lattice Tools — 2–3 min demo narration script

This script is written to match the screenshots in `docs/demo/frames/`.

## 0. Title / positioning

**FeCIM Lattice Tools** is a research-grade, auditable ferroelectric compute-in-memory (FeCIM) simulator and desktop instrument.
It combines: hysteresis physics, crossbar array solvers with IR-drop, circuit-level READ/WRITE/COMPUTE, MNIST inference integration, and an EDA export pipeline.
Every physics parameter is explicitly defined and validated by automated tests.

---

## Frame 001 — Module 1: Hysteresis (Preisach + Landau–Khalatnikov)

**What you’re seeing:** a polarization–electric-field (P–E) hysteresis instrument.

**Key points:**
1. Two physics engines are supported:
   - **Preisach** for multi-level analog state behavior.
   - **Landau–Khalatnikov (L‑K)** for time-domain switching dynamics.
2. The module exposes write/read demos using ISPP-style programming to a discrete target level, with guards and overshoot recovery.

File: `frames/frame_001_hysteresis.png`

---

## Frame 002 — Module 2: Crossbar

**What you’re seeing:** crossbar-array simulation with coupling modes and solver validation.

**Key points:**
1. Models array-level effects: IR drop, sneak paths, and coupling tiers from idealized to full nodal solutions.
2. Validation is research-grade: Kirchhoff-law residual checks and deterministic regression tests.

File: `frames/frame_002_crossbar.png`

---

## Frame 003 — Module 3: MNIST

**What you’re seeing:** end-to-end inference integration that ties array physics to ML accuracy.

**Key points:**
1. Runs MNIST inference with quantization and optional physics-aware noise.
2. Bridges the simulator stack: hysteresis/material parameters → crossbar MVM → classifier output, enabling systems-level tradeoffs.

File: `frames/frame_003_mnist.png`

---

## Frame 004 — Module 4: Circuits (READ/WRITE/COMPUTE)

**What you’re seeing:** circuit-level array simulation (0T1R / 1T1R architectures) treated as a research instrument.

**Key points:**
1. READ/WRITE/COMPUTE paths are validated against Kirchhoff constraints and internal invariants.
2. Includes write-verify behavior and parity checks between GUI dispatch and headless physics harness.

File: `frames/frame_004_circuits.png`

---

## Frame 005 — Module 6: EDA

**What you’re seeing:** an export and verification pipeline for downstream toolchains.

**Key points:**
1. Generates SPICE / Verilog / DEF / LEF / Liberty artifacts from array specs.
2. Export correctness is tested for parseability (e.g., ngspice/Yosys/OpenLane-style constraints) and parameter preservation.

File: `frames/frame_005_eda.png`

---

## Frame 006 — Module 5: Comparison

**What you’re seeing:** comparative modeling across memory/computing baselines.

**Key points:**
1. A structured place to compare FeCIM operating points versus other memory/compute paradigms.
2. Intended for “claims-to-evidence” workflows: comparisons must trace back to explicit models and validated parameters.

File: `frames/frame_006_comparison.png`

---

## Frame 007 — Module 7: Documentation browser

**What you’re seeing:** built-in documentation navigation.

**Key points:**
1. Keeps scientific context close to the instrument: equations, assumptions, and limitations are visible in-app.
2. Supports curriculum/module organization so reviewers can reproduce workflows quickly.

File: `frames/frame_007_docs.png`

---

## Closing

FeCIM Lattice Tools is engineered to be citable and reproducible: deterministic runs, automated regression artifacts, parity tests between GUI and headless harnesses, and an export pipeline with toolchain checks.

Next steps for a reviewer: run `go test ./...` and then execute the headless regression scripts to reproduce published outputs.
