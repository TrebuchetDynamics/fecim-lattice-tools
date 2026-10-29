package docs

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
)

// PublicationPolicy is an explicit, reviewed source index, never a directory walk.
// Hash changes require review here; generate:docs fails closed on source drift.
func PublicationPolicy() []Source {
	return []Source{{ID: "curriculum", Path: "docs/modules/eli5-overview.md", SHA256: "ac610d1ec3d272785c8c15506fd92fb189c41ae7a3950073b53c6204579a47d0", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Curriculum", Omit: map[string]string{"../internals/status.md": "Internal status is not public corpus.", "../internals/api-reference.md": "Internal API documentation is not public corpus."}},
		{ID: "hysteresis", Path: "docs/modules/hysteresis/eli5.md", SHA256: "c4ca870e27ef915aeb651cd443c7967b1f40e31725a28ba322b01b95a2210fa9", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Learning", Omit: map[string]string{}},
		{ID: "crossbar", Path: "docs/modules/crossbar/eli5.md", SHA256: "f93cca6de07afe4ca69feb254bd14c6beb585dcaca904b3672fc712cf441e3b7", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Learning", Omit: map[string]string{}},
		{ID: "mnist", Path: "docs/modules/mnist/eli5.md", SHA256: "1b54dc8cd501f8b2a0477266cf2baa0950142a5da3fd85167f6910acf4119b21", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Learning", Omit: map[string]string{"physics.md": "Outside this curated selection; use repository documentation.", "features.md": "Outside this curated selection; use repository documentation."}},
		{ID: "circuits", Path: "docs/modules/circuits/eli5.md", SHA256: "96001792ffc5e9fde4429c37d1f4fbaa9f46e82e08611b16de667eea80a50f8e", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Learning", Omit: map[string]string{"physics.md": "Outside this curated selection; use repository documentation.", "features.md": "Outside this curated selection; use repository documentation."}},
		{ID: "comparison", Path: "docs/modules/comparison/eli5.md", SHA256: "8a2f6cba0288acfe9f34154fe3947183326101f18a6a6dd3b3377921d2f90dd1", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Learning", Omit: map[string]string{"physics.md": "Outside this curated selection; use repository documentation.", "features.md": "Outside this curated selection; use repository documentation."}},
		{ID: "eda", Path: "docs/modules/eda/eli5.md", SHA256: "79682b232581dca8ed66767d29e0766e0fce3c0d5d5343a8e30ede77e9676c2b", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Learning", Omit: map[string]string{"physics.md": "Outside this curated selection; use repository documentation.", "features.md": "Outside this curated selection; use repository documentation."}},
		{ID: "glossary", Path: "docs/guides/glossary.md", SHA256: "f3aea1ba3b78639d020c37e605d567e50951c2d0ec9a2f82f82de50003f0929c", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Glossary", Omit: map[string]string{}},
		{ID: "trust", Path: "docs/guides/trust.md", SHA256: "f3ce8d58270f1ba7702b872ac33f00e0d0f6354beafc7713a0c3978f274548fc", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Trust", Omit: map[string]string{}},
		{ID: "honesty", Path: "docs/research/honesty-audit.md", SHA256: "5b382ee0fd9d83a222e1b7ed490d78ab83cd33b07430084bebf94f5e6aa7124c", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Trust", Omit: map[string]string{}},
		{ID: "park-correction", Path: "docs/research/validation/park-reference-correction.md", SHA256: "af7924e4f2a63fabcf21c8728b106bbfdfe3ffaa0fe7b7aeef1985773d98a659", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Trust", Omit: map[string]string{}},
		{ID: "validation", Path: "docs/modules/doc-viewer/validation-and-evidence.md", SHA256: "14c75bc2e7091e2a8ffffa4e7ac700656079cbf2221986868354f2eae7c5f492", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Evidence", Omit: map[string]string{}},
		{ID: "citations", Path: "docs/modules/doc-viewer/citations.md", SHA256: "609fe3b7c072dd56c4272679a9ec39203579455f8edd296602c7b3f3506bb6c5", Rights: "MIT", Attribution: "FeCIM Lattice Tools contributors (repository-authored Markdown; publisher assets excluded)", Category: "Citation", Omit: map[string]string{}}}
}

//go:embed corpus.json
var publicContent embed.FS

func Public() (*Corpus, error) {
	data, err := publicContent.ReadFile("corpus.json")
	if err != nil {
		return nil, err
	}
	var sources []Source
	if err = json.Unmarshal(data, &sources); err != nil {
		return nil, err
	}
	if err := ValidatePublication(sources); err != nil {
		return nil, err
	}
	return Build(sources)
}

// ValidatePublication prevents an embedded or imported self-described source
// list from expanding the reviewed public allowlist or changing its rights.
func ValidatePublication(sources []Source) error {
	policy := PublicationPolicy()
	if len(sources) != len(policy) {
		return errors.New("public source count differs from publication policy")
	}
	expected := map[string][]byte{}
	for _, s := range policy {
		s.Text = ""
		b, _ := json.Marshal(s)
		expected[s.ID] = b
	}
	for _, s := range sources {
		s.Text = ""
		b, _ := json.Marshal(s)
		want, ok := expected[s.ID]
		if !ok || !bytes.Equal(b, want) {
			return errors.New("unknown public source, hash, rights or reference policy")
		}
		delete(expected, s.ID)
	}
	return nil
}
