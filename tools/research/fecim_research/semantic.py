from __future__ import annotations

import hashlib
import json
import math
import re
from pathlib import Path
from typing import Any


DEFAULT_EMBEDDING_MODEL = "fecim-hashing-bow-v1"
EMBEDDING_PROVIDER = "local-hashing"
VECTOR_DIMENSION = 64
VECTOR_CACHE = "research/index/lancedb/chunks.jsonl"


def effective_embedding_model(embedding_model: str) -> str:
    return embedding_model.strip() or DEFAULT_EMBEDDING_MODEL


def embed_text(text: str, dimension: int = VECTOR_DIMENSION) -> list[float]:
    vector = [0.0] * dimension
    for token in _tokens(text):
        digest = hashlib.sha256(token.encode("utf-8")).digest()
        bucket = int.from_bytes(digest[:4], "big") % dimension
        vector[bucket] += 1.0
    norm = math.sqrt(sum(value * value for value in vector))
    if norm == 0:
        return vector
    return [round(value / norm, 8) for value in vector]


def build_vector_records(root: Path, chunk_files: list[Path], embedding_model: str) -> list[dict[str, object]]:
    model = effective_embedding_model(embedding_model)
    records: list[dict[str, object]] = []
    for path in sorted(chunk_files):
        for chunk in _read_jsonl(path):
            chunk_id = chunk.get("id")
            if not isinstance(chunk_id, str) or not chunk_id:
                continue
            contents = str(chunk.get("contents", ""))
            text = " ".join(str(chunk.get(key, "")) for key in ["paper_key", "section", "contents"])
            records.append(
                {
                    "id": chunk_id,
                    "paper_key": chunk.get("paper_key", ""),
                    "section": chunk.get("section", ""),
                    "contents": contents,
                    "snippet": _snippet(contents),
                    "chunk_file": _repo_relative(root, path),
                    "source_parser": chunk.get("source_parser", ""),
                    "source_path": chunk.get("source_path", ""),
                    "section_number": chunk.get("section_number"),
                    "chunk_number": chunk.get("chunk_number"),
                    "page_start": chunk.get("page_start"),
                    "page_end": chunk.get("page_end"),
                    "char_start": chunk.get("char_start"),
                    "char_end": chunk.get("char_end"),
                    "sha256": chunk.get("sha256", ""),
                    "embedding_model": model,
                    "embedding_provider": EMBEDDING_PROVIDER,
                    "vector": embed_text(text),
                }
            )
    return records


def write_vector_cache(root: Path, records: list[dict[str, object]]) -> Path:
    path = root / VECTOR_CACHE
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        for record in records:
            f.write(json.dumps(record, ensure_ascii=False, sort_keys=True) + "\n")
    return path


def load_vector_cache(root: Path) -> list[dict[str, Any]]:
    path = root / VECTOR_CACHE
    return [record for record in _read_jsonl(path) if isinstance(record.get("vector"), list)]


def search_vector_records(records: list[dict[str, Any]], query: str, limit: int, embedding_model: str) -> list[tuple[float, str, dict[str, Any]]]:
    query_vector = embed_text(query)
    scored: list[tuple[float, str, dict[str, Any]]] = []
    for record in records:
        docid = str(record.get("id", ""))
        if not docid:
            continue
        score = _dot(query_vector, record.get("vector", []))
        if score > 0:
            scored.append((score, docid, record))
    scored.sort(key=lambda item: (-item[0], str(item[2].get("paper_key", "")), item[1]))
    return scored[:limit]


def _read_jsonl(path: Path) -> list[dict[str, Any]]:
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError:
        return []
    records: list[dict[str, Any]] = []
    for line in lines:
        if not line.strip():
            continue
        try:
            data = json.loads(line)
        except json.JSONDecodeError:
            continue
        if isinstance(data, dict):
            records.append(data)
    return records


def _tokens(text: str) -> list[str]:
    return re.findall(r"[a-z0-9]+", text.lower())


def _dot(left: list[float], right: object) -> float:
    if not isinstance(right, list):
        return 0.0
    total = 0.0
    for a, b in zip(left, right):
        try:
            total += float(a) * float(b)
        except (TypeError, ValueError):
            continue
    return total


def _snippet(text: str, limit: int = 240) -> str:
    compact = re.sub(r"\s+", " ", text).strip()
    if len(compact) <= limit:
        return compact
    return compact[: max(0, limit - 3)].rstrip() + "..."


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()
