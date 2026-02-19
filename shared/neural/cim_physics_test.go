package neural

import (
	"math"
	"testing"
)

func TestForwardCIM_ConductancePathMatchesFP_BoundedError(t *testing.T) {
	net := createTestNetwork()
	in := []float64{0.2, 0.1, 0.4, 0.3}

	fp := net.forwardFP(in, net.QuantWeights1, net.QuantBias1)
	cim := net.forwardCIM(in, net.QuantWeights1, net.QuantBias1)

	if len(fp) != len(cim) {
		t.Fatalf("len mismatch: fp=%d cim=%d", len(fp), len(cim))
	}
	for i := range fp {
		den := math.Max(1e-9, math.Abs(fp[i]))
		rel := math.Abs(cim[i]-fp[i]) / den
		if rel > 0.05 {
			t.Fatalf("output[%d] rel err %.3f > 5%% (fp=%g cim=%g)", i, rel, fp[i], cim[i])
		}
	}
}

func TestDecomposedNoiseSigmaQuadrature(t *testing.T) {
	c := CIMNoiseComponents{ADC: 0.01, Thermal: 0.02, Flicker: 0.03, CellVariation: 0.04}
	want := math.Sqrt(0.01*0.01 + 0.02*0.02 + 0.03*0.03 + 0.04*0.04)
	if math.Abs(c.TotalSigma()-want) > 1e-12 {
		t.Fatalf("total sigma mismatch: got %g want %g", c.TotalSigma(), want)
	}
}

func TestTIA_BandwidthDropsWithGBWLimit(t *testing.T) {
	m := TIAModel{RfOhm: 100e3, CfF: 0.5e-12, GBWHz: 1e6}
	low := m.TransimpedanceMag(1e3)
	high := m.TransimpedanceMag(20e6)
	if !(high < low) {
		t.Fatalf("expected bandwidth-limited gain rolloff, low=%g high=%g", low, high)
	}
}

func TestADCThroughputConstraintReadLatency(t *testing.T) {
	net := createTestNetwork()
	net.Config.ADCConversionTimeS = 100e-9
	net.Config.ADCParallelism = 1
	lat := net.adcReadLatencySecondsLocked(128)
	want := 128 * 100e-9
	if math.Abs(lat-want) > 1e-12 {
		t.Fatalf("latency mismatch: got %g want %g", lat, want)
	}
}

func TestADCSNRModel_6p02NPlus1p76(t *testing.T) {
	for _, bits := range []int{6, 8, 10} {
		theory := adcTheoreticalSNRdB(bits)
		if !validateADCSNR(bits, theory+2.5, 3.0) {
			t.Fatalf("bits=%d expected pass within 3dB", bits)
		}
		if validateADCSNR(bits, theory+3.5, 3.0) {
			t.Fatalf("bits=%d expected fail outside 3dB", bits)
		}
	}
}
