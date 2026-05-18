import hashlib
import json
import shutil
import subprocess
import sys
from pathlib import Path


def collect_chunk_files(root: Path) -> list[Path]:
    chunk_dir = root / "research" / "chunks"
    if not chunk_dir.is_dir():
        return []
    return sorted(chunk_dir.glob("*.jsonl"))


def _sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def _repo_relative(root: Path, path: Path) -> str:
    try:
        return path.resolve().relative_to(root.resolve()).as_posix()
    except ValueError:
        return path.as_posix()


def write_index_manifest(
    root: Path,
    backend: str,
    inputs: list[Path],
    semantic: bool,
    embedding_model: str,
) -> Path:
    manifest_path = root / "research" / "manifests" / "index-latest.json"
    manifest_path.parent.mkdir(parents=True, exist_ok=True)
    data = {
        "backend": backend,
        "semantic": semantic,
        "embedding_model": embedding_model,
        "inputs": [
            {
                "path": _repo_relative(root, path),
                "sha256": _sha(path),
            }
            for path in sorted(inputs)
        ],
        "pyserini_index": "research/index/pyserini",
    }
    manifest_path.write_text(json.dumps(data, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return manifest_path


def run_index(root: Path, semantic: bool, embedding_model: str) -> int:
    if semantic:
        print("semantic indexing is not implemented in the retrieval MVP", file=sys.stderr)
        return 2

    chunks = collect_chunk_files(root)
    if not chunks:
        print("no chunk files found under research/chunks", file=sys.stderr)
        return 1

    index_dir = root / "research" / "index" / "pyserini"
    if index_dir.exists():
        shutil.rmtree(index_dir)
    index_dir.parent.mkdir(parents=True, exist_ok=True)

    command = [
        sys.executable,
        "-m",
        "pyserini.index.lucene",
        "--collection",
        "JsonCollection",
        "--input",
        str(root / "research" / "chunks"),
        "--index",
        str(index_dir),
        "--generator",
        "DefaultLuceneDocumentGenerator",
        "--threads",
        "1",
        "--storePositions",
        "--storeDocvectors",
        "--storeRaw",
    ]
    result = subprocess.run(command, check=False)
    if result.returncode != 0:
        print(
            "Pyserini indexing failed; install pyserini and a compatible Java runtime, then rerun `fecim research index`.",
            file=sys.stderr,
        )
        return result.returncode

    write_index_manifest(root, "pyserini", chunks, semantic=False, embedding_model=embedding_model)
    print(f"indexed {len(chunks)} chunk file(s) into research/index/pyserini")
    return 0
