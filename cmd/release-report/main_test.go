package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSplitCSV(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{",,,", []string{}},
	}
	for _, tc := range cases {
		got := splitCSV(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("splitCSV(%q) = %v (len %d), want %v (len %d)", tc.input, got, len(got), tc.want, len(tc.want))
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("splitCSV(%q)[%d] = %q, want %q", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestDifference(t *testing.T) {
	cases := []struct {
		expected []string
		observed []string
		want     []string
	}{
		{[]string{"a", "b", "c"}, []string{"a", "c"}, []string{"b"}},
		{[]string{"a", "b"}, []string{"a", "b"}, []string{}},
		{[]string{"a", "b"}, []string{}, []string{"a", "b"}},
	}
	for _, tc := range cases {
		got := difference(tc.expected, tc.observed)
		if len(got) != len(tc.want) {
			t.Errorf("difference(%v, %v) = %v, want %v", tc.expected, tc.observed, got, tc.want)
		}
	}
}

func TestRenderMarkdownContainsSections(t *testing.T) {
	r := releaseReport{
		InputDir: "/test",
		DOE: doeCoverage{
			ExpectedMaterials: 1,
			ExpectedModels:    1,
			ExpectedCombos:    1,
			FoundCombos:       1,
			CoverageFrac:      1.0,
		},
		PerMaterial: map[string]perMaterialReport{},
	}
	md := renderMarkdown(r, []string{"hzo"}, []string{"preisach"})
	if !strings.Contains(md, "DOE Coverage") {
		t.Error("markdown missing DOE Coverage section")
	}
	if !strings.Contains(md, "100.0%") {
		t.Error("markdown missing 100% coverage line")
	}
}

func TestRunReportsOutputDirectoryError(t *testing.T) {
	tmp := t.TempDir()
	blockedDir := filepath.Join(tmp, "not-a-directory")
	if err := os.WriteFile(blockedDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write placeholder: %v", err)
	}
	var stderr bytes.Buffer

	code := runReleaseReport([]string{
		"-in", filepath.Join(tmp, "missing-input-ok"),
		"-out-json", filepath.Join(blockedDir, "report.json"),
		"-out-md", filepath.Join(tmp, "report.md"),
	}, &stderr)

	if code != 1 {
		t.Fatalf("exit code=%d, want 1; stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "prepare JSON output directory") {
		t.Fatalf("stderr=%q, want JSON output directory context", stderr.String())
	}
	if strings.Contains(stderr.String(), "panic") {
		t.Fatalf("stderr=%q, must not include panic output", stderr.String())
	}
}
