# Disputed And Weak Claims

This file tracks claims that are contested, weakly sourced, or easy to overstate.

Do not use this file to smuggle unverified claims into the project. Use it to make uncertainty explicit.

## Entry Format

```markdown
## {Claim}

**Status:** `{contested | weak evidence | source mismatch | unresolved}`

| Position | Source | Evidence Level | Notes |
|---|---|---|---|
| Supports | `{key}` | `{peer-reviewed | conference | preprint | other}` | `{specific location and caveat}` |
| Contradicts | `{key}` | `{peer-reviewed | conference | preprint | other}` | `{specific location and caveat}` |

**Project handling:** `{how the repository labels or avoids this claim}`
**Next action:** `{source to read, experiment to run, or wording to change}`
```

## Active Entries

## Guo 2018 NLS citation: DOI identifies a different paper (2026-10-25)

**Status:** `source mismatch` — mandatory research stop; original records not repaired.

| Position | Source | Evidence Level | Notes |
|---|---|---|---|
| Repository attribution | `citations/papers/guo2018_apl_nls.md:1-7,21-39,62`; `docs/research/paper/refs.bib:33-40` | Claimed peer-reviewed journal; identity contradicted | Attributes DOI `10.1063/1.5038038` to Guo, HfO₂ NLS, APL 112, 262903 (2018), and treats 1.5 distribution width / 100 ps attempt time as sourced. |
| Exact DOI identity | [DOI](https://doi.org/10.1063/1.5038038), [Crossref](https://api.crossref.org/works/10.1063/1.5038038), [publisher destination](https://pubs.aip.org/jap/article/124/15/152132/347802/Reservoir-computing-with-a-single-delay-coupled) | Peer-reviewed journal; bibliographic metadata and DOI redirect checked, article body blocked HTTP 403 | Guillaume Dion, Salim Mejaouri and Julien Sylvestre, “Reservoir computing with a single delay-coupled non-linear mechanical oscillator,” Journal of Applied Physics 124(15), 152132 (2018). |

**Reason for mismatch finding:** exact-identifier deposited metadata and DOI redirect agree on a different title, author list and journal. No search snippet was used as verification. The actual scientific contents of the blocked publisher page were not extracted.

**Project handling:** `docs/research/paper/main.tex:74` cites this entry as HfO₂/HZO literature. Source support for `shared/physics/landau.go:116,182-183` and `shared/physics/nls.go:32` remains **NOT VERIFIED**. This does not prove the numerical constants false or that a genuine Guo NLS paper does not exist. No replacement source, figure, measurement conditions or numerical values were invented.

**Next action:** human review of [the stopped audit](reports/research-audit-2026-10-25.md); decide citation/parameter evidence handling and authorize any resumed research. No scientific-content fix or commit was made.

**Integration note (2026-10-28):** Historical stopped-audit finding imported with the owner's merge request. The source mismatch is not repaired, research is not resumed, and no replacement source or parameter value is approved. Raw retrieval/log artifacts remain local; the tracked audit links the public identity sources.
