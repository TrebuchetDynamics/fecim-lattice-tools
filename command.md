# IronLattice-vis Command Reference

**MISSION:** Create world-class visualization demos to help Dr. external research group pitch IronLattice to investors, engineers, and foundry partners.

**PRIMARY REFERENCE:** `ironlattice-transcript.md` (Dr. Tour's Nov 2024 presentation)
**TASK TRACKING:** `TODO.md` (authoritative task list)
**PAPERS:** `opensource/papers/08_Documentation/PAPERS_NEEDED.md`

---

## CURRENT STATUS (2026-01-19)

```
THE IRONLATTICE STORY - 8 DEMOS

Demo 1        Demo 2        Demo 3        Demo 4
"How the      "How we       "What we      "How it fits
memory        compute       can build     in a real
cell works"   in memory"    with it"      chip"
    ↓             ↓             ↓             ↓
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ P-E     │   │Crossbar │   │  MNIST  │   │Peripheral│
│Hysteresis│   │   MVM   │   │  (sim)  │   │ Circuits │
└─────────┘   └─────────┘   └─────────┘   └─────────┘
  ✅ FYNE      ✅ FYNE       ✅ FYNE       ✅ CLI

Demo 5        Demo 6        Demo 7        Demo 8
"1000×        "Scalable     "Real-world   "Why IL
cooler"       3D stack"     challenges"   wins"
    ↓             ↓             ↓             ↓
┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
│ Thermal │   │Multi-   │   │ Non-    │   │Comparison│
│   Map   │   │ Layer   │   │idealities│   │  Chart   │
└─────────┘   └─────────┘   └─────────┘   └─────────┘
  ✅ CLI       🔲 TODO       ✅ in Demo2   🔲 PRIORITY
```

---

## QUICK START - Run All GUIs

```bash
# Demo 1: Ferroelectric Hysteresis (P-E curve, 30 levels)
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis

# Demo 2: Crossbar MVM (IR drop, sneak paths, heatmaps)
cd demo2-crossbar && go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui

# Demo 3: MNIST Neural Network (draw digits, simulation)
# NOTE: IronLattice hardware = 87%, theoretical max = 88%
cd demo3-mnist && go build -o mnist-gui ./cmd/mnist-gui && ./mnist-gui

# Demo 4: Peripheral Circuits (CLI - linearity, timing, power)
cd demo4-circuits && go run ./cmd/circuits --all

# Demo 5: Thermal Simulation (CLI - heat maps)
cd demo5-thermal && go run ./cmd/thermal --realtime

# Run all tests
go test ./...
```

**Build Dependencies (Fyne GUI):**
```bash
# Ubuntu/Debian
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install gcc libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel libXxf86vm-devel
```

---

## TARGET AUDIENCES

| Audience | What They Care About | Key Demos |
|----------|---------------------|-----------|
| **Investors** | ROI, $711B market, 80-90% energy savings | Demo 3, 8 |
| **Engineers** | Physics accuracy, real-world issues | Demo 1, 2, 5, 7 |
| **Foundries** | CMOS compatibility, process flow | Demo 4 |
| **Strategic Partners** | Competitive advantage | Demo 8 |

---

## Demo 1: Hysteresis (Memory Cell Physics) ✅ FYNE GUI

**Story:** "This is how the memory cell works"

**Implemented Features:**
- Mayergoyz Preisach model with full hysteron distribution
- 30 discrete levels clearly shown (LevelIndicator)
- Temperature dependence modeling
- Thread-safe simulation engine
- Real-time P-E hysteresis curve with fade trail
- Material selector (Default HZO, Optimized, IronLattice)
- Waveform selector (Sine, Triangle, Square, Manual)

**Run:**
```bash
cd demo1-hysteresis && go build ./cmd/hysteresis && ./hysteresis
```

**GUI Package:** `demo1-hysteresis/pkg/gui/`
- Custom widgets: `PEPlot`, `LevelIndicator`

**Tests:** 25 passing

---

## Demo 2: Crossbar MVM (Compute-in-Memory) ✅ FYNE GUI

**Story:** "This is how we compute in memory"

**Implemented Features:**
- IR drop analysis with wire resistance modeling
- Sneak path current analysis with visualization
- Interactive heatmap with click-to-select cells
- Three tabbed views: Conductance, IR Drop, Sneak Paths
- 30 discrete conductance levels
- Custom "IronLattice" colormap

**Run:**
```bash
cd demo2-crossbar && go build -o crossbar-gui ./cmd/crossbar-gui && ./crossbar-gui
```

**GUI Package:** `demo2-crossbar/pkg/gui/`
- Custom widgets: `CrossbarHeatmap`, `VectorBarChart`, `DiscreteLevel30Indicator`

**Tests:** 14 passing

---

## Demo 3: MNIST (Neural Network) ✅ FYNE GUI

**Story:** "This is what we can build with it"

> ⚠️ **HARDWARE vs SIMULATION:** IronLattice hardware achieved **87%** with **88% theoretical maximum** (per Dr. Tour). Our simulation uses idealized conditions and may exceed real hardware capabilities.

**Implemented Features:**
- Simulation accuracy varies (idealized, may exceed hardware)
- Interactive 28x28 digit drawing canvas
- Real-time inference as you draw
- Layer activation visualization (input → hidden → output)
- Confusion matrix with clickable cells
- Per-class metrics (precision, recall, F1)
- 30 discrete weight levels

**Run:**
```bash
cd demo3-mnist && go build -o mnist-gui ./cmd/mnist-gui && ./mnist-gui
```

**GUI Package:** `demo3-mnist/pkg/gui/`
- Custom widgets: `DigitCanvas`, `LayerActivationView`, `ConfusionMatrix`, `MetricsPanel`

**Tests:** 9 passing

---

## Demo 4: Peripheral Circuits (System Integration) ✅ CLI

**Story:** "This is how it fits in a real chip"

**Implemented Features:**
- DAC: Digital → Write voltage (5-bit, 30 levels)
- ADC: Analog → Digital level (5-bit)
- TIA: Transimpedance Amplifier
- Charge Pump: 1V → ±1.5V
- INL/DNL linearity analysis
- Timing diagrams
- Power breakdown

**Run:**
```bash
cd demo4-circuits && go run ./cmd/circuits --all
cd demo4-circuits && go run ./cmd/circuits --linearity
cd demo4-circuits && go run ./cmd/circuits --timing
cd demo4-circuits && go run ./cmd/circuits --power
```

**Tests:** 9 passing

---

## Demo 5: Thermal Simulation ✅ CLI

**Story:** "1000× cooler than competition"

**Implemented Features:**
- 2D heat map visualization
- Real-time heat diffusion
- Multi-layer heat coupling
- Hotspot identification
- Thermal throttling warnings
- IronLattice's low-power advantage

**Run:**
```bash
cd demo5-thermal && go run ./cmd/thermal --realtime
```

**Tests:** 17 passing

---

## Demo 8: Technology Comparison 🔲 PRIORITY

**Story:** "Why IronLattice wins vs everyone else"

**Purpose:** The slide Dr. Tour shows investors

```
┌──────────────────────────────────────────────────────────────────┐
│                    COMPUTE PERFORMANCE COMPARISON                 │
├────────────────┬─────────────┬─────────────┬─────────────────────┤
│    Metric      │  DRAM+CPU   │    GPU      │    IronLattice      │
├────────────────┼─────────────┼─────────────┼─────────────────────┤
│ Energy vs NAND │     1×      │    0.1×     │   0.0000001× (10M×) │
│ Speed vs NAND  │     1×      │    100×     │   1,000,000×        │
│ Data Movement  │   O(n²)     │   O(n²)     │        0            │
│ Memory Refresh │   Required  │   Required  │       None          │
│ CMOS Compatible│     Yes     │    Yes      │       Yes           │
│ 30 Analog States│    No      │    No       │       Yes           │
└────────────────┴─────────────┴─────────────┴─────────────────────┘
```

**To Implement:**
- [ ] Side-by-side animated comparison
- [ ] Energy meter visualization (10M× difference)
- [ ] Data center savings calculator
- [ ] Competitive matrix from Dr. Tour's slides

---

## IRONLATTICE SPECS (From Dr. Tour)

| Spec | IronLattice Hardware | Our Simulation | Verification |
|------|---------------------|----------------|--------------|
| Analog states | **30 levels** | ✅ 30 levels | VERIFIED |
| MNIST accuracy | **87%** (88% max) | Variable | ⚠️ SIM ONLY |
| Energy vs NAND | 10M× lower | N/A | UNVERIFIED |
| Energy vs DRAM | 1000× lower | N/A | UNVERIFIED |
| Speed vs NAND | 1M× faster | N/A | UNVERIFIED |
| Data center savings | **80-90%** | N/A | UNVERIFIED |
| CMOS compatible | Standard fab | ✅ Modeled | VERIFIED |
| TRL | **4 (lab only)** | — | VERIFIED |

> ⚠️ Energy claims are from Dr. Tour's presentation and have not been independently verified. IronLattice is at TRL 4 (lab validation), not production.

---

## DR. TOUR'S PHASED MARKET ENTRY

```
PHASE 1               PHASE 2               PHASE 3
┌─────────────┐      ┌─────────────┐      ┌─────────────────┐
│  Replace    │  →   │  Replace    │  →   │  Full Compute-  │
│  NAND Flash │      │  DRAM       │      │  in-Memory      │
└─────────────┘      └─────────────┘      └─────────────────┘
  Easy entry           No refresh           80-90% energy
  No SW changes        1000× lower E        savings
```

---

## PAPER LIBRARY STATUS

**VALID (40+ papers):** See `papers/downloaded/` and `opensource/papers/`

**CORRUPTED (need IEEE access):**
- `Mayergoyz_IEEE_1986.pdf` - Preisach model (CRITICAL)
- `IEEE_CIM_Survey_2023.pdf` - CIM overview
- `Tour_In2Se3_ChemRxiv.pdf` - 2D ferroelectrics

**Full list:** `opensource/papers/08_Documentation/PAPERS_NEEDED.md`

---

## ALL TESTS

```bash
go test ./...   # 110+ tests passing
```

| Package | Tests |
|---------|-------|
| ferroelectric | 20 |
| simulation | 5 |
| crossbar | 14 |
| training (mnist) | 9 |
| peripherals | 9 |
| thermal | 17 |
| multilayer | 17 |
| nonidealities | 20 |
| comparison | 19 |

---

## FILE STRUCTURE

```
ironlattice-vis/
├── demo1-hysteresis/     ✅ P-E curve + Fyne GUI
├── demo2-crossbar/       ✅ Crossbar MVM + Fyne GUI
├── demo3-mnist/          ✅ MNIST (simulation) + Fyne GUI
├── demo4-circuits/       ✅ Peripherals (CLI)
├── demo5-thermal/        ✅ Thermal sim (CLI)
├── demo6-multilayer/     🔲 3D multi-layer
├── demo7-nonidealities/  ✅ (integrated in demo2)
├── demo8-comparison/     🔲 Technology comparison
├── papers/               Scientific papers
├── opensource/papers/    Additional papers + PAPERS_NEEDED.md
├── command.md            This file
├── TODO.md               Strategic task list
└── ironlattice-transcript.md  Dr. Tour's presentation
```

---

## NEXT PRIORITIES

1. **Demo 8: Technology Comparison** - Investor pitch slide
2. **Demo 4 Fyne GUI** - Peripheral circuits visualization
3. **Demo 5 Fyne GUI** - Thermal heat map
4. **Web deployment** - For remote investor presentations
5. **Pitch video** - 2-min demo reel

---

## DR. TOUR QUOTES

> "It's got **30 discrete states**. So it's not 0-1-0-1."

> "We're at **87% validation** here... theoretical is 88%."

> "**Compute in memory** where the same device does the memory and the computation."

> "This could lower the requirements in a data center by **80 to 90%**."

> "Works on a **standard CMOS line** and can translate just like that."

> "There's **no exotic materials** in here. There's no graphene."

---

## WEEBIT NANO PRECEDENT

Dr. Tour's previous spinout (2015):
> "This company Weebit—this is another memory that came out of my lab... it's selling now on the market with three big customers."

**IronLattice is his next one.**

---

*Last updated: 2026-01-19*
