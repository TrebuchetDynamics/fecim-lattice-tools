# Hysteresis Module UI Improvement Plan (v2)

## Executive Summary

Transform module1-hysteresis from a demonstration tool into a highly interactive, responsive, and maintainable FeCIM research workbench. Apply Dr. Shin's expert perspective on ferroelectric memory UI/UX best practices.

---

## Current State Analysis

### Architecture Overview
- **Main File**: `pkg/gui/gui.go` (2863 lines) - monolithic, handles everything
- **Unused Assets**: `labench.go` (keyboard controls using GLFW), `overlay.go` (ASCII overlays) - not wired to Fyne GUI
- **Physics Core**: `pkg/ferroelectric/` - well-tested Preisach model
- **Custom Widgets**: PEPlot (367 lines), LevelIndicator (246 lines), CellVisualizer (223 lines), ModeIndicator (136 lines) - all embedded in gui.go

### Line Count Breakdown (gui.go)
| Section | Lines | Notes |
|---------|-------|-------|
| PEPlot widget | ~367 | Lines 1746-2113 |
| LevelIndicator widget | ~246 | Lines 2119-2365 |
| CellVisualizer widget | ~223 | Lines 2372-2595 |
| ModeIndicator widget | ~136 | Lines 2602-2738 |
| **Total widgets** | ~972 | Can be extracted |
| Controls panel | ~290 | createControlsPanel + related |
| Info/slide/log panels | ~200 | createInfoPanel, createSlidePanel, createLogPanel |
| Simulation loop | ~400 | simulationLoop, waveform generation |
| App struct + lifecycle | ~200 | NewApp, Run, createUI |
| Debug logging | ~100 | saveDebugLog, initDebugLog |
| Slide text generation | ~180 | getSlideText |
| Theme + helpers | ~155 | feCIMTheme, fixedMinWidthLayout |
| **Remaining after extraction** | ~525 | Core orchestration |

### Critical Issues Found

| Issue | Severity | Impact |
|-------|----------|--------|
| **Keyboard shortcuts not working** | HIGH | LabBench uses GLFW keys, Fyne needs its own key handling |
| **No visual phase indicators on plot** | MEDIUM | Users can't see SATURATE/SETTLE/HOLD/READ phases visually |
| **2863 line monolithic file** | HIGH | Difficult to maintain, test, or extend |
| **Interactive mode lacks feedback** | MEDIUM | No animation of E-field ramp during phases |
| **Debug log unbounded growth** | LOW | Memory leak potential in long sessions |
| **Temperature effects not visualized** | MEDIUM | Landau coefficients computed but not displayed |

---

## Improvement Plan

### Phase 1: Keyboard Shortcuts Integration (HIGH PRIORITY)

**Goal**: Add Fyne-native keyboard controls for power users.

**Important Note**: LabBench uses `glfw.Key` constants which are incompatible with Fyne. We will implement Fyne-native keyboard handling directly rather than trying to bridge GLFW.

**Implementation**:
1. Add keyboard handler using Fyne's correct API:
   ```go
   // In run() after window creation:
   a.mainWindow.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
       a.handleKeyPress(ke)
   })
   ```
2. Implement `handleKeyPress()` method with Fyne key mappings:
   - `fyne.KeyE` / `fyne.KeyD` → E-field amplitude (Manual mode only)
   - `fyne.KeyT` / `fyne.KeyG` → Temperature slider (±25K)
   - `fyne.KeyF` / `fyne.KeyV` → Frequency (×2 / ÷2)
   - `fyne.KeyW` → Cycle waveform
   - `fyne.KeySpace` → Pause/Resume
   - `fyne.KeyR` → Reset
   - `fyne.KeySlash` (`?`) → Show keyboard help dialog
3. Add visual feedback when key is pressed (brief status bar flash)
4. Create help dialog showing all shortcuts

**Files to modify**: `pkg/gui/gui.go` (add `handleKeyPress()` method, modify `run()`)
**Estimated changes**: ~100 LOC additions

**Testing verification**:
```bash
go test ./module1-hysteresis/... # Expect all tests pass
# Manual test: Run app, press E/D/T/G/Space/R/? and verify behavior
```

### Phase 2: Visual Phase Indicators (HIGH PRIORITY)

**Goal**: Show current write/read phase visually on the P-E plot and level indicator.

**Implementation**:
1. Add phase indicator to PEPlot widget:
   - Colored banner at top of plot showing phase name + color:
     - SATURATE: Red (#FF6B6B)
     - SETTLE: Orange (#FFA500)
     - HOLD: Green (#4CAF50)
     - READ: Blue (#2196F3)
   - Progress bar showing time remaining in current phase
2. Add target level highlight to LevelIndicator:
   - During WRITE: Target level flashes 3× with 100ms interval, then glows until complete
   - During READ: Soft pulse animation on current level
3. Add E-field trace indicator:
   - Small arrow on plot showing direction of E-field change

**Files to modify**: `pkg/gui/gui.go` (PEPlot renderer, LevelIndicator renderer)
**Estimated changes**: ~180 LOC additions

**Acceptance criteria**:
- [ ] Phase banner visible at top of P-E plot during Write/Read Demo mode
- [ ] Banner color matches phase (SATURATE=red, SETTLE=orange, HOLD=green, READ=blue)
- [ ] Target level flashes exactly 3 times with 100ms intervals on write start
- [ ] E-field direction arrow visible on plot

### Phase 3: Code Refactoring for Maintainability (HIGH PRIORITY)

**Goal**: Break monolithic gui.go into logical modules, targeting <600 lines for app.go.

**New file structure**:
```
pkg/gui/
├── app.go              # App struct, Run(), createUI(), lifecycle (~500 lines)
├── controls.go         # createControlsPanel(), slider/button handlers (~300 lines)
├── info.go             # createInfoPanel(), createSlidePanel(), createLogPanel() (~400 lines)
├── keyboard.go         # handleKeyPress(), showKeyboardHelp() (~120 lines)
├── simulation.go       # simulationLoop(), waveform generation logic (~500 lines)
├── theme.go            # feCIMTheme, colors, fixedMinWidthLayout (~160 lines)
├── widgets/
│   ├── peplot.go       # PEPlot widget + peplotRenderer (~370 lines)
│   ├── level.go        # LevelIndicator widget + levelRenderer (~250 lines)
│   ├── cell.go         # CellVisualizer widget + cellRenderer (~230 lines)
│   ├── mode.go         # ModeIndicator widget + modeRenderer (~140 lines)
│   └── phase.go        # NEW: PhaseIndicator widget (~100 lines)
├── embedded.go         # (existing) BuildContent, Start, Stop
└── labench.go          # (existing) LabBench struct - kept for reference/future Vulkan use
```

**Extraction order** (each step verified with tests):

| Step | Extract | From Lines | To File | Test Command |
|------|---------|------------|---------|--------------|
| 1 | PEPlot | 1746-2113 | widgets/peplot.go | `go test ./module1-hysteresis/...` |
| 2 | LevelIndicator | 2119-2365 | widgets/level.go | `go test ./module1-hysteresis/...` |
| 3 | CellVisualizer | 2372-2595 | widgets/cell.go | `go test ./module1-hysteresis/...` |
| 4 | ModeIndicator | 2602-2738 | widgets/mode.go | `go test ./module1-hysteresis/...` |
| 5 | Theme + colors | 36-47, 2739-2863 | theme.go | `go test ./module1-hysteresis/...` |
| 6 | Controls panel | 506-797 | controls.go | `go test ./module1-hysteresis/...` |
| 7 | Info panels | 815-877, 879-1057 | info.go | `go test ./module1-hysteresis/...` |
| 8 | Simulation loop | 1081-1509 | simulation.go | `go test ./module1-hysteresis/...` |

**After refactoring**: app.go should contain ~500 lines (App struct, NewApp, Run, createUI, updateUI).

**Estimated effort**: ~0 net LOC change (reorganization), but significant refactoring work.

### Phase 4: Interactive Mode Enhancements (MEDIUM PRIORITY)

**Goal**: Make Interactive mode feel more responsive and educational.

**Implementation**:
1. Real-time E-field value display:
   - E-field label updates at 60fps during phase transitions (matching simulation tick)
   - Format: "E: 1.23 → 2.00 MV/cm" showing current and target
2. Animated trace on plot:
   - Trail color changes to highlight the active segment being traced
   - Brighter/thicker line for last 50 points
3. Click feedback on level indicator:
   - Visual ripple animation on click (expanding circle, fades in 200ms)
   - Haptic feedback on mobile (if supported)
4. Add "ERASE" operation:
   - New radio option: WRITE | READ | ERASE
   - ERASE saturates to P=0 (middle level ~15)

**Files to modify**: `pkg/gui/widgets/level.go`, `pkg/gui/simulation.go`, `pkg/gui/controls.go`
**Estimated changes**: ~200 LOC additions

**Acceptance criteria**:
- [ ] E-field label shows "current → target" format during transitions
- [ ] E-field label updates every 16ms (60fps) during animation
- [ ] Click on level indicator shows ripple effect (200ms duration)
- [ ] ERASE radio button available in Interactive mode
- [ ] ERASE operation drives level to 15 (±1)

### Phase 5: Temperature Effects Visualization (MEDIUM PRIORITY)

**Goal**: Show how temperature affects ferroelectric behavior visually.

**Implementation**:
1. Add Landau coefficient display in info panel:
   - α(T) value with color coding:
     - Negative (ferroelectric): Green text
     - Positive (paraelectric): Red text
   - Format: "α(T): -2.5×10⁸ (Ferro)"
2. Add T/Tc ratio visual bar:
   - Horizontal progress bar showing T/Tc percentage
   - Color gradient: Green (<70%) → Yellow (70-90%) → Red (>90%)
3. Add Curie warning:
   - When T > 0.9×Tc, show warning: "⚠ Approaching Tc - ferroelectric properties degrading"
   - Flash warning in status bar

**Files to modify**: `pkg/gui/info.go`
**Estimated changes**: ~100 LOC additions

**Acceptance criteria**:
- [ ] α(T) displayed in info panel with sign-based color (green=negative, red=positive)
- [ ] T/Tc ratio bar visible below temperature slider
- [ ] Warning appears when T > 0.9×Tc (630K for HZO with Tc=700K)

### Phase 6: Bug Fixes and Polish (LOW PRIORITY)

**Goal**: Fix remaining issues and polish UX.

**Implementation**:
1. **Debug log cap**: Limit `wrdDebugLog.Cycles` to last 100 entries
   ```go
   if len(a.wrdDebugLog.Cycles) > 100 {
       a.wrdDebugLog.Cycles = a.wrdDebugLog.Cycles[len(a.wrdDebugLog.Cycles)-100:]
   }
   ```
2. **Cell volume from material**: Replace hardcoded `2e-22` with:
   ```go
   cellVolume := a.material.Area * a.material.Thickness
   ```
3. **Defensive nil checks**: Add guards before accessing material properties
4. **Consistent mutex patterns**: Audit all state reads to use RLock consistently
5. **Loading indicator**: Show spinner during material switch while Preisach reinitializes

**Files to modify**: `pkg/gui/simulation.go`, `pkg/gui/app.go`
**Estimated changes**: ~50 LOC changes

**Acceptance criteria**:
- [ ] Debug log never exceeds 100 cycles (test by running 150+ cycles)
- [ ] Energy calculation uses `material.Area * material.Thickness`
- [ ] No panic when switching materials rapidly

---

## Testing Strategy

**Before each phase**:
```bash
go test ./module1-hysteresis/... -v  # Baseline: all tests pass
```

**After each phase**:
```bash
go test ./module1-hysteresis/... -v  # Regression: same test count, all pass
go test ./... -v                      # Full suite: no cross-module breaks
```

**Manual testing checklist** (after Phase 3):
- [ ] App launches without errors
- [ ] All 7 waveform modes work
- [ ] Material switching works
- [ ] Pause/Resume works
- [ ] Reset clears state
- [ ] Interactive mode click-to-level works

---

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Refactoring breaks existing features | Medium | Run full test suite after EACH extraction step |
| Fyne keyboard handling platform differences | Low | Test on Linux first (dev environment), document any platform quirks |
| Performance regression from animations | Low | Profile if frame time exceeds 16ms; animations use simple transforms |
| Scope creep to Vulkan renderer | Medium | Explicitly marked out of scope; labench.go preserved but not enhanced |
| Widget extraction breaks imports | Medium | Use `pkg/gui/widgets` subpackage, update all imports in same commit |

---

## Implementation Order

1. **Phase 3** (Refactoring) - Do first to make other changes easier to review
2. **Phase 1** (Keyboard) - Quick win, high impact, builds on clean structure
3. **Phase 2** (Visual indicators) - High user value
4. **Phase 6** (Bug fixes) - Can be done incrementally during other phases
5. **Phase 4** (Interactive enhancements) - Nice to have
6. **Phase 5** (Temperature viz) - Educational value, lower priority

---

## Out of Scope

- Vulkan renderer implementation (requires significant GPU programming expertise)
- TUI mode improvements (separate module)
- New waveform types beyond ERASE
- Multi-cell array visualization
- Performance benchmarking suite
- Bridging GLFW keys to Fyne (implementing Fyne-native instead)

---

## Dr. Shin's Expert Recommendations

As a ferroelectric memory expert, I recommend prioritizing:

1. **Keyboard shortcuts** - Essential for rapid experimentation during research
2. **Phase visualization** - Critical for understanding the write mechanism
3. **Temperature effects** - HZO's high Tc is a key differentiator from other ferroelectrics
4. **Clean code structure** - Enables future graduate students to extend the tool

The current physics model (MayergoyzPreisach with 30 hysterons) is scientifically sound. Focus improvements on the UI/UX layer to make the excellent physics more accessible.

---

## PLAN_READY: .omc/plans/hysteresis-ui-improvement.md
