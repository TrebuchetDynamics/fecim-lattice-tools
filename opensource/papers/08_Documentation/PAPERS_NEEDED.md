# Papers to Acquire - Priority Action List

**Updated 2026-01-19 | IronLattice Demo Improvements**

---

## HONESTY CHECK: What We DON'T Have

Before listing papers, let's be clear about what **IronLattice has NOT published**:

| What We Need | Status | Impact on Demos |
|--------------|--------|-----------------|
| **Peer-reviewed paper with data** | NOT PUBLISHED | Can't verify any claims |
| **Chip-level specifications** | NOT DISCLOSED | Demo 8 specs are INVENTED |
| **Detailed P-E curve data** | NOT DISCLOSED | Demo 1 uses generic HZO values |
| **MNIST training details** | NOT DISCLOSED | Demo 3 can't match their approach |
| **Endurance test data** | PARTIALLY (10^9 shown) | We claim 10^11 |
| **Energy measurement methodology** | NOT DISCLOSED | "10M×" is unverified |

**Reality Check:** IronLattice is at **TRL 4** (lab validation). Most of our "data" is extrapolated from a single presentation + standard literature values.

---

## CRITICAL - Corrupted Downloads (Need Replacement)

| File | Size | What It Should Be | Where to Get |
|------|------|-------------------|--------------|
| `Mayergoyz_IEEE_1986.pdf` | 16 bytes | "Mathematical Models of Hysteresis" - Preisach model foundations | IEEE Xplore (institutional access) |
| `IEEE_CIM_Survey_2023.pdf` | 244 bytes | "Compute-in-Memory: Recent Trends and Prospects" | IEEE Xplore |
| `Tour_In2Se3_ChemRxiv.pdf` | 60 bytes | Flash Joule Heating synthesis of In₂Se₃ | ChemRxiv or tour@rice.edu |

**Location:** `opensource/papers/09_CORRUPTED/`

---

## HIGH PRIORITY - Papers to Verify/Contradict Our Claims

### Critical for Honesty

These papers are needed to **validate or correct** claims we're making:

| Paper | Why Critical | Our Claim | Need to Verify |
|-------|--------------|-----------|----------------|
| **Oh et al. IEEE EDL 2017** | 32-level FeFET states | We claim 30 levels work | Scheme C pulse details |
| **Jerry et al. IEDM 2017** | 90% MNIST on FeFET | We claim 95.8% (LIES!) | What accuracy is realistic? |
| **Any FeFET energy measurement** | Compare to NAND/DRAM | We repeat 10M× claim | Is this real or marketing? |
| **FeFET endurance studies** | Verify cycle limits | We claim 10^11 | What's actually demonstrated? |

### For Demo 1 (Hysteresis) Improvements

| Paper | Citation | Why Needed | Our Current Problem |
|-------|----------|------------|---------------------|
| **Mayergoyz Preisach Model** | IEEE Trans. Magnetics, 1986 | Rigorous hysteresis math | Using simplified tanh model |
| **Böscke HfO₂ Foundation** | APL 99, 102903 (2011) | HfO₂ parameters | Using estimated values |
| **HZO switching dynamics** | Various | Pulse timing effects | Using 1ns (real is ~10ns) |

### For Demo 2 (Crossbar 30-Level) Improvements

| Paper | Citation | Why Needed | Our Current Problem |
|-------|----------|------------|---------------------|
| **Oh et al. 32 Levels** | IEEE EDL 38(6), 2017 | Scheme C pulses | May be using wrong programming |
| **Crossbar IR drop analysis** | Various | Realistic non-idealities | Model may be optimistic |
| **Sneak path measurements** | Various | Real current leakage | Model may be optimistic |

### For Demo 3 (MNIST) - MOST CRITICAL

| Paper | Citation | Why Needed | Our Current Problem |
|-------|----------|------------|---------------------|
| **Jerry et al. 90% MNIST** | IEDM 2017 | REAL accuracy data | WE CLAIM 95.8% - IMPOSSIBLE |
| **Hardware MNIST implementations** | Multiple | Realistic expectations | Our simulation is too optimistic |
| **Noise impact studies** | Various | Accuracy vs. variation | We don't model enough noise |

---

## MEDIUM PRIORITY - Enhancement Papers

| Paper | Purpose | Demo |
|-------|---------|------|
| Symmetric potentiation/depression | Training accuracy | Demo 3 |
| Preisach-NN self-calibration | Adaptive modeling | Demo 1 |
| Phase-field TDGL methods | Domain animation | Demo 1 |
| Quantization-aware training | Hardware training | Demo 3 |

---

## Papers That CONTRADICT Our Claims

We should actively seek papers that might **disprove** our assumptions:

| Topic | Search For | Why Important |
|-------|-----------|---------------|
| FeFET reliability limits | Papers showing failures | Are our endurance claims too high? |
| Analog computing accuracy limits | Papers showing accuracy ceilings | Is 87% actually good? |
| Energy measurement standards | How to properly measure | Is "10M×" a fair comparison? |
| CIM non-ideality studies | Real hardware limitations | Are our models too optimistic? |

---

## WHERE TO GET PAPERS

### Free / Open Access
- **arXiv.org** - Many preprints available (already have 30+ from here)
- **Author websites** - external research institution, IBM Research, Intel Labs
- **ResearchGate** - Request directly from authors
- **Google Scholar** - Often links to free PDFs
- **Company white papers** - Weebit Nano, Intel, IBM (free)

### Need Institutional Access
- **IEEE Xplore** - Most device papers (CRITICAL papers above)
- **AIP Publishing** - Applied Physics Letters
- **Nature/Science** - Some foundational papers
- **ACM Digital Library** - CS/algorithm papers

### Contact Directly
- **Dr. Tour's Lab**: tour@rice.edu
  - Request In₂Se₃ paper and ferroelectric work
  - Mention IronLattice visualization project
  - **ASK FOR:** Published data to validate our models

---

## HONEST ASSESSMENT: What We Can Actually Verify

| Claim Category | Can We Verify? | Source |
|----------------|----------------|--------|
| 30 discrete states | YES | Standard FeFET physics, multiple papers |
| Hysteresis loops | YES | Fundamental ferroelectric physics |
| CMOS compatibility | PARTIALLY | Literature supports HZO compatibility |
| 10M× energy claim | NO | No independent measurement published |
| 87% MNIST | NO | Only IronLattice's claim, no paper |
| Endurance 10^12 | NO | They said it's a TARGET not achieved |
| Data center 80-90% | NO | Extrapolation, not measured |

---

## DEMO IMPROVEMENT MAPPING (Updated with Honesty)

| Demo | Critical Paper | Issue to Fix | Honest Status |
|------|----------------|--------------|---------------|
| Demo 1 | Mayergoyz 1986 | Preisach model | Model OK, params unverified |
| Demo 2 | Oh et al. 2017 | Scheme C pulses | Unclear if we do this right |
| Demo 3 | Jerry et al. 2017 | 95.8% claim | **MUST FIX - Claims exceed reality** |
| Demo 8 | ANY chip paper | Invented specs | **MUST FIX - Numbers are made up** |

---

## STATUS TRACKING

- [x] arXiv papers downloaded (30+)
- [x] Papers validated with pdftotext
- [x] Corrupted files identified and isolated
- [x] **HONESTY AUDIT completed** (see HONESTY_AUDIT.md)
- [ ] IEEE Xplore access obtained
- [ ] Mayergoyz paper downloaded
- [ ] Oh et al. Scheme C paper downloaded
- [ ] Jerry et al. 90% MNIST paper downloaded - **CRITICAL FOR DEMO 3 FIX**
- [ ] Dr. Tour contacted for published data
- [ ] Demo 3 accuracy claims corrected
- [ ] Demo 8 specs labeled as estimates

---

## Key Takeaway

**We have been treating unverified marketing claims as facts.**

Dr. Tour said: "we're just at TRL 4... we have a lot of steps we still need to go"

Our demos should reflect this uncertainty, not present IronLattice as a proven technology. The papers we need aren't just for "improvement" - they're to verify whether our claims have any basis in reality.

**Current coverage:** 40+ papers downloaded, 3 corrupted, ~10 critical papers needed for verification

---

## Quick Download Commands

```bash
# Check for corrupted files
find opensource/papers -name "*.pdf" -size -1k -exec ls -la {} \;

# Validate a downloaded PDF
pdftotext paper.pdf - | head -20

# Move corrupted files
mv corrupted.pdf opensource/papers/09_CORRUPTED/
```
