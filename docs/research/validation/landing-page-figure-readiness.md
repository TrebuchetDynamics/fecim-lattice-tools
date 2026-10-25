# Landing-page validation figure: evidence review

Status: **not ready for a headline measured-versus-simulated overlay**.
Reviewed 25 October 2026. This is an evidence inventory, not new physics validation.

## Park HZO fixture

- [The generator](../../../cmd/gen-literature-loops/main.go#L1) explicitly writes **synthetic** P–E loops from built-in Park and Cheema presets.
- [The P–E reference-regression test](../../../validation/literature/module1_pe_loop_test.go) now explicitly classifies Park as synthetic, not literature agreement.
- [Corrected Park provenance](../../../validation/literature/data/park2015_synthetic_shape.provenance.json) labels the reference synthetic/non-experimental and records that the exact historical generator revision is unknown. Its former sidecar had only a dataset ID and DOI, not extraction evidence.
- The [workbench guide](../../guides/research-design-workbench.md#trust-boundary) already distinguishes provisional model-generated fixtures from held-out validation.

Consequently, a low RMSE against this fixture must not be presented as experimental agreement. A fresh `go run ./cmd/gen-literature-loops -preset park -out <scratch.csv>` produces the same 61 field samples but **does not reproduce the committed polarization values byte-for-byte** with the current preset (pointwise polarization RMSE ≈1.0674 µC/cm²). This comparison is not experimental RMSE and does not reconstruct the historical generator version. The generator is evidence of a synthetic-reference workflow, not proof of the exact origin of every committed point. The fixture, consumers, and reports have now been reclassified as synthetic regression evidence. Numeric samples and model coefficients are unchanged; only provenance, labels, paths, and evidence handling were corrected. See [the dated correction audit](park-reference-correction.md).

## BTO candidate

The [quarantined BTO provenance](../../../validation/literature/excluded/bto2021_cryst11101192_hysteresis_digitized.provenance.json) identifies Figure 7 of [Crystals 2021, 11, 1192](https://www.mdpi.com/2073-4352/11/10/1192), reports 121 extracted points and an uncertainty envelope, and formerly labeled it a candidate Tier-1 dataset. It is now explicitly excluded pending independent re-digitization.

This is **BaTiO3/NiFe2O4/BaTiO3**, not HZO. It cannot validate the HZO system evaluator. The article describes a 140/80/140 nm trilayer and P–E measurement at 4 kHz.

The [unchanged quarantined CSV](../../../validation/literature/excluded/bto2021_cryst11101192_hysteresis_digitized.csv) also contains large adjacent changes within its field branches. For example, field increases from approximately 0.321 to 0.343 MV/cm while polarization changes from 16.206 to 0.375 µC/cm². On the descending branch, approximately −0.235 to −0.261 MV/cm changes polarization from −15.902 to +6.440 µC/cm². These observations warrant checking extraction order, axis scaling, and trace selection against the original figure; they do not by themselves identify the cause.

The [preservation test](../../../validation/literature/bto_digitized_integrity_test.go) checks the retained extraction and excluded status, not trace fidelity. The trace and its prior result are outside the active reference-artifact family; a fail-closed provenance check prevents reuse. Do not smooth this reference to make it fit. Re-digitize and independently review the original figure before any reinstatement.

## Acceptance criteria for a future plot

1. Identify an accessible original measurement source, exact figure/panel, material stack, thickness, waveform, frequency, temperature, and redistribution terms.
2. Preserve a traceable extraction artifact and script, calibrated axes, units, branch ordering, point count, and uncertainty.
3. Disclose which parameters were fitted to these points. Call an overlay a **calibration comparison** if those points trained the model; reserve independent-validation language for held-out evidence.
4. Generate the simulation from a recorded commit and explicit preset/settings, with no manual curve reshaping.
5. Compute RMSE in µC/cm² at matched fields **within the same branch**, state interpolation and endpoint handling, and report the number of points. Report normalized RMSE only with its explicit denominator.
6. Produce the plot, metric JSON, data hashes, and caption from one reproducible command. The caption must state the comparison's limits.
7. Add focused failing tests before any new extraction, comparison, or plotting behavior.

Until these conditions are met, retain the README's reproducibility example and describe the available regression/literature checks without claiming a physics proof.

## Public-source refresh for outreach

The [Tour lab publication list](https://jmtour.com/publications-dr-james-tour/) and [IronLattice site](https://www.ironlattice.com/) were inspected during preparation. The company site is positioning/marketing, not an experimental dataset. This inspection does not establish an exhaustive post-January literature search or confirm a new measurement paper suitable for calibration. Recheck dated primary publications before outreach; do not reuse promotional performance numbers as simulator evidence.
