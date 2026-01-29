# Implementation Plan: Shared Code Refactoring

## Requirements Restatement

Consolidate duplicated code patterns across 7 modules into `shared/` package:
1. Path discovery functions (8+ duplicates of `findDataDir()`)
2. Status update helpers with thread-safe `fyne.Do()` wrapping
3. UI update helpers for common widget operations
4. Demo controller for auto-demo loop patterns
5. Array initialization utilities
6. Embedded app base struct (optional composition)

## Current State Analysis

### Already Shared (No Action Needed)
- `shared/widgets/embedded_app.go` - EmbeddedApp interface
- `shared/theme/` - Color palette and theming
- `shared/logging/` - Logging infrastructure
- `shared/utils/font.go` - Bitmap font rendering
- `shared/utils/recover.go` - SafeGo panic recovery
- `shared/utils/drawing.go` - Canvas primitives

### Validated Duplications
| Pattern | Occurrences | Files |
|---------|-------------|-------|
| `findDataDir()` | 8x | module3/gui/app.go, train scripts |
| `findDocsPath()` | 1x | module7/gui/embedded.go |
| `updateStatus()` | 5x | M2, M3, M4, M5, M6 |
| Auto-demo loops | 2x | M2 (context), M3 (mutex) |
| Matrix init | 3x | M2, M4, crossbar pkg |

---

## Implementation Phases

### Phase 1: Path Discovery Helpers
**File:** `shared/utils/path_discovery.go`

Create generic path discovery utilities:
- `FindDirectory(name string, markers ...string) string`
- `FindFile(name string) string`
- `FindDataDir(moduleName string) string` - convenience wrapper

**Changes to modules:**
- `module3-mnist/pkg/gui/app.go` - Replace `findDataDir()` with `utils.FindDataDir("module3-mnist")`
- `module3-mnist/cmd/train-ptq/main.go` - Same
- `module3-mnist/cmd/train-network/main.go` - Same
- `module3-mnist/cmd/train-single-layer/main.go` - Same
- `module3-mnist/train_14levels.go` - Same
- `module7-docs/pkg/gui/embedded.go` - Replace `findDocsPath()` with `utils.FindDirectory("docs")`

**Tests:** `shared/utils/path_discovery_test.go`

---

### Phase 2: Status Update Helper
**File:** `shared/widgets/status_helper.go`

Create thread-safe status updater with cache prevention:
```go
type StatusBar struct {
    label      *widget.Label
    lastText   string
    prefix     string  // e.g., "Status: "
    history    []string
    maxHistory int
}

func NewStatusBar(prefix string) *StatusBar
func (s *StatusBar) GetLabel() *widget.Label
func (s *StatusBar) Update(msg string)      // Thread-safe with fyne.Do()
func (s *StatusBar) Clear()
func (s *StatusBar) GetHistory() []string
```

**Changes to modules:**
- `module2-crossbar/pkg/gui/app.go` - Use `widgets.StatusBar`
- `module3-mnist/pkg/gui/dualmode.go` - Use `widgets.StatusBar`
- `module4-circuits/pkg/gui/app.go` - Use `widgets.StatusBar`
- `module5-comparison/pkg/gui/app.go` - Use `widgets.StatusBar`

**Tests:** `shared/widgets/status_helper_test.go`

---

### Phase 3: UI Update Helpers
**File:** `shared/widgets/ui_helpers.go`

Create thread-safe UI update functions:
```go
func SafeUpdateLabel(label *widget.Label, text string)
func SafeUpdateProgress(progress *widget.ProgressBar, value float64)
func SafeRefresh(obj fyne.CanvasObject)
func SafeShowHide(obj fyne.CanvasObject, show bool)
func SafeEnableDisable(w fyne.Disableable, enable bool)
func SafeSetValue(entry *widget.Entry, value string)
```

**Impact:** Modules can replace scattered `fyne.Do(func() { ... })` patterns with these helpers.

**Tests:** `shared/widgets/ui_helpers_test.go`

---

### Phase 4: Demo Controller
**File:** `shared/widgets/demo_controller.go`

Create reusable demo automation:
```go
type DemoStep struct {
    Name     string
    Duration time.Duration
    Action   func()
}

type DemoController struct {
    steps   []DemoStep
    running bool
    ctx     context.Context
    cancel  context.CancelFunc
    mu      sync.Mutex
}

func NewDemoController(steps []DemoStep) *DemoController
func (d *DemoController) Start()
func (d *DemoController) Stop()
func (d *DemoController) IsRunning() bool
func (d *DemoController) WaitOrStop(duration time.Duration) bool
```

**Changes to modules:**
- `module2-crossbar/pkg/gui/animation.go` - Refactor auto-demo to use controller
- `module3-mnist/pkg/gui/dualmode_demo.go` - Refactor quick demo to use controller

**Tests:** `shared/widgets/demo_controller_test.go`

---

### Phase 5: Array Initialization Helpers
**File:** `shared/utils/array_init.go`

Create matrix utilities:
```go
func InitMatrix2D[T any](rows, cols int, value T) [][]T
func InitMatrix2DFunc[T any](rows, cols int, fn func(i, j int) T) [][]T
func InitMatrixRandom(rows, cols, max int) [][]int
```

**Tests:** `shared/utils/array_init_test.go`

---

### Phase 6: Embedded App Base (Optional)
**File:** `shared/widgets/embedded_base.go`

Create optional composition helper:
```go
type EmbeddedAppBase struct {
    FyneApp   fyne.App
    Window    fyne.Window
    Status    *StatusBar
    IsRunning bool
}

func (b *EmbeddedAppBase) DefaultStart()
func (b *EmbeddedAppBase) DefaultStop()
```

**Note:** This is optional composition, not mandatory inheritance. Modules can embed this struct if they want the boilerplate.

---

## File Creation Summary

| File | Type | Lines |
|------|------|-------|
| `shared/utils/path_discovery.go` | New | ~70 |
| `shared/utils/path_discovery_test.go` | New | ~50 |
| `shared/widgets/status_helper.go` | New | ~60 |
| `shared/widgets/status_helper_test.go` | New | ~40 |
| `shared/widgets/ui_helpers.go` | New | ~80 |
| `shared/widgets/ui_helpers_test.go` | New | ~60 |
| `shared/widgets/demo_controller.go` | New | ~100 |
| `shared/widgets/demo_controller_test.go` | New | ~80 |
| `shared/utils/array_init.go` | New | ~50 |
| `shared/utils/array_init_test.go` | New | ~40 |
| `shared/widgets/embedded_base.go` | New | ~40 |

**Total new shared code:** ~670 lines
**Estimated lines removed from modules:** ~890 lines

---

## Module Changes Summary

| Module | Files Modified | Changes |
|--------|---------------|---------|
| module3-mnist | 5 files | Replace findDataDir(), use StatusBar |
| module2-crossbar | 2 files | Use StatusBar, DemoController |
| module4-circuits | 1 file | Use StatusBar |
| module5-comparison | 1 file | Use StatusBar |
| module7-docs | 1 file | Replace findDocsPath() |

---

## Risks & Mitigations

| Risk | Severity | Mitigation |
|------|----------|------------|
| Breaking existing behavior | Medium | Run `go test ./...` after each phase |
| Fyne thread safety issues | Low | StatusBar uses `fyne.Do()` internally |
| Import cycles | Low | `shared/utils` and `shared/widgets` are leaf packages |
| Generics compatibility | Low | Go 1.21+ already used (check go.mod) |

---

## Execution Order

1. **Phase 1 (Path Discovery)** - Standalone, no dependencies
2. **Phase 2 (Status Helper)** - Standalone, enables Phase 4
3. **Phase 3 (UI Helpers)** - Standalone, utility functions
4. **Phase 4 (Demo Controller)** - Depends on Phase 2 for status integration
5. **Phase 5 (Array Init)** - Standalone, optional
6. **Phase 6 (Embedded Base)** - Depends on Phase 2, optional

---

## Validation Checklist

After each phase:
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes (117+ tests)
- [ ] `go vet ./...` clean
- [ ] Manual test: Run `./fecim-lattice-tools` and verify affected modules

Final validation:
- [ ] All 7 modules load correctly in unified app
- [ ] Status updates display correctly
- [ ] Auto-demo features work in M2 and M3
- [ ] No race conditions (`go test -race ./...`)

---

## WAITING FOR CONFIRMATION

Proceed with this plan? (yes/no/modify)

If you want changes:
- "modify: focus only on Phase 1-3"
- "skip Phase 6"
- "start with Phase 2 instead"
