package docs

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var anchorChars = regexp.MustCompile(`[^a-z0-9-]`)

// Anchor preserves native spelling, with deterministic duplicate suffixes.
func Anchor(s string) string {
	a := anchorChars.ReplaceAllString(strings.ReplaceAll(strings.ToLower(s), " ", "-"), "")
	if a == "" {
		a = "section"
	}
	return a
}
func parseDocument(s Source, paths map[string]string) (Document, error) {
	d := Document{ID: s.ID, Path: s.Path, Title: s.ID, Category: s.Category, SHA256: s.SHA256, Rights: s.Rights, Attribution: s.Attribution, Bytes: len(s.Text), Terms: DetectGlossaryTerms(s.Text), Headings: []Heading{}, Nodes: []Node{}}
	source := []byte(s.Text)
	root := goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough)).Parser().Parse(text.NewReader(source))
	seen := map[string]int{}
	used := map[string]bool{}
	var convert func(ast.Node, int) (Node, error)
	convert = func(a ast.Node, depth int) (Node, error) {
		if depth > 64 {
			return Node{}, errors.New("Markdown nesting limit exceeded")
		}
		n := Node{}
		switch v := a.(type) {
		case *ast.Document:
			n.Kind = "document"
		case *ast.Heading:
			n.Kind = "heading"
			n.Level = v.Level
			label := string(v.Text(source))
			base := Anchor(label)
			count := seen[base]
			n.Anchor = base
			if count > 0 {
				n.Anchor = fmt.Sprintf("%s-%d", base, count)
			}
			for used[n.Anchor] {
				count++
				n.Anchor = fmt.Sprintf("%s-%d", base, count)
			}
			seen[base] = count + 1
			used[n.Anchor] = true
			d.Headings = append(d.Headings, Heading{label, v.Level, n.Anchor})
			if len(d.Headings) == 1 {
				d.Title = label
			}
		case *ast.Paragraph, *ast.TextBlock:
			n.Kind = "paragraph"
		case *ast.Text:
			n.Kind = "text"
			n.Text = string(v.Segment.Value(source))
			if v.SoftLineBreak() || v.HardLineBreak() {
				n.Text += "\n"
			}
		case *ast.String:
			n.Kind = "text"
			n.Text = string(v.Value)
		case *ast.CodeSpan:
			n.Kind = "code"
		case *ast.FencedCodeBlock, *ast.CodeBlock:
			n.Kind = "pre"
			var block strings.Builder
			for i := 0; i < a.Lines().Len(); i++ {
				line := a.Lines().At(i)
				block.Write(line.Value(source))
			}
			n.Text = block.String()
		case *ast.Emphasis:
			if v.Level == 2 {
				n.Kind = "strong"
			} else {
				n.Kind = "em"
			}
		case *ast.List:
			n.Kind = "list"
			n.Ordered = v.IsOrdered()
		case *ast.ListItem:
			n.Kind = "item"
		case *ast.Blockquote:
			n.Kind = "quote"
		case *ast.ThematicBreak:
			n.Kind = "rule"
		case *ast.Link:
			n.Kind = "link"
			u, err := safeReference(s, string(v.Destination), paths, false)
			if err != nil {
				return n, err
			}
			n.URL = u
			if u == "" {
				n.Kind = "span"
			}
		case *ast.AutoLink:
			n.Kind = "link"
			u, err := safeReference(s, string(v.URL(source)), paths, false)
			if err != nil {
				return n, err
			}
			n.URL = u
			n.Text = string(v.Label(source))
		case *ast.Image:
			// This corpus intentionally publishes no images. Explicit omissions retain
			// alt text without an automatic fetch, including for local imports.
			_, err := safeReference(s, string(v.Destination), paths, true)
			if err != nil {
				return n, err
			}
			n.Kind = "span"
		case *ast.RawHTML, *ast.HTMLBlock:
			return n, errors.New("raw HTML is not publishable")
		case *extast.Table:
			n.Kind = "table"
		case *extast.TableHeader:
			n.Kind = "table-head"
		case *extast.TableRow:
			n.Kind = "row"
		case *extast.TableCell:
			n.Kind = "cell"
		case *extast.Strikethrough:
			n.Kind = "del"
		default:
			return n, fmt.Errorf("unsupported Markdown node %s", a.Kind().String())
		}
		for child := a.FirstChild(); child != nil; child = child.NextSibling() {
			out, err := convert(child, depth+1)
			if err != nil {
				return n, err
			}
			n.Children = append(n.Children, out)
		}
		return n, nil
	}
	parsed, err := convert(root, 0)
	if err != nil {
		return d, err
	}
	d.Nodes = parsed.Children
	return d, nil
}
func safeReference(s Source, raw string, paths map[string]string, image bool) (string, error) {
	if reason, ok := s.Omit[raw]; ok && reason != "" {
		return "", nil
	}
	decoded, err := url.PathUnescape(raw)
	if err != nil || strings.ContainsAny(decoded, "\x00\r\n\t\\") {
		return "", errors.New("unsafe URL encoding")
	}
	u, err := url.Parse(decoded)
	if err != nil {
		return "", errors.New("invalid reference")
	}
	if image {
		return "", errors.New("image requires explicit rights/omission")
	}
	if u.Scheme != "" {
		if (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.User != nil {
			return "", errors.New("unsafe external reference")
		}
		return u.String(), nil
	}
	if u.Host != "" || strings.HasPrefix(u.Path, "/") || u.RawQuery != "" {
		return "", errors.New("unsafe internal reference")
	}
	target := s.Path
	if u.Path != "" {
		target = path.Clean(path.Join(path.Dir(s.Path), u.Path))
	}
	id, ok := paths[target]
	if !ok {
		return "", fmt.Errorf("unknown publication reference: %s", raw)
	}
	result := "doc:" + id
	if u.Fragment != "" {
		result += "#" + u.Fragment
	}
	return result, nil
}
