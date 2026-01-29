# Plan: Merge EDA Documentation Folders

## Context

**Original Request:** Merge two EDA documentation folders intelligently
- `docs/eda/` - 3 files (newer, more technical, focused)
- `docs/eda-design-suite/` - 16 files (older, broader coverage, includes archive)

**Goal:** Consolidate into a single canonical `docs/eda/` folder with clear organization

---

## Analysis Summary

### docs/eda/ (Target - Keep as Canonical)
| File | Lines | Content | Decision |
|------|-------|---------|----------|
| API.md | 751 | Complete API reference for Module 6 EDA packages | KEEP as-is |
| ARCHITECTURE.md | 686 | Detailed architecture documentation | KEEP as-is |
| WORKFLOW.md | 1169 | RTL to GDSII workflow guide | KEEP as-is |

### docs/eda-design-suite/ (Source - Migrate Content)
| File | Lines | Content | Decision |
|------|-------|---------|----------|
| README.md | 270 | Module overview with disclaimers | MERGE into new README.md |
| REFERENCES.md | 246 | Scientific references | MOVE to docs/eda/references/ |
| SKY130.md | 266 | PDK integration guide | MOVE to docs/eda/pdk/ |
| eda.opensource.md | Large | Open-source EDA ecosystem | MOVE to docs/eda/ecosystem/ |
| eda.research.md | 593 | Research paper collection | MOVE to docs/eda/references/ |
| eda.eli5.md | 787 | Beginner explanation | MOVE to docs/eda/guides/ |
| eda.demo.md | 317 | Demo walkthrough | MERGE into WORKFLOW.md or MOVE to docs/eda/guides/ |
| eda.integration.md | 582 | OpenLane integration guide | MOVE to docs/eda/guides/ |
| eda.guide.zero_to_asic.md | 118 | Zero to ASIC practical field guide | MOVE to docs/eda/guides/ |
| EDA-LEARN-screen.md | ~100 | GUI screen docs | ARCHIVE (superseded by GUI code) |
| EDA-SPICE-screen.md | 349 | SPICE format explanation | MOVE to docs/eda/guides/ |
| FECIM_TO_WAFER.md | 1186 | Complete fab workflow | MOVE to docs/eda/guides/ |
| OPENLANE_STUDY.md | 393 | OpenLane source analysis | MOVE to docs/eda/references/ |
| OPENLANE_TOOLS_REFERENCE.md | 2371 | Comprehensive CLI reference | MOVE to docs/eda/references/ |
| plan-demo6.md | Large | Implementation plan | ARCHIVE (historical) |
| _archive/ | 4 files | Historical files | KEEP in archive location |

---

## Final Folder Structure

```
docs/eda/
├── README.md                    # NEW: Combined overview (from eda-design-suite/README.md + intro)
├── API.md                       # EXISTING: Keep as-is
├── ARCHITECTURE.md              # EXISTING: Keep as-is
├── WORKFLOW.md                  # EXISTING: Keep as-is
│
├── guides/                      # NEW SUBFOLDER: How-to guides
│   ├── eli5.md                  # FROM: eda.eli5.md (beginner guide)
│   ├── demo.md                  # FROM: eda.demo.md (demo walkthrough)
│   ├── integration.md           # FROM: eda.integration.md (OpenLane integration)
│   ├── zero-to-asic.md          # FROM: eda.guide.zero_to_asic.md (practical field guide)
│   ├── spice-format.md          # FROM: EDA-SPICE-screen.md
│   └── fecim-to-wafer.md        # FROM: FECIM_TO_WAFER.md (fab workflow)
│
├── references/                  # NEW SUBFOLDER: Reference materials
│   ├── scientific.md            # FROM: REFERENCES.md
│   ├── research-papers.md       # FROM: eda.research.md
│   ├── openlane-study.md        # FROM: OPENLANE_STUDY.md
│   └── cli-reference.md         # FROM: OPENLANE_TOOLS_REFERENCE.md
│
├── pdk/                         # NEW SUBFOLDER: PDK-specific docs
│   └── sky130.md                # FROM: SKY130.md
│
├── ecosystem/                   # NEW SUBFOLDER: Ecosystem overview
│   └── opensource-eda.md        # FROM: eda.opensource.md
│
└── _archive/                    # ARCHIVE: Historical/superseded
    ├── plan-demo6.md            # FROM: plan-demo6.md (implementation plan)
    ├── eda-learn-screen.md      # FROM: EDA-LEARN-screen.md
    └── legacy/                  # FROM: eda-design-suite/_archive/
        ├── openlane-validation-report.md
        ├── architecture-1t1r-review.md
        ├── cell-geometry-decision.md
        └── module6-technical-plan.md
```

---

## Detailed Task List

### Task 1: Create New Subfolder Structure
**Files to create:**
- `docs/eda/guides/` (directory)
- `docs/eda/references/` (directory)
- `docs/eda/pdk/` (directory)
- `docs/eda/ecosystem/` (directory)
- `docs/eda/_archive/` (directory)
- `docs/eda/_archive/legacy/` (directory)

**Acceptance Criteria:** All directories exist

---

### Task 2: Create New README.md for docs/eda/
**Action:** Create a new comprehensive README.md that:
1. Includes the disclaimers from eda-design-suite/README.md
2. Provides navigation to all documentation
3. Links to the subfolder structure

**Source Content:**
- Overview from `eda-design-suite/README.md` (lines 1-100)
- Capabilities table
- Critical disclaimers section
- Add navigation links to new structure

**Acceptance Criteria:**
- README.md exists in docs/eda/
- Contains disclaimers about placeholder timing values
- Contains clear navigation to all subfolders

---

### Task 3: Move Guide Documents
**Actions:**
| Source | Destination | Rename |
|--------|-------------|--------|
| eda-design-suite/eda.eli5.md | docs/eda/guides/eli5.md | Yes |
| eda-design-suite/eda.demo.md | docs/eda/guides/demo.md | Yes |
| eda-design-suite/eda.integration.md | docs/eda/guides/integration.md | Yes |
| eda-design-suite/eda.guide.zero_to_asic.md | docs/eda/guides/zero-to-asic.md | Yes |
| eda-design-suite/EDA-SPICE-screen.md | docs/eda/guides/spice-format.md | Yes |
| eda-design-suite/FECIM_TO_WAFER.md | docs/eda/guides/fecim-to-wafer.md | Yes |

**Acceptance Criteria:**
- All 6 files exist in docs/eda/guides/
- Original files removed from eda-design-suite/

---

### Task 4: Move Reference Documents
**Actions:**
| Source | Destination | Rename |
|--------|-------------|--------|
| eda-design-suite/REFERENCES.md | docs/eda/references/scientific.md | Yes |
| eda-design-suite/eda.research.md | docs/eda/references/research-papers.md | Yes |
| eda-design-suite/OPENLANE_STUDY.md | docs/eda/references/openlane-study.md | Yes |
| eda-design-suite/OPENLANE_TOOLS_REFERENCE.md | docs/eda/references/cli-reference.md | Yes |

**Acceptance Criteria:**
- All 4 files exist in docs/eda/references/
- Original files removed from eda-design-suite/

---

### Task 5: Move PDK Document
**Action:**
| Source | Destination | Rename |
|--------|-------------|--------|
| eda-design-suite/SKY130.md | docs/eda/pdk/sky130.md | Yes |

**Acceptance Criteria:**
- File exists in docs/eda/pdk/
- Original file removed from eda-design-suite/

---

### Task 6: Move Ecosystem Document
**Action:**
| Source | Destination | Rename |
|--------|-------------|--------|
| eda-design-suite/eda.opensource.md | docs/eda/ecosystem/opensource-eda.md | Yes |

**Acceptance Criteria:**
- File exists in docs/eda/ecosystem/
- Original file removed from eda-design-suite/

---

### Task 7: Archive Historical Documents
**Actions:**
| Source | Destination |
|--------|-------------|
| eda-design-suite/plan-demo6.md | docs/eda/_archive/plan-demo6.md |
| eda-design-suite/EDA-LEARN-screen.md | docs/eda/_archive/eda-learn-screen.md |
| eda-design-suite/_archive/* | docs/eda/_archive/legacy/ |

**Acceptance Criteria:**
- All archive files exist in docs/eda/_archive/
- Legacy subfolder contains 4 historical files

---

### Task 8: Delete Empty eda-design-suite Folder
**Action:** After all files are migrated, delete the now-empty `docs/eda-design-suite/` folder

**Pre-condition:** All 16 files have been moved or archived (verify with `ls`)
**Acceptance Criteria:** `docs/eda-design-suite/` no longer exists

---

### Task 9: Update Internal Links (COMPREHENSIVE)
**Action:** Search and update any internal links that reference `eda-design-suite/` paths.

**Files to Update (with specific changes):**

| File | Line(s) | Current Link | New Link |
|------|---------|--------------|----------|
| docs/README.md | 20 | `[eda-design-suite/](eda-design-suite/)` | `[eda/](eda/)` |
| docs/README.md | 20 | `[ELI5](eda-design-suite/eda.eli5.md)` | `[ELI5](eda/guides/eli5.md)` |
| docs/opensource-tools/circuit-simulation-tools.md | 1135 | `[EDA-LEARN-screen.md](../eda-design-suite/EDA-LEARN-screen.md)` | `[eda-learn-screen.md](../eda/_archive/eda-learn-screen.md)` |
| docs/opensource-tools/eda-tools.md | 1532 | `[SKY130.md](../eda-design-suite/SKY130.md)` | `[sky130.md](../eda/pdk/sky130.md)` |
| docs/opensource-tools/eda-tools.md | 1533 | `[OPENLANE_STUDY.md](../eda-design-suite/OPENLANE_STUDY.md)` | `[openlane-study.md](../eda/references/openlane-study.md)` |
| docs/opensource-tools/eda-tools.md | 1534 | `[FECIM_TO_WAFER.md](../eda-design-suite/FECIM_TO_WAFER.md)` | `[fecim-to-wafer.md](../eda/guides/fecim-to-wafer.md)` |

**Additional Files to Check:**
- All files in docs/eda/guides/ (update relative paths within moved files)
- All files in docs/eda/references/ (update relative paths within moved files)
- docs/development/scriptReference.md (mentioned in CLAUDE.md)

**Verification Command:**
```bash
grep -r "eda-design-suite" docs/ --include="*.md"
```

**Acceptance Criteria:**
- No broken internal links
- All relative paths updated to new locations
- `grep -r "eda-design-suite" docs/` returns no results (except _archive files if they contain historical references)

---

### Task 10: Update CLAUDE.md Table
**Action:** Update the lookup table in CLAUDE.md to reflect new documentation locations

**Current:**
```markdown
| Fix Fyne GUI issues | `docs/development/GUI/FYNE_NOTES.md` |
```

**Add:**
```markdown
| EDA documentation | `docs/eda/README.md` |
| OpenLane integration | `docs/eda/guides/integration.md` |
| EDA CLI reference | `docs/eda/references/cli-reference.md` |
```

**Acceptance Criteria:** CLAUDE.md points to correct EDA documentation locations

---

## Commit Strategy

### Commit 1: Create folder structure and new README
```
docs(eda): create organized subfolder structure for EDA docs

- Add guides/, references/, pdk/, ecosystem/, _archive/ subfolders
- Create comprehensive README.md with navigation and disclaimers
```

### Commit 2: Migrate guide documents
```
docs(eda): migrate how-to guides from eda-design-suite

Move and rename:
- eda.eli5.md -> guides/eli5.md
- eda.demo.md -> guides/demo.md
- eda.integration.md -> guides/integration.md
- eda.guide.zero_to_asic.md -> guides/zero-to-asic.md
- EDA-SPICE-screen.md -> guides/spice-format.md
- FECIM_TO_WAFER.md -> guides/fecim-to-wafer.md
```

### Commit 3: Migrate reference documents
```
docs(eda): migrate reference materials from eda-design-suite

Move and rename:
- REFERENCES.md -> references/scientific.md
- eda.research.md -> references/research-papers.md
- OPENLANE_STUDY.md -> references/openlane-study.md
- OPENLANE_TOOLS_REFERENCE.md -> references/cli-reference.md
```

### Commit 4: Migrate PDK and ecosystem docs
```
docs(eda): migrate PDK and ecosystem documentation

Move and rename:
- SKY130.md -> pdk/sky130.md
- eda.opensource.md -> ecosystem/opensource-eda.md
```

### Commit 5: Archive historical documents
```
docs(eda): archive superseded documentation

Move to _archive/:
- plan-demo6.md (implementation plan)
- EDA-LEARN-screen.md (superseded by code)
- Legacy archive files from eda-design-suite/_archive/
```

### Commit 6: Remove eda-design-suite and update links
```
docs(eda): complete folder merge, remove eda-design-suite

- Delete empty docs/eda-design-suite/ folder
- Update internal documentation links in:
  - docs/README.md (line 20)
  - docs/opensource-tools/circuit-simulation-tools.md (line 1135)
  - docs/opensource-tools/eda-tools.md (lines 1532-1534)
- Update CLAUDE.md references
```

---

## Success Criteria

1. Single canonical EDA documentation folder at `docs/eda/`
2. Clear organizational structure with subfolders by purpose
3. No duplicate content
4. All 16 source files from eda-design-suite/ preserved (moved or archived)
5. Historical content archived (not deleted)
6. All internal links working (verified with grep)
7. CLAUDE.md updated with correct paths
8. `docs/eda-design-suite/` folder completely removed

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Broken links in other files | Comprehensive link audit in Task 9 with specific file:line references |
| Lost content | Archive rather than delete any uncertain files |
| Missing files during migration | Verify 16-file count before and after each commit |
| Merge conflicts | Small, focused commits that can be reviewed independently |

---

## Notes

- The three existing files in `docs/eda/` (API.md, ARCHITECTURE.md, WORKFLOW.md) are newer and more comprehensive - keep them as-is
- The `eda-design-suite/README.md` has valuable disclaimers that should be preserved in the new README.md
- Some content overlap exists (e.g., OpenLane integration in both WORKFLOW.md and eda.integration.md) - keep both as they serve different purposes (quick reference vs detailed guide)
- `eda.guide.zero_to_asic.md` is a valuable practical guide derived from Matt Venn's work - move to guides/ subfolder
