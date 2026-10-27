# Accessible browser workbench — v1 design

Date: 2026-10-25
Status: Approved direction with seven additions; consolidated written design awaiting review. **Not implemented.** No implementation or commits authorized by this document.

## Intent and source of decisions

Provide **one React application** for the simulation-only research workbench, built for hosted static/browser-only and local/full-engine delivery. Its home view replaces the Astro landing page. A user must be able to inspect models/results and perform the capabilities available in that build without interpreting a plot. Go owns science in both modes: WASM in a Web Worker for browser-only computation; the local Go server for filesystem/large-study/tool operations. Preserve **table first; Go owns every number; cancellation requested is not cancellation finished; plot scope is explicit**. The [accepted delivery decision](../../adr/0005-unified-react-hosted-and-local.md) and [WASM feasibility evidence](../../architecture/browser-wasm-feasibility.md) distinguish approved scope from implemented capabilities.

This document consolidates the approved stack, API boundary, security design, workflow, and user's seven additions from the design conversation. It is a companion to the [core workbench architecture](2026-07-19-research-design-workbench-architecture-design.md), not a claim that browser features already exist. Its manual release gate is [the numbered Orca task scripts](2026-10-25-accessible-browser-workbench-orca-acceptance.md).

### Approved decision register and continuation boundary

These decisions were explicitly approved in the conversation. A resumed session must read this register and the relevant section before asking questions; session restart alone does not reopen a decision. The written consolidation remains available for review, not implementation authorization.

| ID | Approved decision | Owning section |
|---|---|---|
| BW-01 | One React application, two builds: hosted static/browser-only and local/full-engine. Keep Go science and CLI; hosted computation is client-side Go/WASM, never a hosted compute service or TypeScript model. Supersedes WASM-deferred scope. | Unified delivery and capability profiles; ADR 0005 |
| BW-02 | React, TypeScript, Vite; React Aria Components/native HTML; same routes/components in both builds. Local assets embedded in Go; hosted static assets include Go/WASM and its Worker/runtime. | Stack, contracts and build policy |
| BW-03 | Fyne remains fully deprecated. Latest owner scope requires module-by-module retirement when React replacement behavior/accessibility parity is qualified, then deletion. No new Fyne features. React home also replaces Astro: no parallel marketing UI. | React-primary deprecation policy; ADR 0005 |
| BW-04 | Data tables are primary; Go computes values, summaries and scientific analysis; exact values, filtering/pagination and full-data export remain accessible. | Go/browser ownership; sections 2, 4, 5 |
| BW-05 | Approved API boundary, with explicit close, active execution discovery, ETags/If-Match, RFC 9457 errors, server pagination, count-only SSE and health fingerprints. | Local API contract |
| BW-06 | Launch-granted roots, fragment bootstrap, cookie-authenticated SSE, exact Host/Origin protection, mutation header and end-to-end filesystem confinement. | Authentication, confinement and threat model |
| BW-07 | Focus: view heading after navigation; caption with page/sort after pagination; error summary after failed submit. | Section 1 |
| BW-08 | Go supplies display and speakable unit/exponent forms; actual Orca pronunciation is acceptance. | Section 2 |
| BW-09 | Stale-edit comparison is a field-by-field text list, not a visual YAML diff. | Section 3 |
| BW-10 | Browser submits field changes; Go edits YAML nodes, retaining unexposed fields/comments. No-op save is byte-identical. | Section 3 |
| BW-11 | Compare 2–4 runs: metrics as rows, runs as columns, with units and evidence. | Section 4 |
| BW-12 | Each v1 plot has at most one Tab stop; selection is driven from the table. | Section 5 |
| BW-13 | Completion announces run, feasible, Pareto and failure counts. | Section 6 |
| BW-14 | Manual acceptance uses Orca + Firefox on Linux, keyboard only, through numbered scripts with start state, goal, required speech and failure criteria, in both deployment profiles. | Manual acceptance companion |
| BW-15 | Hosted build must never contact/probe localhost or pair with the local engine. Server-only actions show local-run instructions and the persistent capability/privacy disclaimer. Local build retains same-origin security design. | Unified delivery and capability profiles |
| BW-16 | The checked-in example must produce byte-identical results.json through native study/CLI and browser Go/WASM Worker, with identical scientific inputs/versions/revision. No tolerance/field-stripping substitute. | ADR 0005; WASM feasibility and parity gate |

Do not infer that this register grants implementation, commits, task execution, or release authority. No approval is fabricated for unexecuted checks. Newly discovered material conflicts must be recorded in the affected section rather than silently changing approved intent.

## React-primary deprecation policy

Status: Accepted

**Owner decision:** Fyne is **fully deprecated** in favor of React. The latest explicit scope adds one shared React app with hosted/local builds and requires module-by-module Fyne retirement at replacement parity, followed by deletion. This replaces the earlier separate-site/WASM-deferred direction and the intervening no-parity-removal-condition wording; it does not reopen completed work or authorize implementation. Deprecation is this repository's policy, not a claim about upstream Fyne.

- **Primary UI target:** the approved React + TypeScript + Vite browser application over the Go API. Go retains scientific computations, values, units, evidence, persistence and the CLI; no physics or language rewrite.
- **Fully deprecated UI:** existing Fyne commands/modules are legacy code, not a new-feature target. No Fyne-only features/demos. React replacement qualification covers agreed behavior and accessible equivalents; it does not mean replicating inaccessible widget internals. Existing safety/thread rules apply while legacy code remains shipped.
- **Retirement:** qualify and retire Fyne module views one at a time, then delete their UI code/dependencies when unused, preserving Go science/CLI and completed receipts. Actual removal/default-switch/dependency changes need a separately authorized implementation/test batch; no deletion occurs here. Do not pretend an unbuilt React view already replaces a module.
- **One public UI:** replace the Astro site with the same React application's home/docs views. Hosted/static and embedded/local are deployment profiles of that application, not separate sites, route trees or duplicated component implementations. No hosted computation service or desktop wrapper.
- **Completed work remains history:** T-005's verified Fyne result presentation and all existing receipts are retained, not renamed into React delivery or superseded as if unimplemented. The Go foundations T-001..T-004/T-006 remain reusable. At inspection, all six are closed with coordinator-verified receipts; no new runtime tests are claimed by this documentation pass.

### Historical focused deprecation batch and boundary

This subsection records the earlier deprecation-policy planning batch and its then-current checks, not the latest delivery architecture or current task completion. ADR 0005 and the unified profiles below supersede its separate-site/WASM-deferred assumptions. Its requested outcome was to correct UI direction and future work selection, not implement the whole React client. The approved API/security/editing/table/plot/manual-test contract below is retained. It is not reopened or silently treated as delivered. Concrete browser implementation still needs its own full bounded plan, including end-to-end storage confinement and assets/contracts; this deprecation batch must not launch a listener or scaffold/install frontend dependencies.

Documentation owners updated in this batch: this design/decision register, the original core architecture's status notice, README, architecture guide and terminology. Executable agent/skill instructions are implementation-role artifacts, so they are **not edited during planning**. Two complete records under `REACT-UI-POLICY` own the remaining instruction-role work:

| Task | Outcome | Writes / prerequisites |
|---|---|---|
| `T-007-5906cd8f` | Replace Fyne-first agent guidance with React-primary policy | `AGENTS.md`, `CLAUDE.md`, shared FeCIM context; no prerequisite |
| `T-008-5079bc9b` | Retarget build and test skills to the React migration | Canonical builder/labtester skills plus generated Codex/OpenCode metadata; no prerequisite |

Their exact scope, exclusions, literal policy assertions, existing adapter checks and next actions are in `tasks/items/`. Both are independent, with disjoint source ownership and no live/exclusive resource. Protected instruction-path scheduling can still serialize them; this is not a product dependency. Planner's queue call: priority 2, no focus takeover, existing task ranks/contracts/history preserved. Records remain uncommitted and do not authorize the active runner to execute newly saved work. No successor planning daemon is created.

Planner's calls: phased deprecation instead of immediate deletion; no new Fyne feature work; keep the scientific Go core/CLI and legacy regression gates; distinguish React product priority from current runtime availability. Immediate default-switch/deletion was rejected because there is no React workbench distribution in this checkout. Continuing Fyne-first instructions was rejected because it contradicts the owner's explicit decision.

Verification is qualitative: no authoritative plan calls Fyne the future primary UI, no instruction requires new features to gain a Fyne path, no command claims a not-yet-built React distribution, and completed receipts remain unchanged. Future instruction tasks run literal policy checks and existing adapter checks, not tests invented solely to count prose edits. Preserve the existing `Full check:` command and legacy safety obligations. Both instruction owners replace the obsolete `make test-Fyne-ui` reference with the existing `make test-legacy-fyne` where relevant, without changing test targets. All future task checks are proposed/unrun; TDD N/A for this documentation/task-only planning batch.

Builder review: a fresh read-only reviewer reported **“nothing important left”**, confirming both task scopes and their independent lanes. This was a planning/source review, not a runtime or accessibility pass. The reviewer’s shared-context test-alias observation is included in T-007. Final planning checks: `autowork tasks validate --worktree .` **PASS** (13 tasks, 4 goals, no drafts; only this new goal/two tasks pending). Task readback, new policy links, independent write/dependency ownership and literal Python-check syntax **PASS**; scoped `git diff --check` **PASS**. These are consistency receipts, not execution evidence. No instruction, product code, test, dependency or service is changed by this planning batch; no commit is created. Next action after separate execution authorization: T-007's instruction-policy assertion, or independent T-008's skill-policy assertion, then only their scoped changes. The full React client implementation and human Orca gate remain outside this batch.

### Documentation map and current implementation evidence

| Source | Current fact / boundary |
|---|---|
| `workbench/experiment/types.go`; `workbench/presentation/value.go` | Go metrics retain numeric value, unit, model, assumptions and evidence; `presentation.FormatMetric` now supplies separate display/speakable/availability forms. Browser consumption is not yet implemented. |
| `workbench/experiment/analyze.go:28` | Feasibility and Pareto analysis belongs to Go. |
| `workbench/experiment/runner.go`; `progress.go` | The runner accepts a context and an optional nonblocking count-progress channel; this is not yet a browser job registry or SSE endpoint. |
| `workbench/study/study.go` | `study.Run` retains trustworthy partial results and explicit `ResultsState`/`AnalysisAvailable`; the browser still needs its lifecycle adapter. |
| `workbench/project/load.go` | `LoadRoot` provides scoped reads with strict typed YAML; it is not end-to-end confined persistence or a comment-preserving editor. |
| `web/README.md` | `web/` is a static Astro project/marketing site, not this browser client. |
| [Accessibility audit](../../accessibility.md) | Historical Fyne audit, not evidence of browser/Orca acceptance. |

The browser adapter, field-patch editing, ETags and accessible views below remain **planned requirements**. Some reusable Go progress/presentation/outcome foundations are now implemented, not the browser client itself. Scientific equations, result identities, evidence definitions and CLI behavior remain unchanged; Fyne workflows are deprecated compatibility, not the future primary UI.

## Architecture and scope

```text
One React app: home/docs, semantic model/result views, focus, tables, plots
    ├─ hosted + local browser features
    │    ↕ typed Worker requests / Go-owned results and speech
    │  Go/WASM Worker → CPU models, bounded in-memory studies/read-only imports
    └─ local build only
         ↕ authenticated same-origin API (never called by hosted build)
       Go server → granted project roots, large studies/reuse, reports, EDA tools
```

The target is one React application with browser-only and local-engine capabilities, as fixed by [ADR 0005](../../adr/0005-unified-react-hosted-and-local.md). The hosted build is static and does not connect to localhost; the local build embeds the same application and uses same-origin Go APIs. Go/WASM runs browser science in a Worker without filesystem authority. The research workbench remains the first delivery slice; the accepted browser-only model scope below is not a claim that every module view is implemented. Replace Astro with React home/docs and retire Fyne module-by-module after replacement qualification. No implementation, deletion, listener or deployment is performed in this documentation pass.

### Unified delivery and capability profiles

| Feature | Hosted static build | Local embedded build |
|---|---|---|
| Home, docs, glossary, citations, honesty audit | Available without engine; replaces Astro | Same views |
| Hysteresis/ISPP, small CPU crossbar MVM/IR-drop/sneak paths, peripherals, comparisons | Go/WASM Worker; bounded, seeded, in-memory | Same browser views/models; local capabilities additionally available |
| Small sweeps | Capped in Go before allocation; in-memory results | Same capability plus local large-study path |
| User-selected results bundle | Read-only, bounded Go decode/display; labelled imported/unverified | Same view; explicit server verification is a separate operation |
| Disk projects/edits, large sweeps, persistent reuse, stored-run verification | Instruction panel explains how to run `fecim serve`; no connection attempt | Same-origin server API within granted roots |
| EDA generation/external Yosys/OpenROAD/KLayout, application-requested file writes | Server required; no browser export/write path | Local server/tool capability, explicit authority |

Hosted disclaimer remains visibly present and accessible: “Computations run in your browser using the Go model. Project and result data is not uploaded. On-disk projects and edits, large sweeps, run reuse, stored-run verification, EDA and file writes require the local application: run fecim serve. This hosted build does not contact localhost.” Ordinary static asset requests still reach the hosting provider; do not claim zero network metadata. No analytics, data upload or remote computation.

The deployment profile is explicit build configuration, not inferred by probing the network or receiving an auth failure. Hosted mode omits local-engine transports and shows actionable instructions for unavailable features; it does not silently call a hardcoded URL. Local mode retains the approved token/Host/Origin/cookie/CSRF design without any external-origin CORS exception. Hosted state is ephemeral: refreshing may discard in-memory studies and selection, with an honest warning. No scientific data persistence or application file downloads in hosted mode.

All scientific calculations, defaults, validation, summaries, sorting/filtering/paging semantics and numeric speech/display are Go-owned: in the Worker for hosted operations, in the server/shared core where local operations require it. Browser-only selected-file viewing does not attest stored-run integrity or measured-device validation. Rendering, focus and message transport remain TypeScript/React responsibilities.

### Stack, contracts and build policy

- Go retains numerical models, defaults, validation, experiment identities, analysis, summaries and serialization. Share storage-independent operations between native and WASM paths; native persistence is an adapter, not a scientific fork. Thin HTTP/Worker adapters do not invoke CLI prose or reimplement equations. Legacy Fyne calls are historical compatibility, not a second target.
- React + TypeScript + Vite renders the browser workbench. Use React Aria Components where interaction needs accessible focus/keyboard machinery, and native HTML for simpler controls and read-only tables. Framework choice is not evidence of accessibility.
- Go API DTO definitions are authoritative. Generate OpenAPI from them and generate TypeScript request/response types and client bindings from that contract. Pin generators; CI regenerates and rejects drift. Go still validates every input at runtime. Keep presentation DTOs separate from immutable scientific records.
- Build the same React source twice: hosted-static and local-embedded profiles, sharing routes/components and differing in capability adapters. Production assets are built, not committed. CI runs locked frontend builds, generated API/Worker contract checks, freshness checks and the local Go `go:embed` build. Bundle Go/WASM and the matching compiler runtime support file with both profiles where browser computation is enabled. Pin Go/compiler-shim, Node, package manager, lockfile and generators; this document does not invent versions.
- The asset manifest records the frontend-input digest, API-contract digest, expected core compatibility identifier and built-asset fingerprint. The running browser compares its embedded manifest with authenticated health information. Core and frontend hashes need not equal each other; an unexpected pairing is a mismatch.
- Missing assets fail the browser distribution build. A plain `go build` is not itself a frontend build or freshness proof: release CI must reject stale inputs, and `serve` must check the embedded manifest against build-bound frontend inputs/API identity and refuse stale assets **before opening a listener**. A UI health warning additionally catches stale cached browser content. Never claim that `go:embed` alone detects stale assets. Exact build entry-point naming and pinned tool versions belong to implementation planning.
- Node is a development/build dependency, not a requirement for using the distributed binary. No CDN, remote fonts, analytics, service worker, hosted compute, or desktop wrapper is needed for v1.

### Local API contract

Every route below is **planned and local-build-only**, under `/api/v1`. The hosted build never calls these routes on localhost or an external compute service. Browser-only operations use a typed in-process Go/WASM Worker protocol, not this filesystem/job API. `P` denotes an opaque opened project handle and `E` an ephemeral local execution ID; neither is a scientific SHA-256 run ID. All data routes require authentication. Reads never create projects, run models, regenerate reports or export files.

| Method and route | Contract | Existing CLI / source operation |
|---|---|---|
| `GET /health` | App version, core fingerprint, frontend fingerprint, expected compatibility/build identifiers; no project paths/secrets. | New adapter operation |
| `GET /projects` | Discover project candidates only inside launch-granted roots; bounded/paginated discovery, no arbitrary filesystem browsing. | New adapter operation |
| `POST /projects/open` | Open a candidate using root ID + relative project path; return session handle. | CLI receives a directory directly; shared project loading |
| `POST /projects` | Create a validated named subdirectory under a granted root, with starter sweep options; refuse existing destinations. | `project init DIRECTORY`; `study.CreateStarterProject` |
| `GET /projects/P` | Configuration, revision, evidence and provenance. | Shared project loading; no dedicated read CLI |
| `POST /projects/P/validate` | Validate full bundle and return expanded point count/issues without rewriting it. | `project validate DIRECTORY`; `study.ValidateProject` |
| `PUT /projects/P/configuration` | Apply field changes to a named configuration document, guarded by bundle revision; retain original nodes/bytes as specified in section 3. | New shared editing operation; no current CLI equivalent |
| `POST /projects/P/close` | Revoke handle only, never delete data; `409` while an operation is active. | New session operation |
| `POST /projects/P/experiments` | Start asynchronous sweep with bounded workers and idempotency key; `202` + execution ID. | `experiment run DIRECTORY -workers N`; `study.Run` |
| `GET /projects/P/experiments?state=active` | Find active executions after reload. | New job registry |
| `GET /experiments/E` | Current/terminal execution snapshot; unknown or pre-restart execution returns `404`. | New lifecycle adapter |
| `GET /experiments/E/events` | Cookie-authenticated SSE; current snapshot on reconnect; bounded buffers, no durable event-log promise. | New runner progress hook/adapter |
| `POST /experiments/E/cancel` | Request context cancellation; acknowledgement is not terminal confirmation. | Runner context cancellation, already used by Fyne |
| `GET /projects/P/runs` | Server-side filtering, sorting, fixed-page pagination and stable run-ID tie-breaker. | `experiment.LoadRuns`; no dedicated list CLI |
| `GET /projects/P/runs/R` | Integrity-checked run and provenance by validated scientific run ID. | Existing stored-run loading/integrity operations |
| `GET /projects/P/analysis` | Read-only feasibility, Pareto, summaries, plot representation and 2–4-run comparison for explicit selected IDs. | `experiment.Analyze` plus new presentation operations; must not call report-writing wrapper on GET |
| `POST /projects/P/verify` | Explicit run-integrity/provenance review. | Fyne uses `experiment.LoadRuns`; no dedicated verification CLI |
| `POST /projects/P/reports` | Regenerate reports over stored results, not simulations. | `report generate DIRECTORY`; `study.GenerateReport` |
| `GET /projects/P/reports/F` | Download an enumerated JSON/CSV/Markdown report. | Existing report artifacts |
| `POST /projects/P/exports` | Export explicitly selected eligible run without rerunning models. | `export DIRECTORY -run RUN_ID`; `study.ExportRun` |
| `GET /projects/P/exports/R/A` | Download an enumerated artifact of a selected export; never accept a filesystem path. | Existing export artifacts; new download adapter |

One project-mutating operation runs at a time; conflicts return `409`. Closing a tab or dropping SSE does not cancel computation. Jobs/handles/sessions are memory-only and expire on server restart. After reauthentication and reopening the project, an old execution `404` takes the user to integrity-checked runs on disk, not a fabricated terminal result. Rerunning recomputes identities and reuses eligible committed runs. Retain completed execution snapshots in a bounded registry until server shutdown or documented capacity eviction. An evicted ID returns an explicit expired-state response; a fresh server returns `404` for old IDs. Neither response implies loss of scientific runs on disk.

Configuration GET/PUT responses return a strong quoted ETag over a versioned, unambiguous encoding of the three named YAML files' original bytes. `PUT` requires an exact `If-Match`: missing returns `428`, stale returns `412`. Do not accept `If-Match: *` as a concurrency bypass. A field patch changes one named YAML document per save and validates the proposed entire bundle. No-op requests still check the revision. Runs-list responses carry their dataset revision and query scope; a changed dataset resets pagination with an accessible explanation rather than mixing pages from different snapshots.

Errors use RFC 9457 `application/problem+json`: stable problem type, title, status, safe detail/instance, and field-issue extensions where applicable. Do not leak tokens, arbitrary host paths, raw stack traces or project contents in logs/errors. Typed model-domain failed runs remain scientific records, not generic HTTP failures. Citation access uses launch-authorized read-only resources; the browser cannot pass an arbitrary `citation-dir`.

### Authentication, confinement and threat model

The planned launch surface is `fecim serve --root ~/fecim-projects` (illustrative executable shorthand, not an existing command). Repeatable roots are granted **only at launch**. Bind exclusively to `127.0.0.1` on an available port and advertise one canonical origin, `http://127.0.0.1:PORT`. Reject an unexpected Host authority, including port; do not trust forwarded-host headers or enable wildcard listening/proxy mode. No CORS permission headers are emitted.

Bootstrap and session:

1. Generate a cryptographically random 256-bit one-time secret, valid for 60 seconds, and print the launch URL `http://127.0.0.1:PORT/#bootstrap=SECRET` to the terminal. Never log it elsewhere.
2. The minimal public bootstrap shell reads the fragment, immediately removes it with `history.replaceState`, and posts it in JSON to `POST /api/v1/session`. Fragments do not enter HTTP request URLs; do not put secrets in query parameters, paths, browser storage, referrers or SSE URLs.
3. The exchange requires the canonical Origin, JSON content type and `X-FeCIM-Request: 1`. Atomically consume the secret only on a valid exchange; reject replay/expiry. Issue an independent random cookie, then navigate to the token-free application.
4. Cookie is `HttpOnly`, `SameSite=Strict`, host-only, session-scoped with a per-launch name. Server-side validity ends at logout/shutdown. `POST /api/v1/session/logout` revokes it. SSE uses that cookie; EventSource needs no custom authentication header.
5. Renew an expired bootstrap link by explicit terminal action without killing experiments; never expose an unauthenticated token-renewal API. The bootstrap shell/minimal assets and exchange are the documented authentication exception and reveal no project data.

Every mutating request, including open/close/cancel/logout/bootstrap, requires exact canonical Origin, JSON content type and the custom header. Missing or `null` Origin fails. All data reads require the session; reject a foreign Origin when present and cross-origin Fetch Metadata requests, but allow authenticated same-origin GETs without an Origin header, including SSE. Host checks precede application access. Use request/worker/body/connection limits, timeouts appropriate to SSE and bounded event queues; a slow subscriber cannot block the runner.

Use escaped text, no raw project/report HTML insertion, no arbitrary script execution, same-origin embedded assets, a restrictive CSP, anti-framing policy, `nosniff`, no-referrer policy and no-store for sensitive/session responses. Downloads use enumerated artifacts and safe content disposition. Plain loopback HTTP does not claim TLS; require `Secure` cookies if a future HTTPS mode is designed. Cookies are not port-isolated: a hostile same-host local service and malicious same-user software remain outside this local-machine trust boundary.

Filesystem enforcement is end-to-end, not a handler-only string check. Resolve granted roots and acquire traversal-resistant root handles; open projects beneath them, then operate through project-scoped handles. Enforce confinement for configuration, input/citation reads, run artifacts, temporary files, atomic replacements, reports and exports. Reject escaping links, absolute paths, traversal, reserved names and ambiguous paths. Create accepts only a validated single-directory name. Do not authorize by prefix matching or `EvalSymlinks` followed by an unrestricted open. Existing path-based storage needs test-first adaptation to a patched supported Go traversal-resistant API such as `os.Root`, without changing scientific behavior.

| Threat | Defense / limit |
|---|---|
| Hostile website triggering local computation or reads | Session authentication, Host/Origin checks, custom mutation header, SameSite and Fetch Metadata; no CORS grant |
| DNS rebinding | Loopback-only listener and exact canonical Host |
| Traversal, escaping links, check/open races | Root-scoped filesystem operations through every storage path; not just initial validation |
| Untrusted project strings | Escaping, CSP, safe downloads, no HTML/script execution |
| Duplicate requests, stale edits and compute exhaustion | Idempotency, operation locks, revision checks and bounded admission/resources |
| Secret leakage | Fragment bootstrap, single-use expiry, redaction, no-referrer, no browser persistence |
| Malicious privileged/local software | Not a sandbox against same-user malware, hostile extensions, hard-link/mount manipulation or root; disclose rather than claim protection |

ETags and server operation locks protect participating requests, not arbitrary hostile editors. Recheck disk revision immediately before replacement and refuse observed conflicts; do not claim that a generic rename provides an atomic compare-and-swap against noncooperating external writers. The implementation plan must state supported external-edit coordination and test observed conflict handling.

Security references: [Go traversal-resistant APIs](https://go.dev/blog/osroot), [OWASP CSRF guidance](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html), and [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457.html). These are design references, not executed security validation.

The workbench slice includes project editing/local experiment lifecycle, paginated tables, evidence, 2–4-run comparison and supplementary plots, with capability gating in hosted mode. The accepted browser-only scope includes Go/WASM model visualizations and capped small studies; it does not authorize a TypeScript scientific implementation, new physics, general YAML editor, plot-driven selection, remote collaboration or a claim of full expert-module parity today.

## Go/browser ownership

Go supplies scientific values, unit conversions, rounding decisions, evidence labels, sorting/filtering, page numbers/counts, selected-run counts, comparison data, plot coordinates/domain bounds and experiment totals. It also parses and validates numeric input. The browser sends field paths and input strings or intent tokens (sort key, direction, next page, select run); it must not calculate scientific results, aggregate counts, reinterpret units or independently format numbers. Ordinary browser layout and focus mechanics are not scientific calculations.

Display and speech are two presentations of **the same Go-owned value**, not two independently computed measurements. Presentation-only fields must not alter immutable scientific content or run identity. Go retains canonical numeric values for sorting, never sorts formatted strings, and states any rounding; speech must not imply more precision than the displayed value. Comparison and plots use the same snapshot/version as the table.

## 1. Semantic navigation and focus contract

Each view has a unique, descriptive heading. Navigation uses semantic links/buttons, native form controls, tables with captions, scoped headers and named selection controls—not a custom spreadsheet widget. Visible focus and logical Tab order are required.

| Trigger | Focus destination | Required spoken context |
|---|---|---|
| Successful view change, including browser Back/Forward | New view heading, after its content is mounted | View title and heading role/level; e.g. “Compare metrics, heading level 1”. |
| Successful table page change | Table caption, programmatically focusable but not an extra ordinary Tab stop | “Runs, page 2 of 14, sorted by energy ascending”. Go supplies the actual position, sort and filtered scope. |
| Explicit sort/filter action that refreshes the table | Caption after the matching response arrives | Updated sort, page and active scope; reset to page 1 when Go changes the result set. |
| Failed form submit | Error summary, after errors render | “Save failed”, error count and field-specific reasons; summary links move focus to the corresponding field. |
| Background progress or data refresh | Keep current focus | Status is announced without moving the user out of a field, row or heading. |

An obsolete response must not steal focus or announce the wrong page after a newer navigation. Loading and errors leave an identifiable retry path; failed navigation does not pretend the requested page loaded. Error summaries retain entered values. Nonmodal background updates do not automatically open a view or focus an error summary.

The caption must really be reachable by programmatic focus and announced in the supported browser/Orca combination; visual proximity is insufficient. Verify that behavior manually. Avoid duplicate speech from a simultaneously focused heading/caption and redundant live-region echo.

## 2. Display and speakable values

For every presented numeric value, Go supplies its canonical identity/unit and these presentation fields:

- **display:** formatted value with unit, including any explicit approximation/rounding marker;
- **speakable:** complete value and expanded unit/exponent wording;
- **evidence:** the existing evidence label, with a human-readable expansion where needed;
- **availability/reason:** explicit missing, failed or inapplicable state instead of a fabricated zero.

Examples are acceptance fixtures, **not new device measurements**:

| Display | Speakable |
|---|---|
| `15 µC/cm²` | “15 microcoulombs per square centimeter” |
| `1e-12 A` | “1 times 10 to the minus 12 amperes” |
| `2e+3 Ω` | “2 times 10 to the plus 3 ohms” |
| `0 pJ` | “0 picojoules” |

The browser uses Go's speakable text for accessible names/readable cell content, while retaining metric, run and evidence context. Native table reading must expose the speech form too; an aria-label on a nonfocusable element alone is not sufficient proof. Do not hide table headers/children with blanket ARIA replacements or read the same value twice. Labeled fields retain their visible field name in the accessible name and associate units/help appropriately. Plot summaries and conflict values use the same speech contract.

Missing metrics say “Not available” and why; evidence is text, not just color or a badge icon. Unknown unit symbols must not silently pass through as if pronunciation were verified: Go supplies a defined expansion or explicit unsupported-unit status. Unit/exponent pronunciation is a mandatory manual acceptance task, not delegated to default synthesizer handling of Unicode or scientific notation.

## 3. YAML editing, preservation and optimistic concurrency

Implementation planning for the next Go-only slice is saved in the [configuration editing frontier](../plans/2026-10-25-browser-configuration-editing.md): `T-009-1bca6e59` prepares node-preserving validated edits; `T-010-1f9b8e44` adds revision-guarded, root-confined saves. These are planned tasks, not shipped editing, HTTP or React functionality. Raw field conflicts still require later Go presentation and browser/Orca qualification.

The browser submits **field changes**, never a whole YAML document or a browser-reserialized project. Each save identifies the project document, base ETag and allowlisted typed field paths with proposed input. Go parses values, applies accepted changes to its YAML node tree and validates the resulting full configuration before writing. Reject unsupported/ambiguous paths, invalid scalars, duplicate-key ambiguity and unsafe alias changes rather than reconstructing the document and losing data.

Go retains original bytes alongside the parsed nodes. Unexposed, schema-supported fields, ordering, comments and untouched nodes survive editing. This does not weaken strict schema validation or mean arbitrary unknown keys become runnable. Invalid/unsupported documents remain intact and produce an accessible error rather than being stripped into validity.

**No-op guarantee:** after checking that the ETag still matches, an empty patch or edits that resolve to no actual node changes return the original file bytes unchanged. Do not run an encoder and assume it is byte-preserving. Preserve line endings, quoting, whitespace, final newline and comments; do not rewrite the file on a no-op. A changed save may normalize layout through the node encoder, but must retain comments, unexposed field values and valid node relationships. Report save errors without a partial overwrite. Use guarded atomic persistence, detecting concurrent changes before replacement; a stale ETag is never permission to overwrite.

Required future automated tests start RED before production changes:

1. Load/save without edits is byte-identical (byte comparison, not parsed-tree equality), including inline/header comments, unusual spacing/quoting, CRLF and missing final newline.
2. An empty patch and a semantically unchanged field patch leave bytes and ETag unchanged.
3. Editing an exposed field preserves an unexposed supported field and comments; reload passes the existing strict validator.
4. Invalid/unknown paths, duplicate-key ambiguity and unsupported alias edits leave disk bytes unchanged.
5. A stale ETag, including on a no-op request, causes no write; a concurrent writer is not clobbered.
6. Validation/persistence errors do not partially replace the project document.

### Accessible ETag conflict comparison

A rejected stale save opens a conflict view and follows the heading focus rule. Go compares the intended field changes to the latest on-disk nodes, returning a **field-by-field text list** with accessible field labels, units and evidence where applicable. Example fixture:

> Energy limit: yours 50 picojoules; on disk 40 picojoules.

Distinguish missing/deleted values from zero. Optionally include the base value as additional text, but never require understanding a colored side-by-side YAML diff. Every conflicting field must be independently readable with Orca. Unrelated on-disk changes are preserved. Offer explicit reload/discard and review/reapply choices; no silent force-save. Reapply uses a freshly reviewed ETag and normal validation; another disk change causes another conflict. Keep the user's attempted changes available until they choose to discard them.

## 4. Table-first results and compare metrics

The run table is the primary selection and inspection surface. It provides run identity, execution status, feasibility, Pareto status, metrics with units/evidence, and detail links. Filters and sort are explicit. Pagination caption includes page N of M and sort; filtering does not silently make a plot look like the full experiment.

Selection is keyed by immutable run ID, not row number or sort position. It survives pagination/sort; selected runs hidden by a filter are listed in a readable selection summary with remove controls. Go supplies selected count and permissible actions.

**Compare metrics (v1):** select **2 to 4 distinct runs**, then activate “Compare metrics”. Show a native table with **metrics as rows and runs as columns**. Column headers identify runs unambiguously, with full IDs available; row headers include metric name and unit. Each value includes its evidence label and speech form. Include failed/missing values explicitly as “Not available” with reason; never coerce them to zero. Do not compare incompatible units as equal; Go either performs a declared conversion or marks the metric incomparable.

One selected run is insufficient; a fifth selection is refused without removing the first four. Announce the count/range explanation and keep focus on the originating control. Comparison uses stored results; it does not rerun models or calculate browser-side differences. The comparison view follows the view-heading focus rule.

## 5. Plot scope and keyboard boundary

Plots are optional, supplementary views of the table data, not the only route to a result. Each plot names its axes/units, experiment/snapshot, filters, count and scope: full filtered result set versus current page must be explicit. V1 defaults to the **full filtered result set**, not just the visible page; any page-only plot is labeled “current page only”. Missing/unplottable runs are counted and explained by Go.

**At most one sequential Tab stop per plot in v1**, and zero is allowed for a noninteractive figure. No per-point or per-series focus targets. Selection happens in the table and the plot mirrors it; neither keyboard nor pointer plotting interactions introduce a second selection state. Its accessible text summarizes scope and mirrored selection and directs users to the table for values. Legends or export controls, if present outside the plot, are ordinary clearly named controls—not hidden plot navigation. Do not disguise a collection of point controls as one plot stop.

## 6. Experiment lifecycle and announcements

Go owns the authoritative lifecycle and counters. Browser navigation/request abort is not experiment cancellation. A Cancel action must reach the server-owned job, and its receipt is not terminal confirmation.

SSE progress payloads carry only `total`, `completed`, `reused`, `failed`, and `cancelled` counts. `completed` includes every committed/reused resolved point, including failed design points; `reused` and `failed` are subsets and must not be added to `completed`. `cancelled` counts unresolved points abandoned when cancellation has settled; pending is `total - completed - cancelled`. Systemic failure is a separate lifecycle outcome, not invented failed scientific runs. Lifecycle and analyzed terminal details are retrieved from the state endpoint following notification/stream close. The UI may poll state at a bounded rate while cancelling/reconnecting.

Progress is coalesced to at most one ordinary live-region announcement per five seconds, with start and terminal states announced immediately. Intermediate snapshots can differ with worker scheduling despite being counts-only: test bounds/monotonicity, final counts and unchanged run identities/reports, not byte-identical SSE histories. Final analysis supplies feasible/Pareto counts; the browser never derives them from progress. SSE disconnection is not proof of completion; fetch authoritative state. A slow consumer can receive a newer snapshot rather than an unbounded backlog.

| State / event | Required behavior |
|---|---|
| Run accepted / running | Announce start once; progress is bounded/coalesced, not a flood of per-run speech. Keep the current focus. |
| Cancel accepted | Announce “Cancellation requested”; show nonterminal cancelling state, avoid duplicate cancel requests, and retain access to committed results. |
| Cancellation finished | Only after Go confirms workers have stopped and writes have settled/reconciled, announce “Experiment cancelled” with actual committed-run/feasible/Pareto/failure counts and remaining-not-run context. |
| Normal completion | Announce, for the example fixture, “Experiment finished: 8 runs, 6 feasible, 1 Pareto, 0 failed”. Every count comes from Go's final snapshot. |
| Systemic/report failure | Announce a failure with reason and available committed results, not “finished” as if successful. |

“Runs” in the completion summary means committed/reused records in that experiment's analyzed result set, including design-point failures; it does not mean newly evaluated records only. Go also supplies planned/reused/remaining context. Feasible and Pareto counts follow core analysis, not colored plot marks. A normal experiment may finish with design-point failures; announce its nonzero failed count. A cancellation summary is explicitly partial and never counts unevaluated work as failed.

Cancellation racing with completion reports the actual terminal outcome, without claiming both cancelled and finished. A terminal event is announced once per job terminal revision; reconnect/repeated snapshots must not repeatedly interrupt the user. Status remains readable after the announcement. This design requires exposing/reconciling partial committed results beyond the current `study.Run` error outcome; it is not an assertion that the existing wrapper already does so.

## Verification and review gates

Before implementation, review this document and its [numbered manual acceptance scripts](2026-10-25-accessible-browser-workbench-orca-acceptance.md). No personal Orca usage or completed accessibility test is attested by the authoring agent.

Future automated verification must cover presentation/unit/exponent forms, page captions and focus targets, stale navigation responses, validation summaries, field-only patches and no-op byte preservation, ETag races and textual conflicts, 2–4-run comparison, plot Tab-stop limits, authoritative cancellation, and terminal counters/deduplication. Maintain core deterministic outputs and existing scientific evidence boundaries. Automated accessibility scans cannot substitute for Orca task completion.

**Manual release gate:** qualify both hosted and local builds on Orca + Firefox/Linux and keyboard only, using the companion's explicit capability-profile applicability and hosted tasks. Record required/actual speech, focus and pass/fail separately per build; a local pass never certifies hosted WASM behavior. Any failed or unrun applicable required task blocks an accessibility claim. Exact native/WASM example results.json parity is a separate mandatory automated gate, now implemented and run across native Go, Node Go/WASM and headless Firefox Worker; see the validation receipt and feasibility assessment. This does not attest production Worker integration, HZO-loop parity or human Orca acceptance. Tester approval remains a human decision.

TDD: **N/A for this documentation-only consolidation**; it changes no runtime behavior. All implementation and automated tests above are future requirements, not work performed in this change.

Update this spec and its task scripts together when value DTOs, unit vocabulary, table/compare scope, YAML edit semantics, lifecycle events or navigation change. Implementation planning/execution requires a separate user instruction after written-design review.
