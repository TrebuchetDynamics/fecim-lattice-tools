# HfO₂-ZrO₂ (HZO) Material Parameters

## Overview

This document compiles experimentally measured ferroelectric parameters for Hafnium-Zirconium Oxide (HZO) thin films from literature, for use in Ferroelectric CIM simulations.

---

## Remanent Polarization (Pᵣ)

| Structure | Pᵣ (μC/cm²) | 2Pᵣ (μC/cm²) | Thickness | Annealing | Source |
|-----------|-------------|--------------|-----------|-----------|--------|
| TiN/HZO/TiN | 13-26 | 26-52 | - | ~400°C ALD | [General review] |
| TiN/HZO/TiN | 25 | 50 | - | - | [PMC 9740545](https://pmc.ncbi.nlm.nih.gov/articles/PMC9740545/) |
| Pt/HZO/Ni | 19 | 38 | 20 nm | 650°C, 30s | [Springer](https://link.springer.com/article/10.1007/s42341-024-00546-z) |
| HZO (waked-up) | >20 | >40 | 5 nm | - | [Nature Comms](https://www.nature.com/articles/s42005-022-00951-x) |
| HZO multilayer | 9 | 18 | - | - | [ACS Omega](https://pubs.acs.org/doi/10.1021/acsomega.4c10603) |
| **Record (ALA)** | 26.84 | **53.68** | - | ALA | [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S1359645425001478) |
| **Record (400°C)** | ~34.3 | **68.6-69** | - | 400°C | [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S1359645425001478) |
| ZrO₂ seed | 22.3 | 44.6 | - | - | Literature |

### Recommended Values for Simulation

| Parameter | Conservative | Typical | Optimized |
|-----------|--------------|---------|-----------|
| Pᵣ (μC/cm²) | 15 | 25 | 35 |
| 2Pᵣ (μC/cm²) | 30 | 50 | 70 |

---

## Coercive Field (Eᶜ)

| Structure | Eᶜ (MV/cm) | Coercive Voltage | Notes | Source |
|-----------|------------|------------------|-------|--------|
| Pt/HZO/Ni | 1.07 | - | 20nm, 650°C | [Springer](https://link.springer.com/article/10.1007/s42341-024-00546-z) |
| HZO multilayer | 1.2 | - | Standard | [ACS Omega](https://pubs.acs.org/doi/10.1021/acsomega.4c10603) |
| Pure HfO₂ | 1.5 | - | Higher than HZO | Literature |
| HZO (ALA) | **1.02** | - | Record low | [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S1359645425001478) |
| Thin HZO | - | -1.1V / +1.0V | Vᶜ values | [PMC 9740545](https://pmc.ncbi.nlm.nih.gov/articles/PMC9740545/) |
| Waked-up 5nm | - | ~1V | Low voltage | [Nature Comms](https://www.nature.com/articles/s42005-022-00951-x) |
| ZrO₂ seed | - | 2.7V | Higher | Literature |

### Recommended Values for Simulation

| Parameter | High Voltage | Standard | Low Voltage |
|-----------|--------------|----------|-------------|
| Eᶜ (MV/cm) | 1.5 | 1.2 | 1.0 |
| Vᶜ @ 10nm | 1.5V | 1.2V | 1.0V |

---

## Other Parameters

### Relative Permittivity (εᵣ)

| Material | εᵣ | Notes |
|----------|-----|-------|
| HZO (ALA) | **35.9** | Record high |
| HZO (typical) | 25-30 | Standard |
| HfO₂ (amorphous) | ~20 | Non-ferroelectric |

### Depolarization Field

- **Magnitude:** ~1.2 MV/cm (single-domain state)
- **Significance:** Close to coercive field, affects retention

### Film Thickness Range

| Application | Thickness | Notes |
|-------------|-----------|-------|
| Memory devices | 5-15 nm | Typical FeFET gate |
| Research | 10-20 nm | Common study range |
| Wake-up free | 5-10 nm | Ultrathin preferred |

---

## Superlattice-Specific Parameters

### HfO₂/ZrO₂ Superlattice (Verified Data)

| Configuration | Pᵣ (µC/cm²) | Eᶜ (MV/cm) | Endurance | DOI/Source |
|---------------|-------------|------------|-----------|------------|
| Epitaxial 20nm SL | ~15 | 1.4-1.6 | 10⁹ | [10.1038/s41467-025-61758-2](https://doi.org/10.1038/s41467-025-61758-2) |
| Epitaxial 100nm SL | ~33 | 0.85 | 10⁹ | [10.1038/s41467-025-61758-2](https://doi.org/10.1038/s41467-025-61758-2) |
| TiN/HZO-SL/TiN | Variable | Lower | **>5×10¹²** | [PMC11197553](https://pmc.ncbi.nlm.nih.gov/articles/PMC11197553/) |
| 20ns pulse optimized | Variable | - | **>10¹¹** | [10.1109/IRPS48227.2022.9764533](https://doi.org/10.1109/IRPS48227.2022.9764533) |
| La-doped 3D HZO | >33 (2Pᵣ) | - | **>10¹²** | [ResearchGate](https://www.researchgate.net/publication/378376565) |

### Endurance Comparison

| Material | Endurance (cycles) | Notes | Source |
|----------|-------------------|-------|--------|
| Standard HZO | 10⁴-10⁶ | Wake-up + fatigue limits | General literature |
| HZO FeFET | 10⁵ | Typical FeFET | PMC review 2024 |
| HZO capacitor | 10⁸ | MFM structure | PMC review 2024 |
| **HZO superlattice** | **>5×10¹²** | TiN electrodes | [PMC11197553](https://pmc.ncbi.nlm.nih.gov/articles/PMC11197553/) |

### Benefits of Superlattice vs. Solid Solution

1. **Phase Stabilization:** Better orthorhombic phase retention
2. **Lower Eᶜ:** FE-AFE competition reduces barrier (~30% reduction)
3. **Improved Linearity:** Moderate domain switching
4. **Higher Endurance:** Interface defect trapping - **demonstrated >10¹² cycles**

---

## Temperature Dependencies

| Parameter | Temperature Effect |
|-----------|-------------------|
| Pᵣ | Decreases with T |
| Eᶜ | Decreases with T |
| Switching speed | Increases with T |
| Retention | Degrades with T |

### Curie Temperature

- HfO₂ orthorhombic: > 450°C (estimated)
- Operation range: -40°C to 125°C (CMOS compatible)

---

## Simulation Parameter Set

### Default Configuration for Demo 1

```go
// HZO material parameters for simulation
type HZOMaterial struct {
    // Polarization
    Pr      float64 = 25e-6    // C/cm² → convert to SI
    Ps      float64 = 30e-6    // Saturation polarization

    // Field
    Ec      float64 = 1.2e6    // V/cm → 1.2 MV/cm

    // Dielectric
    Epsilon float64 = 30       // Relative permittivity

    // Film
    Thickness float64 = 10e-9  // 10 nm

    // Dynamics
    Tau     float64 = 1e-9     // Switching time constant (ns)
}
```

### Preisach Model Parameters

```go
// Distribution parameters for Preisach model
type PreisachParams struct {
    // Coercive field distribution
    Ec_mean  float64 = 1.2e6   // MV/cm center
    Ec_sigma float64 = 0.3e6   // Distribution width

    // Interaction field distribution
    Eu_mean  float64 = 0       // Centered
    Eu_sigma float64 = 0.5e6   // Interaction spread

    // Saturation
    Ps       float64 = 30e-6   // μC/cm²
}
```

---

## Sources

### With DOIs (Verified)

1. **Polarization Switching Kinetics in HZO** - [PMC9740545](https://pmc.ncbi.nlm.nih.gov/articles/PMC9740545/)
2. **Metal-Ferroelectric-Metal FeFET** - [DOI: 10.1007/s42341-024-00546-z](https://link.springer.com/article/10.1007/s42341-024-00546-z)
3. **Metastable Ferroelectricity in HZO** - [DOI: 10.1038/s42005-022-00951-x](https://www.nature.com/articles/s42005-022-00951-x)
4. **HZO Multilayers Reduced Wake-Up** - [DOI: 10.1021/acsomega.4c10603](https://pubs.acs.org/doi/10.1021/acsomega.4c10603)
5. **Ultra-high Pᵣ via Atomic Layer Annealing** - [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S1359645425001478)
6. **Enhancing ferroelectric stability in HfO₂/ZrO₂ superlattices** - [DOI: 10.1038/s41467-025-61758-2](https://doi.org/10.1038/s41467-025-61758-2) (Nature Commun. 2025)
7. **Correlation between Polarization and High Endurance** - [DOI: 10.1109/IRPS48227.2022.9764533](https://doi.org/10.1109/IRPS48227.2022.9764533) (IEEE IRPS 2022)
8. **HfO₂-based ferroelectric review (endurance)** - [PMC11197553](https://pmc.ncbi.nlm.nih.gov/articles/PMC11197553/) (PMC 2024)
9. **Nature Reviews Materials - HfO₂ fundamentals** - [DOI: 10.1038/s41578-022-00431-2](https://www.nature.com/articles/s41578-022-00431-2)
