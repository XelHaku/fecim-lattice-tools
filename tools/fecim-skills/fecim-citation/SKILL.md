---
name: fecim-citation
description: Verifies and formats FeCIM physics/measurement claims against the project's published-source list and docs/4-research/honesty-audit.md. Use when adding a numeric claim, accuracy figure, or device-parameter assertion to code, docs, PR descriptions, or commit messages.
---

# fecim-citation

Validate any quantitative or device-physics claim before it lands in code, docs, or commit messages. See `tools/fecim-skills/_shared/fecim-context.md` (Honesty-audit policy) for the rule list.

## Workflow

1. **Parse the claim.** Extract: subject (e.g., HZO coercive field), value (e.g., 1.0 MV/cm), context (educational vs validated).

2. **Match to canonical sources** (`tools/fecim-skills/_shared/fecim-context.md` table) and the current source/audit files (`docs/4-research/honesty-audit.md`, `references/`, and `citations/` if present). If the claim corresponds to a published source, format as:
   ```
   <claim text> (Materlik 2015)
   ```

3. **Check honesty-audit removed/unverified list:**
   - "30 analog states" presented as device fact → REPHRASE: "30 analog conductance levels (configurable simulation default)".
   - "87% MNIST accuracy" as a FeCIM device claim → REMOVE; if discussing reservoir computing instead, attribute to HZO FTJ 2025 paper at 98.24% and label "not a FeCIM device claim".
   - Energy multipliers vs NAND/GPUs without measurement evidence → REMOVE or replace with literature-backed comparison.

4. **For uncited claims**, suggest one of:
   - Add the citation if a published source exists.
   - Reframe as "simulation default" with the project's standard wording.
   - Block the claim if no source and no educational framing applies.

5. **Output:**
   ```
   Claim: <as written>
   Status: verified | educational-default | removed-unverified | needs-source
   Suggested wording: <if change needed>
   Citation: <short form, DOI or source path if applicable>
   Evidence boundary: <what the source verifies and what it does not verify>
   ```

## Verification

- Input: "Add 'Our simulator achieves 87% MNIST accuracy' to README.md."
  Expected: status `removed-unverified`; suggested wording reframes as the simulation pipeline's accuracy with explicit educational label, OR points to HZO FTJ 2025's 98.24% with "not a FeCIM device claim" caveat.

## TDD

Citation review is observation — `TDD: N/A`. Code/doc edits triggered by review follow the project's TDD hard-rule per `tools/fecim-skills/_shared/tdd-evidence-template.md`.
