# Peripherals Refactor to Shared Package

**Created:** 2026-01-28
**Status:** READY FOR REVIEW
**Type:** Refactoring
**Risk Level:** Medium (cross-module import changes)

---

## 1. Requirements Summary

### What We're Doing
Refactoring the core peripheral circuit implementations (DAC, ADC, TIA, ChargePump, and analysis functions) from `module4-circuits/pkg/peripherals/` to `shared/peripherals/` to enable code reuse across all modules.

### Why We're Doing It
1. **Code Reuse**: Multiple modules (module1-hysteresis, module2-crossbar, module3-mnist) could benefit from these peripheral models
2. **Single Source of Truth**: Currently `shared/peripherals/defaults.go` duplicates config structs that already exist in module4
3. **Consistency**: Shared peripheral models ensure all modules use identical physics parameters
4. **Maintainability**: Changes to DAC/ADC/TIA models only need to happen in one place

### What Stays in module4
- `gpu_peripherals.go` - Contains hardcoded shader paths to `module4-circuits/shaders/*.spv` (lines 88, 104, 121)
- All shader files in `module4-circuits/shaders/`

---

## 2. Acceptance Criteria

### Must Pass
- [ ] `go build ./...` succeeds with zero errors
- [ ] `go test ./...` passes (all 117+ tests)
- [ ] All existing imports from `fecim-lattice-tools/module4-circuits/pkg/peripherals` work via `fecim-lattice-tools/shared/peripherals`
- [ ] No duplicate type definitions (ADCType, DACConfig, ADCConfig, TIAConfig should exist only once)
- [ ] GPU peripherals continue to work (shader paths intact)

### Definition of Done
- Files moved to `shared/peripherals/`
- Import paths updated in all dependent files
- Duplicate types in `defaults.go` removed
- `defaults.go` refactored to use the full structs
- Documentation references updated (optional, can be follow-up)

---

## 3. Files Inventory

### Files to MOVE (copy then delete originals)

| Source File | Lines | Destination |
|-------------|-------|-------------|
| `module4-circuits/pkg/peripherals/dac.go` | 90 | `shared/peripherals/dac.go` |
| `module4-circuits/pkg/peripherals/adc.go` | 123 | `shared/peripherals/adc.go` |
| `module4-circuits/pkg/peripherals/tia.go` | 101 | `shared/peripherals/tia.go` |
| `module4-circuits/pkg/peripherals/chargepump.go` | 127 | `shared/peripherals/chargepump.go` |
| `module4-circuits/pkg/peripherals/analysis.go` | 264 | `shared/peripherals/analysis.go` |

### Files to KEEP in module4

| File | Reason |
|------|--------|
| `module4-circuits/pkg/peripherals/gpu_peripherals.go` | Shader path dependencies (lines 88, 104, 121) |

### Files to UPDATE (import path changes)

| File | Current Import | New Import |
|------|----------------|------------|
| `module4-circuits/pkg/gui/app.go:18` | `fecim-lattice-tools/module4-circuits/pkg/peripherals` | `fecim-lattice-tools/shared/peripherals` |
| `module4-circuits/pkg/gui/embedded.go:8` | `fecim-lattice-tools/module4-circuits/pkg/peripherals` | `fecim-lattice-tools/shared/peripherals` |
| `module4-circuits/pkg/gui/device_state.go:9` | `fecim-lattice-tools/module4-circuits/pkg/peripherals` | `fecim-lattice-tools/shared/peripherals` |
| `module4-circuits/pkg/gui/device_state_test.go:8` | `fecim-lattice-tools/module4-circuits/pkg/peripherals` | `fecim-lattice-tools/shared/peripherals` |
| `module4-circuits/cmd/circuits/main.go:13` | `fecim-lattice-tools/module4-circuits/pkg/peripherals` | `fecim-lattice-tools/shared/peripherals` |
| `cmd/fecim-lattice-tools/integration_test.go:13` | `fecim-lattice-tools/module4-circuits/pkg/peripherals` | `fecim-lattice-tools/shared/peripherals` |

### Files to MODIFY (remove duplicates)

| File | Changes Needed |
|------|----------------|
| `shared/peripherals/defaults.go` | Remove duplicate `ADCType`, `DACConfig`, `ADCConfig`, `TIAConfig` structs and their methods since the full implementations will now be in the shared package |

---

## 4. Implementation Steps

### Phase 1: Copy Files to Shared (Non-Breaking)

**Task 1.1: Copy dac.go**
```
Source: <local-path>
Dest:   <local-path>
```
- No content changes needed (package name already `peripherals`)

**Task 1.2: Copy adc.go**
```
Source: <local-path>
Dest:   <local-path>
```
- No content changes needed

**Task 1.3: Copy tia.go**
```
Source: <local-path>
Dest:   <local-path>
```
- No content changes needed

**Task 1.4: Copy chargepump.go**
```
Source: <local-path>
Dest:   <local-path>
```
- No content changes needed

**Task 1.5: Copy analysis.go**
```
Source: <local-path>
Dest:   <local-path>
```
- No content changes needed

### Phase 2: Resolve Type Conflicts in defaults.go

**Task 2.1: Remove duplicate ADCType from defaults.go**
The `ADCType` enum is defined in both:
- `adc.go` lines 20-26 (the full version we're keeping)
- `defaults.go` lines 72-83 (duplicate with String() method)

Action: Remove lines 71-97 from `defaults.go` (ADCType type + constants + String method)

**Task 2.2: Remove duplicate DACConfig from defaults.go**
The `DACConfig` struct is defined in both places. After moving `dac.go`:
- `dac.go` has full `DAC` struct with methods
- `defaults.go` has simplified `DACConfig` struct

Action: Keep `DefaultDACConfig()` but have it return a `DAC` struct instead, or remove entirely if redundant.

Decision needed: The `DACConfig` in defaults.go (lines 49-57) duplicates the `DAC` struct in dac.go (lines 10-17).

**Recommended approach**:
- Remove `DACConfig` struct from defaults.go (lines 49-69)
- Remove `ADCConfig` struct from defaults.go (lines 99-121)
- Remove `TIAConfig` struct from defaults.go (lines 145-167)
- Keep the constants (they provide semantic names like `DACVrefHigh`)
- The `Default*()` functions in the individual files serve the same purpose as `Default*Config()`

### Phase 3: Update gpu_peripherals.go to Import Shared

**Task 3.1: Update gpu_peripherals.go imports**
The GPU file needs to reference types from the shared package. Add import:
```go
import (
    sharedperiph "fecim-lattice-tools/shared/peripherals"
)
```

The GPU file's `DefaultDACParams()`, `DefaultADCParams()`, `DefaultTIAParams()` functions (lines 489-527) reference the CPU defaults. These should now reference the shared package's `DefaultDAC()`, `DefaultADC()`, `DefaultTIA()`.

### Phase 4: Update All Import Paths

**Task 4.1: Update module4-circuits/pkg/gui/app.go**
```go
// Line 18: Change from:
"fecim-lattice-tools/module4-circuits/pkg/peripherals"
// To:
"fecim-lattice-tools/shared/peripherals"
```

**Task 4.2: Update module4-circuits/pkg/gui/embedded.go**
```go
// Line 8: Change from:
"fecim-lattice-tools/module4-circuits/pkg/peripherals"
// To:
"fecim-lattice-tools/shared/peripherals"
```

**Task 4.3: Update module4-circuits/pkg/gui/device_state.go**
```go
// Line 9: Change from:
"fecim-lattice-tools/module4-circuits/pkg/peripherals"
// To:
"fecim-lattice-tools/shared/peripherals"
```

**Task 4.4: Update module4-circuits/pkg/gui/device_state_test.go**
```go
// Line 8: Change from:
"fecim-lattice-tools/module4-circuits/pkg/peripherals"
// To:
"fecim-lattice-tools/shared/peripherals"
```

**Task 4.5: Update module4-circuits/cmd/circuits/main.go**
```go
// Line 13: Change from:
"fecim-lattice-tools/module4-circuits/pkg/peripherals"
// To:
"fecim-lattice-tools/shared/peripherals"
```

**Task 4.6: Update cmd/fecim-lattice-tools/integration_test.go**
```go
// Line 13: Change from:
"fecim-lattice-tools/module4-circuits/pkg/peripherals"
// To:
"fecim-lattice-tools/shared/peripherals"
```

### Phase 5: Handle GPU Peripherals Package Conflict

**CRITICAL**: After moving files, `module4-circuits/pkg/peripherals/` will only contain `gpu_peripherals.go`. This file is in package `peripherals`, which will conflict with `shared/peripherals` if both are imported.

**Task 5.1: Rename module4 peripherals package**
Option A (Recommended): Rename `gpu_peripherals.go` package to `gpuperiph`:
```go
// Line 2: Change from:
package peripherals
// To:
package gpuperiph
```
And rename directory: `module4-circuits/pkg/peripherals/` -> `module4-circuits/pkg/gpuperiph/`

Option B: Use import alias everywhere (messier)

**Task 5.2: Update any imports of GPU peripherals**
After renaming, update any files importing the GPU peripherals package.

### Phase 6: Delete Original Files from module4

**Task 6.1: Delete moved files**
After all imports are updated and tests pass:
```
rm module4-circuits/pkg/peripherals/dac.go
rm module4-circuits/pkg/peripherals/adc.go
rm module4-circuits/pkg/peripherals/tia.go
rm module4-circuits/pkg/peripherals/chargepump.go
rm module4-circuits/pkg/peripherals/analysis.go
```

### Phase 7: Verify

**Task 7.1: Build verification**
```bash
go build ./...
```

**Task 7.2: Test verification**
```bash
go test ./...
```

**Task 7.3: Manual verification**
- Run the application
- Navigate to Module 4 (Circuits)
- Verify DAC/ADC/TIA displays work
- Verify GPU peripherals work if available

---

## 5. Risk Identification

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Import path typos | Medium | High (build fails) | Run `go build ./...` after each phase |
| Type conflicts (duplicate ADCType) | High | High (compile error) | Remove duplicates in Phase 2 before Phase 4 |
| GPU peripherals break | Medium | Medium | Keep gpu_peripherals.go shader paths unchanged |
| Missing import updates | Medium | High | Use grep to find ALL occurrences before declaring done |
| Package name conflict | High | High | Rename module4 package to `gpuperiph` |

### Rollback Plan
If issues arise:
1. Git restore all modified files: `git checkout -- .`
2. Delete new files in shared/peripherals/
3. The codebase returns to previous state

---

## 6. Verification Steps

### Build Verification
```bash
cd <local-path>
go build ./...
# Expected: Exit 0, no errors
```

### Test Verification
```bash
go test ./...
# Expected: All 117+ tests pass
```

### Import Verification
```bash
# Should return NO results for old path
grep -r "fecim-lattice-tools/module4-circuits/pkg/peripherals" --include="*.go" .

# Should return results ONLY for gpu_peripherals references (if any remain)
```

### Runtime Verification
1. Run `./fecim-lattice-tools`
2. Navigate to Module 4 (Peripheral Circuits)
3. Verify DAC tab shows voltage conversion
4. Verify ADC tab shows quantization
5. Verify TIA tab shows current-to-voltage conversion
6. Verify Analysis tab shows INL/DNL plots

---

## 7. Task Summary (For Executor)

| # | Task | File(s) | Blocking |
|---|------|---------|----------|
| 1.1-1.5 | Copy 5 files to shared/peripherals/ | dac.go, adc.go, tia.go, chargepump.go, analysis.go | None |
| 2.1-2.2 | Remove duplicate types from defaults.go | shared/peripherals/defaults.go | 1.x |
| 3.1 | Rename module4 peripherals package to gpuperiph | gpu_peripherals.go + directory | 1.x |
| 4.1-4.6 | Update 6 import paths | gui/*.go, main.go, integration_test.go | 2.x, 3.1 |
| 5.1 | Delete original files from module4 | module4-circuits/pkg/peripherals/*.go (except gpu) | 4.x |
| 6.1 | Build verification | - | 5.1 |
| 6.2 | Test verification | - | 6.1 |

---

## 8. Commit Strategy

### Single Commit (Recommended)
All changes in one atomic commit:
```
refactor: move peripheral circuits from module4 to shared for cross-module reuse

- Move dac.go, adc.go, tia.go, chargepump.go, analysis.go to shared/peripherals/
- Remove duplicate type definitions from shared/peripherals/defaults.go
- Rename module4-circuits/pkg/peripherals to gpuperiph (GPU-specific code stays)
- Update all import paths across 6 files

This enables module1-hysteresis, module2-crossbar, and module3-mnist
to use the same peripheral circuit models.
```

---

PLAN_READY: .omc/plans/peripherals-refactor-to-shared.md
