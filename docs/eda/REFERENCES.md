# Scientific References: FeCIM Design Suite

This tool implements simulation models based on published research. It is **not affiliated with or endorsed by** the Tour Group at external research institution or any other research institution.

## Disclaimer

- All models are based on **published literature**, not validated against actual hardware
- Performance claims (energy, speed, endurance) are **from cited papers**, not independently verified
- This is **educational/research software**, not a production tool

---

## 1. Core Device Physics

*Published research that our simulation models are based on.*

* **30-State FeFET Device:** *"Flash In2Se3 for Neuromorphic Computing"* (Shin, Tour, et al., 2025). [View](https://www.researchgate.net/publication/388360521_Flash_In2Se3_for_neuromorphic_computing)
  * *How we use it:* Our 30-level quantization model is based on this demonstrated capability.

* **Flash Joule Heating Synthesis:** *"Stoichiometric Engineering... by Flash-within-Flash"* (2025).
  * *How we use it:* Referenced for material properties; we do not implement this manufacturing process.

* **HZO Ferroelectrics:** Park et al., Adv. Mater. 2015; Cheema et al., Nature 2020.
  * *How we use it:* Hysteresis model parameters (Pr, Ec, Ps) are from these publications.

---

## 2. Open Source EDA Ecosystem

*Tools we generate files for (not tools we provide).*

* **OpenLane:** *"OpenLANE: The Open-Source Digital ASIC Implementation Flow"* (WOSET, 2020). [View](https://woset-workshop.github.io/PDFs/2020/a21.pdf)
  * *How we use it:* We generate Verilog/DEF files compatible with OpenLane format.

* **OpenROAD:** *"Empowering innovation: OpenROAD and the future of open-source EDA"* (EE World, 2024).
  * *How we use it:* Our DEF files use OpenROAD-compatible syntax.

* **GDSFactory:** *"GDSFactory: Build Better Hardware with Better Software"* (IEEE, 2024).
  * *How we use it:* Referenced for future GDSII generation (not currently implemented).

---

## 3. Compute-in-Memory Research

*Academic tools that inspired our approach (we are not affiliated with these projects).*

* **NeuroSim:** *"NeuroSim V1.5: Improved Software Backbone for Benchmarking CIM Accelerators"* (Georgia Tech, 2025).
  * *How we use it:* Referenced for CIM energy modeling methodology.

* **CiMLoop:** *"CiMLoop: A Flexible, Accurate, and Fast Compute-In-Memory Modeling Tool"* (MIT, 2024). [View](https://arxiv.org/pdf/2405.07259)
  * *How we use it:* Referenced for architectural exploration approach.

* **CINM Compiler:** *"CINM (Cinnamon): A Compilation Infrastructure for Heterogeneous CIM"* (ACM ASPLOS, 2024).
  * *How we use it:* Referenced for weight-to-hardware mapping methodology.

---

## 4. Performance Claims (From Literature)

**Important:** The following claims are from published papers, not independently verified by this project.

| Claim | Source | Our Status |
|-------|--------|------------|
| 30 discrete states | Tour Lab 2024/2025 | Simulated, not validated |
| 87% MNIST accuracy | Tour COSM presentation | Target, not achieved |
| 10^9 endurance cycles | Tour Lab (demonstrated) | Used in models |
| 10^12 endurance cycles | Tour Lab (target) | Not demonstrated |
| 10ns switching | Various HZO papers | Used in models |

---

## 5. Market Context (Opinion Pieces)

*These are opinion articles, not peer-reviewed research.*

* *"The Microchip Era Is About to End"* by George Gilder (WSJ, 2025) - [View](https://www.wsj.com/articles/the-microchip-era-is-about-to-end-wafer-scale-integration-computing-ai-3a9d554a)
  * *Note:* This is an opinion piece about wafer-scale integration trends, not a validation of FeCIM technology.

---

## What We Don't Claim

1. **No affiliation** with external research institution, Tour Lab, or any cited institution
2. **No endorsement** from any researcher or company mentioned
3. **No hardware validation** - all models are simulation-only
4. **No production readiness** - this is educational/research software
5. **No ownership** of FeCIM technology - we implement published concepts
