#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

if ! command -v latex >/dev/null 2>&1; then
  echo "latex not found. Install TeX (texlive) to render equations." >&2
  exit 1
fi

if ! command -v dvisvgm >/dev/null 2>&1; then
  echo "dvisvgm not found. Install dvisvgm to render equations." >&2
  exit 1
fi

equations=(
  "frankestein"
  "preisach"
)

for name in "${equations[@]}"; do
  latex -interaction=nonstopmode -halt-on-error -output-format=dvi "${name}.tex"
  dvisvgm --no-fonts -o "${name}.svg" "${name}.dvi"
  rm -f "${name}.aux" "${name}.log" "${name}.dvi"
done
