#!/bin/bash
set -euo pipefail
echo "=== Module 4 Fast PR Gate ==="
env -u DISPLAY -u WAYLAND_DISPLAY \
  go test -tags ci -count=1 -short -timeout 10m \
  -run 'Kirchhoff|CurrentValidation|Thermodynamics|Pattern|MVM|ReadMargin|INL|DNL' \
  ./module4-circuits/pkg/arraysim/... ./module4-circuits/pkg/gui/... ./shared/peripherals/...
echo "✓ Module 4 fast PR gate passed"
