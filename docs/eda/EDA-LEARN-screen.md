# FeCIM Design Suite: Educational EDA Tool

## What This Tool Does

This is an **educational/research-grade EDA tool** for exploring FeCIM (Ferroelectric Compute-in-Memory) array designs. It generates design files (Verilog, DEF, SPICE) that can be used with open-source EDA flows.

**What it is:**
- A simulation and file generation tool
- Based on published research from Tour Lab at external research institution
- Educational and research-grade software
- Generates structural files for EDA integration

**What it is NOT:**
- A production EDA tool (requires extensive validation)
- A foundry or fab service
- Endorsed by or affiliated with Dr. Tour or external research institution
- Capable of building hardware (it generates design files)

---

## Three Operation Modes

### 1. Storage Mode (NAND-like)
Generate high-density non-volatile storage array designs.
- No neural network weights required
- 30 levels per cell (~4.9 bits/cell theoretical density)
- Optimized for retention and endurance parameters

```bash
go run ./cmd/eda-cli -mode storage -rows 256 -cols 256 -name storage_chip
```

### 2. Memory Mode (DRAM-like)
Generate high-speed memory array designs.
- No neural network weights required
- Configured for fast access times
- Non-volatile (no refresh needed)

```bash
go run ./cmd/eda-cli -mode memory -rows 128 -cols 128 -name memory_chip
```

### 3. Compute Mode (AI Accelerator)
Generate analog compute-in-memory array designs.
- Optional: Load pre-trained neural network weights
- Matrix-vector multiply structure
- Quantization to 30 discrete levels

```bash
# With pre-trained weights
go run ./cmd/eda-cli -mode compute -input weights.json -rows 64 -cols 64

# Without weights (programmed later)
go run ./cmd/eda-cli -mode compute -rows 64 -cols 64 -name cim_chip
```

---

## Technology Readiness

**Current Status: Research/Educational Grade**

Per Dr. Tour's November 2024 presentation:
- FeCIM technology is at **TRL 4** (component validation in lab)
- Production readiness requires TRL 7-8
- This tool generates **placeholder cells** for simulation

**What's Generated:**
- Verilog netlists with FeCIM cell instances
- DEF placement files with FIXED coordinates
- SPICE netlists for simulation
- JSON/CSV data exports

**What's NOT Generated:**
- Validated FeCIM cell libraries (don't exist in open PDKs)
- Production-ready GDSII (requires actual cell designs)
- Timing/power characterization (requires silicon data)

---

## Scientific Basis

Models are based on published literature, not validated against actual hardware:

**Device Physics (from literature):**
- HZO ferroelectric hysteresis (Preisach model)
- 30 discrete conductance states (per Tour Lab publications)
- Temperature-dependent behavior models

**EDA Integration:**
- OpenLane-compatible file formats
- SKY130/GF180MCU placeholder cell dimensions
- Standard Verilog/DEF/SPICE syntax

> **[View References](REFERENCES.md)** - Published papers our models are based on

---

## Limitations and Disclaimers

1. **No Hardware Validation**: Models are based on published data, not calibrated to actual devices
2. **Placeholder Cells**: No real FeCIM cells exist in open PDKs - we use dimensional placeholders
3. **Research Grade**: This is educational software, not production EDA tooling
4. **Performance Claims**: Energy/speed comparisons are from published literature, not independently verified
5. **No Affiliation**: This project is not affiliated with or endorsed by external research institution or Dr. Tour

---

## Design Workflow

### Phase 1: Configure
- Select operation mode (Storage, Memory, Compute)
- Set array dimensions
- Choose technology node (SKY130, GF180MCU, IHP)

### Phase 2: Generate
- Create cell assignments
- [Compute mode] Quantize weights to 30 levels
- Calculate programming voltages

### Phase 3: Export
- Generate Verilog netlist
- Generate DEF placement
- Generate SPICE netlist
- Export statistics (JSON/CSV)

### Phase 4: Integration (External)
- Use generated files with OpenLane
- Requires custom FeCIM cell library (not provided)
- Simulation only until real cells available

---

## Why Open Source?

This tool is open source to:
- Enable academic research on FeCIM design methodology
- Provide educational materials for CIM architecture courses
- Accelerate design exploration (not production)
- Demonstrate what a FeCIM EDA flow could look like

**Note:** Open-source does not mean production-ready. Significant development would be needed for actual tape-out.
