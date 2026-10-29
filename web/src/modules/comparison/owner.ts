import type { Schema } from "../../api/types";
import { claimScience, releaseScience } from "../../worker/ownership";
type Transport = {
  snapshot: () => Promise<Schema<"Snapshot">>;
  apply: (c: Schema<"Command">) => Promise<Schema<"Snapshot">>;
  close: () => void;
  projects: () => Promise<{ id: string; name: string }[]>;
};
class ComparisonOwner {
  snapshot?: Schema<"Snapshot">;
  drafts: Record<string, string> = {};
  error?: Error;
  busy = false;
  status = "Loading comparison…";
  projects: { id: string; name: string }[] = [];
  project = "";
  private transport?: Transport;
  private connecting?: Promise<void>;
  private version = 0;
  private listeners = new Set<() => void>();
  subscribe = (fn: () => void) => {
    this.listeners.add(fn);
    return () => {
      this.listeners.delete(fn);
    };
  };
  getSnapshot = () => this.version;
  changed() {
    this.version++;
    this.listeners.forEach((fn) => fn());
  }
  async connect() {
    return (this.connecting ??= (async () => {
      this.transport =
        __FECIM_PROFILE__ === "local"
          ? await import("./local").then((m) => m.localComparison())
          : await import("./hosted").then((m) => m.hostedComparison());
      const snapshot = await this.transport.snapshot(),
        projects = await this.transport.projects();
      this.snapshot = snapshot;
      this.status = snapshot.status;
      this.projects = projects;
      this.project = projects[0]?.id || "";
      this.error = undefined;
      this.changed();
    })().catch((error) => {
      this.transport?.close();
      this.transport = undefined;
      this.snapshot = undefined;
      this.connecting = undefined;
      this.error = error as Error;
      this.status = "Comparison unavailable; retry connection";
      this.changed();
    }));
  }
  draft(id: string, value: string) {
    this.drafts = { ...this.drafts, [id]: value };
    this.changed();
  }
  async action(action: string, format?: string) {
    if (this.busy) throw new Error("One comparison request is already active");
    if (action === "export" && __FECIM_PROFILE__ !== "local") {
      this.status =
        "Exports require the authenticated local application and an opened project. No localhost request was made.";
      this.changed();
      return;
    }
    if (!this.transport || !this.snapshot)
      throw new Error("Retry comparison connection first");
    if (action === "export" && !this.project)
      throw new Error("Open a granted project in Projects before exporting");
    claimScience(this);
    this.busy = true;
    this.error = undefined;
    this.changed();
    const submitted = { ...this.drafts };
    try {
      const fields = Object.fromEntries(
        (this.snapshot.fields || []).map((f) => [
          f.id,
          submitted[f.id] ?? f.input,
        ]),
      );
      const input =
        action === "calculate"
          ? Object.fromEntries(
              Object.entries(fields).filter(([k]) => k !== "output"),
            )
          : action === "sensitivity"
            ? { output: fields.output }
            : action === "export"
              ? { format: format! }
              : {};
      this.snapshot = await this.transport.apply({
        action,
        revision: this.snapshot.revision,
        input,
        resources:
          action === "export" ? [{ id: this.project, kind: "project" }] : [],
      });
      for (const [k, v] of Object.entries(submitted))
        if (
          this.drafts[k] === v &&
          (action === "reset" || (action === "calculate" && k in input))
        )
          delete this.drafts[k];
      this.status = this.snapshot.status;
    } catch (error) {
      this.error = error as Error;
      throw error;
    } finally {
      this.busy = false;
      releaseScience(this);
      this.changed();
    }
  }
  async refreshProjects() {
    if (!this.transport) throw new Error("Connect comparison first");
    this.projects = await this.transport.projects();
    if (!this.projects.some((p) => p.id === this.project))
      this.project = this.projects[0]?.id || "";
    this.changed();
  }
  async retry() {
    this.transport?.close();
    this.transport = undefined;
    this.connecting = undefined;
    await this.connect();
  }
}
export const comparisonOwner = new ComparisonOwner();
