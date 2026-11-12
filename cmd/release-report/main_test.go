package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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

func TestRunReleaseReportAggregation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		passes []bool
		suites []string
		want   bool
	}{
		{"single pass", []bool{true}, []string{"suite"}, true},
		{"single fail", []bool{false}, []string{"suite"}, false},
		{"duplicates pass", []bool{true, true}, []string{"suite", "suite"}, true},
		{"duplicates fail", []bool{false, false}, []string{"suite", "suite"}, false},
		{"mixed same suite", []bool{false, true}, []string{"suite", "suite"}, false},
		{"mixed distinct suites", []bool{false, true}, []string{"z-suite", "a-suite"}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			in := filepath.Join(tmp, "input")
			if err := os.MkdirAll(filepath.Join(in, "nested"), 0o755); err != nil {
				t.Fatal(err)
			}
			writeArtifact := func(path, material, model, suite string, pass bool) {
				t.Helper()
				b, err := json.Marshal(regressionSummary{
					Suite: suite, Model: model, AllPass: pass,
					MaterialSnapshot: materialPhysicsSnapshot{MaterialID: material},
				})
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			paths := []string{filepath.Join(in, "a.json")}
			if len(tc.passes) == 2 {
				paths = append(paths, filepath.Join(in, "nested", "z.json"))
			}
			for i, path := range paths {
				writeArtifact(path, "mixed", "preisach", tc.suites[i], tc.passes[i])
			}
			writeArtifact(filepath.Join(in, "other-model.json"), "mixed", "landau-khalatnikov", "other", true)
			writeArtifact(filepath.Join(in, "passing-p.json"), "passing", "preisach", "other", true)
			writeArtifact(filepath.Join(in, "passing-lk.json"), "passing", "landau-khalatnikov", "other", true)
			writeArtifact(filepath.Join(in, "missing.json"), "missing", "preisach", "other", true)

			for order := 0; order < 2; order++ {
				outJSON, outMD := filepath.Join(tmp, "report.json"), filepath.Join(tmp, "report.md")
				var stderr bytes.Buffer
				if code := runReleaseReport([]string{
					"-in", in, "-out-json", outJSON, "-out-md", outMD,
					"-materials", "mixed,passing,missing", "-models", "preisach,landau-khalatnikov",
				}, &stderr); code != 0 {
					t.Fatalf("order %d: exit=%d stderr=%s", order, code, &stderr)
				}
				b, err := os.ReadFile(outJSON)
				if err != nil {
					t.Fatal(err)
				}
				var report releaseReport
				if err := json.Unmarshal(b, &report); err != nil {
					t.Fatal(err)
				}
				md, err := os.ReadFile(outMD)
				if err != nil {
					t.Fatal(err)
				}
				for material, want := range map[string]struct{ preisach, all bool }{
					"mixed": {tc.want, tc.want}, "passing": {true, true}, "missing": {true, false},
				} {
					pm := report.PerMaterial[material]
					wantModels := map[string]bool{"preisach": want.preisach}
					if material != "missing" {
						wantModels["landau-khalatnikov"] = true
					} else if !reflect.DeepEqual(pm.MissingModels, []string{"landau-khalatnikov"}) {
						t.Errorf("missing models=%v", pm.MissingModels)
					}
					if !reflect.DeepEqual(pm.ByModel, wantModels) || pm.AllModelsPass != want.all {
						t.Errorf("order %d %s: by_model=%v all_models_pass=%v, want %v / %v", order, material, pm.ByModel, pm.AllModelsPass, wantModels, want.all)
					}
					section := strings.Split(strings.Split(string(md), "### "+material+"\n\n")[1], "\n### ")[0]
					for _, line := range []string{
						fmt.Sprintf("- All expected models pass: %v\n", want.all),
						fmt.Sprintf("- preisach: %v\n", want.preisach),
					} {
						if !strings.Contains(section, line) {
							t.Errorf("order %d %s: markdown missing %q: %s", order, material, line, section)
						}
					}
				}
				wantFiles := append(append([]string{}, paths...), filepath.Join(in, "other-model.json"))
				wantSuites := append([]string{}, tc.suites...)
				if order == 1 && len(wantSuites) == 2 {
					wantSuites[0], wantSuites[1] = wantSuites[1], wantSuites[0]
				}
				wantSuites = append(wantSuites, "other")
				pm := report.PerMaterial["mixed"]
				if !reflect.DeepEqual(pm.ObservedFiles, wantFiles) || !reflect.DeepEqual(pm.ObservedSuites, wantSuites) {
					t.Errorf("order %d: provenance=%v / %v, want %v / %v", order, pm.ObservedFiles, pm.ObservedSuites, wantFiles, wantSuites)
				}
				if report.FileCount != len(paths)+4 || report.DOE.FoundCombos != 5 || report.DOE.ExpectedCombos != 6 {
					t.Errorf("coverage/count policy changed: files=%d DOE=%+v", report.FileCount, report.DOE)
				}
				if order == 0 && len(paths) == 2 {
					// Swap filenames without changing either artifact's contents.
					spare := filepath.Join(tmp, "swap.json")
					for _, move := range [][2]string{{paths[0], spare}, {paths[1], paths[0]}, {spare, paths[1]}} {
						if err := os.Rename(move[0], move[1]); err != nil {
							t.Fatal(err)
						}
					}
				}
			}
		})
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
