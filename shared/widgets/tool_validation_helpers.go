//go:build legacy_fyne

// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fmt"
	"strings"

	"fecim-lattice-tools/shared/validation"
)

// ToolStatusText formats a tool status label.
func ToolStatusText(status validation.ToolStatus, name string, mode ToolStatusLabelMode) string {
	symbol := status.Symbol()
	if mode == ToolStatusSymbolWithName {
		return fmt.Sprintf("%s %s", symbol, name)
	}
	return symbol
}

// ToolBusyText formats the temporary "busy" status label.
func ToolBusyText(name string, mode ToolStatusLabelMode) string {
	if mode == ToolStatusSymbolWithName {
		return "... " + name
	}
	return "..."
}

// FormatToolValidationResult formats a validation result into a short status message.
func FormatToolValidationResult(result *validation.ValidationResult, style ToolValidationMessageStyle) string {
	if result == nil {
		switch style {
		case ToolMessageASCII:
			return "X FAIL"
		default:
			return "✗ FAIL"
		}
	}

	notInstalled := result.Error != nil && strings.Contains(strings.ToLower(result.Error.Error()), "not installed")
	if result.Passed {
		switch style {
		case ToolMessageASCII:
			return "V PASS"
		default:
			return "✓ PASS"
		}
	}
	if notInstalled {
		switch style {
		case ToolMessageASCII:
			return "O NOT INSTALLED"
		case ToolMessageUnicodeSkip:
			return "○ SKIP (not installed)"
		default:
			return "○ NOT INSTALLED"
		}
	}

	switch style {
	case ToolMessageASCII:
		return "X FAIL"
	default:
		return "✗ FAIL"
	}
}
