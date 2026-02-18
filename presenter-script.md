# FeCIM Lattice Tools — Presenter Script

## Title
**From Hysteresis to EDA: A 7-Module Tour of Ferroelectric Compute-in-Memory**

## Recommended duration
8–10 minutes

---

## 0) Opening (30–45s)
Hello everyone, and welcome.

Today I’ll show you **FeCIM Lattice Tools**, an educational and research-oriented software platform for understanding ferroelectric compute-in-memory systems.

The tour goes from device physics to system-level design:
1. Hysteresis physics
2. Crossbar arrays
3. Neural-network inference behavior
4. Peripheral-aware circuit operation
5. Technology comparison
6. EDA/export workflow
7. Documentation and curriculum support

This tool is designed to improve intuition, test hypotheses, and support reproducible simulation workflows.

---

## 1) Module 1 — Hysteresis (1.5 min)
We start at the device level with ferroelectric hysteresis.

In this module we visualize polarization–electric-field behavior and explore material-dependent responses.

Key points to highlight:
- Multiple materials can be selected for comparison.
- We support practical operation modes, including waveform-driven and manual targeting.
- Write operation logic can be explored with **ISPP-style program/verify behavior**.
- The module helps connect physical state evolution to programmable levels used later in arrays.

Transition line:
> “Once we understand a single cell’s state evolution, we can scale that behavior to an array.”

---

## 2) Module 2 — Crossbar Arrays (1.5 min)
Now we move from one cell to many cells arranged in a crossbar.

This module lets us inspect:
- Current distribution across the array
- IR-drop effects
- Sneak-path behavior
- Architecture-dependent access constraints

Architectures discussed:
- **0T1R (passive crossbar)**
- **1T1R**
- **2T1R**

The core idea:
- Simpler structures are attractive for density and fabrication.
- Added selector control improves isolation and write/read selectivity.

Transition line:
> “After seeing array physics, the next step is: what does this mean for inference quality?”

---

## 3) Module 3 — MNIST / Inference Behavior (1–1.5 min)
Module 3 explores neural-network inference with quantized multi-level behavior.

What we demonstrate:
- Baseline behavior versus quantized/device-aware behavior
- Accuracy impact under non-ideal conditions
- Fast, intuitive input/output testing using digit-style inference examples
- Energy estimation for inference operations

This makes the trade-off visible:
- Device realism and quantization can affect accuracy,
- but provide insight into practical hardware constraints.

Transition line:
> “Now we connect array behavior and inference behavior to full read/write/compute circuits.”

---

## 4) Module 4 — Circuits + Peripherals (2 min)
This module is the system bridge: **READ, WRITE, and COMPUTE** with peripherals in the loop.

Signal chain emphasis:
- DAC drives voltages
- Crossbar produces currents
- TIA converts current to voltage
- ADC digitizes readout

What to explain during demo:
- Architecture-dependent write disturbance behavior
- Neighbor coupling effects during programming
- Why selector devices can reduce unintended updates
- How matrix-vector style operations map to current summation

Important message:
- This module helps users understand not only ideal equations, but also practical signal-chain constraints.

Transition line:
> “With operation-level behavior understood, we can compare this paradigm against conventional compute options.”

---

## 5) Module 5 — Technology Comparison (45–60s)
Module 5 provides a comparison view across compute options (e.g., CPU/GPU-style baselines versus ferroelectric CiM concepts).

Focus areas:
- Energy trends
- Throughput/efficiency intuition
- Relative positioning, not marketing claims

Presenter note:
Keep claims conservative and evidence-based; present this as a decision-support module, not a final silicon signoff claim.

Transition line:
> “Finally, can we take these ideas into a design-tool flow?”

---

## 6) Module 6 — EDA / Export Flow (1 min)
Module 6 is the forward-looking integration layer.

Purpose:
- Connect simulated array/device concepts to design-flow artifacts
- Explore exports and flow compatibility with open tooling
- Build toward stronger design handoff capability

Current status framing:
- Foundation is present and evolving.
- Peripheral and full-flow depth are active development areas.

Transition line:
> “To make this usable in real learning and research settings, documentation and curriculum matter.”

---

## 7) Module 7 — Documentation & Curriculum (45–60s)
Module 7 is about structured learning and onboarding.

Goals:
- Help students and researchers progress from physics to systems thinking
- Provide guided paths, references, and practical exercises
- Improve reproducibility and communication of results

---

## Closing (30–45s)
FeCIM Lattice Tools is built to make ferroelectric compute-in-memory concepts tangible:
- from hysteresis physics,
- to array effects,
- to inference,
- to peripheral-aware circuits,
- to design-flow exploration,
- and finally to teachable documentation.

Thank you for your time.
If useful, I can also provide:
- a 3-minute compressed version,
- a Spanish version,
- or a live-demo cue sheet with exact click-by-click flow.

---

## Optional live demo cue sheet (quick)
1. Module 1: pick material, show loop behavior and target programming concept
2. Module 2: switch 0T1R → 1T1R → 2T1R; discuss sneak-path/selection
3. Module 3: run quick inference example; compare behavior
4. Module 4: demonstrate read/write/compute and peripheral chain
5. Module 5: show comparison dashboard
6. Module 6: show EDA/export entry points
7. Module 7: show docs/curriculum structure
