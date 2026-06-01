package external_test

// TestMVMNumpyCrossValidation independently solves the same crossbar network
// using scipy.linalg.solve (exact LU) and compares against Go MNA solver.
//
// This is a real external reference cross-validation:
//   - Go:    referenceSolveDense (Gaussian elimination, our MNA stamping)
//   - Python: scipy.linalg.solve on the identical MNA matrix, stamped in numpy
//
// If both implementations stamp the same matrix correctly, results must agree
// within floating-point tolerance (~1e-10 V).
// Disagreement reveals a stamping bug in either Go or Python.
//
// Literature basis for test parameters:
//   - Wire resistance: SKY130 Cu-like 2.2 nΩ·m, 80nm×160nm cross-section
//   - Drive resistance: 1x wire segment (Thevenin model)
//   - Array topologies: 4×4, 8×8, 16×16

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strings"
	"testing"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/validation/external/internal/testsupport"
)

// mnaScript builds the exact same MNA stamping as referenceSolveDense and
// solves using scipy.linalg.solve. Unknowns: [WL_nodes | BL_nodes], rows*cols each.
const mnaScript = `
import numpy as np
from scipy import linalg
import sys, json

d      = json.loads(sys.stdin.read())
rows   = d["rows"]
cols   = d["cols"]
G      = np.array(d["G"])
VWL    = np.array(d["VWL"])
VBL    = np.array(d["VBL"])
gWL    = d["gWL"]      # 1/RWordLine
gBL    = d["gBL"]      # 1/RBitLine
gWLD   = d["gWLD"]     # 1/WLDriveResistance
gBLD   = d["gBLD"]     # 1/BLDriveResistance

n = 2 * rows * cols
A = np.zeros((n, n))
b = np.zeros(n)

def iW(r, c): return r * cols + c
def iB(r, c): return rows * cols + r * cols + c

def stamp_conductance(i, j, g):
    A[i, i] += g; A[j, j] += g; A[i, j] -= g; A[j, i] -= g

def stamp_source(i, g, v):
    A[i, i] += g; b[i] += g * v

for r in range(rows):
    for c in range(cols):
        w  = iW(r, c)
        bl = iB(r, c)

        # WL wire segments
        if c > 0:
            stamp_conductance(w, iW(r, c-1), gWL)
        if c == 0:
            stamp_source(w, gWLD, float(VWL[r]))

        # BL wire segments
        if r > 0:
            stamp_conductance(bl, iB(r-1, c), gBL)
        if r == 0:
            stamp_source(bl, gBLD, float(VBL[c]))

        # Cell conductance
        gcell = float(G[r, c])
        stamp_conductance(w, bl, gcell)

x = linalg.solve(A, b)

V_cell = [[0.0]*cols for _ in range(rows)]
I_cell = [[0.0]*cols for _ in range(rows)]
for r in range(rows):
    for c in range(cols):
        vw = x[iW(r, c)]
        vb = x[iB(r, c)]
        V_cell[r][c] = vw - vb
        I_cell[r][c] = float(G[r, c]) * V_cell[r][c]

print(json.dumps({"V_cell": V_cell, "I_cell": I_cell}))
`

func TestMVMNumpyCrossValidation(t *testing.T) {
	testsupport.RequireCommand(t, "python3", "python3 not installed")
	// Verify scipy available
	testsupport.RequirePythonModule(t, "scipy.linalg", "scipy not installed: run pip3 install scipy")

	// Compute default wire params (mirror WithDefaults logic)
	geom := arraysim.DefaultCellGeometry()
	crossSection := geom.WireWidth * geom.WireThickness
	rPerMeter := geom.MetalResistivity / crossSection
	RWL := rPerMeter * geom.PitchX // Ω/cell WL
	RBL := rPerMeter * geom.PitchY // Ω/cell BL

	t.Logf("Wire params: RWL=%.4f Ω, RBL=%.4f Ω", RWL, RBL)
	t.Logf("Drive params: WLDrive=%.4f Ω, BLDrive=%.4f Ω", RWL, RBL)

	sizes := []int{4, 8, 16}
	G_base := 20e-6 // 20 μS baseline conductance

	type cellErr struct{ vErr, iErr float64 }

	for _, n := range sizes {
		t.Run(fmt.Sprintf("%dx%d", n, n), func(t *testing.T) {
			// Deterministic conductance matrix
			G := make([][]float64, n)
			for r := 0; r < n; r++ {
				G[r] = make([]float64, n)
				for c := 0; c < n; c++ {
					// Vary conductance across cells (realistic spread)
					G[r][c] = G_base * (1 + 0.05*float64(r*n+c)/float64(n*n))
				}
			}

			VWL := make([]float64, n)
			VBL := make([]float64, n)
			for i := 0; i < n; i++ {
				VWL[i] = 0.5 - 0.1*float64(i)/float64(n)
				VBL[i] = -0.05 * float64(i) / float64(n)
			}

			// --- Go solver ---
			params := arraysim.SolveParams{
				WLVoltages:  VWL,
				BLVoltages:  VBL,
				Conductance: G,
				Geometry:    arraysim.DefaultCellGeometry(),
				Wire:        arraysim.WireParams{RWordLine: RWL, RBitLine: RBL},
				// BoundaryParams zero → WithDefaults → WLDrive=RWL, BLDrive=RBL
			}
			solver := arraysim.NewTierASolver()
			goResult, err := solver.Solve(params)
			if err != nil {
				t.Fatalf("Go solver error: %v", err)
			}

			// --- Python/scipy reference ---
			input := map[string]interface{}{
				"rows": n, "cols": n,
				"G": G, "VWL": VWL, "VBL": VBL,
				"gWL": 1.0 / RWL, "gBL": 1.0 / RBL,
				"gWLD": 1.0 / RWL, "gBLD": 1.0 / RBL, // drive = wire
			}
			inputJSON, _ := json.Marshal(input)

			cmd := exec.Command("python3", "-c", mnaScript)
			cmd.Stdin = strings.NewReader(string(inputJSON))
			outBytes, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("Python solver error: %v\noutput: %s", err, outBytes)
			}

			var pyResult struct {
				VCell [][]float64 `json:"V_cell"`
				ICell [][]float64 `json:"I_cell"`
			}
			if err := json.Unmarshal(outBytes, &pyResult); err != nil {
				t.Fatalf("Parse python output: %v\nraw: %s", err, outBytes)
			}

			// --- Compare every cell ---
			maxVErr := 0.0
			maxIErr := 0.0
			worstR, worstC := 0, 0
			for r := 0; r < n; r++ {
				for c := 0; c < n; c++ {
					vErr := math.Abs(goResult.CellVoltages[r][c] - pyResult.VCell[r][c])
					iErr := math.Abs(goResult.CellCurrents[r][c] - pyResult.ICell[r][c])
					if vErr > maxVErr {
						maxVErr = vErr
						worstR, worstC = r, c
					}
					if iErr > maxIErr {
						maxIErr = iErr
					}
				}
			}

			t.Logf("Size %dx%d: maxVErr=%.2e V, maxIErr=%.2e A, worst_cell=(%d,%d)",
				n, n, maxVErr, maxIErr, worstR, worstC)
			t.Logf("  Go V[%d][%d]=%.8f, Py V[%d][%d]=%.8f",
				worstR, worstC, goResult.CellVoltages[worstR][worstC],
				worstR, worstC, pyResult.VCell[worstR][worstC])

			// Hard thresholds: both scipy and Go solve the same matrix with partial pivoting
			// so residuals should be ~eps * cond(A) * ||x|| ≈ 1e-10 V for small arrays
			vThresh := 1e-8  // 10 nV — generous for float64 precision
			iThresh := 1e-12 // 1 pA — generous for float64 precision

			if maxVErr > vThresh {
				t.Errorf("FAIL voltage cross-val %dx%d: maxVErr=%.4e V > %.0e V threshold", n, n, maxVErr, vThresh)
			}
			if maxIErr > iThresh {
				t.Errorf("FAIL current cross-val %dx%d: maxIErr=%.4e A > %.0e A threshold", n, n, maxIErr, iThresh)
			}

			testsupport.WriteExternalArtifact(t, fmt.Sprintf("mvm_numpy_crossval_%dx%d.json", n, n), map[string]interface{}{
				"n": n, "RWL_ohm": RWL, "RBL_ohm": RBL,
				"maxVErr_V": maxVErr, "maxIErr_A": maxIErr,
				"pass_V": maxVErr <= vThresh, "pass_I": maxIErr <= iThresh,
			})
		})
	}
}
