# Ferroelectric compute-in-memory: provenance-first evidence scan v2

Date: 2026-06-18  
Question: What does the peer-reviewed literature say about ferroelectric compute-in-memory devices and architectures, especially FeFET, FeTFT, FTJ, ferroelectric diode, HZO/HfO2, CAM/TCAM, neural accelerators, and evidence gaps?

## Method and limits

This is an evidence scan, not a systematic review. I used ResearchForge (`rforge`) for project setup and source search, then OpenAlex API metadata/abstract retrieval for selected records. No copyrighted full text was downloaded. Claims are therefore limited to titles, metadata, abstracts, and DOI records. `protocol-freeform.txt` records that a reviewer should approve the broad query/screening plan before treating this as a final systematic review.

## Bottom line

Ferroelectric compute-in-memory (FeCIM) is a credible active research area, but most evidence is still **device/prototype, circuit macro, and simulation/architecture** evidence rather than standardized, production-scale accelerator evidence. The strongest practical thread is HfO2/HZO-compatible FeFET/FeTFT work because it aligns with CMOS/HKMG manufacturing. Ferroelectric tunnel junctions/diodes and 2D ferroelectrics are promising but need more array-scale reliability and benchmarking evidence. The best claim hygiene is: **FeCIM has published demonstrations for CIM primitives, analog/MAC arrays, CAM/search, reservoir/HDC, and annealing; broad performance superiority is workload-, device-, and peripheral-assumption-specific.**

## Main device and architecture families

### 1. FeFET / FeTFT compute-in-memory

FeFETs use remanent ferroelectric polarization to shift transistor threshold/conductance. Retrieved sources cover Boolean CIM, analog matrix-vector multiplication, sensor fusion, HDC, and multi-level crossbar macros.

Representative evidence:

- **Computing in memory with FeFETs** (ACM/ISLPED 2018, DOI `10.1145/3218603.3218640`) introduces a FeFET-CIM architecture that can operate as RAM and perform Boolean operations and addition in memory. Abstract-level support: it frames data transfer as the bottleneck and proposes FeFET-based CIM for Boolean and arithmetic operations.
- **First demonstration of in-memory computing crossbar using multi-level Cell FeFET** (Nature Communications 2023, DOI `10.1038/s41467-023-42110-y`) reports a 1FeFET-1R multi-level FeFET crossbar macro for multi-bit MAC using 28 nm HKMG FeFETs. The abstract claims 96.6% handwriting recognition and 91.5% image recognition accuracy.
- **CMOS-compatible compute-in-memory accelerators based on integrated ferroelectric synaptic arrays for CNNs** (Science Advances 2022, DOI `10.1126/sciadv.abm8537`) demonstrates integrated FeTFT synaptic arrays. Its abstract emphasizes the three-terminal FeTFT acting as both nonvolatile memory and access device to reduce leakage/high-power issues in two-terminal crossbars.
- **FeFET Reliability Modeling for In-Memory Computing** (IEEE TED 2023, DOI `10.1109/ted.2023.3313112`) is a reliability-focused source; useful for checking wake-up/fatigue/imprint/variation style risks before overclaiming analog precision.

Interpretation: FeFET/FeTFT is the most mature FeCIM branch in this scan, especially where CMOS-compatible HfO2/HZO stacks and 28 nm HKMG demonstrations appear. However, reported application accuracy or TOPS/W values should be treated as paper-specific until array size, ADC/DAC, write-verify, endurance, retention, and calibration are matched.

### 2. HfO2/HZO ferroelectric materials

HfO2/HZO appears central because it is more CMOS-compatible than many classic ferroelectrics.

Representative evidence:

- **Recent Research for HZO-Based Ferroelectric Memory towards In-Memory Computing Applications** (Electronics 2023, DOI `10.3390/electronics12102297`) reviews HZO ferroelectric memory for IMC and notes low operating voltage, retention, fast switching, analog neuromorphic and digital logic-in-memory possibilities.
- **Ferroelectric Hafnium Oxide Films for In-Memory Computing Applications** (Advanced Electronic Materials 2022, DOI `10.1002/aelm.202200951`) reviews HfO2 materials and device structures including ferroelectric diodes, FeFETs, and FTJs.

Interpretation: HfO2/HZO gives FeCIM its credible semiconductor-integration path. But HZO also brings reliability physics that must be modeled honestly: wake-up, fatigue, imprint, stochastic switching, cycle-to-cycle/device-to-device variation, retention, and temperature dependence.

### 3. Ferroelectric tunnel junctions and ferroelectric diodes

FTJs and ferroelectric diodes use polarization-dependent tunneling/resistance/diode behavior. They are attractive for compact cells and analog synaptic behavior.

Representative evidence:

- **Low-power linear computation using nonlinear ferroelectric tunnel junction memristors** (Nature Electronics 2020, DOI `10.1038/s41928-020-0405-0`) is a key FTJ-CIM paper.
- **High-precision and linear weight updates by subnanosecond pulses in ferroelectric tunnel junction for neuro-inspired computing** (Nature Communications 2022, DOI `10.1038/s41467-022-28303-x`) directly targets precision/linearity of FTJ weight updates.
- **Reconfigurable Compute-In-Memory on Field-Programmable Ferroelectric Diodes** (Nano Letters 2022, DOI `10.1021/acs.nanolett.2c03169`) reports transistor-free CIM for storage, search, and neural-network operations using AlScN ferroelectric diodes; abstract mentions projected sub-0.12 µm² cell footprint at 45 nm and 4-bit neural-network operations.
- **All-ferroelectric implementation of reservoir computing** (Nature Communications 2023, DOI `10.1038/s41467-023-39371-y`) reports volatile and nonvolatile ferroelectric diodes for reservoir and readout networks.

Interpretation: FTJ/diode FeCIM is promising for density and low-current operation, but system validity hinges on selectors/sneak paths, analog-state stability, write variability, and peripheral overhead.

### 4. CAM/TCAM, associative search, and nearest-neighbor workloads

FeCIM is not only neural MAC. CAM/search is a natural fit because nonvolatile state can be compared in place.

Representative evidence:

- **Ferroelectric ternary content-addressable memory for one-shot learning** (Nature Electronics 2019, DOI `10.1038/s41928-019-0321-3`) is a high-impact CAM/one-shot learning paper in the retrieved set.
- **FeFET Multi-Bit Content-Addressable Memories for In-Memory Nearest Neighbor Search** (IEEE TC 2021, DOI `10.1109/tc.2021.3136576`) targets nearest-neighbor search.
- **A Homogeneous FeFET-Based Time-Domain Compute-in-Memory Fabric for Matrix-Vector Multiplication and Associative Search** (IEEE TCAD 2024, DOI `10.1109/tcad.2024.3492994`) proposes one FeFET-based time-domain fabric for both MVM and associative search.

Interpretation: Search/CAM workloads may be better early FeCIM fits than high-precision general neural acceleration because they can exploit nonvolatile matching and tolerate some device-specific constraints.

### 5. Reservoir, hyperdimensional, and temporal computing

Ferroelectric polarization dynamics can be useful, not just a nuisance, for temporal or brain-inspired computing.

Representative evidence:

- **Reservoir computing on a silicon platform with a ferroelectric field-effect transistor** (Communications Engineering 2022, DOI `10.1038/s44172-022-00021-8`) uses ferroelectric polarization dynamics and polarization-charge coupling for short-term memory and nonlinear transform functions.
- **Achieving software-equivalent accuracy for hyperdimensional computing with ferroelectric-based in-memory computing** (Scientific Reports 2022, DOI `10.1038/s41598-022-23116-w`) reports a multi-bit FeFET IMC system for HDC; the abstract claims software-equivalent accuracy plus energy/latency reductions.
- **All-in-Memory Brain-Inspired Computing Using FeFET Synapses** (Frontiers in Electronics 2022, DOI `10.3389/felec.2022.833260`) frames FeFET processing-in-memory for HDC under variation.

Interpretation: These workloads may be well matched to FeCIM because they can tolerate or exploit nonideal dynamics and reduced precision.

### 6. Optimization / annealing

FeCIM is also being explored for combinatorial optimization.

Representative evidence:

- **Ferroelectric compute-in-memory annealer for combinatorial optimization problems** (Nature Communications 2024, DOI `10.1038/s41467-024-46640-x`) reports a FeFET-based CIM annealer for QUBO/Max-Cut-style workloads. Its abstract claims in-situ vector-matrix-vector operations and 75% chip-size saving for Max-Cut matrix compression.

Interpretation: This is a notable non-neural direction. Claims need careful comparison against digital annealers, Ising machines, and memory/peripheral costs.

### 7. 2D ferroelectric devices

2D ferroelectrics and van der Waals heterostructures are a fast-moving exploratory branch.

Representative evidence:

- **Emerging 2D Ferroelectric Devices for In-Sensor and In-Memory Computing** (Advanced Materials 2024, DOI `10.1002/adma.202400332`) reviews 2D ferroelectric devices for in-sensor and in-memory neuromorphic computing.
- **Van der Waals Ferroelectric Semiconductor Field Effect Transistor for In-Memory Computing** (ACS Nano 2023, DOI `10.1021/acsnano.3c01198`) is a device-level in-memory computing paper.
- **Two-dimensional ferroelectric channel transistors integrating ultra-fast memory and neural computing** (Nature Communications 2021, DOI `10.1038/s41467-020-20257-2`) connects 2D ferroelectric channel transistors to memory and neural computing.

Interpretation: 2D FeCIM is promising but less mature for reproducible large-array accelerator claims than HZO/FeFET work.

## Performance claims: safe handling

Retrieved abstracts include large headline numbers, e.g. 96.6% handwriting recognition / 91.5% image recognition in a multi-level FeFET macro, record 13,714 TOPS/W in an IEDM FeFET analog-CIM paper, and large energy/latency improvements for HDC or FTJ acceleration. These should be cited only with paper-specific context. For this repository, avoid statements like “FeCIM achieves X TOPS/W” unless the exact paper, assumptions, array size, precision, and peripheral model are stated.

Safer wording:

> Published FeCIM demonstrations report promising device-, macro-, and workload-specific gains for CIM primitives, neural inference, associative search, reservoir/HDC, and optimization. These claims are not yet universal hardware guarantees and depend strongly on device reliability, array non-idealities, precision, and peripheral overhead.

## Evidence gaps and risks

1. **Analog/multi-level reliability:** retention, endurance, drift, stochastic switching, imprint, wake-up/fatigue, and temperature dependence.
2. **Variation:** device-to-device, cycle-to-cycle, line-edge/material variation, domain stochasticity.
3. **Array non-idealities:** IR drop, sneak/leakage paths, parasitic capacitance, read disturb, write disturb.
4. **Peripheral overhead:** ADC/DAC/sense amplifiers/write drivers/calibration/charge pumps may dominate energy and area.
5. **Benchmark comparability:** different datasets, precisions, array sizes, peripheral assumptions, and baselines make headline numbers hard to compare.
6. **Scaling from cell to system:** device-level low energy does not automatically imply accelerator-level low energy.
7. **Simulation versus measurement:** many architecture results mix measured device characteristics with modeled arrays/systems; label this explicitly.

## Implications for `fecim-lattice-tools`

- Keep the project’s **simulation-only / educational** framing.
- Model FeCIM families separately: FeFET/FeTFT, FTJ/diode, CAM/TCAM, reservoir/HDC, annealer.
- Add provenance tags to any default: “simulation default,” “abstract-reported claim,” “measured device,” “macro demonstration,” or “architecture simulation.”
- Do not present 30 conductance levels, MNIST accuracy, or energy multipliers as general FeCIM facts unless traced to a specific source and assumptions.

## Selected source set

Raw metadata and abstracts are in `selected-openalex-metadata.json`; a compact extraction table is in `evidence-grid.csv`. Search outputs are in `search-openalex-*.txt` and `search-arxiv-*.txt`.
