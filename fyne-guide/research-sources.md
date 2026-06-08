# Research sources

Research performed on 2026-06-05/06 from the FeCIM repo root.

## Primary Fyne sources

- Fyne docs: https://docs.fyne.io/
- Fyne apps gallery: https://apps.fyne.io/
- Fyne GitHub repository: https://github.com/fyne-io/fyne
- Fyne examples repository: https://github.com/fyne-io/examples
- Local Fyne module source: `/home/xel/go/pkg/mod/fyne.io/fyne/v2@v2.7.2`
- Local project dependency: `fyne.io/fyne/v2 v2.7.2`
- Latest update reported by Go tooling: `fyne.io/fyne/v2 v2.7.4`

## Generated/fetched artifacts

| Artifact | Source |
|---|---|
| `raw/docs-sitemap.xml` | `https://docs.fyne.io/sitemap.xml` |
| `raw/apps-sitemap.xml` | `https://apps.fyne.io/sitemap.xml` |
| `raw/apps-all.html` | `https://apps.fyne.io/all.html` |
| `generated/docs-index.json` | Parsed docs sitemap/page title index; source pages were fetched during generation but raw page HTML was not retained to avoid repository bloat. |
| `generated/apps-index.json` | Parsed apps catalog/source/install index; source pages were fetched during generation but raw page HTML was not retained to avoid repository bloat. |
| `raw/changelog/v2.7.3-CHANGELOG.md` | Fyne GitHub raw changelog for tag `v2.7.3`. |
| `raw/changelog/v2.7.4-CHANGELOG.md` | Fyne GitHub raw changelog for tag `v2.7.4`. |
| `generated/godoc-short.md` | `go doc -short` snapshot for key Fyne packages. |
| `generated/module-version.json` | `go list -m -json` and `go list -m -u -json` output for Fyne. |
| `generated/project-fyne-usage.json` | Repo scan of Fyne imports/API token usage. |
| `github-source-notes.md` | Human-readable notes from upstream README, changelog, demo source, and public package tree. |

## Local reference paths inspected

- `/home/xel/go/pkg/mod/fyne.io/fyne/v2@v2.7.2`
- `/home/xel/go/pkg/mod/fyne.io/fyne/v2@v2.7.2/cmd/fyne_demo/tutorials`
- `/tmp/fyne-examples` (temporary shallow clone of `github.com/fyne-io/examples`)

## Notes

- `api.fyne.io` did not resolve during research; docs.fyne.io now includes API docs under `https://docs.fyne.io/api/`.
- `apps-catalog.md` is generated from public app pages and should be treated as a scouting index, not an endorsement of code quality.
- If we clone third-party app repos for deeper review later, do it in `/tmp` or `references/` and document license/status before copying any code.
