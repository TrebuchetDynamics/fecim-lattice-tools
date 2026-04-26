# Paper Implementation Plan

Goal: produce an arXiv-ready tools paper for FeCIM Lattice Tools using LaTeX in this repository.

Current date: April 26, 2026.

## Scope

This is a tools paper, not a new physics or device-measurement paper. The contribution is the software framework, validation methodology, reproducibility surface, educational workflow, and honest limitation boundary.

## Claim Gate

Every public paper claim must have at least one of:

- a citation in `refs.bib`
- a validation artifact in `validation/`, `output/validation/`, or `paper/data/`
- an explicit limitation phrase such as "simulation-only", "educational", "planned", or "not a device measurement"

## Week 1: Skeleton

- [x] Create `paper/` directory.
- [x] Add LaTeX source.
- [x] Add seed bibliography.
- [x] Add build script.
- [x] Add module, validation, and comparison tables.
- [ ] Fill author metadata only after public contact details are approved.
- [ ] Add first architecture figure generated from source.

## Week 2: Methodology

- [ ] Expand the system architecture section.
- [ ] Add equations for Preisach/Landau-Khalatnikov modeling.
- [ ] Add equations for crossbar MVM and parasitic voltage drops.
- [ ] Add peripheral circuit equations and assumptions.
- [ ] Add EDA export flow details.

## Week 3: Validation

- [ ] Promote Module 2 KCL report into paper data.
- [ ] Generate KCL residual plot from artifact JSON.
- [ ] Generate Park 2015 hysteresis overlay from documented provenance.
- [ ] Add SPICE/ngspice comparison only when the artifact is reproducible.
- [ ] Add MNIST result only when the full training/inference artifact and confusion matrix are reproducible.

## Week 4: Related Work and Educational Use

- [ ] Add CrossSim, NeuroSim, AIHWKit, FerroX, and comparable-tool discussion.
- [ ] Add fair comparison table.
- [ ] Add graduate-course and teaching workflow examples.
- [ ] Add reproducibility workflow subsection.

## Week 5: Review

- [ ] Run self-review and save notes in `reviews/`.
- [ ] Run AI review and save notes in `reviews/`.
- [ ] Check every numeric claim against artifact or citation.
- [ ] Check every acronym definition.
- [ ] Check every figure caption.
- [ ] Check BibTeX completeness.

## Week 6: Submission Candidate

- [ ] Build clean PDF.
- [ ] Confirm arXiv category.
- [ ] Freeze artifact-producing commit hash.
- [ ] Save v1 draft in `drafts/`.
- [ ] Submit only if all validation claims are reproducible.

## Build Command

```bash
bash paper/build.sh
```

## Verification Commands

```bash
bash paper/build.sh
git diff --check
python3 scripts/public-release/check_public_release_boundary.py
```

