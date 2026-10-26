# Kim 2020 Figure 3(a): independent red Ref. marker extraction

## Source and limits

Source: Si Joon Kim et al., *Materials* 13(13), 2968 (2020),
[DOI 10.3390/ma13132968](https://doi.org/10.3390/ma13132968), Figure 3(a),
**LEFT red circular Ref. markers**, “RTA after TE dep.” / “After wake-up.”
Copyright 2020 the authors; [CC BY 4.0](https://creativecommons.org/licenses/by/4.0/).
Changes: independently selected original-raster marker centres and affine conversion
to physical units. The original figure/PDF is not redistributed here.

The unchanged [admitted profile](source.provenance.json) owns source identity,
license evidence, measurement conditions and tick-centre calibration. The
[citation record](../../../../../citations/papers/kim2020_tin_hzo_electrodes.md)
explains acquisition and unreported measurement conditions. This is a numerical
extraction, **not experimental validation of the simulator**. No model, material
preset, residual or legacy MDPI/Park sample informed the selection. No legacy CSV
was used, promoted or rewritten. No blue Annealed BE trace, panel 3(b), connecting
line sample, smoothing, interpolation or completion of missing markers is included.

## Selection procedure

1. Reacquire the [publisher PNG](https://pub.mdpi-res.com/materials/materials-13-02968/article_deploy/html/images/materials-13-02968-g003.png)
   into `$TMPDIR`, then verify **961494 bytes, 3629 × 1210 pixels**, and the source
   SHA-256 below **before selecting coordinates**. The CDN may return a different
   optimized PNG even with `?download=1`; this extraction obtained the pinned bytes
   with `?download=1&original=HZO-REF-3`, `Cache-Control: no-cache, no-transform`,
   and `Accept: image/png`. A query suffix does not establish identity; the hash does.
2. Inspect the left plot at original size and enlarged crops. Its plot area is
   approximately x=220..959, y=158..1000 in the original image. All committed
   coordinates are zero-based original pixels: origin top-left, x right, y down.
   Enlargement is for inspection only, never a new coordinate system.
3. Identify red outlined circular/elliptical markers by their white interiors and
   visible perimeter. As a selection aid, inspect enclosed non-red components in
   the original RGB raster (red mask: R-G > 20 and R-B > 20). For isolated markers,
   take the midpoint of the interior bounding extrema, to half-pixel resolution.
   Reject text holes and connecting strokes by visual inspection. The mask is not
   a curve sampler and provides no points along connecting lines.
4. For partially overlapping markers, inspect the remaining outer arcs/sidewalls
   directly. A clipped white-interior bounding box is **not** the marker centre:
   it biases x toward the surviving crescent. The descending points 10..13 use
   visible-arc centre estimates rather than those clipped-hole midpoints. Ascending
   point 11 at (772.5,358.5) is the lower of two overlapping rings: its sidewalls
   and bottom arc remain distinguishable from descending point 2 at (772.5,345.5).
   These five partially overlapped selections carry the wider bound below.
5. Assign the lower branch to **ascending E**, and the upper branch to
   **descending E**, in strictly monotone field order, with zero-based consecutive
   indices. This is a conventional quasi-static branch assignment from plot geometry,
   **not a recovered measurement time sequence**. Opposite branches can share x.

There are **12 ascending + 14 descending = 26 selected markers**. The two apparent
positive saturation positions near (813,334) and (851,326) do not resolve separate
lower/upper centres. Each visible centre is retained once, on the descending
continuation; no ascending duplicate or invented offset is added. Thus the ascending
branch ends at the last distinguishable lower ring near +1.74 MV/cm / +27.42 uC/cm2,
not at the apparently merged +2.12/+2.48 MV/cm endpoints. At the negative end the
paired rings retain distinguishable arcs, including descending point 13 and
ascending point 0; no shared turnaround sample is invented. No unseen markers are
claimed. Endpoint omissions make the branch counts and spans unequal.

## Coverage and selection bounds

| Branch | Count | E span (MV/cm) | P span (uC/cm2) | Adjacent P=0 bracket (indices; E in MV/cm) |
|---|---:|---|---|---|
| ascending | 12 | -2.2464 to +1.7393 | -30.9141 to +27.4238 | 9..10; +0.9858 to +1.3223 |
| descending | 14 | +2.4834 to -2.3744 | +31.4751 to -31.1634 | 8..9; -0.4929 to -0.8483 |

Both branches retain distinct finite fields, both E/P signs, negative saturation,
positive saturation up to the resolvable rings, and the visible switching crossing.
The brackets describe adjacent observed markers only; **no interpolated zero-crossing
sample** is stored. The table is rounded for reading; the CSV retains converter output.

Use a conservative **+/-2 original pixels per axis for isolated marker centres**, and
**+/-4 pixels per axis for ascending 11 and descending 10..13**, covering partial
occlusion and manual perimeter-centre ambiguity. Half-pixel coordinates do not imply
half-pixel accuracy. The profile's separate **+/-2-pixel tick-selection bound** is
unchanged; it did not establish marker-centre accuracy.

With the 633-pixel x and 722-pixel y anchor separations, the marker-only bounds
are about +/-0.0190 MV/cm and +/-0.2493 uC/cm2 (isolated), or +/-0.0379 MV/cm and
+/-0.4986 uC/cm2 (partial overlaps). A conservative bound including independent
+/-2-pixel uncertainty on each anchor is `range*(marker_bound+2)/(separation-4)`
for these in-anchor points: less than +/-0.0382 MV/cm / +/-0.5014 uC/cm2 (isolated)
and +/-0.0573 MV/cm / +/-0.7521 uC/cm2 (partial overlaps). These are **extraction
selection bounds**, not statistical standard deviations, experimental instrument
uncertainties, measured Ps fractions or fitted error bars. No extra decimal places
in the generated CSV change them. Independent human source/trace-fidelity review
against the pinned original remains required before public promotion.

## Offline reproduction and integrity receipt

From the repository root, with an output path that does not already exist:

```sh
go run ./cmd/pe-reference digitize \
  --profile validation/literature/data/experimental/kim2020_fig3a_ref_digitized/source.provenance.json \
  --pixels validation/literature/data/experimental/kim2020_fig3a_ref_digitized/points.pixels.csv \
  --out "$TMPDIR/kim2020-loop-regenerated.csv"
cmp "$TMPDIR/kim2020-loop-regenerated.csv" \
  validation/literature/data/experimental/kim2020_fig3a_ref_digitized/loop.csv
go test ./validation/literature -run '^TestHZOIndependentDigitization$' -count=1
```

The converter is offline, applies the two admitted affine axes without fitting,
and emits deterministic LF-terminated physical CSV. Regeneration needs only the
committed profile and [raw pixels](points.pixels.csv), not the image or old samples.
The [physical CSV](loop.csv) is unedited converter output. SHA-256 covers complete
file bytes, including headers and final newlines. The focused test reads this receipt,
checks all hashes/counts, regenerates bytes, admits physical samples, and rejects
inadequate coverage. It does **not** certify that the selected centres are visually
correct or that a model agrees with measurements.

```json
{
  "dataset_id": "kim2020_fig3a_ref_digitized",
  "source_sha256": "2016943b0e199bf68e4e56783d7edb42324762b442c816cc66b8c6b19046ddfa",
  "sha256": {
    "source.provenance.json": "4cd938944d1aa612549eb000b226704624953e566cdb3b2c8674afa0a17687da",
    "points.pixels.csv": "57ea1f4582646f919da4bb8668337dc696ce38d469fa49372c61ce6374f14dce",
    "loop.csv": "46e38d320b96cf26390c5cef45c8e65a019c74c5766bd9f26aa541008f39fc61"
  },
  "counts": {"ascending": 12, "descending": 14}
}
```
