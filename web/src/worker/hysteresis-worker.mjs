// One production dedicated Worker and its transport-only owner. Scientific
// defaults, operation registry, admission, values and JSON all belong to Go.
// Load as a classic Worker; importing this file on Window installs the client.
(() => {
  const same = (a, b) => a && b && Object.keys(a).length === 4 && Object.keys(b).length === 4 && ['protocol', 'model', 'build', 'go'].every(key => typeof a[key] === 'string' && a[key] === b[key]);
  const hash = async bytes => [...new Uint8Array(await crypto.subtle.digest('SHA-256', bytes))].map(x => x.toString(16).padStart(2, '0')).join('');
  const workerScope = typeof WorkerGlobalScope !== 'undefined' && globalThis instanceof WorkerGlobalScope;
  if (workerScope) {
    let ready = false;
    let initializing = false;
    onmessage = async ({ data }) => {
      try {
        if (data.type === 'init') {
          if (initializing) throw new Error('Worker already initialized');
          initializing = true;
          const load = async (url, expected) => {
            const target = new URL(url, location.href);
            if (target.origin !== location.origin) throw new Error('assets must be same-origin');
            const response = await fetch(target, { cache: 'no-store', credentials: 'omit', redirect: 'error' });
            if (!response.ok) throw new Error(`asset HTTP ${response.status}`);
            const bytes = await response.arrayBuffer();
            if (await hash(bytes) !== expected) throw new Error('asset hash mismatch');
            return bytes;
          };
          const [module, runtime] = await Promise.all([load(data.wasmURL, data.wasmSHA256), load(data.runtimeURL, data.runtimeSHA256)]);
          const blob = URL.createObjectURL(new Blob([runtime], { type: 'text/javascript' }));
          try { importScripts(blob); } finally { URL.revokeObjectURL(blob); }
          const go = new Go();
          const { instance } = await WebAssembly.instantiate(module, go.importObject);
          go.run(instance).catch(error => postMessage({ type: 'fault', error: String(error) }));
          const prefix = data.prefix || 'fecimHysteresis';
          if (!['fecimHysteresis', 'fecimModule'].includes(prefix)) throw new Error('unknown Go bridge');
          globalThis.fecimWorkerPrefix = prefix;
          const identity = JSON.parse(globalThis[prefix + 'Identity']);
          if (!same(identity, data.identity) || identity.build === 'unqualified') throw new Error('Go identity mismatch');
          ready = true;
          postMessage({ type: 'ready', identity });
          return;
        }
        if (!ready || data.type !== 'dispatch' || typeof data.json !== 'string') throw new Error('Worker is not ready or request is malformed');
        // One Go-bounded batch per dispatch; yielding a task permits cancel
        // and progress delivery between units, never an unbounded JS run loop.
        await new Promise(resolve => setTimeout(resolve, 0));
        const json = globalThis[globalThis.fecimWorkerPrefix + 'Dispatch'](data.json);
        postMessage({ type: 'reply', json });
      } catch (error) { postMessage({ type: 'fault', error: String(error) }); }
    };
    return;
  }

  class HysteresisWorkerClient {
    constructor(options) {
      this.options = options;
      this.generation = 0;
      this.revision = 0;
      this.sequence = 0;
      this.state = 'loading';
      this.start();
    }
    start() {
      const generation = ++this.generation;
      this.ready = new Promise((resolve, reject) => {
        this.rejectReady = reject;
        this.startTimer = setTimeout(() => this.fail(new Error('Worker startup timeout')), this.options.timeoutMS || 30000);
        (async () => {
        const target = new URL(this.options.workerURL, location.href);
        if (target.origin !== location.origin) throw new Error('Worker must be same-origin');
        const response = await fetch(target, { cache: 'no-store', credentials: 'omit', redirect: 'error' });
        if (!response.ok) throw new Error(`Worker HTTP ${response.status}`);
        const bytes = await response.arrayBuffer();
        if (await hash(bytes) !== this.options.workerSHA256) throw new Error('Worker hash mismatch');
        if (generation !== this.generation) return;
        this.workerBlob = URL.createObjectURL(new Blob([bytes], { type: 'text/javascript' }));
        const worker = this.worker = new Worker(this.workerBlob);
        worker.onerror = event => { if (generation === this.generation) this.fail(new Error(event.message)); };
        worker.onmessage = ({ data }) => {
          if (generation !== this.generation) return; // discarded Workers cannot publish
          if (data.type === 'fault') { this.fail(new Error(data.error)); return; }
          if (data.type === 'ready') {
            if (!same(data.identity, this.options.identity)) { this.fail(new Error('ready identity mismatch')); return; }
            clearTimeout(this.startTimer);
            URL.revokeObjectURL(this.workerBlob); this.workerBlob = null;
            this.state = 'ready'; resolve(); return;
          }
          if (data.type !== 'reply' || !this.pending) { this.fail(new Error('unsolicited Worker response')); return; }
          try {
            const reply = JSON.parse(data.json);
            const { request } = this.pending;
            if (!same(reply.identity, request.identity) || reply.session !== request.session || reply.request !== request.request || reply.operation !== request.operation || reply.revision !== request.revision + (reply.committed ? 1 : 0) || (!reply.error && !reply.committed)) throw new Error('stale or mismatched Worker response');
            const pending = this.pending;
            this.pending = null; clearTimeout(pending.timer);
            this.revision = reply.revision;
            this.sequence = Math.max(this.sequence, reply.request);
            this.state = reply.error ? 'failed' : 'ready';
            pending.resolve(data.json); // original Go JSON bytes
          } catch (error) { this.fail(error); }
        };
        worker.postMessage({ ...this.options, type: 'init', wasmURL: new URL(this.options.wasmURL, location.href).href, runtimeURL: new URL(this.options.runtimeURL, location.href).href });
        })().catch(error => { if (generation === this.generation) this.fail(error); });
      });
    }
    async send(json) {
      const generation = this.generation;
      await this.ready;
      if (generation !== this.generation) throw new Error('Worker request belongs to a retired generation');
      if (this.pending) throw new Error('one active Worker request allowed');
      const request = JSON.parse(json);
      if (request.revision !== this.revision) throw new Error('stale request revision');
      return new Promise((resolve, reject) => {
        const timer = setTimeout(() => this.fail(new Error('Worker request timeout')), this.options.timeoutMS || 30000);
        this.pending = { request, resolve, reject, timer };
        this.state = 'running';
        this.worker.postMessage({ type: 'dispatch', json });
      });
    }
    async request(operation, input = null) {
      return this.send(JSON.stringify({ identity: this.options.identity, session: this.options.session, request: ++this.sequence, revision: this.revision, operation, input }));
    }
    fail(error) { this.terminate('failed', error); }
    terminate(state = 'cancelled', error = new Error('Worker terminated; ephemeral session discarded')) {
      ++this.generation;
      this.worker?.terminate(); clearTimeout(this.startTimer);
      if (this.workerBlob) { URL.revokeObjectURL(this.workerBlob); this.workerBlob = null; }
      this.rejectReady?.(error);
      if (this.pending) { clearTimeout(this.pending.timer); this.pending.reject(error); this.pending = null; }
      this.state = state;
    }
    recreate() {
      this.terminate();
      this.revision = 0; this.sequence = 0; this.state = 'loading';
      this.start();
      return this.ready;
    }
  }
  globalThis.HysteresisWorkerClient = HysteresisWorkerClient;
})();
