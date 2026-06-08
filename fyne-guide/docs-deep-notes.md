# Deep Fyne docs notes for FeCIM

This is the human-readable digest of the most relevant docs.fyne.io pages for improving FeCIM. See `docs-map.md` and `generated/docs-index.json` for the complete index.

## Runtime and goroutines

Sources: `started/goroutines`, `started/apprun`, local `go doc fyne.io/fyne/v2`.

Key facts:

- Since Fyne `v2.6.0`, Fyne events and callbacks happen on one Fyne event goroutine.
- Any app-created goroutine that calls Fyne APIs must queue UI work through `fyne.Do(func(){ ... })`.
- Use `fyne.DoAndWait(func(){ ... })` when the caller must wait until the UI work is complete, for example screenshot/golden synchronization or image swap correctness.
- `fyne.Do` work is scheduled before the next frame; docs mention this can be up to roughly one frame later.

FeCIM rules:

- Simulation, EDA export, validation, downloads, file I/O, and batch inference should run outside the UI goroutine.
- UI updates from those jobs should go through a small module-local adapter or binding update path, not scattered direct widget mutation.
- `DoAndWait` should be rare and documented because it can serialize tests or runtime paths.

Pattern:

```go
go func() {
    result := runSimulation(snapshot)
    fyne.Do(func() {
        status.SetText(result.Summary)
        plot.SetData(result.Series)
        plot.Refresh()
    })
}()
```

Better long-term pattern:

```go
// compute goroutine updates model/binding; widgets are bound once at construction.
progress.Set(0.75)
statusText.Set("VERIFY pulse 12/16")
```

## Widgets: behavior/state separate from rendering

Sources: `architecture/widgets`, `extend/custom-widget`.

Key facts:

- Standard Fyne widgets expose behavior/state, not detailed look-and-feel.
- Rendering is separated through `fyne.WidgetRenderer`.
- `widget.BaseWidget` reduces boilerplate for custom widgets.
- A custom widget should define behavior/state; its renderer owns canvas objects and layout.
- Renderer interface: `Layout(Size)`, `MinSize() Size`, `Refresh()`, `Objects() []CanvasObject`, `Destroy()`.

FeCIM rules:

- Custom scientific widgets should cache renderer objects; do not rebuild large object trees in `Refresh`.
- State updates should set fields/model data and call `Refresh`, while renderers update existing primitives.
- `Destroy` should stop timers/listeners/goroutines if the widget started any.
- Theme-aware widgets should refresh colors from `theme.Color(...)` or `theme.CurrentForWidget(...)`.

Checklist for a custom FeCIM widget:

- `ExtendBaseWidget` is called in the constructor.
- Public fields represent behavior/state, not internal canvas objects.
- Renderer stores all canvas primitives.
- `MinSize` is cheap and deterministic.
- `Refresh` updates cached objects and calls `canvas.Refresh` only where needed.
- No goroutine leaks; timers have a stop path.
- Tests cover MinSize, renderer object count, theme refresh, and representative interaction.

## Layout and scaling

Sources: `architecture/geometry`, `architecture/scaling`, `extend/custom-layout`, `container/*`, `layout/*`.

Key facts:

- Fyne uses device-independent units and vector graphics.
- Auto-scaling adapts to OS/display DPI and window moves between screens.
- Users can override scale with `FYNE_SCALE` or settings.
- Layouts compute object positions/sizes; direct `Move`/`Resize` only sticks for objects outside layout or inside custom layouts.
- Custom layouts implement `Layout([]CanvasObject, Size)` and `MinSize([]CanvasObject) Size`.

FeCIM rules:

- Avoid pixel-perfect assumptions. Test at desktop and mobile-ish sizes.
- Use `container.NewBorder`, `NewHSplit`, `NewVSplit`, `NewAdaptiveGrid`, and scroll containers before custom layout.
- Use custom layout only for domain-specific visual arrangements: circuit overlays, plot layers, material comparison diagrams.
- Do not rely on fixed pixel dimensions for plot labels or control panels.

## Collections for large data

Sources: `collection/list`, `collection/table`, `collection/tree`, `cmd/fyne_demo/tutorials/collection.go`.

Key facts:

- `widget.List`, `widget.Table`, `widget.Tree`, and `widget.GridWrap` are virtualized/cached collection widgets.
- They do not embed every data item as a widget. They ask for length/children and reuse template objects.
- `List` callbacks: length, create template, update template.
- `Table` callbacks: row/col length, create cell template, update cell.
- `Tree` callbacks: child IDs, branch check, create template, update template.
- Tree node IDs must be unique; for file-like data, use full paths, not display names.

FeCIM targets:

- Crossbar matrices and per-cell conductance/state views.
- Validation results and benchmark tables.
- Simulation logs and event streams.
- Documentation trees and search results.
- EDA artifact browsers.
- Screenshot/reference galleries via `GridWrap`.

Rules:

- `CreateItem`/`CreateCell` should allocate a small reusable template.
- `UpdateItem`/`UpdateCell` should only mutate that template for the given row/cell.
- Keep data access fast and safe; precompute expensive summaries outside callbacks.
- Avoid building `[]fyne.CanvasObject` with thousands of children.

## Data binding

Sources: `explore/binding`, `binding/*`, `cmd/fyne_demo/tutorials/bind.go`, `go doc data/binding`.

Key facts:

- Bindings connect widgets to data sources that change over time.
- Many widgets have `New...WithData` constructors.
- Bindings can be internal (`binding.NewString`) or external (`binding.BindString(&value)`).
- Converters exist: `FloatToStringWithFormat`, `IntToString`, `BoolToString`, `NewSprintf`, etc.
- Collection bindings support lists/trees.
- Preferences can be bound using `BindPreference*`.

FeCIM opportunity:

- Replace manual `SetText` clusters with bound labels.
- Bind sliders/entries/selects to viewmodel state.
- Bind run progress to `ProgressBarWithData`.
- Bind selected material/module/export options to preferences.

Preferred pattern:

```go
status := binding.NewString()
progress := binding.NewFloat()

statusLabel := widget.NewLabelWithData(status)
bar := widget.NewProgressBarWithData(progress)

// background job
progress.Set(0.4)
status.Set("Solving IR drop...")
```

## Validation and forms

Sources: `widget/form`, `data/validation`, `dialog/form`.

Useful APIs:

- `widget.NewForm`, `widget.NewFormItem`.
- `Entry.Validator` and `fyne.StringValidator`.
- `validation.NewRegexp`, `validation.NewTime`, `validation.NewAllStrings`.
- `dialog.ShowForm` for modal validated input.

FeCIM use:

- Validate numeric material parameters before simulation.
- Validate EDA export names/paths/options.
- Mark educational default values as simulation assumptions.
- Put advanced controls into forms with validation rather than loose entries.

## Dialogs, files, storage

Sources: `explore/dialogs`, `dialog`, `storage`.

Useful APIs:

- `dialog.ShowInformation`, `ShowError`, `ShowConfirm`.
- `dialog.NewCustom`, `NewCustomConfirm`, `NewCustomWithoutButtons`.
- `dialog.NewProgress`, `NewProgressInfinite`.
- `ShowFileOpen`, `ShowFileSave`, `ShowFolderOpen`.
- `storage.NewFileURI`, `Reader`, `Writer`, `List`, `Copy`, filters.

FeCIM use:

- Export SPICE/Verilog/Liberty/plots through file save dialogs.
- Open calibration datasets with file filters.
- Show equation/explanation dialogs consistently.
- Use progress dialogs for EDA and long simulation flows.

## Shortcuts, menus, toolbars

Sources: `explore/shortcuts`, `fyne.Menu`, `widget.Toolbar`.

Key facts:

- Shortcuts can be registered on a canvas and reused in menu items.
- Desktop custom shortcuts use `driver/desktop.CustomShortcut`.
- Standard shortcuts exist for copy/cut/paste/undo/redo/select-all.
- Menus and popup menus share `fyne.Menu`/`MenuItem` structures.

FeCIM opportunity:

- Centralize module commands: run, stop, reset, export, open docs, search, copy result.
- Render the same command into toolbar, menu item, button, and shortcut.
- Keep one source of truth for enabled/disabled state.

## Preferences and metadata

Sources: `explore/preferences`, `started/metadata`, `started/packaging`.

Key facts:

- Preferences require a stable app ID via `app.NewWithID("reverse.domain.app")`.
- Preferences support strings, bools, ints, floats, and lists.
- Metadata can live in `FyneApp.toml` and reduce packaging command flags.
- `fyne package` creates desktop bundles/packages for macOS/Linux/Windows.

FeCIM use:

- Persist window layout, split offsets, selected module, selected material, theme/accessibility mode, and recent export paths.
- Add/verify app metadata before release packaging.
- Use bundled resources for icons/docs/assets that must ship with the app.

## Web/mobile/distribution

Sources: `started/webapp`, `started/mobile`, `started/cross-compiling`, `started/distribution`.

Key facts:

- Fyne can run in browsers through WebAssembly using the `fyne` CLI.
- Fyne can package for mobile and desktop platforms.
- Web driver support is documented as supported but not necessarily feature-complete for every app capability.

FeCIM opportunity:

- Desktop remains primary.
- A web/teaching demo may be possible after auditing CGO, file dialogs, Vulkan/GLFW, and package-specific assumptions.
- Tablet/mobile educational viewers may be possible for selected modules if UI complexity is reduced.

## Testing graphical apps

Sources: `started/testing`, `fyne.io/fyne/v2/test`.

Useful APIs:

- `test.NewTempApp`, `test.NewTempWindow`, `test.NewWindow`.
- `test.Tap`, `TapAt`, `Type`, `Drag`, `Scroll`, `FocusNext`, `FocusPrevious`.
- `test.AssertRendersToMarkup`, `AssertRendersToImage`, `RenderObjectToMarkup`.
- `test.WithTestTheme`, `ApplyTheme`.

FeCIM rules:

- Every behavior/UI change starts with a focused failing test per repo policy.
- Prefer unit-level widget tests for command state, validation, bindings, and render contracts.
- Use markup/image goldens for layout regression, but review visual changes before updating goldens.
- Use `fyne.DoAndWait` only when a test must synchronize with the UI queue.
