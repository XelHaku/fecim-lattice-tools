//go:build legacy_fyne

package gui

import (
	"math"
	"testing"
)

func TestWorkloadMACs(t *testing.T) {
	ca := &ComparisonApp{}
	tests := []struct {
		workload string
		want     int
	}{
		{"MNIST", 101632},
		{"ResNet-50", 4000000000},
		{"BERT-Base", 11000000000},
		{"GPT-2", 35000000000},
		{"LLM-70B", 140000000000000},
	}

	for _, tt := range tests {
		ca.currentWorkload = tt.workload
		got := ca.getWorkloadMACs()
		if got != tt.want {
			t.Errorf("workload %s: got %d MACs, want %d", tt.workload, got, tt.want)
		}
	}
}

func TestEnergyPowerCostFormula(t *testing.T) {
	macs := 101632
	energyFJ := cpuEnergyPJPerMAC * 1000
	inferences := 10000.0

	gotEnergy, gotPower, gotCost := energyPowerCost(macs, energyFJ, inferences)

	const (
		expEnergy = 101.632
		expPower  = 1.01632
		expCost   = 0.07419136
	)

	if math.Abs(gotEnergy-expEnergy) > 1e-6 {
		t.Errorf("energy µJ: got %.6f, want %.6f", gotEnergy, expEnergy)
	}
	if math.Abs(gotPower-expPower) > 1e-6 {
		t.Errorf("power W: got %.6f, want %.6f", gotPower, expPower)
	}
	if math.Abs(gotCost-expCost) > 1e-8 {
		t.Errorf("monthly cost: got %.8f, want %.8f", gotCost, expCost)
	}
}

func TestAnnualSavingsForScale(t *testing.T) {
	gpuCost := 10.0
	fecimCost := 1.0

	got := annualSavingsForScale(gpuCost, fecimCost)
	want := 1080000.0

	if math.Abs(got-want) > 1e-6 {
		t.Errorf("annual savings: got %.2f, want %.2f", got, want)
	}
}
