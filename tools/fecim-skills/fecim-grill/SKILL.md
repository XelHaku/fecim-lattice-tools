---
name: fecim-grill
description: Relentlessly interviews the user about a proposed FeCIM physics, simulation, or GUI change before any code is written — covering source citation, educational-vs-validated framing, TDD-RED test, thread-safety, and honesty-audit alignment. Use when starting a non-trivial change to physics, ISPP, crossbar, or GUI logic.
---

# fecim-grill

Domain-specific design grill. Differs from generic grill skills: every branch ties back to FeCIM's TDD hard-rule, honesty-audit, and Fyne thread-safety. See `tools/fecim-skills/_shared/fecim-context.md`.

## Workflow

Use this only when Juan explicitly wants design interrogation or when a non-trivial change is ambiguous enough that coding would risk unsupported physics. Otherwise prefer autonomous execution with documented assumptions. Ask each question one at a time. Do not move on until the user answers. Mark each branch RESOLVED before exit.

1. **What changes, behaviorally?** One sentence. If the answer references a UI element, ask which goroutine/main thread.

2. **Source citation.** Is this grounded in:
   - A published source (Materlik 2015, Park 2015, Alessandri 2018, Guo 2018, HZO FTJ 2025)?
   - An educational simulation default (clearly labeled)?
   - Neither (BLOCK — must reframe before code).

3. **Educational vs validated framing.** If the change touches accuracy/efficiency numbers, run through the honesty-audit removed/unverified list (87% MNIST, 30-states-as-fact, NAND/GPU energy multipliers).

4. **TDD-RED test.** What is the focused failing test that proves the new behavior? Path + name. If the user can't name one, BLOCK until they can.

5. **Thread-safety (Fyne).** If the change runs on a goroutine and touches `*widget.*`, `*canvas.*`, `*container.*`, confirm `fyne.Do(func() { ... })` will wrap the mutation.

6. **UI boundary.** If the change touches `viewmodel/`, confirm zero `fyne.io/...` and `github.com/gogpu/ui` imports added.

7. **Output a one-paragraph design summary** the user signs off before any code is written:
   ```
   Behavior: ...
   Source: ...
   Framing: validated | educational
   RED test: <file:test>
   Thread-safety: covered | n/a
   UI-boundary: ok | n/a
   ```

## Verification

- Input: "Add ISPP guard-band relaxation that allows 4 guard pulses instead of 2."
  Expected: walks all 7 questions; surfaces bug pattern #1 (guard-band sign flip); requires RED test path; produces a sign-off summary.

## TDD

This skill exists to enforce TDD before code is written. It cannot violate TDD itself; output is observation — `TDD: N/A`.
