package docs

// EvidenceBinding identifies exactly what a recorded check exercised. Reading
// a report never runs the check or independently verifies an imported claim.
type EvidenceBinding struct {
	Revision      string `json:"revision"`
	SourceSHA256  string `json:"source_sha256"`
	FixtureSHA256 string `json:"fixture_sha256"`
	Tool          string `json:"tool"`
}
type EvidenceReport struct {
	Binding  EvidenceBinding `json:"binding"`
	Imported bool            `json:"imported"`
	Summary  string          `json:"summary"`
}

func EvidenceStatus(current EvidenceBinding, report *EvidenceReport) string {
	if report == nil {
		return "not run/unverified"
	}
	if report.Imported {
		return "imported/unverified"
	}
	if report.Binding != current || current.Revision == "" || len(current.SourceSHA256) != 64 || len(current.FixtureSHA256) != 64 || current.Tool == "" || report.Summary == "" {
		return "stale"
	}
	return "candidate-bound recorded evidence"
}
