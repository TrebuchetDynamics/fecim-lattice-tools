# Literature Addendum -- February 2026

**Date:** 2026-02-18
**Session:** sciomc-20260218 (7-stage parallel research synthesis)
**Relation to:** `crossbar-circuits-literature-review-2025.md` (2026-02-14)

---

## Purpose

This addendum documents new papers, parameter updates, and findings identified during the February 18, 2026 research session that were not covered in the February 14 crossbar circuits literature review or earlier project research.

---

## 1. New Papers Found in This Research Session

### 1.1 HZO Material Properties and Landau Theory

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| Aturi et al., "FEFETs review" -- Adv. Electron. Mater. | 2026 | 10.1002/aelm.202500402 | Comprehensive FEFET review; 2Pr range 20-55 uC/cm^2; deposition temperature scaling |
| Ahn et al., "HZO nanodots scaling" -- Adv. Funct. Mater. | 2026 | 10.1002/adfm.202511839 | Domain wall velocity 1.2-2.3 m/s in HZO nanodots; Ea=6.65 MV/cm for 7 nm nanodots |
| RSC J. Mater. Chem. C, "Voltage-driven endurance" | 2026 | d5tc03491d | Voltage-driven endurance mechanisms in HZO |
| OAE Microstructures, "HfO2 applications review" | 2025 | microstructures.2025.32 | Comprehensive review; phase discrimination by STEM at grain level |
| Unified Phase-Field Framework (arXiv preprint) | 2026 | arXiv:2602.13959 | Coupled FE/AFE/dielectric phases in HZO; new Landau coefficients |
| Landau modeling HZO/ZrO2 (arXiv preprint) | 2025 | arXiv:2501.05312 | Alternative Landau parameter set for HZO |
| PMC 11983163, HfO2-ZrO2 Multilayers | 2025 | PMC11983163 | La-doped multilayer wake-up comparison; 10x faster wake-up for Zr-bearing films |
| RSC Nanoscale, HZO superlattice stress | 2025 | d4nr05053c | In-plane stress independent of layer count at scale |
| ACS AMI, La-doped HZO >10^12 endurance | 2024 | ResearchGate 378376565 | La-doped 3D-trench MFM >10^12 cycles -- current endurance record |
| ACS Nano, InGaZnO/HZO FeFET arrays | 2025 | 10.1021/acsnano.5c14197 | FeFET array demonstration with HZO |
| Lehninger et al., HfO2 "Game-Changer" review | 2025 | 10.1002/aelm.202400686 | Comprehensive HZO review; retention data |
| PMC 12254504, Epitaxial HfO2/ZrO2 superlattice | 2025 | PMC12254504 | Epitaxial superlattice stability to 100 nm; Ec~0.85 MV/cm |

### 1.2 Preisach Model and FORC Analysis

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| Dynamic Imprint and Recovery in Hf0.2Zr0.8O2 -- Electronics | 2025 | 10.3390/electronics14234593 | FORC-based imprint characterisation; Preisach density shift |
| Sci. Reports, Preisach parameter extraction for fluorite FEs | 2021 | 10.1038/s41598-021-91492-w | Automated Preisach fitting; quantitative FORC parameters for 10 nm HZO |
| Enhanced reliability of HZO with FORC -- JJAP | 2025 | 10.35848/1347-4065/ada163 | FORC reliability assessment of HZO capacitors |
| Phys. Rev. X, Theoretical lower limit of Ec in hafnia | 2025 | PhysRevX.15.021042 | Deep-learning multiscale; NLS/KAI Ec regimes; thickness transition |
| EKAI model for FeFETs -- PMC10934155 | 2024 | PMC10934155 | Extended KAI with grain angle; EBSD-calibrated switching kinetics |
| NLS switching kinetics in HZO -- PMC9740545 | 2022 | PMC9740545 | Quantitative NLS parameters: Ea=10-15 MV/cm, tau_inf=10^-10 s |
| Thickness-driven NLS-to-KAI transition -- Adv. Funct. Mater. | 2025 | 10.1002/adfm.202511380 | Confirmed thickness-driven kinetics transition in wurtzite FEs |
| Multibit HfZrO memcapacitor -- Adv. Funct. Mater. | 2025 | 10.1002/adfm.202531011 | 8-9 stable capacitance states; >10^5 s retention |
| Stochasticity from plasma damage in FE devices -- PMC12397007 | 2025 | PMC12397007 | Process variability dominates over thermal noise |

### 1.3 ISPP and MLC Programming

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| DCC 16-level FeFET -- Science Advances (PMC 11160465) | 2024 | PMC11160465 | Single-pulse 16-level programming via displacement current control |
| HZO FTJ 128-state synapse -- ACS Appl. Mater. Interfaces | 2025 | 10.1021/acsami.5c01547 | 128 states; 2.75% C2C variation; state-of-the-art benchmark |
| FTJ game-changer review -- APL Machine Learning | 2025 | AIP aml 3(2) 020902 | Comprehensive FTJ review; 32+ states |
| Recent advances ferroelectric IMC -- Nano Convergence | 2025 | 10.1186/s40580-025-00520-2 | Comprehensive review; MoS2 FeFET >1000 states |
| FTJ crossbar annealing -- Adv. Intelligent Systems | 2025 | 10.1002/aisy.202500817 | 48x48 crossbar; 8-level uniform programming |
| Ferroelectric NAND Bayesian NN -- Nature Communications | 2025 | s41467-025-61980-y | ISPP for probabilistic weight control |
| Pulse-modulated FeRAM -- Adv. Funct. Mater. | 2025 | 10.1002/adfm.202415511 | Ramp rate control reduces overshoot |
| MLC FeFET program/verify -- Jpn. J. Appl. Phys. | 2025 | 2025JaJAP..64eSP09V | 7-state FeFET crossbar; 0.05V step PV-FS |
| L-ISPP pulse optimisation -- ScienceDirect | 2024 | S1567173924001755 | Logarithmic ISPP improves linearity |

### 1.4 FeCAP Crossbar and Charge-Domain CIM

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| Imec + Georgia Tech, Non-destructive FeCAP readout | 2024-2025 | imec-int.com article | >10^11 read cycles; memory window 8.7 at 0V; read/write endurance decoupled |
| Yeo et al., Nonlinear SAR ADC for FeCAP -- IEEE SSL | 2024 | 10.1109/LSSC.2024.3361953 | Power-of-two nonlinear SAR; zero static power FeCAP macro |
| Bhardwaj et al., Capacitive IMC device-to-systems -- Adv. Intell. Discovery | 2025 | 10.1002/aidi.202500143 | 0.24 fJ/op at 1 MHz; >1 Gbit FeCAP arrays |
| Charge-domain CAM -- Nature Communications | 2025 | s41467-025-63190-y | >10x lower sensing overhead for in-memory search |
| ADC-less CIM (HCiM) -- arXiv | 2024 | arXiv:2403.13577 | Ternary comparators; 12-28x energy vs ADC; 3.5% accuracy cost |
| Memristor-based adaptive ADC -- Nature Communications | 2025 | s41467-025-65233-w | 15.1x energy improvement; nonlinear ramp from memristor column |
| 2T1R1C cells, capacitors reused as CDAC -- IEEE TVLSI | 2025 | 10.1109/TVLSI.2025.3539826 | Cell capacitors double as SAR ADC reference |
| FTJ 60-state device -- Adv. Intelligent Systems | 2024 | 10.1002/aisy.202300554 | 60 programmable states; nonlinearity >1,100 |
| 1763.9 TOPS/W FeFET ACIM chip | 2024 | ResearchGate 400295516 | 520 kb scalable FeFET CIM chip |

### 1.5 Switching Kinetics

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| Sub-nanosecond Si:HfO2 FeFET switching -- Nano Letters | 2023 | 10.1021/acs.nanolett.2c04706 | 300 ps complete switching; 11 decades of time |
| o-t-o switching pathway -- Nature Communications | 2025 | s41467-025-63298-1 | Orthorhombic-tetragonal-orthorhombic phase transition pathway |
| Field-induced structural dynamics HfxZr1-xO2 -- Adv. Materials | 2025 | 10.1002/adma.202510930 | In-situ GIXRD; composition-dependent switching |
| DW nucleation pathways in hafnia -- arXiv | 2023 | arXiv:2311.17290 | O-atom-mediated cross-unit-cell switching; two DW types |
| Multiple-order-parameter DW in hafnia -- PNAS | 2024 | 10.1073/pnas.2406316122 | Complex DW involves polarisation AND tetragonality reversal |
| GNN surrogate for phase-field HZO -- PMC11059552 | 2024 | PMC11059552 | MARE=4.24%, million-fold speedup vs direct phase-field |
| Wake-up DW evolution in HZO -- PMC11789571 | 2025 | PMC11789571 | 90 to 180 degree DW transition; Pr 12 to 28 uC/cm^2 |
| Wake-up mechanism vs Hf content -- PMC12616602 | 2025 | PMC12616602 | Thickness- and composition-driven mechanism shift |
| Jiles-Atherton compact model for HZO -- Adv. Electron. Mater. | 2025 | 10.1002/aelm.202400840 | Physics-based J-A model; frequency-dependent hysteresis |
| Depolarization HZO vs AlScN -- JAP | 2024 | 10.1063/5.0207397 | HZO: 16% domain reversal vs AlScN: <1% |
| Depolarization-controlled reservoir computing -- ACS AMI | 2025 | 10.1021/acsami.5c00213 | 15 nm HZO optimal for reservoir dynamics |
| Oxygen vacancy dynamics in switching -- PMC12368994 | 2025 | PMC12368994 | VO dynamics during switching modes |

### 1.6 CIM Inference and Benchmarking

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| NeuroSim V1.5 -- arXiv | 2025 | arXiv:2505.02314 | Current benchmark framework; FeFET >5 bits/cell; GPU PyTorch |
| CrossSim V3.1 -- Sandia (GitHub) | 2025 | github.com/sandialabs/cross-sim | Leading open-source CIM accuracy simulator; PyTorch/Keras integration |
| Hardware-aware training -- Nature Communications | 2023 | s41467-023-40770-4 | Noise-aware training recovers 2-4 pp accuracy |
| FAST simulator -- Science China | 2025 | 10.1007/s11432-024-4240-x | End-to-end training with IR-drop, variation, stuck-at faults |
| HW-aware quantisation -- ACM TODAES | 2023 | 10.1145/3569940 | ADC/weight co-optimisation framework |
| Ferroelectric memristor RC -- RSC J. Mat. Chem. C | 2026 | d5tc03983e | 98.71% Hand-MNIST reservoir computing |
| FeCAP array MNIST 96.68% -- ScienceDirect | 2025 | S2211285525003702 | Fabricated FeCAP chip measurement |
| 2D FE-gated CIM -- Science Advances | 2024 | 10.1126/sciadv.adp0174 | 99.8% dynamic tracking; 263x GPU power improvement |
| FTJ CNN CIFAR-10 -- ACS AMI | 2025 | 10.1021/acsami.5c00740 | ~92% CIFAR-10 on FTJ hardware |
| Nature Electronics FeRAM 0.24 fJ/op | 2025 | s41928-025-01454-7 | Lowest reported ferroelectric CIM energy |

### 1.7 Measurement Workflows and Tools

| Title / Description | Year | DOI / URL | Key Contribution |
|---------------------|------|-----------|------------------|
| Nano-PUND binary oxides -- AIP APL Materials | 2024 | 10.1063/5.0179847 | Sub-micron PUND for local polarisation mapping |
| AFE-PUND -- Nano Letters | 2025 | 10.1021/acs.nanolett.5c00851 | Modified PUND for antiferroelectrics |
| Charge-pumping PUND -- TechRxiv | 2024 | TechRxiv | Bulk vs interface trapping separation |
| C-V butterfly HZO MFIM simulation -- RSC Nanoscale | 2025 | d4nr03700f | Three-mechanism C-V model confirmed by simulation |
| V-doped HfO2 high endurance -- Nano Letters | 2025 | 10.1021/acs.nanolett.4c05671 | >10^10 endurance with vanadium doping |
| PyOpticon lab automation -- ACS Chem. Mater. | 2025 | 10.1021/acs.chemmater.5c00644 | Python lab instrument control framework |
| FerroX GPU phase-field -- ScienceDirect / GitHub | 2023 | S0010465523001029 | 3D phase-field framework; 15x GPU speedup |
| Retention express test -- Phys. Rev. Applied | 2022 | PhysRevApplied.18.064084 | Power-law retention model validation |

---

## 2. Parameter Updates vs. Previous February 2026 Review

### 2.1 Parameters Confirmed (No Change Needed)

| Parameter | Previous Review Value | This Session Finding | Status |
|-----------|----------------------|---------------------|--------|
| DefaultHZO Pr | 24.5 uC/cm^2 | Within literature midpoint 10-27.5 uC/cm^2 | VALID |
| DefaultHZO Ec | 1.2 MV/cm | Within 1.0-1.5 MV/cm literature range | VALID |
| FeCIMMaterial Pr | 30 uC/cm^2 | Plausible estimate (not disclosed) | VALID |
| MaterlikHfO2 Pr | 20 uC/cm^2 | Conservative for 2015-era data | VALID |
| CryogenicHZO Pr | 75 uC/cm^2 | Literature-backed for 4K enhanced | VALID |
| DefaultHZO EnduranceCycles | 10^10 | Mid-range for standard HZO | VALID |
| 4-bit DAC/ADC sweet spot | Confirmed in Feb 14 review | Reconfirmed by HCiM and NeuroSim data | VALID |
| FeCAP sneak path immunity | Confirmed in Feb 14 review | Reconfirmed by 5 independent papers | VALID |
| 30-level conductance quantisation | Plausible | Within 7-128 experimental range | VALID (label as single-device estimate) |

### 2.2 Parameters Requiring Correction

| Parameter | Previous Value | Corrected Value | Evidence | Priority |
|-----------|---------------|----------------|----------|----------|
| NLS tau_0 (pre-exponential) | 1e-13 s | 1e-10 s | Guo 2018 (APL), PMC9740545; 3+ independent sources | P0 |
| LiteratureSuperlattice Pr | 50 uC/cm^2 | ~22 uC/cm^2 | Best 2025 nanolaminate 2Pr=43.32 uC/cm^2 (IEEE 10787441) | P0 |
| LiteratureSuperlattice EnduranceCycles | 10^10 | 10^10 (keep, but label as conservative; 10^12 demonstrated for La-doped) | ACS AMI 2024 La-doped 3D-trench | P2 (label update) |

### 2.3 New Parameters Not Previously Tracked

| Parameter | Value | Source | Relevance |
|-----------|-------|--------|-----------|
| NLS Lorentz half-width sigma | 0.5-1.5 decades | Gong 2018; multiple NLS studies | NLS distribution shape |
| NLS activation field Ea (consensus) | 2.5 +/- 0.5 MV/cm | Multiple NLS fits | NLS model calibration |
| Domain wall velocity (HZO nanodots) | 1.2-2.3 m/s | Ahn 2026 (AFM) | KAI regime parameter |
| Minimum switching time | 300 ps | Deng 2023 (Nano Lett.) | Switching speed floor |
| FTJ nonlinearity factor | >1,100 | Athle 2024; Youn 2025 | Built-in selector equivalent |
| FeCAP energy per operation | 0.24 fJ at 1 MHz | Bhardwaj 2025 | Energy benchmark |
| Best C2C variation (128-state FTJ) | 2.75% | ACS AMI 2025 | Programming precision benchmark |
| Schottky barrier height (Pt/HZO) | 0.48 eV | ScienceDirect 2023 | Leakage model parameter |
| PF trap depth (O vacancy in HZO) | 0.3-1.0 eV | Multiple | Leakage model parameter |
| Optical dielectric constant (HZO) | 4.0-5.5 | PF slope extraction | Leakage model parameter |
| Retention relaxation exponent beta | 0.01-0.1 | Multiple HZO studies | Retention model |
| Arrhenius Ea for retention | 0.5-1.5 eV | Multiple | Accelerated retention testing |
| PUND pulse width standard | 1-10 us | Consensus | Measurement protocol |
| PUND delay (P to U) | 250 us - 10 ms | Consensus | Measurement protocol |
| FORC smoothing factor SF | 3-5 | Pike 2003 | FORC analysis parameter |

---

## 3. New Findings Not in the Existing Review

### 3.1 Non-Destructive FeCAP Readout
The Feb 14 review did not cover the Imec + Georgia Tech 2024-2025 result demonstrating non-destructive readout for ferroelectric capacitors. This achieves >10^11 read cycles fully decoupled from write endurance (~10^7 cycles), with a record capacitive memory window of 8.7 at 0V DC bias. BEOL-compatible (<400C). This fundamentally changes the inference energy analysis for FeCAP CIM: inference workloads no longer carry a write-budget penalty.

### 3.2 Orthorhombic-Tetragonal-Orthorhombic Switching Pathway
Nature Communications 2025 and Advanced Materials 2025 establish that HZO polarisation switching proceeds via an o-t-o phase transition pathway, not simple polarisation reversal. The non-polar tetragonal phase is a metastable intermediate. This is a qualitative limitation of single-order-parameter Landau-Khalatnikov models that was not discussed in any prior project document. Resolution requires a two-component Landau model coupling polarisation P and tetragonality order parameter eta.

### 3.3 Complex Domain Walls in Hafnia
PNAS 2024 shows that hafnia domain walls involve simultaneous reversal of BOTH polarisation AND tetragonality order parameters. These "complex domain walls" have lower nucleation barriers than simple polarisation-reversal walls and are fundamentally different from perovskite 180-degree walls. This was not covered in any prior review.

### 3.4 Wake-Up as 90-to-180 Degree Domain Wall Transition
PMC11789571 (2025) provides direct atomic-scale TEM imaging showing that wake-up in HZO corresponds to a transition from 90-degree uncharged domain walls (pristine) to 180-degree domain walls (cycled). Remanent polarisation increases from 12 to 28 uC/cm^2 -- a 2.3x increase. This mechanistic picture was not in the prior review.

### 3.5 ADC-Less CIM Architecture
The HCiM paper (arXiv 2403.13577) proposes replacing ADCs entirely with binary/ternary comparators, achieving 12-28x energy reduction at 3.5% CIFAR-10 accuracy cost. This was not covered in the Feb 14 review, which focused on conventional ADC architectures.

### 3.6 CrossSim V3.1 and NeuroSim V1.5 as Reference Tools
Two open-source CIM accuracy simulators have reached maturity: CrossSim V3.1 (Sandia, January 2025) with direct PyTorch/Keras integration, and NeuroSim V1.5 (May 2025) with GPU-accelerated inference simulation. Neither was mentioned in prior project documents. These represent the benchmark tools against which the FeCIM Lattice Tools Module 3 accuracy modelling should be compared.

### 3.7 Jiles-Atherton Compact Model for HZO
A 2025 Advanced Electronic Materials paper (Paasio et al.) presents a physics-based Jiles-Atherton compact model that reproduces experimental P-V hysteresis of HZO capacitors at different fields and temperatures. This is an alternative to Preisach for frequency-dependent hysteresis modelling and was not previously identified.

### 3.8 No Open-Source Ferroelectric Measurement Framework Exists
Stage 7 confirmed that as of 2026, there is no comprehensive open-source Python/Go library for ferroelectric measurement automation (PUND + FORC + C(V) + retention in one framework). Commercial instruments (Radiant, aixACCT) provide proprietary software. Implementing these workflows in Module 1 would be genuinely novel.

### 3.9 GNN Surrogates for Phase-Field
A 2024 study (PMC11059552) trained graph neural networks on polycrystalline HZO phase-field simulation data, achieving 4.24% error (R^2 = 0.952) with million-fold speedup. Validated on structures with up to 1000 grains. This could eventually enable real-time polydomain simulation in the GUI.

### 3.10 Ferroelectric Memcapacitor with 8-9 States
Advanced Functional Materials 2025 demonstrates 8-9 stable, reprogrammable capacitance states in a TiN/HZO/TiN stack with >10^5 s retention and >10^6 cycle endurance. Programming uses progressive voltage pulses. This validates the FeCAP multi-level approach.

---

*This addendum should be read alongside `crossbar-circuits-literature-review-2025.md` and the master synthesis report at `.omc/research/sciomc-20260218/findings/master-synthesis.md`.*
