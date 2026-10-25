# Kim 2020: TiN electrodes and 10 nm HZO

**DOI:** [10.3390/ma13132968](https://doi.org/10.3390/ma13132968)

**Authors:** Si Joon Kim et al.

**Title:** A Comprehensive Study on the Effect of TiN Top and Bottom Electrodes on Atomic Layer Deposited Ferroelectric Hf0.5Zr0.5O2 Thin Films

**Venue:** *Materials* 13(13), 2968 (2020)

**Status:** Source and axis calibration inspected 2026-10-25; no polarization samples extracted and no experimental validation claimed.

## Selected evidence

Use [Figure 3(a)](https://www.mdpi.com/1996-1944/13/13/2968#materials-13-02968-f003), the **LEFT subpanel's red circular Ref. markers**, labelled “RTA after TE dep.” and “After wake-up.” Do not use the blue Annealed BE subpanel, Figure 3(b), connecting lines, or any legacy MDPI/Park CSV.

The [source profile](../../validation/literature/data/experimental/kim2020_fig3a_ref_digitized/source.provenance.json) has the new dataset ID `kim2020_fig3a_ref_digitized`. It records attribution, source identity, measurement limits, and axis anchors only. It does not promote an existing fixture or change any model parameter or legacy validation count.

## Conditions and limits

[Publisher article](https://www.mdpi.com/1996-1944/13/13/2968), sections 2.1–2.2 and section 3's Figure 3 discussion:

| Item | Source-supported description |
|---|---|
| Stack | TiN 90 nm / Hf0.5Zr0.5O2 10 nm / TiN 90 nm |
| Ref. fabrication anneal | 400 °C, 60 s, N2, after top-electrode deposition; no additional bottom-electrode pre-anneal |
| P–E measurement frequency | 10 kHz |
| Wake-up | 10^5 cycles at 2.5 MV/cm |
| Measurement temperature | **Not reported** in the inspected source; 400 °C is the fabrication anneal |
| P–E drive waveform | **Not reported** in the inspected source; do not infer it from the separate pulse write/read experiment |
| Exact plotted device diameter | **Not reported**; section 2.1 gives a fabricated range of 50–200 µm, not this trace's device diameter |

Do not substitute a model's temperature, an assumed waveform, or the midpoint of the diameter range for these unknowns.

## Original raster and acquisition

[Publisher Figure 3 PNG](https://pub.mdpi-res.com/materials/materials-13-02968/article_deploy/html/images/materials-13-02968-g003.png):

- Original dimensions: **3629 × 1210** pixels; **961494** bytes.
- SHA-256: `2016943b0e199bf68e4e56783d7edb42324762b442c816cc66b8c6b19046ddfa`.
- On inspection, the bare URL returned a Cloudflare Polish rewrite: 844940 bytes, SHA-256 `09f2e0cdec0280a5b92b9dc1980fdbfad0a675b99f26956083f5d69cb07062a8`. That version was **not used** for calibration.

Acquire read-only into temporary scratch, bypass the CDN rewrite, and verify the original hash **before selecting pixels**:

```bash
curl -fL -H 'Cache-Control: no-transform' -H 'Accept: image/png' \
  'https://pub.mdpi-res.com/materials/materials-13-02968/article_deploy/html/images/materials-13-02968-g003.png?download=1' \
  -o "$TMPDIR/kim2020-original.png"
sha256sum "$TMPDIR/kim2020-original.png"
```

If the original identity cannot be verified, stop; do not reuse these coordinates with a changed or resized image. No original figure or PDF is committed.

## Axis calibration

Coordinates are **zero-based original-image pixels**, origin at the top-left; x increases rightwards, y downwards. The four anchors are mid-stroke centres of **labelled ticks**, not plot borders. Original-size crops of the left subpanel's bottom and left axes were inspected along with dark-stroke pixel runs.

| Axis | Printed tick | Selected coordinate |
|---|---|---|
| x, electric field | −3 MV/cm | x = 272.5 px |
| x, electric field | +3 MV/cm | x = 905.5 px |
| y, polarization | −45 µC/cm² | y = 939.5 px |
| y, polarization | +45 µC/cm² | y = 217.5 px |

The profile uses machine-readable units `MV/cm` and `uC/cm2`. The two axes define independent affine calibrations; the polarization slope is negative because image y is inverted. Tick-label values are calibration metadata, **not digitized P–E observations**.

A conservative **±2 original pixels per selected coordinate** covers the approximately 3–5-pixel tick-stroke width, antialiasing and manual centre ambiguity. This is an extraction-selection bound, not a fitted standard deviation, measurement uncertainty, or percentage of a model's Ps. Future marker selection needs its own documented bounds and independent visual qualification; this task establishes neither marker accuracy nor trace fidelity.

## Attribution and license evidence

[Publisher-deposited Crossref metadata](https://api.crossref.org/works/10.3390/ma13132968), inspected 2026-10-25, records `license[].content-version = vor` with the URL [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/) and a start date of 2020-07-02. The profile preserves that evidence URL and an attribution identifying the authors, paper, DOI, figure, selected trace and changes.

Attribution: Source: Si Joon Kim et al., *Materials* 13(13), 2968 (2020), DOI 10.3390/ma13132968, Figure 3(a), LEFT red Ref. markers. Copyright 2020 the authors; CC BY 4.0. Changes here are metadata transcription and original-raster tick-centre calibration only. Retain attribution, the license link and an indication of changes in subsequent derivatives. No public promotion or publication approval is established by automated metadata admission.

## Offline reader and verification boundary

[`SourceProfile` and `ReadSourceProfile(io.Reader)`](../../validation/peloop/provenance.go) use schema 1 and standard-library JSON decoding. The reader admits only this qualified dataset/source/trace, with the pinned original hash/dimensions, license evidence, required conditions, explicit not-reported fields, two labelled anchors per axis and a documented positive pixel bound. Incomplete metadata, unsupported units/identity, nonfinite or degenerate anchors, unknown JSON fields and trailing input reject without model defaults or a partially admitted profile.

`Calibration.X` and `.Y` hold `AxisCalibration{Unit, Anchors}`; each `AxisAnchor{Pixel, Value}` records one axis coordinate and its printed physical value. The reader does not fetch the network, check an external image's bytes, digitize samples, compare a model, or certify visual fidelity. Those are separate operations; a metadata test is not experimental agreement.

Focused admission check:

```bash
go test ./validation/peloop -run '^TestSourceProfileAdmission$' -count=1
```

Update this page and the profile together if the source identity, rights evidence, measurement evidence, calibration, or admission contract changes. Independent visual review against the pinned original remains required before public promotion of any future extracted curve.
