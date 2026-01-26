// pkg/gui/embedded.go
// Embeddable documentation viewer for the unified visualizer
package gui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// EmbeddedDocsApp is the embeddable documentation viewer
type EmbeddedDocsApp struct {
	content     fyne.CanvasObject
	currentFile string
	docsPath    string
}

// NewEmbeddedDocsApp creates a new embedded docs app instance
func NewEmbeddedDocsApp() *EmbeddedDocsApp {
	return &EmbeddedDocsApp{}
}

// docEntry represents a documentation file or folder
type docEntry struct {
	name     string
	path     string
	isDir    bool
	children []*docEntry
}

// BuildContent creates the UI content for embedding in the main app
func (app *EmbeddedDocsApp) BuildContent(fyneApp fyne.App, window fyne.Window) fyne.CanvasObject {
	// Find docs directory relative to executable or working directory
	app.docsPath = findDocsPath()

	// Create markdown viewer for content
	contentText := widget.NewRichTextFromMarkdown("# FeCIM Documentation\n\nSelect a document from the tree on the left.")
	contentText.Wrapping = fyne.TextWrapWord

	// Create scrollable content area
	contentScroll := container.NewVScroll(contentText)

	// Create file tree
	tree := app.createDocTree(contentText)

	// Create split view with tree on left, content on right
	split := container.NewHSplit(
		container.NewBorder(
			widget.NewLabelWithStyle("Documentation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			nil, nil, nil,
			tree,
		),
		contentScroll,
	)
	split.SetOffset(0.25) // 25% for tree, 75% for content

	app.content = split
	return app.content
}

// createDocTree builds the documentation file tree widget
func (app *EmbeddedDocsApp) createDocTree(contentWidget *widget.RichText) *widget.Tree {
	// Build tree data
	docs := app.scanDocsDirectory()

	// Map for quick lookup
	pathMap := make(map[string]*docEntry)
	var buildMap func(entries []*docEntry, parent string)
	buildMap = func(entries []*docEntry, parent string) {
		for _, e := range entries {
			pathMap[e.path] = e
			if e.isDir && len(e.children) > 0 {
				buildMap(e.children, e.path)
			}
		}
	}
	buildMap(docs, "")

	tree := widget.NewTree(
		// ChildUIDs - return child IDs for a node
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if uid == "" {
				// Root level
				var ids []widget.TreeNodeID
				for _, d := range docs {
					ids = append(ids, d.path)
				}
				return ids
			}
			if entry, ok := pathMap[uid]; ok && entry.isDir {
				var ids []widget.TreeNodeID
				for _, c := range entry.children {
					ids = append(ids, c.path)
				}
				return ids
			}
			return nil
		},
		// IsBranch - returns true if node has children
		func(uid widget.TreeNodeID) bool {
			if uid == "" {
				return true
			}
			if entry, ok := pathMap[uid]; ok {
				return entry.isDir && len(entry.children) > 0
			}
			return false
		},
		// Create - create a new tree node widget
		func(branch bool) fyne.CanvasObject {
			icon := widget.NewIcon(theme.DocumentIcon())
			label := widget.NewLabel("Document")
			return container.NewHBox(icon, label)
		},
		// Update - update tree node with data
		func(uid widget.TreeNodeID, branch bool, node fyne.CanvasObject) {
			box := node.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)

			if entry, ok := pathMap[uid]; ok {
				label.SetText(entry.name)
				if entry.isDir {
					icon.SetResource(theme.FolderIcon())
				} else {
					icon.SetResource(theme.DocumentIcon())
				}
			}
		},
	)

	// Handle selection - load markdown file
	tree.OnSelected = func(uid widget.TreeNodeID) {
		if entry, ok := pathMap[uid]; ok && !entry.isDir {
			content, err := os.ReadFile(entry.path)
			if err != nil {
				contentWidget.ParseMarkdown("# Error\n\nCould not read file: " + err.Error())
			} else {
				contentWidget.ParseMarkdown(string(content))
			}
			app.currentFile = entry.path
		}
	}

	return tree
}

// scanDocsDirectory recursively scans the docs directory
func (app *EmbeddedDocsApp) scanDocsDirectory() []*docEntry {
	if app.docsPath == "" {
		return []*docEntry{{
			name:  "Docs not found",
			path:  "notfound",
			isDir: false,
		}}
	}

	var entries []*docEntry
	files, err := os.ReadDir(app.docsPath)
	if err != nil {
		return entries
	}

	for _, f := range files {
		entry := app.scanEntry(filepath.Join(app.docsPath, f.Name()), f)
		if entry != nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

// scanEntry recursively scans a directory entry
func (app *EmbeddedDocsApp) scanEntry(path string, info os.DirEntry) *docEntry {
	name := info.Name()

	// Skip hidden files and non-markdown files
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return nil
	}

	if info.IsDir() {
		// Recursively scan directory
		files, err := os.ReadDir(path)
		if err != nil {
			return nil
		}

		var children []*docEntry
		for _, f := range files {
			child := app.scanEntry(filepath.Join(path, f.Name()), f)
			if child != nil {
				children = append(children, child)
			}
		}

		// Only include directories that have markdown content
		if len(children) > 0 {
			return &docEntry{
				name:     name,
				path:     path,
				isDir:    true,
				children: children,
			}
		}
		return nil
	}

	// Only include markdown files
	if strings.HasSuffix(strings.ToLower(name), ".md") {
		return &docEntry{
			name:  name,
			path:  path,
			isDir: false,
		}
	}

	return nil
}

// findDocsPath locates the docs directory
func findDocsPath() string {
	// Try relative to working directory
	candidates := []string{
		"docs",
		"../docs",
		"../../docs",
	}

	// Also try relative to executable
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, "docs"),
			filepath.Join(exeDir, "..", "docs"),
		)
	}

	for _, candidate := range candidates {
		if abs, err := filepath.Abs(candidate); err == nil {
			if info, err := os.Stat(abs); err == nil && info.IsDir() {
				return abs
			}
		}
	}

	return ""
}

// Start is called when this demo tab is selected
func (app *EmbeddedDocsApp) Start() {
	// No background processes to start
}

// Stop is called when this demo tab is deselected
func (app *EmbeddedDocsApp) Stop() {
	// No background processes to stop
}
