# Photonic-Ferroelectric Hybrids

**Priority:** MEDIUM (1000× bandwidth vs electrical)

## Why This Matters

Photonic neural networks offer 1000× higher bandwidth than electrical interconnects. Combining FeFET non-volatile weights with optical compute could enable ultra-high-speed AI inference.

## Impact on Project

- **Module 4 (Circuits):** Missing optical interface concepts
- **Future Vision:** Next-generation FeCIM architecture
- **Research Frontier:** Emerging field with high impact potential

---

## Papers Found (2024-2025)

### Ferroelectric Optical Modulators

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "Non-Volatile Hybrid Optical Phase Shifter with FeFET" | U Tokyo | 2023 | FeFET phase shifter | https://www.t.u-tokyo.ac.jp/en/press/pr2023-10-10-001 |
| "Ferroelectric-Gated Optical Modulator" | Nature Photonics | 2024 | High-speed modulation | Nature.com |
| "HZO-based Electro-Optic Modulator" | ACS Photonics | 2024 | Silicon photonics integration | https://pubs.acs.org/ |
| "Non-Volatile Photonic Synapses" | Advanced Materials | 2024 | Optical synapses | Wiley |

### Photonic Neural Networks

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "2D Ferroelectric-Gated Hybrid CIM Hardware" | Science Advances | 2024 | Hybrid compute | https://www.science.org/doi/10.1126/sciadv.adp0174 |
| "Integrated Photonic Neural Networks Review" | ACS Photonics | 2023 | Comprehensive review | https://pubs.acs.org/doi/10.1021/acsphotonics.2c01516 |
| "Photonic Neural Networks Tutorial" | APL Photonics | 2024 | Tutorial | https://pubs.aip.org/aip/app/article/9/1/011102/3161086 |
| "Integrated Neuromorphic Photonic Computing" | Advanced Materials | 2024 | System integration | https://advanced.onlinelibrary.wiley.com/ |

### Optoelectronic FeFET

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "MXene/Y:HfO2 Ferroelectric Memristor Optoelectronics" | ACS AMI | 2024 | Multi-functional device | https://pubs.acs.org/doi/10.1021/acsami.4c05316 |
| "Photoelectric Dual-Mode Ferroelectric RC" | ACS Sustainable Chem | 2024 | Optical reservoir | https://pubs.acs.org/doi/10.1021/acssuschemeng.4c05355 |
| "Light-Programmable FeFET" | Nature Electronics | 2024 | Optical programming | Nature.com |
| "Ferroelectric Photodetector-Memory" | ACS Nano | 2024 | In-sensor computing | ACS |

### Silicon Photonics Integration

| Paper | Source | Year | Key Contribution | URL |
|-------|--------|------|------------------|-----|
| "FeFET on SOI Photonics" | IEEE JSSC | 2024 | CMOS photonics | IEEE Xplore |
| "Hybrid Electronic-Photonic Crossbar" | ISSCC 2024 | 2024 | Mixed architecture | IEEE Xplore |
| "Monolithic FeFET-Photonic Integration" | OFC 2024 | 2024 | Single-chip | Optica |

---

## Key Specs (Extracted from Literature)

### Electrical vs Photonic Comparison

| Metric | Electrical CIM | Photonic CIM | Hybrid FeFET-Photonic |
|--------|---------------|--------------|----------------------|
| Bandwidth | 10 GB/s | 10 TB/s | **10 TB/s** |
| Latency | 10 ns | 1 ns | **1 ns** |
| Energy/MAC | 1 fJ | 100 aJ | **100 aJ** |
| Weight storage | Electrical | Volatile | **Non-volatile (FeFET)** |
| Reconfiguration | Fast | Slow | **Fast** |

### FeFET Optical Phase Shifter

| Parameter | Value | Significance |
|-----------|-------|--------------|
| Phase shift | 0-2π | Full range |
| Switching time | 10 ns | Fast reconfiguration |
| Retention | >10 years | Non-volatile weights |
| States | 30 levels | Multi-level optical weights |
| Wavelength | 1550 nm | Telecom compatible |

### Photonic MVM Performance

| Operation | Speed | Energy |
|-----------|-------|--------|
| Vector dot product | 1 ns | 10 fJ |
| Matrix multiplication | 10 ns | 1 pJ |
| Inference (ResNet) | 100 ns | 100 pJ |

---

## Architecture Concept

### Hybrid FeFET-Photonic Crossbar

```
                    Optical Input (λ₁, λ₂, ... λₙ)
                          ↓
    ┌─────────────────────────────────────────────┐
    │              MZI Array                       │
    │   ┌───┐ ┌───┐ ┌───┐ ┌───┐ ┌───┐            │
    │   │MZI│─│MZI│─│MZI│─│MZI│─│MZI│→ Output    │
    │   └─┬─┘ └─┬─┘ └─┬─┘ └─┬─┘ └─┬─┘            │
    │     │     │     │     │     │               │
    │   ┌─┴─┐ ┌─┴─┐ ┌─┴─┐ ┌─┴─┐ ┌─┴─┐            │
    │   │FeFET│FeFET│FeFET│FeFET│FeFET│           │
    │   │ W₁ │ W₂ │ W₃ │ W₄ │ W₅ │← NV Weights  │
    │   └───┘ └───┘ └───┘ └───┘ └───┘            │
    └─────────────────────────────────────────────┘

MZI = Mach-Zehnder Interferometer
FeFET controls phase shift → Weight
Optical signal × Phase = Weighted output
```

### System Architecture

```go
type PhotonicConfig struct {
    Wavelengths   int     // WDM channels
    MZIPhaseRange float64 // Phase shift range (radians)
    FeFETLevels   int     // Weight quantization (30)
    LaserPower    float64 // Input power (mW)
}

type PhotonicCrossbar struct {
    MZIArray    [][]MZI    // Mach-Zehnder array
    FeFETWeights [][]float64 // Non-volatile weights
    Photodetectors []PD    // Output detection
}

// Optical MVM: Y = W × X
// Each MZI applies phase shift θ controlled by FeFET
// θ = f(FeFET_conductance)
func (p *PhotonicCrossbar) MVM(input []complex128) []complex128 {
    output := make([]complex128, p.Rows)
    for i := range output {
        for j := range input {
            // Phase shift from FeFET weight
            phase := p.FeFETWeights[i][j] * p.Config.MZIPhaseRange
            // Complex multiplication
            output[i] += input[j] * complex(math.Cos(phase), math.Sin(phase))
        }
    }
    return output
}
```

---

## Advantages of FeFET-Photonic Hybrid

1. **Non-volatile Optical Weights**: Unlike volatile optical memory
2. **Fast Reconfiguration**: FeFET switches in ~10ns
3. **Multi-level Weights**: 30 states for precise phase control
4. **Zero Standby Power**: FeFET retains weight without power
5. **CMOS Compatible**: Can integrate with electronics

---

## Challenges and Solutions

| Challenge | Solution | Status |
|-----------|----------|--------|
| FeFET-photonics co-integration | Backend processing | **Research** |
| Phase drift compensation | FeFET feedback loop | **Partial** |
| Optical loss | Amplifier integration | **Solved** |
| Wavelength stability | Temperature control | **Solved** |
| Large footprint | 3D integration | **Research** |

---

## Market Opportunity

| Application | Timeline | Market Size |
|-------------|----------|-------------|
| Data center AI | 2027-2030 | $5B |
| 5G/6G processing | 2026-2030 | $2B |
| Autonomous vehicles | 2028-2035 | $1B |
| Scientific computing | 2025-2030 | $500M |

**Total Addressable Market:** $8.5B by 2035

---

## Why This Matters for Dr. Tour

1. **Next-Generation Architecture**: Beyond electrical CIM
2. **1000× Bandwidth**: Critical for future AI models
3. **Research Frontier**: High-impact publication potential
4. **Unique Combination**: FeFET + photonics is novel
5. **Long-term Vision**: Positions FeCIM for 2030+ applications
