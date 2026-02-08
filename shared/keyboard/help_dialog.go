package keyboard

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowHelpDialog displays a keyboard shortcuts help dialog.
func ShowHelpDialog(window fyne.Window, shortcuts []Shortcut, moduleSpecific string) {
	helpText := FormatHelpText(shortcuts)
	if moduleSpecific != "" {
		helpText += "\n" + moduleSpecific
	}

	helpLabel := widget.NewLabel(helpText)
	helpLabel.Wrapping = fyne.TextWrapWord

	helpContent := container.NewVScroll(helpLabel)
	helpContent.SetMinSize(fyne.NewSize(400, 500))

	helpDialog := dialog.NewCustom("Keyboard Shortcuts", "Close", helpContent, window)
	helpDialog.Show()
}

// ShowManagerHelpDialog shows a help dialog using shortcuts from a Manager.
func ShowManagerHelpDialog(manager *Manager, moduleSpecific string) {
	ShowHelpDialog(manager.window, manager.GetShortcuts(), moduleSpecific)
}

// QuickHelpText returns a compact one-liner of the most common shortcuts.
func QuickHelpText() string {
	return "Space: Pause | Ctrl+S: Save | Ctrl+E: Export | ?: Help"
}

// HelpTooltip returns a widget with hover tooltip showing keyboard shortcuts.
func HelpTooltip(shortcuts []Shortcut) *widget.Label {
	label := widget.NewLabel("⌨️ Keyboard shortcuts available (press ? for help)")
	label.TextStyle = fyne.TextStyle{Italic: true}
	return label
}
