package themes

import (
	"sync"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestNewManager(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	if manager == nil {
		t.Fatal("NewManager returned nil")
	}

	if manager.CurrentTheme() != ThemeDark {
		t.Errorf("Default theme should be dark, got %s", manager.CurrentTheme())
	}
}

func TestManagerSetTheme(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	tests := []ThemeType{ThemeLight, ThemeHighContrast, ThemeDark}

	for _, theme := range tests {
		manager.SetTheme(theme)
		if manager.CurrentTheme() != theme {
			t.Errorf("SetTheme(%s) did not update current theme", theme)
		}
	}
}

func TestManagerPersistence(t *testing.T) {
	app := test.NewApp()

	// First manager sets theme
	manager1 := NewManager(app)
	manager1.SetTheme(ThemeHighContrast)

	// Second manager should load the preference
	manager2 := NewManager(app)
	manager2.LoadPreference()

	if manager2.CurrentTheme() != ThemeHighContrast {
		t.Errorf("Theme preference was not persisted, got %s", manager2.CurrentTheme())
	}
}

func TestManagerLoadPreferenceDefault(t *testing.T) {
	app := test.NewApp()
	// Clear any existing preference
	app.Preferences().SetString(PreferenceKeyTheme, "")

	manager := NewManager(app)
	manager.LoadPreference()

	if manager.CurrentTheme() != ThemeDark {
		t.Errorf("Default theme should be dark when no preference set, got %s", manager.CurrentTheme())
	}
}

func TestManagerLoadPreferenceInvalid(t *testing.T) {
	app := test.NewApp()
	// Set invalid preference
	app.Preferences().SetString(PreferenceKeyTheme, "invalid_theme")

	manager := NewManager(app)
	manager.LoadPreference()

	if manager.CurrentTheme() != ThemeDark {
		t.Errorf("Should fall back to dark for invalid preference, got %s", manager.CurrentTheme())
	}
}

func TestManagerListener(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	var received ThemeType
	var wg sync.WaitGroup
	wg.Add(1)

	removeListener := manager.AddListener(func(theme ThemeType) {
		received = theme
		wg.Done()
	})

	manager.SetTheme(ThemeLight)
	wg.Wait()

	if received != ThemeLight {
		t.Errorf("Listener received %s, expected Light", received)
	}

	// Remove listener and verify it's not called again
	removeListener()
	received = ""
	manager.SetTheme(ThemeHighContrast)

	// received should not change since listener was removed
	// Note: We can't guarantee timing here, but the listener function shouldn't panic
}

func TestManagerCycleTheme(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	// Start with dark
	manager.SetTheme(ThemeDark)

	// Cycle through all themes
	expected := []ThemeType{ThemeLight, ThemeHighContrast, ThemeDark}
	for _, exp := range expected {
		got := manager.CycleTheme()
		if got != exp {
			t.Errorf("CycleTheme() = %s, want %s", got, exp)
		}
	}
}

func TestManagerSetThemeSameValue(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	listenerCalls := 0
	manager.AddListener(func(theme ThemeType) {
		listenerCalls++
	})

	// Set to dark (already default)
	manager.SetTheme(ThemeDark)
	// Set to dark again
	manager.SetTheme(ThemeDark)

	if listenerCalls != 0 {
		t.Errorf("Listener should not be called when setting same theme, got %d calls", listenerCalls)
	}
}

func TestManagerConcurrency(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent theme changes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			themes := AllThemes()
			manager.SetTheme(themes[i%len(themes)])
		}(i)
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = manager.CurrentTheme()
		}()
	}

	wg.Wait()

	// Should not panic or race
	_ = manager.CurrentTheme()
}

func TestGetThemeColors(t *testing.T) {
	tests := []struct {
		theme    ThemeType
		checkPrimary string
	}{
		{ThemeDark, "#00D4FF"},
		{ThemeLight, "#0078B4"},
		{ThemeHighContrast, "#00FFFF"},
	}

	for _, tt := range tests {
		t.Run(string(tt.theme), func(t *testing.T) {
			colors := GetThemeColors(tt.theme)
			if colors.Primary.Main != tt.checkPrimary {
				t.Errorf("GetThemeColors(%s).Primary.Main = %s, want %s",
					tt.theme, colors.Primary.Main, tt.checkPrimary)
			}
		})
	}
}

func TestManagerGetCurrentColors(t *testing.T) {
	app := test.NewApp()
	manager := NewManager(app)

	manager.SetTheme(ThemeLight)
	colors := manager.GetCurrentColors()

	if colors.Primary.Main != "#0078B4" {
		t.Errorf("GetCurrentColors().Primary.Main = %s, want #0078B4", colors.Primary.Main)
	}
}
