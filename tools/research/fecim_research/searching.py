import json
import re
import sys
from pathlib import Path
from typing import Any

from .indexing import collect_chunk_files


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()


def load_chunk_lookup(root: Path) -> dict[str, dict[str, object]]:
    lookup: dict[str, dict[str, object]] = {}
    for path in collect_chunk_files(root):
        with path.open(encoding="utf-8") as f:
            for line in f:
                if not line.strip():
                    continue
                record = json.loads(line)
                record["chunk_file"] = _repo_relative(root, path)
                chunk_id = record.get("id")
                if isinstance(chunk_id, str):
                    lookup[chunk_id] = record
    return lookup


def render_text_results(rows: list[dict[str, object]]) -> str:
    if not rows:
        return ""
    lines = []
    for row in rows:
        rank = row.get("rank", "")
        paper_key = row.get("paper_key", "")
        score = row.get("score", "")
        section = row.get("section", "")
        docid = row.get("docid", "")
        snippet = row.get("snippet", "")
        lines.append(f"{rank}. {paper_key} score={score} section={section} chunk={docid}")
        if snippet:
            lines.append(str(snippet))
    return "\n".join(lines) + "\n"


def _snippet(text: str, limit: int = 240) -> str:
    compact = re.sub(r"\s+", " ", text).strip()
    if len(compact) <= limit:
        return compact
    return compact[: max(0, limit - 3)].rstrip() + "..."


def _row(rank: int, score: float, docid: str, record: dict[str, Any]) -> dict[str, object]:
    contents = str(record.get("contents", ""))
    return {
        "rank": rank,
        "score": score,
        "docid": docid,
        "paper_key": record.get("paper_key", ""),
        "section": record.get("section", ""),
        "snippet": _snippet(contents),
        "chunk_file": record.get("chunk_file", ""),
        "source_parser": record.get("source_parser", ""),
    }


def run_search(root: Path, query: str, limit: int, json_output: bool) -> int:
    index_dir = root / "research" / "index" / "pyserini"
    if not index_dir.is_dir():
        print("missing BM25 index; run `fecim research index` first", file=sys.stderr)
        return 1

    try:
        from pyserini.search.lucene import LuceneSearcher
    except ImportError:
        print("Pyserini is not installed; install pyserini to run BM25 evidence search.", file=sys.stderr)
        return 1

    searcher = LuceneSearcher(str(index_dir))
    hits = searcher.search(query, k=limit)
    lookup = load_chunk_lookup(root)
    rows = []
    for rank, hit in enumerate(hits, start=1):
        docid = hit.docid
        record = lookup.get(docid, {"id": docid, "contents": ""})
        rows.append(_row(rank, hit.score, docid, record))

    if json_output:
        print(json.dumps(rows, indent=2, sort_keys=True))
    else:
        sys.stdout.write(render_text_results(rows))
    return 0
