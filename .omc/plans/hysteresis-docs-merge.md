# Work Plan: Hysteresis Documentation Merge

**Plan ID:** hysteresis-docs-merge
**Created:** 2026-01-24
**Status:** Ready for Review

---

## Context

### Original Request
Move and merge documentation from `/module1-hysteresis/` into `/docs/hysteresis/`, consolidating overlapping content while preserving unique value from both locations.

### Source Files Analysis

| File | Lines | Unique Value |
|------|-------|--------------|
| `module1-hysteresis/PHYSICS.md` | 549 | **Deep physics**: Atoms/charges basics, dipole concept, hysteron definition, Preisach plane math, code implementation details, verified simulation behavior, τ switching time analysis |
| `module1-hysteresis/ELI5.module1.md` | 331 | **Demo-focused ELI5**: Magic light switch analogy, stubborn magnets analogy, "loop EMERGES" explanation, waveform mode instructions |
| `module1-hysteresis/module1.hysteresis.README.md` | 253 | **Demo documentation**: ASCII UI mockup, controls list, architecture diagram, quick start commands, troubleshooting |

### Target Files Analysis

| File | Lines | Unique Value |
|------|-------|--------------|
| `docs/hysteresis/hysteresis.ELI5.md` | 599 | **Extended ELI5**: Rubber band analogy, parking garage analogy, "perfect module" specification (Parts 7.1-7.11), quick reference card, gap analysis, development effort estimate |
| `docs/hysteresis/hysteresis.opensource.md` | 663 | **Tools catalog**: PyPreisach, FerroX, FERRET, ngspice+OpenVAF, NeuroSim, CrossSim, AIHWKIT, tool comparison matrix, custom implementation tutorial |
| `docs/hysteresis/hysteresis.research.md` | 480 | **Research meta-study**: 50+ paper corpus, Preisach formulation, KAI model, HZO material physics, Landau-Devonshire theory, bibliography |

---

## Content Overlap Analysis

### ELI5 Content (Significant Overlap)

**Overlapping Concepts:**
- Light switch analogy (source) vs. dimmer switch (target) - similar
- Stubborn magnets/hysterons explanation - both files
- 30 levels explanation - both files
- P-E loop walkthrough - both files
- Write vs Read threshold - both files

**Unique to Source (ELI5.module1.md):**
- "Loop EMERGES because each hysteron flips at different voltages" with step-by-step (1-8) explanation
- Specific demo waveform mode instructions (Sine, Random Walk, Write/Read Demo, Manual, Frequency slider)
- "Summary for Kids" table
- "Technical Note: What's Actually Running" section

**Unique to Target (hysteresis.ELI5.md):**
- Rubber band analogy (Part 1)
- Hiking analogy for loop traversal
- "Perfect Module" specification (Parts 7.1-7.11) - 250+ lines
- Quick Reference Card
- Gap Analysis and Development Effort tables
- Learning Resources section (beginner/intermediate/advanced paths)
- Glossary (Part 8)

### Physics Content

**PHYSICS.md is largely unique** - the target files don't have:
- Atoms and charges basics (Part 1)
- Electric field units explanation
- Detailed code snippets showing exact implementation
- τ (switching time) analysis
- "What's Real vs. Simplified" table
- Demo waveform modes table at end

### README Content

**README is module-specific** but overlaps with:
- Target ELI5 Part 7 (features/requirements)
- Target opensource.md (architecture diagram)

---

## Work Objectives

### Core Objective
Consolidate all hysteresis documentation into `docs/hysteresis/` with clear organization, eliminating redundancy while preserving all unique content.

### Deliverables

1. **`docs/hysteresis/hysteresis.physics.md`** (NEW)
   - Renamed and relocated PHYSICS.md
   - Standalone deep physics reference

2. **`docs/hysteresis/hysteresis.ELI5.md`** (UPDATED)
   - Merge unique content from ELI5.module1.md
   - Add "Demo Instructions" section
   - Add "Technical Note" section
   - Preserve all existing content

3. **`docs/hysteresis/hysteresis.demo.md`** (NEW)
   - Extracted from module1.hysteresis.README.md
   - Updated paths for unified app context
   - Module-specific demo documentation

4. **Source files** (DELETED after verification)
   - module1-hysteresis/PHYSICS.md
   - module1-hysteresis/ELI5.module1.md
   - module1-hysteresis/module1.hysteresis.README.md

5. **Cross-references** (UPDATED)
   - Update any internal doc links
   - Update CLAUDE.md if it references these files

### Definition of Done

- [ ] All unique content from source files preserved in target
- [ ] No redundant content in merged files
- [ ] All internal cross-references updated
- [ ] Source files deleted
- [ ] `docs/hysteresis/` has clear, logical organization
- [ ] No broken links in documentation

---

## Guardrails

### Must Have
- Preserve ALL unique physics content from PHYSICS.md
- Preserve "Perfect Module" specification from target ELI5
- Preserve demo-specific instructions for running module1
- Clear file naming convention (hysteresis.*.md)
- Cross-reference between related docs

### Must NOT Have
- Duplicate explanations of the same concept
- Orphaned files in module1-hysteresis/
- Broken internal links
- Loss of any technical accuracy

---

## Task Flow

```
[1] Create hysteresis.physics.md ─────────────────────────────┐
                                                              │
[2] Merge ELI5 content ───────────────────────────────────────┼──> [5] Update cross-refs
                                                              │
[3] Create hysteresis.demo.md ────────────────────────────────┤
                                                              │
[4] Verify all content preserved ─────────────────────────────┘
                                                              │
                                                              v
                                                    [6] Delete source files
                                                              │
                                                              v
                                                    [7] Final verification
```

---

## Detailed TODOs

### TODO 1: Create hysteresis.physics.md
**File:** `<local-path>`

**Action:** Move PHYSICS.md to target location with minor updates

**Steps:**
1. Copy content from `module1-hysteresis/PHYSICS.md`
2. Update header to match naming convention
3. Update any relative paths (e.g., `pkg/ferroelectric/preisach_advanced.go` -> `module1-hysteresis/pkg/ferroelectric/preisach_advanced.go`)
4. Add cross-reference footer pointing to ELI5 and research docs

**Acceptance Criteria:**
- [ ] File exists at `docs/hysteresis/hysteresis.physics.md`
- [ ] All 549 lines of content preserved
- [ ] Paths updated for new location
- [ ] Cross-references added

---

### TODO 2: Merge ELI5 Content
**File:** `<local-path>`

**Action:** Integrate unique sections from ELI5.module1.md

**Content to Add (from source):**

1. **After Part 5 (Hysteron Concept), add new section:**
   ```markdown
   ## Part 5.5: Why the Loop EMERGES (Step by Step)

   [Insert the 8-step "Loop EMERGES because..." explanation from ELI5.module1.md lines 86-133]
   ```

2. **Add new Part 11: Demo Instructions**
   ```markdown
   ## Part 11: Running the Demo

   [Insert "Try It Yourself!" section from ELI5.module1.md lines 251-287]
   [Include all 5 numbered instructions: Sine Wave, Random Walk, Write/Read Demo, Manual, Frequency]
   ```

3. **Add to Part 10 Summary:**
   ```markdown
   ### Technical Note: What's Actually Running

   [Insert table from ELI5.module1.md lines 310-319]
   ```

4. **Add "Summary for Kids" table** (optional, fits ELI5 spirit)
   ```markdown
   ### Summary for Kids

   [Insert table from ELI5.module1.md lines 290-298]
   ```

**Acceptance Criteria:**
- [ ] "Loop EMERGES" step-by-step explanation added
- [ ] Demo instructions section added
- [ ] Technical note table added
- [ ] No duplicate content (check for existing similar sections)
- [ ] Part numbers renumbered if needed

---

### TODO 3: Create hysteresis.demo.md
**File:** `<local-path>`

**Action:** Create demo-specific documentation from README

**Content Structure:**
```markdown
# Hysteresis Demo Documentation

## Overview
[From README lines 1-24]

## Quick Start
[Updated for unified app context]
```bash
# From project root
./launch.sh
# Or build and run
go build -o fecim-visualizer ./cmd/fecim-visualizer && ./fecim-visualizer
# Then select "Hysteresis" tab
```

## UI Layout
[ASCII mockup from README lines 44-73]

## Controls Reference
[Features/controls tables from README lines 79-108]

## Physics Model Summary
[Condensed from README lines 111-167, with link to hysteresis.physics.md for details]

## Architecture
[From README lines 170-186]

## Troubleshooting
[From README lines 224-237]

## References
[From README lines 240-253]
```

**Acceptance Criteria:**
- [ ] Quick start updated for unified app
- [ ] ASCII UI mockup preserved
- [ ] Controls reference complete
- [ ] Links to physics.md for detailed physics
- [ ] Troubleshooting section included

---

### TODO 4: Verify Content Preservation
**Action:** Diff check to ensure no content loss

**Steps:**
1. List all unique sections from source files
2. Verify each exists in target files
3. Create checklist

**Content Checklist:**

From PHYSICS.md:
- [ ] Part 1: Atoms and Charges
- [ ] Part 2: What Makes Ferroelectrics Special
- [ ] Part 3: Hysteresis - The Loop
- [ ] Part 4: Why 30 States
- [ ] Part 5: Hysteron Concept
- [ ] Part 6: Write vs Read Operations
- [ ] Part 7: Minor Loops
- [ ] Part 8: Implementation Details (code snippets)

From ELI5.module1.md:
- [ ] Magic light switch analogy
- [ ] Stubborn magnets explanation
- [ ] 8-step "loop emerges" explanation
- [ ] Demo mode instructions (5 modes)
- [ ] "Summary for Kids" table
- [ ] "Technical Note" table

From README:
- [ ] ASCII UI mockup
- [ ] Waveform modes table
- [ ] GUI controls list
- [ ] Visual indicators list
- [ ] Physics model summary
- [ ] Architecture diagram
- [ ] Troubleshooting section

**Acceptance Criteria:**
- [ ] All items checked off

---

### TODO 5: Update Cross-References
**Action:** Update internal documentation links

**Files to Check:**
1. `docs/hysteresis/hysteresis.ELI5.md` - line 598 references other docs
2. `docs/hysteresis/hysteresis.opensource.md` - lines 598-601 reference module paths
3. `docs/hysteresis/hysteresis.research.md` - lines 302-314 reference module files
4. `CLAUDE.md` - check for any hysteresis doc references

**Specific Updates:**

In `hysteresis.opensource.md`:
- Line 125-145: Update `module1-hysteresis/pkg/ferroelectric/` reference to be absolute
- Line 598: Update `/module1-hysteresis/PHYSICS.md` -> `hysteresis.physics.md`
- Line 599: Update `/module1-hysteresis/ELI5.module1.md` -> `(this file or remove)`

In `hysteresis.research.md`:
- Lines 302-314: Update file references to new locations

In `hysteresis.ELI5.md`:
- Line 512-514: Update paths to new structure

**Acceptance Criteria:**
- [ ] All internal links updated
- [ ] No references to deleted source files

---

### TODO 6: Delete Source Files
**Action:** Remove original files after verification

**Files to Delete:**
1. `<local-path>`
2. `<local-path>`
3. `<local-path>`

**Pre-deletion Verification:**
- Run diff between old and new to confirm nothing lost
- Ensure all cross-references updated

**Acceptance Criteria:**
- [ ] PHYSICS.md deleted
- [ ] ELI5.module1.md deleted
- [ ] module1.hysteresis.README.md deleted
- [ ] No broken links result

---

### TODO 7: Final Verification
**Action:** Validate complete documentation structure

**Checks:**
1. List files in `docs/hysteresis/`:
   - hysteresis.ELI5.md (updated)
   - hysteresis.opensource.md (unchanged)
   - hysteresis.research.md (updated refs)
   - hysteresis.physics.md (NEW)
   - hysteresis.demo.md (NEW)

2. List files in `module1-hysteresis/`:
   - Should have NO .md documentation files (except potentially a minimal README pointer)

3. Test all cross-reference links

**Acceptance Criteria:**
- [ ] 5 files in docs/hysteresis/
- [ ] 0 documentation files in module1-hysteresis/
- [ ] All links work
- [ ] No duplicate content

---

## Commit Strategy

### Commit 1: Add new documentation files
```
docs: add hysteresis.physics.md and hysteresis.demo.md

- Move PHYSICS.md to docs/hysteresis/hysteresis.physics.md
- Create hysteresis.demo.md from module README
- Update paths for new location
```

### Commit 2: Merge ELI5 content
```
docs: merge module ELI5 content into hysteresis.ELI5.md

- Add "Loop EMERGES" step-by-step explanation
- Add demo instructions section
- Add technical note table
- Add summary for kids
```

### Commit 3: Update cross-references
```
docs: update cross-references for hysteresis doc consolidation

- Update links in hysteresis.opensource.md
- Update links in hysteresis.research.md
- Update links in hysteresis.ELI5.md
```

### Commit 4: Remove source files
```
docs: remove redundant module1 documentation files

- Delete module1-hysteresis/PHYSICS.md (moved to docs/)
- Delete module1-hysteresis/ELI5.module1.md (merged)
- Delete module1-hysteresis/module1.hysteresis.README.md (merged)
```

---

## Success Criteria

| Criterion | Measure |
|-----------|---------|
| Content preservation | 100% of unique content from source exists in target |
| Redundancy elimination | No concept explained twice in same detail |
| Organization | Clear file naming, logical structure |
| Maintainability | Single source of truth for each topic |
| Accessibility | Cross-references between related docs |

---

## Final Documentation Structure

```
docs/hysteresis/
├── hysteresis.ELI5.md          # Beginner-friendly, includes demo instructions
├── hysteresis.physics.md       # Deep physics reference (from PHYSICS.md)
├── hysteresis.demo.md          # Demo-specific documentation
├── hysteresis.opensource.md    # Open-source tools catalog
└── hysteresis.research.md      # Research meta-study

module1-hysteresis/
├── cmd/                        # Code only
├── pkg/                        # Code only
└── (no .md files)              # Documentation moved to docs/
```

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Content loss during merge | Low | High | Detailed checklist, git history backup |
| Broken cross-references | Medium | Medium | Systematic link checking |
| Redundant content remains | Low | Low | Review merged files carefully |
| User confusion during transition | Low | Low | Clear commit messages |

---

*Plan generated by Prometheus for RALPLAN workflow*
