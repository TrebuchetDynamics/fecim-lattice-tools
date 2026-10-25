# Fyne version and upgrade notes

## Current project state

- `go.mod` uses `fyne.io/fyne/v2 v2.7.2`.
- `go list -m -u -json fyne.io/fyne/v2` reports latest available update: `v2.7.4`.
- Local module cache inspected: `/home/xel/go/pkg/mod/fyne.io/fyne/v2@v2.7.2`.

## Why `v2.7.4` matters for FeCIM

Relevant fixes from Fyne `v2.7.3` and `v2.7.4` changelogs:

- Tree performance/layout fixes:
  - Tree constantly calling branch/child callbacks.
  - Tree with no data causing layout loops.
- RichText/docs fixes:
  - Markdown HTML entities.
  - Nested lists.
  - Scrolled text render speed.
- Accordion fixes:
  - Inefficient `MinSize` implementation killing CPU.
- Text rendering/wrapping fixes:
  - Wrapped text blur/fuzzy rendering.
  - Word-wrap layout issues.
  - One-character-per-line wrapping issue.
  - Wide character width in `TextGrid`.
- Progress/activity fixes:
  - Infinite progress bar starts automatically again.
- Runtime/platform fixes:
  - Random panic on app launch.
  - macOS main menu rebuild crash with separators.
  - Android/mobile `fyne.Do` and storage fixes.
  - System tray fixes.

These map directly to FeCIM areas:

- Module 7 docs use `RichText`, `Tree`, search, and long scrolling text.
- Multiple modules use `Accordion`, cards, and custom widgets.
- The app has heavy background simulation and UI updates through `fyne.Do`.
- Fyne desktop shell uses menus/windows and could benefit from patch fixes.

## Safe upgrade plan

Because this repo has a TDD hard rule, treat the upgrade as behavior-affecting.

1. Add/update focused tests first:
   - docs tree/search still opens docs and headings;
   - rich text markdown displays nested lists/entities if relevant;
   - accordion-heavy views render without layout loops;
   - infinite progress/activity widgets render in long-running task dialogs;
   - primary desktop smoke test still passes.
2. RED: run the focused tests against `v2.7.2` and record expected current behavior or coverage gap.
3. Upgrade:
   - `go get fyne.io/fyne/v2@v2.7.4`
   - `go mod tidy`
4. GREEN:
   - run focused GUI/module tests;
   - run `go test ./...` if feasible;
   - run screenshot/markup tests for Fyne shell.
5. Watch for visual golden changes from text wrapping/rendering fixes.

## Raw references

- `raw/changelog/v2.7.3-CHANGELOG.md`
- `raw/changelog/v2.7.4-CHANGELOG.md`
- `generated/module-version.json`
