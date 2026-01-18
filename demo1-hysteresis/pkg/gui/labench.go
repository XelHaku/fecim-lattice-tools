// Package gui provides the Lab Bench interactive control interface.
//
// The Lab Bench allows real-time manipulation of simulation parameters
// via keyboard controls, with visual feedback displayed on screen.
//
// Controls:
//   E/D - Increase/Decrease Electric Field amplitude
//   T/G - Increase/Decrease Temperature (modifies Landau coefficients)
//   W   - Cycle waveform (Sine, Triangle, Square)
//   F/V - Increase/Decrease Frequency
//   R   - Reset simulation
//   Space - Pause/Resume
//   Q   - Quit
package gui

import (
	"fmt"
	"math"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// LabBench provides interactive control over simulation parameters.
// It bridges the GUI input to the physics engine in real-time.
type LabBench struct {
	// Electric field control (V/m)
	ElectricFieldAmplitude float64 // Current amplitude
	EFieldMin              float64 // Minimum allowed
	EFieldMax              float64 // Maximum allowed
	EFieldStep             float64 // Step size per keypress

	// Temperature control (K)
	Temperature float64 // Current temperature
	TempMin     float64 // Minimum (typically 0K)
	TempMax     float64 // Maximum (above Tc)
	TempStep    float64 // Step size per keypress

	// Frequency control (Hz)
	Frequency    float64
	FreqMin      float64
	FreqMax      float64
	FreqMultStep float64 // Multiplicative step (e.g., 2x)

	// Waveform control
	WaveformIndex int
	WaveformNames []string

	// State
	Paused  bool
	Running bool

	// Callbacks - invoked when parameters change
	OnEFieldChange     func(float64)
	OnTemperatureChange func(float64)
	OnFrequencyChange  func(float64)
	OnWaveformChange   func(int)
	OnPauseToggle      func(bool)
	OnReset            func()
	OnQuit             func()

	// Derived values (for display)
	LandauAlpha float64 // Current α coefficient (from temperature)
	CoerciveField float64 // Current Ec (temperature-dependent)
}

// NewLabBench creates a Lab Bench with default parameters for HZO simulation.
func NewLabBench() *LabBench {
	return &LabBench{
		// Electric field: 0 to 3 MV/cm, starting at 1.5 MV/cm
		ElectricFieldAmplitude: 1.5e8,
		EFieldMin:              0,
		EFieldMax:              3e8,
		EFieldStep:             0.1e8, // 0.1 MV/cm steps

		// Temperature: 100K to 800K, starting at 300K (room temperature)
		Temperature: 300,
		TempMin:     100,
		TempMax:     800,
		TempStep:    25, // 25K steps

		// Frequency: 1kHz to 100MHz, starting at 1MHz
		Frequency:    1e6,
		FreqMin:      1e3,
		FreqMax:      1e8,
		FreqMultStep: 2.0, // Double or halve

		// Waveforms
		WaveformIndex: 0,
		WaveformNames: []string{"Sine", "Triangle", "Square"},

		// State
		Running: true,
		Paused:  false,
	}
}

// HandleKeyPress processes a GLFW key event and updates parameters.
// Returns true if the event was handled.
func (lb *LabBench) HandleKeyPress(key glfw.Key, action glfw.Action) bool {
	if action != glfw.Press && action != glfw.Repeat {
		return false
	}

	switch key {
	// Electric Field: E = increase, D = decrease
	case glfw.KeyE:
		lb.increaseEField()
		return true
	case glfw.KeyD:
		lb.decreaseEField()
		return true

	// Temperature: T = increase, G = decrease
	case glfw.KeyT:
		lb.increaseTemperature()
		return true
	case glfw.KeyG:
		lb.decreaseTemperature()
		return true

	// Frequency: F = increase, V = decrease
	case glfw.KeyF:
		lb.increaseFrequency()
		return true
	case glfw.KeyV:
		lb.decreaseFrequency()
		return true

	// Waveform: W = cycle
	case glfw.KeyW:
		lb.cycleWaveform()
		return true

	// Pause: Space
	case glfw.KeySpace:
		lb.togglePause()
		return true

	// Reset: R
	case glfw.KeyR:
		lb.reset()
		return true

	// Quit: Q or Escape
	case glfw.KeyQ, glfw.KeyEscape:
		lb.quit()
		return true
	}

	return false
}

// increaseEField steps up the electric field amplitude.
func (lb *LabBench) increaseEField() {
	lb.ElectricFieldAmplitude = math.Min(lb.ElectricFieldAmplitude+lb.EFieldStep, lb.EFieldMax)
	if lb.OnEFieldChange != nil {
		lb.OnEFieldChange(lb.ElectricFieldAmplitude)
	}
}

// decreaseEField steps down the electric field amplitude.
func (lb *LabBench) decreaseEField() {
	lb.ElectricFieldAmplitude = math.Max(lb.ElectricFieldAmplitude-lb.EFieldStep, lb.EFieldMin)
	if lb.OnEFieldChange != nil {
		lb.OnEFieldChange(lb.ElectricFieldAmplitude)
	}
}

// increaseTemperature steps up the temperature.
func (lb *LabBench) increaseTemperature() {
	lb.Temperature = math.Min(lb.Temperature+lb.TempStep, lb.TempMax)
	if lb.OnTemperatureChange != nil {
		lb.OnTemperatureChange(lb.Temperature)
	}
}

// decreaseTemperature steps down the temperature.
func (lb *LabBench) decreaseTemperature() {
	lb.Temperature = math.Max(lb.Temperature-lb.TempStep, lb.TempMin)
	if lb.OnTemperatureChange != nil {
		lb.OnTemperatureChange(lb.Temperature)
	}
}

// increaseFrequency doubles the frequency.
func (lb *LabBench) increaseFrequency() {
	lb.Frequency = math.Min(lb.Frequency*lb.FreqMultStep, lb.FreqMax)
	if lb.OnFrequencyChange != nil {
		lb.OnFrequencyChange(lb.Frequency)
	}
}

// decreaseFrequency halves the frequency.
func (lb *LabBench) decreaseFrequency() {
	lb.Frequency = math.Max(lb.Frequency/lb.FreqMultStep, lb.FreqMin)
	if lb.OnFrequencyChange != nil {
		lb.OnFrequencyChange(lb.Frequency)
	}
}

// cycleWaveform moves to the next waveform type.
func (lb *LabBench) cycleWaveform() {
	lb.WaveformIndex = (lb.WaveformIndex + 1) % len(lb.WaveformNames)
	if lb.OnWaveformChange != nil {
		lb.OnWaveformChange(lb.WaveformIndex)
	}
}

// togglePause flips the paused state.
func (lb *LabBench) togglePause() {
	lb.Paused = !lb.Paused
	if lb.OnPauseToggle != nil {
		lb.OnPauseToggle(lb.Paused)
	}
}

// reset triggers a simulation reset.
func (lb *LabBench) reset() {
	if lb.OnReset != nil {
		lb.OnReset()
	}
}

// quit signals application exit.
func (lb *LabBench) quit() {
	lb.Running = false
	if lb.OnQuit != nil {
		lb.OnQuit()
	}
}

// StatusString returns a formatted string showing current parameters.
// Suitable for on-screen display or console output.
func (lb *LabBench) StatusString() string {
	status := "PAUSED"
	if !lb.Paused {
		status = "RUNNING"
	}

	return fmt.Sprintf(
		"[%s] E: %.2f MV/cm | T: %.0f K | f: %s | Wave: %s",
		status,
		lb.ElectricFieldAmplitude/1e8,
		lb.Temperature,
		formatFrequency(lb.Frequency),
		lb.WaveformNames[lb.WaveformIndex],
	)
}

// ControlsHelp returns a help string listing keyboard controls.
func (lb *LabBench) ControlsHelp() string {
	return `
╔══════════════════════════════════════════╗
║           LAB BENCH CONTROLS             ║
╠══════════════════════════════════════════╣
║  E/D  - Electric Field  (+/- 0.1 MV/cm)  ║
║  T/G  - Temperature     (+/- 25 K)       ║
║  F/V  - Frequency       (x2 / /2)        ║
║  W    - Cycle Waveform                   ║
║ SPACE - Pause/Resume                     ║
║  R    - Reset Simulation                 ║
║  Q    - Quit                             ║
╚══════════════════════════════════════════╝`
}

// formatFrequency converts Hz to human-readable format.
func formatFrequency(f float64) string {
	switch {
	case f >= 1e9:
		return fmt.Sprintf("%.1f GHz", f/1e9)
	case f >= 1e6:
		return fmt.Sprintf("%.1f MHz", f/1e6)
	case f >= 1e3:
		return fmt.Sprintf("%.1f kHz", f/1e3)
	default:
		return fmt.Sprintf("%.0f Hz", f)
	}
}

// TemperatureEffects computes how temperature affects Landau parameters.
// Based on the Curie-Weiss law: α(T) = α₀(T - Tc)/Tc
//
// For HZO, Tc ≈ 700K (orthorhombic phase stability)
// Below Tc: α < 0 → ferroelectric (double-well potential)
// Above Tc: α > 0 → paraelectric (single-well potential)
type TemperatureEffects struct {
	Tc         float64 // Curie temperature (K)
	Alpha0     float64 // Base α coefficient
	AlphaT     float64 // α at current temperature
	IsFerro    bool    // True if T < Tc
}

// ComputeTemperatureEffects calculates Landau parameters at temperature T.
func ComputeTemperatureEffects(T float64) *TemperatureEffects {
	const (
		Tc     = 700.0  // Curie temperature for HZO (K)
		alpha0 = 2.5e8  // Base coefficient (C⁻²·m²·J)
	)

	// Curie-Weiss law
	alphaT := alpha0 * (T - Tc) / Tc

	return &TemperatureEffects{
		Tc:      Tc,
		Alpha0:  alpha0,
		AlphaT:  alphaT,
		IsFerro: T < Tc,
	}
}

// SliderValue represents a normalized slider position [0, 1].
type SliderValue struct {
	Value    float64 // Current value
	Min      float64 // Minimum
	Max      float64 // Maximum
	Label    string  // Display label
	Unit     string  // Unit string
	Format   string  // Printf format for value
}

// Normalized returns the slider position as [0, 1].
func (sv *SliderValue) Normalized() float64 {
	if sv.Max == sv.Min {
		return 0
	}
	return (sv.Value - sv.Min) / (sv.Max - sv.Min)
}

// SetFromNormalized sets the value from a [0, 1] position.
func (sv *SliderValue) SetFromNormalized(norm float64) {
	norm = math.Max(0, math.Min(1, norm))
	sv.Value = sv.Min + norm*(sv.Max-sv.Min)
}

// DisplayString formats the slider value for display.
func (sv *SliderValue) DisplayString() string {
	return fmt.Sprintf(sv.Format, sv.Value) + " " + sv.Unit
}

// GetSliders returns SliderValue representations for GUI rendering.
func (lb *LabBench) GetSliders() []*SliderValue {
	return []*SliderValue{
		{
			Value:  lb.ElectricFieldAmplitude / 1e8, // Convert to MV/cm
			Min:    lb.EFieldMin / 1e8,
			Max:    lb.EFieldMax / 1e8,
			Label:  "Electric Field",
			Unit:   "MV/cm",
			Format: "%.2f",
		},
		{
			Value:  lb.Temperature,
			Min:    lb.TempMin,
			Max:    lb.TempMax,
			Label:  "Temperature",
			Unit:   "K",
			Format: "%.0f",
		},
		{
			Value:  math.Log10(lb.Frequency),
			Min:    math.Log10(lb.FreqMin),
			Max:    math.Log10(lb.FreqMax),
			Label:  "Frequency",
			Unit:   "",
			Format: "", // Use formatFrequency instead
		},
	}
}
