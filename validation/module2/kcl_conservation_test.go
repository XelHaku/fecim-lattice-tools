package module2

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"fecim-lattice-tools/shared/crossbar"
	sharedval "fecim-lattice-tools/shared/validation"
)

const (
	kclValidationCases     = 100
	kclValidationSeed      = int64(20260426)
	kclValidationThreshold = 1e-9
)

type kclValidationReport struct {
	sharedval.ArtifactEnvelope

	Description     string            `json:"description"`
	Cases           int               `json:"cases"`
	Seed            int64             `json:"seed"`
	ThresholdA      float64           `json:"threshold_A"`
	MaxKCLResidualA float64           `json:"max_kcl_residual_A"`
	MaxSolverErrorV float64           `json:"max_solver_error_V"`
	WorstCase       kclValidationCase `json:"worst_case"`
	Pass            bool              `json:"pass"`
}

type kclValidationCase struct {
	Index           int     `json:"index"`
	Rows            int     `json:"rows"`
	Cols            int     `json:"cols"`
	RpRowNormalized float64 `json:"rp_row_normalized"`
	RpColNormalized float64 `json:"rp_col_normalized"`
	MinConductanceS float64 `json:"min_conductance_S"`
	MaxConductanceS float64 `json:"max_conductance_S"`
	MaxVoltageV     float64 `json:"max_voltage_V"`
	Iterations      int     `json:"iterations"`
	SolverErrorV    float64 `json:"solver_error_V"`
}

func TestModule2KCLConservation_PublicValidation(t *testing.T) {
	report := runKCLConservationSuite(t)

	if !report.Pass {
		t.Fatalf("KCL validation failed: max residual %.3e A, threshold %.3e A", report.MaxKCLResidualA, report.ThresholdA)
	}
	if report.Cases != 100 {
		t.Fatalf("unexpected validation case count: got %d, want 100", report.Cases)
	}
}

func runKCLConservationSuite(t *testing.T) kclValidationReport {
	t.Helper()

	rng := rand.New(rand.NewSource(kclValidationSeed))
	report := kclValidationReport{
		ArtifactEnvelope: sharedval.NewEnvelope("RG-VAL-M2-01", "", false),
		Description:      "Module 2 crossbar KCL conservation over deterministic random parasitic-array cases",
		Cases:            kclValidationCases,
		Seed:             kclValidationSeed,
		ThresholdA:       kclValidationThreshold,
	}

	for idx := 0; idx < kclValidationCases; idx++ {
		rows := 2 + rng.Intn(11)
		cols := 2 + rng.Intn(11)
		rpRow := 0.25 + rng.Float64()*4.75
		rpCol := 0.25 + rng.Float64()*4.75
		minG := 10e-6
		maxG := 100e-6

		conductances := randomConductances(rows, cols, minG, maxG, rng)
		applied := randomAppliedVoltages(cols, 0.05, 0.30, rng)

		cfg := crossbar.DefaultSORConfig()
		cfg.MaxIterations = 300
		solver, err := crossbar.NewParasiticSolver(rows, cols, cfg)
		if err != nil {
			t.Fatalf("case %d NewParasiticSolver: %v", idx, err)
		}
		solver.SetParasitics(rpRow, rpCol)
		solver.SetConductances(conductances)

		result, err := solver.SolveMVMWithFallback(applied)
		if err != nil && (result == nil || !result.Converged) {
			t.Fatalf("case %d SolveMVMWithFallback: %v", idx, err)
		}

		maxKCL := maxKCLResidual(result)
		if maxKCL > report.MaxKCLResidualA {
			report.MaxKCLResidualA = maxKCL
			report.WorstCase = kclValidationCase{
				Index:           idx,
				Rows:            rows,
				Cols:            cols,
				RpRowNormalized: rpRow,
				RpColNormalized: rpCol,
				MinConductanceS: minG,
				MaxConductanceS: maxG,
				MaxVoltageV:     maxValue(applied),
				Iterations:      result.Iterations,
				SolverErrorV:    result.MaxError,
			}
		}
		if result.MaxError > report.MaxSolverErrorV {
			report.MaxSolverErrorV = result.MaxError
		}
	}

	report.Pass = report.MaxKCLResidualA < report.ThresholdA
	report.ArtifactEnvelope = sharedval.NewEnvelope("RG-VAL-M2-01", "", report.Pass)
	writeKCLReport(t, report)

	t.Logf("Module 2 KCL validation: cases=%d max_residual=%.3e A threshold=%.3e A artifact=output/validation/module2/kcl_conservation.json",
		report.Cases, report.MaxKCLResidualA, report.ThresholdA)

	return report
}

func maxKCLResidual(result *crossbar.ParasiticMVMResult) float64 {
	rows := len(result.DeviceCurrents)
	cols := len(result.DeviceCurrents[0])
	rowSums := make([][]float64, rows)
	colSums := make([][]float64, rows)

	for i := 0; i < rows; i++ {
		rowSums[i] = make([]float64, cols)
		rowSums[i][cols-1] = result.DeviceCurrents[i][cols-1]
		for j := cols - 2; j >= 0; j-- {
			rowSums[i][j] = rowSums[i][j+1] + result.DeviceCurrents[i][j]
		}
	}

	for i := 0; i < rows; i++ {
		colSums[i] = make([]float64, cols)
	}
	for j := 0; j < cols; j++ {
		colSums[rows-1][j] = result.DeviceCurrents[rows-1][j]
		for i := rows - 2; i >= 0; i-- {
			colSums[i][j] = colSums[i+1][j] + result.DeviceCurrents[i][j]
		}
	}

	maxResidual := 0.0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols-1; j++ {
			maxResidual = maxAbs(maxResidual, rowSums[i][j]-rowSums[i][j+1]-result.DeviceCurrents[i][j])
		}
		maxResidual = maxAbs(maxResidual, rowSums[i][cols-1]-result.DeviceCurrents[i][cols-1])
	}

	for j := 0; j < cols; j++ {
		for i := 0; i < rows-1; i++ {
			maxResidual = maxAbs(maxResidual, colSums[i][j]-colSums[i+1][j]-result.DeviceCurrents[i][j])
		}
		maxResidual = maxAbs(maxResidual, colSums[rows-1][j]-result.DeviceCurrents[rows-1][j])
		maxResidual = maxAbs(maxResidual, colSums[0][j]-result.OutputCurrents[j])
	}

	return maxResidual
}

func randomConductances(rows, cols int, minG, maxG float64, rng *rand.Rand) [][]float64 {
	out := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		out[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			out[i][j] = minG + rng.Float64()*(maxG-minG)
		}
	}
	return out
}

func randomAppliedVoltages(cols int, minV, maxV float64, rng *rand.Rand) []float64 {
	out := make([]float64, cols)
	for j := 0; j < cols; j++ {
		out[j] = minV + rng.Float64()*(maxV-minV)
	}
	return out
}

func writeKCLReport(t *testing.T, report kclValidationReport) {
	t.Helper()

	root := repoRoot(t)
	outDir := filepath.Join(root, "output", "validation", "module2")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("create validation output directory: %v", err)
	}

	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		t.Fatalf("marshal KCL report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(outDir, "kcl_conservation.json"), append(b, '\n'), 0o644); err != nil {
		t.Fatalf("write KCL report: %v", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not find repository root from %s", dir)
		}
		dir = parent
	}
}

func maxValue(values []float64) float64 {
	out := math.Inf(-1)
	for _, v := range values {
		if v > out {
			out = v
		}
	}
	return out
}

func maxAbs(current, candidate float64) float64 {
	if math.Abs(candidate) > current {
		return math.Abs(candidate)
	}
	return current
}
