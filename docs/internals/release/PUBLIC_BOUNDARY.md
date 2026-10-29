# Public Boundary

## Allowed Material
- Source code, tests, validation harnesses, and safe example data or fixtures.
- Build, run, contributor, and maintenance documentation needed to build or review the source tree.
- Research summaries, methodology notes, and educational writeups written in project language.
- DOI lists, arXiv links, metadata records, and paper download scripts.
- Release-policy documents, audit records, and enforcement scripts used to keep the public tree inside the approved boundary.
- Third-party PDFs only when [the canonical PDF audit](THIRD_PARTY_PDF_AUDIT.csv) marks the tracked file as `keep` or `keep-with-conditions`.

## Source-Only Rule
The public repository is source-only.

Hand-authored code, documentation, scripts, audit records, and safe fixture data may be tracked when they are intentionally part of the source release.

Generated outputs may not be tracked. Generated outputs include:
- compiled binaries and packaged builds
- logs, including nested module GUI log directories, screenshots, recordings, and other run artifacts
- validation output, benchmark output, and exported reports
- generated EDA and export files unless they are intentionally curated sample inputs or fixtures

Deterministic validation fixtures that support tests may be tracked only under `testdata/` paths. They must not remain in runtime output directories such as `output/`, `validation/output/`, or module-local `output/` trees.

## Disallowed Material
- `docs/archive/**`
- `docs/research/internal-analysis/**`
- `docs/research/transcripts/COSM_2025_AI_Hardware_Breakthrough/**`
- `docs/research/transcripts/ironlattice-youtube-script.md`
- `docs/research/tour-group-ironlattice-research.md`
- `docs/research/superlattice-material-analysis.md`
- `docs/superpowers/**` and its relocated owner `docs/internals/superpowers/**`
- `docs/research-papers/tourlab/**`
- `docs/research/papers/tourlab/**`
- `demo2-crossbar/pkg/_layers_experimental/kvcache_fjh.go`
- `demo2-crossbar/pkg/layers/kvcache_fjh.go`
- `demo2-inference/pkg/layers/kvcache_fjh.go`
- `module2-crossbar/pkg/_layers_experimental/kvcache_fjh.go`
- `docs/research/papers/DOWNLOAD_PLAN.md`
- `docs/research/tools/research_notes_final.md`
- `validation/literature/_incoming/**`
- Repo-local test logs such as `module1-hysteresis/test_output*.txt`
- Generated build or run artifacts such as binaries, logs, screenshots, recordings, and export output.
- Runtime validation output directories outside `testdata/`, including root `output/`, `validation/output/`, and module-local `output/` trees.
- Personal roleplay material, outreach planning, or NDA/confidential material.
- Third-party PDFs without explicit redistribution evidence recorded in the audit.

## Enforcement and PDF audit maintenance

Run these commands from the repository root with Python 3, Git and ripgrep installed:

```bash
PYTHONDONTWRITEBYTECODE=1 python3 scripts/public-release/check_public_release_boundary.py
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s scripts/public-release -p 'test_public_release_boundary.py' -v
```

The [checker](../../../scripts/public-release/check_public_release_boundary.py) returns zero only when its tracked-path, generated-output, PDF-decision and phrase checks pass. Missing PDF records, conflicting normalized records and scanner errors fail closed. Phrase scanning covers `CLAUDE.md`, `README.md`, `docs/guides`, `docs/modules`, `docs/internals`, `docs/research` and the existing `module1`–`module7` source directories. Only this file and `docs/internals/release/REMOVAL_LOG.md` are excluded from phrase scanning because they define prohibited content. Neither is excluded from tracked-path or PDF checks. Other policy-like documents are not exempt.

The [bootstrap tool](../../../scripts/public-release/bootstrap_pdf_audit.py) updates the canonical audit in place:

```bash
PYTHONDONTWRITEBYTECODE=1 python3 scripts/public-release/bootstrap_pdf_audit.py
```

It retains existing records, including historical rows and blank cells. It normalizes only the exact `docs/4-research/` prefix to `docs/research/` and rejects conflicting normalized keys before writing. New tracked PDFs default to `replace-with-link`, with no redistribution evidence. Review the resulting diff; bootstrap does not grant rights or approve publication.

## Actual repository diagnostic — publication remains blocked

Snapshot: task `T-012-41e51a72`, based on `dba35de3ce3549f69d9e6995a7a88e57030ba09b`, with the mechanical checker/bootstrap and audit repairs described above. The read-only checker command above returned **exit 1**, not clearance. It reported no missing or blocked PDF decisions, no generated-output findings and no scanner errors. It reported these **16 blocked tracked paths** and **three phrase matches**:

```text
Blocked tracked path: docs/internals/superpowers/plans/2026-04-27-fecim-skills.md
Blocked tracked path: docs/internals/superpowers/plans/2026-05-17-research-ingestion-mvp.md
Blocked tracked path: docs/internals/superpowers/plans/2026-06-17-folder-structure-revamp.md
Blocked tracked path: docs/internals/superpowers/plans/2026-07-19-research-design-workbench-core.md
Blocked tracked path: docs/internals/superpowers/plans/2026-09-01-critical-correctness-and-trust-hardening.md
Blocked tracked path: docs/internals/superpowers/plans/2026-10-25-browser-configuration-editing.md
Blocked tracked path: docs/internals/superpowers/plans/2026-10-25-browser-core-prerequisites.md
Blocked tracked path: docs/internals/superpowers/plans/2026-10-25-browser-rooted-persistence.md
Blocked tracked path: docs/internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md
Blocked tracked path: docs/internals/superpowers/plans/2026-10-28-all-modules-web-roadmap.md
Blocked tracked path: docs/internals/superpowers/specs/2026-04-27-fecim-skills-design.md
Blocked tracked path: docs/internals/superpowers/specs/2026-05-17-research-ingestion-design.md
Blocked tracked path: docs/internals/superpowers/specs/2026-06-17-folder-structure-design.md
Blocked tracked path: docs/internals/superpowers/specs/2026-07-19-research-design-workbench-architecture-design.md
Blocked tracked path: docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-design.md
Blocked tracked path: docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md
Blocked phrases found:
docs/internals/release/FILTER_REPO_PATHS.txt:333:docs/research-papers/tourlab/
docs/internals/release/FILTER_REPO_PATHS.txt:334:docs/4-research/papers/tourlab/
docs/internals/journal/README.md:8:- Do not include private outreach names, private addresses, private emails, or confidential feedback.
```

The latter matches include policy-like text, but those ordinary files have no approved exemption. All findings remain unresolved publication blockers; this repair does not remove content or change policy to obtain a zero exit. The publication owner must resolve the findings independently and rerun the actual gate against the final candidate. A passing fixture suite qualifies the implementation, **not permission to publish**.

Audit readback against the base confirmed all **79 unique records** and every non-path cell unchanged: **51 keep / 28 keep-with-conditions**. Only 77 exact legacy-prefix path cells changed. All 76 tracked PDFs now match. These three historical records remain:

- `docs/research/papers/external-research/urban_mining_fjh_2021.pdf`
- `validation/literature/_incoming/mdpi_ma13132968.pdf`
- `validation/literature/_incoming/pzt_nano14050432_mdpires.pdf`

Rights evidence, decisions and obligations were preserved, **not reverified**. A matching audit record is not a new legal clearance or proof that its conditions have been fulfilled. The [release readiness gate](research-software-readiness.md#release-gate--maintainer-approval-required) still applies.

Source bindings (SHA-256 of the repaired tools/audit and unchanged phrase-source files):

| Source | SHA-256 |
| --- | --- |
| `scripts/public-release/check_public_release_boundary.py` | `cb8f64fc72ed69e8a6157a08327a50052f09e6804ba48a5859c74a389bbfdc36` |
| `scripts/public-release/bootstrap_pdf_audit.py` | `3c45da2f444ce0b004deee611fad007ebbd9555126be499ce402540db29cd929` |
| `docs/internals/release/THIRD_PARTY_PDF_AUDIT.csv` | `fa0f8c9878c526cf6c26573ef8d3ee58354fb6a6d0b80bb516f94ae437c24306` |
| `docs/internals/release/FILTER_REPO_PATHS.txt` | `bb3b88ac76b9e584961c952c777d1e5d6001d68dc927e1f752958ff6a7d8ca8c` |
| `docs/internals/journal/README.md` | `cc3ad18f22be41ac7fd63b9face4e58097aeb2a5a5719c1a3b610d6a3d5992db` |
