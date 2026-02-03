# Module 6 EDA GUI - Mermaid Diagram Reference

> **Status (2026-02-03):** This reference targets the earlier multi-panel GUI design.  
> The current GUI exposes two views only: **Builder & Validation** and **Learn**.  
> Treat the rest as legacy reference; see `module6-eda/pkg/gui/app.go` for the live structure.

## Overview

This document provides a quick reference to all Mermaid diagrams documenting the Module 6 EDA GUI structure, created from analysis of:

- `pkg/gui/app.go` - Main window creation (standalone mode)
- `pkg/gui/embedded.go` - Embedded app interface
- `pkg/gui/tabs/builder_validation_tab.go` - Builder & Validation implementation (1286 lines)
- `pkg/gui/tabs/learn_tab.go` - Learning center implementation (387 lines)

## Diagram Categories

### 1. Component Hierarchy & Structure

**Location**: `GUI_ARCHITECTURE.md` - GUI Component Hierarchy section

```
GUI Component Hierarchy
├── Application Entry
│   ├── Standalone Mode (CreateMainWindow)
│   └── Embedded Mode (EmbeddedEDAApp)
├── Main Layout Layer
│   ├── Header with ViewSelector
│   └── Stack Container (View Visibility)
└── Tab Navigation
    ├── Tab 1: Builder & Validation
    └── Tab 2: Learn
```

**Purpose**: Shows high-level application structure and two deployment paths

**Key Points**:
- Both modes use identical tab implementations
- Stack-based view switching in standalone mode
- AppTabs-based navigation in embedded mode
- ArrayConfig struct shared between all components

---

### 2. Builder & Validation Tab - Complete Structure

**Location**: `GUI_ARCHITECTURE.md` - Builder & Validation Tab - Detailed Structure section

```
Builder Tab (1286 lines)
├── TopSection (Configuration & Actions)
│   ├── ConfigSplit (45/55 HSplit)
│   │   ├── CellPanel (7 entry fields)
│   │   └── ArrayPanel (2 entries + 3 architecture buttons)
│   ├── ActionBar (3 buttons)
│   └── StatusBar
├── MainSplit (75/25 VSplit)
│   ├── PreviewTabs (75% height)
│   │   ├── VerilogTab (Code + Stats)
│   │   ├── DEFTab (Code + Stats)
│   │   └── LayoutTab (Images + Buttons)
│   └── ValidationSection (25% height)
│       ├── ValidationRow (Results)
│       ├── OpenLaneRow (Status)
│       └── LogSection (Output)
```

**Purpose**: Shows internal organization of main tab with all component relationships

**Key Points**:
- 7 cell configuration inputs (name, width, height, rise, fall, cap, leakage)
- 2 array dimension inputs (rows, cols)
- 3 image displays (KLayout, OpenROAD, Yosys)
- 4 validation result labels with status symbols
- Comprehensive logging with scrollable output

---

### 3. Learn Tab - Content Organization

**Location**: `GUI_ARCHITECTURE.md` - Learn Tab - Detailed Structure section

```
Learn Tab (387 lines)
├── Header
│   ├── Title: "Learning Center"
│   └── Subtitle
└── Split (25/75 HSplit)
    ├── Sidebar (25%)
    │   ├── SidebarTitle: "Topics"
    │   └── TopicList (3 items)
    │       ├── 1. What is FeCIM EDA?
    │       ├── 2. Crossbar Architecture
    │       └── 3. EDA Files We Generate
    └── ContentScroll (75%)
        └── DynamicContent
            ├── IntroContent (Intro, Visuals, Flow)
            ├── CrossbarContent (Passive, 1T1R, Comparison)
            └── FilesContent (Cards, Generation, Validation)
```

**Purpose**: Shows educational content organization and topic-based rendering

**Key Points**:
- Dynamic content generation per topic
- Canvas-based visualizations (isometric diagrams)
- Comprehensive educational material
- Adaptive grid layouts for file cards

---

### 4. Complete GUI Architecture Flow

**Location**: `GUI_DIAGRAM.md` - Complete GUI Architecture Flow section

**Type**: Large comprehensive flowchart (graph TD with subgraphs)

**Contains**:
- Full component tree from application start to file output
- All UI layers (Entry, Deployment, Main Layout, Tabs)
- Complete data flow through Export and Validation layers
- Threading model with goroutine spawning
- File system interactions
- OpenLane integration points

**Purpose**: "Big picture" reference showing all components and their connections

**Visual Features**:
- Color-coded subgraphs (blue=start, yellow=config, green=data, purple=tabs)
- Component grouping by functional area
- Clear flow paths from user action to output
- Threading boundaries marked

**Best Used For**:
- Understanding how components connect
- Tracing data flow from input to output
- Identifying integration points
- System architecture presentations

---

### 5. Data Flow: Generate All Operation

**Location**: `GUI_ARCHITECTURE.md` - Data Flow: Generate All Operation section

**Type**: Sequence diagram (sequenceDiagram)

**Flow**:
```
User clicks "Generate All"
    ↓
UI disables buttons, parses config
    ↓
GenerateLEF → Write cells/
GenerateLiberty → Write cells/
GenerateCellVerilog → Write cells/
    ↓
GenerateArrayVerilog → Update preview → Write data/
    ↓
generateBuilderDEF → Update preview → Write data/
    ↓
GenerateLayoutImage via KLayout
    ↓
GenerateOpenLaneConfig → Write data/
    ↓
Enable buttons, update status
```

**Purpose**: Shows step-by-step execution of file generation

**Best Used For**:
- Understanding the generation pipeline
- Debugging generation failures
- Verifying all files are created
- Performance profiling

---

### 6. Data Flow: Validation Operation

**Location**: `GUI_ARCHITECTURE.md` - Data Flow: Validation Operation section

**Type**: Sequence diagram (sequenceDiagram)

**Flow**:
```
User clicks "Validate All"
    ↓
[Yosys] ValidateVerilogWithCell → Update label
[DEF] ValidateDEF → Update label
[Cross] CrossCheckFiles → Update label
[Placement] RunPlacementCheckWithCell → Update label (if Docker)
    ↓
Calculate summary (all passed?)
    ↓
Update summary label
    ↓
Enable buttons
```

**Purpose**: Shows validation pipeline with multiple checks

**Validation Types**:
1. **Yosys Syntax Check** - Verilog parsing
2. **DEF Structure Check** - Placement file validation
3. **Cross-File Check** - Consistency across LEF/LIB/V
4. **OpenLane Placement** - DRC and placement validation (requires Docker)

**Best Used For**:
- Understanding what each validation checks
- Troubleshooting validation failures
- Deciding when to skip Docker-dependent checks

---

### 7. State Management & Callbacks

**Location**: `GUI_ARCHITECTURE.md` - State Management & Callbacks section

**Type**: Graph flowchart showing event propagation

**Event Handlers**:
```
Entry OnChanged
    → updateStats()
    → Calculate all metrics
    → fyne.Do() update labels

ArchButtons OnTapped
    → Update cfg.Architecture
    → Auto-set cell dimensions
    → updateArchButtons() visual feedback
    → updateLayoutImage()

ModeSelect OnChanged
    → updateModeHelp()
    → Display context text

GenerateBtn OnTapped
    → Spawn goroutine
    → fyne.Do() UI updates

TopicSelector OnSelected
    → Switch contentScroll.Content
    → Refresh display
```

**Purpose**: Shows event-driven UI update pattern

**Key Pattern**: All heavy lifting in goroutines, UI updates via `fyne.Do()`

---

### 8. Threading Model

**Location**: `GUI_ARCHITECTURE.md` - Threading Model section

**Type**: Flowchart showing thread separation

```
Main UI Thread (Fyne Event Loop)
    ↓
Button Click
    ↓
Spawn Background Goroutine
    ├── Heavy CPU/IO work
    └── Call fyne.Do() for UI updates
        ↓
        Returns to Main UI Thread
        ↓
        Executes UI updates safely
```

**Goroutine Types**:
1. **Generate All** - File I/O, export generation
2. **Validate All** - Tool execution (Yosys, OpenROAD)
3. **Gen Schematic** - Yosys schematic generation
4. **Gen Layout** - KLayout/OpenROAD visualization

**Purpose**: Shows how to avoid UI freezing during long operations

**Best Used For**:
- Understanding thread safety in Fyne
- Debugging UI responsiveness issues
- Adding new long-running operations

---

### 9. Configuration Flow

**Location**: `GUI_ARCHITECTURE.md` - Configuration Flow section

**Type**: Flowchart showing config initialization and propagation

```
Application Start
    ↓
Initialize ArrayConfig
    ├── Rows: 4
    ├── Cols: 4
    ├── Mode: storage
    ├── Architecture: passive
    └── CellWidth/Height: defaults
    ↓
Create Tab Contents (pass config pointer)
    ├── Builder Tab receives pointer
    └── Learn Tab receives nil
    ↓
Parse Entry Widgets
    ↓
Update ArrayConfig fields on change
    ↓
Ripple Effects (updateStats, images, etc)
```

**Purpose**: Shows single source of truth for configuration

**Key Point**: Both tabs see same ArrayConfig, changes instantly visible

---

### 10. Error Handling & User Feedback

**Location**: `GUI_ARCHITECTURE.md` - Error Handling & User Feedback section

**Type**: Decision tree flowchart

```
Operation Requested
    ↓
Try in Goroutine
    ↓
Success?
    ├─ Yes → Update success labels
    │         Log to output
    │         Enable buttons
    │
    └─ No → Catch error
            ├─ File System? → Log error, suggest fix
            ├─ Validation? → Update label ✗, log details
            ├─ Tool Missing? → Log, show help, suggest alternative
            └─ Docker? → Log, show "Pull Image" button
```

**Error Indicators**:
- ✓ Success (green)
- ✗ Failed (red)
- ⊝ Skipped (gray)
- ○ Pending (blue)

---

### 11. Component State Diagram

**Location**: `GUI_DIAGRAM.md` - Component State Diagram section

**Type**: State machine (stateDiagram-v2)

```
Ready
    ↓ (User input changes)
Configuring
    ↓ (Config stable)
Ready
    ↓ (Click "Generate All")
Generating
    ↓ (Files created)
Previewing
    ↓ (Click "Validate All")
Validating
    ↓ (Validation complete)
ValidationResults
    ↓ (If all passed)
Exporting
    ↓ (Export done)
ExportComplete
    ↓ (Ready for next cycle)
Ready
```

**State Properties**:
- **Ready**: Buttons enabled, awaiting action
- **Configuring**: Entry fields active, stats updating
- **Generating**: Heavy file operations in progress
- **Previewing**: Results displayed, ready for validation
- **Validating**: Multiple validation checks running
- **ValidationResults**: Results shown, decision point
- **Exporting**: Package assembly in progress

**Purpose**: High-level application lifecycle

---

### 12. User Interaction Flow Diagram

**Location**: `GUI_DIAGRAM.md` - User Interaction Flow Diagram section

**Type**: Flowchart showing user actions and system reactions

```
User Actions
├─ ChangeRows → UpdateStats → UpdateLabels
├─ ChangeCols → UpdateStats → UpdateLabels
├─ ChangeArch → UpdateArchButtons → UpdateStats → UpdateLabels
├─ ChangeMode → UpdateModeHelp → UpdateLabels
├─ GenClick → DisableUI → RunGeneration → UpdatePreviews → EnableUI
├─ ValClick → DisableUI → RunValidation → UpdateResults → EnableUI
└─ ExportClick → DisableUI → RunExport → UpdateLog → EnableUI
```

**Reaction Pattern**:
1. Disable buttons
2. Run operation in goroutine
3. Collect results via fyne.Do()
4. Update UI widgets
5. Re-enable buttons
6. Show optional dialog

**Purpose**: Shows complete interaction lifecycle

---

## Quick Reference Table

| Diagram | File | Type | Size | Best For |
|---------|------|------|------|----------|
| GUI Component Hierarchy | GUI_ARCHITECTURE.md | graph TD | Small | Understanding structure |
| Builder Tab Detail | GUI_ARCHITECTURE.md | graph TD | Large | Component relationships |
| Learn Tab Detail | GUI_ARCHITECTURE.md | graph TD | Medium | Content organization |
| Generate All Flow | GUI_ARCHITECTURE.md | sequenceDiagram | Medium | Generation pipeline |
| Validate All Flow | GUI_ARCHITECTURE.md | sequenceDiagram | Medium | Validation pipeline |
| State Management | GUI_ARCHITECTURE.md | graph TD | Medium | Event handling |
| Threading Model | GUI_ARCHITECTURE.md | graph LR | Small | Thread coordination |
| Configuration Flow | GUI_ARCHITECTURE.md | graph TD | Medium | Config propagation |
| Error Handling | GUI_ARCHITECTURE.md | graph TD | Medium | Error cases |
| Complete Architecture | GUI_DIAGRAM.md | graph TD | XLarge | Big picture |
| State Machine | GUI_DIAGRAM.md | stateDiagram-v2 | Medium | Lifecycle |
| User Interactions | GUI_DIAGRAM.md | graph LR | Medium | User actions |

## How to Use These Diagrams

### For Understanding the Code
1. Start with **GUI Component Hierarchy** - see overall structure
2. Read **Builder Tab Detail** - understand main components
3. Read **Learn Tab Detail** - understand educational content
4. Study **State Management** - understand how events propagate

### For Adding a Feature
1. Check **State Machine** - where does your feature fit?
2. Review **User Interactions** - what actions trigger it?
3. Study **Threading Model** - is it long-running?
4. Look at **Error Handling** - what can go wrong?

### For Debugging
1. Check **Generate All Flow** - where does it fail?
2. Check **Validate All Flow** - which validation fails?
3. Review **Error Handling** - what error is shown?
4. Check **Threading Model** - is UI freezing?

### For Performance Analysis
1. Review **Threading Model** - are goroutines used?
2. Check **State Management** - how often do updates occur?
3. Review **Complete Architecture** - identify heavy operations
4. Look at performance table in GUI_ARCHITECTURE.md

### For Documentation/Presentation
1. Use **Complete Architecture** for overview
2. Use individual tabs diagrams for deep dives
3. Use **State Machine** for lifecycle explanation
4. Use **Data Flow** diagrams for process understanding

## Diagram Rendering

All diagrams are written in **Mermaid** syntax and will render in:
- GitHub markdown (.md files)
- Notion
- Confluence
- Obsidian
- VS Code (with Mermaid extension)
- Any Mermaid-compatible viewer

Example rendering:
```
[Open GUI_ARCHITECTURE.md in GitHub]
  ↓
Scroll to section
  ↓
Mermaid automatically renders diagram
  ↓
View interactive graph
```

## File Locations

**Documentation Files**:
- `<local-path>` (25KB, 636 lines, 15 diagrams)
- `<local-path>` (21KB, 595 lines, 4 large diagrams)
- `<local-path>` (13KB, 402 lines, reference)
- `<local-path>` (this file)

**Source Code**:
- `<local-path>` - Standalone window
- `<local-path>` - Embedded interface
- `<local-path>` - Main tab (1286 lines)
- `<local-path>` - Learn tab (387 lines)

## Diagram Statistics

- **Total Diagrams**: 19 across 2 files
- **Total Lines of Documentation**: 1,633 lines
- **Total Documentation Size**: 59 KB
- **Mermaid Types Used**: graph, sequenceDiagram, stateDiagram-v2
- **Color Coding**: Yes (16+ colors for visual hierarchy)
- **Emojis for Clarity**: Yes (build, config, labels, buttons, status)

## Next Steps

1. **View the diagrams** - Open `GUI_ARCHITECTURE.md` or `GUI_DIAGRAM.md` in your browser/editor
2. **Study the index** - Read `GUI_INDEX.md` for context and common tasks
3. **Reference as needed** - Use this file to find the right diagram
4. **Share with team** - All files are markdown and GitHub-compatible
