# Autopilot Spec: Fix docs/ Format Issues

## Summary

Fix formatting issues and broken links across 77+ markdown files in `<local-path>`.

## Scope

- 77+ markdown files in docs/
- 1 mermaid diagram (SKY130.md)
- 71 files with ASCII diagrams
- Multiple tables across files

## HIGH Priority Fixes (7 instances)

### Broken `file://` Links

These machine-specific absolute paths must be converted to relative paths:

| File | Line | Issue | Fix |
|------|------|-------|-----|
| `docs/eda/SKY130.md` | 19 | `file://<local-path>` | `../sky130-reference/` |
| `docs/eda/SKY130.md` | 104 | `file://<local-path>` | `../sky130-reference/SKY130_QUICK_REFERENCE.md` |
| `docs/eda/SKY130.md` | 243 | `file://<local-path>` | `../sky130-reference/SKY130_QUICK_REFERENCE.md#metal-layer-stack` |
| `docs/eda/SKY130.md` | 244 | `file://<local-path>` | `../sky130-reference/SKY130_QUICK_REFERENCE.md#fecim-custom-cell-design-guidelines` |
| `docs/eda/SKY130.md` | 245 | `file://<local-path>` | `../sky130-reference/SKY130_QUICK_REFERENCE.md#openlane-integration` |
| `docs/eda/SKY130.md` | 251 | `file://<local-path>` | `../sky130-reference/SKY130_QUICK_REFERENCE.md` |
| `docs/opensource-tools/walkthrough_final.md` | 7 | `file://<local-path>` | `./research_notes_final.md` |

## MEDIUM Priority Fixes (4 instances)

### Old Project Name References

| File | Line | Current | Fix |
|------|------|---------|-----|
| `docs/eda/plan-demo6-improvements.md` | 211 | `github.com/yourusername/ironlattice-vis` | `github.com/your-org/fecim-lattice-tools` |
| `docs/eda/plan-demo6-improvements.md` | 381 | `github.com/[your-username]/ironlattice-vis/tree/main/module6-eda` | `github.com/your-org/fecim-lattice-tools/tree/main/module6-eda` |
| `docs/eda/plan-demo6.md` | 3282 | `github.com/XelHaku/ironlattice-vis/tree/main/module6-eda` | `github.com/your-org/fecim-lattice-tools/tree/main/module6-eda` |
| `docs/eda/plan-demo6.md` | 3354 | `Repository: github.com/XelHaku/ironlattice-vis` | `Repository: github.com/your-org/fecim-lattice-tools` |

## Validated: No Issues

- **Mermaid diagram** in SKY130.md (lines 141-150): Valid `graph LR` syntax
- **Markdown tables**: All have proper `|---|` header separators
- **ASCII art**: Properly enclosed in code blocks

## Out of Scope

- Content rewrites
- Adding new documentation
- Modifying ASCII art inside code blocks
- Changing tables inside code blocks

## Implementation Plan

1. Fix SKY130.md file:// links (6 edits)
2. Fix walkthrough_final.md broken link (1 edit)
3. Update project name in plan-demo6-improvements.md (2 edits)
4. Update project name in plan-demo6.md (2 edits)

**Total: 11 targeted edits across 4 files**
