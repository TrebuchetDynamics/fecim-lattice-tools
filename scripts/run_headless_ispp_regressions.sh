#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${FECIM_REGRESSION_JSON_DIR:-$REPO_ROOT/output/regression}"

mkdir -p "$OUT_DIR"
export FECIM_REGRESSION_JSON_DIR="$OUT_DIR"

echo "[regression] output dir: $FECIM_REGRESSION_JSON_DIR"
echo "[regression] running Preisach + LK headless WRD/ISPP suites"

go test ./module1-hysteresis/pkg/controller \
  -run 'TestHeadlessRegression_WRD_ISPP_(Preisach|LK)$' \
  -count=1 -v

echo "[regression] summaries:"
ls -1 "$OUT_DIR"/*.json
