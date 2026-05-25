// Command gen_golden_loops generates golden regression data for hysteresis.
//
// Usage: gen_golden_loops -material fecim_hzo -output golden/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

type GoldenLoopData struct {
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Generated   string                 `json:"generated"`
	Material    string                 `json:"material"`
	Engine      string                 `json:"engine"`
	Parameters  map[string]interface{} `json:"parameters"`
	Data        struct {
		E []float64 `json:"E"`
		P []float64 `json:"P"`
	} `json:"data"`
}

func runGenGoldenLoops(args []string, stdout, stderr io.Writer) int {
	defaultOut := filepath.Join("module1-hysteresis", "pkg", "ferroelectric", "testdata")
	fs := flag.NewFlagSet("gen_golden_loops", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outDir := fs.String("output", defaultOut, "output directory for golden loop JSON files")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	materials := physics.AllMaterials()

	fmt.Fprintf(stdout, "Found %d materials\n", len(materials))

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fmt.Fprintf(stderr, "prepare output directory %q: %v\n", *outDir, err)
		return 1
	}

	for _, mat := range materials {
		fmt.Fprintf(stdout, "Processing: %s\n", mat.Name)

		if mat.Ec <= 0 || mat.Ps <= 0 {
			fmt.Fprintln(stdout, "  SKIP: missing Ec or Ps")
			continue
		}

		// PREISACH ENGINE
		preisachModel := ferroelectric.NewPreisachModel(mat)
		preisachModel.Reset()

		Emax := 2.0 * mat.Ec
		points := 100

		E_p, P_p := generatePreisachLoop(preisachModel, Emax, points)

		safeName := safeFilename(mat.Name)
		goldenP := GoldenLoopData{
			Version:     "1.5.0",
			Description: fmt.Sprintf("Golden reference hysteresis loop for %s (Preisach engine)", mat.Name),
			Generated:   time.Now().Format("2006-01-02"),
			Material:    mat.Name,
			Engine:      "Preisach",
			Parameters: map[string]interface{}{
				"Emax_multiplier": 2,
				"points":          points,
				"Ec_V_m":          mat.Ec,
				"Ps_C_m2":         mat.Ps,
				"Pr_C_m2":         mat.Pr,
			},
		}
		goldenP.Data.E = E_p
		goldenP.Data.P = P_p

		pFile := filepath.Join(*outDir, fmt.Sprintf("golden_loop_%s_preisach.json", safeName))
		if err := writeJSON(pFile, goldenP); err != nil {
			fmt.Fprintf(stderr, "write golden JSON %q: %v\n", pFile, err)
			return 1
		}
		fmt.Fprintf(stdout, "  ✓ Preisach: %s\n", filepath.Base(pFile))

		fmt.Fprintln(stdout, "  ○ LK: pending (see shared/physics/landau.go)")
	}

	fmt.Fprintln(stdout, "\nDone! Golden loops generated.")
	return 0
}

func main() {
	os.Exit(runGenGoldenLoops(os.Args[1:], os.Stdout, os.Stderr))
}

func generatePreisachLoop(model *ferroelectric.PreisachModel, Emax float64, points int) ([]float64, []float64) {
	if points <= 1 {
		return nil, nil
	}
	E := make([]float64, points)
	P := make([]float64, points)

	step := 2 * Emax / float64(points-1)

	for i := 0; i < points; i++ {
		e := -Emax + step*float64(i)
		P[i] = model.Update(e)
		E[i] = e
	}

	return E, P
}

func safeFilename(name string) string {
	result := []rune{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result = append(result, r)
		} else if r >= 'A' && r <= 'Z' {
			result = append(result, r)
		} else if r == ' ' || r == '-' || r == '(' || r == ')' {
			result = append(result, '_')
		}
	}
	return string(result)
}

func writeJSON(path string, data GoldenLoopData) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	needsClose := true
	defer func() {
		if needsClose {
			_ = f.Close()
		}
	}()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	needsClose = false
	return nil
}
