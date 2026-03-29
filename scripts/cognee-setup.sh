#!/usr/bin/env bash
# cognee-setup.sh — Bootstrap Cognee knowledge engine (CLI-only, no Docker)
#
# Creates a local Python venv and installs cognee.
# Data is stored in .cognee_system/ (repo-local, gitignored).
#
# Usage:
#   ./scripts/cognee-setup.sh          # install / update
#   source .venv-cognee/bin/activate   # then use cognee-cli

set -euo pipefail
cd "$(dirname "$0")/.."

VENV=".venv-cognee"

if [ ! -d "$VENV" ]; then
  echo "Creating Python venv at $VENV ..."
  python3 -m venv "$VENV"
fi

echo "Installing cognee + fastembed ..."
"$VENV/bin/pip" install --upgrade pip --quiet
"$VENV/bin/pip" install --upgrade cognee fastembed --quiet

# Copy .env.example if no .env exists yet
if [ ! -f .env ] && [ -f .env.example ]; then
  cp .env.example .env
  echo "Created .env from .env.example — edit it to add your LLM_API_KEY."
fi

echo ""
echo "Done. Activate and use:"
echo "  source $VENV/bin/activate"
echo "  cognee-cli add \"your text\""
echo "  cognee-cli cognify"
echo "  cognee-cli search \"your query\""
