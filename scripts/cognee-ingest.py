#!/usr/bin/env python3
"""Ingest key FeCIM docs into Cognee knowledge graph."""
import os
# Set env vars BEFORE importing cognee (lru_cache reads at import)
os.environ["COGNEE_SKIP_CONNECTION_TEST"] = "true"
os.environ["ENABLE_BACKEND_ACCESS_CONTROL"] = "false"
os.environ["LLM_PROVIDER"] = "openai"
os.environ["LLM_MODEL"] = "openrouter/openai/gpt-4o-mini"
os.environ["LLM_ENDPOINT"] = "https://openrouter.ai/api/v1"
os.environ["EMBEDDING_PROVIDER"] = "fastembed"
os.environ["EMBEDDING_MODEL"] = "BAAI/bge-small-en-v1.5"
os.environ["EMBEDDING_DIMENSIONS"] = "384"

import cognee
import asyncio
import sys
from pathlib import Path

cognee.config.set_llm_config({
    "llm_provider": "openai",
    "llm_model": "openrouter/openai/gpt-4o-mini",
    "llm_endpoint": "https://openrouter.ai/api/v1",
})

REPO = Path(__file__).resolve().parent.parent

# Key docs to ingest (non-archive, high-value)
DOCS = [
    "CLAUDE.md",
    "status.md",
    # Research
    "docs/4-research/honesty-audit.md",
    "docs/4-research/tour-group-ironlattice-research.md",
    "docs/4-research/superlattice-material-analysis.md",
    # Literature reviews
    "docs/4-research/literature-review/crossbar-circuits-literature-review-2025.md",
    "docs/4-research/literature-review/hzo-hysteresis-validation-data.md",
    "docs/4-research/literature-review/world-class-gap-analysis.md",
    # Validation
    "docs/4-research/validation/CLAIMS-MATRIX.md",
    "docs/4-research/validation/confidence-policy.md",
    "docs/4-research/validation/coverage-boundary.md",
    "docs/4-research/validation/M1-M4-physics-contract.md",
    "docs/4-research/validation/M4-M6-physics-depth-audit.md",
    "docs/4-research/validation/RESEARCH_GRADE_TESTING_STANDARDS.md",
    # Internal analysis
    "docs/4-research/internal-analysis/crossbar-arrays.md",
    "docs/4-research/internal-analysis/hysteresis-physics.md",
    "docs/4-research/internal-analysis/cim-circuits.md",
    "docs/4-research/internal-analysis/circuits.CIM-fundamentals.md",
    # Learning / ELI5
    "docs/2-learn/module1-hysteresis/eli5.md",
    "docs/2-learn/module2-crossbar/eli5.md",
    "docs/2-learn/module3-mnist/eli5.md",
    "docs/2-learn/module4-circuits/eli5.md",
    "docs/2-learn/module5-comparison/eli5.md",
    "docs/2-learn/module6-eda/eli5.md",
    # Dev guides
    "docs/3-develop/gui/FYNE_NOTES.md",
    "docs/3-develop/testing/TESTING.md",
    "docs/3-develop/api-reference.md",
]


async def main():
    added = 0
    skipped = []
    for rel in DOCS:
        p = REPO / rel
        if not p.exists():
            skipped.append(rel)
            continue
        print(f"  adding: {rel}")
        await cognee.add(str(p))
        added += 1

    if skipped:
        print(f"\n  skipped (not found): {skipped}")

    print(f"\n  {added} docs added. Running cognify ...")
    await cognee.cognify()
    print("  Done — knowledge graph built.")


if __name__ == "__main__":
    asyncio.run(main())
