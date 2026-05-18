import json
import tempfile
import unittest
from pathlib import Path

from fecim_research.acquisition import best_oa_pdf_candidate, run_acquire


class FakeResponse:
    def __init__(self, data: bytes, headers: dict[str, str] | None = None):
        self._data = data
        self.headers = headers or {}

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc, traceback):
        return False

    def read(self):
        return self._data


class AcquisitionTest(unittest.TestCase):
    def test_best_oa_pdf_candidate_prefers_open_best_location(self):
        work = {
            "best_oa_location": {
                "is_oa": True,
                "pdf_url": "https://example.org/paper.pdf",
                "landing_page_url": "https://example.org/paper",
                "license": "cc-by",
                "version": "publishedVersion",
                "source": {"display_name": "Example Journal", "type": "journal"},
            },
            "locations": [
                {
                    "is_oa": True,
                    "pdf_url": "https://repository.example/paper.pdf",
                    "landing_page_url": "https://repository.example/paper",
                    "license": "cc-by",
                    "version": "acceptedVersion",
                }
            ],
        }

        candidate = best_oa_pdf_candidate(work)

        self.assertIsNotNone(candidate)
        self.assertEqual(candidate["pdf_url"], "https://example.org/paper.pdf")
        self.assertEqual(candidate["version"], "publishedVersion")

    def test_acquire_downloads_only_openalex_oa_pdf_and_writes_ledgers(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "park2015_advmat_hzo.md"
            citation.parent.mkdir(parents=True)
            citation.write_text(
                "**Key:** `park2015_advmat_hzo`\n"
                "**DOI:** `10.1002/adma.201404531`\n"
                "# Park 2015\n",
                encoding="utf-8",
            )
            calls = []

            def opener(request, timeout):
                url = request.full_url
                calls.append(url)
                if "api.openalex.org/works/" in url:
                    return FakeResponse(
                        json.dumps(
                            {
                                "id": "https://openalex.org/W123",
                                "doi": "https://doi.org/10.1002/adma.201404531",
                                "display_name": "Ferroelectric HZO",
                                "open_access": {"is_oa": True, "oa_status": "gold"},
                                "best_oa_location": {
                                    "is_oa": True,
                                    "pdf_url": "https://publisher.example/park.pdf",
                                    "landing_page_url": "https://publisher.example/park",
                                    "license": "cc-by",
                                    "version": "publishedVersion",
                                    "source": {"display_name": "Advanced Materials", "type": "journal"},
                                },
                            }
                        ).encode("utf-8")
                    )
                if url == "https://publisher.example/park.pdf":
                    return FakeResponse(b"%PDF-1.4\nfixture\n", {"content-type": "application/pdf"})
                raise AssertionError(f"unexpected URL {url}")

            code = run_acquire(root=root, keys=[], download=True, opener=opener)

            self.assertEqual(code, 0)
            self.assertTrue((root / "research" / "papers" / "park2015_advmat_hzo.pdf").exists())
            self.assertTrue((root / "research" / "sources" / "park2015_advmat_hzo.openalex.json").exists())
            self.assertTrue((root / "research" / "sources" / "park2015_advmat_hzo.acquisition.yaml").exists())
            report = json.loads((root / "research" / "reports" / "acquisition-latest.json").read_text())
            self.assertEqual(report["downloaded"], 1)
            self.assertEqual(report["planned"], 1)
            self.assertEqual(report["results"][0]["status"], "downloaded")
            self.assertEqual(report["results"][0]["paper_key"], "park2015_advmat_hzo")
            self.assertIn("api.openalex.org/works/https://doi.org/10.1002/adma.201404531", calls[0])

    def test_acquire_does_not_download_closed_or_missing_pdf_records(self):
        with tempfile.TemporaryDirectory() as td:
            root = Path(td)
            citation = root / "citations" / "papers" / "closed.md"
            citation.parent.mkdir(parents=True)
            citation.write_text("**Key:** `closed`\n**DOI:** `10.1234/closed`\n", encoding="utf-8")

            def opener(request, timeout):
                return FakeResponse(
                    json.dumps(
                        {
                            "id": "https://openalex.org/W999",
                            "doi": "https://doi.org/10.1234/closed",
                            "display_name": "Closed paper",
                            "open_access": {"is_oa": False, "oa_status": "closed"},
                            "best_oa_location": None,
                            "locations": [],
                        }
                    ).encode("utf-8")
                )

            code = run_acquire(root=root, keys=[], download=True, opener=opener)

            self.assertEqual(code, 0)
            self.assertFalse((root / "research" / "papers" / "closed.pdf").exists())
            report = json.loads((root / "research" / "reports" / "acquisition-latest.json").read_text())
            self.assertEqual(report["downloaded"], 0)
            self.assertEqual(report["results"][0]["status"], "no_oa_pdf")


if __name__ == "__main__":
    unittest.main()
