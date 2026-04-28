# M5 Comparison → gogpu/ui Pilot Migration — Design

**Date:** 2026-04-28
**Status:** Spec (awaiting implementation plan)
**Skill applied:** `fecim-gogpu-migrate` (single-module pilot scope)

## Goal

Replace the `ModuleComparison` placeholder in the gogpu/ui shell (`cmd/fecim-lattice-tools-next/`) with a real, read-only viewmodel that surfaces `module5-comparison/pkg/comparison` data. Establish the canonical viewmodel ↔ gogpu/ui rendering pattern that the other six modules will follow in subsequent PRs.

## Non-Goals

- Action plumbing (scenario picker, architecture selector). Read-only MVP only.
- `Start()/Stop()` simulation lifecycle. Comparison data is static; no live updates.
- Migrating any of the other six modules (hysteresis, crossbar, mnist, circuits, eda, docs).
- Hero / market / sensitivity / fabrication-reality / liveslide panels from `module5-comparison/pkg/gui/`. Read-only summary card replaces the placeholder; richer panels come incrementally.
- Renaming `BuildPlaceholderPorts` → `BuildAppPorts` (kept to minimize blast radius; rename in a follow-up if/when a second module migrates).
- Changes to `module5-comparison/pkg/gui/`. Legacy Fyne path keeps working as-is. Both shells coexist.

## Component Architecture

```
shared/viewmodel/
├── types.go                          # (existing) ModulePort, Snapshot, Section, etc.
├── static_module.go                  # (existing) StaticModule
└── comparison/                       # NEW package
    ├── viewmodel.go                  # Module — implements ModulePort
    ├── snapshot.go                   # buildSnapshot — pure function: archs → ModuleSnapshot
    └── viewmodel_test.go             # zero-CGO unit tests

cmd/fecim-lattice-tools-next/
├── main.go                           # MODIFY: buildRoot dispatches comparison ports to buildComparisonView
├── appmodel.go                       # MODIFY: BuildPlaceholderPorts injects real comparison module
├── comparison_view.go                # NEW: gogpu/ui adapter
└── comparison_view_test.go           # NEW: smoke test for buildComparisonView
```

### Boundary rules (per `AGENTS.md` UI boundary)

- `shared/viewmodel/comparison/` imports **only** `shared/viewmodel` and `module5-comparison/pkg/comparison`. Zero `fyne.io/...` or `github.com/gogpu/ui`.
- `cmd/fecim-lattice-tools-next/comparison_view.go` imports `github.com/gogpu/ui/...` and `shared/viewmodel`. Allowed (it's a shell package).
- `module5-comparison/pkg/comparison/` is unchanged. Already UI-neutral.
- `module5-comparison/pkg/gui/` is unchanged.

### Why a sub-package instead of `shared/viewmodel/comparison.go` flat next to `static_module.go`

The six future viewmodels each grow their own state/events/snapshot logic. Sub-packaging now avoids a flat 30+ file `shared/viewmodel/` later and lets per-module tests live next to their viewmodel.

## Viewmodel Design

### `shared/viewmodel/comparison/viewmodel.go`

```go
package comparison

import (
    "fecim-lattice-tools/module5-comparison/pkg/comparison"
    "fecim-lattice-tools/shared/viewmodel"
)

// Module implements viewmodel.ModulePort using pkg/comparison data.
// Read-only MVP — ApplyAction returns ErrUnsupportedAction.
type Module struct {
    architectures []comparison.Architecture
}

// New constructs the comparison viewmodel from the canonical architecture set.
func New() *Module {
    return &Module{architectures: comparison.Architectures()}
}

func (m *Module) Descriptor() viewmodel.ModuleDescriptor {
    return viewmodel.ModuleDescriptor{
        ID:          viewmodel.ModuleComparison,
        Title:       "FeCIM Comparison",
        Description: "Evidence-first technology comparison and scenario analysis.",
        Status:      viewmodel.StatusFunctional,
    }
}

func (m *Module) Snapshot() viewmodel.ModuleSnapshot { return buildSnapshot(m.architectures) }
func (m *Module) ApplyAction(viewmodel.Action) error { return viewmodel.ErrUnsupportedAction }
func (m *Module) Start()                             {}
func (m *Module) Stop()                              {}
```

### `shared/viewmodel/comparison/snapshot.go`

```go
package comparison

import (
    "fmt"
    "time"

    pkg "fecim-lattice-tools/module5-comparison/pkg/comparison"
    "fecim-lattice-tools/shared/viewmodel"
)

// buildSnapshot converts the canonical architecture set into a ModuleSnapshot
// with one Section per architecture and one Metric for the count.
func buildSnapshot(archs []pkg.Architecture) viewmodel.ModuleSnapshot {
    sections := make([]viewmodel.Section, 0, len(archs))
    for _, a := range archs {
        sections = append(sections, viewmodel.Section{
            ID:    string(a.Name),
            Title: a.DisplayName,
            Body: fmt.Sprintf(
                "Energy: %s\nDensity: %s\nEndurance: %s\nRetention: %s\nSpeed: %s\nMaturity: %s",
                a.EnergyPerOp, a.Density, a.Endurance, a.Retention, a.Speed, a.Maturity,
            ),
        })
    }
    metrics := []viewmodel.Metric{
        {ID: "count", Label: "Architectures compared", Value: fmt.Sprintf("%d", len(archs)), Confidence: "deterministic"},
    }
    return viewmodel.ModuleSnapshot{
        Descriptor: viewmodel.ModuleDescriptor{
            ID:          viewmodel.ModuleComparison,
            Title:       "FeCIM Comparison",
            Description: "Evidence-first technology comparison and scenario analysis.",
            Status:      viewmodel.StatusFunctional,
        },
        Metrics:   metrics,
        Sections:  sections,
        UpdatedAt: time.Time{}, // zero time = static snapshot, deterministic in tests
    }
}
```

### Shape decisions

- **Sections per architecture** — each architecture gets one card; 1:1 mapping; no information loss.
- **`Body` is a single `\n`-joined string** — `viewmodel.Section.Body` is `string`. A structured-fields refactor is a separate viewmodel-types PR (YAGNI for now).
- **Single `count` metric** — proves the metrics surface without inventing comparison-specific metrics. Real metrics come in follow-up PRs.
- **`UpdatedAt: time.Time{}`** zero value — explicit "static." Avoids non-determinism in tests.

### Caveat: `pkg/comparison` API names are placeholders

The field names (`Architectures()`, `Architecture.Name`, `EnergyPerOp`, `Density`, etc.) are the design's *expected shape*. The implementation pass reads `module5-comparison/pkg/comparison/architecture.go` first and maps to actual fields. The viewmodel boundary (Sections-by-name + summary-line, count Metric) is what's locked in here.

## gogpu/ui Adapter

### `cmd/fecim-lattice-tools-next/comparison_view.go`

```go
package main

import (
    "github.com/gogpu/ui/primitives"
    "github.com/gogpu/ui/theme/material3"
    "github.com/gogpu/ui/widget"

    "fecim-lattice-tools/shared/viewmodel"
)

// buildComparisonView renders the comparison module's snapshot as a gogpu/ui
// widget tree. Pure: same input → same widget tree, no side effects.
func buildComparisonView(snapshot viewmodel.ModuleSnapshot, theme *material3.Theme) widget.Widget {
    children := []widget.Widget{
        primitives.Text(snapshot.Descriptor.Title).FontSize(20).Bold(),
        primitives.Text(snapshot.Descriptor.Description).FontSize(13),
    }

    for _, m := range snapshot.Metrics {
        children = append(children, primitives.Text(m.Label+": "+m.Value).FontSize(12))
    }

    for _, s := range snapshot.Sections {
        children = append(children, comparisonCard(s, theme))
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

### Wiring `cmd/fecim-lattice-tools-next/main.go`

In `buildRoot`, the existing port loop becomes:

```go
for _, port := range ports {
    snapshot := port.Snapshot()
    switch snapshot.Descriptor.ID {
    case viewmodel.ModuleComparison:
        children = append(children, buildComparisonView(snapshot, theme))
    default:
        children = append(children, moduleCard(snapshot, theme))
    }
}
```

The other 6 modules continue rendering via the generic `moduleCard` placeholder. As future PRs land, the switch grows one case at a time.

### Wiring `cmd/fecim-lattice-tools-next/appmodel.go`

```go
import comparisonvm "fecim-lattice-tools/shared/viewmodel/comparison"

func BuildPlaceholderPorts() []viewmodel.ModulePort {
    descriptors := viewmodel.KnownDescriptors()
    ports := make([]viewmodel.ModulePort, 0, len(descriptors))
    for _, descriptor := range descriptors {
        if descriptor.ID == viewmodel.ModuleComparison {
            ports = append(ports, comparisonvm.New())
            continue
        }
        ports = append(ports, viewmodel.NewStaticModule(descriptor, []viewmodel.Section{
            {ID: "migration-status", Title: "Migration Status",
                Body: "This module is represented by a UI-neutral placeholder while the gogpu/ui shell reaches parity with the current Fyne implementation."},
        }))
    }
    return ports
}
```

### Why a switch in `buildRoot` instead of a port-defined `Render`

Two options were considered:
1. **Port-defined render** — each port exposes `Render(theme) widget.Widget`. UI code is one line.
2. **Switch in shell** *(chosen)* — shell knows how to render each module type via switch on `ModuleID`.

Option 1 forces the viewmodel package to import `gogpu/ui/widget`, which violates the UI-neutrality rule. Option 2 keeps the boundary intact: viewmodels stay UI-neutral; the shell is the only place that knows about widgets.

Trade-off: the switch grows linearly with module count. Move to a `view_dispatch.go` file when it crosses ~150 lines. Not now (≤ 50 lines for all 7 + default).

## Testing

All tests are zero-CGO, run under `make test-next-ui`. TDD-RED first per CLAUDE.md hard-rule.

### `shared/viewmodel/comparison/viewmodel_test.go`

| Test | Asserts |
|---|---|
| `TestModule_Descriptor_ID` | Descriptor returns `viewmodel.ModuleComparison` |
| `TestModule_Descriptor_Status` | Returns `StatusFunctional` (not `StatusPlaceholder`) |
| `TestModule_Snapshot_HasOneSectionPerArchitecture` | `len(snap.Sections) == len(comparison.Architectures())` |
| `TestModule_Snapshot_SectionTitlesMatchArchitectures` | Section titles equal each architecture's display name (set comparison, order-independent) |
| `TestModule_Snapshot_HasCountMetric` | First Metric has ID `count` and Value matches architecture count |
| `TestModule_ApplyAction_ReturnsErrUnsupported` | `ApplyAction(...) == viewmodel.ErrUnsupportedAction` |
| `TestModule_StartStop_AreNoOps` | `Start()`/`Stop()` don't panic, are idempotent |
| `TestBuildSnapshot_DeterministicForSameInput` | Two calls with the same architectures slice yield byte-equal snapshots |

### `cmd/fecim-lattice-tools-next/comparison_view_test.go`

| Test | Asserts |
|---|---|
| `TestBuildComparisonView_ReturnsNonNil` | Returned widget is not nil |
| `TestBuildComparisonView_RendersAllSections` | Widget tree text content includes every section title |
| `TestBuildComparisonView_RendersMetric` | Tree contains the count metric label and value |

If `gogpu/ui` doesn't expose a widget-tree text walker, the second/third tests downgrade to type-shape assertions (returns a `*primitives.BoxWidget`, contains the expected number of children).

### `cmd/fecim-lattice-tools-next/appmodel_test.go` (extend)

| Test | Asserts |
|---|---|
| `TestBuildPlaceholderPorts_ComparisonIsFunctional` | The port with `Descriptor().ID == ModuleComparison` reports `Status == StatusFunctional` |

### Not tested (deliberately)

- **Visual rendering** — gogpu/ui doesn't paint to a framebuffer in unit tests. Smoke-render via `app.Frame()` requires a GPU context. Manual run + future screenshot tests cover this.
- **Architecture data correctness** — `pkg/comparison` already has its own tests. Viewmodel only re-shapes existing data.
- **Fyne side** — untouched.

### Verification commands

- `make test-next-ui` — zero-CGO subset, must pass.
- `make test` (or `go test ./...`) — full suite, must pass (no regression to legacy Fyne path).
- `go vet ./...` — must pass.
- `bash scripts/check-architecture.sh --fast` — Rules 1, 3, 4 must stay green.
- Manual: `CGO_ENABLED=0 go run ./cmd/fecim-lattice-tools-next` and verify the Comparison card now shows the architectures with their summary lines instead of the placeholder text.

## Rollout

1. Single PR off local `main` on branch `feat/m5-comparison-gogpu-migrate`.
2. ~3-6 commits using TDD discipline (one per task in the implementation plan).
3. CI gates: `go vet`, `go test ./...`, `make test-next-ui`, `scripts/check-architecture.sh`. (Skills sync check applies if PR #5 has merged first; otherwise unaffected.)
4. PR description includes manual validation steps.
5. No deprecation. Both shells coexist.

## Open Question for Implementation Pass

`pkg/comparison`'s exact API: the canonical accessor name (`Architectures()` / `AllArchitectures()` / per-arch constructors) and the actual fields on `Architecture`. The implementer reads `module5-comparison/pkg/comparison/architecture.go` first and adapts the snapshot builder. The viewmodel *shape* (Sections-by-name + summary-line, count Metric) is locked; field names are fill-in.
