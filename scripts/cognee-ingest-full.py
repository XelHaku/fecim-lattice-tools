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
    "shared/crossbar/core.go",
    "shared/crossbar/array.go",
    "shared/crossbar/enhanced.go",
    "shared/crossbar/solver.go",
    "shared/crossbar/solver_optimized.go",
    "shared/crossbar/nonidealities.go",
    "shared/crossbar/irdrop.go",
    "shared/crossbar/sneakpath.go",
    "shared/crossbar/sneak_multihop.go",
    "shared/crossbar/drift.go",
    "shared/crossbar/drift_calibration.go",
    "shared/crossbar/device_errors.go",
    "shared/crossbar/uncertainty.go",
    "shared/crossbar/fecap.go",
    "shared/crossbar/thermal.go",
    "shared/crossbar/temperature.go",
    "shared/crossbar/temperature_profile.go",
    "shared/crossbar/nonlinear_iv.go",
    "shared/crossbar/write_disturb.go",
    "shared/crossbar/tiling.go",
    "shared/crossbar/gpu_mvm.go",
    "shared/crossbar/variation_import.go",
    "shared/crossbar/memtorch_export.go",
]

# ── Phase 3: Module1 - Hysteresis ──
HYSTERESIS = [
    "module1-hysteresis/pkg/ferroelectric/material.go",
    "module1-hysteresis/pkg/ferroelectric/preisach.go",
    "module1-hysteresis/pkg/ferroelectric/level_bins.go",
    "module1-hysteresis/pkg/ferroelectric/render.go",
    "module1-hysteresis/pkg/ferroelectric/ferroelectric_test.go",
    "module1-hysteresis/pkg/controller/writer.go",
    "module1-hysteresis/pkg/simulation/engine.go",
    "module1-hysteresis/pkg/simulation/multicell.go",
    "module1-hysteresis/pkg/gui/embedded.go",
    "module1-hysteresis/pkg/gui/gui.go",
    "module1-hysteresis/pkg/gui/controls.go",
    "module1-hysteresis/pkg/gui/sim_loop.go",
    "module1-hysteresis/pkg/algo/calibration.go",
]

# ── Phase 4: Module2 - Crossbar GUI ──
CROSSBAR_GUI = [
    "module2-crossbar/pkg/gui/embedded.go",
    "module2-crossbar/pkg/gui/app.go",
    "module2-crossbar/pkg/gui/heatmap.go",
    "module2-crossbar/pkg/gui/controls.go",
    "module2-crossbar/pkg/gui/app_enhanced.go",
    "module2-crossbar/pkg/gui/tabbed_app.go",
    "module2-crossbar/pkg/gui/widgets.go",
    "module2-crossbar/pkg/gui/animation.go",
    "module2-crossbar/pkg/gui/keyboard.go",
    "module2-crossbar/pkg/network/network.go",
    "module2-crossbar/pkg/weights/weights.go",
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
    "module4-circuits/pkg/arraysim/types.go",
    "module4-circuits/pkg/arraysim/tier_a.go",
    "module4-circuits/pkg/arraysim/tier_b.go",
    "module4-circuits/pkg/arraysim/sensechain.go",
    "module4-circuits/pkg/arraysim/transient.go",
    "module4-circuits/pkg/arraysim/array_config.go",
    "module4-circuits/pkg/arraysim/array_ispp.go",
    "module4-circuits/pkg/arraysim/refsolve_dense.go",
    "module4-circuits/pkg/arraysim/endurance_accuracy.go",
    "module4-circuits/pkg/arraysim/read_margin_analysis.go",
    "module4-circuits/pkg/arraysim/process_variation_mc.go",
    "module4-circuits/pkg/arraysim/design_space_exploration.go",
    "module4-circuits/pkg/arraysim/spice_export.go",
    "module4-circuits/pkg/gui/embedded.go",
    "module4-circuits/pkg/gui/app.go",
    "module4-circuits/pkg/gui/tab_unified.go",
]

# ── Phase 7: Module5 - Technology Comparison ──
COMPARISON = [
    "module5-comparison/pkg/gui/embedded.go",
    "module5-comparison/pkg/gui/app.go",
    "module5-comparison/pkg/gui/evidence_model.go",
    "module5-comparison/pkg/gui/fabrication_reality.go",
    "module5-comparison/pkg/gui/scenario_profiles.go",
    "module5-comparison/pkg/comparison/architecture.go",
    "module5-comparison/pkg/comparison/cacti_baseline.go",
]

# ── Phase 8: Module6 - EDA Pipeline ──
EDA = [
    "module6-eda/pkg/config/types.go",
    "module6-eda/pkg/compiler/types.go",
    "module6-eda/pkg/compiler/compiler.go",
    "module6-eda/pkg/export/def.go",
    "module6-eda/pkg/export/lef.go",
    "module6-eda/pkg/export/liberty.go",
    "module6-eda/pkg/export/spice.go",
    "module6-eda/pkg/export/verilog.go",
    "module6-eda/pkg/export/csv.go",
    "module6-eda/pkg/export/json.go",
    "module6-eda/pkg/export/svg.go",
    "module6-eda/pkg/export/openlane_config.go",
    "module6-eda/pkg/layout/def_generator.go",
    "module6-eda/pkg/layout/verilog_generator.go",
    "module6-eda/pkg/openlane/runner.go",
    "module6-eda/pkg/validate/pdk_bridge.go",
    "module6-eda/pkg/validation/cross_check.go",
    "module6-eda/pkg/validation/def_validator.go",
    "module6-eda/pkg/gui/embedded.go",
    "module6-eda/pkg/gui/app.go",
]

# ── Phase 9: Shared Infrastructure ──
SHARED = [
    "shared/widgets/embedded_base.go",
    "shared/widgets/embedded_app.go",
    "shared/widgets/material_picker.go",
    "shared/widgets/preset_selector.go",
    "shared/widgets/confidence_badge.go",
    "shared/theme/theme.go",
    "shared/logging/logging.go",
    "shared/logging/manager.go",
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
