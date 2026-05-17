//go:build legacy_fyne

package gui

import (
	"testing"

	"fyne.io/fyne/v2"
)

func TestConfusionMatrixArrowNavigation(t *testing.T) {
	cm := NewConfusionMatrix()
	cm.FocusGained()
	if cm.selectedRow != 0 || cm.selectedCol != 0 {
		t.Fatalf("initial focused cell = (%d,%d), want (0,0)", cm.selectedRow, cm.selectedCol)
	}

	cm.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	cm.TypedKey(&fyne.KeyEvent{Name: fyne.KeyDown})
	if cm.selectedRow != 1 || cm.selectedCol != 1 {
		t.Fatalf("after arrow navigation = (%d,%d), want (1,1)", cm.selectedRow, cm.selectedCol)
	}

	cm.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEnd})
	if cm.selectedRow != 9 || cm.selectedCol != 9 {
		t.Fatalf("after End = (%d,%d), want (9,9)", cm.selectedRow, cm.selectedCol)
	}
}

func TestConfusionMatrixKeyboardActivationCallback(t *testing.T) {
	cm := NewConfusionMatrix()
	cm.selectedRow, cm.selectedCol = 2, 3
	cm.matrix[2][3] = 7

	called := false
	cm.OnCellTapped = func(actual, predicted, count int) {
		called = true
		if actual != 2 || predicted != 3 || count != 7 {
			t.Fatalf("unexpected callback payload: (%d,%d,%d)", actual, predicted, count)
		}
	}

	cm.TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if !called {
		t.Fatal("enter key should trigger selected cell callback")
	}
}

func TestOutputBarChartArrowNavigation(t *testing.T) {
	obc := NewOutputBarChart()
	obc.SetValues([]float64{0.1, 0.2, 0.3, 0.4, 0.5})

	obc.TypedKey(&fyne.KeyEvent{Name: fyne.KeyRight})
	if obc.selectedBar != 1 {
		t.Fatalf("selectedBar=%d, want 1", obc.selectedBar)
	}

	obc.TypedKey(&fyne.KeyEvent{Name: fyne.KeyEnd})
	if obc.selectedBar != 4 {
		t.Fatalf("selectedBar=%d, want 4 after End", obc.selectedBar)
	}

	obc.TypedKey(&fyne.KeyEvent{Name: fyne.KeyLeft})
	if obc.selectedBar != 3 {
		t.Fatalf("selectedBar=%d, want 3 after Left", obc.selectedBar)
	}
}
