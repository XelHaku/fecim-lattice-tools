// Package undo provides a command pattern-based undo/redo system for FeCIM parameter changes.
package undo

import (
	"sync"
)

// Command represents a reversible action that can be undone/redone.
type Command interface {
	// Execute performs the command action.
	Execute()
	// Undo reverses the command action.
	Undo()
	// Description returns a human-readable description of the command.
	Description() string
}

// Manager manages the undo/redo history stack.
// It is thread-safe and can be used from multiple goroutines.
type Manager struct {
	mu           sync.Mutex
	undoStack    []Command
	redoStack    []Command
	maxHistory   int
	onChange     func() // Callback when stacks change
	grouping     bool   // Whether we're currently grouping commands
	groupedCmds  []Command
	groupDesc    string
}

// NewManager creates a new undo/redo manager with the specified maximum history size.
func NewManager(maxHistory int) *Manager {
	if maxHistory <= 0 {
		maxHistory = 100
	}
	return &Manager{
		undoStack:  make([]Command, 0, maxHistory),
		redoStack:  make([]Command, 0, maxHistory),
		maxHistory: maxHistory,
	}
}

// SetOnChange sets a callback that is invoked whenever the undo/redo stacks change.
// The callback is invoked with the lock held, so it should not call Manager methods.
func (m *Manager) SetOnChange(fn func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = fn
}

// Execute executes a command and adds it to the undo stack.
// This clears the redo stack since we're creating a new timeline.
func (m *Manager) Execute(cmd Command) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cmd.Execute()

	// If we're grouping commands, add to the group instead of the main stack
	if m.grouping {
		m.groupedCmds = append(m.groupedCmds, cmd)
		return
	}

	m.pushUndo(cmd)
	// Clear redo stack - new action creates new timeline
	m.redoStack = m.redoStack[:0]

	if m.onChange != nil {
		m.onChange()
	}
}

// pushUndo adds a command to the undo stack, respecting max history.
func (m *Manager) pushUndo(cmd Command) {
	if len(m.undoStack) >= m.maxHistory {
		// Remove oldest command
		copy(m.undoStack, m.undoStack[1:])
		m.undoStack = m.undoStack[:len(m.undoStack)-1]
	}
	m.undoStack = append(m.undoStack, cmd)
}

// pushRedo adds a command to the redo stack, respecting max history.
func (m *Manager) pushRedo(cmd Command) {
	if len(m.redoStack) >= m.maxHistory {
		// Remove oldest command
		copy(m.redoStack, m.redoStack[1:])
		m.redoStack = m.redoStack[:len(m.redoStack)-1]
	}
	m.redoStack = append(m.redoStack, cmd)
}

// Undo reverses the most recent command.
// Returns true if a command was undone, false if the undo stack is empty.
func (m *Manager) Undo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.undoStack) == 0 {
		return false
	}

	// Pop from undo stack
	cmd := m.undoStack[len(m.undoStack)-1]
	m.undoStack = m.undoStack[:len(m.undoStack)-1]

	// Undo the command
	cmd.Undo()

	// Push to redo stack
	m.pushRedo(cmd)

	if m.onChange != nil {
		m.onChange()
	}

	return true
}

// Redo re-applies the most recently undone command.
// Returns true if a command was redone, false if the redo stack is empty.
func (m *Manager) Redo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.redoStack) == 0 {
		return false
	}

	// Pop from redo stack
	cmd := m.redoStack[len(m.redoStack)-1]
	m.redoStack = m.redoStack[:len(m.redoStack)-1]

	// Re-execute the command
	cmd.Execute()

	// Push to undo stack
	m.pushUndo(cmd)

	if m.onChange != nil {
		m.onChange()
	}

	return true
}

// CanUndo returns true if there are commands to undo.
func (m *Manager) CanUndo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.undoStack) > 0
}

// CanRedo returns true if there are commands to redo.
func (m *Manager) CanRedo() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.redoStack) > 0
}

// UndoDescription returns the description of the next command to be undone.
// Returns empty string if undo stack is empty.
func (m *Manager) UndoDescription() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.undoStack) == 0 {
		return ""
	}
	return m.undoStack[len(m.undoStack)-1].Description()
}

// RedoDescription returns the description of the next command to be redone.
// Returns empty string if redo stack is empty.
func (m *Manager) RedoDescription() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.redoStack) == 0 {
		return ""
	}
	return m.redoStack[len(m.redoStack)-1].Description()
}

// UndoCount returns the number of commands in the undo stack.
func (m *Manager) UndoCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.undoStack)
}

// RedoCount returns the number of commands in the redo stack.
func (m *Manager) RedoCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.redoStack)
}

// Clear removes all commands from both stacks.
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.undoStack = m.undoStack[:0]
	m.redoStack = m.redoStack[:0]
	m.grouping = false
	m.groupedCmds = nil
	m.groupDesc = ""
	if m.onChange != nil {
		m.onChange()
	}
}

// BeginGroup starts grouping multiple commands into a single undoable action.
// All commands executed until EndGroup is called will be grouped together.
func (m *Manager) BeginGroup(description string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.grouping = true
	m.groupedCmds = make([]Command, 0)
	m.groupDesc = description
}

// EndGroup ends command grouping and adds the grouped commands as a single undoable action.
func (m *Manager) EndGroup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.grouping {
		return
	}

	m.grouping = false

	if len(m.groupedCmds) == 0 {
		m.groupedCmds = nil
		m.groupDesc = ""
		return
	}

	// Create a composite command from all grouped commands
	composite := &CompositeCommand{
		commands:    m.groupedCmds,
		description: m.groupDesc,
	}

	m.pushUndo(composite)
	m.redoStack = m.redoStack[:0]
	m.groupedCmds = nil
	m.groupDesc = ""

	if m.onChange != nil {
		m.onChange()
	}
}

// CompositeCommand groups multiple commands into a single undoable action.
type CompositeCommand struct {
	commands    []Command
	description string
}

// Execute runs all commands in order.
func (c *CompositeCommand) Execute() {
	for _, cmd := range c.commands {
		cmd.Execute()
	}
}

// Undo reverses all commands in reverse order.
func (c *CompositeCommand) Undo() {
	for i := len(c.commands) - 1; i >= 0; i-- {
		c.commands[i].Undo()
	}
}

// Description returns the composite description.
func (c *CompositeCommand) Description() string {
	return c.description
}
