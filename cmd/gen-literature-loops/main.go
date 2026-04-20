// Command gen-literature-loops writes synthetic literature-reference P-E loops
// from built-in material presets.
//
// Usage: gen-literature-loops -preset park -out data.csv
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

func materialForPreset(preset string) (*sharedphysics.HZOMaterial, float64, error) {
	switch preset {
	case "park":
		return sharedphysics.Park2015Fig2aHZO10nm(), 3.0, nil
	case "cheema":
		return sharedphysics.Cheema2020Fig2cHZOSuperlattice5nm(), 4.0, nil
	default:
		return nil, 0, fmt.Errorf("unknown preset %q", preset)
	}
}

func buildSweep(emax float64) []float64 {
	E := make([]float64, 0, 61)
	steps := 31
	for i := 0; i < steps; i++ {
		e := -emax + 2*emax*float64(i)/float64(steps-1)
		E = append(E, e)
	}
	for i := steps - 2; i >= 0; i-- {
		e := -emax + 2*emax*float64(i)/float64(steps-1)
		E = append(E, e)
	}
	return E
}

func main() {
	out := flag.String("out", "", "output csv path")
	preset := flag.String("preset", "park", "preset: park|cheema")
	flag.Parse()
	if *out == "" {
		fmt.Fprintln(os.Stderr, "-out is required")
		os.Exit(2)
	}

	mat, emax, err := materialForPreset(*preset)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// Generate E sweep matching the validator: -Emax..Emax..-Emax with 31 points up.
	E := buildSweep(emax)

	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	f, err := os.Create(*out)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	_ = w.Write([]string{"E_MV_cm", "P_uC_cm2"})

	for _, e := range E {
		p := model.Update(e * 1e8) // C/m2
		pUC := p * 1e2             // uC/cm2
		_ = w.Write([]string{fmt.Sprintf("%0.3f", e), fmt.Sprintf("%0.6f", pUC)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		panic(err)
	}

	fmt.Printf("wrote %s (preset=%s)\n", *out, *preset)
}
