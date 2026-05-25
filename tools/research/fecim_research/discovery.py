from dataclasses import dataclass
from pathlib import Path
import hashlib
import re

from .citations import CitationRecord


DEFAULT_PDF_GLOBS = (
    "docs/4-research/papers/**/*.pdf",
    "research/papers/**/*.pdf",
    "citations/pdfs/**/*.pdf",
)


@dataclass(frozen=True)
class DiscoveredPDF:
    path: Path
    sha256: str
    size: int
    duplicate_of: Path | None = None


@dataclass(frozen=True)
class PDFMatch:
    status: str
    paper_key: str | None
    method: str
    confidence: float


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def discover_pdfs(root: Path, extra_paths: list[Path]) -> list[DiscoveredPDF]:
    paths: set[Path] = set()
    for pattern in DEFAULT_PDF_GLOBS:
        paths.update(root.glob(pattern))
        if pattern.endswith(".pdf"):
            paths.update(root.glob(pattern[:-4] + ".PDF"))
    for extra in extra_paths:
        base = extra if extra.is_absolute() else root / extra
        if base.is_file() and base.suffix.lower() == ".pdf":
            paths.add(base)
        elif base.is_dir():
            paths.update(path for path in base.rglob("*") if path.suffix.lower() == ".pdf")
    out: list[DiscoveredPDF] = []
    canonical_by_sha: dict[str, Path] = {}
    for path in sorted(paths):
        if not path.is_file():
            continue
        digest = sha256_file(path)
        duplicate_of = canonical_by_sha.get(digest)
        if duplicate_of is None:
            canonical_by_sha[digest] = path
        out.append(
            DiscoveredPDF(
                path=path,
                sha256=digest,
                size=path.stat().st_size,
                duplicate_of=duplicate_of,
            )
        )
    return out


def match_pdf_to_record(pdf: DiscoveredPDF, records: dict[str, CitationRecord]) -> PDFMatch:
    stem = pdf.path.stem.lower()
    normalized_stem = _normalize_key(pdf.path.stem)
    for key in sorted(records):
        if _path_matches_record_pdf(pdf.path, records[key].pdf):
            return PDFMatch(status="matched", paper_key=key, method="citation_pdf", confidence=1.0)
    for key in sorted(records):
        if stem == key.lower():
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
    for key in sorted(records):
        if normalized_stem == _normalize_key(key):
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
    for key in sorted(records, key=lambda item: (-len(item), item)):
        if key.lower() in stem:
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
        if _normalize_key(key) in normalized_stem:
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
    return PDFMatch(status="unmatched", paper_key=None, method="none", confidence=0.0)


def _normalize_key(value: str) -> str:
    return re.sub(r"[^a-z0-9]+", "_", value.lower()).strip("_")


def _path_matches_record_pdf(path: Path, record_pdf: str) -> bool:
    pdf_path = record_pdf.strip()
    if not pdf_path or pdf_path.lower() == "not stored":
        return False
    normalized = Path(pdf_path).as_posix()
    candidate = path.as_posix()
    return candidate == normalized or candidate.endswith("/" + normalized)
