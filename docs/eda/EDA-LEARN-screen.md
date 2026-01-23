# FeCIM Design Suite: Universal Chip Design Tool

## The Open-Source Foundry

This is not just an AI compiler. This is not limited to neural networks.
**This is a universal EDA tool for designing FeCIM-based chips.**

### From Simulation to Silicon
In **Modules 1-5**, you modeled the *circuit behavior* (Physics, Arrays, Inference).
In **Module 6**, we generate the *manufacturing files* (Layouts) to build it.

**Module 6 turns your designs into physical reality.**

---

## Three Chip Types

### 1. Storage Chips (NAND Replacement)
Design high-density non-volatile storage arrays.
- **No AI involved** — pure storage like NAND Flash
- 30 levels per cell = 4.9 bits storage density
- Optimized for retention time and endurance
- **No weights needed** — data written by user/controller

```bash
go run ./cmd/eda-cli -mode storage -rows 256 -cols 256 -name storage_chip
```

### 2. Memory Chips (DRAM Replacement)
Design high-speed, zero-refresh memory.
- **No AI involved** — pure memory like DRAM
- 10ns access time, non-volatile
- Optimized for speed and bandwidth
- **No weights needed** — data written by CPU

```bash
go run ./cmd/eda-cli -mode memory -rows 128 -cols 128 -name memory_chip
```

### 3. Compute Chips (AI Accelerator)
Design analog compute-in-memory for neural networks.
- AI/ML inference acceleration
- Matrix-vector multiply in hardware
- **Weights optional** — can pre-program or load later

```bash
# With pre-trained weights
go run ./cmd/eda-cli -mode compute -input weights.json -rows 64 -cols 64

# Without weights (programmed later)
go run ./cmd/eda-cli -mode compute -rows 64 -cols 64 -name cim_chip
```

---

## 🔬 Scientific Validation

We do not make empty claims. Our technology is grounded in peer-reviewed research from the Tour Group at external research institution.

> **[View Full Reference List](REFERENCES.md)**

* **The Physics:** "Flash In2Se3 for Neuromorphic Computing" (Shin et al., 2025) validates our **30-state analog memory**.
* **The Manufacturing:** "Stoichiometric Engineering... by Flash-within-Flash" validates our **Capital Light** process.
* **The Market:** "The Microchip Era Is About to End" (WSJ, Gilder 2025) validates the **Wafer Scale** vision.

---

## 🏗️ Build Actual Hardware

We enable the design of three distinct classes of next-generation silicon:

| Mode | EDA Goal | Output Files |
|------|----------|--------------|
| **Storage** | Optimize retention, density | GDSII, DEF, Verilog |
| **Memory** | Optimize speed (10ns), endurance | SPICE, DEF, Verilog |
| **Compute** | Optimize analog precision, throughput | All formats + weights map |

---

## The Engineering Workflow

This suite guides you through the full semiconductor design lifecycle:

### Phase 1: Configure (Tab 1)
* **Select Mode:** Choose Storage, Memory, or Compute
* **Define Topology:** Set array dimensions (e.g., 256x256) and peripheral circuitry
* **Choose Technology:** SKY130, GF180MCU, or IHP_SG13G2

### Phase 2: Design Generation (Tab 1)
* **Generate Array:** Create cell assignments for all modes
* **[Compute only] Quantize:** Map weights to 30 discrete conductance levels:
    * `Conductance = G_min + (Level / 29) * (G_max - G_min)` (μS)

### Phase 3: Validation (Tab 5)
* **SPICE Simulation:** Run physics-accurate `ngspice` models to prove timing, power, and signal integrity

### Phase 4: Tapeout (Tab 6)
* **Export Files:** Generate Verilog, DEF, SPICE for OpenLane flow
* **OpenLane Integration:** Ready for SKY130 fabrication

---

## Why Open Source?

The FeCIM revolution is about **democratizing access** to post-CMOS performance. By providing a production-grade EDA tool, we empower every engineer to design the future of:
* **Storage** — Replace NAND Flash with 10,000,000x lower energy
* **Memory** — Replace DRAM with zero-refresh non-volatile memory
* **Compute** — Analog AI accelerators that outperform GPUs
