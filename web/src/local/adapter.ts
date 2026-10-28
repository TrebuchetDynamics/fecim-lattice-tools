import type { Identity, Config } from "../hysteresis/types";
export interface BrowserSnapshot {
  Config: Config;
  Rows?: { Index: number; Time: number; E: number; P: number }[];
  Calibration?: unknown;
  Analysis?: unknown;
}
interface Roots {
  roots: { id: string; name: string }[];
  identity: Identity;
  UI: string;
  kinds: string[];
}
export interface FileResult {
  status: string;
  provenance: string;
  value?: { Config?: Config; [key: string]: unknown };
}
async function json<T>(url: string, body?: unknown): Promise<T> {
  const response = await fetch(url, {
    method: body === undefined ? "GET" : "POST",
    credentials: "same-origin",
    cache: "no-store",
    redirect: "error",
    headers:
      body === undefined
        ? {}
        : { "Content-Type": "application/json", "X-FeCIM-Request": "1" },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  if (!response.ok) throw new Error(await response.text());
  return response.json();
}
export async function connectLocal(identity: Identity) {
  const info = await json<Roots>("/api/v1/roots");
  if (
    info.UI !== __FECIM_UI__ ||
    JSON.stringify(info.identity) !== JSON.stringify(identity)
  )
    throw new Error("Local/UI/Worker fingerprint mismatch; rebuild and reload");
  return {
    roots: info.roots,
    kinds: info.kinds,
    file: (
      root: string,
      path: string,
      kind: string,
      operation: "save" | "load",
      snapshot?: BrowserSnapshot,
    ) =>
      json<FileResult>("/api/v1/files", {
        root,
        path,
        kind,
        operation,
        UI: __FECIM_UI__,
        identity,
        snapshot: snapshot || {},
      }),
    logout: () => json("/api/v1/session/logout", {}),
  };
}
export type LocalAdapter = Awaited<ReturnType<typeof connectLocal>>;
