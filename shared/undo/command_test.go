package undo

import (
	"testing"
)

func TestManagerBasic(t *testing.T) {
	m := NewManager(10)

	if m.CanUndo() {
		t.Error("Expected empty undo stack")
	}
	if m.CanRedo() {
		t.Error("Expected empty redo stack")
	}

	// Execute a command
	value := 0
	cmd := NewIntCommand("test", 0, 42, func(v int) { value = v })
	m.Execute(cmd)

	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
	if !m.CanUndo() {
		t.Error("Expected non-empty undo stack")
	}
	if m.CanRedo() {
		t.Error("Expected empty redo stack after execute")
	}

	// Undo
	if !m.Undo() {
		t.Error("Undo should return true")
	}
	if value != 0 {
		t.Errorf("Expected value 0 after undo, got %d", value)
	}
	if m.CanUndo() {
		t.Error("Expected empty undo stack after single undo")
	}
	if !m.CanRedo() {
		t.Error("Expected non-empty redo stack")
	}

	// Redo
	if !m.Redo() {
		t.Error("Redo should return true")
	}
	if value != 42 {
		t.Errorf("Expected value 42 after redo, got %d", value)
	}
}

func TestManagerMultipleCommands(t *testing.T) {
	m := NewManager(10)
	value := 0

	// Execute multiple commands
	for i := 1; i <= 5; i++ {
		oldVal := value
		newVal := i * 10
		cmd := NewIntCommand("test", oldVal, newVal, func(v int) { value = v })
		m.Execute(cmd)
	}

	if value != 50 {
		t.Errorf("Expected value 50, got %d", value)
	}
	if m.UndoCount() != 5 {
		t.Errorf("Expected 5 undo commands, got %d", m.UndoCount())
	}

	// Undo all
	for i := 0; i < 5; i++ {
		m.Undo()
	}
	if value != 0 {
		t.Errorf("Expected value 0 after undoing all, got %d", value)
	}

	// Redo all
	for i := 0; i < 5; i++ {
		m.Redo()
	}
	if value != 50 {
		t.Errorf("Expected value 50 after redoing all, got %d", value)
	}
}

func TestManagerMaxHistory(t *testing.T) {
	m := NewManager(3)
	value := 0

	// Execute more commands than max history
	for i := 1; i <= 5; i++ {
		cmd := NewIntCommand("test", value, i, func(v int) { value = v })
		m.Execute(cmd)
	}

	if m.UndoCount() != 3 {
		t.Errorf("Expected 3 undo commands (max history), got %d", m.UndoCount())
	}

	// Undo all 3 available
	for i := 0; i < 3; i++ {
		m.Undo()
	}
	// Should be at value 2 (commands 3, 4, 5 were available; undoing to before command 3)
	if value != 2 {
		t.Errorf("Expected value 2 after undoing 3 commands, got %d", value)
	}
}

func TestManagerClearRedoOnNewCommand(t *testing.T) {
	m := NewManager(10)
	value := 0

	// Execute two commands
	cmd1 := NewIntCommand("test", 0, 10, func(v int) { value = v })
	cmd2 := NewIntCommand("test", 10, 20, func(v int) { value = v })
	m.Execute(cmd1)
	m.Execute(cmd2)

	// Undo one
	m.Undo()
	if value != 10 {
		t.Errorf("Expected value 10, got %d", value)
	}
	if m.RedoCount() != 1 {
		t.Errorf("Expected 1 redo command, got %d", m.RedoCount())
	}

	// Execute a new command - should clear redo stack
	cmd3 := NewIntCommand("test", 10, 30, func(v int) { value = v })
	m.Execute(cmd3)

	if value != 30 {
		t.Errorf("Expected value 30, got %d", value)
	}
	if m.RedoCount() != 0 {
		t.Errorf("Expected 0 redo commands after new execute, got %d", m.RedoCount())
	}
}

func TestCompositeCommand(t *testing.T) {
	m := NewManager(10)
	val1, val2 := 0, 0

	m.BeginGroup("Set both values")
	cmd1 := NewIntCommand("val1", 0, 10, func(v int) { val1 = v })
	cmd2 := NewIntCommand("val2", 0, 20, func(v int) { val2 = v })
	m.Execute(cmd1)
	m.Execute(cmd2)
	m.EndGroup()

	if val1 != 10 || val2 != 20 {
		t.Errorf("Expected val1=10, val2=20, got val1=%d, val2=%d", val1, val2)
	}
	if m.UndoCount() != 1 {
		t.Errorf("Expected 1 composite undo command, got %d", m.UndoCount())
	}

	// Undo should restore both
	m.Undo()
	if val1 != 0 || val2 != 0 {
		t.Errorf("Expected val1=0, val2=0 after undo, got val1=%d, val2=%d", val1, val2)
	}

	// Redo should set both
	m.Redo()
	if val1 != 10 || val2 != 20 {
		t.Errorf("Expected val1=10, val2=20 after redo, got val1=%d, val2=%d", val1, val2)
	}
}

func TestFloatCommand(t *testing.T) {
	m := NewManager(10)
	value := 1.5

	cmd := NewFloatCommand("Temperature", 1.5, 2.5, func(v float64) { value = v })
	m.Execute(cmd)

	if value != 2.5 {
		t.Errorf("Expected 2.5, got %f", value)
	}

	desc := m.UndoDescription()
	expected := "Temperature: 1.5 → 2.5"
	if desc != expected {
		t.Errorf("Expected description %q, got %q", expected, desc)
	}

	m.Undo()
	if value != 1.5 {
		t.Errorf("Expected 1.5 after undo, got %f", value)
	}
}

func TestStringCommand(t *testing.T) {
	m := NewManager(10)
	value := "hello"

	cmd := NewStringCommand("Mode", "hello", "world", func(v string) { value = v })
	m.Execute(cmd)

	if value != "world" {
		t.Errorf("Expected 'world', got %q", value)
	}

	m.Undo()
	if value != "hello" {
		t.Errorf("Expected 'hello' after undo, got %q", value)
	}
}

func TestBoolCommand(t *testing.T) {
	m := NewManager(10)
	value := false

	cmd := NewBoolCommand("AutoMode", false, true, func(v bool) { value = v })
	m.Execute(cmd)

	if !value {
		t.Error("Expected true")
	}

	desc := m.UndoDescription()
	if desc != "AutoMode: enabled" {
		t.Errorf("Expected 'AutoMode: enabled', got %q", desc)
	}

	m.Undo()
	if value {
		t.Error("Expected false after undo")
	}
}

func TestFuncCommand(t *testing.T) {
	m := NewManager(10)
	executed := false
	undone := false

	cmd := NewFuncCommand("Custom action",
		func() { executed = true },
		func() { undone = true },
	)
	m.Execute(cmd)

	if !executed {
		t.Error("Execute function should have been called")
	}

	m.Undo()
	if !undone {
		t.Error("Undo function should have been called")
	}
}

func TestOnChangeCallback(t *testing.T) {
	m := NewManager(10)
	callCount := 0

	m.SetOnChange(func() {
		callCount++
	})

	// Execute should trigger onChange
	cmd := NewIntCommand("test", 0, 1, func(v int) {})
	m.Execute(cmd)
	if callCount != 1 {
		t.Errorf("Expected callCount 1, got %d", callCount)
	}

	// Undo should trigger onChange
	m.Undo()
	if callCount != 2 {
		t.Errorf("Expected callCount 2, got %d", callCount)
	}

	// Redo should trigger onChange
	m.Redo()
	if callCount != 3 {
		t.Errorf("Expected callCount 3, got %d", callCount)
	}

	// Clear should trigger onChange
	m.Clear()
	if callCount != 4 {
		t.Errorf("Expected callCount 4, got %d", callCount)
	}
}

func TestClear(t *testing.T) {
	m := NewManager(10)

	// Add some commands
	for i := 0; i < 5; i++ {
		cmd := NewIntCommand("test", i, i+1, func(v int) {})
		m.Execute(cmd)
	}

	m.Undo()
	m.Undo()

	if m.UndoCount() != 3 {
		t.Errorf("Expected 3 undo commands, got %d", m.UndoCount())
	}
	if m.RedoCount() != 2 {
		t.Errorf("Expected 2 redo commands, got %d", m.RedoCount())
	}

	m.Clear()

	if m.UndoCount() != 0 {
		t.Errorf("Expected 0 undo commands after clear, got %d", m.UndoCount())
	}
	if m.RedoCount() != 0 {
		t.Errorf("Expected 0 redo commands after clear, got %d", m.RedoCount())
	}
}

func TestSliderCommandMerge(t *testing.T) {
	cmd1 := NewSliderCommand("Temperature", 10.0, 15.0, func(v float64) {}, nil)
	cmd2 := NewSliderCommand("Temperature", 15.0, 20.0, func(v float64) {}, nil)
	cmd3 := NewSliderCommand("Pressure", 1.0, 2.0, func(v float64) {}, nil)

	// Same parameter - should merge
	if !cmd1.MergeWith(cmd2) {
		t.Error("Expected merge to succeed for same parameter")
	}
	if cmd1.oldValue != 10.0 {
		t.Errorf("Expected oldValue to stay 10.0, got %f", cmd1.oldValue)
	}
	if cmd1.newValue != 20.0 {
		t.Errorf("Expected newValue to be 20.0, got %f", cmd1.newValue)
	}

	// Different parameter - should not merge
	if cmd1.MergeWith(cmd3) {
		t.Error("Expected merge to fail for different parameters")
	}
}

func TestSelectCommand(t *testing.T) {
	m := NewManager(10)
	value := "Manual"

	cmd := NewSelectCommand("Waveform", "Manual", "Sine Wave", func(v string) { value = v })
	m.Execute(cmd)

	if value != "Sine Wave" {
		t.Errorf("Expected 'Sine Wave', got %q", value)
	}

	desc := m.UndoDescription()
	expected := "Waveform: Manual → Sine Wave"
	if desc != expected {
		t.Errorf("Expected description %q, got %q", expected, desc)
	}

	m.Undo()
	if value != "Manual" {
		t.Errorf("Expected 'Manual' after undo, got %q", value)
	}
}
