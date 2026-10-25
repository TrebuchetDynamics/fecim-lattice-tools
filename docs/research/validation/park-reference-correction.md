# Park reference correction and BTO exclusion

Corrected **25 October 2026**. This audit separates numerical regression from experimental evidence. It does not change model coefficients or smooth reference samples.

## Corrected classification

| Material | Previous presentation | Current treatment |
|---|---|---|
| Park-context 61-point P–E shape | Digitized Park 2015 Figure 2a, used for literature agreement and calibration claims | **Synthetic regression fixture**, `validation/literature/data/park2015_synthetic_shape.csv`; no measured agreement claimed |
| Separate illustrative calibration-shape JSON | Published Figure 2a trace and uncertainties | Renamed `validation/testdata/literature/park2015_synthetic_calibration_shape.json`; synthetic numerical-fit regression only, historical origin unverified and distinct from the CSV |
| Separate Park figure-metrics JSON | Experimental Pr/Ec extraction with manual-digitization uncertainty | **Unverified extraction**, quarantined at `validation/literature/excluded/park2015_unverified_metrics.json`; not assumed to share the CSV's origin |
| BTO/NFO/BTO Figure 7 extraction | Pixel-digitized candidate included in the accepted P–E artifact family | **Excluded pending independent re-digitization**, preserved under `validation/literature/excluded/`; no smoothing |

The Park fixture's numeric CSV body has SHA-256 `8987dce34a3239d0587326820ad168772d85af72aacd6da4bb5e5bff78141f1b`, unchanged from the former file. New `#` header lines and a provenance sidecar state its non-experimental status and the related generation method. The current command uses `NewPreisachModel`, resets it, applies a 31-up/30-down ±3 MV/cm sweep, and converts polarization from C/m² to µC/cm². The exact historical generator revision is **unknown**. Today's preset yields the same field grid but different polarization samples; the related generator does not authenticate the historical samples.

Retaining a synthetic fixture is useful for checking fitting and numerical stability. It is circular evidence if used to demonstrate that the underlying physics model agrees with measured devices.

## Consumer inventory

### Direct numerical and fitting consumers

- `validation/literature/module1_pe_loop_test.go`: renamed the suite to `TestModule1_PELoop_ReferenceRegression`; Park source, dataset ID, notes, threshold-policy label, logs, and emitted evidence are synthetic/regression-only.
- `validation/literature/module1_pe_chisquared_test.go`: labels reference residual diagnostics, not experimental goodness-of-fit. The 2%-of-Ps scale is a diagnostic choice, not measured digitization uncertainty.
- `shared/physics/preisach_calibration_test.go` and `cmd/calibrate-preisach/main_test.go`: use the renamed synthetic fixture; fitting residual is not experimental agreement.
- `shared/physics/preisach_calibration.go`, `cmd/calibrate-preisach/main.go`, and the two reference-test readers: accept `#` provenance comments without treating them as numeric rows.
- `cmd/gen-literature-loops/`: generated references now declare synthetic origin and generation method in their own headers.

The additional `module1-hysteresis/pkg/ferroelectric/preisach_calibration_test.go` consumer formerly named its distinct JSON “PublishedHZOData.” It now requires explicit synthetic/non-experimental metadata, uses `TestFitSyntheticReference_RMSBelow10PercentScale`, and calls its denominator an arbitrary normalization scale rather than measured Pr. Its optimization fits a tanh shape, not a predictive Preisach model. The corresponding physics-realism audit claim is corrected. Numeric shape and fit constants remain unchanged.

### Derived preset and characterization consumers

- `shared/physics/material_calibrated.go`: legacy exported Park API names remain for compatibility; display names and comments explicitly say synthetic regression. Numeric parameters are unchanged.
- `shared/physics/literature_validation_test.go`: separately embedded approximate Park-attributed anchors lacked extraction provenance; now labeled hand-constructed synthetic anchors, not measured Park agreement.
- `validation/literature/forc_reference.go`, `forc_validation_test.go`, `module1_forc_test.go`, and `module1_switching_kinetics_test.go`: Park-context targets and identifiers are synthetic/model expectations, not measured FORC or switching validation.
- `validation/calibration/calibration_targets.go`: the 1 kHz target is explicitly a simulation default, not a figure-extracted Park calibration point.

### Committed artifacts and aggregate consumers

- Renamed the Park P–E, FORC, and switching-shape snapshots under `validation/testdata/validation_artifacts/literature/` to `*_park2015_synthetic_shape.json`; marked synthetic/non-experimental without changing numeric metrics.
- Corrected the Park row and wording in `pe_chisquared_fit.json`.
- Moved the prior BTO P–E result outside the active artifact family to `validation/testdata/excluded/`, preserving its old numerical result as historical, excluded evidence.
- Updated active reference-artifact contracts from nine entries to eight; a regression success is not an experimental-validation count. The Park contract requires explicit classification and `experimental_validation=false`.
- `scripts/run_literature_validation.sh` reports reference regressions, labels even stale legacy Park artifacts synthetic, excludes stale BTO extraction artifacts, and reports zero experimental validations. It uses the actual `rmse_over_ps` field.
- `scripts/ci/validate_regression_artifacts.py` performs structural artifact checks, not measured-agreement promotion. Quarantined snapshots are outside the active artifact root.
- The quarantined Park JSON is outside recursive `experimental-data/` ingestion. `scripts/cognee-ingest-papers.py` now references this correction instead of the unsupported JSON. Existing external knowledge-engine caches were not modified; rebuild/invalidate them before relying on previously ingested claims.

### Scientific and public documentation

Corrected or qualified: citation source/claim/queue records; the calibration, trust, research-workflow and workbench guides; validation README/planned work; scientific claims matrix; both manuscript drafts and paper plans; historical status/changelog; long-term/testing plans; figure readiness and release-preparation documentation. The dated [honesty-audit entry](../honesty-audit.md#corrected-claims--2026-10-25) records the credibility correction explicitly.

## Deliberately retained literature context

A citation to Park 2015 is not automatically a use of this curve. The following are distinct and are not relabeled as synthetic measurements:

- `DefaultHZO` scalar Pr/Ec plausibility checks in `validation/literature/external_benchmarks_test.go` and corresponding range documentation. These are not predictive or held-out curve validation.
- DefaultHZO simulation goldens and release artifacts that mention a Park-context midpoint but do not consume the shape fixture.
- Bibliography entries, ordinary workbench citation keys, source-discovery tests, and historical research acquisition records.

The separate figure-specific scalar attribution still needs primary-source verification; a generic DOI or an old `validated_at` field is insufficient.

## BTO handling and reinstatement gate

The original BTO CSV is retained byte-for-byte. Its provenance and the former result are explicitly excluded. A fail-closed provenance test prevents this candidate from silently re-entering accepted reference checks. The preservation test verifies quarantine metadata; it is not a trace-quality certificate.

Reinstatement requires independent source-figure extraction with calibrated axes, correct branch ordering, original-image provenance, uncertainty and material/measurement conditions, review of redistribution terms, and reproducible comparison. Do not smooth, stitch, rescale, or retune the old trace to make a model fit. Any later model fitting must be labeled calibration, with held-out evidence required for independent validation.
