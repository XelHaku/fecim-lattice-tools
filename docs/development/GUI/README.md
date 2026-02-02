# FeCIM Lattice Tools - GUI Documentation

Welcome to the GUI module documentation. This directory contains specifications, architecture guides, and troubleshooting information for all interactive visualizations in the FeCIM Lattice Tools project.

## Scope & Audience

- Internal development reference for engineers working on GUI modules
- Assumes familiarity with Fyne, module code structure, and Go conventions
- File paths are repo-relative unless a module root is explicitly called out

## Conventions Used in This Folder

- **file** paths are relative to the module root unless stated otherwise
- **state** lists local widget or app state
- **bindings** lists event handlers or state updates
- **data flow** lists triggers and their downstream updates

## Quick Navigation

### Module Documentation

| Module | Document | Purpose |
|--------|----------|---------|
| **Module 1: Hysteresis** | [GUI.module1.md](GUI.module1.md) | P-E curve visualization and Preisach model behavior |
| **Module 2: Crossbar Array** | [GUI.module2.md](GUI.module2.md) | Matrix-vector multiplication with non-idealities (IR drop, sneak paths, drift) |
| **Module 3: MNIST** | [GUI.module3.md](GUI.module3.md) | Neural network digit recognition training and inference |
| **Module 4: Circuits** | [GUI.module4.md](GUI.module4.md) | DAC/ADC/TIA peripheral circuit simulations |
| **Module 5: Comparison** | [GUI.module5.md](GUI.module5.md) | Technology comparison analysis (FeCIM vs DRAM vs Flash) |
| **Module 6: EDA Tools** | [GUI.module6.md](GUI.module6.md) | EDA workflow, placement, and layout visualization |
| **Module 7: Documentation** | [GUI.module7.md](GUI.module7.md) | In-app documentation viewer with search and navigation |

### Framework & Architecture

| Document | Purpose |
|----------|---------|
| [FYNE_NOTES.md](FYNE_NOTES.md) | Fyne framework patterns, best practices, and threading rules |

### Analysis & Reports

| Document | Purpose |
|----------|---------|
| [../HYPER_ANALYSIS_REPORT.md](../HYPER_ANALYSIS_REPORT.md) | Detailed UI/UX analysis of all module screenshots |

## What's in Each Module Document

Each module documentation file contains:

- **Module Metadata**: Entry point, package path, description
- **Bug Tracker**: Known issues with severity, component, and suggested fixes
- **Screen Layouts**: Detailed component tree with types, purposes, and file locations
- **State Management**: What data each component holds
- **Data Flow**: Event flows, triggers, and state updates
- **Bug Details**: Full technical analysis of each identified issue
- **Notes**: Physics accuracy, thread safety, and implementation patterns

## Bug Tracking Format

Bugs are tracked using a standardized format:

```text
BUG-MX-NNN
├─ M = Module number (1-7)
├─ X = Module identifier (always matches module number)
└─ NNN = Sequential bug number (001, 002, ...)

Example: BUG-M1-004 = Module 1, Bug #4
```

### Severity Levels

| Level | Description | Action Required |
|-------|-------------|-----------------|
| **Critical** | Crash, data loss, or security issue | Fix immediately before release |
| **High** | Functional bug affecting core features | Fix in current release |
| **Medium** | Non-critical bug affecting user experience | Schedule for next release |
| **Low** | Minor issue, edge case, or cosmetic | Document, fix when convenient |

### Bug Status

- `[x]` = Fixed or documented as safe
- `[ ]` = Open/pending

## Common Patterns & Best Practices

### Thread Safety (CRITICAL)

All GUI modules follow this strict pattern:

```go
// ✅ CORRECT: Goroutine updating UI
go func() {
    result := doWork()
    fyne.Do(func() {
        label.SetText(result)
    })
}()

// ❌ WRONG: Goroutine without fyne.Do
go func() {
    label.SetText("Updated")  // Race condition!
}()

// ❌ WRONG: Unnecessary fyne.Do in callbacks
button.OnTapped = func() {
    fyne.Do(func() {  // Not needed - already on main thread
        label.SetText("Clicked")
    })
}()
```

**Rule**: Use `fyne.Do()` ONLY when you create a goroutine with `go`. Never use it in Fyne callbacks or event handlers.

### Rate Limiting for Frequent Updates

Modules with high-frequency updates (60+ FPS) use rate-limited refresh:

```go
// Prevent excessive UI refreshes
type rateLimitedRefresh struct {
    lastRefresh time.Time
    interval    time.Duration
}

func (r *rateLimitedRefresh) ShouldRefresh() bool {
    now := time.Now()
    if now.Sub(r.lastRefresh) >= r.interval {
        r.lastRefresh = now
        return true
    }
    return false
}
```

### Responsive Layout Pattern

Most modules use `AdaptiveLayout` for desktop/mobile responsiveness:

```go
// Desktop: 3-column layout (22% | 75% | 25%)
// Mobile: Tab-based navigation (Info | Main | Controls)

desktop := container.NewHSplit(
    leftColumn,
    container.NewHSplit(
        mainContent,
        rightColumn,
    ),
)

mobile := container.NewAppTabs(
    container.NewTabItem("Info", leftColumn),
    container.NewTabItem("Main", mainContent),
    container.NewTabItem("Controls", rightColumn),
)
```

### Quantization Pattern

All FeCIM modules quantize analog values to 30 discrete levels:

```go
// Standard quantization function (used across all modules)
func QuantizeTo30Levels(analogValue float64) int {
    // Maps [0, maxValue] → [0, 29]
    discreteLevel := int(math.Round(analogValue / maxValue * 29))
    if discreteLevel > 29 {
        discreteLevel = 29
    }
    if discreteLevel < 0 {
        discreteLevel = 0
    }
    return discreteLevel
}
```

## Physics Constants Reference

Constants used consistently across all modules:

| Parameter | Value | Source |
|-----------|-------|--------|
| Discrete Levels | 30 (0-29) | Dr. external research group, COSM 2025 |
| Bits per Cell | 4.9 bits | log₂(30) ≈ 4.91 |
| Pr (Remanent Polarization) | 15-34 µC/cm² | Nature Commun. 2025 |
| Ps (Saturation Polarization) | ~30-35 µC/cm² | Nature Commun. 2025 |
| Ec (Coercive Field) | 1.0-1.5 MV/cm | Nature Commun. 2025 |
| Endurance | 10¹²+ cycles | PMC 2024, IEEE IRPS 2022 |
| Default Material | HZO (HfO₂-ZrO₂) | Dr. Tour research |

## Development Workflow

### Before Making GUI Changes

1. **Read the module document** - Understand the existing component hierarchy and data flow
2. **Check the bug tracker** - Know which issues are already identified
3. **Review FYNE_NOTES.md** - Refresh thread safety and layout patterns
4. **Run tests** - Ensure baseline functionality works

### When Adding a New Component

1. **Update the component tree** in the module document (Screen Layouts section)
2. **Document the state** it manages
3. **Document the data flow** - what events trigger updates?
4. **Note any threading concerns** - is it updated from goroutines?
5. **Add to bug section** if you discover or introduce issues

### When Fixing a Bug

1. **Update the bug status** to `[x]` (marked as fixed)
2. **Add a note** in "BugDetails" section with fix description
3. **Document why it was safe** or explain the fix

### Before Committing

```bash
# 1. Run all tests
go test ./...

# 2. Update any module documentation that changed
# Edit docs/development/GUI/GUI.moduleX.md

# 3. Run the module to verify visually
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools && ./fecim-lattice-tools

# 4. Commit with clear message
git commit -m "feat(module-X): description of change"
```

## Common Issues & Solutions

### Issue: UI doesn't update from goroutine
**Cause**: Missing `fyne.Do()` wrapper
**Solution**: Wrap all UI updates in goroutines with `fyne.Do(func() { ... })`
**Example**: Module 1 (BUG-M1-001)

### Issue: Layout oscillates or grows unexpectedly
**Cause**: Dynamic MinSize changes during updates
**Solution**: Set MinSize once at creation, never change it
**Example**: Module 2 (BUG-M2-003)

### Issue: High-frequency refresh causes jank
**Cause**: Refreshing on every update without rate limiting
**Solution**: Use rate-limited refresh for 60+ FPS updates
**Example**: All modules with animation

### Issue: Race condition on shared state
**Cause**: State accessed without mutex protection
**Solution**: Use `sync.RWMutex` for all shared state
**Example**: Module 1 (BUG-M1-003)

## Module Entry Points

To run or debug a specific module:

```bash
# Module 1: Hysteresis
go run ./module1-hysteresis/cmd/gui/main.go

# Module 2: Crossbar Array
go run ./module2-crossbar/cmd/gui/main.go

# Module 3: MNIST
go run ./module3-mnist/cmd/gui/main.go

# Module 4: Circuits
go run ./module4-circuits/cmd/gui/main.go

# Module 5: Comparison
go run ./module5-comparison/cmd/gui/main.go

# Module 6: EDA
go run ./module6-eda/cmd/gui/main.go

# Module 7: Documentation
# Note: No standalone entry - embedded only, accessed via toolbar icon

# Or the unified launcher
./launch.sh
```

## Widget Reference

### Custom Widgets (Shared)

| Widget | Package | Purpose |
|--------|---------|---------|
| `AdaptiveLayout` | `shared/widgets` | Responsive desktop/mobile layout |
| `PEPlot` | `module1-hysteresis/pkg/gui/widgets` | Hysteresis P-E curve rendering |
| `LevelIndicator` | `module1-hysteresis/pkg/gui/widgets` | 30-level gradient bar with click support |
| `CellVisualizer` | `module1-hysteresis/pkg/gui/widgets` | Visual representation of ferroelectric cell |
| `ModeIndicator` | `module1-hysteresis/pkg/gui/widgets` | WRITE/READ mode display |
| `Heatmap` | `module2-crossbar/pkg/gui` | Crossbar conductance heatmap |
| `CircuitDiagram` | `module4-circuits/pkg/gui` | DAC/ADC/TIA circuit rendering |
| `SearchDialog` | `module7-docs/pkg/gui` | Full-text search with fuzzy matching |
| `BreadcrumbWidget` | `module7-docs/pkg/gui` | Hierarchical path navigation |
| `TableOfContentsWidget` | `module7-docs/pkg/gui` | Auto-generated document ToC |
| `QuickAccessPanel` | `module7-docs/pkg/gui` | Recent & favorite documents |
| `GlossaryPillsWidget` | `module7-docs/pkg/gui` | Detected glossary term buttons |
| `RelatedDocsWidget` | `module7-docs/pkg/gui` | Related document suggestions |

### Standard Fyne Widgets Used

- `widget.Label` - Text display
- `widget.Button` - Clickable buttons
- `widget.Slider` - Numeric input (0-100 range)
- `widget.Select` - Dropdown selection
- `widget.Separator` - Visual dividers
- `widget.RichText` - Styled text (unused currently)
- `container.VBox` - Vertical stacking
- `container.HBox` - Horizontal stacking
- `container.Border` - Bordered layout with zones
- `container.NewVScroll()` - Vertical scrolling

## Resources & Links

### External References
- [Fyne Documentation](https://docs.fyne.io/)
- [Fyne Layout Guide](https://docs.fyne.io/explore/layouts.html)
- [Fyne Container API](https://docs.fyne.io/explore/container.html)

### Internal References
- [Physics Constants & Accuracy Policy](../scriptReference.md#accuracy--honesty-policy)
- [Testing Guide](../TESTING.md)
- [Development Setup](../scriptReference.md)

## Contribution Guidelines

### Documentation Updates

1. Keep module documents synchronized with code
2. Use the standard bug format (BUG-MX-NNN)
3. Include file paths (relative to module root) for all components
4. Update the Analysis Report when making significant UI changes

### Code Quality

1. All UI updates from goroutines must use `fyne.Do()`
2. All state shared between goroutines must use `sync.RWMutex`
3. High-frequency updates (>30 FPS) must use rate limiting
4. MinSize should be set once, never changed dynamically

### Testing

Run before any commit:
```bash
go test ./...
```

All GUI tests are located in `*_test.go` files in each module's `pkg/gui` directory.

## Support & Questions

For questions about GUI documentation or implementation:

1. Check the relevant module document first
2. Review FYNE_NOTES.md for framework patterns
3. Search existing bug issues
4. Check HYPER_ANALYSIS_REPORT.md for UI/UX insights

## Document Version

- **Last Updated**: 2026-02-02
- **Coverage**: Modules 1-7 (Hysteresis, Crossbar, MNIST, Circuits, Comparison, EDA, Documentation)
- **Status**: Active maintenance
