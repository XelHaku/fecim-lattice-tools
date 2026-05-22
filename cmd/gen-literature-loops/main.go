// Command gen-literature-loops writes synthetic literature-reference P-E loops
// from built-in material presets.
//
// Usage: gen-literature-loops -preset park -out data.csv
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
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

func writeLiteratureLoopCSV(out string, mat *sharedphysics.HZOMaterial, emax float64) error {
	// Generate E sweep matching the validator: -Emax..Emax..-Emax with 31 points up.
	E := buildSweep(emax)

	model := ferroelectric.NewPreisachModel(mat)
	model.Reset()

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("create output %q: %w", out, err)
	}
	needsClose := true
	defer func() {
		if needsClose {
			_ = f.Close()
		}
	}()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"E_MV_cm", "P_uC_cm2"}); err != nil {
		return fmt.Errorf("write header to %q: %w", out, err)
	}

	for _, e := range E {
		p := model.Update(e * 1e8) // C/m2
		pUC := p * 1e2             // uC/cm2
		if err := w.Write([]string{fmt.Sprintf("%0.3f", e), fmt.Sprintf("%0.6f", pUC)}); err != nil {
			return fmt.Errorf("write row to %q: %w", out, err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flush output %q: %w", out, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close output %q: %w", out, err)
	}
	needsClose = false
	return nil
}

func runGenLiteratureLoops(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gen-literature-loops", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("out", "", "output csv path")
	preset := fs.String("preset", "park", "preset: park|cheema")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *out == "" {
		fmt.Fprintln(stderr, "-out is required")
		return 2
	}

	mat, emax, err := materialForPreset(*preset)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if err := writeLiteratureLoopCSV(*out, mat, emax); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	fmt.Fprintf(stdout, "wrote %s (preset=%s)\n", *out, *preset)
	return 0
}

func main() {
	os.Exit(runGenLiteratureLoops(os.Args[1:], os.Stdout, os.Stderr))
}
