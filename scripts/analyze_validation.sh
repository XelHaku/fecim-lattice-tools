#!/usr/bin/env bash
# scripts/analyze_validation.sh - Analyzes validation history and prints summary statistics.
# Usage: ./scripts/analyze_validation.sh
set -u -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

BASE_DIR="output/validation"
HISTORY_FILE="$BASE_DIR/history.jsonl"

if [[ ! -f "$HISTORY_FILE" ]]; then
  echo "Validation Analysis"
  echo "==================="
  echo "No history file found at $HISTORY_FILE"
  echo "Nothing to compare yet."
  exit 0
fi

mapfile -t summary_lines < <(grep -v '^\s*$' "$HISTORY_FILE")
summary_count="${#summary_lines[@]}"

if [[ "$summary_count" -lt 2 ]]; then
  echo "Validation Analysis"
  echo "==================="
  echo "Only one history entry present. No regression baseline available."
  exit 0
fi

latest_summary="${summary_lines[$((summary_count-1))]}"
previous_summary="${summary_lines[$((summary_count-2))]}"

if [[ -z "$latest_summary" || -z "$previous_summary" ]]; then
  echo "Validation Analysis"
  echo "==================="
  echo "Only one history entry present. No regression baseline available."
  exit 0
fi

latest_dir=""
previous_dir=""
mapfile -t run_dirs < <(find "$BASE_DIR" -mindepth 1 -maxdepth 1 -type d | sort)
if [[ ${#run_dirs[@]} -ge 2 ]]; then
  latest_dir="${run_dirs[-1]}"
  previous_dir="${run_dirs[-2]}"
fi

python3 - "$latest_summary" "$previous_summary" "$latest_dir" "$previous_dir" <<'PY'
import json
import os
import sys

latest_summary_json = json.loads(sys.argv[1])
previous_summary_json = json.loads(sys.argv[2])
latest_dir = sys.argv[3]
previous_dir = sys.argv[4]

module_files = {
    "m1": "m1_results.json",
    "m4": "m4_results.json",
    "m6": "m6_results.json",
    "shared": "shared_results.json",
    "validation": "validation_results.json",
}

regressions = []

# Rule 2: any module with fewer passes (history summary comparison)
prev_modules = previous_summary_json.get("modules", {})
cur_modules = latest_summary_json.get("modules", {})
for module in sorted(set(prev_modules.keys()) | set(cur_modules.keys())):
    prev_pass = int(prev_modules.get(module, {}).get("pass", 0))
    cur_pass = int(cur_modules.get(module, {}).get("pass", 0))
    if cur_pass < prev_pass:
        regressions.append(
            f"[MODULE REGRESSION] module={module}: pass count {prev_pass} -> {cur_pass}"
        )

# Rule 1: any test PASS -> FAIL (only when two distinct run dirs are available)
if latest_dir and previous_dir and latest_dir != previous_dir:
    def collect_test_statuses(path):
        statuses = {}
        if not os.path.isfile(path):
            return statuses
        with open(path, "r", encoding="utf-8", errors="replace") as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue
                try:
                    obj = json.loads(line)
                except json.JSONDecodeError:
                    continue
                test = obj.get("Test")
                action = obj.get("Action")
                if not test or action not in {"pass", "fail", "skip"}:
                    continue
                statuses[test] = action
        return statuses

    for module, filename in module_files.items():
        prev_status = collect_test_statuses(os.path.join(previous_dir, filename))
        cur_status = collect_test_statuses(os.path.join(latest_dir, filename))
        for test_name, old_status in prev_status.items():
            new_status = cur_status.get(test_name)
            if old_status == "pass" and new_status == "fail":
                regressions.append(
                    f"[TEST REGRESSION] module={module} test={test_name}: PASS -> FAIL"
                )

print("Validation Regression Analysis")
print("==============================")
print(f"Current timestamp:  {latest_summary_json.get('timestamp', 'unknown')}")
print(f"Previous timestamp: {previous_summary_json.get('timestamp', 'unknown')}")
print("")

if regressions:
    print("Result: REGRESSIONS DETECTED")
    print("")
    for item in regressions:
        print(f"- {item}")
    sys.exit(1)

print("Result: No regressions found")
sys.exit(0)
PY