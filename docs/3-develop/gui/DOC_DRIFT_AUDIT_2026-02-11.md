# GUI Documentation Drift Audit — 2026-02-11

Scope: TODO G15 + G16 (docs drift mapping under the current `docs/3-develop/gui/` tree, with historical `docs/development/GUI/` references called out where they still appear), aligned with resize audit/fixes from G14.

## Code vs Docs Checks Performed

### 1) Module 4 (Circuits) — `tab_comparison.go`
- **Code reality (2026-02-11):** comparison dashboard and table now use explicit scroll guards to prevent overlap/clipping during window shrink.
- **Doc update:** `GUI.module4.md`
  - Updated `Last Updated` to `2026-02-11`
  - Added `Recent Changes (2026-02-11)` section documenting:
    - nested `VScroll(HScroll(...))` wrapper on comparison body
    - scroll wrapping in `createCompTableSection()`
    - new regression test `tab_comparison_resize_test.go`

### 2) Module 6 (EDA) — Learn + Builder/Validation
- **Code reality (2026-02-11):**
  - Learn tab content scroll min size reduced (`750x500` -> `360x260`) for narrow-window compatibility.
  - Validation summary grid now wrapped with `HScroll` to prevent truncation/overlap.
- **Doc update:** `GUI.module6.md`
  - Updated `Last Updated` to `2026-02-11`
  - Added `Recent Changes (2026-02-11)` block with both layout changes and test reference.

### 3) GUI docs index
- **Code reality:** docs sweep completed and now tracked as dated artifact.
- **Doc update:** `README.md`
  - Updated `Last Updated` to `2026-02-11`
  - Added pointer to this audit file in `Document Version`.

## New/Updated Evidence Artifacts

### New tests
- `module4-circuits/pkg/gui/tab_comparison_resize_test.go`
- `module6-eda/pkg/gui/tabs/learn_tab_resize_test.go`

### Test commands executed
- `go test ./module4-circuits/pkg/gui -run TestComparisonTab_HasScrollGuardsForResize`
- `go test ./module6-eda/pkg/gui/tabs -run TestMakeLearnTab_ContentScrollUsesCompactMinSize -v`

Both tests passed after code updates.

## Remaining Known Drift (outside this TODO scope)
- This sweep focused on resize/layout behavior and immediate docs impacted by those changes.
- Legacy narrative sections in older module docs may still contain historical recommendations that no longer match current implementation details; these require separate per-module deep cleanup tasks.
