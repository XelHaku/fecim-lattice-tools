package core

import (
	"path/filepath"
	"testing"
	"time"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestPerformance_M3_PERF_02_Throughput measures batch inference throughput.
// Requirement: > 100 images/second for batch=100
// Evidence: total time (s), throughput (img/s), batch size
func TestPerformance_M3_PERF_02_Throughput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping throughput test in short mode")
	}

	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
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

	// Warmup: run 10 inferences to stabilize allocations
	for i := 0; i < 10 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	// Measure throughput for batch of 100 images
	const batchSize = 100
	if len(images) < batchSize {
		t.Fatalf("Need at least %d images, got %d", batchSize, len(images))
	}

	correctCount := 0
	start := time.Now()
	for i := 0; i < batchSize; i++ {
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
		if result.CIMPrediction == labels[i] {
			correctCount++
		}
	}
	elapsed := time.Since(start)

	totalTimeS := elapsed.Seconds()
	throughput := float64(batchSize) / totalTimeS
	accuracy := 100.0 * float64(correctCount) / float64(batchSize)

	t.Logf("M3-PERF-02: Batch Inference Throughput:")
	t.Logf("  Batch size:  %d images", batchSize)
	t.Logf("  Total time:  %.3f s", totalTimeS)
	t.Logf("  Throughput:  %.1f img/s", throughput)
	t.Logf("  Accuracy:    %.1f%% (%d/%d correct)", accuracy, correctCount, batchSize)
	t.Logf("  Avg latency: %.3f ms/image", (totalTimeS*1000.0)/float64(batchSize))

	// Requirement: > 100 images/second
	const minThroughput = 100.0
	if throughput <= minThroughput {
		t.Fatalf("Throughput %.1f img/s <= target %.1f img/s — FAIL", throughput, minThroughput)
	}

	t.Logf("M3-PERF-02: PASS — Throughput %.1f img/s > %.1f img/s requirement",
		throughput, minThroughput)
}

// TestThroughput_LargeBatch measures throughput with larger batch sizes.
// Validates that performance scales with batch size (amortized overhead).
func TestThroughput_LargeBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large batch test in short mode")
	}

	dataDir := filepath.Join("..", "..", "data")
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

	// Test larger batch size (1000 images)
	const batchSize = 1000
	if len(images) < batchSize {
		t.Fatalf("Need at least %d images, got %d", batchSize, len(images))
	}

	start := time.Now()
	for i := 0; i < batchSize; i++ {
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
	}
	elapsed := time.Since(start)

	totalTimeS := elapsed.Seconds()
	throughput := float64(batchSize) / totalTimeS
	avgLatencyMS := (totalTimeS * 1000.0) / float64(batchSize)

	t.Logf("Large Batch Throughput (batch=%d):", batchSize)
	t.Logf("  Total time:  %.3f s", totalTimeS)
	t.Logf("  Throughput:  %.1f img/s", throughput)
	t.Logf("  Avg latency: %.3f ms/image", avgLatencyMS)

	// For large batches, throughput should be even higher (amortized overhead)
	// Expect at least same as small batch (100 img/s minimum)
	const minThroughput = 100.0
	if throughput < minThroughput {
		t.Errorf("Large batch throughput %.1f img/s < %.1f img/s (expected better with larger batch)",
			throughput, minThroughput)
	}
}

// TestThroughput_FP_vs_CIM compares FP and CIM throughput.
// Both paths run in Infer(), but this validates relative performance.
func TestThroughput_FP_vs_CIM(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping FP vs CIM comparison in short mode")
	}

	dataDir := filepath.Join("..", "..", "data")
	images, labels, err := mnist.LoadMNIST(dataDir, false)
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

	const batchSize = 100
	if len(images) < batchSize {
		t.Fatalf("Need at least %d images, got %d", batchSize, len(images))
	}

	// Run batch inference and collect both FP and CIM predictions
	// (Both paths execute in single Infer() call, so timing is combined)
	fpCorrect := 0
	cimCorrect := 0
	start := time.Now()
	for i := 0; i < batchSize; i++ {
		result := net.Infer(images[i])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
		if result.FPPrediction == labels[i] {
			fpCorrect++
		}
		if result.CIMPrediction == labels[i] {
			cimCorrect++
		}
	}
	elapsed := time.Since(start)

	totalTimeS := elapsed.Seconds()
	throughput := float64(batchSize) / totalTimeS
	fpAccuracy := 100.0 * float64(fpCorrect) / float64(batchSize)
	cimAccuracy := 100.0 * float64(cimCorrect) / float64(batchSize)

	t.Logf("FP vs CIM Throughput Comparison (batch=%d):", batchSize)
	t.Logf("  Combined throughput: %.1f img/s", throughput)
	t.Logf("  Total time:          %.3f s", totalTimeS)
	t.Logf("  FP accuracy:         %.1f%% (%d/%d correct)", fpAccuracy, fpCorrect, batchSize)
	t.Logf("  CIM accuracy:        %.1f%% (%d/%d correct)", cimAccuracy, cimCorrect, batchSize)

	// Note: Since both FP and CIM run in same Infer() call,
	// the reported throughput is for dual-path execution.
	// For single-path (CIM only), throughput would be higher.
	t.Logf("  Note: Throughput is for dual-path (FP+CIM) inference in Infer()")
}

// BenchmarkBatchInference benchmarks batch processing performance.
func BenchmarkBatchInference(b *testing.B) {
	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		b.Fatalf("Failed to load MNIST test set: %v", err)
	}

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

	const batchSize = 100
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			imgIdx := (i*batchSize + j) % len(images)
			result := net.Infer(images[imgIdx])
			if result == nil {
				b.Fatal("Infer returned nil")
			}
		}
	}

	// Report effective throughput in benchmark output
	elapsed := b.Elapsed().Seconds()
	totalImages := b.N * batchSize
	throughput := float64(totalImages) / elapsed
	b.ReportMetric(throughput, "img/s")
}

// BenchmarkBatchInference_Size10 benchmarks batch=10 for comparison.
func BenchmarkBatchInference_Size10(b *testing.B) {
	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		b.Fatalf("Failed to load MNIST test set: %v", err)
	}

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

	const batchSize = 10
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			imgIdx := (i*batchSize + j) % len(images)
			_ = net.Infer(images[imgIdx])
		}
	}

	elapsed := b.Elapsed().Seconds()
	throughput := float64(b.N*batchSize) / elapsed
	b.ReportMetric(throughput, "img/s")
}

// BenchmarkBatchInference_Size1000 benchmarks batch=1000 for comparison.
func BenchmarkBatchInference_Size1000(b *testing.B) {
	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		b.Fatalf("Failed to load MNIST test set: %v", err)
	}

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

	const batchSize = 1000
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < batchSize; j++ {
			imgIdx := (i*batchSize + j) % len(images)
			_ = net.Infer(images[imgIdx])
		}
	}

	elapsed := b.Elapsed().Seconds()
	throughput := float64(b.N*batchSize) / elapsed
	b.ReportMetric(throughput, "img/s")
}
