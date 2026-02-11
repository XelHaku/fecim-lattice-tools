# GUI Testing and Screenshot Crawling

This document describes the comprehensive GUI testing and screenshot crawling infrastructure for the FeCIM Lattice Tools application.

## Overview

The FeCIM Lattice Tools project includes several types of GUI tests and screenshot capture systems:

1. **Basic Visual Regression Tests** - Simple screenshot capture for modules
2. **Layout Audit Tests** - Comprehensive layout validation across window sizes
3. **UI Screenshot Crawler** - Advanced automated crawling of all UI states
4. **Xvfb Headless Tests** - Headless GUI testing for CI/CD environments

## UI Screenshot Crawler

The UI Screenshot Crawler is a comprehensive automated system that systematically captures screenshots of all possible UI states including:

- All tabs and nested tab structures
- Known dialogs (Tools, Export, Material picker, etc.)
- Overlays and popups
- Interactive element states
- Different window sizes (desktop, tablet, mobile)

### Features

- **Headless Compatible**: Works with xvfb for CI/CD environments
- **Deterministic**: Opt-in environment variable control
- **Comprehensive**: Discovers and captures all UI states automatically
- **Organized Storage**: Structured screenshot storage under `testdata/screenshots/audit/`
- **Non-Intrusive**: Avoids triggering calibration writes or simulation loops

### Environment Variables

- `FECIM_UI_CRAWL=1` - Enable UI crawler tests
- `DISPLAY` - Set for xvfb headless mode

### Directory Structure

Screenshots are saved to:
```
cmd/fecim-lattice-tools/testdata/screenshots/audit/
├── circuits/
│   ├── base_desktop_1200x800.png
│   ├── base_mobile_390x844.png
│   ├── tabs_desktop_1200x800_set0_tab0_circuit-analysis.png
│   ├── tabs_desktop_1200x800_set0_tab1_material-properties.png
│   ├── dialog_desktop_1200x800_tools.png
│   ├── overlay_desktop_1200x800_learn_0.png
│   └── xvfb/
│       ├── base_desktop_1200x800.png
│       └── tab_set0_idx0_circuit-analysis_1200x800.png
├── hysteresis/
├── mnist/
├── comparison/
├── eda/
└── crossbar/
```

## Running the Tests

### Basic Usage

```bash
# Run crawler for all modules
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler

# Run crawler for specific module (circuits)
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerCircuits
```

### Headless Mode (xvfb)

```bash
# Install xvfb (Ubuntu/Debian)
sudo apt-get install xvfb

# Run with xvfb
xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler

# Run detailed xvfb crawling
xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run XvfbCrawler
```

### CI/CD Integration

For GitHub Actions or similar CI systems:

```yaml
- name: Install xvfb
  run: sudo apt-get update && sudo apt-get install -y xvfb

- name: Run UI Screenshot Crawler
  run: |
    xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler
```

## Test Categories

### 1. Layout Audit Tests

Existing comprehensive layout tests that validate UI across different window sizes:

```bash
# Run layout audit
FECIM_LAYOUT_AUDIT=1 go test -v ./cmd/fecim-lattice-tools/... -run LayoutAudit

# With build tag
go test -tags layoutaudit -v ./cmd/fecim-lattice-tools/... -run LayoutAudit
```

### 2. Visual Regression Tests

Basic visual regression tests for individual modules:

```bash
# Run visual tests
go test -v ./cmd/fecim-lattice-tools/... -run Visual

# Run xvfb visual tests
FECIM_RUN_XVFB=1 xvfb-run -a go test -v ./cmd/fecim-lattice-tools/... -run VisualXvfb
```

### 3. UI Crawler Tests

Advanced automated crawling tests:

```bash
# All modules
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler

# Circuits module only
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerCircuits

# All modules with xvfb
xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run XvfbCrawler
```

## Module-Specific Considerations

### Module 4 (Circuits) - Primary Target

The circuits module is the primary target for comprehensive crawling because:
- Stable UI without font compatibility issues
- No infinite loops in test environment
- Rich interactive elements (tabs, dialogs, overlays)
- Representative of the application's complexity

### Known Limitations

Some modules have limitations in test environments:

- **Crossbar Module**: Uses Bold+Monospace fonts not available in test theme
- **MNIST Module**: Has infinite loop issues with `fyne.Do` in test driver
- **Other Modules**: Generally compatible but may have specific quirks

The crawler automatically skips problematic modules in comprehensive test runs.

## Implementation Details

### Core Components

1. **ui_screenshot_crawler_test.go** - Main crawler implementation
2. **ui_crawler_helpers.go** - Advanced UI discovery functions
3. **ui_crawler_xvfb_test.go** - Xvfb-specific implementations
4. **ui_layout_audit_test.go** - Existing layout audit (extended)

### Key Features

- **Tab Discovery**: Automatically finds and traverses all AppTabs
- **Dialog Detection**: Identifies buttons that trigger dialogs
- **Overlay Capture**: Captures popup overlays and modals
- **Safe Cleanup**: Safely closes dialogs and overlays after capture
- **Error Recovery**: Continues crawling even if individual captures fail

### Screenshot Naming Convention

Screenshots follow a structured naming pattern:
- `base_{size}_{width}x{height}.png` - Base UI state
- `tabs_{size}_{width}x{height}_set{N}_tab{M}_{name}.png` - Tab states
- `dialog_{size}_{width}x{height}_{type}.png` - Dialog states
- `overlay_{size}_{width}x{height}_{trigger}_{index}.png` - Overlay states

## Development Guidelines

### Adding New Modules

To add UI crawling for a new module:

1. Add module to the `modules` slice in `TestUICrawlerAllModules`
2. Create module factory function
3. Handle any module-specific quirks
4. Update this documentation

### Extending Crawling

To add new types of UI elements:

1. Extend `findInteractiveElements` in `ui_crawler_helpers.go`
2. Add new trigger patterns to `findPopupTriggers`
3. Implement capture logic in the main crawler
4. Test with multiple modules

### Debugging

For debugging crawler issues:

1. Enable verbose logging: `-v` flag
2. Check error collection in crawler state
3. Manually inspect generated screenshots
4. Use single module tests for isolation

## Future Enhancements

Potential improvements to the crawler system:

1. **Tooltip Capture** - Reliable cross-platform tooltip detection
2. **Animation Capture** - Video or animated GIF capture
3. **Accessibility Testing** - Screen reader and keyboard navigation validation
4. **Performance Monitoring** - UI rendering performance metrics
5. **Visual Diff Analysis** - Automated comparison against golden images

## Troubleshooting

### Common Issues

1. **Font Errors**: Some modules require specific fonts not available in test environments
2. **Display Issues**: Ensure DISPLAY is set for xvfb mode
3. **Timing Issues**: Increase delays if UI elements aren't fully rendered
4. **Memory Usage**: Large screenshot sets may consume significant memory

### Solutions

- Use xvfb mode for better compatibility
- Skip problematic modules temporarily  
- Adjust timing constants in crawler configuration
- Clean up screenshots periodically to manage disk space

---

## Quick Reference

```bash
# Essential commands
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerCircuits
xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler

# Output location
ls cmd/fecim-lattice-tools/testdata/screenshots/audit/

# Clean up
rm -rf cmd/fecim-lattice-tools/testdata/screenshots/audit/
```