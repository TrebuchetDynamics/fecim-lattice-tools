import "../worker/hysteresis-worker.mjs";
import type { Identity } from "../hysteresis/types";
import { initialQuery, type ResearchAdapter, type Schema } from "./types";

type Manifest = { UI:string; API:string; assets:Record<string,string>; Modules:Schema<"ModuleDescriptor">[]; modules:Record<string,{wasm:string;identity:Identity;prefix:string}> };
let manifestPromise: Promise<Manifest> | undefined;
export function manifest() {
 return manifestPromise ??= fetch("/engine/manifest.json", { credentials:"omit", cache:"no-store", redirect:"error" }).then(async response => {
  if (!response.ok) throw new Error("Asset manifest unavailable");
  const value: Manifest = await response.json();
  if (value.UI !== __FECIM_UI__) throw new Error("UI/Worker asset fingerprint mismatch");
  if (value.API !== __FECIM_API__) throw new Error("API-contract fingerprint mismatch");
  return value;
 }).catch(error => { manifestPromise=undefined; throw error; });
}
export function hostedResearch(): ResearchAdapter {
 let client: InstanceType<typeof HysteresisWorkerClient> | undefined;
 let starting: Promise<void> | undefined;
 let fault: Error | undefined;
 let tail: Promise<unknown> = Promise.resolve();
 let last: Schema<"ResearchResult"> | undefined;
 const ready = () => starting ??= (async () => {
  const m = await manifest(), entry = m.modules.research;
  if (!entry) throw new Error("Research Worker unavailable in this build");
  client = new HysteresisWorkerClient({ identity: entry.identity, prefix:entry.prefix, session:crypto.randomUUID(), workerURL:"/engine/hysteresis-worker.mjs", workerSHA256:m.assets["hysteresis-worker.mjs"], wasmURL:"/engine/"+entry.wasm, wasmSHA256:m.assets[entry.wasm], runtimeURL:"/engine/wasm_exec.js", runtimeSHA256:m.assets["wasm_exec.js"] });
  await client.ready;
 })().catch(error => { starting=undefined;client?.terminate();client=undefined;throw error; });
 const call = (operation:string, input:Partial<Schema<"ResearchInput">> = {}) => {
  const promise = tail.then(async () => {
   if(fault)throw fault;
   let json:string;
   try {await ready();json=await client!.request(operation, { query:initialQuery, ...input });}
   catch(error){client?.terminate();client=undefined;starting=undefined;manifestPromise=undefined;fault=Object.assign(new Error("Browser Worker interrupted; execution outcome unknown. Discard the interrupted study to retry. "+(error as Error).message),{transportFault:true});throw fault;}
   const reply: Schema<"ResearchReply"> = JSON.parse(json);
   if (reply.error) throw new Error(reply.error);
   last=reply.result;return reply.result;
  });
  tail=promise.catch(() => {});return promise;
 };
 const unavailable = async (): Promise<never> => { throw new Error("This operation requires the local application: run fecim-lattice-tools serve --root DIRECTORY. No localhost request was made."); };
 return {
  roots:async()=>[], projects:async()=>({projects:[],page:1,page_size:32,has_more:false}),
  open:unavailable,create:unavailable,
  example:async()=> (await call("example")).configuration!,
  import:async documents => (await call("import",{documents})).configuration!,
  configuration:async()=> {if (!last?.configuration) throw new Error("Open a study first"); return last.configuration;},
  save:unavailable,
  validate:async()=> { const result=await call("validate");return { valid:true,points:result.points }; },
  close:async()=> {if (!fault && (last?.job.state==="running" || last?.job.state==="cancelling")) throw new Error("Cancel or wait before closing");client?.terminate();client=undefined;starting=undefined;last=undefined;fault=undefined;},
  start:async()=> (await call("start")).job,
  state:async()=> (await call("state")).job,
  cancel:async()=> (await call("cancel")).job,
  active:async()=> {if(fault)throw fault;return last && ["running","cancelling"].includes(last.job.state) ? [last.job] : [];},
  analysis:async (_id,query,compare)=> {const r=await call(compare?"compare":"query",{query,compare});if(!r.page || !r.plot)throw new Error("No collected results yet");return {results:r.page,plot:r.plot,comparison:r.comparison,counts:r.counts || {}};},
  inspect:async (_id,run_id)=> (await call("inspect",{run_id})).run_json || "Run unavailable",
  verify:unavailable,reports:unavailable,export:unavailable,
  modules:async()=> (await manifest()).Modules,
 };
}
