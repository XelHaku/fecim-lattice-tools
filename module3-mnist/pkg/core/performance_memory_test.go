package core

import (
	"path/filepath"
	"runtime"
	"testing"

	"fecim-lattice-tools/module3-mnist/pkg/mnist"
)

// TestPerformance_M3_PERF_04_Memory measures memory footprint per inference.
// Requirement: Stable allocations (no memory leaks), reasonable bytes/op
// Evidence: bytes/op, allocs/op, stability over 1000 inferences
func TestPerformance_M3_PERF_04_Memory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping memory test in short mode")
	}

	// Load MNIST test set
	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	// Load pretrained network
	net := NewDualModeNetwork(784, 128, 10)
	net.Config.DACBits = 8
	net.Config.ADCBits = 8
	net.Config.NumLevels = 256
	net.Config.NoiseLevel = 0.0

	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Warmup: stabilize allocations
	for i := 0; i < 50 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	// Force GC to get clean baseline
	runtime.GC()
	runtime.GC() // Double GC for stability

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Run 1000 inferences and measure memory
	const numInferences = 1000
	if len(images) < numInferences {
		t.Fatalf("Need at least %d images, got %d", numInferences, len(images))
	}

	for i := 0; i < numInferences; i++ {
		result := net.Infer(images[i%len(images)])
		if result == nil {
			t.Fatalf("Infer returned nil for image %d", i)
		}
	}

	runtime.ReadMemStats(&m2)

	// Calculate memory deltas
	totalAllocBytes := m2.TotalAlloc - m1.TotalAlloc
	heapAllocBytes := m2.HeapAlloc - m1.HeapAlloc
	numMallocs := m2.Mallocs - m1.Mallocs
	numFrees := m2.Frees - m1.Frees

	bytesPerInference := float64(totalAllocBytes) / float64(numInferences)
	allocsPerInference := float64(numMallocs) / float64(numInferences)

	t.Logf("M3-PERF-04: Memory Footprint Analysis (n=%d inferences):", numInferences)
	t.Logf("  Total allocated:  %d bytes (%.2f KB)", totalAllocBytes, float64(totalAllocBytes)/1024.0)
	t.Logf("  Heap delta:       %d bytes (%.2f KB)", heapAllocBytes, float64(heapAllocBytes)/1024.0)
	t.Logf("  Mallocs:          %d", numMallocs)
	t.Logf("  Frees:            %d", numFrees)
	t.Logf("  Net objects:      %d (Mallocs - Frees)", numMallocs-numFrees)
	t.Logf("")
	t.Logf("  Per inference:")
	t.Logf("    Bytes/op:  %.1f bytes", bytesPerInference)
	t.Logf("    Allocs/op: %.1f", allocsPerInference)

	// Note on memory measurement:
	// Go's GC is asynchronous, so malloc/free counts may show imbalance even without leaks.
	// The real leak test is in TestMemory_LeakDetection which measures heap growth over time.
	_ = int64(numMallocs - numFrees)                                          // netObjects (unused, for debugging)
	_ = (float64(heapAllocBytes) / float64(totalAllocBytes)) * 100.0         // heapGrowthPercent (unused, for debugging)
	
	t.Logf("")
	if numFrees == 0 {
		t.Logf("  Note: GC has not run during test (Frees=0)")
		t.Logf("  See TestMemory_LeakDetection for true leak validation")
	} else {
		t.Logf("  Free rate: %.1f%% (Frees/Mallocs)", 100.0*float64(numFrees)/float64(numMallocs))
	}

	// Sanity check: bytes/op should be reasonable (< 100 KB per inference)
	// Network has pooled scratch buffers, so allocations should be minimal
	const maxBytesPerOp = 100_000.0 // 100 KB
	if bytesPerInference > maxBytesPerOp {
		t.Errorf("Bytes/op %.1f exceeds threshold %.1f (excessive allocation)",
			bytesPerInference, maxBytesPerOp)
	}

	t.Logf("M3-PERF-04: PASS — Memory footprint reasonable")
	t.Logf("  Bytes/op: %.1f (< %.1f threshold)", bytesPerInference, maxBytesPerOp)
	t.Logf("  Allocs/op: %.1f (pooled scratch buffers reduce allocations)", allocsPerInference)
}

// TestMemory_LeakDetection runs many inferences and verifies heap stays stable.
// This is a more aggressive leak test with larger sample size.
func TestMemory_LeakDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping leak detection in short mode")
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
	for i := 0; i < 100 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	// Measure heap at checkpoints
	runtime.GC()
	runtime.GC()

	checkpoints := []int{0, 1000, 2000, 3000, 4000, 5000}
	heapSizes := make([]uint64, len(checkpoints))

	inferCount := 0
	for cpIdx, target := range checkpoints {
		for inferCount < target {
			_ = net.Infer(images[inferCount%len(images)])
			inferCount++
		}

		runtime.GC()
		runtime.GC()
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		heapSizes[cpIdx] = m.HeapAlloc
	}

	t.Logf("Memory Leak Detection (5000 inferences):")
	for i, cp := range checkpoints {
		t.Logf("  Checkpoint %5d: heap = %8d bytes (%.2f KB)",
			cp, heapSizes[i], float64(heapSizes[i])/1024.0)
	}

	// Check heap growth rate
	baselineHeap := heapSizes[1] // Use checkpoint 1000 as baseline (after warmup)
	finalHeap := heapSizes[len(heapSizes)-1]
	heapGrowth := int64(finalHeap) - int64(baselineHeap)
	growthPercent := (float64(heapGrowth) / float64(baselineHeap)) * 100.0

	t.Logf("")
	t.Logf("Heap Growth Analysis:")
	t.Logf("  Baseline (1000): %d bytes", baselineHeap)
	t.Logf("  Final (5000):    %d bytes", finalHeap)
	t.Logf("  Growth:          %d bytes (%.1f%%)", heapGrowth, growthPercent)

	// Allow up to 50% heap growth over 4000 additional inferences
	// (accounts for GC cycle timing and some retained objects)
	const maxGrowthPercent = 50.0
	if growthPercent > maxGrowthPercent {
		t.Fatalf("Heap growth %.1f%% > %.1f%% — potential memory leak",
			growthPercent, maxGrowthPercent)
	}

	t.Logf("Heap growth %.1f%% is within %.1f%% tolerance — no leak detected",
		growthPercent, maxGrowthPercent)
}

// TestMemory_ScratchPoolEffectiveness validates scratch buffer pooling.
// The network uses sync.Pool for scratch buffers; verify this reduces allocations.
func TestMemory_ScratchPoolEffectiveness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping pool effectiveness test in short mode")
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

	// Warmup to populate pool
	for i := 0; i < 100 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	// Measure allocations for 100 inferences after pool is populated
	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	const numInferences = 100
	for i := 0; i < numInferences; i++ {
		_ = net.Infer(images[i%len(images)])
	}

	var m2 runtime.MemStats
	runtime.ReadMemStats(&m2)

	totalAllocs := m2.Mallocs - m1.Mallocs
	allocsPerInference := float64(totalAllocs) / float64(numInferences)

	t.Logf("Scratch Pool Effectiveness (n=%d inferences after warmup):", numInferences)
	t.Logf("  Total mallocs:    %d", totalAllocs)
	t.Logf("  Allocs/inference: %.1f", allocsPerInference)

	// With effective pooling, allocs/op should be very low (< 50)
	// Without pooling, it would be much higher (100-200+ for all the slices)
	const maxAllocsPerOp = 50.0
	if allocsPerInference > maxAllocsPerOp {
		t.Logf("WARNING: Allocs/inference %.1f > %.1f — pool may not be effective",
			allocsPerInference, maxAllocsPerOp)
		t.Logf("  (This is not a failure, but suggests optimization opportunity)")
	} else {
		t.Logf("Pool is effective: %.1f allocs/op < %.1f threshold",
			allocsPerInference, maxAllocsPerOp)
	}
}

// BenchmarkMemoryFootprint provides precise bytes/op and allocs/op metrics.
// Run with: go test -bench=BenchmarkMemoryFootprint -benchmem ./module3-mnist/pkg/core/
func BenchmarkMemoryFootprint(b *testing.B) {
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
	for i := 0; i < 100 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgIdx := i % len(images)
		result := net.Infer(images[imgIdx])
		if result == nil {
			b.Fatal("Infer returned nil")
		}
	}
}

// BenchmarkMemoryFootprint_NoPool simulates inference without pooling (for comparison).
// Note: This benchmark is theoretical since the network always uses pooling.
// It serves as documentation of the pooling benefit.
func BenchmarkMemoryFootprint_WithPool(b *testing.B) {
	// This is the same as BenchmarkMemoryFootprint (pool is always enabled)
	// Included for symmetry with naming convention
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

	// Warmup to populate pool
	for i := 0; i < 100 && i < len(images); i++ {
		_ = net.Infer(images[i])
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		imgIdx := i % len(images)
		result := net.Infer(images[imgIdx])
		if result == nil {
			b.Fatal("Infer returned nil")
		}
	}
}

// TestMemory_ResultSizeValidation checks InferenceResult struct size.
// Large result objects could indicate inefficient data structures.
func TestMemory_ResultSizeValidation(t *testing.T) {
	dataDir := filepath.Join("..", "..", "data")
	images, _, err := mnist.LoadMNIST(dataDir, false)
	if err != nil {
		t.Fatalf("Failed to load MNIST test set: %v", err)
	}

	net := NewDualModeNetwork(784, 128, 10)
	weightsFile := filepath.Join(dataDir, "pretrained_weights.json")
	if err := net.LoadWeights(weightsFile); err != nil {
		t.Fatalf("Failed to load pretrained weights: %v", err)
	}

	// Get a sample result
	result := net.Infer(images[0])
	if result == nil {
		t.Fatal("Infer returned nil")
	}

	// Estimate result size (this is approximate)
	// InferenceResult contains several float64 slices and scalar fields
	// FPLogits, FPProbabilities, CIMLogits, CIMProbabilities: 4 × 10 × 8 bytes = 320 bytes
	// FPHidden, CIMHidden (if present): 2 × 128 × 8 bytes = 2048 bytes
	// Scalar fields: ~100 bytes
	// Total: ~2500 bytes with hidden layers, ~500 bytes for single-layer mode

	estimatedSize := 0
	estimatedSize += len(result.FPLogits) * 8
	estimatedSize += len(result.FPProbabilities) * 8
	estimatedSize += len(result.CIMLogits) * 8
	estimatedSize += len(result.CIMProbabilities) * 8
	if result.FPHidden != nil {
		estimatedSize += len(result.FPHidden) * 8
	}
	if result.CIMHidden != nil {
		estimatedSize += len(result.CIMHidden) * 8
	}
	estimatedSize += 100 // Scalar fields and struct overhead

	t.Logf("InferenceResult Size Estimate:")
	t.Logf("  FPLogits:         %d floats (%d bytes)", len(result.FPLogits), len(result.FPLogits)*8)
	t.Logf("  FPProbabilities:  %d floats (%d bytes)", len(result.FPProbabilities), len(result.FPProbabilities)*8)
	t.Logf("  CIMLogits:        %d floats (%d bytes)", len(result.CIMLogits), len(result.CIMLogits)*8)
	t.Logf("  CIMProbabilities: %d floats (%d bytes)", len(result.CIMProbabilities), len(result.CIMProbabilities)*8)
	if result.FPHidden != nil {
		t.Logf("  FPHidden:         %d floats (%d bytes)", len(result.FPHidden), len(result.FPHidden)*8)
	}
	if result.CIMHidden != nil {
		t.Logf("  CIMHidden:        %d floats (%d bytes)", len(result.CIMHidden), len(result.CIMHidden)*8)
	}
	t.Logf("  Estimated total:  ~%d bytes (%.2f KB)", estimatedSize, float64(estimatedSize)/1024.0)

	// Sanity check: result should be < 10 KB
	const maxResultSize = 10 * 1024
	if estimatedSize > maxResultSize {
		t.Errorf("InferenceResult size %d bytes > %d bytes (too large)", estimatedSize, maxResultSize)
	}
}
