package gpu

import (
	"testing"
)

// ============================================================================
// NewGPUNetwork Tests
// ============================================================================

// TestNewGPUNetwork_Creation verifies NewGPUNetwork creates valid struct.
func TestNewGPUNetwork_Creation(t *testing.T) {
	net := NewGPUNetwork()
	if net == nil {
		t.Fatal("NewGPUNetwork returned nil")
	}
	defer net.Destroy()

	// Struct should always be created, even if GPU unavailable
	// Just check it's non-nil and has IsAvailable method
	isAvailable := net.IsAvailable()
	t.Logf("GPU available: %v", isAvailable)

	if isAvailable {
		// If GPU is available, verify internal state is initialized
		if net.ctx == nil {
			t.Error("GPU available but ctx is nil")
		}
		if net.denseLayer == nil {
			t.Error("GPU available but denseLayer is nil")
		}
		if !net.available {
			t.Error("GPU available but available flag is false")
		}
	} else {
		// If GPU unavailable, verify graceful degradation
		if net.available {
			t.Error("GPU unavailable but available flag is true")
		}
	}
}

// TestNewGPUNetwork_MultipleInstances verifies multiple networks can coexist.
func TestNewGPUNetwork_MultipleInstances(t *testing.T) {
	net1 := NewGPUNetwork()
	defer net1.Destroy()

	net2 := NewGPUNetwork()
	defer net2.Destroy()

	// Both should be valid
	if net1 == nil || net2 == nil {
		t.Fatal("Failed to create multiple GPUNetwork instances")
	}

	// Both should have same availability status
	if net1.IsAvailable() != net2.IsAvailable() {
		t.Error("Different availability status for concurrent networks")
	}
}

// ============================================================================
// GPUNetwork.IsAvailable Tests
// ============================================================================

// TestGPUNetwork_IsAvailable_ReturnsBool verifies IsAvailable returns boolean.
func TestGPUNetwork_IsAvailable_ReturnsBool(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	// Should return a valid boolean (not panic)
	isAvailable := net.IsAvailable()

	// Log the result for debugging
	t.Logf("GPU available: %v", isAvailable)

	// Just verify it's a valid boolean operation (no panic)
	if isAvailable {
		t.Log("GPU acceleration is available")
	} else {
		t.Log("GPU acceleration is NOT available (CPU fallback)")
	}
}

// TestGPUNetwork_IsAvailable_ConsistentAfterDestroy verifies behavior after Destroy.
func TestGPUNetwork_IsAvailable_ConsistentAfterDestroy(t *testing.T) {
	net := NewGPUNetwork()

	availableBefore := net.IsAvailable()
	t.Logf("Available before destroy: %v", availableBefore)

	net.Destroy()

	availableAfter := net.IsAvailable()
	t.Logf("Available after destroy: %v", availableAfter)

	// After destroy, should report unavailable
	if availableAfter {
		t.Error("IsAvailable returned true after Destroy()")
	}
}

// ============================================================================
// GPUNetwork.Forward Tests (Error Handling)
// ============================================================================

// TestGPUNetwork_Forward_UnavailableGPU verifies error when GPU unavailable.
func TestGPUNetwork_Forward_UnavailableGPU(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if net.IsAvailable() {
		t.Skip("GPU is available, cannot test unavailable behavior")
	}

	input := []float32{1.0, 2.0, 3.0}
	layers := []LayerWeights{
		{
			Weights:    []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
			Bias:       []float32{0.1, 0.2},
			Rows:       2,
			Cols:       3,
			Activation: ActivationNone,
		},
	}

	output, err := net.Forward(input, layers)
	if err == nil {
		t.Error("Expected error when GPU unavailable, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output when GPU unavailable, got: %v", output)
	}

	// Verify error message mentions GPU
	errMsg := err.Error()
	if errMsg != "GPU not available" {
		t.Errorf("Unexpected error message: %s", errMsg)
	}
}

// TestGPUNetwork_Forward_EmptyLayers verifies error for empty layer list.
func TestGPUNetwork_Forward_EmptyLayers(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, skipping test")
	}

	input := []float32{1.0, 2.0, 3.0}
	layers := []LayerWeights{} // Empty

	output, err := net.Forward(input, layers)
	if err == nil {
		t.Error("Expected error for empty layers, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output for empty layers, got: %v", output)
	}
}

// TestGPUNetwork_Forward_DimensionMismatch verifies error for mismatched dimensions.
func TestGPUNetwork_Forward_DimensionMismatch(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, skipping test")
	}

	// Input has 5 elements, but layer expects 3
	input := []float32{1.0, 2.0, 3.0, 4.0, 5.0}
	layers := []LayerWeights{
		{
			Weights:    []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
			Bias:       []float32{0.1, 0.2},
			Rows:       2,
			Cols:       3, // Expects 3 inputs, but got 5
			Activation: ActivationNone,
		},
	}

	output, err := net.Forward(input, layers)
	if err == nil {
		t.Error("Expected error for dimension mismatch, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output for dimension mismatch, got: %v", output)
	}
}

// ============================================================================
// GPUNetwork.Destroy Tests
// ============================================================================

// TestGPUNetwork_Destroy_NoPanic verifies Destroy doesn't panic.
func TestGPUNetwork_Destroy_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Destroy panicked: %v", r)
		}
	}()

	net := NewGPUNetwork()
	net.Destroy()

	// Should be safe to call multiple times
	net.Destroy()
	net.Destroy()
}

// TestGPUNetwork_Destroy_ReleasesResources verifies resources are released.
func TestGPUNetwork_Destroy_ReleasesResources(t *testing.T) {
	net := NewGPUNetwork()
	wasAvailable := net.IsAvailable()

	net.Destroy()

	// After destroy, internal state should be cleared
	if net.available {
		t.Error("available flag not cleared after Destroy")
	}
	if net.ctx != nil {
		t.Error("ctx not cleared after Destroy")
	}
	if net.denseLayer != nil {
		t.Error("denseLayer not cleared after Destroy")
	}

	// Should no longer be available
	if net.IsAvailable() {
		t.Error("IsAvailable returned true after Destroy")
	}

	t.Logf("Was available before destroy: %v", wasAvailable)
}

// TestGPUNetwork_Destroy_OperationsFailAfter verifies operations fail after Destroy.
func TestGPUNetwork_Destroy_OperationsFailAfter(t *testing.T) {
	net := NewGPUNetwork()
	wasAvailable := net.IsAvailable()

	net.Destroy()

	// Try to use after destroy
	input := []float32{1.0, 2.0, 3.0}
	layers := []LayerWeights{
		{
			Weights:    []float32{0.1, 0.2, 0.3, 0.4, 0.5, 0.6},
			Bias:       []float32{0.1, 0.2},
			Rows:       2,
			Cols:       3,
			Activation: ActivationNone,
		},
	}

	output, err := net.Forward(input, layers)
	if err == nil {
		t.Error("Expected error after Destroy, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output after Destroy, got: %v", output)
	}

	t.Logf("Was available before destroy: %v", wasAvailable)
}

// ============================================================================
// NewDenseLayerGPU Tests
// ============================================================================

// TestNewDenseLayerGPU_NilContext verifies error for nil context.
func TestNewDenseLayerGPU_NilContext(t *testing.T) {
	layer, err := NewDenseLayerGPU(nil, 10, 20)
	if err == nil {
		t.Error("Expected error for nil context, got nil")
	}
	if layer != nil {
		t.Error("Expected nil layer for nil context")
	}

	expectedMsg := "VulkanContext is nil"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message %q, got %q", expectedMsg, err.Error())
	}
}

// TestNewDenseLayerGPU_InvalidDimensions verifies error for invalid dimensions.
func TestNewDenseLayerGPU_InvalidDimensions(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test NewDenseLayerGPU")
	}

	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"zero_rows", 0, 10},
		{"zero_cols", 10, 0},
		{"negative_rows", -5, 10},
		{"negative_cols", 10, -5},
		{"both_zero", 0, 0},
		{"both_negative", -5, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer, err := NewDenseLayerGPU(net.ctx, tt.rows, tt.cols)
			if err == nil {
				t.Errorf("Expected error for dimensions rows=%d, cols=%d", tt.rows, tt.cols)
			}
			if layer != nil {
				t.Error("Expected nil layer for invalid dimensions")
			}
		})
	}
}

// TestNewDenseLayerGPU_ValidDimensions verifies successful creation.
func TestNewDenseLayerGPU_ValidDimensions(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test NewDenseLayerGPU")
	}

	tests := []struct {
		name string
		rows int
		cols int
	}{
		{"small", 4, 8},
		{"medium", 64, 128},
		{"large", 128, 784},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layer, err := NewDenseLayerGPU(net.ctx, tt.rows, tt.cols)
			if err != nil {
				t.Fatalf("Failed to create layer with rows=%d, cols=%d: %v", tt.rows, tt.cols, err)
			}
			defer layer.Destroy()

			if layer == nil {
				t.Fatal("Layer is nil despite no error")
			}

			// Verify internal state
			if layer.maxRows != tt.rows {
				t.Errorf("maxRows mismatch: expected %d, got %d", tt.rows, layer.maxRows)
			}
			if layer.maxCols != tt.cols {
				t.Errorf("maxCols mismatch: expected %d, got %d", tt.cols, layer.maxCols)
			}
			if !layer.IsAvailable() {
				t.Error("Layer not available after successful creation")
			}
		})
	}
}

// ============================================================================
// DenseLayerGPU.IsAvailable Tests
// ============================================================================

// TestDenseLayerGPU_IsAvailable_ReturnsBool verifies IsAvailable returns boolean.
func TestDenseLayerGPU_IsAvailable_ReturnsBool(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test DenseLayerGPU.IsAvailable")
	}

	layer, err := NewDenseLayerGPU(net.ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}
	defer layer.Destroy()

	// Should return true since GPU is available
	if !layer.IsAvailable() {
		t.Error("IsAvailable returned false for valid layer")
	}
}

// TestDenseLayerGPU_IsAvailable_NilLayer verifies behavior for nil layer.
func TestDenseLayerGPU_IsAvailable_NilLayer(t *testing.T) {
	var layer *DenseLayerGPU

	// Calling IsAvailable on nil should return false (not panic)
	if layer.IsAvailable() {
		t.Error("IsAvailable returned true for nil layer")
	}
}

// TestDenseLayerGPU_IsAvailable_AfterDestroy verifies behavior after Destroy.
func TestDenseLayerGPU_IsAvailable_AfterDestroy(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test DenseLayerGPU.IsAvailable")
	}

	layer, err := NewDenseLayerGPU(net.ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}

	if !layer.IsAvailable() {
		t.Error("Layer not available after creation")
	}

	layer.Destroy()

	// After destroy, should return false
	if layer.IsAvailable() {
		t.Error("IsAvailable returned true after Destroy")
	}
}

// ============================================================================
// DenseLayerGPU.Forward Tests (Error Handling)
// ============================================================================

// TestDenseLayerGPU_Forward_UnavailableGPU verifies error when GPU unavailable.
func TestDenseLayerGPU_Forward_UnavailableGPU(t *testing.T) {
	var layer *DenseLayerGPU // nil layer

	weights := []float32{0.1, 0.2, 0.3, 0.4}
	bias := []float32{0.1, 0.2}
	input := []float32{1.0, 2.0}

	output, err := layer.Forward(weights, bias, input, 2, 2, ActivationNone)
	if err == nil {
		t.Error("Expected error for nil layer, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output for nil layer, got: %v", output)
	}
}

// TestDenseLayerGPU_Forward_InvalidDimensions verifies error handling.
func TestDenseLayerGPU_Forward_InvalidDimensions(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test DenseLayerGPU.Forward")
	}

	layer, err := NewDenseLayerGPU(net.ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}
	defer layer.Destroy()

	// Test various invalid dimension scenarios
	t.Run("zero_dimensions", func(t *testing.T) {
		weights := []float32{}
		bias := []float32{}
		input := []float32{}

		_, err := layer.Forward(weights, bias, input, 0, 0, ActivationNone)
		if err == nil {
			t.Error("Expected error for zero dimensions")
		}
	})

	t.Run("exceed_max_dimensions", func(t *testing.T) {
		// Layer created with max 10x20, try to use 15x25
		rows, cols := 15, 25
		weights := make([]float32, rows*cols)
		bias := make([]float32, rows)
		input := make([]float32, cols)

		_, err := layer.Forward(weights, bias, input, rows, cols, ActivationNone)
		if err == nil {
			t.Error("Expected error for dimensions exceeding max")
		}
	})

	t.Run("weight_size_mismatch", func(t *testing.T) {
		weights := []float32{0.1, 0.2} // Wrong size (expected 4 for 2x2)
		bias := []float32{0.1, 0.2}
		input := []float32{1.0, 2.0}

		_, err := layer.Forward(weights, bias, input, 2, 2, ActivationNone)
		if err == nil {
			t.Error("Expected error for weight size mismatch")
		}
	})

	t.Run("bias_size_mismatch", func(t *testing.T) {
		weights := []float32{0.1, 0.2, 0.3, 0.4}
		bias := []float32{0.1} // Wrong size (expected 2)
		input := []float32{1.0, 2.0}

		_, err := layer.Forward(weights, bias, input, 2, 2, ActivationNone)
		if err == nil {
			t.Error("Expected error for bias size mismatch")
		}
	})

	t.Run("input_size_mismatch", func(t *testing.T) {
		weights := []float32{0.1, 0.2, 0.3, 0.4}
		bias := []float32{0.1, 0.2}
		input := []float32{1.0} // Wrong size (expected 2)

		_, err := layer.Forward(weights, bias, input, 2, 2, ActivationNone)
		if err == nil {
			t.Error("Expected error for input size mismatch")
		}
	})
}

// ============================================================================
// DenseLayerGPU.Destroy Tests
// ============================================================================

// TestDenseLayerGPU_Destroy_NoPanic verifies Destroy doesn't panic.
func TestDenseLayerGPU_Destroy_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Destroy panicked: %v", r)
		}
	}()

	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test DenseLayerGPU.Destroy")
	}

	layer, err := NewDenseLayerGPU(net.ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}

	layer.Destroy()

	// Should be safe to call multiple times
	layer.Destroy()
	layer.Destroy()
}

// TestDenseLayerGPU_Destroy_NilLayer verifies Destroy on nil causes expected panic.
func TestDenseLayerGPU_Destroy_NilLayer(t *testing.T) {
	// Calling methods on nil pointer in Go causes panic - this is expected behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when calling Destroy on nil layer")
		}
	}()

	var layer *DenseLayerGPU
	layer.Destroy() // Will panic - this is Go's expected behavior
}

// TestDenseLayerGPU_Destroy_ReleasesResources verifies resources are released.
func TestDenseLayerGPU_Destroy_ReleasesResources(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test DenseLayerGPU.Destroy")
	}

	layer, err := NewDenseLayerGPU(net.ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}

	layer.Destroy()

	// Verify all buffers are cleared
	if layer.weightBuffer != nil {
		t.Error("weightBuffer not cleared after Destroy")
	}
	if layer.biasBuffer != nil {
		t.Error("biasBuffer not cleared after Destroy")
	}
	if layer.inputBuffer != nil {
		t.Error("inputBuffer not cleared after Destroy")
	}
	if layer.outputBuffer != nil {
		t.Error("outputBuffer not cleared after Destroy")
	}
	if layer.batchInputBuffer != nil {
		t.Error("batchInputBuffer not cleared after Destroy")
	}
	if layer.batchOutputBuffer != nil {
		t.Error("batchOutputBuffer not cleared after Destroy")
	}
	if layer.pipeline != nil {
		t.Error("pipeline not cleared after Destroy")
	}
	if layer.batchPipeline != nil {
		t.Error("batchPipeline not cleared after Destroy")
	}
	if layer.ctx != nil {
		t.Error("ctx not cleared after Destroy")
	}
}

// TestDenseLayerGPU_Destroy_OperationsFailAfter verifies operations fail after Destroy.
func TestDenseLayerGPU_Destroy_OperationsFailAfter(t *testing.T) {
	net := NewGPUNetwork()
	defer net.Destroy()

	if !net.IsAvailable() {
		t.Skip("GPU not available, cannot test DenseLayerGPU.Destroy")
	}

	layer, err := NewDenseLayerGPU(net.ctx, 10, 20)
	if err != nil {
		t.Fatalf("Failed to create layer: %v", err)
	}

	layer.Destroy()

	// Try to use after destroy
	weights := createTestWeights(2, 2)
	bias := createTestBias(2)
	input := createTestInput(2)

	output, err := layer.Forward(weights, bias, input, 2, 2, ActivationNone)
	if err == nil {
		t.Error("Expected error after Destroy, got nil")
	}
	if output != nil {
		t.Errorf("Expected nil output after Destroy, got: %v", output)
	}
}
