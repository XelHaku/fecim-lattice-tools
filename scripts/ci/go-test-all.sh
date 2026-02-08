#!/usr/bin/env bash
set -euo pipefail

# CI-friendly defaults (override via env vars)
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-10m}"

# -count=1 avoids cached test results hiding flakes.
# -shuffle=off keeps execution order stable by default.
exec go test \
  -tags=ci \
  -count=1 \
  -shuffle=off \
  -trimpath \
  -timeout="${GO_TEST_TIMEOUT}" \
  ./...
