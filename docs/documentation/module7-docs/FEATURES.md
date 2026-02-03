# Module 7: Docs - Features

## What This Module Does

- Provides a curriculum-first documentation viewer.
- Offers search, breadcrumbs, ToC, and glossary integration.
- Adds module shortcuts for ELI5, PHYSICS, FEATURES, and TOOLS.

## Primary Components

- `module7-docs/pkg/gui/embedded.go`
- `module7-docs/pkg/gui/search.go`
- `module7-docs/pkg/gui/navigation.go`
- `module7-docs/pkg/gui/glossary_integration.go`

## Key Workflows

- Select a document from the curriculum tree.
- Use module shortcuts to jump between learning layers.
- Use search and glossary pills for cross-topic navigation.

## Extension Points

- Add new category rules in `search.go`.
- Extend the module shortcuts panel for new curriculum layers.
- Customize layout breakpoints in `layout.go`.

## Known Limitations

- Markdown rendering is limited to supported Fyne widgets.
- Search is in-memory and optimized for repo scale.
- No external URLs are fetched or embedded.
