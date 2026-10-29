package docs

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"fecim-lattice-tools/workbench/moduleapi"
)

type History struct {
	Favorites []string `json:"favorites"`
	Recent    []string `json:"recent"`
}
type Output struct {
	Document       *Document       `json:"document,omitempty"`
	Page           Page            `json:"page"`
	DocumentCount  int             `json:"document_count"`
	DecodedBytes   int             `json:"decoded_bytes"`
	CorpusSHA256   string          `json:"corpus_sha256"`
	History        History         `json:"history"`
	EvidenceStatus string          `json:"evidence_status"`
	Report         *EvidenceReport `json:"report,omitempty"`
}
type Port struct {
	corpus                    *Corpus
	revision                  uint64
	query, category, selected string
	page                      int
	history                   History
	report                    *EvidenceReport
	closed                    bool
}

func Descriptor() moduleapi.ModuleDescriptor {
	return moduleapi.ModuleDescriptor{ID: "docs", Title: "Documentation", Available: true, SchemaVersion: moduleapi.SchemaVersion, ModelVersion: "docs-reader/1", Features: []string{"read", "search", "glossary", "citations", "toc", "favorites", "validation-evidence"}, Capabilities: []moduleapi.Capability{{Profile: "hosted", Available: true}, {Profile: "local", Available: true}}}
}
func NewPort() (*Port, error) {
	c, err := Public()
	if err != nil {
		return nil, err
	}
	return &Port{corpus: c, selected: "curriculum", page: 1, history: History{Favorites: []string{}, Recent: []string{}}}, nil
}
func (p *Port) Descriptor() moduleapi.ModuleDescriptor { return Descriptor() }
func (p *Port) Output() (Output, error) {
	if p.closed || p.corpus == nil {
		return Output{}, errors.New("docs session closed")
	}
	page, err := p.corpus.Search(p.query, p.category, p.page)
	if err != nil {
		return Output{}, err
	}
	out := Output{Page: page, DocumentCount: len(p.corpus.Documents), DecodedBytes: p.corpus.Bytes, CorpusSHA256: p.corpus.Digest, History: p.history, EvidenceStatus: EvidenceStatus(EvidenceBinding{}, p.report), Report: p.report}
	if p.selected != "" {
		d, err := p.corpus.Document(p.selected)
		if err != nil {
			return out, err
		}
		out.Document = &d
	}
	return out, nil
}
func (p *Port) Snapshot(ctx context.Context) (moduleapi.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if p.closed {
		return moduleapi.Snapshot{}, errors.New("docs session closed")
	}
	out, err := p.Output()
	if err != nil {
		return moduleapi.Snapshot{}, err
	}
	data, err := json.Marshal(out)
	if err != nil {
		return moduleapi.Snapshot{}, err
	}
	return moduleapi.Snapshot{Revision: p.revision, Descriptor: Descriptor(), Status: "ready", Evidence: []string{out.EvidenceStatus, "educational documentation; not measured device data"}, Fields: []moduleapi.Field{{ID: "docs", Label: "Go documentation output", Input: string(data)}}, Actions: []moduleapi.Action{{ID: "search", Label: "Search", Available: true}, {ID: "open", Label: "Read", Available: true}, {ID: "favorite", Label: "Toggle favorite", Available: true}, {ID: "evidence", Label: "Read imported evidence", Available: true}}, Tables: []moduleapi.TablePage{}, Plots: []moduleapi.Plot{}}, nil
}
func (p *Port) Apply(ctx context.Context, c moduleapi.Command) (moduleapi.Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return moduleapi.Snapshot{}, err
	}
	if p.closed || c.Revision != p.revision {
		return moduleapi.Snapshot{}, errors.New("closed/stale docs session")
	}
	if len(c.Resources) > 0 || len(c.Payload) > 0 {
		return moduleapi.Snapshot{}, errors.New("read-only port has no resource capability")
	}
	for k, v := range c.Input {
		if len(v) > 65536 {
			return moduleapi.Snapshot{}, errors.New("input limit exceeded")
		}
		if k != "query" && k != "category" && k != "page" && k != "id" && k != "report" {
			return moduleapi.Snapshot{}, errors.New("unknown docs input")
		}
	}
	switch c.Action {
	case "search":
		page := 1
		var err error
		if c.Input["page"] != "" {
			page, err = strconv.Atoi(c.Input["page"])
			if err != nil {
				return moduleapi.Snapshot{}, err
			}
		}
		if _, err = p.corpus.Search(c.Input["query"], c.Input["category"], page); err != nil {
			return moduleapi.Snapshot{}, err
		}
		p.query = c.Input["query"]
		p.category = c.Input["category"]
		p.page = page
	case "open":
		id := ResolveAlias(c.Input["id"])
		if _, err := p.corpus.Document(id); err != nil {
			return moduleapi.Snapshot{}, err
		}
		p.selected = id
		p.history.Recent = frontUnique(p.history.Recent, id, 20)
	case "favorite":
		id := c.Input["id"]
		if _, err := p.corpus.Document(id); err != nil {
			return moduleapi.Snapshot{}, err
		}
		found := -1
		for i, x := range p.history.Favorites {
			if x == id {
				found = i
			}
		}
		if found >= 0 {
			p.history.Favorites = append(p.history.Favorites[:found], p.history.Favorites[found+1:]...)
		} else {
			if len(p.history.Favorites) >= 64 {
				return moduleapi.Snapshot{}, errors.New("favorite limit reached")
			}
			p.history.Favorites = append(p.history.Favorites, id)
		}
	case "evidence":
		var report EvidenceReport
		if err := decodeStrict([]byte(c.Input["report"]), &report); err != nil {
			return moduleapi.Snapshot{}, err
		}
		report.Imported = true
		p.report = &report
	default:
		return moduleapi.Snapshot{}, errors.New("unknown docs action")
	}
	p.revision++
	return p.Snapshot(ctx)
}
func frontUnique(list []string, id string, limit int) []string {
	out := []string{id}
	for _, x := range list {
		if x != id && len(out) < limit {
			out = append(out, x)
		}
	}
	return out
}
func (p *Port) Close() { p.closed = true; p.corpus = nil; p.report = nil }

// Explicit aliases replace old quick links; unknown IDs are never guessed.
func ResolveAlias(id string) string {
	switch id {
	case "overview", "docs/documentation", "docs/2-learn/ELI5.md":
		return "curriculum"
	case "docs/GLOSSARY.md":
		return "glossary"
	case "docs/4-research/honesty-audit.md":
		return "honesty"
	}
	return id
}
