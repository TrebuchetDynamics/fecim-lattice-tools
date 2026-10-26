package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"fecim-lattice-tools/validation/peloop"
)

type comparisonFingerprints struct {
	SourceSHA256          string            `json:"source_sha256"`
	DataSHA256            string            `json:"data_sha256"`
	ProfileSHA256         string            `json:"profile_sha256"`
	ModelParametersSHA256 string            `json:"model_parameters_sha256"`
	ModelSourceSHA256     map[string]string `json:"model_source_sha256"`
	ExecutableSHA256      string            `json:"executable_sha256"`
	SourceHashScope       string            `json:"source_hash_scope"`
}

type comparisonGit struct {
	Revision string `json:"revision"`
	Dirty    bool   `json:"dirty"`
}

// One evidence record drives every artifact. No fit threshold or promotion path
// exists: experimental_validation is false even for identically zero residuals.
type comparisonReport struct {
	Schema                 int                    `json:"schema"`
	ComparisonKind         string                 `json:"comparison_kind"`
	ExperimentalValidation bool                   `json:"experimental_validation"`
	Source                 peloop.SourceProfile   `json:"source"`
	Fingerprints           comparisonFingerprints `json:"fingerprints"`
	Git                    comparisonGit          `json:"git"`
	GoVersion              string                 `json:"go_version"`
	Units                  map[string]string      `json:"units"`
	ExactFieldPolicy       string                 `json:"exact_field_policy"`
	Limitations            []string               `json:"limitations"`
	Metrics                peloop.Metrics         `json:"metrics"`
	Baseline               peloop.BaselineResult  `json:"baseline"`
	Reference              []peloop.Sample        `json:"reference"`
}

func runCompare(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("compare", flag.ContinueOnError)
	flags.SetOutput(stderr)
	profilePath := flags.String("profile", "", "local source-profile JSON")
	dataPath := flags.String("data", "", "local branch-labelled physical CSV (at most 2MiB and 10000 points)")
	outPath := flags.String("out-dir", "", "new bundle directory; parent must exist; never overwritten")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *profilePath == "" || *dataPath == "" || *outPath == "" {
		return fmt.Errorf("compare requires --profile, --data and --out-dir paths, with no positional arguments")
	}
	output := filepath.Clean(*outPath)
	if _, err := os.Lstat(output); err == nil {
		return fmt.Errorf("output path already exists: %s", output)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("output: %w", err)
	}
	profileBytes, err := readComparisonInput(*profilePath)
	if err != nil {
		return fmt.Errorf("profile: %w", err)
	}
	profile, err := peloop.ReadSourceProfile(bytes.NewReader(profileBytes))
	if err != nil {
		return err
	}
	data, err := readComparisonInput(*dataPath)
	if err != nil {
		return fmt.Errorf("data: %w", err)
	}
	reference, err := peloop.ReadSamples(bytes.NewReader(data))
	if err != nil {
		return err
	}
	baseline, err := peloop.SimulateBaseline(reference)
	if err != nil {
		return err
	}
	metrics, err := peloop.Compare(reference, baseline.Samples)
	if err != nil {
		return err
	}
	git, fingerprints, err := comparisonProvenance()
	if err != nil {
		return err
	}
	material, err := json.Marshal(baseline.Material)
	if err != nil {
		return err
	}
	fingerprints.SourceSHA256 = profile.SourceAsset.SHA256
	fingerprints.ProfileSHA256 = comparisonHash(profileBytes)
	fingerprints.DataSHA256 = comparisonHash(data)
	fingerprints.ModelParametersSHA256 = comparisonHash(material)
	report := comparisonReport{
		Schema: 1, ComparisonKind: "unfitted-baseline", ExperimentalValidation: false,
		Source: profile, Fingerprints: fingerprints, Git: git, GoVersion: runtime.Version(),
		Units:            map[string]string{"field": "MV/cm", "polarization": "uC/cm2", "rmse": "uC/cm2", "mae": "uC/cm2", "normalized_rmse": "fraction", "material": "native HZOMaterial units (SI except documented eV and GPa fields)"},
		ExactFieldPolicy: "identical branch/index/field at every input position; no interpolation, extrapolation, reordering or cross-branch matching",
		Limitations: []string{
			"Unfitted educational baseline, not predictive device validity. No residual size establishes experimental validation; mismatch is retained, not fitted away.",
			"Measurement temperature, waveform and exact device diameter are not reported; fabrication anneal is not measurement temperature. Constructor 300 K / 1 GPa are model defaults, not inferred source conditions.",
			"The model uses quasi-static Update, not the source's 10 kHz measurement waveform. Two numerical conditioning cycles (512 equal steps per half-cycle, excluding start/including end) after Reset and an initial minimum-field update end at the minimum field; they do not reproduce 100000 measured wake-up cycles.",
			"Lower/upper geometric branches are conventionally assigned ascending/descending E, not a recovered measurement time sequence. Unresolved merged positive endpoints were omitted, not duplicated or completed; branch spans/counts differ.",
			"Independent extraction selection bounds are +/-2 original pixels per axis for isolated centres and +/-4 for ascending 11 and descending 10..13. Tick-anchor selection has a separate +/-2-pixel bound. These are not experimental instrument uncertainties, statistical standard deviations or fitted error bars. See the dataset extraction.md for the physical bound derivation.",
			"Normalized RMSE divides by observed max(abs(reference P)) per branch or pooled points, not measured Ps; zero denominator is null. Pooled errors are point-weighted, not equal-branch-weighted.",
			"Caller-supplied data bytes are fingerprinted, but their measured origin and trace-fidelity are not verified by this command. Source-asset SHA is declared in the admitted profile, not an offline verification of image bytes. Independent review against the pinned original is required before public promotion.",
			"No headline overlay, deployment, source-fidelity certification, experimental validation or publication approval is produced by this offline comparison.",
		},
		Metrics: metrics, Baseline: baseline, Reference: reference,
	}
	metricsJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("report: %w", err)
	}
	metricsJSON = append(metricsJSON, '\n')
	caption := comparisonCaption(report)
	svg := comparisonSVG(report, metricsJSON, caption)

	// Generate and close all files in a sibling staging directory first. The
	// final no-replace rename is the only operation that publishes success.
	staging, err := os.MkdirTemp(filepath.Dir(output), ".pe-reference-*")
	if err != nil {
		return fmt.Errorf("staging output: %w", err)
	}
	defer os.RemoveAll(staging)
	for _, artifact := range []struct {
		name string
		data []byte
	}{
		{"metrics.json", metricsJSON}, {"overlay.svg", []byte(svg)}, {"caption.txt", []byte(caption)},
	} {
		if err := os.WriteFile(filepath.Join(staging, artifact.name), artifact.data, 0644); err != nil {
			return fmt.Errorf("staging %s: %w", artifact.name, err)
		}
	}
	if err := publishComparison(staging, output); err != nil {
		return fmt.Errorf("publish comparison: %w", err)
	}
	return nil
}

func readComparisonInput(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	const limit = 2 * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err == nil && len(data) > limit {
		err = fmt.Errorf("input exceeds 2 MiB")
	}
	return data, err
}

func comparisonHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// Git and source hashes describe the current checkout, not necessarily the
// source that built an externally supplied binary. Hash that binary separately
// so a dirty run is never represented by HEAD alone. Require a local checkout.
func comparisonProvenance() (comparisonGit, comparisonFingerprints, error) {
	git := func(args ...string) (string, error) {
		out, err := exec.Command("git", args...).Output()
		if err != nil {
			return "", fmt.Errorf("comparison requires Git and a local source checkout (%s): %w", strings.Join(args, " "), err)
		}
		return strings.TrimSpace(string(out)), nil
	}
	root, err := git("rev-parse", "--show-toplevel")
	if err != nil {
		return comparisonGit{}, comparisonFingerprints{}, err
	}
	revision, err := git("rev-parse", "HEAD")
	if err != nil {
		return comparisonGit{}, comparisonFingerprints{}, err
	}
	status, err := git("status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return comparisonGit{}, comparisonFingerprints{}, err
	}
	fingerprints := comparisonFingerprints{
		ModelSourceSHA256: map[string]string{},
		SourceHashScope:   "current-checkout non-test .go files in shared/physics, module1-hysteresis/pkg/ferroelectric, validation/peloop and cmd/pe-reference (package roots); not a claim of compiled-source identity; executable SHA separately identifies the running binary",
	}
	for _, dir := range []string{"shared/physics", "module1-hysteresis/pkg/ferroelectric", "validation/peloop", "cmd/pe-reference"} {
		files, err := os.ReadDir(filepath.Join(root, dir))
		if err != nil {
			return comparisonGit{}, comparisonFingerprints{}, err
		}
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".go") || strings.HasSuffix(file.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(dir, file.Name())
			data, err := os.ReadFile(filepath.Join(root, path))
			if err != nil {
				return comparisonGit{}, comparisonFingerprints{}, err
			}
			fingerprints.ModelSourceSHA256[filepath.ToSlash(path)] = comparisonHash(data)
		}
	}
	executable, err := os.Executable()
	if err != nil {
		return comparisonGit{}, comparisonFingerprints{}, err
	}
	file, err := os.Open(executable)
	if err != nil {
		return comparisonGit{}, comparisonFingerprints{}, err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return comparisonGit{}, comparisonFingerprints{}, err
	}
	fingerprints.ExecutableSHA256 = hex.EncodeToString(hash.Sum(nil))
	return comparisonGit{Revision: revision, Dirty: status != ""}, fingerprints, nil
}

func comparisonCaption(r comparisonReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "HZO Figure %s, %s\n%s\n%s; %s\ncomparison_kind=%s; experimental_validation=false\n", r.Source.Figure, r.Source.Trace, r.Source.Article, r.Source.DOI, r.Source.ArticleURL, r.ComparisonKind)
	fmt.Fprintf(&b, "%s\n%s (%s). Changes: digitized marker centres, affine axes, comparison with separate unchanged simulated lines.\n", r.Source.License.Attribution, r.Source.License.Name, r.Source.License.URL)
	fmt.Fprintf(&b, "Field: MV/cm; polarization, RMSE and MAE: uC/cm2; normalized RMSE: fraction. Matching: %s.\n", r.ExactFieldPolicy)
	for _, branch := range []string{"ascending", "descending", "pooled"} {
		m := r.Metrics.Pooled
		if branch != "pooled" {
			m = r.Metrics.PerBranch[branch]
		}
		normalized := "null"
		if m.NormalizedRMSE != nil {
			normalized = comparisonNumber(*m.NormalizedRMSE)
		}
		fmt.Fprintf(&b, "%s: n=%d; RMSE=%s uC/cm2; MAE=%s uC/cm2; normalizer=%s uC/cm2 (%s); normalized RMSE=%s.\n", branch, m.Count, comparisonNumber(m.RMSEUCcm2), comparisonNumber(m.MAEUCcm2), comparisonNumber(m.ObservedMaxAbsReferencePUCcm2), m.Normalizer, normalized)
	}
	for _, limit := range r.Limitations {
		fmt.Fprintln(&b, limit)
	}
	fmt.Fprintf(&b, "Git revision=%s; dirty=%t; Go=%s\n", r.Git.Revision, r.Git.Dirty, r.GoVersion)
	// Preserve the complete conditions, material/settings and fingerprints, not
	// just a selected human-readable parameter subset.
	for _, item := range []struct {
		name  string
		value any
	}{
		{"source profile", r.Source}, {"material", r.Baseline.Material}, {"settings", r.Baseline.Settings}, {"fingerprints", r.Fingerprints},
	} {
		encoded, _ := json.Marshal(item.value) // All values already marshalled successfully above.
		fmt.Fprintf(&b, "%s: %s\n", item.name, encoded)
	}
	return b.String()
}

func comparisonNumber(value float64) string { return strconv.FormatFloat(value, 'g', -1, 64) }

func comparisonSVG(r comparisonReport, metricsJSON []byte, caption string) string {
	escape := func(s string) string { var b bytes.Buffer; _ = xml.EscapeText(&b, []byte(s)); return b.String() }
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="960" height="760" viewBox="0 0 960 760" role="img" aria-labelledby="title description">` + "\n")
	fmt.Fprintf(&b, "<title id=\"title\">HZO unfitted-baseline comparison; experimental_validation=false</title>\n<desc id=\"description\">%s</desc>\n<metadata>%s</metadata>\n", escape(caption), escape(string(metricsJSON)))
	b.WriteString(`<rect width="960" height="760" fill="white"/><g font-family="sans-serif" font-size="14" fill="black">` + "\n")
	b.WriteString(`<text x="60" y="28">HZO unfitted-baseline — experimental_validation=false</text>` + "\n")
	b.WriteString(`<text x="60" y="52">Reference markers (trace fidelity unverified); separate simulated branch lines</text>` + "\n")
	maxE, maxP := 0.0, 0.0
	for _, samples := range [][]peloop.Sample{r.Reference, r.Baseline.Samples} {
		for _, s := range samples {
			maxE = math.Max(maxE, math.Abs(s.FieldMVcm))
			maxP = math.Max(maxP, math.Abs(s.PolarizationUCcm2))
		}
	}
	if maxE == 0 {
		maxE = 1
	}
	if maxP == 0 {
		maxP = 1
	}
	// Normalize before scaling: even large finite residuals stay visible without
	// overflowing a max-min span or silently clipping the mismatch.
	x := func(v float64) float64 { return 460 + 340*(v/maxE) }
	y := func(v float64) float64 { return 320 - 240*(v/maxP) }
	b.WriteString(`<path d="M120 80V560H800 M120 320H800 M460 80V560" stroke="#aaa" fill="none"/>` + "\n")
	for _, e := range []float64{-maxE, 0, maxE} {
		fmt.Fprintf(&b, `<text x="%.3f" y="584" text-anchor="middle">%s</text>`+"\n", x(e), comparisonNumber(e))
	}
	for _, p := range []float64{-maxP, maxP} {
		// Display tick labels fit the left margin; exact values remain in the
		// metadata, caption and per-marker titles rather than clipped digits.
		fmt.Fprintf(&b, `<text x="112" y="%.3f" text-anchor="end">%.5g</text>`+"\n", y(p), p)
	}
	b.WriteString(`<text x="460" y="610" text-anchor="middle">E (MV/cm)</text><text transform="translate(24 320) rotate(-90)" text-anchor="middle">P (uC/cm2)</text>` + "\n")
	for i, branch := range []string{"ascending", "descending"} {
		color := []string{"#a52a2a", "#164b91"}[i]
		var points strings.Builder
		for _, s := range r.Baseline.Samples {
			if s.Branch == branch {
				fmt.Fprintf(&points, "%.3f,%.3f ", x(s.FieldMVcm), y(s.PolarizationUCcm2))
			}
		}
		fmt.Fprintf(&b, `<polyline fill="none" stroke="%s" stroke-width="2" points="%s"><title>%s simulated baseline</title></polyline>`+"\n", color, points.String(), branch)
		for _, s := range r.Reference {
			if s.Branch == branch {
				fmt.Fprintf(&b, `<circle cx="%.3f" cy="%.3f" r="4" fill="white" stroke="%s" stroke-width="2"><title>%s reference %d: E=%s MV/cm; P=%s uC/cm2</title></circle>`+"\n", x(s.FieldMVcm), y(s.PolarizationUCcm2), color, branch, s.PointIndex, comparisonNumber(s.FieldMVcm), comparisonNumber(s.PolarizationUCcm2))
			}
		}
		m := r.Metrics.PerBranch[branch]
		fmt.Fprintf(&b, `<text x="60" y="%d" fill="%s">%s: n=%d; RMSE=%s uC/cm2</text>`+"\n", 638+i*22, color, branch, m.Count, comparisonNumber(m.RMSEUCcm2))
	}
	fmt.Fprintf(&b, `<text x="60" y="682">Pooled: n=%d; RMSE=%s uC/cm2; normalizer=%s uC/cm2 (observed max |P|, not Ps)</text>`+"\n", r.Metrics.Pooled.Count, comparisonNumber(r.Metrics.Pooled.RMSEUCcm2), comparisonNumber(r.Metrics.Pooled.ObservedMaxAbsReferencePUCcm2))
	b.WriteString(`<text x="60" y="706">Quasi-static, not 10 kHz dynamics; measurement temperature/waveform not reported.</text>` + "\n")
	fmt.Fprintf(&b, `<text x="60" y="730">Kim et al. (2020), DOI %s, Fig. %s Ref.; CC BY 4.0; digitized + simulated.</text>`+"\n", escape(r.Source.DOI), escape(r.Source.Figure))
	b.WriteString("</g></svg>\n")
	return b.String()
}
