package undo

import (
	"fmt"
)

// FloatCommand represents a command that changes a float64 value.
type FloatCommand struct {
	name     string
	oldValue float64
	newValue float64
	setter   func(float64)
}

// NewFloatCommand creates a command for changing a float value.
func NewFloatCommand(name string, oldValue, newValue float64, setter func(float64)) *FloatCommand {
	return &FloatCommand{
		name:     name,
		oldValue: oldValue,
		newValue: newValue,
		setter:   setter,
	}
}

// Execute sets the new value.
func (c *FloatCommand) Execute() {
	c.setter(c.newValue)
}

// Undo restores the old value.
func (c *FloatCommand) Undo() {
	c.setter(c.oldValue)
}

// Description returns a human-readable description.
func (c *FloatCommand) Description() string {
	return fmt.Sprintf("%s: %.3g → %.3g", c.name, c.oldValue, c.newValue)
}

// IntCommand represents a command that changes an int value.
type IntCommand struct {
	name     string
	oldValue int
	newValue int
	setter   func(int)
}

// NewIntCommand creates a command for changing an int value.
func NewIntCommand(name string, oldValue, newValue int, setter func(int)) *IntCommand {
	return &IntCommand{
		name:     name,
		oldValue: oldValue,
		newValue: newValue,
		setter:   setter,
	}
}

// Execute sets the new value.
func (c *IntCommand) Execute() {
	c.setter(c.newValue)
}

// Undo restores the old value.
func (c *IntCommand) Undo() {
	c.setter(c.oldValue)
}

// Description returns a human-readable description.
func (c *IntCommand) Description() string {
	return fmt.Sprintf("%s: %d → %d", c.name, c.oldValue, c.newValue)
}

// StringCommand represents a command that changes a string value.
type StringCommand struct {
	name     string
	oldValue string
	newValue string
	setter   func(string)
}

// NewStringCommand creates a command for changing a string value.
func NewStringCommand(name string, oldValue, newValue string, setter func(string)) *StringCommand {
	return &StringCommand{
		name:     name,
		oldValue: oldValue,
		newValue: newValue,
		setter:   setter,
	}
}

// Execute sets the new value.
func (c *StringCommand) Execute() {
	c.setter(c.newValue)
}

// Undo restores the old value.
func (c *StringCommand) Undo() {
	c.setter(c.oldValue)
}

// Description returns a human-readable description.
func (c *StringCommand) Description() string {
	return fmt.Sprintf("%s: %q → %q", c.name, c.oldValue, c.newValue)
}

// BoolCommand represents a command that changes a bool value.
type BoolCommand struct {
	name     string
	oldValue bool
	newValue bool
	setter   func(bool)
}

// NewBoolCommand creates a command for changing a bool value.
func NewBoolCommand(name string, oldValue, newValue bool, setter func(bool)) *BoolCommand {
	return &BoolCommand{
		name:     name,
		oldValue: oldValue,
		newValue: newValue,
		setter:   setter,
	}
}

// Execute sets the new value.
func (c *BoolCommand) Execute() {
	c.setter(c.newValue)
}

// Undo restores the old value.
func (c *BoolCommand) Undo() {
	c.setter(c.oldValue)
}

// Description returns a human-readable description.
func (c *BoolCommand) Description() string {
	if c.newValue {
		return fmt.Sprintf("%s: enabled", c.name)
	}
	return fmt.Sprintf("%s: disabled", c.name)
}

// FuncCommand wraps arbitrary execute/undo functions as a command.
type FuncCommand struct {
	description string
	doFn        func()
	undoFn      func()
}

// NewFuncCommand creates a command from execute and undo functions.
func NewFuncCommand(description string, doFn, undoFn func()) *FuncCommand {
	return &FuncCommand{
		description: description,
		doFn:        doFn,
		undoFn:      undoFn,
	}
}

// Execute runs the do function.
func (c *FuncCommand) Execute() {
	if c.doFn != nil {
		c.doFn()
	}
}

// Undo runs the undo function.
func (c *FuncCommand) Undo() {
	if c.undoFn != nil {
		c.undoFn()
	}
}

// Description returns the command description.
func (c *FuncCommand) Description() string {
	return c.description
}

// SliderCommand represents a slider change with old and new values.
// It includes debouncing logic for continuous slider movements.
type SliderCommand struct {
	name      string
	oldValue  float64
	newValue  float64
	setter    func(float64)
	uiUpdater func(float64) // Optional: update UI (slider, label) when undoing/redoing
}

// NewSliderCommand creates a command for a slider change.
func NewSliderCommand(name string, oldValue, newValue float64, setter func(float64), uiUpdater func(float64)) *SliderCommand {
	return &SliderCommand{
		name:      name,
		oldValue:  oldValue,
		newValue:  newValue,
		setter:    setter,
		uiUpdater: uiUpdater,
	}
}

// Execute sets the new value and updates UI.
func (c *SliderCommand) Execute() {
	c.setter(c.newValue)
	if c.uiUpdater != nil {
		c.uiUpdater(c.newValue)
	}
}

// Undo restores the old value and updates UI.
func (c *SliderCommand) Undo() {
	c.setter(c.oldValue)
	if c.uiUpdater != nil {
		c.uiUpdater(c.oldValue)
	}
}

// Description returns a human-readable description.
func (c *SliderCommand) Description() string {
	return fmt.Sprintf("%s: %.3g → %.3g", c.name, c.oldValue, c.newValue)
}

// MergeWith combines this command with a subsequent slider command if they target the same parameter.
// This is useful for debouncing rapid slider movements into a single undo action.
func (c *SliderCommand) MergeWith(other *SliderCommand) bool {
	if c.name != other.name {
		return false
	}
	// Keep the original oldValue but update newValue
	c.newValue = other.newValue
	return true
}

// SelectCommand represents a selection change (dropdown, radio button, etc.).
type SelectCommand struct {
	name     string
	oldValue string
	newValue string
	setter   func(string)
}

// NewSelectCommand creates a command for a selection change.
func NewSelectCommand(name string, oldValue, newValue string, setter func(string)) *SelectCommand {
	return &SelectCommand{
		name:     name,
		oldValue: oldValue,
		newValue: newValue,
		setter:   setter,
	}
}

// Execute sets the new selection.
func (c *SelectCommand) Execute() {
	c.setter(c.newValue)
}

// Undo restores the old selection.
func (c *SelectCommand) Undo() {
	c.setter(c.oldValue)
}

// Description returns a human-readable description.
func (c *SelectCommand) Description() string {
	return fmt.Sprintf("%s: %s → %s", c.name, c.oldValue, c.newValue)
}
