package neural

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestPerformance_M3_PERF_01_Latency measures single inference latency.
// Requirement: < 10 ms per image (mean), with p95 reported
// Evidence: mean latency (ms), p95 latency (ms), sample count
func TestPerformance_M3_PERF_01_Latency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping latency test in short mode")
	}

	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Load pretrained network with 8-bit CIM configuration
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Warmup: run a few inferences to stabilize allocations
	for i := 0; i < 10 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	// Measure latency for 100 individual inferences
	const numSamples = 100
	if len(images) < numSamples {
		t.Fatalf("Need at least %d images, got %d", numSamples, len(images))
	}

	latencies := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		start := time.Now()
		result := net.Infer(images[i])
		elapsed := time.Since(start)

		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}

		latencies[i] = float64(elapsed.Nanoseconds()) / 1e6 // Convert to milliseconds
	}

	// Compute statistics
	sort.Float64s(latencies)
	mean := 0.0
	for _, lat := range latencies {
		mean += lat
	}
	mean /= float64(numSamples)

	p95Index := int(float64(numSamples) * 0.95)
	if p95Index >= numSamples {
		p95Index = numSamples - 1
	}
	p95 := latencies[p95Index]

	min := latencies[0]
	max := latencies[numSamples-1]
	median := latencies[numSamples/2]

	t.Logf("M3-PERF-01: Single Inference Latency Statistics (n=%d):", numSamples)
	t.Logf("  Mean:   %.3f ms", mean)
	t.Logf("  Median: %.3f ms", median)
	t.Logf("  P95:    %.3f ms", p95)
	t.Logf("  Min:    %.3f ms", min)
	t.Logf("  Max:    %.3f ms", max)

	// Requirement: mean latency < 10 ms per image
	const maxMeanLatencyMS = 10.0
	if mean >= maxMeanLatencyMS {
		t.Fatalf("Mean latency %.3f ms >= target %.3f ms — FAIL", mean, maxMeanLatencyMS)
	}

	// Additional check: p95 should also be reasonable (< 15 ms)
	const maxP95LatencyMS = 15.0
	if p95 >= maxP95LatencyMS {
		t.Errorf("P95 latency %.3f ms >= %.3f ms (warning: high tail latency)", p95, maxP95LatencyMS)
	}

	t.Logf("M3-PERF-01: PASS — Mean latency %.3f ms < %.3f ms requirement",
		mean, maxMeanLatencyMS)
}

// BenchmarkInference provides precise latency measurement via Go benchmarking framework.
// Run with: go test -bench=BenchmarkInference -benchmem ./module3-mnist/pkg/core/
func BenchmarkInference(b *testing.B) {
	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		b.Fatalf("Failed to load MNIST test set: %v", err)
	}

	if len(images) == 0 {
		b.Fatal("No images loaded")
	}

	// Load pretrained network
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		b.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Warmup
	for i := 0; i < 10 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Cycle through test images
		imgIdx := i % len(images)
		result := net.Infer(images[imgIdx])
		if result == nil {
			b.Fatal("Infer returned nil")
		}
	}
}

// BenchmarkInference_FP measures FP-only inference latency (for comparison).
func BenchmarkInference_FP(b *testing.B) {
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		b.Fatalf("Failed to load MNIST test set: %v", err)
	}

	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		b.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Warmup
	for i := 0; i < 10 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgIdx := i % len(images)
		result := net.Infer(images[imgIdx])
		if result == nil {
			b.Fatal("Infer returned nil")
		}
		// Force use of FP result to prevent optimization
		_ = result.FPPrediction
	}
}

// BenchmarkInference_CIM measures CIM-only inference latency (for comparison).
func BenchmarkInference_CIM(b *testing.B) {
	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		b.Fatalf("Failed to load MNIST test set: %v", err)
	}

	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		b.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Warmup
	for i := 0; i < 10 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgIdx := i % len(images)
		result := net.Infer(images[imgIdx])
		if result == nil {
			b.Fatal("Infer returned nil")
		}
		// Force use of CIM result to prevent optimization
		_ = result.CIMPrediction
	}
}

// TestLatency_Determinism verifies that latency is stable (not increasing over time).
// This detects performance regressions and potential memory leaks affecting speed.
func TestLatency_Determinism(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping determinism test in short mode")
	}
	if os.Getenv("FECIM_PERF_DETERMINISM") != "1" {
		t.Skip("Skipping latency determinism test unless FECIM_PERF_DETERMINISM=1 (too environment-sensitive for default CI)")
	}

	dataDir := filepath.Join("..", "..", "module3-mnist", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Warmup
	for i := 0; i < 10 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	// Measure latency in 10 batches of 100 inferences
	const batchSize = 100
	const numBatches = 10
	batchMeans := make([]float64, numBatches)

	for batch := 0; batch < numBatches; batch++ {
		sum := 0.0
		for i := 0; i < batchSize; i++ {
			imgIdx := (batch*batchSize + i) % len(images)
			start := time.Now()
			_ = net.Infer(images[imgIdx])
			elapsed := time.Since(start)
			sum += float64(elapsed.Nanoseconds()) / 1e6
		}
		batchMeans[batch] = sum / float64(batchSize)
	}

	// Check that mean latency doesn't increase significantly over time
	firstMean := batchMeans[0]
	lastMean := batchMeans[numBatches-1]
	drift := ((lastMean - firstMean) / firstMean) * 100.0

	t.Logf("Latency Determinism Check (n=%d batches of %d):", numBatches, batchSize)
	for i, mean := range batchMeans {
		t.Logf("  Batch %2d: %.3f ms (mean)", i+1, mean)
	}
	t.Logf("  First batch mean: %.3f ms", firstMean)
	t.Logf("  Last batch mean:  %.3f ms", lastMean)
	t.Logf("  Drift:            %.2f%%", drift)

	// Allow up to 20% drift (accounts for system noise, GC)
	const maxDriftPercent = 20.0
	if drift > maxDriftPercent {
		t.Fatalf("Latency drift %.2f%% > %.2f%% — potential performance regression or memory leak",
			drift, maxDriftPercent)
	}

	t.Logf("Latency stable: drift %.2f%% within %.2f%% tolerance — PASS", drift, maxDriftPercent)
}

// formatLatency is a helper to format latency values with appropriate precision.
func formatLatency(ms float64) string {
	if ms < 1.0 {
		return fmt.Sprintf("%.2f µs", ms*1000)
	}
	return fmt.Sprintf("%.3f ms", ms)
}
