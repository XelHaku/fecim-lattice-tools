package validation

import (
	"encoding/csv"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/controller"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/crossbar"
	"fecim-lattice-tools/shared/physics"
)

func TestDeterminism_PreisachSameSeedIdenticalOutput(t *testing.T) {
	mat := ferroelectric.DefaultHZO()

	run := func(seed int64) []float64 {
		r := rand.New(rand.NewSource(seed))
		m := ferroelectric.NewPreisachModel(mat)
		m.Reset()
		out := make([]float64, 128)
		for i := range out {
			field := (r.Float64()*2 - 1) * (2.5 * mat.Ec)
			out[i] = m.Update(field)
		}
		return out
	}

	p1 := run(20260213)
	p2 := run(20260213)

	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("Preisach output differs for same seed")
	}
}

func TestDeterminism_LKSameSeedIdenticalOutput(t *testing.T) {
	run := func(seed int64) []float64 {
		r := rand.New(rand.NewSource(seed))
		s := physics.NewLKSolver()
		s.EnableNoise = false
		s.UseNLS = false
		s.P = -math.Abs(s.PMax)
		out := make([]float64, 128)
		for i := range out {
			E := (r.Float64()*2 - 1) * 2.0e9
			s.Step(E, 2e-11)
			out[i] = s.P
		}
		return out
	}

	p1 := run(20260213)
	p2 := run(20260213)

	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("LK output differs for same seed")
	}
}

func TestDeterminism_NoNaNOrInfInOutputs(t *testing.T) {
	mat := ferroelectric.DefaultHZO()
	m := ferroelectric.NewPreisachModel(mat)
	m.Reset()
	for i := 0; i < 256; i++ {
		field := (-2.5 + 5.0*float64(i)/255.0) * mat.Ec
		p := m.Update(field)
		if math.IsNaN(p) || math.IsInf(p, 0) {
			t.Fatalf("Preisach produced invalid value at step %d: %v", i, p)
		}
	}

	s := physics.NewLKSolver()
	s.EnableNoise = false
	s.UseNLS = false
	s.P = -math.Abs(s.PMax)
	for i := 0; i < 256; i++ {
		field := (-2.0 + 4.0*float64(i)/255.0) * 2.0e9
		s.Step(field, 2e-11)
		if math.IsNaN(s.P) || math.IsInf(s.P, 0) {
			t.Fatalf("LK produced invalid value at step %d: %v", i, s.P)
		}
	}
}

func TestDeterminism_CSVSchemaStable(t *testing.T) {
	csvPath := filepath.Join("testdata", "physics_regression", "preisach_loop_default_hzo.csv")
	f, err := os.Open(csvPath)
	if err != nil {
		t.Fatalf("open CSV fixture: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	headers, err := r.Read()
	if err != nil {
		t.Fatalf("read CSV header: %v", err)
	}

	expectedHeaders := []string{"x", "y"}
	if !reflect.DeepEqual(headers, expectedHeaders) {
		t.Fatalf("unexpected CSV headers: got=%v want=%v", headers, expectedHeaders)
	}
	if len(headers) != 2 {
		t.Fatalf("unexpected CSV column count: got=%d want=2", len(headers))
	}

	for row := 2; ; row++ {
		rec, err := r.Read()
		if err != nil {
			break
		}
		if len(rec) != len(expectedHeaders) {
			t.Fatalf("CSV schema drift at row %d: got %d columns, want %d", row, len(rec), len(expectedHeaders))
		}
	}
}

func TestDeterminism_CrossbarSameConfigAndInputSameMVMOutput(t *testing.T) {
	cfg := &crossbar.Config{Rows: 4, Cols: 4, NoiseLevel: 0, ADCBits: 16, DACBits: 16}
	weights := [][]float64{{0.10, 0.25, 0.70, 0.90}, {0.60, 0.30, 0.20, 0.40}, {0.80, 0.55, 0.35, 0.15}, {0.05, 0.45, 0.65, 0.95}}
	input := []float64{0.85, 0.10, 0.55, 0.35}

	run := func() []float64 {
		arr, err := crossbar.NewArray(cfg)
		if err != nil {
			t.Fatalf("NewArray failed: %v", err)
		}
		if err := arr.ProgramWeightMatrix(weights); err != nil {
			t.Fatalf("ProgramWeightMatrix failed: %v", err)
		}
		out, err := arr.MVM(input)
		if err != nil {
			t.Fatalf("MVM failed: %v", err)
		}
		return out
	}

	out1 := run()
	out2 := run()
	if !reflect.DeepEqual(out1, out2) {
		t.Fatalf("non-deterministic MVM output")
	}
}

func TestDeterminism_ISPPSameParametersSamePulseCount(t *testing.T) {
	run := func() (pulses int, finalLevel int, state controller.WriteState) {
		mat := ferroelectric.LiteratureSuperlattice()
		model := ferroelectric.NewPreisachModel(mat)
		model.Reset()

		const numLevels = physics.DefaultLevels
		const targetLevel = 20
		wc := controller.NewWriteController(numLevels, mat.Ec, mat.Ec*2.5, nil)
		wc.PulseDuration = 5e-4
		wc.MaxRetries = physics.DefaultLevels
		wc.Start(targetLevel, true)

		currentField := 0.0
		p := -mat.Ps
		level := determinismLevelFromP(p, mat.Ps, numLevels)

		const maxIters = 40000
		const dt = 5e-6
		for i := 0; i < maxIters; i++ {
			curLevel := determinismLevelFromP(p, mat.Ps, numLevels)
			nextField, done := wc.Update(dt, currentField, curLevel, 0)
			currentField = nextField
			p = model.Update(currentField)
			level = determinismLevelFromP(p, mat.Ps, numLevels)
			if done {
				break
			}
		}

		return wc.TotalPulses + wc.PulseCount, level, wc.State
	}

	p1, l1, s1 := run()
	p2, l2, s2 := run()

	if s1 != controller.StateSuccess || s2 != controller.StateSuccess {
		t.Fatalf("ISPP did not converge in one or both runs: state1=%v state2=%v", s1, s2)
	}
	if l1 != l2 || p1 != p2 {
		t.Fatalf("ISPP determinism mismatch: pulses(%d,%d) level(%d,%d)", p1, p2, l1, l2)
	}
}

func determinismLevelFromP(P, effPs float64, numLevels int) int {
	if numLevels <= 1 {
		return 1
	}
	if effPs == 0 {
		effPs = 1
	}
	n := P / effPs
	if n > 1 {
		n = 1
	}
	if n < -1 {
		n = -1
	}
	maxLevel := numLevels - 1
	lvl0 := int(math.Round((n + 1) / 2 * float64(maxLevel)))
	if lvl0 < 0 {
		lvl0 = 0
	}
	if lvl0 > maxLevel {
		lvl0 = maxLevel
	}
	return lvl0 + 1
}
