#!/usr/bin/env bash
set -euo pipefail

# RG-VAL-05: Release report generator.
# Usage:
#   ./scripts/gen_release_report.sh [input_dir]

IN_DIR=${1:-}
if [[ -z "$IN_DIR" ]]; then
  if [[ -d output/regression ]]; then
    IN_DIR=output/regression
  else
    IN_DIR="${TMPDIR:-/tmp}/fecim-regression"
  fi
fi

OUT_JSON=docs/release/rg-val-05-release-report.json
OUT_MD=docs/release/rg-val-05-release-report.md

go run ./cmd/release-report -in "$IN_DIR" -out-json "$OUT_JSON" -out-md "$OUT_MD"

echo "[release-report] wrote: $OUT_JSON"

echo "[release-report] wrote: $OUT_MD"
