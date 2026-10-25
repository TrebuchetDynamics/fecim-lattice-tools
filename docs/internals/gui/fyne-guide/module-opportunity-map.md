# FeCIM module-by-module Fyne opportunity map

This file translates the Fyne research into concrete UI improvement directions for this repository. It is intentionally documentation-only; each code change still needs TDD.

## Shared shell / `cmd/fecim-lattice-tools*`

Current opportunity:

- Build a command model once and render it as menu, toolbar, buttons, and shortcuts.
- Persist window size, split offsets, last module, and theme via `app.NewWithID` + preferences/bindings.
- Add stable app metadata and packaging commands.
- Standardize long-running job notifications and progress surfaces.

Fyne features:

- `fyne.MainMenu`, `fyne.MenuItem`, `widget.Toolbar`, `driver/desktop.CustomShortcut`.
- `App.Preferences`, `binding.BindPreference*`.
- `fyne.NewNotification`, `dialog.NewProgress`, `widget.ProgressBarWithData`.
- `FyneApp.toml`, `fyne package`.

Suggested TDD slices:

1. Test that command IDs render consistently in menu + toolbar + shortcuts.
2. Test preferences restore the last selected module and split offset.
3. Test a fake long-running command updates progress and sends completion notification.

## Shared UI infrastructure / `shared/widgets`, `shared/themes`, `shared/viewmodel`

Current opportunity:

- Add binding-first viewmodel helpers.
- Add shared virtualized table/list adapters.
- Add shared custom-renderer checklist/helpers.
- Add semantic plot palette tied to Fyne theme.

Fyne features:

- `data/binding`, `widget.New...WithData`, `binding.NewSprintf`.
- `widget.Table`, `widget.ListWithData`, `widget.TreeWithData`.
- `widget.BaseWidget`, `fyne.WidgetRenderer`.
- `theme.Color`, `theme.CurrentForWidget`, settings listeners.

Suggested TDD slices:

1. Add tests for a binding-backed `RunStatus` viewmodel.
2. Add a tiny virtualized `MetricTable` adapter and tests for update callbacks.
3. Add renderer lifecycle tests for shared custom widget base utilities.

## Module 1 — Hysteresis

Current opportunity:

- Improve P-E plot renderer consistency and theme responsiveness.
- Bind material/preset controls to viewmodel state.
- Use progress/binding for ISPP pulse progression.
- Use dialogs/RichText for equations and assumptions with consistent Markdown rendering.

Fyne features:

- `canvas.Line`, `canvas.Text`, `canvas.Raster`, custom widgets.
- `binding.Float`, `binding.String`, `SliderWithData`, `SelectWithData`.
- `ProgressBarWithData`, `RichText`, `dialog.NewCustom`.

Suggested TDD slices:

1. Test plot colors update under light/dark themes.
2. Test selected material binding updates controls and summary labels.
3. Test ISPP progress binding moves from 0 to 1 across a fake run.

## Module 2 — Crossbar

Current opportunity:

- Use `widget.Table` for large conductance/level matrices and current outputs.
- Use `canvas.NewRasterWithPixels` or cached images for heatmaps.
- Add hover/cursor inspection for heatmap cells.
- Use `GridWrap` for preset/device gallery.

Fyne features:

- `widget.TableWithHeaders`, `widget.GridWrap`, `canvas.Raster`.
- `driver/desktop.Hoverable`, custom `Tappable`/`Mouseable` widgets.
- `binding.FloatList`/custom model for matrix summaries.

Suggested TDD slices:

1. Test a matrix table renders row/column counts without constructing all cells as widgets.
2. Test heatmap cell-to-coordinate mapping at two sizes.
3. Test preset gallery selection updates a bound selected preset.

## Module 3 — MNIST

Current opportunity:

- Use progress-bound batch inference UI.
- Use virtualized tables/lists for class metrics, examples, confusion matrix rows.
- Improve digit/weight heatmaps with theme-aware raster palettes.
- Add keyboard-first dataset/example navigation.

Fyne features:

- `ProgressBarWithData`, `widget.Table`, `widget.ListWithData`.
- `canvas.Raster`, `driver/desktop.CustomShortcut`.
- `binding.Int`, `binding.Float`, `binding.String`.

Suggested TDD slices:

1. Test batch progress/status bindings for fake inference.
2. Test confusion matrix table dimensions and cell text update.
3. Test next/previous sample shortcuts change selected sample binding.

## Module 4 — Circuits

Current opportunity:

- Treat circuit mode actions as a command model.
- Use custom layout for circuit overlays rather than scattered direct positions.
- Use virtualized logs/tables for compute/read/write traces.
- Use validation-backed forms for voltage/geometry/process inputs.

Fyne features:

- `fyne.Layout`, `widget.Table`, `widget.List`, `widget.Form`, validators.
- `Toolbar`, `MainMenu`, shortcuts.
- `dialog.ShowConfirm`, `dialog.NewProgress`.

Suggested TDD slices:

1. Test unified action command enablement under read/write/compute modes.
2. Test overlay layout maps key circuit objects at multiple canvas sizes.
3. Test invalid voltage input blocks apply and shows validation text.

## Module 5 — Comparison

Current opportunity:

- Use `Table` for evidence matrices and scenario comparisons.
- Bind scenario assumptions and derived metrics.
- Use charts/custom widgets with shared semantic colors and uncertainty overlays.
- Use `Accordion`/cards to progressively disclose assumptions and caveats.

Fyne features:

- `widget.TableWithHeaders`, `binding.NewSprintf`, `Card`, `Accordion`.
- Theme colors and custom renderers.

Suggested TDD slices:

1. Test scenario selection binding updates all headline metrics.
2. Test evidence table row count and confidence/caveat cell rendering.
3. Test uncertainty overlay renderer under theme changes.

## Module 6 — EDA

Current opportunity:

- Use tree/list/table views for generated artifacts.
- Add progress dialog and cancel-friendly phase model for export flows.
- Use file/folder dialogs and storage filters for artifact destinations.
- Use `TextGrid` for Verilog/SPICE/Liberty previews with line numbers.

Fyne features:

- `widget.Tree`, `widget.List`, `widget.Table`, `TextGrid`.
- `dialog.ShowFolderOpen`, `ShowFileSave`, `storage.NewExtensionFileFilter`.
- `dialog.NewProgress`, `binding.Float`.

Suggested TDD slices:

1. Test artifact tree uses unique full-path IDs.
2. Test export progress phases update a bound progress/status model.
3. Test code preview shows line-numbered generated text.

## Module 7 — Docs

Current opportunity:

- Upgrade Fyne to pick up RichText/Tree/text wrapping fixes.
- Use `DocTabs` for multiple open documents.
- Use `TreeWithData`/bindings for docs navigation and search results.
- Persist favorites/recent docs with preferences.

Fyne features:

- `RichText`, `Tree`, `DocTabs`, `binding.StringList`, preferences.
- `test.AssertRendersToMarkup` for docs layout.

Suggested TDD slices:

1. Test nested Markdown list/entities render after Fyne upgrade.
2. Test docs tree callback does not churn for empty/no-result data.
3. Test favorites persist through preferences in temp app.

## Cross-cutting release/demo opportunities

- Add `FyneApp.toml` metadata.
- Keep desktop as primary target but evaluate `fyne serve`/web demo for a reduced educational build.
- Create screenshots/gifs using existing screenshotter for apps.fyne.io-style gallery presentation.
- Add a small `fyne-guide`-derived checklist to future UI PR templates.

## Recommended sequence

1. Upgrade Fyne patch version with targeted tests.
2. Add binding-backed run status/progress in shared UI.
3. Virtualize one obvious large data view, likely Module 6 artifact tree or Module 2 matrix table.
4. Standardize plot/heatmap theme palette in shared rendering.
5. Add command model/toolbars once shared patterns are stable.
