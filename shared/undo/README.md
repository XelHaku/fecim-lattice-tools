# Undo/Redo System

This package provides a command pattern-based undo/redo system for FeCIM parameter changes.

## Features

- **Command Pattern**: All changes are wrapped in reversible Command objects
- **History Stack**: Maintains a configurable history of actions (default: 100)
- **Keyboard Shortcuts**: Ctrl+Z (undo), Ctrl+Y or Ctrl+Shift+Z (redo)
- **Toolbar Buttons**: Pre-built undo/redo buttons with automatic state management
- **Command Grouping**: Group multiple commands into a single undoable action
- **Global Registry**: Easy access from any module via the global manager

## Quick Start

### Using the Global Manager (Recommended)

The global manager is initialized in the main application. Use it from any module:

```go
import "fecim-lattice-tools/shared/undo"

// Execute a parameter change with undo support
oldValue := 0.5
newValue := 1.0
undo.Execute(undo.NewFloatCommand(
    "Temperature",
    oldValue,
    newValue,
    func(v float64) { myApp.setTemperature(v) },
))

// Undo/redo are handled automatically via keyboard shortcuts and toolbar buttons
```

### Available Command Types

```go
// Float values (temperature, frequency, etc.)
undo.NewFloatCommand(name, oldValue, newValue, setter)

// Integer values (levels, counts, etc.)
undo.NewIntCommand(name, oldValue, newValue, setter)

// String values (modes, material names, etc.)
undo.NewStringCommand(name, oldValue, newValue, setter)

// Boolean values (toggles, flags, etc.)
undo.NewBoolCommand(name, oldValue, newValue, setter)

// Slider values (with optional UI updater)
undo.NewSliderCommand(name, oldValue, newValue, setter, uiUpdater)

// Selection changes (dropdowns, radio buttons)
undo.NewSelectCommand(name, oldValue, newValue, setter)

// Custom execute/undo functions
undo.NewFuncCommand(description, doFn, undoFn)
```

### Grouping Commands

Group multiple commands into a single undoable action:

```go
undo.BeginGroup("Change material settings")
undo.Execute(undo.NewFloatCommand("Temperature", oldTemp, newTemp, setTemp))
undo.Execute(undo.NewIntCommand("Levels", oldLevels, newLevels, setLevels))
undo.Execute(undo.NewStringCommand("Material", oldMat, newMat, setMaterial))
undo.EndGroup()
// User can now undo all three changes with a single Ctrl+Z
```

### Integrating with Sliders

For sliders, use debouncing to avoid creating too many undo entries:

```go
var lastSliderValue float64
var sliderTimer *time.Timer

slider.OnChanged = func(v float64) {
    // Cancel previous timer
    if sliderTimer != nil {
        sliderTimer.Stop()
    }
    
    // Debounce: only record after 300ms of no changes
    sliderTimer = time.AfterFunc(300*time.Millisecond, func() {
        if v != lastSliderValue {
            undo.Execute(undo.NewSliderCommand(
                "Temperature",
                lastSliderValue,
                v,
                setTemperature,
                updateSliderUI,
            ))
            lastSliderValue = v
        }
    })
    
    // Apply immediately for responsiveness
    setTemperature(v)
}
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Main Application                         │
│  ┌─────────────────┐  ┌─────────────────────────────────┐  │
│  │  UndoManager    │  │  UndoToolbar (Buttons)           │  │
│  │  (global)       │◄─┤  Ctrl+Z / Ctrl+Y shortcuts       │  │
│  └────────┬────────┘  └─────────────────────────────────┘  │
│           │                                                  │
│           ▼                                                  │
│  ┌─────────────────────────────────────────────────────────┐│
│  │                 History Stacks                           ││
│  │  ┌───────────────────┐  ┌───────────────────┐          ││
│  │  │    Undo Stack     │  │    Redo Stack     │          ││
│  │  │  [cmd1, cmd2,...] │  │  [cmd3, cmd4,...] │          ││
│  │  └───────────────────┘  └───────────────────┘          ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘

Module Usage:
┌─────────────────────────────────────────────────────────────┐
│  Module (e.g., Hysteresis, Crossbar, MNIST)                 │
│                                                              │
│  slider.OnChanged = func(v float64) {                       │
│      undo.Execute(undo.NewFloatCommand(...))                │
│  }                                                           │
│                                                              │
│  dropdown.OnChanged = func(s string) {                      │
│      undo.Execute(undo.NewSelectCommand(...))               │
│  }                                                           │
└─────────────────────────────────────────────────────────────┘
```

## Files

- `command.go` - Command interface and Manager implementation
- `commands.go` - Built-in command types (Float, Int, String, Bool, etc.)
- `registry.go` - Global manager registry for cross-module access
- `command_test.go` - Unit tests

## Toolbar Widget

The toolbar is provided by `shared/widgets/undo_toolbar.go`:

```go
import "fecim-lattice-tools/shared/widgets"

// Create toolbar (connected to global manager)
undoToolbar := widgets.NewUndoToolbar(undoManager)

// Add to your layout
toolbar := container.NewHBox(
    homeBtn,
    undoToolbar,  // Contains undo and redo buttons
    otherButtons,
)
```

The toolbar automatically:
- Enables/disables buttons based on stack state
- Updates when commands are executed, undone, or redone
- Shows appropriate icons (ContentUndoIcon, ContentRedoIcon)
