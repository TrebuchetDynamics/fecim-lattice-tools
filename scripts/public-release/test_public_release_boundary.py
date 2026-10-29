"""Exercise the release tools through their command-line seams in tracked trees."""

import csv
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest


SCRIPTS = Path(__file__).resolve().parent
AUDIT = "docs/internals/release/THIRD_PARTY_PDF_AUDIT.csv"
FIELDS = ["path", "title", "source_url", "evidence_url", "rights_basis",
          "obligations", "decision", "notes"]
DOC_ROOTS = ["docs/guides", "docs/modules", "docs/internals", "docs/research"]
MODULE_ROOTS = ["module1-hysteresis", "module2-crossbar", "module3-mnist",
                "module4-circuits", "module5-comparison", "module6-eda", "module7-docs"]


class TestPublicReleaseBoundary(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory(dir=os.environ.get("TMPDIR"))
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        subprocess.run(["git", "init", "-q", str(self.root)], check=True)
        for name in ("check_public_release_boundary.py", "bootstrap_pdf_audit.py"):
            target = self.root / "scripts/public-release" / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(SCRIPTS / name, target)
        for path in ["README.md", "CLAUDE.md"] + [
            f"{root}/safe.md" for root in DOC_ROOTS + MODULE_ROOTS
        ]:
            self.track(path, "Safe source material.\n")

    def track(self, path, text="%PDF-1.4\nfixture\n"):
        target = self.root / path
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(text, encoding="utf-8")
        subprocess.run(["git", "add", "--", path], cwd=self.root, check=True)

    def audit(self, rows):
        target = self.root / AUDIT
        target.parent.mkdir(parents=True, exist_ok=True)
        with target.open("w", newline="", encoding="utf-8") as stream:
            writer = csv.DictWriter(stream, fieldnames=FIELDS)
            writer.writeheader()
            writer.writerows(rows)

    def run_tool(self, name="check_public_release_boundary.py", env=None):
        return subprocess.run(
            [sys.executable, str(self.root / "scripts/public-release" / name)],
            cwd=self.root, capture_output=True, text=True,
            env={**os.environ, "PYTHONDONTWRITEBYTECODE": "1", **(env or {})},
        )

    def assert_exit(self, result, code):
        self.assertEqual(result.returncode, code, result.stdout + result.stderr)

    def test_public_journal_privacy_warning_is_not_a_scan_false_positive(self):
        path = "docs/internals/journal/README.md"
        self.track(path, (SCRIPTS.parents[1] / path).read_text(encoding="utf-8"))
        self.audit([])
        self.assert_exit(self.run_tool(), 0)

    def test_canonical_audit_decisions_fail_closed(self):
        path = "docs/research/papers/example.pdf"
        self.track(path)
        for decision, code, finding in [
            ("keep", 0, "checks passed"),
            ("keep-with-conditions", 0, "checks passed"),
            ("replace-with-link", 1, "Blocked PDF decision"),
            ("", 1, "Blocked PDF decision"),
            (None, 1, "Missing PDF audit row"),
        ]:
            with self.subTest(decision=decision):
                self.audit([] if decision is None else [{"path": path, "decision": decision}])
                result = self.run_tool()
                self.assert_exit(result, code)
                self.assertIn(finding, result.stdout)
        (self.root / AUDIT).unlink()
        result = self.run_tool()
        self.assert_exit(result, 1)
        self.assertIn("Missing PDF audit row", result.stdout)

    def test_bootstrap_preserves_cells_history_and_denies_unknown_pdfs(self):
        self.track("docs/research/papers/known.pdf")
        self.track("docs/research/papers/new.pdf")
        known = dict(zip(FIELDS, ["docs/4-research/papers/known.pdf", "",
                                 "https://example.org/source", "https://example.org/evidence",
                                 "Recorded license", "retain notice; indicate changes",
                                 "keep-with-conditions", "Original, quoted\nhistory"]))
        history = {field: "" for field in FIELDS}
        history["path"] = "docs/4-research/papers/historical.pdf"
        untouched = {**history, "path": "docs/4-research-other/historical.pdf"}
        self.audit([known, history, untouched])
        self.assert_exit(self.run_tool("bootstrap_pdf_audit.py"), 0)
        with (self.root / AUDIT).open(newline="", encoding="utf-8") as stream:
            rows = list(csv.DictReader(stream))
        self.assertEqual(rows[:3], [
            {**known, "path": "docs/research/papers/known.pdf"},
            {**history, "path": "docs/research/papers/historical.pdf"}, untouched,
        ])
        self.assertEqual(len(rows), 4)
        self.assertEqual(rows[3]["path"], "docs/research/papers/new.pdf")
        self.assertEqual(rows[3]["decision"], "replace-with-link")
        self.assertEqual(rows[3]["evidence_url"], "")
        before = (self.root / AUDIT).read_bytes()
        self.assert_exit(self.run_tool("bootstrap_pdf_audit.py"), 0)
        self.assertEqual((self.root / AUDIT).read_bytes(), before)
        result = self.run_tool()
        self.assert_exit(result, 1)
        self.assertIn("Blocked PDF decision: docs/research/papers/new.pdf", result.stdout)

    def test_conflicting_normalized_keys_are_rejected_without_rewriting(self):
        self.track("docs/research/papers/known.pdf")
        original = {"path": "docs/4-research/papers/known.pdf", "decision": "keep",
                    "evidence_url": "https://example.org/original"}
        for conflict in [{"decision": "replace-with-link"},
                         {"evidence_url": "https://example.org/different"}]:
            for tool in ["check_public_release_boundary.py", "bootstrap_pdf_audit.py"]:
                with self.subTest(conflict=conflict, tool=tool):
                    self.audit([original, {**original, "path": "docs/research/papers/known.pdf",
                                           **conflict}])
                    before = (self.root / AUDIT).read_bytes()
                    result = self.run_tool(tool)
                    self.assert_exit(result, 1)
                    self.assertIn("Conflicting PDF audit rows", result.stdout + result.stderr)
                    self.assertEqual((self.root / AUDIT).read_bytes(), before)

    def test_only_exact_policy_definitions_are_exempt_from_phrase_scan(self):
        phrase = "Personal roleplay or confidential material is prohibited.\n"
        for path in ["docs/internals/release/PUBLIC_BOUNDARY.md",
                     "docs/internals/release/REMOVAL_LOG.md"]:
            self.track(path, phrase)
        self.assert_exit(self.run_tool(), 0)
        for root in DOC_ROOTS + MODULE_ROOTS:
            with self.subTest(root=root):
                path = f"{root}/PUBLIC_BOUNDARY.md"
                self.track(path, phrase)
                result = self.run_tool()
                self.assert_exit(result, 1)
                self.assertIn(f"{path}:1:", result.stdout)
                (self.root / path).unlink()

    def test_migrated_blocked_and_generated_paths_remain_gated(self):
        for path, finding in [
            ("docs/research/internal-analysis/safe.md", "Blocked tracked path"),
            ("docs/research/transcripts/COSM_2025_AI_Hardware_Breakthrough/safe.md", "Blocked tracked path"),
            ("docs/research/tools/research_notes_final.md", "Blocked tracked path"),
            ("docs/internals/superpowers/plans/safe.md", "Blocked tracked path"),
            ("docs/guides/demo-videos/safe.txt", "Tracked generated output"),
            ("validation/output/safe.txt", "Tracked generated output"),
        ]:
            with self.subTest(path=path):
                self.track(path, "Safe text.\n")
                result = self.run_tool()
                self.assert_exit(result, 1)
                self.assertIn(f"{finding}: {path}", result.stdout)
                subprocess.run(["git", "rm", "-q", "-f", "--", path], cwd=self.root, check=True)
        # Policy phrase exclusions must not remove tracked files from the PDF gate.
        self.track("docs/internals/release/PUBLIC_BOUNDARY.md/unknown.pdf")
        result = self.run_tool()
        self.assert_exit(result, 1)
        self.assertIn("Missing PDF audit row: docs/internals/release/PUBLIC_BOUNDARY.md/unknown.pdf",
                      result.stdout)

    def test_scanner_error_is_not_clearance(self):
        # Inject a deterministic tool failure; all other tests use real rg.
        bin_dir = self.root / "bin"
        bin_dir.mkdir()
        rg = bin_dir / "rg"
        rg.write_text("#!/bin/sh\nprintf 'scanner unavailable\\n' >&2\nexit 2\n", encoding="utf-8")
        rg.chmod(0o755)
        result = self.run_tool(env={"PATH": str(bin_dir) + os.pathsep + os.environ["PATH"]})
        self.assert_exit(result, 1)
        self.assertIn("rg failed with exit code 2", result.stdout)
        self.assertIn("scanner unavailable", result.stdout)


if __name__ == "__main__":
    unittest.main()
