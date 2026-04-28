---
name: fecim-labtester
description: Runs the FeCIM test matrix (full, race, module-scoped, coverage, golden regen) and interprets physics regression failures using the 5 known bug patterns. Use when running tests, debugging test failures, or regenerating physics golden files.
---

# fecim-labtester

Pick the right test invocation, run it, and triage failures against the 5 known bug patterns. See `tools/fecim-skills/_shared/fecim-context.md` for the test matrix and bug list.

## Workflow

1. **Pick scope** by the change being verified:
   - Whole suite: `go test ./...`
   - Race detection: `go test -race ./...`
   - Module: `go test ./module1-hysteresis/...` (or `make test-hys`, `test-xbar`, `test-mnist`, `test-circuits`, `test-shared`)
   - Future shell: `make test-next-ui`
   - Coverage: `go test -cover ./...`

2. **For physics-regression failures**, classify against the 5 known patterns:
   1. Guard-band sign flip
   2. Bounds collapse `[VMin, VMax]`
   3. ACCEPT ±1 guard interaction
   4. Zero-field bounds reset
   5. Preisach Everett zero-clamp

3. **Golden regeneration** is allowed only when divergence is intentional and the user has confirmed:
   ```bash
   FECIM_UPDATE_PHYSICS_GOLDEN=1 go test ./...
   ```
   Diff `validation/testdata/physics_regression/` before committing — unintended changes are bugs.

4. **Output the TDD evidence block** per `tools/fecim-skills/_shared/tdd-evidence-template.md`:
   ```
   RED:  go test ./module1-hysteresis/... -run TestX
         FAIL TestX (Preisach Everett zero-clamp pattern)
   GREEN: go test ./module1-hysteresis/... -run TestX
          ok
   VERIFY: go test ./... && go test -race ./...
   ```

## Verification

- Input: "TestPreisachEverett is failing in module1-hysteresis."
  Expected: maps to known bug pattern #5; suggests product-form Everett vs factorized; runs the targeted test, then full suite, then race.

## TDD

This skill is the TDD verifier itself — every change made under it must produce the RED/GREEN block above before commit.
