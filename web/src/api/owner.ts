import { hostedResearch } from "./hosted";
import { claimScience, releaseScience } from "../worker/ownership";
import { initialQuery, type Analysis, type Configuration, type Job, type Query, type ResearchAdapter, type Schema } from "./types";

export class ResearchOwner {
 adapter?: ResearchAdapter;
 private connecting?: Promise<ResearchAdapter>;
 private listeners = new Set<() => void>();
 private version = 0;
 private poll?: ReturnType<typeof setTimeout>;
 private queryToken = 0;
 projectToken = 0;
 private lastAnnouncement = 0;
 config?: Configuration;
 drafts: Record<string,string> = {};
 job?: Job;
 analysis?: Analysis;
 selected: string[] = [];
 query: Query = {...initialQuery};
 status = "";
 announcement = "";
 expired = false;
 getSnapshot = () => this.version;
 subscribe = (fn:()=>void) => { this.listeners.add(fn);return()=>{this.listeners.delete(fn);}; };
 changed() {this.version++;this.listeners.forEach(fn=>fn());}
 async connect() {
  return this.connecting ??= (__FECIM_PROFILE__ === "local"
   ? import("../local/research").then(m=>m.localResearch())
   : Promise.resolve(hostedResearch())).then(a => {this.adapter=a;return a;}).catch(error=>{this.connecting=undefined;throw error;});
 }
 async accept(config:Configuration) {
  if(this.job && ["running","cancelling","queued"].includes(this.job.state) && !this.expired)throw new Error("Wait for the current experiment to settle before changing projects");
  this.projectToken++;this.queryToken++;this.config=config;this.drafts={};this.analysis=undefined;this.selected=[];this.job=undefined;this.expired=false;this.status="Project opened";this.changed();
  const token=this.projectToken;const a=await this.connect();
  const active=await a.active(config.id);
  if(token!==this.projectToken)return;
  if(active[0]) {claimScience(this);this.job=active[0];this.watch();this.changed();}
 }
 async close() {if(!this.config&&__FECIM_PROFILE__==="local")return;await(await this.connect()).close(this.config?.id||"");this.projectToken++;this.queryToken++;this.config=undefined;this.drafts={};this.analysis=undefined;this.selected=[];this.job=undefined;this.expired=false;this.status="Study closed; open an example or project to continue";clearTimeout(this.poll);releaseScience(this);this.changed();}
 async reload() {
  if(!this.config)return;
  const token=this.projectToken,id=this.config.id,drafts={...this.drafts};
  const config=await(await this.connect()).configuration(id);
  if(token!==this.projectToken)return;
  this.config=config;this.retireDrafts(drafts);this.changed();
 }
 private retireDrafts(submitted:Record<string,string>) {for(const [key,value] of Object.entries(submitted))if(this.drafts[key]===value)delete this.drafts[key];}
 fieldKey(field:Schema<"ConfigurationField">) {return field.document+field.path;}
 draft(field:Schema<"ConfigurationField">,value:string) {this.drafts={...this.drafts,[this.fieldKey(field)]:value};this.changed();}
 async save(document:string,revision=this.config?.revision) {
  if(!this.config || !revision)return;
  const token=this.projectToken,id=this.config.id;
  const submitted=Object.fromEntries((this.config.fields||[]).filter(f=>f.document===document && this.drafts[this.fieldKey(f)]!==undefined).map(f=>[this.fieldKey(f),this.drafts[this.fieldKey(f)]]));
  const changes=(this.config.fields||[]).filter(f=>f.document===document && this.drafts[this.fieldKey(f)]!==undefined).map(f=>({path:f.path,value:this.drafts[this.fieldKey(f)]}));
  const config=await(await this.connect()).save(id,revision,document,changes);
  if(token!==this.projectToken)return;
  this.config=config;this.retireDrafts(submitted);
  this.status="Saved "+document;this.changed();
 }
 async start(workers:number) {
  if(!this.config)return;
  claimScience(this);
  try {this.job=await(await this.connect()).start(this.config.id,this.config.revision,workers,crypto.randomUUID());this.expired=false;this.announcement=this.job.message||"Experiment started";this.lastAnnouncement=Date.now();this.changed();this.watch();}
  catch(error){releaseScience(this);throw error;}
 }
 async cancel() {if(!this.job)return;this.job=await(await this.connect()).cancel(this.job.id);this.announcement=this.job.message||"Cancellation requested";this.changed();this.watch();}
 private watch() {
  clearTimeout(this.poll);
  if(!this.job || !["running","cancelling","queued"].includes(this.job.state))return;
  const id=this.job.id;
  this.poll=setTimeout(()=>{void this.refreshJob(id);},500);
 }
 async refreshJob(id=this.job?.id) {
  if(!id)return;
  try {
   const job=await(await this.connect()).state(id);if(this.job?.id!==id)return;
   const wasTerminal=!this.job || !["running","cancelling","queued"].includes(this.job.state);
   this.job=job;
   const terminal=!["running","cancelling","queued"].includes(job.state);
   if(terminal){releaseScience(this);if(!wasTerminal)this.announcement=job.message||job.state;}
   else if(Date.now()-this.lastAnnouncement>=5000){this.announcement=job.message||job.state;this.lastAnnouncement=Date.now();}
   this.changed();this.watch();
  }catch(error){
   this.status="Connection interrupted. Retry authoritative state; disconnect does not cancel the experiment. "+(error as Error).message;
   if((error as {transportFault?:boolean}).transportFault){this.expired=true;releaseScience(this);this.status="Browser Worker interrupted. Execution outcome is unknown; previous results retained. Explicitly discard the ephemeral study before retrying.";}
   if((error as {problem?:{status:number}}).problem?.status===404 || (error as {problem?:{status:number}}).problem?.status===410){this.expired=true;releaseScience(this);this.status="Execution is no longer known after restart or eviction. Reopen the project and inspect integrity-checked stored runs; no terminal outcome is assumed.";}
   this.changed(); // bounded retry: only explicit Reconnect retries after a fault
  }
 }
 async recover() {if(!this.config)return;const token=this.projectToken;const jobs=await(await this.connect()).active(this.config.id);if(token!==this.projectToken)return;if(jobs[0]){claimScience(this);this.job=jobs[0];this.watch();}else this.status="No active execution found. Inspect stored runs to recover committed work.";this.changed();}
 async results(query:Query=this.query,compare?:string[]) {
  if(!this.config)throw new Error("Open a study first");
  const token=++this.queryToken,project=this.config.id;
  const analysis=await(await this.connect()).analysis(project,query,compare);
  if(token!==this.queryToken || project!==this.config?.id)return false;
  this.analysis=analysis;this.query={...query,page:analysis.results.page,revision:analysis.results.revision};this.changed();return true;
 }
 select(id:string,selected:boolean) {
  if(selected && !this.selected.includes(id) && this.selected.length>=4)throw new Error("Select 2 to 4 distinct runs. A fifth run was refused; the first four remain selected.");
  this.selected=selected?[...new Set([...this.selected,id])]:this.selected.filter(x=>x!==id);this.changed();
 }
}
export const researchOwner = new ResearchOwner();
