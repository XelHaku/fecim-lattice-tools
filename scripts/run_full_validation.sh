#!/usr/bin/env bash
# scripts/run_full_validation.sh - Runs all validation lanes (Module 1 + Module 4 + literature).
# Usage: ./scripts/run_full_validation.sh
set -u -o pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

timestamp="$(date +%Y-%m-%d-%H%M)"
out_base="output/validation"
out_dir="$out_base/$timestamp"
mkdir -p "$out_dir"

commit="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"

TEST_TIMEOUT="${VALIDATION_TEST_TIMEOUT:-600s}"

run_module_tests() {
  local module_path="$1"
  local output_file="$2"

  echo "Running: timeout ${TEST_TIMEOUT} go test -short -count=1 -json ./${module_path}/..."
  if ! timeout "${TEST_TIMEOUT}" go test -short -count=1 -json "./${module_path}/..." > "$output_file" 2>&1; then
    rc=$?
    if [[ $rc -eq 124 ]]; then
      echo "Warning: ${module_path} timed out after ${TEST_TIMEOUT} (details in ${output_file})"
    else
      echo "Warning: tests reported failures for ${module_path} (details in ${output_file})"
    fi
  fi
}

extract_counts() {
  local file="$1"
  python3 - "$file" <<'PY'
import json
import sys

path = sys.argv[1]
counts = {"pass": 0, "fail": 0, "skip": 0}

with open(path, "r", encoding="utf-8", errors="replace") as f:
    for line in f:
        line = line.strip()
        if not line:
            continue
        try:
            obj = json.loads(line)
        except json.JSONDecodeError:
            continue
        if "Test" not in obj:
            continue
        action = obj.get("Action")
        if action in counts:
            counts[action] += 1

print(f'{counts["pass"]} {counts["fail"]} {counts["skip"]}')
PY
}

run_module_tests "module1-hysteresis" "$out_dir/m1_results.json"
run_module_tests "module4-circuits" "$out_dir/m4_results.json"
run_module_tests "module6-eda" "$out_dir/m6_results.json"
run_module_tests "shared" "$out_dir/shared_results.json"
run_module_tests "validation" "$out_dir/validation_results.json"

read -r m1_pass m1_fail m1_skip <<< "$(extract_counts "$out_dir/m1_results.json")"
read -r m4_pass m4_fail m4_skip <<< "$(extract_counts "$out_dir/m4_results.json")"
read -r m6_pass m6_fail m6_skip <<< "$(extract_counts "$out_dir/m6_results.json")"
read -r shared_pass shared_fail shared_skip <<< "$(extract_counts "$out_dir/shared_results.json")"
read -r val_pass val_fail val_skip <<< "$(extract_counts "$out_dir/validation_results.json")"

total_pass=$((m1_pass + m4_pass + m6_pass + shared_pass + val_pass))
total_fail=$((m1_fail + m4_fail + m6_fail + shared_fail + val_fail))

summary_file="$out_dir/validation_summary.json"
cat > "$summary_file" <<JSON
{"timestamp":"$timestamp","commit":"$commit","modules":{"m1":{"pass":$m1_pass,"fail":$m1_fail,"skip":$m1_skip},"m4":{"pass":$m4_pass,"fail":$m4_fail,"skip":$m4_skip},"m6":{"pass":$m6_pass,"fail":$m6_fail,"skip":$m6_skip},"shared":{"pass":$shared_pass,"fail":$shared_fail,"skip":$shared_skip},"validation":{"pass":$val_pass,"fail":$val_fail,"skip":$val_skip}},"total_pass":$total_pass,"total_fail":$total_fail}
JSON

history_file="$out_base/history.jsonl"
mkdir -p "$out_base"
cat "$summary_file" >> "$history_file"
echo >> "$history_file"

echo "Validation run complete."
echo "Output directory: $out_dir"
echo "Summary: $summary_file"