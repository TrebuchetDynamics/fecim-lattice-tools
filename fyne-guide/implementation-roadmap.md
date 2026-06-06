# Fyne mega-improvement implementation roadmap

Source inputs: `fyne-guide/improvement-backlog.md`, `module-opportunity-map.md`, `version-and-upgrade-notes.md`, `docs-deep-notes.md`, `github-source-notes.md`.

This roadmap is sorted from easiest/safest to hardest/highest-risk. Every behavior or dependency change must follow repo TDD policy.

## 0. Safety constraints from current repo state

- Worktree is very dirty with many unrelated edits and deletions.
- `go.mod`/`go.sum` are already dirty from GoGPU dependency removal.
- Avoid broad `go mod tidy`, visual golden churn, generated rewrites, or touching unrelated modified files.
- Prefer new focused tests and narrowly-scoped edits.

## 1. Fyne patch upgrade guard and dependency bump — easiest / highest leverage

Why first:

- Guide identifies Fyne `v2.7.4` as immediate value over current `v2.7.2`.
- Patch fixes align with FeCIM: RichText/docs, Tree, Accordion, text wrapping, infinite progress, menu/runtime stability.
- Smallest production delta: `go.mod` Fyne line + `go.sum` hashes.

TDD slice:

1. Add a focused test that requires `fyne.io/fyne/v2 >= v2.7.4` in `go.mod`.
2. RED: test fails on current `v2.7.2`.
3. GREEN: bump only Fyne module requirement and sums.
4. Validate with the new focused test and a cheap package compile/test where feasible.

Risk: dependency upgrades can alter visual goldens. Defer broad screenshot updates.

## 2. Shared progress binding adapter

Why second:

- Long-running task UX appears across modules.
- Existing `shared/progress` already has Fyne widget surface and tests.
- Binding adapter can improve app UI without rewiring every module immediately.

TDD slice:

1. Add a test for a public `ProgressBindings` adapter exposing `binding.String` status and `binding.Float` progress.
2. Implement minimal adapter from existing `progress.Progress` callbacks.
3. Validate `go test ./shared/progress`.

Risk: `shared/progress` is already dirty; inspect ownership before editing.

## 3. Module 6 EDA artifact tree model

Why third:

- Virtualized trees are a clear Fyne win.
- Artifact browsers naturally use unique full-path IDs.
- Can implement model/test before UI wiring.

TDD slice:

1. Add model test: generated artifact paths produce unique tree node IDs and stable children.
2. Implement tree model in Module 6 or shared viewmodel.
3. Later bind to `widget.Tree`.

Risk: Module 6 has many dirty files; keep model isolated.

## 4. Binding-backed run status viewmodel

Why fourth:

- Guide found very low `data/binding` usage.
- A shared status model reduces manual `fyne.Do`/`SetText` fan-out.

TDD slice:

1. Test `RunStatus` exposes operation, phase, detail, progress, and terminal state through bindings.
2. Implement in shared UI/viewmodel package.
3. Use in one module only after adapter is stable.

Risk: API design needs care to avoid shallow helper sprawl.

## 5. Module 2 matrix virtualization

Why fifth:

- Crossbar matrices are a perfect `widget.Table` use case.
- Performance and memory improvement visible to users.

TDD slice:

1. Test a matrix table/model reports row/column counts and updates cell text without allocating all cells.
2. Implement data/template callbacks.
3. Integrate into a low-risk matrix/results pane.

Risk: visual layout may change; screenshot tests may need review.

## 6. Shared plot/heatmap theme palette

Why sixth:

- Many custom renderers exist across modules.
- A semantic palette reduces inconsistent theme handling.

TDD slice:

1. Test palette exposes semantic colors for axis/grid/background/success/warning/error under light/dark themes.
2. Implement shared palette API.
3. Migrate one plot/heatmap as pilot.

Risk: visual golden churn; migrate gradually.

## 7. Command model + toolbar/menu/shortcut unification

Why later:

- High UX payoff, but touches shell and many modules.
- Needs stable command vocabulary.

TDD slice:

1. Test command IDs render consistently into menu items and toolbar actions.
2. Implement a tiny shell command registry.
3. Migrate one module command cluster.

Risk: existing shell is heavily dirty; avoid until ownership is clear.

## 8. Docs module DocTabs/search/tree modernization

Why later:

- RichText/Tree fixes improve this after Fyne upgrade.
- User-visible, but existing Module 7 files are dirty.

TDD slice:

1. Test nested markdown/entities and tree empty-state behavior.
2. Add DocTabs for multiple open docs.
3. Persist favorites/recent docs.

Risk: visual and behavioral changes across docs UI.

## 9. Packaging metadata and release polish

Why later:

- Low code complexity but release-impacting.
- Needs owner decisions: app ID, icon, website, version, target OS matrix.

TDD/validation slice:

1. Add metadata lint/test once owner confirms values.
2. Add `FyneApp.toml`.
3. Add packaging docs/scripts.

Risk: app identity/release decisions.

## 10. Web/mobile/tablet demo builds — hardest

Why last:

- Requires dependency/rendering audit: CGO, GLFW/Vulkan, file dialogs, storage, native-only paths.
- Potentially broad build-tag work.

TDD slice:

1. Add build-tag smoke test or CI script for a reduced web/mobile target.
2. Isolate unsupported desktop-only features behind adapters.
3. Build a reduced educational demo.

Risk: high platform/build complexity.

## Current recommended execution

Start with item 1: Fyne patch upgrade guard and narrow dependency bump.

If item 1 passes, next recommended slice is item 2 only after inspecting current `shared/progress` dirty diffs for ownership.
