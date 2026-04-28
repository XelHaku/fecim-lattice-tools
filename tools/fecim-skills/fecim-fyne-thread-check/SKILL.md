---
name: fecim-fyne-thread-check
description: Audits Go code for goroutine-to-widget access without fyne.Do(...) wrapping, the project's most common GUI freeze cause. Use when reviewing a PR that adds goroutines, async I/O, or simulation tickers in any pkg/gui/ or shell package.
---

# fecim-fyne-thread-check

Find places where a goroutine touches a Fyne widget without `fyne.Do(...)` wrapping. See `tools/fecim-skills/_shared/fecim-context.md` (UI thread-safety rule).

## Workflow

1. **Define audit scope.** Default: `module*/pkg/gui/`, `cmd/fecim-lattice-tools/`. Narrow to changed files for PR review:
   ```bash
   git diff --name-only main...HEAD | rg '\.go$' | rg 'pkg/gui|cmd/fecim'
   ```

2. **Find goroutine launches:**
   ```bash
   rg -nU 'go func\(' <scope>
   ```

3. **For each match, examine the body** for direct mutation of:
   - `*widget.*` (e.g., `Label.SetText`, `Button.SetText`, `Entry.SetText`, `ProgressBar.SetValue`)
   - `*canvas.*` (e.g., `canvas.Refresh`, `*canvas.Image.Image = ...`)
   - `*container.*` (`Add`, `Remove`, `Refresh`)
   - Direct field assignment to any `fyne.CanvasObject`

4. **Verify the call is wrapped:**
   - GOOD: `fyne.Do(func() { label.SetText("done") })`
   - GOOD: helper function whose body is itself wrapped
   - BAD: bare `label.SetText(...)` inside `go func()`

5. **Output a violation list:**
   ```
   <file>:<line>: <symbol>.<method>(...) inside goroutine — needs fyne.Do
     Suggested:
       fyne.Do(func() { <symbol>.<method>(...) })
   ```

6. **Cross-reference** `docs/3-develop/gui/FYNE_NOTES.md#threading-critical` for nuanced cases (animation tickers, blocking dialogs).

## Verification

- Input: PR adds `go func() { mylabel.SetText("done") }()` in `module1-hysteresis/pkg/gui/simulation.go`.
  Expected: skill flags `simulation.go:<line>: mylabel.SetText` and suggests the wrap.

## TDD

Audit is observation — `TDD: N/A`. Any code change made to fix a violation triggers the project's TDD hard-rule.
