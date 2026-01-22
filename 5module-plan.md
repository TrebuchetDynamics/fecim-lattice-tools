# Complete 5-Demo Consolidation Plan - Option A

**Project:** Ferroelectric CIM Visualizer  
**Author:** @XelHaku  
**Date:** 2026-01-21  
**Goal:** Transform 8 scattered demos into 5 world-class, cohesive demonstrations

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [The 5-Demo Architecture](#2-the-5-demo-architecture)
3. [Phase 1: Consolidation (Week 1-2)](#3-phase-1-consolidation-week-1-2)
4. [Phase 2: Demo 3 Enhancement (Week 3-6)](#4-phase-2-demo-3-enhancement-week-3-6)
5. [Phase 3: Demo 5 Creation (Week 7-8)](#5-phase-3-demo-5-creation-week-7-8)
6. [Phase 4: Polish & Documentation (Week 9-10)](#6-phase-4-polish--documentation-week-9-10)
7. [Implementation Details](#7-implementation-details)
8. [Testing Strategy](#8-testing-strategy)
9. [Migration Guide](#9-migration-guide)
10. [Success Metrics](#10-success-metrics)

---

## 1. Executive Summary

### Current State
- 8 demos listed in README
- 4 fully working (Demos 1-4)
- 2 partially working (Demos 5, 7)
- 2 not implemented (Demos 6, 8)
- Code duplication between demos
- Confusing narrative for investors/users

### Target State (10 weeks)
- **5 world-class demos** with clear narrative
- **Demo 3 (MNIST)** as flagship showcase
- **Demo 5 (Comparison)** as technical briefing tool
- Clean code reuse (Demo 3 imports Demo 2)
- Professional documentation
- Single command to run all demos

### The 5-Demo Story

```
THE FeCIM NARRATIVE
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

"How does the          "How do we          "What can we
 memory cell work?"     compute with it?"   build with it?"
      ↓                      ↓                    ↓
┌──────────┐               │  DEMO 2  │    →    │  DEMO 3  │
│Hysteresis│          │ Crossbar │         │  MNIST   │
│          │          │   +      │         │  87%     │
│30 levels │          │Non-Ideal │         │FLAGSHIP  │
└──────────┘          └──────────┘         └──────────┘
  PHYSICS              COMPUTE            APPLICATION

"How does it fit      "Why does FeCIM
 in a real chip?"      beat everything?"
      ↓                      ↓
┌──────────┐          ┌──────────┐
│  DEMO 4  │    →     │  DEMO 5  │
│ Circuits │          │Comparison│
│  System  │          │ Investor │
│   CMOS   │          │  Pitch   │
└──────────┘          └──────────┘
  SYSTEM               BUSINESS
```

---

## 2. The 5-Demo Architecture

### Demo 1: "The Memory Cell" (Hysteresis)
**Status:** ✅ Complete - No changes needed  
**Audience:** Everyone (foundational)  
**Duration:** 2-3 minutes

```
┌─────────────────────────────────────────┐
│ Demo 1: Ferroelectric Hysteresis       │
├─────────────────────────────────────────┤
│ • P-E curve with live animation         │
│ • 30 discrete analog levels             │
│ • Material selector (HZO, optimized)    │
│ • Waveform modes (sine, triangle, etc.) │
│ • Preisach model simulation             │
│                                         │
│ Key Insight:                            │
│ "It's got 30 discrete states.           │
│  So it's not 0-1-0-1."                  │
│              — Dr. external research group           │
└─────────────────────────────────────────┘
```

**Implementation:** Keep as-is, update docs only.

---

### Demo 2: "The Crossbar Computer" (Enhanced with Non-Idealities)
**Status:** ✅ Complete + Merge Demo 7  
**Audience:** Engineers, technical investors  
**Duration:** 5-7 minutes

```
┌─────────────────────────────────────────┐
│ Demo 2: Crossbar MVM + Non-Idealities   │
├─────────────────────────────────────────┤
│ 4 TABS:                                 │
│                                         │
│ [1] Ideal MVM                           │
│   • Interactive crossbar heatmap        │
│   • Click cells to program weights      │
│   • Matrix-vector multiply animation    │
│   • Perfect operation                   │
│                                         │
│ [2] IR Drop Analysis                    │
│   • Wire resistance model               │
│   • Voltage gradient heatmap            │
│   • Worst-case corner identification    │
│   • Impact on accuracy                  │
│                                         │
│ [3] Sneak Path Currents                 │
│   • Parasitic current visualization     │
│   • Target cell vs interference         │
│   • SNR degradation                     │
│   • 1T1R vs 1R mitigation               │
│                                         │
│ [4] Drift & Variation                   │
│   • Conductance drift over time         │
│   • Device-to-device variation          │
│   • Cycle-to-cycle variation            │
│   • FeCIM vs ReRAM vs PCM comparison    │
│                                         │
│ Key Insight:                            │
│ "We handle real-world challenges        │
│  better than competition"               │
└─────────────────────────────────────────┘
```

**Implementation:** Refactor Demo 2 GUI, merge Demo 7 code.

---

### Demo 3: "The AI Brain" (MNIST - THE FLAGSHIP)
**Status:** ✅ Working → 🚀 World-Class Enhancement  
**Audience:** Investors, press, broad audience  
**Duration:** 10-15 minutes (with guided tour)

```
┌─────────────────────────────────────────────────────────────┐
│ Demo 3: MNIST Neural Network on FeCIM - FLAGSHIP           │
├─────────────────────────────────────────────────────────────┤
│ 4-ZONE INTERACTIVE LAYOUT:                                  │
│                                                             │
│ ┌──────────────┬──────────────────────────────────────┐    │
│ │ ZONE 1:      │ ZONE 2: LIVE INFERENCE              │    │
│ │ DRAWING      │                                      │    │
│ │              │ [Draw digit → See prediction]       │    │
│ │ [28×28       │                                      │    │
│ │  canvas]     │ FP:  "3" (94%)  ✓                   │    │
│ │              │ CIM: "3" (89%)  ✓ Match             │    │
│ │              │                                      │    │
│ │              │ Energy: 5.1 μJ (vs GPU: 51 mJ)      │    │
│ │              │ → 10,000× savings                   │    │
│ └──────────────┴──────────────────────────────────────┘    │
│ ┌──────────────┬──────────────────────────────────────┐    │
│ │ ZONE 3:      │ ZONE 4: WEIGHT VISUALIZATION        │    │
│ │ HARDWARE     │                                      │    │
│ │ KNOBS        │ Layer: [Input→Hidden ▼]             │    │
│ │              │                                      │    │
│ │ Levels: 30   │ [Crossbar heatmap: 128×784]         │    │
│ │ [●─────]     │ [30 discrete colors visible]        │    │
│ │              │                                      │    │
│ │ Noise: 0.01  │ Blue = negative weight               │    │
│ │ [──●───]     │ White = zero                         │    │
│ │              │ Red = positive weight                │    │
│ │ ADC: 6 bits  │                                      │    │
│ │ DAC: 8 bits  │ Distinct levels: 30 (FeCIM max)     │    │
│ │              │                                      │    │
│ │ [Run Quick   │                                      │    │
│ │  Test]       │                                      │    │
│ │ FP:  98.2%   │                                      │    │
│ │ CIM: 87.1% ✓ │                                      │    │
│ └──────────────┴──────────────────────────────────────┘    │
│                                                             │
│ FAILURE MODE PRESETS:                                      │
│ [Ideal] [Quant Cliff] [Noisy] [Broken ADC]                │
│                                                             │
│ GUIDED TOUR MODE: (7 steps, 3 minutes)                     │
│ [Start Tour] → Teaches "Why 30 levels?" story              │
│                                                             │
│ Key Insight:                                                │
│ "We're at 87% validation here...                           │
│  theoretical is 88%."                                       │
│              — Dr. external research group                               │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:** Full enhancement per MNIST plan (Phases 1-4).

---

### Demo 4: "The Chip System" (Circuits)
**Status:** ✅ Complete - Minor enhancements  
**Audience:** Foundries, chip designers  
**Duration:** 5-7 minutes

```
┌─────────────────────────────────────────┐
│ Demo 4: Peripheral Circuits & System    │
├─────────────────────────────────────────┤
│ • DAC/ADC conversion visualization      │
│ • Timing diagrams (write/read cycles)   │
│ • Charge pump operation                 │
│ • TIA (Transimpedance Amplifier)        │
│ • Power breakdown by component          │
│ • CMOS compatibility checklist          │
│                                         │
│ ENHANCEMENTS:                           │
│ • Add process flow diagram              │
│ • Add HZO superlattice layer stack      │
│ • Add "No exotic materials" section     │
│                                         │
│ Key Insight:                            │
│ "Works on a standard CMOS line          │
│  and can translate just like that."     │
│              — Dr. external research group           │
└─────────────────────────────────────────┘
```

**Implementation:** Add CMOS integration section, keep core as-is.

---

### Demo 5: "Why FeCIM Wins" (NEW - Technical Briefing)
**Status:** 🔲 Build from scratch  
**Audience:** Investors, executives, strategic partners  
**Duration:** 8-10 minutes

```
┌─────────────────────────────────────────────────────────────┐
│ Demo 5: Technology Comparison & Business Case              │
├─────────────────────────────────────────────────────────────┤
│ 5 INTERACTIVE SECTIONS:                                     │
│                                                             │
│ [1] ENERGY COMPARISON                                       │
│     ════════════════════════════════════════                │
│     CPU+DRAM  ████████████████████ 1000 pJ                  │
│     GPU+HBM   ████████           100 pJ                     │
│     FeCIM     █                   10 pJ                     │
│                                                             │
│     → 10,000,000× better than NAND                          │
│     → 1,000× better than DRAM                               │
│                                                             │
│ [2] COMPETITIVE MATRIX                                      │
│     ┌──────────┬──────┬──────┬──────┬──────┐               │
│     │ Feature  │FeCIM │ NAND │ReRAM │ PCM  │               │
│     ├──────────┼──────┼──────┼──────┼──────┤               │
│     │ Energy   │  ✅  │  ❌  │  🟡  │  🟡  │               │
│     │ Speed    │  ✅  │  ❌  │  ✅  │  ❌  │               │
│     │ Endure   │  ✅  │  ❌  │  ❌  │  🟡  │               │
│     │ CMOS     │  ✅  │  ✅  │  🟡  │  🟡  │               │
│     │ 30 lvls  │  ✅  │  ✅  │  ❌  │  ✅  │               │
│     │ CIM      │  ✅  │  ❌  │  🟡  │  🟡  │               │
│     └──────────┴──────┴──────┴──────┴──────┘               │
│                                                             │
│     Only FeCIM has ✅ across ALL categories                 │
│                                                             │
│ [3] DATA CENTER SAVINGS CALCULATOR                          │
│     Input: [1000] GPUs running inference                    │
│                                                             │
│     Current (GPU):                                          │
│     • Power: 250 kW                                         │
│     • Cost: $50,000/day                                     │
│     • CO₂: 2,190 tons/year                                  │
│                                                             │
│     With FeCIM:                                             │
│     • Power: 2.5 kW (-99%)                                  │
│     • Cost: $500/day (-99%)                                 │
│     • CO₂: 22 tons/year (-99%)                              │
│                                                             │
│     Annual Savings: $18.2M                                  │
│                                                             │
│ [4] MARKET OPPORTUNITY                                      │
│     AI Semiconductor Market:                                │
│     • 2025: $163B                                           │
│     • 2030: $403B (CAGR 20%)                                │
│     • FeCIM TAM: $50-100B                                   │
│                                                             │
│     Addressable Markets:                                    │
│     • NAND Flash replacement (Phase 1)                      │
│     • DRAM replacement (Phase 2)                            │
│     • Full CIM compute (Phase 3)                            │
│                                                             │
│ [5] TRL PROGRESSION                                         │
│     TRL 1  2  3  4  5  6  7  8  9                           │
│      ○  ○  ○  ●  ○  ○  ○  ○  ○                             │
│               ↑                                             │
│          WE ARE HERE                                        │
│     "Component validation in lab"                           │
│                                                             │
│     Next Milestones:                                        │
│     • TRL 5: Prototype (6 months)                           │
│     • TRL 6: Pilot line (12 months)                         │
│     • TRL 7: Pre-production (18 months)                     │
│     • TRL 8: Qualified (24 months)                          │
│                                                             │
│ Key Insight:                                                │
│ "This could lower the requirements                          │
│  in a data center by 80 to 90%."                            │
│              — Dr. external research group                               │
└─────────────────────────────────────────────────────────────┘
```

**Implementation:** New demo built on Fyne, reuse comparison framework.

---

## 3. Phase 1: Consolidation (Week 1-2)

### Goal
Merge Demo 7 into Demo 2, archive Demo 6, prepare for Demo 5 creation.

### Task 1.1: Merge Demo 7 → Demo 2 (5 days)

#### Current State
```
module2-crossbar/
├── pkg/
│   ├── crossbar/
│   │   ├── array.go           # Basic MVM
│   │   └── config.go
│   └── gui/
│       └── app.go              # Single view

demo7-nonidealities/
├── pkg/
│   ├── irdrop/
│   │   └── analysis.go         # Voltage drop model
│   ├── sneak/
│   │   └── paths.go            # Parasitic currents
│   └── drift/
│       └── model.go            # Conductance drift
```

#### Target State
```
module2-crossbar/
├── pkg/
│   ├── crossbar/
│   │   ├── array.go
│   │   ├── config.go
│   │   ├── irdrop.go          # ← Moved from demo7
│   │   ├── sneak.go           # ← Moved from demo7
│   │   └── drift.go           # ← Moved from demo7
│   └── gui/
│       ├── app.go
│       ├── tabs/
│       │   ├── ideal_tab.go    # Tab 1: Normal MVM
│       │   ├── irdrop_tab.go   # Tab 2: IR drop
│       │   ├── sneak_tab.go    # Tab 3: Sneak paths
│       │   └── drift_tab.go    # Tab 4: Variation
│       └── shared_widgets.go
```

#### Implementation Steps

**Day 1-2: Move code**
```bash
# 1. Copy non-ideality models
cp demo7-nonidealities/pkg/irdrop/*.go module2-crossbar/pkg/crossbar/
cp demo7-nonidealities/pkg/sneak/*.go module2-crossbar/pkg/crossbar/
cp demo7-nonidealities/pkg/drift/*.go module2-crossbar/pkg/crossbar/

# 2. Update package declarations
# Change: package irdrop → package crossbar
# Change: package sneak → package crossbar
# Change: package drift → package crossbar

# 3. Update imports in moved files
# Old: "multilayer-ferroelectric-cim-visualizer/demo7-nonidealities/pkg/irdrop"
# New: "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
```

**Day 3: Create tabbed GUI**

```go
// module2-crossbar/pkg/gui/tabs/ideal_tab.go

package tabs

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
)

type IdealTab struct {
    crossbarArray *crossbar.Array
    heatmapView   *fyne.Container
    controlPanel  *fyne.Container
}

func NewIdealTab(array *crossbar.Array) *IdealTab {
    tab := &IdealTab{
        crossbarArray: array,
    }
    tab.createUI()
    return tab
}

func (t *IdealTab) createUI() {
    // Existing Demo 2 heatmap code
    t.heatmapView = createCrossbarHeatmap(t.crossbarArray)
    t.controlPanel = createControlPanel()
}

func (t *IdealTab) Content() fyne.CanvasObject {
    return container.NewBorder(
        nil, // top
        t.controlPanel, // bottom
        nil, // left
        nil, // right
        t.heatmapView, // center
    )
}
```

```go
// module2-crossbar/pkg/gui/tabs/irdrop_tab.go

package tabs

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
)

type IRDropTab struct {
    crossbarArray *crossbar.Array
    voltageMap    *fyne.Container
    analysisPanel *fyne.Container
    statsLabel    *widget.Label
}

func NewIRDropTab(array *crossbar.Array) *IRDropTab {
    tab := &IRDropTab{
        crossbarArray: array,
        statsLabel:    widget.NewLabel(""),
    }
    tab.createUI()
    return tab
}

func (t *IRDropTab) createUI() {
    // Use code from demo7-nonidealities
    t.voltageMap = createVoltageGradientHeatmap(t.crossbarArray)
    
    t.analysisPanel = container.NewVBox(
        widget.NewLabel("IR Drop Analysis"),
        t.statsLabel,
        widget.NewButton("Identify Worst Corner", func() {
            t.findWorstCorner()
        }),
    )
}

func (t *IRDropTab) findWorstCorner() {
    // Call crossbar.AnalyzeIRDrop()
    result := crossbar.AnalyzeIRDrop(t.crossbarArray)
    t.statsLabel.SetText(result.Summary())
}

func (t *IRDropTab) Content() fyne.CanvasObject {
    return container.NewBorder(
        nil,
        t.analysisPanel,
        nil,
        nil,
        t.voltageMap,
    )
}
```

**Day 4: Wire up tabs in main app**

```go
// module2-crossbar/pkg/gui/app.go

package gui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
    "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/gui/tabs"
)

type CrossbarApp struct {
    fyneApp fyne.App
    window  fyne.Window
    array   *crossbar.Array
    
    // Tabs
    idealTab  *tabs.IdealTab
    irdropTab *tabs.IRDropTab
    sneakTab  *tabs.SneakTab
    driftTab  *tabs.DriftTab
}

func NewCrossbarApp() *CrossbarApp {
    app := &CrossbarApp{
        fyneApp: app.NewWithID("com.fecim.crossbar-demo"),
    }
    
    // Initialize crossbar
    app.array = crossbar.NewArray(16, 16, crossbar.DefaultConfig())
    
    // Create tabs
    app.idealTab = tabs.NewIdealTab(app.array)
    app.irdropTab = tabs.NewIRDropTab(app.array)
    app.sneakTab = tabs.NewSneakTab(app.array)
    app.driftTab = tabs.NewDriftTab(app.array)
    
    app.createUI()
    return app
}

func (a *CrossbarApp) createUI() {
    a.window = a.fyneApp.NewWindow("Demo 2: Crossbar MVM + Non-Idealities")
    
    tabs := container.NewAppTabs(
        container.NewTabItem("Ideal MVM", a.idealTab.Content()),
        container.NewTabItem("IR Drop Analysis", a.irdropTab.Content()),
        container.NewTabItem("Sneak Paths", a.sneakTab.Content()),
        container.NewTabItem("Drift & Variation", a.driftTab.Content()),
    )
    
    a.window.SetContent(tabs)
    a.window.Resize(fyne.NewSize(1200, 800))
}

func (a *CrossbarApp) Run() {
    a.window.ShowAndRun()
}
```

**Day 5: Testing and cleanup**

```bash
# Test each tab independently
cd module2-crossbar
go test ./pkg/crossbar -v
go test ./pkg/gui/tabs -v

# Run integrated demo
go build -o crossbar-gui ./cmd/crossbar-gui
./crossbar-gui

# Verify all 4 tabs work:
# 1. Ideal MVM → heatmap, click cells, MVM animation
# 2. IR Drop → voltage gradient, worst corner
# 3. Sneak Paths → parasitic currents, SNR
# 4. Drift → time-series plot, variation stats
```

#### Acceptance Criteria
- [ ] All 4 tabs functional
- [ ] No crashes when switching tabs
- [ ] Demo 7 code fully integrated
- [ ] Tests pass: `go test ./module2-crossbar/...`

---

### Task 1.2: Archive Demo 6 & Demo 7 (1 day)

```bash
# 1. Create archive directory
mkdir -p docs/archive/removed-demos

# 2. Document why demos were removed
cat > docs/archive/removed-demos/README.md << 'EOF'
# Archived Demos

## Demo 6: Multi-Layer 3D Stack

**Reason for removal:** Too niche, adds complexity without clarity.

**What it showed:** 3D stacking of crossbar layers with via connections.

**Why it's not needed:**
- FeCIM is at TRL 4 (lab validation)
- 3D stacking is TRL 2-3 (years away from practical)
- Better to show static 3D diagram in Demo 4 or Demo 5
- Focus resources on demos that matter NOW

**Preserved artifacts:**
- Conceptual diagrams → docs/archive/removed-demos/demo6-diagrams/
- Code snippets → docs/archive/removed-demos/demo6-code/

## Demo 7: Non-Idealities

**Reason for removal:** Merged into Demo 2 as tabs.

**What it showed:** IR drop, sneak paths, conductance drift.

**New location:** module2-crossbar/pkg/gui/tabs/
- Tab 2: IR Drop Analysis
- Tab 3: Sneak Paths
- Tab 4: Drift & Variation

**All functionality preserved and enhanced.**
EOF

# 3. Move old code to archive
mv demo6-multilayer docs/archive/removed-demos/demo6-multilayer
mv demo7-nonidealities docs/archive/removed-demos/demo7-nonidealities

# 4. Update .gitignore
echo "docs/archive/removed-demos/demo6-multilayer/" >> .gitignore
echo "docs/archive/removed-demos/demo7-nonidealities/" >> .gitignore
```

#### Acceptance Criteria
- [ ] Archive directory created
- [ ] Removal rationale documented
- [ ] Old demos moved (not deleted - for reference)
- [ ] .gitignore updated

---

### Task 1.3: Update Project Structure (2 days)

#### Update README.md

```markdown
# Ferroelectric CIM Visualizer

**5 World-Class Demos for Ferroelectric Compute-in-Memory**

[![Demos](https://img.shields.io/badge/Demos-5%2F5-brightgreen.svg)]()

---

## The FeCIM Story: 5 Core Demos

```
Demo 1: "How the memory cell works"        ✅ Fyne GUI
Demo 2: "How we compute in memory"         ✅ Fyne GUI (4 tabs)
Demo 3: "What we can build with it"        ✅ Fyne GUI (FLAGSHIP)
Demo 4: "How it fits in a real chip"       ✅ Fyne GUI
Demo 5: "Why FeCIM beats everything"       ✅ Fyne GUI (NEW)
```

---

## Quick Start

```bash
# Install dependencies (Ubuntu/Debian)
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# Run all 5 demos from unified launcher
go build ./cmd/fecim-visualizer && ./fecim-visualizer

# Or run individual demos:
./module1-hysteresis/hysteresis
./module2-crossbar/crossbar-gui
./module3-mnist/mnist-gui
./module4-circuits/circuits-gui
./demo5-comparison/comparison-gui
```

---

## Demo Details

### Demo 1: Ferroelectric Hysteresis ✅

[... existing Demo 1 content ...]

### Demo 2: Crossbar MVM + Non-Idealities ✅

**NEW: 4-tab interface shows ideal and real-world behavior**

**Tab 1: Ideal MVM**
- Interactive crossbar heatmap
- Click cells to program weights
- Matrix-vector multiply animation

**Tab 2: IR Drop Analysis**
- Wire resistance modeling
- Voltage gradient visualization
- Worst-case corner identification

**Tab 3: Sneak Path Currents**
- Parasitic current visualization
- SNR degradation analysis
- 1T1R vs 1R comparison

**Tab 4: Drift & Variation**
- Conductance drift over time
- Device-to-device variation
- FeCIM vs ReRAM vs PCM

[... continue ...]

### Demo 3: MNIST Neural Network (FLAGSHIP) 🚀

**NEW: Enhanced with 4-zone interactive layout**

[... see Section 4 for full details ...]

### Demo 4: Peripheral Circuits ✅

[... existing Demo 4 content + minor enhancements ...]

### Demo 5: Technology Comparison (NEW) ⭐

**Purpose:** Investor pitch - why FeCIM wins

**Features:**
- Side-by-side energy comparison
- Competitive technology matrix
- Data center savings calculator
- Market opportunity ($403B by 2030)
- TRL progression roadmap

[... see Section 5 for full details ...]

---

## Removed Demos (Archived)

**Demo 6 (Multi-Layer 3D):** Archived - too far from current TRL 4
**Demo 7 (Non-Idealities):** Merged into Demo 2 as tabs

See `docs/archive/removed-demos/README.md` for details.
```

#### Update File Structure

```bash
# New structure
multilayer-ferroelectric-cim-visualizer/
├── module1-hysteresis/          ✅ Keep as-is
├── module2-crossbar/            ✅ Enhanced with tabs
├── module3-mnist/               🚀 Full enhancement (Phase 2)
├── module4-circuits/            ✅ Minor enhancements
├── demo5-comparison/          🔲 Build in Phase 3
├── shared/                    Shared utilities
├── docs/
│   ├── README.md
│   ├── module1-hysteresis.md
│   ├── module2-crossbar.md      ← Updated with tab docs
│   ├── module3-mnist.md         ← Full enhancement plan
│   ├── module4-circuits.md
│   ├── demo5-comparison.md    ← New
│   └── archive/
│       └── removed-demos/
│           ├── README.md
│           ├── demo6-multilayer/
│           └── demo7-nonidealities/
├── scripts/
│   ├── build-all.sh           ← Updated for 5 demos
│   ├── test-all.sh            ← Updated for 5 demos
│   └── demo-launcher.sh       ← New: launches all 5
└── README.md                  ← Updated with 5-demo story
```

#### Create Unified Launcher

```bash
# scripts/demo-launcher.sh

#!/bin/bash

cat << "EOF"
╔════════════════════════════════════════════════════════════╗
║     FERROELECTRIC CIM VISUALIZER - DEMO LAUNCHER          ║
║                                                            ║
║  5 World-Class Demos                                       ║
╚════════════════════════════════════════════════════════════╝
EOF

echo ""
echo "Select demo to run:"
echo ""
echo "  1) Demo 1: Ferroelectric Hysteresis (2-3 min)"
echo "  2) Demo 2: Crossbar MVM + Non-Idealities (5-7 min)"
echo "  3) Demo 3: MNIST Neural Network - FLAGSHIP (10-15 min)"
echo "  4) Demo 4: Peripheral Circuits (5-7 min)"
echo "  5) Demo 5: Technology Comparison - Technical Briefing (8-10 min)"
echo ""
echo "  A) Run ALL demos sequentially"
echo "  Q) Quit"
echo ""
read -p "Choice: " choice

case $choice in
    1)
        echo "Launching Demo 1: Hysteresis..."
        cd module1-hysteresis && ./hysteresis
        ;;
    2)
        echo "Launching Demo 2: Crossbar + Non-Idealities..."
        cd module2-crossbar && ./crossbar-gui
        ;;
    3)
        echo "Launching Demo 3: MNIST (FLAGSHIP)..."
        cd module3-mnist && ./mnist-gui
        ;;
    4)
        echo "Launching Demo 4: Circuits..."
        cd module4-circuits && ./circuits-gui
        ;;
    5)
        echo "Launching Demo 5: Comparison..."
        cd demo5-comparison && ./comparison-gui
        ;;
    A|a)
        echo "Running all demos sequentially..."
        echo "Demo 1 will start in 3 seconds..."
        sleep 3
        ./module1-hysteresis/hysteresis &
        wait
        ./module2-crossbar/crossbar-gui &
        wait
        ./module3-mnist/mnist-gui &
        wait
        ./module4-circuits/circuits-gui &
        wait
        ./demo5-comparison/comparison-gui &
        wait
        ;;
    Q|q)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac
```

```bash
chmod +x scripts/demo-launcher.sh
```

#### Acceptance Criteria (Phase 1)
- [ ] README.md reflects 5 demos (not 8)
- [ ] Archive directory properly documented
- [ ] Unified launcher script works
- [ ] All build/test scripts updated
- [ ] Demo 2 has 4 functional tabs

---

## 4. Phase 2: Demo 3 Enhancement (Week 3-6)

**This is the FLAGSHIP demo - gets the most attention.**

Refer to the complete MNIST enhancement plan from earlier (Sections 1-8 of that document).

### Summary Timeline

**Week 3: Core Infrastructure**
- Implement `core/quantize.go` with symmetric mapping
- Implement `core/network.go` (DualModeNetwork)
- Implement `core/inference.go` (dual-path forward)
- Write unit tests (target: 85% coverage)

**Week 4: GUI Basics**
- Implement 4-zone layout
- Drawing canvas (reuse existing)
- Result panel (FP vs CIM comparison)
- Control panel (sliders)
- Weight panel (heatmap)

**Week 5: Educational Features**
- Add 4 failure mode presets
- Add "Quick Test" button
- Add info dialogs (Why 30?, Hardware Reality)
- Add energy efficiency display
- Implement Guided Tour (7 steps)

**Week 6: Polish**
- Documentation updates
- Full test suite
- Benchmark against literature
- Bug fixes and optimization

### Key Implementation Details

See **Section 5 (Code Specifications)** from the MNIST enhancement plan for:
- Complete code for `DualModeNetwork`
- `QuantizeWeights()` function
- Dual inference engine
- All 4 GUI zones (drawing, results, controls, weights)

### Acceptance Criteria (Phase 2)
- [ ] 4-zone layout functional
- [ ] FP vs CIM dual inference works
- [ ] Quantization slider (1-30) affects results
- [ ] Failure mode presets work correctly
- [ ] Guided Tour runs without crashes
- [ ] Tests pass: `go test ./module3-mnist/... -cover`
- [ ] Coverage > 80% for core package
- [ ] Quick Test shows ~87% CIM accuracy with noise=0.08

---

## 5. Phase 3: Demo 5 Creation (Week 7-8)

### Goal
Build technical briefing tool with comparison charts, calculators, market data.

### Architecture

```
demo5-comparison/
├── cmd/
│   └── comparison-gui/
│       └── main.go
├── pkg/
│   ├── comparison/
│   │   ├── energy.go          # Energy comparison data
│   │   ├── competitive.go     # Competitive matrix
│   │   ├── market.go          # Market size data
│   │   └── trl.go             # TRL progression model
│   ├── calculator/
│   │   └── savings.go         # Data center savings calculator
│   └── gui/
│       ├── app.go
│       ├── sections/
│       │   ├── energy_section.go
│       │   ├── competitive_section.go
│       │   ├── calculator_section.go
│       │   ├── market_section.go
│       │   └── trl_section.go
│       └── charts/
│           ├── bar_chart.go
│           ├── matrix_chart.go
│           └── timeline_chart.go
└── data/
    ├── energy_comparison.json
    ├── competitive_matrix.json
    └── market_data.json
```

### Week 7: Core Data & Calculations

#### Day 1-2: Energy Comparison

```go
// demo5-comparison/pkg/comparison/energy.go

package comparison

// EnergyMetrics represents energy consumption for a technology
type EnergyMetrics struct {
    Technology       string
    EnergyPerMAC_pJ  float64  // picojoules
    EnergyPerInf_uJ  float64  // microjoules (MNIST inference)
    PowerPerChip_mW  float64
    Verified         bool     // Is this measured or claimed?
    Source           string
}

// GetEnergyComparison returns comparison data
func GetEnergyComparison() []EnergyMetrics {
    return []EnergyMetrics{
        {
            Technology:      "CPU + DRAM",
            EnergyPerMAC_pJ: 1000.0,
            EnergyPerInf_uJ: 101.6,  // 101,632 MACs × 1000 pJ
            PowerPerChip_mW: 15000.0,
            Verified:        true,
            Source:          "Intel Xeon datasheet",
        },
        {
            Technology:      "GPU + HBM",
            EnergyPerMAC_pJ: 100.0,
            EnergyPerInf_uJ: 10.16,
            PowerPerChip_mW: 300000.0, // 300W
            Verified:        true,
            Source:          "NVIDIA V100 datasheet",
        },
        {
            Technology:      "FPGA",
            EnergyPerMAC_pJ: 50.0,
            EnergyPerInf_uJ: 5.08,
            PowerPerChip_mW: 75000.0,
            Verified:        true,
            Source:          "Xilinx Versal",
        },
        {
            Technology:      "ReRAM CIM",
            EnergyPerMAC_pJ: 10.0,
            EnergyPerInf_uJ: 1.016,
            PowerPerChip_mW: 100.0,
            Verified:        false,
            Source:          "Weebit Nano claims",
        },
        {
            Technology:      "FeCIM (HZO)",
            EnergyPerMAC_pJ: 0.05,   // 50 fJ
            EnergyPerInf_uJ: 0.00508,
            PowerPerChip_mW: 10.0,
            Verified:        false,
            Source:          "Jerry et al. IEDM 2017 + Dr. Tour",
        },
    }
}

// CalculateEnergyRatio returns how much better technology A is vs B
func CalculateEnergyRatio(techA, techB string) float64 {
    data := GetEnergyComparison()
    
    var energyA, energyB float64
    for _, m := range data {
        if m.Technology == techA {
            energyA = m.EnergyPerMAC_pJ
        }
        if m.Technology == techB {
            energyB = m.EnergyPerMAC_pJ
        }
    }
    
    if energyA == 0 {
        return 0
    }
    
    return energyB / energyA
}
```

#### Day 3: Competitive Matrix

```go
// demo5-comparison/pkg/comparison/competitive.go

package comparison

type CompetitiveFeature string

const (
    FeatureEnergy     CompetitiveFeature = "Energy Efficiency"
    FeatureSpeed      CompetitiveFeature = "Write/Read Speed"
    FeatureEndurance  CompetitiveFeature = "Endurance (cycles)"
    FeatureCMOS       CompetitiveFeature = "CMOS Compatible"
    FeatureCIM        CompetitiveFeature = "Compute-in-Memory"
    FeatureLevels     CompetitiveFeature = "Analog Levels (30+)"
    FeatureRetention  CompetitiveFeature = "Retention (10yr)"
)

type Rating int

const (
    Poor     Rating = 0  // ❌
    Fair     Rating = 1  // 🟡
    Good     Rating = 2  // ✅
)

type Technology struct {
    Name     string
    Features map[CompetitiveFeature]Rating
    Notes    string
}

func GetCompetitiveMatrix() []Technology {
    return []Technology{
        {
            Name: "FeCIM (HZO)",
            Features: map[CompetitiveFeature]Rating{
                FeatureEnergy:    Good,
                FeatureSpeed:     Good,
                FeatureEndurance: Good,
                FeatureCMOS:      Good,
                FeatureCIM:       Good,
                FeatureLevels:    Good,
                FeatureRetention: Good,
            },
            Notes: "Only technology with ✅ across all categories",
        },
        {
            Name: "3D NAND Flash",
            Features: map[CompetitiveFeature]Rating{
                FeatureEnergy:    Poor,
                FeatureSpeed:     Poor,
                FeatureEndurance: Poor,
                FeatureCMOS:      Good,
                FeatureCIM:       Poor,
                FeatureLevels:    Good, // TLC/QLC
                FeatureRetention: Good,
            },
            Notes: "Legacy technology, high energy",
        },
        {
            Name: "ReRAM (Weebit)",
            Features: map[CompetitiveFeature]Rating{
                FeatureEnergy:    Good,
                FeatureSpeed:     Good,
                FeatureEndurance: Poor,
                FeatureCMOS:      Fair,
                FeatureCIM:       Fair,
                FeatureLevels:    Poor, // Binary/4-level
                FeatureRetention: Fair,
            },
            Notes: "Good speed, but endurance and variability issues",
        },
        {
            Name: "PCM (Phase Change)",
            Features: map[CompetitiveFeature]Rating{
                FeatureEnergy:    Fair,
                FeatureSpeed:     Poor,
                FeatureEndurance: Fair,
                FeatureCMOS:      Fair,
                FeatureCIM:       Fair,
                FeatureLevels:    Good, // Multi-level
                FeatureRetention: Good,
            },
            Notes: "Slow writes, drift issues",
        },
        {
            Name: "MRAM (STT/SOT)",
            Features: map[CompetitiveFeature]Rating{
                FeatureEnergy:    Fair,
                FeatureSpeed:     Good,
                FeatureEndurance: Good,
                FeatureCMOS:      Fair,
                FeatureCIM:       Poor,
                FeatureLevels:    Poor, // Binary
                FeatureRetention: Good,
            },
            Notes: "Fast but binary only, CIM challenging",
        },
    }
}

func (t *Technology) Score() int {
    score := 0
    for _, rating := range t.Features {
        score += int(rating)
    }
    return score
}
```

#### Day 4-5: Data Center Calculator

```go
// demo5-comparison/pkg/calculator/savings.go

package calculator

import "fmt"

type DataCenterConfig struct {
    NumGPUs           int
    GPUPower_W        float64
    GPUUtilization    float64  // 0-1
    HoursPerDay       float64
    ElectricityCost   float64  // $/kWh
    CO2PerKWh         float64  // kg CO2 per kWh
}

type SavingsResult struct {
    // Current (GPU)
    CurrentPower_kW      float64
    CurrentEnergy_kWh    float64  // per day
    CurrentCost_day      float64
    CurrentCO2_tons_year float64
    
    // With FeCIM
    FeCIMPower_kW        float64
    FeCIMEnergy_kWh      float64
    FeCIMCost_day        float64
    FeCIMCO2_tons_year   float64
    
    // Savings
    PowerReduction_pct   float64
    CostSavings_day      float64
    CostSavings_year     float64
    CO2Savings_tons_year float64
}

func CalculateSavings(config DataCenterConfig) SavingsResult {
    // GPU baseline
    totalGPUPower_W := float64(config.NumGPUs) * config.GPUPower_W * config.GPUUtilization
    currentPower_kW := totalGPUPower_W / 1000.0
    currentEnergy_kWh := currentPower_kW * config.HoursPerDay
    currentCost_day := currentEnergy_kWh * config.ElectricityCost
    currentCO2_kg_day := currentEnergy_kWh * config.CO2PerKWh
    currentCO2_tons_year := currentCO2_kg_day * 365 / 1000.0
    
    // FeCIM (assume 10,000× lower energy)
    fecimEnergyRatio := 10000.0
    fecimEnergy_kWh := currentEnergy_kWh / fecimEnergyRatio
    fecimPower_kW := fecimEnergy_kWh / config.HoursPerDay
    fecimCost_day := fecimEnergy_kWh * config.ElectricityCost
    fecimCO2_kg_day := fecimEnergy_kWh * config.CO2PerKWh
    fecimCO2_tons_year := fecimCO2_kg_day * 365 / 1000.0
    
    // Savings
    powerReduction_pct := (1.0 - fecimPower_kW/currentPower_kW) * 100.0
    costSavings_day := currentCost_day - fecimCost_day
    costSavings_year := costSavings_day * 365
    co2Savings_tons_year := currentCO2_tons_year - fecimCO2_tons_year
    
    return SavingsResult{
        CurrentPower_kW:      currentPower_kW,
        CurrentEnergy_kWh:    currentEnergy_kWh,
        CurrentCost_day:      currentCost_day,
        CurrentCO2_tons_year: currentCO2_tons_year,
        
        FeCIMPower_kW:        fecimPower_kW,
        FeCIMEnergy_kWh:      fecimEnergy_kWh,
        FeCIMCost_day:        fecimCost_day,
        FeCIMCO2_tons_year:   fecimCO2_tons_year,
        
        PowerReduction_pct:   powerReduction_pct,
        CostSavings_day:      costSavings_day,
        CostSavings_year:     costSavings_year,
        CO2Savings_tons_year: co2Savings_tons_year,
    }
}

func (r *SavingsResult) Summary() string {
    return fmt.Sprintf(`
Data Center Savings Analysis
════════════════════════════════════════════

CURRENT (GPU):
  Power:        %.1f kW
  Energy/day:   %.1f kWh
  Cost/day:     $%.2f
  CO₂/year:     %.1f tons

WITH FeCIM:
  Power:        %.3f kW (%.1f%% reduction)
  Energy/day:   %.3f kWh
  Cost/day:     $%.2f
  CO₂/year:     %.1f tons

ANNUAL SAVINGS:
  Cost:         $%.2f million
  CO₂:          %.1f tons
`,
        r.CurrentPower_kW,
        r.CurrentEnergy_kWh,
        r.CurrentCost_day,
        r.CurrentCO2_tons_year,
        
        r.FeCIMPower_kW,
        r.PowerReduction_pct,
        r.FeCIMEnergy_kWh,
        r.FeCIMCost_day,
        r.FeCIMCO2_tons_year,
        
        r.CostSavings_year / 1000000.0,
        r.CO2Savings_tons_year,
    )
}
```

### Week 8: GUI Implementation

#### Energy Comparison Section

```go
// demo5-comparison/pkg/gui/sections/energy_section.go

package sections

import (
    "fmt"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "fyne.io/fyne/v2/canvas"
    "image/color"
    "multilayer-ferroelectric-cim-visualizer/demo5-comparison/pkg/comparison"
)

type EnergySection struct {
    data         []comparison.EnergyMetrics
    chartCanvas  *fyne.Container
    detailsPanel *fyne.Container
}

func NewEnergySection() *EnergySection {
    es := &EnergySection{
        data: comparison.GetEnergyComparison(),
    }
    es.createUI()
    return es
}

func (es *EnergySection) createUI() {
    // Create bar chart
    es.chartCanvas = es.createBarChart()
    
    // Create details table
    detailsTable := widget.NewTable(
        func() (int, int) {
            return len(es.data) + 1, 4 // +1 for header
        },
        func() fyne.CanvasObject {
            return widget.NewLabel("")
        },
        func(id widget.TableCellID, cell fyne.CanvasObject) {
            label := cell.(*widget.Label)
            
            if id.Row == 0 {
                // Header
                headers := []string{"Technology", "Energy/MAC", "Verified", "Source"}
                label.SetText(headers[id.Col])
                label.TextStyle = fyne.TextStyle{Bold: true}
            } else {
                // Data
                metric := es.data[id.Row-1]
                switch id.Col {
                case 0:
                    label.SetText(metric.Technology)
                case 1:
                    label.SetText(fmt.Sprintf("%.2f pJ", metric.EnergyPerMAC_pJ))
                case 2:
                    if metric.Verified {
                        label.SetText("✅ Yes")
                    } else {
                        label.SetText("⚠️ Claimed")
                    }
                case 3:
                    label.SetText(metric.Source)
                }
            }
        },
    )
    
    detailsTable.SetColumnWidth(0, 150)
    detailsTable.SetColumnWidth(1, 100)
    detailsTable.SetColumnWidth(2, 80)
    detailsTable.SetColumnWidth(3, 200)
    
    es.detailsPanel = container.NewVBox(
        widget.NewLabel("Energy Comparison Details"),
        detailsTable,
    )
}

func (es *EnergySection) createBarChart() *fyne.Container {
    bars := container.NewVBox()
    
    // Find max value for scaling
    maxEnergy := 0.0
    for _, m := range es.data {
        if m.EnergyPerMAC_pJ > maxEnergy {
            maxEnergy = m.EnergyPerMAC_pJ
        }
    }
    
    for _, m := range es.data {
        // Create bar
        barWidth := (m.EnergyPerMAC_pJ / maxEnergy) * 800.0
        
        var barColor color.Color
        if m.Technology == "FeCIM (HZO)" {
            barColor = color.RGBA{0, 212, 255, 255} // Cyan (highlight)
        } else {
            barColor = color.RGBA{100, 100, 100, 255}
        }
        
        bar := canvas.NewRectangle(barColor)
        bar.SetMinSize(fyne.NewSize(float32(barWidth), 40))
        
        label := widget.NewLabel(fmt.Sprintf("%s: %.2f pJ", m.Technology, m.EnergyPerMAC_pJ))
        
        row := container.NewHBox(label, bar)
        bars.Add(row)
    }
    
    // Add comparison ratio
    ratio := comparison.CalculateEnergyRatio("FeCIM (HZO)", "CPU + DRAM")
    ratioLabel := widget.NewLabelWithStyle(
        fmt.Sprintf("FeCIM is %.0f× more efficient than CPU+DRAM", ratio),
        fyne.TextAlignCenter,
        fyne.TextStyle{Bold: true},
    )
    
    return container.NewVBox(
        widget.NewLabel("Energy per MAC (picojoules)"),
        bars,
        ratioLabel,
    )
}

func (es *EnergySection) Content() fyne.CanvasObject {
    return container.NewVBox(
        widget.NewLabelWithStyle("Section 1: Energy Efficiency", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        es.chartCanvas,
        widget.NewSeparator(),
        es.detailsPanel,
    )
}
```

#### Calculator Section

```go
// demo5-comparison/pkg/gui/sections/calculator_section.go

package sections

import (
    "fmt"
    "strconv"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "multilayer-ferroelectric-cim-visualizer/demo5-comparison/pkg/calculator"
)

type CalculatorSection struct {
    // Inputs
    numGPUsEntry    *widget.Entry
    utilizationSlider *widget.Slider
    hoursEntry      *widget.Entry
    costEntry       *widget.Entry
    
    // Output
    resultsText     *widget.Label
    
    // Current config
    config          calculator.DataCenterConfig
}

func NewCalculatorSection() *CalculatorSection {
    cs := &CalculatorSection{
        config: calculator.DataCenterConfig{
            NumGPUs:         1000,
            GPUPower_W:      300.0,
            GPUUtilization:  0.7,
            HoursPerDay:     24.0,
            ElectricityCost: 0.10,
            CO2PerKWh:       0.5,
        },
    }
    cs.createUI()
    return cs
}

func (cs *CalculatorSection) createUI() {
    cs.numGPUsEntry = widget.NewEntry()
    cs.numGPUsEntry.SetText("1000")
    cs.numGPUsEntry.OnChanged = func(s string) {
        if val, err := strconv.Atoi(s); err == nil {
            cs.config.NumGPUs = val
            cs.calculate()
        }
    }
    
    cs.utilizationSlider = widget.NewSlider(0, 1)
    cs.utilizationSlider.Value = 0.7
    cs.utilizationSlider.Step = 0.1
    cs.utilizationSlider.OnChanged = func(v float64) {
        cs.config.GPUUtilization = v
        cs.calculate()
    }
    
    cs.hoursEntry = widget.NewEntry()
    cs.hoursEntry.SetText("24")
    cs.hoursEntry.OnChanged = func(s string) {
        if val, err := strconv.ParseFloat(s, 64); err == nil {
            cs.config.HoursPerDay = val
            cs.calculate()
        }
    }
    
    cs.costEntry = widget.NewEntry()
    cs.costEntry.SetText("0.10")
    cs.costEntry.OnChanged = func(s string) {
        if val, err := strconv.ParseFloat(s, 64); err == nil {
            cs.config.ElectricityCost = val
            cs.calculate()
        }
    }
    
    cs.resultsText = widget.NewLabel("")
    cs.resultsText.Wrapping = fyne.TextWrapWord
    
    // Initial calculation
    cs.calculate()
}

func (cs *CalculatorSection) calculate() {
    result := calculator.CalculateSavings(cs.config)
    cs.resultsText.SetText(result.Summary())
}

func (cs *CalculatorSection) Content() fyne.CanvasObject {
    inputsForm := container.NewVBox(
        widget.NewLabel("Data Center Configuration:"),
        container.NewHBox(
            widget.NewLabel("Number of GPUs:"),
            cs.numGPUsEntry,
        ),
        container.NewHBox(
            widget.NewLabel("GPU Utilization:"),
            cs.utilizationSlider,
            widget.NewLabel(fmt.Sprintf("%.0f%%", cs.utilizationSlider.Value*100)),
        ),
        container.NewHBox(
            widget.NewLabel("Hours/day:"),
            cs.hoursEntry,
        ),
        container.NewHBox(
            widget.NewLabel("Electricity cost ($/kWh):"),
            cs.costEntry,
        ),
    )
    
    return container.NewVBox(
        widget.NewLabelWithStyle("Section 3: Data Center Savings Calculator", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
        inputsForm,
        widget.NewSeparator(),
        widget.NewLabel("Results:"),
        container.NewScroll(cs.resultsText),
    )
}
```

#### Main App with Sections

```go
// demo5-comparison/pkg/gui/app.go

package gui

import (
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/app"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
    "multilayer-ferroelectric-cim-visualizer/demo5-comparison/pkg/gui/sections"
)

type ComparisonApp struct {
    fyneApp fyne.App
    window  fyne.Window
    
    energySection      *sections.EnergySection
    competitiveSection *sections.CompetitiveSection
    calculatorSection  *sections.CalculatorSection
    marketSection      *sections.MarketSection
    trlSection         *sections.TRLSection
}

func NewComparisonApp() *ComparisonApp {
    app := &ComparisonApp{
        fyneApp: app.NewWithID("com.fecim.comparison-demo"),
    }
    
    app.energySection = sections.NewEnergySection()
    app.competitiveSection = sections.NewCompetitiveSection()
    app.calculatorSection = sections.NewCalculatorSection()
    app.marketSection = sections.NewMarketSection()
    app.trlSection = sections.NewTRLSection()
    
    app.createUI()
    return app
}

func (a *ComparisonApp) createUI() {
    a.window = a.fyneApp.NewWindow("Demo 5: Technology Comparison - Technical Briefing")
    
    // Create tabs for each section
    tabs := container.NewAppTabs(
        container.NewTabItem("1. Energy Efficiency", a.energySection.Content()),
        container.NewTabItem("2. Competitive Matrix", a.competitiveSection.Content()),
        container.NewTabItem("3. Savings Calculator", a.calculatorSection.Content()),
        container.NewTabItem("4. Market Opportunity", a.marketSection.Content()),
        container.NewTabItem("5. TRL Roadmap", a.trlSection.Content()),
    )
    
    header := widget.NewLabelWithStyle(
        "Demo 5: Why FeCIM Wins - Technology Comparison",
        fyne.TextAlignCenter,
        fyne.TextStyle{Bold: true},
    )
    
    footer := widget.NewLabel(
        `Dr. external research group: "This could lower the requirements in a data center by 80 to 90%."`
    )
    
    content := container.NewBorder(
        header,  // top
        footer,  // bottom
        nil,     // left
        nil,     // right
        tabs,    // center
    )
    
    a.window.SetContent(content)
    a.window.Resize(fyne.NewSize(1200, 800))
}

func (a *ComparisonApp) Run() {
    a.window.ShowAndRun()
}
```

```go
// demo5-comparison/cmd/comparison-gui/main.go

package main

import "multilayer-ferroelectric-cim-visualizer/demo5-comparison/pkg/gui"

func main() {
    app := gui.NewComparisonApp()
    app.Run()
}
```

### Acceptance Criteria (Phase 3)
- [ ] All 5 sections functional
- [ ] Energy bar chart renders correctly
- [ ] Competitive matrix displays all technologies
- [ ] Calculator produces reasonable numbers
- [ ] Market data displays properly
- [ ] TRL timeline shows current position
- [ ] Tests pass: `go test ./demo5-comparison/...`
- [ ] Demo builds and runs: `./demo5-comparison/comparison-gui`

---

## 6. Phase 4: Polish & Documentation (Week 9-10)

### Week 9: Documentation

#### Task 4.1: Individual Demo READMEs

Update each demo's README with consistent structure:

**Template:**
```markdown
# Demo [N]: [Name]

**Purpose:** [One-line description]  
**Audience:** [Who should watch this]  
**Duration:** [Estimated time]  
**Status:** ✅ Complete

## Quick Start

```bash
cd demo[N]-[name]
go build -o [name]-gui ./cmd/[name]-gui
./[name]-gui
```

## Overview

[What this demo shows]

## Key Features

[Bulleted list of features]

## User Guide

[How to use the demo]

## Technical Details

[Architecture, algorithms used]

## Educational Insights

[What users should learn]

## Key Quote

> "[Dr. Tour quote relevant to this demo]"  
> — Dr. external research group, external research institution

## References

[Papers, sources]

## Tests

```bash
go test ./pkg/... -v
```
```

Apply to:
- `module1-hysteresis/README.md`
- `module2-crossbar/README.md` (update with tab info)
- `module3-mnist/README.md` (full enhancement docs)
- `module4-circuits/README.md`
- `demo5-comparison/README.md` (new)

#### Task 4.2: Main README Polish

Update main README.md with:
1. GIF screenshots of each demo
2. Feature comparison table
3. "From Idea to Demo" timeline
4. Contribution guidelines
5. Citation information

#### Task 4.3: Architecture Documentation

Create `docs/ARCHITECTURE.md`:

```markdown
# Ferroelectric CIM Visualizer - Architecture

## Overview

This project uses a modular architecture with shared utilities and independent demos.

## Directory Structure

```
multilayer-ferroelectric-cim-visualizer/
├── module1-hysteresis/          # Standalone demo
├── module2-crossbar/            # Shared crossbar package
├── module3-mnist/               # Imports demo2/crossbar
├── module4-circuits/            # Standalone demo
├── demo5-comparison/          # Standalone demo
├── shared/                    # Shared utilities
│   ├── theme/                 # FeCIM branding
│   ├── logger/                # Logging
│   └── utils/                 # Common functions
└── docs/                      # Documentation
```

## Code Reuse

### Demo 3 → Demo 2 Dependency

Demo 3 (MNIST) imports crossbar simulation from Demo 2:

```go
import "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"
```

This provides:
- `crossbar.Array` - crossbar data structure
- `ProgramWeights()` - weight programming
- `MatVecMul()` - matrix-vector multiply
- `Config` - noise, ADC/DAC configuration

### Shared Theme

All demos use consistent FeCIM branding:

```go
import "multilayer-ferroelectric-cim-visualizer/shared/theme"

app.Settings().SetTheme(theme.FeCIMTheme())
```

## Testing Strategy

Each demo has its own test suite:

```bash
go test ./module1-hysteresis/... -v
go test ./module2-crossbar/... -v
go test ./module3-mnist/... -v
go test ./module4-circuits/... -v
go test ./demo5-comparison/... -v
```

## Build Process

```bash
# Build all demos
scripts/build-all.sh

# Run all tests
scripts/test-all.sh

# Launch unified interface
scripts/demo-launcher.sh
```
```

---

### Week 10: Final Polish

#### Task 4.4: Create Demo Screenshots/GIFs

```bash
# For each demo, record:
# 1. 30-second GIF showing key interaction
# 2. Static screenshot showing full UI
# 3. "Failure mode" screenshot (for Demo 3)

# Tools:
# - peek (Linux screen recorder → GIF)
# - ffmpeg (video → GIF conversion)
# - scrot (screenshots)

# Save to:
docs/screenshots/
├── module1-hysteresis.gif
├── module1-hysteresis-static.png
├── module2-crossbar-tabs.gif
├── module2-crossbar-irdrop.png
├── module3-mnist-drawing.gif
├── module3-mnist-4zones.png
├── module3-mnist-failure.png
├── module4-circuits.gif
├── module4-circuits-timing.png
├── demo5-comparison-energy.gif
└── demo5-comparison-calculator.png
```

#### Task 4.5: YouTube Video Script (3-5 minutes)

```markdown
# Ferroelectric CIM Visualizer - Demo Reel Script

**Duration:** 4 minutes  
**Audience:** Investors, engineers, press

## Script

[0:00-0:15] HOOK
"What if AI used 10,000 times less energy?
Not in 10 years. Right now.
This is ferroelectric compute-in-memory.
Let me show you."

[0:15-0:45] DEMO 1: The Memory Cell
"Traditional memory stores 0 or 1.
This stores 30 analog levels.
Watch the hysteresis curve trace out as I apply voltage.
Each loop is a stable memory state.
30 states = more precision = better AI."

[0:45-1:30] DEMO 2: The Crossbar Computer
"Here's how we compute with it.
This is a crossbar array - a grid of memory cells.
When I apply input voltages, currents flow through each cell.
The array multiplies an entire matrix in one step.
No data movement. No waiting.

But real hardware has problems.
[Switch to IR Drop tab]
Voltage drops along the wires.
[Switch to Sneak Path tab]
Parasitic currents interfere.

FeCIM handles these better than competition."

[1:30-2:30] DEMO 3: The AI Brain (FLAGSHIP)
"Now the payoff: real AI.
I'll draw a digit... there.
The network runs in two modes:
Digital (ideal): 94% confident it's a 3.
FeCIM (hardware): 89% confident. Still correct.

Watch what happens if I crank up the noise.
[Adjust slider]
Now it misclassifies. But reset to realistic settings...
[Reset]
87% accuracy. That matches Dr. Tour's hardware.

Energy used: 5 microjoules.
A GPU would use 50 millijoules.
10,000 times more."

[2:30-3:15] DEMO 4: The System
"This isn't just a memory cell.
It's a full system: DAene. No exotic materials.
Works on existing fabs.
That's why it can ship."

[3:15-3:50] DEMO 5: Why FeCIM Wins
"Here's the business case.

Energy comparison: FeCIM uses 0.05 picojoules per operation.
GPUs use 100. That's 2000× better.

Competitive matrix: only FeCIM checks every box.
Energy, speed, endurance, CMOS compatible.

Data center calculator: 1000 GPUs running inference.
Current cost: $50,000 per day.
With FeCIM: $500 per day.
Annual savings: $18 million."

[3:50-4:00] CLOSE
"This is ferroelectric compute-in-memory.
Not a concept. Working hardware.
87% accuracy on real chips.

The future of AI is here."

[End card: GitHub repo, contact info]
```

#### Task 4.6: Create CONTRIBUTING.md

```markdown
# Contributing to Ferroelectric CIM Visualizer

Thank you for your interest in contributing!

## How to Contribute

### 1. Report Issues

Found a bug? Have a suggestion?

[Open an issue](https://github.com/XelHaku/multilayer-ferroelectric-cim-visualizer/issues)

Include:
- Demo name and version
- Steps to reproduce
- Expected vs actual behavior
- Screenshots if relevant

### 2. Improve Documentation

Documentation improvements are always welcome:
- Fix typos
- Add examples
- Clarify explanations
- Add references to papers

### 3. Code Contributions

#### Setup
```bash
git clone https://github.com/XelHaku/multilayer-ferroelectric-cim-visualizer
cd multilayer-ferroelectric-cim-visualizer
go mod download
scripts/build-all.sh
scripts/test-all.sh
```

#### Guidelines
- Follow Go conventions (gofmt, golint)
- Add tests for new features
- Update relevant README
- Keep demos independent (except Demo 3 → Demo 2 import)

#### Pull Request Process
1. Fork the repository
2. Create feature branch: `git checkout -b feature/your-feature`
3. Commit changes: `git commit -m "Add: your feature"`
4. Push: `git push origin feature/your-feature`
5. Open Pull Request with description

### 4. Add Scientific References

Know of relevant papers?

Add to:
- `docs/papers/` (with proper citation)
- Update demo READMEs with reference links

## Code of Conduct

Be respectful, constructive, and professional.

## Questions?

Contact: [@XelHaku](https://github.com/XelHaku)

## Acknowledgments

This project is inspired by Dr. external research group's ferroelectric CIM research at external research institution.
```

---

## 7. Implementation Details

### Shared Utilities Package

Create a shared package for consistency across demos:

```
shared/
├── theme/
│   ├── fecim_theme.go         # FeCIM branding colors
│   └── colors.go              # Color constants
├── logger/
│   └── logger.go              # Unified logging
├── widgets/
│   ├── heatmap.go             # Reusable heatmap widget
│   ├── progress_indicator.go # Custom progress bar
│   └── tour_overlay.go        # Guided tour UI
└── utils/
    ├── math.go                # Common math functions
    └── file.go                # File I/O helpers
```

#### FeCIM Theme

```go
// shared/theme/fecim_theme.go

package theme

import (
    "image/color"
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/theme"
)

var (
    ColorFeCIMBlue   = color.RGBA{0, 50, 100, 255}
    ColorFeCIMCyan   = color.RGBA{0, 212, 255, 255}
    ColorFeCIMDark   = color.RGBA{0, 20, 40, 255}
    ColorFeCIMLight  = color.RGBA{230, 230, 230, 255}
)

type fecimTheme struct{}

func FeCIMTheme() fyne.Theme {
    return &fecimTheme{}
}

func (t *fecimTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
    switch name {
    case theme.ColorNameBackground:
        return ColorFeCIMBlue
    case theme.ColorNameForeground:
        return ColorFeCIMLight
    case theme.ColorNamePrimary:
        return ColorFeCIMCyan
    case theme.ColorNameButton:
        return color.RGBA{0, 70, 130, 255}
    default:
        return theme.DefaultTheme().Color(name, variant)
    }
}

func (t *fecimTheme) Font(style fyne.TextStyle) fyne.Resource {
    return theme.DefaultTheme().Font(style)
}

func (t *fecimTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
    return theme.DefaultTheme().Icon(name)
}

func (t *fecimTheme) Size(name fyne.ThemeSizeName) float32 {
    return theme.DefaultTheme().Size(name)
}
```

#### Usage in Demos

```go
// In any demo's main.go

import "multilayer-ferroelectric-cim-visualizer/shared/theme"

func main() {
    app := app.NewWithID("com.fecim.demo")
    app.Settings().SetTheme(theme.FeCIMTheme())
    
    // ... rest of app
}
```

---

## 8. Testing Strategy

### Test Coverage Goals

| Demo | Target Coverage | Critical Paths |
|------|----------------|----------------|
| Demo 1 | 80% | Preisach model, hysteresis loop |
| Demo 2 | 85% | MVM, IR drop, sneak paths |
| Demo 3 | 85% | Quantization, dual inference |
| Demo 4 | 75% | DAC/ADC models |
| Demo 5 | 70% | Calculations, data accuracy |

### Integration Tests

```go
// tests/integration/demo_flow_test.go

package integration

import (
    "testing"
    "time"
)

func TestDemo1ToDemo3Flow(t *testing.T) {
    // Simulates user journey: Demo 1 → Demo 2 → Demo 3
    
    // 1. Demo 1: Learn about 30 levels
    // 2. Demo 2: See how crossbar computes
    // 3. Demo 3: Use crossbar in neural network
    
    // This tests that concepts build on each other
}

func TestAllDemosBuild(t *testing.T) {
    demos := []string{
        "module1-hysteresis",
        "module2-crossbar",
        "module3-mnist",
        "module4-circuits",
        "demo5-comparison",
    }
    
    for _, demo := range demos {
        t.Run(demo, func(t *testing.T) {
            // Verify demo builds without errors
            // (actual build test would use exec.Command)
        })
    }
}
```

### Benchmark Tests

```go
// module3-mnist/pkg/core/quantize_bench_test.go

package core

import "testing"

func BenchmarkQuantizeWeights_30Levels(b *testing.B) {
    weights := make([][]float64, 128)
    for i := range weights {
        weights[i] = make([]float64, 784)
        for j := range weights[i] {
            weights[i][j] = float64(j-392) / 392.0
        }
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = QuantizeWeights(weights, 30)
    }
}

func BenchmarkDualInference(b *testing.B) {
    net := NewDualModeNetwork(784, 128, 10)
    net.LoadWeights("../../data/pretrained_30_h128.json")
    
    input := make([]float64, 784)
    for i := range input {
        input[i] = float64(i) / 784.0
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = net.Infer(input)
    }
}
```

---

## 9. Migration Guide

### For Users

**If you cloned the repository before consolidation:**

```bash
# 1. Backup your local changes
git stash

# 2. Pull latest changes
git pull origin main

# 3. Rebuild all demos
scripts/build-all.sh

# 4. Run new launcher
scripts/demo-launcher.sh
```

**What changed:**
- Demo 7 merged into Demo 2 (now has tabs)
- Demo 6 archived (see `docs/archive/`)
- Demo 8 became Demo 5 (new comparison tool)
- Demo 3 significantly enhanced
- Main README updated to show 5 demos

**What stayed the same:**
- Demo 1 unchanged
- Demo 4 unchanged (minor docs update)
- All existing functionality preserved

### For Contributors

**If you have a pull request targeting old structure:**

1. Rebase onto new `main` branch
2. Update imports if targeting Demo 7:
   - Old: `demo7-nonidealities/pkg/irdrop`
   - New: `module2-crossbar/pkg/crossbar`
3. Update README references (8 demos → 5 demos)

---

## 10. Success Metrics

### Completion Checklist

#### Phase 1: Consolidation (Week 1-2)
- [ ] Demo 7 code merged into Demo 2
- [ ] Demo 2 has 4 functional tabs
- [ ] Demo 6 archived with documentation
- [ ] README updated to reflect 5 demos
- [ ] Unified launcher script works
- [ ] All existing tests still pass

#### Phase 2: Demo 3 Enhancement (Week 3-6)
- [ ] Core quantization function implemented
- [ ] DualModeNetwork class working
- [ ] 4-zone GUI layout complete
- [ ] FP vs CIM toggle functional
- [ ] Hardware controls affect inference
- [ ] Weight heatmap shows 30 levels
- [ ] Failure mode presets work
- [ ] Guided tour runs successfully
- [ ] Quick test shows ~87% with noise=0.08
- [ ] Tests pass with >80% coverage

#### Phase 3: Demo 5 Creation (Week 7-8)
- [ ] Energy comparison bar chart renders
- [ ] Competitive matrix displays correctly
- [ ] Savings calculator produces reasonable results
- [ ] Market section shows opportunity
- [ ] TRL timeline displays progression
- [ ] All 5 sections accessible via tabs
- [ ] Demo builds and runs without errors

#### Phase 4: Polish (Week 9-10)
- [ ] All 5 demo READMEs updated
- [ ] Main README polished with GIFs
- [ ] ARCHITECTURE.md created
- [ ] CONTRIBUTING.md created
- [ ] Screenshots captured for all demos
- [ ] YouTube video script finalized
- [ ] All tests pass: `go test ./...`
- [ ] Build script works: `scripts/build-all.sh`

### Quality Metrics

| Metric | Target | How to Measure |
|--------|--------|----------------|
| Build success rate | 100% | `scripts/build-all.sh` exits 0 |
| Test coverage (demo3) | >80% | `go test -cover ./module3-mnist/pkg/core` |
| Test coverage (overall) | >75% | `go test -cover ./...` |
| Demo launch time | <3s | Time from launch to GUI visible |
| MNIST accuracy (sim) | ~97% | 30 levels, noise=0.01 |
| MNIST accuracy (realistic) | ~87% | 30 levels, noise=0.08 |
| Documentation completeness | 100% | All READMEs have all sections |
| GIF recording | 5/5 demos | One GIF per demo |

### User Experience Metrics

| Metric | Target | How to Validate |
|--------|--------|-----------------|
| New user can launch demo | <5 min | Fresh clone → running demo |
| Demo 3 guided tour completes | 100% | No crashes during 7 steps |
| Failure modes are obvious | Yes | User sees clear visual change |
| Energy savings calculator makes sense | Yes | Results match Dr. Tour's claims |
| TRL progression is clear | Yes | User understands "we are at TRL 4" |

---

## Timeline Summary

```
WEEK 1-2:  Consolidation
  ├─ Merge Demo 7 → Demo 2
  ├─ Archive Demo 6
  ├─ Update documentation
  └─ Create unified launcher

WEEK 3-6:  Demo 3 Enhancement (FLAGSHIP)
  ├─ Week 3: Core infrastructure
  ├─ Week 4: GUI implementation
  ├─ Week 5: Educational features
  └─ Week 6: Polish & testing

WEEK 7-8:  Demo 5 Creation
  ├─ Week 7: Data models & calculations
  └─ Week 8: GUI sections & charts

WEEK 9-10: Final Polish
  ├─ Week 9: Documentation
  └─ Week 10: Screenshots, video, QA

══════════════════════════════════════════
TOTAL: 10 weeks to world-class portfolio
══════════════════════════════════════════
```

---

## Next Steps

### Immediate Actions (Today)

1. **Review this plan** - Confirm approach makes sense
2. **Choose starting point:**
   - Option A: Start Phase 1 (consolidation)
   - Option B: Start Phase 2 (Demo 3 enhancement) in parallel
   - Option C: Create skeleton structure first

3. **Set up tracking:**
   ```bash
   # Create project board or checklist
   cp TODO.md TODO.backup.md
   # Update TODO.md with Phase 1 tasks
   ```

4. **Create feature branch:**
   ```bash
   git checkout -b consolidate-5-demos
   ```

### Weekly Milestones

**End of Week 2:** Demo 2 has tabs, Demo 6/7 archived  
**End of Week 4:** Demo 3 GUI complete  
**End of Week 6:** Demo 3 fully enhanced  
**End of Week 8:** Demo 5 complete  
**End of Week 10:** All docs finalized, video published

---

## Questions to Resolve Before Starting

1. **Priority:** Should we do Phase 1 (consolidation) first, or can we work on Phase 2 (Demo 3) in parallel?
   - **Pro of parallel:** Faster to flagship demo
   - **Con of parallel:** Risk of merge conflicts

2. **Demo 3 weights:** Do we need to retrain for multiple hidden sizes (64/128/256), or just keep 128?
   - **Recommendation:** Start with 128 only, add others later if needed

3. **Demo 5 data:** Can we use Dr. Tour's claimed numbers (10M× energy) even though they're not independently verified?
   - **Recommendation:** Yes, but clearly label as "claimed" vs "verified"

4. **Testing:** Should we write tests as we go, or batch at the end?
   - **Recommendation:** As we go for core logic, batch for GUI

5. **Video:** Record video ourselves or hire professional?
   - **Recommendation:** DIY screen recording + script voiceover, can upgrade later

---

## Conclusion

This plan transforms **8 scattered demos into 5 world-class demonstrations** with:

✅ **Clear narrative:** Physics → Compute → Application → System → Business  
✅ **Flagship showcase:** Demo 3 (MNIST) with full enhancement  
✅ **Investor pitch:** Demo 5 (Comparison) with calculator & charts  
✅ **Clean architecture:** Demo 3 imports Demo 2, no duplication  
✅ **Professional docs:** READMEs, screenshots, video script  
✅ **10-week timeline:** Realistic and achievable  

**The result:** A portfolio that helps Dr. Tour pitch FeCIM to investors, foundries, and the world.

---

**Ready to start?** Let me know which phase you want to tackle first, and I'll provide detailed day-by-day implementation guidance.
