//go:build legacy_fyne

// Package gui provides logging helpers for module 4 UI actions and inputs.
package gui

import (
	"sync"

	"fecim-lattice-tools/shared/logging"
)

var circuitsLogOnce sync.Once
var circuitsLog *logging.Logger

func getCircuitsLog() *logging.Logger {
	circuitsLogOnce.Do(func() {
		circuitsLog = logging.NewLogger("circuits")
	})
	return circuitsLog
}

func logAction(format string, args ...interface{}) {
	if !logging.IsVerbose(logging.VerbosityDebug) {
		return
	}
	getCircuitsLog().Debug("ACTION: "+format, args...)
}

func logInput(format string, args ...interface{}) {
	if !logging.IsVerbose(logging.VerbosityDebug) {
		return
	}
	getCircuitsLog().Debug("INPUT: "+format, args...)
}

func opModeLabel(mode OpMode) string {
	switch mode {
	case OpModeRead:
		return "READ"
	case OpModeWrite:
		return "WRITE"
	case OpModeCompute:
		return "COMPUTE"
	default:
		return "IDLE"
	}
}

func dacModeLabel(mode DACMode) string {
	switch mode {
	case DACManual:
		return "MANUAL"
	case DACReadPreset:
		return "READ_PRESET"
	case DACWritePreset:
		return "WRITE_PRESET"
	case DACInputVector:
		return "INPUT_VECTOR"
	case DACRandom:
		return "RANDOM"
	default:
		return "UNKNOWN"
	}
}

func dacRangeLabel(mode DACRangeMode) string {
	switch mode {
	case DACRangeRead:
		return "READ"
	case DACRangeWrite:
		return "WRITE"
	default:
		return "UNKNOWN"
	}
}
