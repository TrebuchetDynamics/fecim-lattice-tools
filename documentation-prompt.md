# FeCIM Curriculum Documentation Prompt

## Role And Operating Mode

- Adopt a combined senior-researcher scientific tone inspired by Dr. external research group and Dr. Jaeho Shin.
  Do not impersonate or claim to be them.
- You are an expert software engineer, documentation systems engineer, and PhD-level curriculum designer.
- Operate autonomously. Only ask questions when blocked by missing inputs or files.
- If an ambiguity remains, choose the most reasonable default and record the choice in the report.
- Headless-first: use CLI + file inspection only. Do not run GUI unless explicitly requested.

## Objective

- Build a concise, PhD-ready curriculum in `docs/documentation/` covering physics, math, and software
  of FeCIM Lattice Tools.
- Maintain scientific honesty: separate verified results from aspirational claims and cite sources
  where present in this repo.
- Ensure Module 7 (documentation viewer) presents the curriculum cleanly and predictably, prioritizing
  layout clarity and navigation over raw file listing.
- Iterate with minimal, verifiable changes.

## Non-Negotiable Curriculum Structure

```
docs/documentation/
  README.md
  MODULES.md
  research-papers/README.md
  module1-hysteresis/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
  module2-crossbar/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
  module3-mnist/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
  module4-circuits/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
  module5-comparison/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
  module6-eda/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
  module7-docs/
    ELI5.md
    PHYSICS.md
    FEATURES.md
    OPENSOURCE-TOOLS.md
```

## Content Standards

- Keep content concise with short sections, flat bullet lists, and compact tables.
- Use consistent headings so Module 7 ToC and search are reliable.
- Always state what the simulator simplifies vs reality.
- Distinguish demonstrated vs modeled vs aspirational.

### Required Section Templates

ELI5.md
- Learning Objectives
- Intuition
- Key Analogies
- What The Simulator Simplifies
- Next Steps

PHYSICS.md
- Prerequisites
- Core Model
- Key Equations (Simplified)
- Parameters And Units
- Assumptions And Limits
- Where It Lives In Code
- Sources

FEATURES.md
- What This Module Does
- Primary Components
- Key Workflows
- Extension Points
- Known Limitations

OPENSOURCE-TOOLS.md
- When To Use External Tools
- Recommended Tools (with short rationale)
- Integration Notes (links to in-repo docs where applicable)

## Module 7 Requirements (Curriculum-First UI)

- Default root is `docs/documentation/`, not the full `docs/` tree.
- Sidebar order at root: module folders (module1..module7), then `research-papers`, then README/MODULES.
- Provide quick-access shortcuts for the current module: ELI5, PHYSICS, FEATURES, OPENSOURCE-TOOLS.
- Category mapping by filename:
  - `ELI5.md` -> ELI5
  - `PHYSICS.md` -> Physics
  - `FEATURES.md` -> Guide
  - `OPENSOURCE-TOOLS.md` -> Guide
- Click behavior is deterministic: favorites toggles must not trigger document selection.

## Documentation Alignment

- Update `docs/development/GUI/GUI.module7.md` to reflect curriculum-first navigation and UI behavior.
- Update `docs/development/ARCHITECTURE.md` only if Module 7 details are missing or incorrect.

## Validation

- Required: `go test ./module7-docs/...`
- Optional: headless structure checks (rg/find/python) for `docs/documentation` completeness.
- If a command fails, fix and re-run until it succeeds or report a clear blocker.

## Execution Rules

- No human intermediaries: run commands, inspect logs, make edits, validate independently.
- Always check `logs/` for the most recent run and quote key evidence in the report.
- Keep changes minimal and targeted; avoid refactors unless required for correctness.
- If you add temporary scripts, remove them before final output.
- If blocked, report the exact error output and the last command run.

## Deliverable

Provide a concise report including:
- What was created/updated in `docs/documentation/`
- How Module 7 was updated to present the curriculum
- Validation commands run and key evidence (including logs)
- Any gaps, issues, or follow-ups
