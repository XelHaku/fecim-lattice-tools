package ferroelectric

import (
	"math"
	"testing"
)

func TestTanhEverett_SymmetricPairMatchesProductForm(t *testing.T) {
	everett := &TanhEverett{
		Ps:    0.30,
		Ec:    1.0,
		Delta: 0.2,
	}

	for _, x := range []float64{0, 0.2, 0.6, 1.0, 3.0} {
		alpha := everett.Ec + x
		beta := -everett.Ec - x

		got := everett.Calculate(alpha, beta)
		// Product-form identity for symmetric pairs (α=Ec+x, β=-Ec-x):
		//   E(α,β) = [1+tanh(x/Δ)]² · Ps/4
		tanhVal := math.Tanh(x / everett.Delta)
		expected := (1.0 + tanhVal) * (1.0 + tanhVal) * everett.Ps / 4.0

		if math.Abs(got-expected) > math.Abs(expected)*1e-8+1e-12 {
			t.Fatalf("x=%.3f: got %.12f, expected %.12f", x, got, expected)
		}
	}
}

func TestPreisachModel_SetTemperatureUpdatesEverett(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	tempK := 400.0
	model.SetTemperature(tempK)

	deltaT := tempK - 300.0
	expectedEc := material.Ec + material.TempCoeffEc*deltaT
	expectedPs := material.Ps + material.TempCoeffPr*deltaT
	if expectedEc < 1e5 {
		expectedEc = 1e5
	}
	if expectedPs < 1e-6 {
		expectedPs = 1e-6
	}

	gotEc := model.GetEffectiveEc()
	if math.Abs(gotEc-expectedEc) > math.Abs(expectedEc)*1e-8+1e-12 {
		t.Fatalf("Ec mismatch: got %.6e, expected %.6e", gotEc, expectedEc)
	}

	gotPs := model.everett.Ps
	const epsilon0 = 8.854e-12
	chi := 0.0
	if material.EpsilonLF > 1 {
		chi = epsilon0 * (material.EpsilonLF - 1)
	} else if material.Epsilon > 1 {
		chi = epsilon0 * (material.Epsilon - 1)
	}
	revPSat := chi * expectedEc
	expectedIrrev := expectedPs - revPSat
	if expectedIrrev < 0 {
		expectedIrrev = 0
	}
	if math.Abs(gotPs-expectedIrrev) > math.Abs(expectedIrrev)*1e-8+1e-12 {
		t.Fatalf("Ps mismatch: got %.6e, expected %.6e", gotPs, expectedIrrev)
	}

	expectedDelta := tuneDeltaForPr(expectedEc, model.stack.SaturationE, expectedIrrev, material.Pr)
	if math.Abs(model.everett.Delta-expectedDelta) > math.Abs(expectedDelta)*1e-8+1e-12 {
		t.Fatalf("Delta mismatch: got %.6e, expected %.6e", model.everett.Delta, expectedDelta)
	}
}

func TestPreisachModel_SetStressUpdatesEverett(t *testing.T) {
	material := DefaultHZO()
	model := NewPreisachModel(material)

	model.SetStress(2.0)

	expectedEc := material.Ec * (1.0 + 0.05*(2.0-1.0))
	if expectedEc < 1e5 {
		expectedEc = 1e5
	}
	expectedPs := material.Ps
	if expectedPs < 1e-6 {
		expectedPs = 1e-6
	}

	gotEc := model.GetEffectiveEc()
	if math.Abs(gotEc-expectedEc) > math.Abs(expectedEc)*1e-8+1e-12 {
		t.Fatalf("Ec mismatch: got %.6e, expected %.6e", gotEc, expectedEc)
	}

	gotPs := model.everett.Ps
	const epsilon0 = 8.854e-12
	chi := 0.0
	if material.EpsilonLF > 1 {
		chi = epsilon0 * (material.EpsilonLF - 1)
	} else if material.Epsilon > 1 {
		chi = epsilon0 * (material.Epsilon - 1)
	}
	revPSat := chi * expectedEc
	expectedIrrev := expectedPs - revPSat
	if expectedIrrev < 0 {
		expectedIrrev = 0
	}
	if math.Abs(gotPs-expectedIrrev) > math.Abs(expectedIrrev)*1e-8+1e-12 {
		t.Fatalf("Ps mismatch: got %.6e, expected %.6e", gotPs, expectedIrrev)
	}
}
