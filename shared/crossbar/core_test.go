package crossbar

import "testing"

// mockCore is a minimal CrossbarCore implementation for interface contract testing.
type mockCore struct {
	rows, cols int
	matrix     [][]float64
}

func newMockCore(rows, cols int) *mockCore {
	m := make([][]float64, rows)
	for i := range m {
		m[i] = make([]float64, cols)
	}
	return &mockCore{rows: rows, cols: cols, matrix: m}
}

func (m *mockCore) MVM(input []float64) ([]float64, error) {
	return make([]float64, m.rows), nil
}
func (m *mockCore) VMM(input []float64) ([]float64, error) {
	return make([]float64, m.cols), nil
}
func (m *mockCore) ProgramWeightMatrix(weights [][]float64) error { return nil }
func (m *mockCore) GetConductanceMatrix() [][]float64             { return m.matrix }
func (m *mockCore) Rows() int                                     { return m.rows }
func (m *mockCore) Cols() int                                     { return m.cols }

// Compile-time assertion: mockCore must satisfy CrossbarCore.
var _ CrossbarCore = (*mockCore)(nil)

func TestCrossbarCoreInterface_MockImplementation(t *testing.T) {
	var core CrossbarCore = newMockCore(4, 4)

	out, err := core.MVM([]float64{0.5, 0.5, 0.5, 0.5})
	if err != nil {
		t.Fatalf("MVM: unexpected error: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("MVM output length: got %d, want 4", len(out))
	}

	out, err = core.VMM([]float64{0.5, 0.5, 0.5, 0.5})
	if err != nil {
		t.Fatalf("VMM: unexpected error: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("VMM output length: got %d, want 4", len(out))
	}

	weights := make([][]float64, 4)
	for i := range weights {
		weights[i] = make([]float64, 4)
		for j := range weights[i] {
			weights[i][j] = float64(i*4+j) / 16.0
		}
	}
	if err := core.ProgramWeightMatrix(weights); err != nil {
		t.Fatalf("ProgramWeightMatrix: unexpected error: %v", err)
	}

	g := core.GetConductanceMatrix()
	if len(g) != 4 {
		t.Fatalf("GetConductanceMatrix rows: got %d, want 4", len(g))
	}
	if len(g[0]) != 4 {
		t.Fatalf("GetConductanceMatrix cols: got %d, want 4", len(g[0]))
	}

	if core.Rows() != 4 {
		t.Fatalf("Rows: got %d, want 4", core.Rows())
	}
	if core.Cols() != 4 {
		t.Fatalf("Cols: got %d, want 4", core.Cols())
	}
}

func TestCrossbarCoreInterface_ArraySatisfiesInterface(t *testing.T) {
	// Verify *Array satisfies CrossbarCore at runtime (beyond the compile-time check).
	cfg := &Config{Rows: 4, Cols: 4, ADCBits: 4, DACBits: 4}
	arr, err := NewArray(cfg)
	if err != nil {
		t.Fatalf("NewArray: %v", err)
	}

	var core CrossbarCore = arr
	if core.Rows() != 4 || core.Cols() != 4 {
		t.Fatalf("Array as CrossbarCore: Rows=%d Cols=%d, want 4 4", core.Rows(), core.Cols())
	}

	out, err := core.MVM([]float64{0.5, 0.5, 0.5, 0.5})
	if err != nil {
		t.Fatalf("Array.MVM via interface: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("Array.MVM via interface output length: got %d, want 4", len(out))
	}
}

// mockDevice is a minimal DeviceModel implementation for interface contract testing.
type mockDevice struct {
	levels     int
	gMin, gMax float64
}

func (d *mockDevice) ApplyWriteError(targetG float64) float64 { return targetG }
func (d *mockDevice) ReadNoise(storedG float64) float64       { return storedG }
func (d *mockDevice) DriftError(storedG, _ float64) float64   { return storedG }
func (d *mockDevice) Levels() int                             { return d.levels }
func (d *mockDevice) ConductanceRange() (float64, float64)    { return d.gMin, d.gMax }

// Compile-time assertion: mockDevice must satisfy DeviceModel.
var _ DeviceModel = (*mockDevice)(nil)

func TestDeviceModelInterface_MockImplementation(t *testing.T) {
	var dev DeviceModel = &mockDevice{levels: 30, gMin: 0.01e-6, gMax: 1.0e-6}

	g := dev.ApplyWriteError(0.5e-6)
	if g <= 0 {
		t.Fatalf("ApplyWriteError: got %v, want > 0", g)
	}

	g = dev.ReadNoise(0.5e-6)
	if g <= 0 {
		t.Fatalf("ReadNoise: got %v, want > 0", g)
	}

	g = dev.DriftError(0.5e-6, 3600)
	if g <= 0 {
		t.Fatalf("DriftError: got %v, want > 0", g)
	}

	if dev.Levels() != 30 {
		t.Fatalf("Levels: got %d, want 30", dev.Levels())
	}

	gMin, gMax := dev.ConductanceRange()
	if gMin >= gMax {
		t.Fatalf("ConductanceRange: min >= max: %v >= %v", gMin, gMax)
	}
}
