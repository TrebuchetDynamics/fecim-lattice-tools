# Fyne capability matrix

Sources: `docs-map.md`, `generated/godoc-short.md`, local `fyne.io/fyne/v2@v2.7.2`, docs.fyne.io, apps.fyne.io, and `github.com/fyne-io/examples`.

## Core app/window/runtime

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| App identity and preferences | `app.NewWithID`, `fyne.App.Preferences`, `binding.BindPreference*` | Persist selected module, material presets, theme, last export folders, warning acknowledgements. |
| Multiple windows | `App.NewWindow`, `Window.Show`, `Window.SetOnClosed`, desktop splash window | Separate plot/export windows, detached docs window, splash/loading screen. |
| Main menu and menus | `fyne.NewMainMenu`, `fyne.NewMenuItem`, popup menus | File/export actions, module navigation, recent configs, help/references. |
| Notifications | `fyne.NewNotification`, `App.SendNotification` | Long simulation completion, EDA export complete/fail, benchmark finished. |
| Shortcuts | `fyne.Shortcut`, `Canvas.AddShortcut`, desktop custom shortcuts | Ctrl+F docs search, Ctrl+E export, Ctrl+R run, module-specific hotkeys. |
| Goroutine-safe UI updates | `fyne.Do`, `fyne.DoAndWait` | Required for simulations, screenshot tests, timers, and background export pipelines. |

## Layout and navigation

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| Responsive layout | `container.NewBorder`, `NewHSplit`, `NewVSplit`, `NewAdaptiveGrid`, `layout.NewSpacer` | Stable desktop/mobile/tablet views, split plot/controls panels. |
| Tabs | `container.NewAppTabs`, `container.NewDocTabs` | Module tabs, open docs pages, compare experiment workspaces. |
| Virtualized collections | `widget.List`, `widget.Table`, `widget.Tree`, `widget.GridWrap` | Large calibration files, crossbar matrices, result tables, logs, documentation trees. |
| Scroll and clipping | `container.NewVScroll`, `NewHScroll`, `container.NewClip` | Long docs/plots/logs without oversized layout cost. |
| Inner windows | `container.NewInnerWindow`, `NewMultipleWindows` | Floating inspectors, plot details, parameter panels inside one canvas. |

## Widgets and forms

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| Basic inputs | `Entry`, `Select`, `SelectEntry`, `Slider`, `Check`, `RadioGroup`, `CheckGroup`, `DateEntry` | Simulation parameters and material/device selection. |
| Validation | `fyne.StringValidator`, `data/validation`, `Entry.Validator`, `Form` | Clamp and explain invalid physics inputs before running. |
| Actions | `Button`, importance levels, `Toolbar`, `Menu` | Primary run/export actions, danger reset actions, toolbar shortcuts. |
| Status/progress | `ProgressBar`, `ProgressBarInfinite`, `Activity` | ISPP convergence, MNIST batch runs, EDA flow progress, background loading. |
| Rich docs | `RichText`, Markdown, hyperlinks, `TextGrid` | Documentation module, code/netlist previews, logs with line numbers. |
| Cards/accordion | `Card`, `Accordion` | Group educational explanations, collapsible advanced controls. |

## Drawing, plots, and animation

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| Vector primitives | `canvas.Line`, `Rectangle`, `Circle`, `Arc`, `Polygon`, `Text` | P-E loops, schematic overlays, gauges, annotations. |
| Raster/data images | `canvas.NewRaster`, `NewRasterWithPixels`, `NewImageFromImage` | Crossbar heatmaps, MNIST digits, layout previews, scientific renderers. |
| Theme-aware graphics | `theme.Color`, `Settings().AddListener` | Plots stay legible in light/dark/high-contrast themes. |
| Animation | `fyne.NewAnimation`, `canvas.NewPositionAnimation`, color/size animations | Pulse playback, switching dynamics, guided tutorials, progress highlights. |
| Custom widgets | `widget.BaseWidget`, `CreateRenderer`, `WidgetRenderer` | Reusable plot widgets, responsive cards, custom scientific controls. |

## Data and persistence

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| Data binding | `binding.New*`, `Bind*`, `New*WithData`, `NewSprintf` | Connect viewmodels to widgets with fewer manual refresh paths. |
| Bound collections | `ListWithData`, `Table` callbacks, `TreeWithData` | Efficient live result streams and docs/search trees. |
| Preferences binding | `BindPreferenceString/Bool/Float/Int/List` | Persist settings without bespoke plumbing. |
| File/storage abstraction | `storage.NewFileURI`, `Reader`, `Writer`, dialogs with URI closers | Cross-platform file open/save/export and mobile/web storage readiness. |
| Resource bundling | `fyne bundle`, `fyne.Resource`, `NewStaticResource` | Embed icons, screenshots, docs, material presets, EDA templates. |

## Dialogs and user assistance

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| Information/errors | `dialog.ShowInformation`, `ShowError`, `ShowConfirm` | Clear simulation/EDA failure messages and confirmations. |
| Custom dialogs | `dialog.NewCustom`, `NewCustomConfirm`, `ShowCustomWithoutButtons` | Equation explanations, validation dashboards, export previews. |
| File/folder dialogs | `ShowFileOpen`, `ShowFileSave`, `ShowFolderOpen`, filters | Load calibration datasets, save SPICE/Verilog/Liberty/plots. |
| Progress dialogs | `dialog.NewProgress`, `NewProgressInfinite` | Long-running export/simulation tasks with cancel-ready design. |

## Platform/distribution

| Capability | Fyne tooling/docs | FeCIM use |
|---|---|---|
| Desktop packaging | `fyne package`, app metadata | Build installable Linux/macOS/Windows artifacts. |
| WebAssembly | docs: Run in a Browser | Demo/teaching build if simulation/runtime dependencies allow it. |
| Mobile packaging | Android/iOS docs | Educational tablet viewer or demos. |
| System tray | desktop app systray docs/API | Background long job monitoring, quick launch recent files. |
| Cross-compiling | docs: cross-compiling | Release automation and CI artifacts. |

## Testing and quality

| Capability | Fyne APIs | FeCIM use |
|---|---|---|
| Headless widget tests | `fyne.io/fyne/v2/test` | Existing GUI tests; expand for every new workflow. |
| Golden markup/images | `AssertRendersToMarkup`, `AssertRendersToImage`, `RenderObjectToMarkup` | Catch layout regressions and screenshot drift. |
| Interaction helpers | `test.Tap`, `Type`, `Drag`, `Scroll`, `FocusNext` | TDD for form validation, module navigation, plot controls. |
| Temporary app/window | `test.NewTempApp`, `NewTempWindow` | Isolated tests without global app leakage. |

## Highest-value gaps for this repo

1. **Data binding:** Current scan found many direct widget updates but very little `data/binding` usage. Bound viewmodels would reduce refresh bugs.
2. **Virtualized data:** Use `Table/List/Tree/GridWrap` for large matrices/logs/docs instead of constructing hundreds/thousands of widgets.
3. **Theme-aware custom renderers:** Standardize renderer lifecycle, cached objects, and theme refresh for custom plots/heatmaps.
4. **Packaging metadata:** Make FeCIM installable with proper app ID, icon, metadata, and release commands.
5. **Upgrade patch line:** Move to `v2.7.4` after TDD, because it fixes UI issues relevant to RichText, Tree, Accordion, text wrapping, and progress widgets.
