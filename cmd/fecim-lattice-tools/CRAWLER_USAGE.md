# UI Screenshot Crawler - Usage Examples

## Quick Start

The UI Screenshot Crawler is now implemented and ready to use. Here are practical examples:

### 1. Basic Infrastructure Verification (No GUI Required)

```bash
# Verify crawler infrastructure without GUI
go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerInfrastructure
```

This test verifies:
- Environment variable detection
- Directory creation
- Module instantiation
- Configuration validation
- Helper function correctness

### 2. Run UI Crawler for Circuits Module

```bash
# Enable crawler and run for circuits module
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerCircuits
```

This will capture:
- Base UI states at multiple resolutions (desktop, tablet, mobile)
- All tab configurations
- Known dialogs (Tools, Export, etc.)
- Popup overlays
- Interactive element states

### 3. Headless Mode with xvfb

```bash
# Install xvfb first (Ubuntu/Debian)
sudo apt-get install xvfb

# Run in headless mode
xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerCircuits

# Run detailed xvfb crawler
xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run XvfbCrawler
```

### 4. All Modules (Where Compatible)

```bash
# Run crawler for all supported modules
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerAllModules
```

Note: Some modules (Crossbar, MNIST) are automatically skipped due to test environment compatibility issues.

## Output Location

Screenshots are saved to:
```
cmd/fecim-lattice-tools/testdata/screenshots/audit/
├── circuits/
│   ├── base_desktop_1200x800.png
│   ├── base_mobile_390x844.png  
│   ├── base_tablet_768x1024.png
│   ├── tabs_desktop_1200x800_set0_tab0_circuit-analysis.png
│   ├── tabs_desktop_1200x800_set0_tab1_material-properties.png
│   ├── dialog_desktop_1200x800_tools.png
│   └── overlay_desktop_1200x800_learn_0.png
└── xvfb/
    └── (xvfb-specific screenshots)
```

## Integration with Existing Tests

The crawler extends existing test infrastructure:

- **Layout Audit Tests** (`FECIM_LAYOUT_AUDIT=1`) - Existing comprehensive layout validation
- **Visual Regression Tests** - Basic module screenshot capture 
- **UI Crawler Tests** (`FECIM_UI_CRAWL=1`) - New comprehensive state crawler

All systems can run independently or together.

## CI/CD Integration Example

```yaml
name: GUI Tests

on: [push, pull_request]

jobs:
  gui-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Go
        uses: actions/setup-go@v3
        with:
          go-version: 1.19
          
      - name: Install xvfb
        run: sudo apt-get update && sudo apt-get install -y xvfb
        
      - name: Run Infrastructure Tests
        run: go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerInfrastructure
        
      - name: Run UI Crawler
        run: xvfb-run -a env FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawler
        
      - name: Upload Screenshots
        uses: actions/upload-artifact@v3
        with:
          name: ui-screenshots
          path: cmd/fecim-lattice-tools/testdata/screenshots/audit/
```

## Troubleshooting

### Common Issues

1. **"DISPLAY is not set"** - Use xvfb: `xvfb-run -a env FECIM_UI_CRAWL=1 go test...`

2. **Font errors** - Some modules have font compatibility issues, these are automatically skipped

3. **Test hangs** - Without display or xvfb, GUI tests will hang. Use verification tests for basic checks.

### Debug Mode

To see detailed logging:
```bash
FECIM_UI_CRAWL=1 go test -v ./cmd/fecim-lattice-tools/... -run UICrawlerCircuits -args -test.verbose
```

## Current Implementation Status

✅ **Completed:**
- Core crawler infrastructure
- Circuits module (primary target)
- Comprehensive tab traversal
- Dialog and overlay detection
- Headless xvfb support
- Verification and dry-run tests
- Organized screenshot storage
- Environment variable control

🚧 **Planned Enhancements:**
- Integration of additional modules (hysteresis, comparison, EDA)
- Tooltip capture implementation
- Visual diff analysis against golden images
- Animation/video capture
- Performance metrics collection

The crawler is fully functional for the circuits module and provides a solid foundation for extending to other modules as needed.