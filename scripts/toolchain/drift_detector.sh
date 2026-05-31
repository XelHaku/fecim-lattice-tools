#!/usr/bin/env bash
set -u

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/tool_versions.sh
source "$SCRIPT_DIR/lib/tool_versions.sh"

PIN_FILE="${PIN_FILE:-tools/external/README.md}"

if [[ ! -f "$PIN_FILE" ]]; then
  echo "error: $PIN_FILE not found" >&2
  exit 2
fi

printf "%-12s | %-16s | %-16s | %s\n" "tool" "pinned" "current" "status"
printf "%s\n" "----------------------------------------------------------------"

exit_code=0
for tool in ngspice iverilog verilator go python; do
  pinned="$(extract_pin "$tool" "$PIN_FILE")"
  current="$(get_current_version "$tool")"

  if [[ -z "$current" ]]; then
    status="missing"
    exit_code=1
  elif [[ -z "$pinned" ]]; then
    status="missing-pin"
    exit_code=1
  elif [[ "$tool" == "go" ]]; then
    if [[ "$current" == "$pinned"* ]]; then
      status="match"
    else
      status="drift"
      exit_code=1
    fi
  else
    if [[ "$current" == "$pinned" ]]; then
      status="match"
    else
      status="drift"
      exit_code=1
    fi
  fi

  printf "%-12s | %-16s | %-16s | %s\n" "$tool" "$pinned" "${current:-n/a}" "$status"
done

exit $exit_code
