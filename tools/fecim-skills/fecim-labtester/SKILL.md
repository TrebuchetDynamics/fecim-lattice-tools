---
name: fecim-labtester
description: Runs the FeCIM test matrix (full, race, module-scoped, coverage, golden regen) and interprets physics regression failures using the 5 known bug patterns. Use when running tests, debugging test failures, or regenerating physics golden files.
---

# fecim-labtester

React web is the primary UI target; Fyne is deprecated compatibility. The React client is planned, not an already-shipped distribution. For browser work, inspect current implementation and supported test tooling before choosing checks; do not invent a React package/test command or treat the Astro marketing site in `web/` as the client.

Keep Go core/CLI and physics regression coverage, plus legacy Fyne guards while shipped. Fyne widget mutations from goroutines still require `fyne.Do(func() { ... })`; deprecation does not waive thread safety or prove browser/accessibility acceptance.

Pick the right test invocation, run it, and triage failures against the 5 known bug patterns. See `tools/fecim-skills/_shared/fecim-context.md` for the Go test matrix and bug list, and `docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md` for the approved browser direction.

## Workflow

1. **Pick scope** by the change being verified:
   - Full check (unchanged): `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`
   - Whole suite: `go test ./...`
   - Counted whole suite: `go test -json ./...` followed by a pass/fail/skip/package summary
   - Race detection: `go test -race ./...`
   - Module: `go test ./module1-hysteresis/...` (or `make test-hys`, `test-xbar`, `test-mnist`, `test-circuits`, `test-shared`)
   - Deprecated Fyne compatibility: `make test-legacy-fyne`
   - Coverage: `go test -cover ./...`
   - Diff hygiene: `git diff --check`

2. **For physics-regression failures**, classify against the 5 known patterns:
   1. Guard-band sign flip
   2. Bounds collapse `[VMin, VMax]`
   3. ACCEPT ±1 guard interaction
   4. Zero-field bounds reset
   5. Preisach Everett zero-clamp

3. **Golden regeneration** is allowed only when divergence is intentional and the user has confirmed:
   ```bash
   FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...
   ```
   Diff `validation/testdata/physics_regression/` before committing — unintended changes are bugs.

4. **Output the TDD evidence block** per `tools/fecim-skills/_shared/tdd-evidence-template.md`:
   ```
   RED:  go test ./module1-hysteresis/... -run TestX
         FAIL TestX (Preisach Everett zero-clamp pattern)
   GREEN: go test ./module1-hysteresis/... -run TestX
          ok
   VERIFY: go test ./... && go test -race ./... && git diff --check
   ```

## Verification

- Input: "TestPreisachEverett is failing in module1-hysteresis."
  Expected: maps to known bug pattern #5; suggests product-form Everett vs factorized; runs the targeted test, then full suite, then race.

## TDD

This skill is the TDD verifier itself — every change made under it must produce the RED/GREEN block above before commit.
