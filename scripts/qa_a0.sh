#!/usr/bin/env bash
# scripts/qa_a0.sh - Quick QA check: runs all tests and reports summary.
# Usage: ./scripts/qa_a0.sh [output.json]
set -euo pipefail

# Ensure Go is discoverable in cron/non-interactive shells.
export PATH="/usr/local/go/bin:${PATH}"

cd "$(dirname "$0")/.."

LIST_TOTAL=$(go list ./... | wc -l | tr -d ' ')

JSON_PATH="${1:-/tmp/fecim_gotest.json}"

go test -json ./... > "$JSON_PATH"
TEST_EC=$?

# Aggregate package status deterministically (no jq dependency).
# NOTE: gotest_pkgsum exits 1 if fail>0, so we capture output first.
set +e
PKG_SUM_LINE=$(go run ./scripts/gotest_pkgsum/main.go "$JSON_PATH")
PKG_SUM_EC=$?
set -e

JSON_TOTAL=$(echo "$PKG_SUM_LINE" | sed -n 's/.*total=\([0-9]\+\).*/\1/p')

echo "LIST_TOTAL=${LIST_TOTAL} JSON_TOTAL=${JSON_TOTAL}"
echo "$PKG_SUM_LINE"

if [[ -z "${JSON_TOTAL}" ]]; then
  echo "ERROR: could not parse JSON_TOTAL from PKG_SUM output" >&2
  exit 2
fi

if [[ "${LIST_TOTAL}" != "${JSON_TOTAL}" ]]; then
  echo "ERROR: package total mismatch (LIST_TOTAL != JSON_TOTAL) — possible truncation/partial capture" >&2
  exit 2
fi

# Prefer the deterministic package summary exit code; it encodes fail>0.
if [[ $PKG_SUM_EC -ne 0 ]]; then
  exit $PKG_SUM_EC
fi

# If tests itself failed but pkgsum didn't catch it (shouldn't happen), fail.
if [[ $TEST_EC -ne 0 ]]; then
  exit $TEST_EC
fi
