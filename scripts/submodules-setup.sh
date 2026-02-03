#!/usr/bin/env bash
set -euo pipefail

# Configure this repo to avoid fetching nested submodules that are optional.
# Run this once per machine after cloning.

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

# Avoid recursive submodule updates for this repo

git config --local submodule.recurse false
git config --local fetch.recurseSubmodules false

# Disable nested submodules inside cross-sim
if [[ -d opensource/crossbar/cross-sim ]]; then
  git -C opensource/crossbar/cross-sim config submodule.data.update none
  git -C opensource/crossbar/cross-sim config submodule.pretrained_models.update none
  # Deinit only if these were initialized (silence warnings if not present)
  git -C opensource/crossbar/cross-sim submodule deinit -f applications/dnn/data applications/dnn/pretrained_models 2>/dev/null || true
fi

# Disable nested submodules inside ferret
if [[ -d opensource/hysteresys/ferret ]]; then
  git -C opensource/hysteresys/ferret config submodule.ScalFMM.update none
  # Deinit only if initialized (silence warnings if not present)
  git -C opensource/hysteresys/ferret submodule deinit -f ScalFMM 2>/dev/null || true
fi

cat <<'MSG'
Submodule setup complete.

Use:
  git submodule update --init --depth 1

Avoid:
  git submodule update --init --recursive
  git pull --recurse-submodules
MSG
