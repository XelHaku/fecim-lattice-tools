// Package simulation provides the time-stepping simulation engine.
package simulation

import (
	"math"
	"sync"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

// State represents the current simulation state.
type State struct {
	Time          float64 // Simulation time (s)
	Voltage       float64 // Applied voltage (V)
	ElectricField float64 // Electric field (V/m)
	Polarization  float64 // Current polarization (C/m²)
	NormPol       float64 // Normalized polarization (-1 to +1)

	// History for plotting
	VoltageHistory []float64
	PolHistory     []float64
	MaxHistory     int
}

// Engine manages the ferroelectric simulation.
type Engine struct {
	model    *ferroelectric.PreisachModel
	material *ferroelectric.HZOMaterial
	state    *State

	// Simulation parameters
	dt float64 // Time step (s)

	// Thread-safe state (protected by mu)
	mu      sync.RWMutex
	running bool
	paused  bool

	// Waveform generation
	waveform  WaveformType
	frequency float64 // Hz
	amplitude float64 // V
}

// WaveformType defines the input voltage waveform.
type WaveformType int

const (
	WaveformSine WaveformType = iota
	WaveformTriangle
	WaveformSquare
	WaveformManual
)

// NewEngine creates a new simulation engine.
func NewEngine(material *ferroelectric.HZOMaterial) *Engine {
	model := ferroelectric.NewPreisachModel(material)

	return &Engine{
		model:     model,
		material:  material,
		state:     newState(1000),
		dt:        1e-9, // 1 ns time step
		waveform:  WaveformSine,
		frequency: 1e6, // 1 MHz default
		amplitude: material.CoerciveVoltage() * 2,
	}
}

func newState(maxHistory int) *State {
	return &State{
		VoltageHistory: make([]float64, 0, maxHistory),
		PolHistory:     make([]float64, 0, maxHistory),
		MaxHistory:     maxHistory,
	}
}

// Start begins the simulation loop.
func (e *Engine) Start() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = true
	e.paused = false
}

// Stop halts the simulation.
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.running = false
}

// Pause toggles the paused state.
func (e *Engine) Pause() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.paused = !e.paused
}

// IsPaused returns true if simulation is paused.
func (e *Engine) IsPaused() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.paused
}

// IsRunning returns true if simulation is running.
func (e *Engine) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.running
}

// Step advances the simulation by one time step.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) Step() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.running || e.paused {
		return
	}

	// Generate voltage based on waveform
	e.state.Voltage = e.generateVoltage(e.state.Time)

	// Convert voltage to electric field
	e.state.ElectricField = e.state.Voltage / e.material.Thickness

	// Update polarization via Preisach model
	e.state.Polarization = e.model.Update(e.state.ElectricField)
	e.state.NormPol = e.model.NormalizedPolarization()

	// Record history
	e.recordHistory()

	// Advance time
	e.state.Time += e.dt
}

// generateVoltage produces the input voltage at time t.
func (e *Engine) generateVoltage(t float64) float64 {
	if e.waveform == WaveformManual {
		return e.state.Voltage // Use manually set value
	}

	omega := 2 * math.Pi * e.frequency
	phase := omega * t

	switch e.waveform {
	case WaveformSine:
		return e.amplitude * math.Sin(phase)

	case WaveformTriangle:
		// Triangle wave from -A to +A
		p := math.Mod(phase, 2*math.Pi) / (2 * math.Pi)
		if p < 0.25 {
			return e.amplitude * (4 * p)
		} else if p < 0.75 {
			return e.amplitude * (2 - 4*p)
		} else {
			return e.amplitude * (4*p - 4)
		}

	case WaveformSquare:
		if math.Sin(phase) >= 0 {
			return e.amplitude
		}
		return -e.amplitude

	default:
		return 0
	}
}

// recordHistory saves current state for plotting.
func (e *Engine) recordHistory() {
	s := e.state

	// Add new values
	s.VoltageHistory = append(s.VoltageHistory, s.Voltage)
	s.PolHistory = append(s.PolHistory, s.NormPol)

	// Trim if too long
	if len(s.VoltageHistory) > s.MaxHistory {
		s.VoltageHistory = s.VoltageHistory[1:]
		s.PolHistory = s.PolHistory[1:]
	}
}

// SetVoltage manually sets the voltage (for WaveformManual mode).
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetVoltage(v float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.Voltage = v
}

// SetWaveform changes the voltage waveform type.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetWaveform(w WaveformType) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.waveform = w
}

// SetFrequency changes the waveform frequency.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetFrequency(f float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.frequency = f
}

// SetAmplitude changes the waveform amplitude.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) SetAmplitude(a float64) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.amplitude = a
}

// State returns a copy of the current simulation state.
// Thread-safe: returns a copy to prevent data races.
func (e *Engine) State() State {
	e.mu.RLock()
	defer e.mu.RUnlock()
	// Return a copy of the state to prevent race conditions
	stateCopy := *e.state
	// Deep copy the history slices
	stateCopy.VoltageHistory = make([]float64, len(e.state.VoltageHistory))
	copy(stateCopy.VoltageHistory, e.state.VoltageHistory)
	stateCopy.PolHistory = make([]float64, len(e.state.PolHistory))
	copy(stateCopy.PolHistory, e.state.PolHistory)
	return stateCopy
}

// Reset clears the simulation state.
// Thread-safe: uses mutex to protect state modifications.
func (e *Engine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.model.Reset()
	e.state = newState(e.state.MaxHistory)
}

// GetHysteresisData returns P-E data for plotting the hysteresis loop.
func (e *Engine) GetHysteresisData() ([]float64, []float64) {
	Emax := e.amplitude / e.material.Thickness
	return e.model.GetHysteresisLoop(Emax, 100)
}

// RunRealtime runs the simulation in real-time with the given callback.
func (e *Engine) RunRealtime(updateCallback func(*State), targetFPS int) {
	ticker := time.NewTicker(time.Second / time.Duration(targetFPS))
	defer ticker.Stop()

	stepsPerFrame := int(1.0 / (e.dt * float64(targetFPS)))

	for range ticker.C {
		if !e.IsRunning() {
			break
		}

		if !e.IsPaused() {
			// Run multiple physics steps per frame
			for i := 0; i < stepsPerFrame; i++ {
				e.Step()
			}
		}

		// Call the update callback
		if updateCallback != nil {
			e.mu.RLock()
			state := e.state
			e.mu.RUnlock()
			updateCallback(state)
		}
	}
}
