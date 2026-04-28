---
name: fecim-honesty-audit
description: Enforces docs/4-research/honesty-audit.md policy by scanning PR diffs, READMEs, and presentation material for removed/unverified claims (87% MNIST, 30-states-as-fact, energy multipliers vs NAND/GPUs). Use before committing docs, PRs, or release notes that include accuracy or efficiency numbers.
---

# fecim-honesty-audit

Scan a diff, README, or PR description for the project's removed/unverified claim list before it lands. See `tools/fecim-skills/_shared/fecim-context.md` (Honesty-audit policy).

## Workflow

1. **Take input.** Either:
   - PR diff: `git diff main...HEAD -- '*.md' README.md`
   - Specific file(s).
   - Free-text the user pastes.

2. **Regex-scan for trigger phrases (case-insensitive):**
   - `\b87%?\s*MNIST\b`
   - `\b30\s*(analog\s+)?(states|levels)\b` (followed by `device|hardware|fact` → red flag; with `simulation default|configurable` → ok)
   - `vs\.?\s*(NAND|GPUs?|DRAM)` near energy/power numbers
   - `\bX×\s*(less|more)\s*(energy|power)\b`
   - Numeric percentages near `MNIST|CIFAR|accuracy|efficiency`

3. **Classify each hit** as one of:
   - `verified` — has a published citation in the project's list.
   - `educational-default` — properly labeled as a simulation default.
   - `removed-unverified` — matches the audit's removal list; must change.

4. **For `removed-unverified`, suggest the approved rephrasing:**
   - "30 analog states" → "30 analog conductance levels (configurable simulation default)"
   - "87% MNIST accuracy" → remove, OR attribute the 98.24% HZO FTJ 2025 figure with "not a FeCIM device claim".
   - Energy multipliers without source → remove or replace with literature-backed comparison.

5. **Output a structured report:**
   ```
   File:line: <claim text>
   Status: verified | educational-default | removed-unverified
   Suggested change: <wording or DELETE>
   ```

   Exit `0` if all clean, exit `1` (or print `CHANGES REQUESTED`) if any `removed-unverified`.

## Verification

- Input: a README diff with `our chip achieves 87% MNIST accuracy at 1000× lower energy than GPUs`.
  Expected: 2 hits flagged, both `removed-unverified`; suggests removal or reframing.

## TDD

Audit is observation — `TDD: N/A`. Wording changes are documentation-only and qualify under CLAUDE.md's `TDD: N/A` carve-out.
