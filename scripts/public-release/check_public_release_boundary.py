#!/usr/bin/env python3

from __future__ import annotations

import csv
import re
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
AUDIT_PATH = ROOT / "docs/public-release/THIRD_PARTY_PDF_AUDIT.csv"
DISALLOWED_PATHS = [
    "docs/archive/",
    "docs/4-research/internal-analysis/",
    "docs/4-research/transcripts/COSM_2025_AI_Hardware_Breakthrough/",
    "docs/4-research/transcripts/ironlattice-youtube-script.md",
    "docs/4-research/tour-group-ironlattice-research.md",
    "docs/4-research/superlattice-material-analysis.md",
]
BAN_RE = re.compile(r"restricted|under nda|internal repo=|internal draft", re.IGNORECASE)
SCAN_ROOTS = [
    "CLAUDE.md",
    "README.md",
    "docs/2-learn",
    "docs/3-develop",
    "docs/4-research",
    "module5-comparison",
    "module6-eda",
]


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
    if not AUDIT_PATH.exists():
        return {}
    with AUDIT_PATH.open(newline="", encoding="utf-8") as fh:
        reader = csv.DictReader(fh)
        decisions = {}
        for row in reader:
            path = (row.get("path") or "").strip()
            if path:
                decisions[path] = (row.get("decision") or "").strip()
        return decisions


def main() -> int:
    failures: list[str] = []

    for path in tracked_files():
        if any(path.startswith(disallowed) for disallowed in DISALLOWED_PATHS):
            failures.append(f"Blocked tracked path: {path}")

    pdf_decisions = load_pdf_decisions()
    for path in tracked_pdfs():
        decision = pdf_decisions.get(path)
        if decision is None:
            failures.append(f"Missing PDF audit row: {path}")
            continue
        if decision not in {"keep", "keep-with-conditions"}:
            failures.append(f"Blocked PDF decision: {path} -> {decision or '<missing>'}")

    scan = subprocess.run(
        ["rg", "-n", "-i", BAN_RE.pattern, *SCAN_ROOTS],
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
