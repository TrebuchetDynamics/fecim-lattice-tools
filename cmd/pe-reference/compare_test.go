package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/shared/physics"
	"fecim-lattice-tools/validation/peloop"
)

// Catches lost residuals/provenance, promotion of low error to validation, and
// publication before admission or over an existing caller-owned path.
func TestCompareCommandEvidenceBoundary(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "pe-reference")
	if out, err := exec.Command("go", "build", "-o", binary, ".").CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	dataset, err := filepath.Abs("../../validation/literature/data/experimental/kim2020_fig3a_ref_digitized")
	if err != nil {
		t.Fatal(err)
	}
	profile := filepath.Join(dataset, "source.provenance.json")
	data := filepath.Join(dataset, "loop.csv")
	read := func(path string) []byte {
		t.Helper()
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	hash := func(b []byte) string { sum := sha256.Sum256(b); return hex.EncodeToString(sum[:]) }
	invoke := func(profile, data, out string, extra ...string) error {
		t.Helper()
		args := append([]string{"compare", "--profile", profile, "--data", data, "--out-dir", out}, extra...)
		output, err := exec.Command(binary, args...).CombinedOutput()
		if err == nil && len(output) != 0 {
			t.Fatalf("unexpected output: %s", output)
		}
		if err != nil {
			return fmt.Errorf("%w: %s", err, output)
		}
		return nil
	}
	if runtime.GOOS != "linux" {
		dir := t.TempDir()
		before := snapshotComparisonPaths(t, dir)
		err := invoke(profile, data, filepath.Join(dir, "output"))
		if err == nil || !strings.Contains(err.Error(), "atomic no-replace") {
			t.Fatalf("unsupported platform must fail closed at publication: %v", err)
		}
		if !reflect.DeepEqual(before, snapshotComparisonPaths(t, dir)) {
			t.Fatal("unsupported platform left output or staging debris")
		}
		return
	}
	type evidence struct {
		ComparisonKind         string               `json:"comparison_kind"`
		ExperimentalValidation bool                 `json:"experimental_validation"`
		Source                 peloop.SourceProfile `json:"source"`
		Fingerprints           struct {
			Source      string            `json:"source_sha256"`
			Data        string            `json:"data_sha256"`
			Profile     string            `json:"profile_sha256"`
			Parameters  string            `json:"model_parameters_sha256"`
			Executable  string            `json:"executable_sha256"`
			ModelSource map[string]string `json:"model_source_sha256"`
		} `json:"fingerprints"`
		Git struct {
			Revision string `json:"revision"`
			Dirty    *bool  `json:"dirty"`
		} `json:"git"`
		Units            map[string]string     `json:"units"`
		ExactFieldPolicy string                `json:"exact_field_policy"`
		Limitations      []string              `json:"limitations"`
		Metrics          peloop.Metrics        `json:"metrics"`
		Baseline         peloop.BaselineResult `json:"baseline"`
		Reference        []peloop.Sample       `json:"reference"`
	}
	var measured evidence
	for _, mode := range []string{"measured", "zero residual"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			input := data
			if mode == "zero residual" {
				// Deliberately model-derived test input, never a committed measured fixture.
				var csv strings.Builder
				csv.WriteString("branch,point_index,field_MV_cm,polarization_uC_cm2\n")
				for _, s := range measured.Baseline.Samples {
					fmt.Fprintf(&csv, "%s,%d,%.17g,%.17g\n", s.Branch, s.PointIndex, s.FieldMVcm, s.PolarizationUCcm2)
				}
				input = filepath.Join(dir, "model-derived-test-only.csv")
				if err := os.WriteFile(input, []byte(csv.String()), 0600); err != nil {
					t.Fatal(err)
				}
			}
			var first map[string][]byte
			for i := 0; i < 2; i++ {
				out := filepath.Join(dir, fmt.Sprint("bundle-", i))
				if err := invoke(profile, input, out); err != nil {
					t.Fatalf("compare: %v", err)
				}
				files, err := os.ReadDir(out)
				if err != nil || len(files) != 3 {
					t.Fatalf("bundle files: %v %v", files, err)
				}
				bundle := map[string][]byte{}
				for _, name := range []string{"metrics.json", "overlay.svg", "caption.txt"} {
					bundle[name] = read(filepath.Join(out, name))
				}
				if i == 1 {
					if !reflect.DeepEqual(first, bundle) {
						t.Fatal("same inputs produced different bundle bytes")
					}
					continue
				}
				first = bundle
				var report evidence
				if err := json.Unmarshal(bundle["metrics.json"], &report); err != nil {
					t.Fatal(err)
				}
				if report.ComparisonKind != "unfitted-baseline" || report.ExperimentalValidation {
					t.Fatal("comparison promoted to validation")
				}
				if report.Fingerprints.Source != "2016943b0e199bf68e4e56783d7edb42324762b442c816cc66b8c6b19046ddfa" || report.Fingerprints.Profile != hash(read(profile)) || report.Fingerprints.Data != hash(read(input)) || report.Fingerprints.Executable != hash(read(binary)) {
					t.Fatal("incorrect input/source/executable fingerprints")
				}
				materialJSON, _ := json.Marshal(report.Baseline.Material)
				if report.Fingerprints.Parameters != hash(materialJSON) || !reflect.DeepEqual(report.Baseline.Material, *physics.DefaultHZO()) {
					t.Fatal("material snapshot/fingerprint changed")
				}
				if len(report.Fingerprints.ModelSource) == 0 {
					t.Fatal("missing source hashes")
				}
				for path, want := range report.Fingerprints.ModelSource {
					if hash(read(filepath.Join("../..", path))) != want {
						t.Fatalf("stale source hash: %s", path)
					}
				}
				revision, err := exec.Command("git", "rev-parse", "HEAD").Output()
				if err != nil || report.Git.Revision != strings.TrimSpace(string(revision)) || report.Git.Dirty == nil {
					t.Fatal("missing Git revision/dirty state")
				}
				status, err := exec.Command("git", "status", "--porcelain", "--untracked-files=all").Output()
				if err != nil || *report.Git.Dirty != (len(status) > 0) {
					t.Fatal("incorrect dirty state")
				}
				if report.Units["field"] != "MV/cm" || report.Units["polarization"] != "uC/cm2" || report.Units["rmse"] != "uC/cm2" || report.Units["normalized_rmse"] != "fraction" || !strings.Contains(report.ExactFieldPolicy, "no interpolation") {
					t.Fatal("missing units/exact-field policy")
				}
				settings := report.Baseline.Settings
				if settings.TemperatureK != 300 || settings.StressGPa != 1 || settings.ConditioningCycles != 2 || settings.StepsPerHalfCycle != 512 || !settings.Reset {
					t.Fatal("wrong baseline settings")
				}
				if report.Source.Conditions.MeasurementTemperature.Status != "not reported" || report.Source.Conditions.Waveform.Status != "not reported" {
					t.Fatal("unknown conditions lost")
				}
				caption := string(bundle["caption.txt"])
				for _, text := range []string{"unfitted-baseline", "experimental_validation=false", "quasi-static", "10 kHz", "not reported", "not measured Ps", "not a recovered measurement time sequence", "trace-fidelity", "not verified", "omitted", "+/-4", "not experimental instrument uncertainties"} {
					if !strings.Contains(caption, text) || !strings.Contains(strings.Join(report.Limitations, " "), text) && text != "unfitted-baseline" && text != "experimental_validation=false" {
						t.Fatalf("lost evidence limit %q", text)
					}
				}
				inputSamples, err := peloop.ReadSamples(bytes.NewReader(read(input)))
				if err != nil || !reflect.DeepEqual(inputSamples, report.Reference) {
					t.Fatal("report did not retain the supplied reference values/order")
				}
				if len(report.Reference) != 26 || len(report.Baseline.Samples) != 26 || report.Metrics.Pooled.Count != 26 || report.Metrics.PerBranch["ascending"].Count != 12 || report.Metrics.PerBranch["descending"].Count != 14 {
					t.Fatal("incorrect branch/pooled counts")
				}
				var sumSquared, sumAbs, maxP float64
				for j, ref := range report.Reference {
					sim := report.Baseline.Samples[j]
					if ref.Branch != sim.Branch || ref.PointIndex != sim.PointIndex || ref.FieldMVcm != sim.FieldMVcm {
						t.Fatal("field mismatch")
					}
					d := sim.PolarizationUCcm2 - ref.PolarizationUCcm2
					sumSquared += d * d
					sumAbs += math.Abs(d)
					maxP = math.Max(maxP, math.Abs(ref.PolarizationUCcm2))
				}
				m := report.Metrics.Pooled
				if math.Abs(m.RMSEUCcm2-math.Sqrt(sumSquared/26)) > 1e-12 || math.Abs(m.MAEUCcm2-sumAbs/26) > 1e-12 || m.ObservedMaxAbsReferencePUCcm2 != maxP || m.NormalizedRMSE == nil || math.Abs(*m.NormalizedRMSE-m.RMSEUCcm2/maxP) > 1e-12 {
					t.Fatal("residual/normalizer was altered")
				}
				if mode == "measured" {
					if m.RMSEUCcm2 <= 1 {
						t.Fatal("measured mismatch hidden")
					}
					measured = report
				} else if m.RMSEUCcm2 != 0 {
					t.Fatal("zero-residual seam did not exercise low error")
				}
				for name, metric := range report.Metrics.PerBranch {
					if !strings.Contains(caption, name) || !strings.Contains(caption, strconv.FormatFloat(metric.RMSEUCcm2, 'g', -1, 64)) {
						t.Fatal("caption lost branch metrics")
					}
				}
				var svg struct {
					Metadata string `xml:"metadata"`
					Desc     string `xml:"desc"`
				}
				if err := xml.Unmarshal(bundle["overlay.svg"], &svg); err != nil {
					t.Fatal(err)
				}
				if svg.Desc != caption || !bytes.Equal([]byte(svg.Metadata), bundle["metrics.json"]) {
					t.Fatal("SVG/JSON/caption evidence diverged")
				}
				decoder := xml.NewDecoder(bytes.NewReader(bundle["overlay.svg"]))
				markers, lines := 0, 0
				for {
					token, err := decoder.Token()
					if err == io.EOF {
						break
					}
					if err != nil {
						t.Fatal(err)
					}
					if e, ok := token.(xml.StartElement); ok {
						if e.Name.Local == "circle" {
							markers++
						}
						if e.Name.Local == "polyline" {
							lines++
						}
					}
				}
				if markers != 26 || lines != 2 {
					t.Fatalf("measured markers=%d simulated branch lines=%d", markers, lines)
				}
			}
		})
	}
	t.Run("destination created after preflight", func(t *testing.T) {
		dir := t.TempDir()
		staging, output := filepath.Join(dir, "staging"), filepath.Join(dir, "output")
		for _, path := range []string{staging, output} {
			if err := os.Mkdir(path, 0700); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(staging, "metrics.json"), []byte("complete staged evidence"), 0600); err != nil {
			t.Fatal(err)
		}
		before := snapshotComparisonPaths(t, dir)
		// Exercise the final publication seam directly: os.Rename would wrongly
		// replace this empty directory even after a successful preflight check.
		if err := publishComparison(staging, output); err == nil {
			t.Fatal("publication overwrote a late caller-owned destination")
		}
		if !reflect.DeepEqual(before, snapshotComparisonPaths(t, dir)) {
			t.Fatal("no-clobber publication changed existing paths")
		}
	})
	for _, mode := range []string{"bad data", "bad profile", "missing data", "missing profile", "existing file", "existing empty directory", "existing bundle", "symlink", "dangling symlink", "missing out", "unknown flag", "positional"} {
		t.Run(mode, func(t *testing.T) {
			dir := t.TempDir()
			out := filepath.Join(dir, "output")
			p, d := profile, data
			extra := []string{}
			switch mode {
			case "bad data", "bad profile":
				bad := filepath.Join(dir, "bad")
				if err := os.WriteFile(bad, []byte("invalid"), 0600); err != nil {
					t.Fatal(err)
				}
				if mode == "bad data" {
					d = bad
				} else {
					p = bad
				}
			case "missing data":
				d += ".missing"
			case "missing profile":
				p += ".missing"
			case "existing file":
				if err := os.WriteFile(out, []byte("preserve me"), 0600); err != nil {
					t.Fatal(err)
				}
			case "existing empty directory", "existing bundle":
				if err := os.Mkdir(out, 0700); err != nil {
					t.Fatal(err)
				}
				if mode == "existing bundle" {
					if err := os.WriteFile(filepath.Join(out, "metrics.json"), []byte("preserve me"), 0600); err != nil {
						t.Fatal(err)
					}
				}
			case "symlink", "dangling symlink":
				target := data
				if mode == "dangling symlink" {
					target += ".missing"
				}
				if err := os.Symlink(target, out); err != nil {
					t.Fatal(err)
				}
			case "missing out":
				out = ""
			case "unknown flag":
				extra = []string{"--online"}
			case "positional":
				extra = []string{"unexpected"}
			}
			before := snapshotComparisonPaths(t, dir)
			if err := invoke(p, d, out, extra...); err == nil {
				t.Fatal("invalid request returned success")
			}
			if after := snapshotComparisonPaths(t, dir); !reflect.DeepEqual(before, after) {
				t.Fatalf("rejection changed caller paths: before=%v after=%v", before, after)
			}
		})
	}
}

func snapshotComparisonPaths(t *testing.T, dir string) map[string]string {
	t.Helper()
	result := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}
			result[path] = "symlink:" + target
		} else if entry.IsDir() {
			result[path] = "directory"
		} else {
			b, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			result[path] = string(b)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}
