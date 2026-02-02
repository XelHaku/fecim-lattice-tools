// Package widgets provides shared UI components for FeCIM visualizers.
package widgets

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/validation"
)

// ToolStatusLabelMode controls how the tool status labels are rendered.
type ToolStatusLabelMode int

const (
	// ToolStatusSymbolOnly shows only the status symbol.
	ToolStatusSymbolOnly ToolStatusLabelMode = iota
	// ToolStatusSymbolWithName shows symbol plus tool name.
	ToolStatusSymbolWithName
)

// ToolValidationMessageStyle controls result text formatting.
type ToolValidationMessageStyle int

const (
	// ToolMessageUnicode uses ✓/✗/○ for PASS/FAIL/NOT INSTALLED.
	ToolMessageUnicode ToolValidationMessageStyle = iota
	// ToolMessageASCII uses V/X/O for PASS/FAIL/NOT INSTALLED.
	ToolMessageASCII
	// ToolMessageUnicodeSkip uses ○ SKIP (not installed) for missing tools.
	ToolMessageUnicodeSkip
)

// ToolValidationOptions configures the shared tool validation UI.
type ToolValidationOptions struct {
	Window               fyne.Window
	ButtonLabel          string
	DialogTitle          string
	StatusLabelMode      ToolStatusLabelMode
	MessageStyle         ToolValidationMessageStyle
	IncludeInstall       bool
	IncludeInstallNotes  bool
	DialogFooter         string
}

// ToolValidationWidgets bundles the status labels and button for tool validation.
type ToolValidationWidgets struct {
	CrossSimStatus   *widget.Label
	BadCrossbarStatus *widget.Label
	Button           *widget.Button
	Update           func()
}

// NewToolValidationWidgets creates status labels and a validation button.
func NewToolValidationWidgets(opts ToolValidationOptions) *ToolValidationWidgets {
	if opts.ButtonLabel == "" {
		opts.ButtonLabel = "Validate Tools"
	}
	if opts.DialogTitle == "" {
		opts.DialogTitle = "Tool Validation"
	}

	statusText := func(status validation.ToolStatus, name string) string {
		symbol := status.Symbol()
		if opts.StatusLabelMode == ToolStatusSymbolWithName {
			return fmt.Sprintf("%s %s", symbol, name)
		}
		return symbol
	}

	busyText := func(name string) string {
		if opts.StatusLabelMode == ToolStatusSymbolWithName {
			return "... " + name
		}
		return "..."
	}

	crosssimStatus := widget.NewLabel(statusText(validation.StatusUnknown, "CrossSim"))
	badcrossbarStatus := widget.NewLabel(statusText(validation.StatusUnknown, "BadCrossbar"))
	crosssimStatus.TextStyle = fyne.TextStyle{Monospace: true}
	badcrossbarStatus.TextStyle = fyne.TextStyle{Monospace: true}

	update := func() {
		crosssimInfo := validation.CrossSimInfo()
		badcrossbarInfo := validation.BadCrossbarInfo()
		SafeUpdateLabel(crosssimStatus, statusText(crosssimInfo.Status, "CrossSim"))
		SafeUpdateLabel(badcrossbarStatus, statusText(badcrossbarInfo.Status, "BadCrossbar"))
	}

	formatResult := func(result *validation.ValidationResult) string {
		notInstalled := result.Error != nil && strings.Contains(strings.ToLower(result.Error.Error()), "not installed")
		if result.Passed {
			switch opts.MessageStyle {
			case ToolMessageASCII:
				return "V PASS"
			default:
				return "✓ PASS"
			}
		}
		if notInstalled {
			switch opts.MessageStyle {
			case ToolMessageASCII:
				return "O NOT INSTALLED"
			case ToolMessageUnicodeSkip:
				return "○ SKIP (not installed)"
			default:
				return "○ NOT INSTALLED"
			}
		}
		switch opts.MessageStyle {
		case ToolMessageASCII:
			return "X FAIL"
		default:
			return "✗ FAIL"
		}
	}

	toolsBtn := widget.NewButton(opts.ButtonLabel, func() {
		go func() {
			SafeUpdateLabel(crosssimStatus, busyText("CrossSim"))
			SafeUpdateLabel(badcrossbarStatus, busyText("BadCrossbar"))

			var installResults []*validation.InstallResult
			if opts.IncludeInstall {
				installResults = validation.InstallToolsIfNeeded()
			}

			validateResults := validation.ValidateAllTools()

			messages := make([]string, 0, len(validateResults))
			for i, r := range validateResults {
				installMsg := ""
				if opts.IncludeInstall && opts.IncludeInstallNotes && i < len(installResults) &&
					installResults[i].Output != "Already installed" && installResults[i].Success {
					installMsg = " (installed)"
				}
				messages = append(messages, fmt.Sprintf("%s: %s%s", r.Tool, formatResult(r), installMsg))
			}

			SafeUpdateLabel(crosssimStatus, statusText(validation.CrossSimInfo().Status, "CrossSim"))
			SafeUpdateLabel(badcrossbarStatus, statusText(validation.BadCrossbarInfo().Status, "BadCrossbar"))

			if opts.Window != nil {
				dialogText := strings.Join(messages, "\n")
				if opts.DialogFooter != "" {
					dialogText = dialogText + "\n\n" + opts.DialogFooter
				}
				content := widget.NewLabel(dialogText)
				content.Wrapping = fyne.TextWrapWord
				dialog.ShowCustom(opts.DialogTitle, "Close", content, opts.Window)
			}
		}()
	})

	// Initial status check in background
	go update()

	return &ToolValidationWidgets{
		CrossSimStatus:   crosssimStatus,
		BadCrossbarStatus: badcrossbarStatus,
		Button:           toolsBtn,
		Update:           update,
	}
}
