# Preisach P–E calibration

Fit a Preisach material preset to a digitized polarization–electric-field loop:

```bash
go run ./cmd/calibrate-preisach \
  -csv validation/literature/data/park2015_fig2a_hzo_10nm.csv \
  -preset default_hzo \
  -out /tmp/park2015-hzo-calibration.json
```

The input CSV columns are `E_MV_cm,P_uC_cm2`. The command prints a human-readable fit summary and writes a schema-versioned JSON artifact containing the fitted parameters, fit-error metrics, preset, input filename, SHA-256 of the exact input bytes, and units. `-out` is optional; when provided, it creates a new file and refuses to overwrite an existing path.

The artifact labels the result as a fit to supplied P–E data, **not device validation**. Preserve the original data and source citation alongside the artifact. The fit is specific to this model, preset, input data, and fitting objective; it does not establish measured-device performance.

This command currently exports calibration evidence but does not automatically rewrite material presets or inject fitted parameters into the research workbench evaluator. Apply fitted parameters explicitly and record that change in the study design until that integration exists.
