# External Comparator Baselines (Locked)

This directory stores immutable baseline artifacts used for cross-tool validation.

## Policy

1. **Baselines are versioned by tool + tool version + scenario.**
2. **Never overwrite historical baseline files.** Add a new versioned folder/file.
3. Every baseline update must include:
   - rationale for update,
   - tool version change and/or model change,
   - expected metric impact summary.
4. CI/comparator runs should reference explicit baseline paths, not "latest".

## Current structure

- `heracles/` — active Heracles reference data
- `crosssim/` — placeholder for CrossSim baselines
- `ngspice/` — placeholder for ngspice baselines

## Naming suggestion

Use names like:

`<scenario>__tool-<version>__date-YYYYMMDD.<ext>`

Examples:

- `pe_loop_hzo__tool-v0.4.0__date-20260213.csv`
- `read_margin_8x8__tool-42__date-20260213.json`

## Change log template

When adding a baseline, append in commit message or review note:

- Tool/version:
- Scenario:
- Source command:
- Metrics impacted:
- Reviewer:
