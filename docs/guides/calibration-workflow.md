# Preisach P–E calibration

Exercise Preisach fitting with a **synthetic regression shape**, not a digitized Park measurement:

```bash
go run ./cmd/calibrate-preisach \
  -csv validation/literature/data/park2015_synthetic_shape.csv \
  -preset default_hzo \
  -out /tmp/park-synthetic-calibration.json
```

The sample's historical generator revision is unknown; numeric samples are retained for regression only. Its provenance header records the related model-generation workflow. A small fit residual is not agreement with Park 2015. For device calibration, supply independently traceable measured data instead.

The input CSV columns are `E_MV_cm,P_uC_cm2`. Lines beginning with `#` are provenance comments. The command prints a human-readable fit summary and writes a schema-versioned JSON artifact containing the fitted parameters, fit-error metrics, preset, input filename, SHA-256 of the exact input bytes, and units. `-out` is optional; when provided, it creates a new file and refuses to overwrite an existing path.

The artifact labels the result as a fit to supplied P–E data, **not device validation**. Preserve the original data and source citation alongside the artifact. The fit is specific to this model, preset, input data, and fitting objective; it does not establish measured-device performance.

This command currently exports calibration evidence but does not automatically rewrite material presets or inject fitted parameters into the research workbench evaluator. Apply fitted parameters explicitly and record that change in the study design until that integration exists.
