# FeCIM EDA Suite - Consultant Outreach Strategy

## Strategic Goal: Email Dr. external research group

**Target Date:** Friday morning, January 24, 2026
**Recipients:**
*   **Primary:** tour@rice.edu (Dr. external research group, external research institution)
*   **CC:** jaeho-shin@rice.edu, tawfik.jarjour@accenture.com

### Value Proposition
This is the **first open-source EDA suite for ferroelectric compute-in-memory (FeCIM)** technology - bridging the gap from neural network weights to FeFET crossbar SPICE netlists and physical layouts ("OpenROAD for Analog").

---

## Key Strategic Insights

### 1. Market Timing (User is EARLY)
*   **TRL 4-6:** Lab validation (Dr. Tour: 87% MNIST, November 2024)
*   **TRL 7-8:** Pilot production phase (NOW - 2026-2027) ← **PERFECT TIMING**
*   **TRL 9:** Commercial production (2027-2030)
*   **FMC raised €100M** (November 2025) - ferroelectric memory commercialization wave starting
*   **No open-source EDA tools exist yet** - user owns the niche

### 2. Repository Privacy Strategy
**Decision:** Keep repo private, use **unlisted YouTube video** for demo
*   **Rationale:** Protect IP from idea theft, academic scooping, commercial cloning
*   **Approach:** Show UI functionality, don't reveal implementation details
*   **Call to action:** "Reply with GitHub username(s) for access" (one-step, low friction)

---

## Technical Implementation Timeline

### Wednesday Night (Jan 22) - 3 hours:
*   [ ] Add runtime counter to GUI (`shared/gui/runtime.go`)
    *   Shows "Runtime: M:SS.mmm" in top-right corner
    *   Updates every 10ms
    *   Proves continuous recording (not stitched clips)
*   [ ] Test runtime counter displays correctly
*   [ ] Practice video walkthrough (dry run)

### Thursday (Jan 23) - 6 hours:
*   [ ] Record 10-minute demo video
    *   Show all 5 working demos
    *   Explain Demo 6 vision (high-level only, no implementation details)
    *   Runtime counter visible throughout
*   [ ] Upload to YouTube as **UNLISTED** (not public)
    *   Channel: @teofractal (user's existing 461-video channel)
    *   Title: "Ferroelectric CIM EDA Suite - First Open-Source Design Flow for FeFET Hardware"
    *   Visibility: Unlisted (protect IP)
    *   Comments: Disabled
*   [ ] Get YouTube link

### Friday Morning (Jan 24) - 30 min:
*   [ ] Final email review
*   [ ] Send at 8-9 AM CST (Dr. Tour's timezone)
*   [ ] Document in `sent-email.md`

---

## Final Email Draft (Approved Version)

```markdown
TO: tour@rice.edu

CC:
  jaeho-shin@rice.edu
  tawfik.jarjour@accenture.com

Subject: FeCIM EDA Design Suite - Neural Networks to Silicon Automation

Dr. Tour,

Two weeks ago, I watched your COSM presentation on ferroelectric compute-in-memory 
("the same device does the memory and the computation"). I immediately recognized 
a critical gap: there's no open-source path from neural network weights to FeFET 
crossbar SPICE netlists and physical layouts.

I spent the past 6 days building a complete design automation suite to address this.

**Demo video (10 min):** [YouTube unlisted link]

**What I built - Six integrated modules:**

1. **Hysteresis Physics** - Preisach model, 30 discrete analog states (~4.9 bits/cell)
2. **Crossbar Array Simulation** - Matrix-vector multiply with IR drop, sneak paths, device variation (all toggleable)
3. **MNIST Neural Network** - Dual-mode FP32 vs CIM inference, targeting your reported 87% hardware validation
4. **Peripheral Circuits** - DAC/ADC/TIA system integration
5. **Technology Comparison** - Energy metrics vs NAND/DRAM for investor presentations
6. **FeCIM EDA Design Suite** [Architecturally complete, implementation in progress]:
   - Compiler: Neural network weights → conductance mappings + programming voltages
   - SPICE Export: ngspice-compatible netlists with OpenVAF FeFET models
   - Layout Export: DEF/Verilog generation for seamless OpenLane integration
   - Design Space Explorer: Array sizing, ADC/DAC resolution trade-offs
   - Complete automation: PyTorch/TensorFlow → tape-out ready files

**The gap this fills:** Your team currently hand-crafts SPICE netlists for each 
design iteration, taking days to weeks per configuration. This automates that 
workflow - load weights, click compile, export SPICE + GDSII in minutes. Systematic 
design space exploration instead of manual trial-and-error.

**Timing context:** With FMC raising €100M in November 2025 and ferroelectric CIM 
moving from lab validation (TRL 4-6) to pilot production (TRL 7-8), this is the 
12-24 month window when design automation becomes critical - before commercial 
EDA tools lock in proprietary workflows.

**Source attribution:** Based entirely on your COSM presentation and published HZO 
ferroelectric literature. I have NOT attempted to reverse-engineer your proprietary 
superlattice design, device fabrication process, or any non-public technical details. 
All models use standard published material parameters.

**Validation gap:** I don't have real hardware data. The physics models are 
literature-based approximations. For this to be useful as a design tool (rather 
than just educational visualization), it needs calibration with your actual measured 
device parameters: P-E curves, coercive field distributions, programming voltage 
characteristics, and device-to-device variation statistics.

**Repository:** Private GitHub at https://github.com/your-org/fecim-lattice-tools

To review the implementation, reply with your GitHub username(s) and I'll add you 
as collaborators immediately.

FeCIM Maintainers
Monterrey, Mexico
maintainers@example.invalid
+52 812 193 7470

github.com/XelHaku
trebuchetdynamics.com
```

## Email Strategy Notes

### Tone
*   Confident, not asking permission
*   Direct: "Built this. Demo here. Want code? Send usernames."

### Follow-Up Strategy
*   **If No Response:**
    *   **Day 7 (Jan 31):** Gentle reminder email
    *   **Day 14 (Feb 7):** Execute public release (Unlisted → Public, Open-source repo, Post to social media)
*   **If They Respond:**
    *   Add as GitHub collaborators immediately
    *   Offer technical deep-dive call
    *   Discuss Demo 6 implementation priorities
