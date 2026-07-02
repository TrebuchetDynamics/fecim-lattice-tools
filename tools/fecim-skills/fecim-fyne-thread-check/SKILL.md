---
name: fecim-fyne-thread-check
description: Audits Go code for goroutine-to-widget access without fyne.Do(...) wrapping, the project's most common GUI freeze cause. Use when reviewing a PR that adds goroutines, async I/O, or simulation tickers in any pkg/gui/ or shell package.
---

# fecim-fyne-thread-check

Find places where a goroutine touches a Fyne widget without `fyne.Do(...)` wrapping. See `tools/fecim-skills/_shared/fecim-context.md` (UI thread-safety rule).

## Workflow

1. **Run the automated guardrail first.** It scans default Fyne GUI scopes and exits non-zero on direct mutation-like widget calls inside goroutine literals without `fyne.Do`/`fyne.DoAndWait`:
   ```bash
   make fyne-thread-check
   # or scoped:
   go run ./tools/fyne-thread-check module6-eda/pkg/gui/tabs/builder_validation_tab.go
   ```

2. **Define audit scope.** Default: `module*/pkg/gui/`, `cmd/fecim-lattice-tools/`. Narrow to changed files for PR review:
   ```bash
   git diff --name-only main...HEAD | grep '\.go$' | grep -E 'pkg/gui|cmd/fecim'
   ```
   Use `rg` if available; on this host it may be missing, so `grep` is the default-safe fallback.

3. **Find goroutine launches:**
   ```bash
   rg -nU 'go func\(' <scope> 2>/dev/null || grep -RIn 'go func(' <scope>
   ```

4. **For each match, examine the body** for direct mutation of:
   - `*widget.*` (e.g., `Label.SetText`, `Button.SetText`, `Entry.SetText`, `ProgressBar.SetValue`)
   - `*canvas.*` (e.g., `canvas.Refresh`, `*canvas.Image.Image = ...`)
   - `*container.*` (`Add`, `Remove`, `Refresh`)
   - Direct field assignment to any `fyne.CanvasObject`

5. **Verify the call is wrapped:**
   - GOOD: `fyne.Do(func() { label.SetText("done") })`
   - GOOD: helper function whose body is itself wrapped
   - BAD: bare `label.SetText(...)` inside `go func()`

6. **Output a violation list:**
   ```
   <file>:<line>: <symbol>.<method>(...) inside goroutine — needs fyne.Do
     Suggested:
       fyne.Do(func() { <symbol>.<method>(...) })
   ```

7. **Cross-reference** `docs/3-develop/gui/FYNE_NOTES.md#threading-critical` for nuanced cases (animation tickers, blocking dialogs).

## Verification

- Input: PR adds `go func() { mylabel.SetText("done") }()` in `module1-hysteresis/pkg/gui/simulation.go`.
  Expected: skill flags `simulation.go:<line>: mylabel.SetText` and suggests the wrap.

## TDD

Audit is observation — `TDD: N/A`. Any code change made to fix a violation triggers the project's TDD hard-rule.
