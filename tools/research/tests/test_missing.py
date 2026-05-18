import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.missing import build_missing_report, run_missing


class MissingPapersTest(unittest.TestCase):
    def test_build_missing_report_lists_only_citation_records_without_matched_pdf(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "stored", doi="10.1000/stored")
            self._write_paper(root, "missing_with_doi", doi="10.1000/missing")
            self._write_paper(root, "missing_without_doi")
            self._write_pdf(root, "stored")
            self._write_acquisition(root, "missing_with_doi", "no_oa_pdf")

            report = build_missing_report(root)

            self.assertEqual(report["total_records"], 3)
            self.assertEqual(report["stored"], 1)
            self.assertEqual(report["missing"], 2)
            self.assertEqual(report["missing_with_doi"], 1)
            self.assertEqual(report["missing_without_doi"], 1)
            self.assertEqual([item["paper_key"] for item in report["items"]], ["missing_with_doi", "missing_without_doi"])
            self.assertEqual(report["items"][0]["status"], "needs_acquire")
            self.assertEqual(report["items"][0]["last_acquisition_status"], "no_oa_pdf")
            self.assertEqual(report["items"][0]["download_command"], "fecim research acquire missing_with_doi --download")
            self.assertEqual(report["items"][1]["status"], "missing_doi")
            self.assertEqual(report["items"][1]["download_command"], "")

    def test_run_missing_writes_git_trackable_report(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            self._write_paper(root, "missing_with_doi", doi="10.1000/missing")

            code = run_missing(root)

            self.assertEqual(code, 0)
            report_path = root / "research" / "reports" / "missing-papers-latest.json"
            self.assertTrue(report_path.exists())
            payload = json.loads(report_path.read_text())
            self.assertEqual(payload["missing"], 1)
            self.assertEqual(payload["items"][0]["paper_key"], "missing_with_doi")

    def test_explicit_existing_pdf_path_counts_as_stored_even_when_filename_differs(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            pdf = root / "docs" / "papers" / "Publisher Download Name.pdf"
            pdf.parent.mkdir(parents=True)
            pdf.write_bytes(b"%PDF-1.7\nfixture\n")
            self._write_paper(
                root,
                "park2015_advmat_hzo",
                doi="10.1000/stored",
                pdf="docs/papers/Publisher Download Name.pdf",
            )

            report = build_missing_report(root)

            self.assertEqual(report["stored"], 1)
            self.assertEqual(report["missing"], 0)
            self.assertEqual(report["items"], [])

    def _write_paper(self, root: Path, key: str, doi: str = "", pdf: str = ""):
        path = root / "citations" / "papers" / f"{key}.md"
        path.parent.mkdir(parents=True, exist_ok=True)
        lines = [f"# {key}", f"**Key:** `{key}`"]
        if doi:
            lines.append(f"**DOI:** `{doi}`")
        if pdf:
            lines.append(f"**PDF:** `{pdf}`")
        path.write_text("\n".join(lines) + "\n", encoding="utf-8")

    def _write_pdf(self, root: Path, key: str):
        path = root / "research" / "papers" / f"{key}.pdf"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(b"%PDF-1.7\nfixture\n")

    def _write_acquisition(self, root: Path, key: str, status: str):
        path = root / "research" / "sources" / f"{key}.acquisition.yaml"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(f"paper_key: {key}\nstatus: {status}\n", encoding="utf-8")


if __name__ == "__main__":
    unittest.main()
