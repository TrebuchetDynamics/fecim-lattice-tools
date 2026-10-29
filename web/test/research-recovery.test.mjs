import {test} from 'node:test';
import assert from 'node:assert/strict';
import {mkdtempSync,readFileSync,writeFileSync,mkdirSync,rmSync,copyFileSync} from 'node:fs';
import {join} from 'node:path';
import {pathToFileURL} from 'node:url';
import ts from 'typescript';

// Exercise the public owner/adapter seam and Worker transport, not science.
const temp=mkdtempSync(join(process.env.TMPDIR,'research-recovery-'));
for(const name of ['api/owner','api/hosted','api/types','worker/ownership']) {
 const path=join(temp,name+'.mjs');mkdirSync(join(path,'..'),{recursive:true});
 const js=ts.transpileModule(readFileSync('src/'+name+'.ts','utf8'),{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.ESNext}}).outputText.replace(/(from\s+["'])(\.[^"']+)(["'])/g,(_,a,b,c)=>a+b+(b.endsWith('.mjs')?'':'.mjs')+c);
 writeFileSync(path,js);
}
copyFileSync('src/worker/hysteresis-worker.mjs',join(temp,'worker/hysteresis-worker.mjs'));
globalThis.__FECIM_PROFILE__='hosted';globalThis.__FECIM_UI__='ui';globalThis.__FECIM_API__='api';
const {ResearchOwner}=await import(pathToFileURL(join(temp,'api/owner.mjs')));
const {claimScience,releaseScience}=await import(pathToFileURL(join(temp,'worker/ownership.mjs')));
const {hostedResearch}=await import(pathToFileURL(join(temp,'api/hosted.mjs')));
test.after(()=>rmSync(temp,{recursive:true,force:true}));
const job={id:'real-adapter-job',state:'running'};
const config={id:'project',revision:'revision',fields:[]};
test('adoption and recovery claim tab ownership, refusing a busy HYS without cancelling the job',async()=>{
 const owner=new ResearchOwner();owner.connect=async()=>({active:async()=>[job]});
 const hys={};claimScience(hys);
 await assert.rejects(owner.accept(config),/Another scientific computation/);
 assert.equal(owner.config.id,'project');assert.equal(owner.job,undefined);
 releaseScience(hys);await owner.recover();
 assert.throws(()=>claimScience(hys),/Another scientific computation/);
 releaseScience(owner);clearTimeout(owner.poll);
});
test('hosted transport failure retains unknown job until explicit close, then creates a fresh client',async()=>{
 let count=0,fault=false,firstFetch=true,fetches=0;
 globalThis.fetch=async()=>{fetches++;if(firstFetch){firstFetch=false;throw new Error('Transient startup failure');}return {ok:true,json:async()=>({UI:'ui',API:'api',assets:{},modules:{research:{identity:{},wasm:'research.wasm',prefix:'fecimModule'}}})};};
 globalThis.HysteresisWorkerClient=class {
  ready=Promise.resolve();constructor(){count++;}
  async request(op){if(fault)throw new Error('Worker disconnected');return JSON.stringify({result:{configuration:config,job:op==='start'?job:{id:'none',state:'idle'}}});}
  terminate(){}
 };
 const owner=new ResearchOwner();const adapter=await owner.connect();
 await assert.rejects(adapter.example(),e=>e.transportFault===true);
 assert.equal(owner.config,undefined);
 await assert.rejects(adapter.example(),e=>e.transportFault===true);
 assert.equal(fetches,1);
 await owner.close(); // Public reset must work even before a configuration exists.
 await owner.accept(await adapter.example());assert.equal(fetches,2);
 await adapter.start();fault=true;
 await assert.rejects(adapter.state(),e=>e.transportFault===true);
 await adapter.close();fault=false;await adapter.example();assert.equal(count,2);
});
