# Docs Folder Structure Revamp Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Remove all numeric directory prefixes from `docs/`, absorb all orphan folders and loose top-level files into four clean semantic sections.

**Architecture:** Pure filesystem restructure using `git mv` to preserve history. No production Go code is modified. CLAUDE.md updated at the end to fix the four broken path references. TDD: N/A — documentation-only change with no production-code behavior change.

**Tech Stack:** git, bash

---

### Task 1: Rename the four main sections

**Files:**
- Rename: `docs/1-getting-started/` → `docs/guides/`
- Rename: `docs/2-learn/` → `docs/modules/`
- Rename: `docs/3-develop/` → `docs/internals/`
- Rename: `docs/4-research/` → `docs/research/`

- [ ] **Step 1: Rename all four dirs**

```bash
git mv docs/1-getting-started docs/guides
git mv docs/2-learn docs/modules
git mv docs/3-develop docs/internals
git mv docs/4-research docs/research
```

- [ ] **Step 2: Verify**

```bash
ls docs/
```
Expected: `guides/  internals/  modules/  research/` visible alongside the remaining orphan dirs and loose files.

- [ ] **Step 3: Commit**

```bash
git commit -m "docs: rename numbered sections to semantic names"
```

---

### Task 2: Rename module subdirs (drop moduleN- prefix)

**Files:**
- Rename: `docs/modules/module1-hysteresis/` → `docs/modules/hysteresis/`
- Rename: `docs/modules/module2-crossbar/` → `docs/modules/crossbar/`
- Rename: `docs/modules/module3-mnist/` → `docs/modules/mnist/`
- Rename: `docs/modules/module4-circuits/` → `docs/modules/circuits/`
- Rename: `docs/modules/module5-comparison/` → `docs/modules/comparison/`
- Rename: `docs/modules/module6-eda/` → `docs/modules/eda/`
- Rename: `docs/modules/module7-docs/` → `docs/modules/doc-viewer/`

- [ ] **Step 1: Rename all module subdirs**

```bash
git mv docs/modules/module1-hysteresis docs/modules/hysteresis
git mv docs/modules/module2-crossbar docs/modules/crossbar
git mv docs/modules/module3-mnist docs/modules/mnist
git mv docs/modules/module4-circuits docs/modules/circuits
git mv docs/modules/module5-comparison docs/modules/comparison
git mv docs/modules/module6-eda docs/modules/eda
git mv docs/modules/module7-docs docs/modules/doc-viewer
```

- [ ] **Step 2: Verify**

```bash
ls docs/modules/
```
Expected: `README.md  circuits/  comparison/  crossbar/  doc-viewer/  eda/  eli5-overview.md  hysteresis/  mnist/`

- [ ] **Step 3: Commit**

```bash
git commit -m "docs: drop moduleN- prefix from module subdirs"
```

---

### Task 3: Absorb orphans into guides/

Merge both `assets/` subdirs into a single flat `guides/assets/`, move presentations, and relocate two loose top-level files.

**Files:**
- Move: `docs/assets/hysteresis_readme.png` → `docs/guides/assets/`
- Move: `docs/assets/reference-screenshots/*.png` (5 files) → `docs/guides/assets/`
- Move: `docs/assets/screenshots/*.png` (7 files) → `docs/guides/assets/`
- Move: `docs/presentations/` → `docs/guides/presentations/`
- Move: `docs/GLOSSARY.md` → `docs/guides/glossary.md`
- Move: `docs/TRUST.md` → `docs/guides/trust.md`
- Delete: empty `docs/assets/` dir

- [ ] **Step 1: Merge assets into guides/assets/**

```bash
mkdir -p docs/guides/assets
git mv docs/assets/hysteresis_readme.png docs/guides/assets/
git mv docs/assets/reference-screenshots/crossbar-heatmap-16x16.png docs/guides/assets/
git mv docs/assets/reference-screenshots/crossbar-heatmap-8x8.png docs/guides/assets/
git mv docs/assets/reference-screenshots/hysteresis-material-comparison.png docs/guides/assets/
git mv docs/assets/reference-screenshots/hysteresis-p-e-loop.png docs/guides/assets/
git mv docs/assets/reference-screenshots/mvm-diagram.png docs/guides/assets/
git mv docs/assets/screenshots/circuits-ispp-convergence.png docs/guides/assets/
git mv docs/assets/screenshots/circuits-pvt-corners.png docs/guides/assets/
git mv docs/assets/screenshots/comparison-architecture-bars.png docs/guides/assets/
git mv docs/assets/screenshots/docs-overview.png docs/guides/assets/
git mv docs/assets/screenshots/eda-design-overview.png docs/guides/assets/
git mv docs/assets/screenshots/hysteresis_fyne.png docs/guides/assets/
git mv docs/assets/screenshots/mnist-accuracy-sweep.png docs/guides/assets/
```

- [ ] **Step 2: Move presentations and loose top-level files**

```bash
git mv docs/presentations docs/guides/presentations
git mv docs/GLOSSARY.md docs/guides/glossary.md
git mv docs/TRUST.md docs/guides/trust.md
```

- [ ] **Step 3: Remove now-empty assets subdirs**

```bash
rmdir docs/assets/reference-screenshots docs/assets/screenshots docs/assets
```

- [ ] **Step 4: Verify**

```bash
ls docs/guides/assets/
ls docs/guides/presentations/
```
Expected: 13 image files in `docs/guides/assets/`. `presenter-script.md` in `docs/guides/presentations/`.

- [ ] **Step 5: Commit**

```bash
git commit -m "docs: absorb assets, presentations, glossary, trust into guides/"
```

---

### Task 4: Absorb orphans into internals/ and create audits/

**Files:**
- Move: `docs/adr/` → `docs/internals/adr/`
- Move: `docs/notebook/` → `docs/internals/journal/`
- Move: `docs/public-release/` → `docs/internals/release/`
- Move: `docs/superpowers/` → `docs/internals/superpowers/`
- Move: `docs/SECURITY.md` → `docs/internals/security.md`
- Move: `docs/HOW_TO_BREAK_THIS.md` → `docs/internals/how-to-break-this.md`
- Create: `docs/internals/audits/`
- Move: `docs/internals/HYPER_ANALYSIS_REPORT.md` → `docs/internals/audits/`
- Move: `docs/internals/TODO_MARKER_AUDIT_2026-02-18.md` → `docs/internals/audits/`
- Move: `docs/internals/AUTOIMPROVEMENT-KG-PROMPT.md` → `docs/internals/audits/`
- Move: `docs/internals/gui/DOC_DRIFT_AUDIT_2026-02-11.md` → `docs/internals/audits/`

- [ ] **Step 1: Move orphan dirs into internals/**

```bash
git mv docs/adr docs/internals/adr
git mv docs/notebook docs/internals/journal
git mv docs/public-release docs/internals/release
git mv docs/superpowers docs/internals/superpowers
```

- [ ] **Step 2: Move loose top-level files into internals/**

```bash
git mv docs/SECURITY.md docs/internals/security.md
git mv docs/HOW_TO_BREAK_THIS.md docs/internals/how-to-break-this.md
```

- [ ] **Step 3: Create audits/ and move stray session-artifact files**

```bash
mkdir -p docs/internals/audits
git mv docs/internals/HYPER_ANALYSIS_REPORT.md docs/internals/audits/
git mv docs/internals/TODO_MARKER_AUDIT_2026-02-18.md docs/internals/audits/
git mv docs/internals/AUTOIMPROVEMENT-KG-PROMPT.md docs/internals/audits/
git mv docs/internals/gui/DOC_DRIFT_AUDIT_2026-02-11.md docs/internals/audits/
```

- [ ] **Step 4: Verify**

```bash
ls docs/internals/
ls docs/internals/audits/
```
Expected: `adr/  api-reference.md  architecture/  audits/  automation/  gui/  journal/  release/  superpowers/  testing/  ...` in `docs/internals/`. Four files in `docs/internals/audits/`.

- [ ] **Step 5: Commit**

```bash
git commit -m "docs: absorb adr, notebook, public-release, superpowers, security into internals/"
```

---

### Task 5: Absorb orphans into research/

**Files:**
- Move: `docs/paper/` → `docs/research/paper/`
- Rename: `docs/research/opensource-tools/` → `docs/research/tools/`
- Move: `docs/opensource-tools-summary.md` → `docs/research/tools/summary.md`
- Move: `docs/PREDICTIONS.md` → `docs/research/predictions.md`
- Move: `docs/research/PHYSICS_REALISM_AUDIT.md` → `docs/research/validation/`
- Move: `docs/research/PHYSICS_REALISM_AUDIT_ADDENDUM_2026-02.md` → `docs/research/validation/`
- Move: `docs/documentation/module4-circuits/M4-OBS-05-RESULTS.md` → `docs/research/validation/module4/observations/`
- Delete: empty `docs/documentation/` dir

- [ ] **Step 1: Move paper/ and rename opensource-tools/**

```bash
git mv docs/paper docs/research/paper
git mv docs/research/opensource-tools docs/research/tools
```

- [ ] **Step 2: Move loose top-level files into research/**

```bash
git mv docs/opensource-tools-summary.md docs/research/tools/summary.md
git mv docs/PREDICTIONS.md docs/research/predictions.md
```

- [ ] **Step 3: Move physics audit files into validation/**

```bash
git mv docs/research/PHYSICS_REALISM_AUDIT.md docs/research/validation/
git mv docs/research/PHYSICS_REALISM_AUDIT_ADDENDUM_2026-02.md docs/research/validation/
```

- [ ] **Step 4: Move stray documentation/ file and remove empty dir**

```bash
git mv docs/documentation/module4-circuits/M4-OBS-05-RESULTS.md docs/research/validation/module4/observations/
rmdir docs/documentation/module4-circuits docs/documentation
```

- [ ] **Step 5: Verify**

```bash
ls docs/research/
ls docs/research/tools/
ls docs/research/validation/ | head -5
```
Expected: `paper/` and `tools/` present in `docs/research/`. `tools/` contains former opensource-tools content plus `summary.md`. `PHYSICS_REALISM_AUDIT*.md` files present in `docs/research/validation/`.

- [ ] **Step 6: Commit**

```bash
git commit -m "docs: absorb paper, tools, predictions, physics audits into research/"
```

---

### Task 6: Update CLAUDE.md path references

**Files:**
- Modify: `CLAUDE.md` (root of repo, not inside docs/)

CLAUDE.md has 8 path references across 4 unique paths, all of which now point to moved locations.

- [ ] **Step 1: Apply all four substitutions**

```bash
sed -i 's|docs/3-develop/api-reference.md|docs/internals/api-reference.md|g' CLAUDE.md
sed -i 's|docs/3-develop/testing/TESTING.md|docs/internals/testing/TESTING.md|g' CLAUDE.md
sed -i 's|docs/2-learn/module6-eda/README.md|docs/modules/eda/README.md|g' CLAUDE.md
sed -i 's|docs/4-research/honesty-audit.md|docs/research/honesty-audit.md|g' CLAUDE.md
```

- [ ] **Step 2: Verify all 8 references are updated**

```bash
grep "docs/" CLAUDE.md
```
Expected output (8 lines, no `docs/1-`, `docs/2-`, `docs/3-`, `docs/4-` remaining):
```
**Full reference:** See `docs/internals/api-reference.md` for detailed lookups.
| Find a function | `docs/internals/api-reference.md` |
| Fix an error | `docs/internals/testing/TESTING.md` |
| Add a feature | `docs/internals/api-reference.md` |
| Run/understand tests | `docs/internals/testing/TESTING.md` |
| EDA documentation | `docs/modules/eda/README.md` |
... accuracy over marketing claims. Full audit: `docs/research/honesty-audit.md`.
Full test documentation: `docs/internals/testing/TESTING.md`
```

- [ ] **Step 3: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md path references after restructure"
```

---

### Task 7: Verify and final cleanup

- [ ] **Step 1: Confirm no old numbered prefixes remain**

```bash
find docs/ -type d | grep -E '/[1-4]-'
```
Expected: no output.

- [ ] **Step 2: Confirm no stray top-level content remains**

```bash
ls docs/
```
Expected: exactly `README.md  guides/  internals/  modules/  research/`

- [ ] **Step 3: Confirm no moduleN- dirs remain**

```bash
find docs/ -type d | grep -E 'module[0-9]-'
```
Expected: no output.

- [ ] **Step 4: Run tests to confirm nothing broke**

```bash
go test ./... -short -timeout 60s 2>&1 | tail -5
```
Expected: all packages PASS. (The restructure touches no Go source — this is a sanity check that no test fixture paths reference the old doc locations.)

- [ ] **Step 5: Final commit if any cleanup needed**

```bash
git status
```
If clean, done. If any unexpected loose files remain, move them to the appropriate section and commit with `docs: final cleanup after folder restructure`.
