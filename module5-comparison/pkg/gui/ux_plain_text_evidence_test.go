package gui

import (
	"strings"
	"testing"
)

func TestM5UX03_PlainTextEvidencePanelContent(t *testing.T) {
	run := BuildScenarioRun(ScenarioBaseline, PresetScenarioConfig(ScenarioBaseline))
	txt := PlainTextEvidence("Energy chart evidence", run)
	if !strings.Contains(txt, "Energy chart evidence") || strings.Contains(txt, "±") || !strings.Contains(txt, "not a statistical interval") || !strings.Contains(txt, "Not available") {
		t.Fatalf("plain-text evidence must withhold CO2 and statistical uncertainty: %s", txt)
	}
}
