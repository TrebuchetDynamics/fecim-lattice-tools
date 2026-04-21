#!/usr/bin/env python3

from __future__ import annotations

import csv
import subprocess
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
AUDIT_PATH = ROOT / "docs/public-release/THIRD_PARTY_PDF_AUDIT.csv"
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
        return {row["path"]: row for row in reader if row.get("path")}


def build_row(path: str, existing: dict[str, str] | None = None) -> dict[str, str]:
    row = {field: "" for field in FIELDS}
    row["path"] = path
    if existing:
        for field in FIELDS:
            if existing.get(field):
                row[field] = existing[field]
    if not row["title"]:
        row["title"] = Path(path).stem
    if not row["decision"]:
        row["decision"] = DEFAULT_DECISION
    if not row["notes"]:
        row["notes"] = DEFAULT_NOTES
    return row


def main() -> int:
    existing = load_existing()
    rows = [build_row(path, existing.get(path)) for path in tracked_pdfs()]
    AUDIT_PATH.parent.mkdir(parents=True, exist_ok=True)
    with AUDIT_PATH.open("w", newline="", encoding="utf-8") as fh:
        writer = csv.DictWriter(fh, fieldnames=FIELDS)
        writer.writeheader()
        writer.writerows(rows)
    print(f"Wrote {len(rows)} PDF audit rows to {AUDIT_PATH.relative_to(ROOT)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
