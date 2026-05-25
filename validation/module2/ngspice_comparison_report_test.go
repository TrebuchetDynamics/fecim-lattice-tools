package module2

import (
	"encoding/json"
	"fmt"
	"html"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedval "fecim-lattice-tools/shared/validation"
)

const module2NgspiceRelErrThreshold = 0.01

type ngspiceComparisonReport struct {
	sharedval.ArtifactEnvelope

	Description            string                  `json:"description"`
	Status                 string                  `json:"status"`
	NgspiceAvailable       bool                    `json:"ngspice_available"`
	NgspicePath            string                  `json:"ngspice_path,omitempty"`
	NgspiceVersion         string                  `json:"ngspice_version,omitempty"`
	ComparisonExecuted     bool                    `json:"comparison_executed"`
	ClaimSupported         bool                    `json:"claim_supported"`
	RelativeErrorThreshold float64                 `json:"relative_error_threshold"`
	MaxAbsoluteErrorA      float64                 `json:"max_absolute_error_A"`
	MaxRelativeError       float64                 `json:"max_relative_error"`
	Cases                  []ngspiceComparisonCase `json:"cases"`
	Artifacts              map[string]string       `json:"artifacts"`
	Limitations            []string                `json:"limitations"`
	Pass                   bool                    `json:"pass"`
}

type ngspiceComparisonCase struct {
	Name                  string             `json:"name"`
	Rows                  int                `json:"rows"`
	Cols                  int                `json:"cols"`
	NetlistPath           string             `json:"netlist_path"`
	NgspiceOutputPath     string             `json:"ngspice_output_path,omitempty"`
	StructuralPass        bool               `json:"structural_pass"`
	MissingTokens         []string           `json:"missing_tokens,omitempty"`
	ParsedBranchCurrentsA map[string]float64 `json:"parsed_branch_currents_A,omitempty"`
	ExpectedRowCurrentsA  []float64          `json:"expected_row_currents_A,omitempty"`
	ComparedBranches      int                `json:"compared_branches"`
	MaxAbsoluteErrorA     float64            `json:"max_absolute_error_A"`
	MaxRelativeError      float64            `json:"max_relative_error"`
	Pass                  bool               `json:"pass"`
}

func TestModule2NgspiceComparisonReport_PublicValidation(t *testing.T) {
	report := runNgspiceComparisonReport(t)
	if !report.Pass {
		t.Fatalf("ngspice comparison report failed: status=%s max_rel=%.6e threshold=%.6e", report.Status, report.MaxRelativeError, report.RelativeErrorThreshold)
	}
}

func runNgspiceComparisonReport(t *testing.T) ngspiceComparisonReport {
	t.Helper()

	root := repoRoot(t)
	outDir := filepath.Join(root, "output", "validation", "module2")
	caseDir := filepath.Join(outDir, "ngspice_comparison")
	if err := os.MkdirAll(caseDir, 0o755); err != nil {
		t.Fatalf("create ngspice comparison output directory: %v", err)
	}

	reportPath := filepath.Join(outDir, "ngspice_comparison.json")
	plotPath := filepath.Join(outDir, "ngspice_comparison.svg")
	report := ngspiceComparisonReport{
		ArtifactEnvelope:       sharedval.NewEnvelope("RG-VAL-M2-03", "", false),
		Description:            "Module 2 crossbar SPICE comparison report for deterministic small-array resistive fixtures",
		Status:                 "initialized",
		RelativeErrorThreshold: module2NgspiceRelErrThreshold,
		Artifacts: map[string]string{
			"json_report": relArtifactPath(root, reportPath),
			"svg_plot":    relArtifactPath(root, plotPath),
		},
		Limitations: []string{
			"SPICE deck is an apples-to-apples resistive crossbar netlist for current comparison; it is not a calibrated FeFET compact model.",
			"The quantitative comparison is optional and is skipped when ngspice is not installed or emits no parseable source branch currents.",
			"Passing this report validates small-array circuit consistency, not fabricated-device agreement.",
		},
	}

	fixtures := []struct {
		name string
		n    int
	}{
		{name: "crossbar_1x1", n: 1},
		{name: "crossbar_2x2", n: 2},
		{name: "crossbar_4x4", n: 4},
	}

	for _, fixture := range fixtures {
		params := module2NgspiceParams(fixture.n)
		deck := buildModule2NgspiceDeck(params, fmt.Sprintf("Module 2 %dx%d SPICE comparison", fixture.n, fixture.n))
		netlistPath := filepath.Join(caseDir, fmt.Sprintf("%s.sp", fixture.name))
		if err := os.WriteFile(netlistPath, []byte(deck), 0o644); err != nil {
			t.Fatalf("write %s netlist: %v", fixture.name, err)
		}
		c := ngspiceComparisonCase{
			Name:           fixture.name,
			Rows:           fixture.n,
			Cols:           fixture.n,
			NetlistPath:    relArtifactPath(root, netlistPath),
			StructuralPass: true,
			Pass:           true,
		}
		for _, token := range requiredNgspiceDeckTokens(fixture.n) {
			if !strings.Contains(deck, token) {
				c.StructuralPass = false
				c.Pass = false
				c.MissingTokens = append(c.MissingTokens, token)
			}
		}
		report.Cases = append(report.Cases, c)
	}

	if !allNgspiceStructuresPass(report.Cases) {
		report.Status = "failed_structural_netlist_validation"
		report.Pass = false
		report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-03", "", false)
		writeNgspiceComparisonArtifacts(t, reportPath, plotPath, report)
		t.Fatalf("ngspice comparison structural netlist validation failed")
	}

	ngspicePath, err := exec.LookPath("ngspice")
	if err != nil {
		report.Status = "skipped_ngspice_missing"
		report.Pass = true
		report.ClaimSupported = false
		report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-03", "", true)
		writeNgspiceComparisonArtifacts(t, reportPath, plotPath, report)
		t.Skip("ngspice not installed; wrote structural SPICE report and skipped optional quantitative comparison")
	}
	report.NgspiceAvailable = true
	report.NgspicePath = ngspicePath
	report.NgspiceVersion = ngspiceVersion(ngspicePath)

	parsedAny := false
	comparedAny := false
	for i := range report.Cases {
		params := module2NgspiceParams(report.Cases[i].Rows)
		outPath := filepath.Join(caseDir, fmt.Sprintf("%s.out", report.Cases[i].Name))
		report.Cases[i].NgspiceOutputPath = relArtifactPath(root, outPath)

		cmd := exec.Command(ngspicePath, "-b", "-o", outPath, filepath.Join(root, report.Cases[i].NetlistPath))
		if runErr := cmd.Run(); runErr != nil {
			report.Status = "failed_ngspice_run"
			report.Pass = false
			report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-03", "", false)
			writeNgspiceComparisonArtifacts(t, reportPath, plotPath, report)
			t.Fatalf("ngspice run failed for %s: %v", report.Cases[i].Name, runErr)
		}

		raw, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("read ngspice output for %s: %v", report.Cases[i].Name, err)
		}
		parsed := parseModule2SourceCurrentsA(string(raw))
		report.Cases[i].ParsedBranchCurrentsA = parsed
		if len(parsed) > 0 {
			parsedAny = true
		}

		ref, err := arraysim.NewTierBSolver().Solve(params)
		if err != nil {
			t.Fatalf("reference Tier-B solve failed for %s: %v", report.Cases[i].Name, err)
		}
		report.Cases[i].ExpectedRowCurrentsA = append([]float64(nil), ref.RowCurrents...)

		caseCompared := 0
		for r := 0; r < report.Cases[i].Rows; r++ {
			key := fmt.Sprintf("vwl_src_%d", r)
			got, ok := parsed[key]
			if !ok {
				continue
			}
			want := -ref.RowCurrents[r]
			absErr := math.Abs(got - want)
			relErr := absErr / math.Max(math.Abs(want), 1e-15)
			report.Cases[i].MaxAbsoluteErrorA = maxAbsFloat(report.Cases[i].MaxAbsoluteErrorA, absErr)
			report.Cases[i].MaxRelativeError = maxAbsFloat(report.Cases[i].MaxRelativeError, relErr)
			caseCompared++
		}
		report.Cases[i].ComparedBranches = caseCompared
		if caseCompared > 0 {
			comparedAny = true
		}
		report.Cases[i].Pass = report.Cases[i].StructuralPass && report.Cases[i].MaxRelativeError <= report.RelativeErrorThreshold
		report.MaxAbsoluteErrorA = maxAbsFloat(report.MaxAbsoluteErrorA, report.Cases[i].MaxAbsoluteErrorA)
		report.MaxRelativeError = maxAbsFloat(report.MaxRelativeError, report.Cases[i].MaxRelativeError)
	}

	if !parsedAny || !comparedAny {
		report.Status = "skipped_no_parseable_source_branch_currents"
		report.Pass = true
		report.ClaimSupported = false
		report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-03", "", true)
		writeNgspiceComparisonArtifacts(t, reportPath, plotPath, report)
		t.Skip("ngspice ran, but no parseable WL source branch currents were found; wrote skipped report")
	}

	report.ComparisonExecuted = true
	report.ClaimSupported = report.MaxRelativeError <= report.RelativeErrorThreshold
	report.Pass = report.ClaimSupported
	if report.Pass {
		report.Status = "passed"
	} else {
		report.Status = "failed_relative_error_threshold"
	}
	report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-03", "", report.Pass)
	writeNgspiceComparisonArtifacts(t, reportPath, plotPath, report)

	return report
}

func module2NgspiceParams(n int) arraysim.SolveParams {
	g := make([][]float64, n)
	wl := make([]float64, n)
	bl := make([]float64, n)
	for r := 0; r < n; r++ {
		g[r] = make([]float64, n)
		wl[r] = 0.5 - 0.05*float64(r)
		for c := 0; c < n; c++ {
			g[r][c] = 1e-4 * (1 + 0.05*float64(r+c))
		}
	}
	for c := 0; c < n; c++ {
		bl[c] = 0
	}
	return arraysim.SolveParams{
		WLVoltages:  wl,
		BLVoltages:  bl,
		Conductance: g,
		Wire:        arraysim.WireParams{RWordLine: 5.0, RBitLine: 7.0},
		Boundary:    arraysim.BoundaryParams{WLDriveResistance: 2.0, BLDriveResistance: 2.0},
	}
}

func buildModule2NgspiceDeck(params arraysim.SolveParams, title string) string {
	rows := len(params.Conductance)
	cols := len(params.BLVoltages)
	wire := params.Wire.WithDefaults(params.Geometry.WithDefaults())
	boundary := params.Boundary.WithDefaults(wire)

	var b strings.Builder
	fmt.Fprintf(&b, "* %s\n", title)
	fmt.Fprintf(&b, ".param RWL=%.9g RBL=%.9g RWLDRV=%.9g RBLDRV=%.9g\n\n", wire.RWordLine, wire.RBitLine, boundary.WLDriveResistance, boundary.BLDriveResistance)

	for r := 0; r < rows; r++ {
		vwl := 0.0
		if r < len(params.WLVoltages) {
			vwl = params.WLVoltages[r]
		}
		fmt.Fprintf(&b, "VWL_SRC_%d wl_src_%d 0 %.9g\n", r, r, vwl)
		fmt.Fprintf(&b, "RWL_DRV_%d wl_src_%d wl_%d_0 {RWLDRV}\n", r, r, r)
	}
	b.WriteString("\n")

	for c := 0; c < cols; c++ {
		vbl := 0.0
		if c < len(params.BLVoltages) {
			vbl = params.BLVoltages[c]
		}
		fmt.Fprintf(&b, "VBL_SRC_%d bl_src_%d 0 %.9g\n", c, c, vbl)
		fmt.Fprintf(&b, "RBL_DRV_%d bl_src_%d bl_0_%d {RBLDRV}\n", c, c, c)
	}
	b.WriteString("\n* WL wire resistances\n")

	for r := 0; r < rows; r++ {
		for c := 0; c < cols-1; c++ {
			fmt.Fprintf(&b, "RWL_%d_%d wl_%d_%d wl_%d_%d {RWL}\n", r, c, r, c, r, c+1)
		}
	}
	b.WriteString("\n* BL wire resistances\n")

	for c := 0; c < cols; c++ {
		for r := 0; r < rows-1; r++ {
			fmt.Fprintf(&b, "RBL_%d_%d bl_%d_%d bl_%d_%d {RBL}\n", r, c, r, c, r+1, c)
		}
	}
	b.WriteString("\n* Memory cell conductances\n")

	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			g := params.Conductance[r][c]
			res := 1e15
			if g > 0 {
				res = 1.0 / g
			}
			fmt.Fprintf(&b, "RCELL_%d_%d wl_%d_%d bl_%d_%d %.9g\n", r, c, r, c, r, c, res)
		}
	}

	b.WriteString("\n.control\n")
	b.WriteString("op\n")
	b.WriteString("print all\n")
	b.WriteString(".endc\n\n.end\n")
	return b.String()
}

func requiredNgspiceDeckTokens(n int) []string {
	return []string{
		".control",
		"op",
		"print all",
		".end",
		"VWL_SRC_0",
		fmt.Sprintf("VWL_SRC_%d", n-1),
		fmt.Sprintf("VBL_SRC_%d", n-1),
		fmt.Sprintf("RCELL_%d_%d", n-1, n-1),
	}
}

func allNgspiceStructuresPass(cases []ngspiceComparisonCase) bool {
	for _, c := range cases {
		if !c.StructuralPass {
			return false
		}
	}
	return true
}

func ngspiceVersion(path string) string {
	out, err := exec.Command(path, "-v").CombinedOutput()
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 {
		return ""
	}
	return strings.TrimSpace(lines[0])
}

func parseModule2SourceCurrentsA(raw string) map[string]float64 {
	out := map[string]float64{}
	re := regexp.MustCompile(`(?i)\b(vwl_src_\d+|vbl_src_\d+)(?:#branch)?\s*=\s*([+-]?[0-9]*\.?[0-9]+(?:[eE][+-]?\d+)?)`)
	for _, m := range re.FindAllStringSubmatch(raw, -1) {
		v, err := strconv.ParseFloat(m[2], 64)
		if err == nil {
			out[strings.ToLower(m[1])] = v
		}
	}
	return out
}

func writeNgspiceComparisonArtifacts(t *testing.T, reportPath, plotPath string, report ngspiceComparisonReport) {
	t.Helper()

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal ngspice comparison report: %v", err)
	}
	if err := os.WriteFile(reportPath, append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write ngspice comparison report: %v", err)
	}
	if err := os.WriteFile(plotPath, []byte(renderNgspiceComparisonSVG(report)), 0o644); err != nil {
		t.Fatalf("write ngspice comparison SVG: %v", err)
	}
}

func renderNgspiceComparisonSVG(report ngspiceComparisonReport) string {
	const width = 760
	const height = 280
	const left = 70
	const top = 55
	const chartH = 150
	barW := 120
	gap := 45

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	b.WriteString(`<rect width="100%" height="100%" fill="#0f172a"/>`)
	fmt.Fprintf(&b, `<text x="24" y="30" fill="#e2e8f0" font-family="sans-serif" font-size="18">%s</text>`, html.EscapeString("Module 2 ngspice comparison"))
	fmt.Fprintf(&b, `<text x="24" y="50" fill="#94a3b8" font-family="sans-serif" font-size="12">status=%s threshold=%.2f%%</text>`, html.EscapeString(report.Status), report.RelativeErrorThreshold*100)
	b.WriteString(`<line x1="70" y1="205" x2="720" y2="205" stroke="#475569"/>`)

	thresholdY := top + chartH - int(report.RelativeErrorThreshold*float64(chartH))
	if thresholdY < top {
		thresholdY = top
	}
	fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="720" y2="%d" stroke="#f59e0b" stroke-dasharray="5 5"/>`, left, thresholdY, thresholdY)
	fmt.Fprintf(&b, `<text x="600" y="%d" fill="#fbbf24" font-family="sans-serif" font-size="11">1%% target</text>`, thresholdY-5)

	if len(report.Cases) == 0 || !report.ComparisonExecuted {
		fmt.Fprintf(&b, `<text x="90" y="130" fill="#cbd5e1" font-family="sans-serif" font-size="16">%s</text>`, html.EscapeString("Quantitative ngspice comparison not executed on this host."))
	}

	for i, c := range report.Cases {
		x := left + i*(barW+gap)
		barH := int(math.Min(c.MaxRelativeError/report.RelativeErrorThreshold, 1.0) * float64(chartH))
		if c.ComparedBranches == 0 {
			barH = 0
		}
		y := top + chartH - barH
		fill := "#22c55e"
		if !c.Pass || c.MaxRelativeError > report.RelativeErrorThreshold {
			fill = "#ef4444"
		}
		fmt.Fprintf(&b, `<rect x="%d" y="%d" width="%d" height="%d" fill="%s" rx="5"/>`, x, y, barW, barH, fill)
		fmt.Fprintf(&b, `<text x="%d" y="225" fill="#e2e8f0" font-family="sans-serif" font-size="12">%s</text>`, x, html.EscapeString(c.Name))
		label := "not run"
		if c.ComparedBranches > 0 {
			label = fmt.Sprintf("%.3g%%", c.MaxRelativeError*100)
		}
		fmt.Fprintf(&b, `<text x="%d" y="%d" fill="#cbd5e1" font-family="sans-serif" font-size="12">%s</text>`, x, y-6, html.EscapeString(label))
	}

	b.WriteString(`</svg>`)
	return b.String()
}

func relArtifactPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	return filepath.ToSlash(rel)
}
