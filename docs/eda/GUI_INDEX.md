# Module 6 EDA GUI - Complete Documentation Index

> **Status (2026-02-03):** This index targets an earlier multi-panel GUI design.  
> The current GUI exposes two views only: **Builder & Validation** and **Learn**.  
> Use this as legacy reference; see `module6-eda/pkg/gui/app.go` for the live structure.

## Quick Navigation

| Document | Purpose | Best For |
|----------|---------|----------|
| **GUI_ARCHITECTURE.md** | Detailed component breakdown, data flows, threading | Understanding how components interact |
| **GUI_DIAGRAM.md** | Complete Mermaid diagram with all flows | Visual reference, presentation |
| **This File** | Quick navigation & summary | Getting oriented |

## What's Documented

### 1. Component Hierarchy & Structure
- **Standalone vs Embedded Modes**: Two deployment paths with shared implementation
- **Tab Navigation**: Builder & Validation tab vs Learn tab
- **Builder Tab Sections**: Configuration, Preview, Validation
- **Learn Tab Topics**: Introduction, Crossbar Architecture, EDA Files

**See**: `GUI_ARCHITECTURE.md` - GUI Component Hierarchy section

### 2. Detailed Component Breakdowns

#### Builder & Validation Tab
- **Configuration Section**: Cell and Array panels with live stats
- **Preview Tabs**: Verilog, DEF, and Layout visualization
- **Validation Section**: Results display and logging
- **Image Generation**: KLayout, OpenROAD, Yosys outputs

**See**: `GUI_ARCHITECTURE.md` - Builder & Validation Tab - Detailed Structure

#### Learn Tab
- **Topic Selector**: 3 educational topics
- **Dynamic Content**: Context-sensitive educational material
- **Visuals**: Isometric diagrams and comparison tables

**See**: `GUI_ARCHITECTURE.md` - Learn Tab - Detailed Structure

### 3. Data Flows & Operations

#### Generate All Operation
- Parses user input from entry fields
- Generates cell library files (LEF, LIB, Verilog)
- Generates array Verilog and DEF
- Creates layout images via KLayout
- Outputs OpenLane configuration

**See**: `GUI_ARCHITECTURE.md` - Data Flow: Generate All Operation

#### Validation Operation
- Yosys Verilog syntax validation
- DEF placement structure validation
- Cross-file consistency checking
- OpenROAD placement validation via Docker

**See**: `GUI_ARCHITECTURE.md` - Data Flow: Validation Operation

### 4. State Management

#### Configuration Flow
- Single `ArrayConfig` struct shared across tabs
- Real-time updates from entry widgets
- Architecture-dependent cell dimensions
- Statistics auto-calculated

**See**: `GUI_ARCHITECTURE.md` - Configuration Flow

#### Callback System
- `OnChanged` handlers for entry fields
- `OnTapped` handlers for architecture buttons
- `OnSelected` handlers for topic selection
- All updates marshalled through `fyne.Do()`

**See**: `GUI_ARCHITECTURE.md` - State Management & Callbacks

### 5. Threading Model

#### Main UI Thread
- Fyne event loop handles UI rendering
- Button clicks and user input processed

#### Background Goroutines
- Generate All: File I/O and export functions
- Validate All: Tool execution (Yosys, OpenROAD)
- Gen Schematic: Yosys schematic generation
- Gen Layout: KLayout/OpenROAD visualization

#### Thread Synchronization
- All UI updates use `fyne.Do()` for safe marshalling
- Buttons disabled during operations
- Status labels updated on completion

**See**: `GUI_ARCHITECTURE.md` - Threading Model

### 6. Error Handling

#### Operation Types
- File system errors (permission, space)
- Validation failures (syntax, constraints)
- Tool missing (Docker, Yosys, OpenROAD)
- Docker issues (image not pulled)

#### User Feedback
- Status labels with ✓/✗/⊝ indicators
- Detailed log output with timestamps
- Optional error dialogs for critical failures
- Help suggestions when tools unavailable

**See**: `GUI_ARCHITECTURE.md` - Error Handling & User Feedback

### 7. UI Widget Organization

#### Input Widgets (Entry Fields)
- 7 cell configuration inputs
- 2 array dimension inputs
- Mode selector dropdown
- 3 architecture toggle buttons

#### Output Widgets (Display)
- 6 statistics labels (auto-calculated)
- 2 code preview areas (Verilog, DEF)
- 3 image canvases (KLayout, OpenROAD, Yosys)
- 4 validation result labels
- 1 log output with scroll

**See**: `GUI_ARCHITECTURE.md` - UI Widget Organization

## Key Concepts

### Deployment Modes

**Standalone Mode**
```go
// cmd/eda-gui/main.go
app := fyneApp.NewApp()
w := gui.CreateMainWindow(app)
w.ShowAndRun()
```

**Embedded Mode**
```go
// cmd/fecim-lattice-tools/main.go
edaApp := gui.NewEmbeddedEDAApp()
content := edaApp.BuildContent(fyneApp, window)
// Returns AppTabs for unified visualizer
```

### Configuration Sharing

Both modes use identical `ArrayConfig` struct:
- Rows: 4 (default)
- Cols: 4 (default)
- Mode: "storage" (default)
- Architecture: "passive" (default)
- CellWidth/CellHeight: Architecture-dependent

Changes in one tab instantly visible in the other.

### File Generation Pipeline

```
User Input
    ↓
Parse Config
    ↓
Generate Cell Library (LEF, LIB, V)
    ↓
Generate Array Files (Verilog, DEF)
    ↓
Generate Layout Images (KLayout/OpenROAD)
    ↓
Generate Config (OpenLane JSON)
    ↓
Update Previews & Status
```

### Validation Pipeline

```
User Clicks "Validate All"
    ↓
Yosys Syntax Check
    ↓
DEF Structure Check
    ↓
Cross-File Consistency Check
    ↓
OpenLane Placement Check (if Docker available)
    ↓
Display Results & Summary
```

## File Locations

### Source Code
```
module6-eda/
├── cmd/
│   ├── eda-gui/main.go          # Standalone entry point
│   └── eda-cli/main.go          # CLI entry point
├── pkg/gui/
│   ├── app.go                   # CreateMainWindow
│   ├── embedded.go              # EmbeddedEDAApp
│   └── tabs/
│       ├── builder_validation_tab.go
│       ├── learn_tab.go
│       └── learn_visuals*.go
├── pkg/export/                  # Generation functions
├── pkg/validation/              # Validation functions
└── pkg/openlane/               # OpenLane integration
```

### Generated Output
```
data/
├── fecim_crossbar_NxM.v        # Array Verilog
├── fecim_crossbar_NxM.def      # Array DEF
├── fecim_crossbar_NxM.png      # KLayout image
├── fecim_crossbar_NxM_openroad.png
├── fecim_crossbar_NxM_schematic.png
└── config.json                  # OpenLane config

cells/
├── fecim_bitcell/               # Passive cell
│   ├── fecim_bitcell.lef
│   ├── fecim_bitcell.lib
│   └── fecim_bitcell.v
├── fecim_1t1r_bitcell/         # 1T1R cell
└── fecim_2t1r_bitcell/         # 2T1R cell
```

## Mermaid Diagrams Provided

### GUI_ARCHITECTURE.md
1. **GUI Component Hierarchy** - Top-level structure and deployment modes
2. **Builder & Validation Tab - Detailed Structure** - All components and their relationships
3. **Learn Tab - Detailed Structure** - Topic-based content organization
4. **Data Flow: Generate All Operation** - Sequence diagram of generation process
5. **Data Flow: Validation Operation** - Sequence diagram of validation process
6. **State Management & Callbacks** - Event handling and updates
7. **Threading Model** - Goroutine coordination
8. **Configuration Flow** - How configuration propagates through system
9. **Error Handling & User Feedback** - Error cases and responses

### GUI_DIAGRAM.md
1. **Complete GUI Architecture Flow** - Comprehensive component relationship diagram
2. **Component State Diagram** - Application states and transitions
3. **User Interaction Flow Diagram** - User actions and system reactions
4. **Component Relationship Matrix** - Dependencies and data flow table

## Common Tasks

### I want to understand...

**How the GUI is structured?**
- Read: `GUI_ARCHITECTURE.md` - GUI Component Hierarchy
- View: `GUI_DIAGRAM.md` - Complete GUI Architecture Flow

**How data flows from user input to file output?**
- Read: `GUI_ARCHITECTURE.md` - Data Flow sections
- View: `GUI_DIAGRAM.md` - User Interaction Flow Diagram

**What happens during validation?**
- Read: `GUI_ARCHITECTURE.md` - Data Flow: Validation Operation
- See code: `pkg/validation/` and `builder_validation_tab.go` lines 752-898

**How threads are coordinated?**
- Read: `GUI_ARCHITECTURE.md` - Threading Model
- See code: `builder_validation_tab.go` - All goroutine patterns use `fyne.Do()`

**How configuration is shared between tabs?**
- Read: `GUI_ARCHITECTURE.md` - Configuration Flow
- See code: `app.go` lines 19-28 and `embedded.go` lines 25-34

**How the Learn tab displays content?**
- Read: `GUI_ARCHITECTURE.md` - Learn Tab - Detailed Structure
- See code: `learn_tab.go` - `MakeLearnTab()` function

**How to add a new feature?**
- Read: `GUI_ARCHITECTURE.md` - Design Patterns section
- Study: Existing button callbacks in `builder_validation_tab.go`

### I want to modify...

**The configuration defaults:**
- File: `app.go` lines 20-28 or `embedded.go` lines 26-34
- Change: ArrayConfig struct initialization

**The UI layout:**
- File: `builder_validation_tab.go` lines 1022-1175
- Method: Modify container declarations and split offsets

**The validation logic:**
- File: `builder_validation_tab.go` lines 753-898
- Method: Add/modify validation blocks in goroutine

**The cell/array dimensions:**
- File: `builder_validation_tab.go` lines 260-294 (architecture buttons)
- Method: Update architecture-specific dimensions

**The Learn tab content:**
- Files: `learn_tab.go` and `learn_visuals*.go`
- Method: Modify content generator functions

## Performance Notes

| Operation | Duration | Thread | Blocking |
|-----------|----------|--------|----------|
| Update stats | <1ms | Main | No |
| Generate files | 50-100ms | Background | No |
| Generate images | 1-5s | Background | No |
| Validate Verilog | 1-3s | Background | No |
| Validate placement | 2-5s | Background | No |

All heavy operations run in goroutines with UI updates via `fyne.Do()`.

## Testing the GUI

### Manual Testing Checklist

**Configuration**
- [ ] Change rows, verify stats update
- [ ] Change columns, verify stats update
- [ ] Toggle architecture, verify dimensions change
- [ ] Select mode, verify help text changes

**Generation**
- [ ] Click "Generate All", verify files created
- [ ] Check `data/` directory for outputs
- [ ] View Verilog preview
- [ ] View DEF preview
- [ ] Verify stats labels update

**Validation**
- [ ] Click "Validate All", verify results appear
- [ ] Check log for validation details
- [ ] Verify summary label shows pass/fail

**Layout Images**
- [ ] Click "Gen Schematic (Yosys)", check for schematic
- [ ] Click "Gen Layout (OpenROAD)", check for layout
- [ ] Verify status labels update

**Learn Tab**
- [ ] Click topic 0, view intro content
- [ ] Click topic 1, view crossbar diagrams
- [ ] Click topic 2, view file format cards

**Docker Integration**
- [ ] Check Docker status on startup
- [ ] If Docker missing, verify "Pull Image" button appears
- [ ] Click "Pull Image" and monitor log

## Code Quality Standards

### Entry Point: app.go
- Creates main window with header and view selector
- Initializes ArrayConfig shared structure
- Sets up Stack container for view switching

### Tab Implementation: builder_validation_tab.go
- ~1200 lines implementing entire tab
- Organized in sections: config, preview, validation, export
- All goroutines use `fyne.Do()` for UI updates
- Button states managed with Enable/Disable

### Learning Tab: learn_tab.go
- ~400 lines implementing learning center
- Dynamic content rendering based on topic selection
- Canvas-based visuals for diagrams
- Consistent formatting with bullet lists and cards

### Embedded Mode: embedded.go
- ~60 lines providing embeddable interface
- Implements `BuildContent()`, `Start()`, `Stop()` pattern
- Returns AppTabs for unified visualizer

## Architecture Patterns Used

1. **Shared Configuration Pattern** - ArrayConfig pointer passed to all tabs
2. **Goroutine + fyne.Do Pattern** - Long operations in background, UI updates marshalled
3. **Preview + Validation Pattern** - Real-time preview with explicit validation
4. **Conditional UI Pattern** - Dynamic button visibility based on system state
5. **Tab Content Switching Pattern** - Dynamic content generation per topic

## Known Limitations

- Liberty timing values are placeholders (real values require SPICE characterization)
- KLayout image generation requires Docker with OpenLane image
- Placement validation requires OpenROAD (Docker or native installation)
- No real-time collaboration between users
- Configuration not persisted between sessions

## Future Enhancement Opportunities

1. Save/load configuration profiles
2. Batch generation for multiple array sizes
3. Compare multiple configurations side-by-side
4. Export configuration as PDF report
5. Live preview updates (more responsive)
6. Undo/redo for configuration changes
7. Tool version detection and compatibility checking
8. Advanced mode with custom cell definitions
