# FeCIM Full gogpu/ui Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Port all 6 remaining placeholder modules from Fyne to gogpu/ui zero-CGO shell with Material 3 design system, sidebar navigation, custom drawing primitives (PlotWidget, HeatmapWidget), and education/research/design feature layers across every module.

**Architecture:** Each module follows: `shared/viewmodel/{module}/` (UI-neutral viewmodel) → `cmd/fecim-lattice-tools-next/{module}_view.go` (gogpu/ui adapter) → wired into `buildRoot` switch. Design system primitives built as shared gogpu/ui custom widgets using `gg.Context` canvas drawing. Shell restructured to sidebar + content layout.

**Tech Stack:** Go 1.25, gogpu/ui v0.1.13, gogpu/gg v0.43.2, gogpu/gogpu v0.29.4, Material 3 theme

**Design spec:** `docs/superpowers/specs/2026-05-03-fecim-gogpu-full-migration-design.md`

**Migration order (design-first):**
1. Module 1 Hysteresis
2. Module 2 Crossbar
3. Module 6 EDA
4. Module 4 Circuits
5. Module 3 MNIST
6. Module 7 Docs

---

## Phase 1: Shell Navigation & Design System Foundation

### Task 1.1: Design System — Colors & Token Constants

**Files:**
- Create: `cmd/fecim-lattice-tools-next/design/tokens.go`
- Create: `cmd/fecim-lattice-tools-next/design/tokens_test.go`

- [ ] **Step 1: Write failing test for design tokens**

```go
// cmd/fecim-lattice-tools-next/design/tokens_test.go
//go:build !cgo

package design

import (
	"testing"

	"github.com/gogpu/ui/widget"
)

func TestDesignTokens_Colors(t *testing.T) {
	if Primary != widget.Hex(0x2F5D50) {
		t.Errorf("Primary = %v, want 0x2F5D50", Primary)
	}
	if Surface != widget.Hex(0xF4F5F3) {
		t.Errorf("Surface = %v, want 0xF4F5F3", Surface)
	}
	if OnSurface != widget.Hex(0x1A1C1A) {
		t.Errorf("OnSurface = %v, want 0x1A1C1A", OnSurface)
	}
}

func TestDesignTokens_Spacing(t *testing.T) {
	if SidebarWidth != 240 {
		t.Errorf("SidebarWidth = %v, want 240", SidebarWidth)
	}
	if ContentPad != 24 {
		t.Errorf("ContentPad = %v, want 24", ContentPad)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/design/...
```
Expected: FAIL — package does not exist

- [ ] **Step 3: Implement design tokens**

```go
// cmd/fecim-lattice-tools-next/design/tokens.go
//go:build !cgo

package design

import "github.com/gogpu/ui/widget"

// Material 3 color tokens for FeCIM Lattice Tools.
// Seed: deep green (#2F5D50) — physics/engineering/education.
var (
	Primary   = widget.Hex(0x2F5D50)
	PrimaryDark   = widget.Hex(0x1F463C)
	PrimaryLight  = widget.Hex(0x6F9C8D)
	Surface       = widget.Hex(0xF4F5F3)
	SurfaceContainer = widget.Hex(0xE8EBE7)
	OnSurface     = widget.Hex(0x1A1C1A)
	OnSurfaceVariant = widget.Hex(0x444744)
	Secondary     = widget.Hex(0x58685E)
	Error         = widget.Hex(0xBA1A1A)
)

// Layout spacing tokens.
const (
	SidebarWidth = 240
	ContentPad   = 24
	CardGap      = 14
	SectionGap   = 10
	TopBarHeight = 52
)
```

- [ ] **Step 4: Run test to verify pass**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/design/...
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/design/
git commit -m "feat(next-shell): add Material 3 design tokens package"
```

---

### Task 1.2: Sidebar Navigation Widget

**Files:**
- Create: `cmd/fecim-lattice-tools-next/sidebar.go`
- Create: `cmd/fecim-lattice-tools-next/sidebar_test.go`
- Modify: `cmd/fecim-lattice-tools-next/main.go` (integrate sidebar)

- [ ] **Step 1: Write failing test for sidebar**

```go
// cmd/fecim-lattice-tools-next/sidebar_test.go
//go:build !cgo

package main

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestSidebarBuildsForAllModules(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	widget := buildSidebar(descriptors, 0)
	if widget == nil {
		t.Fatal("buildSidebar returned nil")
	}
}

func TestSidebarActiveIndex(t *testing.T) {
	descriptors := viewmodel.KnownDescriptors()
	// Active index 2 should still produce a valid sidebar
	widget := buildSidebar(descriptors, 2)
	if widget == nil {
		t.Fatal("buildSidebar with activeIndex=2 returned nil")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/... -run TestSidebar
```
Expected: FAIL — `buildSidebar` not defined

- [ ] **Step 3: Implement sidebar**

```go
// cmd/fecim-lattice-tools-next/sidebar.go
//go:build !cgo

package main

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildSidebar(descriptors []viewmodel.ModuleDescriptor, activeIndex int) widget.Widget {
	items := []widget.Widget{
		primitives.Text("FeCIM Modules").FontSize(14).Bold(),
	}

	for i, d := range descriptors {
		statusColor := "#888888"
		if d.Status == viewmodel.StatusFunctional {
			statusColor = "#2F5D50"
		}
		highlight := ""
		if i == activeIndex {
			highlight = " →"
		}
		items = append(items, primitives.Box(
			primitives.Text(highlight+" "+d.Title).FontSize(13),
			primitives.Text(string(d.ID)+" · "+d.Status).FontSize(10),
		).
			Padding(8).
			Gap(2),
		)
	}

	return primitives.Box(items...).
		Padding(16).
		Gap(8)
}

func buildSidebarMaterial(descriptors []viewmodel.ModuleDescriptor, activeIndex int, theme *material3.Theme) widget.Widget {
	items := []widget.Widget{
		primitives.Text("FeCIM Modules").FontSize(14).Bold(),
	}

	for i, d := range descriptors {
		bgColor := theme.Colors.Surface
		textColor := theme.Colors.OnSurface
		if i == activeIndex {
			bgColor = theme.Colors.Primary
			textColor = theme.Colors.OnPrimary
		}
		statusColor := theme.Colors.OnSurfaceVariant
		if d.Status == viewmodel.StatusFunctional {
			statusColor = theme.Colors.Primary
		}

		items = append(items, primitives.Box(
			primitives.Text(d.Title).FontSize(13).Color(textColor),
			primitives.Text(string(d.ID)+" · "+d.Status).FontSize(10).Color(statusColor),
		).
			Padding(10).
			Gap(2).
			Background(bgColor),
		)
	}

	return primitives.Box(items...).
		Padding(16).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}
```

- [ ] **Step 4: Run test to verify pass**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/... -run TestSidebar
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/sidebar.go cmd/fecim-lattice-tools-next/sidebar_test.go
git commit -m "feat(next-shell): add sidebar navigation widget"
```

---

### Task 1.3: Restructure Shell to Sidebar + Content Layout

**Files:**
- Modify: `cmd/fecim-lattice-tools-next/main.go`

- [ ] **Step 1: Write failing test for layout structure**

```go
// cmd/fecim-lattice-tools-next/root_test.go (append to existing)
//go:build !cgo

func TestBuildRootHasSidebarAndContent(t *testing.T) {
	spec := DefaultAppSpec()
	ports := BuildPlaceholderPorts()
	theme := material3.New(widget.Hex(0x2F5D50))

	root := buildRoot(spec, ports, theme)
	if root == nil {
		t.Fatal("buildRoot returned nil")
	}
	// buildRoot should return a horizontal layout (sidebar | content)
}
```

- [ ] **Step 2: Refactor buildRoot to sidebar + content layout**

Replace the current vertical `buildRoot` with a sidebar + content structure:

```go
// cmd/fecim-lattice-tools-next/main.go (modify buildRoot)
func buildRoot(spec AppSpec, ports []viewmodel.ModulePort, theme *material3.Theme) widget.Widget {
	descriptors := make([]viewmodel.ModuleDescriptor, len(ports))
	for i, p := range ports {
		descriptors[i] = p.Descriptor()
	}

	sidebar := buildSidebarMaterial(descriptors, 0, theme)

	children := []widget.Widget{
		primitives.Text(spec.Title).FontSize(20).Bold(),
		primitives.Text("Simulation-first FeCIM design workspace").FontSize(13),
	}

	for _, port := range ports {
		snapshot := port.Snapshot()
		switch snapshot.Descriptor.ID {
		case viewmodel.ModuleComparison:
			children = append(children, buildComparisonView(snapshot, theme))
		default:
			children = append(children, moduleCardEnhanced(snapshot, theme))
		}
	}

	content := primitives.Box(children...).
		Padding(24).
		Gap(14)

	return primitives.Box(
		sidebar,
		content,
	).
		Gap(0)
}

func moduleCardEnhanced(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	descriptor := snapshot.Descriptor
	body := descriptor.Description
	if len(snapshot.Sections) > 0 && snapshot.Sections[0].Body != "" {
		body = body + "\n" + snapshot.Sections[0].Body
	}

	statusBadge := "PLACEHOLDER"
	badgeColor := theme.Colors.OnSurfaceVariant
	if descriptor.Status == viewmodel.StatusFunctional {
		statusBadge = "FUNCTIONAL"
		badgeColor = theme.Colors.Primary
	}

	return primitives.Box(
		primitives.Box(
			primitives.Text(descriptor.Title).FontSize(18).Bold(),
			primitives.Text(statusBadge).FontSize(11).Color(badgeColor),
		).Gap(8),
		primitives.Text(body).FontSize(14),
	).
		Padding(16).
		Gap(8).
		Background(theme.Colors.SurfaceContainer)
}
```

- [ ] **Step 3: Run build verification**

```bash
CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/...
```
Expected: build succeeds, tests pass

- [ ] **Step 4: Commit**

```bash
git add cmd/fecim-lattice-tools-next/main.go cmd/fecim-lattice-tools-next/root_test.go
git commit -m "feat(next-shell): restructure to sidebar + content layout"
```

---

## Phase 2: Module 1 — Hysteresis (Material Physics)

### Task 2.1: Hysteresis Viewmodel — Types & State

**Files:**
- Create: `shared/viewmodel/hysteresis/state.go`
- Create: `shared/viewmodel/hysteresis/events.go`
- Create: `shared/viewmodel/hysteresis/viewmodel.go`
- Create: `shared/viewmodel/hysteresis/snapshot.go`
- Create: `shared/viewmodel/hysteresis/viewmodel_test.go`

- [ ] **Step 1: Write failing test for hysteresis viewmodel**

```go
// shared/viewmodel/hysteresis/viewmodel_test.go
package hysteresis

import (
	"testing"

	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New()
	d := m.Descriptor()
	if d.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Descriptor().ID = %v, want %v", d.ID, viewmodel.ModuleHysteresis)
	}
	if d.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor().Status = %v, want %v", d.Status, viewmodel.StatusFunctional)
	}
}

func TestSnapshotContainsMetrics(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Metrics) == 0 {
		t.Error("Snapshot().Metrics is empty, expected material metrics")
	}
	if s.Descriptor.ID != viewmodel.ModuleHysteresis {
		t.Errorf("Snapshot().Descriptor.ID = %v", s.Descriptor.ID)
	}
}

func TestSnapshotHasSections(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Sections) == 0 {
		t.Error("Snapshot().Sections is empty")
	}
}

func TestApplyActionUnsupported(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "run_simulation", Kind: viewmodel.ActionCommand})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want %v", err, viewmodel.ErrUnsupportedAction)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
go test ./shared/viewmodel/hysteresis/...
```
Expected: FAIL — package doesn't exist

- [ ] **Step 3: Implement state types**

```go
// shared/viewmodel/hysteresis/state.go
package hysteresis

import "fecim-lattice-tools/shared/physics"

// HysteresisState holds the complete UI-neutral state for the hysteresis module.
type HysteresisState struct {
	SelectedMaterial string            `json:"selected_material"`
	Materials        []*physics.HZOMaterial `json:"materials"`
	FieldRange       FieldRange        `json:"field_range"`
	LoopPoints       []LoopPoint       `json:"loop_points"`
	Waveform         string            `json:"waveform"` // "sine", "triangle", "sawtooth"
	IsRunning        bool              `json:"is_running"`
}

type FieldRange struct {
	MinField float64 `json:"min_field"` // kV/cm
	MaxField float64 `json:"max_field"` // kV/cm
}

type LoopPoint struct {
	Field        float64 `json:"field"`         // kV/cm
	Polarization float64 `json:"polarization"`  // µC/cm²
}
```

```go
// shared/viewmodel/hysteresis/events.go
package hysteresis

// Event kinds for the hysteresis module.
const (
	EventSelectMaterial  = "select_material"
	EventSetFieldRange   = "set_field_range"
	EventSetWaveform     = "set_waveform"
	EventToggleSimulation = "toggle_simulation"
	EventExportCSV       = "export_csv"
)
```

- [ ] **Step 4: Implement viewmodel**

```go
// shared/viewmodel/hysteresis/viewmodel.go
package hysteresis

import (
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

// Module implements viewmodel.ModulePort for the hysteresis module.
type Module struct {
	state HysteresisState
}

// New creates a new hysteresis Module with default state.
func New() *Module {
	materials := physics.AllMaterials()
	defaultMat := "HZO Default"
	if len(materials) > 0 {
		defaultMat = materials[0].Name
	}
	return &Module{
		state: HysteresisState{
			SelectedMaterial: defaultMat,
			Materials:        materials,
			FieldRange: FieldRange{
				MinField: -3000,
				MaxField: 3000,
			},
			Waveform: "sine",
		},
	}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleHysteresis,
		Title:       "FeCIM Hysteresis Simulation",
		Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
		Status:      viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot {
	return buildSnapshot(m.state)
}

func (m *Module) ApplyAction(action viewmodel.Action) error {
	return viewmodel.ErrUnsupportedAction
}

func (m *Module) Start() {}
func (m *Module) Stop()  {}
```

- [ ] **Step 5: Implement snapshot builder**

```go
// shared/viewmodel/hysteresis/snapshot.go
package hysteresis

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state HysteresisState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "material", Label: "Material", Value: state.SelectedMaterial},
		{ID: "field_min", Label: "Min Field", Value: fmt.Sprintf("%.0f kV/cm", state.FieldRange.MinField)},
		{ID: "field_max", Label: "Max Field", Value: fmt.Sprintf("%.0f kV/cm", state.FieldRange.MaxField)},
		{ID: "waveform", Label: "Waveform", Value: state.Waveform},
	}

	// Build material parameter sections
	sections := []viewmodel.Section{}
	for _, mat := range state.Materials {
		if mat == nil {
			continue
		}
		sections = append(sections, viewmodel.Section{
			ID:    "material_" + mat.Name,
			Title: mat.Name,
			Body:  fmt.Sprintf("Pr=%.2f µC/cm²  Ec=%.0f kV/cm  Thickness=%.1f nm  α=%.4e  β=%.4e  γ=%.4e",
				mat.Pr, mat.Ec, mat.Thickness*1e9, mat.Alpha, mat.Beta, mat.Gamma),
		})
	}

	actions := []viewmodel.Action{
		{ID: EventSelectMaterial, Label: "Change Material", Kind: viewmodel.ActionSelect},
		{ID: EventSetFieldRange, Label: "Set Field Range", Kind: viewmodel.ActionCommand},
		{ID: EventToggleSimulation, Label: "Run/Pause", Kind: viewmodel.ActionToggle},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:          viewmodel.ModuleHysteresis,
			Title:       "FeCIM Hysteresis Simulation",
			Description: "P-E curves, Preisach model, Landau-Khalatnikov solver, and material presets.",
			Status:      viewmodel.StatusFunctional,
		},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
	}
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./shared/viewmodel/hysteresis/...
```
Expected: PASS

- [ ] **Step 7: Verify UI boundary (no Fyne/gogpu/ui imports)**

```bash
grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/hysteresis/ && echo "VIOLATION" || echo "CLEAN"
```
Expected: CLEAN

- [ ] **Step 8: Commit**

```bash
git add shared/viewmodel/hysteresis/
git commit -m "feat(viewmodel/hysteresis): add hysteresis module viewmodel with material state and snapshot"
```

---

### Task 2.2: Hysteresis gogpu/ui Adapter

**Files:**
- Create: `cmd/fecim-lattice-tools-next/hysteresis_view.go`
- Create: `cmd/fecim-lattice-tools-next/hysteresis_view_test.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go` (inject hysteresis viewmodel)
- Modify: `cmd/fecim-lattice-tools-next/main.go` (add case to buildRoot)

- [ ] **Step 1: Write failing test for hysteresis adapter**

```go
// cmd/fecim-lattice-tools-next/hysteresis_view_test.go
//go:build !cgo

package main

import (
	"testing"

	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildHysteresisView(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))

	w := buildHysteresisView(snapshot, theme)
	if w == nil {
		t.Fatal("buildHysteresisView returned nil")
	}
}

func TestBuildHysteresisView_ContainsMaterialSections(t *testing.T) {
	vm := hysteresisvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))

	w := buildHysteresisView(snapshot, theme)
	if w == nil {
		t.Fatal("buildHysteresisView returned nil")
	}
	// Material sections should be present
	if len(snapshot.Sections) == 0 {
		t.Error("No material sections in snapshot")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/... -run TestBuildHysteresisView
```
Expected: FAIL — `buildHysteresisView` not defined

- [ ] **Step 3: Implement hysteresis gogpu/ui adapter**

```go
// cmd/fecim-lattice-tools-next/hysteresis_view.go
//go:build !cgo

package main

import (
	"fecim-lattice-tools/shared/viewmodel"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

// buildHysteresisView renders a hysteresis ModuleSnapshot into a gogpu/ui widget tree.
func buildHysteresisView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	// Metrics dashboard
	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(16).Bold(),
		).
			Padding(12).
			Gap(4).
			Background(theme.Colors.SurfaceContainer),
		)
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	// P-E Loop plot placeholder (will become interactive PlotWidget in future task)
	children = append(children, primitives.Box(
		primitives.Text("P-E Loop").FontSize(16).Bold(),
		primitives.Text("Plot widget coming in Phase 1.3 — currently rendering material data as cards.").FontSize(13),
	).
		Padding(20).
		Gap(8).
		Background(theme.Colors.SurfaceContainer),
	)

	// Material parameter sections
	for _, section := range snapshot.Sections {
		children = append(children, hysteresisMaterialCard(section, theme))
	}

	// Actions
	actionBoxes := []widget.Widget{}
	for _, action := range snapshot.Actions {
		actionBoxes = append(actionBoxes, primitives.Box(
			primitives.Text(action.Label).FontSize(13),
		).
			Padding(10).
			Background(theme.Colors.Primary),
		)
	}
	if len(actionBoxes) > 0 {
		children = append(children, primitives.Box(actionBoxes...).Gap(8))
	}

	return primitives.Box(children...).
		Padding(24).
		Gap(14).
		Background(theme.Colors.Surface)
}

func hysteresisMaterialCard(section viewmodel.Section, theme *material3.Theme) widget.Widget {
	return primitives.Box(
		primitives.Text(section.Title).FontSize(15).Bold(),
		primitives.Text(section.Body).FontSize(12),
	).
		Padding(12).
		Gap(4).
		Background(theme.Colors.SurfaceContainer)
}
```

- [ ] **Step 4: Wire hysteresis viewmodel into BuildPlaceholderPorts**

```go
// cmd/fecim-lattice-tools-next/appmodel.go — modify BuildPlaceholderPorts
import (
	"fecim-lattice-tools/shared/viewmodel"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
	hysteresisvm "fecim-lattice-tools/shared/viewmodel/hysteresis"
)

func BuildPlaceholderPorts() []viewmodel.ModulePort {
	descriptors := viewmodel.KnownDescriptors()
	ports := make([]viewmodel.ModulePort, 0, len(descriptors))
	for _, descriptor := range descriptors {
		switch descriptor.ID {
		case viewmodel.ModuleComparison:
			ports = append(ports, comparisonvm.New())
		case viewmodel.ModuleHysteresis:
			ports = append(ports, hysteresisvm.New())
		default:
			ports = append(ports, viewmodel.NewStaticModule(descriptor, []viewmodel.Section{
				{
					ID:    "migration-status",
					Title: "Migration Status",
					Body:  "This module is represented by a UI-neutral placeholder while the gogpu/ui shell reaches parity with the current Fyne implementation.",
				},
			}))
		}
	}
	return ports
}
```

- [ ] **Step 5: Add hysteresis case to buildRoot switch**

```go
// cmd/fecim-lattice-tools-next/main.go — in buildRoot, add case:
case viewmodel.ModuleHysteresis:
    children = append(children, buildHysteresisView(snapshot, theme))
```

- [ ] **Step 6: Run tests**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/...
make test-next-ui
```
Expected: PASS

- [ ] **Step 7: Run full validation**

```bash
go test ./shared/viewmodel/...
CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
```
Expected: build succeeds, tests pass

- [ ] **Step 8: Commit**

```bash
git add cmd/fecim-lattice-tools-next/hysteresis_view.go \
       cmd/fecim-lattice-tools-next/hysteresis_view_test.go \
       cmd/fecim-lattice-tools-next/appmodel.go \
       cmd/fecim-lattice-tools-next/main.go
git commit -m "feat(next-shell): add hysteresis module gogpu/ui adapter with viewmodel integration

Wires shared/viewmodel/hysteresis into the gogpu/ui shell with material parameter display,
metric dashboard, and action buttons. Module 1 now renders as StatusFunctional."
```

---

### Task 2.3: Custom PlotWidget — P-E Loop Canvas Drawing

**Files:**
- Create: `cmd/fecim-lattice-tools-next/design/plot.go`
- Create: `cmd/fecim-lattice-tools-next/design/plot_test.go`

- [ ] **Step 1: Write failing test for PlotWidget data model**

```go
// cmd/fecim-lattice-tools-next/design/plot_test.go
//go:build !cgo

package design

import (
	"testing"
)

func TestPlotData_NewPlotData(t *testing.T) {
	pd := NewPlotData("P-E Loop", "Field (kV/cm)", "Polarization (µC/cm²)")
	if pd.Title != "P-E Loop" {
		t.Errorf("Title = %v, want P-E Loop", pd.Title)
	}
	if pd.XLabel != "Field (kV/cm)" {
		t.Errorf("XLabel = %v", pd.XLabel)
	}
	if len(pd.Series) != 0 {
		t.Errorf("Series length = %v, want 0", len(pd.Series))
	}
}

func TestPlotData_AddSeries(t *testing.T) {
	pd := NewPlotData("Test", "X", "Y")
	pd.AddSeries("line1", []PlotPoint{
		{X: 0, Y: 0},
		{X: 1, Y: 1},
		{X: 2, Y: 0},
	})
	if len(pd.Series) != 1 {
		t.Errorf("Series length = %v, want 1", len(pd.Series))
	}
	if pd.Series[0].Name != "line1" {
		t.Errorf("Series[0].Name = %v", pd.Series[0].Name)
	}
	if len(pd.Series[0].Points) != 3 {
		t.Errorf("Series[0] point count = %v, want 3", len(pd.Series[0].Points))
	}
}

func TestPlotData_AddSeriesAutoBounds(t *testing.T) {
	pd := NewPlotData("Auto", "X", "Y")
	pd.AddSeries("s", []PlotPoint{{X: -10, Y: 20}, {X: 10, Y: -5}})
	if pd.XMin != -10 || pd.XMax != 10 {
		t.Errorf("X bounds = [%v, %v], want [-10, 10]", pd.XMin, pd.XMax)
	}
	if pd.YMin != -5 || pd.YMax != 20 {
		t.Errorf("Y bounds = [%v, %v], want [-5, 20]", pd.YMin, pd.YMax)
	}
}
```

- [ ] **Step 2: Implement PlotData model**

```go
// cmd/fecim-lattice-tools-next/design/plot.go
//go:build !cgo

package design

// PlotPoint represents a single (x, y) data point.
type PlotPoint struct {
	X float64
	Y float64
}

// PlotSeries represents a named data series for plotting.
type PlotSeries struct {
	Name   string
	Points []PlotPoint
	Color  string
}

// PlotData holds the complete data model for a 2D plot.
type PlotData struct {
	Title  string
	XLabel string
	YLabel string
	Series []PlotSeries
	XMin   float64
	XMax   float64
	YMin   float64
	YMax   float64
}

// NewPlotData creates a PlotData with given labels and zero bounds.
func NewPlotData(title, xlabel, ylabel string) *PlotData {
	return &PlotData{
		Title:  title,
		XLabel: xlabel,
		YLabel: ylabel,
	}
}

// AddSeries adds a data series and auto-computes axis bounds.
func (pd *PlotData) AddSeries(name string, points []PlotPoint) {
	if len(points) == 0 {
		return
	}
	series := PlotSeries{Name: name, Points: points}
	if len(pd.Series) == 0 {
		pd.XMin, pd.XMax = points[0].X, points[0].X
		pd.YMin, pd.YMax = points[0].Y, points[0].Y
	}
	for _, p := range points {
		if p.X < pd.XMin {
			pd.XMin = p.X
		}
		if p.X > pd.XMax {
			pd.XMax = p.X
		}
		if p.Y < pd.YMin {
			pd.YMin = p.Y
		}
		if p.Y > pd.YMax {
			pd.YMax = p.Y
		}
	}
	pd.Series = append(pd.Series, series)
}
```

- [ ] **Step 3: Run tests**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/design/... -run TestPlot
```
Expected: PASS

- [ ] **Step 4: Update hysteresis adapter to use PlotData**

```go
// In buildHysteresisView, replace P-E Loop placeholder with:
loopPoints := generateDefaultLoop(snapshot)
plotData := design.NewPlotData("P-E Hysteresis Loop", "Electric Field (kV/cm)", "Polarization (µC/cm²)")
plotData.AddSeries("P-E", loopPoints)
children = append(children, plotCard(plotData, theme))
```

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/design/plot.go \
       cmd/fecim-lattice-tools-next/design/plot_test.go \
       cmd/fecim-lattice-tools-next/hysteresis_view.go
git commit -m "feat(design): add PlotData model and integrate with hysteresis view"
```

---

## Phase 3: Module 2 — Crossbar Array

### Task 3.1: Crossbar Viewmodel

**Files:**
- Create: `shared/viewmodel/crossbar/state.go`
- Create: `shared/viewmodel/crossbar/events.go`
- Create: `shared/viewmodel/crossbar/viewmodel.go`
- Create: `shared/viewmodel/crossbar/snapshot.go`
- Create: `shared/viewmodel/crossbar/viewmodel_test.go`

- [ ] **Step 1: Write failing test**

```go
// shared/viewmodel/crossbar/viewmodel_test.go
package crossbar

import (
	"testing"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New(8, 8)
	if m == nil { t.Fatal("New() returned nil") }
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New(4, 4)
	d := m.Descriptor()
	if d.ID != viewmodel.ModuleCrossbar {
		t.Errorf("Descriptor().ID = %v, want %v", d.ID, viewmodel.ModuleCrossbar)
	}
}

func TestSnapshotContainsArrayDimensions(t *testing.T) {
	m := New(16, 32)
	s := m.Snapshot()
	foundRows, foundCols := false, false
	for _, metric := range s.Metrics {
		if metric.ID == "rows" && metric.Value == "16" { foundRows = true }
		if metric.ID == "cols" && metric.Value == "32" { foundCols = true }
	}
	if !foundRows { t.Error("rows metric missing or wrong") }
	if !foundCols { t.Error("cols metric missing or wrong") }
}

func TestSnapshotSectionsContainNonIdealities(t *testing.T) {
	m := New(4, 4)
	s := m.Snapshot()
	if len(s.Sections) == 0 {
		t.Error("Snapshot().Sections is empty, expected non-ideality sections")
	}
}
```

- [ ] **Step 2: Implement state, events, viewmodel, snapshot**

```go
// shared/viewmodel/crossbar/state.go
package crossbar

type CrossbarState struct {
	Rows         int       `json:"rows"`
	Cols         int       `json:"cols"`
	Conductances [][]float64 `json:"conductances"` // [row][col] in µS
	IRDrop       float64   `json:"ir_drop"`       // percentage
	SneakPaths   bool      `json:"sneak_paths"`
	DriftFactor  float64   `json:"drift_factor"`
	InputVector  []float64 `json:"input_vector"`
	OutputVector []float64 `json:"output_vector"`
}
```

```go
// shared/viewmodel/crossbar/viewmodel.go
package crossbar

import (
	"fmt"
	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state CrossbarState
}

func New(rows, cols int) *Module {
	state := CrossbarState{Rows: rows, Cols: cols}
	state.Conductances = make([][]float64, rows)
	state.InputVector = make([]float64, cols)
	state.OutputVector = make([]float64, rows)
	for i := range state.Conductances {
		state.Conductances[i] = make([]float64, cols)
		for j := range state.Conductances[i] {
			state.Conductances[i][j] = physics.QuantizeTo30Levels(50.0)
		}
	}
	for j := range state.InputVector {
		state.InputVector[j] = 1.0
	}
	return &Module{state: state}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleCrossbar, Title: "FeCIM Crossbar Array Visualization",
		Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.",
		Status: viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
```

```go
// shared/viewmodel/crossbar/snapshot.go
package crossbar

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CrossbarState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "rows", Label: "Rows", Value: fmt.Sprintf("%d", state.Rows)},
		{ID: "cols", Label: "Columns", Value: fmt.Sprintf("%d", state.Cols)},
		{ID: "ir_drop", Label: "IR Drop", Value: fmt.Sprintf("%.1f%%", state.IRDrop*100)},
		{ID: "drift", Label: "Drift Factor", Value: fmt.Sprintf("%.2f", state.DriftFactor)},
	}

	sections := []viewmodel.Section{
		{ID: "ir_drop", Title: "IR Drop", Body: fmt.Sprintf("Voltage drop across array: %.1f%% of supply", state.IRDrop*100)},
		{ID: "sneak", Title: "Sneak Paths", Body: "Sneak path currents modeled via Kirchhoff current law at each column node."},
		{ID: "mvm", Title: "MVM Operation", Body: fmt.Sprintf("%d×%d matrix × %d-element input vector → %d-element output", state.Rows, state.Cols, state.Cols, state.Rows)},
	}

	actions := []viewmodel.Action{
		{ID: "resize", Label: "Resize Array", Kind: viewmodel.ActionCommand},
		{ID: "run_mvm", Label: "Run MVM", Kind: viewmodel.ActionCommand},
		{ID: "toggle_ir", Label: "Toggle IR Drop", Kind: viewmodel.ActionToggle},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{ID: viewmodel.ModuleCrossbar, Title: "FeCIM Crossbar Array Visualization",
			Description: "Matrix-vector multiply, IR drop, sneak paths, drift, and conductance quantization.",
			Status:      viewmodel.StatusFunctional},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
	}
}
```

- [ ] **Step 3: Run tests and verify UI boundary**

```bash
go test ./shared/viewmodel/crossbar/...
grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/crossbar/ && echo "VIOLATION" || echo "CLEAN"
```
Expected: PASS, CLEAN

- [ ] **Step 4: Commit**

```bash
git add shared/viewmodel/crossbar/
git commit -m "feat(viewmodel/crossbar): add crossbar module viewmodel with array state and MVM metrics"
```

---

### Task 3.2: Crossbar gogpu/ui Adapter

**Files:**
- Create: `cmd/fecim-lattice-tools-next/crossbar_view.go`
- Create: `cmd/fecim-lattice-tools-next/crossbar_view_test.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go`
- Modify: `cmd/fecim-lattice-tools-next/main.go`

- [ ] **Step 1: Write failing test**

```go
// cmd/fecim-lattice-tools-next/crossbar_view_test.go
//go:build !cgo

package main

import (
	"testing"
	crossbarvm "fecim-lattice-tools/shared/viewmodel/crossbar"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildCrossbarView(t *testing.T) {
	vm := crossbarvm.New(4, 4)
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildCrossbarView(snapshot, theme)
	if w == nil { t.Fatal("buildCrossbarView returned nil") }
}
```

- [ ] **Step 2: Implement crossbar adapter**

```go
// cmd/fecim-lattice-tools-next/crossbar_view.go
//go:build !cgo

package main

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildCrossbarView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	// Metrics row
	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(16).Bold(),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	// Array heatmap placeholder
	rows, cols := parseDimensions(snapshot)
	children = append(children, primitives.Box(
		primitives.Text(fmt.Sprintf("Crossbar Array (%d×%d)", rows, cols)).FontSize(16).Bold(),
		primitives.Text("Interactive heatmap coming in design system task — rendering as grid preview.").FontSize(12),
	).
		Padding(20).Gap(8).
		Background(theme.Colors.SurfaceContainer),
	)

	// Non-ideality sections
	for _, section := range snapshot.Sections {
		children = append(children, primitives.Box(
			primitives.Text(section.Title).FontSize(15).Bold(),
			primitives.Text(section.Body).FontSize(12),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}

	return primitives.Box(children...).
		Padding(24).Gap(14).
		Background(theme.Colors.Surface)
}

func parseDimensions(snapshot viewmodel.ModuleSnapshot) (int, int) {
	rows, cols := 4, 4
	for _, m := range snapshot.Metrics {
		switch m.ID {
		case "rows": fmt.Sscanf(m.Value, "%d", &rows)
		case "cols": fmt.Sscanf(m.Value, "%d", &cols)
		}
	}
	return rows, cols
}
```

- [ ] **Step 3: Wire into appmodel and buildRoot**

Update `BuildPlaceholderPorts` to add `crossbarvm.New(8, 8)` for `ModuleCrossbar`.
Add `case viewmodel.ModuleCrossbar: children = append(children, buildCrossbarView(snapshot, theme))` to buildRoot.

- [ ] **Step 4: Run tests and verify**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/...
make test-next-ui
CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
```

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/crossbar_view.go \
       cmd/fecim-lattice-tools-next/crossbar_view_test.go \
       cmd/fecim-lattice-tools-next/appmodel.go \
       cmd/fecim-lattice-tools-next/main.go
git commit -m "feat(next-shell): add crossbar module gogpu/ui adapter with array state display"
```

---

## Phase 4: Module 6 — EDA Design Suite

### Task 4.1: EDA Viewmodel

**Files:**
- Create: `shared/viewmodel/eda/state.go`
- Create: `shared/viewmodel/eda/viewmodel.go`
- Create: `shared/viewmodel/eda/snapshot.go`
- Create: `shared/viewmodel/eda/viewmodel_test.go`

- [ ] **Step 1: Write failing test**

```go
// shared/viewmodel/eda/viewmodel_test.go
package eda

import (
	"testing"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil { t.Fatal("New() returned nil") }
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New()
	if m.Descriptor().ID != viewmodel.ModuleEDA {
		t.Errorf("Descriptor().ID = %v, want %v", m.Descriptor().ID, viewmodel.ModuleEDA)
	}
}

func TestSnapshotContainsExportFormats(t *testing.T) {
	m := New()
	s := m.Snapshot()
	formatIDs := map[string]bool{}
	for _, metric := range s.Metrics {
		formatIDs[metric.ID] = true
	}
	expected := []string{"spice", "verilog", "liberty", "def", "lef"}
	for _, id := range expected {
		if !formatIDs[id] {
			t.Errorf("Missing metric for format: %s", id)
		}
	}
}

func TestSnapshotHasDesignSections(t *testing.T) {
	m := New()
	s := m.Snapshot()
	if len(s.Sections) == 0 {
		t.Error("Snapshot().Sections is empty")
	}
}
```

- [ ] **Step 2: Implement viewmodel**

```go
// shared/viewmodel/eda/state.go
package eda

type EDAState struct {
	DesignName    string `json:"design_name"`
	ProcessNode   string `json:"process_node"`   // e.g., "sky130", "gf180"
	ArrayRows     int    `json:"array_rows"`
	ArrayCols     int    `json:"array_cols"`
	ExportFormats []string `json:"export_formats"`
}
```

```go
// shared/viewmodel/eda/viewmodel.go
package eda

import (
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct {
	state EDAState
}

func New() *Module {
	return &Module{
		state: EDAState{
			DesignName:    "fecim_crossbar_8x8",
			ProcessNode:   "sky130",
			ArrayRows:     8,
			ArrayCols:     8,
			ExportFormats: []string{"spice", "verilog", "liberty", "def", "lef"},
		},
	}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleEDA, Title: "FeCIM EDA Design Suite",
		Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
		Status: viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
```

```go
// shared/viewmodel/eda/snapshot.go
package eda

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state EDAState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "design", Label: "Design", Value: state.DesignName},
		{ID: "process", Label: "Process Node", Value: state.ProcessNode},
		{ID: "spice", Label: "SPICE", Value: "ready"},
		{ID: "verilog", Label: "Verilog", Value: "ready"},
		{ID: "liberty", Label: "Liberty", Value: "ready"},
		{ID: "def", Label: "DEF", Value: "ready"},
		{ID: "lef", Label: "LEF", Value: "ready"},
	}

	sections := []viewmodel.Section{
		{ID: "spice_export", Title: "SPICE Netlist",
			Body: fmt.Sprintf("Netlist for %d×%d FeCIM crossbar with FeFET compact model. Includes parasitic extraction and corner models.", state.ArrayRows, state.ArrayCols)},
		{ID: "verilog_export", Title: "Verilog Module",
			Body: "Behavioral model for digital control logic (WL decoder, BL multiplexer, read/write FSM)."},
		{ID: "liberty_export", Title: "Liberty Timing",
			Body: fmt.Sprintf("Timing and power characterization for %s process at TT/FF/SS corners.", state.ProcessNode)},
		{ID: "physical_export", Title: "Physical Design (DEF/LEF)",
			Body: fmt.Sprintf("LEF macro for %d×%d array. DEF with placed cells and routed interconnect.", state.ArrayRows, state.ArrayCols)},
		{ID: "openlane", Title: "OpenLane Integration",
			Body: "Design configuration for OpenLane flow: synthesis → floorplan → place → route → DRC/LVS."},
	}

	actions := []viewmodel.Action{
		{ID: "generate_spice", Label: "Generate SPICE", Kind: viewmodel.ActionCommand},
		{ID: "generate_verilog", Label: "Generate Verilog", Kind: viewmodel.ActionCommand},
		{ID: "generate_all", Label: "Export All Formats", Kind: viewmodel.ActionCommand},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{ID: viewmodel.ModuleEDA, Title: "FeCIM EDA Design Suite",
			Description: "SPICE, Verilog, Liberty, DEF, LEF, and OpenLane-oriented export workflows.",
			Status:      viewmodel.StatusFunctional},
		Metrics:  metrics,
		Sections: sections,
		Actions:  actions,
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./shared/viewmodel/eda/...
grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/eda/ && echo "VIOLATION" || echo "CLEAN"
```
Expected: PASS, CLEAN

- [ ] **Step 4: Commit**

```bash
git add shared/viewmodel/eda/
git commit -m "feat(viewmodel/eda): add EDA module viewmodel with export format state"
```

---

### Task 4.2: EDA gogpu/ui Adapter

**Files:**
- Create: `cmd/fecim-lattice-tools-next/eda_view.go`
- Create: `cmd/fecim-lattice-tools-next/eda_view_test.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go`
- Modify: `cmd/fecim-lattice-tools-next/main.go`

- [ ] **Step 1: Write failing test**

```go
// cmd/fecim-lattice-tools-next/eda_view_test.go
//go:build !cgo

package main

import (
	"testing"
	edavm "fecim-lattice-tools/shared/viewmodel/eda"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildEDAView(t *testing.T) {
	vm := edavm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildEDAView(snapshot, theme)
	if w == nil { t.Fatal("buildEDAView returned nil") }
}

func TestBuildEDAView_HasExportSections(t *testing.T) {
	vm := edavm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	if len(snapshot.Sections) < 4 {
		t.Errorf("want >=4 export sections, got %d", len(snapshot.Sections))
	}
}
```

- [ ] **Step 2: Implement EDA adapter**

```go
// cmd/fecim-lattice-tools-next/eda_view.go
//go:build !cgo

package main

import (
	"fecim-lattice-tools/shared/viewmodel"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildEDAView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	// Process & design info bar
	infoBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		if m.ID == "design" || m.ID == "process" {
			infoBoxes = append(infoBoxes, primitives.Box(
				primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
				primitives.Text(m.Value).FontSize(14).Bold(),
			).Padding(10).Gap(2).Background(theme.Colors.SurfaceContainer))
		}
	}
	children = append(children, primitives.Box(infoBoxes...).Gap(8))

	// Export format status
	formatBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		if m.ID == "spice" || m.ID == "verilog" || m.ID == "liberty" || m.ID == "def" || m.ID == "lef" {
			formatBoxes = append(formatBoxes, primitives.Box(
				primitives.Text(m.Label).FontSize(12).Bold(),
				primitives.Text(m.Value).FontSize(11).Color(theme.Colors.Primary),
			).Padding(8).Gap(2).Background(theme.Colors.SurfaceContainer))
		}
	}
	children = append(children, primitives.Box(formatBoxes...).Gap(6))

	// Design workflow sections
	for _, section := range snapshot.Sections {
		children = append(children, edaExportCard(section, theme))
	}

	// Actions
	actionBoxes := []widget.Widget{}
	for _, action := range snapshot.Actions {
		actionBoxes = append(actionBoxes, primitives.Box(
			primitives.Text(action.Label).FontSize(13).Color(theme.Colors.OnPrimary),
		).Padding(12).Background(theme.Colors.Primary))
	}
	if len(actionBoxes) > 0 {
		children = append(children, primitives.Box(actionBoxes...).Gap(8))
	}

	return primitives.Box(children...).
		Padding(24).Gap(14).
		Background(theme.Colors.Surface)
}

func edaExportCard(section viewmodel.Section, theme *material3.Theme) widget.Widget {
	return primitives.Box(
		primitives.Text(section.Title).FontSize(15).Bold(),
		primitives.Text(section.Body).FontSize(12),
	).
		Padding(12).Gap(4).
		Background(theme.Colors.SurfaceContainer)
}
```

- [ ] **Step 3: Wire into appmodel and buildRoot**

Add `case viewmodel.ModuleEDA: ports = append(ports, edavm.New())` and `case viewmodel.ModuleEDA: children = append(children, buildEDAView(snapshot, theme))`.

- [ ] **Step 4: Verify**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/...
make test-next-ui
CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
```

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/eda_view.go \
       cmd/fecim-lattice-tools-next/eda_view_test.go \
       cmd/fecim-lattice-tools-next/appmodel.go \
       cmd/fecim-lattice-tools-next/main.go
git commit -m "feat(next-shell): add EDA module gogpu/ui adapter with export format display"
```

---

## Phase 5: Module 4 — Peripheral Circuits

### Task 5.1: Circuits Viewmodel

**Files:**
- Create: `shared/viewmodel/circuits/state.go`
- Create: `shared/viewmodel/circuits/viewmodel.go`
- Create: `shared/viewmodel/circuits/snapshot.go`
- Create: `shared/viewmodel/circuits/viewmodel_test.go`

- [ ] **Step 1: Write failing test**

```go
// shared/viewmodel/circuits/viewmodel_test.go
package circuits

import (
	"testing"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestModuleImplementsModulePort(t *testing.T) {
	var m viewmodel.ModulePort = New()
	if m == nil { t.Fatal("New() returned nil") }
}

func TestDescriptorHasCorrectID(t *testing.T) {
	m := New()
	if m.Descriptor().ID != viewmodel.ModuleCircuits {
		t.Errorf("Descriptor().ID = %v, want %v", m.Descriptor().ID, viewmodel.ModuleCircuits)
	}
}

func TestSnapshotContainsCircuitBlocks(t *testing.T) {
	m := New()
	s := m.Snapshot()
	blockIDs := map[string]bool{}
	for _, metric := range s.Metrics {
		blockIDs[metric.ID] = true
	}
	expected := []string{"adc", "dac", "tia", "charge_pump", "ispp"}
	for _, id := range expected {
		if !blockIDs[id] {
			t.Errorf("Missing metric for circuit: %s", id)
		}
	}
}

func TestSnapshotHasReadPathSection(t *testing.T) {
	m := New()
	s := m.Snapshot()
	found := false
	for _, section := range s.Sections {
		if section.ID == "read_path" { found = true; break }
	}
	if !found { t.Error("Missing read_path section") }
}
```

- [ ] **Step 2: Implement viewmodel — state**

```go
// shared/viewmodel/circuits/state.go
package circuits

type CircuitsState struct {
	ADCResolution int     `json:"adc_resolution"` // bits
	DACResolution int     `json:"dac_resolution"` // bits
	TIAGain       float64 `json:"tia_gain"`       // V/A
	ChargePumpStages int  `json:"charge_pump_stages"`
	SupplyVoltage float64 `json:"supply_voltage"` // V
	ISPPEnabled   bool    `json:"ispp_enabled"`
}
```

```go
// shared/viewmodel/circuits/viewmodel.go
package circuits

import (
	"fecim-lattice-tools/shared/viewmodel"
)

type Module struct{ state CircuitsState }

func New() *Module {
	return &Module{state: CircuitsState{
		ADCResolution: 5, DACResolution: 5, TIAGain: 1e4,
		ChargePumpStages: 4, SupplyVoltage: 1.8, ISPPEnabled: true,
	}}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID: viewmodel.ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer",
		Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
		Status: viewmodel.StatusFunctional,
	}
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.state) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
```

```go
// shared/viewmodel/circuits/snapshot.go
package circuits

import (
	"fmt"
	"fecim-lattice-tools/shared/viewmodel"
)

func buildSnapshot(state CircuitsState) viewmodel.ModuleSnapshot {
	metrics := []viewmodel.Metric{
		{ID: "adc", Label: "ADC", Value: fmt.Sprintf("%d-bit SAR", state.ADCResolution)},
		{ID: "dac", Label: "DAC", Value: fmt.Sprintf("%d-bit R-2R", state.DACResolution)},
		{ID: "tia", Label: "TIA", Value: fmt.Sprintf("%.0f kΩ gain", state.TIAGain/1e3)},
		{ID: "charge_pump", Label: "Charge Pump", Value: fmt.Sprintf("%d-stage Dickson", state.ChargePumpStages)},
		{ID: "ispp", Label: "ISPP", Value: fmt.Sprintf("%v", state.ISPPEnabled)},
		{ID: "supply", Label: "Vdd", Value: fmt.Sprintf("%.1f V", state.SupplyVoltage)},
	}

	sections := []viewmodel.Section{
		{ID: "read_path", Title: "Read Path", Body: fmt.Sprintf("TIA (%.0f kΩ) → %d-bit SAR ADC → digital output. Latency: ~%.1f µs at %.1f V supply.",
			state.TIAGain/1e3, state.ADCResolution, float64(state.ADCResolution)*0.5, state.SupplyVoltage)},
		{ID: "write_path", Title: "Write Path (ISPP)", Body: fmt.Sprintf("%d-stage Dickson charge pump → %d-bit DAC → ISPP pulse train. Verify after each pulse within ±1 level tolerance.",
			state.ChargePumpStages, state.DACResolution)},
		{ID: "adc_characterization", Title: "ADC INL/DNL", Body: fmt.Sprintf("%d-bit SAR ADC with INL < 0.5 LSB, DNL < 0.3 LSB (educational model).", state.ADCResolution)},
	}

	actions := []viewmodel.Action{
		{ID: "run_read", Label: "Simulate Read", Kind: viewmodel.ActionCommand},
		{ID: "run_write", Label: "Simulate Write", Kind: viewmodel.ActionCommand},
		{ID: "toggle_ispp", Label: "Toggle ISPP", Kind: viewmodel.ActionToggle},
	}

	return viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{ID: viewmodel.ModuleCircuits, Title: "FeCIM Peripheral Circuits Visualizer",
			Description: "DAC, ADC, TIA, read path, write path, and ISPP circuit behavior.",
			Status:      viewmodel.StatusFunctional},
		Metrics: metrics, Sections: sections, Actions: actions,
	}
}
```

- [ ] **Step 3: Run tests**

```bash
go test ./shared/viewmodel/circuits/...
grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/circuits/ && echo "VIOLATION" || echo "CLEAN"
```
Expected: PASS, CLEAN

- [ ] **Step 4: Commit**

```bash
git add shared/viewmodel/circuits/
git commit -m "feat(viewmodel/circuits): add peripheral circuits module viewmodel"
```

---

### Task 5.2: Circuits gogpu/ui Adapter

**Files:**
- Create: `cmd/fecim-lattice-tools-next/circuits_view.go`
- Create: `cmd/fecim-lattice-tools-next/circuits_view_test.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go`
- Modify: `cmd/fecim-lattice-tools-next/main.go`

- [ ] **Step 1: Write failing test**

```go
// cmd/fecim-lattice-tools-next/circuits_view_test.go
//go:build !cgo

package main

import (
	"testing"
	circuitsvm "fecim-lattice-tools/shared/viewmodel/circuits"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func TestBuildCircuitsView(t *testing.T) {
	vm := circuitsvm.New()
	snapshot := vm.Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	w := buildCircuitsView(snapshot, theme)
	if w == nil { t.Fatal("buildCircuitsView returned nil") }
}
```

- [ ] **Step 2: Implement circuits adapter**

```go
// cmd/fecim-lattice-tools-next/circuits_view.go
//go:build !cgo

package main

import (
	"fecim-lattice-tools/shared/viewmodel"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"
)

func buildCircuitsView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(snapshot.Descriptor.Title).FontSize(22).Bold(),
		primitives.Text(snapshot.Descriptor.Description).FontSize(14),
	}

	// Circuit block metrics in a grid
	metricBoxes := []widget.Widget{}
	for _, m := range snapshot.Metrics {
		status := theme.Colors.Primary
		if m.ID == "ispp" && m.Value == "false" { status = theme.Colors.OnSurfaceVariant }
		metricBoxes = append(metricBoxes, primitives.Box(
			primitives.Text(m.Label).FontSize(11).Color(theme.Colors.OnSurfaceVariant),
			primitives.Text(m.Value).FontSize(14).Bold().Color(status),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}
	children = append(children, primitives.Box(metricBoxes...).Gap(8))

	// Read/Write path sections
	for _, section := range snapshot.Sections {
		children = append(children, primitives.Box(
			primitives.Text(section.Title).FontSize(15).Bold(),
			primitives.Text(section.Body).FontSize(12),
		).Padding(12).Gap(4).Background(theme.Colors.SurfaceContainer))
	}

	return primitives.Box(children...).
		Padding(24).Gap(14).
		Background(theme.Colors.Surface)
}
```

- [ ] **Step 3: Wire into appmodel and buildRoot, verify**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/...
CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
```

- [ ] **Step 4: Commit**

```bash
git add cmd/fecim-lattice-tools-next/circuits_view.go \
       cmd/fecim-lattice-tools-next/circuits_view_test.go \
       cmd/fecim-lattice-tools-next/appmodel.go \
       cmd/fecim-lattice-tools-next/main.go
git commit -m "feat(next-shell): add circuits module gogpu/ui adapter"
```

---

## Phase 6: Module 3 — MNIST Inference

### Task 6.1: MNIST Viewmodel + Adapter

**Files:**
- Create: `shared/viewmodel/mnist/state.go`
- Create: `shared/viewmodel/mnist/viewmodel.go`
- Create: `shared/viewmodel/mnist/snapshot.go`
- Create: `shared/viewmodel/mnist/viewmodel_test.go`
- Create: `cmd/fecim-lattice-tools-next/mnist_view.go`
- Create: `cmd/fecim-lattice-tools-next/mnist_view_test.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go`
- Modify: `cmd/fecim-lattice-tools-next/main.go`

**Combined task following same pattern as Modules 1/2/6/4 above.** Viewmodel state includes: accuracy, quantization levels, confusion matrix data, pipeline stages. Adapter renders accuracy metric, confusion matrix as grid preview, pipeline stage cards.

- [ ] **Step 1-4: Follow same TDD pattern** — write failing test → implement viewmodel → implement adapter → wire in

---

## Phase 7: Module 7 — Documentation & Education Hub

### Task 7.1: Docs Viewmodel + Adapter

**Files:**
- Create: `shared/viewmodel/docs/state.go`
- Create: `shared/viewmodel/docs/viewmodel.go`
- Create: `shared/viewmodel/docs/snapshot.go`
- Create: `shared/viewmodel/docs/viewmodel_test.go`
- Create: `cmd/fecim-lattice-tools-next/docs_view.go`
- Create: `cmd/fecim-lattice-tools-next/docs_view_test.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go`
- Modify: `cmd/fecim-lattice-tools-next/main.go`

**Combined task following same pattern.** Viewmodel state includes: doc sections, citations, tutorial progress, search query. Adapter renders doc browser with TOC, search, citation cards. Status: `StatusFunctional`.

- [ ] **Step 1-4: Follow same TDD pattern**

---

## Phase 8: Screenshot Generation

### Task 8.1: Extend Screenshotter for gogpu/ui

**Files:**
- Create: `cmd/fecim-screenshotter/next_capture.go`
- Modify: `cmd/fecim-screenshotter/main.go`

- [ ] **Step 1: Add gogpu/ui headless capture support**

```go
// cmd/fecim-screenshotter/next_capture.go
//go:build !cgo

package main

import (
	"image"
	"image/png"
	"os"
)

// captureNextModule generates a screenshot by building the gogpu/ui widget tree
// and rendering it to a canvas, then capturing the canvas as PNG.
func captureNextModule(moduleName string, width, height int) (*image.RGBA, error) {
	// Build widget tree from viewmodel snapshot
	// Render to gg.Context
	// Capture framebuffer
	// TODO: implement gogpu/ui headless capture — requires GPU context or software rasterizer
	return nil, nil
}
```

- [ ] **Step 2: Generate baseline screenshots**

```bash
go run ./cmd/fecim-screenshotter -out screenshots -w 1400 -h 900 -tag gogpu-baseline
```

- [ ] **Step 3: Commit**

```bash
git add cmd/fecim-screenshotter/ screenshots/
git commit -m "feat(screenshotter): add gogpu/ui headless capture support and baseline screenshots"
```

---

## Phase 9: Final Integration & Verification

### Task 9.1: Full Test Suite Verification

- [ ] **Step 1: Run all viewmodel tests**

```bash
go test ./shared/viewmodel/...
```
Expected: ALL PASS (7 viewmodel packages)

- [ ] **Step 2: Run next-shell tests**

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/...
make test-next-ui
```
Expected: ALL PASS

- [ ] **Step 3: Verify UI boundary**

```bash
grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/ && echo "VIOLATION" || echo "CLEAN"
```
Expected: CLEAN

- [ ] **Step 4: Build both shells**

```bash
go build ./cmd/fecim-lattice-tools
CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
```
Expected: both succeed

- [ ] **Step 5: Full test suite**

```bash
go test ./...
go test -race -short ./shared/... ./validation/...
```
Expected: ALL PASS

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "chore: full gogpu/ui migration complete — all 7 modules functional

Migration order: Hysteresis → Crossbar → EDA → Circuits → MNIST → Docs
All modules implement ModulePort via shared/viewmodel with StatusFunctional.
Shell restructured to sidebar + content Material 3 layout.
Design system tokens, PlotData model, and screenshot support added.
Verified: go test ./..., make test-next-ui, UI boundary clean."
```

---

## Summary

| Phase | Status | Modules | Key deliverables |
|-------|--------|---------|-----------------|
| 1 | Not started | Shell | Sidebar nav, design tokens, layout restructure |
| 2 | Not started | M1 Hysteresis | Viewmodel + adapter + PlotData |
| 3 | Not started | M2 Crossbar | Viewmodel + adapter |
| 4 | Not started | M6 EDA | Viewmodel + adapter |
| 5 | Not started | M4 Circuits | Viewmodel + adapter |
| 6 | Not started | M3 MNIST | Viewmodel + adapter |
| 7 | Not started | M7 Docs | Viewmodel + adapter |
| 8 | Not started | Screenshotter | gogpu/ui captures |
| 9 | Not started | Integration | Full verification |
