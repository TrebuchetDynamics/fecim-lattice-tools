# Fyne Guide for FeCIM Lattice Tools

This folder collects Fyne research for improving the FeCIM desktop UI.

## What is here

| File | Purpose |
|---|---|
| `capability-matrix.md` | What Fyne can do and where each feature can help FeCIM. |
| `api-cheatsheet.md` | Practical Fyne package/API cheatsheet for FeCIM work. |
| `docs-deep-notes.md` | Deep docs digest: goroutines, widgets, layout, binding, testing, packaging. |
| `improvement-backlog.md` | Concrete improvement ideas, ranked for this repo. |
| `module-opportunity-map.md` | Module-by-module Fyne improvement opportunities and TDD slices. |
| `version-and-upgrade-notes.md` | Current Fyne version, latest available version, and relevant fixes. |
| `docs-map.md` | Generated map of docs.fyne.io pages from the sitemap. |
| `apps-catalog.md` | Generated catalog notes from apps.fyne.io with example project links. |
| `official-demo-notes.md` | Notes from `fyne_demo` and `github.com/fyne-io/examples`. |
| `example-projects-to-study.md` | Curated official/gallery app references to study next. |
| `github-source-notes.md` | Findings from `github.com/fyne-io/fyne` and the local Fyne module source tree. |
| `research-sources.md` | Source links and generated artifact notes. |
| `generated/` | Machine-readable indexes and `go doc` snapshots. |
| `raw/` | Raw fetched sitemaps/pages/changelogs for traceability. |

## Quick takeaways

- Project uses Fyne heavily already: `generated/project-fyne-usage.json` found 1000+ Go files importing Fyne packages.
- Biggest near-term leverage: upgrade `fyne.io/fyne/v2` from `v2.7.2` to `v2.7.4`, because newer patch releases fix RichText, Tree, Accordion, text wrapping, infinite progress, and scrolled text performance issues relevant to this app.
- UI architecture opportunity: use Fyne data binding more. The scan saw very low `binding.` usage compared with direct widget mutation.
- Performance opportunity: move large metrics/results views to virtualized `widget.Table`, `widget.List`, `widget.Tree`, and `widget.GridWrap` instead of building many static widgets.
- Visualization opportunity: standardize scientific plots/heatmaps around theme-aware canvas/raster widgets with cached renderers and explicit refresh contracts.
- GitHub/source study confirms we should use public Fyne packages only, avoid fighting layouts, and treat `fyne_demo` as canonical widget reference.
- Module opportunity map now turns the Fyne research into safe TDD slices for modules 1–7 and the shared shell.

TDD: N/A for this folder creation/update because this is documentation/research only and does not change production behavior.
