# Work Plan: Module5 Investor-Compelling Technology Comparison Redesign

**Plan ID**: module5-investor-redesign
**Created**: 2026-01-28
**Status**: READY FOR EXECUTION

---

## 1. Context

### 1.1 Original Request
Redesign module5-comparison to create an investor-compelling technology comparison tool that uses ONLY peer-reviewed, verified data from HONESTY_AUDIT.md with DOI citations.

### 1.2 Research Findings

**Current State Analysis:**
- Module5 currently uses UNVERIFIED claims from Dr. Tour's COSM 2025 talk (TRL 4, Tier 5 source)
- Energy constants (`cpuEnergyPJPerMAC = 1000`, `gpuEnergyPJPerMAC = 100`, `fecimEnergyPJPerMAC = 1`) are NOT peer-reviewed
- Hero visualizations display "80-90% DATA CENTER ENERGY REDUCTION" which is UNVERIFIED
- Competitive matrix uses arbitrary checkmarks without DOI citations
- Market projections ($721B) lack proper source citations

**Verified Data Available (from HONESTY_AUDIT.md):**

| Category | Claim | Value | Source | DOI |
|----------|-------|-------|--------|-----|
| Energy vs NAND | Improvement | 25-100x | Samsung Nature 2025 | 10.1038/s41586-025-09793-3 |
| Energy vs GPU (LLM) | Improvement | 70,000x | Nature Comp. Sci. 2025 | 10.1038/s43588-025-00854-1 |
| Power savings | vs NAND | 96% | Samsung Nature 2025 | 10.1038/s41586-025-09793-3 |
| Endurance | Max demonstrated | 10^12 cycles | Nano Letters 2024 | 10.1021/acs.nanolett.4c05671 |
| Endurance | Sliding FE | >10^11 cycles | Science 2024 | 10.1126/science.adp3575 |
| MNIST Accuracy | Best | 98.24% | ScienceDirect 2025 | 10.1016/j.jallcom.2025.034309 |
| MNIST Accuracy | Nature | 96.6% | Nature Commun. 2023 | 10.1038/s41467-023-42110-y |
| Analog States | Maximum | 140 levels | Song, Adv. Science 2024 | 10.1002/advs.202308588 |
| Analog States | Baseline | 32 levels | Oh, IEEE EDL 2017 | 10.1109/LED.2017.2698083 |
| 3D Integration | BEOL | 22nm | CEA-Leti Dec 2024 | (Industry demo) |
| Layer Roadmap | Target | 512 layers | Samsung Nature 2025 | 10.1038/s41586-025-09793-3 |
| Automotive | AEC-Q100 Grade 0 | -40C to 150C | Fraunhofer IPMS 2024 | (Industry cert) |
| Cryogenic | Write speed | 20x faster @ 77K | IEEE 2023 | (Conference) |
| Cryogenic | TCAM energy | 1.36 aJ/bit @ 4K | npj Unconv. Comp. 2025 | 10.1038/s44335-025-00039-z |
| Material Pr | Room temp | 15-34 uC/cm^2 | Nature Commun. 2025 | 10.1038/s41467-025-61758-2 |
| Material Pr | Cryogenic 4K | 75 uC/cm^2 | Adv. Elec. Mat. 2024 | 10.1002/aelm.202300879 |
| Material Ec | Standard | 0.6-1.5 MV/cm | Nature Commun. 2025 | 10.1038/s41467-025-61758-2 |

---

## 2. Work Objectives

### 2.1 Core Objective
Create an technical briefing technology comparison tool that presents ONLY peer-reviewed, verified claims with DOI citations, making the FeCIM value proposition crystal clear while maintaining scientific integrity.

### 2.2 Deliverables
1. **Data Layer**: New verified data structures with DOI citations
2. **Hero Visualizations**: Redesigned with honest, verified metrics
3. **Competitive Matrix**: Honest technology comparison with sources
4. **ROI Calculator**: Real-world workload calculations with peer-reviewed baselines
5. **TRL Transparency**: Clear technology readiness level indicators
6. **Investment Thesis Tab**: Clear narrative with cited evidence

### 2.3 Definition of Done
- [ ] All displayed claims have DOI citations or are marked [UNVERIFIED]
- [ ] No removed/contradicted claims appear (87% MNIST, 10Mx energy)
- [ ] Competitive matrix uses peer-reviewed benchmarks only
- [ ] ROI calculator uses verifiable energy baselines
- [ ] TRL status prominently displayed on all projections
- [ ] All tests pass: `go test ./module5-comparison/...`
- [ ] Visual consistency with shared/theme FeCIM branding

---

## 3. Guardrails

### 3.1 MUST Have
- DOI citations for every numerical claim
- Clear [VERIFIED] / [UNVERIFIED] / [ESTIMATED] labels
- TRL 4 disclaimer on all FeCIM projections
- Honest competitive positioning (FeCIM has weaknesses)
- Sources visible to investors (not hidden in tooltips)

### 3.2 MUST NOT Have
- 87% MNIST claim (REMOVED - below peer-reviewed benchmarks)
- 10,000,000x energy claim (REMOVED - no peer-reviewed data)
- "80-90% data center reduction" without qualification
- Checkmarks without cited evidence
- Any claim without tier 1-2 source OR clear [ESTIMATED] label

---

## 4. Task Flow

```
[T1: Data Layer] --> [T2: Hero Redesign] --> [T5: Integration]
                 --> [T3: Competitive Matrix Redesign] --> [T5]
                 --> [T4: ROI Calculator Redesign] --> [T5]
                                                  --> [T6: Tests]
```

---

## 5. Detailed TODOs

### T1: Create Verified Data Layer
**File**: `module5-comparison/pkg/comparison/verified_data.go` (NEW)
**Estimated**: 150 lines

**Acceptance Criteria**:
- [ ] Define `VerifiedClaim` struct with DOI, Source, Value, Tier fields
- [ ] Create `VerificationStatus` enum: VERIFIED, PLAUSIBLE, UNVERIFIED, ESTIMATED
- [ ] Populate all claims from HONESTY_AUDIT.md with DOIs
- [ ] Export functions to retrieve claims by category
- [ ] Include citation formatter for display

**Code Outline**:
```go
type VerifiedClaim struct {
    Category   string           // "energy", "endurance", "accuracy", etc.
    Claim      string           // Human-readable claim
    Value      string           // "25-100x" or "98.24%"
    NumericMin float64          // For calculations
    NumericMax float64          // For calculations (range)
    Source     string           // "Samsung Nature 2025"
    DOI        string           // "10.1038/s41586-025-09793-3"
    Tier       int              // 1-5 (1=peer-reviewed journal)
    Status     VerificationStatus
}

// Pre-populated verified claims
var EnergyVsNAND = VerifiedClaim{
    Category:   "energy",
    Claim:      "FeFET vs NAND energy efficiency",
    Value:      "25-100x improvement",
    NumericMin: 25,
    NumericMax: 100,
    Source:     "Samsung Nature 2025",
    DOI:        "10.1038/s41586-025-09793-3",
    Tier:       1,
    Status:     Verified,
}
```

---

### T2: Redesign Hero Visualizations
**File**: `module5-comparison/pkg/gui/hero.go` (MODIFY)
**Estimated**: 200 lines modified

**Acceptance Criteria**:
- [ ] Replace "80-90%" with verified "25-100x vs NAND" (Samsung Nature 2025)
- [ ] Add "70,000x vs GPU for LLM workloads" (Nature Comp. Sci. 2025)
- [ ] Show verification badge with DOI on each hero claim
- [ ] Remove unverified energy bar animation (replace with verified comparison)
- [ ] Add prominent TRL indicator

**Key Changes**:

1. **AnimatedEnergyRace** - Replace current implementation:
   - OLD: "80-90% DATA CENTER ENERGY REDUCTION"
   - NEW: "25-100x MORE EFFICIENT THAN NAND" [VERIFIED: Samsung Nature 2025, DOI:10.1038/s41586-025-09793-3]
   - Add secondary: "70,000x vs GPU (LLM workloads)" [VERIFIED]

2. **Remove hardcoded unverified constants**:
   ```go
   // REMOVE these - not peer-reviewed:
   // cpuEnergyPJPerMAC = 1000.0
   // gpuEnergyPJPerMAC = 100.0
   // fecimEnergyPJPerMAC = 1.0

   // REPLACE with verified baseline comparisons
   ```

3. **Add citation display widget**:
   ```go
   type CitationBadge struct {
       Claim  VerifiedClaim
       // Displays: "SOURCE: Samsung Nature 2025 | DOI: 10.1038/..."
   }
   ```

---

### T3: Redesign Competitive Matrix
**File**: `module5-comparison/pkg/gui/market.go` (MODIFY)
**Estimated**: 250 lines modified

**Acceptance Criteria**:
- [ ] Replace arbitrary checkmarks with cited benchmarks
- [ ] Add actual competitor specifications with sources
- [ ] Include TRL column for honest technology readiness comparison
- [ ] Show FeCIM weaknesses honestly (TRL 4, no production scale)
- [ ] Add sortable columns for investor comparison

**New Competitive Matrix Structure**:

| Technology | TRL | Endurance | Energy Efficiency | MNIST Accuracy | Integration | Source |
|------------|-----|-----------|-------------------|----------------|-------------|--------|
| FeFET/HZO | 4-6 | 10^12 cycles | 25-100x vs NAND | 98.24% | 22nm BEOL | Multiple DOIs |
| ReRAM (Crossbar) | 7-8 | 10^6-10^8 | 10x vs DRAM | 95% | Mature | IEEE refs |
| MRAM (Everspin) | 9 | 10^15 | 1x (baseline) | N/A | Production | Datasheet |
| PCM (Intel Optane) | 9 (EOL) | 10^8 | 3x vs NAND | N/A | Discontinued | Intel specs |
| GPU (NVIDIA H100) | 9 | N/A | Baseline | 99%+ | Production | NVIDIA specs |

**Honest FeCIM Weaknesses to Display**:
- TRL 4-6 (not production ready)
- No demonstrated scale manufacturing
- Cycle endurance varies by implementation
- Tour-specific device specs NOT peer-reviewed

---

### T4: Redesign ROI Calculator
**File**: `module5-comparison/pkg/gui/widgets.go` (MODIFY)
**Estimated**: 200 lines modified

**Acceptance Criteria**:
- [ ] Use VERIFIED energy baselines only
- [ ] Show calculation methodology with citations
- [ ] Add confidence intervals based on data ranges
- [ ] Clear "PROJECTED" labels on FeCIM estimates
- [ ] Input validation for realistic workloads

**Calculation Methodology**:
```go
// Use verified comparison ratios, not absolute energy values
type ROICalculation struct {
    Workload           string
    NANDBaseline       float64  // Industry standard (verifiable)
    FeCIMImprovement   Range    // 25-100x (Samsung Nature 2025)
    ConfidenceBand     string   // "Based on peer-reviewed range"
}

// Display format:
// "Projected Savings: $X - $Y million/year"
// "Based on: 25-100x efficiency improvement (Samsung Nature 2025)"
// "Confidence: MEDIUM (range reflects published variance)"
```

---

### T5: Create Investment Thesis Tab
**File**: `module5-comparison/pkg/gui/thesis.go` (NEW)
**Estimated**: 200 lines

**Acceptance Criteria**:
- [ ] Clear value proposition with cited evidence
- [ ] Technology readiness timeline
- [ ] Risk factors with honest assessment
- [ ] Market opportunity with proper citations
- [ ] Competitive moat analysis

**Content Structure**:

1. **Why FeCIM Matters** (1 slide)
   - Memory wall problem (cite Horowitz 2014 ISSCC)
   - Compute-in-memory solution
   - HfO2-ZrO2 physics advantage

2. **Verified Performance** (1 slide)
   - 25-100x vs NAND [DOI]
   - 70,000x vs GPU for LLM [DOI]
   - 10^12 cycle endurance [DOI]
   - 98.24% MNIST accuracy [DOI]

3. **Technology Readiness** (1 slide)
   - TRL 4-6 (Laboratory to Prototype)
   - 22nm BEOL demonstrated (CEA-Leti)
   - Production roadmap: 2-3 years

4. **Risk Factors** (1 slide)
   - Manufacturing scale-up uncertainty
   - Competing technologies (ReRAM, MRAM)
   - No production revenue yet
   - Academic-to-commercial gap

5. **Investment Thesis** (1 slide)
   - $721B addressable market (cite sources)
   - Phased commercialization strategy
   - Capital-light fabless model
   - Clear technical differentiation

---

### T6: Update Tests
**File**: `module5-comparison/pkg/gui/*_test.go` (MODIFY)
**Estimated**: 100 lines

**Acceptance Criteria**:
- [ ] Test that no removed claims appear in UI
- [ ] Test that all displayed claims have DOI citations
- [ ] Test ROI calculator uses verified ranges
- [ ] Test TRL indicators are present

**New Test Cases**:
```go
func TestNoRemovedClaims(t *testing.T) {
    // Verify "87% MNIST" and "10Mx energy" don't appear
}

func TestAllClaimsHaveDOI(t *testing.T) {
    // Verify every VerifiedClaim has non-empty DOI
}

func TestTRLIndicatorsPresent(t *testing.T) {
    // Verify TRL badge appears on FeCIM projections
}
```

---

## 6. Commit Strategy

### Commit 1: Data Layer
```
feat(comparison): add verified data layer with DOI citations

- Add VerifiedClaim struct with source tracking
- Populate all claims from HONESTY_AUDIT.md
- Export functions for claim retrieval by category
```

### Commit 2: Hero Redesign
```
refactor(comparison): replace unverified hero claims with cited data

- Replace "80-90%" with "25-100x vs NAND" (Samsung Nature 2025)
- Add citation badges to hero visualizations
- Remove hardcoded unverified energy constants
```

### Commit 3: Competitive Matrix
```
refactor(comparison): honest competitive matrix with sources

- Replace arbitrary checkmarks with cited benchmarks
- Add TRL column for technology readiness
- Display FeCIM weaknesses honestly
```

### Commit 4: ROI Calculator
```
refactor(comparison): ROI calculator with verified baselines

- Use peer-reviewed comparison ratios
- Add confidence intervals from data ranges
- Clear "PROJECTED" labels on estimates
```

### Commit 5: Investment Thesis
```
feat(comparison): add investment thesis tab with cited evidence

- Clear value proposition with DOI citations
- Technology readiness timeline
- Honest risk factor assessment
```

### Commit 6: Tests
```
test(comparison): add verification and honesty tests

- Test no removed claims appear
- Test all claims have DOI citations
- Test TRL indicators present
```

---

## 7. Success Criteria

### Investor Due-Diligence Ready
- [ ] Every claim traceable to peer-reviewed source
- [ ] Clear distinction between verified/projected data
- [ ] Honest technology readiness assessment
- [ ] No claims that contradict peer-reviewed literature

### Technical Quality
- [ ] All tests pass
- [ ] No resize loops on Wayland (use LayoutCache pattern)
- [ ] 30 FPS animation cap maintained
- [ ] Consistent FeCIM theme branding

### User Experience
- [ ] Clear, compelling value proposition
- [ ] DOI citations accessible (not hidden)
- [ ] TRL status immediately visible
- [ ] Interactive ROI calculator with honest projections

---

## 8. Files to Modify/Create

| File | Action | Lines Est. |
|------|--------|------------|
| `pkg/comparison/verified_data.go` | CREATE | 150 |
| `pkg/gui/hero.go` | MODIFY | 200 |
| `pkg/gui/market.go` | MODIFY | 250 |
| `pkg/gui/widgets.go` | MODIFY | 200 |
| `pkg/gui/thesis.go` | CREATE | 200 |
| `pkg/gui/app.go` | MODIFY | 50 |
| `pkg/gui/*_test.go` | MODIFY | 100 |

**Total Estimated**: ~1150 lines

---

## 9. Dependencies

- `shared/theme` - FeCIM branding colors
- `shared/widgets` - LayoutCache pattern for Wayland stability
- `shared/logging` - Debug logging infrastructure
- `fyne.io/fyne/v2` - GUI framework

---

## 10. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Reduced "wow factor" without exaggerated claims | Medium | Focus on verified 70,000x LLM claim (impressive and true) |
| Investor confusion about TRL | Medium | Clear visual TRL timeline with milestones |
| Competitor data accuracy | Low | Use only published datasheets and peer-reviewed papers |
| ROI calculator precision | Medium | Show ranges, not point estimates |

---

---

## 11. Additional Implementation Notes (Added by Critic Review)

### 11.1 Existing Code Patterns to Follow

Based on codebase analysis, the following patterns MUST be maintained:

1. **Widget Creation Pattern** (from `hero.go`, `market.go`):
   ```go
   type MyWidget struct {
       widget.BaseWidget
       mu           sync.RWMutex  // Always use mutex for state
       animProgress float64       // Animation state
       container    *fyne.Container
       renderer     fyne.WidgetRenderer  // BUG-M5-003: Cache renderer
   }
   ```

2. **BUG-M5-004 Fix Pattern** - Threshold-based text updates:
   ```go
   // Track integer progress to avoid unnecessary text formatting
   lastProgressPct int
   needsTextUpdate bool

   func (w *Widget) UpdateAnimation(dt float64) {
       newPct := int(100 * w.animProgress)
       if newPct != w.lastProgressPct {
           w.needsTextUpdate = true
           w.lastProgressPct = newPct
       }
   }
   ```

3. **Embedded Interface** (from `embedded.go`):
   ```go
   type EmbeddedApp interface {
       BuildContent(fyneApp fyne.App, parentWindow fyne.Window) fyne.CanvasObject
       Start()  // Called when tab selected
       Stop()   // Called when tab deselected
   }
   ```

### 11.2 Color Palette (from `hero.go`)

Use existing technical briefing colors:
```go
var (
    heroTextColor   = color.RGBA{240, 244, 248, 255} // Off-white
    heroCyanColor   = color.RGBA{0, 212, 255, 255}   // FeCIM cyan
    heroGreenColor  = color.RGBA{46, 204, 113, 255}  // Success
    heroRedColor    = color.RGBA{231, 76, 60, 255}   // Baseline/warning
    heroAmberColor  = color.RGBA{243, 156, 18, 255}  // Caution
    heroMutedColor  = color.RGBA{160, 180, 200, 255} // Secondary
    estimatedColor  = color.RGBA{255, 191, 0, 255}   // Unverified/estimated
)
```

### 11.3 Test Patterns (from existing `*_test.go`)

Follow existing test style:
```go
func TestNoRemovedClaims(t *testing.T) {
    // Check constants don't contain removed values
    // Check strings don't contain "87%" or "10,000,000x"
}

func TestVerifiedDataHasDOI(t *testing.T) {
    for _, claim := range AllVerifiedClaims() {
        if claim.Status == Verified && claim.DOI == "" {
            t.Errorf("Verified claim %q missing DOI", claim.Claim)
        }
    }
}
```

### 11.4 Files to NOT Modify

- `pkg/comparison/render.go` - CLI rendering (not related to GUI redesign)
- `cmd/` directories - Entry points only
- `pkg/gui/liveslide.go` - Enums/modes (still needed)

### 11.5 Key Constants to Remove/Replace

**REMOVE these unverified constants from `hero.go` and `app.go`:**
```go
// These are from Tour COSM 2025 (Tier 5, not peer-reviewed):
const cpuEnergyPJ   = 1000.0  // REMOVE
const gpuEnergyPJ   = 100.0   // REMOVE
const fecimEnergyPJ = 1.0     // REMOVE
```

**REPLACE with verified ratio-based approach:**
```go
// Samsung Nature 2025: DOI 10.1038/s41586-025-09793-3
const (
    FeCIMVsNANDMin = 25.0   // Conservative
    FeCIMVsNANDMax = 100.0  // Optimistic
)

// Nature Comp. Sci. 2025: DOI 10.1038/s43588-025-00854-1
const FeCIMVsGPU_LLM = 70000.0  // LLM workloads only
```

---

**PLAN_READY: .omc/plans/module5-investor-redesign.md**
