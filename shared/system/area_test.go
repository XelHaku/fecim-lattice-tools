package system

import "testing"

func TestCellAreaUM2_AllNodes(t *testing.T) {
	nodes := []TechnologyNode{Node130nm, Node65nm, Node28nm, Node22nm, Node14nm}
	cells := []CellType{CellFeFET, CellRRAM, CellPCM, CellFeCAP}
	for _, node := range nodes {
		for _, cell := range cells {
			m := NewCrossbarAreaModel(64, 64, node, cell)
			a := m.CellAreaUM2()
			if a <= 0 {
				t.Errorf("CellAreaUM2(%s,%s) = %g, want > 0", node, cell, a)
			}
		}
	}
}

func TestArrayAreaUM2(t *testing.T) {
	m := NewCrossbarAreaModel(64, 64, Node65nm, CellFeFET)
	arr := m.ArrayAreaUM2()
	cell := m.CellAreaUM2()
	expected := float64(64*64) * cell
	if arr != expected {
		t.Errorf("ArrayAreaUM2() = %g, want %g", arr, expected)
	}
	if arr <= 0 {
		t.Errorf("ArrayAreaUM2() = %g, want > 0", arr)
	}
}

func TestPeripheralAreaUM2(t *testing.T) {
	m := NewCrossbarAreaModel(64, 64, Node65nm, CellFeFET)
	p := m.PeripheralAreaUM2(4)
	if p <= 0 {
		t.Errorf("PeripheralAreaUM2(4) = %g, want > 0", p)
	}
	// More bits → more area
	p8 := m.PeripheralAreaUM2(8)
	if p8 <= p {
		t.Errorf("PeripheralAreaUM2(8)=%g should be > PeripheralAreaUM2(4)=%g", p8, p)
	}
}

func TestTotalAreaUM2(t *testing.T) {
	m := NewCrossbarAreaModel(64, 64, Node65nm, CellFeFET)
	total := m.TotalAreaUM2(4)
	arr := m.ArrayAreaUM2()
	per := m.PeripheralAreaUM2(4)
	if total != arr+per {
		t.Errorf("TotalAreaUM2(4) = %g, want %g", total, arr+per)
	}
	if total <= 0 {
		t.Errorf("TotalAreaUM2(4) = %g, want > 0", total)
	}
}
