// Qualification only. Go owns science and every serialized scientific byte.
import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { resolve, join, dirname } from 'node:path';
import { pathToFileURL } from 'node:url';
import { createServer } from 'node:http';
import { runInThisContext } from 'node:vm';
import { createHash, webcrypto } from 'node:crypto';
import { gzipSync, brotliCompressSync, constants } from 'node:zlib';
import { cpus, platform, arch } from 'node:os';

const [mode, wasmPath, shimPath, inputPath, dir] = process.argv.slice(2);
if (!dir) throw new Error('usage: hysteresis_harness.mjs MODE WASM SHIM INPUT OUTPUT');
const wasm = readFileSync(wasmPath), shim = readFileSync(shimPath);
const driver = readFileSync('web/src/worker/hysteresis-worker.mjs');
const fixture = JSON.parse(readFileSync(inputPath, 'utf8'));
// Keep original Go request bytes: parse/stringify erases signed zero.
const requests = (fixture.parity || fixture).map(q => typeof q === 'string' ? q : JSON.stringify(q));
const identity = JSON.parse(requests[0]).identity;
const sha = bytes => createHash('sha256').update(bytes).digest('hex');
const br = bytes => brotliCompressSync(bytes, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } });
const sizes = bytes => ({ raw: bytes.length, gzip: gzipSync(bytes, { level: 9 }).length, brotli: br(bytes).length, sha256: sha(bytes) });
const payload = { wasm: sizes(wasm), runtime: sizes(shim), worker: sizes(driver) };
const total = Object.values(payload).reduce((a, s) => ({ raw: a.raw + s.raw, gzip: a.gzip + s.gzip, brotli: a.brotli + s.brotli }), { raw: 0, gzip: 0, brotli: 0 });
globalThis.crypto ||= webcrypto;
runInThisContext(shim.toString(), { filename: shimPath });
const go = new Go();
const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
go.run(instance).catch(error => { console.error(error); process.exitCode = 1; });
if (JSON.stringify(JSON.parse(fecimHysteresisIdentity)) !== JSON.stringify(identity)) throw new Error('Node Go identity mismatch');
const nodeReplies = requests.map(q => fecimHysteresisDispatch(q));
writeFileSync(join(dir, 'node-results.json'), `[${nodeReplies.join(',')}]\n`);

const playwrightPath = process.env.FECIM_PLAYWRIGHT_MODULE || resolve('web/node_modules/playwright/index.mjs');
if (!existsSync(playwrightPath)) throw new Error('locked Playwright required; see README.md overrides');
const { firefox } = await import(pathToFileURL(playwrightPath));
const routes = new Map([
 ['/', ['text/html', Buffer.from('<!doctype html><title>Production hysteresis qualification</title>')]],
 ['/hysteresis.wasm', ['application/wasm', wasm]],
 ['/wasm_exec.js', ['text/javascript', shim]],
 ['/hysteresis-worker.mjs', ['text/javascript', driver]],
]);
const encoded = new Map([...routes].map(([url, [mime, bytes]]) => [url, [mime, br(bytes)]]));
const transfer = [];
let shaped = false, nextByteAt = 0;
const delay = ms => new Promise(resolve => setTimeout(resolve, Math.max(0, ms)));
const server = createServer(async (request, response) => {
 const route = encoded.get(request.url);
 if (!route) { response.writeHead(404).end(); return; }
 const [mime, bytes] = route;
 const profile = shaped;
 transfer.push({ url: request.url, bytes: bytes.length, shaped: profile });
 if (profile) await delay(100); // added request round-trip delay, private server only
 response.writeHead(200, { 'Content-Type': mime, 'Content-Encoding': 'br', 'Cache-Control': 'no-store', 'Content-Length': bytes.length });
 // Reserve every chunk on one shared timeline: concurrent module/runtime
 // fetches share 10,000,000 bits/s, not a per-resource 10 Mbit/s allowance.
 for (let i = 0; i < bytes.length; i += 16384) {
  const chunk = bytes.subarray(i, i + 16384);
  if (profile) { nextByteAt = Math.max(performance.now(), nextByteAt) + chunk.length * 8 / 10_000; await delay(nextByteAt - performance.now()); }
  if (response.destroyed) break;
  response.write(chunk);
 }
 response.end();
});
let browser;
try {
 const base = Number(process.env.AUTOWORK_PORT_BASE || 0);
 await new Promise((resolve, reject) => { server.once('error', reject); server.listen(base ? base + 1 : 0, '127.0.0.1', resolve); });
 browser = await firefox.launch({ headless: true, ...(process.env.FECIM_FIREFOX_EXECUTABLE ? { executablePath: process.env.FECIM_FIREFOX_EXECUTABLE } : {}) });
 const origin = `http://127.0.0.1:${server.address().port}`;
 const options = { workerURL: '/hysteresis-worker.mjs', wasmURL: '/hysteresis.wasm', runtimeURL: '/wasm_exec.js', identity, wasmSHA256: sha(wasm), runtimeSHA256: sha(shim), workerSHA256: sha(driver), session: 'qualification' };
 const context = await browser.newContext();
 const page = await context.newPage();
 page.setDefaultTimeout(90_000);
 await page.goto(origin);
 await page.addScriptTag({ url: '/hysteresis-worker.mjs' });
 const measured = await page.evaluate(async ({ requests, options, mode }) => {
  // Identity is a typed object, not a JSON-property-order contract.
  const reordered = Object.fromEntries(Object.entries(options.identity).reverse());
  const client = new HysteresisWorkerClient({ ...options, identity: reordered });
  const replies = [], admission = [];
  try {
   await client.ready;
   for (const originalJSON of requests) {
    const q = JSON.parse(originalJSON); // metadata only, never scientific reserialization
    const before = performance.now();
    const json = await client.send(originalJSON);
    const elapsedMS = performance.now() - before;
    const reply = JSON.parse(json);
    if (reply.error) throw new Error(`production ${q.operation}: ${reply.error}`);
    replies.push(json); admission.push({ operation: q.operation, elapsedMS, input: q.input, revision: reply.revision });
   }
   const lifecycle = [];
   if (mode === 'lifecycle') {
    // Real cancellation, stale revision rejection, unavailable mode and
    // subsequent replacement run through the production Go Worker.
    const rejected = JSON.parse(await client.request('unsupported.operation'));
    if (!rejected.error?.includes('unsupported operation') || rejected.committed) throw new Error('Go protocol rejection missing');
    const unchanged = JSON.parse(await client.request('core.snapshot'));
    if (unchanged.error || unchanged.result.Config.Seed !== 42) throw new Error('invalid operation changed live session');
    const cancelled = JSON.parse(replies.at(-1));
    if (cancelled.result.State !== 'cancelled' && cancelled.result.Status !== 'cancelled') throw new Error('untruthful cancellation');
    let stale = false;
    try { await client.send(requests[0]); } catch { stale = true; }
    if (!stale) throw new Error('stale revision published');
    await client.recreate();
    let retiredSettled = false;
    const retired = client.request('session.new', {}).then(() => { throw new Error('retired same-turn request published'); }, () => { retiredSettled = true; });
    await client.recreate(); // no intervening yield after request()
    let raceTimer;
    try {
     await Promise.race([retired, new Promise((_, reject) => { raceTimer = setTimeout(() => reject(new Error('retired same-turn send remained pending')), 1000); })]);
    } finally { clearTimeout(raceTimer); }
    if (!retiredSettled || client.pending) throw new Error('retired send occupied replacement');
    const first = JSON.parse(await client.request('session.new', {}));
    if (first.error || first.revision !== 1) throw new Error('recreate did not discard session');
    const retiredHandler = client.worker.onmessage;
    const pending = client.request('core.snapshot');
    // Let send enqueue before terminating; no mocked numerical happy path.
    await new Promise(resolve => setTimeout(resolve, 0));
    client.terminate();
    await pending.then(() => { throw new Error('terminated request published'); }, () => {});
    if (client.state !== 'cancelled') throw new Error('incorrect hard-cancel status');
    await client.recreate();
    await client.request('session.new', {});
    // A late error from the retired generation must not poison the new session.
    const live = client.request('core.snapshot');
    retiredHandler({ data: { type: 'fault', error: 'injected late retired Worker error' } });
    if (JSON.parse(await live).error) throw new Error('retired generation poisoned replacement');
    // Explicit response corruption only: intercepted production message is
    // given a stale request ID. The client must fail closed, not publish it.
    const handler = client.worker.onmessage;
    client.worker.onmessage = event => {
     if (event.data.type === 'reply') { const r = JSON.parse(event.data.json); r.request++; event = { data: { type: 'reply', json: JSON.stringify(r) } }; }
     handler(event);
    };
    await client.request('core.snapshot').then(() => { throw new Error('fault-injected stale reply published'); }, () => {});
    if (client.state !== 'failed') throw new Error('stale reply not terminal');
    for (const override of [{ workerSHA256: '0'.repeat(64) }, { wasmSHA256: '0'.repeat(64) }, { runtimeSHA256: '0'.repeat(64) }, { identity: { ...options.identity, build: 'drifted-build' } }, { wasmURL: '/missing.wasm' }]) {
     const broken = new HysteresisWorkerClient({ ...options, ...override });
     try {
      await broken.ready.then(() => { throw new Error('drifted/missing asset admitted'); }, () => {});
      if (broken.state !== 'failed') throw new Error('startup fault not terminal');
     } finally { broken.terminate(); }
    }
    lifecycle.push('real cancel', 'Go rejects unsupported without mutation', 'stale request rejected', 'terminate rejects pending', 'recreate resets session', 'same-turn retired send rejected', 'retired generation ignored', 'fault-injected stale response rejected', 'Worker/module/runtime hash drift rejected', 'build drift rejected', 'missing asset rejected');
   }
   return { json: `[${replies.join(',')}]\n`, admission, lifecycle, userAgent: navigator.userAgent };
  } finally { client.terminate(); }
 }, { requests, options, mode });
 writeFileSync(join(dir, 'firefox-results.json'), measured.json);
 const cold = [];
 if (mode === 'budgets') {
  if (!fixture.cold) throw new Error('missing Go-defined cold Sine fixture');
  if (total.raw > 6 * 2 ** 20 || total.brotli > 1.5 * 2 ** 20 || total.gzip > 2 * 2 ** 20) throw new Error(`production payload guardrail missed: ${JSON.stringify(total)}`);
  for (let repeat = 0; repeat < 3; repeat++) {
   shaped = true; nextByteAt = 0;
   const coldContext = await browser.newContext();
   const coldPage = await coldContext.newPage();
   await coldPage.goto(origin); // driver marker starts before client/Worker fetch
   const firstTransfer = transfer.length;
   const record = await coldPage.evaluate(async ({ options, trace }) => {
    const start = performance.now();
    await import('/hysteresis-worker.mjs');
    const client = new HysteresisWorkerClient(options);
    try {
     await client.ready;
     let final;
     for (const q of trace) { final = JSON.parse(await client.send(q)); if (final.error) throw new Error(final.error); }
     if (final.result.Total < 32 || final.result.Rows.length !== 32) throw new Error('cold marker lacks complete loop table');
     return { elapsedMS: performance.now() - start, marker: 'cold-sine-loop-complete: Go pause + first 32-row page after full signed cycle', total: final.result.Total };
    } finally { client.terminate(); }
   }, { options, trace: fixture.cold });
   record.transfer = transfer.slice(firstTransfer);
   cold.push(record);
   await coldContext.close();
   if (record.elapsedMS > 5000) throw new Error(`cold production Sine guardrail missed: ${record.elapsedMS} ms`);
  }
 }
 // These are per-dispatch engineering admission bounds, not the unrelated
 // example-only 250 ms calculation budget or whole-job convergence promises.
 const admissionLimitsMS = { dispatch: 1000, cancellation: 500 };
 const perOperation = Object.fromEntries([...new Set(measured.admission.map(x => x.operation))].map(operation => {
  const samples = measured.admission.filter(x => x.operation === operation).map(x => x.elapsedMS);
  return [operation, { count: samples.length, maxMS: Math.max(...samples), limitMS: operation === 'job.cancel' ? admissionLimitsMS.cancellation : admissionLimitsMS.dispatch }];
 }));
 const measurements = {
  identity, payload, total, node: process.version, firefox: browser.version(), firefoxExecutable: process.env.FECIM_FIREFOX_EXECUTABLE || firefox.executablePath(),
  playwright: JSON.parse(readFileSync(join(dirname(playwrightPath), 'package.json'))).version,
  platform: platform(), arch: arch(), cpu: cpus()[0]?.model, logicalCPUs: cpus().length, browserUserAgent: measured.userAgent,
  admission: measured.admission, perOperation, admissionLimitsMS,
  admittedCaps: { sessionJobs: 1, coreProgramStepsPerAdvance: 256, analysisWorkUnitsPerAdvance: 1, retainedSamples: 100000, plotPoints: 2000, tableRows: 32, programmingDomains: 128, forcReversals: 64, forcResolution: 256, requestBytes: 1048576 },
  lifecycle: measured.lifecycle, cold,
  byteFilter: 'Sum raw and separately gzip-9/Brotli-11 encoded production hysteresis.wasm + matching wasm_exec.js + hysteresis-worker.mjs. Browser cold transfer includes driver import and dedicated Worker fetch (driver twice), module/runtime; excludes empty HTML navigation before marker. No React-route claim.',
  shaping: 'Private server only: added 100 ms request RTT, one aggregate serialized 10,000,000 bit/s chunk scheduler across responses, Brotli served with no-store. Three fresh browser contexts/workers; no host-wide networking or CPU throttling.',
 };
 writeFileSync(join(dir, `measurements-${mode}.json`), JSON.stringify(measurements, null, 2) + '\n');
 console.log(JSON.stringify({ mode, identity, total, cold, lifecycle: measured.lifecycle, perOperation }, null, 2));
 if (mode === 'budgets' && Object.values(perOperation).some(m => m.maxMS > m.limitMS)) throw new Error('production per-operation admission guardrail missed; see measurements-budgets.json');
 await context.close();
} finally {
 if (browser) await browser.close();
 await new Promise(resolve => server.close(resolve));
}
process.exit(process.exitCode || 0);
