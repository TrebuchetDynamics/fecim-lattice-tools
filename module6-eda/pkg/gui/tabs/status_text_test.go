package tabs

import (
	"os"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/module6-eda/pkg/config"
)

func TestBuilderValidationStatusLabelsCarryNonColorStatePrefix(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp: %v", err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	cfg := &config.ArrayConfig{
		Rows:         4,
		Cols:         4,
		Mode:         "storage",
		Architecture: "passive",
		Technology:   "sky130",
		CellWidth:    0.46,
		CellHeight:   2.72,
	}
	root := MakeBuilderValidationTab(cfg, nil)

	if exact := findLabelWithText(root, "Not generated"); exact != nil {
		t.Fatal("image status labels must include a non-color state prefix, got bare Not generated")
	}
	if exact := findLabelWithText(root, "Not validated"); exact != nil {
		t.Fatal("validation status labels must include a non-color state prefix, got bare Not validated")
	}
	if prefixed := countLabelsWithText(root, edaStatusPending("Not generated")); prefixed != 3 {
		t.Fatalf("expected 3 prefixed image status labels, got %d", prefixed)
	}
	if prefixed := countLabelsWithText(root, edaStatusPending("Not validated")); prefixed != 4 {
		t.Fatalf("expected 4 prefixed validation status labels, got %d", prefixed)
	}
}

func TestEDAStatusTextCarriesNonColorStatePrefix(t *testing.T) {
	cases := []struct {
		name string
		got  string
		want string
	}{
		{"pending", edaStatusPending("Not generated"), "○ Not generated"},
		{"running", edaStatusRunning("Generating"), "… Generating"},
		{"success", edaStatusSuccess("Generated", "fecim.png"), "✓ Generated: fecim.png"},
		{"warning", edaStatusWarning("DOT only", "install graphviz"), "⚠ DOT only: install graphviz"},
		{"failure", edaStatusFailure("Failed", "missing Docker"), "✗ Failed: missing Docker"},
		{"skipped", edaStatusSkipped("Skipped", "headless"), "○ Skipped: headless"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("status text = %q, want %q", tc.got, tc.want)
			}
			if !strings.ContainsAny(tc.got[:3], "○…✓⚠✗") {
				t.Fatalf("status text %q lacks non-color state prefix", tc.got)
			}
		})
	}
}

func findLabelWithText(root fyne.CanvasObject, text string) *widget.Label {
	var found *widget.Label
	walkObjects(root, func(obj fyne.CanvasObject) {
		if found != nil {
			return
		}
		if label, ok := obj.(*widget.Label); ok && label.Text == text {
			found = label
		}
	})
	return found
}

func countLabelsWithText(root fyne.CanvasObject, text string) int {
	count := 0
	walkObjects(root, func(obj fyne.CanvasObject) {
		if label, ok := obj.(*widget.Label); ok && label.Text == text {
			count++
		}
	})
	return count
}
