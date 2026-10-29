# Validation and evidence

This page reads evidence; it does not run tests. The default candidate status is **not run/unverified**. Simulation results are not measured device data. Real Orca with Firefox on Linux remains a separate, unmet human qualification gate for both profiles.

## What the checks establish

| Check | Purpose | Limit |
| --- | --- | --- |
| Physics regression | Compare Go model output with retained fixtures | A synthetic fixture is not an experimental measurement |
| Documentation workflow | Exercise real indexing, glossary aliases, counts and anchors | Automated behavior is not human screen-reader qualification |
| Publication safety | Reject unknown sources, changed hashes, unsupported rights and unsafe references | Only the explicit public allowlist is published |
| Browser workflow | Read, search and navigate the actual Go-generated corpus | No blanket scientific Pass or Measured badge |

## Read a recorded report

A report must identify the exact source revision, source SHA-256, fixture SHA-256 and tool version. Missing reports show **not run/unverified**. Reports for another binding show **stale**. An imported report always shows **imported/unverified**; importing it is not verification. Only a separately supplied, exact-bound trusted record can show **candidate-bound recorded evidence**. Recorded evidence is not a claim of experimental agreement or release qualification.

The browser accepts a small local JSON report for reading only. It does not upload that report, execute its contents, run commands, or persist it automatically. Hosted builds contain no raw private runner receipts.

```json
{
  "binding": {
    "revision": "exact-candidate-revision",
    "source_sha256": "64-hex-source-digest",
    "fixture_sha256": "64-hex-fixture-digest",
    "tool": "exact-tool-version"
  },
  "summary": "Description of the recorded check, not an unconditional Pass"
}
```

## Assumptions and provenance

The educational conductance baseline is 30 levels, configurable, not a verified device-state count. Literature coefficients and synthetic regression assertions must retain their own conditions. The Park-context shape is synthetic regression data, not a digitized measured Park loop; unsupported scalar reports and the disputed BTO trace do not become valid by being read here.

Read [What You Can Trust](../../guides/trust.md), the [honesty audit](../../research/honesty-audit.md), the [Park correction](../../research/validation/park-reference-correction.md), and the [citation guide](citations.md).
