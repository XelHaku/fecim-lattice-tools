package multilayer

import (
	"math"
	"testing"
)

// TestNewLayer verifies layer creation.
func TestNewLayer(t *testing.T) {
	layer := NewLayer("test", 64, 32)

	if layer.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", layer.Name)
	}
	if layer.Rows != 64 {
		t.Errorf("Expected 64 rows, got %d", layer.Rows)
	}
	if layer.Cols != 32 {
		t.Errorf("Expected 32 cols, got %d", layer.Cols)
	}
	if layer.Levels != 30 {
		t.Errorf("Expected 30 levels, got %d", layer.Levels)
	}
	if len(layer.Weights) != 64 {
		t.Errorf("Expected 64 weight rows, got %d", len(layer.Weights))
	}
	if len(layer.Weights[0]) != 32 {
		t.Errorf("Expected 32 weight cols, got %d", len(layer.Weights[0]))
	}
}

// TestNewStack verifies stack creation.
func TestNewStack(t *testing.T) {
	stack := NewStack("test-stack")

	if stack.Name != "test-stack" {
		t.Errorf("Expected name 'test-stack', got '%s'", stack.Name)
	}
	if len(stack.Layers) != 0 {
		t.Errorf("Expected 0 layers, got %d", len(stack.Layers))
	}
	if stack.Technology != "IronLattice" {
		t.Errorf("Expected 'IronLattice', got '%s'", stack.Technology)
	}
}

// TestMNISTStack verifies MNIST stack configuration.
func TestMNISTStack(t *testing.T) {
	stack := MNISTStack()

	if len(stack.Layers) != 3 {
		t.Errorf("Expected 3 layers, got %d", len(stack.Layers))
	}

	// Check layer dimensions
	expectedDims := []struct{ rows, cols int }{
		{784, 128},
		{128, 64},
		{64, 10},
	}

	for i, exp := range expectedDims {
		layer := stack.Layers[i]
		if layer.Rows != exp.rows || layer.Cols != exp.cols {
			t.Errorf("Layer %d: expected %dx%d, got %dx%d",
				i, exp.rows, exp.cols, layer.Rows, layer.Cols)
		}
	}
}

// TestAddLayer verifies layer connectivity validation.
func TestAddLayer(t *testing.T) {
	stack := NewStack("test")

	layer1 := NewLayer("L1", 16, 8)
	layer2 := NewLayer("L2", 8, 4)
	layer3 := NewLayer("L3", 16, 4) // Wrong size

	// First layer should always succeed
	err := stack.AddLayer(layer1)
	if err != nil {
		t.Errorf("First layer should succeed: %v", err)
	}

	// Matching dimensions should succeed
	err = stack.AddLayer(layer2)
	if err != nil {
		t.Errorf("Matching dimensions should succeed: %v", err)
	}

	// Mismatched dimensions should fail
	err = stack.AddLayer(layer3)
	if err == nil {
		t.Error("Mismatched dimensions should fail")
	}
}

// TestTotalCells verifies cell counting.
func TestTotalCells(t *testing.T) {
	stack := SmallStack()

	// 16*8 + 8*4 = 128 + 32 = 160
	expected := 16*8 + 8*4
	if stack.TotalCells() != expected {
		t.Errorf("Expected %d cells, got %d", expected, stack.TotalCells())
	}
}

// TestBitsPerCell verifies bit calculation.
func TestBitsPerCell(t *testing.T) {
	stack := SmallStack()

	// log2(30) ≈ 4.9
	expected := math.Log2(30)
	actual := stack.BitsPerCell()

	if math.Abs(actual-expected) > 0.01 {
		t.Errorf("Expected %.2f bits/cell, got %.2f", expected, actual)
	}
}

// TestForward verifies forward pass.
func TestForward(t *testing.T) {
	stack := SmallStack()

	// Create input
	input := make([]float64, 16)
	for i := range input {
		input[i] = float64(i) / 16.0
	}

	output, err := stack.Forward(input)
	if err != nil {
		t.Errorf("Forward failed: %v", err)
	}

	// Output should match last layer cols
	if len(output) != 4 {
		t.Errorf("Expected 4 outputs, got %d", len(output))
	}
}

// TestForwardInputValidation verifies input size validation.
func TestForwardInputValidation(t *testing.T) {
	stack := SmallStack()

	// Wrong input size
	input := make([]float64, 8) // Should be 16
	_, err := stack.Forward(input)
	if err == nil {
		t.Error("Should fail with wrong input size")
	}
}

// TestNewViaNetwork verifies via network creation.
func TestNewViaNetwork(t *testing.T) {
	stack := SmallStack()
	viaNet := NewViaNetwork(stack)

	// SmallStack has 2 layers, so 1 via array
	if len(viaNet.Arrays) != 1 {
		t.Errorf("Expected 1 via array, got %d", len(viaNet.Arrays))
	}

	// First layer has 8 outputs = 8 vias
	if viaNet.TotalVias != 8 {
		t.Errorf("Expected 8 vias, got %d", viaNet.TotalVias)
	}
}

// TestViaStats verifies via statistics.
func TestViaStats(t *testing.T) {
	stack := MNISTStack()
	viaNet := NewViaNetwork(stack)
	stats := viaNet.GetStats(stack.FootprintArea())

	// MNIST: 128 + 64 = 192 vias
	expectedVias := 128 + 64
	if stats.TotalVias != expectedVias {
		t.Errorf("Expected %d vias, got %d", expectedVias, stats.TotalVias)
	}

	if stats.PropagationDelay <= 0 {
		t.Error("Propagation delay should be positive")
	}
}

// TestEnergyEstimate verifies energy calculations.
func TestEnergyEstimate(t *testing.T) {
	stack := SmallStack()
	estimates := stack.EstimateEnergy()

	if len(estimates) != 2 {
		t.Errorf("Expected 2 estimates, got %d", len(estimates))
	}

	for i, est := range estimates {
		if est.TotalEnergy <= 0 {
			t.Errorf("Layer %d: energy should be positive", i)
		}
		if est.TraditionalComp <= 1 {
			t.Errorf("Layer %d: traditional should use more energy", i)
		}
	}
}

// TestDataFlowStats verifies data flow analysis.
func TestDataFlowStats(t *testing.T) {
	stack := SmallStack()
	stats := stack.AnalyzeDataFlow()

	if len(stats) != 2 {
		t.Errorf("Expected 2 stats, got %d", len(stats))
	}

	// Check first layer
	if stats[0].InputSize != 16 {
		t.Errorf("Expected input size 16, got %d", stats[0].InputSize)
	}
	if stats[0].MACOperations != 16*8 {
		t.Errorf("Expected %d MACs, got %d", 16*8, stats[0].MACOperations)
	}
	if stats[0].CIMAdvantage <= 1 {
		t.Error("CIM advantage should be > 1")
	}
}

// TestStackMetrics verifies physical metrics.
func TestStackMetrics(t *testing.T) {
	stack := MNISTStack()

	height := stack.StackHeight()
	if height <= 0 {
		t.Error("Stack height should be positive")
	}

	area := stack.FootprintArea()
	if area <= 0 {
		t.Error("Footprint area should be positive")
	}

	arealDensity := stack.ArealDensity()
	if arealDensity <= 0 {
		t.Error("Areal density should be positive")
	}

	t.Logf("MNIST Stack: %.0fnm height, %.2fµm² area, %.2f bits/µm²",
		height, area, arealDensity)
}

// TestViaYield verifies yield estimation.
func TestViaYield(t *testing.T) {
	stack := MNISTStack()
	viaNet := NewViaNetwork(stack)

	// Low defect density should give high yield
	yield := viaNet.EstimateViaYield(0.1) // 0.1 defects/cm²
	if yield < 0.99 {
		t.Errorf("Expected high yield, got %.2f", yield)
	}

	// High defect density should give lower yield
	yieldLow := viaNet.EstimateViaYield(1000.0) // 1000 defects/cm²
	if yieldLow >= yield {
		t.Error("Higher defect density should lower yield")
	}
}

// TestLayerUtilization verifies utilization calculation.
func TestLayerUtilization(t *testing.T) {
	stack := MNISTStack()
	util := stack.LayerUtilization()

	if len(util) != 3 {
		t.Errorf("Expected 3 utilization values, got %d", len(util))
	}

	// First layer is largest, should be 1.0
	if math.Abs(util[0]-1.0) > 0.01 {
		t.Errorf("First layer utilization should be 1.0, got %.2f", util[0])
	}

	// Other layers should be less
	for i := 1; i < len(util); i++ {
		if util[i] >= util[i-1] {
			t.Errorf("Layer %d utilization should decrease", i)
		}
	}
}

// TestRenderer verifies render output.
func TestRenderer(t *testing.T) {
	stack := SmallStack()
	renderer := DefaultRenderer()

	// Test 3D view
	view := renderer.Render3DView(stack)
	if len(view) == 0 {
		t.Error("3D view should not be empty")
	}

	// Test exploded view
	exploded := renderer.RenderExplodedView(stack)
	if len(exploded) == 0 {
		t.Error("Exploded view should not be empty")
	}

	// Test metrics
	metrics := renderer.RenderMetrics(stack)
	if len(metrics) == 0 {
		t.Error("Metrics should not be empty")
	}
}

// TestDataFlowVisualization verifies data flow rendering.
func TestDataFlowVisualization(t *testing.T) {
	stack := SmallStack()
	renderer := DefaultRenderer()

	input := make([]float64, 16)
	for i := range input {
		input[i] = float64(i) / 16.0
	}

	flow := renderer.RenderDataFlow(stack, input)
	if len(flow) == 0 {
		t.Error("Data flow visualization should not be empty")
	}
}
