package themes

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestScaledThemeDelegatesColorFontIcon(t *testing.T) {
	base := GetTheme(ThemeDark)
	scaled := NewScaledTheme(base, 1.25)

	if scaled.Color(theme.ColorNamePrimary, theme.VariantDark) == nil {
		t.Fatal("ScaledTheme.Color should delegate to base theme")
	}
	if scaled.Font(fyne.TextStyle{}) == nil {
		t.Fatal("ScaledTheme.Font should delegate to base theme")
	}
	if scaled.Icon(theme.IconNameHelp) == nil {
		t.Fatal("ScaledTheme.Icon should delegate to base theme")
	}
}

func TestManagerAppAndCycleFallback(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()
	m := NewManager(app)

	if m.App() != app {
		t.Fatal("App() should return bound fyne app")
	}

	// Force unknown current theme to exercise CycleTheme fallback branch.
	m.currentTheme = ThemeType("unknown")
	if got := m.CycleTheme(); got != ThemeDark {
		t.Fatalf("expected CycleTheme fallback to dark, got %s", got)
	}
}

func TestThemeColorSwitchCoverageAcrossVariants(t *testing.T) {
	themes := []fyne.Theme{&DarkTheme{}, &LightTheme{}, &HighContrastTheme{}, &PresentationTheme{}}
	colorNames := []fyne.ThemeColorName{
		theme.ColorNameBackground,
		theme.ColorNameForeground,
		theme.ColorNamePrimary,
		theme.ColorNameButton,
		theme.ColorNameHover,
		theme.ColorNamePressed,
		theme.ColorNameDisabledButton,
		theme.ColorNameInputBackground,
		theme.ColorNameInputBorder,
		theme.ColorNameDisabled,
		theme.ColorNamePlaceHolder,
		theme.ColorNameSuccess,
		theme.ColorNameError,
		theme.ColorNameWarning,
		theme.ColorNameSeparator,
		theme.ColorNameShadow,
		theme.ColorNameOverlayBackground,
		theme.ColorNameMenuBackground,
		theme.ColorNameHeaderBackground,
		theme.ColorNameSelection,
		theme.ColorNameFocus,
		theme.ColorNameHyperlink,
		theme.ColorNameForegroundOnPrimary,
		theme.ColorNameForegroundOnSuccess,
		theme.ColorNameForegroundOnWarning,
		theme.ColorNameForegroundOnError,
		theme.ColorNameScrollBar,
		fyne.ThemeColorName("unknown_color_case"), // default branch
	}

	for _, th := range themes {
		for _, n := range colorNames {
			if th.Color(n, theme.VariantDark) == nil {
				t.Fatalf("theme %T returned nil for color %s", th, n)
			}
		}
	}

	// High-contrast Size switch coverage for default branch as well.
	hc := &HighContrastTheme{}
	if hc.Size(theme.SizeNameText) <= 0 || hc.Size(theme.SizeNamePadding) <= 0 {
		t.Fatal("high contrast size variants should be positive")
	}
}
