# Scientific Validation: FeCIM Design Suite

The FeCIM Design Suite is based on peer-reviewed research from the Tour Group at external research institution and the broader open-source EDA community.

## Core Technology: FeFET Physics

The ferroelectric field-effect transistors (FeFETs) used in this design suite are based on:
- HfO₂-ZrO₂ (HZO) superlattice structures
- In₂Se₃ (indium selenide) 2D ferroelectrics via Flash Joule Heating
- 30 discrete analog conductance states
- Rewritable with 10⁶ to 10¹² cycle endurance

## 1. Core Device Physics
*Research validating the underlying materials and manufacturing process.*

*   **The 30-State Device:** *"Flash In2Se3 for Neuromorphic Computing"* (Shin, Tour, et al., 2025). [View](https://www.researchgate.net/publication/388360521_Flash_In2Se3_for_neuromorphic_computing)
    *   *Significance:* Validates the 30-state analog memory and Compute-in-Memory capability of In2Se3.
*   **The Manufacturing Process:** *"Stoichiometric Engineering... by Flash-within-Flash"* (2025).
    *   *Significance:* Validates the "Capital Light" flash synthesis method.
*   **Lithium Recovery:** *"Two-step flash Joule heating method recovers lithium‑ion battery materials"* (Science Advances, 2025).
    *   *Significance:* Validates the scalability of Flash Joule Heating.

## 2. Open Source EDA Ecosystem
*Research validating the production-readiness of the tools we use for digital and physical implementation.*

*   **The Digital Flow:** *"OpenLANE: The Open-Source Digital ASIC Implementation Flow"* (WOSET, 2020). [View](https://woset-workshop.github.io/PDFs/2020/a21.pdf)
    *   *Significance:* Proves that open-source tools can produce manufacturable GDSII.
*   **The Physical Engine:** *"Empowering innovation: OpenROAD and the future of open-source EDA"* (EE World, 2024).
    *   *Significance:* Validates the "No-Human-In-The-Loop" routing capability.
*   **Layout Automation:** *"GDSFactory: Build Better Hardware with Better Software"* (IEEE, 2024).
    *   *Significance:* Validates Python-driven layout generation for custom analog structures.

## 3. Compute-in-Memory (CIM) Design Tools
*Research validating the methodologies for modeling and simulating CIM arrays.*

*   **Benchmarking Backbone:** *"NeuroSim V1.5: Improved Software Backbone for Benchmarking CIM Accelerators"* (arXiv, 2025).
    *   *Significance:* The industry standard for estimating CIM energy and area (error < 5% vs silicon).
*   **System Modeling:** *"CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool"* (MIT, IEEE 2024). [View](https://arxiv.org/pdf/2405.07259)
    *   *Significance:* Validates our architectural exploration approach (Tab 3).
*   **Compilation:** *"CINM (Cinnamon): A Compilation Infrastructure for Heterogeneous CIM"* (ACM ASPLOS, 2024).
    *   *Significance:* Demonstrates 51x performance improvement via compiler-based offloading.
*   **Simulator:** *"FAST: Functional Array Simulator for Memristor-based ANNs"* (Nature Communications, 2024).
    *   *Significance:* Validates hardware-software co-design methodologies.

## 4. Neuromorphic Architectures
*Research validating the system-level application of CIM chips.*

*   **Analog AI:** *"An analog-AI chip for energy-efficient speech recognition"* (Nature, 2023).
    *   *Significance:* Proof of concept for fully analog inference chips (IBM NorthPole).
*   **LLM Acceleration:** *"Memory Is All You Need: CIM Architectures for LLM Inference"* (arXiv, 2024).
    *   *Significance:* Validates the application of CIM to modern Transformer workloads.
*   **Spiking Networks:** *"Fully memristive SNN for energy-efficient graph learning"* (Science, 2025).
    *   *Significance:* Demonstrates extremely low power (1.93 pJ/op) spiking neural networks.

---

**Market Context:** *"The Microchip Era Is About to End"* by George Gilder (WSJ, 2025) - [View Article](https://www.wsj.com/articles/the-microchip-era-is-about-to-end-wafer-scale-integration-computing-ai-3a9d554a)
