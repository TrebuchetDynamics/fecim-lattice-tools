// Feature-owned native/Node/Firefox qualification of the production CPU bridge.
// Canonical scientific strings are never parsed, rounded or rewritten by JS.
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
  go.run(instance).catch((e) => parentPort.postMessage({ error: String(e) }));
  parentPort.on("message", (data) =>
    parentPort.postMessage({
      reply: fecimModuleDispatch(data),
      memory: instance.exports.mem.buffer.byteLength,
    }),
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
    throw Error("Circuit payload cap exceeded: " + JSON.stringify(budget));
  let browser, server;
  const nodeProfiles = [];
  try {
    const outputs = [];
    for (const fixture of fixtures) {
      const w = new Worker(new URL(import.meta.url), {
          workerData: { wasm: wasmPath, shim: shimPath },
        }),
        receive = () => once(w, "message").then(([d]) => d);
      if ((await receive()) !== "ready")
        throw Error("Node Worker startup failed");
      let sequence = 1,
        revision = 0,
        memory = 0,
        heap = 0;
      const call = async (op, command = { action: "", revision: 0 }) => {
        const pending = receive();
        w.postMessage(
          JSON.stringify({
            identity: fixture.identity,
            session: fixture.session,
            request: sequence++,
            revision,
            operation: op,
            input: { command },
          }),
        );
        const message = await pending;
        memory = Math.max(memory, message.memory);
        const reply = JSON.parse(message.reply);
        revision = reply.revision;
        if (reply.error) throw Error(reply.error);
        heap = Math.max(heap, reply.result.resources.heap_alloc_bytes);
        return reply.result;
      };
      const started = performance.now(),
        result = [];
      try {
        for (const command of fixture.commands) {
          const science = ["write", "read", "compute"].includes(command.action);
          let r = await call(science ? "start" : "action", command);
          if (science) {
            for (
              let i = 0;
              i < 5000 && ["running", "cancelling"].includes(r.job.state);
              i++
            ) {
              await new Promise((r) => setTimeout(r, 1));
              r = await call("state");
            }
            if (r.job.state !== "completed" || !r.results_json)
              throw Error(
                "Node science did not complete: " + JSON.stringify(r.job),
              );
            result.push(r.results_json);
          }
        }
        const ms = performance.now() - started;
        if (ms > 15000 || memory > 128 * 1024 ** 2 || heap > 128 * 1024 ** 2)
          throw Error("Node CPU/memory profile cap exceeded");
        nodeProfiles.push({
          session: fixture.session,
          ms,
          wasmMemoryBytes: memory,
          observedGoHeapBytes: heap,
        });
        outputs.push(result);
      } finally {
        await w.terminate();
      }
    }
    writeFileSync(join(dir, "node-circuits.json"), JSON.stringify(outputs));
    const routes = new Map([
      [
        "/",
        [
          "text/html",
          Buffer.from("<!doctype html><title>Circuits parity</title>"),
        ],
      ],
      ["/circuits.wasm", ["application/wasm", wasm]],
      ["/wasm_exec.js", ["text/javascript", shim]],
      ["/worker.mjs", ["text/javascript", driver]],
    ]);
    server = createServer((req, res) => {
      const a = routes.get(req.url);
      if (!a) {
        res.writeHead(404).end();
        return;
      }
      res.writeHead(200, { "Content-Type": a[0], "Cache-Control": "no-store" });
      res.end(a[1]);
    });
    const base = Number(process.env.AUTOWORK_PORT_BASE || 0);
    await new Promise((resolve, reject) => {
      server.once("error", reject);
      server.listen(base ? base + 3 : 0, "127.0.0.1", resolve);
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
        : "",
      executablePath =
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
        const outputs = [],
          profiles = [];
        let interrupted = 0,
          completionRaces = 0;
        for (const fixture of fixtures) {
          const client = new HysteresisWorkerClient({
            identity: fixture.identity,
            session: fixture.session,
            prefix: "fecimModule",
            workerURL: "/worker.mjs",
            workerSHA256: hashes.driver,
            wasmURL: "/circuits.wasm",
            wasmSHA256: hashes.wasm,
            runtimeURL: "/wasm_exec.js",
            runtimeSHA256: hashes.shim,
          });
          let heap = 0;
          const call = async (op, command = { action: "", revision: 0 }) => {
            const reply = JSON.parse(await client.request(op, { command }));
            if (reply.error) throw Error(reply.error);
            heap = Math.max(heap, reply.result.resources.heap_alloc_bytes);
            return reply.result;
          };
          try {
            await client.ready;
            const started = performance.now(),
              results = [];
            for (const command of fixture.commands) {
              const science = ["write", "read", "compute"].includes(
                command.action,
              );
              let r = await call(science ? "start" : "action", command);
              if (science) {
                for (
                  let i = 0;
                  i < 5000 && ["running", "cancelling"].includes(r.job.state);
                  i++
                ) {
                  await new Promise((r) => setTimeout(r, 1));
                  r = await call("state");
                }
                if (r.job.state !== "completed" || !r.results_json)
                  throw Error("Firefox science did not complete");
                results.push(r.results_json);
              }
            }
            const ms = performance.now() - started;
            if (ms > 15000 || heap > 128 * 1024 ** 2)
              throw Error("Firefox CPU/Go-heap profile cap exceeded");
            profiles.push({
              session: fixture.session,
              ms,
              observedGoHeapBytes: heap,
            });
            outputs.push(results);
            const canonical = (await call("state")).results_json;
            let rejected = false;
            try {
              await call("action", {
                action: "configure",
                input: { rows: "16" },
                revision: 0,
              });
            } catch {
              rejected = true;
            }
            if (!rejected || (await call("state")).results_json !== canonical)
              throw Error("invalid configuration was not atomic");
            await call("start", {
              action: "write",
              input: { target: "0" },
              revision: 0,
            });
            let r = await call("cancel");
            if (!r.job.cancellation_requested) {
              if (r.job.state !== "completed")
                throw Error("cancel neither acknowledged nor complete");
              completionRaces++;
              continue;
            }
            for (
              let i = 0;
              i < 5000 && ["running", "cancelling"].includes(r.job.state);
              i++
            ) {
              await new Promise((r) => setTimeout(r, 1));
              r = await call("state");
            }
            if (r.job.state !== "cancelled" || r.results_json !== canonical)
              throw Error("cancel changed previously committed science");
            interrupted++;
          } finally {
            client.terminate();
          }
        }
        if (!interrupted) throw Error("no actual CPU write was interrupted");
        return { outputs, profiles, interrupted, completionRaces };
      },
      {
        fixtures,
        hashes: { driver: hash(driver), wasm: hash(wasm), shim: hash(shim) },
      },
    );
    writeFileSync(
      join(dir, "firefox-circuits.json"),
      JSON.stringify(result.outputs),
    );
    const receipt = {
      runtimes: ["native", "node-worker", "firefox-production-worker"],
      budget,
      nodeProfiles,
      firefoxProfiles: result.profiles,
      interrupted: result.interrupted,
      completionRaces: result.completionRaces,
      browser: browser.version(),
      memoryBoundary:
        "Node committed WASM linear memory and both Workers' Go heap at observed responses; not process RSS or a peak-heap claim",
    };
    writeFileSync(join(dir, "circuits-resource.json"), JSON.stringify(receipt));
    console.log(JSON.stringify(receipt));
  } finally {
    await browser?.close();
    if (server) await new Promise((r) => server.close(r));
  }
}
