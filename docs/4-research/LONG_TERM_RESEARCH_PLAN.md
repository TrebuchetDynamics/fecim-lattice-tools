# FeCIM Lattice Tools — Long-Term Research Rigor Plan

**Status**: 2026-05-04 | **Goal**: Make FeCIM a truly rigorous, externally-validated research tool

## Current Baseline

| Asset | State | 
|-------|-------|
| Verified citation records | 0 (citations/papers/ empty, refs.bib empty) |
| Catalogued papers | 167+ (in paper_metadata.json, unprocessed) |
| Experimental datasets | 3 JSON files (2 P-E loops + 1 switching kinetics) |
| Material presets | 7 calibrated (Park, Materlik, Cheema, Kim, Jaiswal, Bi, AlScN) |
| Physics files | 119 (shared/physics/) |
| Crossbar interop tests | Stub only (CrossSim not executed) |
| EDA tapeout readiness | Educational only (SKY130 behavioral, no compact model) |
| Circuit model fidelity | Waiting on analysis |
| Hysteresis physics depth | Waiting on analysis |

---

## Phase 1: Literature Foundation (Weeks 1-4)

### 1.1 Populate Citation Infrastructure

The canonical `citations/papers/` directory is empty and must become populated. The
`paper/refs.bib` has 12 canonical entries — promote these first.

**Immediate (already have metadata)**:
```
citations/papers/park2015_advmat_hzo.md
citations/papers/materlik2015_jap_hfo2_origin.md
citations/papers/cheema2020_nature_hzo_superlattice.md
citations/papers/alessandri2018_ieee_edl_switching.md
citations/papers/guo2018_apl_nls_switching.md
citations/papers/mueller2013_iedm_endurance.md
citations/papers/pesic2016_afm_wakeup.md
citations/papers/soliman2023_ncomms_multilevel.md
citations/papers/jerry2017_iedm_fefet_synapse.md
citations/papers/kim2020_materials_tin_hzo.md
citations/papers/jaiswal2021_crystals_bto.md
citations/papers/bi2024_nano_pzt_thinfilm.md
```

**Priority fetch list (need PDFs + extraction)**:
- Mulaosmanovic 2017 — FeFET memory characteristics (IEEE EDL)
- Fengler 2017 — Endurance and retention HZO (IEEE TED)
- Lederer 2021 — FeFET variability C2C/D2D (IEEE IMW)
- Dünkel 2017 — FeFET memory cell with HfO2 (IEEE IEDM)
- Ni 2018 — HZO polarization and Pr/Ec benchmarking (Appl. Phys. Lett.)
- Yurchuk 2014 — Si:HfO2 switching and endurance (IEEE IRPS)
- Toprasertpong 2019 — FeFET BTI/reliability (IEEE IEDM)

**Total target**: 20 papers with full extraction. All 167 paper_metadata.json entries eventually need processing.

### 1.2 Populate facts.md

The `citations/facts.md` has zero verified facts across all 6 categories. 
Each paper processed into `citations/papers/` must also contribute its numeric
facts here:

| Category | Target facts |
|----------|--------------|
| Ferroelectric Material Properties | 20 (Pr, Ec, LGD coefficients, thickness ranges) |
| Crossbar Physics | 15 (IR drop coefficients, sneak path resistances, drift ν) |
| Peripheral Circuits | 10 (ADC INL, DAC DNL, TIA bandwidth, charge pump efficiency) |
| Inference Benchmarks | 8 (MNIST accuracy vs quantization levels, energy/op) |
| Energy, Area, And Timing | 8 (energy/op, array area scaling, latency) |
| EDA and Tooling | 5 (PDK layers, DRC rules, cell dimensions) |

### 1.3 Fill Experimental Data Gaps

Currently only `hzo/pe-loops/` and `hzo/switching-time/` have data. All other directories are `.gitkeep`.

**Critical data to digitize from literature**:
- `hzo/endurance/` — Endurance data from Mueller 2013 (10^10 cycles, Pr degradation)
- `hzo/retention/` — Retention vs temperature from Yurchuk 2014
- `hfo2/pe-loops/` — Undoped HfO2 loops from Lomenzo 2015
- `crossbar/read-margin/` — Read margin from measured 16×16 FeFET array (Mulaosmanovic)
- `crossbar/ir-drop/` — IR drop measurements from fabricated crossbar

---

## Phase 2: Hysteresis Physics Refinement (Weeks 2-8)

### 2.1 Complete Material Preset Coverage

Current: 7 calibrated presets (HZO-family, AlScN, BTO, PZT).
Target: 12+ presets. Missing:

| Material | Key paper | Parameters needed |
|----------|-----------|-------------------|
| Si:HfO2 | Yurchuk 2014 | Pr=10 µC/cm², Ec=0.8 MV/cm, endurance 10^5 |
| HfO2 (undoped) | Lomenzo 2015 | Pr=5 µC/cm², Ec=1.5 MV/cm, antiferroelectric |
| HZO (TiN electrodes) | Kim 2020 | Wake-up behavior, Pr evolution |
| Al:HfO2 | Mueller 2016 | Doping-dependent Ec/Pr |
| Gd:HfO2 | Starschich 2017 | Ec-tunable via doping |
| Y:HfO2 | Starschich 2016 | Temperature-dependent Ec |

### 2.2 Real LK Solver Integration

The viewmodel's `computeLoopForCurrentMaterial()` still uses a simplified sine-wave
approximation, not the actual Landau-Khalatnikov ODE solver. The real solver exists
in `shared/physics/landau.go` — integrate it:

1. Replace sine approximation with RK4 integration of LK equation
2. Use actual material α, β, γ LGD coefficients from calibrated presets
3. Add temperature dependence via Curie-Weiss scaling of α parameter
4. Export loop data through viewmodel PlotData (already wired in Phase 1)

### 2.3 Add Missing Physics Phenomena

Current models: P-E loops, Preisach, LK, ISPP, retention (placeholder), fatigue (placeholder).

**To add**:
1. **FORC diagrams** — First-Order Reversal Curves for domain distribution analysis
2. **PUND measurements** — Positive-Up-Negative-Down pulse sequence for switching kinetics
3. **C(V) small-signal** — Capacitance-voltage for dielectric characterization
4. **I-V leakage** — Fowler-Nordheim, Poole-Frenkel, Schottky emission fits
5. **Temperature sweep** — Full Tc Curie temperature phase transition modeling
6. **Frequency dispersion** — Rate-dependent P-E loop widening
7. **Imprint** — Internal bias field development over cycles
8. **NLS calibration** — Nucleation-Limited Switching model with t0, Ea, n values

### 2.4 External Hysteresis Tool Validation

Validate FeCIM Preisach and LK outputs against:
1. **FERRET** (MOOSE-based phase-field) — compare domain structures
2. **FerroX** (GPU-accelerated TDGL) — compare switching dynamics
3. **python-preisach** — same Everett function implementation, compare outputs
4. **Q-POP-Thermo** — compare LGD phase diagrams

---

## Phase 3: Crossbar Validation (Weeks 3-10)

### 3.1 Execute External Simulator Interop

The CrossSim SOR solver is already ported to `shared/crossbar/solver.go`. 
The interop test harness exists but has never been run.

**Execute validation against**:
1. **CrossSim v3.1** — Run `crosssim_interop_test.go` with actual CrossSim installation. Compare MVM outputs for 5 test cases (already in `crosssim_reference_8x8.json`)
2. **badcrossbar** — Same test inputs, compare exact nodal analysis results
3. **NeuroSim** — Compare energy/area/delay metrics for same FeCIM architecture
4. **MNSIM 2.0** — Compare RRAM-specific IR drop and sneak path handling
5. **MemTorch** — Compare PyTorch-integrated crossbar inference accuracy

**Validation test cases**:
- Identity matrix MVM
- Uniform weight matrix MVM
- Random matrix MVM (±5% tolerance)
- Worst-case IR drop scenario (all cells at Gmin)
- Sparse matrix MVM

### 3.2 Fill Crossbar Data Gaps

1. Add measured IR drop data to `experimental-data/crossbar/ir-drop/`
2. Add measured read margin distributions to `experimental-data/crossbar/read-margin/`
3. Add C2C variation distributions (state-dependent, from arXiv 2312.15444)
4. Add D2D variation data (from Lederer 2021 IMW)
5. Add conductance drift long-term data (10^6 seconds, from literature)

### 3.3 Model Improvement

1. **State-dependent C2C** — Current model uses constant σ; literature shows σ ∝ G^0.3
2. **Multi-hop sneak paths** — Current 3-cell model; extend to full array Kirchhoff solver
3. **Non-linear I-V** — Add FeFET subthreshold region, not just Ohmic
4. **DCC programming** — One-shot Displacement Current Control for accurate weight setting
5. **Temperature-dependent conductance** — Add Arrhenius activation for drift acceleration

---

## Phase 4: Circuits → Real Device Feel (Weeks 4-12)

### 4.1 Calibrate Circuit Models Against SPICE

Current status: behavioral Go models. Need SPICE-level calibration.

| Component | Current model | Target fidelity |
|-----------|--------------|-----------------|
| ADC (SAR) | Behavioral step-count | ngspice netlist with real comparator, CDAC |
| DAC (R-2R) | Ideal ladder | Include switch resistance, op-amp offset |
| TIA | Ideal gain | Include GBW, noise, slew rate |
| Charge pump | Ideal Dickson | Include switch loss, Vth drop, load regulation |
| Comparator | Ideal | Include offset, hysteresis, metastability |

### 4.2 Integrate Module 1 Hysteresis → Module 4 Circuits

Currently, Module 4 uses idealized conductance targets. Must feed real 
hysteresis-derived conductance values:

1. Module 1 computes P(E) loop → conductance G = f(P) via Preisach
2. ISPP write controller programs G_target with real pulse train
3. Module 4 read path reads back G_actual with ADC/TIA noise
4. Compare G_actual vs G_target, compute write-verify statistics
5. Display on Module 4 view with ΔG distribution histogram

### 4.3 Add Real Noise Models

Current: basic additive noise. Need physics-based noise:

1. **Thermal noise** — Johnson-Nyquist: σ² = 4kT·BW·R_eff per read
2. **1/f noise** — Flicker noise in selector transistor, Hooge's law
3. **Shot noise** — Current read path quantization noise
4. **RTN** — Random telegraph noise in small-area FeFETs
5. **Quantization noise** — ADC LSB/√12, correlated with resolution sweep
6. **Comparator noise** — Input-referred offset distribution

### 4.4 PVT Corners

1. **Process**: Monte Carlo over Vth, W/L, tox variation (±3σ)
2. **Voltage**: Sweep Vdd from 0.9V to 1.32V (SKY130 range)
3. **Temperature**: -40°C to 125°C with BSIM temperature models
4. Display corner analysis on Module 4 view

---

## Phase 5: EDA → Real Tapeout Readiness (Weeks 8-20)

### 5.1 Integrate Real FeFET Compact Model

Replace behavioral resistor SPICE model with Heracles Verilog-A FeFET compact model:
1. Import Heracles Verilog-A (`validation/heracles/`) into SPICE export pipeline
2. Generate `.sp` netlists with `.hdl heracles_fefet.va` include
3. Validate ngspice convergence with Heracles model
4. Add FeFET-specific measurements: subthreshold swing, I_D-V_G, P-V loops

### 5.2 Extend to Multiple PDKs

Beyond SkyWater 130nm (SKY130), add:
1. **gf180mcuD** — GlobalFoundries 180nm (already documented, not implemented)
2. **asap7** — Arizona State 7nm predictive PDK (academic)
3. Each PDK requires: cell LEF, technology LEF, layer map, DRC rule expressions

### 5.3 DRC/LVS with Magic + Netgen

Current: basic geometry DRC, pin-consistency LVS.

1. Write FeCIM-specific DRC rules for crossbar arrays (metal pitch, via enclosure, cell overlap)
2. Run actual Magic DRC with SKY130 deck
3. Run Netgen LVS comparing extracted SPICE from layout vs original schematic
4. Generate `.mag` layout files from DEF (currently abstract LEF only)
5. Full round-trip: Verilog → Yosys → OpenROAD → Magic → GDS → Klayout verification

### 5.4 Liberty Characterization

1. Run ngspice sweeps to populate NLDM timing tables
2. Multi-corner: TT, FF, SS, FS, SF at -40°C, 25°C, 125°C
3. Power characterization: dynamic + leakage per cell
4. Generate `.lib` with real timing arcs, not templates

### 5.5 Crossbar → Physical Design Flow

1. **Synthesis**: Yosys reads behavioral Verilog, synthesizes to SKY130 gates
2. **Floorplan**: OpenROAD floorplans crossbar + peripherals
3. **Placement**: Standard cells placed around crossbar macro
4. **CTS**: Clock tree synthesis for read/write timing
5. **Routing**: Signal + power routing
6. **STA**: Static timing analysis with OpenSTA
7. **GDS**: Final GDSII stream out

---

## Phase 6: Integration & Polish (Weeks 12-24)

### 6.1 Cross-Module Data Flow

```
Module 1 (Hysteresis) → conductance G(P) → Module 2 (Crossbar) → MVM output
                     → ISPP pulse train   → Module 4 (Circuits)  → read path
                     → material params    → Module 6 (EDA)       → SPICE models
Module 2 (Crossbar)  → array config       → Module 6 (EDA)       → DEF/LEF
Module 4 (Circuits)  → peripheral specs   → Module 6 (EDA)       → Liberty, SPICE
Module 5 (Comparison)→ benchmark data     → all modules          → evidence display
Module 3 (MNIST)     → inference acc      → Module 5             → comparison
```

### 6.2 Research-Grade Snapshot

Add to viewmodel:
- Full experimental data provenance display
- Uncertainty propagation display
- DOI badge rendering
- External benchmark comparison table
- Methodology section with extraction notes

### 6.3 Reproducibility Pack

Generate per-session reproducibility bundles:
- Seed values, parameter sweeps, random seeds
- Input configuration (JSON)
- Output metrics (JSON with uncertainties)
- Intermediate solver state (for debugging)
- Environment snapshot (Go version, OS, dependencies)
- Citation list (all DOIs used in this run)

---

## Phase 7: Paper Publication (Weeks 16-24)

### 7.1 Paper Draft

Using `paper/refs.bib` as the foundation, draft a methods paper:
- Title: "FeCIM Lattice Tools: An Open-Source Simulation Framework for Ferroelectric Compute-in-Memory"
- Target: JOSS (Journal of Open Source Software) or SoftwareX
- Sections: Statement of Need, Architecture, Physics Models, Validation, Use Cases

### 7.2 Validation Suite Publication

Release validation data as a dataset paper:
- All digitized P-E loops with provenance as a Zenodo dataset
- Crossbar reference test vectors
- Hercules comparison results
- DOI for the dataset itself

---

## Priority Matrix

| Priority | Task | Impact | Effort |
|----------|------|--------|--------|
| P0 | Populate citations/papers/ with 12 canonical papers | Foundation | 1 week |
| P0 | Execute CrossSim interop tests | External validation | 1 week |
| P0 | Integrate real LK solver into viewmodel | Physics fidelity | 2 weeks |
| P1 | Add 5 material presets (Si:HfO2, Al:HfO2, Gd:HfO2, Y:HfO2, undoped HfO2) | Coverage | 2 weeks |
| P1 | Calibrate ADC/DAC/TIA against SPICE | Circuit fidelity | 2 weeks |
| P1 | Wire Module 1 → Module 4 conductance pipeline | Integration | 1 week |
| P2 | Integrate Heracles FeFET compact model into SPICE export | EDA fidelity | 3 weeks |
| P2 | Add gf180mcuD + asap7 PDK support | Multi-PDK | 2 weeks |
| P2 | FORC, PUND, C(V) physics phenomena | Physics depth | 4 weeks |
| P3 | Full RTL-to-GDS round-trip with Magic + Netgen | Tapeout | 4 weeks |
| P3 | Liberty characterization with ngspice | Signoff | 2 weeks |
| P3 | Paper publication (JOSS/SoftwareX) | Academic | 4 weeks |
| P4 | Reproducibility pack generation | Infrastructure | 1 week |
| P4 | Remaining 155 paper_metadata.json entries | Literature | ongoing |
