package circuits

import "testing"

func TestPVTCornersWorkflowComputesDefaultInvestigationSummary(t *testing.T) {
	updated := newPVTCornersWorkflow(CircuitsState{
		SupplyVoltage: 1.8,
		ADCResolution: 5,
		TIAGain:       1e4,
	}).compute()

	if !near(updated.ENOBtt, 4.4150374993, 1e-9) || !near(updated.ENOBff, 4.5145731728, 1e-9) || !near(updated.ENOBss, 4.2995602819, 1e-9) {
		t.Fatalf("ENOB corners = FF %.10f TT %.10f SS %.10f", updated.ENOBff, updated.ENOBtt, updated.ENOBss)
	}
	if !near(updated.ADCNoiseLSB, 0.0162379763, 1e-9) {
		t.Fatalf("ADC noise LSB = %.10f", updated.ADCNoiseLSB)
	}
	if !near(updated.SNRdB, 31.86, 1e-9) {
		t.Fatalf("SNR = %.10f", updated.SNRdB)
	}
	if updated.PVTTemperatureSweep != "pass -40/25/85/125 C" {
		t.Fatalf("temperature sweep = %q", updated.PVTTemperatureSweep)
	}
	if updated.PVTProcessYield != 1 || updated.PVTPassSamples != 20 || updated.PVTSamples != 20 {
		t.Fatalf("process yield = %.3f (%d/%d)", updated.PVTProcessYield, updated.PVTPassSamples, updated.PVTSamples)
	}
	if !near(updated.PVTENOBNoiseCeiling, 13.6146448901, 1e-9) || updated.PVTENOBCeilingBits != 16 {
		t.Fatalf("noise ceiling = %.10f at %d bits", updated.PVTENOBNoiseCeiling, updated.PVTENOBCeilingBits)
	}
}
