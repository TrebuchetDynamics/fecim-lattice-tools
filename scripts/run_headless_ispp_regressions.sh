#!/usr/bin/env bash
set -euo pipefail

# Required: this regression lane is fully headless (no display stack).
if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then
  echo "[regression] ERROR: DISPLAY/WAYLAND_DISPLAY detected; run this lane fully headless." >&2
  echo "[regression] Hint: unset DISPLAY WAYLAND_DISPLAY" >&2
  exit 1
fi

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${FECIM_REGRESSION_JSON_DIR:-$REPO_ROOT/output/regression}"

mkdir -p "$OUT_DIR"
export FECIM_REGRESSION_JSON_DIR="$OUT_DIR"

echo "[regression] output dir: $FECIM_REGRESSION_JSON_DIR"
echo "[regression] running Preisach + LK headless WRD/ISPP suites"

go test ./module1-hysteresis/pkg/controller \
  -run 'TestHeadlessRegression_WRD_ISPP_(Preisach|LK)$' \
  -count=1 -v | tee "$OUT_DIR/test.log"

echo "[regression] per-material verdicts:" 
grep -E 'VERDICT material=' -n "$OUT_DIR/test.log" || true

# RG-VAL-03: enforce required-material verdict coverage.
PROFILE="${FECIM_MATERIAL_PROFILE:-pr}"
REQUIRED_MATS=$(go run ./cmd/material-profile -profile "$PROFILE" -sep ' ')

missing=0
for m in $REQUIRED_MATS; do
  if ! grep -q "VERDICT material=$m" "$OUT_DIR/test.log"; then
    echo "[regression] ERROR: missing required material verdict: $m (profile=$PROFILE)" >&2
    missing=1
  fi
done

if [[ "$missing" -ne 0 ]]; then
  echo "[regression] FAIL: required-material verdicts incomplete (profile=$PROFILE)" >&2
  exit 1
fi

echo "[regression] PASS: required-material verdicts complete (profile=$PROFILE)"

echo "[regression] summaries:"
ls -1 "$OUT_DIR"/*.json
