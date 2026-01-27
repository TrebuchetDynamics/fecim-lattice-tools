# MNIST Module Fixes TODO

**Generated:** 2026-01-27
**Module:** `module3-mnist/`
**Status:** Pending Review

This document tracks all identified issues, bugs, and improvements for the MNIST module based on a comprehensive expert review.

---

## Table of Contents

1. [Critical Issues (Must Fix)](#critical-issues-must-fix)
2. [High Priority Issues](#high-priority-issues)
3. [Medium Priority Issues](#medium-priority-issues)
4. [Low Priority Issues](#low-priority-issues)
5. [Security Issues](#security-issues)
6. [Architectural Debt](#architectural-debt)
7. [Documentation Gaps](#documentation-gaps)
8. [Test Coverage Gaps](#test-coverage-gaps)

---

## Critical Issues (Must Fix)

### CRIT-001: Nil Pointer Dereference in `softmax` Function
- **File:** `pkg/core/network_inference.go:212`
- **Issue:** Accessing `x[0]` without checking if slice is empty causes panic
- **Impact:** Runtime crash if empty input passed
- **Fix:**
```go
func softmax(x []float64) []float64 {
    if len(x) == 0 {
        return nil
    }
    max := x[0]
    // ...
}
```
- [ ] Implement fix
- [ ] Add unit test for empty input

### CRIT-002: Nil Pointer Dereference in `quantizeADC`
- **File:** `pkg/core/network_inference.go:290-291`
- **Issue:** Accessing `values[0]` without empty check
- **Impact:** Runtime crash
- **Fix:** Add `if len(values) == 0 { return values }` at function start
- [ ] Implement fix
- [ ] Add unit test

### CRIT-003: Inconsistent Minimum Levels Bound
- **File:** `pkg/core/network_config.go:8-9` vs `:96-97`
- **Issue:** `SetNumLevels` allows 1 level, but `SetLayer1Levels` requires 2. `QuantizeWeights` requires `levels >= 2`
- **Impact:** Setting 1 level causes quantization to fail
- **Fix:** Change `SetNumLevels` to clamp to 2, not 1
- [ ] Implement fix
- [ ] Add regression test

---

## High Priority Issues

### HIGH-001: Ignored Crossbar Creation Errors
- **File:** `pkg/gui/embedded.go:36-37`, `pkg/gui/app.go:141-151`
- **Issue:** `crossbar.NewArray` errors silently ignored with `_`
- **Impact:** Nil pointer panic if crossbar creation fails
- **Fix:** Check and handle/log errors
- [ ] Fix in embedded.go
- [ ] Fix in app.go

### HIGH-002: Race Condition in `tryLoadQATWeights`
- **File:** `pkg/gui/dualmode_inference.go:386-445`
- **Issue:** `currentQATLevel` accessed without synchronization
- **Impact:** Data race, potential inconsistent state
- **Fix:** Use atomic operations or mutex for `currentQATLevel`
- [ ] Add mutex protection
- [ ] Run with `-race` flag to verify

### HIGH-003: InferCIMOnly Uses FP Weights Instead of Quantized
- **File:** `pkg/core/network_inference.go:159-177`
- **Issue:** Function named "CIMOnly" but uses `net.FPWeights1/2` instead of `net.QuantWeights1/2`
- **Impact:** Semantically incorrect, misleading function behavior
- **Fix:** Use quantized weights in CIM path
- [ ] Implement fix
- [ ] Add test verifying CIM uses quantized weights

### HIGH-004: Deprecated `rand.Seed` Usage
- **File:** `pkg/training/network_test.go:144`
- **Issue:** `rand.Seed(42)` deprecated since Go 1.20
- **Fix:** Use `rand.New(rand.NewSource(42))` for reproducible tests
- [ ] Update test file

### HIGH-005: Complex Error Channel Handling
- **File:** `pkg/gui/dualmode_inference.go:352-382`
- **Issue:** `updateWeightHeatmapWithProgress` receives from same channel it sends to
- **Impact:** Confusing code, potential deadlock
- **Fix:** Restructure error handling
- [ ] Refactor function

### HIGH-006: Start() Goroutines May Race with UI
- **File:** `pkg/gui/dualmode.go:219-265`
- **Issue:** Goroutines started in `Start()` may race with Fyne's UI initialization
- **Impact:** Potential UI corruption or panic
- **Fix:** Ensure proper synchronization with UI ready state
- [ ] Review and fix timing

### HIGH-007: Magic Number for Energy Ratio
- **File:** `pkg/gui/dualmode_inference.go:192-194`
- **Issue:** `10000` hardcoded instead of using `EnergyRatioGPU` constant
- **Fix:** Use `EnergyRatioGPU` from `dualmode.go:40`
- [ ] Replace all hardcoded values

### HIGH-008: Missing Thread Safety in DigitCanvas
- **File:** `pkg/gui/canvas.go:29-47`
- **Issue:** `pixels` array has no mutex for concurrent access
- **Impact:** Potential data race
- **Fix:** Add mutex or document single-thread access requirement
- [ ] Add synchronization

### HIGH-009: Redundant GetQuantWeights Calls
- **File:** `pkg/gui/dualmode_weights.go:144-149`
- **Issue:** Calls `GetQuantWeights()` twice, inefficient and could get inconsistent values
- **Fix:** Call once and use results
- [ ] Refactor function

---

## Medium Priority Issues

### MED-001: Duplicate Code in runInference and updateResultDisplays
- **File:** `pkg/gui/dualmode_inference.go:29-112` and `:169-234`
- **Issue:** ~90% identical code between functions
- **Fix:** Have `runInference` call `updateResultDisplays`
- [ ] Refactor to remove duplication

### MED-002: Unused Exported Constant
- **File:** `pkg/gui/dualmode.go:29`
- **Issue:** `MNISTTotalMACs` defined but never used
- **Fix:** Either use it or remove it
- [ ] Remove or use constant

### MED-003: Debug Print Statements in Production
- **Files:**
  - `pkg/gui/dualmode.go:178, 188-191, 199`
  - `pkg/gui/dualmode_weights.go:23, 29, 35, etc.`
- **Issue:** `fmt.Println` debug statements should use logging
- **Fix:** Use `mnistLog.Printf()` or remove
- [ ] Clean up debug prints

### MED-004: Inconsistent Max Quantization Level
- **File:** `pkg/core/network_quantization.go:30-31`
- **Issue:** Clamps to 31, but `FeCIMLevels` is 30
- **Fix:** Use consistent constants
- [ ] Standardize bounds

### MED-005: Bug in generateSyntheticData for Digit 7
- **File:** `pkg/gui/app.go:789-798`
- **Issue:** Uses `labels[0]` instead of `i` when drawing digit 7
- **Fix:** Change `images[labels[0]]` to `images[i]`
- [ ] Fix bug
- [ ] Add test

### MED-006: Non-Reproducible Training RNG
- **File:** `pkg/gui/dualmode_controls.go:391-392`
- **Issue:** Local RNG seeded with random value, not reproducible
- **Fix:** Consider fixed seed option for debugging
- [ ] Add debug mode with fixed seed

### MED-007: Misleading forwardCIM Function
- **File:** `pkg/core/network_inference.go:194-197`
- **Issue:** `forwardCIM` just calls `forwardFP`, misleading name
- **Fix:** Remove wrapper or add documentation
- [ ] Document or refactor

### MED-008: Missing Input Validation in Infer
- **File:** `pkg/core/network_inference.go:8-11`
- **Issue:** No validation that `len(input) == net.InputSize`
- **Fix:** Add validation
- [ ] Add input length check

### MED-009: Potential Nil Tooltip Access
- **File:** `pkg/gui/dualmode_weights.go:311-312`
- **Issue:** `h.app.window` could be nil when creating tooltip
- **Fix:** Add nil check
- [ ] Add window nil check

### MED-010: Silent Weight Dimension Mismatch
- **File:** `pkg/training/network.go:352-366`
- **Issue:** Silently ignores weights that don't fit
- **Fix:** Log warning on dimension mismatch
- [ ] Add warning log

### MED-011: Unused Exported Function
- **File:** `pkg/training/network.go:49-61`
- **Issue:** `NewMNISTNetworkWithWeights` has no callers
- **Fix:** Remove or document intended use
- [ ] Remove or document

### MED-012: fmt.Println in Library Code
- **File:** `pkg/training/network.go:522`
- **Issue:** Production library prints to stdout
- **Fix:** Use logging or remove
- [ ] Remove print statement

### MED-013: Memory Allocation from Untrusted File Data
- **File:** `pkg/mnist/loader.go:84, 131`
- **Issue:** Allocates memory based on file header without validation
- **Impact:** Potential memory exhaustion with malicious file
- **Fix:** Add sanity limits (e.g., `maxMNISTImages = 100000`)
- [ ] Add bounds validation

---

## Low Priority Issues

### LOW-001: Inconsistent Error Message Formatting
- **File:** `pkg/core/quantize.go:19`
- **Issue:** Some errors use "quantize:" prefix, others don't
- **Fix:** Standardize error message format
- [ ] Standardize errors

### LOW-002: Variable Shadowing
- **File:** `pkg/gui/dualmode_weights.go:143-144`
- **Issue:** `w2` naming confusing when `w` is parameter
- **Fix:** Rename for clarity
- [ ] Rename variable

### LOW-003: Magic Numbers in Tests
- **File:** `pkg/training/network_test.go:143-145`
- **Issue:** Magic numbers `42`, `64`, `100`, `0.1`
- **Fix:** Use named constants
- [ ] Add constants

### LOW-004: Missing Godoc for Exported Variable
- **File:** `pkg/core/network.go:168-169`
- **Issue:** `AvailableQATLevels` lacks documentation
- **Fix:** Add godoc comment
- [ ] Add documentation

### LOW-005: Inconsistent FeCIM Capitalization
- **File:** `pkg/gui/app.go:39`
- **Issue:** `feCIMTheme` (lowercase) vs `FeCIMDefaultLevels` (uppercase)
- **Fix:** Standardize naming
- [ ] Standardize names

### LOW-006: Missing binary.Read Error Checks
- **File:** `pkg/mnist/loader.go:70-73, 123-124`
- **Issue:** Error returns not checked
- **Fix:** Check all `binary.Read` errors
- [ ] Add error handling

---

## Security Issues

### SEC-001: Unsafe Type Assertion in Test
- **File:** `pkg/core/integration_test.go:419`
- **Issue:** `r.(error)` panics if recover value isn't error type
- **Fix:** Use comma-ok idiom
- [ ] Fix type assertion

### SEC-002: Empty Slice Access (Multiple Locations)
- **Files:**
  - `pkg/core/network_inference.go:212` (softmax)
  - `pkg/core/network_inference.go:234` (argmax)
  - `pkg/core/quantize.go:407`
- **Issue:** Accessing first element without bounds check
- **Fix:** Add length checks
- [ ] Fix all locations

---

## Architectural Debt

### ARCH-001: God Object DualModeApp
- **File:** `pkg/gui/dualmode.go:55-150`
- **Issue:** 97 lines of field declarations, 50+ fields mixing concerns
- **Impact:** Hard to maintain, test, and extend
- **Fix:** Decompose into:
  - `NetworkController` - network state and operations
  - `InferencePresenter` - results display
  - `ControlsPresenter` - hardware config UI
  - `DemoController` - animation/demo logic
- [ ] Plan decomposition
- [ ] Implement refactor

### ARCH-002: Dual Network Implementations
- **Files:** `pkg/core/DualModeNetwork` and `pkg/training/MNISTNetwork`
- **Issue:** Two separate implementations with duplicated functionality
- **Impact:** Maintenance burden, confusion
- **Fix:** Consolidate to single implementation
- [ ] Analyze dependencies
- [ ] Plan consolidation

### ARCH-003: Missing Interfaces
- **Issue:** Core types lack interfaces, preventing mocking/testing
- **Needed:**
  - `NetworkInferer` interface for inference operations
  - `WeightLoader` interface for weight I/O
  - `DataLoader` interface for MNIST data
- [ ] Define interfaces
- [ ] Refactor implementations

### ARCH-004: GUI Business Logic
- **Files:** `pkg/gui/dualmode_inference.go`, `pkg/gui/dualmode_controls.go`
- **Issue:** Inference orchestration, QAT loading in GUI layer
- **Fix:** Extract to `pkg/controller` package
- [ ] Create controller package
- [ ] Move business logic

### ARCH-005: Crossbar Coupling in Training
- **File:** `pkg/training/network.go:11`
- **Issue:** Training package depends on `module2-crossbar`
- **Impact:** Circular conceptual dependency
- **Fix:** Consider interface abstraction
- [ ] Evaluate abstraction

---

## Documentation Gaps

### DOC-001: Missing Architecture Overview
- **Issue:** No document explaining dual-mode architecture
- **Fix:** Create `docs/mnist/mnist.architecture.md`
- [ ] Write architecture doc

### DOC-002: Missing API Reference
- **Issue:** Public APIs lack comprehensive documentation
- **Fix:** Add godoc comments to all exported functions
- [ ] Document pkg/core exports
- [ ] Document pkg/gui exports
- [ ] Document pkg/training exports

### DOC-003: Missing Developer Guide
- **Issue:** No guide for extending/modifying module
- **Fix:** Create `docs/mnist/mnist.development.md`
- [ ] Write developer guide

### DOC-004: Outdated Improvement Plan References
- **File:** `docs/mnist/mnist-module-improvements-plan.md`
- **Issue:** References to non-existent files (e.g., `liveslide.go`)
- **Fix:** Update file references
- [ ] Audit and update references

---

## Test Coverage Gaps

### TEST-001: No GUI Package Tests
- **Package:** `pkg/gui/`
- **Issue:** Zero test files for ~7,000 lines of code
- **Priority:** HIGH
- **Fix:** Add tests for:
  - [ ] `runInference` logic
  - [ ] `tryLoadQATWeights` logic
  - [ ] `changeHiddenSize` logic
  - [ ] Preset application

### TEST-002: Integration Tests Skip on Missing Data
- **File:** `pkg/training/network_test.go:236-238`
- **Issue:** Tests skip when MNIST data missing instead of using synthetic
- **Fix:** Add synthetic data fallback
- [ ] Create synthetic test data generator

### TEST-003: No Concurrency Tests for GUI
- **Issue:** No tests verify thread safety of GUI components
- **Fix:** Add tests with `-race` flag
- [ ] Add concurrency tests

### TEST-004: Hardcoded Test Paths
- **File:** `pkg/training/network_test.go:196`
- **Issue:** Uses `/tmp/test_weights.json`
- **Fix:** Use `t.TempDir()`
- [ ] Update test paths

---

## Progress Tracking

| Category | Total | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 3 | 0 | 3 |
| High | 9 | 0 | 9 |
| Medium | 13 | 0 | 13 |
| Low | 6 | 0 | 6 |
| Security | 2 | 0 | 2 |
| Architecture | 5 | 0 | 5 |
| Documentation | 4 | 0 | 4 |
| Tests | 4 | 0 | 4 |
| **Total** | **46** | **0** | **46** |

---

## Recommended Fix Order

1. **Critical Issues** - Fix immediately (CRIT-001, CRIT-002, CRIT-003)
2. **Security Issues** - Fix before any release (SEC-001, SEC-002)
3. **High Priority** - Fix in next sprint (HIGH-001 through HIGH-009)
4. **Test Coverage** - Add tests alongside fixes
5. **Medium Priority** - Address incrementally
6. **Architecture** - Plan for major refactor cycle
7. **Low Priority** - Address opportunistically
8. **Documentation** - Update as code changes

---

## References

- Code Review Agent Report (2026-01-27)
- Security Review Agent Report (2026-01-27)
- Architecture Analysis Agent Report (2026-01-27)
- Exploration Agent Report (2026-01-27)

---

*This document should be updated as issues are fixed. Check off items as they are completed.*
