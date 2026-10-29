// Common module qualification: real production Go envelope, Node runtime and
// actual Firefox production Worker. Scientific serialization is never parsed
// and reserialized by JavaScript; the original Go results_json bytes are saved.
import { readFileSync, writeFileSync, existsSync } from "node:fs";
import { resolve, join } from "node:path";
import { pathToFileURL } from "node:url";
import { runInThisContext } from "node:vm";
import { createHash, webcrypto } from "node:crypto";
import { createServer } from "node:http";
const [wasmPath,shimPath,inputPath,dir]=process.argv.slice(2);
const wasm=readFileSync(wasmPath),shim=readFileSync(shimPath),driver=readFileSync("web/src/worker/hysteresis-worker.mjs");
const fixture=JSON.parse(readFileSync(inputPath,"utf8"));
const hash=bytes=>createHash("sha256").update(bytes).digest("hex");
globalThis.crypto ||= webcrypto;
runInThisContext(shim.toString(),{filename:shimPath});
const go=new Go();const {instance}=await WebAssembly.instantiate(wasm,go.importObject);
go.run(instance).catch(error=>{console.error(error);process.exitCode=1;});
if(JSON.stringify(JSON.parse(fecimModuleIdentity))!==JSON.stringify(fixture.identity))throw new Error("Node module identity mismatch");
let request={...fixture};
async function dispatch(operation){const reply=JSON.parse(fecimModuleDispatch(JSON.stringify({...request,operation})));if(reply.error)throw new Error(reply.error);request={...request,request:request.request+1,revision:reply.revision};return reply.result;}
await dispatch("example");await dispatch("start");
let result;
for(let i=0;i<5000;i++){result=await dispatch("state");if(result.job.state!=="running")break;await new Promise(r=>setTimeout(r,1));}
if(result.job.state!=="finished"||!result.results_json)throw new Error("Node module did not finish successful science");
writeFileSync(join(dir,"node-results.json"),result.results_json);
const playwright=process.env.FECIM_PLAYWRIGHT_MODULE||resolve("web/node_modules/playwright/index.mjs");
const {firefox}=await import(pathToFileURL(playwright));
const routes=new Map([["/",["text/html",Buffer.from("<!doctype html><title>Production module parity</title>")]],["/module.wasm",["application/wasm",wasm]],["/wasm_exec.js",["text/javascript",shim]],["/worker.mjs",["text/javascript",driver]]]);
const server=createServer((req,res)=>{const found=routes.get(req.url);if(!found){res.writeHead(404).end();return;}res.writeHead(200,{"Content-Type":found[0],"Cache-Control":"no-store"});res.end(found[1]);});
let browser;
try{
 const base=Number(process.env.AUTOWORK_PORT_BASE||0);
 await new Promise((resolve,reject)=>{server.once("error",reject);server.listen(base?base+1:0,"127.0.0.1",resolve);});
 const cached=process.env.AUTOWORK_CACHE?join(process.env.AUTOWORK_CACHE,"playwright-1.60.0/firefox-1522/firefox/firefox"):"";
 const executablePath=process.env.FECIM_FIREFOX_EXECUTABLE||(existsSync(cached)?cached:undefined);
 browser=await firefox.launch({headless:true,...(executablePath?{executablePath}:{})});
 const page=await browser.newPage();await page.goto(`http://127.0.0.1:${server.address().port}`);await page.addScriptTag({url:"/worker.mjs"});
 const output=await page.evaluate(async ({fixture,hashes})=>{
  const client=new HysteresisWorkerClient({identity:fixture.identity,session:fixture.session,prefix:"fecimModule",workerURL:"/worker.mjs",workerSHA256:hashes.driver,wasmURL:"/module.wasm",wasmSHA256:hashes.wasm,runtimeURL:"/wasm_exec.js",runtimeSHA256:hashes.shim});
  const call=async operation=>{const reply=JSON.parse(await client.request(operation,fixture.input));if(reply.error)throw new Error(reply.error);return reply.result;};
  try{
   await client.ready;await call("example");await call("start");
   let result;for(let i=0;i<5000;i++){result=await call("state");if(result.job.state!=="running")break;await new Promise(r=>setTimeout(r,1));}
   if(result.job.state!=="finished"||!result.results_json)throw new Error("Firefox module did not finish science");
   const canonical=result.results_json;
   await call("start");const requested=await call("cancel");
   if(!requested.job.cancellation_requested)throw new Error("Go did not acknowledge cancellation");
   for(let i=0;i<5000;i++){result=await call("state");if(!["running","cancelling"].includes(result.job.state))break;await new Promise(r=>setTimeout(r,1));}
   if(result.job.state!=="cancelled"||result.job.progress.cancelled<1)throw new Error("Worker cancel did not interrupt Go point evaluation");
   let rejected=false;try{await client.send(JSON.stringify(fixture));}catch{rejected=true;}if(!rejected)throw new Error("stale request admitted");
   return {canonical,state:result.job.state,cancelled:result.job.progress.cancelled};
  }finally{client.terminate();}
 },{fixture,hashes:{driver:hash(driver),wasm:hash(wasm),shim:hash(shim)}});
 writeFileSync(join(dir,"firefox-results.json"),output.canonical);
 console.log(JSON.stringify({runtimes:["native","node","firefox"],productionEnvelope:true,originalBytes:true,cancellation:output.state,cancelled:output.cancelled,browser:browser.version()}));
}finally{await browser?.close();await new Promise(r=>server.close(r));}
process.exit(0);
