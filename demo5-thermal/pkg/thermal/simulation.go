package thermal

import (
	"math"
	"sync"
)

// ThermalSim represents a 2D thermal simulation grid.
// Implements heat diffusion equation: dT/dt = α∇²T + Q/(ρ*cp)
// where α is thermal diffusivity and Q is heat generation rate.
type ThermalSim struct {
	Grid              [][]float64 // Temperature grid (°C)
	PowerMap          [][]float64 // Heat generation map (W/m²)
	Width             int         // Grid width
	Height            int         // Grid height
	Conductivity      float64     // Thermal conductivity (W/m·K)
	Diffusivity       float64     // Thermal diffusivity (m²/s)
	VolumetricHeatCap float64     // ρ*cp volumetric heat capacity (J/m³·K)
	AmbientTemp       float64     // Ambient temperature (°C)
	MaxTemp           float64     // Maximum safe operating temperature (°C)
	CellSize          float64     // Size of each grid cell (m)
	mu                sync.RWMutex
}

// LayerConfig holds thermal properties for a single layer.
type LayerConfig struct {
	Name         string  // Layer name
	Thickness    float64 // Layer thickness (m)
	Conductivity float64 // Thermal conductivity (W/m·K)
	Diffusivity  float64 // Thermal diffusivity (m²/s)
	HeatGen      float64 // Base heat generation (W/m²)
}

// HotspotInfo contains information about a detected hotspot.
type HotspotInfo struct {
	X           int     // X coordinate
	Y           int     // Y coordinate
	Temperature float64 // Temperature at hotspot (°C)
	Severity    float64 // Severity (0-1, where 1 is at MaxTemp)
}

// ThermalWarning represents a thermal warning condition.
type ThermalWarning struct {
	Level       int     // Warning level (1-3)
	Message     string  // Warning message
	MaxTemp     float64 // Maximum temperature detected
	AverageTemp float64 // Average temperature
	Hotspots    int     // Number of hotspots
}

// NewThermalSim creates a new thermal simulation with given dimensions.
func NewThermalSim(width, height int) *ThermalSim {
	grid := make([][]float64, height)
	powerMap := make([][]float64, height)
	for i := range grid {
		grid[i] = make([]float64, width)
		powerMap[i] = make([]float64, width)
	}

	return &ThermalSim{
		Grid:              grid,
		PowerMap:          powerMap,
		Width:             width,
		Height:            height,
		Conductivity:      150.0,  // Silicon ~150 W/m·K
		Diffusivity:       8.8e-5, // Silicon thermal diffusivity
		VolumetricHeatCap: 1.63e6, // Silicon ρ*cp ≈ 2330*700 J/m³·K
		AmbientTemp:       25.0,   // Room temperature
		MaxTemp:           85.0,   // Typical max for CMOS
		CellSize:          1e-6,   // 1 µm grid cells
	}
}

// DefaultThermalSim creates a thermal simulation sized for a crossbar array.
func DefaultThermalSim() *ThermalSim {
	sim := NewThermalSim(32, 32)
	sim.Reset()
	return sim
}

// Reset sets the entire grid to ambient temperature.
func (t *ThermalSim) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			t.Grid[y][x] = t.AmbientTemp
			t.PowerMap[y][x] = 0
		}
	}
}

// SetPower sets heat generation at a specific location.
func (t *ThermalSim) SetPower(x, y int, power float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if x >= 0 && x < t.Width && y >= 0 && y < t.Height {
		t.PowerMap[y][x] = power
	}
}

// SetPowerMap sets the entire power map at once.
func (t *ThermalSim) SetPowerMap(powerMap [][]float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for y := 0; y < t.Height && y < len(powerMap); y++ {
		for x := 0; x < t.Width && x < len(powerMap[y]); x++ {
			t.PowerMap[y][x] = powerMap[y][x]
		}
	}
}

// Step advances the simulation by dt seconds using explicit finite difference.
func (t *ThermalSim) Step(dt float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Create temporary grid for next state
	newGrid := make([][]float64, t.Height)
	for i := range newGrid {
		newGrid[i] = make([]float64, t.Width)
	}

	// Stability condition: dt <= dx²/(4α)
	maxDt := t.CellSize * t.CellSize / (4 * t.Diffusivity)
	if dt > maxDt {
		dt = maxDt
	}

	// Heat diffusion coefficient
	alpha := t.Diffusivity * dt / (t.CellSize * t.CellSize)

	// Apply heat diffusion equation with boundary conditions
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			// Current temperature
			T := t.Grid[y][x]

			// Laplacian (∇²T) using 5-point stencil with Neumann boundary conditions
			var Tup, Tdown, Tleft, Tright float64

			if y > 0 {
				Tup = t.Grid[y-1][x]
			} else {
				Tup = T // Adiabatic boundary
			}
			if y < t.Height-1 {
				Tdown = t.Grid[y+1][x]
			} else {
				Tdown = t.AmbientTemp // Bottom: heat sink
			}
			if x > 0 {
				Tleft = t.Grid[y][x-1]
			} else {
				Tleft = T // Adiabatic boundary
			}
			if x < t.Width-1 {
				Tright = t.Grid[y][x+1]
			} else {
				Tright = T // Adiabatic boundary
			}

			laplacian := Tup + Tdown + Tleft + Tright - 4*T

			// Heat generation term: dT/dt = Q'' / (ρ*c_p * thickness)
			// PowerMap is W/m² (heat flux), divide by volumetric heat capacity and thickness
			// This gives heating rate in K/s
			heatGen := t.PowerMap[y][x] / (t.VolumetricHeatCap * t.CellSize)

			// Update temperature
			newGrid[y][x] = T + alpha*laplacian + heatGen*dt

			// Clamp to physical limits
			if newGrid[y][x] < t.AmbientTemp {
				newGrid[y][x] = t.AmbientTemp
			}
		}
	}

	// Copy new grid to current grid
	for y := 0; y < t.Height; y++ {
		copy(t.Grid[y], newGrid[y])
	}
}

// StepMultiple runs multiple simulation steps.
func (t *ThermalSim) StepMultiple(steps int, dt float64) {
	for i := 0; i < steps; i++ {
		t.Step(dt)
	}
}

// GetTemperature returns the temperature at a specific location.
func (t *ThermalSim) GetTemperature(x, y int) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if x >= 0 && x < t.Width && y >= 0 && y < t.Height {
		return t.Grid[y][x]
	}
	return t.AmbientTemp
}

// GetMaxTemperature returns the maximum temperature in the grid.
func (t *ThermalSim) GetMaxTemperature() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	maxTemp := t.AmbientTemp
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			if t.Grid[y][x] > maxTemp {
				maxTemp = t.Grid[y][x]
			}
		}
	}
	return maxTemp
}

// GetMinTemperature returns the minimum temperature in the grid.
func (t *ThermalSim) GetMinTemperature() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	minTemp := math.MaxFloat64
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			if t.Grid[y][x] < minTemp {
				minTemp = t.Grid[y][x]
			}
		}
	}
	return minTemp
}

// GetAverageTemperature returns the average temperature across the grid.
func (t *ThermalSim) GetAverageTemperature() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	sum := 0.0
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			sum += t.Grid[y][x]
		}
	}
	return sum / float64(t.Width*t.Height)
}

// FindHotspots identifies cells above a temperature threshold.
func (t *ThermalSim) FindHotspots(threshold float64) []HotspotInfo {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var hotspots []HotspotInfo
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			if t.Grid[y][x] > threshold {
				severity := (t.Grid[y][x] - threshold) / (t.MaxTemp - threshold)
				if severity > 1.0 {
					severity = 1.0
				}
				hotspots = append(hotspots, HotspotInfo{
					X:           x,
					Y:           y,
					Temperature: t.Grid[y][x],
					Severity:    severity,
				})
			}
		}
	}
	return hotspots
}

// CheckThermalWarning checks for thermal throttling conditions.
func (t *ThermalSim) CheckThermalWarning() *ThermalWarning {
	maxTemp := t.GetMaxTemperature()
	avgTemp := t.GetAverageTemperature()

	// Warning thresholds
	warnThreshold := t.MaxTemp * 0.75  // 75% of max → warning
	critThreshold := t.MaxTemp * 0.90  // 90% of max → critical
	alertThreshold := t.MaxTemp * 0.60 // 60% of max → alert

	hotspots := t.FindHotspots(alertThreshold)

	if maxTemp > critThreshold {
		return &ThermalWarning{
			Level:       3,
			Message:     "CRITICAL: Thermal throttling required!",
			MaxTemp:     maxTemp,
			AverageTemp: avgTemp,
			Hotspots:    len(hotspots),
		}
	} else if maxTemp > warnThreshold {
		return &ThermalWarning{
			Level:       2,
			Message:     "WARNING: Approaching thermal limit",
			MaxTemp:     maxTemp,
			AverageTemp: avgTemp,
			Hotspots:    len(hotspots),
		}
	} else if len(hotspots) > 0 {
		return &ThermalWarning{
			Level:       1,
			Message:     "ALERT: Hotspots detected",
			MaxTemp:     maxTemp,
			AverageTemp: avgTemp,
			Hotspots:    len(hotspots),
		}
	}
	return nil
}

// GetGridCopy returns a copy of the current temperature grid.
func (t *ThermalSim) GetGridCopy() [][]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	gridCopy := make([][]float64, t.Height)
	for y := 0; y < t.Height; y++ {
		gridCopy[y] = make([]float64, t.Width)
		copy(gridCopy[y], t.Grid[y])
	}
	return gridCopy
}

// ThermalResistance calculates thermal resistance between two points.
func (t *ThermalSim) ThermalResistance(x1, y1, x2, y2 int) float64 {
	// R_th = L / (k * A)
	// For grid: R = distance * cellSize / (k * cellSize²) = distance / (k * cellSize)
	dx := float64(x2 - x1)
	dy := float64(y2 - y1)
	distance := math.Sqrt(dx*dx+dy*dy) * t.CellSize
	area := t.CellSize * t.CellSize
	return distance / (t.Conductivity * area)
}

// TotalHeatGeneration returns total heat being generated in the grid.
func (t *ThermalSim) TotalHeatGeneration() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	total := 0.0
	for y := 0; y < t.Height; y++ {
		for x := 0; x < t.Width; x++ {
			total += t.PowerMap[y][x]
		}
	}
	return total
}

// SteadyStateAnalysis estimates steady-state temperature given current power map.
// Uses iterative solver to reach equilibrium.
func (t *ThermalSim) SteadyStateAnalysis(maxIterations int, tolerance float64) int {
	iterations := 0
	for iterations < maxIterations {
		oldMax := t.GetMaxTemperature()
		t.Step(1e-6) // Small time step
		newMax := t.GetMaxTemperature()

		if math.Abs(newMax-oldMax) < tolerance {
			break
		}
		iterations++
	}
	return iterations
}
