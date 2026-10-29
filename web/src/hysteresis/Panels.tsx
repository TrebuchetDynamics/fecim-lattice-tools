import { useEffect, useRef, useState } from "react";
import { Engine } from "./engine";
import { Field } from "./Field";
import { GoData, ResultsTable } from "./GoData";
import type { Budget, FORC, ProgramConfig, Results } from "./types";

export function ProgramPanel({
  data,
  engine,
  settingsStale,
}: {
  data: Results;
  engine: Engine;
  settingsStale: boolean;
}) {
  const cat = data.catalogue!;
  const [config, setConfig] = useState<ProgramConfig>(cat.Program);
  const [budget, setBudget] = useState<Budget>(cat.ProgramBudget);
  const [target, setTarget] = useState(String(cat.Target));
  const program = data.program;
  const locked = !!data.active || data.failed || settingsStale;
  const programStale =
    !!program && JSON.stringify(config) !== JSON.stringify(program.Config);
  const canQueue =
    data.active === "program" && program?.Config.Mode === "write-read";
  return (
    <details>
      <summary>Programming and calibration</summary>
      <p>
        Select the core Manual drive and apply settings before programming.
        Discrete levels are configurable educational bins, not demonstrated
        analog device states. Calibration is in-memory simulation; “fresh” does
        not mean every target converged. Inspect fallback counts and reasons
        below.
      </p>
      <form
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          if (!locked) void engine.programNew(config);
        }}
      >
        <fieldset disabled={locked}>
          <legend>Programming protocol</legend>
          <div className="lab-grid">
            <Field label="Programming mode">
              <select
                value={config.Mode}
                onChange={(event) =>
                  setConfig({ ...config, Mode: event.target.value })
                }
              >
                {cat.ProgramModes.map((value) => (
                  <option key={value}>{value}</option>
                ))}
              </select>
            </Field>
            <Field label="ISPP method">
              <select
                value={config.Method}
                onChange={(event) =>
                  setConfig({ ...config, Method: event.target.value })
                }
              >
                {cat.Methods.map((value) => (
                  <option key={value}>{value}</option>
                ))}
              </select>
            </Field>
            {(
              [
                "Domains",
                "WorkLimit",
                "PulseLimit",
                "CalibrationWorkLimit",
                "GuardFraction",
              ] as const
            ).map((key) => (
              <label key={key}>
                {key}
                <input
                  type="number"
                  step="any"
                  value={config[key]}
                  onChange={(event) =>
                    setConfig({ ...config, [key]: Number(event.target.value) })
                  }
                />
              </label>
            ))}
          </div>
          <button type="submit">Initialize programmer</button>
        </fieldset>
      </form>
      {settingsStale && (
        <p>
          Apply edited loop settings before starting programming or calibration.
        </p>
      )}
      {programStale && (
        <p>
          Programmer settings changed; initialize the programmer before using
          this configuration.
        </p>
      )}
      {program && (
        <>
          <div className="metric-strip">
            <p>
              Program status
              <strong data-testid="program-status">{program.Status}</strong>
            </p>
            <p>
              Phase<strong>{program.Phase}</strong>
            </p>
            <p>
              Target / queued
              <strong>
                {program.Target} / {program.QueuedTarget}
              </strong>
            </p>
            <p>
              Read level<strong>{program.ReadLevel}</strong>
            </p>
            <p>
              Calibration
              <strong data-testid="calibration-status">
                {program.Calibration.Status}
              </strong>
            </p>
          </div>
          {program.Reason && <p>{program.Reason}</p>}
          <p>{program.Calibration.Reason}</p>
          <fieldset
            disabled={
              settingsStale ||
              programStale ||
              (locked && !canQueue) ||
              program.Status === "legacy_protocol_unavailable"
            }
          >
            <legend>Explicit programming actions</legend>
            <div className="lab-grid">
              <Field
                label="Target level"
                hint="Native 1-based target; the core’s current level is zero-based."
              >
                <input
                  type="number"
                  step="1"
                  value={target}
                  onChange={(event) => setTarget(event.target.value)}
                />
              </Field>
              <label>
                Program seconds per step
                <input
                  disabled={locked}
                  type="number"
                  step="any"
                  value={budget.DT}
                  onChange={(event) =>
                    setBudget({ ...budget, DT: Number(event.target.value) })
                  }
                />
              </label>
              <label>
                Program steps per batch
                <input
                  disabled={locked}
                  type="number"
                  step="1"
                  value={budget.Steps}
                  onChange={(event) =>
                    setBudget({ ...budget, Steps: Number(event.target.value) })
                  }
                />
              </label>
            </div>
          </fieldset>
          <div className="actions">
            <button
              aria-disabled={
                settingsStale ||
                programStale ||
                (locked && !canQueue) ||
                program.Status === "legacy_protocol_unavailable"
              }
              onClick={() => {
                if (
                  !settingsStale &&
                  !programStale &&
                  (!locked || canQueue) &&
                  program.Status !== "legacy_protocol_unavailable"
                )
                  engine.programAction("write-target", Number(target), budget);
              }}
            >
              Program target
            </button>
            <button
              aria-disabled={
                locked ||
                programStale ||
                program.Status === "legacy_protocol_unavailable"
              }
              onClick={() => {
                if (
                  !locked &&
                  !programStale &&
                  program.Status !== "legacy_protocol_unavailable"
                )
                  engine.programAction("calibrate", 0, budget);
              }}
            >
              Calibrate in memory
            </button>
          </div>
          <details>
            <summary>Controller, calibration values and availability</summary>
            <GoData
              value={{
                Config: program.Config,
                Controller: program.Controller,
                Calibration: program.Calibration,
                Work: program.Work,
                OwnsDrive: program.OwnsDrive,
              }}
            />
          </details>
        </>
      )}
    </details>
  );
}

export function AnalysisPanel({
  data,
  engine,
  settingsStale,
}: {
  data: Results;
  engine: Engine;
  settingsStale: boolean;
}) {
  const cat = data.catalogue!;
  const [kind, setKind] = useState(cat.Diagnostics[0]);
  const [forc, setFORC] = useState<FORC>(cat.FORC);
  const [reference, setReference] = useState<{
    Format: string;
    Source: string;
    Material: string;
    Data: string;
  }>();
  const [fileError, setFileError] = useState("");
  const [reading, setReading] = useState(false);
  const [referenceName, setReferenceName] = useState("");
  const referenceRead = useRef(0);
  const fileInput = useRef<HTMLInputElement>(null);
  useEffect(
    () => () => {
      referenceRead.current++;
    },
    [],
  );
  const [source, setSource] = useState("");
  const [submittedInputs, setSubmittedInputs] = useState<string>();
  const inputs = JSON.stringify([
    kind,
    forc,
    reference,
    source,
    reading,
    fileError,
  ]);
  const caption = useRef<HTMLTableCaptionElement>(null);
  const result = data.analysis,
    page = data.analysisPage;
  const [referencePage, setReferencePage] = useState(0);
  useEffect(() => setReferencePage(0), [result]);
  async function readReference(file?: File) {
    const token = ++referenceRead.current;
    setFileError("");
    setReference(undefined);
    setReferenceName(file?.name || "");
    setReading(!!file);
    if (!file) return;
    try {
      // Base64 and metadata still require Go envelope/reference validation.
      if (file.size > cat.ReferenceBytes)
        throw new Error("Reference exceeds Go byte limit");
      const bytes = new Uint8Array(await file.arrayBuffer());
      if (token !== referenceRead.current) return;
      let binary = "";
      for (const byte of bytes) binary += String.fromCharCode(byte);
      const csv = file.name.toLowerCase().endsWith(".csv");
      setReference({
        Format: csv ? "csv" : "json",
        // JSON provenance belongs to Go's parser, not invented UI metadata.
        Source: csv ? file.name : "",
        Material: csv ? data.core!.Material.ID : "",
        Data: btoa(binary),
      });
    } catch (error) {
      if (token === referenceRead.current) setFileError(String(error));
    } finally {
      if (token === referenceRead.current) setReading(false);
    }
  }
  const references = result?.References as unknown as
    | { Source: string; Points: unknown[]; [key: string]: unknown }[]
    | undefined;
  const stale =
    !!result && result.Context.ConfigRevision !== data.core!.ConfigRevision;
  return (
    <details>
      <summary>Diagnostics and reference context</summary>
      <p>
        Independent Preisach diagnostics; selected loop model, temperature and
        stress are context only, not applied. PUND uses the legacy demo profile.
        FORC uses the existing GUI diagnostic, not a physical L-K experiment.
        Applied parameters and evidence come from Go.
      </p>
      <form
        noValidate
        onSubmit={(event) => {
          event.preventDefault();
          if (
            settingsStale ||
            data.active ||
            data.failed ||
            reading ||
            fileError
          )
            return;
          setSubmittedInputs(inputs);
          engine.analyze({
            Kind: kind,
            ...(kind === cat.Diagnostics[1] ? { FORC: forc } : {}),
            ...(reference
              ? {
                  References: [
                    { ...reference, Source: source || reference.Source },
                  ],
                }
              : {}),
          });
        }}
      >
        <fieldset disabled={!!data.active || data.failed || settingsStale}>
          <legend>Diagnostic inputs</legend>
          <Field label="Diagnostic">
            <select
              value={kind}
              onChange={(event) => setKind(event.target.value)}
            >
              {cat.Diagnostics.map((value) => (
                <option key={value}>{value}</option>
              ))}
            </select>
          </Field>
          {kind === cat.Diagnostics[1] && (
            <div className="lab-grid">
              {(["EmaxVm", "Reversals", "Resolution"] as const).map((key) => (
                <label key={key}>
                  {key}
                  <input
                    type="number"
                    step="any"
                    value={forc[key]}
                    onChange={(event) =>
                      setFORC({ ...forc, [key]: Number(event.target.value) })
                    }
                  />
                </label>
              ))}
            </div>
          )}
          <label>
            Reference source or citation
            <input
              value={source}
              onChange={(event) => setSource(event.target.value)}
            />
          </label>
          <label>
            Reference branch file (CSV or JSON)
            <input
              type="file"
              ref={fileInput}
              accept=".csv,.json"
              onChange={(event) => void readReference(event.target.files?.[0])}
            />
            <span className="hint">
              Read only into memory; not uploaded, saved, fitted or verified.
              CSV header E_MV_cm,P_uC_cm2; JSON source/material/E_V_m/P_C_m2.
              One monotonic branch.
            </span>
          </label>
          {reading && <p role="status">Reading reference into memory…</p>}
          {reference && (
            <p>
              Reference ready in memory: {referenceName}. Go validates on
              Analyze.
            </p>
          )}
          {fileError && <p role="alert">{fileError}</p>}
        </fieldset>
        <button
          type="submit"
          className="primary"
          aria-disabled={
            !!data.active ||
            data.failed ||
            settingsStale ||
            reading ||
            !!fileError
          }
        >
          Analyze
        </button>
        <button
          type="button"
          disabled={!!data.active}
          onClick={() => {
            referenceRead.current++;
            setReading(false);
            setReference(undefined);
            setReferenceName("");
            setFileError("");
            if (fileInput.current) fileInput.current.value = "";
          }}
        >
          Clear reference
        </button>
      </form>
      {settingsStale && (
        <p>
          Apply edited loop settings before starting diagnostics; retained
          diagnostics use their original applied settings.
        </p>
      )}
      {data.progress && (
        <p>
          Diagnostic state:{" "}
          <strong data-testid="analysis-status">{data.progress.State}</strong> ·
          Work units {data.progress.Completed} of {data.progress.Total}
        </p>
      )}
      {result && submittedInputs !== inputs && (
        <p data-testid="diagnostic-settings-state">
          Edited diagnostic inputs; retained results stale until Analyze runs
          with these inputs.
        </p>
      )}
      {stale && (
        <p>
          Diagnostic results stale: loop settings changed. Results retain their
          original context and revisions.
        </p>
      )}
      {result && (
        <>
          <section
            data-testid="analysis-context"
            aria-label="Diagnostic identity and evidence"
          >
            <h3>Diagnostic identity and context</h3>
            {result.Notices.map((text) => (
              <p key={text}>{text}</p>
            ))}
            <GoData
              values={result.ScientificValues}
              value={{
                Diagnostic: {
                  ...(result.Diagnostic as Record<string, unknown>),
                  Applied: result.AppliedParameters,
                },
                Context: result.Context,
                Evidence: result.Evidence,
                ExperimentalValidation: result.ExperimentalValidation,
                ReferenceContext: result.ReferenceContext,
                PUNDCharges: result.PUND?.Charges,
              }}
            />
          </section>
          <Field label="Analysis table">
            <select
              value={data.table}
              disabled={!!data.active}
              onChange={(event) =>
                void engine.pageAnalysis(event.target.value, 0)
              }
            >
              {cat.Tables[result.Kind].map((value) => (
                <option key={value}>{value}</option>
              ))}
            </select>
          </Field>
          {page && (
            <>
              <ResultsTable
                id="analysis-table"
                page={page}
                keys={cat.AnalysisColumns[data.table!].map(
                  (column) => column.Key,
                )}
                labels={Object.fromEntries(
                  cat.AnalysisColumns[data.table!].map((column) => [
                    column.Key,
                    column.Label,
                  ]),
                )}
                caption={
                  (data.table || "") +
                  " · Rows " +
                  (page.Total ? page.Offset + 1 : 0) +
                  "–" +
                  (page.Offset + page.Rows.length) +
                  " of " +
                  page.Total +
                  " · Go revision " +
                  page.Revision
                }
                captionRef={caption}
              />
              <div className="actions">
                <button
                  disabled={!!data.active || page.Offset === 0}
                  onClick={() =>
                    void engine
                      .pageAnalysis(
                        data.table!,
                        Math.max(0, page.Offset - cat.PageSize),
                      )
                      .then(() => caption.current?.focus())
                  }
                >
                  Previous analysis page
                </button>
                <button
                  disabled={
                    !!data.active || page.Offset + cat.PageSize >= page.Total
                  }
                  onClick={() =>
                    void engine
                      .pageAnalysis(data.table!, page.Offset + cat.PageSize)
                      .then(() => caption.current?.focus())
                  }
                >
                  Next analysis page
                </button>
                <button
                  disabled={!!data.active || !page.Total}
                  onClick={() =>
                    void engine
                      .pageAnalysis(
                        data.table!,
                        Math.floor((page.Total - 1) / cat.PageSize) *
                          cat.PageSize,
                      )
                      .then(() => caption.current?.focus())
                  }
                >
                  Last analysis page
                </button>
              </div>
              <details>
                <summary>Canonical diagnostic page</summary>
                <GoData
                  value={{
                    Revision: page.Revision,
                    Offset: page.Offset,
                    Total: page.Total,
                    Rows: page.Rows.map((row) =>
                      Object.fromEntries(
                        cat.AnalysisColumns[data.table!].map((column) => [
                          column.Label,
                          row[column.Key],
                        ]),
                      ),
                    ),
                  }}
                />
              </details>
            </>
          )}
          {references?.map((ref, index) => (
            <section
              key={index}
              aria-label={"Imported reference " + ref.Source}
            >
              <h3>{ref.Source} — imported, unverified</h3>
              <GoData
                value={{
                  ...ref,
                  Points: ref.Points.slice(
                    referencePage * cat.PageSize,
                    (referencePage + 1) * cat.PageSize,
                  ),
                }}
              />
              <p>
                Reference rows {referencePage * cat.PageSize + 1}–
                {Math.min(
                  (referencePage + 1) * cat.PageSize,
                  ref.Points.length,
                )}{" "}
                of {ref.Points.length}
              </p>
              <div className="actions">
                <button
                  disabled={!referencePage}
                  onClick={() => setReferencePage(referencePage - 1)}
                >
                  Previous reference page
                </button>
                <button
                  disabled={
                    (referencePage + 1) * cat.PageSize >= ref.Points.length
                  }
                  onClick={() => setReferencePage(referencePage + 1)}
                >
                  Next reference page
                </button>
              </div>
            </section>
          ))}
        </>
      )}
    </details>
  );
}
