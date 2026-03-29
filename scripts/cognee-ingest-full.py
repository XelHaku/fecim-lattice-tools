#!/usr/bin/env python3
"""Full ingest of FeCIM repo into Cognee knowledge graph.

Feeds: physics code, crossbar models, neural network code, circuits,
EDA pipeline, validation, research docs, GUI code, and more.
"""
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

# ── Phase 1: Core Physics Engine ──
PHYSICS = [
    # Landau-Khalatnikov solver
    "shared/physics/lk_solver.go",
    "shared/physics/lk_solver_test.go",
    "shared/physics/lk_rk4.go",
    "shared/physics/lk_rk4_test.go",
    "shared/physics/lk_types.go",
    # Preisach hysteresis model
    "shared/physics/preisach.go",
    "shared/physics/preisach_test.go",
    "shared/physics/preisach_types.go",
    "shared/physics/preisach_distribution.go",
    "shared/physics/preisach_everett.go",
    # ISPP write controller
    "shared/physics/ispp.go",
    "shared/physics/ispp_test.go",
    "shared/physics/ispp_types.go",
    # Temperature & retention
    "shared/physics/retention.go",
    "shared/physics/wakeup.go",
    "shared/physics/dispersion.go",
    "shared/physics/temperature.go",
]

# ── Phase 2: Crossbar Array (core compute-in-memory) ──
CROSSBAR = [
    "shared/crossbar/array.go",
    "shared/crossbar/array_test.go",
    "shared/crossbar/mvm.go",
    "shared/crossbar/mvm_test.go",
    "shared/crossbar/quantize.go",
    "shared/crossbar/quantize_test.go",
    "shared/crossbar/nonidealities.go",
    "shared/crossbar/nonidealities_test.go",
    "shared/crossbar/irdrop.go",
    "shared/crossbar/sneak.go",
    "shared/crossbar/drift.go",
    "shared/crossbar/conductance.go",
    "shared/crossbar/types.go",
    "shared/crossbar/config.go",
    "shared/crossbar/preset.go",
]

# ── Phase 3: Module1 - Hysteresis ──
HYSTERESIS = [
    "module1-hysteresis/pkg/ferroelectric/material.go",
    "module1-hysteresis/pkg/ferroelectric/material_test.go",
    "module1-hysteresis/pkg/ferroelectric/hysteresis.go",
    "module1-hysteresis/pkg/ferroelectric/types.go",
    "module1-hysteresis/pkg/controller/controller.go",
    "module1-hysteresis/pkg/controller/modes.go",
    "module1-hysteresis/pkg/gui/embedded.go",
    "module1-hysteresis/pkg/gui/app.go",
]

# ── Phase 4: Module2 - Crossbar GUI ──
CROSSBAR_GUI = [
    "module2-crossbar/pkg/gui/embedded.go",
    "module2-crossbar/pkg/gui/app.go",
    "module2-crossbar/pkg/gui/heatmap.go",
    "module2-crossbar/pkg/gui/controls.go",
]

# ── Phase 5: Module3 - MNIST Neural Network ──
MNIST = [
    "module3-mnist/pkg/training/network.go",
    "module3-mnist/pkg/training/network_test.go",
    "module3-mnist/pkg/training/single_layer.go",
    "module3-mnist/pkg/training/trainer_foundation.go",
    "module3-mnist/pkg/mnist/loader.go",
    "module3-mnist/pkg/gui/embedded.go",
    "module3-mnist/pkg/gui/dualmode.go",
    "module3-mnist/pkg/gui/quantization_widget.go",
    "module3-mnist/pkg/core/shim.go",
    "module3-mnist/README.md",
    "module3-mnist/FEATURES.md",
]

# ── Phase 6: Module4 - Circuits/Peripherals ──
CIRCUITS = [
    "module4-circuits/pkg/peripherals/dac.go",
    "module4-circuits/pkg/peripherals/adc.go",
    "module4-circuits/pkg/peripherals/tia.go",
    "module4-circuits/pkg/peripherals/types.go",
    "module4-circuits/pkg/peripherals/pipeline.go",
    "module4-circuits/pkg/peripherals/sense_chain.go",
    "module4-circuits/pkg/gui/embedded.go",
    "module4-circuits/pkg/gui/app.go",
    "module4-circuits/pkg/solver/solver.go",
]

# ── Phase 7: Module5 - Technology Comparison ──
COMPARISON = [
    "module5-comparison/pkg/gui/embedded.go",
    "module5-comparison/pkg/gui/app.go",
    "module5-comparison/pkg/gui/comparison_data.go",
]

# ── Phase 8: Module6 - EDA Pipeline ──
EDA = [
    "module6-eda/pkg/config/types.go",
    "module6-eda/pkg/gui/embedded.go",
    "module6-eda/pkg/gui/app.go",
    "module6-eda/pkg/gui/pipeline.go",
    "module6-eda/pkg/synthesis/synthesis.go",
    "module6-eda/pkg/place/place.go",
    "module6-eda/pkg/route/route.go",
    "module6-eda/pkg/drc/drc.go",
]

# ── Phase 9: Shared Infrastructure ──
SHARED = [
    "shared/widgets/embedded_base.go",
    "shared/widgets/heatmap.go",
    "shared/widgets/plot.go",
    "shared/theme/theme.go",
    "shared/logging/logger.go",
]

# ── Phase 10: Validation & Benchmarks ──
VALIDATION = [
    "validation/literature.go",
    "validation/statistics.go",
    "validation/interfaces.go",
    "validation/readiness_report.go",
    "validation/benchmarks/benchmark_suite.go",
    "validation/configvalidator/validator.go",
    "validation/configvalidator/calibration.go",
    "validation/configvalidator/preisach.go",
    "validation/configvalidator/README.md",
]

# ── Phase 11: Main app & CLI ──
APP = [
    "cmd/fecim-lattice-tools/main.go",
    "cmd/fecim-lattice-tools/launcher.go",
    "cmd/fecim-lattice-tools/mode.go",
    "cmd/fecim-lattice-tools/seed.go",
    "cmd/fecim-lattice-tools/subcommands.go",
]

# ── Phase 12: All remaining docs (non-archive) ──
DOCS_RESEARCH = [
    "docs/4-research/honesty-audit.md",
    "docs/4-research/tour-group-ironlattice-research.md",
    "docs/4-research/superlattice-material-analysis.md",
    "docs/4-research/literature-review/crossbar-circuits-literature-review-2025.md",
    "docs/4-research/literature-review/hzo-hysteresis-validation-data.md",
    "docs/4-research/literature-review/world-class-gap-analysis.md",
    "docs/4-research/validation/CLAIMS-MATRIX.md",
    "docs/4-research/validation/confidence-policy.md",
    "docs/4-research/validation/coverage-boundary.md",
    "docs/4-research/validation/M1-M4-physics-contract.md",
    "docs/4-research/validation/M4-M6-physics-depth-audit.md",
    "docs/4-research/validation/RESEARCH_GRADE_TESTING_STANDARDS.md",
    "docs/4-research/validation/verification-vs-validation.md",
    "docs/4-research/internal-analysis/crossbar-arrays.md",
    "docs/4-research/internal-analysis/hysteresis-physics.md",
    "docs/4-research/internal-analysis/cim-circuits.md",
    "docs/4-research/internal-analysis/circuits.CIM-fundamentals.md",
    "docs/4-research/opensource-tools/tool-comparison-matrix.md",
    "docs/4-research/opensource-tools/memristor-rram-tools.md",
    "docs/4-research/opensource-tools/circuit-analysis-libraries.md",
    "docs/4-research/papers/TOPIC_SUMMARIES.md",
    "docs/4-research/papers/RESEARCH_GAP_ANALYSIS.md",
    "docs/4-research/papers/external-research/README.md",
    "docs/4-research/transcripts/COSM_2025_AI_Hardware_Breakthrough/ironlattice-transcript.md",
]

DOCS_LEARN = [
    "docs/2-learn/module1-hysteresis/eli5.md",
    "docs/2-learn/module1-hysteresis/tools.md",
    "docs/2-learn/module1-hysteresis/run-modes.md",
    "docs/2-learn/module2-crossbar/eli5.md",
    "docs/2-learn/module3-mnist/eli5.md",
    "docs/2-learn/module3-mnist/README.md",
    "docs/2-learn/module3-mnist/tools.md",
    "docs/2-learn/module4-circuits/eli5.md",
    "docs/2-learn/module4-circuits/features.md",
    "docs/2-learn/module4-circuits/tools.md",
    "docs/2-learn/module5-comparison/eli5.md",
    "docs/2-learn/module5-comparison/features.md",
    "docs/2-learn/module6-eda/eli5.md",
]

DOCS_DEV = [
    "docs/3-develop/gui/FYNE_NOTES.md",
    "docs/3-develop/gui/GUI.module2.md",
    "docs/3-develop/gui/GUI.module5.md",
    "docs/3-develop/gui/GUI.module7.md",
    "docs/3-develop/testing/TESTING.md",
    "docs/3-develop/api-reference.md",
]

DOCS_ROOT = [
    "CLAUDE.md",
    "README.md",
    "CHANGELOG.md",
    "CONTRIBUTING.md",
    "AGENTS.md",
    "TODO.md",
    "status.md",
]

# ── Phase 13: Paper topic summaries (by-topic READMEs) ──
PAPER_TOPICS = [
    f"docs/4-research/papers/by-topic/{d}/README.md"
    for d in [
        "01-ferroelectric-materials", "02-training-algorithms",
        "03-simulation-tools", "04-cim-architectures",
        "05-neuromorphic", "06-photonic-computing",
        "07-memory-architectures", "08-industry-reports",
        "09-reviews-surveys", "10-cim-compilers-mapping",
        "11-reservoir-computing", "12-spiking-neural-networks",
        "13-in-memory-training", "14-transformer-llm-accelerators",
        "15-3d-stacking-architectures", "16-photonic-ferroelectric-hybrids",
        "17-security-cryptography", "18-ald-process-control",
        "19-variability-yield", "20-manufacturing-integration",
        "21-3d-stacking", "22-automotive-harsh-env",
        "23-cryogenic-operation",
    ]
]

ALL_PHASES = [
    ("Physics Engine", PHYSICS),
    ("Crossbar Array", CROSSBAR),
    ("Module1: Hysteresis", HYSTERESIS),
    ("Module2: Crossbar GUI", CROSSBAR_GUI),
    ("Module3: MNIST", MNIST),
    ("Module4: Circuits", CIRCUITS),
    ("Module5: Comparison", COMPARISON),
    ("Module6: EDA", EDA),
    ("Shared Infra", SHARED),
    ("Validation", VALIDATION),
    ("Main App", APP),
    ("Research Docs", DOCS_RESEARCH),
    ("Learning Docs", DOCS_LEARN),
    ("Dev Docs", DOCS_DEV),
    ("Root Docs", DOCS_ROOT),
    ("Paper Topics", PAPER_TOPICS),
]


async def main():
    total_added = 0
    total_skipped = 0

    for phase_name, files in ALL_PHASES:
        added = 0
        skipped = []
        for rel in files:
            p = REPO / rel
            if not p.exists():
                skipped.append(rel)
                total_skipped += 1
                continue
            await cognee.add(str(p))
            added += 1
            total_added += 1
        status = f"  {phase_name}: {added} added"
        if skipped:
            status += f", {len(skipped)} skipped"
        print(status)

    print(f"\nTotal: {total_added} files added, {total_skipped} skipped")
    print("Running cognify (this may take several minutes)...")
    await cognee.cognify()
    print("Done — knowledge graph built.")


if __name__ == "__main__":
    asyncio.run(main())
