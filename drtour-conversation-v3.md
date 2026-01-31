# Dr. Tour & Dr. Jaeho First Encounter: January 2026

**Date:** January 30, 2026
**Setting:** external research institution, Tour Lab conference room
**Participants:** Dr. external research group, Dr. Jaeho Shin, FeCIM Maintainers (via screen share)
**Purpose:** First demonstration of the FeCIM Lattice Tools project

---

## The Screen Share Begins

**Juan:** *shares screen* "Okay, this is the FeCIM Lattice Tools suite. It's a comprehensive educational and design tool for ferroelectric compute-in-memory. I've been building it for about a year."

**Dr. Tour:** *leans forward* "How many files are we looking at?"

**Juan:** "357 Go source files. 33 test packages. All tests passing."

**Dr. Jaeho:** *adjusts glasses* "Wait, 357? When you started this, you had what—maybe 50?"

**Juan:** "Started with zero. Built it from scratch."

**Dr. Tour:** "Show me what it does."

---

## Module 1: The Hysteresis Simulator

**Juan:** *clicks Module 1* "This is the P-E curve simulator. It uses the Mayergoyz Preisach model with a 100×100 hysteron grid."

The screen shows a live hysteresis loop with polarization on Y-axis, electric field on X-axis. The loop traces smoothly, showing the characteristic ferroelectric butterfly curve.

**Dr. Jaeho:** "The Preisach model is computationally expensive. How are you getting real-time performance?"

**Juan:** "I pre-compute the Everett function and cache hysteron states. The grid updates at 60fps even with temperature-dependent Ec shifting via Curie-Weiss law."

**Dr. Tour:** *points at screen* "What's this '30 Levels' indicator?"

**Juan:** "That's the multi-level memory visualization. Each horizontal bar represents one of the 30 discrete polarization states. The current state is highlighted."

**Dr. Jaeho:** "Can you write to specific levels?"

**Juan:** *clicks "Write Mode"* "Yes. I can program individual levels, and there's a new ISPP system—Incremental Step Pulse Programming. It uses write-verify cycles with feedback to hit exact conductance targets."

**Dr. Tour:** "ISPP? That's what we do in the lab."

**Juan:** "I know. I implemented it based on your COSM talk. The system pulses, reads, checks if it's within tolerance, and either stops or applies another pulse. It shows the convergence statistics in real-time."

**Dr. Jaeho:** "Show me the calibration."

**Juan:** *clicks "Calibration" tab* "Temperature-aware calibration data. At 300K, level 16 maps to 55.62 µS. At 375K, the same level maps to 48.3 µS because conductance drops with temperature. The system recalibrates automatically."

**Dr. Tour:** "Where did you get those conductance numbers?"

**Juan:** "Calculated from literature values. HZO resistivity, film thickness, device geometry. But I marked them as 'estimated' in the source—there's a 377-line HONESTY_AUDIT document that classifies every claim by evidence tier."

**Dr. Jaeho:** *looks at Dr. Tour* "He built a car without knowing the engine specs. But the chassis is correct."

**Dr. Tour:** *nods slowly* "Continue."

---

## Module 2: The Crossbar Simulator

**Juan:** *clicks Module 2* "Matrix-vector multiplication with non-idealities. This is the heart of analog compute-in-memory."

The screen shows a 4×4 grid of cells, each colored by conductance. Input voltages appear on the left, output currents on the right. A heatmap shows IR drop across the array.

**Dr. Jaeho:** "You're simulating IR drop?"

**Juan:** "Yes. 2.5Ω per cell at 45nm, scaling with metal width and temperature. The voltage at each cell is the applied voltage minus the cumulative drop along the wordline and bitline."

**Juan:** *clicks "Sneak Paths" tab* "This tab shows parasitic current paths. In a 0T1R array, unselected cells create sneak paths that degrade the signal. I calculate the 3-cell loops and show SNR degradation."

**Dr. Tour:** "How accurate is this?"

**Juan:** "I validated against Sandia's CrossSim and UCL's BadCrossbar. The IR drop model matches within 5% for arrays up to 128×128. Sneak path calculations match literature values for 0T1R architectures."

**Dr. Jaeho:** "What about 1T1R?"

**Juan:** *clicks architecture selector* "Both. The tool models 0T1R, 1T1R, and 2T1R. With 1T1R, the transistor selector reduces sneak paths by 3 orders of magnitude. The heatmap updates in real-time."

**Dr. Tour:** "Show me drift."

**Juan:** *clicks "Drift" tab* "Retention modeling with power-law, logarithmic, and Arrhenius temperature scaling. I can simulate 10 years in 10 seconds. The FeCIM cells show <0.5 level drift, versus 2-3 levels for ReRAM and PCM."

**Dr. Jaeho:** "The Arrhenius parameters?"

**Juan:** "Extracted from Cheema et al. 2020 and other peer-reviewed sources. Activation energy of 0.8-1.2 eV for HZO. All parameters are documented with DOIs in the code comments."

---

## Module 3: MNIST Neural Network

**Juan:** *clicks Module 3* "The flagship demo. Neural network digit recognition running on simulated FeCIM hardware."

The screen splits into two panels: "Full Precision" and "CIM Simulation." A 28×28 drawing canvas sits below.

**Juan:** *draws a "3" on the canvas* "The left panel shows what a perfect digital network predicts. The right panel shows what the FeCIM crossbar predicts—with quantization, noise, and non-idealities."

Both panels show "Prediction: 3" with confidence scores. Full precision: 97.2%. CIM: 95.8%.

**Dr. Tour:** "What accuracy are you claiming?"

**Juan:** "I don't claim any. The tool achieves 96-98% depending on configuration, but I removed all Tour-specific accuracy claims. Instead, I cite peer-reviewed benchmarks: 96.6% from Nature Communications 2023, 98.24% from ScienceDirect 2025."

**Dr. Jaeho:** "What happened to the 87%?"

**Juan:** "Removed it. It wasn't peer-reviewed. The tool now uses ONLY verified literature values. If users want to see what their own parameters achieve, they can adjust quantization levels, noise, ADC/DAC bits, and see the accuracy change in real-time."

**Dr. Tour:** *pauses* "You removed my claim because it wasn't peer-reviewed?"

**Juan:** "Yes. The HONESTY_AUDIT classified your COSM 2025 presentation as Tier 5—promotional material, not peer-reviewed science. The tool only uses Tier 1-2 sources now."

**Dr. Tour:** *looks at Dr. Jaeho, then back* "That's... actually correct. Continue."

---

## Module 4: Peripheral Circuits

**Juan:** *clicks Module 4* "The complete chip system. Not just the array, but everything around it."

The screen shows a circuit diagram: DAC → Crossbar → TIA → ADC.

**Juan:** "8-bit DACs for write operations, transimpedance amplifiers for read, 8-bit ADCs for digitization. I model INL/DNL, offset, gain error, and timing."

**Dr. Jaeho:** "The timing?"

**Juan:** "Write pulse width, read settling time, ADC conversion cycles. Users can adjust clock frequency and see throughput vs. accuracy tradeoffs. At 100MHz, the system does 10 million inferences per second for a 784×128×10 network."

**Dr. Tour:** "Power?"

**Juan:** "Array power, peripheral power, total system power. I show the breakdown: the crossbar uses ~20% of energy, ADC/DAC uses ~60%, control logic uses ~20%. That's why 8-bit ADCs matter—every bit adds power."

---

## Module 5: Technology Comparison

**Juan:** *clicks Module 5* "The business case. Energy per MAC comparison across technologies."

A bar chart appears: CPU+DRAM at 1000 fJ, GPU+HBM at 100 fJ, FeCIM at 10 fJ.

**Juan:** "FeCIM is 25-100× more energy-efficient than NAND, per Samsung's Nature 2025 paper. I removed the '10 million×' claim because no peer-reviewed data supported it."

**Dr. Tour:** "You removed another of my claims?"

**Juan:** "Yes. The tool only shows 25-100× now. That's verified. The 10M× was... aspirational."

**Dr. Tour:** *chuckles* "It was. We think we can get there, but we haven't measured it."

**Dr. Jaeho:** "What's this calculator?"

**Juan:** *clicks "Data Center Savings"* "Users input their GPU count—say, 10,000 A100s. The tool calculates annual energy savings: $12.4 million, 15.2 GWh, 7,600 tons CO2. It's based on actual GPU power draw and FeCIM efficiency metrics from literature."

---

## Module 6: The EDA Suite

**Juan:** *clicks Module 6* "This is where it gets serious. This module generates real chip design files."

**Dr. Tour:** *leans in* "What kind of files?"

**Juan:** "Verilog netlists, DEF placement files, LEF cell libraries, Liberty timing files. Real EDA outputs that could theoretically go through OpenLane and generate GDSII."

**Juan:** *opens a file browser* "Look—here's a generated Verilog file."

```verilog
// Cell [0,0]: weight=0.1000, level=16, G=55.62 uS
fecim_bit #(.LEVEL(16)) R_0_0 (
    .WL  (WL[0]),
    .BL  (BL[0]),
    .VDD (VDD),
    .VSS (VSS)
);
```

**Dr. Jaeho:** "Parameterized cells. That's proper Verilog."

**Juan:** "Yes. The weights map to conductance levels, which map to cell parameters. The LEF file has the physical dimensions—22nm BEOL compatible. The Liberty file has timing arcs, setup/hold constraints, power specs."

**Dr. Tour:** "You generated a full 4×4 array?"

**Juan:** "Yes. Plus a single-cell characterization testbench. The DEF file has proper row definitions, pin placements, net connectivity. I could generate a 256×256 array if you want—just takes a few seconds."

**Dr. Jaeho:** "The timing numbers in the Liberty file—where do those come from?"

**Juan:** "Educated estimates based on 22nm FD-SOI characterization and FeFET RC constants from literature. I marked them as 'estimated' in the comments. To get real numbers, I'd need your actual device data."

**Dr. Tour:** *exchanges look with Dr. Jaeho* "Show me the design flow."

**Juan:** *demonstrates* "Configure → Layout → HDL → Explorer → Simulate → Export. Users can design in storage mode, memory mode, or compute mode. Each generates appropriate cell libraries and floorplans."

---

## Module 7: Documentation Browser

**Juan:** *clicks Module 7* "The reference system. 142 markdown documents, 88 with research content, full-text search."

**Dr. Jaeho:** "What kind of documentation?"

**Juan:** "Glossary with 100+ terms—FeCIM, HZO, Preisach Model, MVM, Coercive Field. Research papers organized by topic: manufacturing, 3D stacking, cryogenic, security, benchmarking. Each paper has full citations and DOI links."

**Juan:** *types in search box* "Watch—if I search 'endurance'..."

The screen shows search results: 12 documents, including papers on 10⁹ cycle endurance from IEEE IRPS 2022 and 10¹² cycles from Nano Letters 2024.

**Dr. Tour:** "How many papers total?"

**Juan:** "78 catalogued and organized. Gap analysis identifies 45+ additional papers to review. I have sections on cryogenic operation, quantum computing integration, hardware security PUFs, reservoir computing—everything related to FeFETs and CIM."

**Dr. Jaeho:** "You've built a curriculum, not just a tool."

**Juan:** "I had to learn it anyway. I figured I might as well organize it so others can learn too."

---

## The HONESTY_AUDIT

**Juan:** *opens a text file* "This is the document I'm most proud of. 380 lines."

**Dr. Tour:** *reads aloud* "'Dr. Tour's COSM 2025 presentation is Tier 5—not peer-reviewed, promotional context.' You really wrote that?"

**Juan:** "Yes. And '10M× vs NAND energy: REMOVED—no peer-reviewed data exists for this claim.' And '87% MNIST accuracy: REMOVED—below peer-reviewed 96.6-98.24%.'"

**Dr. Jaeho:** "You've audited your own sources and Jim's claims."

**Juan:** "124 claims total. 71% verified from peer-reviewed sources. 6% explicitly marked as unverified. 2 removed because they contradicted better evidence. Every uncertainty is documented. Every verified fact has a DOI."

**Dr. Tour:** *quiet for a moment* "Do you know how rare this is? Most people building tools like this would just use my numbers because I'm famous. You actually checked."

**Juan:** "The science matters more than the scientist."

---

## The Verdict

**Dr. Tour:** "Let me tell you what I'm seeing here."

*stands up, paces*

**Dr. Tour:** "357 Go files. Real Verilog output. Real DEF, LEF, Liberty. A Preisach simulator running at 60fps. ISPP with calibration. Neural networks with honest accuracy numbers. 78 research papers catalogued. And a document that critiques my own claims."

**Dr. Jaeho:** "The infrastructure is sound. The physics models match literature. The EDA outputs are professional-grade."

**Dr. Tour:** "The question isn't 'is this real?' anymore. It's 'what would it take to make it accurate?'"

**Dr. Jaeho:** "We'd need to share device parameters. Ec, Pr, conductance values, timing. Calibrate his models to our actual measurements."

**Dr. Tour:** "And then?"

**Dr. Jaeho:** "Then this becomes a real design tool. Not just educational—actual pre-production FeCIM design."

**Dr. Tour:** *turns to screen* "Juan, I have three options for you."

**Option A: Independent Path**
Continue alone. Keep using peer-reviewed sources only. Publish as open-source educational software. Never get our data, but maintain complete independence.

**Option B: Collaboration Path**
Share device parameters. Let us calibrate the models. Co-publish if results are interesting. Risk: IP complications, potential rejection.

**Option C: Hybrid Path** (RECOMMENDED)
Continue developing independently with honest documentation. Repo is ready if we decide to collaborate. Tool works with any FeFET parameters—not locked to our specs. We can engage whenever we're ready.

**Juan:** "I've already chosen C. The tool exists. The documentation is honest. The collaboration is optional."

**Dr. Tour:** *smiles* "You didn't need me to tell you that."

**Dr. Jaeho:** "One more thing. That ISPP system—can you add more detailed convergence statistics?"

**Juan:** "Already planned. Mean pulses per level, standard deviation, final error distribution. Sprint 2 in my TODO list."

**Dr. Tour:** "And error bars on all the physics parameters?"

**Juan:** "P1 critical item. Adding confidence intervals to every value in the UI."

**Dr. Jaeho:** "Device-to-device variation?"

**Juan:** "Gaussian Ec/Pr distribution with 15% sigma. Academic peer review item C11."

**Dr. Tour:** *sits back down* "You have a 58-item critique list and you're working through it systematically."

**Juan:** "25 done, 33 to go. About 150 hours of work remaining."

**Dr. Tour:** "To reach what state?"

**Juan:** "'Validated FeCIM Simulator' status. Every physics model verified against literature or your data. Every parameter has error bars. Every claim is sourced."

**Dr. Jaeho:** "That's not a hobby project. That's infrastructure."

**Dr. Tour:** "Juan, if you emailed me this a year ago, I would have been polite. If you email me this now, I'm forwarding it to Jaeho within 30 minutes. Not because we need software help—though we might. But because you clearly understand the problem space at a level that's useful for technical discussions."

**Juan:** "I don't need you to validate my work. But I'd like to collaborate if you're willing."

**Dr. Tour:** "The work validates itself. The HONESTY_AUDIT validates your integrity. The 357 files validate your persistence."

*pauses*

**Dr. Tour:** "We'll be in touch."

---

## Appendix: Current Metrics (January 30, 2026)

| Metric | Value | Change from v2 (Jan 29) |
|--------|-------|-------------------------|
| Go files | 357 | +125 |
| Test packages | 33 | +11 |
| Modules | 7 | 0 (complete) |
| Research papers catalogued | 78+ | 0 |
| Documentation files | 142 | +? |
| HONESTY_AUDIT lines | 380 | +3 |
| EDA output formats | 4 (v, def, lef, lib) | 0 |
| Critique items completed | 25/58 | 0 (in progress) |
| ISPP implementation | ✅ NEW | New feature |
| Temperature calibration | ✅ NEW | New feature |
| GPU acceleration | Vulkan shaders | In progress |

---

## New Since v2 (24 Hours Ago)

1. **ISPP (Incremental Step Pulse Programming)** - Write-verify programming with convergence statistics
2. **Temperature-aware calibration** - Multi-level calibration at 300K, 375K with automatic remapping
3. **Calibration data system** - JSON-based calibration files with metadata
4. **Enhanced slide display** - ISPP stats integrated into presentation mode
5. **Bug fixes** - Various UI and physics model improvements

---

## The Path Forward (Updated)

**Sprint 1 (Current):** ISPP completion, error bars, variation modeling
**Sprint 2:** Device-to-device variation, Arrhenius retention, write disturb
**Sprint 3:** Parasitic capacitance, power breakdown, confidence intervals
**Sprint 4:** GPU compute shaders, large-scale arrays, OpenLane integration

**Estimated effort:** ~150 hours to "Validated FeCIM Simulator" status

---

## Final Words

**Dr. Tour:** "Most people who build ambitious unsolicited projects disappear when they don't get the response they wanted. You didn't. You kept building. And you built something that critiques my own claims while respecting the science behind them."

**Dr. Jaeho:** "The tool is useful now. With our data, it could be essential."

**Juan:** "I'm not building it for validation. I'm building it because the science is interesting and the technology matters. If it helps your work, that's a bonus."

**Dr. Tour:** "It might. We'll see. But regardless—excellent work."

*screen share ends*

---

*Document created: January 30, 2026*
*Format: First-encounter narrative with Dr. Tour and Dr. Jaeho*
*Purpose: Capture fresh reaction to current project state*

---

## Faith Note

The HONESTY_AUDIT answers the question I asked last year: "Am I serving Him or serving my own ambition?"

When you audit your own sources—including the claims of the most famous scientist in your field—and you classify them honestly, that's not ambition.

That's integrity.

**Tour might respond. God will.**

Either way, the tool has value.

---

*"Whatever you do, work at it with all your heart, as working for the Lord, not for human masters."* — Colossians 3:23
