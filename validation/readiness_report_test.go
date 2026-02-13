package validation

import (
	"encoding/json"
	"testing"
)

func TestBuildExecutiveReadinessReport_ValidScoresAndTasks(t *testing.T) {
	report := BuildExecutiveReadinessReport(ReadinessInputs{
		TestPassRate:          0.9,
		CalibrationStatus:     0.8,
		ExportQuality:         0.85,
		DocumentationCoverage: 0.75,
	})

	if report.EducationScore < 0 || report.EducationScore > 100 {
		t.Fatalf("education score out of range: %d", report.EducationScore)
	}
	if report.ResearchScore < 0 || report.ResearchScore > 100 {
		t.Fatalf("research score out of range: %d", report.ResearchScore)
	}
	if report.DesignScore < 0 || report.DesignScore > 100 {
		t.Fatalf("design score out of range: %d", report.DesignScore)
	}
	if len(report.NextTasks) != 5 {
		t.Fatalf("expected 5 next tasks, got %d", len(report.NextTasks))
	}
}

func TestReadinessReportJSON_ProducesValidJSON(t *testing.T) {
	payload, err := ReadinessReportJSON(ReadinessInputs{
		TestPassRate:          1,
		CalibrationStatus:     1,
		ExportQuality:         1,
		DocumentationCoverage: 1,
	})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var report ReadinessReport
	if err := json.Unmarshal(payload, &report); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if report.EducationScore != 100 || report.ResearchScore != 100 || report.DesignScore != 100 {
		t.Fatalf("expected perfect scores, got %+v", report)
	}
}
