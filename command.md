/ralph-loop:/ralph-loop "# Fyne GUI Deep Refactor & Bug Fix Session

## Problem Statement
The app has severe layout bugs where elements grow uncontrollably when clicking buttons. Root causes identified:
1. Custom widget renderers calling `r.Refresh()` from `Layout()` - causes layout cycles
2. Renderers using `widget.Size()` in Refresh which returns the (possibly huge) allocated size instead of MinSize
3. VBox/HBox containers distributing extra space to children instead of using spacers
4. Duplicated buggy renderer code across 6+ modules with same anti-patterns

## Your Mission
You are a Fyne GUI expert. Systematically refactor and fix all GUI code:

### 1. AUDIT ALL CUSTOM RENDERERS
Search for pattern: `func (r *.*Renderer) Layout`
Every renderer must follow this pattern:
```go
func (r *myRenderer) Layout(size fyne.Size) {
    // Position/resize objects based on SIZE parameter, NOT widget.Size()
    r.content.Resize(size)
    r.content.Move(fyne.NewPos(0, 0))
}

func (r *myRenderer) Refresh() {
    // Regenerate objects, use MinSize for drawing constraints
    // NEVER call Layout from here
}
```

### 2. CREATE SHARED BASE WIDGETS
Create `shared/widgets/base_renderer.go` with:
- `ConstrainedWidget` base type that enforces MinSize limits
- Helper functions for common patterns
- Debug logging for layout events

### 3. ADD VERBOSE DEBUG MODE
Add to each module's GUI code:
```go
var debugLayout = os.Getenv("FYNE_DEBUG_LAYOUT") != ""

func debugLog(format string, args ...interface{}) {
    if debugLayout {
        fmt.Printf("[LAYOUT] "+format+"\n", args...)
    }
}
```
Log: MinSize calls, Layout calls with size, Refresh calls

### 4. FIX CONTAINER PATTERNS
Replace:
```go
container.NewVBox(item1, item2, item3)  // BAD - stretches items
```
With:
```go
container.NewVBox(item1, item2, item3, layout.NewSpacer())  // GOOD - spacer absorbs extra
```

### 5. FILES TO CHECK
Priority order:
- module1-hysteresis/pkg/gui/gui.go (peplotRenderer, cellRenderer, modeRenderer, levelRenderer)
- module2-crossbar/pkg/gui/liveslide.go (modeIndicatorBoxRenderer)
- module2-crossbar/pkg/gui/widgets.go (ALL custom widgets)
- module2-crossbar/pkg/gui/heatmap.go
- module3-mnist/pkg/gui/liveslide.go
- module5-comparison/pkg/gui/liveslide.go
- cmd/fecim-visualizer/launcher.go (demoCardRenderer)

### 6. TEST METHODOLOGY
After each fix:
```bash
FYNE_DEBUG_LAYOUT=1 go build ./cmd/fecim-visualizer && ./fecim-visualizer
```
Click every button, resize window, switch tabs. Elements must NOT grow.

### 7. SUCCESS CRITERIA
- No element grows when clicking buttons
- Window resizes smoothly without layout jumps
- Custom widgets stay at their MinSize
- Debug output shows stable layout calls (not infinite loops)

## Run Command
```
FYNE_DEBUG_LAYOUT=1 ./fecim-visualizer 2>&1 | grep -E "\[LAYOUT\]|panic|fatal"
```

---
Keep working until ALL layout bugs are fixed. Be systematic. Test after each change.
" --max-iterations 20000 --completion-promise "DONE HYPER ANALYSIS"

