// Package docs provides bounded, read-only documentation indexing and presentation.
package docs

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	MaxDocuments     = 2048
	MaxDocumentBytes = 1 << 20
	MaxCorpusBytes   = 10 << 20
	PageSize         = 20
)

type Source struct {
	ID          string            `json:"id"`
	Path        string            `json:"path"`
	SHA256      string            `json:"sha256"`
	Rights      string            `json:"rights"`
	Attribution string            `json:"attribution"`
	Category    string            `json:"category"`
	Omit        map[string]string `json:"omitted_references,omitempty"`
	Text        string            `json:"text"`
}
type Node struct {
	Kind     string `json:"kind"`
	Text     string `json:"text,omitempty"`
	URL      string `json:"url,omitempty"`
	Level    int    `json:"level,omitempty"`
	Anchor   string `json:"anchor,omitempty"`
	Ordered  bool   `json:"ordered,omitempty"`
	Children []Node `json:"children,omitempty"`
}
type Heading struct {
	Text   string `json:"text"`
	Level  int    `json:"level"`
	Anchor string `json:"anchor"`
}
type Document struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	SHA256      string    `json:"sha256"`
	Rights      string    `json:"rights"`
	Attribution string    `json:"attribution"`
	Bytes       int       `json:"bytes"`
	Terms       []string  `json:"terms"`
	Headings    []Heading `json:"headings"`
	Nodes       []Node    `json:"nodes"`
}
type Corpus struct {
	Sources   []Source   `json:"sources"`
	Documents []Document `json:"documents"`
	Bytes     int        `json:"bytes"`
	Digest    string     `json:"digest"`
	index     *SearchIndex
}
type Page struct {
	Query    string         `json:"query"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Results  []SearchResult `json:"results"`
}

func Hash(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,79}$`)

func publicPath(p string) bool {
	return strings.HasPrefix(p, "docs/guides/") || strings.HasPrefix(p, "docs/modules/") || p == "docs/research/honesty-audit.md" || p == "docs/research/validation/park-reference-correction.md"
}

// Build accepts only explicitly supplied publication sources. It never walks a
// directory, follows a network reference, reads cwd or writes history.
func Build(sources []Source) (*Corpus, error) { return build(sources, false) }
func build(sources []Source, granted bool) (*Corpus, error) {
	if len(sources) > MaxDocuments {
		return nil, errors.New("document limit exceeded")
	}
	c := &Corpus{Sources: append([]Source{}, sources...), Documents: []Document{}, index: newIndex()}
	sort.Slice(c.Sources, func(i, j int) bool { return c.Sources[i].ID < c.Sources[j].ID })
	byPath := map[string]string{}
	seen := map[string]bool{}
	for _, s := range c.Sources {
		if !idPattern.MatchString(s.ID) || seen[s.ID] || !publicPath(s.Path) || path.Clean(s.Path) != s.Path || strings.ContainsAny(s.Path, "\\\x00") || !strings.HasSuffix(s.Path, ".md") {
			return nil, fmt.Errorf("unknown, duplicate or private source: %s", s.ID)
		}
		if _, ok := byPath[s.Path]; ok {
			return nil, errors.New("duplicate source path")
		}
		if (s.Rights != "MIT" && !(granted && s.Rights == "local/unverified")) || s.Attribution == "" {
			return nil, fmt.Errorf("unsupported rights: %s", s.ID)
		}
		if len(s.Text) > MaxDocumentBytes || c.Bytes+len(s.Text) > MaxCorpusBytes || !utf8.ValidString(s.Text) {
			return nil, errors.New("decoded corpus byte/encoding limit exceeded")
		}
		if Hash([]byte(s.Text)) != s.SHA256 {
			return nil, fmt.Errorf("source hash mismatch: %s", s.ID)
		}
		seen[s.ID] = true
		byPath[s.Path] = s.ID
		c.Bytes += len(s.Text)
	}
	for _, s := range c.Sources {
		d, err := parseDocument(s, byPath)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", s.ID, err)
		}
		c.Documents = append(c.Documents, d)
		c.index.add(d, s.Text)
	}
	// Resolve fragments only after every document has real, unique anchors.
	for _, d := range c.Documents {
		if err := c.checkFragments(d.Nodes); err != nil {
			return nil, err
		}
	}
	var identity strings.Builder
	for _, s := range c.Sources {
		fmt.Fprintf(&identity, "%s\x00%s\x00%s\x00%s\n", s.ID, s.Path, s.SHA256, s.Rights)
	}
	c.Digest = Hash([]byte(identity.String()))
	return c, nil
}
func (c *Corpus) checkFragments(nodes []Node) error {
	for _, n := range nodes {
		if strings.HasPrefix(n.URL, "doc:") {
			target := strings.TrimPrefix(n.URL, "doc:")
			parts := strings.SplitN(target, "#", 2)
			d, err := c.Document(parts[0])
			if err != nil {
				return err
			}
			if len(parts) == 2 {
				ok := false
				for _, h := range d.Headings {
					if h.Anchor == parts[1] {
						ok = true
					}
				}
				if !ok {
					return fmt.Errorf("unknown anchor: %s", n.URL)
				}
			}
		}
		if err := c.checkFragments(n.Children); err != nil {
			return err
		}
	}
	return nil
}
func (c *Corpus) Document(id string) (Document, error) {
	for _, d := range c.Documents {
		if d.ID == id {
			return d, nil
		}
	}
	return Document{}, errors.New("unknown document")
}
func (c *Corpus) Search(query, category string, page int) (Page, error) {
	p := Page{Query: query, Page: page, PageSize: PageSize, Results: []SearchResult{}}
	if len(query) > 256 || page < 1 || page > MaxDocuments || !utf8.ValidString(query) {
		return p, errors.New("query/page limit exceeded")
	}
	hits := c.index.Query(strings.TrimSpace(query), 0)
	for _, hit := range hits {
		if category != "" && hit.Category != category {
			continue
		}
		if p.Total >= (page-1)*PageSize && len(p.Results) < PageSize {
			p.Results = append(p.Results, hit)
		}
		p.Total++
	}
	return p, nil
}
