import { expect, test } from "@playwright/test";
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { join } from "node:path";

// Register per-spec hooks: each spec gets its own grant and single-use launch link.
export function useResearchServer() {
  const server = {
    port: Number(process.env.AUTOWORK_PORT_BASE || 20200) + 12,
    root: "",
    launch: "",
  };
  if (process.env.FECIM_WEB_PROFILE !== "local") return server;

  let child: ReturnType<typeof spawn> | undefined;
  test.beforeAll(async () => {
    test.setTimeout(420_000);
    server.root = mkdtempSync(join(process.env.TMPDIR!, "research-local-"));
    const binary = join(server.root, "fecim");
    execFileSync("node", ["src/local/build-native.mjs", binary], { timeout: 360_000 });
    child = spawn(binary, ["serve", "--root", server.root, "--port", String(server.port)]);
    let output = "";
    child.stdout!.on("data", data => { output += data; });
    child.stderr!.on("data", data => { output += data; });
    await expect.poll(() => {
      server.launch = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)?.[0] || "";
      return Boolean(server.launch);
    }, { message: "local /: embedded server should emit a launch link", timeout: 30_000 }).toBe(true);
  });
  test.afterAll(async () => {
    if (child?.pid && child.exitCode === null && child.signalCode === null) {
      const exited = new Promise(resolve => child!.once("exit", resolve));
      child.kill("SIGTERM");
      await exited;
    }
    if (server.root) rmSync(server.root, { recursive: true, force: true });
  });
  return server;
}
