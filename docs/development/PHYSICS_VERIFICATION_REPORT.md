# Physics Verification Report - FeCIM EDA Module 6

**Date:** 2026-02-03
**Scope:** Verify physics constants and geometry values used by Module 6 (EDA) GUI, compiler, and LEF/DEF/Liberty exporters
**Status:** ⚠️ PARTIALLY RESOLVED — timing/placeholder constants updated; geometry inconsistencies remain

---

## Executive Summary

Recent updates corrected the **timing, input capacitance, and leakage placeholders** to realistic baseline values in `module6-eda/pkg/config/types.go`. However, **cell dimensions for 1T1R and 2T1R remain inconsistent across the Builder UI, compiler defaults, and LEF export logic**, which can produce mismatched area calculations and exported layouts.

### Critical Findings

| Category | Status | Severity | Notes |
|----------|--------|----------|-------|
| Passive Cell Dimensions | ✅ CONSISTENT | - | 0.460 × 2.720 µm across UI/compiler/export |
| 1T1R Cell Dimensions | ⚠️ INCONSISTENT | HIGH | UI vs compiler vs LEF min-width mismatch |
| 2T1R Cell Dimensions | ⚠️ INCONSISTENT | HIGH | UI vs compiler vs LEF min-width mismatch |
| Timing Placeholders | ✅ UPDATED | LOW | 10 ns rise/fall now in config | 
| Input Capacitance | ✅ UPDATED | LOW | 0.015 pF baseline | 
| Leakage Power | ✅ UPDATED | LOW | 0.0003 nW baseline | 
| DBU Standard | ✅ CORRECT | - | 1000 DBU/µm | 
| Utilization Calculation | ✅ CORRECT | - | 100% for packed array | 

---

## 1. CELL DIMENSIONS CONSISTENCY

### 1.1 Passive Cell (0T1R) - ✅ CONSISTENT

**Current Values (aligned):**
- **Builder UI:** 0.460 × 2.720 µm (`module6-eda/pkg/gui/tabs/builder_validation_tab.go`)
- **Compiler Defaults:** 0.46 × 2.72 µm (`module6-eda/pkg/compiler/types.go`)
- **LEF Export:** uses `cfg.Width/Height` without overrides (`module6-eda/pkg/export/lef.go`)

**Conclusion:** Passive cell dimensions are consistent across UI, compiler, and export.

---

### 1.2 1T1R Cell - ⚠️ INCONSISTENT

| Pipeline Stage | Width (µm) | Height (µm) | Source |
|----------------|------------|-------------|--------|
| Builder UI | 0.460 | 4.070 | `module6-eda/pkg/gui/tabs/builder_validation_tab.go` |
| Compiler Defaults | 0.920 | 3.400 | `module6-eda/pkg/compiler/types.go` (`With1T1R`) |
| LEF Export Minimums | ≥0.920 | ≥3.400 | `module6-eda/pkg/export/lef.go` (`Generate1T1RLEF`) |

**Observed Behavior:**
- If the Builder UI values are used (0.460 × 4.070), LEF export **overrides width to 0.920** but keeps height at 4.070 → exported macro is **0.920 × 4.070**.
- Compiler calculations still assume **0.920 × 3.400**, so area and utilization calculations diverge from LEF output.

**Impact:** Layout outputs, reported area, and utilization are not consistent across pipeline stages.

---

### 1.3 2T1R Cell - ⚠️ INCONSISTENT

| Pipeline Stage | Width (µm) | Height (µm) | Source |
|----------------|------------|-------------|--------|
| Builder UI | 0.920 | 4.070 | `module6-eda/pkg/gui/tabs/builder_validation_tab.go` |
| Compiler Defaults | 1.380 | 3.400 | `module6-eda/pkg/compiler/types.go` (`With2T1R`) |
| LEF Export Minimums | ≥1.380 | ≥3.400 | `module6-eda/pkg/export/lef.go` (`Generate2T1RLEF`) |

**Observed Behavior:**
- Builder UI values (0.920 × 4.070) are **overridden to 1.380 × 4.070** during LEF export.
- Compiler calculations still assume **1.380 × 3.400**, again diverging from exported LEF.

**Impact:** The geometry pipeline is internally inconsistent for 2T1R arrays.

---

## 2. TIMING & PLACEHOLDER PARAMETER VERIFICATION

### 2.1 Rise/Fall Times - ✅ UPDATED

**Current Values:**
- **RiseTime:** 10.0 ns
- **FallTime:** 10.0 ns

**Source:** `module6-eda/pkg/config/types.go`

**Status:** Updated to a realistic baseline for HfO₂ FeFET switching. Still a placeholder until device characterization is available.

---

### 2.2 Input Capacitance - ✅ UPDATED

**Current Value:**
- **InputCap:** 0.015 pF (15 fF)

**Source:** `module6-eda/pkg/config/types.go`

**Status:** Updated to a mid‑range FeFET cell estimate. Placeholder pending measured data.

---

### 2.3 Leakage Power - ✅ UPDATED

**Current Value:**
- **LeakagePower:** 0.0003 nW (0.3 pW)

**Source:** `module6-eda/pkg/config/types.go`

**Status:** Updated to a realistic low‑leakage baseline. Placeholder pending measured data.

---

## 3. EDA STANDARDS VERIFICATION

### 3.1 Database Units - ✅ CORRECT

**Current Value:** 1000 DBU/µm

**Status:** Industry standard; no action required.

---

### 3.2 Utilization Calculation - ✅ CORRECT

**Current Formula:**
```
utilization = (cell_area / array_total_area) × 100%
```

**Status:** Correct for a packed crossbar array; no action required.

---

## 4. REQUIRED CORRECTIONS (NEXT STEPS)

1. **Choose a single canonical 1T1R geometry** (width/height) and propagate it consistently across:
   - Builder UI defaults
   - Compiler defaults (`With1T1R`)
   - LEF min‑width/min‑height overrides
2. **Choose a single canonical 2T1R geometry** and propagate it consistently across the same pipeline.
3. **Document the chosen geometry basis** (e.g., SKY130 site multiples, HVL track height) in both UI and export comments.

Once geometry is aligned, Module 6 EDA outputs will be consistent across UI, compiler, and LEF export.
