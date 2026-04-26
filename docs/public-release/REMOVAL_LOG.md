# Removal Log

## Current-Tree Deletions
- `docs/archive/` moved to the private archive; removed from the public candidate and remains included in the history rewrite list.
- `docs/4-research/internal-analysis/` moved to the private archive; removed from the public candidate and remains included in the history rewrite list.
- `docs/4-research/transcripts/COSM_2025_AI_Hardware_Breakthrough/` removed as transcript/slide material with restricted access-restricted risk.
- `docs/4-research/transcripts/ironlattice-youtube-script.md` removed as audience-specific outreach material.
- `docs/4-research/tour-group-ironlattice-research.md` removed as research planning material.
- `docs/4-research/superlattice-material-analysis.md` removed as speculative proprietary-material analysis.
- Generated screenshots, recordings, exports, `output/validation/` content, and nested module GUI log directories remain outside the source-only boundary and in the rewrite manifest.
- Backup `.bak` files and scratch backup artifacts were removed from the public candidate and remain history rewrite targets where listed in `FILTER_REPO_PATHS.txt`.
- Root and module-local compiled binaries removed from the public candidate.
- Demo recordings, screenshot frames, and dated root screenshot artifacts removed from the public candidate.
- Deterministic validation JSON fixtures moved from runtime `output/` trees into `validation/testdata/release_artifacts/`.

## History Rewrite Targets
- All paths listed in `FILTER_REPO_PATHS.txt`.
- Every PDF row in `THIRD_PARTY_PDF_AUDIT.csv` whose `decision` is `replace-with-link` or `remove`.
- All generated-output paths listed in `FILTER_REPO_PATHS.txt`, including compiled binaries, demo recordings/screenshots, and runtime validation output directories.
