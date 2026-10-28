# FeCIM research audit — 2026-10-25

**Integration note (2026-10-28):** This is the historical stopped audit from `review/concurrent-research-2026-10-25` (base `816d17c7`), saved under the owner's merge request. Repository line numbers, retrieval counts and access results below describe that run, not a new comprehensive audit. Raw evidence filenames refer to the retained local research worktree, not tracked checkout files; logs, interrupted retrieval state and machine/session records are deliberately not published. Public DOI/Crossref/publisher URLs remain in the finding. This integration does not resume research, repair citations/parameters or grant scientific/publication approval.

**Status: STOPPED at the user's mandatory published-output error gate. Comprehensive research is NOT complete.**

## Ten most important findings / audit observations

1. **Confirmed bibliographic contradiction affecting a manuscript:** `guo2018` associates DOI `10.1063/1.5038038` with Guo and HfO₂ nucleation-limited switching. The exact DOI identifies a mechanical-oscillator reservoir-computing paper by **Guillaume Dion, Salim Mejaouri and Julien Sylvestre**, *Journal of Applied Physics* **124**, article **152132** (2018). This is a **peer-reviewed journal** source, not the claimed HfO₂ paper. Evidence: [DOI](https://doi.org/10.1063/1.5038038), [publisher destination](https://pubs.aip.org/jap/article/124/15/152132/347802/Reservoir-computing-with-a-single-delay-coupled), [Crossref exact-DOI record](https://api.crossref.org/works/10.1063/1.5038038). Publisher body access returned HTTP 403; bibliographic identity is supported by the deposited metadata and DOI redirect, not by a search snippet.
2. **Publication-facing propagation is present:** `docs/research/paper/main.tex:74` cites `guo2018` as HfO₂/HZO ferroelectric literature; `docs/research/paper/refs.bib:33-40` supplies the mismatched identity. This is a manuscript-draft error; no claim is made that an already published article contains it. No separate error in the JOSS draft, README, or CITATION.cff was established before stopping.
3. **The citation problem also reaches parameter provenance:** `shared/physics/landau.go:116,182-183` attributes `NLSSigma=1.5` (dimensionless) and `TauInf=1.0e-10 s` (100 ps, exact unit conversion) to Guo, APL 112, 262903 (2018). `shared/physics/nls.go:32` repeats the 100 ps attribution. The cited record does not verify these numbers. Their physical correctness is **NOT VERIFIED**; the mismatch does not itself show they are numerically false.
4. **The source record overstates verification:** `citations/papers/guo2018_apl_nls.md:21-39,62` presents those constants as directly sourced and labels the paper read. That status is not evidence of the claimed source. No figure, table, or page supporting the constants was verified.
5. **Park curve provenance already has a correction in the dirty worktree:** `docs/research/validation/park-reference-correction.md` and `citations/papers/park2015_advmat_hzo.md:63-67` identify the retained curve as synthetic regression data. This audit observed that prior correction; it did not independently reconstruct the historical generator or verify Park's figure. Exact Park figure, scalar values, sample conditions, and agreement with measurements remain **NOT VERIFIED**.
6. **The current README appropriately calls 30 levels educational:** `README.md:148` explicitly denies validated-device status. Whether any peer-reviewed HZO/superlattice paper now demonstrates exactly 30 states, with retention and variability, remains **NOT VERIFIED**. An interrupted search is not evidence of absence.
7. **The repository's 98.24% MNIST claim remains NOT VERIFIED in this run:** `citations/facts.md:166-167,186` points to a 2025 Journal of Alloys and Compounds reference. The primary page was not successfully read. No accuracy, architecture, state count, or measured-versus-simulated classification is adopted here.
8. **The broad sweep is only partial:** 26 separate queries were prepared, but only the first query has completed source jobs. Saved coverage reports 59 library records and 50 unique DOIs; these are metadata counts, not included or verified FeCIM papers. Nine query/source jobs completed; 1030 pairs remained pending in the interrupted checkpoint. No topic is considered comprehensively searched.
9. **Access failures limit evidence:** DOI requests via web_read hit a Jina domain block; direct Wiley/AIP publisher requests returned HTTP 403. Lens, NASA ADS and Dimensions credentials were unavailable; ChemRxiv returned HTTP 403. The Semantic Scholar API key was absent, although its first query returned one record. No challenge bypass or restricted full-text acquisition was attempted.
10. **No scientific-content repair or parameter replacement was made:** retrieval was terminated when the manuscript citation contradiction was confirmed. All material constants, the remaining topics, citation expansion, and final scientific approvals remain incomplete. No new paper record was promoted; no commits were made.

## Method and limits

Requested scope: all eleven user topics, a repository-wide scientific claim audit, primary-source checks, and every material constant. Requested depth: **Comprehensive**. Explicit save destination: `citations/`. Retrieval used `rforge dev (a743265, 2026-10-20T01:16:43Z)`, `--sources all`, 26 queries, `--limit 10`, `--depth comprehensive`, and `--max-requests 120`. The maximum-request budget and the later stop mean this invocation could not establish exhaustive coverage. All-source expansion included irrelevant databases; source counts must not be confused with domain coverage.

Exact-DOI Crossref metadata and a DOI redirect established the stop finding. Direct public HTTP inspection was used after web tools failed. The wrong source's identity was checked, but its scientific results were not extracted. No new quantitative scientific claim is approved by this report. Existing facts are not presumed correct.

The initial worktree was extensively dirty, including citations, model code, manuscript text, and synthetic-fixture corrections. No agent tool call repaired or overwrote those files. Final status comparison detected two additional modified paths outside this audit: `docs/internals/testing/TESTING.md` and `validation/literature/forc_validation_test.go`. This agent did not edit them; concurrent/external changes cannot be ruled out. They were left untouched, and a byte-for-byte unchanged-worktree claim cannot be made. The linked research-forge shared contract file was missing. No tests or live scholarly APIs were added to CI. **TDD: N/A**, research artifacts only.

## Confirmed claims

- **Repository behavior/documentation fact:** current README labels the default 30 levels configurable educational discretization (`README.md:148`), not measured hardware evidence.
- **Repository provenance fact:** current Park correction explicitly classifies its retained curve as synthetic, not experimentally digitized. This is not confirmation of a Park material constant.
- **External bibliographic identity:** DOI `10.1063/1.5038038` identifies Dion, Mejaouri and Sylvestre's 2018 mechanical-oscillator paper, not the claimed Guo HfO₂ paper. See saved exact-DOI response and redirect headers.
- **No new FeCIM scientific numeric claim was verified before the stop.**

## Contradicted claims

| Repository claim | Location | Primary identity evidence | Impact |
|---|---|---|---|
| DOI `10.1063/1.5038038` identifies Guo et al., “Nucleation-limited polarization switching in ferroelectric HfO2 capacitors,” Applied Physics Letters 112, 262903 (2018) | `docs/research/paper/refs.bib:33-40`; `citations/refs.bib:38-45`; `citations/papers/guo2018_apl_nls.md:1-7` | Exact-DOI Crossref response identifies Dion/Mejaouri/Sylvestre, Journal of Applied Physics 124(15), 152132; DOI redirects to the same publisher article | Manuscript `main.tex:74` cites this entry as HfO₂/HZO context. Source identity is wrong; parameter support through this record is unestablished. |

This finding does **not** establish that no genuine Guo NLS paper exists. Finding a replacement source or DOI remains future research. No replacement DOI is guessed.

## Claims with no supporting source verified

- `NLSSigma=1.5` and `TauInf/Tau0=1.0e-10 s` as directly extracted from the claimed Guo reference: **NOT VERIFIED**. Thickness, temperature, electrodes, wake-up state, excitation frequency, cycle count and source figure/table/page: **NOT VERIFIED**.
- Park-attributed `Pr=24 µC/cm²`, `Ec=1.0 MV/cm`, Si doping, 10 nm thickness, and Figure 2a attribution in the existing source record: **NOT VERIFIED** in this run. No measured curve was authenticated.
- Exactly 30 reliably distinguishable HZO or superlattice states, associated retention/variation, and any James Tour/Rice/IronLattice claim: **NOT VERIFIED**.
- The existing 98.24% MNIST result and other benchmark claims: **NOT VERIFIED**.
- All other scientific claims/material constants not yet inventoried: **NOT VERIFIED**. This list is explicitly not an exhaustive repository claim ledger.

## Parameters that should change / suggested values

**No numerical replacement is justified by this audit.** The 1.5 distribution width and 100 ps time should be treated as unverified simulation defaults unless independently sourced; that is an evidence-status recommendation for human review, not a numerical correction. No changes were applied. Do not substitute the mechanical-oscillator paper as an HfO₂ parameter source.

## Important papers not yet cited

**Not assessed before the mandatory stop.** The mechanical-oscillator paper is identified solely to document the wrong DOI; it is not recommended as FeCIM literature or added to the paper library. Candidate retrieval results are unscreened and must not be promoted automatically.

## Conflicts

The source-identity conflict is also logged in `citations/disputed.md`. Both positions are retained: the repository's Guo/APL/HfO₂ entry versus the DOI's Dion/JAP/mechanical-oscillator identity. The conclusion is limited to bibliographic mismatch because exact-identifier metadata and redirect independently agree. No conflicting measured device results were compared.

## Topic coverage / outstanding work

| Topic | Status at stop |
|---|---|
| 1. HfO₂/HZO origin and reliability | First query partially retrieved; no primary numeric extraction |
| 2. Superlattices/nanolaminates | Prepared queries; NOT VERIFIED |
| 3. Multilevel FeFET/FeCap states | Prepared queries; NOT VERIFIED |
| 4. FeFET/FeCap CIM arrays and energy | Prepared queries; NOT VERIFIED |
| 5. Crossbar non-idealities | Prepared queries; NOT VERIFIED |
| 6. Peripheral energy/area | Prepared queries; NOT VERIFIED |
| 7. Modeling | One source-identity contradiction confirmed; models/parameters otherwise NOT VERIFIED |
| 8. Hardware benchmarks | Prepared queries; NOT VERIFIED |
| 9. Simulators/tools | Prepared queries; OSS search plan and license/security study not executed |
| 10. Rice/Tour/IronLattice | Prepared queries; NOT VERIFIED; no journal/talk/press conflation endorsed |
| 11. 2025-onward developments | Dedicated variants prepared; NOT VERIFIED |

Citation expansion on three seeds, full constant inventory, accepted extraction grid, gap analysis of included literature, and the complete claim audit were **not performed**, honoring the stop. An audit-only evidence grid/gaps file records this limitation without claiming rforge extraction acceptance.

## Artifacts and verification

Evidence directory: `citations/reports/fecim-research-2026-10-25/`.

- `queries.txt`, `plan.md`: requested scope and checklist.
- `retrieval/`: raw results, failures, job artifacts and interrupted batch checkpoint.
- `coverage-stats.log`: `rforge search stats` output after termination.
- `guo-doi-crossref.json`: deposited metadata, saved from exact-DOI Crossref API.
- `guo-doi-headers.txt`: DOI redirect to AIP, then HTTP 403; no accessible article-body claim.
- `stop-claim-locations.txt`: line-numbered repository propagation evidence.
- `evidence-grid.json`, `evidence-gaps.json`: audit-level evidence and incompleteness, not accepted scientific extraction.
- `provenance.json`: queries, source outcomes, interrupted status, all output paths and errors.
- `verification.json`: JSON/source-identity/checkpoint checks and baseline worktree preservation check.

Canonical `facts.md`, existing paper records, bibliographies, code, README and manuscripts were not repaired. No new paper record was added because the relevant alleged source was not established and the mandatory stop took precedence over completion of the output list. The report and dispute log preserve the finding for human review.

## Required human decision

Review the wrong-source finding and decide how to handle the manuscript citation and parameter evidence labels. Authorize resumption only after acknowledging this stop. A future pass must find genuine switching sources, verify original pages/figures and conditions, and continue all eleven topics; it must not silently turn this partial run into a completed review. No final scientific claims, screening decisions, full-text acquisition, or export approvals were self-issued.
