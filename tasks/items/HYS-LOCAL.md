+++
schema = 1
id = "HYS-LOCAL"
title = "Serve the same React Hysteresis UI with authenticated confined file operations"
short = "Local launch and safe file workflows"
status = "done"
goal = "HYS-REACT"
section = "Next"
rank = 6000
depends_on = ["HYS-UI"]
touches = ["workbench/server/", "internal/browserui/", "cmd/fecim-lattice-tools/subcommands.go", "cmd/fecim-lattice-tools/serve.go", "cmd/fecim-lattice-tools/serve_test.go", "cmd/fecim-lattice-tools/README.md", "web/src/local/", "web/src/hysteresis/", "web/test/e2e/hysteresis-local.spec.ts", "web/package.json", "web/package-lock.json", "Makefile", ".gitignore", ".github/workflows/ci.yml", "docs/guides/installation.md", "docs/internals/testing/wasm-parity-gate.md"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Add fecim-lattice-tools serve --root DIRECTORY as an additive command, embedded local profile/assets, matching fingerprints and same-origin file adapter. Follow accepted bootstrap/session/Host/Origin/CSRF/root-handle design. Explicitly load/save existing supported preset/calibration formats and export enumerated Go-rendered PE/diagnostic/log artifacts; no auto-save or arbitrary SVG/HTML payload. Hysteresis computation stays the same browser Worker; supplied snapshots are labelled browser simulation data, not verified persisted studies. Generated browser assets are build outputs, not committed binaries."
acceptance = ["Given built local assets and a private ephemeral port, when serve starts/bootstrap exchanges/logout/replay/foreign Host/Origin/CSRF/oversize requests occur, then only the exact loopback origin/session is admitted, tokens never enter URLs/logs/storage after fragment removal and incompatible/missing asset fingerprints fail clearly before operations; workbench/server/server_test.go::TestHysteresisLocalSession and cmd serve tests.", "Given a granted temporary root and valid supported typed snapshot/preset/calibration, when explicit local save/load/export runs, then existing formats/Go serialization are preserved and traversal/symlink/absolute/outside-root/overwrite conflicts fail without writes outside grants; TestHysteresisLocalFiles checks enumerated kinds, 8 MiB pre-read cap and provenance, not client verification claims.", "Given hosted/local builds from the same candidate, when real browser file workflows/launch/error/conflict actions execute, then identical Hysteresis components provide truthful accessible confirmations/local instructions, local writes need explicit action and hosted constructs no local transport; test/e2e/hysteresis-local.spec.ts plus existing CLI/repo-surface guards keep all old defaults working."]
checks = ["env CGO_ENABLED=1 xvfb-run -a go test ./cmd/fecim-lattice-tools ./workbench/server ./internal/browserui -count=1", "env CGO_ENABLED=1 go test -race ./workbench/server -count=1", "cd web && npm run build:hosted && npm run build:local && npm run test:e2e -- test/e2e/hysteresis-local.spec.ts", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m"]
exclusions = ["Go owns science/defaults/display/speech; simulation-only evidence; no hosted scientific upload/persistence/localhost probing; existing CLI/Fyne/default builds remain. No launch-default change, Fyne deletion, new physical model, golden/tolerance weakening, deployment or service changes.", "No public deployment/push, shared service changes or ungranted external operations; test servers use private ephemeral ports."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/HYS-LOCAL.toml"
+++

- Sources: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract; docs/internals/adr/0005-unified-react-hosted-and-local.md
- Produces: Additive serve command, embedded local profile with typed authenticated file API, enumeration/root confinement/security tests and actual local-profile integration evidence. No default-launch switch.
- Consumes: HYS-UI same route/components and local adapter callbacks/build outputs; HYS-WORKER fingerprints; CORE/PROGRAM/ANALYSIS typed artifacts and Go serialization. Reuse existing traversal-resistant workbench root/storage primitives.
- Repo obligations triggered: Design Authentication/confinement/threat model and ADR build/profile boundary; installation/command help names actual additive command, no claim shorthand is an existing executable. Retain TestDefaultRepoGraphExposesFynePackages/TestGoFilesDoNotCarryLegacyFyneBuildTags and all old entry points. Explicit protected Makefile/gitignore/manifests/CI writes include generated asset staging/embedding; plain Go builds must compile and report browser-assets-unavailable rather than require generated files. Unit tests inject an in-memory fs; actual browser integration builds and embeds real assets. Private ports/scratch only; no production listener/service changes.
- Next action: Write TestHysteresisLocalSession for missing/drifted assets and rejected foreign Origin before implementing serve; run RED. Then use temporary granted roots for the file/security cases.
- Verification status: New tests/commands are proposed/unrun; no implementation evidence or manual qualification is claimed by planning.
- Pending live: Real Orca + Firefox/Linux keyboard/speech qualification for both hosted and authenticated local profiles remains NOT RUN; follow docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md. Automated evidence does not replace that human gate.
