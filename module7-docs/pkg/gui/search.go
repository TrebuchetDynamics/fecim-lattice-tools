// Package gui provides the documentation viewer with search capabilities.
package gui

import (
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"unicode"

	"fecim-lattice-tools/shared/mathutil"
	docscore "fecim-lattice-tools/workbench/modules/docs"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// IndexEntry represents a term occurrence in a document.
type IndexEntry = docscore.IndexEntry

// SearchDocMetadata stores cached metadata about a document for search purposes.
// This extends the basic DocMetadata with additional search-relevant fields.
type SearchDocMetadata = docscore.SearchDocMetadata

// SearchIndex provides full-text search over documentation.
type SearchIndex struct {
	mu       sync.RWMutex
	index    map[string][]IndexEntry       // term -> documents
	docs     map[string]*SearchDocMetadata // path -> metadata
	docsPath string
}

// SearchResult represents a single search result.
type SearchResult = docscore.SearchResult

const (
	titleMatchBoost    = 3.0
	headingMatchBoost  = 2.0
	glossaryMatchBoost = 1.5
	exactMatchBoost    = 1.5
)

// searchToken represents a parsed token with position info.
type searchToken struct {
	term      string
	position  int
	inTitle   bool
	inHeading bool
}

// NewSearchIndex creates a new search index (built lazily on first query).
func NewSearchIndex(docsPath string) *SearchIndex {
	si := &SearchIndex{
		index:    make(map[string][]IndexEntry),
		docs:     make(map[string]*SearchDocMetadata),
		docsPath: docsPath,
	}
	// Index is built lazily on first Query() call to speed up app startup
	return si
}

// Build walks all .md files and builds the inverted index.
func (si *SearchIndex) Build() {
	si.mu.Lock()
	defer si.mu.Unlock()

	// Clear existing index
	si.index = make(map[string][]IndexEntry)
	si.docs = make(map[string]*SearchDocMetadata)

	if si.docsPath == "" {
		return
	}

	// Walk all markdown files
	filepath.Walk(si.docsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden files/directories
		if strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "_") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process markdown files
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			return nil
		}

		si.indexDocument(path, info)
		return nil
	})
}

// indexDocument indexes a single markdown document.
func (si *SearchIndex) indexDocument(path string, info os.FileInfo) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	text := string(content)

	// Extract metadata
	metadata := si.extractMetadata(path, text, info)
	si.docs[path] = metadata

	// Tokenize and index
	tokens := si.tokenize(text)

	// Build term frequency map
	termFreq := make(map[string]*IndexEntry)
	for _, tok := range tokens {
		entry, exists := termFreq[tok.term]
		if !exists {
			entry = &IndexEntry{
				DocPath:   path,
				Frequency: 0,
				InTitle:   false,
				InHeading: false,
			}
			termFreq[tok.term] = entry
		}
		entry.Frequency++
		if tok.inTitle {
			entry.InTitle = true
		}
		if tok.inHeading {
			entry.InHeading = true
		}
	}

	// Generate snippets and add to index
	for term, entry := range termFreq {
		entry.Snippet = si.extractSnippet(text, term)
		si.index[term] = append(si.index[term], *entry)
	}
}

// extractMetadata parses document metadata.
func (si *SearchIndex) extractMetadata(path, content string, info os.FileInfo) *SearchDocMetadata {
	meta := &SearchDocMetadata{
		Path:         path,
		LastModified: info.ModTime(),
	}

	// Extract title from first # heading
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			meta.Title = strings.TrimPrefix(trimmed, "# ")
			break
		}
	}
	if meta.Title == "" {
		// Use filename without extension as fallback
		meta.Title = strings.TrimSuffix(filepath.Base(path), ".md")
	}

	// Detect category based on path/filename
	meta.Category = si.detectCategory(path)

	// Calculate reading time (words / 200 wpm)
	wordCount := len(strings.Fields(content))
	meta.ReadingTime = int(math.Ceil(float64(wordCount) / 200.0))
	if meta.ReadingTime < 1 {
		meta.ReadingTime = 1
	}

	// Detect glossary terms
	meta.GlossaryTerms = si.detectGlossaryTerms(content)

	return meta
}

// detectCategory determines the document category from its path.
func (si *SearchIndex) detectCategory(path string) string {
	pathLower := strings.ToLower(path)
	filenameLower := strings.ToLower(filepath.Base(path))

	// Check filename first (higher priority)
	switch filenameLower {
	case "eli5.md":
		return "ELI5"
	case "physics.md":
		return "Physics"
	case "features.md", "opensource-tools.md":
		return "Guide"
	}
	if strings.Contains(filenameLower, "research") {
		return "Research"
	}
	if strings.Contains(filenameLower, "demo") {
		return "Demo"
	}

	// Check path for folder names
	if strings.Contains(pathLower, "/research-papers/") {
		return "Research"
	}
	if strings.Contains(pathLower, "/cim/") || strings.Contains(pathLower, "/crossbar/") {
		return "Physics"
	}

	return "Guide"
}

// detectGlossaryTerms finds glossary terms in the content.
func (si *SearchIndex) detectGlossaryTerms(content string) []string {
	return DetectGlossaryTerms(content)
}

// tokenize splits content into searchable tokens.
func (si *SearchIndex) tokenize(content string) []searchToken {
	var tokens []searchToken
	lines := strings.Split(content, "\n")
	position := 0
	inTitle := false
	inHeading := false

	for lineIdx, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect if this is a title (first # heading)
		isFirstHeading := lineIdx == 0 || (position == 0 && strings.HasPrefix(trimmed, "# "))
		if strings.HasPrefix(trimmed, "# ") && isFirstHeading {
			inTitle = true
		} else {
			inTitle = false
		}

		// Detect any heading
		inHeading = strings.HasPrefix(trimmed, "#")

		// Extract words from line
		words := si.extractWords(line)
		for _, word := range words {
			if len(word) >= 2 { // Skip single-char tokens
				tokens = append(tokens, searchToken{
					term:      word,
					position:  position,
					inTitle:   inTitle,
					inHeading: inHeading,
				})
				position++
			}
		}
	}

	return tokens
}

// extractWords splits a line into lowercase words, stripping punctuation.
func (si *SearchIndex) extractWords(line string) []string {
	var words []string

	// Split on whitespace
	parts := strings.Fields(line)
	for _, part := range parts {
		// Strip punctuation and convert to lowercase
		word := strings.Map(func(r rune) rune {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				return unicode.ToLower(r)
			}
			return -1 // Remove the character
		}, part)

		if word != "" {
			words = append(words, word)
		}
	}

	return words
}

// extractSnippet gets ~100 chars of context around a term match.
func (si *SearchIndex) extractSnippet(content, term string) string {
	contentLower := strings.ToLower(content)
	termLower := strings.ToLower(term)

	idx := strings.Index(contentLower, termLower)
	if idx == -1 {
		// Return beginning of content
		if len(content) > 100 {
			return content[:100] + "..."
		}
		return content
	}

	// Get context around match
	start := idx - 40
	if start < 0 {
		start = 0
	}
	end := idx + len(term) + 60
	if end > len(content) {
		end = len(content)
	}

	// Find word boundaries
	for start > 0 && content[start] != ' ' && content[start] != '\n' {
		start--
	}
	for end < len(content) && content[end] != ' ' && content[end] != '\n' {
		end++
	}

	snippet := strings.TrimSpace(content[start:end])
	// Remove excessive whitespace/newlines
	snippet = regexp.MustCompile(`\s+`).ReplaceAllString(snippet, " ")

	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}

// Query performs a fuzzy search with TF-IDF ranking.
func (si *SearchIndex) Query(query string, limit int) []SearchResult {
	si.mu.RLock()
	needsBuild := len(si.docs) == 0 && si.docsPath != ""
	si.mu.RUnlock()
	if needsBuild {
		si.Build()
	}
	si.mu.RLock()
	defer si.mu.RUnlock()
	return docscore.Rank(si.index, si.docs, query, limit)
}

// findMatchingTerms returns terms that match (exact or fuzzy).
func (si *SearchIndex) findMatchingTerms(query string) []string {
	var matches []string

	for term := range si.index {
		// Exact match
		if term == query {
			matches = append(matches, term)
			continue
		}

		// Prefix match
		if strings.HasPrefix(term, query) {
			matches = append(matches, term)
			continue
		}

		// Contains match (for longer queries)
		if len(query) >= 3 && strings.Contains(term, query) {
			matches = append(matches, term)
			continue
		}

		// Fuzzy match using edit distance for short terms
		if len(query) >= 3 && len(term) >= 3 {
			if editDistance(term, query) <= 1 {
				matches = append(matches, term)
			}
		}
	}

	return matches
}

// editDistance calculates Levenshtein distance between two strings.
func editDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Optimization: if strings differ by more than 2 in length, skip
	if mathutil.AbsInt(len(a)-len(b)) > 2 {
		return mathutil.AbsInt(len(a) - len(b))
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := range matrix[0] {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = min(min(
				matrix[i-1][j]+1,  // deletion
				matrix[i][j-1]+1), // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func hasGlossaryOverlap(glossarySet map[string]struct{}, docTerms []string) bool {
	for _, term := range docTerms {
		if _, ok := glossarySet[term]; ok {
			return true
		}
	}
	return false
}

// GetDocMetadata returns metadata for a document path.
// If index hasn't been built yet, indexes just this document.
func (si *SearchIndex) GetDocMetadata(path string) *SearchDocMetadata {
	si.mu.RLock()
	meta, exists := si.docs[path]
	si.mu.RUnlock()

	if exists {
		return meta
	}

	// Index just this one document if not found
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}

	si.mu.Lock()
	defer si.mu.Unlock()
	si.indexDocument(path, info)
	return si.docs[path]
}

// SearchDialog provides a modal search dialog with keyboard navigation.
type SearchDialog struct {
	widget.BaseWidget

	index       *SearchIndex
	entry       *widget.Entry
	resultsList *widget.List
	results     []SearchResult
	selected    int
	dialog      dialog.Dialog
	parent      fyne.Window
	OnSelected  func(path string)

	mu sync.Mutex
}

// NewSearchDialog creates a new search dialog.
func NewSearchDialog(index *SearchIndex, parent fyne.Window, onSelected func(path string)) *SearchDialog {
	sd := &SearchDialog{
		index:      index,
		parent:     parent,
		OnSelected: onSelected,
		results:    []SearchResult{},
		selected:   -1,
	}

	sd.entry = widget.NewEntry()
	sd.entry.SetPlaceHolder("Search documentation...")
	sharedwidgets.SetAccessibleLabel(sd.entry, "Documentation search query")
	sd.entry.OnChanged = sd.onQueryChanged

	sd.resultsList = widget.NewList(
		func() int {
			sd.mu.Lock()
			defer sd.mu.Unlock()
			return len(sd.results)
		},
		func() fyne.CanvasObject {
			categoryBadge := canvas.NewText("Guide ", theme.ForegroundColor())
			categoryBadge.TextStyle = fyne.TextStyle{Bold: true}
			categoryBadge.TextSize = 14
			titleLabel := widget.NewLabel("Title")
			return container.NewHBox(
				widget.NewIcon(theme.DocumentIcon()),
				container.NewVBox(
					container.NewHBox(categoryBadge, titleLabel),
					widget.NewLabel("Snippet..."),
				),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			sd.mu.Lock()
			if id >= len(sd.results) {
				sd.mu.Unlock()
				return
			}
			result := sd.results[id]
			sd.mu.Unlock()

			box := obj.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			textBox := box.Objects[1].(*fyne.Container)
			titleRow := textBox.Objects[0].(*fyne.Container)
			categoryBadge := titleRow.Objects[0].(*canvas.Text)
			titleLabel := titleRow.Objects[1].(*widget.Label)
			snippetLabel := textBox.Objects[1].(*widget.Label)

			// Set icon based on category
			icon.SetResource(sd.getCategoryIcon(result.Category))

			titleLabel.SetText(result.Title)
			titleLabel.TextStyle = fyne.TextStyle{Bold: true}

			categoryText := result.Category
			if categoryText != "" {
				categoryText = categoryText + " "
			}
			categoryBadge.Text = categoryText
			if badgeColor, ok := CategoryColors[result.Category]; ok {
				categoryBadge.Color = badgeColor
			} else {
				categoryBadge.Color = theme.ForegroundColor()
			}
			categoryBadge.Refresh()

			// Truncate snippet
			snippet := result.Snippet
			if result.MatchType != "" {
				snippet = strings.Title(result.MatchType) + " • " + snippet
			}
			if len(snippet) > 80 {
				snippet = snippet[:77] + "..."
			}
			snippetLabel.SetText(snippet)
			snippetLabel.TextStyle = fyne.TextStyle{}
		},
	)

	sd.resultsList.OnSelected = func(id widget.ListItemID) {
		sd.mu.Lock()
		if id < len(sd.results) {
			path := sd.results[id].DocPath
			sd.mu.Unlock()
			sd.selectResult(path)
		} else {
			sd.mu.Unlock()
		}
	}

	sd.ExtendBaseWidget(sd)
	return sd
}

// getCategoryIcon returns an appropriate icon for the document category.
func (sd *SearchDialog) getCategoryIcon(category string) fyne.Resource {
	switch category {
	case "ELI5":
		return theme.HelpIcon()
	case "Physics":
		return theme.ComputerIcon()
	case "Research":
		return theme.SearchIcon()
	case "Demo":
		return theme.MediaPlayIcon()
	case "Guide":
		return theme.ListIcon()
	default:
		return theme.DocumentIcon()
	}
}

// onQueryChanged handles search input changes.
func (sd *SearchDialog) onQueryChanged(query string) {
	if sd.index == nil {
		return
	}

	results := sd.index.Query(query, 10)

	sd.mu.Lock()
	sd.results = results
	sd.selected = -1
	if len(results) > 0 {
		sd.selected = 0
	}
	sd.mu.Unlock()

	sd.resultsList.Refresh()
	if sd.selected >= 0 {
		sd.resultsList.Select(sd.selected)
	}
}

// selectResult handles result selection.
func (sd *SearchDialog) selectResult(path string) {
	if sd.dialog != nil {
		sd.dialog.Hide()
	}
	if sd.OnSelected != nil {
		sd.OnSelected(path)
	}
}

// MoveSelection moves the selection up or down.
func (sd *SearchDialog) MoveSelection(delta int) {
	sd.mu.Lock()
	count := len(sd.results)
	if count == 0 {
		sd.mu.Unlock()
		return
	}

	sd.selected += delta
	if sd.selected < 0 {
		sd.selected = count - 1
	} else if sd.selected >= count {
		sd.selected = 0
	}
	selected := sd.selected
	sd.mu.Unlock()

	sd.resultsList.Select(selected)
}

// ConfirmSelection selects the current highlighted result.
func (sd *SearchDialog) ConfirmSelection() {
	sd.mu.Lock()
	if sd.selected >= 0 && sd.selected < len(sd.results) {
		path := sd.results[sd.selected].DocPath
		sd.mu.Unlock()
		sd.selectResult(path)
	} else {
		sd.mu.Unlock()
	}
}

// Show displays the search dialog.
func (sd *SearchDialog) Show() {
	// Clear previous state
	sd.entry.SetText("")
	sd.mu.Lock()
	sd.results = []SearchResult{}
	sd.selected = -1
	sd.mu.Unlock()
	sd.resultsList.Refresh()

	content := container.NewBorder(
		sd.entry,
		nil, nil, nil,
		sd.resultsList,
	)

	// Create custom dialog
	sd.dialog = dialog.NewCustom("Search Documentation", "Cancel", content, sd.parent)
	sd.dialog.Resize(fyne.NewSize(600, 400))
	sd.dialog.Show()

	// Focus the entry
	sd.parent.Canvas().Focus(sd.entry)
}

// CreateRenderer implements fyne.Widget.
func (sd *SearchDialog) CreateRenderer() fyne.WidgetRenderer {
	content := container.NewBorder(
		sd.entry,
		nil, nil, nil,
		sd.resultsList,
	)
	return widget.NewSimpleRenderer(content)
}

// searchShortcut implements fyne.Shortcut for Cmd/Ctrl+K.
type searchShortcut struct{}

func (s *searchShortcut) ShortcutName() string {
	return "Search"
}

// SetupSearchShortcut configures search shortcuts (Cmd/Ctrl+K and /).
func SetupSearchShortcut(window fyne.Window, searchDialog *SearchDialog) {
	if window == nil || searchDialog == nil {
		return
	}

	// Add keyboard shortcut for Cmd/Ctrl+K
	ctrlK := &desktop.CustomShortcut{
		KeyName:  fyne.KeyK,
		Modifier: fyne.KeyModifierShortcutDefault,
	}
	window.Canvas().AddShortcut(ctrlK, func(shortcut fyne.Shortcut) {
		searchDialog.Show()
	})

	// Add keyboard shortcut for / (quick find style)
	slashShortcut := &desktop.CustomShortcut{
		KeyName:  fyne.KeySlash,
		Modifier: 0,
	}
	window.Canvas().AddShortcut(slashShortcut, func(shortcut fyne.Shortcut) {
		searchDialog.Show()
	})

	// Custom shortcut handling via key events on entry
	searchDialog.entry.OnSubmitted = func(s string) {
		searchDialog.ConfirmSelection()
	}

	// Note: For proper Cmd+K support, use TypedShortcut on the window's canvas
	// This requires implementing a custom shortcut type or using Fyne's built-in
	// shortcut mechanism with a custom key handler.
	window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		// Handle up/down arrows when search dialog is visible
		if searchDialog.dialog != nil {
			switch key.Name {
			case fyne.KeyUp:
				searchDialog.MoveSelection(-1)
			case fyne.KeyDown:
				searchDialog.MoveSelection(1)
			case fyne.KeyReturn:
				searchDialog.ConfirmSelection()
			case fyne.KeyEscape:
				searchDialog.dialog.Hide()
			}
		}
	})
}

// HighlightMatch wraps matching text with emphasis markers.
func HighlightMatch(text, query string) string {
	if query == "" {
		return text
	}

	// Case-insensitive replacement
	re, err := regexp.Compile("(?i)(" + regexp.QuoteMeta(query) + ")")
	if err != nil {
		return text
	}

	return re.ReplaceAllString(text, "**$1**")
}

// GetCategoryColor returns a color name for the category.
func GetCategoryColor(category string) string {
	switch category {
	case "ELI5":
		return "green"
	case "Physics":
		return "cyan"
	case "Research":
		return "purple"
	case "Demo":
		return "amber"
	case "Guide":
		return "teal"
	default:
		return "gray"
	}
}
