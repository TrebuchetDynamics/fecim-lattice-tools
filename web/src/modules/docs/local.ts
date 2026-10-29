// Loaded only by an explicit local-profile action. No discovery/probing is
// compiled into the hosted documentation screen.
import createClient from "openapi-fetch";
import type { paths } from "../../api/generated";
import type { Schema } from "../../api/types";
import { snapshotOutput } from "./data";

export async function localDocs() {
  const client = createClient<paths>({
    credentials: "same-origin",
    cache: "no-store",
    redirect: "error",
  });
  client.use({
    onRequest({ request }) {
      if (request.method !== "GET") {
        request.headers.set("Content-Type", "application/json");
        request.headers.set("X-FeCIM-Request", "1");
      }
      return request;
    },
  });
  function unwrap<T>(r: { data?: T; error?: unknown; response: Response }): T {
    if (!r.response.ok || r.error)
      throw new Error(
        "Local documentation request rejected; check the grant, format and current session",
      );
    return r.data!;
  }
  let snapshot = unwrap(
    await client.POST("/api/v1/modules/{module}/sessions", {
      params: { path: { module: "docs" } },
      body: {},
    }),
  );
  return {
    roots: async () => {
      const response = await fetch("/api/v1/roots", {
        credentials: "same-origin",
        cache: "no-store",
        redirect: "error",
      });
      if (!response.ok)
        throw new Error("Open the authenticated local launch link first");
      const value = await response.json();
      if (value.UI !== __FECIM_UI__)
        throw new Error("Local/UI fingerprint mismatch");
      return value.roots as { id: string; name: string }[];
    },
    projects: async () =>
      unwrap(
        await client.GET("/api/v1/projects", {
          params: { query: { page: 1, page_size: 32 } },
        }),
      ),
    openProject: async (root: string, path: string) =>
      unwrap(
        await client.POST("/api/v1/projects/open", { body: { root, path } }),
      ),
    bind: async (root: string, path: string) =>
      unwrap(
        await client.POST("/api/v1/resources", {
          body: { root, path, kind: "reference" },
        }),
      ),
    async call(
      action: string,
      input: Record<string, string> = {},
      resources: Schema<"ResourceReference">[] = [],
    ) {
      snapshot = unwrap(
        await client.POST("/api/v1/module-sessions/{session}/actions", {
          params: { path: { session: snapshot.session } },
          body: { action, input, resources, revision: snapshot.revision },
        }),
      );
      return snapshotOutput(snapshot);
    },
    async close() {
      unwrap(
        await client.POST("/api/v1/module-sessions/{session}/close", {
          params: { path: { session: snapshot.session } },
          body: {},
        }),
      );
    },
  };
}
export type LocalDocs = Awaited<ReturnType<typeof localDocs>>;
