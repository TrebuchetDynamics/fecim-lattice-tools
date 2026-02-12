#!/usr/bin/env bash
set -euo pipefail

# Run LK headless mode on 3 deterministic targets and summarize perf evidence.
# Outputs:
#  - Raw run log in logs/lk-perf-evidence-<timestamp>.log
#  - Parsed summary printed to stdout

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p logs
stamp="$(date +%Y%m%d_%H%M%S)"
out_log="logs/lk-perf-evidence-${stamp}.log"

export FECIM_ISPP_TARGET_LEVELS="lo,mid,hi"
export FECIM_HEADLESS_FAST="${FECIM_HEADLESS_FAST:-1}"
export FECIM_ISPP_STEPS_PER_PULSE="${FECIM_ISPP_STEPS_PER_PULSE:-400}"
export FECIM_ISPP_MAX_PULSES="${FECIM_ISPP_MAX_PULSES:-1200}"

cmd=(go run ./cmd/fecim-lattice-tools --logger --verbosity debug --mode hysteresis --engine lk)

echo "[lk-perf] running: ${cmd[*]}" | tee "$out_log"
"${cmd[@]}" >> "$out_log" 2>&1

echo "[lk-perf] raw log: $out_log"
echo

echo "=== LK PERF (3 targets) ==="
awk '
  /LK_PERF ISPP_/ {
    label=""; steps=""; dtMin=""; dtMean=""; dtMax=""; solverMs="";
    n=split($0, a, /[[:space:]]+/);
    for (i=1; i<=n; i++) {
      if (a[i] ~ /^ISPP_/) label=a[i];
      else if (a[i] ~ /^steps=/) { split(a[i],kv,"="); steps=kv[2]; }
      else if (a[i] ~ /^dtMin=/) { split(a[i],kv,"="); dtMin=kv[2]; }
      else if (a[i] ~ /^dtMean=/) { split(a[i],kv,"="); dtMean=kv[2]; }
      else if (a[i] ~ /^dtMax=/) { split(a[i],kv,"="); dtMax=kv[2]; }
      else if (a[i] ~ /^solverMs=/) { split(a[i],kv,"="); solverMs=kv[2]; }
    }
    printf("%-16s steps=%-6s dtMin=%-11s dtMean=%-11s dtMax=%-11s solverMs=%s\n", label, steps, dtMin, dtMean, dtMax, solverMs);
  }
' "$out_log"

echo
echo "=== LK ISPP convergence + overshoot accounting ==="
awk '
  /ISPP step [0-9]+/ {
    line=$0;
    gsub(/,/, "", line);
    n=split(line, a, /[[:space:]]+/);
    label=""; target=""; attempts=""; success=""; overs=""; maxd=""; stuck="";
    for (i=1; i<=n; i++) {
      if (a[i] ~ /^\(T[0-9]+_L[0-9]+\):$/) {
        label=a[i];
        gsub(/[():]/, "", label);
      } else if (a[i] ~ /^targetLevel=/) { split(a[i],kv,"="); target=kv[2]; }
      else if (a[i] ~ /^attempts=/) { split(a[i],kv,"="); attempts=kv[2]; }
      else if (a[i] ~ /^success=/) { split(a[i],kv,"="); success=kv[2]; }
      else if (a[i] ~ /^overshoots=/) { split(a[i],kv,"="); overs=kv[2]; }
      else if (a[i] ~ /^maxLevelDelta=/) { split(a[i],kv,"="); maxd=kv[2]; }
      else if (a[i] ~ /^stuckBreakers=/) { split(a[i],kv,"="); stuck=kv[2]; }
    }
    printf("%-10s target=%-4s attempts=%-5s success=%-5s overshoots=%-4s maxDelta=%-4s stuckBreakers=%s\n", label, target, attempts, success, overs, maxd, stuck);
  }
' "$out_log"
