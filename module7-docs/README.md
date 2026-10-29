# Module 7: Documentation

The shared web application provides a read-only learning library at `/docs` (also `/modules/docs`). The legacy Fyne viewer remains available; this port does not change launch defaults or qualify legacy retirement.

## Web reader

Both hosted and local builds ship the same curated corpus: currently **13 repository-authored Markdown documents**, including six module lessons, a learning overview, glossary, citation guide, trust/honesty explanations, the Park correction, and **Validation and evidence**. The displayed document, byte and search counts come from Go, not the old fabricated “230+” reference count. The citation guide contains four bibliographic DOI links, not four redistributed papers.

- Read complete escaped text, lists, code and semantic tables before loading WASM.
- Explicit search loads a dedicated read-only Go Worker. Search retains native TF-IDF/title/heading weighting, prefix/fuzzy matching and glossary aliases. Empty and unmatched queries return zero hits. Category plus author/title/DOI queries filter actual content; pages contain at most 20 results.
- Navigate library groups, breadcrumbs, citations and a Go-generated table of contents. Fragment URLs identify actual unique headings; navigation moves focus to the selected heading.
- Favorites and recent pages are ephemeral. No browser storage or automatic `.omc` write is used by the web reader.
- Local only: explicitly connect authenticated grants, name a Markdown file inside a launch-granted root, and assign a stable document ID. Go checks its bytes and parses it locally. Imported content is unverified and never enters the public build.
- Local history is off by default. Select an existing project inside a granted root and opt in, then use **Save history** or **Load history** explicitly. Only `docs-history.json` inside that opened project is used. Opting in or searching does not write it. Disable the checkbox to revoke that session's history capability. Existing history must be loaded before replacing it; a changed file or another session's save causes a conflict rather than an overwrite. Malformed history is rejected without changing the reader.

Reading documentation never starts checks, uploads scientific data, probes localhost from hosted builds, or acquires/cancels an experiment Worker. Native search/persistence compatibility remains unchanged apart from sharing the real ranking and glossary logic.

## Publication and regeneration

The source-of-truth allowlist is [`workbench/modules/docs/public.go`](../workbench/modules/docs/public.go): exact paths, SHA-256 values, MIT rights, attribution and explicit omitted-reference reasons. It is not a recursive scan of `docs/`. No internal corpus, private receipts, paper PDFs or publisher images are published. Links outside the selection must be explicitly omitted; their labels remain readable. The current selection contains no image assets. Unknown sources, hash drift, unsupported rights, raw HTML, unknown/unsafe links, remote images and unresolved fragments fail generation.

From the repository root, with the pinned Go/Node tools described in [`web/README.md`](../web/README.md):

```bash
cd web
npm run generate:docs
npm run check:docs
npm run build:hosted
npm run build:local
```

The generator is `go run ./cmd/fecim-docs-index` (repository root); `--check` verifies output without writes. The approved Markdown stays in its canonical location. `web/public/docs-manifest.json`, `web/public/docs-content/*.json` and `workbench/modules/docs/corpus.json` are deterministic, checked-in generated build inputs. Both build commands regenerate and validate them. Do not hand-edit them or automatically repin source hashes: review source/rights/reference changes in the allowlist, regenerate, then commit the complete output. There is no new scientific dataset.

Limits are Go-owned: 2048 documents, 1 MiB per decoded document, 10 MiB total decoded corpus, 256 query bytes, 20 results per page, 64 favorites and 20 recent pages. A limit violation is reported, not silently truncated into a successful index.

## Validation and evidence

The [Validation and evidence page](../docs/modules/doc-viewer/validation-and-evidence.md) replaces the legacy dashboard's unconditional Pass/Measured presentation in the web reader. No trusted candidate report is bundled: the default is **not run/unverified**. Reports bound to other source/fixture/tool identities are stale; user imports remain **imported/unverified**, even when their summary claims a pass. Reading a report never runs a test or independently verifies it. The [citation guide](../docs/modules/doc-viewer/citations.md), [trust guide](../docs/guides/trust.md) and [honesty audit](../docs/research/honesty-audit.md) retain the scientific boundaries.

Automated coverage exists in `TestDocsWebWorkflow`, `TestDocsWebSafety`, `TestDocsWebEvidence`, `TestDocsWebPort`, `TestDocsWebGrantedHistory`, and `web/test/e2e/docs.spec.ts`. It exercises real Go index output, publication rejection, native grants, lazy-search recent-page retention, opt-in history save/load and malformed-file handling, keyboard/focus, escaped imported report text, zoom, code/tables and automated accessibility in both profiles. Temporary granted documents exercise real browser pagination; Go tests also cover stale-writer conflicts and linear allocation for near-limit code blocks. It does **not** establish measured device accuracy, release qualification or actual Orca speech. **Real Orca + Firefox/Linux acceptance remains NOT RUN for both profiles.**

```bash
go test ./workbench/modules/docs ./shared/viewmodel/docs -count=1
env CGO_ENABLED=1 xvfb-run -a go test ./module7-docs/pkg/gui -run 'Test.*(Index|Query|Search|Glossary|ToC|Integrity)' -count=1
go test ./shared/viewmodel -run TestAllModulesE2E -count=1
cd web && npm test && npm run typecheck && npm run test:e2e -- test/e2e/docs.spec.ts
# After both builds, exercise the local profile against a real embedded server:
FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/docs.spec.ts
```

## Implementation map

| Path | Responsibility |
| --- | --- |
| `workbench/modules/docs/` | GUI-independent bounded corpus, native-derived ranking/aliases, constrained Markdown AST, publication policy, evidence labels and read-only Worker protocol |
| `shared/viewmodel/docs/` | Compatibility actions with real search/counts; self-supplying native web port and explicit rooted capabilities |
| `cmd/fecim-docs-index/` | Deterministic publication generator and drift check |
| `cmd/fecim-docs-wasm/` | Standalone Go/WASM entry; the feature-owned web entry is discovered by the shared builder |
| `web/src/modules/docs/` | Feature Screen, semantic reader and profile-isolated transport |
| `module7-docs/pkg/gui/` | Preserved Fyne browser, index/search wrapper, navigation, glossary and legacy persistence |

Related: [documentation index](../docs/README.md), [learning modules](../docs/modules/README.md), [technical glossary](../docs/guides/glossary.md).
