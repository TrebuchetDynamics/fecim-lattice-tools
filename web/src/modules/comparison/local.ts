// Dynamically imported only by the authenticated local build. No hosted probe.
import createClient from "openapi-fetch";
import type { paths } from "../../api/generated";
import type { Schema } from "../../api/types";
const client = createClient<paths>({
  credentials: "same-origin",
  cache: "no-store",
  redirect: "error",
});
client.use({
  onRequest({ request }) {
    if (request.method !== "GET") {
      request.headers.set("X-FeCIM-Request", "1");
      request.headers.set("Content-Type", "application/json");
    }
    return request;
  },
});
function unwrap<T>(r: { data?: T; error?: unknown; response: Response }): T {
  if (!r.response.ok || r.error)
    throw new Error(
      "Comparison local request rejected. Check authentication, fields, project and revision.",
    );
  return r.data!;
}
export async function localComparison() {
  const roots = await fetch("/api/v1/roots", {
    credentials: "same-origin",
    cache: "no-store",
    redirect: "error",
  });
  if (!roots.ok)
    throw new Error("Authentication required: open the terminal launch link.");
  if ((await roots.json()).UI !== __FECIM_UI__)
    throw new Error("Local/UI fingerprint mismatch");
  let snapshot = unwrap(
    await client.POST("/api/v1/modules/{module}/sessions", {
      params: { path: { module: "comparison" } },
      body: {},
    }),
  );
  const session = snapshot.session;
  return {
    snapshot: async () => snapshot,
    apply: async (command: Schema<"Command">) =>
      (snapshot = unwrap(
        await client.POST("/api/v1/module-sessions/{session}/actions", {
          params: { path: { session } },
          body: command,
        }),
      )),
    close: () => {
      void client.POST("/api/v1/module-sessions/{session}/close", {
        params: { path: { session } },
        body: {},
      });
    },
    projects: async () => {
      const out: { id: string; name: string }[] = [];
      let page = 1;
      do {
        const p = unwrap(
          await client.GET("/api/v1/projects", {
            params: { query: { page, page_size: 32 } },
          }),
        );
        for (const pjt of p.projects || [])
          if (pjt.open_id) out.push({ id: pjt.open_id, name: pjt.path });
        if (!p.has_more) break;
        page++;
      } while (page <= 32);
      return out;
    },
  };
}
