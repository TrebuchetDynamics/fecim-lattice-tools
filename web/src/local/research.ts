// This module is dynamically imported only by the local build branch. Hosted
// artifacts do not contain an HTTP scientific transport or engine discovery.
import createClient from "openapi-fetch";
import type { paths } from "../api/generated";
import { RequestError, type ResearchAdapter, type Schema, type Problem } from "../api/types";
const client = createClient<paths>({ credentials: "same-origin", cache: "no-store", redirect: "error" });
client.use({ onRequest({ request }) {
 if (request.method !== "GET") { request.headers.set("X-FeCIM-Request", "1"); request.headers.set("Content-Type", "application/json"); }
 return request;
}});
function unwrap<T>(reply: { data?: T; error?: unknown; response: Response }): T {
 if (reply.error || !reply.response.ok) {
  const p = reply.error as Partial<Problem> | undefined;
  throw new RequestError({ type: p?.type || "urn:fecim:request", title: p?.title || "Request failed", status: reply.response.status, detail: p?.detail || "Request failed; retry or reopen the project", ...p });
 }
 return reply.data!;
}
export async function localResearch(): Promise<ResearchAdapter> {
 // Reuse the same authenticated roots endpoint and existing build identities.
 const response = await fetch("/api/v1/roots", { credentials: "same-origin", cache: "no-store", redirect: "error" });
 if (!response.ok) throw new Error("Authentication required. Open the one-time launch link from the terminal.");
 const roots = await response.json() as { roots: {id:string;name:string}[]; UI:string };
 if (roots.UI !== __FECIM_UI__) throw new Error("Local/UI fingerprint mismatch; rebuild and reload");
 const p = (project: string) => ({ path: { project } });
 return {
  roots: async () => roots.roots,
  projects: async (page = 1) => unwrap(await client.GET("/api/v1/projects", { params: { query: { page, page_size: 32 } } })),
  open: async (root, path) => unwrap(await client.POST("/api/v1/projects/open", { body: { root, path } })),
  create: async (root, name) => unwrap(await client.POST("/api/v1/projects", { body: { root, name, options: { ConductanceLevels:[16,30],Rows:[16,32],ADCBits:[4,6] } } })),
  example: async () => { throw new Error("Use a granted native project or create a starter project"); },
  import: async () => { throw new Error("Open projects inside launch-granted roots"); },
  configuration: async id => unwrap(await client.GET("/api/v1/projects/{project}/configuration", { params: p(id) })),
  save: async (id, revision, document, changes) => unwrap(await client.PUT("/api/v1/projects/{project}/configuration", { params: { ...p(id), header: { "If-Match": revision } }, body: { document: document as Schema<"PatchBody">["document"], changes } })),
  validate: async id => unwrap(await client.POST("/api/v1/projects/{project}/validate", { params: p(id), body: {} })),
  close: async id => { unwrap(await client.POST("/api/v1/projects/{project}/close", { params: p(id), body: {} })); },
  start: async (id, revision, workers, key) => unwrap(await client.POST("/api/v1/projects/{project}/experiments", { params: { ...p(id), header: { "If-Match": revision, "Idempotency-Key": key } }, body: { workers } })),
  state: async execution => unwrap(await client.GET("/api/v1/experiments/{execution}", { params: { path: { execution } } })),
  cancel: async execution => unwrap(await client.POST("/api/v1/experiments/{execution}/cancel", { params: { path: { execution } }, body: {} })),
  active: async id => unwrap(await client.GET("/api/v1/projects/{project}/experiments", { params: { ...p(id), query: { state:"active" } } })).experiments || [],
  analysis: async (id, query, compare) => unwrap(await client.GET("/api/v1/projects/{project}/analysis", { params: { ...p(id), query: { ...query, compare: compare?.join(",") } } })),
  inspect: async (id, run) => JSON.stringify(unwrap(await client.GET("/api/v1/projects/{project}/runs/{run}", { params: { path: { project:id, run } } })), null, 2),
  verify: async id => unwrap(await client.POST("/api/v1/projects/{project}/verify", { params: p(id), body: {} })),
  reports: async id => unwrap(await client.POST("/api/v1/projects/{project}/reports", { params: p(id), body: {} })),
  export: async (id, run) => unwrap(await client.POST("/api/v1/projects/{project}/exports", { params: p(id), body: { run_id:run } })),
  modules: async () => unwrap(await client.GET("/api/v1/modules")).modules || [],
 };
}
