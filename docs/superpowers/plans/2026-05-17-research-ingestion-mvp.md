# Research Ingestion MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the first file-first research retrieval slice: discover local papers, write git-trackable metadata/parse/chunk artifacts, build a BM25 search cache, and expose it through `fecim research ingest/index/search`.

**Architecture:** The Go CLI adds a `research` subcommand that shells out to repo-local Python tooling. Python owns the research pipeline and writes committed ledger artifacts under `research/`; Pyserini indexes are rebuildable caches under `research/index/`. This plan intentionally implements retrieval first and leaves claim-audit CI for a later plan.

**Tech Stack:** Go standard library CLI, Python 3 standard library for ingestion/chunking, optional GROBID HTTP service, optional Marker command, Pyserini for Lucene BM25 indexing/search when installed.

---

## Scope

This plan implements the retrieval MVP from [research ingestion design](../specs/2026-05-17-research-ingestion-design.md):

- `fecim research ingest`
- `fecim research index`
- `fecim research search "query"`
- `research/` file ledger layout
- citation-key matching from existing `citations/papers/*.md`
- quarantine for unmatched PDFs
- deterministic chunk JSONL
- Pyserini-backed BM25 cache, with clear failure when Pyserini is missing

Deferred to a follow-up plan:

- `fecim research cite`
- `fecim research claim-scan`
- `fecim research audit`
- `fecim research graph`
- `make research-audit`
- legal OA download/acquisition
- semantic LanceDB index
- machine claim extraction

## File Structure

Create:

- `research/README.md` — explains ledger vs cache boundaries.
- `research/.gitignore` — ignores rebuildable caches while keeping the ledger trackable.
- `research/papers/.gitkeep` — user PDF drop zone.
- `research/sources/.gitkeep` — normalized source metadata.
- `research/parsed/.gitkeep` — parser outputs.
- `research/chunks/.gitkeep` — chunk JSONL files.
- `research/extracted/.gitkeep` — generated claim candidates for later work.
- `research/graphs/.gitkeep` — graph exports for later work.
- `research/manifests/.gitkeep` — run/index manifests.
- `research/reports/.gitkeep` — unmatched/duplicate/failure reports.
- `research/index/.gitkeep` — preserves the cache parent while cache contents stay ignored.
- `citations/claims/README.md` — describes the reviewed claim registry.
- `citations/claims/.gitkeep` — preserves directory.
- `cmd/fecim-lattice-tools/research_subcommand.go` — Go wrapper for Python research tool.
- `cmd/fecim-lattice-tools/research_subcommand_test.go` — Go command dispatch tests.
- `tools/research/research_cli.py` — Python entrypoint called by Go.
- `tools/research/fecim_research/__init__.py` — package marker.
- `tools/research/fecim_research/cli.py` — argparse command routing.
- `tools/research/fecim_research/paths.py` — repo path resolution and directory constants.
- `tools/research/fecim_research/citations.py` — citation record loading and key extraction.
- `tools/research/fecim_research/discovery.py` — PDF discovery, hashing, duplicate detection, matching.
- `tools/research/fecim_research/yamlio.py` — small deterministic YAML emitter for known ledger schemas.
- `tools/research/fecim_research/parsing.py` — GROBID/Marker wrappers and parse status records.
- `tools/research/fecim_research/chunking.py` — deterministic section/chunk JSONL writer.
- `tools/research/fecim_research/ingest.py` — ingest orchestration.
- `tools/research/fecim_research/indexing.py` — Pyserini index command wrapper and manifest writer.
- `tools/research/fecim_research/searching.py` — Pyserini search wrapper and evidence output.
- `tools/research/tests/test_cli.py`
- `tools/research/tests/test_discovery.py`
- `tools/research/tests/test_chunking.py`
- `tools/research/tests/test_ingest.py`
- `tools/research/tests/test_indexing.py`
- `tools/research/tests/test_searching.py`

Modify:

- `cmd/fecim-lattice-tools/subcommands.go` — dispatch `research` and advertise it in help.
- `Makefile` — add a lightweight `test-research` target for Python unit tests and Go command tests.

## Task 1: Add Research Ledger Layout

**Files:**

- Create: `research/README.md`
- Create: `research/.gitignore`
- Create: `research/papers/.gitkeep`
- Create: `research/sources/.gitkeep`
- Create: `research/parsed/.gitkeep`
- Create: `research/chunks/.gitkeep`
- Create: `research/extracted/.gitkeep`
- Create: `research/graphs/.gitkeep`
- Create: `research/manifests/.gitkeep`
- Create: `research/reports/.gitkeep`
- Create: `research/index/.gitkeep`
- Create: `citations/claims/README.md`
- Create: `citations/claims/.gitkeep`
- Test: `cmd/fecim-lattice-tools/research_layout_test.go`

- [ ] **Step 1: Write the failing layout contract test**

Create `cmd/fecim-lattice-tools/research_layout_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResearchLedgerLayoutExists(t *testing.T) {
	root := repoRoot()
	requiredDirs := []string{
		"research/papers",
		"research/sources",
		"research/parsed",
		"research/chunks",
		"research/extracted",
		"research/graphs",
		"research/manifests",
		"research/reports",
		"citations/claims",
	}
	for _, dir := range requiredDirs {
		info, err := os.Stat(filepath.Join(root, dir))
		if err != nil {
			t.Fatalf("expected %s to exist: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestResearchGitignoreKeepsLedgerAndIgnoresCaches(t *testing.T) {
	root := repoRoot()
	body, err := os.ReadFile(filepath.Join(root, "research/.gitignore"))
	if err != nil {
		t.Fatalf("read research/.gitignore: %v", err)
	}
	text := string(body)
	required := []string{
		"/index/pyserini/",
		"/index/lancedb/",
		"/index/models/",
		"/.cache/",
		"!/index/.gitkeep",
	}
	for _, phrase := range required {
		if !strings.Contains(text, phrase) {
			t.Fatalf("research/.gitignore must contain %q", phrase)
		}
	}
}
```

- [ ] **Step 2: Run the layout test and confirm RED**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run 'TestResearchLedgerLayoutExists|TestResearchGitignoreKeepsLedgerAndIgnoresCaches' -count=1
```

Expected: FAIL because `research/` and `citations/claims/` do not exist yet.

- [ ] **Step 3: Add the ledger directories and README files**

Create `research/README.md`:

```markdown
# Research Ledger

This directory stores git-trackable research ingestion artifacts.

Canonical reviewed claims and paper records remain under `citations/`.
This directory stores the retrieval ledger: normalized source metadata,
parser outputs, chunks, manifests, reports, and rebuildable search cache
manifests.

Tracked:

- `sources/`
- `parsed/`
- `chunks/`
- `extracted/`
- `graphs/`
- `manifests/`
- `reports/`

Ignored rebuildable caches:

- `index/pyserini/`
- `index/lancedb/`
- `index/models/`
- `.cache/`
```

Create `research/.gitignore`:

```gitignore
/index/pyserini/
/index/lancedb/
/index/models/
/.cache/

!/index/
!/index/.gitkeep
!/index/README.md
```

Create `citations/claims/README.md`:

```markdown
# Reviewed Claim Registry

Reviewed scientific claims live here as one YAML file per claim ID.

Generated claim candidates belong under `research/extracted/` and are not
citable facts until a human-reviewed record exists in this directory.
```

Create `.gitkeep` files in every new empty directory listed for this task.

- [ ] **Step 4: Run the layout test and confirm GREEN**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run 'TestResearchLedgerLayoutExists|TestResearchGitignoreKeepsLedgerAndIgnoresCaches' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 1**

Run:

```bash
git add research citations/claims cmd/fecim-lattice-tools/research_layout_test.go
git commit -m "feat(research): add file-first ledger layout"
```

Record RED and GREEN command summaries in the commit notes or implementation handoff.

## Task 2: Add Go `research` Subcommand Wrapper

**Files:**

- Create: `cmd/fecim-lattice-tools/research_subcommand.go`
- Create: `cmd/fecim-lattice-tools/research_subcommand_test.go`
- Modify: `cmd/fecim-lattice-tools/subcommands.go`

- [ ] **Step 1: Write failing Go dispatch tests**

Create `cmd/fecim-lattice-tools/research_subcommand_test.go`:

```go
package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestDispatchResearchSubcommandUsesResearchRunner(t *testing.T) {
	var got []string
	previous := researchRunner
	researchRunner = func(args []string) error {
		got = append([]string(nil), args...)
		return nil
	}
	defer func() { researchRunner = previous }()

	if err := dispatchSubcommand([]string{"research", "search", "HZO coercive field"}); err != nil {
		t.Fatalf("dispatch research: %v", err)
	}
	want := []string{"search", "HZO coercive field"}
	if len(got) != len(want) {
		t.Fatalf("research runner args len=%d want=%d args=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("arg %d=%q want %q", i, got[i], want[i])
		}
	}
}

func TestDispatchResearchSubcommandPropagatesRunnerError(t *testing.T) {
	previous := researchRunner
	researchRunner = func(args []string) error {
		return errors.New("research tool failed")
	}
	defer func() { researchRunner = previous }()

	err := dispatchSubcommand([]string{"research", "ingest"})
	if err == nil || !strings.Contains(err.Error(), "research tool failed") {
		t.Fatalf("expected runner error, got %v", err)
	}
}

func TestRootUsageListsResearchSubcommand(t *testing.T) {
	var buf bytes.Buffer
	printRootUsage(&buf)
	text := buf.String()
	if !strings.Contains(text, "research") {
		t.Fatalf("root usage must mention research subcommand:\n%s", text)
	}
	if !strings.Contains(text, "research ingest") {
		t.Fatalf("root usage must include research example:\n%s", text)
	}
}
```

- [ ] **Step 2: Run the Go research dispatch tests and confirm RED**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run 'TestDispatchResearchSubcommand|TestRootUsageListsResearchSubcommand' -count=1
```

Expected: FAIL because `researchRunner` and `research` dispatch do not exist.

- [ ] **Step 3: Implement the Go wrapper**

Create `cmd/fecim-lattice-tools/research_subcommand.go`:

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var researchRunner = runResearchTool

func runResearchSubcommand(args []string) error {
	if len(args) == 0 {
		args = []string{"--help"}
	}
	return researchRunner(args)
}

func runResearchTool(args []string) error {
	python := os.Getenv("FECIM_RESEARCH_PYTHON")
	if python == "" {
		python = "python3"
	}
	root := filepath.Clean(filepath.Join(".", "tools", "research", "research_cli.py"))
	cmdArgs := append([]string{root}, args...)
	cmd := exec.Command(python, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("research tool: %w", err)
	}
	return nil
}
```

Modify `dispatchSubcommand` in `cmd/fecim-lattice-tools/subcommands.go`:

```go
	case "research":
		return runResearchSubcommand(args[1:])
```

Modify `printRootUsage` in `cmd/fecim-lattice-tools/subcommands.go` to include:

```go
	fmt.Fprintln(w, "  research           paper ingestion, indexing, and evidence search")
```

Add examples:

```go
	fmt.Fprintln(w, "  fecim-lattice-tools research ingest")
	fmt.Fprintln(w, "  fecim-lattice-tools research search \"HZO coercive field Preisach\"")
```

- [ ] **Step 4: Run the Go research dispatch tests and confirm GREEN**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run 'TestDispatchResearchSubcommand|TestRootUsageListsResearchSubcommand' -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit Task 2**

Run:

```bash
git add cmd/fecim-lattice-tools/research_subcommand.go cmd/fecim-lattice-tools/research_subcommand_test.go cmd/fecim-lattice-tools/subcommands.go
git commit -m "feat(research): add CLI subcommand wrapper"
```

## Task 3: Add Python Research CLI Skeleton

**Files:**

- Create: `tools/research/research_cli.py`
- Create: `tools/research/fecim_research/__init__.py`
- Create: `tools/research/fecim_research/cli.py`
- Create: `tools/research/fecim_research/paths.py`
- Create: `tools/research/tests/test_cli.py`
- Modify: `Makefile`

- [ ] **Step 1: Write failing Python CLI tests**

Create `tools/research/tests/test_cli.py`:

```python
import io
import unittest
from contextlib import redirect_stdout

from fecim_research.cli import main


class CLITest(unittest.TestCase):
    def test_help_lists_core_commands(self):
        out = io.StringIO()
        with self.assertRaises(SystemExit) as ctx, redirect_stdout(out):
            main(["--help"])
        self.assertEqual(ctx.exception.code, 0)
        text = out.getvalue()
        self.assertIn("ingest", text)
        self.assertIn("index", text)
        self.assertIn("search", text)

    def test_unknown_command_fails(self):
        with self.assertRaises(SystemExit) as ctx:
            main(["unknown"])
        self.assertNotEqual(ctx.exception.code, 0)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run the Python CLI tests and confirm RED**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_cli.py' -v
```

Expected: FAIL because `fecim_research.cli` does not exist.

- [ ] **Step 3: Implement the Python CLI skeleton**

Create `tools/research/research_cli.py`:

```python
#!/usr/bin/env python3
from fecim_research.cli import main


if __name__ == "__main__":
    raise SystemExit(main())
```

Create `tools/research/fecim_research/__init__.py`:

```python
"""File-first research ingestion tools for FeCIM Lattice Tools."""
```

Create `tools/research/fecim_research/paths.py`:

```python
from pathlib import Path


def repo_root(start: Path | None = None) -> Path:
    current = (start or Path.cwd()).resolve()
    for candidate in [current, *current.parents]:
        if (candidate / "go.mod").exists() and (candidate / "citations").is_dir():
            return candidate
    raise RuntimeError("could not locate repository root")


def research_root(root: Path) -> Path:
    return root / "research"
```

Create `tools/research/fecim_research/cli.py`:

```python
import argparse
from pathlib import Path

from .paths import repo_root


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="fecim research")
    parser.add_argument("--repo-root", type=Path, default=None)
    sub = parser.add_subparsers(dest="command", required=True)

    ingest = sub.add_parser("ingest", help="discover, parse, and chunk local papers")
    ingest.add_argument("paths", nargs="*", help="optional extra PDF roots")

    index = sub.add_parser("index", help="build rebuildable search indexes")
    index.add_argument("--semantic", action="store_true", help="build local semantic index")
    index.add_argument("--embedding-model", default="", help="local embedding model name")

    search = sub.add_parser("search", help="search evidence chunks")
    search.add_argument("query", help="search query")
    search.add_argument("--json", action="store_true", help="emit JSON results")
    search.add_argument("--limit", type=int, default=10)

    return parser


def main(argv: list[str] | None = None) -> int:
    parser = build_parser()
    args = parser.parse_args(argv)
    root = args.repo_root.resolve() if args.repo_root else repo_root()

    if args.command == "ingest":
        from .ingest import run_ingest

        return run_ingest(root=root, extra_paths=[Path(p) for p in args.paths])
    if args.command == "index":
        from .indexing import run_index

        return run_index(root=root, semantic=args.semantic, embedding_model=args.embedding_model)
    if args.command == "search":
        from .searching import run_search

        return run_search(root=root, query=args.query, limit=args.limit, json_output=args.json)
    parser.error(f"unknown command {args.command}")
    return 2
```

- [ ] **Step 4: Add `test-research` Make target**

Modify `Makefile` `.PHONY` line to include `test-research`.

Add:

```make
test-research:
	PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -v
	CGO_ENABLED=0 $(GO) test ./cmd/fecim-lattice-tools -run 'TestResearch|TestDispatchResearch|TestRootUsageListsResearch' -count=1
```

- [ ] **Step 5: Run the Python CLI tests and confirm GREEN**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_cli.py' -v
```

Expected: PASS.

- [ ] **Step 6: Run Go wrapper tests after Python entrypoint exists**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run 'TestDispatchResearchSubcommand|TestRootUsageListsResearchSubcommand' -count=1
```

Expected: PASS.

- [ ] **Step 7: Commit Task 3**

Run:

```bash
git add Makefile tools/research
git commit -m "feat(research): add Python research CLI skeleton"
```

## Task 4: Implement PDF Discovery, Hashing, And Citation-Key Matching

**Files:**

- Create: `tools/research/fecim_research/citations.py`
- Create: `tools/research/fecim_research/discovery.py`
- Create: `tools/research/fecim_research/yamlio.py`
- Create: `tools/research/tests/test_discovery.py`

- [ ] **Step 1: Write failing discovery tests**

Create `tools/research/tests/test_discovery.py`:

```python
import tempfile
import unittest
from pathlib import Path

from fecim_research.citations import load_citation_records
from fecim_research.discovery import discover_pdfs, match_pdf_to_record, sha256_file


class DiscoveryTest(unittest.TestCase):
    def test_discovers_pdf_roots_and_hashes_files(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF-1.4\nfixture\n")

            found = discover_pdfs(root, extra_paths=[])
            self.assertEqual(len(found), 1)
            self.assertEqual(found[0].path, pdf)
            self.assertEqual(found[0].sha256, sha256_file(pdf))

    def test_loads_existing_citation_keys_from_markdown(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text(
                "# Park 2015\n\n"
                "**Key:** `park2015_advmat_hzo`\n"
                "**DOI:** `10.1002/adma.201404531`\n"
                "**Title:** `Ferroelectric HZO`\n",
                encoding="utf-8",
            )
            records = load_citation_records(root)
            self.assertIn("park2015_advmat_hzo", records)
            self.assertEqual(records["park2015_advmat_hzo"].doi, "10.1002/adma.201404531")

    def test_matches_pdf_filename_to_existing_citation_key(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            paper = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            paper.parent.mkdir(parents=True)
            paper.write_text("**Key:** `park2015_advmat_hzo`\n", encoding="utf-8")
            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            records = load_citation_records(root)
            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records)
            self.assertEqual(match.paper_key, "park2015_advmat_hzo")
            self.assertEqual(match.status, "matched")
            self.assertEqual(match.method, "filename")

    def test_unmatched_pdf_is_quarantined(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "research" / "papers" / "unknown.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")

            found = discover_pdfs(root, extra_paths=[])[0]
            match = match_pdf_to_record(found, records={})
            self.assertEqual(match.status, "unmatched")
            self.assertIsNone(match.paper_key)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run discovery tests and confirm RED**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_discovery.py' -v
```

Expected: FAIL because discovery modules do not exist.

- [ ] **Step 3: Implement citation record loading**

Create `tools/research/fecim_research/citations.py`:

```python
from dataclasses import dataclass
from pathlib import Path
import re


FIELD_RE = re.compile(r"^\*\*(?P<name>[^*]+):\*\*\s*`?(?P<value>[^`\n]+)`?", re.MULTILINE)


@dataclass(frozen=True)
class CitationRecord:
    key: str
    path: Path
    title: str = ""
    doi: str = ""
    arxiv_id: str = ""


def _fields(text: str) -> dict[str, str]:
    out: dict[str, str] = {}
    for match in FIELD_RE.finditer(text):
        out[match.group("name").strip().lower()] = match.group("value").strip()
    return out


def load_citation_records(root: Path) -> dict[str, CitationRecord]:
    records: dict[str, CitationRecord] = {}
    papers_dir = root / "citations" / "papers"
    if not papers_dir.exists():
        return records
    for path in sorted(papers_dir.glob("*.md")):
        text = path.read_text(encoding="utf-8", errors="replace")
        fields = _fields(text)
        key = fields.get("key") or path.stem
        records[key] = CitationRecord(
            key=key,
            path=path,
            title=fields.get("title", ""),
            doi=fields.get("doi", ""),
            arxiv_id=fields.get("arxiv", ""),
        )
    return records
```

- [ ] **Step 4: Implement deterministic YAML emitter**

Create `tools/research/fecim_research/yamlio.py`:

```python
from collections.abc import Mapping, Sequence


def dumps_yaml(value: object, indent: int = 0) -> str:
    lines: list[str] = []
    _emit(value, lines, indent)
    return "\n".join(lines) + "\n"


def _scalar(value: object) -> str:
    if value is None:
        return "null"
    if isinstance(value, bool):
        return "true" if value else "false"
    text = str(value)
    if text == "" or text.lower() in {"null", "true", "false"} or any(c in text for c in ":#[]{}"):
        return '"' + text.replace('"', '\\"') + '"'
    return text


def _emit(value: object, lines: list[str], indent: int) -> None:
    pad = " " * indent
    if isinstance(value, Mapping):
        for key in sorted(value.keys()):
            item = value[key]
            if isinstance(item, (Mapping, list, tuple)):
                lines.append(f"{pad}{key}:")
                _emit(item, lines, indent + 2)
            else:
                lines.append(f"{pad}{key}: {_scalar(item)}")
        return
    if isinstance(value, Sequence) and not isinstance(value, (str, bytes, bytearray)):
        for item in value:
            if isinstance(item, Mapping):
                lines.append(f"{pad}-")
                _emit(item, lines, indent + 2)
            else:
                lines.append(f"{pad}- {_scalar(item)}")
        return
    lines.append(f"{pad}{_scalar(value)}")
```

- [ ] **Step 5: Implement PDF discovery and matching**

Create `tools/research/fecim_research/discovery.py`:

```python
from dataclasses import dataclass
from pathlib import Path
import hashlib

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
    for extra in extra_paths:
        base = extra if extra.is_absolute() else root / extra
        if base.is_file() and base.suffix.lower() == ".pdf":
            paths.add(base)
        elif base.is_dir():
            paths.update(base.rglob("*.pdf"))
    out: list[DiscoveredPDF] = []
    for path in sorted(paths):
        if not path.is_file():
            continue
        out.append(DiscoveredPDF(path=path, sha256=sha256_file(path), size=path.stat().st_size))
    return out


def match_pdf_to_record(pdf: DiscoveredPDF, records: dict[str, CitationRecord]) -> PDFMatch:
    stem = pdf.path.stem.lower()
    for key in sorted(records):
        if stem == key.lower() or key.lower() in stem:
            return PDFMatch(status="matched", paper_key=key, method="filename", confidence=0.95)
    return PDFMatch(status="unmatched", paper_key=None, method="none", confidence=0.0)
```

- [ ] **Step 6: Run discovery tests and confirm GREEN**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_discovery.py' -v
```

Expected: PASS.

- [ ] **Step 7: Commit Task 4**

Run:

```bash
git add tools/research/fecim_research/citations.py tools/research/fecim_research/discovery.py tools/research/fecim_research/yamlio.py tools/research/tests/test_discovery.py
git commit -m "feat(research): discover and match local PDFs"
```

## Task 5: Implement Parser Output Selection And Chunking

**Files:**

- Create: `tools/research/fecim_research/parsing.py`
- Create: `tools/research/fecim_research/chunking.py`
- Create: `tools/research/tests/test_chunking.py`

- [ ] **Step 1: Write failing chunking tests**

Create `tools/research/tests/test_chunking.py`:

```python
import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.chunking import chunk_markdown, write_chunks_jsonl


class ChunkingTest(unittest.TestCase):
    def test_chunks_markdown_by_heading_and_size(self):
        text = "# Title\n\n## Results\n\nHZO coercive field text.\n\nMore remanent polarization text."
        chunks = chunk_markdown("park2015_advmat_hzo", text, max_chars=40)
        self.assertGreaterEqual(len(chunks), 2)
        self.assertEqual(chunks[0]["paper_key"], "park2015_advmat_hzo")
        self.assertIn("contents", chunks[0])
        self.assertTrue(chunks[0]["id"].startswith("park2015_advmat_hzo::sec-"))

    def test_write_chunks_jsonl_is_deterministic(self):
        chunks = chunk_markdown("park2015_advmat_hzo", "## Results\n\nHZO coercive field text.", max_chars=100)
        with tempfile.TemporaryDirectory() as td:
            path = Path(td) / "chunks.jsonl"
            write_chunks_jsonl(path, chunks)
            lines = path.read_text(encoding="utf-8").splitlines()
            self.assertEqual(len(lines), 1)
            record = json.loads(lines[0])
            self.assertEqual(record["paper_key"], "park2015_advmat_hzo")
            self.assertEqual(record["chunk_number"], 1)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run chunking tests and confirm RED**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_chunking.py' -v
```

Expected: FAIL because `chunking.py` does not exist.

- [ ] **Step 3: Implement parser status helpers**

Create `tools/research/fecim_research/parsing.py`:

```python
from dataclasses import dataclass, asdict
from pathlib import Path
import json
import os
import shutil
import subprocess
import urllib.request


@dataclass(frozen=True)
class ParseResult:
    paper_key: str
    parser: str
    status: str
    output_path: str
    message: str


def write_parse_manifest(path: Path, results: list[ParseResult]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = {"results": [asdict(r) for r in results]}
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")


def run_marker_if_configured(pdf: Path, out_md: Path) -> ParseResult:
    cmd = os.environ.get("FECIM_MARKER_CMD", "").strip()
    if not cmd:
        return ParseResult("", "marker", "skipped", str(out_md), "FECIM_MARKER_CMD is not set")
    out_md.parent.mkdir(parents=True, exist_ok=True)
    result = subprocess.run([*cmd.split(), str(pdf), str(out_md)], text=True, capture_output=True)
    if result.returncode != 0:
        return ParseResult("", "marker", "failed", str(out_md), result.stderr.strip())
    return ParseResult("", "marker", "ok", str(out_md), "marker completed")


def copy_sidecar_markdown_if_present(pdf: Path, out_md: Path) -> ParseResult:
    sidecar = pdf.with_suffix(".md")
    if not sidecar.exists():
        return ParseResult("", "sidecar-markdown", "skipped", str(out_md), "no sidecar markdown")
    out_md.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(sidecar, out_md)
    return ParseResult("", "sidecar-markdown", "ok", str(out_md), f"copied {sidecar}")
```

Note: full GROBID HTTP integration is added in Task 6. This task establishes parser result records and a sidecar Markdown path for deterministic tests.

- [ ] **Step 4: Implement deterministic chunking**

Create `tools/research/fecim_research/chunking.py`:

```python
from __future__ import annotations

import hashlib
import json
import re
from pathlib import Path


HEADING_RE = re.compile(r"^(#{1,6})\s+(?P<title>.+)$", re.MULTILINE)


def _sha(text: str) -> str:
    return hashlib.sha256(text.encode("utf-8")).hexdigest()


def _sections(markdown: str) -> list[tuple[str, str]]:
    matches = list(HEADING_RE.finditer(markdown))
    if not matches:
        return [("Body", markdown.strip())] if markdown.strip() else []
    sections: list[tuple[str, str]] = []
    for i, match in enumerate(matches):
        start = match.end()
        end = matches[i + 1].start() if i + 1 < len(matches) else len(markdown)
        title = match.group("title").strip()
        body = markdown[start:end].strip()
        if body:
            sections.append((title, body))
    return sections


def chunk_markdown(paper_key: str, markdown: str, max_chars: int = 1800) -> list[dict[str, object]]:
    chunks: list[dict[str, object]] = []
    chunk_number = 1
    for section_number, (section, body) in enumerate(_sections(markdown), start=1):
        paragraphs = [p.strip() for p in re.split(r"\n\s*\n", body) if p.strip()]
        current = ""
        for paragraph in paragraphs:
            candidate = paragraph if not current else current + "\n\n" + paragraph
            if current and len(candidate) > max_chars:
                chunks.append(_record(paper_key, section, section_number, chunk_number, current))
                chunk_number += 1
                current = paragraph
            else:
                current = candidate
        if current:
            chunks.append(_record(paper_key, section, section_number, chunk_number, current))
            chunk_number += 1
    return chunks


def _record(paper_key: str, section: str, section_number: int, chunk_number: int, contents: str) -> dict[str, object]:
    return {
        "id": f"{paper_key}::sec-{section_number:02d}::chunk-{chunk_number:03d}",
        "paper_key": paper_key,
        "contents": contents,
        "section": section,
        "section_number": section_number,
        "chunk_number": chunk_number,
        "source_parser": "marker",
        "source_path": f"research/parsed/{paper_key}/marker.md",
        "page_start": None,
        "page_end": None,
        "char_start": None,
        "char_end": None,
        "sha256": _sha(contents),
    }


def write_chunks_jsonl(path: Path, chunks: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8") as f:
        for chunk in chunks:
            f.write(json.dumps(chunk, sort_keys=True, ensure_ascii=False) + "\n")
```

- [ ] **Step 5: Run chunking tests and confirm GREEN**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_chunking.py' -v
```

Expected: PASS.

- [ ] **Step 6: Commit Task 5**

Run:

```bash
git add tools/research/fecim_research/parsing.py tools/research/fecim_research/chunking.py tools/research/tests/test_chunking.py
git commit -m "feat(research): chunk parsed paper text"
```

## Task 6: Implement `research ingest`

**Files:**

- Create: `tools/research/fecim_research/ingest.py`
- Create: `tools/research/tests/test_ingest.py`
- Modify: `tools/research/fecim_research/parsing.py`

- [ ] **Step 1: Write failing ingest tests**

Create `tools/research/tests/test_ingest.py`:

```python
import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.ingest import run_ingest


class IngestTest(unittest.TestCase):
    def test_ingest_writes_source_chunk_manifest_and_unmatched_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `park2015_advmat_hzo`\n**DOI:** `10.1002/adma.201404531`\n", encoding="utf-8")

            pdf = root / "research" / "papers" / "park2015_advmat_hzo.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF fixture")
            pdf.with_suffix(".md").write_text("## Results\n\nHZO coercive field evidence.", encoding="utf-8")

            unknown = root / "research" / "papers" / "unknown.pdf"
            unknown.write_bytes(b"%PDF unknown")

            code = run_ingest(root=root, extra_paths=[])
            self.assertEqual(code, 0)

            self.assertTrue((root / "research" / "sources" / "park2015_advmat_hzo.yaml").exists())
            self.assertTrue((root / "research" / "parsed" / "park2015_advmat_hzo" / "marker.md").exists())
            self.assertTrue((root / "research" / "chunks" / "park2015_advmat_hzo.jsonl").exists())
            report = json.loads((root / "research" / "reports" / "unmatched-pdfs.json").read_text(encoding="utf-8"))
            self.assertEqual(len(report["unmatched"]), 1)
            self.assertIn("unknown.pdf", report["unmatched"][0]["path"])


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Run ingest tests and confirm RED**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_ingest.py' -v
```

Expected: FAIL because `ingest.py` does not exist.

- [ ] **Step 3: Add GROBID HTTP helper**

Append to `tools/research/fecim_research/parsing.py`:

```python
def run_grobid_if_available(pdf: Path, out_tei: Path) -> ParseResult:
    raw_url = os.environ.get("FECIM_GROBID_URL", "").strip()
    if not raw_url:
        return ParseResult("", "grobid", "skipped", str(out_tei), "FECIM_GROBID_URL is not set")
    url = raw_url.rstrip("/")
    endpoint = f"{url}/api/processFulltextDocument"
    try:
        boundary = "----fecimresearchboundary"
        data = pdf.read_bytes()
        body = (
            f"--{boundary}\r\n"
            f'Content-Disposition: form-data; name="input"; filename="{pdf.name}"\r\n'
            "Content-Type: application/pdf\r\n\r\n"
        ).encode("utf-8") + data + f"\r\n--{boundary}--\r\n".encode("utf-8")
        request = urllib.request.Request(
            endpoint,
            data=body,
            headers={"Content-Type": f"multipart/form-data; boundary={boundary}"},
            method="POST",
        )
        with urllib.request.urlopen(request, timeout=20) as response:
            text = response.read().decode("utf-8", errors="replace")
    except Exception as exc:
        return ParseResult("", "grobid", "failed", str(out_tei), str(exc))
    out_tei.parent.mkdir(parents=True, exist_ok=True)
    out_tei.write_text(text, encoding="utf-8")
    return ParseResult("", "grobid", "ok", str(out_tei), "grobid completed")
```

- [ ] **Step 4: Implement ingest orchestration**

Create `tools/research/fecim_research/ingest.py`:

```python
from pathlib import Path
import json

from .chunking import chunk_markdown, write_chunks_jsonl
from .citations import load_citation_records
from .discovery import discover_pdfs, match_pdf_to_record
from .parsing import copy_sidecar_markdown_if_present, run_grobid_if_available, run_marker_if_configured, write_parse_manifest
from .yamlio import dumps_yaml


def run_ingest(root: Path, extra_paths: list[Path]) -> int:
    records = load_citation_records(root)
    pdfs = discover_pdfs(root, extra_paths)
    unmatched: list[dict[str, object]] = []
    processed = 0

    for pdf in pdfs:
        match = match_pdf_to_record(pdf, records)
        if match.status != "matched" or match.paper_key is None:
            unmatched.append({"path": str(pdf.path), "sha256": pdf.sha256, "size": pdf.size, "status": "unmatched"})
            continue
        paper_key = match.paper_key
        _write_source(root, paper_key, pdf, match)
        parsed_dir = root / "research" / "parsed" / paper_key
        grobid = run_grobid_if_available(pdf.path, parsed_dir / "grobid.tei.xml")
        sidecar = copy_sidecar_markdown_if_present(pdf.path, parsed_dir / "marker.md")
        marker = sidecar if sidecar.status == "ok" else run_marker_if_configured(pdf.path, parsed_dir / "marker.md")
        write_parse_manifest(parsed_dir / "manifest.json", [grobid, marker])

        marker_path = parsed_dir / "marker.md"
        if marker_path.exists():
            chunks = chunk_markdown(paper_key, marker_path.read_text(encoding="utf-8", errors="replace"))
            write_chunks_jsonl(root / "research" / "chunks" / f"{paper_key}.jsonl", chunks)
            processed += 1

    reports = root / "research" / "reports"
    reports.mkdir(parents=True, exist_ok=True)
    (reports / "unmatched-pdfs.json").write_text(
        json.dumps({"unmatched": unmatched}, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
    _write_manifest(root, processed, len(unmatched))
    print(f"ingest complete: processed={processed} unmatched={len(unmatched)}")
    return 0


def _write_source(root: Path, paper_key: str, pdf, match) -> None:
    record = {
        "paper_key": paper_key,
        "status": "matched",
        "match": {"method": match.method, "confidence": match.confidence},
        "pdf": {"path": str(pdf.path), "sha256": pdf.sha256, "size": pdf.size, "acquisition": "local"},
        "citation_record": f"citations/papers/{paper_key}.md",
    }
    path = root / "research" / "sources" / f"{paper_key}.yaml"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(dumps_yaml(record), encoding="utf-8")


def _write_manifest(root: Path, processed: int, unmatched: int) -> None:
    path = root / "research" / "manifests" / "ingest-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        json.dumps({"processed": processed, "unmatched": unmatched}, indent=2, sort_keys=True) + "\n",
        encoding="utf-8",
    )
```

- [ ] **Step 5: Run ingest tests and confirm GREEN**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_ingest.py' -v
```

Expected: PASS.

- [ ] **Step 6: Run the CLI against the testable sidecar path**

Run:

```bash
PYTHONPATH=tools/research python3 tools/research/research_cli.py --repo-root . ingest
```

Expected: exits `0`, writes reports under `research/reports/`, and does not require GROBID or Marker for papers without matching sidecar Markdown.

- [ ] **Step 7: Commit Task 6**

Run:

```bash
git add tools/research/fecim_research/ingest.py tools/research/fecim_research/parsing.py tools/research/tests/test_ingest.py
git commit -m "feat(research): ingest matched local papers"
```

## Task 7: Implement BM25 Indexing And Evidence Search

**Files:**

- Create: `tools/research/fecim_research/indexing.py`
- Create: `tools/research/fecim_research/searching.py`
- Create: `tools/research/tests/test_indexing.py`
- Create: `tools/research/tests/test_searching.py`

- [ ] **Step 1: Write failing indexing tests**

Create `tools/research/tests/test_indexing.py`:

```python
import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.indexing import collect_chunk_files, write_index_manifest


class IndexingTest(unittest.TestCase):
    def test_collect_chunk_files_sorted(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunks = root / "research" / "chunks"
            chunks.mkdir(parents=True)
            (chunks / "b.jsonl").write_text("{}", encoding="utf-8")
            (chunks / "a.jsonl").write_text("{}", encoding="utf-8")
            got = collect_chunk_files(root)
            self.assertEqual([p.name for p in got], ["a.jsonl", "b.jsonl"])

    def test_write_index_manifest_records_backend_and_inputs(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "a.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text("{\"id\":\"a\",\"contents\":\"text\"}\n", encoding="utf-8")
            path = write_index_manifest(root, "pyserini", [chunk], semantic=False, embedding_model="")
            data = json.loads(path.read_text(encoding="utf-8"))
            self.assertEqual(data["backend"], "pyserini")
            self.assertFalse(data["semantic"])
            self.assertEqual(len(data["inputs"]), 1)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 2: Write failing search formatting tests**

Create `tools/research/tests/test_searching.py`:

```python
import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.searching import load_chunk_lookup, render_text_results


class SearchingTest(unittest.TestCase):
    def test_load_chunk_lookup_reads_jsonl_chunks(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            chunk = root / "research" / "chunks" / "park.jsonl"
            chunk.parent.mkdir(parents=True)
            chunk.write_text(
                json.dumps({"id": "park::sec-01::chunk-001", "paper_key": "park", "contents": "HZO coercive field evidence"}) + "\n",
                encoding="utf-8",
            )
            lookup = load_chunk_lookup(root)
            self.assertIn("park::sec-01::chunk-001", lookup)

    def test_render_text_results_includes_score_key_and_snippet(self):
        rows = [
            {
                "rank": 1,
                "score": 7.5,
                "docid": "park::sec-01::chunk-001",
                "paper_key": "park",
                "section": "Results",
                "snippet": "HZO coercive field evidence",
            }
        ]
        text = render_text_results(rows)
        self.assertIn("1. park", text)
        self.assertIn("score=7.5", text)
        self.assertIn("HZO coercive field evidence", text)


if __name__ == "__main__":
    unittest.main()
```

- [ ] **Step 3: Run indexing/search tests and confirm RED**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_indexing.py' -v
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_searching.py' -v
```

Expected: FAIL because indexing and searching modules do not exist.

- [ ] **Step 4: Implement indexing wrapper**

Create `tools/research/fecim_research/indexing.py`:

```python
from pathlib import Path
import hashlib
import json
import subprocess
import sys


def collect_chunk_files(root: Path) -> list[Path]:
    chunks = root / "research" / "chunks"
    if not chunks.exists():
        return []
    return sorted(chunks.glob("*.jsonl"))


def _sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def write_index_manifest(root: Path, backend: str, inputs: list[Path], semantic: bool, embedding_model: str) -> Path:
    path = root / "research" / "manifests" / "index-latest.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = {
        "backend": backend,
        "semantic": semantic,
        "embedding_model": embedding_model,
        "inputs": [{"path": str(p), "sha256": _sha(p)} for p in inputs],
        "pyserini_index": "research/index/pyserini",
    }
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    return path


def run_index(root: Path, semantic: bool, embedding_model: str) -> int:
    if semantic:
        print("semantic indexing is not implemented in the retrieval MVP", file=sys.stderr)
        return 2
    inputs = collect_chunk_files(root)
    if not inputs:
        print("no chunk files found under research/chunks", file=sys.stderr)
        return 1
    index_dir = root / "research" / "index" / "pyserini"
    index_dir.parent.mkdir(parents=True, exist_ok=True)
    if index_dir.exists():
        import shutil

        shutil.rmtree(index_dir)
    cmd = [
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
    result = subprocess.run(cmd)
    if result.returncode != 0:
        print("pyserini indexing failed; install Pyserini and Java, then retry", file=sys.stderr)
        return result.returncode
    write_index_manifest(root, "pyserini", inputs, semantic=False, embedding_model="")
    print(f"indexed {len(inputs)} chunk files into {index_dir}")
    return 0
```

- [ ] **Step 5: Implement search wrapper and evidence formatting**

Create `tools/research/fecim_research/searching.py`:

```python
from pathlib import Path
import json
import sys


def load_chunk_lookup(root: Path) -> dict[str, dict[str, object]]:
    lookup: dict[str, dict[str, object]] = {}
    for path in sorted((root / "research" / "chunks").glob("*.jsonl")):
        for line in path.read_text(encoding="utf-8").splitlines():
            if not line.strip():
                continue
            record = json.loads(line)
            record["chunk_file"] = str(path)
            lookup[str(record["id"])] = record
    return lookup


def render_text_results(rows: list[dict[str, object]]) -> str:
    lines: list[str] = []
    for row in rows:
        lines.append(
            f'{row["rank"]}. {row["paper_key"]}  score={row["score"]}  sec={row.get("section", "")}'
        )
        lines.append(f'   chunk: {row["docid"]}')
        lines.append(f'   Snippet: {row["snippet"]}')
    return "\n".join(lines) + ("\n" if lines else "")


def _snippet(text: str, limit: int = 240) -> str:
    compact = " ".join(text.split())
    return compact if len(compact) <= limit else compact[: limit - 1] + "..."


def run_search(root: Path, query: str, limit: int, json_output: bool) -> int:
    index_dir = root / "research" / "index" / "pyserini"
    if not index_dir.exists():
        print("missing BM25 index; run `fecim research index` first", file=sys.stderr)
        return 1
    try:
        from pyserini.search.lucene import LuceneSearcher
    except Exception as exc:
        print(f"pyserini is unavailable: {exc}", file=sys.stderr)
        return 1

    lookup = load_chunk_lookup(root)
    searcher = LuceneSearcher(str(index_dir))
    hits = searcher.search(query, k=limit)
    rows: list[dict[str, object]] = []
    for rank, hit in enumerate(hits, start=1):
        chunk = lookup.get(hit.docid, {})
        contents = str(chunk.get("contents", ""))
        rows.append(
            {
                "rank": rank,
                "score": hit.score,
                "docid": hit.docid,
                "paper_key": chunk.get("paper_key", hit.docid.split("::", 1)[0]),
                "section": chunk.get("section", ""),
                "snippet": _snippet(contents),
                "chunk_file": chunk.get("chunk_file", ""),
                "source_parser": chunk.get("source_parser", ""),
            }
        )
    if json_output:
        print(json.dumps({"query": query, "results": rows}, indent=2, sort_keys=True))
    else:
        print(render_text_results(rows), end="")
    return 0
```

- [ ] **Step 6: Run indexing/search unit tests and confirm GREEN**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_indexing.py' -v
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -p 'test_searching.py' -v
```

Expected: PASS.

- [ ] **Step 7: Run optional Pyserini smoke test if Pyserini is installed**

Run:

```bash
PYTHONPATH=tools/research python3 - <<'PY'
import importlib.util
raise SystemExit(0 if importlib.util.find_spec("pyserini") else 1)
PY
```

If that command exits `0`, run:

```bash
PYTHONPATH=tools/research python3 tools/research/research_cli.py --repo-root . index
PYTHONPATH=tools/research python3 tools/research/research_cli.py --repo-root . search "HZO coercive field" --limit 3
```

Expected: index command creates `research/index/pyserini/`; search prints ranked evidence. If Pyserini is missing, record that optional smoke test was skipped.

- [ ] **Step 8: Commit Task 7**

Run:

```bash
git add tools/research/fecim_research/indexing.py tools/research/fecim_research/searching.py tools/research/tests/test_indexing.py tools/research/tests/test_searching.py
git commit -m "feat(research): index and search paper chunks"
```

## Task 8: Add End-To-End CLI Contract

**Files:**

- Create: `cmd/fecim-lattice-tools/research_e2e_test.go`
- Modify: `tools/research/tests/test_cli.py`

- [ ] **Step 1: Write failing Go wrapper e2e test with injected runner**

Create `cmd/fecim-lattice-tools/research_e2e_test.go`:

```go
package main

import "testing"

func TestResearchCommandsForwardExpectedArguments(t *testing.T) {
	var calls [][]string
	previous := researchRunner
	researchRunner = func(args []string) error {
		calls = append(calls, append([]string(nil), args...))
		return nil
	}
	defer func() { researchRunner = previous }()

	commands := [][]string{
		{"research", "ingest"},
		{"research", "index"},
		{"research", "search", "HZO coercive field Preisach"},
	}
	for _, cmd := range commands {
		if err := dispatchSubcommand(cmd); err != nil {
			t.Fatalf("dispatch %v: %v", cmd, err)
		}
	}
	if len(calls) != 3 {
		t.Fatalf("calls=%d want 3", len(calls))
	}
	if calls[0][0] != "ingest" || calls[1][0] != "index" || calls[2][0] != "search" {
		t.Fatalf("unexpected forwarded calls: %#v", calls)
	}
}
```

- [ ] **Step 2: Run e2e test and confirm RED or GREEN**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run TestResearchCommandsForwardExpectedArguments -count=1
```

Expected: PASS if Task 2 implementation already covers this. If it fails, fix `runResearchSubcommand` argument forwarding only.

- [ ] **Step 3: Add Python CLI command import regression**

Append to `tools/research/tests/test_cli.py`:

```python
    def test_core_commands_import_without_optional_dependencies(self):
        import fecim_research.ingest
        import fecim_research.indexing
        import fecim_research.searching

        self.assertTrue(hasattr(fecim_research.ingest, "run_ingest"))
        self.assertTrue(hasattr(fecim_research.indexing, "run_index"))
        self.assertTrue(hasattr(fecim_research.searching, "run_search"))
```

- [ ] **Step 4: Run focused Go and Python tests**

Run:

```bash
go test ./cmd/fecim-lattice-tools -run 'TestResearch' -count=1
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -v
```

Expected: PASS.

- [ ] **Step 5: Commit Task 8**

Run:

```bash
git add cmd/fecim-lattice-tools/research_e2e_test.go tools/research/tests/test_cli.py
git commit -m "test(research): cover CLI forwarding and imports"
```

## Task 9: Final Verification

**Files:**

- Modify only if prior tasks require small corrections.

- [ ] **Step 1: Run research test target**

Run:

```bash
make test-research
```

Expected: PASS.

- [ ] **Step 2: Run command package tests**

Run:

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools -count=1
```

Expected: PASS.

- [ ] **Step 3: Run Python full research tests directly**

Run:

```bash
PYTHONPATH=tools/research python3 -m unittest discover -s tools/research/tests -v
```

Expected: PASS.

- [ ] **Step 4: Run git diff check for generated cache leakage**

Run:

```bash
git status --short
```

Expected: no tracked files under `research/index/pyserini/`, `research/index/lancedb/`, `research/index/models/`, or `research/.cache/`. Existing unrelated dirty files may remain; do not revert them.

- [ ] **Step 5: Confirm no final cleanup commit is required**

If Task 9 changed no files, do not create an empty commit. If Task 9 exposed a specific failing file, return to the task that owns that file, apply the fix there, rerun that task's focused tests, and commit with that task's commit message pattern.

## Handoff Notes

- This plan intentionally avoids a mandatory Pyserini install in unit tests. The actual `index` and `search` commands require Pyserini at runtime and must fail with actionable messages when it is unavailable.
- GROBID and Marker are optional during tests. Sidecar Markdown lets tests exercise deterministic chunking without requiring parser services.
- Do not commit PDFs unless the repository policy changes. The ledger should commit metadata, parser outputs, chunks, manifests, and reports.
- Do not modify unrelated dirty files. At plan time, unrelated dirty files were present in `data/calibrations/literature_superlattice.json`, `internal/gogpuapp/root_test.go`, `shared/viewmodel/circuits/*.go`, `shared/viewmodel/circuits/viewmodel_test.go`, and untracked binaries.
