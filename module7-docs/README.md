# Module 7: Documentation

In-app documentation browser and educational curriculum viewer for fecim-lattice-tools. Provides searchable, navigable access to all project documentation from within the GUI.

## Overview

Module 7 embeds the project's documentation tree into the unified Fyne application. It renders Markdown content, supports full-text search, glossary cross-referencing, and persistent navigation state so users can explore ferroelectric CIM concepts without leaving the tool.

## Package Structure

### `pkg/gui/` — Fyne Documentation Browser

- **embedded.go** — Embeddable app for unified launcher
- **layout.go** — Documentation viewer layout (sidebar + content pane)
- **navigation.go** — Tree navigation for the `docs/` directory structure
- **search.go** — Full-text search across all documentation files
- **glossary_integration.go** — Inline glossary term highlighting and tooltips
- **persistence.go** — Save/restore last-viewed document and scroll position

## Key Types and Functions

| Type / Function | Package | Description |
|---|---|---|
| `EmbeddedDocsApp` | `pkg/gui` | Pluggable app for the unified launcher |
| `DocNavigation` | `pkg/gui` | Tree-based document navigator |
| `DocSearch` | `pkg/gui` | Full-text search engine |
| `GlossaryIntegration` | `pkg/gui` | Term detection and tooltip injection |
| `Persistence` | `pkg/gui` | Navigation state save/restore |

## Testing

```bash
# Run all module 7 tests
go test ./module7-docs/...

# Verbose
go test -v ./module7-docs/pkg/gui/...
```

Key test suite:
- `pkg/gui/` — Documentation loading, search indexing, navigation state

## Related Documentation

- `docs/documentation/` — Module index and curriculum structure
- `docs/documentation/module7-docs/` — ELI5, features
- `docs/GLOSSARY.md` — Master glossary used by the integration layer
- Repository root `README.md`
