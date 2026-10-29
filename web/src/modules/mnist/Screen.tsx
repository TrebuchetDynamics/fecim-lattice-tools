import { useEffect, useRef, useState } from "react";
import type { Schema } from "../../api/types";
import { Value } from "../../research/Shared";
import { claimScience, releaseScience } from "../../worker/ownership";
import type { Local } from "./local";
import "./mnist.css";

const local = __FECIM_PROFILE__ === "local";
const field = (s: Schema<"Snapshot"> | undefined, id: string) =>
  s?.fields?.find((f) => f.id === id)?.input || "";
function Table({ table }: { table: Schema<"TablePage"> }) {
  return (
    <div className="table-scroll">
      <table id={"mnist-" + table.id}>
        <caption tabIndex={-1}>{table.caption}</caption>
        <thead>
          <tr>
            {table.columns?.map((c, i) => (
              <th key={i} scope="col">
                {c}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {table.rows?.map((row, i) => (
            <tr key={i}>
              {row?.map((c, j) =>
                j === 0 ? (
                  <th key={j} scope="row">
                    <Value value={c.presentation} />
                  </th>
                ) : (
                  <td key={j}>
                    <Value value={c.presentation} />
                  </td>
                ),
              )}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
const configFields: Record<string, string> = {
  levels: "Uniform levels",
  layer1: "Layer 1 levels",
  layer2: "Layer 2 levels",
  adc_bits: "ADC bits",
  dac_bits: "DAC bits",
  noise: "Noise coefficient",
  seed: "Noise seed",
};
export default function Screen() {
  const [adapter, setAdapter] = useState<Local>(),
    [snapshot, setSnapshot] = useState<Schema<"Snapshot">>(),
    [cfg, setCfg] = useState<Record<string, string>>({}),
    [pending, setPending] = useState(false),
    [unknown, setUnknown] = useState(false),
    [error, setError] = useState(""),
    [status, setStatus] = useState("Connecting local capability…"),
    [job, setJob] = useState<Schema<"Job">>();
  const [root, setRoot] = useState(""),
    [model, setModel] = useState("pretrained_weights.json"),
    [images, setImages] = useState("t10k-images-idx3-ubyte.gz"),
    [labels, setLabels] = useState("t10k-labels-idx1-ubyte.gz"),
    [sample, setSample] = useState("0"),
    [start, setStart] = useState("0"),
    [count, setCount] = useState("500"),
    [kind, setKind] = useState("levels"),
    [values, setValues] = useState("[2,4,8,16,30]"),
    [layer, setLayer] = useState("1"),
    [caseIndex, setCaseIndex] = useState("0"),
    [pixels, setPixels] = useState(() => Array<number>(784).fill(0)),
    [raw, setRaw] = useState(""),
    [row, setRow] = useState("0"),
    [col, setCol] = useState("0"),
    [intensity, setIntensity] = useState("1"),
    [imported, setImported] = useState("");
  const owner = useRef({}),
    lock = useRef(false),
    alive = useRef(true),
    timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined),
    alert = useRef<HTMLParagraphElement>(null),
    scope = useRef<HTMLElement>(null),
    canvas = useRef<HTMLCanvasElement>(null),
    importSequence = useRef(0),
    lastAnnouncement = useRef(0);
  const [announcement, setAnnouncement] = useState("");
  const active =
      !!job && ["running", "queued", "cancelling"].includes(job.state),
    locked = pending || active || unknown || !adapter;
  const configFrom = (s: Schema<"Snapshot">) => {
    const c = JSON.parse(field(s, "config")) as Record<string, unknown>;
    setCfg(
      Object.fromEntries(Object.entries(c).map(([k, v]) => [k, String(v)])),
    );
  };
  const publish = (s: Schema<"Snapshot">) => {
    if (!alive.current) return;
    setSnapshot(s);
    setStatus(s.status);
  };
  const publishJob = (j: Schema<"Job">) => {
    if (!alive.current) return;
    setJob(j);
    setStatus(
      j.state +
        " — " +
        (j.message || "") +
        ` (${j.progress.completed}/${j.progress.total} completed)`,
    );
    const active = ["running", "queued", "cancelling"].includes(j.state);
    if (!active) releaseScience(owner.current);
    if (!active || Date.now() - lastAnnouncement.current > 5000) {
      setAnnouncement(j.state + " — " + (j.message || ""));
      lastAnnouncement.current = Date.now();
    }
  };
  function watch(a: Local, id: string) {
    clearTimeout(timer.current);
    timer.current = setTimeout(() => {
      void a
        .job(id)
        .then(async (j) => {
          if (!alive.current) return;
          publishJob(j);
          if (["running", "queued", "cancelling"].includes(j.state))
            watch(a, id);
          else publish(await a.refresh());
        })
        .catch((e) => {
          if (alive.current) {
            setUnknown(true);
            setError(
              "Execution outcome unknown; previous results retained. Reconnect or cancel by job ID. " +
                (e as Error).message,
            );
          }
        });
    }, 100);
  }
  useEffect(() => {
    alive.current = true;
    if (local) {
      void import("./local")
        .then((m) => m.connect())
        .then(async (a) => {
          if (!alive.current) return;
          setAdapter(a);
          publish(a.snapshot);
          configFrom(a.snapshot);
          setRoot(a.roots[0]?.id || "");
          const id = field(a.snapshot, "job_id");
          if (id) {
            setUnknown(true);
            const j = await a.job(id);
            if (!alive.current) return;
            publishJob(j);
            setUnknown(false);
            if (["running", "queued", "cancelling"].includes(j.state)) {
              claimScience(owner.current);
              watch(a, id);
            }
          }
        })
        .catch((e) => {
          if (alive.current) {
            setUnknown(true);
            setError((e as Error).message);
          }
        });
    }
    return () => {
      alive.current = false;
      clearTimeout(timer.current);
      releaseScience(owner.current);
    };
  }, []);
  useEffect(() => {
    if (error && alert.current?.getClientRects().length) alert.current.focus();
  }, [error]);
  useEffect(() => {
    const context = canvas.current?.getContext("2d");
    if (!context) return;
    context.fillStyle = "#101a24";
    context.fillRect(0, 0, 280, 280);
    pixels.forEach((v, i) => {
      context.fillStyle = `rgb(${v * 255} ${v * 255} ${v * 255})`;
      context.fillRect((i % 28) * 10, Math.floor(i / 28) * 10, 10, 10);
    });
  }, [pixels]);
  async function action(
    action: string,
    input: Record<string, string> = {},
    resources?: Schema<"ResourceReference">[],
  ) {
    if (!adapter || !snapshot || lock.current) return;
    lock.current = true;
    setPending(true);
    setError("");
    try {
      const next = await adapter.apply({
        action,
        input,
        resources,
        revision: snapshot.revision,
      });
      publish(next);
      if (action === "configure") configFrom(next);
      if (action === "view")
        requestAnimationFrame(() => {
          const caption = scope.current?.querySelector<HTMLElement>(
            "#mnist-weights caption",
          );
          if (caption?.getClientRects().length) caption.focus();
        });
    } catch (e) {
      setError((e as Error).message);
    } finally {
      lock.current = false;
      setPending(false);
    }
  }
  async function load(dataset: boolean) {
    if (!adapter || !snapshot || lock.current) return;
    lock.current = true;
    setPending(true);
    setError("");
    try {
      const refs = dataset
        ? [
            await adapter.bind(root, images, "dataset"),
            await adapter.bind(root, labels, "dataset"),
          ]
        : [await adapter.bind(root, model, "model")];
      publish(
        await adapter.apply({
          action: dataset ? "load_dataset" : "load_model",
          resources: refs,
          revision: snapshot.revision,
        }),
      );
      if (!dataset) setLayer("1");
    } catch (e) {
      setError((e as Error).message);
    } finally {
      lock.current = false;
      setPending(false);
    }
  }
  async function run(action: string) {
    if (!adapter || !snapshot || lock.current) return;
    lock.current = true;
    setPending(true);
    setError("");
    let attempted = false;
    try {
      claimScience(owner.current);
      attempted = true;
      const j = await adapter.start({
        action,
        revision: snapshot.revision,
        input: action === "infer" ? {} : { start, count, kind, values },
      });
      publishJob(j);
      setUnknown(false);
      watch(adapter, j.id);
    } catch (e) {
      setUnknown(attempted);
      setError(
        (attempted
          ? "Start rejected or execution outcome unknown; reconnect before retrying. "
          : "") + (e as Error).message,
      );
    } finally {
      // A lost start receipt does not prove the server rejected execution.
      if (!attempted) releaseScience(owner.current);
      lock.current = false;
      setPending(false);
    }
  }
  async function reconnect(cancel = false) {
    if (!adapter || lock.current) return;
    lock.current = true;
    setPending(true);
    setError("");
    clearTimeout(timer.current);
    try {
      const current = await adapter.refresh();
      publish(current);
      const id = field(current, "job_id");
      if (id) {
        const j = cancel ? await adapter.cancel(id) : await adapter.job(id);
        publishJob(j);
        if (["running", "queued", "cancelling"].includes(j.state)) {
          claimScience(owner.current);
          watch(adapter, id);
        } else publish(await adapter.refresh());
      } else {
        setJob(undefined);
        releaseScience(owner.current);
      }
      setUnknown(false);
    } catch (e) {
      setUnknown(true);
      setError(
        "Execution outcome unknown; no success claimed. " +
          (e as Error).message,
      );
    } finally {
      lock.current = false;
      setPending(false);
    }
  }
  function setPixel(y: number, x: number, value: number) {
    if (
      !Number.isInteger(y) ||
      !Number.isInteger(x) ||
      y < 0 ||
      y > 27 ||
      x < 0 ||
      x > 27 ||
      !Number.isFinite(value) ||
      value < 0 ||
      value > 1
    ) {
      setError("Choose row/column 0..27 and intensity 0..1.");
      return;
    }
    setPixels((old) => {
      const next = [...old];
      next[y * 28 + x] = value;
      return next;
    });
    setRaw("");
  }
  async function readImport(file: File | undefined) {
    const seq = ++importSequence.current;
    setImported("");
    setError("");
    if (!file) return;
    if (file.size > 2 * 1024 * 1024) {
      setError("Read-only result import exceeds 2 MiB.");
      return;
    }
    try {
      const text = await file.text();
      if (seq === importSequence.current) setImported(text);
    } catch {
      if (seq === importSequence.current)
        setError("Result import could not be read.");
    }
  }
  function exportResult() {
    const blob = new Blob([field(snapshot, "result")], {
        type: "application/json",
      }),
      url = URL.createObjectURL(blob),
      a = document.createElement("a");
    a.href = url;
    a.download = "mnist-result.json";
    a.click();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
  const weights = snapshot?.tables?.find((t) => t.id === "weights");
  const can = (id: string) =>
    !!snapshot?.actions?.find((a) => a.id === id)?.available;
  return (
    <section ref={scope} className="section mnist-workbench">
      <div className="mnist-kicker">MODULE 03 / PAIRED CPU INFERENCE</div>
      <h1>MNIST</h1>
      <p className="mnist-notice">
        Educational simulation, not a hardware claim. Go owns preprocessing,
        inference, quantization and evidence. No model or dataset is downloaded
        automatically.
      </p>
      <h2>Model, input, evidence</h2>
      <p>
        Compare full precision with CIM probabilities and hidden activations. A
        drawing has no accuracy. Labelled evaluation records exact file
        fingerprints, actual subset, seed, configuration and confusion counts.
      </p>
      {error && (
        <p ref={alert} role="alert" tabIndex={-1}>
          {error}
        </p>
      )}
      <p className="sr-only" role="status" aria-live="polite">
        {announcement}
      </p>
      {!local ? (
        <>
          <p>
            Inference, dataset evaluation, weights and sweeps require the local
            application: build the local browser profile, then run{" "}
            <code>fecim-lattice-tools serve --root DIRECTORY</code> and open its
            terminal launch link. This hosted page does not contact the server
            or upload model, input or result bytes. Ordinary static asset
            requests still occur.
          </p>
          <label>
            Import result for read-only inspection
            <input
              type="file"
              onChange={(e) => void readImport(e.target.files?.[0])}
            />
          </label>
          <p>Imported/unverified — not a locally verified evaluation</p>
          <pre data-testid="mnist-import">
            {imported ||
              "No imported result. Reload discards it; no hosted inference or export."}
          </pre>
        </>
      ) : (
        <>
          <p data-testid="mnist-status">{status}</p>
          {adapter && !adapter.reloadRecovery && (
            <p>
              Local session identity storage is unavailable. Do not reload an
              active job; keep this page open to cancel or inspect it.
            </p>
          )}
          <button
            disabled={pending || !adapter}
            onClick={() => void reconnect()}
          >
            Reconnect authoritative state
          </button>
          <button
            disabled={
              pending || !adapter || !(job?.id || field(snapshot, "job_id"))
            }
            onClick={() => void reconnect(true)}
          >
            Cancel MNIST job
          </button>
          <div className="mnist-grid">
            <div>
              <fieldset disabled={locked}>
                <legend>01 / Explicit rooted files</legend>
                <label>
                  Granted root
                  <select
                    value={root}
                    onChange={(e) => setRoot(e.target.value)}
                  >
                    {adapter?.roots.map((r) => (
                      <option key={r.id} value={r.id}>
                        {r.name}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Model path
                  <input
                    value={model}
                    onChange={(e) => setModel(e.target.value)}
                  />
                </label>
                <button onClick={() => void load(false)}>
                  Load selected model
                </button>
                <p>
                  Select exact baseline/QAT bytes; no nearest-model fallback.
                </p>
                <label>
                  IDX images path
                  <input
                    value={images}
                    onChange={(e) => setImages(e.target.value)}
                  />
                </label>
                <label>
                  IDX labels path
                  <input
                    value={labels}
                    onChange={(e) => setLabels(e.target.value)}
                  />
                </label>
                <button onClick={() => void load(true)}>
                  Load labelled corpus
                </button>
              </fieldset>
              <fieldset disabled={locked}>
                <legend>02 / Configuration</legend>
                {Object.entries(configFields).map(([key, label]) => (
                  <label key={key}>
                    {label}
                    <input
                      value={cfg[key] || ""}
                      onChange={(e) =>
                        setCfg({ ...cfg, [key]: e.target.value })
                      }
                    />
                  </label>
                ))}
                <label>
                  Quantization method
                  <select
                    value={cfg.quantization || "uniform-ptq"}
                    onChange={(e) =>
                      setCfg({ ...cfg, quantization: e.target.value })
                    }
                  >
                    <option value="uniform-ptq">
                      Uniform PTQ / requantize
                    </option>
                    <option value="per-layer-ptq">Per-layer PTQ</option>
                    <option value="model-qat">Declared model QAT levels</option>
                  </select>
                </label>
                <label>
                  Architecture
                  <select
                    value={cfg.architecture || "784-H-10"}
                    onChange={(e) =>
                      setCfg({ ...cfg, architecture: e.target.value })
                    }
                  >
                    <option value="784-H-10">784 → H → 10</option>
                    <option value="784-10">
                      784 → 10 (requires calibration weights)
                    </option>
                  </select>
                </label>
                {(["ir_drop", "sneak"] as const).map((key) => (
                  <label key={key}>
                    <input
                      type="checkbox"
                      checked={cfg[key] === "true"}
                      onChange={(e) =>
                        setCfg({ ...cfg, [key]: String(e.target.checked) })
                      }
                    />
                    {key === "ir_drop" ? "IR drop" : "Sneak paths"}
                  </label>
                ))}
                <button onClick={() => void action("configure", cfg)}>
                  Apply configuration
                </button>
                {["ideal", "hardware", "noisy"].map((preset) => (
                  <button
                    key={preset}
                    onClick={() => void action("configure", { preset })}
                  >
                    {preset[0].toUpperCase() + preset.slice(1)} preset
                  </button>
                ))}
              </fieldset>
            </div>
            <div>
              <fieldset disabled={locked}>
                <legend>03 / Input — choose one source</legend>
                <canvas
                  ref={canvas}
                  width={280}
                  height={280}
                  aria-hidden="true"
                  onPointerDown={(e) => {
                    e.currentTarget.setPointerCapture(e.pointerId);
                    const r = e.currentTarget.getBoundingClientRect();
                    setPixel(
                      Math.floor(((e.clientY - r.top) * 28) / r.height),
                      Math.floor(((e.clientX - r.left) * 28) / r.width),
                      Number(intensity),
                    );
                  }}
                  onPointerMove={(e) => {
                    if (e.buttons) {
                      const r = e.currentTarget.getBoundingClientRect();
                      setPixel(
                        Math.floor(((e.clientY - r.top) * 28) / r.height),
                        Math.floor(((e.clientX - r.left) * 28) / r.width),
                        Number(intensity),
                      );
                    }
                  }}
                />
                <p>
                  Pointer drawing is supplementary. Use row, column and
                  intensity to create the same input with a keyboard, or provide
                  all 784 pixels as JSON.
                </p>
                <label>
                  Pixel row
                  <input value={row} onChange={(e) => setRow(e.target.value)} />
                </label>
                <label>
                  Pixel column
                  <input value={col} onChange={(e) => setCol(e.target.value)} />
                </label>
                <label>
                  Pixel intensity
                  <input
                    value={intensity}
                    onChange={(e) => setIntensity(e.target.value)}
                  />
                </label>
                <button
                  onClick={() =>
                    setPixel(Number(row), Number(col), Number(intensity))
                  }
                >
                  Set pixel
                </button>
                <button
                  onClick={() => {
                    setPixels(Array<number>(784).fill(0));
                    setRaw("");
                    void action("clear");
                  }}
                >
                  Clear input
                </button>
                <label>
                  Raw input pixels (784 JSON values in [0,1])
                  <textarea
                    value={raw || JSON.stringify(pixels)}
                    onChange={(e) => setRaw(e.target.value)}
                  />
                </label>
                <button
                  onClick={() =>
                    void action("input", {
                      pixels: raw || JSON.stringify(pixels),
                      source: "drawing",
                    })
                  }
                >
                  Use drawing
                </button>
                <button
                  onClick={() =>
                    void action("input", {
                      pixels: raw || JSON.stringify(pixels),
                      source: "input",
                    })
                  }
                >
                  Use unlabelled input
                </button>
                <label>
                  Dataset sample index
                  <input
                    value={sample}
                    onChange={(e) => setSample(e.target.value)}
                  />
                </label>
                <button
                  disabled={!can("sample")}
                  onClick={() => void action("sample", { index: sample })}
                >
                  Use dataset sample
                </button>
                <p>
                  Drawing normalization runs once in Go. Dataset and
                  programmatic pixels bypass drawing normalization.
                </p>
                <button
                  disabled={!can("infer")}
                  onClick={() => void run("infer")}
                >
                  Run paired inference
                </button>
              </fieldset>
              <fieldset disabled={locked}>
                <legend>04 / Labelled evaluation and sweeps</legend>
                <label>
                  Subset start
                  <input
                    value={start}
                    onChange={(e) => setStart(e.target.value)}
                  />
                </label>
                <label>
                  Selected images
                  <input
                    value={count}
                    onChange={(e) => setCount(e.target.value)}
                  />
                </label>
                <button
                  disabled={!can("evaluate")}
                  onClick={() => void run("evaluate")}
                >
                  Evaluate labelled subset
                </button>
                <label>
                  Sweep parameter
                  <select
                    value={kind}
                    onChange={(e) => setKind(e.target.value)}
                  >
                    <option value="levels">Uniform levels</option>
                    <option value="adc">ADC bits</option>
                    <option value="noise">Noise coefficient</option>
                  </select>
                </label>
                <label>
                  Sweep cases (JSON)
                  <input
                    value={values}
                    onChange={(e) => setValues(e.target.value)}
                  />
                </label>
                <button
                  disabled={!can("sweep")}
                  onClick={() => void run("sweep")}
                >
                  Run selected sweep
                </button>
                <p>
                  Evaluation: at most 10000 images. Sweeps: explicitly choose at
                  most 500 images and 64 cases. Jobs have a 15-minute
                  cancellation deadline. No silent sample reduction.
                </p>
              </fieldset>
            </div>
          </div>
          <h2>Identity and retained results</h2>
          <dl data-testid="mnist-identity">
            {snapshot?.fields
              ?.filter((f) =>
                [
                  "hidden",
                  "model_sha256",
                  "dataset_sha256",
                  "corpus_count",
                  "source",
                  "result_fresh",
                ].includes(f.id),
              )
              .map((f) => (
                <div key={f.id}>
                  <dt>
                    {f.label}: {f.input}
                  </dt>
                </div>
              ))}
          </dl>
          <ul data-testid="mnist-evidence">
            {snapshot?.evidence?.map((e, i) => (
              <li key={i}>{e}</li>
            ))}
          </ul>
          <fieldset disabled={locked || !weights}>
            <legend>Weight inspection</legend>
            <label>
              Weight layer
              <select
                value={layer}
                onChange={(e) => {
                  setLayer(e.target.value);
                  void action("view", { layer: e.target.value });
                }}
              >
                <option value="1">Hidden weights</option>
                <option value="2">Output weights</option>
                <option value="3">Calibration weights (when present)</option>
              </select>
            </label>
            <button
              disabled={!weights || weights.page <= 1}
              onClick={() =>
                void action("view", { page: String((weights?.page || 1) - 1) })
              }
            >
              Previous weight page
            </button>
            <button
              disabled={!weights || weights.page >= weights.total}
              onClick={() =>
                void action("view", { page: String((weights?.page || 1) + 1) })
              }
            >
              Next weight page
            </button>
            <label>
              Confusion case
              <input
                value={caseIndex}
                onChange={(e) => setCaseIndex(e.target.value)}
              />
            </label>
            <button onClick={() => void action("view", { case: caseIndex })}>
              Inspect confusion case
            </button>
          </fieldset>
          {snapshot?.tables?.map((t) => (
            <Table key={t.id} table={t} />
          ))}
          <button
            disabled={!snapshot || field(snapshot, "result") === "null"}
            onClick={exportResult}
          >
            Export complete result JSON
          </button>
          <details>
            <summary>Complete Go result JSON and provenance</summary>
            <pre data-testid="mnist-result">
              {field(snapshot, "result") || "No result"}
            </pre>
          </details>
        </>
      )}
    </section>
  );
}
