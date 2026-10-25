#!/usr/bin/env bash
set -euo pipefail

# Reference-regression runner (includes synthetic fixtures; not empirical validation).
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

unset DISPLAY
unset WAYLAND_DISPLAY

OUT_DIR="${FECIM_LITERATURE_JSON_DIR:-$ROOT/output/validation/literature}"
export FECIM_LITERATURE_JSON_DIR="$OUT_DIR"
mkdir -p "$OUT_DIR"

echo "[lit] out_dir=$OUT_DIR"

# Run suite
FECIM_LITERATURE_JSON_DIR="$OUT_DIR" go test -v ./validation/literature/... | tee /tmp/literature_validation.out

# Summary table from JSON artifacts
python3 - <<'PY'
import glob, json, os
out = os.environ.get('FECIM_LITERATURE_JSON_DIR', 'output/validation/literature')
files = sorted(glob.glob(os.path.join(out, 'module1_pe_loop_*.json')))
print('\n[lit] Summary (reference regressions, not experimental validation)')
print('material_id\tcomparison_kind\tPr_err_%\tEc_err_%\tRMSE_over_Ps\tArea_err_%\tPASS')
excluded = 0
superseded = 0
shown = 0
has_current_park = any(os.path.basename(f) == 'module1_pe_loop_park2015_synthetic_shape.json' for f in files)
for f in files:
    with open(f) as stream:
        d = json.load(stream)
    material = d['material_id']
    if material == 'bto2021_cryst11101192_hysteresis_digitized':
        excluded += 1
        continue  # Also exclude stale artifacts from before quarantine.
    if material == 'park2015_hzo_10nm' and has_current_park:
        superseded += 1
        continue
    kind = d.get('comparison_kind', 'reference-regression')
    if material.startswith('park2015'):
        kind = 'synthetic-regression'  # Includes stale legacy filenames.
    print(f"{material}\t{kind}\t{d['pr_err_pct']:.2f}\t{d['ec_err_pct']:.2f}\t{d['rmse_over_ps']:.4f}\t{d['loop_area_err_pct']:.2f}\t{d['pass']}")
    shown += 1
print(f"\n[lit] reference_artifacts={shown} experimental_validations=0 excluded={excluded} superseded={superseded} dir={out}")
PY
