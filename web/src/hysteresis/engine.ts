import "../worker/hysteresis-worker.mjs";
import { claimScience, releaseScience } from "../worker/ownership";
import type {
  AnalysisResult,
  AnalysisRow,
  Budget,
  Core,
  CoreRow,
  Manifest,
  Page,
  Program,
  Progress,
  Reply,
  Results,
  SessionInfo,
} from "./types";

// Transport/lifecycle only. Scientific state and revisions are accepted only
// from the identity-checking production client and Go's response envelopes.
class GoInputError extends Error {
  constructor(
    message: string,
    readonly committed = false,
  ) {
    super(message);
  }
}
export class Engine {
  data: Results = { revision: 0, status: "Loading Go Worker…", failed: false };
  private client?: InstanceType<typeof HysteresisWorkerClient>;
  private tail: Promise<unknown> = Promise.resolve();
  private generation = 0;
  private loop = 0;
  private alive = true;
  private operations = new Set<string>();
  private lastAnnouncement = 0;
  private scientificRequests = 0;
  constructor(private changed: (data: Results) => void) {}
  private publish(patch: Partial<Results>) {
    if (this.alive) {
      if (patch.active) {
        try { claimScience(this); } catch (error) {
          this.data = { ...this.data, error: String(error instanceof Error ? error.message : error) };
          this.changed(this.data); return;
        }
      }
      this.data = { ...this.data, ...patch };
      if (!this.data.active && this.scientificRequests === 0) releaseScience(this);
      this.changed(this.data);
    }
  }
  private enqueue<T>(fn: () => Promise<T>): Promise<T | undefined> {
    const generation = this.generation;
    const owner = this.loop;
    const promise = this.tail.then(async () => {
      if (!this.alive || generation !== this.generation) return;
      try {
        return await fn();
      } catch (error) {
        if (generation !== this.generation || !this.alive) return;
        // A rejected earlier command must not cancel a newer queued job.
        // Only the current owner may retire the pump and release its UI lock.
        const ownsLoop = owner === this.loop;
        if (ownsLoop) this.loop++;
        this.publish({
          error: String(error instanceof Error ? error.message : error),
          ...(ownsLoop ? { active: undefined } : {}),
          failed: !(error instanceof GoInputError),
          status:
            "Failed; previous results retained, not a successful computation",
        });
      }
    });
    this.tail = promise;
    return promise;
  }
  async start() {
    const generation = ++this.generation;
    this.loop++;
    this.client?.terminate();
    this.tail = Promise.resolve();
    this.operations.clear();
    this.publish({
      core: undefined,
      corePage: undefined,
      program: undefined,
      progress: undefined,
      analysis: undefined,
      analysisPage: undefined,
      active: undefined,
      revision: 0,
      error: undefined,
      failed: false,
      status: "Loading Go Worker…",
    });
    await this.enqueue(async () => {
      const response = await fetch("/engine/manifest.json", {
        cache: "no-store",
        credentials: "omit",
        redirect: "error",
      });
      if (!response.ok) throw new Error("Manifest HTTP " + response.status);
      const manifest: Manifest = await response.json();
      if (generation !== this.generation) return;
      if (manifest.UI !== __FECIM_UI__)
        throw new Error("UI/Worker asset fingerprint mismatch");
      if (manifest.API !== undefined && manifest.API !== __FECIM_API__)
        throw new Error("API-contract fingerprint mismatch");
      this.publish({ catalogue: manifest.Catalogue });
      const client = (this.client = new HysteresisWorkerClient({
        identity: manifest.identity,
        session: crypto.randomUUID(),
        workerURL: "/engine/hysteresis-worker.mjs",
        workerSHA256: manifest.assets["hysteresis-worker.mjs"],
        wasmURL: "/engine/hysteresis.wasm",
        wasmSHA256: manifest.assets["hysteresis.wasm"],
        runtimeURL: "/engine/wasm_exec.js",
        runtimeSHA256: manifest.assets["wasm_exec.js"],
      }));
      await client.ready;
      if (generation !== this.generation) return;
      const initial = await this.send<SessionInfo>("session.new", {});
      this.operations = new Set(initial.Operations.map((op) => op.name));
      this.publish({
        core: initial.Core,
        status: "Ready; paused. Computations and results are ephemeral.",
      });
      await this.corePage(0);
    });
  }
  // Collect Go-owned SI rows only after an explicit local operation. No
  // numerical conversion/formatting or automatic persistence happens here.
  async fileSnapshot() {
    if (this.data.active)
      throw new Error("Pause or cancel computation before local files");
    await this.tail;
    if (!this.data.core || this.data.failed)
      throw new Error("Ready Go state required");
    const core = this.data.core;
    const rows: { Index: number; Time: number; E: number; P: number }[] = [];
    for (
      let offset = 0;
      offset < core.ViewCount;
      offset += this.data.catalogue!.PageSize
    ) {
      const page = await this.send<Page<CoreRow>>("core.page", {
        Offset: offset,
      });
      if (page.Revision !== core.Revision)
        throw new Error("State changed during export; pause and retry");
      for (const row of page.Rows)
        rows.push({
          Index: row.Index,
          Time: row.Time,
          E: row.E.SI,
          P: row.P.SI,
        });
    }
    return {
      Config: core.Config,
      Rows: rows,
      Calibration: this.data.program?.Calibration,
      Analysis: this.data.analysis,
    };
  }
  dispose() {
    this.alive = false;
    this.generation++;
    this.loop++;
    this.client?.terminate();
    releaseScience(this);
  }
  private async send<T>(operation: string, input: unknown = null): Promise<T> {
    if (operation !== "session.new" && !this.operations.has(operation))
      throw new Error("Go capability unavailable: " + operation);
    const computes = !/page$|snapshot$|result$|session.new/.test(operation);
    if (computes) { try { claimScience(this); } catch (error) { throw new GoInputError((error as Error).message); } }
    if (computes) this.scientificRequests++;
    let reply: Reply<T>;
    try { reply = JSON.parse(await this.client!.request(operation, input)); }
    finally { if (computes) { this.scientificRequests--; if (!this.data.active) releaseScience(this); } }
    this.publish({ revision: reply.revision });
    if (reply.error) {
      // Go can commit partial work and a terminal failure in the same envelope.
      // Preserve that returned state; an error is never scientific success.
      if (reply.committed && reply.result) {
        if (operation.startsWith("core."))
          this.publish({ core: reply.result as unknown as Core });
        else if (operation.startsWith("program.")) {
          const program = reply.result as unknown as Program;
          this.publish({ program, core: program.Core });
        } else if (operation.startsWith("analysis."))
          this.publish({ progress: reply.result as unknown as Progress });
      }
      throw new GoInputError(reply.error, reply.committed);
    }
    if (!reply.committed || reply.result === undefined)
      throw new Error("Incomplete Go reply");
    return reply.result;
  }
  private async corePage(offset: number) {
    const corePage = await this.send<Page<CoreRow>>("core.page", {
      Offset: offset,
    });
    this.publish({ corePage });
  }
  private async refresh() {
    const core = await this.send<Core>("core.snapshot");
    this.publish({ core });
    if (this.data.program)
      this.publish({ program: await this.send<Program>("program.snapshot") });
    await this.corePage(0);
  }
  private stop() {
    this.loop++;
    this.publish({ active: undefined });
  }
  apply(command: unknown) {
    this.stop();
    return this.enqueue(async () => {
      const core = await this.send<Core>("core.apply", command);
      this.publish({
        core,
        error: undefined,
        status: "Settings/state applied by Go",
      });
      if (this.data.program)
        this.publish({ program: await this.send<Program>("program.snapshot") });
      await this.corePage(0);
    });
  }
  page(offset: number) {
    return this.enqueue(() => this.corePage(offset));
  }
  async step(budget: Budget) {
    this.stop();
    return this.enqueue(async () => {
      await this.send<Core>("core.advance", budget); // Go validates while paused, before Run
      this.publish({
        core: await this.send<Core>("core.apply", { Kind: "run" }),
        error: undefined,
      });
      this.publish({ core: await this.send<Core>("core.advance", budget) });
      this.publish({
        core: await this.send<Core>("core.apply", { Kind: "pause" }),
        status: "Batch completed; paused",
      });
      await this.corePage(0);
    });
  }
  run(budget: Budget) {
    if (this.data.active) return;
    const token = ++this.loop;
    this.publish({ active: "core", error: undefined, status: "Running loop" });
    void this.enqueue(async () => {
      if (token !== this.loop) return;
      this.publish({ error: undefined, status: "Running loop" });
      await this.send<Core>("core.advance", budget); // paused admission; no scientific steps
      this.publish({
        core: await this.send<Core>("core.apply", { Kind: "run" }),
      });
    }).then(() => this.pump("core", budget, token));
  }
  pause() {
    if (this.data.active && this.data.active !== "core") {
      this.publish({
        status:
          "Pause is unavailable for programming/diagnostics; Cancel stops the active job.",
      });
      return Promise.resolve();
    }
    this.stop();
    return this.enqueue(async () => {
      this.publish({
        core: await this.send<Core>("core.apply", { Kind: "pause" }),
        status: "Paused",
      });
      await this.corePage(0);
    });
  }
  cancel() {
    this.stop();
    return this.enqueue(async () => {
      await this.send("job.cancel");
      await this.refresh();
      if (this.data.progress?.State === "running")
        this.publish({
          progress: { ...this.data.progress, State: "cancelled" },
          analysis: undefined,
          analysisPage: undefined,
        });
      this.publish({
        status: "Cancelled; completed work retained in memory",
        error: undefined,
      });
    });
  }
  reset() {
    this.stop();
    return this.enqueue(async () => {
      if (this.data.progress?.State === "running") {
        const progress = await this.send<Progress>("job.cancel");
        this.publish({
          progress,
          analysis: undefined,
          analysisPage: undefined,
        });
      }
      if (
        this.data.program &&
        this.data.program.Status !== "legacy_protocol_unavailable"
      ) {
        const program = await this.send<Program>("program.apply", {
          Kind: "reset",
        });
        this.publish({ program, core: program.Core });
      } else
        this.publish({
          core: await this.send<Core>("core.apply", { Kind: "reset" }),
        });
      this.publish({ status: "Reset; loop history cleared", error: undefined });
      await this.corePage(0);
    });
  }
  programNew(config: unknown) {
    this.stop();
    return this.enqueue(async () => {
      const program = await this.send<Program>("program.new", config);
      this.publish({
        program,
        core: program.Core,
        error: undefined,
        status: program.Reason || "Programmer initialized",
      });
    });
  }
  programAction(kind: string, target: number, budget: Budget) {
    if (
      this.data.active === "program" &&
      this.data.program?.Config.Mode === "write-read" &&
      kind === "write-target"
    ) {
      void this.enqueue(async () => {
        try {
          const program = await this.send<Program>("program.apply", {
            Kind: kind,
            Target: target,
          });
          this.publish({ program, core: program.Core, error: undefined });
        } catch (error) {
          if (error instanceof GoInputError && !error.committed) {
            this.publish({
              error: error.message,
              status: "Target rejected; active write/read continues",
            });
            return;
          }
          throw error;
        }
      });
      return;
    }
    if (this.data.active) return;
    const token = ++this.loop;
    this.publish({
      active: "program",
      error: undefined,
      status:
        kind === "calibrate" ? "Calibrating in memory" : "Programming target",
    });
    void this.enqueue(async () => {
      if (token !== this.loop) return;
      this.publish({ error: undefined });
      await this.send<Program>("program.advance", budget); // validate idle schedule before taking drive
      const program = await this.send<Program>(
        "program.apply",
        kind === "write-target"
          ? { Kind: kind, Target: target }
          : { Kind: kind },
      );
      this.publish({ program, core: program.Core });
    }).then(() => this.pump("program", budget, token));
  }
  analyze(input: unknown) {
    if (this.data.active) return;
    const token = ++this.loop;
    this.publish({
      active: "analysis",
      error: undefined,
      analysis: undefined,
      analysisPage: undefined,
      status: "Analyzing independent diagnostic",
    });
    void this.enqueue(async () => {
      if (token !== this.loop) return;
      this.publish({ error: undefined });
      this.publish({
        progress: await this.send<Progress>("analysis.new", input),
      });
    }).then(() => this.pump("analysis", undefined, token));
  }
  private async pump(
    kind: "core" | "program" | "analysis",
    budget: Budget | undefined,
    token: number,
  ) {
    while (this.alive && token === this.loop && this.data.active === kind) {
      await this.enqueue(async () => {
        if (token !== this.loop) return;
        if (kind === "core")
          this.publish({ core: await this.send<Core>("core.advance", budget) });
        else if (kind === "program") {
          const program = await this.send<Program>("program.advance", budget);
          this.publish({ program, core: program.Core });
        } else
          this.publish({
            progress: await this.send<Progress>(
              "analysis.advance",
              this.data.catalogue!.AnalysisBudget,
            ),
          });
        if (performance.now() - this.lastAnnouncement >= 5000) {
          this.lastAnnouncement = performance.now();
          this.publish({
            status:
              kind === "analysis"
                ? "Analyzing: " +
                  this.data.progress!.Completed +
                  " of " +
                  this.data.progress!.Total
                : kind === "program"
                  ? "Programming: " + this.data.program!.Phase
                  : "Running loop; " +
                    this.data.core!.Count +
                    " retained samples",
          });
        }
      });
      if (token !== this.loop) return;
      const done =
        kind === "program"
          ? !this.data.program?.OwnsDrive
          : kind === "analysis"
            ? this.data.progress?.State !== "running"
            : false;
      if (done) {
        this.publish({
          active: undefined,
          status:
            kind === "analysis"
              ? "Analysis complete"
              : "Programming " + this.data.program!.Status,
        });
        await this.enqueue(async () => {
          if (kind === "analysis") {
            const analysis = await this.send<AnalysisResult>("analysis.result");
            const table = this.data.catalogue!.Tables[analysis.Kind][0];
            this.publish({ analysis, table });
            await this.analysisPage(table, 0);
          }
          await this.refresh();
        });
        return;
      }
      // Yield to keyboard/cancel between every Go-bounded operation.
      // Playback changes cadence, never physical DT. Short waits keep cancel responsive.
      const delay =
        kind === "core"
          ? this.data.catalogue!.PlaybackCadenceMS /
            this.data.core!.Config.TimeScale
          : this.data.catalogue!.PlaybackCadenceMS;
      const until = performance.now() + delay;
      while (this.alive && token === this.loop && performance.now() < until)
        await new Promise((resolve) =>
          setTimeout(resolve, Math.min(100, until - performance.now())),
        );
    }
  }
  private async analysisPage(table: string, offset: number) {
    this.publish({
      analysisPage: await this.send<Page<AnalysisRow>>("analysis.page", {
        Table: table,
        Offset: offset,
      }),
      table,
    });
  }
  pageAnalysis(table: string, offset: number) {
    return this.enqueue(() => this.analysisPage(table, offset));
  }
}
