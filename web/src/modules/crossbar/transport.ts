import "../../worker/hysteresis-worker.mjs";
import { manifest } from "../../api/hosted";
import type { Schema } from "../../api/types";
export type Snapshot = Schema<"Snapshot">;
export type Job = Schema<"Job">;
export type Command = Schema<"Command">;
export interface Update {
  snapshot: Snapshot;
  job: Job;
}
export interface Transport {
  read(): Promise<Update>;
  action(command: Command): Promise<Update>;
  start(command: Command): Promise<Update>;
  cancel(): Promise<Update>;
  close(): Promise<void>;
  roots?: { id: string; name: string }[];
  openProject?: (root: string, path: string) => Promise<string>;
}
export async function connect(): Promise<Transport> {
  if (__FECIM_PROFILE__ === "local") return (await import("./local")).connect();
  const m = await manifest(),
    entry = m.modules.crossbar;
  if (!entry) throw new Error("Crossbar Worker is unavailable in this build");
  const client = new HysteresisWorkerClient({
    identity: entry.identity,
    prefix: entry.prefix,
    session: crypto.randomUUID(),
    workerURL: "/engine/hysteresis-worker.mjs",
    workerSHA256: m.assets["hysteresis-worker.mjs"],
    wasmURL: "/engine/" + entry.wasm,
    wasmSHA256: m.assets[entry.wasm],
    runtimeURL: "/engine/wasm_exec.js",
    runtimeSHA256: m.assets["wasm_exec.js"],
  });
  try {
    await client.ready;
  } catch (error) {
    client.terminate();
    throw error;
  }
  const call = async (
    operation: string,
    command?: Command,
  ): Promise<Update> => {
    const reply = JSON.parse(
      await client.request(operation, {
        command: command || { action: "", revision: 0 },
      }),
    ) as { error?: string; result: Update };
    if (reply.error) throw new Error(reply.error);
    return reply.result;
  };
  return {
    read: () => call("state"),
    action: (c) => call("action", c),
    start: (c) => call("start", c),
    cancel: () => call("cancel"),
    close: async () => client.terminate(),
  };
}
