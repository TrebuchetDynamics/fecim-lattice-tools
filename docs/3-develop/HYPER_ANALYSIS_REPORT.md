# Hyper Analysis Report

**Date:** 2026-03-05  
**Repo path used:** `<local-path>`  
**Purpose:** identify the highest-confidence gaps between the current repo state and a research-grade, agent-reliable FeCIM workflow.

## Verified baseline

- Repository breadth: `go list ./... | wc -l` returned `107` packages.
- UI heartbeat already reported for this repo path:
  - UI/a11y/theme/widget tests: PASS (`5/5` package targets)
  - Playtest layout + a11y suite: PASS (`7/7` named tests)
  - Screenshot automation: recovered from `0/8` to `8/8` captures after the root-window fallback fix
- TODO archive still shows substantial shipped scope (`~260+` completed items), so the main remaining work is evidence hardening rather than feature emptiness.

## Priority Findings

### 1. Startup viewport restore could relaunch off-screen

- Evidence: `cmd/fecim-lattice-tools/main.go` restored persisted `window_width/window_height` with a minimum clamp only.
- Observed symptom from the UI heartbeat: startup repeatedly reported `Loaded window size: 1268x1394` while screenshots were captured on a `1400x900` display.
- Impact: hidden controls, vertical clipping risk, and inconsistent screenshot evidence in headless review runs.
- Action taken on 2026-03-05: restored sizes are now clamped to the canonical `1400x900` audit envelope and covered by unit tests in `cmd/fecim-lattice-tools/window_prefs_test.go`.

### 2. Agent/doc routing drift is real and measurable

- Evidence: `rg -n "docs/development/|docs/testing/" -g '*.md' -g '!docs/archive/**'` returned `19` live references.
- Examples:
  - `AGENTS.md`
  - `docs/3-develop/gui/GUI.module4.md`
  - `docs/3-develop/gui/GUI.module6.md`
- Impact: agents and maintainers are routed to stale paths during debugging, testing, and UI review.
- Action taken on 2026-03-05: `AGENTS.md` now points at current `docs/3-develop/...` paths and the archived script reference that still exists.

### 3. Literature reproducibility is not yet fully direct-data backed

- Evidence: `rg -n '"is_placeholder_for_refinement"\\s*:\\s*true' validation/literature/data/*.provenance.json | wc -l` returned `5`.
- Evidence: `rg -n '"is_placeholder_for_refinement"\\s*:\\s*false' validation/literature/data/*.provenance.json | wc -l` returned `1`.
- Placeholder datasets currently include:
  - `validation/literature/data/pzt2024_nano14050432_fig2_thinfilm_traceB.provenance.json`
  - `validation/literature/data/pzt2024_nano14050432_fig2_thinfilm.provenance.json`
  - `validation/literature/data/bto2021_cryst11101192_hysteresis.provenance.json`
  - `validation/literature/data/alscn2022_pmc9607415_fig6a_pt_200nm.provenance.json`
  - `validation/literature/data/alscn2022_pmc9607415_fig6b_mo_200nm.provenance.json`
- Impact: research claims remain partially anchored to estimated or interim traces rather than fully committed point-by-point digitizations.

### 4. Citation debt remains concentrated in important docs

- Evidence: `rg -n "CITATION NEEDED" docs module* shared validation config -g '!docs/archive/**' | wc -l` returned `9`.
- Concentration:
  - `docs/2-learn/module2-crossbar/architecture.md`
  - `docs/4-research/PHYSICS_REALISM_AUDIT.md`
- Impact: some explanatory or numerical claims still outrun DOI-backed provenance, especially around drift, endurance, temperature, and calibrated physics defaults.

### 5. Repository health dashboard is stale enough to mislead

- Evidence from current commands:
  - `go list ./... | wc -l` returned `107`, while `docs/3-develop/repo-health.md` still reports `85`.
  - `gofmt -l .` currently returned `17` paths, while the dashboard still reports `5`.
  - `go tool cover -func=coverage.out` fails because the existing `coverage.out` references missing paths (`module1-hysteresis/pkg/gui/simulation.go: no such file or directory`).
- Impact: the dashboard is no longer a trustworthy release gate or planning source until it is regenerated from a fresh run.

### 6. Local markdown search is operationally unstable on this host

- Evidence:
  - `qmd status` emits repeated CUDA build failures: `CUDA Toolkit not found`.
  - `qmd query` initiated a `1.28 GB` generation-model download before returning any search results.
- Impact: the intended pre-scan workflow can stall active work and consume time/bandwidth instead of producing evidence.
- Current operational policy: use `rg` and direct file reads whenever `qmd` cold-starts or emits CUDA/bootstrap output.

## Research-Grade Gap Summary

The repo is already strong on deterministic tests and breadth. The main blockers to calling it research-grade are:

1. unresolved placeholder literature datasets,
2. remaining citation debt in physics-facing docs,
3. stale operational dashboards and doc routing,
4. toolchain workflows (`qmd`) that are not yet reliable enough for evidence-first iteration.

## Recommended Next Slice

1. Convert the `5` placeholder provenance files into direct digitized datasets with explicit uncertainty notes and non-placeholder flags.
2. Burn down the `9` remaining `CITATION NEEDED` markers, starting with crossbar drift/endurance/temperature claims.
3. Regenerate `docs/3-develop/repo-health.md` from fresh `go list`, `gofmt -l`, and coverage artifacts so it becomes trustworthy again.
4. Add a lightweight `scripts/research_grade_audit.sh` gate so these counts can be rerun without manual repo inspection.

## Commands Used

```bash
go list ./... | wc -l
gofmt -l .
go tool cover -func=coverage.out
rg -n "docs/development/|docs/testing/" -g '*.md' -g '!docs/archive/**'
rg -n "CITATION NEEDED" docs module* shared validation config -g '!docs/archive/**'
rg -n '"is_placeholder_for_refinement"\\s*:\\s*true' validation/literature/data/*.provenance.json
rg -n '"is_placeholder_for_refinement"\\s*:\\s*false' validation/literature/data/*.provenance.json
qmd status
qmd query "research grade validation acceptance criteria fecim" -c project_fecim -n 8
```
