package tabs

import "strings"

func edaStatusPending(label string) string         { return edaStatusJoin("○", label, "") }
func edaStatusRunning(label string) string         { return edaStatusJoin("…", label, "") }
func edaStatusSkipped(label, detail string) string { return edaStatusJoin("○", label, detail) }
func edaStatusSuccess(label, detail string) string { return edaStatusJoin("✓", label, detail) }
func edaStatusWarning(label, detail string) string { return edaStatusJoin("⚠", label, detail) }
func edaStatusFailure(label, detail string) string { return edaStatusJoin("✗", label, detail) }

func edaStatusJoin(symbol, label, detail string) string {
	label = strings.TrimSpace(label)
	detail = strings.TrimSpace(detail)
	if label == "" {
		label = "Status"
	}
	if detail == "" {
		return symbol + " " + label
	}
	return symbol + " " + label + ": " + detail
}
