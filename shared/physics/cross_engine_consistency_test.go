package physics_test

import (
	"math"
	"testing"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/shared/physics"
)

func lkExtractPrEcAtTemp(t *testing.T, tempK float64) (pr, ec float64) {
	t.Helper()

	mat := physics.DefaultHZO()
	s := physics.NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false
	s.UseMaterialAlpha = false // ensure alpha(T) coupling is active
	s.Temperature = tempK
	s.UpdateParams()
	s.SetState(-math.Abs(mat.Pr))

	eMax := 3.0 * mat.Ec
	const (
		nPtsHalf      = 241
		stepsPerPoint = 400
		dt            = 2e-12
	)

	fields := make([]float64, 0, 2*nPtsHalf)
	pols := make([]float64, 0, 2*nPtsHalf)

	for i := 0; i < nPtsHalf; i++ {
		E := -eMax + (2*eMax*float64(i))/float64(nPtsHalf-1)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}
	for i := 0; i < nPtsHalf; i++ {
		E := eMax - (2*eMax*float64(i))/float64(nPtsHalf-1)
		for k := 0; k < stepsPerPoint; k++ {
			s.Step(E, dt)
		}
		fields = append(fields, E)
		pols = append(pols, s.GetState())
	}

	// Pr from E=0 crossings.
	var prVals []float64
	for i := 1; i < len(fields); i++ {
		if fields[i-1] == 0 {
			prVals = append(prVals, math.Abs(pols[i-1]))
			continue
		}
		if fields[i-1]*fields[i] <= 0 {
			dx := fields[i] - fields[i-1]
			if dx != 0 {
				f := -fields[i-1] / dx
				if f >= 0 && f <= 1 {
					p0 := pols[i-1] + f*(pols[i]-pols[i-1])
					prVals = append(prVals, math.Abs(p0))
				}
			}
		}
	}
	if len(prVals) == 0 {
		t.Fatalf("LK: failed to extract Pr at T=%.1fK", tempK)
	}
	for _, v := range prVals {
		pr += v
	}
	pr /= float64(len(prVals))

	// Ec from P=0 crossings.
	var ecVals []float64
	for i := 1; i < len(pols); i++ {
		if pols[i-1]*pols[i] <= 0 {
			dy := pols[i] - pols[i-1]
			if dy != 0 {
				f := -pols[i-1] / dy
				if f >= 0 && f <= 1 {
					ec0 := fields[i-1] + f*(fields[i]-fields[i-1])
					ecVals = append(ecVals, math.Abs(ec0))
				}
			}
		}
	}
	if len(ecVals) == 0 {
		t.Fatalf("LK: failed to extract Ec at T=%.1fK", tempK)
	}
	for _, v := range ecVals {
		ec += v
	}
	ec /= float64(len(ecVals))

	return pr, ec
}

func preisachPrEcAtTemp(tempK float64) (pr, ec float64) {
	mat := ferroelectric.DefaultHZO()
	model := ferroelectric.NewPreisachModel(mat)
	model.SetTemperature(tempK)

	satE := 5.0 * model.GetEffectiveEc()
	model.Reset()
	model.Update(-satE)
	model.Update(satE)
	pr = math.Abs(model.Update(0))
	ec = math.Abs(model.GetEffectiveEc())
	return pr, ec
}

func collectCrossEngineSweep(t *testing.T) (temps, prPreisach, prLK, ecPreisach, ecLK []float64) {
	t.Helper()
	for temp := 200.0; temp <= 450.0; temp += 25.0 {
		prP, ecP := preisachPrEcAtTemp(temp)
		prL, ecL := lkExtractPrEcAtTemp(t, temp)
		temps = append(temps, temp)
		prPreisach = append(prPreisach, prP)
		prLK = append(prLK, prL)
		ecPreisach = append(ecPreisach, ecP)
		ecLK = append(ecLK, ecL)
	}
	return
}

func TestCrossEngine_PrVsTemperature(t *testing.T) {
	temps, prPreisach, prLK, _, _ := collectCrossEngineSweep(t)
	for i := 1; i < len(temps); i++ {
		if prPreisach[i] > prPreisach[i-1]+1e-9 {
			t.Fatalf("Preisach Pr not monotonically decreasing: Tprev=%.1fK Prprev=%.6e, T=%.1fK Pr=%.6e", temps[i-1], prPreisach[i-1], temps[i], prPreisach[i])
		}
		if prLK[i] > prLK[i-1]+1e-9 {
			t.Fatalf("LK Pr not monotonically decreasing: Tprev=%.1fK Prprev=%.6e, T=%.1fK Pr=%.6e", temps[i-1], prLK[i-1], temps[i], prLK[i])
		}
	}
}

func TestCrossEngine_EcVsTemperature(t *testing.T) {
	temps, _, _, ecPreisach, ecLK := collectCrossEngineSweep(t)
	for i := 1; i < len(temps); i++ {
		if ecPreisach[i] > ecPreisach[i-1]+1e-6 {
			t.Fatalf("Preisach Ec not monotonically decreasing: Tprev=%.1fK Ecprev=%.6e, T=%.1fK Ec=%.6e", temps[i-1], ecPreisach[i-1], temps[i], ecPreisach[i])
		}
		if ecLK[i] > ecLK[i-1]+1e-6 {
			t.Fatalf("LK Ec not monotonically decreasing: Tprev=%.1fK Ecprev=%.6e, T=%.1fK Ec=%.6e", temps[i-1], ecLK[i-1], temps[i], ecLK[i])
		}
	}
}

func TestCrossEngine_PrMagnitudeAgreement(t *testing.T) {
	// Compare default engine operating points at room temperature.
	// For LK we keep default configured alpha mode (UseMaterialAlpha=true after ConfigureFromMaterial).
	prPreisach, _ := preisachPrEcAtTemp(300.0)

	mat := physics.DefaultHZO()
	s := physics.NewLKSolver()
	s.ConfigureFromMaterial(mat)
	s.UseNLS = false
	s.EnableNoise = false
	s.SetState(-math.Abs(mat.Pr))

	satE := 5.0 * mat.Ec
	for i := 0; i < 3000; i++ {
		s.Step(satE, 2e-12)
	}
	for i := 0; i < 3000; i++ {
		s.Step(0, 2e-12)
	}
	prLK := math.Abs(s.GetState())

	if prPreisach <= 0 || prLK <= 0 {
		t.Fatalf("invalid Pr values: Preisach=%.6e LK=%.6e", prPreisach, prLK)
	}

	relDiff := math.Abs(prPreisach-prLK) / math.Max(prPreisach, prLK)
	if relDiff > 0.30 {
		t.Fatalf("Pr mismatch exceeds 30%% at 300K: Preisach=%.6e LK=%.6e relDiff=%.2f%%", prPreisach, prLK, relDiff*100)
	}
}
