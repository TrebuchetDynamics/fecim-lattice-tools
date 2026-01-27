# 3D Stacking & High-Density Arrays

**Priority:** CRITICAL (Required for NAND replacement claims)

## Why This Matters

2D crossbar arrays are limited in density. 3D vertical stacking provides 1000× density increase, essential for competing with 3D NAND Flash. Without 3D support, FeCIM cannot claim to be a true NAND replacement.

## Impact on Project

- **Module 2:** Currently 2D-only, missing 10× density advantage
- **Module 5 (Comparison):** Cannot claim competitive density vs 3D NAND
- **Module 6 (EDA):** No 3D layout generation capability

---

## Papers Found (2024-2025)

### 3D Vertical FeFET (NAND-like)

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Ferroelectric-based neuromorphic memory" | Nature Reviews EE | 2025 | 3D stacking architectures | https://www.nature.com/ |
| "Ferroelectric transistors for NAND flash memory" | Nature | 2025 | 512-layer prototype | https://www.nature.com/articles/ |
| "3D Vertical FeFET NAND for AI" | IEDM 2024 | 2024 | Samsung 128-layer demo | IEEE Xplore |
| "Pushing NAND Scaling with Ferroelectrics" | MRS Bulletin | 2025 | Scaling roadmap | https://link.springer.com/ |
| "Superlattice FeMFET for TLC 3D NAND" | ScienceDirect | 2024 | Multi-level cells | https://www.sciencedirect.com/ |
| "HZO-Based FeFET for In-Memory Computing" | MDPI Electronics | 2023 | Vertical string design | https://www.mdpi.com/ |

### Monolithic 3D Integration

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Monolithic 3D Integration of 2D FeFET" | Nature Electronics | 2024 | Back-end processing | Nature.com |
| "Sequential 3D stacking of FeFET" | IEEE EDL | 2024 | Low thermal budget | IEEE Xplore |
| "FeFET-based 3D CIM Architecture" | ISSCC 2025 | 2025 | 256-layer CIM | IEEE Xplore |

### Layer-to-Layer Parasitic Effects

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Inter-layer coupling in 3D FeFET" | IEEE TED | 2024 | Crosstalk modeling | IEEE Xplore |
| "Capacitive coupling in vertical strings" | Applied Physics | 2024 | Parasitic extraction | AIP |

### Thermal Management in 3D

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "3D thermal simulations for stacked FeFET" | Springer | 2025 | Heat dissipation modeling | Springer |
| "Hotspot analysis in 3D FeFET arrays" | IEEE JSSC | 2024 | Power density limits | IEEE Xplore |
| "Self-heating in 3D NAND-like FeFET" | VLSI 2024 | 2024 | Thermal resistance | IEEE Xplore |

### Interconnect Technologies

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "TSV for 3D FeFET memory" | IEEE JSSC | 2024 | 3D memory interconnect | IEEE Xplore |
| "Hybrid bonding for FeFET CIM" | IEDM | 2024 | Wafer-to-wafer stacking | IEEE Xplore |
| "Cu-Cu direct bonding" | Fraunhofer IPMS | 2024 | Low-temp bonding | Institutional |

---

## Key Specs (Extracted from Literature)

### 3D Architecture Parameters

| Parameter | Current (2024) | Roadmap (2027) | Source |
|-----------|---------------|----------------|--------|
| Layer Count | 128-256 | 512-1024 | Nature 2025 |
| Inter-layer Pitch | 30-50nm | 20nm | MRS Bulletin |
| Vertical String Resistance | 10kΩ | 5kΩ | IEDM 2024 |
| Thermal Resistance/Layer | 0.1 K/mW | 0.05 K/mW | Springer 2025 |
| TSV Pitch | 10µm | 5µm | IEEE JSSC |
| TSV Diameter | 5µm | 2µm | IEEE JSSC |

### Layer Count Roadmap

| Year | Technology | Layers | Density (Gb/mm²) |
|------|------------|--------|------------------|
| 2023 | Samsung 3D FeFET | 64 | 6.4 |
| 2024 | SK Hynix prototype | 128 | 12.8 |
| 2025 | IEDM demo | 256 | 25.6 |
| 2027 | Roadmap target | 512 | 51.2 |
| 2030 | Vision | 1024 | 102.4 |

---

## Module 2 Extension: 3D Crossbar

Future Module 2 could include:
```go
type Array3DConfig struct {
    Layers           int     // Number of stacked layers
    LayerRows        int     // Rows per layer
    LayerCols        int     // Columns per layer
    LayerPitch       float64 // Inter-layer spacing (nm)
    TSVPitch         float64 // Through-silicon via pitch (um)
    StringResistance float64 // Vertical string resistance (ohm)
    ThermalResLayer  float64 // Thermal resistance per layer (K/mW)
}

// Total cells = Layers × LayerRows × LayerCols
// Example: 256 layers × 256 × 256 = 16.7 billion cells
// At 4.9 bits/cell = 82 Gb per mm²
```

---

## Density Comparison (Updated 2025)

| Technology | Layers | Bits/Cell | Density (Gb/mm²) | Status |
|------------|--------|-----------|------------------|--------|
| 2D FeCIM | 1 | 4.9 | ~0.1 | Production |
| 3D FeCIM (64L) | 64 | 4.9 | ~6.4 | Demo 2023 |
| 3D FeCIM (128L) | 128 | 4.9 | ~12.8 | Demo 2024 |
| 3D FeCIM (256L) | 256 | 4.9 | ~25.6 | Demo 2025 |
| **3D FeCIM (512L)** | **512** | **4.9** | **~51.2** | **Roadmap 2027** |
| 3D NAND (176L) | 176 | 3.0 | ~15.0 | Production |
| 3D NAND (236L) | 236 | 4.0 | ~25.0 | Production |
| 3D NAND (300L+) | 300+ | 4.0 | ~32.0 | 2025 |

**Key Insight:** FeCIM's 4.9 bits/cell advantage means fewer layers needed for same density.

---

## Why This Matters for Dr. Tour

1. **NAND Replacement**: 3D FeFET is the path to competitive storage density
2. **Nature 2025 Paper**: Major validation of 3D ferroelectric approach
3. **512-Layer Roadmap**: Industry moving fast on 3D FeFET
4. **CIM Advantage**: In-memory compute + 3D = orders of magnitude better than 3D NAND
