---
name: fecim-gogpu-migrate
description: Migrates a Fyne tab/component to the gogpu/ui zero-CGO shell via the shared/viewmodel UI-neutral bridge. Use when porting a module from cmd/fecim-lattice-tools to cmd/fecim-lattice-tools-next, or when extracting UI-coupled logic into the viewmodel layer.
---

# fecim-gogpu-migrate

Port a Fyne tab/component to the future zero-CGO `gogpu/ui` shell. The viewmodel layer is the UI-neutral bridge — see `tools/fecim-skills/_shared/fecim-context.md` (UI boundary rule).

## Workflow

1. **Identify the Fyne-coupled file.** Read `module*/pkg/gui/<name>.go`. List which types touch `fyne.io/...` and which are pure data.

2. **Extract pure state and event interface into `shared/viewmodel/<name>/`:**
   - `state.go` — typed snapshot of what the UI displays.
   - `events.go` — events the UI sends back (button presses, value changes).
   - `<name>.go` — the viewmodel: pure-Go reducer over events, returns new state.

   No `fyne.io/...` or `github.com/gogpu/ui` imports allowed in `shared/viewmodel/`. Verify:
   ```bash
   grep -r 'fyne.io\|gogpu/ui' shared/viewmodel/ && echo VIOLATION
   ```

3. **Write a `_test.go` for the viewmodel:**
   - Drive events, assert on state.
   - This is RED-first per CLAUDE.md TDD hard-rule.

4. **Reimplement the Fyne adapter** to render `state` and dispatch `events` to the viewmodel. Do not duplicate logic.

5. **Add (or stub) the gogpu/ui adapter** at `cmd/fecim-lattice-tools-next/...` rendering the same viewmodel.

6. **Verify both shells:**
   ```bash
   go test ./shared/viewmodel/... && go test ./module*/pkg/gui/... && make test-next-ui
   go build ./cmd/fecim-lattice-tools && CGO_ENABLED=0 go build ./cmd/fecim-lattice-tools-next
   git diff --check
   ```

7. **Output the TDD evidence block** per `tools/fecim-skills/_shared/tdd-evidence-template.md`.

## Verification

- Input: "Migrate module1-hysteresis/pkg/gui/simulation.go to viewmodel."
  Expected: lists Fyne types in scope; proposes `shared/viewmodel/hysteresis_simulation/` layout; writes failing viewmodel test first.

## TDD

Full TDD applies. Behavior change ≠ moving code only — the viewmodel test must demonstrate the reducer logic before the adapter is touched.
