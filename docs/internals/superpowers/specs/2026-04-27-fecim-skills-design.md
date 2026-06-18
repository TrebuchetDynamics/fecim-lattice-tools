# FeCIM Skills — Design

**Date:** 2026-04-27
**Status:** Spec (awaiting implementation plan)
**Inspiration:** [mattpocock/skills](https://github.com/mattpocock/skills)

## Goal


## Non-Goals

- Rewriting `superpowers:grill-me` — `fecim-grill` is FeCIM-specific and complementary.
- Auto-syncing skills from the mattpocock upstream repo.
- Packaging as an Anthropic Claude Code plugin.

## Skill Catalog

Each skill lives in `tools/fecim-skills/<name>/SKILL.md` with mattpocock-format YAML frontmatter (`name`, `description`).

| # | Skill | description: (frontmatter, abbreviated) | Core workflow |
|---|---|---|---|
| 1 | **fecim-researcher** | Surveys FeCIM domain knowledge by searching `references/`, `citations/`, `docs/4-research/`, and the local Cognee KG, then synthesizes a cited research note. Use when investigating a physics topic, evaluating a paper, or grounding a design decision in literature. | Identify question and scope → search `references/`, `docs/4-research/`, `experimental-data/` → if Cognee env vars set, query KG via `scripts/cognee-search.py` (else skip) → cite findings using canonical short form (Materlik 2015, Park 2015, etc.) → output structured note: question, sources, finding, gaps, recommended next step. |
| 2 | **fecim-citation** | Verifies and formats FeCIM physics/measurement claims against the project's published-source list and `docs/4-research/honesty-audit.md`. Use when adding a numeric claim, accuracy figure, or device-parameter assertion to code, docs, PR descriptions, or commit messages. | Parse claim → match to published-physics list (Materlik 2015, Park 2015, Alessandri 2018, Guo 2018, 2025 HZO FTJ paper) → check honesty-audit "removed/unverified" list (87% MNIST, 30-level-as-spec, NAND/GPU energy multipliers) → output citation in canonical format OR "needs rephrasing as educational" with suggested wording. |
| 4 | **fecim-labtester** | Runs the FeCIM test matrix (full, race, module-scoped, coverage, golden regen) and interprets physics regression failures using the 5 known bug patterns. Use when running tests, debugging test failures, or regenerating physics golden files. | Pick scope (`./...`, module, package) → run with appropriate flags (`-race`, `-cover`, `FECIM_UPDATE_PHYSICS_GOLDEN=1`) → on failure, classify against 5 known patterns from `MEMORY.md` (guard-band sign flip, bounds collapse, ACCEPT ±1 interaction, zero-field bounds reset, Preisach Everett zero-clamp) → output RED/GREEN evidence block per CLAUDE.md TDD rule. |
| 7 | **fecim-honesty-audit** | Enforces `docs/4-research/honesty-audit.md` policy by scanning PR diffs, READMEs, and presentation material for removed/unverified claims (87% MNIST, 30-states-as-fact, energy multipliers vs NAND/GPUs). Use before committing docs, PRs, or release notes that include accuracy or efficiency numbers. | Take diff or doc as input → regex-scan for trigger phrases + numeric patterns → for each hit, classify: verified / educational-default / removed-unverified → for removed-unverified, suggest the approved rephrasing pattern from the audit doc → output: pass / changes-requested with line-level annotations. |

### Cross-cutting conventions for all 8 skills

- Each `SKILL.md` ≤ ~150 lines.
- Long reference content (honesty-audit rule list, 5 bug patterns) lives in `tools/fecim-skills/_shared/fecim-context.md` and is referenced — not duplicated — to avoid drift.
- Every skill that produces a code-mutating action ends its workflow with the TDD evidence block from `_shared/tdd-evidence-template.md`.
- No skill is allowed to skip the TDD hard-rule from CLAUDE.md.

## Directory Layout

```
tools/fecim-skills/
├── README.md                          # what these are, how to install
├── _shared/
│   ├── fecim-context.md               # FeCIM domain primer (citations, modules, honesty rules)
│   └── tdd-evidence-template.md       # RED/GREEN/verification block per CLAUDE.md
├── fecim-researcher/SKILL.md
├── fecim-citation/SKILL.md
├── fecim-builder/SKILL.md
├── fecim-labtester/SKILL.md
├── fecim-grill/SKILL.md
├── fecim-honesty-audit/SKILL.md
└── fecim-fyne-maintain/SKILL.md

scripts/install-fecim-skills.sh        # generates per-harness adapters (idempotent)
scripts/test-install-fecim-skills.sh   # tests for the install script
```

Generated adapter targets (**committed** to the repo, not gitignored, so the repo works for users without running the install script; regenerated on edit):

```
.claude/skills/<name>/SKILL.md         # symlink to canonical, copy fallback
.codex/AGENTS.md                       # managed block referencing canonical files
.opencode/command/<name>.md            # frontmatter wrapper that includes canonical
```

## Adapter Format Per Harness

### Canonical (single source)

```yaml
---
name: fecim-researcher
description: Surveys FeCIM domain knowledge by searching references/, citations/, ...
---

# fecim-researcher
[body]
```

### 1. Claude Code — `.claude/skills/<name>/SKILL.md`

Relative symlink to canonical. On Windows/CI fallback: copy with managed-block header:
```
<!-- generated-from: tools/fecim-skills/<name>/SKILL.md -->
<!-- do not edit; run scripts/install-fecim-skills.sh -->
```

### 2. Codex — managed block in `.codex/AGENTS.md`

```markdown
<!-- fecim-skills:start -->
## FeCIM Skills

When the user's request matches the trigger description below, follow the workflow in the linked file.

- **fecim-researcher** — Surveys FeCIM domain knowledge ... → see `tools/fecim-skills/fecim-researcher/SKILL.md`
- **fecim-citation** — Verifies and formats FeCIM physics claims ... → see `tools/fecim-skills/fecim-citation/SKILL.md`
- ... (one bullet per skill)

For all skills above, the canonical body is the linked SKILL.md file. Read it before acting.
<!-- fecim-skills:end -->
```

Block is regenerated each install — content outside the markers is preserved.

### 3. opencode — `.opencode/command/<name>.md`

```markdown
---
description: <copied from canonical description: field>
agent: build
---

Read and follow the workflow in `tools/fecim-skills/<name>/SKILL.md`.

Use the user's request as `$ARGUMENTS`.
```

`agent: build` selects opencode's default "build" agent (its general code-writing agent). Each skill is invocable as `/fecim-researcher <args>` in opencode.

## Install Flow

`scripts/install-fecim-skills.sh`:

1. For each `tools/fecim-skills/<name>/SKILL.md`:
   - Write Claude Code adapter (symlink, copy fallback).
   - Update Codex managed block.
   - Write opencode command file.
2. Each generated artifact carries a header comment: `# Generated by scripts/install-fecim-skills.sh — edit tools/fecim-skills/<name>/SKILL.md instead`.
3. `make install-skills` invokes the script.
4. CI runs `scripts/install-fecim-skills.sh --check`; non-zero exit if any generated adapter is stale.

## Failure Modes Covered

| Failure | Behavior |
|---|---|
| User edits a generated adapter | Reverted on next install; CI flags drift if pushed |
| Symlink fails (Windows, some CIs) | Script falls back to copy with managed-block header |
| `.codex/AGENTS.md` already exists with user content | Script only touches the `<!-- fecim-skills:start -->`/`...:end -->` block |
| `.opencode/command/` doesn't exist | Script creates it (`mkdir -p`) |
| Canonical SKILL.md frontmatter missing `description:` | Script fails loudly with the offending file path |

## Testing

`scripts/test-install-fecim-skills.sh` (bash, runs in temp dir):

| Test | Setup | Assertion |
|---|---|---|
| Fresh install | Empty `.claude/`, `.codex/`, `.opencode/` | All 8 adapters appear; canonical files unchanged |
| Idempotent re-install | Run install twice | Diff between runs 1 and 2 is empty |
| Codex preserves user content | Pre-existing `.codex/AGENTS.md` with custom prose outside the markers | Custom prose still present; managed block updated |
| `--check` detects drift | Modify a generated file, run `--check` | Non-zero exit, error names the drifted file |
| Missing `description:` | Stub a SKILL.md without `description:` | Install fails loudly, names the file |
| Symlink fallback | `ln -s` disabled in temp env | Adapter is a copy with managed-block header |

Runs via `make test-skills` and in CI.

Each `SKILL.md` also includes a "Verification" section listing 1–3 example invocations and expected behavior. These are documentation, not automated tests, but they keep skills auditable.

```markdown
## Verification
- Input: a PR adding `go func() { mylabel.SetText("done") }()` in `module1-hysteresis/pkg/gui/`
```

## TDD Compliance per CLAUDE.md

The install script (which IS behavior) is under full TDD:
- **RED**: write `scripts/test-install-fecim-skills.sh` first; one scenario fails.
- **GREEN**: minimal `install-fecim-skills.sh` change to pass.
- **REFACTOR**: with all scenarios green.

The 8 SKILL.md files are documentation/configuration, not behavior changes — they qualify under CLAUDE.md's "Documentation-only … may use `TDD: N/A`" carve-out.

## Rollout

1. Land the install script + tests with one skill (`fecim-builder`) wired through to validate the pipeline end-to-end on all three harnesses.
2. Land the remaining 7 skills in a second PR (smaller change, content-only).
3. Update `AGENTS.md` and `docs/3-develop/README.md` with one line pointing to `tools/fecim-skills/README.md`.
4. No deprecation — nothing existing replaces.

## Out of Scope (Explicit)

- Rewriting `superpowers:grill-me`.
- Auto-updating skills from the mattpocock upstream repo.
- Anthropic Claude Code plugin packaging.
