# FeCIM Themes Package

This package provides comprehensive theme support for the FeCIM Lattice Tools Fyne GUI applications.

## Features

- **Three theme options**: Dark (original FeCIM), Light, and High Contrast
- **Persistence**: User's theme preference is saved and restored
- **Theme Manager**: Centralized theme management with listener support
- **Ready-to-use widgets**: Theme switcher and quick toggle button

## Usage

### Basic Setup

```go
import (
    "fecim-lattice-tools/shared/themes"
    "fyne.io/fyne/v2/app"
)

func main() {
    fyneApp := app.NewWithID("com.example.myapp")
    
    // Create theme manager and load saved preference
    themeManager := themes.NewManager(fyneApp)
    themeManager.LoadPreference()
    
    // ... rest of your app
}
```

### Adding Theme Switcher to Settings

```go
// Full theme selector with preview
switcher := themes.NewThemeSwitcher(themeManager)
settingsPanel.Add(switcher)

// Or just a quick toggle button
toggleBtn := themes.CreateQuickToggle(themeManager)
toolbar.Add(toggleBtn)
```

### Listening for Theme Changes

```go
removeListener := themeManager.AddListener(func(t themes.ThemeType) {
    // React to theme change
    log.Println("Theme changed to:", t)
})

// Later, when done:
removeListener()
```

### Programmatic Theme Control

```go
// Set a specific theme
themeManager.SetTheme(themes.ThemeLight)

// Cycle through themes
themeManager.CycleTheme()

// Get current theme
current := themeManager.CurrentTheme()
```

## Available Themes

### Dark (FeCIM) - Default
The original FeCIM blue-based theme optimized for low-light environments.
- Background: Deep blue (#003264)
- Primary: Cyan (#00D4FF)
- Good contrast ratios for accessibility

### Light
A light theme for daytime use and bright environments.
- Background: Near-white (#FAFCFF)
- Primary: Deep blue (#0078B4)
- Comfortable for extended use in daylight

### High Contrast
Maximum contrast theme for accessibility and low-vision users.
- Background: Pure black (#000000)
- Primary: Pure cyan (#00FFFF)
- 10% larger text for better readability
- Maximum color contrast

## Theme Colors Reference

Each theme provides consistent semantic colors:

| Purpose | Dark | Light | High Contrast |
|---------|------|-------|---------------|
| Background | #003264 | #FAFCFF | #000000 |
| Primary | #00D4FF | #0078B4 | #00FFFF |
| Success | #57CC99 | #32A064 | #00FF00 |
| Warning | #FFC857 | #C88C28 | #FFFF00 |
| Error | #FF6B6B | #C83C3C | #FF0000 |

## Integration with Custom Widgets

For widgets that need theme-aware colors:

```go
colors := themeManager.GetCurrentColors()

// Use colors directly
bgColor := colors.Background.Main // Hex string like "#003264"

// Or use GetThemedColor for custom colors
accentColor := themes.GetThemedColor(
    themeManager.CurrentTheme(),
    darkVariant,
    lightVariant,
    hcVariant,
)
```

## Preference Storage

Theme preferences are stored using Fyne's preferences system with the key `app.theme`. This is automatically handled by the Manager.

## Testing

Run the tests:
```bash
go test ./shared/themes/...
```

## Migration from shared/theme

If you're migrating from the old `shared/theme` package:

1. Update imports from `fecim-lattice-tools/shared/theme` to `fecim-lattice-tools/shared/themes`
2. Replace `&sharedtheme.FeCIMTheme{}` with theme manager initialization
3. The old color constants are available as `themes.DarkPrimary`, `themes.DarkBackground`, etc.
