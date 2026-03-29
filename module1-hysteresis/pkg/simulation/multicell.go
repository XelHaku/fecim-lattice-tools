package simulation

import (
	"fmt"
	"sync"

	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
)

type CellCoord struct{ Row, Col int }

type CellState struct {
	Voltage       float64
	ElectricField float64
	Polarization  float64
	NormPol       float64
}

type MultiCellArray struct {
	rows, cols int
	material   *ferroelectric.HZOMaterial
	models     [][]*ferroelectric.PreisachModel
	states     [][]CellState
	mu         sync.RWMutex
}

func NewMultiCellArray(rows, cols int, material *ferroelectric.HZOMaterial) (*MultiCellArray, error) {
	if rows <= 0 || cols <= 0 {
		return nil, fmt.Errorf("invalid array dimensions: %dx%d", rows, cols)
	}
	if material == nil {
		return nil, fmt.Errorf("material cannot be nil")
	}
	m := &MultiCellArray{rows: rows, cols: cols, material: material, models: make([][]*ferroelectric.PreisachModel, rows), states: make([][]CellState, rows)}
	for r := 0; r < rows; r++ {
		m.models[r] = make([]*ferroelectric.PreisachModel, cols)
		m.states[r] = make([]CellState, cols)
		for c := 0; c < cols; c++ {
			m.models[r][c] = ferroelectric.NewPreisachModel(material)
		}
	}
	return m, nil
}

func (m *MultiCellArray) Size() (int, int) { m.mu.RLock(); defer m.mu.RUnlock(); return m.rows, m.cols }

func (m *MultiCellArray) StepCell(row, col int, voltage float64) (CellState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.inBounds(row, col) {
		return CellState{}, fmt.Errorf("cell index out of bounds: (%d,%d)", row, col)
	}
	return m.stepCellLocked(row, col, voltage), nil
}

func (m *MultiCellArray) StepWithVoltageMap(voltageMap [][]float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(voltageMap) != m.rows {
		return fmt.Errorf("voltage map rows mismatch: got %d, want %d", len(voltageMap), m.rows)
	}
	for r := range voltageMap {
		if len(voltageMap[r]) != m.cols {
			return fmt.Errorf("voltage map cols mismatch at row %d: got %d, want %d", r, len(voltageMap[r]), m.cols)
		}
	}
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.cols; c++ {
			m.stepCellLocked(r, c, voltageMap[r][c])
		}
	}
	return nil
}

func (m *MultiCellArray) StepWithSelector(cells []CellCoord, voltage float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, cell := range cells {
		if !m.inBounds(cell.Row, cell.Col) {
			return fmt.Errorf("cell index out of bounds: (%d,%d)", cell.Row, cell.Col)
		}
		m.stepCellLocked(cell.Row, cell.Col, voltage)
	}
	return nil
}

func (m *MultiCellArray) GetCellState(row, col int) (CellState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if !m.inBounds(row, col) {
		return CellState{}, fmt.Errorf("cell index out of bounds: (%d,%d)", row, col)
	}
	return m.states[row][col], nil
}

func (m *MultiCellArray) Snapshot() [][]CellState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([][]CellState, m.rows)
	for r := 0; r < m.rows; r++ {
		out[r] = make([]CellState, m.cols)
		copy(out[r], m.states[r])
	}
	return out
}

func (m *MultiCellArray) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for r := 0; r < m.rows; r++ {
		for c := 0; c < m.cols; c++ {
			m.models[r][c].Reset()
			m.states[r][c] = CellState{}
		}
	}
}

func (m *MultiCellArray) stepCellLocked(row, col int, voltage float64) CellState {
	var field float64
	if m.material.Thickness > 0 {
		field = voltage / m.material.Thickness
	}
	p := m.models[row][col].Update(field)
	s := CellState{Voltage: voltage, ElectricField: field, Polarization: p, NormPol: m.models[row][col].NormalizedPolarization()}
	m.states[row][col] = s
	return s
}
func (m *MultiCellArray) inBounds(row, col int) bool {
	return row >= 0 && row < m.rows && col >= 0 && col < m.cols
}
