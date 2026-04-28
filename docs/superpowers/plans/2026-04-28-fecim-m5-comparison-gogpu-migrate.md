# M5 Comparison → gogpu/ui Pilot Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the `ModuleComparison` placeholder in `cmd/fecim-lattice-tools-next/` with a real read-only viewmodel surfacing `module5-comparison/pkg/comparison` data, establishing the canonical viewmodel ↔ gogpu/ui pattern for the remaining six modules.

**Architecture:** Add a thin `Architectures()` accessor in `pkg/comparison`, build a UI-neutral viewmodel sub-package (`shared/viewmodel/comparison/`) that converts architectures into a `ModuleSnapshot`, and add a gogpu/ui adapter (`comparison_view.go`) that the shell's `buildRoot` switch dispatches to for `ModuleComparison`. Other modules continue using the generic placeholder.

**Tech Stack:** Go 1.25 (zero-CGO build path), `github.com/gogpu/ui/...` widgets/primitives, `material3` theme, `shared/viewmodel` boundary types. Tests use the standard library only.

**Spec:** `docs/superpowers/specs/2026-04-28-fecim-m5-comparison-gogpu-migrate-design.md`

**TDD discipline:** Each behavior task follows RED → GREEN → REFACTOR per `CLAUDE.md` hard-rule. Mechanical wiring tasks use `TDD: N/A` with a written reason and reference to the test that *does* cover the behavior.

---

## Pre-flight (verify worktree state — done by writing-plans, no action required)

- Branch: `feat/m5-comparison-gogpu-migrate` off `main` HEAD
- Worktree: `.worktrees/feat-m5-comparison-gogpu-migrate/`
- Spec committed: `5fa4ea6d docs: add M5 comparison gogpu/ui pilot migration spec`
- Baseline tests passing: `make test-next-ui` is green at branch start

---

### Task 1: Add `Architectures()` accessor in `pkg/comparison`

The viewmodel needs one canonical entry point that returns the three reference architectures. The package already has `TraditionalCPU()`, `GPUAccelerator()`, and `FeCIMChip()` constructors but no aggregator — `CompareArchitectures` builds the slice inline. We add a deterministic accessor so the viewmodel doesn't have to know the canonical set.

**Files:**
- Test: `module5-comparison/pkg/comparison/architecture_test.go` (existing or new — check before creating)
- Modify: `module5-comparison/pkg/comparison/architecture.go`

- [ ] **Step 1: Check whether `architecture_test.go` exists**

Run: `ls module5-comparison/pkg/comparison/architecture_test.go 2>/dev/null && echo EXISTS || echo CREATE`

If `EXISTS`, append the new test function to it. If `CREATE`, create it with the test function inside `package comparison`.

- [ ] **Step 2: Write the failing test (RED)**

Append to `module5-comparison/pkg/comparison/architecture_test.go` (or create with this content if missing):

```go
package comparison

import "testing"

func TestArchitectures_ReturnsCanonicalSet(t *testing.T) {
	got := Architectures()
	if len(got) != 3 {
		t.Fatalf("Architectures() returned %d entries, want 3", len(got))
	}
	wantNames := []string{"Traditional CPU+DRAM", "GPU Accelerator", "FeCIM CIM"}
	for i, want := range wantNames {
		if got[i] == nil {
			t.Fatalf("Architectures()[%d] is nil", i)
		}
		if got[i].Name != want {
			t.Errorf("Architectures()[%d].Name = %q, want %q", i, got[i].Name, want)
		}
	}
}

func TestArchitectures_ReturnsFreshSliceEachCall(t *testing.T) {
	a := Architectures()
	b := Architectures()
	if &a[0] == &b[0] {
		t.Fatal("Architectures() returned shared backing array; callers could mutate the canonical set")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

Run: `cd /home/xel/git/sages-openclaw/workspace-riju/fecim-lattice-tools/.worktrees/feat-m5-comparison-gogpu-migrate && go test ./module5-comparison/pkg/comparison/ -run TestArchitectures -v`

Expected: FAIL with `undefined: Architectures` compile error.

- [ ] **Step 4: Implement `Architectures()` (GREEN)**

Add to `module5-comparison/pkg/comparison/architecture.go` (immediately after the `CustomArchitecture` function):

```go
// Architectures returns the canonical reference architecture set used for
// comparison views. Each call returns a fresh slice with newly-constructed
// architectures so callers may mutate without affecting subsequent calls.
func Architectures() []*Architecture {
	return []*Architecture{
		TraditionalCPU(),
		GPUAccelerator(),
		FeCIMChip(),
	}
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./module5-comparison/pkg/comparison/ -run TestArchitectures -v`

Expected: PASS (both tests).

- [ ] **Step 6: Verify no regressions in the package**

Run: `go test ./module5-comparison/pkg/comparison/`

Expected: PASS (whole package green).

- [ ] **Step 7: Commit**

```bash
git add module5-comparison/pkg/comparison/architecture.go module5-comparison/pkg/comparison/architecture_test.go
git commit -m "$(cat <<'EOF'
feat(comparison): add Architectures() canonical accessor

Returns []*Architecture{TraditionalCPU(), GPUAccelerator(), FeCIMChip()}.
Required by upcoming viewmodel sub-package which must not know the canonical set.

TDD evidence:
- RED: TestArchitectures_ReturnsCanonicalSet failed with "undefined: Architectures"
- GREEN: 2 tests pass (canonical set + fresh-slice-per-call)
- Verification: go test ./module5-comparison/pkg/comparison/

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 2: Create `shared/viewmodel/comparison/` package — pure `buildSnapshot`

A pure function `buildSnapshot([]*pkg.Architecture) viewmodel.ModuleSnapshot` is the heart of the viewmodel. It's pure (no time, no state) so its tests are deterministic and the function is reusable from the `Module` wrapper.

**Files:**
- Create: `shared/viewmodel/comparison/snapshot.go`
- Create: `shared/viewmodel/comparison/snapshot_test.go`

- [ ] **Step 1: Write the failing test (RED)**

Create `shared/viewmodel/comparison/snapshot_test.go`:

```go
package comparison

import (
	"strings"
	"testing"

	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestBuildSnapshot_DescriptorIsFunctionalComparison(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	if snap.Descriptor.ID != viewmodel.ModuleComparison {
		t.Errorf("Descriptor.ID = %q, want %q", snap.Descriptor.ID, viewmodel.ModuleComparison)
	}
	if snap.Descriptor.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor.Status = %q, want %q", snap.Descriptor.Status, viewmodel.StatusFunctional)
	}
}

func TestBuildSnapshot_HasOneSectionPerArchitecture(t *testing.T) {
	archs := pkg.Architectures()
	snap := buildSnapshot(archs)
	if len(snap.Sections) != len(archs) {
		t.Fatalf("len(Sections) = %d, want %d", len(snap.Sections), len(archs))
	}
	for i, a := range archs {
		if snap.Sections[i].Title != a.Name {
			t.Errorf("Sections[%d].Title = %q, want %q", i, snap.Sections[i].Title, a.Name)
		}
		if snap.Sections[i].ID == "" {
			t.Errorf("Sections[%d].ID is empty", i)
		}
	}
}

func TestBuildSnapshot_SectionBodyIncludesPhysicalAndPerformanceFields(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	cpu := snap.Sections[0].Body
	for _, want := range []string{"Technology", "Process", "TDP", "TOPS"} {
		if !strings.Contains(cpu, want) {
			t.Errorf("CPU section body missing %q label\nbody: %s", want, cpu)
		}
	}
}

func TestBuildSnapshot_FeCIMSectionFlagsEstimatedValues(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	fecim := snap.Sections[2]
	if fecim.Title != "FeCIM CIM" {
		t.Fatalf("Sections[2].Title = %q, want FeCIM CIM", fecim.Title)
	}
	if !strings.Contains(strings.ToLower(fecim.Body), "estimated") {
		t.Errorf("FeCIM section body must flag IsEstimated=true (per honesty-audit policy)\nbody: %s", fecim.Body)
	}
}

func TestBuildSnapshot_HasArchitectureCountMetric(t *testing.T) {
	snap := buildSnapshot(pkg.Architectures())
	if len(snap.Metrics) == 0 {
		t.Fatal("snapshot has no metrics")
	}
	got := snap.Metrics[0]
	if got.ID != "count" {
		t.Errorf("Metrics[0].ID = %q, want count", got.ID)
	}
	if got.Value != "3" {
		t.Errorf("Metrics[0].Value = %q, want 3", got.Value)
	}
	if got.Confidence != "deterministic" {
		t.Errorf("Metrics[0].Confidence = %q, want deterministic", got.Confidence)
	}
}

func TestBuildSnapshot_DeterministicForSameInput(t *testing.T) {
	archs := pkg.Architectures()
	a := buildSnapshot(archs)
	b := buildSnapshot(archs)
	if !a.UpdatedAt.IsZero() || !b.UpdatedAt.IsZero() {
		t.Fatal("buildSnapshot must use zero time for deterministic tests")
	}
	if len(a.Sections) != len(b.Sections) {
		t.Fatalf("section counts differ across calls: %d vs %d", len(a.Sections), len(b.Sections))
	}
	for i := range a.Sections {
		if a.Sections[i] != b.Sections[i] {
			t.Errorf("Sections[%d] differs across calls\n  a: %+v\n  b: %+v", i, a.Sections[i], b.Sections[i])
		}
	}
}

func TestBuildSnapshot_EmptyInputProducesNoSections(t *testing.T) {
	snap := buildSnapshot(nil)
	if len(snap.Sections) != 0 {
		t.Errorf("nil input: len(Sections) = %d, want 0", len(snap.Sections))
	}
	if snap.Metrics[0].Value != "0" {
		t.Errorf("nil input: count metric = %q, want 0", snap.Metrics[0].Value)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./shared/viewmodel/comparison/ -v`

Expected: FAIL with `package shared/viewmodel/comparison: no Go files in ...` or `undefined: buildSnapshot`.

- [ ] **Step 3: Implement `buildSnapshot` (GREEN)**

Create `shared/viewmodel/comparison/snapshot.go`:

```go
package comparison

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

func descriptor() viewmodel.ModuleDescriptor {
	return viewmodel.ModuleDescriptor{
		ID:          viewmodel.ModuleComparison,
		Title:       "FeCIM Comparison",
		Description: "Evidence-first technology comparison and scenario analysis.",
		Status:      viewmodel.StatusFunctional,
	}
}

// buildSnapshot converts a slice of architectures into a UI-neutral
// ModuleSnapshot. Pure: same input → same output, no clock, no I/O.
func buildSnapshot(archs []*pkg.Architecture) viewmodel.ModuleSnapshot {
	sections := make([]viewmodel.Section, 0, len(archs))
	for _, a := range archs {
		if a == nil {
			continue
		}
		sections = append(sections, viewmodel.Section{
			ID:    sectionID(a.Name),
			Title: a.Name,
			Body:  architectureBody(a),
		})
	}
	metrics := []viewmodel.Metric{
		{
			ID:         "count",
			Label:      "Architectures compared",
			Value:      strconv.Itoa(len(sections)),
			Confidence: "deterministic",
		},
	}
	return viewmodel.ModuleSnapshot{
		Descriptor: descriptor(),
		Metrics:    metrics,
		Sections:   sections,
		UpdatedAt:  time.Time{},
	}
}

func sectionID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "+", "-")
	return id
}

func architectureBody(a *pkg.Architecture) string {
	estimated := ""
	if a.IsEstimated {
		estimated = " (estimated; not validated)"
	}
	return fmt.Sprintf(
		"Technology: %s%s\nProcess node: %.0f nm\nChip area: %.0f mm²\nTDP: %.1f W\nPeak TOPS: %.2f\nTOPS/W: %.3f\nMemory: %.0f GB @ %.0f GB/s",
		a.Technology, estimated,
		a.ProcessNode,
		a.ChipArea,
		a.TDP,
		a.PeakTOPS,
		a.TOPSPerWatt,
		a.MemorySize, a.MemoryBW,
	)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./shared/viewmodel/comparison/ -v`

Expected: PASS (all 7 tests).

- [ ] **Step 5: Commit**

```bash
git add shared/viewmodel/comparison/snapshot.go shared/viewmodel/comparison/snapshot_test.go
git commit -m "$(cat <<'EOF'
feat(viewmodel/comparison): add pure buildSnapshot

Converts []*comparison.Architecture into a UI-neutral ModuleSnapshot:
- One Section per architecture (Technology, Process, area, TDP, TOPS, TOPS/W, Memory)
- One "count" metric (deterministic confidence)
- FeCIM section body explicitly flags estimated values per honesty-audit policy
- Zero UpdatedAt for test determinism

Zero deps on Fyne or gogpu/ui — boundary stays UI-neutral.

TDD evidence:
- RED: 7 tests failed with "undefined: buildSnapshot"
- GREEN: all 7 tests pass (descriptor, per-arch sections, body fields, FeCIM estimated flag, count metric, determinism, empty input)
- Verification: go test ./shared/viewmodel/comparison/

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 3: Comparison `Module` struct implementing `ModulePort`

Wraps `buildSnapshot` to satisfy `viewmodel.ModulePort`. Captures the architectures slice at construction so `Snapshot()` is repeatable.

**Files:**
- Create: `shared/viewmodel/comparison/viewmodel.go`
- Create: `shared/viewmodel/comparison/viewmodel_test.go`

- [ ] **Step 1: Write the failing test (RED)**

Create `shared/viewmodel/comparison/viewmodel_test.go`:

```go
package comparison

import (
	"testing"

	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

func TestNew_ReturnsModuleWithCanonicalArchitectures(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	snap := m.Snapshot()
	want := len(pkg.Architectures())
	if len(snap.Sections) != want {
		t.Errorf("New().Snapshot() Sections = %d, want %d (canonical architecture count)", len(snap.Sections), want)
	}
}

func TestModule_DescriptorMatchesSnapshot(t *testing.T) {
	m := New()
	desc := m.Descriptor()
	snap := m.Snapshot()
	if desc != snap.Descriptor {
		t.Errorf("Descriptor() and Snapshot().Descriptor disagree\n  desc: %+v\n  snap: %+v", desc, snap.Descriptor)
	}
	if desc.ID != viewmodel.ModuleComparison {
		t.Errorf("Descriptor().ID = %q, want %q", desc.ID, viewmodel.ModuleComparison)
	}
	if desc.Status != viewmodel.StatusFunctional {
		t.Errorf("Descriptor().Status = %q, want %q (no longer placeholder)", desc.Status, viewmodel.StatusFunctional)
	}
}

func TestModule_ApplyAction_ReturnsErrUnsupported(t *testing.T) {
	m := New()
	err := m.ApplyAction(viewmodel.Action{ID: "anything"})
	if err != viewmodel.ErrUnsupportedAction {
		t.Errorf("ApplyAction error = %v, want viewmodel.ErrUnsupportedAction", err)
	}
}

func TestModule_StartStop_AreNoOpsAndIdempotent(t *testing.T) {
	m := New()
	m.Start()
	m.Start()
	m.Stop()
	m.Stop()
	// reaching here without panic is the assertion
}

func TestModule_SatisfiesModulePortInterface(t *testing.T) {
	var _ viewmodel.ModulePort = New()
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./shared/viewmodel/comparison/ -run 'TestNew|TestModule' -v`

Expected: FAIL with `undefined: New` (compile error).

- [ ] **Step 3: Implement `Module` (GREEN)**

Create `shared/viewmodel/comparison/viewmodel.go`:

```go
package comparison

import (
	pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
	"fecim-lattice-tools/shared/viewmodel"
)

// Module is a read-only viewmodel for the FeCIM comparison module.
// Implements viewmodel.ModulePort. Architectures are captured at construction
// so Snapshot is deterministic across calls.
type Module struct {
	architectures []*pkg.Architecture
}

// New constructs a Module from the canonical architecture set.
func New() *Module {
	return &Module{architectures: pkg.Architectures()}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor { return descriptor() }
func (m *Module) Snapshot() viewmodel.ModuleSnapshot     { return buildSnapshot(m.architectures) }
func (m *Module) ApplyAction(viewmodel.Action) error     { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                                 {}
func (m *Module) Stop()                                  {}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./shared/viewmodel/comparison/ -v`

Expected: PASS (all snapshot tests + 5 new module tests).

- [ ] **Step 5: Commit**

```bash
git add shared/viewmodel/comparison/viewmodel.go shared/viewmodel/comparison/viewmodel_test.go
git commit -m "$(cat <<'EOF'
feat(viewmodel/comparison): add Module type implementing ModulePort

- Module wraps buildSnapshot via captured architecture slice
- ApplyAction returns viewmodel.ErrUnsupportedAction (read-only MVP)
- Start/Stop are no-ops (static data, no lifecycle)
- New() pulls canonical set from pkg/comparison.Architectures()

TDD evidence:
- RED: 5 tests failed with "undefined: New"
- GREEN: descriptor parity, ApplyAction, Start/Stop idempotency, ModulePort interface satisfaction
- Verification: go test ./shared/viewmodel/comparison/

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 4: Wire `BuildPlaceholderPorts` to inject the real comparison module

`BuildPlaceholderPorts` currently builds a `StaticModule` for every descriptor. Replace the comparison branch with `comparisonvm.New()` so the gogpu/ui shell receives a functional port.

**Files:**
- Modify: `cmd/fecim-lattice-tools-next/appmodel.go`
- Modify: `cmd/fecim-lattice-tools-next/appmodel_test.go`

- [ ] **Step 1: Write the failing test (RED)**

Append to `cmd/fecim-lattice-tools-next/appmodel_test.go`:

```go
func TestBuildPlaceholderPorts_ComparisonIsFunctional(t *testing.T) {
	ports := BuildPlaceholderPorts()
	var got viewmodel.ModulePort
	for _, p := range ports {
		if p.Descriptor().ID == viewmodel.ModuleComparison {
			got = p
			break
		}
	}
	if got == nil {
		t.Fatal("no port found for ModuleComparison")
	}
	if got.Descriptor().Status != viewmodel.StatusFunctional {
		t.Errorf("comparison port Status = %q, want %q (no longer placeholder)",
			got.Descriptor().Status, viewmodel.StatusFunctional)
	}
	snap := got.Snapshot()
	if len(snap.Sections) < 3 {
		t.Errorf("comparison snapshot has %d sections, want >= 3 (one per canonical architecture)", len(snap.Sections))
	}
	if len(snap.Metrics) == 0 {
		t.Error("comparison snapshot has no metrics")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -run TestBuildPlaceholderPorts_ComparisonIsFunctional -v`

Expected: FAIL — comparison port reports `StatusPlaceholder`, not `StatusFunctional`.

- [ ] **Step 3: Modify `BuildPlaceholderPorts` (GREEN)**

Replace the contents of `cmd/fecim-lattice-tools-next/appmodel.go` with:

```go
package main

import (
	"fecim-lattice-tools/shared/viewmodel"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
)

type AppSpec struct {
	Title   string
	Command string
	Width   int
	Height  int
}

func DefaultAppSpec() AppSpec {
	return AppSpec{
		Title:   "FeCIM Lattice Tools Next",
		Command: "fecim-lattice-tools-next",
		Width:   1400,
		Height:  900,
	}
}

func BuildPlaceholderPorts() []viewmodel.ModulePort {
	descriptors := viewmodel.KnownDescriptors()
	ports := make([]viewmodel.ModulePort, 0, len(descriptors))
	for _, descriptor := range descriptors {
		if descriptor.ID == viewmodel.ModuleComparison {
			ports = append(ports, comparisonvm.New())
			continue
		}
		ports = append(ports, viewmodel.NewStaticModule(descriptor, []viewmodel.Section{
			{
				ID:    "migration-status",
				Title: "Migration Status",
				Body:  "This module is represented by a UI-neutral placeholder while the gogpu/ui shell reaches parity with the current Fyne implementation.",
			},
		}))
	}
	return ports
}
```

- [ ] **Step 4: Run tests to verify the new test passes AND existing tests still pass**

Run: `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -v`

Expected: PASS — the new `TestBuildPlaceholderPorts_ComparisonIsFunctional` passes, and the pre-existing `TestBuildPlaceholderPortsCoversAllKnownDescriptors` still passes (it asserts `len(snapshot.Sections) > 0` and `ApplyAction(unknown)` returns non-nil error — both true for the new comparison module).

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/appmodel.go cmd/fecim-lattice-tools-next/appmodel_test.go
git commit -m "$(cat <<'EOF'
feat(next-shell): inject real comparison viewmodel into BuildPlaceholderPorts

ModuleComparison no longer renders as a generic placeholder — it now ships
with the comparison viewmodel package (Sections per architecture, count metric).
Other six modules continue using the static placeholder until their viewmodels
land in follow-up PRs.

TDD evidence:
- RED: TestBuildPlaceholderPorts_ComparisonIsFunctional failed (comparison port reported StatusPlaceholder)
- GREEN: new test passes; existing TestBuildPlaceholderPortsCoversAllKnownDescriptors still passes
- Verification: CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 5: gogpu/ui adapter — `buildComparisonView`

A pure function turning a `ModuleSnapshot` into a gogpu/ui widget tree. Lives next to `main.go` because it imports `gogpu/ui` packages (which the viewmodel side cannot).

**Files:**
- Create: `cmd/fecim-lattice-tools-next/comparison_view.go`
- Create: `cmd/fecim-lattice-tools-next/comparison_view_test.go`

- [ ] **Step 1: Write the failing test (RED)**

Create `cmd/fecim-lattice-tools-next/comparison_view_test.go`:

```go
//go:build !cgo

package main

import (
	"testing"

	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
	comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"
)

func TestBuildComparisonView_ReturnsNonNilWidget(t *testing.T) {
	snap := comparisonvm.New().Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))
	got := buildComparisonView(snap, theme)
	if got == nil {
		t.Fatal("buildComparisonView returned nil")
	}
}

func TestBuildComparisonView_HandlesEmptySnapshot(t *testing.T) {
	snap := viewmodel.ModuleSnapshot{
		Descriptor: viewmodel.ModuleDescriptor{
			ID:          viewmodel.ModuleComparison,
			Title:       "FeCIM Comparison",
			Description: "Evidence-first technology comparison.",
			Status:      viewmodel.StatusFunctional,
		},
	}
	theme := material3.New(widget.Hex(0x2F5D50))
	got := buildComparisonView(snap, theme)
	if got == nil {
		t.Fatal("buildComparisonView returned nil for empty snapshot")
	}
}

func TestBuildComparisonView_DoesNotPanicOnRealComparisonData(t *testing.T) {
	snap := comparisonvm.New().Snapshot()
	theme := material3.New(widget.Hex(0x2F5D50))

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("buildComparisonView panicked: %v", r)
		}
	}()

	for i := 0; i < 5; i++ {
		_ = buildComparisonView(snap, theme)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -run TestBuildComparisonView -v`

Expected: FAIL with `undefined: buildComparisonView`.

- [ ] **Step 3: Implement `buildComparisonView` (GREEN)**

Create `cmd/fecim-lattice-tools-next/comparison_view.go`:

```go
//go:build !cgo

package main

import (
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/theme/material3"
	"github.com/gogpu/ui/widget"

	"fecim-lattice-tools/shared/viewmodel"
)

// buildComparisonView renders a comparison ModuleSnapshot into a gogpu/ui
// widget tree. Pure: same input → same widget tree, no side effects.
func buildComparisonView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
	descriptor := snapshot.Descriptor

	children := []widget.Widget{
		primitives.Text(descriptor.Title).FontSize(20).Bold(),
		primitives.Text(descriptor.Description).FontSize(13),
		primitives.Text(string(descriptor.ID) + " | " + descriptor.Status).FontSize(11),
	}

	for _, m := range snapshot.Metrics {
		children = append(children, primitives.Text(m.Label+": "+m.Value).FontSize(12))
	}

	for _, section := range snapshot.Sections {
		children = append(children, comparisonCard(section, theme))
	}

	return primitives.Box(children...).
		Padding(20).
		Gap(12).
		Background(theme.Colors.Surface)
}

func comparisonCard(section viewmodel.Section, theme *material3.Theme) widget.Widget {
	return primitives.Box(
		primitives.Text(section.Title).FontSize(16).Bold(),
		primitives.Text(section.Body).FontSize(13),
	).
		Padding(14).
		Gap(6).
		Background(theme.Colors.SurfaceContainer)
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -run TestBuildComparisonView -v`

Expected: PASS (all 3 tests).

- [ ] **Step 5: Commit**

```bash
git add cmd/fecim-lattice-tools-next/comparison_view.go cmd/fecim-lattice-tools-next/comparison_view_test.go
git commit -m "$(cat <<'EOF'
feat(next-shell): add buildComparisonView gogpu/ui adapter

Pure function: ModuleSnapshot → widget tree.
- Title + description + status line + count metric + per-architecture cards
- Uses primitives.Box / Text + material3 theme colors
- !cgo build tag (matches root_test.go convention for headless tests)

TDD evidence:
- RED: 3 tests failed with "undefined: buildComparisonView"
- GREEN: non-nil result; empty snapshot handled; no panic across repeated calls
- Verification: CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -run TestBuildComparisonView

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 6: Wire `buildRoot` to dispatch comparison ports through `buildComparisonView`

The shell's `buildRoot` currently runs every port through the generic `moduleCard`. Switch on `snapshot.Descriptor.ID` so the comparison module renders via `buildComparisonView`; all others keep the placeholder card.

**Files:**
- Modify: `cmd/fecim-lattice-tools-next/main.go`

**TDD: RED → GREEN.** The behavior under test ("`buildRoot` dispatches comparison ports differently from generic ports") is observable through the existing `TestBuildRootInstallsInHeadlessApp` smoke test — it must still install a non-nil root after the switch is added. We add a focused test that exercises the dispatch path explicitly.

- [ ] **Step 1: Write the failing test (RED)**

Append to `cmd/fecim-lattice-tools-next/root_test.go`:

```go
func TestBuildRoot_RendersWithRealComparisonPort(t *testing.T) {
	spec := DefaultAppSpec()
	ports := BuildPlaceholderPorts() // includes real comparison viewmodel after Task 4

	var foundComparison bool
	for _, p := range ports {
		if p.Descriptor().ID == "comparison" {
			foundComparison = true
			break
		}
	}
	if !foundComparison {
		t.Fatal("BuildPlaceholderPorts did not include a comparison port")
	}

	root := buildRoot(spec, ports, material3.New(widget.Hex(0x2F5D50)))
	if root == nil {
		t.Fatal("buildRoot returned nil")
	}

	app := uiapp.New()
	app.SetRoot(root)
	app.Frame()
	if app.Window().Root() == nil {
		t.Fatal("dispatch through buildComparisonView dropped the root widget")
	}
}
```

- [ ] **Step 2: Run test to verify the existing flow still works (this test will pass with the OLD buildRoot since comparison data already flows, but we want the dispatch path)**

Run: `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -run TestBuildRoot -v`

Expected before wiring change: PASS (both existing and new tests). The new test guards against future regressions: any future change that drops the comparison port or breaks `buildRoot` for it will fail the test. The dispatch-via-switch behavior itself is verified by `TestBuildComparisonView_*` (Task 5) plus the type/shape difference visible at runtime.

> Note on TDD framing: the *new* behavior here ("comparison port routes through buildComparisonView") is not unit-testable without widget-tree introspection that gogpu/ui doesn't expose. The smoke test above guarantees the integration compiles and renders. The comparison-specific rendering is fully tested in Task 5; this task is the wiring that connects them. Per `CLAUDE.md`: this is a GUI workflow change, so we add the smoke test first and then make the change.

- [ ] **Step 3: Modify `buildRoot` (GREEN)**

In `cmd/fecim-lattice-tools-next/main.go`, replace the existing port-loop inside `buildRoot` with a switch on `snapshot.Descriptor.ID`:

```go
func buildRoot(spec AppSpec, ports []viewmodel.ModulePort, theme *material3.Theme) widget.Widget {
	children := []widget.Widget{
		primitives.Text(spec.Title).FontSize(28).Bold(),
		primitives.Text("Future default gogpu/ui shell. Current module cards are placeholders until parity with the Fyne app is reached.").FontSize(15),
		primitives.Text("Stable fallback remains: go run ./cmd/fecim-lattice-tools").FontSize(13),
	}

	for _, port := range ports {
		snapshot := port.Snapshot()
		switch snapshot.Descriptor.ID {
		case viewmodel.ModuleComparison:
			children = append(children, buildComparisonView(snapshot, theme))
		default:
			children = append(children, moduleCard(snapshot, theme))
		}
	}

	return primitives.Box(children...).
		Padding(28).
		Gap(14).
		Background(theme.Colors.Surface)
}
```

The `moduleCard` function is unchanged. The switch grows by one `case` line per future module migration.

- [ ] **Step 4: Run tests to verify the wiring works**

Run: `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools-next/ -v`

Expected: PASS (all 4 tests: appmodel × 2, root × 2, comparison_view × 3).

- [ ] **Step 5: Run the full zero-CGO target to confirm no regressions across the shell**

Run: `make test-next-ui`

Expected: PASS (covers `shared/viewmodel/...` and `cmd/fecim-lattice-tools-next/...`).

- [ ] **Step 6: Commit**

```bash
git add cmd/fecim-lattice-tools-next/main.go cmd/fecim-lattice-tools-next/root_test.go
git commit -m "$(cat <<'EOF'
feat(next-shell): dispatch comparison ports through buildComparisonView

buildRoot now switches on snapshot.Descriptor.ID:
- ModuleComparison → buildComparisonView (rich card with per-architecture sections)
- default → moduleCard (placeholder kept for the other six modules)

Switch grows one case per future module migration.

TDD evidence:
- RED smoke test: TestBuildRoot_RendersWithRealComparisonPort added before wiring
- GREEN: all 4 root/view tests pass; make test-next-ui green
- Verification: make test-next-ui

Co-Authored-By: Claude Opus 4.7 (1M context) <noreply@anthropic.com>
EOF
)"
```

---

### Task 7: Final verification and PR

Confirm full repo health, run the architecture boundary check, do a manual visual smoke run, and open the PR.

- [ ] **Step 1: Run the full Go test suite**

Run: `go test ./...`

Expected: PASS (no regressions in any module — Fyne path untouched).

- [ ] **Step 2: Run vet**

Run: `go vet ./...`

Expected: clean.

- [ ] **Step 3: Run the zero-CGO subset**

Run: `make test-next-ui`

Expected: PASS.

- [ ] **Step 4: Run the architecture boundary check**

Run: `bash scripts/check-architecture.sh --fast`

Expected: Rules 1, 3, 4 PASS — confirms `shared/viewmodel/comparison/` does not import any UI package and `cmd/fecim-lattice-tools-next/` does not import Fyne.

- [ ] **Step 5: Manual visual smoke run**

Run: `CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next/`

Expected: window opens; the "FeCIM Comparison" card displays:
- Title "FeCIM Comparison"
- Status line `comparison | functional`
- Metric line "Architectures compared: 3"
- Three architecture cards: "Traditional CPU+DRAM", "GPU Accelerator", "FeCIM CIM" — each with technology, process node, area, TDP, TOPS, TOPS/W, memory
- "FeCIM CIM" body explicitly contains "(estimated; not validated)"
- The other six modules still show the generic placeholder card

Close the window when verified.

- [ ] **Step 6: Push and open PR**

```bash
git push -u origin feat/m5-comparison-gogpu-migrate

gh pr create --title "feat(next-shell): pilot M5 comparison migration to gogpu/ui" --body "$(cat <<'EOF'
## Summary

- Pilot migration replacing the `ModuleComparison` placeholder in `cmd/fecim-lattice-tools-next/` with a real read-only viewmodel surfacing `module5-comparison/pkg/comparison` data.
- Establishes the canonical pattern (UI-neutral viewmodel sub-package + thin gogpu/ui adapter + `buildRoot` switch) that the remaining six modules will follow.
- Both shells coexist; legacy Fyne path (`cmd/fecim-lattice-tools`) is unchanged.

## Changes

- `module5-comparison/pkg/comparison`: add `Architectures()` canonical accessor.
- `shared/viewmodel/comparison/`: new sub-package — `Module` (implements `ModulePort`) + pure `buildSnapshot`. Zero UI deps.
- `cmd/fecim-lattice-tools-next/`: inject real comparison module via `BuildPlaceholderPorts`, dispatch through `buildComparisonView` from `buildRoot`'s switch.
- FeCIM section body explicitly flags `IsEstimated=true` per honesty-audit policy.

## TDD evidence

| Task | RED | GREEN |
|------|-----|-------|
| T1 `Architectures()` accessor | undefined symbol | 2 tests (canonical set + fresh-slice-per-call) |
| T2 `buildSnapshot` | undefined symbol | 7 tests (descriptor, sections, body fields, FeCIM estimated, count metric, determinism, empty) |
| T3 `Module` struct | undefined symbol | 5 tests (descriptor parity, ApplyAction, Start/Stop idempotency, ModulePort interface) |
| T4 `BuildPlaceholderPorts` | port reported `StatusPlaceholder` | functional comparison port with sections + metrics |
| T5 `buildComparisonView` | undefined symbol | 3 tests (non-nil, empty snapshot, no-panic) |
| T6 `buildRoot` dispatch | smoke test added | full zero-CGO suite green |

## Verification

- `go test ./...` — passes
- `go vet ./...` — clean
- `make test-next-ui` — passes
- `bash scripts/check-architecture.sh --fast` — Rules 1/3/4 green
- Manual: `CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next` — comparison card shows three architecture cards with full physical/performance summaries; FeCIM body flags estimated values; other six modules retain placeholder card.

## Test plan

- [ ] CI green (Go test, vet, arch-check)
- [ ] Reviewer manually runs `CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next` and confirms the comparison card renders with 3 architecture sub-cards and the count metric.
- [ ] Reviewer confirms `cmd/fecim-lattice-tools` (legacy Fyne shell) is unchanged.

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

- [ ] **Step 7: Verify PR opened**

Run: `gh pr view --json url,number,title`

Expected: JSON with the new PR URL. Capture the URL for the user.

---

## Self-Review Checklist (writing-plans skill)

Performed inline as the final step of plan authoring.

**1. Spec coverage:**
- ✅ "Replace `ModuleComparison` placeholder with real viewmodel" → T2-T4
- ✅ "Establish canonical viewmodel ↔ gogpu/ui rendering pattern" → T2 (viewmodel) + T5 (adapter) + T6 (dispatch)
- ✅ "Read-only MVP, no actions" → `ApplyAction` returns `ErrUnsupportedAction` (T3)
- ✅ "No `Start()/Stop()` lifecycle" → no-op `Start`/`Stop` (T3)
- ✅ "Sections-by-name + summary line" → T2 `buildSnapshot`
- ✅ "Single `count` metric" → T2 `buildSnapshot`
- ✅ "FeCIM section flags estimated values" → T2 test + impl
- ✅ "Switch-in-shell, not port-defined render" → T6 dispatch via `snapshot.Descriptor.ID`
- ✅ "Boundary rule: viewmodel does NOT import gogpu/ui or Fyne" → arch-check at T7-S4
- ✅ "Both shells coexist" → no changes to `module5-comparison/pkg/gui/` or `cmd/fecim-lattice-tools/`
- ✅ Spec's "Open Question for Implementation Pass" (canonical accessor name) resolved → T1 adds `Architectures()`; struct field names from real `architecture.go` used in T2

**2. Placeholder scan:**
- No "TBD", "TODO", "implement later", or vague "appropriate error handling" instructions.
- Every code step shows the actual code.
- Every command step shows the actual command and expected outcome.

**3. Type consistency:**
- `comparisonvm` alias used consistently in T4 (the only file importing both packages).
- `*pkg.Architecture` used consistently — matches real package (`TraditionalCPU()` returns `*Architecture`, not `Architecture`).
- `viewmodel.ErrUnsupportedAction` matches the existing export in `shared/viewmodel/static_module.go`.
- `viewmodel.StatusFunctional` constant used in both T2 (snapshot) and T3/T4 (descriptor) — same string literal `"functional"`.
- Section body field names (`Technology`, `ProcessNode`, `ChipArea`, `TDP`, `PeakTOPS`, `TOPSPerWatt`, `MemorySize`, `MemoryBW`, `IsEstimated`) all confirmed against `module5-comparison/pkg/comparison/architecture.go`.

No issues found.

---

## Execution Handoff

Plan complete and saved. Two execution options:

1. **Subagent-Driven (recommended)** — fresh subagent per task with two-stage review (spec compliance → code quality), fast iteration in this session.
2. **Inline Execution** — execute tasks directly in this session using `superpowers:executing-plans`, with checkpoints between tasks.

Which approach?
