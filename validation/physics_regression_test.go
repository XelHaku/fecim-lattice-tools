package validation

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

const (
	physicsRegressionDataDir = "testdata/physics_regression"
	physicsRegressionVersion = "v1"
	updatePhysicsGoldenEnv    = "FECIM_UPDATE_PHYSICS_GOLDEN"
)

type curveReference struct {
	Version     string                 `json:"version"`
	Scenario    string                 `json:"scenario"`
	Description string                 `json:"description"`
	Generated   string                 `json:"generated_utc"`
	Parameters  map[string]interface{} `json:"parameters"`
	Data        struct {
		X []float64 `json:"x"`
		Y []float64 `json:"y"`
	} `json:"data"`
}

func rmsError(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}
	if len(a) == 0 {
		return 0
	}
	sum := 0.0
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(a)))
}

func maxAbsError(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.Inf(1)
	}
	maxErr := 0.0
	for i := range a {
		e := math.Abs(a[i] - b[i])
		if e > maxErr {
			maxErr = e
		}
	}
	return maxErr
}

func loadCurveReference(t *testing.T, path string) curveReference {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open golden reference %s: %v", path, err)
	}
	defer f.Close()

	var ref curveReference
	if err := json.NewDecoder(f).Decode(&ref); err != nil {
		t.Fatalf("decode golden reference %s: %v", path, err)
	}
	return ref
}

func saveCurveReference(path string, ref curveReference) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(ref, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func saveCurveCSV(path string, x, y []float64) error {
	if len(x) != len(y) {
		return fmt.Errorf("x/y length mismatch: %d vs %d", len(x), len(y))
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write([]string{"x", "y"}); err != nil {
		return err
	}
	for i := range x {
		row := []string{
			strconvFormatFloat(x[i]),
			strconvFormatFloat(y[i]),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

func strconvFormatFloat(v float64) string {
	// Use a stable representation suitable for diffs.
	return fmt.Sprintf("%.12e", v)
}

func generatePreisachLoop() (xE, yP []float64, params map[string]interface{}) {
	mat := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(mat)
	Emax := 2.0 * mat.Ec
	points := 201
	xE, yP = model.GetHysteresisLoop(Emax, points)
	params = map[string]interface{}{
		"material": "DefaultHZO (module1 PreisachModel)",
		"Emax":     Emax,
		"points":   points,
		"Ec":       mat.Ec,
		"Ps":       mat.Ps,
		"Pr":       mat.Pr,
	}
	return xE, yP, params
}

func generateLKLoop() (xE, yP []float64, params map[string]interface{}) {
	s := physics.NewLKSolver()
	s.EnableNoise = false
	s.UseNLS = false

	// Deterministic loop: triangular E(t) waveform.
	Emax := 2.5e9 // V/m (25 MV/cm)
	stepsPerHalf := 400
	dt := 2e-11 // seconds

	// Start at negative saturation.
	s.P = -math.Abs(s.PMax)
	s.Time = 0

	// Generate E sequence: -Emax -> +Emax -> -Emax (inclusive endpoints).
	xE = make([]float64, 0, 2*stepsPerHalf+1)
	yP = make([]float64, 0, 2*stepsPerHalf+1)

	for i := 0; i <= stepsPerHalf; i++ {
		E := -Emax + (2*Emax)*float64(i)/float64(stepsPerHalf)
		s.Step(E, dt)
		xE = append(xE, E)
		yP = append(yP, s.P)
	}
	for i := 1; i <= stepsPerHalf; i++ {
		E := Emax - (2*Emax)*float64(i)/float64(stepsPerHalf)
		s.Step(E, dt)
		xE = append(xE, E)
		yP = append(yP, s.P)
	}

	params = map[string]interface{}{
		"solver":        "shared/physics.LKSolver",
		"Emax":          Emax,
		"stepsPerHalf":  stepsPerHalf,
		"dt_seconds":    dt,
		"enableNoise":   s.EnableNoise,
		"useNLS":        s.UseNLS,
		"beta":          s.Beta,
		"gamma":         s.Gamma,
		"rho":           s.Rho,
		"q12":           s.Q12,
		"stress_pa":     s.Stress,
		"k_dep":         s.K_dep,
		"seriesR_ohm":   s.SeriesResistance,
		"thickness_m":   s.Thickness,
		"area_m2":       s.Area,
		"curieTemp_K":   s.CurieTemp,
		"curieConst_K":  s.CurieConst,
		"initialP":      -math.Abs(s.PMax),
		"pMax":          s.PMax,
		"useEffVisc":    s.UseEffectiveViscosity,
		"useMatAlpha":   s.UseMaterialAlpha,
		"temperature_K": s.Temperature,
	}
	return xE, yP, params
}

func TestPhysicsRegressionCurves(t *testing.T) {
	update := os.Getenv(updatePhysicsGoldenEnv) != ""

	tests := []struct {
		scenario     string
		jsonFile     string
		csvFile      string
		generate     func() (x, y []float64, params map[string]interface{})
		xTolMaxAbs   func(params map[string]interface{}) float64
		yTolRMS      func(params map[string]interface{}) float64
		yTolMaxAbs   func(params map[string]interface{}) float64
		minPoints    int
		assumptions  string
	}{
		{
			scenario: "preisach_loop_default_hzo",
			jsonFile: filepath.Join(physicsRegressionDataDir, "preisach_loop_default_hzo.json"),
			csvFile:  filepath.Join(physicsRegressionDataDir, "preisach_loop_default_hzo.csv"),
			generate: generatePreisachLoop,
			xTolMaxAbs: func(params map[string]interface{}) float64 {
				// E sampling should be stable; allow tiny numeric jitter.
				return 1e-12
			},
			yTolRMS: func(params map[string]interface{}) float64 {
				ps, _ := params["Ps"].(float64)
				return 0.02 * ps
			},
			yTolMaxAbs: func(params map[string]interface{}) float64 {
				ps, _ := params["Ps"].(float64)
				return 0.03 * ps
			},
			minPoints: 200,
			assumptions: "PreisachModel hysteresis loop is deterministic for DefaultHZO and fixed point count.",
		},
		{
			scenario: "lk_loop_default",
			jsonFile: filepath.Join(physicsRegressionDataDir, "lk_loop_default.json"),
			csvFile:  filepath.Join(physicsRegressionDataDir, "lk_loop_default.csv"),
			generate: generateLKLoop,
			xTolMaxAbs: func(params map[string]interface{}) float64 {
				return 1e-9 // V/m
			},
			yTolRMS: func(params map[string]interface{}) float64 {
				pMax, _ := params["pMax"].(float64)
				return 0.03 * pMax
			},
			yTolMaxAbs: func(params map[string]interface{}) float64 {
				pMax, _ := params["pMax"].(float64)
				return 0.05 * pMax
			},
			minPoints: 700,
			assumptions: "LK loop uses a fixed triangular E(t) waveform and fixed dt with noise/NLS disabled; output should be stable across platforms for the same solver code.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.scenario, func(t *testing.T) {
			x, y, params := tc.generate()
			if len(x) < tc.minPoints || len(y) < tc.minPoints {
				t.Fatalf("generated curve too short: len=%d", len(x))
			}
			if len(x) != len(y) {
				t.Fatalf("generated x/y length mismatch: %d vs %d", len(x), len(y))
			}

			if update {
				ref := curveReference{
					Version:     physicsRegressionVersion,
					Scenario:    tc.scenario,
					Description: tc.assumptions,
					Generated:   time.Now().UTC().Format(time.RFC3339),
					Parameters:  params,
				}
				ref.Data.X = x
				ref.Data.Y = y

				if err := saveCurveReference(tc.jsonFile, ref); err != nil {
					t.Fatalf("save golden json: %v", err)
				}
				if err := saveCurveCSV(tc.csvFile, x, y); err != nil {
					t.Fatalf("save golden csv: %v", err)
				}
				t.Logf("updated golden curve: %s (set %s= to disable)", tc.jsonFile, updatePhysicsGoldenEnv)
				return
			}

			golden := loadCurveReference(t, tc.jsonFile)
			if golden.Version != physicsRegressionVersion {
				t.Fatalf("golden version mismatch: got %q want %q", golden.Version, physicsRegressionVersion)
			}
			if golden.Scenario != tc.scenario {
				t.Fatalf("golden scenario mismatch: got %q want %q", golden.Scenario, tc.scenario)
			}

			xErrMax := maxAbsError(x, golden.Data.X)
			yErrRMS := rmsError(y, golden.Data.Y)
			yErrMax := maxAbsError(y, golden.Data.Y)

			xTol := tc.xTolMaxAbs(params)
			yTolRMS := tc.yTolRMS(params)
			yTolMax := tc.yTolMaxAbs(params)

			if xErrMax > xTol {
				t.Errorf("x max|err| too large: %e (tol %e)", xErrMax, xTol)
			}
			if yErrRMS > yTolRMS {
				t.Errorf("y RMS err too large: %e (tol %e)", yErrRMS, yTolRMS)
			}
			if yErrMax > yTolMax {
				t.Errorf("y max|err| too large: %e (tol %e)", yErrMax, yTolMax)
			}

			t.Logf("curve ok: xMaxErr=%e yRMSErr=%e yMaxErr=%e", xErrMax, yErrRMS, yErrMax)
		})
	}
}
