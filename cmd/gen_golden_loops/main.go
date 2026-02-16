// Command gen_golden_loops generates golden regression data for hysteresis.
//
// Usage: gen_golden_loops -material fecim_hzo -output golden/
package main

import (
	"encoding/json"
	"fmt"
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

func main() {
	materials := physics.AllMaterials()
	
	fmt.Printf("Found %d materials\n", len(materials))
	
	outDir := "<local-path>"
	os.MkdirAll(outDir, 0755)
	
	for _, mat := range materials {
		fmt.Printf("Processing: %s\n", mat.Name)
		
		if mat.Ec <= 0 || mat.Ps <= 0 {
			fmt.Printf("  SKIP: missing Ec or Ps\n")
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
		
		pFile := filepath.Join(outDir, fmt.Sprintf("golden_loop_%s_preisach.json", safeName))
		writeJSON(pFile, goldenP)
		fmt.Printf("  ✓ Preisach: %s\n", filepath.Base(pFile))
		
		fmt.Printf("  ○ LK: pending (see shared/physics/landau.go)\n")
	}
	
	fmt.Println("\nDone! Golden loops generated.")
}

func generatePreisachLoop(model *ferroelectric.PreisachModel, Emax float64, points int) ([]float64, []float64) {
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

func writeJSON(path string, data GoldenLoopData) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Printf("ERROR creating %s: %v\n", path, err)
		return
	}
	defer f.Close()
	
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		fmt.Printf("ERROR encoding %s: %v\n", path, err)
	}
}
