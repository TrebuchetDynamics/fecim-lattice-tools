# Iron Lattice Design Suite: Production Grade EDA

## The Open-Source Foundry

This is not a toy. This is not just a "translator" for AI weights.
**This is the open-source production tool for the Iron Lattice platform.**

### From Module 4 to Silicon
In **Module 4**, you modeled the *circuit behavior* (Schematics).
In **Module 6**, we generate the *manufacturing files* (Layouts) to build it.

**Module 6 turns your Module 4 circuits into physical reality.**

---

## 🔬 Scientific Validation

We do not make empty claims. Our technology is grounded in peer-reviewed research from the Tour Group at external research institution.

> **[View Full Reference List](REFERENCES.md)**

*   **The Physics:** "Flash In2Se3 for Neuromorphic Computing" (Shin et al., 2025) validates our **30-state analog memory**.
*   **The Manufacturing:** "Stoichiometric Engineering... by Flash-within-Flash" validates our **Capital Light** process.
*   **The Market:** "The Microchip Era Is About to End" (WSJ, Gilder 2025) validates the **Wafer Scale** vision.

---

## 🏗️ Build Actual Hardware

We enable the design of three distinct classes of next-generation silicon:

### 1. High-Density Storage (NAND Replacement)
Design multi-terabit non-volatile storage arrays.
*   **EDA Goal:** Optimize for **Retention** and **Density**.
*   **The Spec:** 10,000,000x lower energy than Flash. 90% lower voltage.
*   **Output:** GDSII layouts for dense 3D vertical strings.

### 2. High-Speed Memory (DRAM Replacement)
Design ultra-fast, restoration-free memory cost-optimized for caches.
*   **EDA Goal:** Optimize for **Speed** (10ns switching) and **Endurance** ($10^{12}$ cycles).
*   **The Spec:** Zero refresh cycles. Non-volatile.
*   **Output:** SPICE netlists for sense amplifiers and row drivers.

### 3. Neuromorphic GPUs (The "AI Killer")
Design massively parallel Compute-in-Memory accelerators.
*   **EDA Goal:** Maximize **Analog Precision** (30 states) and **Throughput**.
*   **The Spec:** Matrix-Vector Multiplication *inside* the array.
*   **Output:** Configurable connect-logic for the 30-state lattice.

---

## The Engineering Workflow

This suite guides you through the full semiconductor design lifecycle:

### Phase 1: Architecture (Tabs 1 & 3)
*   **Configure Physics:** Choose your superlattice composition (Storage Mode vs Compute Mode).
*   **Define Topology:** Set array dimensions (e.g., 256x256 tiles) and peripheral circuitry.

### Phase 2: Synthesis (Tab 1)
*   **Compile:** Map your logic (or data retention requirements) to the physical lattice.
*   **Quantize:** For AI, map weights to the 30 discrete conductance levels:
    *   `Conductance = 1.0 + (Level / 29) * 99.0` ($\mu S$)

### Phase 3: Validation (Tab 4)
*   **SPICE Simulation:** Run physics-accurate `ngspice` models to prove timing, power, and signal integrity before spending millions on fabrication.

### Phase 4: Tapeout (Tab 5)
*   **Export GDSII:** Generate the final geometric files for the standard CMOS foundry.
*   **Capital Light:** Ready for standard manufacturing lines.

---

## Why Open Source?

The Iron Lattice revolution is about **democratizing access** to post-silicon performance. By providing a production-grade EDA tool, we empower every engineer to design the future of:
*   **The Data Center** (Pizza-box sized supercomputers)
*   **The Edge** (Extending battery life by orders of magnitude)
*   **The Wafer** (Full wafer-scale integration)
