# IronLattice Research Papers Library
**Organized by Category**  
**Last Updated:** 2026-01-18

## 📁 Directory Structure

### 01_Core_Materials/
HZO ferroelectrics, material physics, domain dynamics
- HZO discovery and optimization
- Ferroelectric domain wall physics  
- AlScN alternative materials
- Temperature resilience studies

### 02_Training_Algorithms/
Quantization-aware training, low-precision networks
- Quantization-aware training (QAT)
- Low-precision neural networks (5-bit)
- Variation-resilient training
- Analog AI hardware co-design

### 03_Simulation_Tools/
NeuroSim, CrossSim, IBM AIHWKit, FerroX
- Circuit-level simulation (NeuroSim, DNNNeuroSim)
- Crossbar simulation (CrossSim)
- IBM analog AI toolkit (AIHWKit)
- Ferroelectric simulation (FerroX, PEtra)

### 04_CIM_Architectures/
Crossbar arrays, compute-in-memory designs
- Crossbar non-idealities (IR drop, sneak paths)
- 3D FeFET architectures
- Hybrid CIM designs
- Neuromorphic hardware

### 05_2D_Materials/
In₂Se₃ and 2D ferroelectric semiconductors
- 2D ferroelectric materials review
- Context for Dr. Tour's work

### 06_Industry_Reports/
ITRS/IRDS roadmaps, wafer-scale integration
- IEEE technology roadmaps
- Wafer-scale integration studies
- Industry demonstrations (Tsinghua)

### 07_Reviews_Surveys/
Comprehensive reviews of memory technologies
- Flash vs emerging NVM comparison
- Memory technology surveys

### 08_Documentation/
Catalogs, guides, technical dossiers
- **TECHNICAL_DOSSIER.md** - Extracted specifications
- **IMPLEMENTATION_GUIDE.md** - Actionable recommendations
- **COMPREHENSIVE_ANALYSIS.md** - Full research synthesis
- **PAPERS_CATALOG.md** - Complete bibliography
- **PAPERS_NEEDED.md** - Manual acquisition list
- Download session summaries

### 09_CORRUPTED/
Files needing re-download
- IEEE_CIM_Survey_2023.pdf (244 bytes)
- Mayergoyz_IEEE_1986.pdf (16 bytes)
- Tour_In2Se3_ChemRxiv.pdf (60 bytes)

## 📊 Statistics (Updated 2026-01-19)

**Total Papers:** 40+ PDFs
**Valid Papers:** ALL REVALIDATED
**Corrupted:** 3 papers (in 09_CORRUPTED)
**Total Size:** ~120 MB
**Coverage:** 95%

**FIXED on 2026-01-19:** 11 papers with wrong content were redownloaded with correct arXiv versions.

## 🎯 Quick Access

**For Demo Fixes:**
- Demo 1 (Hysteresis): See `TECHNICAL_DOSSIER.md` Section 1 (Mayergoyz)
- Demo 2 (30 levels): See `TECHNICAL_DOSSIER.md` Section 2 (Scheme C)
- Demo 3 (90% MNIST): See `TECHNICAL_DOSSIER.md` Section 3 (75ns pulses)

**For Implementation:**
- `08_Documentation/IMPLEMENTATION_GUIDE.md`
- `08_Documentation/TECHNICAL_DOSSIER.md`

**For Research:**
- `08_Documentation/COMPREHENSIVE_ANALYSIS.md`
- `08_Documentation/PAPERS_CATALOG.md`

## 🔍 Finding Papers

Use grep to search across all papers:
```bash
cd /path/to/papers
grep -r "search term" . --include="*.md"
```

Or browse by category folder above.
