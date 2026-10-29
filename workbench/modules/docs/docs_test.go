package docs

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func fixture(t *testing.T, texts map[string]string) *Corpus {
	t.Helper()
	sources := []Source{}
	for id, body := range texts {
		sources = append(sources, Source{ID: id, Path: "docs/guides/" + id + ".md", SHA256: Hash([]byte(body)), Rights: "MIT", Attribution: "FeCIM contributors", Category: "Guide", Text: body})
	}
	c, err := Build(sources)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestDocsWebWorkflow(t *testing.T) {
	c := fixture(t, map[string]string{"memory": "# Ferroelectric Memory\n\nFerroelectric materials exhibit spontaneous polarization.\n\n## Reading\n\n```go\n// # not a heading\n```\n\n## Reading\n", "crossbar": "# Crossbar Arrays\n\nCrossbar architecture for compute-in-memory.", "learning": "# Machine Learning\n\nNeural networks and ferroelectric devices."})
	for _, tc := range []struct {
		q string
		n int
	}{{"ferroelectric", 2}, {"crossbar", 1}, {"memory", 2}, {"neural", 1}, {"nonexistent", 0}, {"", 0}} {
		p, err := c.Search(tc.q, "", 1)
		if err != nil || p.Total != tc.n {
			t.Fatalf("%q: %+v %v", tc.q, p, err)
		}
	}
	d, err := c.Document("memory")
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Headings) != 3 || d.Headings[2].Anchor != "reading-1" {
		t.Fatalf("actual anchors: %+v", d.Headings)
	}
	if strings.Join(d.Terms, ",") != "FeCIM,Hysteresis Loop,Remnant Polarization" && !strings.Contains(strings.Join(d.Terms, ","), "Remnant Polarization") {
		t.Fatalf("aliases: %v", d.Terms)
	}
	if err = ValidatePublication(c.Sources); err == nil {
		t.Fatal("unreviewed sources admitted as public corpus")
	}
	if _, err = c.Document("unknown"); err == nil {
		t.Fatal("unknown document accepted")
	}
	if _, err = c.Search(strings.Repeat("a", 257), "", 1); err == nil {
		t.Fatal("unbounded query")
	}
	p, _ := c.Search("ferroelectric", "unknown", 1)
	if p.Total != 0 {
		t.Fatal("category ignored")
	}
	a, _ := json.Marshal(c)
	pages := map[string]string{}
	for i := 0; i < 21; i++ {
		pages[fmt.Sprintf("page-%02d", i)] = "# Commonneedle\n\nA real bounded document."
	}
	paged := fixture(t, pages)
	first, _ := paged.Search("commonneedle", "", 1)
	second, _ := paged.Search("commonneedle", "", 2)
	if first.Total != 21 || len(first.Results) != 20 || len(second.Results) != 1 || second.Results[0].DocPath != "page-20" {
		t.Fatalf("paging: %+v %+v", first, second)
	}
	again, _ := Build(c.Sources)
	x, _ := json.Marshal(again)
	if string(x) != string(a) {
		t.Fatal("nondeterministic corpus")
	}
}

func TestDocsWebSafety(t *testing.T) {
	for _, text := range []string{"# Reading\n\n# Reading\n\n# Reading-1\n", "# Title\n\n## 标题\n\n## 标题\n"} {
		c := fixture(t, map[string]string{"anchors": text})
		seen := map[string]bool{}
		for _, h := range c.Documents[0].Headings {
			if seen[h.Anchor] {
				t.Fatalf("duplicate anchor %s", h.Anchor)
			}
			seen[h.Anchor] = true
		}
	}
	good := Source{ID: "safe", Path: "docs/guides/safe.md", Text: "# Safe\n\n[link](https://doi.org/10.1038/test)\n", Rights: "MIT", Attribution: "FeCIM contributors"}
	good.SHA256 = Hash([]byte(good.Text))
	for _, bad := range []Source{
		{ID: "private", Path: "docs/internals/private.md", Text: "# Private", Rights: "MIT", Attribution: "x"},
		{ID: "unknown", Path: "docs/guides/a.md", Text: "# Unknown", Rights: "unknown", Attribution: "x"},
		{ID: "hash", Path: "docs/guides/a.md", Text: "# Hash", Rights: "MIT", Attribution: "x", SHA256: "bad"},
	} {
		if _, err := Build([]Source{bad}); err == nil {
			t.Fatalf("accepted invalid source %s", bad.ID)
		}
	}
	for _, text := range []string{"<script>alert(1)</script>", "[bad](javascript:alert%281%29)", "![track](https://evil.invalid/pixel.png)", "[private](../internals/private.md)", "[bad](%2f%2fevil.invalid)", "[bad](missing.md)"} {
		s := good
		s.Text = text
		s.SHA256 = Hash([]byte(text))
		if _, err := Build([]Source{s}); err == nil {
			t.Fatalf("accepted unsafe reference %s", text)
		}
	}
	for _, n := range []int{MaxDocumentBytes + 1, MaxCorpusBytes + 1} {
		s := good
		s.Text = strings.Repeat("a", n)
		s.SHA256 = Hash([]byte(s.Text))
		if _, err := Build([]Source{s}); err == nil {
			t.Fatal("size admitted")
		}
	}
	many := []Source{}
	for i := 0; i < MaxDocuments+1; i++ {
		s := good
		s.ID = fmt.Sprintf("doc-%d", i)
		s.Path = "docs/guides/" + s.ID + ".md"
		many = append(many, s)
	}
	if _, err := Build(many); err == nil {
		t.Fatal("document count limit admitted")
	}
	total := []Source{}
	for i := 0; i < 11; i++ {
		s := good
		s.ID = fmt.Sprintf("large-%d", i)
		s.Path = "docs/guides/" + s.ID + ".md"
		s.Text = strings.Repeat("x", MaxDocumentBytes)
		s.SHA256 = Hash([]byte(s.Text))
		total = append(total, s)
	}
	if _, err := Build(total); err == nil {
		t.Fatal("total decoded corpus limit admitted")
	}
	c := fixture(t, map[string]string{"safe": "# Safe\n\n| Value | Meaning |\n| --- | --- |\n| 30 | Educational |\n"})
	if len(c.Documents[0].Nodes) == 0 {
		t.Fatal("empty AST")
	}
}

func TestDocsWebEvidence(t *testing.T) {
	current := EvidenceBinding{Revision: "candidate", SourceSHA256: strings.Repeat("a", 64), FixtureSHA256: strings.Repeat("b", 64), Tool: "go1.26.10"}
	for _, tc := range []struct {
		report *EvidenceReport
		status string
	}{
		{nil, "not run/unverified"},
		{&EvidenceReport{Binding: EvidenceBinding{Revision: "old"}}, "stale"},
		{&EvidenceReport{Binding: current, Imported: true, Summary: "Pass"}, "imported/unverified"},
		{&EvidenceReport{Binding: current, Summary: "TestDocsWebWorkflow recorded"}, "candidate-bound recorded evidence"},
	} {
		if got := EvidenceStatus(current, tc.report); got != tc.status {
			t.Fatalf("%+v: %s", tc.report, got)
		}
	}
}
