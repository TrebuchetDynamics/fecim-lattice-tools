# Open Source Ferroelectric & Crossbar Simulation Tools

A curated collection of open-source tools relevant to the IronLattice demos.

---

## 1. CrossSim (Sandia National Laboratories)

**Repository:** https://github.com/sandialabs/cross-sim  
**Version:** 3.1  
**License:** BSD-3-Clause  
**Language:** Python  
**GPU Support:** ✅ CuPy (CUDA 12.3)

### Description
CrossSim is a GPU-accelerated simulator for analog in-memory computing. It models various device and circuit non-idealities for crossbar arrays.

### Key Features
- **Neural network interface:** PyTorch and Keras integration
- **Hardware-aware training:** Backpropagation through analog layers
- **Device models:** RRAM, PCM, FeFET, SRAM configurable
- **Non-idealities:** Programming errors, conductance drift, read noise
- **Numpy-like API:** Drop-in replacement for matrix operations

### Requirements
```
numpy==1.26.3
scipy==1.11.4
tensorflow==2.17.0
pytorch==2.2.1
cupy==12.3.0  # for GPU
```

### Relevance to IronLattice
- **Demo 2 (Crossbar):** Reference implementation for MVM simulation
- **Demo 2 (Non-idealities):** Built-in models for IR drop, variation
- Tutorials for ISCA 2024 and NICE 2024 conferences

### Clone & Install
```bash
git clone https://github.com/sandialabs/cross-sim
cd cross-sim
pip install .
git submodule init && git submodule update --progress  # 1.2GB data
```

---

## 2. WaCPro (Waveform and Crossbar Programmer)

**Repository:** https://github.com/DUTh-FET/WaCPro  
**Version:** 1.0.3  
**Author:** Ioannis K. Chatzipaschalis (DUTh-FET)  
**Language:** Python + PyQt GUI  
**GPU Support:** ❌

### Description
Open-source GUI application for waveform generation and crossbar programming, designed for memristor and crossbar experiments.

### Key Features
- **GUI-based design:** Intuitive pulse configuration
- **Waveform types:** Step, Ramp, Half-Sine for rows; Square for columns
- **Visualization:** Preview pulses, heatmaps over time
- **Export formats:** .txt, .csv, .mat (MATLAB)
- **AWG support:** SCPI control for arbitrary waveform generators
- **Energy calculation:** Logs crossbar write energies

### Parameters
- Total Duration, Time Step, Pulse Width, Spacing
- Start Delay, Column Min/Max Voltage
- Writing Waveform selection

### Relevance to IronLattice
- **Demo 2 (Crossbar):** Reference for waveform/pulse visualization
- **Hardware integration:** Could connect to real AWG for testing

### Clone & Install
```bash
git clone https://github.com/DUTh-FET/WaCPro
cd WaCPro
pip install -r requirements.txt
python main.py
```

---

## 3. PEtra (Polarization Electric Field Tracer)

**Repository:** https://github.com/IONICS-Lab/PEtra (search for correct URL)  
**Focus:** Polymer ferroelectric characterization  
**Type:** Hardware + Software  
**License:** Open-source (research/educational)

### Description
Open-source and versatile PE loop tracer designed for polymeric piezoelectrics like P(VDF-TrFE).

### Key Features
- **Ultra-low current sensitivity:** Down to 2 pA
- **Adjustable gain:** 10³ to 10⁷ V/A
- **Frequency range:** 0.1 Hz to 200 Hz
- **Target materials:** Screen-printed P(VDF-TrFE) transducers

### Application Areas
- Ferroelectric polymer characterization
- Fabrication process optimization
- Flexible/wearable ultrasound transducers

### Relevance to IronLattice
- **Demo 1 (Hysteresis):** Reference for P-E loop visualization
- **Measurement methodology:** Understanding real-world PE loop measurement

---

## 4. ferro_scripts (WMD-group)

**Repository:** https://github.com/WMD-group/ferro_scripts  
**Author:** Walsh Materials Design Group (Imperial College London)  
**Language:** Python  
**License:** MIT

### Description
Python script for generating ferroelectric hysteresis loops based on the Garrity et al. 2014 model.

### Key Features
- **Physics-based model:** Based on [Phys. Rev. Lett. 112, 127601 (2014)](https://journals.aps.org/prl/abstract/10.1103/PhysRevLett.112.127601)
- **YAML configuration:** Material parameters in config files
- **Included materials:** BaTiO₃ (bto_params.yaml), CrCA

### Input Parameters (YAML)
| Parameter | Description |
|-----------|-------------|
| `cell_dims` | Unit cell dimensions in bohr |
| `energy_data` | Q vs Energy table |
| `chi_data` | Q vs Polarizability table |
| `remnant_polarisation` | Pr in μC/cm² |
| `Emax` | Maximum field in kV/cm |
| `Esamples` | Number of field points |

### Requirements
```
numpy, scipy, matplotlib, pyyaml
```

### Usage
```bash
cd ferro_scripts
python hysteresis.py parameters.yaml
```

### Relevance to IronLattice
- **Demo 1 (Hysteresis):** Algorithm reference for P-E loop generation
- **Alternative model:** Comparison to Preisach approach

---

## 5. Preisachmodel (fddf22)

**Repository:** https://github.com/fddf22/Preisachmodel  
**Language:** Python  
**License:** MIT

### Description
Forward and numerically inverted Preisach model implementation based on the formulation by Isaak D. Mayergoyz.

### Key Features
- **Forward model:** Input field → Output polarization
- **Inverse model:** Numerical inversion for control applications
- **Classical formulation:** Based on Mayergoyz's original theory

### References
1. "Mathematical models of hysteresis" - Isaak D. Mayergoyz
2. "Identification and Inversion of Magnetic Hysteresis for Sinusoidal Magnetization" - Kozek & Gross
3. "Removing numerical instabilities in Preisach model using genetic algorithms" - Consolo et al.
4. "Analytical Approximation of Preisach Distribution Functions" - Janos Fuezi (IEEE TMag 2003)

### Relevance to IronLattice
- **Demo 1 (Hysteresis):** Direct reference for Preisach implementation
- **Already have:** Similar implementation in `demo1-hysteresis/pkg/ferroelectric/preisach.go`

### Clone & Install
```bash
git clone https://github.com/fddf22/Preisachmodel
cd Preisachmodel
pip install numpy scipy matplotlib
python preisach.py
```

---

## Comparison Table

| Tool | Focus | Language | GPU | UI | Demo Relevance |
|------|-------|----------|-----|----|----|
| **CrossSim** | Crossbar MVM | Python | ✅ | ❌ | Demo 2 |
| **WaCPro** | Waveform/Crossbar | Python | ❌ | ✅ GUI | Demo 2 |
| **PEtra** | P-E Loop Measurement | Python | ❌ | ❌ | Demo 1 |
| **ferro_scripts** | Hysteresis Simulation | Python | ❌ | ❌ | Demo 1 |
| **Preisachmodel** | Preisach Hysteresis | Python | ❌ | ❌ | Demo 1 |

---

## Additional Resources

### FerroX (Berkeley Lab)
- **Purpose:** GPU phase-field simulation of ferroelectrics
- **Paper:** arXiv:2210.15668 (already downloaded)
- **Framework:** AMReX-based, 15x GPU speedup
- **Relevance:** Demo 3 TDGL implementation reference

### IBM AIHWKit
- **Purpose:** Hardware-aware neural network training
- **Repository:** https://github.com/IBM/aihwkit
- **Documentation:** https://aihwkit.readthedocs.io
- **Relevance:** Demo 2 training with non-idealities

---

## Dr. external research group Collaborators

### Jaeho Shin (external research institution)
- **Focus:** 2D ferroelectric semiconductors, neuromorphic computing
- **Key Paper:** "α‐In2Se3 Synthesized by FWF for Neuromorphic Computing" (Adv. Electronic Materials, Nov 2024)
- **Device:** 2D ferroelectric semiconductor FET artificial synaptic device
- **Achievement:** 87% MNIST accuracy with In2Se3 FeFET

### Other Tour Lab Publications
- Flash Joule Heating synthesis methods
- HfO₂/ZrO₂ superlattice ferroelectrics
- Neuromorphic computing architectures
