---
Module: module6-eda
Name: FeCIM Design Suite - EDA
Entry: module6-eda/cmd/eda-gui/main.go
Package: multilayer-ferroelectric-cim-visualizer/module6-eda/pkg/gui
Description: |
  Educational array builder that demonstrates how FeCIM crossbar arrays could
  integrate with open-source EDA tools (OpenLane/SKY130). Generates LEF/DEF/Verilog
  files with placeholder timing values - NOT production-ready. Real fabrication
  requires validated FeFET SPICE models and characterized timing.
---

Bugs:
  (None currently tracked)

Screens:
  - name: MainWindow
    file:app.go:15
    description: Standalone window with tab selector dropdown
    layout:
      - Header (VBox):
          file:app.go:86-96
          components:
            - ViewSelector (Select):
                type: widget.Select
                purpose: Switch between Builder and Learn tabs
                state: Selected index (0 = Builder, 1 = Learn)
                file:app.go:46
                bindings: OnChanged switches tab visibility
            - Banner (Label):
                type: widget.Label
                purpose: Displays "Generate fabrication-ready files for OpenLane/SKY130"
                file:app.go:83-84
            - Separator:
                type: widget.Separator
                file:app.go:95
      - ContentContainer (Stack):
          file:app.go:50
          description: Container with all tabs stacked, show/hide based on selection
          components:
            - BuilderContent (Canvas):
                type: fyne.CanvasObject
                purpose: Builder & Validation tab
                file:tabs/builder_validation_tab.go:27
                visibility: Controlled by viewSelector.OnChanged
                state: Shown when viewSelector == "1. Builder & Validation"
            - LearnContent (Canvas):
                type: fyne.CanvasObject
                purpose: Educational Learn tab
                file:tabs/learn_tab.go:26
                visibility: Controlled by viewSelector.OnChanged
                state: Shown when viewSelector == "2. Learn"

  - name: EmbeddedApp
    file:embedded.go:24
    description: Embedded version for unified visualizer (AppTabs instead of Stack)
    layout:
      - AppTabs (Container):
          file:embedded.go:37-40
          components:
            - Tab 1 Builder:
                type: container.TabItem
                purpose: Builder & Validation
                content: tabs.MakeBuilderValidationTab()
            - Tab 2 Learn:
                type: container.TabItem
                purpose: Educational content
                content: tabs.MakeLearnTab()

  - name: BuilderValidationTab
    file:tabs/builder_validation_tab.go:27
    description: |
      Unified tab combining cell/array configuration, Verilog/DEF preview,
      validation (Yosys, DEF syntax, LEF/LIB/V cross-check), and export
    layout:
      - TopSection (VBox):
          file:tabs/builder_validation_tab.go:718-723
          components:
            - ConfigSplit (HSplit):
                file:tabs/builder_validation_tab.go:626-627
                offset: 0.5
                left: CellPanel
                right: ArrayPanel
                components:
                  - CellPanel (VBox):
                      file:tabs/builder_validation_tab.go:590-594
                      components:
                        - CellTitle (Label):
                            type: widget.Label
                            text: "Cell Config"
                            style: Bold
                            file:tabs/builder_validation_tab.go:591
                        - CellForm (Form):
                            type: widget.Form
                            file:tabs/builder_validation_tab.go:580-588
                            fields:
                              - nameEntry:
                                  purpose: Cell name
                                  default: "fecim_bitcell"
                                  file:tabs/builder_validation_tab.go:29-30
                              - widthEntry:
                                  purpose: Cell width (μm)
                                  default: "0.460"
                                  file:tabs/builder_validation_tab.go:32-33
                                  bindings: OnChanged -> updateStats()
                              - heightEntry:
                                  purpose: Cell height (μm)
                                  default: "2.720"
                                  file:tabs/builder_validation_tab.go:35-36
                                  bindings: OnChanged -> updateStats()
                              - riseEntry:
                                  purpose: Rise time (ns) - PLACEHOLDER
                                  default: "0.1"
                                  file:tabs/builder_validation_tab.go:39-40
                              - fallEntry:
                                  purpose: Fall time (ns) - PLACEHOLDER
                                  default: "0.1"
                                  file:tabs/builder_validation_tab.go:42-43
                              - capEntry:
                                  purpose: Input capacitance (pF) - PLACEHOLDER
                                  default: "0.002"
                                  file:tabs/builder_validation_tab.go:45-46
                              - leakageEntry:
                                  purpose: Leakage power (nW) - PLACEHOLDER
                                  default: "0.001"
                                  file:tabs/builder_validation_tab.go:48-49
                        - CellAreaLabel (Label):
                            type: widget.Label
                            purpose: Display calculated cell area
                            file:tabs/builder_validation_tab.go:73
                            bindings: Updated by updateStats()
                  - ArrayPanel (VBox):
                      file:tabs/builder_validation_tab.go:617-623
                      components:
                        - ArrayTitle (Label):
                            type: widget.Label
                            text: "Array Config"
                            style: Bold
                            file:tabs/builder_validation_tab.go:618
                        - ArrayForm (Form):
                            type: widget.Form
                            file:tabs/builder_validation_tab.go:597-602
                            fields:
                              - rowsEntry:
                                  purpose: Array rows
                                  default: "4"
                                  file:tabs/builder_validation_tab.go:76-77
                                  bindings: OnChanged -> updateStats(), updates cfg.Rows
                              - colsEntry:
                                  purpose: Array columns
                                  default: "4"
                                  file:tabs/builder_validation_tab.go:79-80
                                  bindings: OnChanged -> updateStats(), updates cfg.Cols
                              - modeSelect:
                                  purpose: Operation mode selection
                                  options: ["storage", "memory", "compute"]
                                  default: cfg.Mode
                                  file:tabs/builder_validation_tab.go:99-104
                                  bindings: OnChanged -> updates cfg.Mode, updateModeHelp()
                              - archSelect:
                                  purpose: Architecture selection
                                  options: ["passive", "1t1r"]
                                  default: cfg.Architecture
                                  file:tabs/builder_validation_tab.go:106-109
                                  bindings: OnChanged -> updates cfg.Architecture
                        - ModeHelpText (Label):
                            type: widget.Label
                            purpose: Explain selected mode
                            wrapping: TextWrapWord
                            file:tabs/builder_validation_tab.go:83-104
                            bindings: Updated by updateModeHelp() when mode changes
                        - StatsBox (VBox):
                            file:tabs/builder_validation_tab.go:606-615
                            components:
                              - StatsTitle (Label):
                                  text: "Statistics"
                                  style: Bold
                                  file:tabs/builder_validation_tab.go:605
                              - TotalLabel:
                                  purpose: Total Cells count
                                  file:tabs/builder_validation_tab.go:112
                                  bindings: Updated by updateStats()
                              - AreaLabel:
                                  purpose: Array area (μm²)
                                  file:tabs/builder_validation_tab.go:113
                                  bindings: Updated by updateStats()
                              - WLLengthLabel:
                                  purpose: Word line length
                                  file:tabs/builder_validation_tab.go:114
                                  bindings: Updated by updateStats()
                              - BLLengthLabel:
                                  purpose: Bit line length
                                  file:tabs/builder_validation_tab.go:115
                                  bindings: Updated by updateStats()
                              - DensityLabel:
                                  purpose: Cell density (cells/μm²)
                                  file:tabs/builder_validation_tab.go:116
                                  bindings: Updated by updateStats()
                              - UtilizationLabel:
                                  purpose: Utilization percentage
                                  file:tabs/builder_validation_tab.go:117
                                  bindings: Updated by updateStats()
            - Separator (widget.Separator):
                file:tabs/builder_validation_tab.go:720
            - ActionButtons (HBox):
                file:tabs/builder_validation_tab.go:647-651
                components:
                  - GenerateAllBtn (Button):
                      type: widget.Button
                      text: "Generate All"
                      purpose: Generate LEF/LIB/V/Verilog/DEF/Config files
                      file:tabs/builder_validation_tab.go:271-352
                      async: true (runs in goroutine)
                      bindings:
                        - Disables all 3 action buttons during execution
                        - Updates statusLabel, logOutput
                        - Generates files in cells/ and output/exports/
                        - Updates verilogPreview, defPreview, layoutViz
                        - Re-enables buttons when complete
                  - ValidateAllBtn (Button):
                      type: widget.Button
                      text: "Validate All"
                      purpose: Run Yosys, DEF, cross-check, optional placement validation
                      file:tabs/builder_validation_tab.go:355-486
                      async: true
                      bindings:
                        - Disables action buttons
                        - Updates yosysResult, defResult, crossResult, placementResult
                        - Writes to logOutput
                        - Shows validation summary
                  - ExportPackageBtn (Button):
                      type: widget.Button
                      text: "Export Package"
                      purpose: Export complete package to output/exports/{design}/
                      file:tabs/builder_validation_tab.go:489-575
                      async: true
                      bindings:
                        - Creates directory structure
                        - Exports all files with README
                        - Shows dialog with absolute path
            - StatusBar (HBox):
                file:tabs/builder_validation_tab.go:712-715
                components:
                  - StatusLabel (Label):
                      type: widget.Label
                      purpose: Show current operation status
                      file:tabs/builder_validation_tab.go:264
                      bindings: Updated by all async operations
      - PreviewTabs (AppTabs):
          file:tabs/builder_validation_tab.go:640-644
          description: Middle section showing generated file previews
          components:
            - VerilogTab:
                components:
                  - VerilogStatsLabel:
                      purpose: Show instances/lines/size
                      file:tabs/builder_validation_tab.go:179
                      bindings: Updated after Verilog generation
                  - VerilogPreview (MultiLineEntry):
                      type: widget.MultiLineEntry
                      purpose: Show generated Verilog code
                      wrapping: TextWrapOff
                      file:tabs/builder_validation_tab.go:168-170
                      bindings: Updated by generateAllBtn
            - DEFTab:
                components:
                  - DEFStatsLabel:
                      purpose: Show components count and filename
                      file:tabs/builder_validation_tab.go:180
                      bindings: Updated after DEF generation
                  - DEFPreview (MultiLineEntry):
                      type: widget.MultiLineEntry
                      purpose: Show generated DEF placement
                      wrapping: TextWrapOff
                      file:tabs/builder_validation_tab.go:172-174
                      bindings: Updated by generateAllBtn
            - LayoutTab:
                components:
                  - LayoutViz (Label):
                      type: widget.Label
                      purpose: ASCII art visualization of crossbar layout
                      style: Monospace
                      file:tabs/builder_validation_tab.go:176-177
                      bindings: Updated by makeBuilderLayoutVisualization()
      - ValidationSection (VBox):
          file:tabs/builder_validation_tab.go:705-709
          description: Bottom section with validation results and log
          components:
            - ValidationSplit (HSplit):
                file:tabs/builder_validation_tab.go:693-697
                offset: 0.65 (65% validation, 35% OpenLane)
                components:
                  - ValidationResultsPanel (VBox):
                      file:tabs/builder_validation_tab.go:688-691
                      components:
                        - ValidationSummary (Label):
                            type: widget.Label
                            purpose: Overall pass/fail status
                            style: Bold
                            file:tabs/builder_validation_tab.go:186-187
                            bindings: Updated by validateAllBtn
                        - ValidationRow (HBox):
                            file:tabs/builder_validation_tab.go:654-662
                            components:
                              - YosysResult (Label):
                                  purpose: Verilog syntax validation result
                                  file:tabs/builder_validation_tab.go:183
                                  values: ["Not validated", "...", "✓ PASS", "✗ FAIL"]
                              - DEFResult (Label):
                                  purpose: DEF syntax validation result
                                  file:tabs/builder_validation_tab.go:184
                                  values: ["Not validated", "...", "✓ PASS", "✗ FAIL"]
                              - CrossResult (Label):
                                  purpose: LEF/LIB/V cross-check result
                                  file:tabs/builder_validation_tab.go:185
                                  values: ["Not validated", "...", "✓ PASS", "✗ FAIL"]
                              - PlacementResult (Label):
                                  purpose: OpenLane placement validation result
                                  file:tabs/builder_validation_tab.go:205
                                  values: ["Not validated", "...", "✓ PASS", "✗ ERROR", "⊝ SKIP"]
                  - OpenLanePanel (VBox):
                      file:tabs/builder_validation_tab.go:668-675
                      components:
                        - OpenLaneHelpText (Label):
                            text: "Optional: Enable placement validation if OpenLane/Docker is installed"
                            wrapping: TextWrapWord
                            file:tabs/builder_validation_tab.go:665-666
                        - DockerStatus (Label):
                            purpose: Show Docker image status
                            file:tabs/builder_validation_tab.go:203
                            values: ["Checking...", "✓ Docker image ready", "○ Docker image not pulled", "✗ OpenLane not available", "✓ Native tools detected"]
                            bindings: Updated by goroutine on tab load (line 208-231)
                        - PDKStatus (Label):
                            purpose: Show SKY130A PDK status
                            file:tabs/builder_validation_tab.go:204
                            values: ["Checking...", "✓ SKY130A PDK ready", "○ PDK not installed"]
                            bindings: Updated by goroutine on tab load
                        - PullImageBtn (Button):
                            type: widget.Button
                            text: "Pull OpenLane Image"
                            purpose: Download OpenLane Docker image
                            file:tabs/builder_validation_tab.go:234-258
                            async: true
                            visibility: Hidden if Docker not available (line 678-685)
                        - EnablePlacementCheck (Check):
                            type: widget.Check
                            text: "Enable OpenLane Placement Check"
                            purpose: Enable optional placement validation
                            file:tabs/builder_validation_tab.go:261
                            bindings: Checked during validateAllBtn execution (line 427)
            - LogSection (VBox):
                file:tabs/builder_validation_tab.go:699-703
                components:
                  - LogHeader (HBox):
                      components:
                        - LogTitle (Label):
                            text: "Validation Log"
                            style: Bold
                        - ClearLogBtn (Button):
                            purpose: Clear log output
                            file:tabs/builder_validation_tab.go:198-200
                  - LogOutput (MultiLineEntry):
                      type: widget.MultiLineEntry
                      purpose: Show validation/generation logs
                      wrapping: TextWrapWord
                      style: Monospace
                      file:tabs/builder_validation_tab.go:188-190
                      bindings: Updated by addLog() function (line 192-196)

  - name: LearnTab
    file:tabs/learn_tab.go:26
    description: Educational content with 3 topics explaining EDA concepts
    layout:
      - Header (VBox):
          file:tabs/learn_tab.go:96-102
          components:
            - HeaderTitle (Label):
                text: "FeCIM Array Builder - Learning Center"
                style: Bold, Centered
                file:tabs/learn_tab.go:97
            - HeaderSubtitle (Label):
                text: "Understanding OpenLane and where our tool fits in"
                file:tabs/learn_tab.go:98
            - Separator:
                file:tabs/learn_tab.go:99
      - Split (HSplit):
          file:tabs/learn_tab.go:92-93
          offset: 0.25 (25% sidebar, 75% content)
          components:
            - Sidebar (VBox):
                file:tabs/learn_tab.go:85-89
                components:
                  - SidebarTitle (Label):
                      text: "Topics"
                      style: Bold, Centered
                      file:tabs/learn_tab.go:81
                  - TopicSelector (List):
                      type: widget.List
                      purpose: Select learning topic
                      file:tabs/learn_tab.go:34-52
                      items:
                        - "1. What is FeCIM EDA?"
                        - "2. The Crossbar Architecture"
                        - "3. EDA Files We Generate"
                      bindings: OnSelected -> updates ContentScroll.Content (line 59-75)
                      default: Selected index 0 (line 78)
            - ContentScroll (Scroll):
                file:tabs/learn_tab.go:55-56
                minSize: 750x500
                components:
                  - DynamicContent (VBox):
                      purpose: Show content based on topic selection
                      options:
                        - makeIntroContent() (topic 0)
                        - makeCrossbarContent() (topic 1)
                        - makeFilesContent() (topic 2)

  - name: LearnContent_Intro
    file:tabs/learn_tab.go:118
    description: Topic 1 content with OpenLane flow explanation
    components:
      - Title (Label):
          text: "What is FeCIM EDA?"
          style: Bold
          file:tabs/learn_tab.go:119
      - Intro (Label):
          purpose: Explain educational nature and placeholder disclaimer
          wrapping: TextWrapWord
          file:tabs/learn_tab.go:121-122
      - OperationModesVisual (Canvas):
          purpose: Show 3 FeCIM modes (Storage, Memory, Compute)
          file:tabs/learn_visuals.go:640
          size: 640x300 (enforced by sizedContainer)
      - OpenLaneFlowDiagram (Canvas):
          purpose: Show RTL-to-GDSII pipeline with highlighted stages
          file:tabs/learn_visuals.go:41
          size: 800x340 (enforced by sizedContainer)
          highlights: Verilog, Floorplan (LEF), Placement (DEF) in CYAN
      - StagesExplained (Label):
          purpose: Text explanation of each OpenLane stage
          wrapping: TextWrapWord
          file:tabs/learn_tab.go:132-137
      - DoColumns (Grid):
          purpose: Two-column "What we do" vs "What we don't do"
          file:tabs/learn_tab.go:154
          columns: 2
          left: DoList (bullet list)
          right: DontList (bullet list)
      - DisclaimerCard (Card):
          purpose: Show disclaimer about not being affiliated with Rice/Tour
          file:tabs/learn_tab.go:157-158

  - name: LearnContent_Crossbar
    file:tabs/learn_tab.go:189
    description: Topic 2 content explaining Passive vs 1T1R architectures
    components:
      - Title (Label):
          text: "The Crossbar Architecture"
          style: Bold
          file:tabs/learn_tab.go:190
      - PassiveSection (VBox):
          components:
            - PassiveTitle (Label):
                text: "Passive Crossbar"
                style: Bold
                file:tabs/learn_tab.go:193
            - PassiveDesc (Label):
                purpose: Describe passive architecture pros/cons
                wrapping: TextWrapWord
                file:tabs/learn_tab.go:194-197
            - PassiveDiagram (Canvas):
                purpose: Isometric 3x3 passive crossbar visualization
                file:tabs/learn_visuals.go:222
                size: 620x440 (enforced by sizedContainer)
                shows: WL/BL layers, FeFET pillars, labels
      - 1T1RSection (VBox):
          components:
            - 1T1RTitle (Label):
                text: "1T1R (1 Transistor + 1 Resistor)"
                style: Bold
                file:tabs/learn_tab.go:202
            - 1T1RDesc (Label):
                purpose: Describe 1T1R architecture pros/cons
                wrapping: TextWrapWord
                file:tabs/learn_tab.go:203-206
            - 1T1RDiagram (Canvas):
                purpose: Isometric 3x3 1T1R crossbar visualization
                file:tabs/learn_visuals.go:360
                size: 620x440 (enforced by sizedContainer)
                shows: WL/BL/SL layers, transistor symbols, FeFET devices
      - ComparisonTable (Canvas):
          purpose: Visual comparison table (Passive vs 1T1R)
          file:tabs/learn_visuals.go:527
          size: 480x230 (enforced by sizedContainer)
          rows: Property, Cell Size, Area, Sneak Path, Max Array, Port Count, Fab Complex
      - SneakPathExplanation (VBox):
          components:
            - SneakPathTitle (Label):
                text: "The Sneak Path Problem"
                style: Bold
                file:tabs/learn_tab.go:215
            - SneakPathDesc (Label):
                purpose: Explain sneak path current issue
                wrapping: TextWrapWord
                file:tabs/learn_tab.go:216-217
      - RecommendationCard (Card):
          purpose: Show array size recommendations
          file:tabs/learn_tab.go:220-221
          text: "<= 16x16: Passive | 32x32: Either | >= 64x64: 1T1R"

  - name: LearnContent_Files
    file:tabs/learn_tab.go:257
    description: Topic 3 content showing EDA file formats
    components:
      - Title (Label):
          text: "EDA Files We Generate"
          style: Bold
          file:tabs/learn_tab.go:258
      - CardsGrid (Grid):
          purpose: Show 4 file format preview cards
          file:tabs/learn_tab.go:266
          columns: 2
          cards:
            - LEFPreviewCard:
                file:tabs/learn_visuals.go:790
                shows: LEF MACRO example with SIZE and PIN
                size: 360x224
            - DEFPreviewCard:
                file:tabs/learn_visuals.go:801
                shows: DEF COMPONENTS with FIXED placement
                size: 360x224
            - VerilogPreviewCard:
                file:tabs/learn_visuals.go:812
                shows: Verilog module with cell instantiation
                size: 360x224
            - LibertyPreviewCard:
                file:tabs/learn_visuals.go:824
                shows: Liberty library with timing (PLACEHOLDER)
                size: 360x224
      - FilePurposes (Label):
          purpose: Explain purpose of each file format
          wrapping: TextWrapWord
          file:tabs/learn_tab.go:270-281
          content: LEF, DEF, Verilog, Liberty, OpenLane Config explanations
      - ReferencesCard (Canvas):
          purpose: IEEE-formatted references with categories
          file:tabs/learn_visuals.go:840
          size: 420x310
          categories:
            - "EDA / OpenLane"
            - "FeCIM Device Physics"
            - "CIM Architecture"

DataFlow:
  - trigger: User changes rows/cols/width/height
    source: rowsEntry.OnChanged, colsEntry.OnChanged, widthEntry.OnChanged, heightEntry.OnChanged
    updates:
      - cfg.Rows, cfg.Cols (ArrayConfig)
      - cfg.CellWidth, cfg.CellHeight (ArrayConfig)
      - All stats labels (totalLabel, areaLabel, wlLengthLabel, blLengthLabel, densityLabel, utilizationLabel, cellAreaLabel)
    file:tabs/builder_validation_tab.go:120-159

  - trigger: User changes mode selection
    source: modeSelect.OnChanged
    updates:
      - cfg.Mode (ArrayConfig)
      - modeHelpText (displays mode explanation)
    file:tabs/builder_validation_tab.go:99-104

  - trigger: User changes architecture selection
    source: archSelect.OnChanged
    updates:
      - cfg.Architecture (ArrayConfig)
    file:tabs/builder_validation_tab.go:106-109

  - trigger: User clicks "Generate All"
    source: generateAllBtn.OnClick
    updates:
      - Disables all 3 action buttons
      - statusLabel -> "Generating..."
      - Generates cell files (LEF/LIB/V) to cells/fecim_bitcell/
      - Generates array Verilog to output/exports/fecim_crossbar_NxM.v
      - Generates DEF to output/exports/fecim_crossbar_NxM.def
      - Generates OpenLane config to output/exports/config.json
      - Updates verilogPreview, defPreview, layoutViz
      - Updates verilogStatsLabel, defStatsLabel
      - Re-enables buttons, statusLabel -> "All files generated"
    file:tabs/builder_validation_tab.go:271-352

  - trigger: User clicks "Validate All"
    source: validateAllBtn.OnClick
    updates:
      - Disables all 3 action buttons
      - statusLabel -> "Validating..."
      - Runs Yosys validation -> updates yosysResult
      - Runs DEF validation -> updates defResult
      - Runs LEF/LIB/V cross-check -> updates crossResult
      - (Optional) Runs OpenLane placement check -> updates placementResult
      - Updates validationSummary, logOutput
      - Re-enables buttons, statusLabel -> "All validations passed" or "Some validations failed"
    file:tabs/builder_validation_tab.go:355-486

  - trigger: User clicks "Export Package"
    source: exportPackageBtn.OnClick
    updates:
      - Disables all 3 action buttons
      - statusLabel -> "Exporting package..."
      - Creates directory output/exports/fecim_crossbar_NxM/
      - Copies all files (cells/, Verilog, DEF, config.json, design JSON, README)
      - Shows dialog with absolute path
      - Re-enables buttons, statusLabel -> "Package exported to ..."
    file:tabs/builder_validation_tab.go:489-575

  - trigger: User clicks "Pull OpenLane Image"
    source: pullImageBtn.OnClick
    updates:
      - dockerStatus -> "Pulling image..."
      - Executes docker pull via openlane.Manager
      - Updates dockerStatus -> "✓ Docker image ready" or "✗ Pull failed"
      - Adds progress messages to logOutput
    file:tabs/builder_validation_tab.go:234-258

  - trigger: Tab loads
    source: Tab initialization
    updates:
      - Spawns goroutine to detect OpenLane mode
      - Updates dockerStatus based on Docker availability
      - Updates pdkStatus based on SKY130A installation
      - Hides pullImageBtn if Docker not available
    file:tabs/builder_validation_tab.go:208-231, 678-685

  - trigger: User selects Learn topic
    source: topicSelector.OnSelected
    updates:
      - contentScroll.Content switches to makeIntroContent(), makeCrossbarContent(), or makeFilesContent()
      - contentScroll.Refresh()
    file:tabs/learn_tab.go:59-75

  - trigger: User switches view (standalone mode)
    source: viewSelector.OnChanged
    updates:
      - Shows selected tab content
      - Hides all other tab contents
      - Updates currentView state
    file:app.go:56-70

SharedState:
  - ArrayConfig:
      type: config.ArrayConfig
      purpose: Shared configuration between Builder tab controls
      file:config/types.go:28
      fields:
        - Rows: int (default 4)
        - Cols: int (default 4)
        - Mode: string (storage/memory/compute)
        - Architecture: string (passive/1t1r)
        - Technology: string (sky130)
        - CellWidth: float64 (0.46 μm)
        - CellHeight: float64 (2.72 μm)
      usage:
        - Passed to MakeBuilderValidationTab() (line module6-eda/pkg/gui/app.go:31, embedded.go:38)
        - Updated by form entries (rowsEntry, colsEntry, modeSelect, archSelect)
        - Read by export/validation functions

BugDetails: []

ArchitectureNotes:
  - pattern: Minimal embedded wrapper
    description: |
      EmbeddedEDAApp is extremely simple - just wraps CreateModuleContent().
      Unlike other modules, no complex state management.
    file:embedded.go:14-21

  - pattern: Shared config pointer
    description: |
      ArrayConfig is created once in app.go or embedded.go and passed by pointer
      to MakeBuilderValidationTab(). All form controls update this shared instance.
    file:app.go:20-28, embedded.go:26-34

  - pattern: Goroutine safety
    description: |
      All async operations (generateAllBtn, validateAllBtn, exportPackageBtn, pullImageBtn)
      wrap UI updates in fyne.Do() to ensure thread safety.
    example: module6-eda/pkg/gui/tabs/builder_validation_tab.go:277-280, 301-303

  - pattern: Helper function for stats
    description: |
      updateStats() calculates derived metrics (area, density, utilization) from
      user inputs and updates all stat labels. Called by OnChanged handlers.
    file:tabs/builder_validation_tab.go:120-159

  - pattern: Mode help text
    description: |
      updateModeHelp() provides contextual explanation when user selects a mode.
      Helps educate users about storage/memory/compute differences.
    file:tabs/builder_validation_tab.go:86-97

  - pattern: Validation workflow
    description: |
      Generate -> Validate -> Export three-step workflow with clear button states.
      Buttons disable during async ops to prevent race conditions.
    file:tabs/builder_validation_tab.go:267-575

  - pattern: Optional OpenLane integration
    description: |
      OpenLane placement validation is optional (checkbox-controlled).
      Docker status checked on startup, pull button shown only if needed.
    file:tabs/builder_validation_tab.go:208-231, 427-461

  - pattern: Educational visuals
    description: |
      Learn tab uses custom canvas rendering (learn_visuals.go) for diagrams.
      sizedContainer() enforces minimum sizes so VBox/VScroll allocate proper space.
    file:tabs/learn_tab.go:14-23, learn_visuals.go

  - pattern: File format cards
    description: |
      FileFormatCard() creates reusable styled cards with syntax-highlighted previews.
      Used for LEF/DEF/Verilog/Liberty examples in Learn tab.
    file:tabs/learn_visuals.go:757-787

  - pattern: IEEE references
    description: |
      ReferencesCard() shows properly formatted academic references grouped by category.
      Emphasizes educational nature of the tool.
    file:tabs/learn_visuals.go:840-960

EducationalDisclaimers:
  - location: Learn Tab Intro
    text: |
      "This is an educational tool that demonstrates how FeCIM arrays could integrate
      with open-source EDA. All timing values are placeholders - real values require
      SPICE characterization with validated models."
    file:tabs/learn_tab.go:121-122

  - location: Config Types
    text: |
      "⚠️ PLACEHOLDER TIMING VALUES
      These are estimates requiring FeFET characterization via SPICE simulation
      Real values need: SPICE compact model + Liberty characterization"
    file:config/types.go:18-24

  - location: Cell Config
    text: |
      "PLACEHOLDER timing values"
    file:config/types.go:48-52

TestCoverage:
  - file:config/types_test.go
    description: Unit tests for config types

KeyFeatures:
  - name: Unified Builder Tab
    description: |
      Consolidated 6 previous tabs into 1 tab with logical sections:
      - Cell/Array config (top)
      - Preview tabs (middle)
      - Validation + OpenLane status (bottom)
    benefit: Cleaner UX, less tab switching

  - name: Real-time Statistics
    description: |
      As user types rows/cols/dimensions, stats update immediately:
      - Total cells, array area, WL/BL length, density, utilization
    file:tabs/builder_validation_tab.go:120-159

  - name: Mode-aware Help
    description: |
      When user selects storage/memory/compute, help text explains the mode.
    file:tabs/builder_validation_tab.go:86-104

  - name: Comprehensive Validation
    description: |
      4 validation checks: Yosys (Verilog), DEF syntax, LEF/LIB/V cross-check,
      optional OpenLane placement
    file:tabs/builder_validation_tab.go:355-486

  - name: Package Export
    description: |
      Export complete design package with all files + README to
      output/exports/{design}/ ready for OpenLane integration
    file:tabs/builder_validation_tab.go:489-575

  - name: Docker Integration
    description: |
      Detects OpenLane Docker image, offers pull button if missing,
      optional placement validation if available
    file:tabs/builder_validation_tab.go:208-231, 234-258

  - name: Educational Learn Tab
    description: |
      3 topics with custom visuals:
      1. What is FeCIM EDA? (OpenLane flow diagram, operation modes)
      2. Crossbar Architecture (Passive vs 1T1R isometric diagrams, comparison table)
      3. EDA Files (LEF/DEF/Verilog/Liberty format cards, references)
    file:tabs/learn_tab.go

  - name: Isometric Crossbar Visualizations
    description: |
      Beautiful 3D-style diagrams showing WL/BL layers, FeFET devices,
      transistor symbols (1T1R), with color-coded legends
    file:tabs/learn_visuals.go:222-520

  - name: File Format Preview Cards
    description: |
      Styled cards showing actual LEF/DEF/Verilog/Liberty syntax examples
    file:tabs/learn_visuals.go:790-833

  - name: Academic References
    description: |
      IEEE-formatted references grouped by category (EDA, FeCIM Physics, CIM Architecture)
    file:tabs/learn_visuals.go:840-960

FilesGenerated:
  - path: cells/fecim_bitcell/fecim_bitcell.lef
    purpose: Cell abstract (LEF)
    generator: export.GenerateLEF(cellCfg)
    file:tabs/builder_validation_tab.go:293

  - path: cells/fecim_bitcell/fecim_bitcell.lib
    purpose: Timing library (Liberty) - PLACEHOLDER VALUES
    generator: export.GenerateLiberty(cellCfg)
    file:tabs/builder_validation_tab.go:294

  - path: cells/fecim_bitcell/fecim_bitcell.v
    purpose: Cell Verilog model (behavioral)
    generator: export.GenerateCellVerilog(cellCfg)
    file:tabs/builder_validation_tab.go:295

  - path: output/exports/fecim_crossbar_NxM.v
    purpose: Array Verilog netlist
    generator: export.GenerateArrayVerilog(*cfg)
    file:tabs/builder_validation_tab.go:300

  - path: output/exports/fecim_crossbar_NxM.def
    purpose: Physical placement (DEF)
    generator: generateBuilderDEF(*cfg)
    file:tabs/builder_validation_tab.go:319

  - path: output/exports/config.json
    purpose: OpenLane configuration
    generator: export.GenerateOpenLaneConfig(*cfg)
    file:tabs/builder_validation_tab.go:339

  - path: output/exports/fecim_crossbar_NxM/
    purpose: Complete package directory (created by Export Package)
    contains: cells/, Verilog, DEF, config.json, design JSON, README.md
    file:tabs/builder_validation_tab.go:500-555

ValidationTools:
  - name: Yosys
    purpose: Verilog syntax validation
    function: validation.ValidateVerilogWithCell()
    file:tabs/builder_validation_tab.go:379-387

  - name: DEF Parser
    purpose: DEF syntax validation + stats extraction
    function: validation.ValidateDEF()
    file:tabs/builder_validation_tab.go:394-404

  - name: Cross-check
    purpose: Verify LEF/LIB/V cell/pin names match
    function: validation.CrossCheckFiles()
    file:tabs/builder_validation_tab.go:415-424

  - name: OpenLane Placement
    purpose: Optional - validate DEF placement with OpenROAD
    function: validation.RunPlacementCheck()
    file:tabs/builder_validation_tab.go:442-459
    enabled: Only if enablePlacementCheck is checked

IntegrationPoints:
  - name: EmbeddedApp Interface
    methods: [BuildContent, Start, Stop]
    file:embedded.go:44-60
    purpose: Embeds into unified visualizer (cmd/fecim-visualizer)

  - name: Standalone Entry
    file: cmd/eda-gui/main.go:10-14
    purpose: Run as standalone app via CreateMainWindow()

  - name: Export Package Integration
    purpose: Exported packages can be directly copied to OpenLane designs/
    workflow: "Copy output/exports/fecim_crossbar_NxM/ to designs/ -> run flow.tcl"
    file:tabs/builder_validation_tab.go:537-554

ExternalDependencies:
  - name: OpenLane Docker
    optional: true
    purpose: Placement validation
    detection: openlane.NewManager().DetectMode()
    file:tabs/builder_validation_tab.go:209-210

  - name: SKY130A PDK
    optional: true
    purpose: Placement validation
    detection: manager.IsPDKInstalled()
    file:tabs/builder_validation_tab.go:225-229

  - name: Yosys
    optional: false (for validation)
    purpose: Verilog syntax checking
    file:tabs/builder_validation_tab.go:379

UIPatterns:
  - pattern: Button state management
    description: |
      During async operations, all 3 action buttons are disabled to prevent
      concurrent file writes. Re-enabled on completion.
    file:tabs/builder_validation_tab.go:272-275, 345-348

  - pattern: Status label updates
    description: |
      Single statusLabel shows current operation state, updated via fyne.Do()
    file:tabs/builder_validation_tab.go:264, 278, 344

  - pattern: Log output pattern
    description: |
      addLog() helper writes timestamped messages to logOutput with fyne.Do()
    file:tabs/builder_validation_tab.go:192-196

  - pattern: Progress feedback
    description: |
      Generate/Validate/Export show step-by-step progress in log:
      "[1/6] Generating cell library..."
    file:tabs/builder_validation_tab.go:508-537

  - pattern: Dialog confirmation
    description: |
      Export Package shows success dialog with absolute path on completion
    file:tabs/builder_validation_tab.go:567-571

PerformanceNotes:
  - operation: Generate All
    timing: ~100-300ms for 4x4 array
    bottleneck: File I/O (sequential writes)
    file:tabs/builder_validation_tab.go:271-352

  - operation: Validate All
    timing: ~500ms-2s (depends on Yosys)
    bottleneck: External Yosys process
    file:tabs/builder_validation_tab.go:355-486

  - operation: Export Package
    timing: ~200-500ms
    bottleneck: Directory creation + file copies
    file:tabs/builder_validation_tab.go:489-575

  - operation: Pull Docker Image
    timing: ~5-10 minutes (first time)
    bottleneck: Network download
    file:tabs/builder_validation_tab.go:234-258

FutureEnhancements:
  - feature: GDS2 export
    description: Generate actual layout GDS2 file (requires Magic/KLayout integration)

  - feature: Timing characterization wizard
    description: Guide user through SPICE setup for real Liberty values

  - feature: Batch export
    description: Generate multiple array sizes at once (4x4, 8x8, 16x16, etc.)

  - feature: Template library
    description: Save/load array configurations as templates

  - feature: Advanced placement
    description: Custom placement algorithms beyond fixed grid

  - feature: Power analysis
    description: Estimate power consumption based on array config
```
