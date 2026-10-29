package moduleapi

import (
	"fecim-lattice-tools/workbench/hysteresis"
	"fecim-lattice-tools/workbench/study"
)

// Research envelopes are shared by the production Worker and native parity
// runner. Documents are bounded selected bytes, never filesystem authority.
type ResearchRequest struct {
	Identity  hysteresis.Identity `json:"identity"`
	Session   string              `json:"session"`
	Request   uint64              `json:"request"`
	Revision  uint64              `json:"revision"`
	Operation string              `json:"operation"`
	Input     ResearchInput       `json:"input"`
}
type ResearchInput struct {
	Documents          map[string][]byte  `json:"documents,omitempty"`
	Query              study.ResultsQuery `json:"query"`
	Compare            []string           `json:"compare,omitempty"`
	RepositoryRevision string             `json:"repository_revision,omitempty"`
	RunID              string             `json:"run_id,omitempty"`
}
type ResearchConfiguration struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	Revision string               `json:"revision"`
	Fields   []ConfigurationField `json:"fields"`
	Evidence string               `json:"evidence"`
}
type ResearchResult struct {
	Plot          *study.ResultsPlot       `json:"plot,omitempty"`
	Configuration *ResearchConfiguration   `json:"configuration,omitempty"`
	Job           Job                      `json:"job"`
	Page          *study.ResultsPage       `json:"page,omitempty"`
	Comparison    *study.ResultsComparison `json:"comparison,omitempty"`
	Counts        map[string]int           `json:"counts,omitempty"`
	Points        int                      `json:"points"`
	ResultsJSON   string                   `json:"results_json,omitempty"`
	RunJSON       string                   `json:"run_json,omitempty"`
}
type ResearchReply struct {
	Identity  hysteresis.Identity `json:"identity"`
	Session   string              `json:"session"`
	Request   uint64              `json:"request"`
	Revision  uint64              `json:"revision"`
	Operation string              `json:"operation"`
	Committed bool                `json:"committed"`
	Error     string              `json:"error,omitempty"`
	Result    ResearchResult      `json:"result"`
}
