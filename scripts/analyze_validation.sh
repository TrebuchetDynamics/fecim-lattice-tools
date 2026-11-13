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

python3 - "$latest_summary" "$previous_summary" "$BASE_DIR" <<'PY'
import json
import os
import re
import sys

latest_summary_json = json.loads(sys.argv[1])
previous_summary_json = json.loads(sys.argv[2])
base_dir = sys.argv[3]

module_files = {
    "m1": "m1_results.json",
    "m4": "m4_results.json",
    "m6": "m6_results.json",
    "shared": "shared_results.json",
    "validation": "validation_results.json",
}

print("Validation Regression Analysis")
print("==============================")
print(f"Current timestamp:  {latest_summary_json.get('timestamp', 'unknown')}")
print(f"Previous timestamp: {previous_summary_json.get('timestamp', 'unknown')}")
print("")


def unavailable(context):
    print(f"Result: Comparison unavailable / incomplete: {context}")
    sys.exit(1)


def selected_dir(summary, label):
    timestamp = summary.get("timestamp")
    if not isinstance(timestamp, str) or not re.fullmatch(r"[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9]{4}", timestamp):
        unavailable(f"{label} timestamp is missing or invalid: {timestamp!r}")
    path = os.path.join(base_dir, timestamp)
    if os.path.islink(path) or not os.path.isdir(path):
        unavailable(f"{label} selected run directory is missing or unsafe: {path}")
    return path


latest_dir = selected_dir(latest_summary_json, "current")
previous_dir = selected_dir(previous_summary_json, "previous")
if latest_dir == previous_dir:
    unavailable("selected timestamps name the same directory; overwritten evidence cannot be compared")

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

# Rule 1: any test PASS -> FAIL in lanes named by either selected summary.
def collect_test_statuses(path):
    statuses = {}
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
            statuses[(obj.get("Package", ""), test)] = action
    return statuses


for module, filename in module_files.items():
    if module not in prev_modules and module not in cur_modules:
        continue
    try:
        prev_status = collect_test_statuses(os.path.join(previous_dir, filename))
        cur_status = collect_test_statuses(os.path.join(latest_dir, filename))
    except OSError as error:
        unavailable(f"required lane {module} log cannot be read: {error}")
    # Missing or invalid legacy statuses are not evidence of command success.
    prev_exit = prev_modules.get(module, {}).get("exit_code")
    cur_exit = cur_modules.get(module, {}).get("exit_code")
    if (type(prev_exit) is int and 0 <= prev_exit <= 255
            and type(cur_exit) is int and 0 <= cur_exit <= 255
            and prev_exit == 0 and cur_exit != 0):
        regressions.append(
            f"[LANE EXIT REGRESSION] module={module}: exit_code {prev_exit} -> {cur_exit}"
        )
    for identity, old_status in prev_status.items():
        package, test_name = identity
        new_status = cur_status.get(identity)
        if old_status == "pass" and new_status == "fail":
            regressions.append(
                f"[TEST REGRESSION] module={module} package={package!r} test={test_name!r}: PASS -> FAIL"
            )

if regressions:
    print("Result: REGRESSIONS DETECTED")
    print("")
    for item in regressions:
        print(f"- {item}")
    sys.exit(1)

print("Result: No regressions found")
sys.exit(0)
PY