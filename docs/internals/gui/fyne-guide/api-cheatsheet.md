# Fyne API cheatsheet for FeCIM

This cheatsheet condenses the key packages from `generated/godoc-short.md` and maps them to FeCIM usage.

## `fyne.io/fyne/v2`

Core interfaces/types:

- App/window/runtime: `App`, `Window`, `Driver`, `Canvas`, `Clipboard`, `Notification`, `MainMenu`, `Menu`, `MenuItem`.
- UI object model: `CanvasObject`, `Widget`, `WidgetRenderer`, `Container`, `Layout`.
- Input/events: `Tappable`, `SecondaryTappable`, `DoubleTappable`, `Draggable`, `Scrollable`, `Focusable`, `Shortcutable`, `KeyEvent`, `PointEvent`, `DragEvent`, `ScrollEvent`.
- Geometry/text: `Position`, `Size`, `TextStyle`, `TextAlign`, `TextWrap`, `TextTruncation`.
- Theme/resources: `Theme`, `Resource`, `StaticResource`, `ThemedResource`, `ThemeColorName`, `ThemeIconName`, `ThemeSizeName`.
- Storage: `URI`, `URIReadCloser`, `URIWriteCloser`, `ListableURI`, `Storage`.
- Threading: `Do`, `DoAndWait`.

High-value FeCIM APIs:

```go
fyne.Do(func() { /* UI update from background goroutine */ })
fyne.DoAndWait(func() { /* synchronized UI update */ })
fyne.NewNotification("Run complete", "MNIST batch finished")
fyne.NewMainMenu(...)
fyne.NewMenuItemWithIcon("Export", theme.DocumentSaveIcon(), exportFn)
fyne.NewStaticResource("preset.json", data)
```

## `app`

- `app.New()` creates an app.
- `app.NewWithID(id string)` creates an app with stable preferences storage.
- `app.SetMetadata` sets app metadata.

FeCIM default should prefer `app.NewWithID(...)` for preference-backed UI.

## `canvas`

Drawing primitives:

- `NewLine`, `NewRectangle`, `NewCircle`, `NewArc`, `NewPolygon`, `NewText`.
- `NewImageFromFile`, `NewImageFromImage`, `NewImageFromResource`, `NewImageFromURI`.
- `NewRaster`, `NewRasterWithPixels`, `NewRasterFromImage`.
- Gradients: `NewHorizontalGradient`, `NewVerticalGradient`, `NewLinearGradient`, `NewRadialGradient`.
- Animations: `NewPositionAnimation`, `NewSizeAnimation`, `NewColorRGBAAnimation`, `fyne.NewAnimation`.
- `canvas.Refresh(obj)`.

FeCIM use:

- Plot axes/traces: `Line`, `Text`, `Rectangle`.
- Heatmaps/MNIST digits: `Raster` or `Image`.
- Circuit diagrams: vector primitives and custom layout.
- Educational animations: Fyne animations with stop controls.

## `container`

Standard composition:

- `NewVBox`, `NewHBox`, `NewBorder`, `NewCenter`, `NewPadded`, `NewStack`, `NewMax`.
- `NewGridWithColumns`, `NewGridWithRows`, `NewAdaptiveGrid`, `NewGridWrap`.
- `NewHSplit`, `NewVSplit`.
- `NewVScroll`, `NewHScroll`, `NewScroll`.
- `NewAppTabs`, `NewDocTabs`.
- `NewInnerWindow`, `NewMultipleWindows`.
- `NewThemeOverride`.

FeCIM use:

- Module shell: `Border` + `Split` + tabs.
- Responsive cards: adaptive grids.
- Long docs/logs: scroll containers.
- Detached document/workspace UX: `DocTabs` or inner windows.

## `layout`

Layout primitives:

- `NewVBoxLayout`, `NewHBoxLayout`, `NewBorderLayout`, `NewCenterLayout`, `NewFormLayout`, `NewPaddedLayout`, `NewStackLayout`.
- `NewGridLayout`, `NewGridLayoutWithColumns`, `NewGridLayoutWithRows`, `NewGridWrapLayout`.
- `NewAdaptiveGridLayout`, `NewRowWrapLayout`.
- `NewSpacer`.

FeCIM use:

- Prefer containers over directly using layout objects unless custom composition needs it.
- Use `Spacer` for toolbar/header alignment.

## `widget`

Inputs/actions/status:

- Actions: `Button`, `Toolbar`, `Menu`, `Hyperlink`.
- Inputs: `Entry`, `Select`, `SelectEntry`, `Check`, `CheckGroup`, `RadioGroup`, `Slider`, `Calendar`, `DateEntry`.
- Display: `Label`, `Icon`, `Card`, `Accordion`, `Separator`, `RichText`, `TextGrid`, `FileIcon`.
- Progress: `ProgressBar`, `ProgressBarInfinite`, `Activity`.
- Forms: `Form`, `FormItem`.
- Collections: `List`, `Table`, `Tree`, `GridWrap`.
- Popups: `PopUp`, `PopUpMenu`, `ShowPopUp*`.
- Custom widgets: `BaseWidget`, `NewSimpleRenderer`.

FeCIM use:

- `RichText` for docs/equations.
- `TextGrid` for generated code/netlist/log previews.
- `Table` for dense numeric output.
- `Tree` for docs/artifacts hierarchy.
- `Toolbar` for module command surfaces.
- `ProgressBarWithData` for long run phases.

## `data/binding`

Important bindings:

- Scalar: `NewString`, `NewBool`, `NewFloat`, `NewInt`, `NewURI`, `NewUntyped`.
- External: `BindString`, `BindBool`, `BindFloat`, `BindInt`, `BindURI`, `BindUntyped`.
- Lists: `NewStringList`, `NewFloatList`, `BindStringList`, etc.
- Trees: `NewStringTree`, `BindStringTree`, etc.
- Struct/map: `BindStruct`, `NewUntypedMap`, `BindUntypedMap`.
- Conversion: `FloatToStringWithFormat`, `IntToString`, `BoolToString`, `StringToFloat`, `NewSprintf`.
- Preferences: `BindPreferenceString`, `BindPreferenceBool`, `BindPreferenceFloat`, `BindPreferenceInt`, list variants.
- Listener: `NewDataListener`.

FeCIM use:

- One binding-backed state object per module panel.
- Bind progress, selected presets, status text, and validation messages.
- Use list/tree bindings for logs/artifact/document trees.

## `data/validation`

- `NewRegexp` for constrained names/IDs.
- `NewTime` for time-formatted input.
- `NewAllStrings` to combine validators.

FeCIM use:

- Validate numeric fields with custom `fyne.StringValidator`.
- Validate EDA module/cell names and export paths.

## `dialog`

Dialog APIs:

- Info/error/confirm: `ShowInformation`, `ShowError`, `ShowConfirm`, `NewInformation`, `NewError`, `NewConfirm`.
- Input/forms: `ShowEntryDialog`, `ShowForm`, `NewEntryDialog`, `NewForm`.
- Custom: `ShowCustom`, `ShowCustomConfirm`, `ShowCustomWithoutButtons`, `NewCustom*`.
- Files: `ShowFileOpen`, `ShowFileSave`, `ShowFolderOpen`, `NewFileOpen`, `NewFileSave`, `NewFolderOpen`.
- Progress: `NewProgress`, `NewProgressInfinite`.
- Color: `ShowColorPicker`, `NewColorPicker`.

FeCIM use:

- Equation details, export preview, run confirmation, validation errors, long task progress.

## `storage`

Storage APIs:

- URI creation/parsing: `NewURI`, `ParseURI`, `NewFileURI`.
- I/O: `Reader`, `Writer`, `OpenFileFromURI`, `SaveFileToURI`, `Appender`.
- File operations: `List`, `Copy`, `Move`, `Delete`, `Exists`, `Parent`, `Child`.
- Filters: `NewExtensionFileFilter`, `NewMimeTypeFileFilter`.

FeCIM use:

- Cross-platform import/export and future mobile/web readiness.

## `theme`

Useful capabilities:

- Built-in themes: `DarkTheme`, `LightTheme`, `DefaultTheme`.
- Current values: `Color`, `ColorForWidget`, `Current`, `CurrentForWidget`, `Size`, `Icon`.
- Common colors: foreground/background/primary/error/success/warning/disabled/input/hover/focus.
- Icons: document, folder, search, menu, settings, media, navigation, content add/remove/copy/save, warning/error/info, etc.
- JSON/custom themes: `FromJSON`, `FromJSONWithFallback`.

FeCIM use:

- Consistent semantic colors for plots and validation states.
- Icon vocabulary for commands and docs.
- Accessibility/high-contrast theme layer.

## `test`

Test helpers:

- App/window/canvas: `NewTempApp`, `NewTempWindow`, `NewWindow`, `Canvas`, `NewCanvas`.
- Interactions: `Tap`, `TapAt`, `TapCanvas`, `Type`, `TypeOnCanvas`, `Drag`, `Scroll`, `FocusNext`, `FocusPrevious`.
- Rendering/goldens: `RenderToMarkup`, `RenderObjectToMarkup`, `AssertRendersToMarkup`, `AssertObjectRendersToMarkup`, `AssertRendersToImage`, `AssertObjectRendersToImage`.
- Themes: `WithTestTheme`, `ApplyTheme`, `KnownThemeVariants`.
- Renderer access: `WidgetRenderer`, `TempWidgetRenderer`.

FeCIM use:

- Required for TDD of all UI changes.
- Prefer focused widget tests before full screenshot crawls.

## `driver/desktop`

Desktop extensions:

- `CustomShortcut` for platform desktop keybindings.
- Hover/mouse/cursor interfaces: `Hoverable`, `Mouseable`, `Cursorable`.
- Desktop app/driver/canvas interfaces.

FeCIM use:

- Keyboard command model.
- Hover tooltips/inspectors for plots and heatmaps.
- Cursor affordances for drag/zoom/inspect modes.
