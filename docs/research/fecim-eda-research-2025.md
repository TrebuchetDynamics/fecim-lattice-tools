# FeCIM and EDA Research Compilation (2024-2026)

> Comprehensive research on Ferroelectric Compute-in-Memory technology, OpenLane, and open-source EDA tools.
> Generated: January 2026

---

## Table of Contents

1. [FeCIM Academic Research](#1-fecim-academic-research)
2. [OpenLane and OpenROAD Ecosystem](#2-openlane-and-openroad-ecosystem)
3. [Analog and Memory EDA Tools](#3-analog-and-memory-eda-tools)
4. [Crossbar Array Simulation](#4-crossbar-array-simulation)
5. [Neuromorphic Design Tools](#5-neuromorphic-design-tools)
6. [Recommendations for FeCIM Design](#6-recommendations-for-fecim-design)
7. [Complete Reference List](#7-complete-reference-list)

---

## 1. FeCIM Academic Research

### 1.1 HfO₂-ZrO₂ Superlattice Ferroelectric Research

#### Recent Publications (2024-2025)

| Source | Year | Key Finding |
|--------|------|-------------|
| ACS Applied Nano Materials | 2024 | HfO2-ZrO2 superlattice with 4×[HfO2(10Å)-ZrO2(5Å)] structure achieving robust ferroelectric polarization with 500°C annealing |
| Materials Horizons | 2024 | Self-rectifying ferroelectric tunnel junction using HfO2/ZrO2/HfO2 superlattice for neuromorphic synapses |
| ACS Omega | 2024 | Solution-processed HfO2-ZrO2 multilayer films (50nm thick) with reduced wake-up effects |

#### Key Material Properties Achieved

| Parameter | Value | Notes |
|-----------|-------|-------|
| Remanent Polarization (2Pr) | 26-45 µC/cm² | Varies by structure |
| Endurance | >10¹² cycles | With La doping |
| Coercive Voltage | Reduced | Through element inhomogeneity optimization |
| CMOS Compatibility | 400°C processing | BEOL compatible |

### 1.2 Multi-Level Cell Ferroelectric Memory

#### Ferroelectric Tunnel Junctions (FTJs)

**Advanced Intelligent Systems (2024)**: W/HfₓZr₁₋ₓO₂/TiN FTJ
- **60 programmable conductance states**
- Dynamic range of 10
- Current density >3 A/m² at Vread=0.3V
- Nonlinearity >1100

**Nature Communications (2022)**: High-performance FTJ
- **256 conductance states (8-bit)**
- Cycle-to-cycle variation: ~2.06%
- Linearity: <1
- Endurance: >10⁹ cycles
- Switching speed: ~630 ps pulses at ≤5V
- **32 conductance states (5-bit)** with 1.6% variation

#### Ferroelectric FETs (FeFETs)

**Nature Communications (2023)**: Multi-level cell FeFET crossbar
- **7 distinct threshold voltage states** (~3 bits/cell)
- **96.6% MNIST accuracy**
- **885 TOPS/W energy efficiency**
- 28nm HKMG technology validation

**ResearchGate (2024)**: 1k FeFET crossbar
- **7 distinct VT states** with program/verify scheme

### 1.3 Ferroelectric Compute-in-Memory Architectures

#### Key Publications

**Nature Communications (March 2024)** - "Ferroelectric compute-in-memory annealer for combinatorial optimization problems"
- FeFET-based crossbar for vector-matrix-vector (VMV) multiplication
- 75% chip size savings for Max-Cut problems
- Column-shared ADCs with Shift-and-Add processing units

**Science Advances (2024)** - "Two-dimensional fully ferroelectric-gated hybrid computing-in-memory"
- On/off ratios: >10⁷
- Endurance: >10¹²
- Retention: >10 years
- Device variation: ~0.3-0.5%

**Scientific Reports (April 2024)** - "Ferroelectric capacitors and field-effect transistors as in-memory computing elements"
- Low-voltage operation: 1.2 V through interfacial layer engineering
- Selector-less crossbar operation
- Eliminates sneak paths inherently

**arXiv (October 2024)** - "FeBiM: Ferroelectric Bayesian Inference Engine"
- Storage density: 26.32 Mb/mm²
- Computing efficiency: 581.40 TOPS/W
- One FeFET per cell design

**arXiv (January 2025)** - "Dendritic Computing with Multi-Gate FeFETs"
- Novel neuron design mimicking dendrites
- Local nonlinear computations
- Smaller crossbar arrays for efficiency

### 1.4 Neural Network Acceleration Performance

#### MNIST/Fashion-MNIST Results

| Accuracy | Architecture | Source | Year |
|----------|--------------|--------|------|
| **98.78%** | 24×24 ferroelectric memristor array | Journal of Materiomics | 2025 |
| **96.6%** | FeFET crossbar, 885 TOPS/W | Nature Communications | 2024 |
| **97.6%** | 1F-1T FeFET MLP | IEEE | 2024 |
| **91.5%** | FeFET crossbar | IEEE | 2024 |
| **88.78%** | Reservoir computing | Nature | 2025 |

#### Energy Efficiency Benchmarks

| Efficiency | Architecture | Source |
|------------|--------------|--------|
| **885 TOPS/W** | Multi-level FeFET crossbar | Nature Communications 2023 |
| **581.40 TOPS/W** | FeBiM Bayesian inference | arXiv 2024 |
| **111.6 TOPS/W** | Charge-domain CiM | IEEE 2024 |
| **2.612 POPS/J** | Proposed MAC accelerator | Research 2024 |

#### Integration Density
- Area efficiency: 88.5 bits/µm² (including peripherals)
- 4F² crossbar architecture capability

### 1.5 Non-Idealities and Solutions

#### Ferroelectric Advantages Over Resistive Memories

**From Nano Convergence (2024)** - "Ferroelectric capacitive memories: devices, arrays, and applications":

| Issue | FeCAP Advantage |
|-------|-----------------|
| IR Drop | Much higher device resistance than interconnects; free from IR drop issues |
| Sneak Paths | Capacitor nature inherently suppresses sneak currents |
| Static Power | Negligible static power consumption |
| Selector | No selector required (capacitive nature) |

#### Remaining Challenges (2024)

**Reliability Issues**:
- Endurance degradation: Electron trapping, oxygen vacancy migration
- Retention failures: Depolarization fields in ultrathin films
- Wake-up/fatigue effects: Requires cycling for optimal performance
- Imprint effects: Asymmetric polarization switching

**Variability Sources**:
- Device-to-device (D2D) variation: Limits precise weight representation
- Cycle-to-cycle (C2C) variation: Affects programming reliability
- Trapped charges and defects: Introduce asymmetry in hysteresis
- Oxygen vacancies: Create non-ideal switching behavior

**Mitigation Strategies**:
- Dual defect-shielding layers (achieved >10¹³ endurance)
- Electrode engineering (tungsten for tensile stress)
- Interface layer optimization
- Controlled element inhomogeneity

### 1.6 Technology Comparison (2024)

| Technology | Write Endurance | Cell Size | Advantages | Challenges |
|------------|----------------|-----------|------------|------------|
| **FeRAM/FeFET** | >10¹² cycles | Sub-10nm capable | CMOS compatible, sneak-path free, low static power, multi-level | Scaling <5nm, wake-up effects, retention |
| **ReRAM** | 10⁶-10⁹ cycles | 4F² | High density, >7 bits/cell | IR drop, sneak paths, requires selector |
| **PCM** | 10⁸-10⁹ cycles | Larger | Proven technology | High power, slow write, thermal crosstalk |
| **MRAM** | >10¹⁵ cycles | Moderate | Ultra-fast, high endurance | Limited multi-level, complex fabrication |

### 1.7 Preisach Model for Hysteresis Simulation

**Physica Scripta (June 2024)** - "The Preisach model of hysteresis: fundamentals and applications"
- Comprehensive review of Preisach model generalizations
- Applications in ferroelectric and ferromagnetic systems
- Stochastic external impact modeling

**Nature Communications (2018)** - "Physical reality of the Preisach model for organic ferroelectrics"
- Strong physical basis for polarization switching
- Valid across all time scales

**Scientific Reports (2021)** - "Extraction of Preisach model parameters for fluorite-structure ferroelectrics"
- Fast, robust modeling framework for HfO₂-based materials
- Automated parameter determination

---

## 2. OpenLane and OpenROAD Ecosystem

### 2.1 OpenLane - Automated RTL to GDSII Flow

**Repository**: https://github.com/The-OpenROAD-Project/OpenLane

#### Architecture
- Automated RTL to GDSII flow based on multiple integrated components
- Currently in maintenance mode (version 1.0.X)
- **OpenLane 2** (https://github.com/chipfoundry/openlane2) is next-generation rewrite
- **LibreLane** recommended for new designs

#### Key Components Integrated
| Component | Function |
|-----------|----------|
| OpenROAD | Placement and routing engine |
| Yosys | Logic synthesis framework |
| Magic | Layout design and verification |
| Netgen | Circuit netlist comparison (LVS) |
| KLayout | Mask layout viewer and editor |
| CVC | Circuit validity checking |
| SPEF-Extractor | Parasitic extraction |

#### Track Record
- **600+ successful tapeouts** through Google MPW shuttle and Efabless chipIgnite

### 2.2 OpenROAD - Core Physical Design Platform

**Repository**: https://github.com/The-OpenROAD-Project/OpenROAD

#### Purpose
Autonomous 24-hour RTL-to-GDSII flow with no human intervention (DARPA IDEA program)

#### Core Tools and Stages

```
1. Synthesis → Yosys (RTL to netlist)
2. Floorplanning → Chip area, IO placement, power distribution
3. Placement → RePlAce (analytical placement, cells as charged particles)
4. Clock Tree Synthesis (CTS) → Buffer insertion and optimization
5. Routing → Global routing + detailed routing
6. Timing Analysis → OpenSTA (static timing analyzer)
7. Parasitic Extraction (PEX) → Wire R and C modeling
8. Finishing → GDSII generation
```

#### Architecture
- Common database: **OpenDB**
- API bindings: Tcl and Python
- GUI for visualization and debugging
- Multi-threading support
- Technology support: 7nm to 180nm

### 2.3 Process Design Kits (PDKs)

#### Sky130 PDK
**Repository**: https://github.com/google/skywater-pdk

| Property | Value |
|----------|-------|
| Technology | 180nm-130nm hybrid (originally Cypress) |
| Voltage | 1.8V internal, 5.0V I/Os |
| Features | Local interconnect, SONOS, MiM capacitors |
| License | Apache 2.0 |
| Adoption | 50+ universities |

#### GF180MCU PDK
**Repository**: https://github.com/google/gf180mcu-pdk

| Property | Value |
|----------|-------|
| Technology | GlobalFoundries 180nm MCU bulk |
| Libraries | 7- and 9-track digital cells, 3.3V-10V devices |
| Capacity | 16+ million wafers/year → 22+ million by 2026 |
| Applications | Motor controllers, RFID, MCUs, PMICs, IoT |
| License | Apache 2.0 |

### 2.4 Digital Design Tools

#### Yosys - Open SYnthesis Suite
**Repository**: https://github.com/YosysHQ/yosys

| Feature | Description |
|---------|-------------|
| Language Support | Verilog-2005 |
| Architecture | Modular "passes" for customizable flows |
| Output Formats | Verilog, BLIF, EDIF, BTOR, SMT-LIB |
| Formal Methods | Property checking, equivalence |
| License | ISC (GPL-compatible) |

#### OpenSTA - Static Timing Analysis
**Repository**: https://github.com/The-OpenROAD-Project/OpenSTA

- Gate-level static timing verifier
- TCL command interpreter
- Network adapter architecture
- Dual licensed: GPL v3 + commercial

### 2.5 Layout and Verification Tools

#### Magic VLSI
**Repository**: https://github.com/RTimothyEdwards/magic
**Website**: http://opencircuitdesign.com/magic/

| Feature | Capability |
|---------|------------|
| DRC | Real-time checking, hierarchical DRC |
| Extraction | Transistor/resistor extraction, SPICE netlists |
| Formats | CIF, GDS, LEF, DEF, SPICE |
| License | Berkeley open-source |

#### KLayout - Layout Viewer and Editor
**Repository**: https://github.com/klayout/klayout
**Website**: https://www.klayout.de

- GDS and OASIS file viewer/editor
- Multi-layout overlay
- Parametric cells (PCells)
- Ruby scripting
- LVS and PEX capabilities

#### Netgen - LVS Tool
**Repository**: https://github.com/RTimothyEdwards/netgen
**Website**: http://opencircuitdesign.com/netgen/

- Compares SPICE or Verilog netlists
- Hierarchical LVS support
- Tcl extension

### 2.6 Integration: Digital Design Flow

```
RTL (Verilog/VHDL)
    ↓
[Yosys] → Logic Synthesis → Gate-level netlist
    ↓
[OpenROAD/OpenLane Flow]
    ├─ Floorplanning (chip area, I/O, power grid)
    ├─ Placement (RePlAce, timing-driven)
    ├─ CTS (clock tree synthesis)
    ├─ Routing (global + detailed)
    ├─ [OpenSTA] → Timing analysis
    └─ Parasitic extraction
    ↓
[Magic] → Layout finishing + DRC
    ↓
GDSII (manufacturing)
    ↓
[Netgen] → LVS verification
    ↓
[KLayout] → Final layout viewing/editing
```

---

## 3. Analog and Memory EDA Tools

### 3.1 Analog Layout Automation

#### ALIGN (Analog Layout, Intelligently Generated from Netlists)
**Repository**: https://github.com/ALIGN-analoglayout/ALIGN-public
**Website**: https://align-analoglayout.github.io/ALIGN-public/

| Property | Description |
|----------|-------------|
| Developers | University of Minnesota, Texas A&M, Intel (DARPA IDEA) |
| Input | Unannotated SPICE netlists |
| Output | GDSII layout |
| Deployment | Docker: `darpaalign/align-public:latest` |

**Architecture (4 stages)**:
1. Circuit Annotation: Hierarchical netlist representation
2. Design Rule Abstraction: PDK → JSON format
3. Primitive Cell Generation: Transistor structures → GDSII
4. Placement and Routing: Hierarchical assembly

#### MAGICAL (Machine Generated Analog IC Layout)
**Repository**: https://github.com/magical-eda/MAGICAL

| Property | Description |
|----------|-------------|
| Developers | DARPA IDEA program |
| Validation | 40nm 1GS/s ΔΣ ADC |
| Enhancement | LayoutCopilot (2025): LLM-powered multi-agent frameworks |

#### BAG (Berkeley Analog Generator)
**Repository**: https://github.com/ucb-art/BAG_framework

- Python-based framework interfacing with Cadence Virtuoso
- Automatically generates layout, schematic, LVS, PEX
- **BagNet Enhancement**: DNN-based discriminator, 2+ orders of magnitude improvement

### 3.2 Schematic Capture and Simulation

#### Xschem - Schematic Capture
**Repository**: https://github.com/StefanSchippers/xschem
**Website**: https://xschem.sourceforge.io/stefan/index.html

| Property | Description |
|----------|-------------|
| Purpose | VLSI/ASIC/Analog schematic editor |
| Netlists | VHDL, Spice, Verilog |
| License | GNU GPL |

#### ngspice - SPICE Simulator
**Repository**: https://github.com/ngspice/ngspice
**Website**: https://ngspice.sourceforge.io/

| Feature | Description |
|---------|-------------|
| Heritage | Successor to Berkeley SPICE 3f.5 |
| XSPICE | Analog behavioral modeling + digital co-simulation |
| Cider | Numerical device simulator |
| Verilog-A | OpenVAF/OSDI interface |
| License | BSD-3-Clause |

#### Xyce - Parallel SPICE Simulator
**Repository**: https://github.com/Xyce/Xyce
**Website**: https://xyce.sandia.gov

| Feature | Description |
|---------|-------------|
| Developer | Sandia National Labs |
| Analysis | DC, transient, AC, noise, harmonic balance |
| Scale | Millions of devices with parallel computing |
| Verilog-A | Compiler included |

#### OpenVAF - Verilog-A Compiler
**Repository**: https://github.com/pascalkuthe/OpenVAF
**Website**: https://openvaf.semimod.de

| Feature | Description |
|---------|-------------|
| Performance | 10x faster than commercial alternatives |
| Interface | OSDI (Open Source Device Interface) |
| Status | Community fork: OpenVAF-Reloaded (active 2024) |

### 3.3 Memory Compilers

#### OpenRAM - SRAM Compiler
**Repository**: https://github.com/VLSIDA/OpenRAM
**Website**: https://openram.org

| Output | Description |
|--------|-------------|
| Layouts | Physical SRAM layouts |
| Netlists | Complete circuit netlists |
| Models | Timing and power models |
| P&R | Placement and routing models |
| License | BSD 3-clause |

#### OpenRRAM - Resistive RAM Compiler
**Repository**: https://github.com/akashlevy/OpenRRAM

- Based on OpenRAM framework
- Alternative: https://github.com/akdimitri/RRAM_COMPILER (Imperial College)
- TSMC 180nm technology support

### 3.4 Analog Design Flow

```
Circuit Concept
    ↓
[Xschem] → Schematic capture → SPICE netlist
    ↓
[ngspice/Xyce] → Circuit simulation & verification
    ↓
[ALIGN or BAG] → Automated layout generation
    ↓
[Magic] → Manual layout editing + DRC + extraction
    ↓
[Netgen] → LVS (schematic vs. layout)
    ↓
[ngspice] → Post-layout simulation (with parasitics)
    ↓
[KLayout] → Final layout review
    ↓
GDSII
```

---

## 4. Crossbar Array Simulation

### 4.1 CrossSim (Sandia National Labs)
**Repository**: https://github.com/sandialabs/cross-sim

| Feature | Description |
|---------|-------------|
| Language | Python with GPU acceleration |
| Capabilities | Fast internal circuit simulator for parasitic resistance |
| Applications | Neural networks, signal processing, linear systems |
| Parameterizability | Extensive system-level options |

### 4.2 NeuroSim
**Repository**: https://github.com/neurosim

| Feature | Description |
|---------|-------------|
| Framework | DNN+NeuroSim for end-to-end benchmarking |
| Accuracy | Chip-level error under 1% after calibration |
| Coverage | Device, circuit, and algorithm levels |

### 4.3 Additional CIM Benchmarking Tools

| Tool | Repository | Description |
|------|------------|-------------|
| ZigZag-IMC | https://github.com/KULeuven-MICAS/zigzag-imc | Analytical CIM performance model |
| MNSIM 2.0 | - | Behavior-level modeling for analog/digital CIM |
| CiMLoop | - | Statistical energy modeling |
| CIMinus | - | Unified DNN pruning to system-level cost |

### 4.4 ReRAM/Memristor Simulation

#### MemTorch
**Repository**: https://github.com/coreylammie/MemTorch

- PyTorch integration for memristive deep learning
- RRAM modeling with device non-idealities
- CPU and CUDA support

#### NeuroPack
- Algorithm-level Python simulator
- Complete hierarchical framework for SNNs
- Various neuron models and learning rules

#### DL-RSIM
- TensorFlow-based reliability simulation
- ReRAM-based CNN accelerator focus

#### Basic Python Implementations
- HP Labs Ion Drift and Yakopcic models: https://github.com/thomast8/Memristor-Models
- GUI-based simulator: https://github.com/DuttaAbhigyan/Memristor-Simulation-Using-Python

---

## 5. Neuromorphic Design Tools

### 5.1 PyNN
- Most widespread common interface to neuromorphic hardware
- Supports: SpiNNaker, BrainScaleS, Heidelberg Spikey
- Python API for simulator-independent specification
- EU Human Brain Project development

### 5.2 Lava (Intel)
- Framework for brain-inspired neural networks
- Target: Loihi digital neuromorphic hardware

### 5.3 SpiNNaker2
| Property | Value |
|----------|-------|
| Hardware | 153 ARM cores per chip, 19MB SRAM, 2GB DRAM |
| Accelerators | Machine Learning (MAC) and Neuromorphic (Exp/Log) |
| Scale | 10 million core target |
| Dresden Supercomputer | 5+ million cores (5 billion neurons) |
| Fabricated Units | 34,500+ (as of 2026) |

### 5.4 NIR (Neuromorphic Intermediate Representation)
- Common reference frame for digital neuromorphic computations
- Hybrid systems: continuous-time dynamics + discrete events
- Abstracts discretization and hardware constraints

### 5.5 Machine Learning for EDA

**NVIDIA Research Highlights**:
- "Large Language Model for Standard Cell Layout Design Optimization" (Best Paper, LAD 2024)
- "INSTA: Ultra-Fast, Differentiable Statistical Static Timing Analysis" (Best Paper, DAC 2025)

**Curated Resources**: https://github.com/Thinklab-SJTU/awesome-ai4eda

**Key Techniques**:
- Graph Neural Networks for placement
- Reinforcement learning for design space exploration
- Computer vision for layout tasks

---

## 6. Recommendations for FeCIM Design

Based on your ferroelectric compute-in-memory project using HfO2-ZrO2 superlattices with 30 analog states:

### 6.1 Circuit Simulation
- **Xyce** or **Ngspice** with **OpenVAF** for custom FeFET device models
- Implement Preisach model parameters matching specifications (Pr ~25 µC/cm², Ec ~1 MV/cm)

### 6.2 Crossbar Array Simulation
- **CrossSim** (Sandia) for fast GPU-accelerated MVM operations
- **NeuroSim** for device-to-system benchmarking with non-idealities
- **ZigZag-IMC** for system-level performance exploration

### 6.3 Analog Layout Automation
- **ALIGN** for automated analog peripheral circuits (DAC/ADC/TIA)
- **MAGICAL** if design space exploration with ML optimization needed

### 6.4 Memory Compiler
- **OpenRRAM** as starting point - adaptable to ferroelectric devices
- Custom modifications needed for multi-level cell (30 states vs. binary)

### 6.5 Digital Integration
- **OpenROAD** flow for digital control logic and interface circuits
- **SkyWater SKY130** PDK for prototyping if fabrication access available

### 6.6 Machine Learning EDA
- **BagNet** for layout optimization of analog peripherals
- **NVIDIA Research** techniques for placement optimization

### 6.7 Neuromorphic Applications
- **PyNN** for algorithm development
- **NIR** for portability across platforms

### 6.8 TCAD for Device Physics
- Commercial tools (Sentaurus TCAD) with Preisach model
- Calibrate to 30-level quantization scheme

---

## 7. Complete Reference List

### 7.1 FeCIM and Ferroelectric Memory

#### Multi-Level Cell FeFET & Performance
- [First demonstration of in-memory computing crossbar using multi-level Cell FeFET](https://www.nature.com/articles/s41467-023-42110-y) - Nature Communications 2023
- [Reliable multi-level cell programming in FeFET arrays](https://www.researchgate.net/publication/390350071) - 2024
- [Device Feasibility Analysis of Multi-level FeFETs](https://ieeexplore.ieee.org/document/10595900/) - IEEE 2024

#### Ferroelectric Tunnel Junctions
- [Ferroelectric Tunnel Junction Memristors for In-Memory Computing](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202300554) - Advanced Intelligent Systems 2024
- [High-precision weight updates by subnanosecond pulses in FTJ](https://www.nature.com/articles/s41467-022-28303-x) - Nature Communications 2022

#### HfO₂-ZrO₂ Materials & Endurance
- [HfO2-ZrO2 Superlattice Ferroelectric FETs](https://pubs.acs.org/doi/10.1021/acsanm.4c04974) - ACS Applied Nano Materials 2024
- [Ferroelectric HfO2-ZrO2 Multilayers with Reduced Wake-Up](https://pubs.acs.org/doi/10.1021/acsomega.4c10603) - ACS Omega 2024
- [La Doped HZO-Based 3D-Trench Capacitors (>10¹² endurance)](https://colab.ws/articles/10.1109/led.2024.3368225) - IEEE LED 2024
- [Optimization of ferroelectricity and endurance of HZO](https://www.sciopen.com/article/10.26599/JAC.2024.9220916) - Journal Advanced Ceramics 2024

#### Compute-in-Memory Architectures
- [Ferroelectric CiM annealer for combinatorial optimization](https://www.nature.com/articles/s41467-024-46640-x) - Nature Communications March 2024
- [2D fully ferroelectric-gated hybrid CiM](https://www.science.org/doi/10.1126/sciadv.adp0174) - Science Advances 2024
- [Ferroelectric capacitors and FETs as in-memory computing elements](https://www.nature.com/articles/s41598-024-59298-8) - Scientific Reports April 2024
- [CMOS-compatible CiM accelerators based on ferroelectric arrays](https://www.science.org/doi/full/10.1126/sciadv.abm8537) - Science Advances

#### Non-Idealities & Array Design
- [Ferroelectric capacitive memories: devices, arrays, applications](https://link.springer.com/article/10.1186/s40580-024-00463-0) - Nano Convergence 2024
- [Cross-Layer Framework for Ferroelectric Capacitor-Based CiM](https://dl.acm.org/doi/abs/10.1109/ASP-DAC58780.2024.10473887) - ASP-DAC 2024
- [Error-Aware Training for In-RRAM Computing](https://dl.acm.org/doi/10.1145/3711830) - ACM JETC 2024

#### Reviews & Advances
- [Recent advances in ferroelectric materials, devices, and CiM](https://pmc.ncbi.nlm.nih.gov/articles/PMC12592630/) - PMC 2024/2025
- [Hafnium oxide-based ferroelectric FETs: materials to applications](https://pubs.aip.org/aip/jap/article/138/1/010701/3351745/) - J. Applied Physics 2024
- [Ferroelectric Hafnium Oxide: Game-Changer for Nanoelectronics](https://advanced.onlinelibrary.wiley.com/doi/10.1002/aelm.202400686) - Advanced Electronic Materials 2025
- [Ferroelectric memristor crossbar arrays for neuromorphic computing](https://www.sciencedirect.com/science/article/abs/pii/S2211285525004963) - Journal of Materiomics May 2025

#### arXiv Preprints (2024-2025)
- [FeBiM: Ferroelectric Bayesian Inference Engine](https://arxiv.org/html/2410.19356) - October 2024
- [Dendritic Computing with Multi-Gate FeFETs](https://arxiv.org/abs/2505.01635) - January 2025
- [Roadmap to Neuromorphic Computing with Emerging Technologies](https://arxiv.org/html/2407.02353v1) - July 2024
- [Full Spectrum of 3D Ferroelectric Memory Architectures](https://arxiv.org/pdf/2504.09713) - April 2025

#### Preisach Modeling
- [Preisach model of hysteresis: fundamentals and applications](https://ui.adsabs.harvard.edu/abs/2024PhyS...99f2008S/abstract) - Physica Scripta June 2024
- [Extraction of Preisach model parameters for fluorite ferroelectrics](https://www.nature.com/articles/s41598-021-91492-w) - Scientific Reports 2021
- [Physical reality of Preisach model for organic ferroelectrics](https://www.nature.com/articles/s41467-018-06717-w) - Nature Communications 2018

### 7.2 OpenLane and OpenROAD

#### Core Projects
- [OpenLane GitHub](https://github.com/The-OpenROAD-Project/OpenLane) - RTL to GDSII flow
- [OpenROAD GitHub](https://github.com/The-OpenROAD-Project/OpenROAD) - Unified application
- [OpenROAD Documentation](https://openroad.readthedocs.io/) - Official docs
- [OpenROAD Flow Scripts](https://github.com/The-OpenROAD-Project/OpenROAD-flow-scripts) - ORFS automation

#### PDKs
- [Sky130 PDK](https://github.com/google/skywater-pdk) - SkyWater open PDK
- [GF180 PDK](https://github.com/google/gf180mcu-pdk) - GlobalFoundries open PDK
- [SkyWater Technology](https://www.skywatertechnology.com/sky130-open-source-pdk/) - Official page

#### Synthesis & Analysis
- [Yosys GitHub](https://github.com/YosysHQ/yosys) - Open synthesis suite
- [Yosys Documentation](https://yosys.readthedocs.io) - User manual
- [OpenSTA GitHub](https://github.com/The-OpenROAD-Project/OpenSTA) - Timing analysis

### 7.3 Layout & Verification

- [Magic GitHub](https://github.com/RTimothyEdwards/magic) - VLSI layout tool
- [Magic Website](http://opencircuitdesign.com/magic/) - Official site
- [KLayout Website](https://www.klayout.de) - Layout viewer
- [Netgen GitHub](https://github.com/RTimothyEdwards/netgen) - LVS tool
- [Netgen Website](http://opencircuitdesign.com/netgen/) - Official site

### 7.4 Analog Tools

- [Xschem GitHub](https://github.com/StefanSchippers/xschem) - Schematic capture
- [Xschem Website](https://xschem.sourceforge.io/stefan/index.html) - Documentation
- [ngspice GitHub](https://github.com/ngspice/ngspice) - SPICE simulator
- [ngspice Website](https://ngspice.sourceforge.io/) - Official site
- [Xyce GitHub](https://github.com/Xyce/Xyce) - Parallel SPICE
- [Xyce Website](https://xyce.sandia.gov/) - Sandia National Labs
- [PySpice GitHub](https://github.com/PySpice-org/PySpice) - Python interface
- [OpenVAF GitHub](https://github.com/pascalkuthe/OpenVAF) - Verilog-A compiler
- [OpenVAF Website](https://openvaf.semimod.de/) - Documentation

### 7.5 Analog Layout Automation

- [ALIGN GitHub](https://github.com/ALIGN-analoglayout/ALIGN-public) - Analog layout automation
- [ALIGN Website](https://align-analoglayout.github.io/ALIGN-public/) - Documentation
- [MAGICAL GitHub](https://github.com/magical-eda/MAGICAL) - ML-based analog layout
- [BAG Framework](https://github.com/ucb-art/BAG_framework) - Berkeley Analog Generator
- [BagNet Paper](https://ieeexplore.ieee.org/document/8942062/) - DNN-boosted optimizer

### 7.6 Memory Compilers

- [OpenRAM GitHub](https://github.com/VLSIDA/OpenRAM) - SRAM compiler
- [OpenRAM Website](https://openram.org/) - Documentation
- [OpenRRAM GitHub](https://github.com/akashlevy/OpenRRAM) - RRAM adaptation
- [RRAM Compiler (Imperial College)](https://github.com/akdimitri/RRAM_COMPILER) - Research implementation

### 7.7 Crossbar Array & CIM Simulation

- [CrossSim GitHub](https://github.com/sandialabs/cross-sim) - Sandia crossbar simulator
- [NeuroSim GitHub](https://github.com/neurosim) - Stanford CiM benchmarking
- [ZigZag-IMC GitHub](https://github.com/KULeuven-MICAS/zigzag-imc) - System-level exploration
- [DNN+NeuroSim Paper](https://www.frontiersin.org/journals/artificial-intelligence/articles/10.3389/frai.2021.659060/full) - Validation
- [MemTorch GitHub](https://github.com/coreylammie/MemTorch) - Memristive DL framework
- [Memristor Models GitHub](https://github.com/thomast8/Memristor-Models) - HP Labs models

### 7.8 Neuromorphic Computing

- [SpiNNaker2 Paper](https://arxiv.org/html/2401.04491v1) - 10 million core system
- [SpiNNaker2 Overview](https://open-neuromorphic.org/neuromorphic-computing/hardware/spinnaker-2-university-of-dresden/) - Hardware
- [NIR Paper](https://www.nature.com/articles/s41467-024-52259-9) - Neuromorphic intermediate representation
- [EBRAINS Neuromorphic](https://www.ebrains.eu/modelling-simulation-and-computing/computing/neuromorphic-computing/) - Platform access

### 7.9 Machine Learning for EDA

- [NVIDIA EDA Research](https://research.nvidia.com/labs/electronic-design-automation/) - Latest ML techniques
- [Awesome AI4EDA GitHub](https://github.com/Thinklab-SJTU/awesome-ai4eda) - Curated papers
- [AI-Native EDA Paper](https://arxiv.org/html/2403.07257v2) - Large circuit models

### 7.10 Ferroelectric Memory TCAD

- [Dual-Bit FeFET Paper](https://www.nature.com/articles/s44335-025-00030-8) - npj 2025
- [HfO2 FeFET Reliability](https://link.springer.com/article/10.1007/s00542-025-05919-9) - Microsystem Tech 2025
- [Recessed Channel FeFET](https://link.springer.com/article/10.1007/s40042-024-01079-7) - Korean Physical Society 2024

### 7.11 CIM Benchmarking

- [Analog vs Digital CIM Paper](https://arxiv.org/html/2405.14978v1) - Quantitative modeling 2024
- [High Precision CIM](https://www.nature.com/articles/s44335-025-00044-2) - npj 2025

### 7.12 Additional Resources

- [OpenROAD Project Website](https://theopenroadproject.org/) - Latest updates
- [Google Open Source Silicon Blog](https://opensource.googleblog.com/2022/08/GlobalFoundries-joins-Googles-open-source-silicon-initiative.html) - GF180 announcement
- [Ultimate Guide to Open Source EDA](https://anysilicon.com/the-ultimate-guide-to-open-source-eda-tools/) - Overview
- [Basilisk Multi-million Gate SoC](https://arxiv.org/html/2405.04257v2) - Case study

---

## Notes

### Project-Specific Observation
The FeCIM Visualizer project specifies "30 discrete analog states (~4.9 bits/cell)". Most academic literature reports:
- 32 states (5-bit)
- 60 states
- Up to 256 states (8-bit)

Consider verifying or documenting the source of the 30-state specification.

### Key Takeaways

1. **FeCIM advantages**: Inherent sneak-path elimination, no selector required, >10¹² endurance, CMOS compatibility
2. **State of the art**: 885 TOPS/W efficiency, 96.6% MNIST accuracy with FeFET crossbars
3. **Open-source EDA**: 600+ successful tapeouts with OpenLane/OpenROAD
4. **Simulation tools**: CrossSim and NeuroSim provide comprehensive CIM modeling
5. **Analog automation**: ALIGN and MAGICAL maturing for practical analog layout

---

*Document compiled from web research conducted January 2026*
