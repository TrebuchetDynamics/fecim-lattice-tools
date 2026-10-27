# ADR 0005: One React application, hosted and local builds

Status: Accepted product/architecture direction from the owner's explicit scope change; implementation and commits are not authorized by this ADR. Supersedes the separate Astro-site/React-workbench direction and the earlier WASM-deferred decision. Does not erase completed work or its evidence.

## Decision

Maintain one React + TypeScript + Vite application and route/component/accessibility system, built with two deployment profiles:

- **Hosted static:** public home/landing content and browser-only research features. Physics, scientific defaults, calculations, analysis and numeric speech/display remain Go, compiled to `GOOS=js GOARCH=wasm` and run in a Web Worker. No hosted computation service and no localhost connection/probing, CORS pairing or automatic local-engine discovery. A feature requiring the server presents local-run instructions, not a failed loopback request.
- **Local:** the same React application embedded in the Go binary, served by the planned `fecim serve` command. Full supported local capabilities use same-origin Go APIs and the existing loopback Host/Origin, session/bootstrap, CSRF and root-confinement design. Browser-only capabilities remain available here too; do not build a second UI for them.

Replace the Astro application with the unified React application's home view, retaining useful public content/citations/links. Retire deprecated Fyne module views as their React replacements meet the agreed behavior/accessibility parity, then delete the retired Fyne UI code. This latest owner instruction reinstates module-by-module replacement qualification as a removal condition; it does not make Fyne a new-feature target or authorize deletion during design work. Preserve the scientific Go core/CLI and completed IDs/receipts.

## Capability boundary

**Browser without local engine:** docs, glossary, citations, honesty audit; hysteresis/ISPP visualization; small CPU crossbar MVM/IR-drop/sneak-path models; peripheral models; comparison views; capped small in-memory sweeps; read-only viewing of a user-selected result bundle. Selectable files are explicit user inputs, not directory access. No TypeScript numerical model, scientific default, independent metric formatting or derived-physics implementation.

**Local engine:** on-disk projects/edits, large sweeps, persistent run reuse, stored-run verification, EDA generation/tool flows and all application-requested file writes. Hosted studies/results stay in memory: no automatic download, IndexedDB/OPFS/localStorage scientific persistence, filesystem API, report export or writable project bundle. Test comparison can serialize results.json into memory without writing a user file. Ordinary browser caching of static application assets is not scientific data persistence.

A persistent visible, screen-reader-readable hosted notice must say what runs in the browser, what requires `fecim serve`, and that project/measurement/result data is not uploaded. Explain that ordinary requests for static site assets still occur; do not promise the hosting provider sees no HTTP metadata. Do not include third-party analytics or remote computation. Deliberately following citation links is navigation, not scientific data upload.

## Execution and identity

Use a typed Go/WASM worker adapter for inputs/outputs, errors, count progress and cancellation; React owns presentation/focus, not physics. Enforce work/size/dimension/step caps in Go before allocation; TypeScript may expose them but is not the authority. The exact performance-supported caps require measurement before enabling a model; no unmeasured responsiveness claim is accepted. Long CPU loops must yield at bounded work boundaries; terminating/restarting the worker is a hard-cancel recovery path for ephemeral computations, not confirmation of local-server cancellation or deletion of persisted runs.

Keep native persistence/root handles, external tools, GPU/Vulkan and file logging out of the hosted execution path. Browser adapter needs data/byte-oriented configuration/citation input and an in-memory study path reusing current validation, Expand/ID/Evaluate/Analyze and report serializers. Do not duplicate equations or bypass failed-result validation. Imported results are displayed as imported/unverified; opening a file is not the server's stored-run integrity verification operation.

## Required acceptance

1. **Exact parity:** the checked-in example study produces byte-identical `results.json` through the existing native study/CLI path and browser WASM worker. Use identical configuration, input bytes/hashes, seeds, evaluator/model/schema versions and repository revision; compare original JSON bytes, not parsed objects, rounded numbers or tolerance envelopes. No target-specific field stripping. Fix shared calculation/serialization differences under scientific regression review; never weaken this gate silently.
2. **One UI:** hosted and local builds share views and accessible semantics; deployment capability adapters differ, not route/component implementations. Home replaces the Astro landing page. No Fyne removal is declared complete before its replacement is qualified and code is actually deleted in a separately authorized change.
3. **Hosted isolation:** automated browser requests must show no localhost/loopback/private-network probing and no scientific data upload during import, calculation or server-only actions. Missing-server actions remain navigable instructions. No hidden localhost fallback inside WASM, JS or a service worker.
4. **Accessibility:** Orca + Firefox/Linux and keyboard-only acceptance apply to both builds within their capability profiles, with identical table-first/speakable-values/focus contracts. WASM computation must not freeze the main UI or flood announcements.
5. **Build coherence:** embed and serve the same pinned Go WASM/runtime support version and frontend/API/worker contract identifiers. Build both profiles in CI; asset drift or an incompatible worker/core/UI pairing cannot ship silently. No runtime Node installation is required by users.

## Alternatives rejected

- Separate marketing site and app: duplicates navigation/content and contradicts the one-UI requirement.
- Hosted page connecting to localhost: contradicts the explicit latest decision and would require a different cross-origin security design.
- Physics/defaults in TypeScript: creates scientific model drift and another provenance authority.
- Native Go persistence/CLI entry points copied directly into browser execution: compilation alone does not provide browser filesystem/process capabilities.
- GPU/WebGPU rewrite or a different WASM compiler now: not needed to reuse existing CPU Go models; no performance evidence justifies this new kernel/semantic surface.

## Evidence and limits

See [WASM feasibility assessment](../architecture/browser-wasm-feasibility.md) for package compile checks, source findings and existing native/Node-WASM test evidence. The [validation-gate receipt](../testing/wasm-parity-gate.md) now records exact example-report parity natively, in Node Go/WASM and in a real headless Firefox Worker, plus measured payload/timing and proposed budgets. This is validation-only execution, not the production Worker/input API. Production integration, HZO-loop parity, actual hosted load budgets and dual-build Orca acceptance remain unimplemented/unverified. The [browser design](../superpowers/specs/2026-10-25-accessible-browser-workbench-design.md) and [manual acceptance gate](../superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md) own interaction/security requirements.

This decision authorizes documentation of accepted intent only. No dependency, manifest, fixture, test implementation, Astro/Fyne deletion, deployment, service startup or commit is performed by this decision record.
