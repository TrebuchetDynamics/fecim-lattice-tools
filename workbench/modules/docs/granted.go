package docs

import "errors"

// AddGranted receives bytes only from the native capability adapter. The
// read-only Worker exposes no operation that calls it. Nothing is published.
func (p *Port) AddGranted(id string, text []byte) error {
	if p.closed || p.corpus == nil {
		return errors.New("docs session closed")
	}
	if !idPattern.MatchString(id) || id == "" {
		return errors.New("choose a stable lowercase document ID")
	}
	if _, err := p.corpus.Document(id); err == nil {
		return errors.New("document ID already exists")
	}
	source := Source{ID: id, Path: "docs/guides/granted/" + id + ".md", Text: string(text), SHA256: Hash(text), Category: "Local (imported/unverified)", Rights: "local/unverified", Attribution: "Explicitly granted local document; rights and scientific claims unverified"}
	sources := append(append([]Source{}, p.corpus.Sources...), source)
	corpus, err := build(sources, true)
	if err != nil {
		return err
	}
	p.corpus = corpus
	p.selected = id
	return nil
}
func (p *Port) RestoreHistory(h History) error {
	if p.closed || p.corpus == nil {
		return errors.New("docs session closed")
	}
	if h.Favorites == nil || h.Recent == nil {
		return errors.New("history requires favorites and recent arrays")
	}
	if len(h.Favorites) > 64 || len(h.Recent) > 20 {
		return errors.New("history entry limit exceeded")
	}
	for _, list := range [][]string{h.Favorites, h.Recent} {
		seen := map[string]bool{}
		for _, id := range list {
			if _, err := p.corpus.Document(id); err != nil {
				return errors.New("history contains unknown or unavailable document")
			}
			if seen[id] {
				return errors.New("duplicate history entry")
			}
			seen[id] = true
		}
	}
	p.history = h
	return nil
}
