# IronLattice Demo Honesty Audit

**Critical Assessment of Claims vs. Reality**

**Date:** 2026-01-19

---

## Purpose

This document provides a brutally honest assessment of what our demos claim versus what IronLattice/Dr. Tour actually stated. Claims are categorized as:

| Category | Definition |
|----------|------------|
| **VERIFIED** | Directly matches Dr. Tour's statements in transcript |
| **EXTRAPOLATED** | Derived from his statements but not directly stated |
| **INVENTED** | Made up with no basis in his statements |
| **MISLEADING** | Could deceive readers about IronLattice's actual status |

---

## CRITICAL ISSUES (Must Fix)

### 1. MNIST Accuracy: MISLEADING

**What Dr. Tour Said:**
> "We're at 87% validation here... theoretical is 88%"

**What We Claimed:**
- "95.8% achieved" (multiple README files)
- "Exceeds 87% target"

**The Problem:**
- Our simulation claims to **exceed the theoretical maximum** stated by Dr. Tour
- 88% is the **theoretical max** for their architecture, not a baseline to beat
- We're comparing simulation (idealized) to hardware (real constraints)

**Status:** MISLEADING - partially fixed, but claims still exist in code

---

### 2. Performance Numbers in Demo 8: INVENTED

**File:** `demo8-comparison/pkg/comparison/architecture.go`

```go
// IronLatticeChip() - THESE NUMBERS ARE MADE UP
TDP:         5,   // Where did this come from? NOT in transcript
PeakTOPS:    50,  // INVENTED - no basis
TOPSPerWatt: 10,  // INVENTED - this is THE key claim
ChipArea:    50,  // INVENTED
```

**What Dr. Tour Said:**
- TRL 4 (lab validation only)
- No chip-level specs disclosed
- No TOPS numbers mentioned
- No chip area mentioned

**The Problem:**
We invented specific chip-level performance numbers that have NO basis in Dr. Tour's presentation. These numbers then drive all our "10x", "100x" comparison claims.

**Status:** INVENTED - needs warning labels or removal

---

### 3. Material Parameters: PARTIALLY INVENTED

**File:** `demo1-hysteresis/pkg/ferroelectric/material.go`

| Parameter | Our Value | Dr. Tour Said | Status |
|-----------|-----------|---------------|--------|
| `EnduranceCycles` | 1e11 | "10^9 demonstrated, need 10^12" | **OPTIMISTIC** |
| `RetentionTime` | 3.15e9 (100 years) | "10^7 seconds" (~116 days demonstrated) | **OPTIMISTIC** |
| `Pr` | 30 μC/cm² | Not specified | **ASSUMED** |
| `Ps` | 35 μC/cm² | Not specified | **ASSUMED** |
| `Ec` | 1.0 MV/cm | Not specified | **ASSUMED** |
| `Tau` | 1 ns | "10ns switching" on slides | **OPTIMISTIC** |

**The Problem:**
- We claim 100 year retention when only ~116 days was demonstrated
- We claim 10^11 endurance when only 10^9 was demonstrated
- We claim 1ns switching when 10ns was shown

**Status:** OPTIMISTIC - should be labeled as "target" not "current"

---

### 4. Energy Claims: UNVERIFIED

**What Dr. Tour Said:**
| Claim | Source |
|-------|--------|
| "10,000,000× lower energy than NAND" | Presentation slide |
| "1,000,000× faster than NAND" | Presentation slide |
| "1,000× lower energy than DRAM" | Presentation slide |
| "80-90% data center reduction" | Spoken + slide |

**The Problem:**
These are **marketing claims** for a **TRL 4** technology:
- No peer-reviewed publication supports these numbers
- No independent verification
- Dr. Tour himself said: "we're just at TRL 4... commercialization is 9"
- The slide data shows limited cycle testing (10^2 cycles in one graph)

**Our Treatment:**
We present these as **verified facts** throughout the demos without noting:
- These are unverified claims
- TRL 4 means lab demonstration only
- No independent testing has confirmed these numbers

**Status:** UNVERIFIED - should have disclaimers

---

### 5. "1000x Lower Power" / "1000x Cooler": EXTRAPOLATED

**File:** `demo5-thermal/cmd/thermal/main.go:59`
```go
fmt.Println("  IronLattice: 1000x lower power = cool operation")
```

**What Dr. Tour Said:**
- "1000× lower read/write energy than DRAM"
- "80-90% data center energy reduction"

**The Problem:**
- "1000x lower power" is our simplification of complex energy claims
- Power ≠ Energy (power = energy/time)
- "1000x cooler" is a further extrapolation not in the transcript

**Status:** EXTRAPOLATED - may be misleading

---

## Claims Audit Table

### VERIFIED Claims (Safe to Use)

| Claim | Source Quote | Demo |
|-------|--------------|------|
| 30 discrete analog states | "It's got 30 discrete states" | All |
| 87% MNIST accuracy | "We're at 87% validation here" | Demo 3 |
| 88% theoretical maximum | "theoretical is 88%" | Demo 3 |
| TRL 4 status | "we're just at TRL 4" | Documentation |
| CMOS compatible | "Works on standard CMOS line" | All |
| Non-volatile memory | "nonvolatile" | All |
| Compute-in-memory | "same device does memory and computation" | All |
| 80-90% data center reduction | "lower requirements by 80 to 90%" | Demo 5, 8 |
| Ferroelectric superlattice | "ferroelectric super lattice" | Demo 1 |
| No exotic materials | "no graphene, no exotic materials" | Documentation |

### EXTRAPOLATED Claims (Needs Disclaimer)

| Claim | Derived From | Risk |
|-------|--------------|------|
| 10M× energy improvement | Slide claim (unverified) | Medium |
| 1M× speed improvement | Slide claim (unverified) | Medium |
| 1000× vs DRAM | Slide claim (unverified) | Medium |
| Material parameters (Pr, Ps, Ec) | Literature values | Low |
| Hysteresis loop shape | Standard ferroelectric physics | Low |

### INVENTED Claims (Needs Removal or Strong Disclaimer)

| Claim | Location | Problem |
|-------|----------|---------|
| IronLattice chip specs (5W TDP, 50 TOPS) | Demo 8 | NO BASIS |
| 95.8% MNIST accuracy | Demo 3 | Exceeds stated theoretical max |
| 10^11 endurance | material.go | Only 10^9 demonstrated |
| 100 year retention | material.go | Only 10^7 seconds demonstrated |
| 1ns switching time | material.go | 10ns shown on slides |
| Comparison metrics (10x, 100x vs GPU) | Demo 8 | Based on invented specs |

---

## Recommendations

### Immediate Fixes Required

1. **Demo 3 MNIST:**
   - Remove all "95.8%" claims or clearly label as "simulation only"
   - State: "IronLattice hardware achieved 87% with 88% theoretical maximum"

2. **Demo 8 Comparisons:**
   - Add massive disclaimer that IronLattice specs are ESTIMATED
   - Or remove specific TOPS/Watt claims entirely
   - Note TRL 4 status prominently

3. **Material Parameters:**
   - Change `EnduranceCycles: 1e11` to `1e9` (demonstrated) or label as "target"
   - Change `RetentionTime: 3.15e9` to `1e7` (demonstrated) or label as "target"
   - Add "// ESTIMATED - not disclosed by IronLattice" comments

4. **Energy Claims:**
   - Add disclaimer: "These claims are from Dr. Tour's presentation and have not been independently verified"
   - Note TRL 4 status wherever energy claims appear

### Documentation Standards Going Forward

All performance claims should include:

```markdown
| Metric | Claimed Value | Source | Verification Status |
|--------|---------------|--------|---------------------|
| MNIST accuracy | 87% | Dr. Tour presentation | Demonstrated (TRL4) |
| Endurance | 10^12 cycles | Target | NOT YET ACHIEVED |
| Energy vs NAND | 10^7× lower | Claim | UNVERIFIED |
```

---

## What Dr. Tour Actually Said (Key Quotes)

### Limitations Acknowledged:

> "We still have to get this up to the required 10^12 cycles and eventually hopefully even higher than that."

> "We're just at TRL 4... commercialization is 9. So we have a lot of steps we still need to go."

> "As far as electromagnetic interference... we have no idea. I just don't know."

> "We haven't raised a penny to date."

### Claims Made (But Unverified):

> "10 million times lower read write energy than NAND flash"

> "Million times faster than NAND flash"

> "This could lower the requirements in a data center by 80 to 90%"

---

## Conclusion

Our demos tell a compelling story but contain several claims that:

1. **Exceed IronLattice's stated capabilities** (95.8% vs 87% MNIST)
2. **Use invented specifications** (chip-level TOPS numbers)
3. **Present unverified marketing claims as facts** (10M× energy improvement)
4. **Use optimistic material parameters** (100 year retention vs 116 days demonstrated)

**These issues undermine credibility** and could mislead investors, researchers, or partners who rely on our visualizations.

### Action Items Priority

| Priority | Issue | Action |
|----------|-------|--------|
| **P0** | MNIST 95.8% claim | Remove or prominently disclaim |
| **P0** | Demo 8 invented specs | Add warning banner or remove |
| **P1** | Material parameters | Label as "target" vs "demonstrated" |
| **P1** | Energy claims | Add "unverified" disclaimers |
| **P2** | 1000x claims | Clarify as extrapolation |

---

## Verification Checklist

- [x] Remove or disclaim 95.8% MNIST accuracy (2026-01-19: Fixed in README.md, command.md, demo3.README.md, ELI5.demo3.md)
- [x] Add TRL4 warnings to Demo 8 comparisons (2026-01-19: Added warning banner to main.go and comments to architecture.go)
- [x] Split material parameters into "demonstrated" vs "target" (2026-01-19: Updated IronLatticeMaterial() to use demonstrated values, added IronLatticeMaterialTarget())
- [x] Add "unverified claim" labels to energy comparisons (2026-01-19: Added UNVERIFIED labels to command.md specs table)
- [x] Review all README files for unsupported claims (2026-01-19: Fixed main README.md, command.md)
- [x] Add prominent disclaimer to main README (2026-01-19: Added TRL4 disclaimer banner to README.md)
