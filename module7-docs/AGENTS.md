<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# Module 7: In-App Documentation Viewer

## Purpose

Module 7 provides an in-app documentation viewer integrated into the unified GUI. It presents project documentation, API references, tutorials, and educational content within the same Fyne application as the simulation modules.

Key features:
- File-tree based navigation through docs/ directory
- Search index with keyword and content search
- Breadcrumb navigation and history tracking
- Table of contents for long documents
- Glossary integration with term highlighting
- Keyboard shortcuts for navigation
- Module quick-access shortcuts
- Rich text rendering with links and formatting

## Key Files

### Core GUI
- `pkg/gui/embedded.go` - EmbeddedDocsApp, main UI builder, file tree navigation
- `pkg/gui/app.go` - Application lifecycle (BuildContent, Start, Stop)

### Navigation & Search
- `pkg/gui/navigation.go` - Navigation logic and state management
- `pkg/gui/search.go` - Search index and query matching
- `pkg/gui/persistence.go` - History tracking and navigation state persistence
- `pkg/gui/keyboard.go` - Keyboard shortcuts and input handling

### Layout & Widgets
- `pkg/gui/layout.go` - Custom layout managers for doc viewer
- Breadcrumb, TableOfContents, DocumentMetadata, GlossaryPills, SearchDialog, ModuleShortcuts widget implementations

### Content Processing
- `pkg/gui/glossary_integration.go` - Glossary term matching and highlighting

### Testing
- `pkg/gui/docs_test.go` - Basic functionality tests
- `pkg/gui/docs_integrity_test.go` - Content validation and structure tests

## Subdirectories

```
module7-docs/
├── pkg/
│   └── gui/                     # Fyne GUI components and application
├── README.md
└── AGENTS.md                   # This file
```

## For AI Agents

### Working

**Current State:**
- File tree navigation of docs/ directory (recursive)
- Search index built on startup, supports keyword and full-text search
- Navigation history with back/forward support
- Breadcrumb widget shows current location
- Table of contents extracted from markdown headers
- Glossary integration highlights defined terms
- Keyboard shortcuts for common operations (Ctrl+F for search, Alt+Left/Right for nav)
- Module quick-access panel for jumping to specific modules
- Document metadata displayed (path, last updated, related docs)

**Task Pattern:**
1. Understand EmbeddedDocsApp structure (UI components, state, file tree)
2. Docs are loaded from `utils.FindDirectory("docs/documentation")` or fallback to "docs"
3. File tree is built recursively from docs/ structure
4. Search index built on startup via `NewSearchIndex()`
5. Navigation state persisted via `DocsHistory`

**Key Patterns:**
- EmbeddedDocsApp implements shared widget interface: BuildContent(), Start(), Stop()
- docEntry tree structure mirrors file system hierarchy
- Search is async (non-blocking) with ranked results
- History is bidirectional (back/forward) with current position tracking
- Keyboard input processed via standard Go keyboard events
- Content rendering uses Fyne RichText widget

### Testing

**Test Files:**
- `pkg/gui/docs_test.go` - Basic navigation, search, history tests
- `pkg/gui/docs_integrity_test.go` - Documentation structure validation, link checking, metadata verification

**Run Tests:**
```bash
go test ./module7-docs/...                    # All tests
go test -v ./module7-docs/pkg/gui             # GUI tests
go test -v ./module7-docs/pkg/gui -run Integrity  # Integrity checks
```

**Test Coverage Notes:**
- Navigation: file tree traversal, path handling
- Search: index building, query matching, result ranking
- History: back/forward, state persistence
- Content: markdown parsing, link extraction, metadata
- Integrity: document structure, missing files, circular references

### Patterns

**File Tree Navigation:**
- Recursive traversal of docs/ directory on startup
- docEntry tree structure with name, path, isDir, children
- Tree widget shows directory structure with expand/collapse
- Clicking file loads content into RichText widget

**Search System:**
- SearchIndex built from all .md files at startup
- Keyword matching: split on whitespace/punctuation
- Full-text search: basic substring matching (future: regex support)
- Results ranked by relevance (exact matches first, then partial)
- Search is case-insensitive

**Navigation History:**
- DocsHistory maintains stack of visited documents
- Current position tracked, enables back/forward
- History persisted to disk for session continuation
- Limit history size to prevent memory growth

**Content Rendering:**
- RichText widget supports markdown-like formatting (bold, italic, monospace)
- Links extracted and made clickable (internal navigation)
- Headers used for table of contents extraction
- Code blocks preserved with monospace rendering

**Keyboard Shortcuts:**
- Ctrl+F: Open search dialog
- Alt+Left / Alt+Right: Navigate back/forward
- Escape: Close search or dialog
- Enter: Navigate to selected search result
- Tab: Move between UI elements (standard Fyne behavior)

**Glossary Integration:**
- Glossary terms extracted from docs/
- GlossaryPillsWidget shows relevant terms for current document
- Term highlighting in content (future enhancement)
- Clicking term navigates to glossary definition

## Dependencies

**Internal:**
- `shared/widgets` - EmbeddedAppBase interface and reusable widgets
- `shared/utils` - Directory finding utilities
- `shared/logging` - Logging infrastructure

**External:**
- `fyne.io/fyne/v2` - GUI framework (container, layout, widget, theme)
- Standard Go packages (os, path/filepath, sort, strings, net/url)

## MANUAL

### Adding Documentation

1. **Create Markdown File** in `docs/documentation/` or subdirectory:
   ```markdown
   # Document Title

   ## Section 1
   Content here...

   ### Subsection
   More content...
   ```

2. **Search Index Updates Automatically** on app restart

3. **Glossary Terms** are extracted from `docs/GLOSSARY.md` (if it exists):
   ```markdown
   ## Term Name
   Definition of the term...
   ```

4. **Links** within documentation should use relative paths:
   ```markdown
   [Link Text](../other-doc.md)
   [Cross-module](../../module2-crossbar/README.md)
   ```

### Customizing UI Components

**Breadcrumbs:**
- Automatically built from current file path
- Clicking breadcrumb navigates to that level

**Table of Contents:**
- Generated from markdown headers (# H1, ## H2, etc)
- Clicking TOC entry jumps to section (future: anchor support)

**Metadata Widget:**
- Shows: file path, size, last modified date
- Displays related documents (same directory, same topic)

**Module Shortcuts:**
- Quick access to each module's documentation
- Edit in `createUIComponents()` to match module structure

**Search Dialog:**
- Appears on Ctrl+F or search button click
- Shows ranked results with preview
- Result preview shows surrounding context

### Keyboard Navigation Optimization

For accessibility, test with screen readers:
```bash
# Enable screen reader mode (varies by platform)
FECIM_A11Y=screenreader ./fecim-lattice-tools
```

Keyboard-only navigation should work:
- Tab: Move between widgets
- Enter: Activate buttons/select results
- Arrow keys: Scroll content and navigate trees

### Debugging Search

Enable debug logging via `FECIM_DEBUG=docs`:
```bash
FECIM_DEBUG=docs ./fecim-lattice-tools
```

This logs:
- Search index build progress
- Query execution and result ranking
- History operations (push/pop/navigate)
- Content loading and rendering

### Content Integrity Checks

The integrity test validates:
- All linked files exist and are readable
- Markdown syntax is valid
- No circular references in linked docs
- Glossary terms are defined
- Metadata is complete (title, summary, etc)

Run validation:
```bash
go test -v ./module7-docs/pkg/gui -run Integrity
```

### Extending with New Widget

1. **Create Widget** in `pkg/gui/` (e.g., `my_widget.go`):
   ```go
   type MyWidget struct {
       // State
   }

   func (w *MyWidget) Render() fyne.CanvasObject {
       return container.NewVBox(...)
   }
   ```

2. **Integrate** into EmbeddedDocsApp.createUIComponents():
   ```go
   app.myWidget = NewMyWidget()
   ```

3. **Add Tests** in `docs_test.go`

### Future Enhancements

Potential improvements (not yet implemented):
- Full-text search with ranking (tf-idf)
- Anchor links within documents (#section-heading)
- Dark mode support for readability
- Document version history
- Collaborative editing (future: if multi-user)
- LaTeX math rendering
- Diagram support (Mermaid, Graphviz)

---

**Last Updated:** 2026-02-13
