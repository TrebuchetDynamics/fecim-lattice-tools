// Feature-owned parity fixture: identical typed Go requests through Node and the
// production Firefox Worker. Only original Go scientific JSON bytes are saved.
import { readFileSync, writeFileSync, existsSync } from "node:fs";
import { resolve, join } from "node:path";
import { pathToFileURL } from "node:url";
import { runInThisContext } from "node:vm";
import { createHash, webcrypto } from "node:crypto";
import { createServer } from "node:http";
const [wasmPath, shimPath, inputPath, dir] = process.argv.slice(2);
const wasm = readFileSync(wasmPath),
  shim = readFileSync(shimPath),
  driver = readFileSync("web/src/worker/hysteresis-worker.mjs");
const requests = JSON.parse(readFileSync(inputPath, "utf8"));
const hash = (b) => createHash("sha256").update(b).digest("hex");
const save = (runtime, replies) => {
  if (replies.length !== requests.length)
    throw new Error("incomplete comparison transcript");
  for (let i = 0; i < replies.length; i++) {
    const r = JSON.parse(replies[i]);
    if (
      r.error ||
      !r.committed ||
      !r.results_json ||
      !r.result.tables?.some((t) => t.id === "energy" && t.rows.length === 3)
    )
      throw new Error("unsuccessful comparison science");
    writeFileSync(join(dir, `${runtime}-${i}.json`), r.results_json);
    writeFileSync(join(dir, `${runtime}-reply-${i}.json`), replies[i]);
  }
};
globalThis.crypto ||= webcrypto;
runInThisContext(shim.toString(), { filename: shimPath });
const go = new Go(),
  { instance } = await WebAssembly.instantiate(wasm, go.importObject);
go.run(instance).catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
if (
  JSON.stringify(JSON.parse(fecimModuleIdentity)) !==
  JSON.stringify(requests[0].identity)
)
  throw new Error("Node identity mismatch");
save(
  "node",
  requests.map((q) => fecimModuleDispatch(JSON.stringify(q))),
);
const playwright =
  process.env.FECIM_PLAYWRIGHT_MODULE ||
  resolve("web/node_modules/playwright/index.mjs");
const { firefox } = await import(pathToFileURL(playwright));
const routes = new Map([
  [
    "/",
    [
      "text/html",
      Buffer.from("<!doctype html><title>Comparison production parity</title>"),
    ],
  ],
  ["/comparison.wasm", ["application/wasm", wasm]],
  ["/wasm_exec.js", ["text/javascript", shim]],
  ["/worker.mjs", ["text/javascript", driver]],
]);
const server = createServer((req, res) => {
  const found = routes.get(req.url);
  if (!found) {
    res.writeHead(404).end();
    return;
  }
  res.writeHead(200, { "Content-Type": found[0], "Cache-Control": "no-store" });
  res.end(found[1]);
});
let browser;
try {
  const base = Number(process.env.AUTOWORK_PORT_BASE || 0);
  await new Promise((resolve, reject) => {
    server.once("error", reject);
    server.listen(base ? base + 2 : 0, "127.0.0.1", resolve);
  });
  const cached = process.env.AUTOWORK_CACHE
    ? join(
        process.env.AUTOWORK_CACHE,
        "playwright-1.60.0/firefox-1522/firefox/firefox",
      )
    : "";
  const executablePath =
    process.env.FECIM_FIREFOX_EXECUTABLE ||
    (existsSync(cached) ? cached : undefined);
  browser = await firefox.launch({
    headless: true,
    ...(executablePath ? { executablePath } : {}),
  });
  const page = await browser.newPage();
  await page.goto(`http://127.0.0.1:${server.address().port}`);
  await page.addScriptTag({ url: "/worker.mjs" });
  const replies = await page.evaluate(
    async ({ requests, hashes }) => {
      const first = requests[0];
      const client = new HysteresisWorkerClient({
        identity: first.identity,
        session: first.session,
        prefix: "fecimModule",
        workerURL: "/worker.mjs",
        workerSHA256: hashes.driver,
        wasmURL: "/comparison.wasm",
        wasmSHA256: hashes.wasm,
        runtimeURL: "/wasm_exec.js",
        runtimeSHA256: hashes.shim,
      });
      try {
        await client.ready;
        const out = [];
        for (const q of requests)
          out.push(await client.send(JSON.stringify(q)));
        let rejected = false;
        try {
          await client.send(JSON.stringify(first));
        } catch {
          rejected = true;
        }
        if (!rejected) throw new Error("stale comparison request admitted");
        return out;
      } finally {
        client.terminate();
      }
    },
    {
      requests,
      hashes: { driver: hash(driver), wasm: hash(wasm), shim: hash(shim) },
    },
  );
  save("firefox", replies);
  console.log(
    JSON.stringify({
      runtimes: ["native", "node", "firefox"],
      operations: requests.length,
      productionEnvelope: true,
      originalScientificBytes: true,
      staleRejected: true,
      browser: browser.version(),
    }),
  );
} finally {
  await browser?.close();
  await new Promise((resolve) => server.close(resolve));
}
process.exit(0);
