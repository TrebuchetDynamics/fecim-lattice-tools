# Contributing to FeCIM Lattice Tools

Thanks for contributing to **FeCIM Lattice Tools**. This project is a Go-based educational/simulation suite for ferroelectric compute-in-memory (FeCIM) concepts.

This guide describes how to build, test, validate physics behavior, and submit high-quality changes.

---

## Project Structure

The repository is organized around **7 modules** plus shared infrastructure:

- `module1-hysteresis/` - Ferroelectric hysteresis engines (Preisach, LK), ISPP/WRD flows, material behavior
- `module2-crossbar/` - Crossbar array behavior, MVM paths, non-idealities (IR drop, sneak effects, drift)
- `module3-mnist/` - MNIST inference demos and CIM-vs-reference pipeline checks
- `module4-circuits/` - Peripheral read/write chain, ADC/DAC/TIA behavior, circuit-level modeling
- `module5-comparison/` - Technology comparison views and modeled trade-off reporting
- `module6-eda/` - EDA-oriented generation/export, layout/routing foundations, integration helpers
- `module7-docs/` - In-app documentation browser, curriculum organization, discoverability tooling

Shared and cross-cutting directories:

- `shared/` - Reusable packages (physics, CLI helpers, UI/common widgets, IO, logging, validation helpers)
- `validation/` - Integration/regression validation suites and golden test data
- `config/` - Configuration files and physics/config validation inputs
- `cmd/fecim-lattice-tools/` - Main unified application entrypoint

---

## Build

Build the main application from repo root:

```bash
go build ./cmd/fecim-lattice-tools/
```

---

## Test

Run full test suite:

```bash
go test ./...
```

Run with race detector (required before claiming concurrency-sensitive fixes):

```bash
go test -race ./...
```

When changing specific module behavior, add/adjust targeted tests near affected packages in addition to running the full suite.

---

## Code Style and Static Checks

Before opening a PR, ensure:

1. Code is formatted with `gofmt`
2. Static analysis passes with `go vet`
3. Tests pass (`go test ./...` and `go test -race ./...`)

Recommended commands:

```bash
gofmt -w .
go vet ./...
go test ./...
go test -race ./...
```

Guidelines:

- Prefer small, focused functions with clear names.
- Keep module-specific logic inside its module; move shared behavior into `shared/`.
- Avoid silent failure paths; return actionable errors with context.
- Preserve deterministic behavior in tests (seed randomness explicitly when needed).

---

## Physics Validation Requirements

Physics-facing changes (models, equations, calibration, tolerance windows, default parameters, units, conversion logic) must include validation evidence.

### Required practices

- **Golden files**: use/update golden references in `testdata/` where applicable.
- **Tolerance checks**: assertions should compare measured vs expected values using explicit tolerances.
- **Unit-aware output**: report values with units and deltas when relevant.

Example reporting style:

- `P_r = 24.7 uC/cm^2 (expected 25.0, delta 1.2%, within 5% tolerance)`

### When changing expected outputs

If a golden file changes, include in PR description:

- Why the change is scientifically/model-wise correct
- Which tests/regressions were updated
- Previous vs new expected behavior and tolerance rationale

---

## Pull Request Guidelines

Please follow this checklist for every PR:

1. **Scope clearly**: one logical change set per PR.
2. **Describe intent**: include problem, approach, and impacted modules.
3. **Link evidence**: include test output and, for physics changes, validation evidence.
4. **Update docs**: keep docs/help text/tooltips aligned with behavior changes.
5. **Keep CI green**: no failing tests, vet issues, or race regressions.

Suggested PR template content:

- Summary of change
- Modules/directories touched
- Test commands run and results
- Physics validation notes (if applicable)
- Follow-up work (if any)

---

## Commit Message Conventions (used in this project)

This repository follows a conventional, action-oriented style visible in history.

### Preferred format

```text
<type>(<optional-scope>): <imperative summary>
```

Common types used in this repo:

- `feat` - new functionality
- `fix` - bug fixes/corrections
- `perf` - performance improvements
- `test` - test additions/coverage/regression work
- `docs` - documentation updates
- `refactor` - structural code changes without behavior intent
- `chore` - maintenance/non-feature work
- `audit` / `race-safety` / `ux` - specialized maintenance categories used by the team

Optional tracker tags are often appended in parentheses, for example:

- `test: improve crossbar array and comparison coverage (COV-15, COV-20)`
- `ux: add keyboard shortcuts for docs search and EDA builder (UXP-09, UXP-12)`

### Commit quality expectations

- Use imperative tense ("add", "fix", "reduce", "align").
- Keep subject line concise and specific.
- Group related file changes in one commit; avoid mixed unrelated edits.
- For reversions/cleanup, state it explicitly in subject.

---

## Quick Contributor Workflow

```bash
# 1) Create branch
git checkout -b <type>/<short-topic>

# 2) Implement + format + verify
gofmt -w .
go vet ./...
go test ./...
go test -race ./...

# 3) Commit with convention
git commit -m "<type>(<scope>): <summary>"

# 4) Push and open PR
git push -u origin <branch>
```

Thanks again for helping improve FeCIM Lattice Tools.
