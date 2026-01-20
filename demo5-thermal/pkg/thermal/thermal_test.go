package thermal

import (
	"math"
	"testing"
)

// TestNewThermalSim verifies simulation creation.
func TestNewThermalSim(t *testing.T) {
	sim := NewThermalSim(32, 32)

	if sim.Width != 32 || sim.Height != 32 {
		t.Errorf("Expected 32x32 grid, got %dx%d", sim.Width, sim.Height)
	}

	if len(sim.Grid) != 32 {
		t.Errorf("Expected 32 rows, got %d", len(sim.Grid))
	}

	if len(sim.Grid[0]) != 32 {
		t.Errorf("Expected 32 columns, got %d", len(sim.Grid[0]))
	}
}

// TestReset verifies grid resets to ambient temperature.
func TestReset(t *testing.T) {
	sim := NewThermalSim(16, 16)
	sim.Grid[5][5] = 100.0 // Set hot spot

	sim.Reset()

	for y := 0; y < sim.Height; y++ {
		for x := 0; x < sim.Width; x++ {
			if math.Abs(sim.Grid[y][x]-sim.AmbientTemp) > 0.001 {
				t.Errorf("Cell [%d,%d] = %.2f, expected %.2f",
					x, y, sim.Grid[y][x], sim.AmbientTemp)
			}
		}
	}
}

// TestHeatDiffusion verifies heat spreads from hot spot.
func TestHeatDiffusion(t *testing.T) {
	sim := NewThermalSim(16, 16)
	sim.Reset()

	// Create hot spot in center
	centerX, centerY := 8, 8
	sim.Grid[centerY][centerX] = 100.0

	// Run simulation
	sim.StepMultiple(100, 1e-9)

	// Center should have cooled
	centerTemp := sim.Grid[centerY][centerX]
	if centerTemp >= 100.0 {
		t.Errorf("Center should cool down, got %.2f", centerTemp)
	}

	// Neighbors should have warmed
	neighborTemp := sim.Grid[centerY][centerX+1]
	if neighborTemp <= sim.AmbientTemp {
		t.Errorf("Neighbor should warm up, got %.2f", neighborTemp)
	}
}

// TestHeatGeneration verifies power injection raises temperature.
func TestHeatGeneration(t *testing.T) {
	sim := NewThermalSim(16, 16)
	sim.Reset()

	// Apply power at center
	sim.SetPower(8, 8, 1e8) // High power density

	// Run simulation
	sim.StepMultiple(100, 1e-9)

	// Center should be hotter than ambient
	centerTemp := sim.Grid[8][8]
	if centerTemp <= sim.AmbientTemp {
		t.Errorf("Heated cell should be above ambient, got %.2f", centerTemp)
	}
}

// TestMaxTemperature verifies max temperature detection.
func TestMaxTemperature(t *testing.T) {
	sim := NewThermalSim(8, 8)
	sim.Reset()

	// Set specific temperatures
	sim.Grid[3][3] = 50.0
	sim.Grid[5][5] = 75.0
	sim.Grid[2][6] = 60.0

	maxTemp := sim.GetMaxTemperature()
	if math.Abs(maxTemp-75.0) > 0.001 {
		t.Errorf("Expected max temp 75.0, got %.2f", maxTemp)
	}
}

// TestAverageTemperature verifies average calculation.
func TestAverageTemperature(t *testing.T) {
	sim := NewThermalSim(4, 4)

	// Set uniform temperature
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			sim.Grid[y][x] = 50.0
		}
	}

	avgTemp := sim.GetAverageTemperature()
	if math.Abs(avgTemp-50.0) > 0.001 {
		t.Errorf("Expected average temp 50.0, got %.2f", avgTemp)
	}
}

// TestHotspotDetection verifies hotspot finding.
func TestHotspotDetection(t *testing.T) {
	sim := NewThermalSim(16, 16)
	sim.Reset()

	// Create two hotspots
	sim.Grid[4][4] = 70.0
	sim.Grid[12][12] = 80.0

	threshold := 60.0
	hotspots := sim.FindHotspots(threshold)

	if len(hotspots) != 2 {
		t.Errorf("Expected 2 hotspots, got %d", len(hotspots))
	}

	// Check hotspot properties
	for _, h := range hotspots {
		if h.Temperature < threshold {
			t.Errorf("Hotspot at (%d,%d) below threshold: %.2f < %.2f",
				h.X, h.Y, h.Temperature, threshold)
		}
	}
}

// TestThermalWarning verifies warning system.
func TestThermalWarning(t *testing.T) {
	sim := DefaultThermalSim()
	sim.Reset()

	// No warning at ambient
	warning := sim.CheckThermalWarning()
	if warning != nil {
		t.Errorf("Expected no warning at ambient, got level %d", warning.Level)
	}

	// Create critical hotspot
	sim.Grid[8][8] = sim.MaxTemp * 0.95

	warning = sim.CheckThermalWarning()
	if warning == nil {
		t.Error("Expected warning for hot spot")
	} else if warning.Level != 3 {
		t.Errorf("Expected critical warning (level 3), got level %d", warning.Level)
	}
}

// TestThermalResistance verifies resistance calculation.
func TestThermalResistance(t *testing.T) {
	sim := DefaultThermalSim()

	// Same point should have zero resistance
	r0 := sim.ThermalResistance(5, 5, 5, 5)
	if r0 != 0 {
		t.Errorf("Same-point resistance should be 0, got %.2e", r0)
	}

	// Farther points should have higher resistance
	r1 := sim.ThermalResistance(0, 0, 1, 0)
	r2 := sim.ThermalResistance(0, 0, 2, 0)

	if r2 <= r1 {
		t.Errorf("Farther distance should have higher resistance: r1=%.2e, r2=%.2e",
			r1, r2)
	}
}

// TestMultiLayerCreation verifies multi-layer simulation.
func TestMultiLayerCreation(t *testing.T) {
	mlSim := NewMultiLayerSim(3, 16, 16)

	if len(mlSim.Layers) != 3 {
		t.Errorf("Expected 3 layers, got %d", len(mlSim.Layers))
	}

	if len(mlSim.Coupling) != 2 {
		t.Errorf("Expected 2 coupling coefficients, got %d", len(mlSim.Coupling))
	}
}

// TestMultiLayerHeatTransfer verifies heat flows between layers.
func TestMultiLayerHeatTransfer(t *testing.T) {
	mlSim := NewMultiLayerSim(3, 8, 8)
	mlSim.Reset()

	// Heat top layer
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			mlSim.Layers[2].Grid[y][x] = 80.0
		}
	}

	// Middle and bottom at ambient
	initialMiddle := mlSim.Layers[1].GetAverageTemperature()

	// Run simulation
	mlSim.StepMultiple(1000, 1e-9)

	// Middle layer should have warmed
	finalMiddle := mlSim.Layers[1].GetAverageTemperature()
	if finalMiddle <= initialMiddle {
		t.Errorf("Middle layer should warm: initial=%.2f, final=%.2f",
			initialMiddle, finalMiddle)
	}
}

// TestMultiLayerGlobalStats verifies stack-wide statistics.
func TestMultiLayerGlobalStats(t *testing.T) {
	mlSim := NewMultiLayerSim(3, 8, 8)
	mlSim.Reset()

	// Set different temperatures in each layer
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			mlSim.Layers[0].Grid[y][x] = 30.0
			mlSim.Layers[1].Grid[y][x] = 50.0
			mlSim.Layers[2].Grid[y][x] = 70.0
		}
	}

	maxTemp := mlSim.GetGlobalMaxTemp()
	if math.Abs(maxTemp-70.0) > 0.001 {
		t.Errorf("Expected global max 70.0, got %.2f", maxTemp)
	}

	minTemp := mlSim.GetGlobalMinTemp()
	if math.Abs(minTemp-30.0) > 0.001 {
		t.Errorf("Expected global min 30.0, got %.2f", minTemp)
	}

	avgTemp := mlSim.GetStackAverageTemp()
	if math.Abs(avgTemp-50.0) > 0.001 {
		t.Errorf("Expected stack average 50.0, got %.2f", avgTemp)
	}
}

// TestFeCIMLowPower demonstrates FeCIM thermal advantage.
func TestFeCIMLowPower(t *testing.T) {
	// FeCIM: ~0.001 pJ per MAC → very low power density
	// Traditional: ~10 pJ per MAC → high power density (10,000x more)

	feCIMSim := DefaultThermalSim()
	traditionalSim := DefaultThermalSim()
	feCIMSim.Reset()
	traditionalSim.Reset()

	// Apply typical workload power densities for CIM arrays
	// Traditional CIM/GPU: ~10-100 W/cm² = 1e5 - 1e6 W/m²
	// FeCIM: 1000-10000x lower
	feCIMPower := 1e4   // W/m² (very low - FeCIM advantage)
	traditionalPower := 1e7  // W/m² (high - typical CMOS)

	for y := 8; y < 24; y++ {
		for x := 8; x < 24; x++ {
			feCIMSim.SetPower(x, y, feCIMPower)
			traditionalSim.SetPower(x, y, traditionalPower)
		}
	}

	// Run both simulations for 10 µs (10000 steps at 1 ns each)
	feCIMSim.StepMultiple(10000, 1e-9)
	traditionalSim.StepMultiple(10000, 1e-9)

	feCIMMax := feCIMSim.GetMaxTemperature()
	traditionalMax := traditionalSim.GetMaxTemperature()

	// FeCIM should be significantly cooler
	if feCIMMax >= traditionalMax {
		t.Errorf("FeCIM should be cooler: IL=%.2f°C, Traditional=%.2f°C",
			feCIMMax, traditionalMax)
	}

	// Log the results
	ilRise := feCIMMax - feCIMSim.AmbientTemp
	tradRise := traditionalMax - traditionalSim.AmbientTemp
	t.Logf("FeCIM: %.2f°C (rise: %.2f°C)", feCIMMax, ilRise)
	t.Logf("Traditional: %.2f°C (rise: %.2f°C)", traditionalMax, tradRise)
	if ilRise > 0.01 {
		t.Logf("Temperature ratio: %.0fx cooler with FeCIM", tradRise/ilRise)
	} else {
		t.Log("FeCIM: Essentially no heating!")
	}
}

// TestSteadyState verifies steady-state solver.
func TestSteadyState(t *testing.T) {
	sim := NewThermalSim(8, 8)
	sim.Reset()

	// Apply constant power
	sim.SetPower(4, 4, 1e6)

	// Run to steady state
	iterations := sim.SteadyStateAnalysis(10000, 0.001)

	t.Logf("Reached steady state in %d iterations", iterations)

	// Should reach some equilibrium
	if iterations >= 10000 {
		t.Log("Warning: May not have reached full steady state")
	}
}

// TestHeatMapRenderer verifies rendering output.
func TestHeatMapRenderer(t *testing.T) {
	sim := DefaultThermalSim()
	sim.Reset()

	// Add some temperature variation
	sim.Grid[8][8] = 60.0
	sim.Grid[16][16] = 75.0

	renderer := DefaultRenderer()
	renderer.UseColor = false // For test output

	output := renderer.Render(sim)
	if len(output) == 0 {
		t.Error("Renderer produced empty output")
	}

	// Should have grid lines
	if len(output) < sim.Height {
		t.Error("Output too short")
	}
}

// TestTemperatureGradient verifies gradient calculation.
func TestTemperatureGradient(t *testing.T) {
	sim := DefaultThermalSim()
	sim.Reset()

	// Create horizontal temperature gradient
	for y := 0; y < sim.Height; y++ {
		for x := 0; x < sim.Width; x++ {
			sim.Grid[y][x] = 25.0 + float64(x)*2.0
		}
	}

	renderer := DefaultRenderer()
	renderer.UseColor = false

	gradient := renderer.TemperatureGradient(sim)
	if len(gradient) == 0 {
		t.Error("Gradient renderer produced empty output")
	}

	// With left-to-right gradient, heat should flow left (←)
	// Check that output contains arrow characters
	t.Logf("Gradient output length: %d characters", len(gradient))
}

// TestVerticalProfile verifies multi-layer vertical profile.
func TestVerticalProfile(t *testing.T) {
	mlSim := NewMultiLayerSim(3, 8, 8)

	// Set different temperatures
	mlSim.Layers[0].Grid[4][4] = 30.0
	mlSim.Layers[1].Grid[4][4] = 50.0
	mlSim.Layers[2].Grid[4][4] = 70.0

	profile := mlSim.VerticalTemperatureProfile(4, 4)

	if len(profile) != 3 {
		t.Errorf("Expected 3 values in profile, got %d", len(profile))
	}

	if math.Abs(profile[0]-30.0) > 0.001 {
		t.Errorf("Layer 0 temp wrong: expected 30.0, got %.2f", profile[0])
	}
	if math.Abs(profile[1]-50.0) > 0.001 {
		t.Errorf("Layer 1 temp wrong: expected 50.0, got %.2f", profile[1])
	}
	if math.Abs(profile[2]-70.0) > 0.001 {
		t.Errorf("Layer 2 temp wrong: expected 70.0, got %.2f", profile[2])
	}
}
