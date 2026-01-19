# Papers to Acquire - Priority Action List
**Updated 2026-01-19 | IronLattice Demo Improvements**

---

## CRITICAL - Corrupted Downloads (Need Replacement)

| File | Size | What It Should Be | Where to Get |
|------|------|-------------------|--------------|
| `Mayergoyz_IEEE_1986.pdf` | 16 bytes | "Mathematical Models of Hysteresis" - Preisach model foundations | IEEE Xplore (institutional access) |
| `IEEE_CIM_Survey_2023.pdf` | 244 bytes | "Compute-in-Memory: Recent Trends and Prospects" | IEEE Xplore |
| `Tour_In2Se3_ChemRxiv.pdf` | 60 bytes | Flash Joule Heating synthesis of In₂Se₃ | ChemRxiv or tour@rice.edu |

**Location:** `opensource/papers/09_CORRUPTED/`

---

## HIGH PRIORITY - Demo-Specific Papers

### For Demo 1 (Hysteresis) Improvements

| Paper | Citation | Why Needed | Search Terms |
|-------|----------|------------|--------------|
| **Mayergoyz Preisach Model** | IEEE Trans. Magnetics, Vol. 22, No. 5, 1986 | Stack-based voltage history, "wiping-out" property | "Mayergoyz Preisach 1986" |
| **Böscke HfO₂ Foundation** | Appl. Phys. Lett. 99, 102903 (2011) | HfO₂ ferroelectricity discovery, orthorhombic phase | DOI: 10.1063/1.3634052 |
| **Domain Wall Dynamics** | Various IEEE EDL | Grain-by-grain switching animation | "HZO domain wall dynamics" |

### For Demo 2 (Crossbar 30-Level) Improvements

| Paper | Citation | Why Needed | Search Terms |
|-------|----------|------------|--------------|
| **Oh et al. 32 Levels** | IEEE Electron Device Lett. 38(6), 2017 | **Scheme C incremental amplitude pulses** | "HfZrOx 32 levels conductance" |
| **1T1R FeFET Array** | IEDM/VLSI papers | Sneak path suppression architecture | "1T1R FeFET crossbar" |
| **IR Drop Modeling** | Various | Enhanced wire resistance modeling | "crossbar IR drop analysis" |

### For Demo 3 (MNIST 87%+) Improvements

| Paper | Citation | Why Needed | Search Terms |
|-------|----------|------------|--------------|
| **Jerry et al. 90% MNIST** | IEDM 2017 | **75ns pulse width optimization**, symmetric updates | "FeFET synapse MNIST 90%" |
| **QAT for Analog AI** | arXiv (multiple) | Quantization-aware training implementation | "quantization aware training analog" |
| **RUSD Algorithm** | Various | Sign-based gradient descent for limited precision | "sign SGD analog neural network" |

---

## MEDIUM PRIORITY - Enhancement Papers

| Paper | Purpose | Demo |
|-------|---------|------|
| Symmetric potentiation/depression curves | Better training convergence | Demo 3 |
| Preisach-NN self-calibration | Adaptive hysteresis modeling | Demo 1 |
| Noise robustness analysis methods | Accuracy vs. variation plots | Demo 3 |
| Phase-field TDGL numerical methods | Domain evolution animation | Demo 1 |

---

## NICE TO HAVE - Future Enhancements

| Paper | Purpose |
|-------|---------|
| Spiking neural networks on FeFET | Alternative inference mode |
| Wafer-scale FeFET integration | Manufacturing scalability |
| Weebit ReRAM comparison studies | Competitive positioning |
| Flash Joule Heating other materials | Dr. Tour's synthesis methods |

---

## WHERE TO GET PAPERS

### Free / Open Access
- **arXiv.org** - Many preprints available (already have 30+ from here)
- **Author websites** - external research institution, IBM Research, Intel Labs
- **ResearchGate** - Request directly from authors
- **Google Scholar** - Often links to free PDFs
- **Company white papers** - Weebit Nano, Intel, IBM (free)

### Need Institutional Access
- **IEEE Xplore** - Most device papers (CRITICAL papers above)
- **AIP Publishing** - Applied Physics Letters
- **Nature/Science** - Some foundational papers
- **ACM Digital Library** - CS/algorithm papers

### Contact Directly
- **Dr. Tour's Lab**: tour@rice.edu
  - Request In₂Se₃ paper and ferroelectric work
  - Mention IronLattice visualization project

---

## QUICK DOWNLOAD COMMANDS

```bash
# Check for corrupted files
find opensource/papers -name "*.pdf" -size -1k -exec ls -la {} \;

# Validate a downloaded PDF
pdftotext paper.pdf - | head -20

# Move corrupted files
mv corrupted.pdf opensource/papers/09_CORRUPTED/
```

---

## DEMO IMPROVEMENT MAPPING

| Demo | Critical Paper | Key Technique |
|------|----------------|---------------|
| Demo 1 | Mayergoyz 1986 | Discrete Preisach Model with history stack |
| Demo 2 | Oh et al. 2017 | Scheme C: V_prog[n] = V_start + n×ΔV |
| Demo 3 | Jerry et al. 2017 | 75ns pulse width for symmetric updates |

---

## STATUS TRACKING

- [x] arXiv papers downloaded (30+)
- [x] Papers validated with pdftotext
- [x] Corrupted files identified and isolated
- [ ] IEEE Xplore access obtained
- [ ] Mayergoyz paper downloaded
- [ ] Oh et al. Scheme C paper downloaded
- [ ] Jerry et al. 90% MNIST paper downloaded
- [ ] Dr. Tour contacted for In₂Se₃ paper

**Current coverage:** 40+ papers downloaded, 3 corrupted, ~10 more critical papers needed
