#!/usr/bin/env python3

from __future__ import annotations

import csv
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
AUDIT_PATH = ROOT / "docs/internals/release/THIRD_PARTY_PDF_AUDIT.csv"
FIELDS = [
    "path",
    "title",
    "source_url",
    "evidence_url",
    "rights_basis",
    "obligations",
    "decision",
    "notes",
]
DEFAULT_DECISION = "replace-with-link"
DEFAULT_NOTES = "Explicit redistribution evidence is required before changing this row to keep."


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


def load_existing() -> dict[str, dict[str, str]]:
    if not AUDIT_PATH.exists():
        return {}
    with AUDIT_PATH.open(newline="", encoding="utf-8") as fh:
        reader = csv.DictReader(fh)
        existing = {}
        for row in reader:
            path = row.get("path") or ""
            if not path:
                continue
            # Only this exact historical prefix has an approved migration.
            if path.startswith("docs/4-research/"):
                path = "docs/research/" + path[len("docs/4-research/"):]
            row["path"] = path
            if path in existing and existing[path] != row:
                raise SystemExit(f"Conflicting PDF audit rows: {path}")
            existing[path] = row
        return existing


def build_row(path: str) -> dict[str, str]:
    row = {field: "" for field in FIELDS}
    row["path"] = path
    if not row["title"]:
        row["title"] = Path(path).stem
    if not row["decision"]:
        row["decision"] = DEFAULT_DECISION
    if not row["notes"]:
        row["notes"] = DEFAULT_NOTES
    return row


def main() -> int:
    existing = load_existing()
    # Retain historical records and every existing cell, including blank cells.
    for path in tracked_pdfs():
        if path not in existing:
            existing[path] = build_row(path)
    rows = list(existing.values())
    AUDIT_PATH.parent.mkdir(parents=True, exist_ok=True)
    with AUDIT_PATH.open("w", newline="", encoding="utf-8") as fh:
        writer = csv.DictWriter(fh, fieldnames=FIELDS)
        writer.writeheader()
        writer.writerows(rows)
    print(f"Wrote {len(rows)} PDF audit rows to {AUDIT_PATH.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
