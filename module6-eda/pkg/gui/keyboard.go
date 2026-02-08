// Package gui provides Fyne-based GUI components for EDA suite.
package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// SetupKeyboard initializes keyboard shortcuts for the EDA window.
func SetupKeyboard(w fyne.Window, viewSelector *widget.Select) {
	// Register Ctrl+S for save
	ctrlS := &desktop.CustomShortcut{
		KeyName:  fyne.KeyS,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlS, func(_ fyne.Shortcut) {
		showInfoDialog(w, "Save", "Save: Feature coming soon")
	})

	// Register Ctrl+E for export
	ctrlE := &desktop.CustomShortcut{
		KeyName:  fyne.KeyE,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlE, func(_ fyne.Shortcut) {
		showInfoDialog(w, "Export", "Export: Use the Export buttons in each tab")
	})

	// Register Ctrl+R for reset
	ctrlR := &desktop.CustomShortcut{
		KeyName:  fyne.KeyR,
		Modifier: fyne.KeyModifierControl,
	}
	w.Canvas().AddShortcut(ctrlR, func(_ fyne.Shortcut) {
		if viewSelector != nil && len(viewSelector.Options) > 0 {
			viewSelector.SetSelected(viewSelector.Options[0])
		}
	})

	// Handle non-modifier keys
	w.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeySpace:
			// Toggle between views
			if viewSelector != nil {
				cycleViewSelector(viewSelector)
			}

		case fyne.Key1:
			if viewSelector != nil && len(viewSelector.Options) > 0 {
				viewSelector.SetSelected(viewSelector.Options[0])
			}

		case fyne.Key2:
			if viewSelector != nil && len(viewSelector.Options) > 1 {
				viewSelector.SetSelected(viewSelector.Options[1])
			}

		case fyne.KeySlash:
			ShowKeyboardHelp(w)

		case fyne.KeyLeft:
			if viewSelector != nil {
				prevView(viewSelector)
			}

		case fyne.KeyRight:
			if viewSelector != nil {
				nextView(viewSelector)
			}
		}
	})
}

// cycleViewSelector cycles to the next view
func cycleViewSelector(selector *widget.Select) {
	if selector == nil || len(selector.Options) == 0 {
		return
	}
	currentIdx := 0
	for i, opt := range selector.Options {
		if opt == selector.Selected {
			currentIdx = i
			break
		}
	}
	nextIdx := (currentIdx + 1) % len(selector.Options)
	selector.SetSelected(selector.Options[nextIdx])
}

// nextView switches to the next view
func nextView(selector *widget.Select) {
	cycleViewSelector(selector)
}

// prevView switches to the previous view
func prevView(selector *widget.Select) {
	if selector == nil || len(selector.Options) == 0 {
		return
	}
	currentIdx := 0
	for i, opt := range selector.Options {
		if opt == selector.Selected {
			currentIdx = i
			break
		}
	}
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(selector.Options) - 1
	}
	selector.SetSelected(selector.Options[prevIdx])
}

// showInfoDialog shows a simple info dialog
func showInfoDialog(w fyne.Window, title, message string) {
	dialog.ShowInformation(title, message, w)
}

// ShowKeyboardHelp displays a dialog with all keyboard shortcuts
func ShowKeyboardHelp(w fyne.Window) {
	helpText := `Keyboard Shortcuts:

Navigation:
  ←/→       Switch views
  Space     Cycle through views
  1         Go to Builder & Validation
  2         Go to Learn

Data:
  Ctrl+S    Save (coming soon)
  Ctrl+E    Export (use tab buttons)
  Ctrl+R    Reset to first view

Help:
  / or ?    Show this help dialog

Tips:
• Use number keys (1-2) to quickly jump to views
• Each tab has its own export functionality
• Generated files are educational examples only`

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(360, 320))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, w)
	helpDialog.Show()
}
