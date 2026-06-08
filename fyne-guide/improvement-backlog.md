# FeCIM Fyne improvement backlog

Evidence from `generated/project-fyne-usage.json`:

- Fyne is already central: 1065 Go files import Fyne packages.
- Heavy direct UI updates: `fyne.Do` 1352 occurrences and `fyne.DoAndWait` 360 occurrences.
- Heavy custom widget surface: `widget.BaseWidget` token count 455.
- Low data binding use: `binding.` only 4 occurrences.
- Virtualized collection use is present but limited relative to data-heavy UI: `widget.NewList` 28, `widget.NewTable` 12, `widget.NewTree` 4, `widget.NewGridWrap` 0.
- `widget.NewToolbar`, `widget.NewActivity`, `container.NewDocTabs` were not detected.

## P0 — Upgrade and stabilize Fyne patch version

**Goal:** move from `v2.7.2` to `v2.7.4`.

Why: patch fixes hit our exact surfaces: Tree, RichText, Accordion, text wrapping/rendering, progress, launch panic, menu crash, system tray.

TDD slice:

1. Add focused GUI tests for docs tree/RichText, accordion views, and progress indicators.
2. Run current tests and record RED/coverage evidence.
3. Upgrade `go get fyne.io/fyne/v2@v2.7.4 && go mod tidy`.
4. Run targeted tests plus `go test ./...`.
5. Accept/update visual goldens only after review.

## P1 — Introduce a binding-backed viewmodel layer for controls

**Problem:** Direct widget mutation and ad-hoc refreshes make complex simulations hard to reason about.

**Use Fyne:** `data/binding`, `New*WithData`, `BindPreference*`, `NewSprintf`, `DataListener`.

Candidate targets:

- Material/device parameter controls.
- Run status/progress labels.
- Selected module and selected material/device presets.
- Export path/options.
- Read-only metrics panels that currently call many `SetText`/`Refresh` operations.

Expected payoff:

- Fewer goroutine/UI-thread bugs.
- Easier tests: assert bound values rather than walking every widget.
- Less manual `fyne.Do` fan-out.

## P1 — Virtualize large result views

**Problem:** FeCIM naturally displays large matrices, calibration tables, logs, and docs indexes. Static widget grids do not scale.

**Use Fyne:**

- `widget.Table` / `NewTableWithHeaders` for metrics matrices, crossbar arrays, truth tables, and EDA data.
- `widget.ListWithData` for simulation log streams and result lists.
- `widget.TreeWithData` for docs, EDA file trees, and module output artifacts.
- `widget.GridWrap` for screenshot/reference gallery, materials, or generated artifact cards.

Implementation notes:

- Follow official demo pattern: `Length`, `CreateItem`, `UpdateItem` callbacks.
- Keep cell templates small and reused.
- Use `SetColumnWidth`, `SetRowHeight` only for exceptional cells.
- Keep model data immutable or guarded; update via binding/list refresh contracts.

## P1 — Standardize scientific visualization widgets

**Problem:** The repo has many custom renderers and rasters; each can drift in theme behavior, resize behavior, and refresh cost.

**Use Fyne:** `canvas.NewRasterWithPixels`, `canvas.Image`, vector primitives, cached `WidgetRenderer`, `theme.Color`, `Settings().AddListener`.

Create or formalize shared helpers for:

- Theme-aware plot palette: background, axes, grid, warning/success/error colors.
- Renderer lifecycle checklist: cache renderer, don't rebuild object trees in `Refresh`, implement `Destroy`, avoid goroutine leaks.
- Raster invalidation: recompute data image off UI thread, assign/refresh on UI thread.
- Export path: render plot data to image/SVG/CSV consistently.

Good references:

- Official examples `clock`: theme-aware custom canvas/layout.
- Official examples `fractal`: `NewRasterWithPixels` for computed visualization.
- Fyne demo `collection.go`: virtualized views for large datasets.

## P2 — Add first-class toolbars and command model

**Problem:** Actions are scattered across buttons, cards, dialogs, and menus.

**Use Fyne:** `widget.Toolbar`, `fyne.MainMenu`, `fyne.Shortcut`, `MenuItem`, `Button.Importance`.

Recommended pattern:

- Define module command structs: ID, label, icon, shortcut, enabled binding, handler.
- Render the same commands into toolbar, menu, buttons, and keyboard shortcuts.
- Keep destructive actions as `DangerImportance`/confirm dialogs.

Candidate commands:

- Run simulation, stop/cancel, reset, export, import preset, open docs, copy results, save screenshot.

## P2 — Improve long-running task UX

**Problem:** Simulations, validation, and EDA export can take time.

**Use Fyne:** `ProgressBarWithData`, `ProgressBarInfinite`, `Activity`, `dialog.NewProgress`, notifications.

Targets:

- EDA export progress with phases.
- ISPP/MNIST batch run progress.
- Background docs/search indexing.
- Non-blocking cancelable dialogs.

Design rule:

- Background goroutines do compute/I/O only.
- UI updates go through `fyne.Do`/bindings.
- Use `DoAndWait` only when required for screenshot/golden synchronization.

## P2 — Package like a polished desktop app

**Use Fyne docs:** packaging, metadata, distribution, cross-compiling.

Actions:

- Verify app ID/name/version/icon metadata.
- Add release commands around `fyne package` where appropriate.
- Bundle docs/icons/resources with `fyne bundle` or `fyne.Resource`.
- Consider web demo build only after dependencies and rendering path are audited.

## P2 — Use preferences intentionally

**Use Fyne:** `App.Preferences`, `binding.BindPreference*`.

Persist:

- Last selected module.
- Last selected material/device presets.
- Window size/layout split offsets.
- Theme or high-contrast mode.
- Recent export directories and docs favorites.

## P3 — App store/gallery learnings

From `apps-catalog.md`, recurring successful Fyne patterns:

- Cross-platform single-purpose GUI with clear workflows.
- Local-first/privacy-first processing.
- File drag/drop or file dialogs.
- Tray/menu-bar utility modes.
- Progress-rich long-running tasks.
- Good install metadata and screenshots.

FeCIM can borrow this by making each module feel like a focused tool while still sharing a unified shell.

## Guardrails

- Apply repo TDD rule for every code change.
- Keep Fyne widget mutations on the Fyne UI goroutine using `fyne.Do`/bindings.
- Treat screenshot/markup changes as reviewable UI changes, not automatic golden churn.
- Avoid rebuilding widget trees on every update; prefer cached widgets, bindings, or collection refreshes.
