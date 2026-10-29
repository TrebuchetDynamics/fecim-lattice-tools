#!/usr/bin/env python3

from __future__ import annotations

import re
import subprocess
from pathlib import Path

from bootstrap_pdf_audit import load_existing


ROOT = Path(__file__).resolve().parents[2]
DISALLOWED_PATHS = (
    "docs/archive/",
    "docs/superpowers/",
    "docs/internals/superpowers/",
    "docs/research/internal-analysis/",
    "docs/research/transcripts/COSM_2025_AI_Hardware_Breakthrough/",
    "docs/research/transcripts/ironlattice-youtube-script.md",
    "docs/research/tour-group-ironlattice-research.md",
    "docs/research/superlattice-material-analysis.md",
    "docs/research/papers/tourlab/",
    "docs/research/papers/DOWNLOAD_PLAN.md",
    "docs/research/tools/research_notes_final.md",
    "docs/4-research/internal-analysis/",
    "docs/4-research/transcripts/COSM_2025_AI_Hardware_Breakthrough/",
    "docs/4-research/transcripts/ironlattice-youtube-script.md",
    "docs/4-research/tour-group-ironlattice-research.md",
    "docs/4-research/superlattice-material-analysis.md",
    "docs/research-papers/tourlab/",
    "docs/4-research/papers/tourlab/",
    "docs/4-research/papers/DOWNLOAD_PLAN.md",
    "docs/4-research/opensource-tools/research_notes_final.md",
    "demo2-crossbar/pkg/_layers_experimental/kvcache_fjh.go",
    "demo2-crossbar/pkg/layers/kvcache_fjh.go",
    "demo2-inference/pkg/layers/kvcache_fjh.go",
    "module2-crossbar/pkg/_layers_experimental/kvcache_fjh.go",
    "validation/literature/_incoming/",
)
GENERATED_PATH_PREFIXES = (
    "artifacts/",
    "exports/",
    "output/",
    "recordings/",
    "screenshots/",
    "docs/1-getting-started/demo-videos/",
    "docs/guides/demo-videos/",
    "validation/output/",
    "validation/literature/output/",
    "shared/physics/output/",
    "module1-hysteresis/pkg/controller/output/",
    "module4-circuits/pkg/arraysim/output/",
    "module4-circuits/pkg/gui/output/",
)
GENERATED_EXACT_PATHS = (
    "crossbar",
    "hysteresis",
    "gen_golden_loops",
    "module1-hysteresis/hysteresis",
    "module1-hysteresis/test_output.txt",
    "module1-hysteresis/test_output_final.txt",
    "module2-crossbar/inference",
    "module3-mnist/mnist",
)
BAN_RE = re.compile(
    r"confidential|under nda|\bnda\b|private repo\s*=|fictional roleplay|"
    r"personal roleplay|contact-strategy|outreach planning|recommended email enhancement|"
    r"james\s+tour|tourlab|rice university|investor|investment pitch|investor pitch|"
    r"highly fundable|premium pricing",
    re.IGNORECASE,
)
SCAN_ROOTS = (
    "CLAUDE.md",
    "README.md",
    "docs/guides",
    "docs/modules",
    "docs/internals",
    "docs/research",
)
MODULE_SCAN_ROOTS = (
    "module1-hysteresis",
    "module2-crossbar",
    "module3-mnist",
    "module4-circuits",
    "module5-comparison",
    "module6-eda",
    "module7-docs",
)
# These files define prohibited phrases; no file or PDF gate excludes them.
PHRASE_POLICY_PATHS = (
    "docs/internals/release/PUBLIC_BOUNDARY.md",
    "docs/internals/release/REMOVAL_LOG.md",
)


def tracked_files() -> list[str]:
    result = subprocess.run(
        ["git", "ls-files"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(result.stderr.strip() or "git ls-files failed")
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]


def tracked_pdfs() -> list[str]:
    result = subprocess.run(
        ["git", "ls-files", "*.pdf"],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if result.returncode != 0:
        raise SystemExit(result.stderr.strip() or "git ls-files failed")
    return [line.strip() for line in result.stdout.splitlines() if line.strip()]


def load_pdf_decisions() -> dict[str, str]:
    return {
        path: (row.get("decision") or "").strip()
        for path, row in load_existing().items()
    }


def is_generated_output(path: str) -> bool:
    if path in GENERATED_EXACT_PATHS:
        return True
    if path.startswith(GENERATED_PATH_PREFIXES):
        return True
    if path.startswith("logs/"):
        return True
    if "/logs/" in path:
        return True
    if path.endswith(".log"):
        return True
    if path.endswith(".mp4"):
        return True
    if re.match(r"\d{4}-\d{2}-\d{2}.*\.png$", path):
        return True
    return False


def main() -> int:
    failures: list[str] = []
    tracked = tracked_files()

    for path in tracked:
        if any(path.startswith(disallowed) for disallowed in DISALLOWED_PATHS):
            failures.append(f"Blocked tracked path: {path}")
        if is_generated_output(path):
            failures.append(f"Tracked generated output: {path}")

    pdf_decisions = load_pdf_decisions()
    for path in tracked_pdfs():
        decision = pdf_decisions.get(path)
        if decision is None:
            failures.append(f"Missing PDF audit row: {path}")
            continue
        if decision not in {"keep", "keep-with-conditions"}:
            failures.append(f"Blocked PDF decision: {path} -> {decision or '<missing>'}")

    scan_roots = [*SCAN_ROOTS, *(
        path for path in MODULE_SCAN_ROOTS if (ROOT / path).exists()
    )]
    policy_globs = [arg for path in PHRASE_POLICY_PATHS for arg in ("--glob", f"!/{path}")]
    scan = subprocess.run(
        ["rg", "-n", "-i", *policy_globs, BAN_RE.pattern, *scan_roots],
        cwd=ROOT,
        capture_output=True,
        text=True,
        check=False,
    )
    if scan.returncode == 0:
        failures.append("Blocked phrases found:\n" + scan.stdout.strip())
    elif scan.returncode != 1:
        failures.append(f"rg failed with exit code {scan.returncode}:\n{scan.stderr.strip()}")

    if failures:
        for failure in failures:
            print(failure)
        return 1

    print("Public release boundary checks passed.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
