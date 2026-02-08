#!/usr/bin/env bash
set -euo pipefail

GO_TEST_RACE_TIMEOUT="${GO_TEST_RACE_TIMEOUT:-20m}"

# Keep the race set small/targeted to control runtime.
# Avoid GUI-heavy packages to keep CI headless and deterministic.
RACE_PKGS=(
  "./shared/compute"
  "./shared/gpu"
  "./shared/io"
  "./shared/logging"
  "./shared/peripherals"
  "./shared/physics"
  "./shared/recording"
  "./shared/theme"
  "./shared/utils"
  "./module1-hysteresis/pkg/algo"
  "./module1-hysteresis/pkg/ferroelectric"
  "./module1-hysteresis/pkg/simulation"
  "./module2-crossbar/pkg/crossbar"
  "./module2-crossbar/pkg/network"
  "./module2-crossbar/pkg/training"
  "./module2-crossbar/pkg/weights"
  "./module3-mnist/pkg/core"
  "./module3-mnist/pkg/mnist"
  "./module3-mnist/pkg/training"
  "./module4-circuits/pkg/arraysim"
  "./module5-comparison/pkg/comparison"
  "./module6-eda/pkg/compiler"
  "./module6-eda/pkg/config"
)

exec go test \
  -tags=ci \
  -race \
  -count=1 \
  -shuffle=off \
  -trimpath \
  -timeout="${GO_TEST_RACE_TIMEOUT}" \
  "${RACE_PKGS[@]}"
