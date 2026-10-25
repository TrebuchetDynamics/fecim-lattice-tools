# Goal 1 delivery — security follow-up

Date: 2026-10-25. Preparation delivery base: `86b9f1ce` on `delivery/goal1` (PR #23). The owner approved updating the Go toolchain and vulnerable dependencies. This is a new delivery receipt, not a replacement for the historical preparation evidence.

## Bounded changes

- Require patched Go **1.26.10** in `go.mod`; both CI jobs and Nightly read that file. Go 1.26 is a supported release line under the [official release policy](https://go.dev/doc/devel/release#policy). The language/toolchain minimum increases deliberately; older Go installations with automatic toolchain selection enabled download the required version.
- Upgrade `golang.org/x/image` to **v0.45.0** and `golang.org/x/net` to **v0.60.0**. Their module graph selects `x/sys v0.48.0` and `x/text v0.42.0`. No simulator source, physics constants, reference samples, or acceptance tolerances change.
- Install Wayland and xkbcommon development libraries in CI/Nightly. The existing GLFW 3.4 dependency builds both X11 and Wayland by default; PR CI job `121294900895` failed on missing `wayland-client-core.h`. The required external Verilog installation/version/test contract remains intact.
- Upgrade the static marketing site from Astro 6 to **7.4.2**, with a regenerated npm lockfile. Semver-compatible updates left three advisories including critical Astro issues; the fixed release requires the major upgrade. Follow the [official Astro 7 migration guide](https://docs.astro.build/en/guides/upgrade-to/v7/). Retain `output: 'static'` and explicitly preserve HTML-aware whitespace via `compressHTML: true`. This does not implement a browser workbench or authorize a deployment.
- Keep the Playwright preview process in the foreground with `--ignore-lock`. Astro 7 can auto-background under detected coding agents; Playwright must own startup and shutdown rather than silently reuse an orphan server. The tests and CI non-reuse policy are unchanged.

## RED → GREEN

- **RED:** `make govulncheck` on the delivery base reported **24 reachable vulnerabilities** in the installed Go 1.26.2 standard library and x/image. **GREEN:** the same command after the updates reports **No vulnerabilities found**. The initial v0.43.0 candidate still had a module-only x/image advisory; v0.45.0 removes it too.
- **RED:** `npm audit` on the original lockfile reported **13 vulnerabilities**, including one critical. A compatible-only audit fix left three, requiring Astro 7. **GREEN:** `npm audit` after a clean `npm ci` reports **0 vulnerabilities**.
- **RED:** `CI=1 DEBUG=pw:webserver npm run test:e2e` with a fresh, unavailable port failed because the preview launcher exited after backgrounding. **GREEN:** with `--ignore-lock`, the same command proves cold startup, **16 passed / 4 expected viewport-specific skips**, and server shutdown; port 4321 is free afterward. An earlier run that reused an already-running preview server is not the cold-start certificate.
- **RED:** public PR CI failed on missing Wayland headers. Local compilation uses the required headers; the revised hosted job must pass before merging. The first local workflow edit also failed the existing exact external-Verilog installation contract; preserving its original command and installing the additional libraries separately restores the contract without weakening it.

Dependency lockfile regeneration and documentation are metadata-only; no new simulator behavior is introduced. Existing automated security gates and browser tests provide the failing checks above; no gate is disabled or made non-fatal.

## Local verification

On the final dependency/configuration tree:

- `make ci`: **PASS** (format, vet, security gate, short suite, architecture/build, research tests, citation registry audit).
- `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`: **PASS**.
- `go mod verify` and `go mod tidy -diff`: **PASS**.
- `npm ci && npm audit && npm test && npm run build` in `web/`: **PASS**, zero npm vulnerabilities.
- `CI=1 DEBUG=pw:webserver npm run test:e2e` in `web/`: **PASS**, 16 passed / 4 expected skips, fresh foreground preview and shutdown.
- Adapter sync and all eight installer tests: **PASS**.
- Display-tagged screenshotter suite, including the four automatic-display cleanup failure paths: **PASS**.
- `go test -race -short ./shared/... ./validation/...`: **PASS** on the initial security candidate; x/image subsequently advanced from v0.43.0 to v0.45.0. The final dependency tree requires hosted race verification; do not misattribute the earlier local race result to the final image version.
- Independent read-only review: no remaining blockers after reviewing the final x/image and foreground-preview changes. Review independently confirmed module verification/security scan and the cold-start browser receipt.

Local logs are retained outside the source tree under `/tmp/fecim-goal1-security-*`. Hosted PR/main checks are separate evidence, not implied by local passes. Goal 1 is complete only after the public root is clean, the Park correction is public, and CI is green on main. No release, DOI, new experimental validation, Orca pass, submission, or outreach is claimed here.

## Original checkout isolation

The following pre-existing paths remain untouched in the original checkout and are excluded by the approved five-topic delivery scope:

- `.gitignore`
- `README.md`
- `docs/internals/CONTEXT.md`
- `docs/internals/architecture/ARCHITECTURE.md`
- `docs/internals/superpowers/specs/2026-07-19-research-design-workbench-architecture-design.md`
- `docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md`
- `cmd/fecim-lattice-tools/eda_full_app_race_test.go`
- `docs/internals/superpowers/plans/2026-10-25-browser-configuration-editing.md`

The divergent local main and its later HZO/browser/autowork commits are not merged by this delivery.
