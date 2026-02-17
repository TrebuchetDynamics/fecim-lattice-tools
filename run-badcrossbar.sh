#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENV="$SCRIPT_DIR/opensource/badcrossbar/.venv"

source "$VENV/bin/activate"
exec python3 "$@"
