# Autonomous KG-Driven Self-Improvement Prompt

> Copy-paste this prompt to any Claude Code session to trigger a full KG-driven audit, fix, and improvement cycle. It feeds the Cognee knowledge graph, queries it for issues, and swarms agents to fix everything it finds.

---

## The Prompt

```
Feed the Cognee knowledge graph with all new/changed files since the last ingest, then use the KG to systematically find and fix issues across the entire codebase. Use swarm agents for parallel execution.

## Phase 1: KG Update
1. Run `scripts/cognee-ingest-full.py` to re-ingest all source code, docs, and configs
2. Run `scripts/cognee-ingest-papers.py` to ingest academic papers, validation data, and experimental datasets
3. Wait for cognify to complete

## Phase 2: KG-Driven Issue Detection
Query the Cognee KG with these categories (run all queries, collect all findings):

### Physics & Parameters
- "What physics parameters are inconsistent between code, YAML configs, and documentation?"
- "What paper citations in code comments don't match the actual parameter values used?"
- "What experimental data exists that isn't used for validation or calibration?"
- "What physics models are referenced in documentation but not implemented?"
- "What material preset values disagree with their cited literature sources?"

### Code Quality
- "What functions or types are duplicated across modules?"
- "What cross-module imports violate the shared/ boundary rule?"
- "What dead code or unused exports exist?"
- "What magic numbers in tests should use named constants?"
- "What goroutines are spawned without proper shutdown mechanisms?"

### Robustness & Safety
- "What mathematical operations could produce NaN, Inf, or division by zero?"
- "What config constructors could panic or produce invalid state?"
- "What file I/O operations lack error handling or path sanitization?"
- "What exported functions lack input validation at system boundaries?"
- "What MVM pipeline stages silently clamp or drop errors?"

### Documentation & Ontology
- "What inconsistencies exist between code implementations and documentation claims?"
- "What naming inconsistencies exist across modules for the same concepts?"
- "What documentation references stale file paths that no longer exist?"
- "What units are inconsistent between modules?"
- "What claims lack proper evidence status caveats (Demonstrated/Modeled/Aspirational)?"

### Architecture & Tests
- "What module interfaces are inconsistent with the embedded app pattern?"
- "What test coverage gaps exist for critical physics functions?"
- "What YAML vs hardcoded fallback paths produce different runtime behavior?"
- "What keyboard shortcuts or UI defaults are inconsistent across modules?"
- "What validation tests reference nonexistent test functions?"

## Phase 3: Swarm Fix Agents
For each category of findings, launch parallel fix agents:
- Physics fixes: correct parameters to match literature, align YAML with hardcoded
- Code fixes: deduplicate, remove dead code, add validation guards
- Doc fixes: update stale paths, correct claims, add caveats
- Robustness fixes: add div-by-zero guards, constructor validation, NaN checks
- Architecture fixes: fix interface compliance, keyboard handlers, module boundaries

Rules for fix agents:
- Read files before editing
- Run `go build ./...` after changes
- Run `go test -short ./...` to verify no regressions
- Use literature citations for any physics parameter changes
- Don't modify test assertions without understanding why they exist
- Make changes opt-in when they affect simulation behavior (add config flags)

## Phase 4: Verification
1. Run full build: `go build ./...`
2. Run full test suite: `go test -short ./...`
3. If any failures, fix them before finishing
4. Re-ingest changed files into KG for next cycle
5. Summarize all changes made

## Key Files
- `.env` — Cognee LLM config (OpenRouter + fastembed)
- `scripts/cognee-setup.sh` — Bootstrap Cognee venv
- `scripts/cognee-ingest-full.py` — Full repo ingest (source + docs)
- `scripts/cognee-ingest-papers.py` — Academic papers + validation ingest
- `CLAUDE.md` — Project rules and conventions

## Cognee Config (must set env vars BEFORE import)
```python
import os
os.environ["COGNEE_SKIP_CONNECTION_TEST"] = "true"
os.environ["ENABLE_BACKEND_ACCESS_CONTROL"] = "false"
os.environ["LLM_PROVIDER"] = "openai"
os.environ["LLM_MODEL"] = "openrouter/openai/gpt-4o-mini"
os.environ["LLM_ENDPOINT"] = "https://openrouter.ai/api/v1"
os.environ["EMBEDDING_PROVIDER"] = "fastembed"
os.environ["EMBEDDING_MODEL"] = "BAAI/bge-small-en-v1.5"
os.environ["EMBEDDING_DIMENSIONS"] = "384"
```

## What Makes This Work
- The KG finds *relationships* grep can't: "this doc claims X but code implements Y"
- Cross-domain queries: physics ↔ code ↔ docs ↔ papers ↔ tests
- Swarm parallelism: 5-10 agents fixing different domains simultaneously
- Literature-backed: physics changes cite papers, not guesses
- Non-breaking: new features are opt-in, existing tests must pass
```

---

## Quick Start (one-liner)

```
Feed Cognee KG with all project files, query it for physics/code/doc/architecture issues, and fix everything found using swarm agents. Academic papers first. All tests must pass.
```
