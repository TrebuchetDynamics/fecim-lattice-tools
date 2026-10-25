# Excluded reference material

These files are retained for audit, not accepted calibration or experimental validation.

- `bto2021_cryst11101192_hysteresis_digitized.csv`: original numerical trace retained unchanged. Large within-branch jumps require independent re-digitization, not smoothing. Its sidecar records excluded status.
- `park2015_unverified_metrics.json`: separate historical scalar values with unsupported figure-extraction provenance. It is not the 61-point synthetic Park shape and is not presumed to share its origin. It has been removed from `experimental-data/` ingestion.

See [the dated correction inventory](../../../docs/research/validation/park-reference-correction.md). No file in this directory may count as accepted experimental evidence merely because it contains a DOI or old validation timestamp.
