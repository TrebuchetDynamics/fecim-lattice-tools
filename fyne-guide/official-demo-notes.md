# Official demo and example project notes

Sources:
- Local module cache: `/home/xel/go/pkg/mod/fyne.io/fyne/v2@v2.7.2`
- Fyne examples repo cloned to `/tmp/fyne-examples` during research (`github.com/fyne-io/examples`).

## Official `fyne_demo` tutorial areas

- `advanced.go`
- `animation.go`
- `bind.go`
- `canvas.go`
- `collection.go`
- `container.go`
- `data.go`
- `dialog.go`
- `icons.go`
- `theme.go`
- `welcome.go`
- `widget.go`
- `window.go`

The official demo is important because it shows supported patterns for: virtualized collections, custom layouts, theme-aware canvas drawing, dialogs, windows, data binding, and widgets.

## Example repository files

- `main.go`
- `data.go`
- `xkcd/main.go`
- `tictactoe/grid.go`
- `tictactoe/board.go`
- `fractal/main.go`
- `clock/clock.go`
- `bugs/main.go`
- `bugs/button.go`
- `bugs/bundled.go`
- `bugs/board_test.go`
- `bugs/board.go`
- `img/icon/icon.go`
- `img/icon/bundled.go`

Notable examples:

- `clock`: theme-aware custom canvas/layout, live updates via `fyne.Do`, settings listener for theme changes.
- `fractal`: `canvas.NewRasterWithPixels` for computed/scientific visualization; custom `fyne.Layout` for resize behavior.
- `xkcd`: HTTP-backed form, image loading, background download, `fyne.Do` UI update.
- `bugs`/`tictactoe`: custom tappable widgets and simple game-state separation.
- moved standalone repos referenced by README: `fyne-io/calculator`, `fyne-io/solitaire`, `fyne-io/life`.
