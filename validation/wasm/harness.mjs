// Validation driver only: transport, browser automation and measurements.
// All scientific values/identities/analysis/JSON come from the compiled Go code.
import { readFileSync, writeFileSync, existsSync } from 'node:fs';
import { resolve, join, dirname } from 'node:path';
import { pathToFileURL } from 'node:url';
import { createServer } from 'node:http';
import { runInThisContext } from 'node:vm';
import { gzipSync, brotliCompressSync, constants } from 'node:zlib';
import { cpus, platform, arch } from 'node:os';
import { execFileSync } from 'node:child_process';

const [wasmPath, shimPath, inputPath, outputDir] = process.argv.slice(2);
if (!outputDir) throw new Error('usage: node harness.mjs WASM SHIM INPUT OUTPUT_DIR');
const wasm = readFileSync(wasmPath);
const shim = readFileSync(shimPath);
const input = readFileSync(inputPath, 'utf8');

const workerSource = `
const began = performance.now();
importScripts('/wasm_exec.js');
const go = new Go();
(async () => {
  const { instance } = await WebAssembly.instantiateStreaming(fetch('/study.wasm'), go.importObject);
  go.run(instance).catch(error => postMessage({ type: 'error', error: String(error) }));
  postMessage({ type: 'ready', startupMS: performance.now() - began });
})().catch(error => postMessage({ type: 'error', error: String(error) }));
onmessage = ({ data }) => {
  try {
    const start = performance.now();
    const result = JSON.parse(globalThis.fecimParityRun(data.input));
    postMessage({ type: 'result', id: data.id, ...result, runMS: performance.now() - start });
  } catch (error) { postMessage({ type: 'error', error: String(error) }); }
};
`;

let browser;
let server;
try {
  const startupStart = performance.now();
  runInThisContext(shim.toString('utf8'), { filename: shimPath });
  const go = new globalThis.Go();
  const { instance } = await WebAssembly.instantiate(wasm, go.importObject);
  go.run(instance).catch(error => { console.error(error); process.exitCode = 1; });
  const nodeStartupMS = performance.now() - startupStart;
  const nodeStart = performance.now();
  const nodeResult = JSON.parse(globalThis.fecimParityRun(input));
  const nodeRunMS = performance.now() - nodeStart;
  if (nodeResult.error) throw new Error(`Node Go worker: ${nodeResult.error}`);
  writeFileSync(join(outputDir, 'node-results.json'), nodeResult.resultsJSON);

  const playwrightPath = process.env.FECIM_PLAYWRIGHT_MODULE || resolve('web/node_modules/playwright/index.mjs');
  if (!existsSync(playwrightPath)) throw new Error('Playwright is required; install the locked web development dependencies or set FECIM_PLAYWRIGHT_MODULE');
  const { firefox } = await import(pathToFileURL(playwrightPath));
  const routes = new Map([
    ['/', ['text/html', Buffer.from('<!doctype html><title>Go WASM parity fixture</title>')]],
    ['/study.wasm', ['application/wasm', wasm]],
    ['/wasm_exec.js', ['text/javascript', shim]],
    ['/worker.js', ['text/javascript', Buffer.from(workerSource)]],
  ]);
  server = createServer((request, response) => {
    const route = routes.get(request.url);
    if (!route) { response.writeHead(404).end(); return; }
    response.writeHead(200, { 'Content-Type': route[0], 'Cache-Control': 'no-store' });
    response.end(route[1]);
  });
  await new Promise((resolve, reject) => { server.once('error', reject); server.listen(0, '127.0.0.1', resolve); });
  browser = await firefox.launch({ headless: true, ...(process.env.FECIM_FIREFOX_EXECUTABLE ? { executablePath: process.env.FECIM_FIREFOX_EXECUTABLE } : {}) });
  const page = await browser.newPage();
  page.setDefaultTimeout(60_000);
  await page.goto(`http://127.0.0.1:${server.address().port}/`);
  let browserTimer;
  const browserDeadline = new Promise((_, reject) => {
    browserTimer = setTimeout(() => reject(new Error('Firefox Worker did not finish within 60 seconds')), 60_000);
  });
  const measured = await Promise.race([page.evaluate(async input => {
    const before = performance.now();
    const worker = new Worker('/worker.js');
    const results = [];
    let readyMS;
    try {
      await new Promise((resolve, reject) => {
        worker.onerror = event => reject(new Error(event.message));
        worker.onmessage = ({ data }) => {
          if (data.type === 'error') reject(new Error(data.error));
          if (data.type === 'ready') { readyMS = performance.now() - before; resolve(); }
        };
      });
      // First invocation is cold. Six further invocations expose repeatability
      // and give a small descriptive sample, not a statistical performance claim.
      for (let id = 0; id < 7; id++) {
        results.push(await new Promise((resolve, reject) => {
          worker.onerror = event => reject(new Error(event.message));
          worker.onmessage = ({ data }) => {
            if (data.type === 'error' || data.error) reject(new Error(data.error));
            else if (data.type === 'result' && data.id === id) resolve(data);
          };
          worker.postMessage({ id, input });
        }));
      }
      if (results.some(result => result.resultsJSON !== results[0].resultsJSON)) throw new Error('Repeated Firefox calculations changed results.json bytes');
      return { resultsJSON: results[0].resultsJSON, workerReadyMS: readyMS, runMS: results.map(r => r.runMS), userAgent: navigator.userAgent };
    } finally { worker.terminate(); }
  }, input), browserDeadline]).finally(() => clearTimeout(browserTimer));
  writeFileSync(join(outputDir, 'firefox-results.json'), measured.resultsJSON);

  const sizes = bytes => ({ raw: bytes.length, gzip: gzipSync(bytes, { level: 9 }).length, brotli: brotliCompressSync(bytes, { params: { [constants.BROTLI_PARAM_QUALITY]: 11 } }).length });
  const measurements = {
    go: execFileSync('go', ['version'], { encoding: 'utf8' }).trim(),
    node: process.version, firefox: browser.version(), firefoxExecutable: process.env.FECIM_FIREFOX_EXECUTABLE || firefox.executablePath(),
    playwright: JSON.parse(readFileSync(join(dirname(playwrightPath), 'package.json'), 'utf8')).version,
    platform: platform(), arch: arch(), cpu: cpus()[0]?.model, logicalCPUs: cpus().length,
    wasm: sizes(wasm), runtimeShim: sizes(shim), workerShim: sizes(Buffer.from(workerSource)),
    nodeStartupMS, nodeRunMS, firefoxWorkerReadyMS: measured.workerReadyMS, firefoxRunMS: measured.runMS,
    browserUserAgent: measured.userAgent,
    conditions: 'Headless Firefox, real dedicated Worker, uncompressed assets over loopback HTTP, no network/CPU throttling; cache disabled; one fresh worker and seven calculations. Not CDN or cold-device load performance.',
  };
  writeFileSync(join(outputDir, 'measurements.json'), JSON.stringify(measurements, null, 2) + '\n');
  console.log(JSON.stringify(measurements, null, 2));
} finally {
  if (browser) await browser.close();
  if (server) await new Promise(resolve => server.close(resolve));
}
// The validation Go main deliberately remains alive to service calls. This
// driver owns that instance, so terminate it after artifacts are written.
process.exit(process.exitCode || 0);
