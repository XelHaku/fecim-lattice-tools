package ferroelectric

import (
	"math"
	"testing"

	sharedphysics "fecim-lattice-tools/shared/physics"
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

func TestPreisach_TemperatureSweep(t *testing.T) {
	material := sharedphysics.MaterlikHfO2()
	model := NewPreisachModel(material)
	model.SetStress(0.0) // isolate temperature effect

	model.SetTemperature(250.0)
	ec250 := model.GetEffectiveEc()
	model.SetTemperature(300.0)
	ec300 := model.GetEffectiveEc()
	model.SetTemperature(350.0)
	ec350 := model.GetEffectiveEc()

	if !(ec250 > ec300 && ec300 > ec350) {
		t.Fatalf("expected Ec to decrease as temperature approaches Tc: Ec250=%.6e Ec300=%.6e Ec350=%.6e", ec250, ec300, ec350)
	}
}

func TestPreisach_StressEffect(t *testing.T) {
	material := sharedphysics.MaterlikHfO2()
	model := NewPreisachModel(material)
	model.SetTemperature(300.0)
	model.SetStress(0.0)
	ecNoStress := model.GetEffectiveEc()
	model.SetStress(1.0)
	ecWithStress := model.GetEffectiveEc()

	if math.Abs(ecWithStress-ecNoStress) <= math.Abs(ecNoStress)*1e-6 {
		t.Fatalf("expected stress to modify Ec via electrostriction: Ec(0GPa)=%.6e Ec(1GPa)=%.6e", ecNoStress, ecWithStress)
	}
}

func TestPreisach_LK_TemperatureConsistency(t *testing.T) {
	material := sharedphysics.MaterlikHfO2()
	model := NewPreisachModel(material)
	model.SetStress(0.0)

	temps := []float64{250.0, 300.0, 350.0}
	prevPreisachEc := math.Inf(1)
	prevLKEc := math.Inf(1)

	for _, temp := range temps {
		model.SetTemperature(temp)
		preisachEc := model.GetEffectiveEc()
		lkEc := material.CoerciveFieldAtTemp(temp)

		if !(preisachEc < prevPreisachEc) {
			t.Fatalf("preisach Ec trend mismatch at T=%.1fK: prev=%.6e curr=%.6e", temp, prevPreisachEc, preisachEc)
		}
		if lkEc > 0 && !(lkEc < prevLKEc) {
			t.Fatalf("LK Ec trend mismatch at T=%.1fK: prev=%.6e curr=%.6e", temp, prevLKEc, lkEc)
		}

		prevPreisachEc = preisachEc
		if lkEc > 0 {
			prevLKEc = lkEc
		}
	}
}
