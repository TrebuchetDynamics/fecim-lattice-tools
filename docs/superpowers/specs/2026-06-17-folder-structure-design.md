# Docs Folder Structure Revamp

**Date:** 2026-06-17
**Status:** Approved

## Problem

The `docs/` tree has two structural issues:

1. The four main content sections use numeric prefixes (`1-getting-started`, `2-learn`, `3-develop`, `4-research`) that add noise without adding meaning.
2. Several directories (`adr/`, `assets/`, `documentation/`, `notebook/`, `paper/`, `presentations/`, `public-release/`, `superpowers/`) and loose top-level files live outside the numbered scheme with no clear home.

## Goals

- Remove all numeric prefixes from directory names across the tree.
- Give every file and directory a clear semantic home — nothing loose at the top level except `README.md`.
- Keep internal structure of deeply-nested content (e.g., `research/papers/by-topic/`) unchanged to avoid broad link churn.

## New Top-Level Shape

```
docs/
├── README.md
├── guides/
├── modules/
├── internals/
└── research/
```

## Section Mappings

### `guides/` (← `1-getting-started/` + orphans)

User-facing onboarding and reference docs.

```
guides/
├── README.md
├── installation.md
├── cli-reference.md
├── runbook.md
├── research-workflow.md
├── reproducibility-guide.md
├── glossary.md                    (← top-level GLOSSARY.md)
├── trust.md                       (← top-level TRUST.md)
├── assets/                        (← assets/screenshots/ + assets/reference-screenshots/ merged flat)
└── presentations/                 (← presentations/)
    └── presenter-script.md
```

`assets/screenshots/` and `assets/reference-screenshots/` merge into a single flat `guides/assets/` — the distinction between reference and current screenshots is not worth a separate directory.

### `modules/` (← `2-learn/`)

Per-module educational docs. The `moduleN-` prefix is dropped from each subdirectory name.

```
modules/
├── README.md
├── eli5-overview.md
├── hysteresis/        (← module1-hysteresis/)
├── crossbar/          (← module2-crossbar/)
├── mnist/             (← module3-mnist/)
├── circuits/          (← module4-circuits/)
├── comparison/        (← module5-comparison/)
├── eda/               (← module6-eda/)
└── doc-viewer/        (← module7-docs/)
```

Internal file structure of each module subdir is unchanged.

### `internals/` (← `3-develop/` + orphans)

Developer-facing: architecture, testing, GUI specs, ADRs, release tooling.

```
internals/
├── README.md
├── api-reference.md
├── accessibility.md
├── code-quality.md
├── config-reference.md
├── config-migration.md
├── known-limitations.md
├── memory-optimization.md
├── repo-health.md
├── security.md                    (← top-level SECURITY.md)
├── how-to-break-this.md           (← top-level HOW_TO_BREAK_THIS.md)
├── units-reference.md
├── architecture/                  (unchanged)
├── adr/                           (← docs/adr/)
├── audits/                        (NEW — collects stray session-artifact files)
│   ├── HYPER_ANALYSIS_REPORT.md
│   ├── TODO_MARKER_AUDIT_2026-02-18.md
│   ├── AUTOIMPROVEMENT-KG-PROMPT.md
│   └── DOC_DRIFT_AUDIT_2026-02-11.md  (moved out of gui/)
├── automation/                    (unchanged)
├── gui/                           (unchanged, minus DOC_DRIFT_AUDIT moved above)
├── journal/                       (← notebook/)
├── release/                       (← public-release/)
├── superpowers/                   (unchanged)
└── testing/                       (unchanged)
```

### `research/` (← `4-research/` + orphans)

Literature, physics validation, paper, and tools.

```
research/
├── README.md
├── error-propagation.md
├── honesty-audit.md
├── physics-models.md
├── physics-validation.md
├── predictions.md                 (← top-level PREDICTIONS.md)
├── literature-review/             (unchanged)
├── paper/                         (← docs/paper/)
├── tools/                         (← opensource-tools/ + top-level opensource-tools-summary.md)
├── transcripts/                   (unchanged)
└── validation/                    (unchanged, plus two files moved in from research root)
    ├── PHYSICS_REALISM_AUDIT.md
    ├── PHYSICS_REALISM_AUDIT_ADDENDUM_2026-02.md
    └── ... (audits/ baselines/ claims/ contracts/ external/ module1/ module4/ policies/ reviewer/ tools/)
```

The stray `docs/documentation/module4-circuits/M4-OBS-05-RESULTS.md` (orphan folder, single file) moves to `research/validation/module4/observations/` where the rest of the module4 observation files live.

`research/papers/by-topic/` numbered subdirs (`01-ferroelectric-materials/`, etc.) are left unchanged.

## What Does Not Change

- All file names (only directories are renamed/moved).
- Internal structure within `research/papers/by-topic/` and `research/validation/`.
- Internal structure of each module subdir under `modules/`.
- The `superpowers/` content under `internals/superpowers/`.

## CLAUDE.md Impact

`CLAUDE.md` references several paths. After the move the affected paths are:

| Old path | New path |
|---|---|
| `docs/3-develop/api-reference.md` | `docs/internals/api-reference.md` |
| `docs/3-develop/testing/TESTING.md` | `docs/internals/testing/TESTING.md` |
| `docs/2-learn/module6-eda/README.md` | `docs/modules/eda/README.md` |
| `docs/4-research/honesty-audit.md` | `docs/research/honesty-audit.md` |

These four lines in `CLAUDE.md` must be updated as part of implementation.
