+++
schema = 1
id = "HYS-UI"
title = "Replace Astro with the shared interactive React Hysteresis application"
short = "React Hysteresis in one app"
status = "done"
goal = "HYS-REACT"
section = "Next"
rank = 5000
depends_on = ["HYS-WORKER"]
touches = ["web/src/", "web/index.html", "web/package.json", "web/package-lock.json", "web/tsconfig.json", "web/vite.config.ts", "web/astro.config.mjs", "web/public/_redirects", "web/test/landing-contract.test.mjs", "web/test/e2e/", "web/playwright.config.mjs", "web/README.md", "web/wrangler.toml", ".github/workflows/ci.yml", "docs/internals/testing/wasm-parity-gate.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Implement one locked React/TypeScript/Vite web app and route/component registry replacing Astro home, retaining useful public content/citations/deployment behavior. Render CORE, PROGRAM and ANALYSIS controls/results from actual Go Worker envelopes: all implemented modes, Go-derived summaries/units/availability, stateful loop, programming/calibration, diagnostics/context, native paged tables and bounded supplementary plots. Produce hosted and local profiles from identical views; local-only buttons remain adapter callbacks until HYS-LOCAL. No fake numerical fixtures on happy paths."
acceptance = ["Given real production Worker data, when users change material/model/drive, run/pause/reset/cancel/program/calibrate/analyze and page results, then shared Hysteresis controls act on Go state with correct revisions, all retained values/Go speech are accessible and stale/error/unavailable states are explicit; test/e2e/hysteresis.spec.ts exercises all workflow families.", "Given keyboard/narrow-width/zoom and background progress, when navigating inputs/results/advanced sections, then native labels/headers/page captions/focus persist, no plot-only action exists, no row virtualization or sample announcement flood occurs; hysteresis.spec.ts checks DOM/focus/automated accessibility, without claiming Orca acceptance.", "Given either build profile and every hosted action/fault path, when assets load and computations/server-only controls run, then there is one route tree/home, no TS science/default duplication, no hosted private-network request/scientific upload/persistence/download and actual route payload/cold-result guardrails pass; hosted-isolation.spec.ts and updated landing-contract.test.mjs retain independent content/deployment/no-Fyne-WASM guards."]
checks = ["cd web && npm ci && npm test", "cd web && npm run typecheck && npm run build:hosted && npm run build:local", "cd web && npm run test:e2e -- test/e2e/hysteresis.spec.ts test/e2e/hosted-isolation.spec.ts", "env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1"]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HYS-UI.toml"
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: One React route tree with hosted/local build outputs, working production Hysteresis UI, locked build/test scripts and dual-profile CI/actual route-budget/isolation evidence. Sole React route/component registry owner.
- Consumes: HYS-WORKER production typed envelope, Go catalogue/capabilities, canonical/display/speech/revision pages and build/runtime fingerprints. Preserves its Worker shim; UI lifecycle client is transport/focus only.
- Repo obligations triggered: ADR one UI, hosted isolation, budget measurement, dual-profile CI and Orca limitations. Replace explicit Astro-only assertions in web/test/landing-contract.test.mjs:14–41 while retaining useful content/deployment/no-Fyne-WASM invariants; update existing browser consumers instead of deleting safety tests. Protected package/lock/tsconfig/Vite/CI/wrangler paths explicitly declared. Build scripts use pinned Go/runtime and generated assets, no runtime Node for end users. Keep docs truthful about unimplemented local adapter until HYS-LOCAL.
- Next action: Update landing-contract.test.mjs to expect one React app while preserving its independent safety/content assertions; run expected RED, then implement minimal shell with a real Go loop before advanced panels.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.
- Pending live: Real Orca + Firefox/Linux keyboard/speech qualification remains NOT RUN for hosted and authenticated local profiles (HYS-ORCA). Automated DOM/focus/keyboard/accessibility evidence is not Orca acceptance.
