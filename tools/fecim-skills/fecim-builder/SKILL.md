---
name: fecim-builder
description: Runs build flows for the canonical Fyne shell and module commands on this Go 1.25 monorepo. Use when building, packaging, or debugging build failures in cmd/fecim-lattice-tools or module GUI commands.
---

# fecim-builder


See `tools/fecim-skills/_shared/fecim-context.md` (Build target matrix) for the canonical CGO/entry-point mapping.

## Workflow

1. **Identify the target** — ask the user if unclear:
   - Canonical Fyne shell: `cmd/fecim-lattice-tools`

2. **Set the build environment:**
   - Canonical: `export CGO_ENABLED=0`.

3. **Run the build:**
   - Canonical single-binary: `CGO_ENABLED=0 go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools`
   - Canonical launch: `./launch.sh`
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
   - For the canonical app, confirm `CGO_ENABLED=0` was respected (`go env CGO_ENABLED` should print `0` in the same shell).

## Verification

- Input: "Build the legacy GUI."
- Input: "Build the canonical shell." with missing libvulkan-dev.
  Expected: succeeds since libvulkan is optional, OR triages with the table above.

## TDD

Build invocations are observation, not behavior change — `TDD: N/A`. Any code change discovered during triage triggers the project's TDD hard-rule. See `tools/fecim-skills/_shared/tdd-evidence-template.md`.
