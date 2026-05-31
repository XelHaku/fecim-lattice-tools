#!/usr/bin/env python3
"""Ingest academic papers, physics audits, validation, and experimental data into Cognee KG."""
import os
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
from pathlib import Path

cognee.config.set_llm_config({
    "llm_provider": "openai",
    "llm_model": "openrouter/openai/gpt-4o-mini",
    "llm_endpoint": "https://openrouter.ai/api/v1",
})

REPO = Path(__file__).resolve().parent.parent

# Papers and research not yet ingested
NEW_DOCS = [
    # Physics audits and models
    "docs/4-research/PHYSICS_REALISM_AUDIT.md",
    "docs/4-research/PHYSICS_REALISM_AUDIT_ADDENDUM_2026-02.md",
    "docs/4-research/physics-models.md",
    "docs/4-research/physics-validation.md",
    "docs/4-research/error-propagation.md",
    "docs/4-research/README.md",
    # Literature review addendum
    "docs/4-research/literature-review/literature-addendum-2026-02.md",
    # Paper indices
    "docs/4-research/papers/README.md",
    "docs/4-research/papers/PAPERS_INDEX.md",
    "docs/4-research/papers/ORGANIZATION.md",
    "docs/4-research/papers/NEW_PAPERS_2026-02-01.md",
    # Validation results not yet ingested
    "docs/4-research/validation/baselines/baseline-2026-02-13.md",
    "docs/4-research/validation/baselines/compatibility-matrix.md",
    "docs/4-research/validation/policies/eda-trust-boundary.md",
    "docs/4-research/validation/claims/m1-m4-claim-matrix.md",
    "docs/4-research/validation/module1/observations/resolution-2026-02-12.md",
    "docs/4-research/validation/module4/investigations/summary.md",
    "docs/4-research/validation/module4/observations/ops-resolution-2026-02-12.md",
    "docs/4-research/validation/baselines/quarterly-review-checklist.md",
    "docs/4-research/validation/reviewer/quickstart.md",
    "docs/4-research/validation/tools/verox-feasibility.md",
    "docs/4-research/validation/module4/validation-checklist.md",
    "docs/4-research/validation/module4/investigations/01-read-margin-selector-ron.md",
    "docs/4-research/validation/module4/investigations/02-wordline-rc-delay.md",
    "docs/4-research/validation/module4/investigations/03-half-select-disturb.md",
    "docs/4-research/validation/module4/investigations/04-enob-refinement.md",
    "docs/4-research/validation/module4/investigations/05-dickson-charge-pump.md",
    "docs/4-research/validation/module4/investigations/06-dynamic-topsw.md",
    "docs/4-research/validation/module4/investigations/07-spice-netlist-export.md",
    # Opensource tools analysis
    "docs/4-research/opensource-tools/circuit-analysis-libraries.md",
    "docs/4-research/opensource-tools/data-acquisition-tools.md",
    "docs/4-research/opensource-tools/walkthrough_final.md",
    "docs/4-research/opensource-tools/opensource-crossbar.md",
    # Experimental data (JSON → treat as text)
    "experimental-data/README.md",
    "experimental-data/hzo/pe-loops/park2015_advmat_hzo_10nm_fig2a.json",
    "experimental-data/hzo/pe-loops/materlik2015_jap_hfzro2_temperature_lk.json",
    "experimental-data/hzo/switching-time/jerry2017_iedm_fefet_synapse_switching.json",
    # Key validation Go source (contract tests, literature tests)
    "validation/literature.go",
    "validation/readiness_report.go",
    "validation/heracles/heracles_reference.go",
    "validation/heracles/spice_parser.go",
    "validation/heracles/spice_netlist.go",
    "validation/calibration/calibration_targets.go",
    "validation/comparator/comparator.go",
    "validation/literature/forc_reference.go",
    # Calibrated materials
    "shared/physics/material_calibrated.go",
    "shared/physics/material_config.go",
    # Config loading
    "config/physics/physics.go",
]


async def main():
    total = 0
    skipped = []
    for rel in NEW_DOCS:
        p = REPO / rel
        if not p.exists():
            skipped.append(rel)
            continue
        await cognee.add(str(p))
        total += 1

    print(f"Added {total} files, {len(skipped)} skipped")
    if skipped:
        print(f"Skipped: {skipped[:5]}...")

    print("Running cognify...")
    await cognee.cognify()
    print("Done — KG updated with academic papers and validation data.")


if __name__ == "__main__":
    asyncio.run(main())
