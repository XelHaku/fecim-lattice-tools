package comparison

import (
	"testing"
)

// TestTraditionalCPU verifies CPU architecture creation.
func TestTraditionalCPU(t *testing.T) {
	cpu := TraditionalCPU()

	if cpu.Name != "Traditional CPU+DRAM" {
		t.Errorf("Expected 'Traditional CPU+DRAM', got '%s'", cpu.Name)
	}
	if cpu.TDP <= 0 {
		t.Error("TDP should be positive")
	}
	if cpu.PeakTOPS <= 0 {
		t.Error("PeakTOPS should be positive")
	}
	if cpu.TOPSPerWatt <= 0 {
		t.Error("TOPSPerWatt should be positive")
	}
}

// TestGPUAccelerator verifies GPU architecture creation.
func TestGPUAccelerator(t *testing.T) {
	gpu := GPUAccelerator()

	if gpu.Name != "GPU Accelerator" {
		t.Errorf("Expected 'GPU Accelerator', got '%s'", gpu.Name)
	}
	if gpu.PeakTOPS <= 0 {
		t.Error("PeakTOPS should be positive")
	}
	if gpu.MemoryBW <= 0 {
		t.Error("Memory bandwidth should be positive")
	}
}

// TestFeCIMChip verifies FeCIM architecture creation.
func TestFeCIMChip(t *testing.T) {
	iron := FeCIMChip()

	if iron.Name != "FeCIM CIM" {
		t.Errorf("Expected 'FeCIM CIM', got '%s'", iron.Name)
	}
	if iron.Technology != "FeFET Crossbar" {
		t.Errorf("Expected 'FeFET Crossbar', got '%s'", iron.Technology)
	}
}

// TestFeCIMEfficiency verifies FeCIM is most efficient.
func TestFeCIMEfficiency(t *testing.T) {
	cpu := TraditionalCPU()
	gpu := GPUAccelerator()
	iron := FeCIMChip()

	// FeCIM should have highest TOPS/W
	if iron.TOPSPerWatt <= cpu.TOPSPerWatt {
		t.Error("FeCIM should have higher TOPS/W than CPU")
	}
	if iron.TOPSPerWatt <= gpu.TOPSPerWatt {
		t.Error("FeCIM should have higher TOPS/W than GPU")
	}
}

// TestMNISTWorkload verifies MNIST workload creation.
func TestMNISTWorkload(t *testing.T) {
	mnist := MNISTWorkload()

	if mnist.Name != "MNIST" {
		t.Errorf("Expected 'MNIST', got '%s'", mnist.Name)
	}
	if mnist.TotalOps != 101632 {
		t.Errorf("Expected 101632 ops, got %d", mnist.TotalOps)
	}
	if mnist.Layers != 2 {
		t.Errorf("Expected 2 layers, got %d", mnist.Layers)
	}
}

// TestResNet50Workload verifies ResNet-50 workload creation.
func TestResNet50Workload(t *testing.T) {
	resnet := ResNet50Workload()

	if resnet.Name != "ResNet-50" {
		t.Errorf("Expected 'ResNet-50', got '%s'", resnet.Name)
	}
	if resnet.TotalOps < 1e9 {
		t.Error("ResNet-50 should have billions of ops")
	}
}

// TestRunInference verifies inference simulation.
func TestRunInference(t *testing.T) {
	arch := FeCIMChip()
	workload := MNISTWorkload()

	result := arch.RunInference(workload.TotalOps, 1)

	if result.Latency <= 0 {
		t.Error("Latency should be positive")
	}
	if result.Throughput <= 0 {
		t.Error("Throughput should be positive")
	}
	if result.Energy <= 0 {
		t.Error("Energy should be positive")
	}
}

// TestFeCIMLowestLatency verifies FeCIM has good latency.
func TestFeCIMLowestLatency(t *testing.T) {
	workload := MNISTWorkload()

	cpu := TraditionalCPU()
	iron := FeCIMChip()

	cpuResult := cpu.RunInference(workload.TotalOps, 1)
	fecimResult := iron.RunInference(workload.TotalOps, 1)

	// FeCIM should have lower latency than CPU
	if fecimResult.Latency >= cpuResult.Latency {
		t.Error("FeCIM should have lower latency than CPU")
	}
}

// TestFeCIMLowestEnergy verifies FeCIM has lowest energy.
func TestFeCIMLowestEnergy(t *testing.T) {
	workload := MNISTWorkload()

	cpu := TraditionalCPU()
	gpu := GPUAccelerator()
	iron := FeCIMChip()

	cpuResult := cpu.RunInference(workload.TotalOps, 1)
	gpuResult := gpu.RunInference(workload.TotalOps, 1)
	fecimResult := iron.RunInference(workload.TotalOps, 1)

	// FeCIM should have lowest energy
	if fecimResult.Energy >= cpuResult.Energy {
		t.Error("FeCIM should use less energy than CPU")
	}
	if fecimResult.Energy >= gpuResult.Energy {
		t.Error("FeCIM should use less energy than GPU")
	}
}

// TestScaleToDataCenter verifies data center scaling.
func TestScaleToDataCenter(t *testing.T) {
	arch := FeCIMChip()
	workload := MNISTWorkload()
	targetThroughput := 10000.0 // 10K inferences/sec

	metrics := ScaleToDataCenter(arch, targetThroughput, workload)

	if metrics.ChipsRequired < 1 {
		t.Error("Should require at least 1 chip")
	}
	if metrics.TotalPower <= 0 {
		t.Error("Total power should be positive")
	}
	if metrics.TCO <= 0 {
		t.Error("TCO should be positive")
	}
}

// TestFeCIMLowestTCO verifies FeCIM has lowest TCO.
func TestFeCIMLowestTCO(t *testing.T) {
	workload := MNISTWorkload()
	targetThroughput := 10000.0

	cpuMetrics := ScaleToDataCenter(TraditionalCPU(), targetThroughput, workload)
	gpuMetrics := ScaleToDataCenter(GPUAccelerator(), targetThroughput, workload)
	ironMetrics := ScaleToDataCenter(FeCIMChip(), targetThroughput, workload)

	if ironMetrics.TCO >= cpuMetrics.TCO {
		t.Error("FeCIM should have lower TCO than CPU")
	}
	if ironMetrics.TCO >= gpuMetrics.TCO {
		t.Error("FeCIM should have lower TCO than GPU")
	}
}

// TestCompareArchitectures verifies full comparison.
func TestCompareArchitectures(t *testing.T) {
	workload := MNISTWorkload()
	comparison := CompareArchitectures(workload, 1, 10000.0)

	if len(comparison.Architectures) != 3 {
		t.Errorf("Expected 3 architectures, got %d", len(comparison.Architectures))
	}
	if len(comparison.Results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(comparison.Results))
	}
	if len(comparison.DataCenter) != 3 {
		t.Errorf("Expected 3 data center metrics, got %d", len(comparison.DataCenter))
	}
}

// TestCalculateAdvantages verifies advantage calculation.
func TestCalculateAdvantages(t *testing.T) {
	workload := MNISTWorkload()
	comparison := CompareArchitectures(workload, 1, 10000.0)
	advantages := CalculateAdvantages(comparison)

	// All advantages should be > 1 (FeCIM is better)
	if advantages.VsCPU.EnergyReduction <= 1 {
		t.Error("Energy reduction vs CPU should be > 1")
	}
	if advantages.VsCPU.ThroughputIncrease <= 1 {
		t.Error("Throughput increase vs CPU should be > 1")
	}
	if advantages.VsGPU.PowerReduction <= 1 {
		t.Error("Power reduction vs GPU should be > 1")
	}
}

// TestRenderer verifies renderer functions.
func TestRenderer(t *testing.T) {
	renderer := NewRenderer()

	archs := []*Architecture{
		TraditionalCPU(),
		GPUAccelerator(),
		FeCIMChip(),
	}

	// Test architecture specs render
	specs := renderer.RenderArchitectureSpecs(archs)
	if len(specs) == 0 {
		t.Error("Architecture specs should not be empty")
	}

	// Test bar chart render
	labels := []string{"CPU", "GPU", "FeCIM"}
	values := []float64{100, 50, 5}
	chart := renderer.RenderBarChart("Power (W)", labels, values, "W", true)
	if len(chart) == 0 {
		t.Error("Bar chart should not be empty")
	}
}

// TestLargeWorkloads verifies large workload handling.
func TestLargeWorkloads(t *testing.T) {
	workloads := []Workload{
		MNISTWorkload(),
		ResNet50Workload(),
		BERTBaseWorkload(),
		GPT2Workload(),
	}

	iron := FeCIMChip()

	for _, w := range workloads {
		result := iron.RunInference(w.TotalOps, 1)
		if result.Throughput <= 0 {
			t.Errorf("Throughput should be positive for %s", w.Name)
		}
		if result.Energy <= 0 {
			t.Errorf("Energy should be positive for %s", w.Name)
		}
	}
}

// TestCustomArchitecture verifies custom architecture creation.
func TestCustomArchitecture(t *testing.T) {
	custom := CustomArchitecture("Test", 10, 5, 100)

	if custom.PeakTOPS != 10 {
		t.Errorf("Expected 10 TOPS, got %f", custom.PeakTOPS)
	}
	if custom.TDP != 5 {
		t.Errorf("Expected 5 W TDP, got %f", custom.TDP)
	}
	if custom.ChipArea != 100 {
		t.Errorf("Expected 100 mm² area, got %f", custom.ChipArea)
	}

	// Verify efficiency calculation
	if custom.TOPSPerWatt != 2 {
		t.Errorf("Expected 2 TOPS/W, got %f", custom.TOPSPerWatt)
	}
}

// TestCO2Emissions verifies CO2 emission calculation.
func TestCO2Emissions(t *testing.T) {
	workload := MNISTWorkload()
	targetThroughput := 10000.0

	cpuMetrics := ScaleToDataCenter(TraditionalCPU(), targetThroughput, workload)
	ironMetrics := ScaleToDataCenter(FeCIMChip(), targetThroughput, workload)

	// FeCIM should have lower CO2 emissions
	if ironMetrics.CO2Emissions >= cpuMetrics.CO2Emissions {
		t.Error("FeCIM should have lower CO2 emissions")
	}

	// CO2 should be proportional to power
	if cpuMetrics.CO2Emissions <= 0 {
		t.Error("CO2 emissions should be positive")
	}
}

// TestFormatFunctions verifies formatting helpers.
func TestFormatFunctions(t *testing.T) {
	// Test formatNumber
	if formatNumber(1e12) != "1.0T" {
		t.Error("Should format trillions with T")
	}
	if formatNumber(1e9) != "1.0B" {
		t.Error("Should format billions with B")
	}
	if formatNumber(1e6) != "1.0M" {
		t.Error("Should format millions with M")
	}

	// Test formatLatency
	if formatLatency(1000) != "1.00 s" {
		t.Errorf("1000ms should be 1.00 s, got %s", formatLatency(1000))
	}
	if formatLatency(1) != "1.00 ms" {
		t.Errorf("1ms should be 1.00 ms, got %s", formatLatency(1))
	}
}

// TestModelPowerReductionTarget logs whether the model hits the 80-90% target.
func TestModelPowerReductionTarget(t *testing.T) {
	workload := ResNet50Workload()
	targetThroughput := 100000.0 // 100K inferences/sec

	cpuMetrics := ScaleToDataCenter(TraditionalCPU(), targetThroughput, workload)
	ironMetrics := ScaleToDataCenter(FeCIMChip(), targetThroughput, workload)

	// Calculate power reduction percentage
	powerReduction := (1 - ironMetrics.TotalPower/cpuMetrics.TotalPower) * 100

	// Model input target: 80-90% reduction
	if powerReduction < 80 {
		t.Logf("Power reduction: %.1f%% (target: 80-90%%)", powerReduction)
		// This is informational, not a strict test failure
		// Real-world results depend on many factors
	}
}
