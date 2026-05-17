//go:build legacy_fyne

package themes

import (
	"sync"

	"fyne.io/fyne/v2"

	"fecim-lattice-tools/shared/accessibility"
)

const (
	// PreferenceKeyTheme is the key used to store theme preference
	PreferenceKeyTheme = "app.theme"
)

// Manager handles theme switching and persistence
type Manager struct {
	app          fyne.App
	currentTheme ThemeType
	listeners    []func(ThemeType)
	mu           sync.RWMutex
}

// NewManager creates a new theme manager for the given Fyne app
func NewManager(app fyne.App) *Manager {
	return &Manager{
		app:          app,
		currentTheme: ThemeDark, // Default theme
		listeners:    make([]func(ThemeType), 0),
	}
}

// LoadPreference loads the saved theme preference and applies it.
// Call this during app initialization.
func (m *Manager) LoadPreference() {
	m.mu.Lock()
	defer m.mu.Unlock()

	prefs := m.app.Preferences()
	saved := prefs.StringWithFallback(PreferenceKeyTheme, string(ThemeDark))

	// Validate the saved theme
	themeType := ThemeType(saved)
	switch themeType {
	case ThemeDark, ThemeLight, ThemeHighContrast, ThemePresentation:
		m.currentTheme = themeType
	default:
		m.currentTheme = ThemeDark
	}

	// Apply the theme
	m.applyThemeLocked()
}

// CurrentTheme returns the currently active theme type
func (m *Manager) CurrentTheme() ThemeType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentTheme
}

// SetTheme changes the current theme and persists the preference
func (m *Manager) SetTheme(t ThemeType) {
	m.mu.Lock()
	if m.currentTheme == t {
		m.mu.Unlock()
		return
	}
	m.currentTheme = t

	// Apply the theme
	m.applyThemeLocked()

	// Persist the preference
	m.app.Preferences().SetString(PreferenceKeyTheme, string(t))

	// Copy listeners for notification outside lock
	listeners := make([]func(ThemeType), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	// Notify listeners (skip nil entries from removed listeners)
	for _, listener := range listeners {
		if listener != nil {
			listener(t)
		}
	}
}

// AddListener adds a callback that will be called when the theme changes.
// Returns a function to remove the listener.
func (m *Manager) AddListener(fn func(ThemeType)) func() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, fn)
	idx := len(m.listeners) - 1

	return func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		// Remove listener by setting to nil (avoids slice reordering issues)
		if idx < len(m.listeners) {
			m.listeners[idx] = nil
		}
	}
}

// CycleTheme cycles through available themes in order
func (m *Manager) CycleTheme() ThemeType {
	m.mu.RLock()
	current := m.currentTheme
	m.mu.RUnlock()

	themes := AllThemes()
	for i, t := range themes {
		if t == current {
			next := themes[(i+1)%len(themes)]
			m.SetTheme(next)
			return next
		}
	}

	// Fallback
	m.SetTheme(ThemeDark)
	return ThemeDark
}

// ApplyAccessibilityPreferences reapplies the currently selected theme with
// accessibility preferences (e.g., large text scaling).
func (m *Manager) ApplyAccessibilityPreferences() {
	m.mu.Lock()
	m.applyThemeLocked()
	m.mu.Unlock()
}

func (m *Manager) applyThemeLocked() {
	base := GetTheme(m.currentTheme)
	scale := accessibility.TextScale(m.app)
	if scale != 1.0 {
		m.app.Settings().SetTheme(NewScaledTheme(base, scale))
		return
	}
	m.app.Settings().SetTheme(base)
}

// App returns the Fyne app associated with this manager
func (m *Manager) App() fyne.App {
	return m.app
}

// GetCurrentColors returns the color palette for the current theme.
// This is useful for custom widgets that need theme-aware colors.
func (m *Manager) GetCurrentColors() ThemeColors {
	m.mu.RLock()
	t := m.currentTheme
	m.mu.RUnlock()

	return GetThemeColors(t)
}

// ThemeColors provides easy access to common theme colors
type ThemeColors struct {
	Primary    ColorSet
	Background ColorSet
	Surface    ColorSet
	Text       ColorSet
	Success    ColorSet
	Warning    ColorSet
	Error      ColorSet
}

// ColorSet contains a main color and related variants
type ColorSet struct {
	Main     string
	Light    string
	Dark     string
	Contrast string
}

// GetThemeColors returns the color palette for the given theme type
func GetThemeColors(t ThemeType) ThemeColors {
	switch t {
	case ThemeLight:
		return ThemeColors{
			Primary:    ColorSet{Main: "#0078B4", Light: "#3399CC", Dark: "#005580", Contrast: "#FFFFFF"},
			Background: ColorSet{Main: "#FAFCFF", Light: "#FFFFFF", Dark: "#F0F4FA", Contrast: "#142332"},
			Surface:    ColorSet{Main: "#F0F4FA", Light: "#FFFFFF", Dark: "#E6ECF5", Contrast: "#142332"},
			Text:       ColorSet{Main: "#142332", Light: "#506478", Dark: "#000000", Contrast: "#FFFFFF"},
			Success:    ColorSet{Main: "#32A064", Light: "#4FC080", Dark: "#288050", Contrast: "#FFFFFF"},
			Warning:    ColorSet{Main: "#C88C28", Light: "#E0A840", Dark: "#A07020", Contrast: "#000000"},
			Error:      ColorSet{Main: "#C83C3C", Light: "#E05050", Dark: "#A03030", Contrast: "#FFFFFF"},
		}
	case ThemeHighContrast:
		return ThemeColors{
			Primary:    ColorSet{Main: "#00FFFF", Light: "#80FFFF", Dark: "#00CCCC", Contrast: "#000000"},
			Background: ColorSet{Main: "#000000", Light: "#141420", Dark: "#000000", Contrast: "#FFFFFF"},
			Surface:    ColorSet{Main: "#14141E", Light: "#28283C", Dark: "#000000", Contrast: "#FFFFFF"},
			Text:       ColorSet{Main: "#FFFFFF", Light: "#FFFFFF", Dark: "#C8C8C8", Contrast: "#000000"},
			Success:    ColorSet{Main: "#00FF00", Light: "#80FF80", Dark: "#00CC00", Contrast: "#000000"},
			Warning:    ColorSet{Main: "#FFFF00", Light: "#FFFF80", Dark: "#CCCC00", Contrast: "#000000"},
			Error:      ColorSet{Main: "#FF0000", Light: "#FF8080", Dark: "#CC0000", Contrast: "#FFFFFF"},
		}
	case ThemePresentation:
		return ThemeColors{
			Primary:    ColorSet{Main: "#50C8FF", Light: "#80DDFF", Dark: "#30A0DD", Contrast: "#0A0A12"},
			Background: ColorSet{Main: "#0A0A12", Light: "#181824", Dark: "#000000", Contrast: "#F8F8FF"},
			Surface:    ColorSet{Main: "#181824", Light: "#282838", Dark: "#0A0A12", Contrast: "#F8F8FF"},
			Text:       ColorSet{Main: "#F8F8FF", Light: "#FFFFFF", Dark: "#B4B9C8", Contrast: "#0A0A12"},
			Success:    ColorSet{Main: "#64E6A0", Light: "#90F0C0", Dark: "#40C080", Contrast: "#0A0A12"},
			Warning:    ColorSet{Main: "#FFDC50", Light: "#FFE880", Dark: "#DDB830", Contrast: "#0A0A12"},
			Error:      ColorSet{Main: "#FF5A5A", Light: "#FF8080", Dark: "#DD4040", Contrast: "#F8F8FF"},
		}
	default: // ThemeDark
		return ThemeColors{
			Primary:    ColorSet{Main: "#00D4FF", Light: "#66E5FF", Dark: "#00A8CC", Contrast: "#003264"},
			Background: ColorSet{Main: "#003264", Light: "#004682", Dark: "#002850", Contrast: "#F0F4F8"},
			Surface:    ColorSet{Main: "#004682", Light: "#005AA0", Dark: "#003264", Contrast: "#F0F4F8"},
			Text:       ColorSet{Main: "#F0F4F8", Light: "#FFFFFF", Dark: "#A0B4C8", Contrast: "#003264"},
			Success:    ColorSet{Main: "#57CC99", Light: "#80DDB5", Dark: "#40A080", Contrast: "#003264"},
			Warning:    ColorSet{Main: "#FFC857", Light: "#FFD880", Dark: "#CCA040", Contrast: "#003264"},
			Error:      ColorSet{Main: "#FF6B6B", Light: "#FF9090", Dark: "#CC5050", Contrast: "#F0F4F8"},
		}
	}
}
