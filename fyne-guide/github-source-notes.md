# Fyne GitHub/source notes

Source focus: `https://github.com/fyne-io/fyne` plus the local module cache at `/home/xel/go/pkg/mod/fyne.io/fyne/v2@v2.7.2`.

## What Fyne is

From the upstream README:

- Fyne is an easy-to-use UI toolkit and app API written in Go.
- It targets desktop and mobile apps from a single codebase.
- Fyne apps are designed to be portable and work without pre-installed UI libraries.
- The official demo app (`fyne_demo`) is the fastest way to explore supported widgets, containers, dialogs, themes, windows, and collections.

## Development prerequisites

Upstream README notes:

- Go is required.
- A C compiler and system development tools are required for native builds.
- Standard library usage:

```bash
go get fyne.io/fyne/v2@latest
go mod tidy
```

FeCIM note: this repo already depends on `fyne.io/fyne/v2`, currently `v2.7.2`.

## Official demo commands

```bash
go install fyne.io/fyne/v2/cmd/fyne_demo@latest
fyne_demo
```

Local demo source inspected:

- `cmd/fyne_demo/main.go`
- `cmd/fyne_demo/tutorials/canvas.go`
- `cmd/fyne_demo/tutorials/collection.go`
- `cmd/fyne_demo/tutorials/container.go`
- `cmd/fyne_demo/tutorials/dialog.go`
- `cmd/fyne_demo/tutorials/widget.go`
- `cmd/fyne_demo/tutorials/window.go`
- `cmd/fyne_demo/tutorials/bind.go`
- `cmd/fyne_demo/tutorials/theme.go`

FeCIM note: use `fyne_demo` as canonical behavior reference before writing custom versions of widgets/layouts.

## Fyne utility commands from README

Install the Fyne CLI:

```bash
go install fyne.io/tools/cmd/fyne@latest
```

Install current app into OS app location:

```bash
fyne install
```

Package mobile apps:

```bash
fyne package -os android -appID my.domain.appname
fyne install -os android
fyne package -os ios -appID my.domain.appname
fyne package -os iossimulator -appID my.domain.appname
```

Release app-store artifacts:

```bash
fyne release -os ios -certificate "Apple Distribution" -profile "My App Distribution" -appID "com.example.myapp"
```

Desktop packaging is covered in docs and summarized in `version-and-upgrade-notes.md` / `docs-deep-notes.md`.

## Source tree inventory

High-value public packages in the Fyne repo/module:

| Path | Purpose for FeCIM |
|---|---|
| `app` | App creation, metadata, preferences, settings. |
| `canvas` | Drawing primitives, raster/image rendering, animation. |
| `container` | Layout containers, tabs, split panes, scroll, navigation, clipping. |
| `data/binding` | Binding view state to widgets and preferences. |
| `data/validation` | Validation helpers for forms/entries. |
| `dialog` | Information, error, form, file, progress, custom dialogs. |
| `driver/desktop` | Desktop-specific shortcuts, hover/mouse/cursor behavior. |
| `driver/mobile` | Mobile driver integrations. |
| `driver/software` | Software driver path useful for tests/offscreen contexts. |
| `layout` | Layout implementations and spacers. |
| `storage` | Cross-platform URI/file abstraction and filters. |
| `test` | Headless GUI testing and rendering/golden helpers. |
| `theme` | Built-in themes, icons, colors, sizes, fonts. |
| `widget` | Standard widgets, collections, rich text, forms, progress, custom widget base. |

Source tree also includes:

- `cmd/fyne`: CLI entry point inside toolkit repo.
- `cmd/fyne_demo`: showcase app.
- `cmd/fyne_settings`: GUI for global Fyne settings like theme and scaling.
- `cmd/hello`: minimal hello-world app.
- `internal/*`: rendering, driver, cache, painter, theme internals. Do not import these from FeCIM.
- `theme/icons`, `theme/font`, `lang/translations`: built-in resources and translations.

## Version/changelog findings

Current FeCIM dependency:

- `fyne.io/fyne/v2 v2.7.2`

Latest discovered via `go list -m -u`:

- `v2.7.4`

Important Fyne version notes from local/GitHub changelogs:

### `v2.7.0`

Major additions:

- Canvas `Arc`, `Polygon`, `Square`, image cover fill, rounded image/corner support.
- New containers: `Navigation`, `Clip`.
- `RowWrap` layout.
- Generics for data-binding `List` and `Tree`.
- JSON theme fallback.
- System tray left-tap window support.
- RichText bullet start number.
- Entry validation option.
- Large rendering/data/theme/TextGrid performance improvements.

FeCIM relevance:

- Better visual primitives for circuit/plot UI.
- Better data binding collection support.
- Better theme and TextGrid performance for docs/export previews.

### `v2.7.2`

Relevant fixes already present in this repo's dependency:

- Accordion crash fix.
- Extended list focus fix.
- TextGrid row refresh fix.
- Entry focus/click fixes.
- Main menu Alt-Tab/Wayland focus fixes.
- Theme extension fix.
- RichText alignment fix.

### `v2.7.3` and `v2.7.4`

Relevant fixes not yet in FeCIM dependency:

- Tree callback/layout fixes.
- Accordion CPU/MinSize fixes.
- RichText nested list and HTML entity fixes.
- Text wrapping and scrolled text rendering improvements.
- Infinite progress auto-start fix.
- Random launch panic fix.
- Main menu rebuild crash fix on macOS.
- System tray fixes.

FeCIM relevance:

- Module 7 docs use Tree/RichText and long scrolling text.
- Multiple modules use Accordion/custom widgets.
- Long-running jobs need reliable progress widgets.
- Desktop shell/menu stability matters.

## Things FeCIM should avoid from source study

- Do not import `fyne.io/fyne/v2/internal/...`; internal packages are toolkit implementation details.
- Do not duplicate standard widgets unless there is a domain-specific reason.
- Do not fight the layout system with manual `Move`/`Resize` inside normal containers.
- Do not call Fyne APIs from app goroutines without `fyne.Do` / bindings.
- Do not rebuild thousands of widgets for data-heavy views; use virtualized collections.

## Best source-level references for future implementation

- `cmd/fyne_demo/tutorials/collection.go`: correct `List`, `Table`, `Tree`, `GridWrap` patterns.
- `cmd/fyne_demo/tutorials/bind.go`: binding-backed labels, sliders, progress, lists, forms.
- `cmd/fyne_demo/tutorials/widget.go`: standard widget variants and importance styling.
- `cmd/fyne_demo/tutorials/window.go`: multiple windows, fixed/centered windows, notifications, splash window.
- `cmd/fyne_settings/settings/appearance.go`: theme/settings UI reference.
- `test` package docs/source: best reference for headless GUI test patterns.

## Direct FeCIM recommendations from GitHub/source notes

1. **Upgrade patch line:** add focused tests, then move `fyne.io/fyne/v2` to `v2.7.4`.
2. **Run `fyne_demo`:** use it as the UI vocabulary reference for widgets before custom work.
3. **Adopt CLI packaging metadata:** add or verify app metadata and package commands for installable builds.
4. **Use public packages only:** keep FeCIM abstractions on top of public Fyne packages.
5. **Centralize Fyne patterns:** use shared helpers for binding, progress, commands, and custom renderer lifecycle.
