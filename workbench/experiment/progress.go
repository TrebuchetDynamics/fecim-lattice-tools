package experiment

// ProgressSnapshot reports counts for the expanded candidates of one Run.
// Completed includes reused and failed committed points; Reused and Failed are
// overlapping subsets of Completed. Cancelled counts unresolved points only
// after workers settle on context cancellation, not on a systemic error.
// Snapshots contain no scientific results or provenance.
type ProgressSnapshot struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Reused    int `json:"reused"`
	Failed    int `json:"failed"`
	Cancelled int `json:"cancelled"`
}

func (p ProgressSnapshot) send(ch chan<- ProgressSnapshot) {
	select {
	case ch <- p:
	default:
	}
}
