---
name: fecim-builder
description: Use when building, packaging, or debugging build failures for FeCIM's React web target, existing Go core/CLI, or deprecated Fyne compatibility commands in this Go 1.25 monorepo.
---

# fecim-builder


React web is the primary UI target; Fyne is deprecated compatibility, not the direction for new UI work. The React client is planned, not an already-shipped distribution. Preserve the Go core/CLI and working Fyne entry points until replacements are separately verified.

See `docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md` (Stack, contracts and build policy) for the approved browser direction, and `tools/fecim-skills/_shared/fecim-context.md` (Build target matrix) for existing Go CGO/entry-point mapping.

## Workflow

1. **Identify the target** — ask the user if unclear:
   - Primary UI target: React web. Inspect current implementation and documented build tooling first; do not invent a React package/build command or treat `web/` (the Astro marketing site) as the client.
   - Existing Go core/CLI and deprecated Fyne shell: `cmd/fecim-lattice-tools`.

2. **Set the build environment:**
   - Existing Go command below: `export CGO_ENABLED=0` (not a frontend build setting).

3. **Run the build:**
   - Existing Go/CLI and deprecated Fyne compatibility binary: `CGO_ENABLED=0 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`
   - Deprecated Fyne compatibility launch: `./launch.sh`
   - Whole repo: `go build ./...`
   - Preflight first: verify repo path, `go`, `git`, and `rg`/fallback using `_shared/fecim-context.md`.

4. **On failure, triage:**

   | Symptom | Cause | Fix |
   |---|---|---|
   | `fatal error: GL/gl.h: No such file` | Missing OpenGL headers | Report blocker with command output; operator/admin installs packages. |
   | `cannot find -lvulkan` | Vulkan loader missing | Report blocker for legacy/Vulkan path; optional dep may be omitted when not required. |
   | `gcc not found` | CGO toolchain missing | Report blocker with `command -v gcc` evidence; operator/admin installs packages. |
   | `package github.com/Fyne: cannot find module` | Fyne import in non-shell pkg | UI-boundary violation; move logic to `shared/viewmodel/` per AGENTS.md |

5. **Verify:**
   - Binary exists and is executable.
   - For the existing Go command, confirm `CGO_ENABLED=0` was respected (`go env CGO_ENABLED` should print `0` in the same shell).
   - A Go build alone does not build or verify React assets or prove a browser distribution is ready.

## Verification

- Input: "Build the legacy GUI."
- Input: "Build the existing Go/CLI binary." with missing libvulkan-dev.
  Expected: succeeds since libvulkan is optional, OR triages with the table above.
- Input: "Build the React workbench."
  Expected: checks current implementation/tooling and reports any missing browser build path, without substituting a Go-only build or the Astro marketing site.

## TDD

Build invocations are observation, not behavior change — `TDD: N/A`. Any code change discovered during triage triggers the project's TDD hard-rule. See `tools/fecim-skills/_shared/tdd-evidence-template.md`.
