// Reader-owned search extracted from module7-docs. No GUI or filesystem imports.
package docs

import (
	"fecim-lattice-tools/shared/mathutil"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
)

// IndexEntry represents a term occurrence in a document.
type IndexEntry struct {
	DocPath   string
	Frequency int    // number of term occurrences
	InTitle   bool   // term found in document title
	InHeading bool   // term found in a heading
	Snippet   string // context around match
}

// SearchDocMetadata stores cached metadata about a document for search purposes.
// This extends the basic DocMetadata with additional search-relevant fields.
type SearchDocMetadata struct {
	Path          string
	Title         string   // extracted from first # heading
	Category      string   // ELI5, Physics, Research, Demo, Guide
	ReadingTime   int      // estimated minutes (words / 200)
	GlossaryTerms []string // detected glossary terms
	LastModified  time.Time
}

// SearchResult represents a single search result.
type SearchResult struct {
	DocPath   string
	Title     string
	Category  string  // ELI5, Physics, Research, Demo, Guide
	Snippet   string  // ~100 chars with match context
	MatchType string  // "title", "heading", "content", "glossary"
	Relevance float64 // TF-IDF score
}

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

type SearchIndex struct {
	index map[string][]IndexEntry
	docs  map[string]*SearchDocMetadata
}

func Rank(index map[string][]IndexEntry, docs map[string]*SearchDocMetadata, query string, limit int) []SearchResult {
	return (&SearchIndex{index: index, docs: docs}).Query(query, limit)
}
func newIndex() *SearchIndex {
	return &SearchIndex{index: map[string][]IndexEntry{}, docs: map[string]*SearchDocMetadata{}}
}
func (si *SearchIndex) add(d Document, text string) {
	si.docs[d.ID] = &SearchDocMetadata{Path: d.ID, Title: d.Title, Category: d.Category, GlossaryTerms: d.Terms}
	freq := map[string]*IndexEntry{}
	for _, tok := range si.tokenize(text) {
		e := freq[tok.term]
		if e == nil {
			e = &IndexEntry{DocPath: d.ID}
			freq[tok.term] = e
		}
		e.Frequency++
		e.InTitle = e.InTitle || tok.inTitle
		e.InHeading = e.InHeading || tok.inHeading
	}
	for term, e := range freq {
		e.Snippet = si.extractSnippet(text, term)
		si.index[term] = append(si.index[term], *e)
	}
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
	if query == "" || len(si.docs) == 0 {
		return nil
	}

	// Parse query into terms
	queryTerms := si.extractWords(query)
	if len(queryTerms) == 0 {
		return nil
	}

	// Calculate document scores
	docScores := make(map[string]float64)
	docMatchTypes := make(map[string]string)
	docSnippets := make(map[string]string)

	totalDocs := float64(len(si.docs))

	for _, term := range queryTerms {
		// Find matching entries (exact and fuzzy)
		matchingTerms := si.findMatchingTerms(term)

		for _, matchTerm := range matchingTerms {
			entries, exists := si.index[matchTerm]
			if !exists {
				continue
			}

			// IDF: log(N / df) where df is document frequency.
			// Clamp to a small floor so common terms still retain ranking signal
			// from title/heading boosts and frequency.
			idf := math.Log(totalDocs / float64(len(entries)))
			if idf <= 0 {
				idf = 0.1
			}

			for _, entry := range entries {
				// TF: 1 + log(frequency)
				tf := 1.0 + math.Log(float64(entry.Frequency))

				// Base score
				score := tf * idf

				// Boost for title matches
				if entry.InTitle {
					score *= titleMatchBoost
					docMatchTypes[entry.DocPath] = "title"
				} else if entry.InHeading {
					score *= headingMatchBoost
					if docMatchTypes[entry.DocPath] != "title" {
						docMatchTypes[entry.DocPath] = "heading"
					}
				} else if docMatchTypes[entry.DocPath] == "" {
					docMatchTypes[entry.DocPath] = "content"
				}

				// Exact match bonus
				if matchTerm == term {
					score *= exactMatchBoost
				}

				docScores[entry.DocPath] += score

				// Keep best snippet
				if entry.Snippet != "" {
					if existing, ok := docSnippets[entry.DocPath]; !ok || len(entry.Snippet) > len(existing) {
						docSnippets[entry.DocPath] = entry.Snippet
					}
				}
			}
		}
	}

	// Check for glossary term matches
	glossaryQueryTerms := DetectGlossaryTerms(query)
	if len(glossaryQueryTerms) > 0 {
		glossarySet := make(map[string]struct{}, len(glossaryQueryTerms))
		for _, term := range glossaryQueryTerms {
			glossarySet[term] = struct{}{}
		}

		for path, meta := range si.docs {
			if meta == nil || len(meta.GlossaryTerms) == 0 {
				continue
			}
			if !hasGlossaryOverlap(glossarySet, meta.GlossaryTerms) {
				continue
			}
			if score, ok := docScores[path]; ok {
				docScores[path] = score * glossaryMatchBoost
				if docMatchTypes[path] == "" {
					docMatchTypes[path] = "glossary"
				}
			}
		}
	}

	// Convert to results and sort
	var results []SearchResult
	for path, score := range docScores {
		meta := si.docs[path]
		title := path
		category := ""
		if meta != nil {
			title = meta.Title
			category = meta.Category
		}

		snippet := docSnippets[path]
		if snippet == "" && meta != nil {
			snippet = "Category: " + meta.Category
		}

		results = append(results, SearchResult{
			DocPath:   path,
			Title:     title,
			Category:  category,
			Snippet:   snippet,
			MatchType: docMatchTypes[path],
			Relevance: score,
		})
	}

	// Sort by relevance (descending)
	sort.Slice(results, func(i, j int) bool {
		if results[i].Relevance == results[j].Relevance {
			return results[i].DocPath < results[j].DocPath
		}
		return results[i].Relevance > results[j].Relevance
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	return results
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

	sort.Strings(matches)
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
