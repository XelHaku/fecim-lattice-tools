# Plan: EDA Learn UI Improvement

## Context

### Original Request
Fix and improve the EDA/Learn UI in module6-eda to provide a better learning experience with fewer views, better diagrams, and improved content organization.

### Current State Analysis
The Learn tab currently has 6 topics spread across 2 files:
- `learn_tab.go` (372 lines) - Topic definitions and content functions
- `learn_visuals.go` (967 lines) - Diagram rendering functions

**Current 6 Topics:**
1. **Overview** - What the tool does/doesn't do, disclaimers, operation modes visual
2. **OpenLane Flow** - RTL-to-GDSII pipeline explanation with flow diagram
3. **Where We Fit In** - Same flow diagram with highlights + crossbar diagram + explanation
4. **What We Generate** - 4 file format preview cards (LEF, DEF, Verilog, Liberty)
5. **Cell Types** - Passive vs 1T1R comparison with diagrams and table
6. **References** - IEEE-style citations

### Problems Identified
1. **Too many views** - 6 topics for what should be streamlined learning
2. **Repeated content** - Flow diagram appears in topics 2 AND 3 (lines 141, 185)
3. **Dense text** - ASCII-art style headers like "WHAT WE DO:" and "----------"
4. **Small diagrams** - Crossbar visualizations at 420x310px (lines 360, 523)
5. **Disjointed narrative** - Topics don't flow as a learning journey
6. **Technical jargon** - "RTL-to-GDSII" without beginner explanation
7. **Scattered disclaimers** - Disclaimers in Overview AND References AND What We Generate

---

## Work Objectives

### Core Objective
Consolidate the Learn tab from 6 topics to 3 well-structured topics that tell a coherent story with larger, clearer diagrams.

### Deliverables
1. Consolidated topic structure (6 -> 3)
2. Larger, clearer diagrams (40-60% size increase)
3. Progressive narrative flow (What is this? -> How does it work? -> Technical details)
4. Single consolidated disclaimer location
5. Cleaner text formatting (no ASCII headers)

### Definition of Done
- [ ] Learn tab loads without errors
- [ ] Only 3 topics in sidebar
- [ ] All diagrams visible and appropriately sized
- [ ] No duplicated content
- [ ] Single disclaimer section
- [ ] User can navigate all 3 topics

---

## Proposed New Structure

### Topic 1: "What is FeCIM EDA?" (Merged: Overview + Where We Fit In + OpenLane Flow educational content)
**Purpose:** Introduction and context - answers "What am I looking at?"

**Content:**
- Brief intro (2-3 sentences, no ASCII headers)
- LARGER Operation Modes diagram (current: 420x200, new: 560x260)
- OpenLane Flow diagram WITH highlights showing our contribution (removes duplication)
- **"OpenLane Stages Explained" section** (preserved from makeOpenLaneFlowContent lines 145-169)
- Compact "What we do / What we don't do" as two columns with Unicode bullets
- Single disclaimer banner at bottom (using widget.NewCard)

**Key Changes:**
- Remove duplicated flow diagram from topic 2
- Move operation modes diagram to be more prominent
- **Preserve educational "Stages Explained" content** from the removed makeOpenLaneFlowContent()
- Clean up text formatting (no ASCII headers or dashed underlines)

### Topic 2: "The Crossbar Architecture" (Merged: Cell Types)
**Purpose:** Explain the physical structure - answers "How does it work?"

**Content:**
- Passive crossbar diagram with title and description (540x400)
- 1T1R crossbar diagram with title and description (540x400)
- Passive vs 1T1R comparison table
- Sneak path problem explanation (simplified prose)
- Array size recommendations (in styled card)

**Key Changes:**
- **STACK DIAGRAMS VERTICALLY** (not side-by-side) - fits 600px scroll width
- Increase diagram size by ~30% (420x310 -> 540x400)
- Simplify sneak path explanation to clean prose (no ASCII art)
- Add recommendation box using widget.NewCard

### Topic 3: "EDA Files We Generate" (Merged: What We Generate + References)
**Purpose:** Technical details - answers "What does it produce?"

**Content:**
- 4 file format preview cards (LEF, DEF, Verilog, Liberty) - LARGER
- Brief explanation of each format's purpose (inline, not separate section)
- References section at bottom (consolidated)

**Key Changes:**
- Increase card sizes
- Move references from separate topic to bottom of this topic
- Remove redundant disclaimer (already in Topic 1)

---

## Content Reorganization Map

| Old Location | New Location | Notes |
|--------------|--------------|-------|
| makeOverviewContent() -> intro text | Topic 1: cleaned up prose | Remove ASCII headers |
| makeOverviewContent() -> OperationModesVisual() | Topic 1: LARGER version (560x260) | |
| makeOpenLaneFlowContent() -> OpenLaneFlowDiagram(false) | REMOVED (duplicate) | |
| **makeOpenLaneFlowContent() -> "THE STAGES EXPLAINED" (lines 145-169)** | **Topic 1: preserved as "Stages Explained" section** | **CRITICAL: Keep educational content** |
| makeWhereWeFitContent() -> OpenLaneFlowDiagram(true) | Topic 1: always shows highlights | API changed to no parameter |
| makeWhereWeFitContent() -> IsometricCrossbar() | Topic 2: LARGER version (540x400) | Stacked vertically |
| makeWhatWeGenerateContent() -> file cards | Topic 3: LARGER cards (380x200) | |
| makeCellTypesContent() -> diagrams + table | Topic 2: LARGER diagrams | Both stacked vertically |
| makeReferencesContent() -> ReferencesCard() | Topic 3: bottom section | |

---

## Diagram Improvements

### 1. OpenLane Flow Diagram (learn_visuals.go:41-179)
**Current:** 620x250px
**New:** 720x300px
**Changes:**
- Increase boxW from 120 to 140
- Increase boxH from 55 to 65
- Increase container size to 720x300
- Always show highlights (remove `showOurContribution` parameter - always true)

### 2. Isometric Crossbar (learn_visuals.go:227-363)
**Current:** 420x310px
**New:** 540x400px
**Changes:**
- Increase cellSize from 42 to 52
- Increase layerGap from 55 to 70
- Adjust startX/startY for new size
- Increase container to 540x400

### 3. 1T1R Crossbar (learn_visuals.go:365-526)
**Current:** 420x310px
**New:** 540x400px
**Changes:**
- Same scale increases as passive crossbar
- Ensure consistent sizing

### 4. Operation Modes Visual (learn_visuals.go:645-756)
**Current:** 420x200px
**New:** 560x260px
**Changes:**
- Increase boxW from 110 to 140
- Increase boxH from 90 to 110
- Adjust circle position

### 5. File Format Cards (learn_visuals.go:763-839)
**Current:** 340x175px
**New:** 380x200px
**Changes:**
- Increase card width from 340 to 380
- Increase content area height

---

## Guardrails

### Must Have
- All 3 new topics functional and navigable
- All diagrams render correctly at new sizes
- No runtime errors or panics
- Proper use of `fyne.Do()` for UI updates from goroutines
- References still accessible (in Topic 3)

### Must NOT Have
- More than 3 topics in sidebar
- Duplicated diagrams or content
- ASCII-art style headers (WHAT WE DO:, ----------)
- Multiple disclaimer sections
- Breaking changes to the embedded app interface

---

## Implementation Tasks

### Task 1: Update Topic Structure in learn_tab.go
**File:** `<local-path>`
**Lines:** 16-65

**Changes:**
1. Update topics array (lines 16-23) from 6 items to 3:
   ```go
   topics := []string{
       "1. What is FeCIM EDA?",
       "2. The Crossbar Architecture",
       "3. EDA Files We Generate",
   }
   ```

2. Update switch statement (lines 43-65) to map to new content functions:
   ```go
   switch id {
   case 0:
       content = makeIntroContent()
   case 1:
       content = makeCrossbarContent()
   case 2:
       content = makeFilesContent()
   }
   ```

**Acceptance Criteria:**
- [ ] Sidebar shows exactly 3 topics
- [ ] Clicking each topic loads correct content
- [ ] No compilation errors

---

### Task 2: Create makeIntroContent() Function
**File:** `<local-path>`
**Lines:** 92-135 (replace makeOverviewContent)
**DEPENDS ON:** Task 7 (API change must be done first so OpenLaneFlowDiagram() call compiles)

**Content Structure:**
```go
func makeIntroContent() fyne.CanvasObject {
    // Title
    title := widget.NewLabelWithStyle("What is FeCIM EDA?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

    // Clean intro paragraph (no ASCII headers)
    intro := widget.NewLabel(`Module 6 is an Array Builder that generates EDA files for integrating
FeCIM crossbar arrays into the OpenLane RTL-to-GDSII flow.

This is an educational tool that demonstrates how ferroelectric compute-in-memory
arrays could integrate with open-source EDA tools. All timing values are placeholders
that would require SPICE characterization with validated device models.`)
    intro.Wrapping = fyne.TextWrapWord

    // Operation Modes Visual (LARGER - 560x260)
    modesVisual := OperationModesVisual()

    // OpenLane Flow Diagram (WITH highlights - always)
    flowDiagram := OpenLaneFlowDiagram()  // No parameter - always shows highlights

    // Two-column layout for capabilities
    doColumn := makeBulletList("What We Do",
        "Generate LEF files (cell abstracts)",
        "Generate Liberty files (timing placeholders)",
        "Generate Verilog netlists (behavioral models)",
        "Generate DEF files (physical placement)",
        "Export OpenLane configuration")

    dontColumn := makeBulletList("What We Don't Do",
        "Provide validated FeFET device models",
        "Generate production-ready layouts",
        "Characterize real timing values",
        "Fabricate chips")

    columnsLayout := container.NewGridWithColumns(2, doColumn, dontColumn)

    // Disclaimer banner (styled)
    disclaimer := widget.NewCard("", "", widget.NewLabel(
        "This project is not affiliated with or endorsed by external research institution, Dr. external research group, or any foundry."))

    return container.NewVBox(title, widget.NewSeparator(), intro, ...)
}

// Helper function for bullet lists (add to learn_tab.go)
func makeBulletList(header string, items ...string) fyne.CanvasObject {
    headerLabel := widget.NewLabelWithStyle(header, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

    var listItems []fyne.CanvasObject
    listItems = append(listItems, headerLabel)

    for _, item := range items {
        bullet := widget.NewLabel("  " + string('\u2022') + " " + item)  // Unicode bullet
        bullet.Wrapping = fyne.TextWrapWord
        listItems = append(listItems, bullet)
    }

    return container.NewVBox(listItems...)
}
```

**Two-Column Layout Details:**
- **Container:** `container.NewGridWithColumns(2, doColumn, dontColumn)`
- **Each column:** VBox containing:
  - Header: `widget.NewLabelWithStyle(text, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})`
  - Bullet items: `widget.NewLabel("  \u2022 " + itemText)` (Unicode bullet U+2022)
- **Headers:** Placed INSIDE each column at top, not above the grid
- **No ASCII:** Replace "WHAT WE DO:\n-----------" with clean header label + bullet list

**Acceptance Criteria:**
- [ ] Clean text formatting (no ASCII art, no dashed underlines)
- [ ] Operation modes diagram visible and larger
- [ ] Flow diagram shows highlights
- [ ] Two-column layout with proper bullet points (Unicode, not ASCII)
- [ ] Single disclaimer at bottom using widget.NewCard for styling

---

### Task 3: Create makeCrossbarContent() Function
**File:** `<local-path>`
**Lines:** 282-349 (replace makeCellTypesContent)
**DEPENDS ON:** Task 7 (API change must be done first)

**CRITICAL: Diagram Layout Decision**
The two 540px-wide diagrams cannot fit side-by-side in the 600px scroll area.
**Solution:** Stack diagrams VERTICALLY with labels between them.

**Content Structure:**
```go
func makeCrossbarContent() fyne.CanvasObject {
    title := widget.NewLabelWithStyle("Cell Types: Passive vs 1T1R",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

    // PASSIVE CROSSBAR - stacked vertically, NOT side-by-side
    passiveTitle := widget.NewLabelWithStyle("Passive Crossbar",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
    passiveDiagram := IsometricCrossbar(3, 3, true)  // 540x400px

    passiveDesc := widget.NewLabel(`Simple structure with direct cell connections.
Ports: WL[], BL[], VDD, VSS  |  Cell Size: 0.46 x 2.72 um
Pros: Dense packing, lower complexity  |  Cons: Sneak path currents, limited to ~32x32`)
    passiveDesc.Wrapping = fyne.TextWrapWord

    // 1T1R CROSSBAR - below passive
    oneToneRTitle := widget.NewLabelWithStyle("1T1R Crossbar (1 Transistor + 1 Resistor)",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
    oneToneRDiagram := Isometric1T1RCrossbar(3, 3)  // 540x400px

    oneToneRDesc := widget.NewLabel(`Transistor-isolated cells eliminate sneak paths.
Ports: WL[], BL[], SL[], VDD, VSS  |  Cell Size: 0.92 x 2.72 um (2x width)
Pros: No sneak paths, scales to 128x128+  |  Cons: Larger area, complex routing`)
    oneToneRDesc.Wrapping = fyne.TextWrapWord

    // Comparison table
    comparisonTable := CellComparisonTable()

    // Sneak path explanation (simplified - prose, no ASCII)
    sneakTitle := widget.NewLabelWithStyle("The Sneak Path Problem",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
    sneakExplain := widget.NewLabel(`In passive arrays, unintended current paths form through
neighboring cells. Reading cell (0,0) creates parallel paths through adjacent cells,
with error growing as N squared for an NxN array.`)
    sneakExplain.Wrapping = fyne.TextWrapWord

    // Recommendation box (styled card)
    recommendation := widget.NewCard("Array Size Recommendation", "",
        widget.NewLabel("Up to 16x16: Passive  |  32x32: Either  |  64x64+: 1T1R required"))

    return container.NewVBox(
        title, widget.NewSeparator(),
        passiveTitle, passiveDiagram, passiveDesc,
        widget.NewSeparator(),
        oneToneRTitle, oneToneRDiagram, oneToneRDesc,
        widget.NewSeparator(),
        comparisonTable,
        widget.NewSeparator(),
        sneakTitle, sneakExplain,
        recommendation,
    )
}
```

**Key Changes:**
- **STACK VERTICALLY** - Each diagram on its own row with title above and description below
- Remove side-by-side layout that would overflow 600px scroll width
- Simplify sneak path explanation to prose (no ASCII diagram)
- Add visual recommendation box using `widget.NewCard`

**Acceptance Criteria:**
- [ ] Passive diagram visible with title and description
- [ ] 1T1R diagram visible below passive with title and description
- [ ] Both diagrams at 540x400px (larger than before)
- [ ] No horizontal overflow in scroll area
- [ ] Comparison table intact
- [ ] Sneak path explanation as clean prose
- [ ] Recommendation in styled card

---

### Task 4: Create makeFilesContent() Function
**File:** `<local-path>`
**Lines:** 227-280 (replace makeWhatWeGenerateContent)
**DEPENDS ON:** Task 7 (API change must be done first)

**Content Structure:**
```go
func makeFilesContent() fyne.CanvasObject {
    title := widget.NewLabelWithStyle("EDA Files We Generate",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

    intro := widget.NewLabel("The Array Builder generates the following files for OpenLane integration:")
    intro.Wrapping = fyne.TextWrapWord

    // 2x2 grid of file format cards (LARGER - 380x200 each)
    lefCard := LEFPreviewCard()
    defCard := DEFPreviewCard()
    verilogCard := VerilogPreviewCard()
    libertyCard := LibertyPreviewCard()

    cardsRow1 := container.NewHBox(lefCard, defCard)
    cardsRow2 := container.NewHBox(verilogCard, libertyCard)

    // File purposes (clean prose, no ASCII headers)
    purposesTitle := widget.NewLabelWithStyle("File Purposes",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
    purposes := widget.NewLabel(`LEF (Library Exchange Format): Defines cell geometry and pin locations as an abstract view.

DEF (Design Exchange Format): Physical placement with X,Y coordinates. The FIXED keyword prevents auto-placement.

Verilog Netlist: Structural description of the array with black-box cell instantiations.

Liberty (.lib): Timing information for synthesis. All values are placeholders requiring SPICE characterization.

OpenLane Config (JSON): Points OpenLane to our custom files.`)
    purposes.Wrapping = fyne.TextWrapWord

    // References section (moved from separate topic)
    refsTitle := widget.NewLabelWithStyle("References",
        fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
    refsCard := ReferencesCard()

    return container.NewVBox(
        title, widget.NewSeparator(),
        intro,
        cardsRow1, cardsRow2,
        widget.NewSeparator(),
        purposesTitle, purposes,
        widget.NewSeparator(),
        refsTitle, refsCard,
    )
}
```

**Key Changes:**
- Use LARGER file format cards (380x200px)
- Include references inline at bottom
- Remove redundant disclaimer (consolidated in Topic 1)
- Clean prose descriptions (no "FILE PURPOSES:\n--------------")

**Acceptance Criteria:**
- [ ] All 4 file cards visible in 2x2 layout
- [ ] Cards larger and readable (380x200px)
- [ ] References section at bottom
- [ ] No duplicate disclaimer

---

### Task 5: Remove Obsolete Functions and Preserve Educational Content
**File:** `<local-path>`
**DEPENDS ON:** Tasks 2, 3, 4 (content functions must be created first)

**Functions to Remove:**
- `makeOpenLaneFlowContent()` (lines 137-179) - merged into makeIntroContent
- `makeWhereWeFitContent()` (lines 181-225) - merged into makeIntroContent
- `makeReferencesContent()` (lines 351-371) - merged into makeFilesContent

**CRITICAL: Preserve Educational Content from makeOpenLaneFlowContent()**
The "THE STAGES EXPLAINED" content (lines 143-169) contains valuable learning material.
This content MUST be preserved by moving it to makeIntroContent().

**Content to Preserve (from lines 145-169):**
```
1. SYNTHESIS (Yosys) - Converts behavioral Verilog to gate-level netlist
2. FLOORPLAN - Defines die area and I/O pin locations
3. PLACEMENT (RePlAce + OpenDP) - Assigns X,Y coordinates to every cell
4. CTS (Clock Tree Synthesis) - Distributes clock signal evenly
5. ROUTING (TritonRoute) - Draws metal wire connections
6. SIGNOFF & GDSII - DRC/LVS verification, final output
```

**How to Preserve:**
Add the following to makeIntroContent() AFTER the flow diagram:
```go
// Stage explanations (preserved from makeOpenLaneFlowContent)
stagesTitle := widget.NewLabelWithStyle("OpenLane Stages Explained",
    fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

stagesContent := widget.NewLabel(`1. Synthesis (Yosys): Converts behavioral Verilog to gate-level netlist
   Example: "a & b" becomes sky130_fd_sc_hd__and2_1

2. Floorplan: Defines die area and I/O pin locations

3. Placement (RePlAce + OpenDP): Assigns X,Y coordinates to every cell

4. Clock Tree Synthesis: Distributes clock signal evenly
   Note: FeCIM arrays often skip this stage

5. Routing (TritonRoute): Draws metal wire connections

6. Signoff & GDSII: DRC/LVS verification, final output`)
stagesContent.Wrapping = fyne.TextWrapWord
```

**Acceptance Criteria:**
- [ ] No unused functions remain
- [ ] No compilation errors
- [ ] All references to removed functions updated
- [ ] "Stages Explained" educational content preserved in makeIntroContent()
- [ ] Stage descriptions maintain their educational value

---

### Task 6: Scale Up Diagram Sizes in learn_visuals.go
**File:** `<local-path>`

**6a. OpenLaneFlowDiagram (lines 41-179)**
- Line 43: `boxW := float32(140)` (was 120)
- Line 44: `boxH := float32(65)` (was 55)
- Line 176: `cont.Resize(fyne.NewSize(720, 300))` (was 620, 250)
- Remove `showOurContribution` parameter - always show highlights

**6b. IsometricCrossbar (lines 227-363)**
- Line 232: `cellSize := float32(52)` (was 42)
- Line 242: `layerGap := float32(70)` (was 55)
- Line 360: `cont.Resize(fyne.NewSize(540, 400))` (was 420, 310)

**6c. Isometric1T1RCrossbar (lines 365-526)**
- Line 370: `cellSize := float32(52)` (was 42)
- Line 377: `layerGap := float32(70)` (was 55)
- Line 523: `cont.Resize(fyne.NewSize(540, 400))` (was 420, 310)

**6d. OperationModesVisual (lines 645-756)**
- Line 657: `boxW := float32(140)` (was 110)
- Line 658: `boxH := float32(110)` (was 90)
- Line 753: `cont.Resize(fyne.NewSize(560, 260))` (was 420, 200)

**6e. FileFormatCard (lines 763-793)**
- Line 766: `headerBg.Resize(fyne.NewSize(380, 36))` (was 340, 32)
- Line 776: `contentBg.Resize(fyne.NewSize(380, 160))` (was 340, 140)
- Line 790: `card.Resize(fyne.NewSize(380, 200))` (was 340, 175)

**Acceptance Criteria:**
- [ ] All diagrams render at new sizes
- [ ] No clipping or overflow issues
- [ ] Proportions maintained
- [ ] Labels still readable

---

### Task 7: Update OpenLaneFlowDiagram API
**File:** `<local-path>`
**Lines:** 41-179

**Change:**
```go
// Before
func OpenLaneFlowDiagram(showOurContribution bool) fyne.CanvasObject

// After
func OpenLaneFlowDiagram() fyne.CanvasObject
```

- Remove the boolean parameter
- Always show contribution highlights (lines 95-99 become unconditional)
- Update all call sites in learn_tab.go

**Acceptance Criteria:**
- [ ] Function signature updated
- [ ] All call sites updated
- [ ] Highlights always visible
- [ ] No compilation errors

---

### Task 8: Test and Verify
**DEPENDS ON:** All previous tasks

**Step 1: Run Unit Tests**
```bash
cd <local-path>
go test ./module6-eda/...
```

**Step 2: Build Application**
```bash
go build -o fecim-visualizer ./cmd/fecim-visualizer
```

**Step 3: Manual Verification**
```bash
./fecim-visualizer
```
1. Open EDA tab
2. Navigate to Learn sub-tab
3. Click each of the 3 topics:
   - Topic 1 "What is FeCIM EDA?" - verify operation modes + flow diagram + two-column layout + disclaimer
   - Topic 2 "The Crossbar Architecture" - verify BOTH diagrams render (stacked vertically), table, sneak path explanation
   - Topic 3 "EDA Files We Generate" - verify all 4 file cards + references section
4. Verify diagrams render correctly at larger sizes
5. Verify no duplicate content (flow diagram appears ONCE in Topic 1)
6. Verify single disclaimer location (only in Topic 1)
7. Verify educational "Stages Explained" content is present in Topic 1
8. Scroll through Topic 2 - verify no horizontal overflow

**Acceptance Criteria:**
- [ ] `go test ./module6-eda/...` passes with no failures
- [ ] `go build` completes without errors
- [ ] App runs without panics
- [ ] All 3 topics accessible and navigable
- [ ] All diagrams visible and larger (25%+ increase)
- [ ] Topic 2 diagrams stacked vertically without horizontal overflow
- [ ] No duplicated content
- [ ] Educational content preserved

---

## Task Dependencies

**CRITICAL: Task 7 (API change) must complete BEFORE Tasks 2-4 can call OpenLaneFlowDiagram()**

```
PHASE 1: API and Infrastructure (must complete first)
    Task 7 (OpenLaneFlowDiagram API change - remove bool parameter)
    Task 6 (diagram sizes) - can run parallel with Task 7

PHASE 2: Content Functions (depends on Phase 1)
    Task 1 (topic structure)
        └─> Task 2 (makeIntroContent) - calls OpenLaneFlowDiagram()
        └─> Task 3 (makeCrossbarContent)
        └─> Task 4 (makeFilesContent)

PHASE 3: Cleanup (depends on Phase 2)
    Task 5 (remove obsolete functions)

PHASE 4: Verification
    Task 8 (test and verify)
```

**Execution Order:**
1. Task 6 + Task 7 (parallel - both in learn_visuals.go)
2. Task 1 (topic structure in learn_tab.go)
3. Tasks 2, 3, 4 (content functions - can be parallel)
4. Task 5 (cleanup obsolete functions)
5. Task 8 (test)

---

## Commit Strategy

### Commit 1: Update Diagram API and Sizes (learn_visuals.go)
- Task 7: OpenLaneFlowDiagram API change (remove bool parameter, always show highlights)
- Task 6: All diagram size increases
- **Rationale:** API must change before content functions can compile

### Commit 2: Refactor Learn Tab Structure (learn_tab.go)
- Task 1: Update topic structure (6 -> 3 topics)
- Task 2: Create makeIntroContent (includes preserved "Stages Explained" content)
- Task 3: Create makeCrossbarContent (diagrams stacked vertically)
- Task 4: Create makeFilesContent
- Task 5: Remove obsolete functions

### Commit 3: Final Polish (if needed)
- Any adjustments from testing
- Minor text/layout tweaks

---

## Success Criteria

| Criteria | Verification |
|----------|--------------|
| 3 topics only | Count items in sidebar |
| Larger diagrams | Visual inspection (at least 25% larger) |
| No duplicates | Flow diagram appears once |
| Single disclaimer | Search for "DISCLAIMER" - one instance |
| Clean formatting | No "----" or "====" in text |
| App runs | `go build && ./fecim-visualizer` succeeds |
| All tests pass | `go test ./module6-eda/...` |

---

## Estimated Effort

| Task | Complexity | Time Estimate |
|------|------------|---------------|
| Task 1 | Low | 5 min |
| Task 2 | Medium | 15 min |
| Task 3 | Medium | 10 min |
| Task 4 | Medium | 10 min |
| Task 5 | Low | 5 min |
| Task 6 | Medium | 15 min |
| Task 7 | Low | 5 min |
| Task 8 | Low | 10 min |
| **Total** | | **~75 min** |

---

## Files Modified

1. `<local-path>`
   - Topic list reduction
   - New content functions
   - Removed obsolete functions

2. `<local-path>`
   - Diagram size increases
   - OpenLaneFlowDiagram API change
