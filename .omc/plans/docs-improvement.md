# Plan: Agent-Optimized Documentation Improvement

**Created:** 2026-01-25
**Status:** READY
**Estimated Effort:** Medium (2-3 hours)

---

## Context

### Original Request
Improve `docs/development/scriptReference.md` for agent consumption and update `CLAUDE.md` to reference it properly.

### Problem Statement
Current documentation is comprehensive but not optimized for AI agent lookups:
- No decision trees for common tasks
- No error resolution guide
- Missing thread safety annotations
- No file dependency information
- Functions lack parameter/return types
- CLAUDE.md duplicates information instead of referencing scriptReference.md

### Research Findings
- 522 lines in scriptReference.md - good foundation but prose-heavy
- 131 lines in CLAUDE.md - contains duplicated info from scriptReference
- 6 modules with consistent embedded.go pattern
- `fyne.Do()` used in 30+ locations for thread-safe UI updates
- All modules follow same import pattern from multilayer-ferroelectric-cim-visualizer/

---

## Work Objectives

### Core Objective
Transform documentation into agent-optimized format where any codebase question can be answered in 3 lookups or fewer.

### Deliverables
1. **Improved scriptReference.md** with:
   - Quick Decision Trees (table format)
   - Common Patterns Section with code examples
   - Error Resolution Guide
   - File Dependency Graph
   - Thread Safety annotations
   - Full function signatures with types

2. **Updated CLAUDE.md** with:
   - "For Agents" section pointing to scriptReference.md
   - Removed duplication (reference scriptReference.md instead)
   - Quick-start decision tree

### Definition of Done
- [ ] Agent can find "how to add a new demo" in 1 lookup
- [ ] Agent can find "how to fix UI update crash" in 1 lookup
- [ ] Agent can find correct import path in 1 lookup
- [ ] CLAUDE.md is under 100 lines (currently 131)
- [ ] No information duplication between files

---

## Guardrails

### MUST Have
- Preserve all physics constants and citations
- Maintain scientific accuracy policy section
- Keep all existing function references (add types, don't remove)
- Use table format for decision trees

### MUST NOT Have
- Prose paragraphs for lookup information (use tables)
- Duplicated information between CLAUDE.md and scriptReference.md
- Broken links or file references
- Changes to actual source code

---

## Task Flow

```
[T1: Restructure scriptReference.md]
    |
    v
[T2: Add Decision Trees]
    |
    v
[T3: Add Error Resolution Guide]
    |
    v
[T4: Add Thread Safety Annotations]
    |
    v
[T5: Update CLAUDE.md]
    |
    v
[T6: Verify Cross-References]
```

---

## Detailed TODOs

### T1: Restructure scriptReference.md Header
**File:** `docs/development/scriptReference.md`
**Priority:** HIGH
**Effort:** 15 min

Add agent-focused header section at top of file:

```markdown
# FeCIM Script Reference (Agent-Optimized)

## Quick Navigation (For AI Agents)

| I need to... | Go to section |
|--------------|---------------|
| Find a function | [Quick Function Lookups](#quick-function-lookups) |
| Add a new demo | [Adding New Demo Pattern](#adding-new-demo-pattern) |
| Fix a UI crash | [Error Resolution](#error-resolution-guide) |
| Understand imports | [Import Patterns](#import-patterns) |
| Check thread safety | [Thread Safety Guide](#thread-safety-guide) |
| Find file dependencies | [Module Dependencies](#module-dependencies) |
```

**Acceptance Criteria:**
- [ ] Table of contents uses "I need to..." format
- [ ] All section anchors are valid
- [ ] Navigation table is in first 20 lines

---

### T2: Add Decision Trees Section
**File:** `docs/development/scriptReference.md`
**Priority:** HIGH
**Effort:** 30 min

Add after Quick Navigation:

```markdown
## Decision Trees

### "I need to modify..." Decision Tree

| Modify What | File Location | Key Type/Function |
|-------------|---------------|-------------------|
| Crossbar physics | `module2-crossbar/pkg/crossbar/array.go` | `Config`, `Array`, `MVM()` |
| Hysteresis model | `module1-hysteresis/pkg/ferroelectric/preisach.go` | `PreisachModel`, `Update()` |
| MNIST inference | `module3-mnist/pkg/core/network.go` | `DualModeNetwork`, `Infer()` |
| Circuit peripherals | `module4-circuits/pkg/peripherals/*.go` | `DAC`, `ADC`, `TIA` |
| Theme/colors | `shared/theme/theme.go` | `ColorPrimary`, `ColorBackground` |
| Non-idealities | `module2-crossbar/pkg/crossbar/nonidealities.go` | `AnalyzeIRDrop()`, `AnalyzeSneakPaths()` |

### "I need to add..." Decision Tree

| Add What | Template File | Required Interface |
|----------|---------------|-------------------|
| New demo module | Copy `module4-circuits/pkg/gui/embedded.go` | `NewEmbedded*App()`, `BuildContent()`, `Start()`, `Stop()` |
| New tab to existing demo | See `module2-crossbar/pkg/gui/tabs/*.go` | Tab struct with `CreateContent()` method |
| New widget | `shared/widgets/*.go` | Implement `fyne.Widget` interface |
| New physics test | `*_test.go` in same package | `Test*` function with `*testing.T` |

### "I need to debug..." Decision Tree

| Problem | First Check | Solution Pattern |
|---------|-------------|------------------|
| UI not updating | Missing `fyne.Do()` wrapper | Wrap UI update in `fyne.Do(func() { ... })` |
| Nil pointer in GUI | Widget not initialized | Check `BuildContent()` called before `Start()` |
| Wrong quantization | Check `FeCIMLevels` constant | Use `crossbar.QuantizeTo30Levels()` |
| Import error | Check module path | Use `multilayer-ferroelectric-cim-visualizer/module*` |
| Test fails on CI | GUI test without display | Mock or skip with `t.Skip()` |
```

**Acceptance Criteria:**
- [ ] All decision trees use table format
- [ ] File paths are absolute from repo root
- [ ] Each row links to a specific file and function

---

### T3: Add Error Resolution Guide
**File:** `docs/development/scriptReference.md`
**Priority:** HIGH
**Effort:** 25 min

Add new section:

```markdown
## Error Resolution Guide

### Common Errors and Fixes

| Error Message | Cause | Fix |
|---------------|-------|-----|
| `panic: runtime error: invalid memory address` | UI update from goroutine | Wrap in `fyne.Do(func() { ... })` |
| `undefined: crossbar.NewArray` | Wrong import path | Import `multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar` |
| `type *XxxApp has no field or method BuildContent` | Missing embedded interface | Implement full interface: `BuildContent()`, `Start()`, `Stop()` |
| `cannot use x (type float64) as type int` | Quantization type mismatch | Use `int(crossbar.QuantizeTo30Levels(x))` for level index |
| `fyne: no OpenGL context` | GUI test without display | Add `t.Skip("Requires display")` or mock |
| `weights not loaded` | Missing weight file | Call `LoadWeights()` before `Infer()` |
| `conductance out of range` | Value > 1.0 or < 0.0 | Normalize to [0, 1] before `ProgramWeight()` |

### Thread Safety Errors

| Symptom | Check | Fix |
|---------|-------|-----|
| Random crashes on UI update | Update from goroutine? | Use `fyne.Do()` |
| Race condition warnings | Shared state access? | Use mutex or channel |
| Frozen UI | Blocking on main thread? | Move to goroutine with `fyne.Do()` callback |

### Build Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `go: module not found` | Module not in go.mod | Run `go mod tidy` |
| `cgo: pkg-config not found` | Missing system deps | Install `libgl1-mesa-dev` (Linux) |
| `vulkan headers not found` | Missing Vulkan SDK | Install `vulkan-sdk` or skip Vulkan mode |
```

**Acceptance Criteria:**
- [ ] At least 10 common errors covered
- [ ] Each error has exact error message, cause, and fix
- [ ] Thread safety section separate from general errors

---

### T4: Add Thread Safety and Import Patterns
**File:** `docs/development/scriptReference.md`
**Priority:** MEDIUM
**Effort:** 25 min

Add sections:

```markdown
## Thread Safety Guide

### Functions Requiring fyne.Do() Wrapper

Any function that updates UI from a goroutine MUST use `fyne.Do()`:

| Component | Safe Pattern |
|-----------|--------------|
| Label text | `fyne.Do(func() { label.SetText("new") })` |
| Container add | `fyne.Do(func() { container.Add(widget) })` |
| Refresh widget | `fyne.Do(func() { widget.Refresh() })` |
| Progress bar | `fyne.Do(func() { progress.SetValue(0.5) })` |
| Status updates | `fyne.Do(func() { app.updateStatus("msg") })` |

### Thread-Safe Patterns

```go
// WRONG - will crash randomly
go func() {
    label.SetText("Updated")  // NO: direct UI update
}()

// CORRECT - thread-safe
go func() {
    result := heavyComputation()
    fyne.Do(func() {
        label.SetText(result)  // YES: wrapped in fyne.Do()
    })
}()
```

## Import Patterns

### Standard Module Imports

```go
import (
    // Core crossbar
    "multilayer-ferroelectric-cim-visualizer/module2-crossbar/pkg/crossbar"

    // Hysteresis model
    "multilayer-ferroelectric-cim-visualizer/module1-hysteresis/pkg/ferroelectric"

    // MNIST network
    "multilayer-ferroelectric-cim-visualizer/module3-mnist/pkg/core"

    // Circuit peripherals
    "multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/peripherals"

    // Comparison
    "multilayer-ferroelectric-cim-visualizer/module5-comparison/pkg/comparison"

    // EDA compiler
    "multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/compiler"

    // Shared theme
    "multilayer-ferroelectric-cim-visualizer/shared/theme"

    // Fyne GUI
    "fyne.io/fyne/v2"
    "fyne.io/fyne/v2/container"
    "fyne.io/fyne/v2/widget"
)
```

## Module Dependencies

### Dependency Graph (which modules import which)

```
cmd/fecim-visualizer
    ├── shared/theme
    ├── module1-hysteresis/pkg/gui
    ├── module2-crossbar/pkg/gui
    ├── module3-mnist/pkg/gui
    ├── module4-circuits/pkg/gui
    └── module5-comparison/pkg/gui

module3-mnist/pkg/core
    └── module2-crossbar/pkg/crossbar  (for quantization)

module4-circuits/pkg/peripherals
    └── (no internal deps - standalone)

module5-comparison/pkg/comparison
    └── (no internal deps - uses own architecture types)
```

### Safe to Modify (no dependents)
- `shared/theme/` - Only imported by GUI packages
- `module*-*/pkg/gui/` - Only imported by main app
- `docs/` - No code imports

### Modify with Care (has dependents)
- `module2-crossbar/pkg/crossbar/` - Imported by module3-mnist
- `shared/widgets/` - Imported by all GUI packages
```

**Acceptance Criteria:**
- [ ] All fyne.Do() patterns documented with examples
- [ ] Import paths are copy-paste ready
- [ ] Dependency graph shows what imports what

---

### T5: Update CLAUDE.md
**File:** `CLAUDE.md`
**Priority:** HIGH
**Effort:** 20 min

Rewrite to remove duplication and add agent section:

```markdown
# CLAUDE.md - FeCIM Lattice Tools

## For AI Agents

**Full reference:** See `docs/development/scriptReference.md` for:
- Function lookups and signatures
- Decision trees for common tasks
- Error resolution guide
- Thread safety patterns

**Quick decision:**
| I need to... | Look in |
|--------------|---------|
| Find a function | `docs/development/scriptReference.md#quick-function-lookups` |
| Fix an error | `docs/development/scriptReference.md#error-resolution-guide` |
| Add a feature | `docs/development/scriptReference.md#decision-trees` |
| Run tests | See [Testing](#testing) below |

## Overview

Go-based lattice tool suite for Ferroelectric Compute-in-Memory (FeCIM).

**Core concept**: 30 discrete analog states per cell (~4.9 bits/cell).

> **Primary Source**: Dr. external research group, COSM 2025 - [Transcript](docs/videos/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md)

## Build & Run

```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
```

## Key Rules

### Do
- Use `fyne.Do(func() { ... })` for all UI updates from goroutines
- Quantize to 30 levels: `crossbar.QuantizeTo30Levels(value)`
- Follow the embedded app interface pattern
- Run `go test ./...` before committing

### Don't
- Modify `module2-crossbar/pkg/_layers_experimental/` - archived
- Add demos without implementing the embedded interface
- Use blocking operations on the main UI thread
- Commit binaries

## Physics Constants

| Parameter | Value | Source |
|-----------|-------|--------|
| FeCIM Levels | 30 | Dr. Tour COSM 2025 |
| Pr | 15-34 uC/cm^2 | Nature Commun. 2025 |
| Ec | 1.0-1.5 MV/cm | Nature Commun. 2025 |
| Endurance | 10^12+ cycles | PMC 2024, IEEE IRPS 2022 |

## Accuracy & Honesty Policy

Scientific accuracy over marketing claims. See full policy in `docs/development/scriptReference.md`.

| Claim | Status |
|-------|--------|
| 30 analog states | Verified (Dr. Tour + peer-reviewed) |
| 87% MNIST accuracy | Verified |
| 10^12 cycle endurance | Target (literature shows path) |
| 10Mx vs NAND energy | Unverified (Dr. Tour claim only) |

## Testing

```bash
go test ./...                            # All tests
go test ./module2-crossbar/pkg/crossbar  # Crossbar only
```

Full test documentation: `docs/development/TESTING.md`

## Git Conventions

- Commit: `type: description` (feat, fix, docs, refactor, test, chore)
- Run tests before pushing

## Ignore

- `logs/`, `output/`, `docs/archive/`
- `module2-crossbar/pkg/_layers_experimental/`
```

**Acceptance Criteria:**
- [ ] CLAUDE.md is under 100 lines
- [ ] "For AI Agents" section is first content section
- [ ] All detailed info references scriptReference.md
- [ ] No prose duplication - just tables and key rules

---

### T6: Add Common Patterns Section with Examples
**File:** `docs/development/scriptReference.md`
**Priority:** MEDIUM
**Effort:** 20 min

Add after Decision Trees:

```markdown
## Common Patterns

### Adding a New Demo Module

1. Create directory structure:
```
module7-newdemo/
    cmd/newdemo-gui/main.go
    pkg/newdemo/logic.go
    pkg/gui/
        app.go
        embedded.go  # Required for unified app
```

2. Implement embedded interface in `pkg/gui/embedded.go`:
```go
package gui

import "fyne.io/fyne/v2"

type EmbeddedNewDemoApp struct {
    // internal state
}

func NewEmbeddedNewDemoApp() *EmbeddedNewDemoApp {
    return &EmbeddedNewDemoApp{}
}

func (app *EmbeddedNewDemoApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
    // Create and return UI content
    return widget.NewLabel("New Demo")
}

func (app *EmbeddedNewDemoApp) Start() {
    // Called when tab selected - start animations, load data
}

func (app *EmbeddedNewDemoApp) Stop() {
    // Called when tab deselected - stop animations, cleanup
}
```

3. Register in `cmd/fecim-visualizer/main.go`:
```go
import newdemo "multilayer-ferroelectric-cim-visualizer/module7-newdemo/pkg/gui"

// In main(), add to DemoApp:
newDemoApp := newdemo.NewEmbeddedNewDemoApp()
```

### Adding a Physics Parameter

1. Define constant in appropriate package:
```go
// module2-crossbar/pkg/crossbar/array.go
const (
    FeCIMLevels = 30  // From Dr. Tour COSM 2025
    NewParam    = 42  // Add citation comment
)
```

2. Add to physics tests in `*_test.go`:
```go
func TestNewParamPhysics(t *testing.T) {
    // Test the physical validity
}
```

3. Document in `CLAUDE.md` Physics Constants table.

### Updating Quantization Logic

Location: `module2-crossbar/pkg/crossbar/array.go`

```go
// QuantizeTo30Levels maps [0,1] to one of 30 discrete levels
func QuantizeTo30Levels(value float64) float64 {
    if value < 0 {
        value = 0
    } else if value > 1 {
        value = 1
    }
    level := int(value * float64(FeCIMLevels-1) + 0.5)
    return float64(level) / float64(FeCIMLevels-1)
}
```

To modify quantization:
1. Change `FeCIMLevels` constant (with citation)
2. Update tests in `module2-crossbar/pkg/crossbar/physics_test.go`
3. Run `go test ./module2-crossbar/...` to verify
```

**Acceptance Criteria:**
- [ ] At least 3 complete code examples
- [ ] Examples are copy-paste ready
- [ ] Each pattern has numbered steps

---

### T7: Verify Cross-References
**File:** Both files
**Priority:** LOW
**Effort:** 10 min

Steps:
1. Check all section anchors in scriptReference.md match links
2. Verify CLAUDE.md references to scriptReference.md are valid
3. Confirm no broken file paths

**Acceptance Criteria:**
- [ ] All `#anchor` links work
- [ ] All `docs/development/` paths exist
- [ ] `go test ./...` still passes

---

## Commit Strategy

### Single Commit
```
docs: optimize scriptReference.md and CLAUDE.md for AI agent consumption

- Add decision trees for common tasks (modify/add/debug)
- Add error resolution guide with exact error messages
- Add thread safety guide with fyne.Do() patterns
- Add import patterns and module dependencies
- Add common patterns with code examples
- Refactor CLAUDE.md to reference scriptReference.md
- Reduce CLAUDE.md from 131 to <100 lines

Agents can now find answers in 3 lookups or fewer.
```

---

## Success Criteria

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| Lookups to find "add demo" | 3+ (scan file) | 1 (decision tree) | <=3 |
| Lookups to fix UI crash | Unknown | 1 (error table) | <=3 |
| CLAUDE.md lines | 131 | <100 | <100 |
| Duplication | High | None | None |
| Decision tree coverage | 0 | 3 tables | 3+ |
| Error resolutions | 0 | 10+ | 10+ |

---

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Broken anchor links | Medium | T7 verification step |
| Missing common error | Low | Add as discovered |
| Info loss in CLAUDE.md trim | Low | Reference scriptReference.md |

---

## Notes for Executor

- All changes are in `.md` files only - no code changes
- Use exact file paths shown
- Preserve existing good content - add structure around it
- Test anchor links after editing
