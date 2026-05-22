package simulation

import (
	"sync"
	"testing"
	"time"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

// TestNewEngineNilMaterialCreatesInertEngine verifies nil material does not bind an implicit default or panic.
func TestNewEngineNilMaterialCreatesInertEngine(t *testing.T) {
	engine := NewEngine(nil)
	if engine == nil {
		t.Fatal("expected inert engine, got nil")
	}
	if engine.material != nil || engine.model != nil {
		t.Fatalf("expected nil material/model for inert engine, got material=%v model=%v", engine.material, engine.model)
	}
	if engine.state == nil {
		t.Fatal("expected inert engine to retain usable state")
	}
	if engine.amplitude != 0 {
		t.Fatalf("expected zero amplitude for inert engine, got %.3e V", engine.amplitude)
	}

	engine.Start()
	engine.SetWaveform(WaveformManual)
	engine.SetVoltage(1.0)
	engine.Step()

	state := engine.State()
	if state.Time != 0 {
		t.Fatalf("expected nil-material Step to skip time integration, got %.3e s", state.Time)
	}
	if state.ElectricField != 0 {
		t.Fatalf("expected nil-material Step to clamp electric field, got %.3e V/m", state.ElectricField)
	}
	if len(state.VoltageHistory) != 0 || len(state.PolHistory) != 0 {
		t.Fatalf("expected no history for nil-material engine, got voltage=%d polarization=%d", len(state.VoltageHistory), len(state.PolHistory))
	}

	E, P := engine.GetHysteresisData()
	if len(E) != 0 || len(P) != 0 {
		t.Fatalf("expected no hysteresis loop for nil-material engine, got E=%d P=%d", len(E), len(P))
	}

	engine.Reset()
	if got := engine.State().Time; got != 0 {
		t.Fatalf("expected reset inert engine time to remain 0, got %.3e s", got)
	}
}

// TestEngineStartStop verifies thread-safe start/stop operations
func TestEngineStartStop(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	if engine.IsRunning() {
		t.Error("Engine should not be running initially")
	}

	engine.Start()
	if !engine.IsRunning() {
		t.Error("Engine should be running after Start()")
	}

	engine.Stop()
	if engine.IsRunning() {
		t.Error("Engine should not be running after Stop()")
	}
}

// TestEnginePause verifies thread-safe pause operations
func TestEnginePause(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	engine.Start()

	if engine.IsPaused() {
		t.Error("Engine should not be paused after Start()")
	}

	engine.Pause()
	if !engine.IsPaused() {
		t.Error("Engine should be paused after Pause()")
	}

	engine.Pause()
	if engine.IsPaused() {
		t.Error("Engine should not be paused after second Pause()")
	}
}

// TestEngineConcurrentAccess verifies no data races under concurrent access
func TestEngineConcurrentAccess(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	var wg sync.WaitGroup
	const goroutines = 10
	const iterations = 100

	// Start the engine
	engine.Start()

	// Spawn multiple goroutines doing concurrent operations
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				engine.IsRunning()
				engine.IsPaused()
				engine.Step()
			}
		}()
	}

	// Also do pause/unpause in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < iterations; j++ {
			engine.Pause()
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()
	engine.Stop()

	// If we get here without race detector complaints, we're good
}

// TestEngineStep verifies simulation advances
func TestEngineStep(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	engine.Start()
	initialTime := engine.State().Time

	for i := 0; i < 10; i++ {
		engine.Step()
	}

	if engine.State().Time <= initialTime {
		t.Error("Simulation time should advance after steps")
	}
}

// TestEngineStepRejectsNonPhysicalThickness verifies invalid thickness prevents integration.
func TestEngineStepRejectsNonPhysicalThickness(t *testing.T) {
	tests := []struct {
		name      string
		thickness float64
	}{
		{name: "zero thickness", thickness: 0},
		{name: "negative thickness", thickness: -1e-9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			material := *ferroelectric.DefaultHZO()
			material.Thickness = tt.thickness
			engine := NewEngine(&material)
			engine.SetWaveform(WaveformManual)
			engine.SetVoltage(0.5)
			engine.Start()

			engine.Step()

			state := engine.State()
			if state.Time != 0 {
				t.Fatalf("expected no time integration for thickness %.3e m, got time %.3e s", tt.thickness, state.Time)
			}
			if len(state.VoltageHistory) != 0 || len(state.PolHistory) != 0 {
				t.Fatalf("expected no history for thickness %.3e m, got voltage=%d polarization=%d", tt.thickness, len(state.VoltageHistory), len(state.PolHistory))
			}
		})
	}
}

// TestEngineReset verifies reset clears state
func TestEngineReset(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	engine.Start()
	for i := 0; i < 100; i++ {
		engine.Step()
	}

	engine.Reset()

	if engine.State().Time != 0 {
		t.Errorf("Time should be 0 after reset, got %v", engine.State().Time)
	}
}

// =============================================================================
// WAVEFORM GENERATION TESTS
// =============================================================================

// TestSineWaveformGeneration verifies sine wave produces correct values.
func TestSineWaveformGeneration(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.SetWaveform(WaveformSine)
	engine.SetFrequency(1e6) // 1 MHz
	engine.SetAmplitude(1.0) // 1 V

	// Sine wave should oscillate between -1 and +1
	engine.Start()
	minV, maxV := 0.0, 0.0

	// Run through one full period
	period := 1.0 / 1e6 // 1 µs
	steps := int(period / engine.dt)

	for i := 0; i < steps*2; i++ {
		engine.Step()
		v := engine.State().Voltage
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	// Should reach near ±amplitude
	if maxV < 0.9 || minV > -0.9 {
		t.Errorf("Sine wave amplitude issue: min=%.4f, max=%.4f (expected ±1)", minV, maxV)
	}
}

// TestTriangleWaveformGeneration verifies triangle wave produces correct values.
func TestTriangleWaveformGeneration(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.SetWaveform(WaveformTriangle)
	engine.SetFrequency(1e6)
	engine.SetAmplitude(1.0)

	engine.Start()
	minV, maxV := 0.0, 0.0

	// Run through multiple periods
	for i := 0; i < 10000; i++ {
		engine.Step()
		v := engine.State().Voltage
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}

	// Should reach near ±amplitude
	if maxV < 0.9 || minV > -0.9 {
		t.Errorf("Triangle wave amplitude issue: min=%.4f, max=%.4f", minV, maxV)
	}
}

// TestManualWaveformMode verifies manual voltage control.
func TestManualWaveformMode(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.SetWaveform(WaveformManual)
	engine.SetVoltage(0.5)
	engine.Start()

	// Run steps
	for i := 0; i < 10; i++ {
		engine.Step()
	}

	// Voltage should remain at set value
	if engine.State().Voltage != 0.5 {
		t.Errorf("Manual voltage should be 0.5, got %f", engine.State().Voltage)
	}
}

// =============================================================================
// PHYSICS RESPONSE TESTS
// =============================================================================

// TestPolarizationRespondsToField verifies ferroelectric response.
func TestPolarizationRespondsToField(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.SetWaveform(WaveformManual)
	engine.Start()

	// Apply positive field beyond coercive
	engine.SetVoltage(material.CoerciveVoltage() * 2)
	for i := 0; i < 100; i++ {
		engine.Step()
	}
	posP := engine.State().NormPol

	// Reset and apply negative field
	engine.Reset()
	engine.Start()
	engine.SetVoltage(-material.CoerciveVoltage() * 2)
	for i := 0; i < 100; i++ {
		engine.Step()
	}
	negP := engine.State().NormPol

	// Positive field should give higher polarization than negative
	if posP <= negP {
		t.Errorf("Polarization mismatch: P(+E)=%.4f should be > P(-E)=%.4f", posP, negP)
	}

	t.Logf("Polarization response: P(+E)=%.4f, P(-E)=%.4f", posP, negP)
}

// TestNormalizedPolarizationBounds verifies P_norm in [-1, 1].
func TestNormalizedPolarizationBounds(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.SetWaveform(WaveformSine)
	engine.Start()

	for i := 0; i < 10000; i++ {
		engine.Step()
		p := engine.State().NormPol
		if p < -1.1 || p > 1.1 {
			t.Errorf("Normalized P=%.4f outside bounds at step %d", p, i)
		}
	}
}

// =============================================================================
// HISTORY RECORDING TESTS
// =============================================================================

// TestHistoryRecording verifies voltage/polarization history.
func TestHistoryRecording(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.Start()

	// Run steps
	for i := 0; i < 500; i++ {
		engine.Step()
	}

	state := engine.State()
	if len(state.VoltageHistory) == 0 {
		t.Error("Voltage history should not be empty")
	}
	if len(state.PolHistory) == 0 {
		t.Error("Polarization history should not be empty")
	}
	if len(state.VoltageHistory) != len(state.PolHistory) {
		t.Error("Voltage and polarization histories should have same length")
	}
}

// TestHistoryMaxLimit verifies history trimming.
func TestHistoryMaxLimit(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	maxHist := 100
	engine.state.MaxHistory = maxHist
	engine.Start()

	// Run more steps than max history
	for i := 0; i < 500; i++ {
		engine.Step()
	}

	if len(engine.State().VoltageHistory) > maxHist {
		t.Errorf("History exceeded max: %d > %d",
			len(engine.State().VoltageHistory), maxHist)
	}
}

// =============================================================================
// MATERIAL CONFIGURATION TESTS
// =============================================================================

// TestEngineWithDifferentMaterials verifies all material types work.
func TestEngineWithDifferentMaterials(t *testing.T) {
	materials := []*ferroelectric.HZOMaterial{
		ferroelectric.DefaultHZO(),
		ferroelectric.FeCIMMaterial(),
		ferroelectric.FeCIMMaterialTarget(),
	}

	for _, mat := range materials {
		t.Run(mat.Name, func(t *testing.T) {
			engine := NewEngine(mat)
			if engine == nil {
				t.Fatal("NewEngine returned nil")
			}

			engine.Start()
			for i := 0; i < 100; i++ {
				engine.Step()
			}

			if engine.State().Time == 0 {
				t.Error("Simulation should have advanced")
			}
		})
	}
}

// =============================================================================
// HYSTERESIS DATA TESTS
// =============================================================================

func TestGetHysteresisDataRejectsNonPhysicalThickness(t *testing.T) {
	tests := []struct {
		name      string
		thickness float64
	}{
		{name: "zero thickness", thickness: 0},
		{name: "negative thickness", thickness: -1e-9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			material := *ferroelectric.DefaultHZO()
			material.Thickness = tt.thickness
			engine := NewEngine(&material)

			E, P := engine.GetHysteresisData()
			if len(E) != 0 || len(P) != 0 {
				t.Fatalf("expected no hysteresis data for thickness %.3e m, got E=%d P=%d", tt.thickness, len(E), len(P))
			}
		})
	}
}

// TestGetHysteresisData verifies loop data generation.
func TestGetHysteresisData(t *testing.T) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	E, P := engine.GetHysteresisData()

	if len(E) == 0 || len(P) == 0 {
		t.Fatal("Hysteresis data is empty")
	}

	if len(E) != len(P) {
		t.Errorf("E and P length mismatch: %d vs %d", len(E), len(P))
	}

	// E should span both positive and negative
	hasPos, hasNeg := false, false
	for _, e := range E {
		if e > 0 {
			hasPos = true
		}
		if e < 0 {
			hasNeg = true
		}
	}

	if !hasPos || !hasNeg {
		t.Error("E field should span both positive and negative values")
	}

	t.Logf("Hysteresis loop: %d points", len(E))
}

// =============================================================================
// BENCHMARKS
// =============================================================================

func BenchmarkEngineStep(b *testing.B) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Step()
	}
}

func BenchmarkEngineStepWithLargeHistory(b *testing.B) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)
	engine.state.MaxHistory = 10000
	engine.Start()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Step()
	}
}

func BenchmarkGetHysteresisData(b *testing.B) {
	material := ferroelectric.DefaultHZO()
	engine := NewEngine(material)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.GetHysteresisData()
	}
}
