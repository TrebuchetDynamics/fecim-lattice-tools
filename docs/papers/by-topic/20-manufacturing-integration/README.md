# Manufacturing & Process Integration

**Priority:** CRITICAL (Must have for production claims)

## Why This Matters

FeCIM technology must integrate with existing CMOS manufacturing flows. Without process integration documentation, we cannot claim "manufacturable" or discuss foundry compatibility.

## Impact on Project

- **Module 6 (EDA):** Missing process corners (slow/typical/fast)
- **Module 4 (Circuits):** No BEOL thermal constraints
- **Documentation:** Cannot claim "manufacturable" without process specs

---

## Papers Found (2024-2025)

### BEOL/FEOL Integration

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Ferroelectric Memories" | Fraunhofer IPMS | 2024 | Automotive-grade process flow | Institutional |
| "Ferroelectric materials, devices, and chips" | Sci China Info Sci | 2025 | 200mm/300mm wafer-scale | Institutional |
| "Recent Progress in Emerging 2D Ferroelectrics" | Advanced Materials | 2025 | 2D FE fabrication | https://onlinelibrary.wiley.com/ |
| "Hafnia-based FeFET: A promising device for memory" | Materials Today Advances | 2025 | Complete CMOS integration | https://www.sciencedirect.com/ |

### Sub-5nm Node Scaling

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Recent advances in ferroelectric capacitors" | Nano Convergence | 2025 | Future node compatibility | https://nanoconvergencejournal.springeropen.com/ |
| "HfO2-Based FeFET for Next-Generation" | IEEE JAP | 2024 | Interface engineering | IEEE Xplore |
| "Ferroelectric HfZrO for Next-Gen Memory" | Advanced Functional Materials | 2024 | 10nm node integration | https://onlinelibrary.wiley.com/ |

### ALD Process Control (NEW)

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Precursor Purge Time Effects on PE-ALD HZO" | ACS Omega | 2025 | Process optimization | https://pubs.acs.org/doi/10.1021/acsomega.5c01112 |
| "Ferroelectric Enhancement via Bottom Electrode Oxidation" | ACS AEM | 2024 | Interface engineering | https://pubs.acs.org/doi/10.1021/acsaelm.3c01502 |
| "HfO2/ZrO2 Multilayer Ferroelectricity Modulation" | Materials Letters | 2022 | Superlattice control | https://www.sciencedirect.com/ |
| "Cocktail Precursor ALD for HZO" | ACS AMI | 2025 | Novel deposition | https://pubs.acs.org/doi/10.1021/acsami.4c21964 |
| "Epitaxial ALD of Ferroelectric HZO" | Adv Functional Materials | 2024 | Crystalline quality | https://onlinelibrary.wiley.com/ |

### Doping Optimization

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| La, Y, Si doping studies | ACS/APL | 2024 | Threshold voltage tuning | Various |
| "Doping Engineering for FeFET" | Nature Electronics | 2024 | Multi-element doping | Nature.com |

### Thermal Budget

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Ferroelectric-based neuromorphic memory" | Nature Reviews EE | 2025 | Low-temp processing for 3D | https://www.nature.com/ |
| "Low-Temperature Crystallization of HZO" | ACS Applied Materials | 2024 | <400°C processing | https://pubs.acs.org/ |

---

## Key Specs (Extracted from Literature)

### Process Parameters

| Parameter | Typical Value | Range | Source |
|-----------|---------------|-------|--------|
| ALD Temp | 280°C | 200-350°C | ACS Omega 2025 |
| Anneal Temp | 500°C | 400-600°C | Multiple |
| HZO Thickness | 10nm | 5-20nm | Nano Convergence |
| Hf:Zr Ratio | 1:1 | 0.5:1 - 1.5:1 | ACS AMI 2025 |
| Electrode | TiN | TiN, TaN, W | Adv Func Mat |

### Process Integration Checklist

- [x] ALD precursor sequences (HfO₂/ZrO₂ cycles) - Documented
- [x] Crystallization anneal temperatures (400-600°C) - Confirmed
- [x] Interface layer requirements (SiO₂, TiN) - TiN preferred
- [ ] Process corner definitions (SS/TT/FF) - Needs fab data
- [ ] Thermal budget constraints for BEOL integration - Partial

---

## Module 6 Integration

Extract these parameters for EDA tool:
```go
type ProcessConfig struct {
    Node           string  // "SKY130", "GF180", "IHP_SG13G2"
    AnnealTemp     float64 // Crystallization temperature (°C)
    ThermalBudget  float64 // Max allowed (°C × seconds)
    InterfaceLayer string  // "SiO2", "TiN", "none"
    Corner         string  // "slow", "typical", "fast"
    HZOThickness   float64 // nm
    HfZrRatio      float64 // 1.0 = equal
}
```

---

## Industry Status (2025)

| Company | Node | Status | Notes |
|---------|------|--------|-------|
| GlobalFoundries | 22FDX | Production | FeFET option |
| Samsung | 14nm | Development | FinFET FeFET |
| TSMC | 16nm | Research | HZO integration |
| Fraunhofer IPMS | 130nm | Production | Automotive qualified |
| SK Hynix | 3D NAND | Development | 512-layer target |
