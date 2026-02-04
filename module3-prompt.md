Role

  - You are an expert software engineer and neural network / analog CIM scientist.
  - Operate fully autonomously. Do not ask questions unless genuinely blocked by missing inputs/files.
  - If an ambiguity remains, choose the most reasonable default and proceed; document the choice.

Objective

  - Ensure the Module 3 MNIST neural network implementation fully matches the equations, data flow,
    and behavior described in docs/neural-network/mnist.architecture.md, docs/neural-network/mnist.demo.md,
    docs/neural-network/mnist.research.md, and module3-mnist/FEATURES.md.
  - Make any required code + documentation updates to achieve fidelity and verify via CLI output and logs.
  - Improve Module 3 documentation quality and ensure referenced papers are downloaded into the repo's
    research-papers area when possible.

Tasks

  1. Inference and quantization fidelity (no approximations unless explicitly called out)

  - Verify network topology (784 -> 128 -> 10) and single-layer mode (784 -> 10) in code and docs.
  - Validate FP path math (linear layers, ReLU, softmax, normalization) and output probabilities.
  - Validate CIM path math: weight quantization to N levels, DAC/ADC quantization, noise injection,
    per-layer quantization, and output scaling order.
  - Confirm disagreement metrics (KL divergence or similar), accuracy tracking, and confusion matrix logic.
  - Verify energy and performance models shown in GUI match documented formulas and defaults.
  - Cross-check variable names, units, and parameter mappings between code and docs.
  - If gaps are found, implement fixes and update docs accordingly.

  2. Data and weight handling

  - Validate MNIST IDX parsing, bounds checks, and sanity limits for dataset sizes.
  - Verify weight file loading, QAT level selection, and fallback behavior.
  - Ensure quantization stats and level counts reflect actual quantized weights.
  - Confirm deterministic behavior when a debug seed is configured.

  3. Architecture and UI alignment

  - Confirm DualModeApp and MNISTApp controls map to NetworkConfig ranges and defaults.
  - Ensure UI updates from goroutines use fyne.Do() consistently.
  - Check that labels/claims (accuracy targets, hardware claims) align with honesty policy and docs.

  4. Architecture documentation

  - Update docs/neural-network/mnist.architecture.md and docs/development/ARCHITECTURE.md to reflect
    Module 3 data flow, responsibilities, and interfaces.
  - Update docs/development/GUI/GUI.module3.md if UI behavior changes.

Validation

  - Run: go test ./module3-mnist/...
  - Run: go test -race ./module3-mnist/pkg/core -run TestConcurrent
  - If GUI verification is required: go run ./cmd/fecim-lattice-tools mnist
  - If CLI verification is required: go run ./cmd/fecim-lattice-tools mnist cli -evaluate (dataset required)
  - Use logs to confirm dual-path inference, quantization, and noise settings are exercised.
  - If any command fails, fix and re-run until it succeeds or a clear blocker exists.

Execution Rules (Autonomous)

  - No human intermediaries: run commands, inspect logs, make edits, and validate independently.
  - Always check logs in logs/ for the most recent run and quote key evidence in the report.
  - Prefer minimal, targeted changes over refactors unless required for correctness.
  - Keep code changes within the smallest possible surface area.
  - If a new CLI flag or headless pathway is required for validation, implement it.
  - If tests or validation scripts are needed, add them temporarily, run, then remove before final output.
  - Never skip validation; if blocked, report exact error output and the last command run.

Deliverable

  - A concise report that includes:
      - What was verified (inference math, quantization, metrics, UI alignment)
      - Documentation changes made (file paths + summary)
      - Any gaps, issues, or follow-ups needed
