//go:build legacy_fyne

package themes

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"fecim-lattice-tools/shared/accessibility"
)

func TestManagerApplyAccessibilityPreferences_LargeTextScale(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	m := NewManager(app)
	m.SetTheme(ThemeDark)

	baseSize := app.Settings().Theme().Size("text")
	accessibility.SetLargeTextMode(app, true)
	m.ApplyAccessibilityPreferences()

	scaledSize := app.Settings().Theme().Size("text")
	if scaledSize <= baseSize {
		t.Fatalf("expected scaled text size > base size, got base=%f scaled=%f", baseSize, scaledSize)
	}
}
