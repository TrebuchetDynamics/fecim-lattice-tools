package gui

import (
	"fmt"
	"image/color"
	"regexp"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	fecimTheme "fecim-lattice-tools/shared/theme"
	"fecim-lattice-tools/shared/widgets"
)

// GlossaryPillsWidget displays detected glossary terms as clickable pills
type GlossaryPillsWidget struct {
	widget.BaseWidget
	terms       []string
	onTermClick func(term string)
	window      fyne.Window
	container   *fyne.Container
}

// NewGlossaryPillsWidget creates a new glossary pills widget
func NewGlossaryPillsWidget(window fyne.Window) *GlossaryPillsWidget {
	w := &GlossaryPillsWidget{
		window:    window,
		container: container.NewHBox(),
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetTerms updates the displayed terms
func (g *GlossaryPillsWidget) SetTerms(terms []string) {
	g.terms = terms
	g.rebuild()
	g.Refresh()
}

// DetectTerms scans markdown content for glossary terms
func (g *GlossaryPillsWidget) DetectTerms(markdownContent string) []string {
	return DetectGlossaryTerms(markdownContent)
}

// CreateRenderer implements fyne.Widget
func (g *GlossaryPillsWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(g.container)
}

func (g *GlossaryPillsWidget) rebuild() {
	g.container.Objects = nil

	for _, term := range g.terms {
		termCopy := term // Capture for closure
		btn := widget.NewButton(term, func() {
			if g.onTermClick != nil {
				g.onTermClick(termCopy)
			} else {
				// Default behavior: show glossary popup
				widgets.ShowGlossary(termCopy, g.window)
			}
		})
		btn.Importance = widget.LowImportance
		g.container.Add(btn)
	}
}

// DetectGlossaryTerms scans content for glossary terms
func DetectGlossaryTerms(content string) []string {
	// Convert content to lowercase for case-insensitive matching
	lowerContent := strings.ToLower(content)

	// Find unique terms that appear as whole words
	foundTerms := make(map[string]bool)

	for _, entry := range widgets.TermsData {
		termLower := strings.ToLower(entry.Term)
		// Create regex for whole-word match
		pattern := `\b` + regexp.QuoteMeta(termLower) + `\b`
		matched, err := regexp.MatchString(pattern, lowerContent)
		if err == nil && matched {
			// Store with original casing
			foundTerms[entry.Term] = true
		}
	}

	// Convert to sorted slice
	terms := make([]string, 0, len(foundTerms))
	for term := range foundTerms {
		terms = append(terms, term)
	}
	sort.Strings(terms)

	return terms
}

// CategoryBadge displays a colored category indicator
type CategoryBadge struct {
	widget.BaseWidget
	category string
	color    color.Color
}

// CategoryColors maps category names to colors
var CategoryColors = map[string]color.Color{
	"ELI5":     fecimTheme.ColorSuccess, // Green - beginner friendly
	"Physics":  fecimTheme.ColorPrimary, // Cyan - technical
	"Research": fecimTheme.ColorPurple,  // Purple - academic
	"Demo":     fecimTheme.ColorWarning, // Amber - practical
	"Guide":    fecimTheme.ColorAccent,  // Teal - howto
}

// NewCategoryBadge creates a new category badge
func NewCategoryBadge(category string) *CategoryBadge {
	badgeColor, ok := CategoryColors[category]
	if !ok {
		badgeColor = fecimTheme.ColorPrimary // Default to cyan
	}

	b := &CategoryBadge{
		category: category,
		color:    badgeColor,
	}
	b.ExtendBaseWidget(b)
	return b
}

// CreateRenderer implements fyne.Widget
func (b *CategoryBadge) CreateRenderer() fyne.WidgetRenderer {
	// Create colored border line
	border := canvas.NewRectangle(b.color)
	border.SetMinSize(fyne.NewSize(4, 0))

	// Create label
	label := canvas.NewText(b.category, theme.ForegroundColor())
	label.TextStyle = fyne.TextStyle{Bold: true}

	// Container with border and text
	content := container.NewBorder(
		nil, nil,
		border, nil,
		container.NewPadded(label),
	)

	return widget.NewSimpleRenderer(content)
}

// DocumentMetadataWidget displays document metadata with category, reading time, and glossary terms
type DocumentMetadataWidget struct {
	widget.BaseWidget
	title         string
	category      string
	readingTime   int // minutes
	glossaryTerms []string
	window        fyne.Window
	container     *fyne.Container
}

// NewDocumentMetadataWidget creates a new document metadata widget
func NewDocumentMetadataWidget(window fyne.Window) *DocumentMetadataWidget {
	w := &DocumentMetadataWidget{
		window:    window,
		container: container.NewVBox(),
	}
	w.ExtendBaseWidget(w)
	return w
}

// SetMetadata updates the metadata display
func (d *DocumentMetadataWidget) SetMetadata(title, category string, readingTime int, terms []string) {
	d.title = title
	d.category = category
	d.readingTime = readingTime
	d.glossaryTerms = terms
	d.rebuild()
	d.Refresh()
}

// CreateRenderer implements fyne.Widget
func (d *DocumentMetadataWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(d.container)
}

func (d *DocumentMetadataWidget) rebuild() {
	d.container.Objects = nil

	// First row: Category badge | Reading time
	firstRow := container.NewHBox()

	if d.category != "" {
		badge := NewCategoryBadge(d.category)
		firstRow.Add(badge)
	}

	if d.readingTime > 0 {
		if len(firstRow.Objects) > 0 {
			firstRow.Add(widget.NewLabel("|"))
		}
		readingLabel := widget.NewLabel(formatReadingTime(d.readingTime))
		readingLabel.TextStyle = fyne.TextStyle{Italic: true}
		firstRow.Add(readingLabel)
	}

	if len(firstRow.Objects) > 0 {
		d.container.Add(firstRow)
	}

	// Second row: Key Terms pills
	if len(d.glossaryTerms) > 0 {
		termsRow := container.NewHBox(
			widget.NewLabel("Key Terms:"),
		)

		pills := NewGlossaryPillsWidget(d.window)
		pills.SetTerms(d.glossaryTerms)
		termsRow.Add(pills)

		d.container.Add(termsRow)
	}
}

func formatReadingTime(minutes int) string {
	if minutes == 1 {
		return "1 min read"
	}
	return fmt.Sprintf("%d min read", minutes)
}
