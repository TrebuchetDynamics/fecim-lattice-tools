# Research software release and JOSS readiness

Status: local preparation only; no release, DOI registration, or submission authorized.
Evidence checked 25 October 2026. Recheck live policies before publication.

The [first-pass verification receipt](public-preparation-verification.md) records the original GUI and adapter failures. The [follow-up correction audit](../../research/validation/park-reference-correction.md) records the Park evidence reclassification and BTO exclusion. Recheck the [follow-up verification and full diff](reference-correction-verification.md) before release; local preparation does not establish scientific validation or JOSS eligibility.

## What is established

- MIT license: [LICENSE](../../../LICENSE).
- Citation metadata: [CITATION.cff](../../../CITATION.cff); no invented version, release date, ORCID, or DOI.
- An existing local `v1.0.0-rc1` tag points to `e29dd222` (14 February 2026). Do not overwrite it or assume `v0.1.0` is the agreed next version.
- GitHub's releases endpoint returned an empty list during this review.
- GitHub reports repository creation on 26 April 2026. This does **not** prove when it first became public.
- The eight-run example demonstrates reproducibility, not calibrated physics or measured performance.
- [Internal notes](../README.md#internal-notes-and-publication-preparation) have moved; instruction entry points and CI-required generated adapters remain.
- A [JOSS manuscript draft](../../research/paper/joss/paper.md) exists separately from the older LaTeX manuscript. It is not submission-ready.

Read-only checks used:

```bash
git tag --list
git show -s --format='%h %as %s' v1.0.0-rc1
gh api repos/TrebuchetDynamics/fecim-lattice-tools/releases
gh api repos/TrebuchetDynamics/fecim-lattice-tools --jq '{created_at,visibility}'
```

## Release gate — maintainer approval required

- [ ] Choose the next version explicitly, reconciling the earlier release-candidate tag.
- [ ] Review the complete candidate diff, including pre-existing work; confirm authorship and affiliation in the citation file.
- [ ] Record full-suite, applicable race, CLI reproduction, and external-tool verification against the exact candidate commit. Old status notes are not fresh evidence.
- [ ] Review third-party data redistribution and the existing [public release boundary](PUBLIC_BOUNDARY.md).
- [ ] Approve the release notes and exact commit, then create the tag and GitHub release as separate authorized actions.
- [ ] Add actual version and release date to citation metadata as part of release preparation; keep them consistent with the published release.

## Zenodo gate — account owner action required

Enable the repository in Zenodo **before** publishing the release intended for automatic archival. Follow the official [repository enablement](https://help.zenodo.org/docs/github/enable-repository/) and [release archival](https://help.zenodo.org/docs/github/archive-software/github-upload/) instructions.

- [ ] Account owner authorizes the GitHub integration and enables this repository.
- [ ] Validate `CITATION.cff` and confirm archive metadata.
- [ ] Publish the approved GitHub release and verify that its archive record resolves.
- [ ] Record the actual version DOI and distinguish it from the concept DOI.
- [ ] Add the real DOI to citation metadata and README only after verification.

A Zenodo software DOI provides a citable archive; it does not imply peer review or scientific validation.

## JOSS gate — eligibility is not yet established

The current [submission requirements](https://joss.readthedocs.io/en/latest/submitting.html) require more than tests, documentation, and an open license:

- [ ] Establish **more than six months of public development**, with distributed, iterative activity. On the verification date, even the reported repository creation date was less than six calendar months ago; creation date is not public-history evidence.
- [ ] Document **actual research use**, at least by the developers: a named research workflow, question, inputs, outputs, and evidence of use. A deterministic example or intended audience alone does not establish research impact.
- [ ] Review open-source workflow, support expectations, governance, contribution guidance, CI, changelog, and release evidence. Stars and forks are not eligibility gates.
- [ ] Complete and human-review the [draft and bibliography](../../research/paper/joss/README.md), including a defensible comparison with related tools and why this project is a separate contribution.
- [ ] Provide the complete AI tool/model/version inventory and assistance scope for code, tests, docs, and manuscript. The human author must personally verify and attest to review, validation, and core design decisions.
- [ ] Confirm authorship, affiliations, funding, and conflicts of interest; do not infer these from Git history.
- [ ] Render the manuscript with JOSS tooling and check citations, word count (currently 750–1750), and all required sections.
- [ ] Obtain explicit approval before submitting. JOSS's current policy prohibits AI-generated author/editor/reviewer conversational interactions except translation.

The [paper format](https://joss.readthedocs.io/en/latest/paper.html) requires Summary, Statement of need, State of the field, Software design, Research impact statement, AI usage disclosure, Acknowledgements, and References. A validation plot can support the software's evidence story, but does not substitute for eligibility.

## Validation and outreach gate

- [ ] Resolve the [figure provenance issues](../../research/validation/landing-page-figure-readiness.md) before adding a headline experimental overlay.
- [ ] Keep fitting/calibration and held-out validation distinct, and report units, sample conditions, branch alignment, extraction uncertainty, RMSE definition, and reproducible generation command.
- [ ] Refresh Tour/IronLattice primary-source research immediately before drafting outreach. Company marketing and interviews are not measurement publications; an HZO model must not be described as calibrated to an unspecified IronLattice device.
- [ ] Approve any email separately; nothing in this preparation sends messages.

Update this checklist when JOSS policy, release state, figure evidence, or research-use evidence changes.
