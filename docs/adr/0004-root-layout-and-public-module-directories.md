# ADR 0004: Root Layout and Public Module Directories

Root cleanup should move loose support material into Purpose Directories while keeping the numbered `moduleN-*` directories stable as Public Module Directories. We chose this over a broad `modules/` rename because the numbers are part of the educational product vocabulary, docs, screenshots, and user learning path, while reusable implementation ownership is already handled by `shared/` seams.

## Consequences

- Root Collateral such as paper drafts, notebooks, screenshots, prompts, external-tool notes, and validation protocols should live under `docs/`, `tools/`, or another Purpose Directory instead of the repository root.
- `module1-hysteresis` through `module7-docs` remain public landmarks for now; make them thinner adapters over `shared/` rather than renaming them for cosmetic symmetry.
- Generated binaries and output should be rebuilt into ignored artifact locations, not tracked at the repository root.
