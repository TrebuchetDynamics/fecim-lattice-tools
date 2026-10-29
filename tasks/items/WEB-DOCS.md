+++
schema = 1
id = "WEB-DOCS"
title = "Port documentation search and learning to the shared web application"
short = "Real web docs and search"
status = "done"
goal = "WEB-ALL"
section = "Next"
rank = 2000
depends_on = ["WEB-WORKBENCH"]
touches = ["workbench/modules/docs/", "shared/viewmodel/docs/", "cmd/fecim-docs-wasm/", "cmd/fecim-docs-index/", "web/src/modules/docs/", "web/public/docs-manifest.json", "web/public/docs-content/", "web/test/e2e/docs.spec.ts", "web/package.json", "web/package-lock.json", "docs/README.md", "module7-docs/README.md", "module7-docs/pkg/gui/"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement M7 roadmap in one module Screen using a Go reader/index/search core and curated manifest; real navigation/search/glossary/citations/ToC/favorites, not the canned docs viewmodel. Ship the same approved corpus in both builds, local-only additional granted docs and explicit opt-in root history. Extract GUI-bound search logic without importing Fyne or auto .omc writes. Real Go query/count/provenance plus escaped constrained AST, safe links/assets and actual anchors."
acceptance = ["Given curated and temporary corpus fixtures, when Go indexes/searches/filters/queries glossary/anchors, then matches/counts/aliases/hashes are real and deterministic and bounds/unknown/empty queries report honestly; TestDocsWebWorkflow reuses the real native search/glossary oracles, replacing fabricated matches/counts.", "Given hostile Markdown/HTML/URL/image/import or internal/private path fixtures, when build/Go parsing/React rendering runs, then publication allowlist/attribution holds, scripts/remote auto-fetch/internal corpus leakage fail, and hosted search/history make no scientific uploads/local calls/persistent writes; TestDocsWebSafety and docs.spec.ts use actual Go index outputs. TestDocsWebEvidence also preserves the mounted Validation view's evidence-reading outcome while replacing unconditional legacy Pass/Measured badges with missing/stale/imported-unverified/candidate-bound recorded-evidence status; reads do not execute tests.", "Given both profiles with keyboard/zoom, when browsing/searching/paging/following citation/ToC/glossary and opting into local history, then headings/anchors/counts/focus and complete text/code/table reading are accessible; docs.spec.ts proves automated behavior while actual Orca remains a separately unmet manual gate."]
checks = ["go test ./workbench/modules/docs ./shared/viewmodel/docs -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./module7-docs/pkg/gui -run 'Test.*(Index|Query|Search|Glossary|ToC|Integrity)' -count=1", "cd web && npm run generate:docs && npm test && npm run typecheck && npm run build:hosted && npm run build:local && npm run test:e2e -- test/e2e/docs.spec.ts", "go test ./shared/viewmodel -run TestAllModulesE2E -count=1"]
exclusions = ["No implementation/deployment/service change follows from planning; future implementation preserves Go science/CLI, original example-byte parity, existing qualification gates and no hosted scientific upload/persistence/localhost probing.", "No automatic EDA-tool/model/dataset/PDK/container acquisition or installation; explicitly scoped pinned build/test dependency provisioning is allowed only during authorized implementation. No new physics/TS scientific model, fabricated accuracy/provenance/tool pass, premature Fyne deletion/default switch, public release or push."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/WEB-DOCS.toml"
+++


- Sources: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md; docs/internals/adr/0005-unified-react-hosted-and-local.md; docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md
- Produces: Module docs web port and Screen at the installed path; bounded Go corpus/search/AST and build generator/manifests with actual counts/rights; read-only search Worker. No shared registry edits.
- Consumes: WEB-WORKBENCH common module protocol/client and route discovery; existing real module7 index/glossary behavior, curated approved repository/citation files, not recursive internal/PDF publication.
- Repo obligations triggered: M7 roadmap corpus/caps/history/XSS decisions and existing docs integrity tests. Preserve UI functionality by moving reusable logic, not deleting native search before qualification. Public content manifest and generator command are task-owned new deliverables; build fails unknown/hash/rights/unsafe reference, may explicitly omit an unlicensed asset and retain a citation link. Owner evidence/real Orca still pending; automated tests cannot qualify it. Generated content is build output, manifest/source index policy explicit; no new scientific dataset.
- Next action: Add TestDocsWebWorkflow with a small real corpus showing an unmatched query returns no hits, not canned text; watch RED, then extract real indexing.
- Verification status: All new feature tests/commands are proposed and unrun. Source inspection is not numerical/runtime/manual/tool validation. No implementation or dispatch in this planning occurrence.

- Builder disposition: Builder P2 mounted-validation disposition resolved: roadmap assigns a named Validation and evidence page to WEB-DOCS, with curated source/check/evidence descriptions and truthful revision/hash/tool/fixture status, no new runner, no raw private receipts and no blanket Pass/Measured claims. TestDocsWebEvidence covers absent/stale/imported reports. The hardcoded dashboard behavior is intentionally retired, not silently treated as scientifically valid. All GUI files needed for search/index/ToC extraction are bounded by module7-docs/pkg/gui/; existing CLI/native compatibility remains until later qualified retirement.

- CI ownership: WEB-WORKBENCH owns shared workflow/build discovery once; this task supplies real module tests/Screen/entry artifacts picked up by it, without editing a common CI or route registry. Declared cold all-module test/presentation overlaps still need merge review; no artificial dependency is added.
- Pending live: Real Orca + Firefox/Linux human acceptance for hosted and authenticated local Docs workflows remains NOT RUN: record actual speech/focus, citation/ToC/glossary navigation, zoom and complete text/code/table reading with candidate/tool/voice fingerprints. Automated tests do not qualify it.
