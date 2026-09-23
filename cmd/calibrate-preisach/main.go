// Command calibrate-preisach fits Preisach model parameters to measured P-E loop data.
//
// Usage: calibrate-preisach -csv data.csv -preset default_hzo
// Prints the best-fit summary to stdout.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

type calibrationFit struct {
	Ec, Ps, Pr float64
	RMSEFS     float64
	PrErrPct   float64
	EcErrPct   float64
	AreaErrPct float64
}

type calibrationArtifact struct {
	SchemaVersion int    `json:"schema_version"`
	Model         string `json:"model"`
	Preset        string `json:"preset"`
	Input         struct {
		File   string `json:"file"`
		SHA256 string `json:"sha256"`
		Units  string `json:"units"`
	} `json:"input"`
	Fit struct {
		EcVM       float64 `json:"ec_v_m"`
		PsCM2      float64 `json:"ps_c_m2"`
		PrCM2      float64 `json:"pr_c_m2"`
		RMSEFS     float64 `json:"rmse_fs"`
		PrErrPct   float64 `json:"pr_error_percent"`
		EcErrPct   float64 `json:"ec_error_percent"`
		AreaErrPct float64 `json:"area_error_percent"`
	} `json:"fit"`
	Boundary string `json:"boundary"`
}

func main() {
	os.Exit(runCalibratePreisach(os.Args[1:], os.Stdout, os.Stderr))
}

func runCalibratePreisach(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("calibrate-preisach", flag.ContinueOnError)
	flags.SetOutput(stderr)
	csvPath := flags.String("csv", "", "CSV with E_MV_cm,P_uC_cm2")
	preset := flags.String("preset", "default_hzo", "base preset: default_hzo|literature_superlattice")
	output := flags.String("out", "", "optional new JSON calibration artifact path")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *csvPath == "" {
		if _, err := fmt.Fprintln(stderr, "-csv is required"); err != nil {
			return 2
		}
		return 2
	}

	inputBytes, err := os.ReadFile(*csvPath)
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "load csv: %v\n", err); writeErr != nil {
			return 1
		}
		return 1
	}
	E, P, err := parseCSVLoop(inputBytes)
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "load csv: %v\n", err); writeErr != nil {
			return 1
		}
		return 1
	}

	base := baseMaterial(*preset)
	if base == nil {
		if _, writeErr := fmt.Fprintf(stderr, "unknown preset %q\n", *preset); writeErr != nil {
			return 1
		}
		return 1
	}

	// Grid search over Ec, Ps, Pr ratio.
	// Keep search small: pragmatic calibration for matching digitized loop shape.
	targetFS := maxAbs(P)
	if targetFS == 0 {
		targetFS = 1
	}

	b := calibrationFit{RMSEFS: math.Inf(1)}

	// Reference metrics from digitized curve using same estimators.
	prRef := estimatePrAtZeroField(P, E)
	ecRef := estimateEc(P, E)
	areaRef := loopAreaEnergyDensity(E, P)

	for ecScale := 0.6; ecScale <= 1.6; ecScale += 0.05 {
		for psScale := 0.6; psScale <= 1.8; psScale += 0.05 {
			for prRatio := 0.3; prRatio <= 0.95; prRatio += 0.05 {
				m := *base
				m.Ec = base.Ec * ecScale
				m.Ps = base.Ps * psScale
				m.Pr = m.Ps * prRatio
				// Disable reversible dielectric contribution for fitting loop shape.
				m.Epsilon = 1
				m.EpsilonLF = 1

				Psim := simulatePreisachLoop(&m, E)
				rmse := rmse(Psim, P)
				rmseFS := rmse / targetFS

				prSim := estimatePrAtZeroField(Psim, E)
				ecSim := estimateEc(Psim, E)
				areaSim := loopAreaEnergyDensity(E, Psim)

				prErr := pctErr(prSim, prRef)
				ecErr := pctErr(ecSim, ecRef)
				areaErr := pctErr(areaSim, areaRef)

				// Objective: prioritize RMSEFS, but penalize Pr/Ec/area mismatches.
				score := rmseFS + 0.01*prErr/100.0 + 0.01*ecErr/100.0 + 0.005*areaErr/100.0
				bestScore := b.RMSEFS + 0.01*b.PrErrPct/100.0 + 0.01*b.EcErrPct/100.0 + 0.005*b.AreaErrPct/100.0

				if score < bestScore {
					b = calibrationFit{Ec: m.Ec, Ps: m.Ps, Pr: m.Pr, RMSEFS: rmseFS, PrErrPct: prErr, EcErrPct: ecErr, AreaErrPct: areaErr}
				}
			}
		}
	}

	if _, err := fmt.Fprintf(stdout, "best Ec=%0.3e V/m (%0.3f MV/cm)\nbest Ps=%0.3e C/m2 (%0.3f uC/cm2)\nbest Pr=%0.3e C/m2 (%0.3f uC/cm2)\nmetrics: rmseFS=%0.4f prErr=%0.2f%% ecErr=%0.2f%% areaErr=%0.2f%%\nref: pr=%0.3f uC/cm2 ec=%0.3f MV/cm area=%0.3e J/m3\n", b.Ec, b.Ec/1e8, b.Ps, b.Ps*1e2, b.Pr, b.Pr*1e2, b.RMSEFS, b.PrErrPct, b.EcErrPct, b.AreaErrPct, prRef, ecRef, areaRef); err != nil {
		return 1
	}
	if *output != "" {
		if err := writeCalibrationArtifact(*output, *csvPath, *preset, b, inputBytes); err != nil {
			if _, writeErr := fmt.Fprintf(stderr, "write calibration artifact: %v\n", err); writeErr != nil {
				return 1
			}
			return 1
		}
	}
	return 0
}

func writeCalibrationArtifact(path, inputPath, preset string, fit calibrationFit, input []byte) error {
	artifact := calibrationArtifact{SchemaVersion: 1, Model: "Preisach", Preset: preset, Boundary: "fit to supplied P-E data; simulation calibration, not device validation"}
	artifact.Input.File = filepath.Base(inputPath)
	digest := sha256.Sum256(input)
	artifact.Input.SHA256 = hex.EncodeToString(digest[:])
	artifact.Input.Units = "E_MV_cm,P_uC_cm2"
	artifact.Fit.EcVM = fit.Ec
	artifact.Fit.PsCM2 = fit.Ps
	artifact.Fit.PrCM2 = fit.Pr
	artifact.Fit.RMSEFS = fit.RMSEFS
	artifact.Fit.PrErrPct = fit.PrErrPct
	artifact.Fit.EcErrPct = fit.EcErrPct
	artifact.Fit.AreaErrPct = fit.AreaErrPct
	blob, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	blob = append(blob, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.Write(blob); err != nil {
		return errors.Join(err, file.Close(), os.Remove(path))
	}
	if err := file.Close(); err != nil {
		return errors.Join(err, os.Remove(path))
	}
	return nil
}

func baseMaterial(name string) *sharedphysics.HZOMaterial {
	switch name {
	case "default_hzo":
		return sharedphysics.DefaultHZO()
	case "literature_superlattice":
		return sharedphysics.LiteratureSuperlattice()
	default:
		return nil
	}
}

func loadCSVLoop(path string) ([]float64, []float64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return parseCSVLoop(data)
}

func parseCSVLoop(data []byte) ([]float64, []float64, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	if len(rows) < 2 {
		return nil, nil, fmt.Errorf("CSV must contain a header and at least one data row")
	}
	E := make([]float64, 0, len(rows)-1)
	P := make([]float64, 0, len(rows)-1)
	for i := 1; i < len(rows); i++ {
		if len(rows[i]) < 2 {
			return nil, nil, fmt.Errorf("row %d: expected at least 2 columns", i+1)
		}
		e, err := strconv.ParseFloat(rows[i][0], 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d: parse E_MV_cm: %w", i+1, err)
		}
		p, err := strconv.ParseFloat(rows[i][1], 64)
		if err != nil {
			return nil, nil, fmt.Errorf("row %d: parse P_uC_cm2: %w", i+1, err)
		}
		if math.IsNaN(e) || math.IsInf(e, 0) || math.IsNaN(p) || math.IsInf(p, 0) {
			return nil, nil, fmt.Errorf("row %d: field and polarization must be finite", i+1)
		}
		E = append(E, e)
		P = append(P, p)
	}
	return E, P, nil
}

func simulatePreisachLoop(mat *sharedphysics.HZOMaterial, E_MV_cm []float64) []float64 {
	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()
	out := make([]float64, len(E_MV_cm))
	for i, e := range E_MV_cm {
		P_c_m2 := model.Update(e * 1e8)
		out[i] = P_c_m2 * 1e2
	}
	return out
}

func estimatePrAtZeroField(P, E []float64) float64 {
	best := math.Inf(1)
	pr := 0.0
	for i := range E {
		d := math.Abs(E[i])
		if d < best {
			best = d
			pr = math.Abs(P[i])
		}
	}
	return pr
}

func estimateEc(P, E []float64) float64 {
	cross := make([]float64, 0, 4)
	for i := 1; i < len(P); i++ {
		if P[i-1] == 0 {
			cross = append(cross, math.Abs(E[i-1]))
			continue
		}
		if (P[i-1] > 0) != (P[i] > 0) {
			t := math.Abs(P[i-1]) / (math.Abs(P[i-1]) + math.Abs(P[i]))
			e0 := E[i-1] + t*(E[i]-E[i-1])
			cross = append(cross, math.Abs(e0))
		}
	}
	if len(cross) == 0 {
		return 0
	}
	s := 0.0
	for _, c := range cross {
		s += c
	}
	return s / float64(len(cross))
}

func rmse(a, b []float64) float64 {
	s := 0.0
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return math.Sqrt(s / float64(len(a)))
}

func maxAbs(x []float64) float64 {
	m := 0.0
	for _, v := range x {
		if a := math.Abs(v); a > m {
			m = a
		}
	}
	return m
}

func loopAreaEnergyDensity(E_MV_cm, P_uC_cm2 []float64) float64 {
	w := 0.0
	for i := 1; i < len(E_MV_cm); i++ {
		e1 := E_MV_cm[i-1] * 1e8
		e2 := E_MV_cm[i] * 1e8
		p1 := P_uC_cm2[i-1] * 1e-2
		p2 := P_uC_cm2[i] * 1e-2
		dp := p2 - p1
		w += 0.5 * (e1 + e2) * dp
	}
	return math.Abs(w)
}

func pctErr(sim, ref float64) float64 {
	den := math.Abs(ref)
	if den == 0 {
		return math.Inf(1)
	}
	return math.Abs(sim-ref) / den * 100
}
