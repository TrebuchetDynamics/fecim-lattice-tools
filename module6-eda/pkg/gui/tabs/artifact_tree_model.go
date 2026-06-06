package tabs

import (
	"path"
	"sort"
	"strings"
)

// ArtifactTreeModel adapts generated EDA artifact paths into stable IDs for a
// Fyne widget.Tree. IDs are normalized full paths so duplicate basenames in
// different directories remain unique.
type ArtifactTreeModel struct {
	children map[string][]string
	branches map[string]bool
	labels   map[string]string
}

// NewArtifactTreeModel builds a deterministic tree model from artifact paths.
func NewArtifactTreeModel(paths []string) *ArtifactTreeModel {
	m := &ArtifactTreeModel{
		children: map[string][]string{"": {}},
		branches: map[string]bool{"": true},
		labels:   map[string]string{"": "Artifacts"},
	}

	childSets := map[string]map[string]struct{}{"": {}}
	ensureChildSet := func(id string) map[string]struct{} {
		if _, ok := childSets[id]; !ok {
			childSets[id] = map[string]struct{}{}
		}
		return childSets[id]
	}

	for _, raw := range paths {
		artifactID := normalizeArtifactTreeID(raw)
		if artifactID == "" {
			continue
		}

		segments := strings.Split(artifactID, "/")
		parent := ""
		for idx, segment := range segments {
			id := segment
			if parent != "" {
				id = parent + "/" + segment
			}

			ensureChildSet(parent)[id] = struct{}{}
			m.labels[id] = segment

			if idx < len(segments)-1 {
				m.branches[id] = true
				ensureChildSet(id)
			}
			parent = id
		}
	}

	for parent, set := range childSets {
		children := make([]string, 0, len(set))
		for child := range set {
			children = append(children, child)
		}
		sort.Strings(children)
		m.children[parent] = children
	}

	return m
}

// ChildUIDs returns deterministic child IDs for the given tree node ID.
func (m *ArtifactTreeModel) ChildUIDs(uid string) []string {
	children := m.children[m.ID(uid)]
	out := make([]string, len(children))
	copy(out, children)
	return out
}

// IsBranch reports whether uid is a branch node.
func (m *ArtifactTreeModel) IsBranch(uid string) bool {
	return m.branches[m.ID(uid)]
}

// Label returns the display label for uid.
func (m *ArtifactTreeModel) Label(uid string) string {
	id := m.ID(uid)
	if label, ok := m.labels[id]; ok {
		return label
	}
	return path.Base(id)
}

// ID normalizes raw path input into the stable tree ID used by this model.
func (m *ArtifactTreeModel) ID(raw string) string {
	return normalizeArtifactTreeID(raw)
}

func normalizeArtifactTreeID(raw string) string {
	clean := path.Clean(strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/"))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	if clean == "." {
		return ""
	}
	return clean
}
