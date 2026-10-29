// Feature-owned qualification: actual production Go bridge in a Node worker
// thread and the production browser Worker/client in Firefox. JS never computes
// or reserializes scientific fields; only the original canonical strings travel.
import { readFileSync, writeFileSync, existsSync } from "node:fs";
import { resolve, join } from "node:path";
import { pathToFileURL } from "node:url";
import { runInThisContext } from "node:vm";
import { once } from "node:events";
import { createHash, webcrypto } from "node:crypto";
import { createServer } from "node:http";
import {
  Worker,
  isMainThread,
  parentPort,
  workerData,
} from "node:worker_threads";
import { brotliCompressSync, gzipSync } from "node:zlib";
if (!isMainThread) {
  globalThis.crypto ||= webcrypto;
  runInThisContext(readFileSync(workerData.shim, "utf8"));
  const go = new Go(),
    { instance } = await WebAssembly.instantiate(
      readFileSync(workerData.wasm),
      go.importObject,
    );
  go.run(instance).catch((error) =>
    parentPort.postMessage({ error: String(error) }),
  );
  parentPort.on("message", (data) =>
    parentPort.postMessage(fecimModuleDispatch(data)),
  );
  parentPort.postMessage("ready");
} else {
  const [wasmPath, shimPath, inputPath, dir] = process.argv.slice(2),
    fixtures = JSON.parse(readFileSync(inputPath, "utf8"));
  const wasm = readFileSync(wasmPath),
    shim = readFileSync(shimPath),
    driver = readFileSync("web/src/worker/hysteresis-worker.mjs"),
    hash = (b) => createHash("sha256").update(b).digest("hex");
  const raw = Buffer.concat([wasm, shim, driver]),
    budget = {
      raw: raw.length,
      brotli: brotliCompressSync(raw).length,
      gzip: gzipSync(raw).length,
    };
  if (
    budget.raw > 6 * 1024 ** 2 ||
    budget.brotli > 1.5 * 1024 ** 2 ||
    budget.gzip > 2 * 1024 ** 2
  )
    throw new Error(
      "Crossbar Worker payload gate exceeded: " + JSON.stringify(budget),
    );
  let browser, server;
  try {
    const outputs = [];
    for (const fixture of fixtures) {
      const w = new Worker(new URL(import.meta.url), {
        workerData: { wasm: wasmPath, shim: shimPath },
      });
      const receive = () => once(w, "message").then(([data]) => data);
      if ((await receive()) !== "ready")
        throw new Error("Node Worker startup failed");
      let req = { ...fixture };
      const call = async (op) => {
        const pending = receive();
        w.postMessage(JSON.stringify({ ...req, operation: op }));
        const r = JSON.parse(await pending);
        if (r.error) throw new Error(r.error);
        req = { ...req, request: req.request + 1, revision: r.revision };
        return r.result;
      };
      try {
        await call("action");
        await call("start");
        let result;
        for (let i = 0; i < 5000; i++) {
          result = await call("state");
          if (!["running", "cancelling"].includes(result.job.state)) break;
          await new Promise((r) => setTimeout(r, 1));
        }
        if (result.job.state !== "succeeded" || !result.results_json)
          throw new Error("Node did not finish science");
        outputs.push(result.results_json);
      } finally {
        await w.terminate();
      }
    }
    writeFileSync(join(dir, "node-crossbar.json"), JSON.stringify(outputs));
    const routes = new Map([
      [
        "/",
        [
          "text/html",
          Buffer.from("<!doctype html><title>Crossbar parity</title>"),
        ],
      ],
      ["/crossbar.wasm", ["application/wasm", wasm]],
      ["/wasm_exec.js", ["text/javascript", shim]],
      ["/worker.mjs", ["text/javascript", driver]],
    ]);
    server = createServer((req, res) => {
      const asset = routes.get(req.url);
      if (!asset) {
        res.writeHead(404).end();
        return;
      }
      res.writeHead(200, {
        "Content-Type": asset[0],
        "Cache-Control": "no-store",
      });
      res.end(asset[1]);
    });
    const base = Number(process.env.AUTOWORK_PORT_BASE || 0);
    await new Promise((resolve, reject) => {
      server.once("error", reject);
      server.listen(base ? base + 2 : 0, "127.0.0.1", resolve);
    });
    const { firefox } = await import(
      pathToFileURL(
        process.env.FECIM_PLAYWRIGHT_MODULE ||
          resolve("web/node_modules/playwright/index.mjs"),
      )
    );
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
    const result = await page.evaluate(
      async ({ fixtures, hashes }) => {
        const outputs = [];
        let interrupted = 0, completionRaces = 0;
        for (const fixture of fixtures) {
          const client = new HysteresisWorkerClient({
            identity: fixture.identity,
            session: fixture.session,
            prefix: "fecimModule",
            workerURL: "/worker.mjs",
            workerSHA256: hashes.driver,
            wasmURL: "/crossbar.wasm",
            wasmSHA256: hashes.wasm,
            runtimeURL: "/wasm_exec.js",
            runtimeSHA256: hashes.shim,
          });
          const call = async (op, input = fixture.input) => {
            const reply = JSON.parse(await client.request(op, input));
            if (reply.error) throw new Error(reply.error);
            return reply.result;
          };
          try {
            await client.ready;
            await call("action");
            await call("start");
            let r;
            for (let i = 0; i < 5000; i++) {
              r = await call("state");
              if (!["running", "cancelling"].includes(r.job.state)) break;
              await new Promise((r) => setTimeout(r, 1));
            }
            if (r.job.state !== "succeeded" || !r.results_json)
              throw new Error("Firefox did not finish science");
            const canonical = r.results_json;
            outputs.push(canonical);
            // Admission is Go-owned and atomic, not a TypeScript input restriction.
            let rejected = false;
            try {
              await call("action", {
                command: { action: "configure", input: { rows: "33" } },
              });
            } catch {
              rejected = true;
            }
            if (!rejected) throw new Error("hosted dimension cap bypassed");
            if ((await call("state")).results_json !== canonical)
              throw new Error("invalid action lost previous result");
            await call("start");
            const cancel = await call("cancel");
            if (!cancel.job.cancellation_requested) {
              if (cancel.job.state !== "succeeded" || cancel.results_json !== canonical) throw new Error("cancel neither acknowledged nor already complete");
              completionRaces++;
              continue;
            }
            for (let i = 0; i < 5000; i++) {
              r = await call("state");
              if (!["running", "cancelling"].includes(r.job.state)) break;
              await new Promise((r) => setTimeout(r, 1));
            }
            if (
              r.job.state !== "cancelled" ||
              r.job.progress.cancelled !== 1 ||
              r.results_json !== canonical
            )
              throw new Error(
                "real Worker cancellation did not retain previous science",
              );
            interrupted++;
          } finally {
            client.terminate();
          }
        }
        if (!interrupted) throw new Error("no real CPU computation was interrupted");
        return {
          outputs,
          interrupted, completionRaces,
          cancellation: "cancelled",
          profiles: fixtures.length,
        };
      },
      {
        fixtures,
        hashes: { driver: hash(driver), wasm: hash(wasm), shim: hash(shim) },
      },
    );
    writeFileSync(
      join(dir, "firefox-crossbar.json"),
      JSON.stringify(result.outputs),
    );
    console.log(
      JSON.stringify({
        runtimes: ["native", "node-worker", "firefox-worker"],
        profiles: result.profiles,
        cancellation: result.cancellation,
        interrupted: result.interrupted, completionRaces: result.completionRaces,
        budget,
        browser: browser.version(),
      }),
    );
  } finally {
    await browser?.close();
    if (server) await new Promise((r) => server.close(r));
  }
  process.exit(0);
}
