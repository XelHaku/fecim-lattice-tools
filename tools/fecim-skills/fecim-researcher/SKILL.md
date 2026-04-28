---
name: fecim-researcher
description: Surveys FeCIM domain knowledge by searching references/, citations/, docs/4-research/, and the local Cognee KG, then synthesizes a cited research note. Use when investigating a physics topic, evaluating a paper, or grounding a design decision in literature.
---

# fecim-researcher

Survey FeCIM literature and project knowledge to ground a design decision or answer a physics question. See `tools/fecim-skills/_shared/fecim-context.md` for the canonical citation list.

## Workflow

1. **Define the question.** Write it in one sentence. If it has multiple sub-questions, split them and run the workflow per sub-question.

2. **Search local sources** in this order:
   - `docs/4-research/` (audits, validation notes, error propagation)
   - `references/` (academic papers, simulation benchmarks)
   - `citations/` (project's citation registry, if present)
   - `experimental-data/` (HZO, HfO2, crossbar characterization)

   Use `rg` with focused patterns (e.g., `rg -i "preisach" docs/4-research/ references/`).

3. **If Cognee is configured locally** (`.env` has `LLM_API_KEY`, `.cognee_system/` exists), query the KG:
   ```bash
   python3 - <<'PY'
   import os, asyncio
   os.environ["COGNEE_SKIP_CONNECTION_TEST"] = "true"
   os.environ["ENABLE_BACKEND_ACCESS_CONTROL"] = "false"
   import cognee
   async def main():
       results = await cognee.search("YOUR QUERY HERE")
       print(results)
   asyncio.run(main())
   PY
   ```
   Otherwise skip silently.

4. **Cite findings** using the canonical short forms from `tools/fecim-skills/_shared/fecim-context.md`. Never invent a citation; if no source supports a claim, label it "no source found, requires verification".

5. **Output a structured note**:
   ```
   Question: <...>
   Sources: <list with short citation forms>
   Finding: <2-5 sentences>
   Gaps: <what is unsupported, what should be measured>
   Recommended next step: <action>
   ```

## Verification

- Input: "What HZO Landau parameters do we use?"
  Expected: searches `module1-hysteresis/pkg/ferroelectric/material.go` and `references/`, cites Materlik 2015, returns parameters with units.

## TDD

Research output is observation — `TDD: N/A`. Any code change suggested by the research triggers the project's TDD hard-rule.
