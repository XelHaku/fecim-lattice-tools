<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-13 | Updated: 2026-02-13 -->

# module6-eda/pkg/gui

## Purpose

Provides Fyne-based GUI for the EDA (Electronic Design Automation) module. Displays tabbed interface for design flow stages: circuit specification, physical design, verification, and results. Integrates with EDA tool chains and provides visual feedback on design progress, constraint violations, and optimization metrics.

## Key Files

| File | Description |
|------|-------------|
| `app.go` | Main application window and tab structure (2.5KB). |
| `embedded.go` | Implements embedded app interface: `BuildContent()`, `Start()`, `Stop()`. |
| `keyboard.go` | Keyboard controls for EDA operations (3.7KB). |
| `tabs/` | Subdirectory with tab implementations for each EDA stage. |
| `widgets/` | Custom widgets for EDA-specific visualizations. |

## For AI Agents

### Working In This Directory

**Embedded App Interface:**

Implements standard interface (see `embedded.go`):

```go
type EmbeddedApp interface {
    BuildContent() fyne.CanvasObject
    Start()
    Stop()
}
```

**Tab Structure (tabs/ subdirectory):**

Each EDA stage is a separate tab:
- Specification tab: Circuit parameters, constraints
- Design tab: Physical layout, device sizing
- Verification tab: Simulation results, constraint checking
- Results tab: Optimization outcomes, reports

**Keyboard Controls:**

Common operations mapped to keyboard shortcuts (see `keyboard.go`):
- Tab navigation
- Design parameter adjustment
- Simulation execution
- Export results

### Testing Requirements

```bash
# Run all EDA GUI tests
go test ./module6-eda/pkg/gui -v

# Run embedded interface test
go test ./module6-eda/pkg/gui -run TestEmbedded -v

# Run tab tests
go test ./module6-eda/pkg/gui -run TestTabs -v

# Run keyboard controls
go test ./module6-eda/pkg/gui -run TestKeyboard -v
```

### Common Patterns

- **Tab switching**: Each tab is independent with own state; coordinated via top-level app struct
- **Design parameter updates**: Modify simulation parameters, re-run verification
- **Results display**: Tables, plots, and reports for design metrics
- **Keyboard control**: Dispatch key events to active tab handler

## Dependencies

### Internal

- `module6-eda/pkg/core` - EDA tool integration (not in scope here)
- `shared/widgets` - Custom Fyne widgets
- `shared/theme` - Application theme

### External

- `fyne.io/fyne/v2` - GUI framework

<!-- MANUAL: Last edited 2026-02-13. EDA module GUI is lightweight tabbed interface. -->
