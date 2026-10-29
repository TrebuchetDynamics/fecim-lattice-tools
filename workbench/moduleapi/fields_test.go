package moduleapi

import (
	"fecim-lattice-tools/workbench/project"
	"strings"
	"testing"
)

func TestConfigurationFieldPresentation(t *testing.T) {
	field := FieldFromConfiguration(project.ConfigurationField{Document: "design.yaml", Path: "/device/g_min_s", Kind: "float", Value: "1e-06"}, "S")
	if field.Value != "1e-06" || field.Label != "Minimum conductance" || field.Unit != "S" || !strings.Contains(field.Presentation.Speakable, "siemens") || !strings.Contains(field.Presentation.Speakable, "times 10") {
		t.Fatalf("canonical editor input/Go speech: %+v", field)
	}
	seed := FieldFromConfiguration(project.ConfigurationField{Document: "sweep.yaml", Path: "/seed", Kind: "int64", Value: "9223372036854775807"}, "")
	if seed.Presentation.Speakable != "9223372036854775807" || seed.Value != "9223372036854775807" {
		t.Fatalf("integer precision lost: %+v", seed)
	}
}
