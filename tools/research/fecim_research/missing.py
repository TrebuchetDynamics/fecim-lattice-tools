from __future__ import annotations

from pathlib import Path
import re

from .citations import CitationRecord, load_citation_records
from .discovery import discover_pdfs, match_pdf_to_record
from .reporting import write_content_addressed_report


STATUS_RE = re.compile(r"^status:\s*(?P<status>.+)$", re.MULTILINE)


def run_missing(root: Path) -> int:
    report = build_missing_report(root)
    report = write_content_addressed_report(
        root,
        "research/reports/missing-papers-latest.json",
        "research/reports/missing-papers",
        report,
    )

    print(
        "research missing complete: "
        f"total={report['total_records']} missing={report['missing']} "
        f"with_doi={report['missing_with_doi']} without_doi={report['missing_without_doi']} "
        f"report=research/reports/missing-papers-latest.json"
    )
    return 0


def build_missing_report(root: Path) -> dict[str, object]:
    records = load_citation_records(root)
    stored_keys = _stored_pdf_keys(root, records)
    items = [_missing_item(root, record) for key, record in sorted(records.items()) if key not in stored_keys]
    missing_with_doi = sum(1 for item in items if item["status"] == "needs_acquire")
    missing_without_doi = sum(1 for item in items if item["status"] == "missing_doi")
    return {
        "total_records": len(records),
        "stored": len(stored_keys),
        "missing": len(items),
        "missing_with_doi": missing_with_doi,
        "missing_without_doi": missing_without_doi,
        "items": items,
    }


def _stored_pdf_keys(root: Path, records: dict[str, CitationRecord]) -> set[str]:
    keys: set[str] = set()
    for key, record in records.items():
        if _explicit_pdf_exists(root, record):
            keys.add(key)
    for pdf in discover_pdfs(root, extra_paths=[]):
        match = match_pdf_to_record(pdf, records)
        if match.paper_key in records:
            keys.add(match.paper_key)
    return keys


def _explicit_pdf_exists(root: Path, record: CitationRecord) -> bool:
    pdf = record.pdf.strip()
    if not pdf or pdf.lower() == "not stored":
        return False
    path = Path(pdf)
    if path.is_absolute() or ".." in path.parts:
        return False
    return (root / path).is_file()


def _missing_item(root: Path, record: CitationRecord) -> dict[str, object]:
    has_doi = bool(record.doi)
    return {
        "paper_key": record.key,
        "citation_path": _rel(root, record.path),
        "title": record.title,
        "doi": record.doi,
        "status": "needs_acquire" if has_doi else "missing_doi",
        "last_acquisition_status": _last_acquisition_status(root, record.key),
        "download_command": f"fecim research acquire {record.key} --download" if has_doi else "",
    }


def _last_acquisition_status(root: Path, paper_key: str) -> str:
    path = root / "research" / "sources" / f"{paper_key}.acquisition.yaml"
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return ""
    match = STATUS_RE.search(text)
    if match is None:
        return ""
    return _unquote(match.group("status").strip())


def _unquote(value: str) -> str:
    if len(value) >= 2 and value[0] == value[-1] == '"':
        return value[1:-1].replace('\\"', '"')
    if len(value) >= 2 and value[0] == value[-1] == "'":
        return value[1:-1]
    return value


def _rel(root: Path, path: Path) -> str:
    try:
        return str(path.relative_to(root))
    except ValueError:
        return str(path)
