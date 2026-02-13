#!/usr/bin/env bash
set -u

PIN_FILE="tools/external/README.md"

if [[ ! -f "$PIN_FILE" ]]; then
  echo "error: $PIN_FILE not found" >&2
  exit 2
fi

get_current_version() {
  local tool="$1"
  case "$tool" in
    "ngspice")
      command -v ngspice >/dev/null 2>&1 || { echo ""; return; }
      ngspice -v 2>&1 | sed -n 's/.*ngspice-\([0-9][0-9]*\).*/\1/p' | head -n1
      ;;
    "iverilog")
      command -v iverilog >/dev/null 2>&1 || { echo ""; return; }
      iverilog -V 2>&1 | sed -n 's/.*version \([0-9][0-9.]*\).*/\1/p' | head -n1
      ;;
    "verilator")
      command -v verilator >/dev/null 2>&1 || { echo ""; return; }
      verilator --version 2>&1 | awk '{print $2}' | head -n1
      ;;
    "go")
      command -v go >/dev/null 2>&1 || { echo ""; return; }
      go version | awk '{print $3}' | sed 's/^go//' | head -n1
      ;;
    "python")
      command -v python3 >/dev/null 2>&1 || { echo ""; return; }
      python3 --version 2>&1 | awk '{print $2}' | head -n1
      ;;
    *)
      echo ""
      ;;
  esac
}

extract_pin() {
  local key="$1"
  case "$key" in
    ngspice)
      grep -E '^\| ngspice \|' "$PIN_FILE" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    iverilog)
      grep -E '^\| Icarus Verilog .*\|' "$PIN_FILE" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    verilator)
      grep -E '^\| Verilator \|' "$PIN_FILE" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    go)
      grep -E '^\| Go toolchain \|' "$PIN_FILE" | awk -F'|' '{gsub(/`/,"",$3); gsub(/x$/,"",$3); gsub(/ /,"",$3); print $3}'
      ;;
    python)
      grep -E '^\| Python scientific stack .*\|' "$PIN_FILE" | awk -F'|' '{gsub(/`/,"",$3); gsub(/ /,"",$3); split($3,a,","); sub(/^numpy==/,"",a[1]); print a[1]}'
      ;;
  esac
}

printf "%-12s | %-16s | %-16s | %s\n" "tool" "pinned" "current" "status"
printf "%s\n" "----------------------------------------------------------------"

exit_code=0
for tool in ngspice iverilog verilator go python; do
  pinned="$(extract_pin "$tool")"
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
