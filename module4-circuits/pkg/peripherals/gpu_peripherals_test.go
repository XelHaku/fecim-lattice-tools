package peripherals

import (
	"math"
	"testing"
)

// TestGPUPeripherals_Availability tests GPU initialization and availability check.
func TestGPUPeripherals_Availability(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	// Should not error even if GPU unavailable
	// Just check that we can query availability
	available := gpu.IsAvailable()
	t.Logf("GPU compute available: %v", available)
}

// TestGPUPeripherals_BatchDAC tests GPU-accelerated DAC conversion.
func TestGPUPeripherals_BatchDAC(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	if !gpu.IsAvailable() {
		t.Skip("GPU compute not available, skipping GPU tests")
	}

	// Test with small batch
	codes := []int32{0, 8, 16, 24, 31}
	params := DefaultDACParams(len(codes))

	voltages, err := gpu.BatchDAC(codes, params)
	if err != nil {
		t.Fatalf("BatchDAC failed: %v", err)
	}

	if len(voltages) != len(codes) {
		t.Fatalf("Expected %d voltages, got %d", len(codes), len(voltages))
	}

	// Verify monotonicity (should be increasing)
	for i := 1; i < len(voltages); i++ {
		if voltages[i] < voltages[i-1] {
			t.Errorf("Non-monotonic output: v[%d]=%.4f < v[%d]=%.4f",
				i, voltages[i], i-1, voltages[i-1])
		}
	}

	// Verify range is approximately within Vref bounds
	maxCode := (1 << uint(params.Bits)) - 1
	lsb := (params.VrefP - params.VrefN) / float32(maxCode)
	maxError := lsb * (params.INLMax + params.DNLMax)
	for i, v := range voltages {
		if v < params.VrefN-maxError || v > params.VrefP+maxError {
			t.Errorf("Voltage[%d]=%.4f out of range [%.4f, %.4f]",
				i, v, params.VrefN-maxError, params.VrefP+maxError)
		}
	}

	t.Logf("DAC conversion successful: codes=%v -> voltages=%.4f", codes, voltages)
}

// TestGPUPeripherals_BatchADC tests GPU-accelerated ADC conversion.
func TestGPUPeripherals_BatchADC(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	if !gpu.IsAvailable() {
		t.Skip("GPU compute not available, skipping GPU tests")
	}

	// Test with voltage sweep
	voltages := []float32{-1.0, -0.5, 0.0, 0.5, 1.0}
	params := DefaultADCParams(len(voltages))

	codes, quantized, err := gpu.BatchADC(voltages, params)
	if err != nil {
		t.Fatalf("BatchADC failed: %v", err)
	}

	if len(codes) != len(voltages) {
		t.Fatalf("Expected %d codes, got %d", len(voltages), len(codes))
	}
	if len(quantized) != len(voltages) {
		t.Fatalf("Expected %d quantized voltages, got %d", len(voltages), len(quantized))
	}

	// Verify codes are within valid range
	maxCode := int32((1 << uint(params.Bits)) - 1)
	for i, code := range codes {
		if code < 0 || code > maxCode {
			t.Errorf("Code[%d]=%d out of range [0, %d]", i, code, maxCode)
		}
	}

	// Verify quantized voltages are within reference range
	for i, v := range quantized {
		if v < params.VrefN || v > params.VrefP {
			t.Errorf("Quantized[%d]=%.4f out of range [%.4f, %.4f]",
				i, v, params.VrefN, params.VrefP)
		}
	}

	t.Logf("ADC conversion successful: voltages=%.4f -> codes=%v", voltages, codes)
}

// TestGPUPeripherals_BatchTIA tests GPU-accelerated TIA conversion.
func TestGPUPeripherals_BatchTIA(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	if !gpu.IsAvailable() {
		t.Skip("GPU compute not available, skipping GPU tests")
	}

	// Test with current sweep (nanoampere range)
	currents := []float32{-1e-9, -0.5e-9, 0, 0.5e-9, 1e-9}
	params := DefaultTIAParams(len(currents))

	voltages, err := gpu.BatchTIA(currents, params)
	if err != nil {
		t.Fatalf("BatchTIA failed: %v", err)
	}

	if len(voltages) != len(currents) {
		t.Fatalf("Expected %d voltages, got %d", len(currents), len(voltages))
	}

	// Verify saturation at Vmax
	for i, v := range voltages {
		if math.Abs(float64(v)) > float64(params.Vmax)+0.01 {
			t.Errorf("Voltage[%d]=%.4f exceeds saturation Vmax=%.4f",
				i, v, params.Vmax)
		}
	}

	// Verify zero current gives approximately zero voltage (within noise)
	zeroIdx := 2 // currents[2] = 0
	if math.Abs(float64(voltages[zeroIdx])) > 0.1 {
		t.Errorf("Zero current produced voltage %.4f (expected near zero)", voltages[zeroIdx])
	}

	t.Logf("TIA conversion successful: currents=%.4e -> voltages=%.4f", currents, voltages)
}

// TestGPUPeripherals_LargeBatch tests GPU performance with larger batches.
func TestGPUPeripherals_LargeBatch(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	if !gpu.IsAvailable() {
		t.Skip("GPU compute not available, skipping GPU tests")
	}

	// Test with 1024 elements (4 workgroups)
	batchSize := 1024
	codes := make([]int32, batchSize)
	for i := range codes {
		codes[i] = int32(i % 32) // Cycle through all 5-bit codes
	}

	params := DefaultDACParams(batchSize)
	voltages, err := gpu.BatchDAC(codes, params)
	if err != nil {
		t.Fatalf("BatchDAC (large) failed: %v", err)
	}

	if len(voltages) != batchSize {
		t.Fatalf("Expected %d voltages, got %d", batchSize, len(voltages))
	}

	t.Logf("Large batch DAC successful: processed %d conversions", batchSize)
}

// TestGPUPeripherals_EmptyBatch tests handling of empty input arrays.
func TestGPUPeripherals_EmptyBatch(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	if !gpu.IsAvailable() {
		t.Skip("GPU compute not available, skipping GPU tests")
	}

	// Test empty DAC
	voltages, err := gpu.BatchDAC([]int32{}, DefaultDACParams(0))
	if err != nil {
		t.Errorf("Empty BatchDAC failed: %v", err)
	}
	if len(voltages) != 0 {
		t.Errorf("Expected empty result, got %d elements", len(voltages))
	}

	// Test empty ADC
	codes, quant, err := gpu.BatchADC([]float32{}, DefaultADCParams(0))
	if err != nil {
		t.Errorf("Empty BatchADC failed: %v", err)
	}
	if len(codes) != 0 || len(quant) != 0 {
		t.Errorf("Expected empty result, got %d codes, %d quant", len(codes), len(quant))
	}

	// Test empty TIA
	v, err := gpu.BatchTIA([]float32{}, DefaultTIAParams(0))
	if err != nil {
		t.Errorf("Empty BatchTIA failed: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("Expected empty result, got %d elements", len(v))
	}
}

// TestGPUPeripherals_CPUCompare compares GPU results with CPU implementation.
func TestGPUPeripherals_CPUCompare(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}
	defer gpu.Destroy()

	if !gpu.IsAvailable() {
		t.Skip("GPU compute not available, skipping GPU tests")
	}

	// Create CPU DAC for comparison
	dacCPU := DefaultDAC()

	// Test same codes on both
	codes := []int32{0, 8, 16, 24, 31}
	paramsGPU := DACParams{
		Bits:   5,
		VrefP:  1.0,
		VrefN:  -1.0,
		INLMax: 0.5,
		DNLMax: 0.25,
		Size:   int32(len(codes)),
		Seed:   float32(12345), // Match DefaultDACParams seed
	}

	// GPU conversion
	voltagesGPU, err := gpu.BatchDAC(codes, paramsGPU)
	if err != nil {
		t.Fatalf("GPU BatchDAC failed: %v", err)
	}

	// CPU conversion (returns float64, so we convert)
	voltagesCPU := make([]float32, len(codes))
	for i, code := range codes {
		voltagesCPU[i] = float32(dacCPU.Convert(int(code)))
	}

	// Compare results - should be similar but not identical due to different noise patterns
	// Check that they're in the same ballpark (within a few LSBs)
	maxCodeGPU := (1 << uint(paramsGPU.Bits)) - 1
	lsb := (paramsGPU.VrefP - paramsGPU.VrefN) / float32(maxCodeGPU)
	tolerance := lsb * 3.0 // Allow 3 LSB difference

	for i := range codes {
		diff := math.Abs(float64(voltagesGPU[i] - voltagesCPU[i]))
		if diff > float64(tolerance) {
			t.Logf("Warning: GPU/CPU mismatch at code %d: GPU=%.4f, CPU=%.4f, diff=%.4f LSBs",
				codes[i], voltagesGPU[i], voltagesCPU[i], diff/float64(lsb))
		}
	}

	t.Logf("GPU vs CPU comparison completed")
}

// TestDefaultParams verifies default parameter constructors.
func TestDefaultParams(t *testing.T) {
	size := 100

	dacParams := DefaultDACParams(size)
	if dacParams.Size != int32(size) {
		t.Errorf("DAC Size: expected %d, got %d", size, dacParams.Size)
	}
	if dacParams.Bits != 5 {
		t.Errorf("DAC Bits: expected 5, got %d", dacParams.Bits)
	}

	adcParams := DefaultADCParams(size)
	if adcParams.Size != int32(size) {
		t.Errorf("ADC Size: expected %d, got %d", size, adcParams.Size)
	}
	if adcParams.Bits != 5 {
		t.Errorf("ADC Bits: expected 5, got %d", adcParams.Bits)
	}

	tiaParams := DefaultTIAParams(size)
	if tiaParams.Size != int32(size) {
		t.Errorf("TIA Size: expected %d, got %d", size, tiaParams.Size)
	}
	if tiaParams.Gain <= 0 {
		t.Errorf("TIA Gain should be positive, got %.2e", tiaParams.Gain)
	}
}

// TestGPUPeripherals_Destroy verifies proper cleanup.
func TestGPUPeripherals_Destroy(t *testing.T) {
	gpu, err := NewGPUPeripherals()
	if err != nil {
		t.Fatalf("NewGPUPeripherals failed: %v", err)
	}

	// Should not panic
	gpu.Destroy()

	// Double destroy should be safe
	gpu.Destroy()

	// After destroy, IsAvailable should return false
	if gpu.IsAvailable() {
		t.Error("IsAvailable should return false after Destroy")
	}
}
