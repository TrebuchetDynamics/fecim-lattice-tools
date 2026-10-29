import { useEffect, useRef, useState } from "react";
import { Engine } from "./engine";
import { Field } from "./Field";
import { GoData, QuantityValue, ResultsTable } from "./GoData";
import { Plot } from "./Plot";
import { LocalActions } from "./LocalActions";
import { ProgramPanel, AnalysisPanel } from "./Panels";
import type { Budget, Config, Results } from "./types";

export default function Hysteresis() {
  const [data, setData] = useState<Results>({
    revision: 0,
    status: "Loading Go Worker…",
    failed: false,
  });
  const engine = useRef<Engine | null>(null);
  const [draft, setDraft] = useState<Config>();
  const lastApplied = useRef<Config | undefined>(undefined);
  const [budget, setBudget] = useState<Budget>();
  const [manual, setManual] = useState("0");
  const coreCaption = useRef<HTMLTableCaptionElement>(null);
  const resultsHeading = useRef<HTMLHeadingElement>(null);
  useEffect(() => {
    const current = (engine.current = new Engine(setData));
    void current.start();
    return () => current.dispose();
  }, []);
  useEffect(() => {
    if (!data.core) {
      lastApplied.current = undefined;
      setDraft(undefined);
      return;
    }
    const previous = lastApplied.current;
    setDraft((draft) =>
      !draft || JSON.stringify(draft) === JSON.stringify(previous)
        ? data.core!.Config
        : draft,
    );
    lastApplied.current = data.core.Config;
  }, [data.core?.ConfigRevision]);
  useEffect(() => {
    if (data.catalogue) setBudget(data.catalogue.LoopBudget);
  }, [data.catalogue]);
  const core = data.core,
    cat = data.catalogue;
  const stale =
    !!draft && !!core && JSON.stringify(draft) !== JSON.stringify(core.Config);
  const locked = !!data.active || data.failed;
  const ready = !!core && !!draft && !!cat && !!budget;
  const setting = (key: keyof Config, value: string) =>
    setDraft((previous) =>
      previous
        ? {
            ...previous,
            [key]: typeof previous[key] === "number" ? Number(value) : value,
          }
        : previous,
    );
  return (
    <>
      <div className="lab-header">
        <p className="eyebrow">
          01 / Browser research · {__FECIM_PROFILE__} profile
        </p>
        <h1 className="lab-title">Hysteresis lab</h1>
        <p>
          Stateful Go physics. Table-first results. Ephemeral, education-phase
          simulation — not device measurements.
        </p>
      </div>
      <aside className="notice" aria-label="Computation and privacy">
        <p>
          Computations run in your browser using the Go model.
          {__FECIM_PROFILE__ === "hosted"
            ? " Project and result data is not uploaded. On-disk projects and edits, large sweeps, run reuse, stored-run verification, EDA and file writes require the local application: run fecim-lattice-tools serve --root DIRECTORY (design shorthand: fecim serve). This hosted build does not contact localhost."
            : " Explicit local file actions send browser snapshots only to this authenticated same-origin server and its launch-granted roots. They do not verify persisted studies. No automatic file writes."}
        </p>
        <p>
          Ordinary static asset requests still occur and the host can see HTTP
          metadata. No analytics or remote solver. Hosted and local profiles use
          the same Hysteresis views and browser Go Worker.
        </p>
      </aside>
      <p
        className="status-line"
        role="status"
        aria-live="polite"
        aria-atomic="true"
      >
        {data.status}
      </p>
      {data.error && (
        <p role="alert" className="error">
          {data.error}
        </p>
      )}
      <div className="actions">
        <button onClick={() => void engine.current?.start()}>
          Discard and restart Worker
        </button>
        {(data.failed || (!core && data.error)) && (
          <button onClick={() => void engine.current?.start()}>
            Retry Worker
          </button>
        )}
      </div>
      {ready && (
        <>
          <form
            noValidate
            onSubmit={(event) => {
              event.preventDefault();
              void engine.current!.apply({ Kind: "configure", Config: draft });
            }}
          >
            <fieldset disabled={locked}>
              <legend>Material and drive</legend>
              <div className="lab-grid">
                <Field label="Material">
                  <select
                    value={draft.MaterialID}
                    onChange={(event) => {
                      const material = core.Catalogue.find(
                        (item) => item.ID === event.target.value,
                      )!;
                      setDraft({
                        ...draft,
                        MaterialID: material.ID,
                        Levels: material.DefaultLevels,
                        RangeFraction: material.DefaultRangeFraction,
                      });
                    }}
                  >
                    {core.Catalogue.map((item) => (
                      <option value={item.ID} key={item.ID}>
                        {item.Parameters.Name}
                      </option>
                    ))}
                  </select>
                </Field>
                <Field label="Model">
                  <select
                    value={draft.Model}
                    onChange={(event) => setting("Model", event.target.value)}
                  >
                    {core.Models.map((value) => (
                      <option key={value}>{value}</option>
                    ))}
                  </select>
                </Field>
                <Field label="Drive">
                  <select
                    value={draft.Mode}
                    onChange={(event) => setting("Mode", event.target.value)}
                  >
                    {core.Modes.map((value) => (
                      <option key={value}>{value}</option>
                    ))}
                  </select>
                </Field>
                {(
                  [
                    ["Frequency", "Frequency"],
                    ["TimeScale", "TimeScale"],
                    ["Temperature", "Temperature"],
                    ["Stress", "Stress"],
                    ["Levels", "Levels"],
                    ["RangeFraction", "RangeFraction"],
                    ["HistoryLimit", "History"],
                  ] as const
                ).map(([key, controlKey]) => {
                  const control = core.Controls[controlKey];
                  return (
                    <Field
                      key={key}
                      label={
                        (key === "TimeScale"
                          ? "Playback scale"
                          : key === "RangeFraction"
                            ? "Range fraction"
                            : key === "HistoryLimit"
                              ? "History limit"
                              : key) +
                        " (" +
                        control.Unit +
                        ")"
                      }
                      hint={control.Reason || "SI input; validated by Go"}
                    >
                      <input
                        name={key}
                        type="number"
                        step="any"
                        min={control.Min}
                        max={control.Max}
                        value={draft[key]}
                        disabled={!control.Enabled}
                        onChange={(event) => setting(key, event.target.value)}
                      />
                    </Field>
                  );
                })}
                <label>
                  Seed
                  <input
                    name="Seed"
                    type="number"
                    step="1"
                    value={draft.Seed}
                    onChange={(event) => setting("Seed", event.target.value)}
                  />
                </label>
              </div>
              <button type="submit" className="primary">
                Apply settings
              </button>
            </fieldset>
          </form>
          <p data-testid="settings-state">
            {stale
              ? "Edited settings; previous results stale until applied and recomputed"
              : "Current settings"}
          </p>
          <fieldset disabled={locked}>
            <legend>Physical stepping and manual field</legend>
            <p className="hint">
              DT is explicit physical seconds per Go step, independent of
              playback scale. The opening batch schedule is generated by Go.
              Changing frequency does not silently retune DT.
            </p>
            <div className="lab-grid">
              <label>
                Seconds per step
                <input
                  type="number"
                  step="any"
                  value={budget.DT}
                  onChange={(event) =>
                    setBudget({ ...budget, DT: Number(event.target.value) })
                  }
                />
              </label>
              <label>
                Steps per batch
                <input
                  type="number"
                  step="1"
                  value={budget.Steps}
                  onChange={(event) =>
                    setBudget({ ...budget, Steps: Number(event.target.value) })
                  }
                />
              </label>
              <Field
                label={"Manual field (" + core.Controls.ManualField.Unit + ")"}
                hint={core.Controls.ManualField.Reason}
              >
                <input
                  type="number"
                  step="any"
                  value={manual}
                  disabled={!core.Controls.ManualField.Enabled}
                  onChange={(event) => setManual(event.target.value)}
                />
              </Field>
            </div>
            <button
              disabled={!core.Controls.ManualField.Enabled}
              onClick={() =>
                void engine.current!.apply({
                  Kind: "set-field",
                  Field: Number(manual),
                })
              }
            >
              Set field
            </button>
          </fieldset>
          <div className="actions">
            <button
              className="primary"
              aria-disabled={locked || stale}
              onClick={() => {
                if (!locked && !stale) engine.current!.run(budget);
              }}
            >
              Run
            </button>
            <button
              aria-disabled={!!data.active && data.active !== "core"}
              aria-describedby="pause-reason"
              onClick={() => void engine.current!.pause()}
            >
              Pause
            </button>
            <button
              aria-disabled={locked || stale}
              onClick={() => {
                if (!locked && !stale) void engine.current!.step(budget);
              }}
            >
              Advance batch
            </button>
            <button onClick={() => void engine.current!.cancel()}>
              Cancel
            </button>
            <button onClick={() => void engine.current!.reset()}>Reset</button>
            <button onClick={() => resultsHeading.current?.focus()}>
              Read results
            </button>
          </div>
          <p id="pause-reason" className="hint">
            Pause applies to the loop. Programming and diagnostics support
            Cancel instead; Reset cancels an active job before clearing loop
            history.
          </p>
          <div className="metric-strip">
            <p>
              State
              <strong className="state" data-testid="core-state">
                {core.State}
              </strong>
            </p>
            <p>
              Electric field
              <strong>
                <QuantityValue value={core.E} />
              </strong>
            </p>
            <p>
              Polarization
              <strong>
                <QuantityValue value={core.P} />
              </strong>
            </p>
            <p>
              Retained samples
              <strong data-testid="retained-count">{core.Count}</strong>
            </p>
            <p>
              Go revision<strong data-testid="revision">{data.revision}</strong>
            </p>
          </div>
          <section aria-labelledby="loop-results">
            <h2 id="loop-results" ref={resultsHeading} tabIndex={-1}>
              Loop results
            </h2>
            <p>
              {core.Evidence} · {core.Dropped} dropped ring samples · Core
              revision {core.Revision}. Rows are bound to their displayed
              revision, not relabelled by later work.
            </p>
            {data.corePage && data.corePage.Revision !== core.Revision && (
              <p>
                Table is stale while computation advances; pause to read the
                current page.
              </p>
            )}
            <Field label="History view">
              <select
                value={core.View}
                disabled={!!data.active}
                onChange={(event) =>
                  void engine.current!.apply({
                    Kind: "set-view",
                    View: event.target.value,
                  })
                }
              >
                {cat.Views.map((value) => (
                  <option key={value}>{value}</option>
                ))}
              </select>
            </Field>
            {data.corePage && (
              <>
                <ResultsTable
                  id="core-table"
                  page={data.corePage}
                  keys={["Index", "TimeValue", "Branch", "E", "P"]}
                  labels={{ TimeValue: cat.CoreTimeLabel }}
                  caption={
                    (data.corePage.Position || "") +
                    " · Core revision " +
                    data.corePage.Revision
                  }
                  captionRef={coreCaption}
                />
                <div className="actions">
                  <button
                    disabled={!!data.active || data.corePage.Offset === 0}
                    onClick={() =>
                      void engine
                        .current!.page(
                          Math.max(0, data.corePage!.Offset - cat.PageSize),
                        )
                        .then(() => coreCaption.current?.focus())
                    }
                  >
                    Previous core page
                  </button>
                  <button
                    disabled={
                      !!data.active ||
                      data.corePage.Offset + cat.PageSize >= data.corePage.Total
                    }
                    onClick={() =>
                      void engine
                        .current!.page(data.corePage!.Offset + cat.PageSize)
                        .then(() => coreCaption.current?.focus())
                    }
                  >
                    Next core page
                  </button>
                  <button
                    disabled={!!data.active || !data.corePage.Total}
                    onClick={() =>
                      void engine
                        .current!.page(
                          Math.floor(
                            (data.corePage!.Total - 1) / cat.PageSize,
                          ) * cat.PageSize,
                        )
                        .then(() => coreCaption.current?.focus())
                    }
                  >
                    Last core page
                  </button>
                </div>
              </>
            )}
            <details>
              <summary>Canonical SI rows and units</summary>
              <p>
                Time is in seconds; E and P retain their exact canonical SI
                beside Go display and speech. These are the same page rows, not
                new calculations.
              </p>
              <GoData value={data.corePage} />
            </details>
            <details>
              <summary>Supplementary loop plot</summary>
              <Plot core={core} />
            </details>
          </section>
          <ProgramPanel
            data={data}
            engine={engine.current!}
            settingsStale={stale}
          />
          <AnalysisPanel
            data={data}
            engine={engine.current!}
            settingsStale={stale}
          />
          <details>
            <summary>Model assumptions, material and build identity</summary>
            <ul>
              {core.Assumptions.map((text) => (
                <li key={text}>{text}</li>
              ))}
            </ul>
            <GoData
              values={core.ScientificValues}
              value={{
                Material: core.Material,
                Config: core.Config,
                Controls: core.Controls,
                Identity: cat.Identity,
                UI: __FECIM_UI__,
              }}
            />
          </details>
        </>
      )}
      <LocalActions data={data} engine={engine.current} />
    </>
  );
}
