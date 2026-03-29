package benchmarks

import "testing"

// mnistAccuracyThreshold is the minimum CIM inference accuracy required for the
// 8-bit quantized MNIST benchmark (M3-ACC-02).
const mnistAccuracyThreshold = 0.80

func TestRunMNISTBenchmarkDeterministic(t *testing.T) {
	cfg := BenchmarkConfig{
		Seed:              12345,
		Samples:           64,
		AccuracyThreshold: mnistAccuracyThreshold,
		LatencyNsMax:      100,
		EnergyPJMax:       10,
	}
	a := RunMNISTBenchmark(cfg)
	b := RunMNISTBenchmark(cfg)

	if a != b {
		t.Fatalf("benchmark is not deterministic:\nA=%+v\nB=%+v", a, b)
	}
}

func TestRunMNISTBenchmarkThresholds(t *testing.T) {
	res := RunMNISTBenchmark(BenchmarkConfig{Seed: 7, Samples: 32, AccuracyThreshold: 0.99, LatencyNsMax: 80, EnergyPJMax: 4.5})
	if res.AccuracyPass && res.LatencyPass && res.EnergyPass {
		t.Fatalf("expected strict thresholds to trip at least one failure, got %+v", res)
	}
	if res.OverallPass != (res.AccuracyPass && res.LatencyPass && res.EnergyPass) {
		t.Fatalf("overall pass inconsistent with component passes")
	}
}
