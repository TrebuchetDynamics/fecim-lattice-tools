import { useEffect, useRef, useState, useSyncExternalStore } from "react";
import { researchOwner as owner } from "../api/owner";
import { initialQuery, type Query, type Schema } from "../api/types";
import { ErrorSummary, FieldInput, Metric } from "./Shared";

export type ResearchView = "projects"|"configuration"|"experiment"|"results";
const local = __FECIM_PROFILE__ === "local";
export default function Workbench({ view }: {view:ResearchView}) {
 useSyncExternalStore(owner.subscribe,owner.getSnapshot);
 const [error,setError]=useState<Error>();
 const [document,setDocument]=useState<string>();
 useEffect(()=>{setError(undefined);},[view]);
 const generation=useRef(0),currentView=useRef(view),alive=useRef(true);
 if(currentView.current!==view){currentView.current=view;generation.current++;}
 useEffect(()=>{alive.current=true;return()=>{alive.current=false;generation.current++;};},[]);
 const act=async<T,>(fn:()=>Promise<T>)=>{const request=++generation.current,project=owner.projectToken;try{setError(undefined);const value=await fn();return alive.current&&request===generation.current?value:undefined;}catch(e){if(alive.current&&request===generation.current&&(project===owner.projectToken||view==="projects"))setError(e instanceof Error?e:new Error(String(e)));return undefined;}};
 return <section className="section research"><p className="eyebrow">Research workbench · {__FECIM_PROFILE__}</p><h1>{view[0].toUpperCase()+view.slice(1)}</h1>
 <aside className="notice" aria-label="Research capabilities and privacy"><p>Simulation only — not validated device measurements. {local?"Project operations use this authenticated same-origin server and launch-granted roots. Writes are explicit; browser navigation does not cancel experiments.":"Computations run in your browser using the Go model. Project and result data is not uploaded. On-disk projects and edits, large sweeps, run reuse, stored-run verification, EDA and file writes require the local application: run fecim-lattice-tools serve --root DIRECTORY. This hosted build does not contact localhost."}</p>{!local&&<p>Refreshing discards in-memory studies and selection. No scientific downloads or persistence. Ordinary static asset requests still expose HTTP metadata to the host.</p>}</aside>
 <nav className="actions" aria-label="Research views">{(["projects","configuration","experiment","results"] as const).map(v=><a className="button" key={v} href={"/"+v} aria-current={view===v?"page":undefined}>{v[0].toUpperCase()+v.slice(1)}</a>)}</nav>
 {error&&<ErrorSummary error={error} clear={()=>setError(undefined)} document={document} reload={()=>void act(()=>owner.reload())} reapply={revision=>void act(()=>owner.save(document!,revision))}/>}
 <p data-testid="research-status" role="status" aria-live="polite">{owner.status}</p>
 {view==="projects"?<Projects act={act}/>:!owner.config?<p>Open or create a study in <a href="/projects">Projects</a>. {local?"After a server restart, reopen the project to inspect integrity-checked stored runs.":"Hosted state is ephemeral."}</p>:<>
 <p>Opened study: {owner.config.name||owner.config.id}. Evidence: {owner.config.evidence}</p>
 {view==="configuration"&&<Configuration act={act} setDocument={setDocument}/>}
 {view==="experiment"&&<Experiment act={act}/>}
 {view==="results"&&<Results act={act}/>}
 </>}
 </section>;
}
type Act=<T>(fn:()=>Promise<T>)=>Promise<T|undefined>;
function Projects({act}:{act:Act}) {
 const [roots,setRoots]=useState<{id:string;name:string}[]>([]),[root,setRoot]=useState("");
 const [projects,setProjects]=useState<Schema<"Projects">>();
 const [name,setName]=useState(""),[path,setPath]=useState(""),[busy,setBusy]=useState(false);
 const alive=useRef(true);
 useEffect(()=>{alive.current=true;void act(async()=>{const a=await owner.connect();const [r,p]=await Promise.all([a.roots(),a.projects()]);if(alive.current){setRoots(r);setRoot(r[0]?.id||"");setProjects(p);}});return()=>{alive.current=false;};},[]);
 const open=(fn:()=>Promise<Schema<"Configuration">>)=>void act(async()=>{setBusy(true);try{await owner.accept(await fn());}finally{setBusy(false);}});
 return <>
 {local?<><label>Granted project root<select aria-label="Granted project root" value={root} onChange={e=>setRoot(e.target.value)}>{roots.map(r=><option key={r.id} value={r.id}>{r.name}</option>)}</select></label>
 <form onSubmit={e=>{e.preventDefault();open(async()=>(await owner.connect()).create(root,name));}}><label>Project name<input value={name} onChange={e=>setName(e.target.value)} required/></label><button disabled={busy||!root}>Create project</button></form>
 <form onSubmit={e=>{e.preventDefault();open(async()=>(await owner.connect()).open(root,path));}}><label>Relative project path<input value={path} onChange={e=>setPath(e.target.value)} required/></label><button disabled={busy||!root}>Open project</button></form>
 <h2>Discovered projects</h2><ul>{projects?.projects?.map(p=><li key={p.root+p.path}>{p.path} <button onClick={()=>open(async()=>(await owner.connect()).open(p.root,p.path))}>Open {p.path}</button></li>)}</ul>
 <button disabled={!projects||projects.page<=1} onClick={()=>void act(async()=>setProjects(await(await owner.connect()).projects(projects!.page-1)))}>Previous projects</button><button disabled={!projects?.has_more} onClick={()=>void act(async()=>setProjects(await(await owner.connect()).projects(projects!.page+1)))}>Next projects</button>
 </>:<><button disabled={busy} onClick={()=>open(async()=>(await owner.connect()).example())}>Open small example</button><p>The admitted example has eight points and 16/32-row arrays. The original 64-row example remains native-only. Go enforces at most 256 points and 32 by 32 before expansion; inputs are never silently truncated.</p>
 <label>Import study byte bundle<input type="file" multiple accept=".yaml,.yml,.json" onChange={e=>{const files=Array.from(e.target.files||[]);void act(async()=>{
  if(!files.length)return;
  setBusy(true);try{
   if(files.reduce((n,f)=>n+f.size,0)>8<<20)throw new Error("Study bytes exceed 8 MiB");
   let docs:Record<string,string>={};
   if(files.length===1&&files[0].name.endsWith(".json")){const bundle=JSON.parse(await files[0].text());docs=bundle.documents; if(!docs||typeof docs!=="object")throw new Error("Expected a documents map of file names to base64 bytes");}
   else for(const f of files){if(docs[f.name])throw new Error("Duplicate document name");const bytes=new Uint8Array(await f.arrayBuffer());let binary="";for(const byte of bytes)binary+=String.fromCharCode(byte);docs[f.name]=btoa(binary);}
   await owner.accept(await(await owner.connect()).import(docs));
  }finally{setBusy(false);}
 });}}/></label><p>Select project.yaml, design.yaml, sweep.yaml and any named inputs, or a JSON byte bundle with a documents map of base64 bytes. Imported provenance is unverified.</p></>}
 {!local&&!owner.config&&<button disabled={busy} onClick={()=>void act(()=>owner.close())}>Reset research Worker</button>}
 {owner.expired&&!local&&<p>Worker execution is unknown. Previous results are retained until you explicitly discard this study.</p>}
 {owner.config&&<><h2>Current project</h2><p>{owner.config.name||owner.config.id}</p><button disabled={busy} onClick={()=>void act(()=>owner.close())}>{owner.expired&&!local?"Discard interrupted study":"Close project"}</button></>}
 </>;
}
function Configuration({act,setDocument}:{act:Act;setDocument:(doc:string)=>void}) {
 const [validation,setValidation]=useState("");
 const [busy,setBusy]=useState(false);
 return <><button onClick={()=>void act(async()=>{const v=await(await owner.connect()).validate(owner.config!.id);setValidation(v.valid?`Configuration valid: ${v.points} points`:"Configuration invalid");})}>Validate configuration</button><p data-testid="validation" role="status">{validation}</p>
 {!local&&<p>Read-only imported configuration. Editing on-disk projects requires the local application.</p>}
 {["project.yaml","design.yaml","sweep.yaml"].map(document=><form key={document} onSubmit={e=>{e.preventDefault();setDocument(document);void act(async()=>{setBusy(true);try{await owner.save(document);}finally{setBusy(false);}});}}><fieldset><legend>{document}</legend>{owner.config!.fields?.filter(f=>f.document===document).map(f=><FieldInput key={f.path} field={f} readOnly={!local} input={owner.drafts[owner.fieldKey(f)]??f.value} change={v=>owner.draft(f,v)}/>)}{local&&<button disabled={busy}>Save {document}</button>}</fieldset></form>)}
 {local&&<button onClick={()=>void act(()=>owner.reload())}>Reload configuration and discard edits</button>}
 </>;
}
function Experiment({act}:{act:Act}) {
 const [workers,setWorkers]=useState("1");
 const [starting,setStarting]=useState(false);
 const active=!!owner.job&&["running","cancelling","queued"].includes(owner.job.state)&&!owner.expired;
 return <><form onSubmit={e=>{e.preventDefault();void act(async()=>{setStarting(true);try{await owner.start(Number(workers));}finally{setStarting(false);}});}}>{local&&<label>Workers<select aria-label="Workers" value={workers} onChange={e=>setWorkers(e.target.value)}>{["1","2","3","4"].map(n=><option key={n}>{n}</option>)}</select></label>}<button disabled={active||starting}>Run study</button></form>
 <button disabled={!active||owner.job?.cancellation_requested} onClick={()=>void act(()=>owner.cancel())}>Cancel experiment</button>
 {local&&<><button onClick={()=>void act(()=>owner.recover())}>Discover active executions</button><button disabled={!owner.job||owner.expired} onClick={()=>void act(()=>owner.refreshJob())}>Reconnect authoritative state</button><p>Polling is bounded to two state requests per second. On interruption it stops until explicit reconnect. Disconnect and navigation never mean cancellation; committed records remain on disk and repeated studies reuse eligible IDs.</p></>}
 <p data-testid="experiment-state">{owner.job?.message||owner.job?.state||"No execution started"}</p>
 <p>{owner.announcement}</p>
 {owner.job&&<dl><dt>State</dt><dd>{owner.expired?"Unknown — execution no longer known":owner.job.state}</dd><dt>Planned</dt><dd>{owner.job.progress.total}</dd><dt>Completed (including reused and failed)</dt><dd>{owner.job.progress.completed}</dd><dt>Reused</dt><dd>{owner.job.progress.reused}</dd><dt>Failed</dt><dd>{owner.job.progress.failed}</dd><dt>Cancelled, not evaluated</dt><dd>{owner.job.progress.cancelled}</dd><dt>Results state</dt><dd>{owner.job.results_state}</dd></dl>}
 </>;
}
function Results({act}:{act:Act}) {
 const [draft,setDraft]=useState<Query>(owner.query),[comparison,setComparison]=useState(false),[detail,setDetail]=useState("");
 const [action,setAction]=useState(""),[reports,setReports]=useState<string[]>([]),[artifacts,setArtifacts]=useState<Schema<"Artifacts">>();
 const [selectionStatus,setSelectionStatus]=useState("");
 const caption=useRef<HTMLTableCaptionElement>(null),heading=useRef<HTMLHeadingElement>(null),alive=useRef(true);
 useEffect(()=>{alive.current=true;void act(()=>owner.results());return()=>{alive.current=false;};},[]);
 const load=(q:Query)=>void act(async()=>{if(await owner.results(q)&&alive.current){setComparison(false);caption.current?.focus();}});
 useEffect(()=>{if(comparison)heading.current?.focus();},[comparison]);
 const page=owner.analysis?.results;
 return <><form onSubmit={e=>{e.preventDefault();load({...draft,page:1,revision:owner.query.revision});}}><label>Filter runs<select aria-label="Filter runs" value={draft.filter} onChange={e=>setDraft({...draft,filter:e.target.value})}>{["all","success","failed","feasible","pareto"].map(f=><option key={f}>{f}</option>)}</select></label><label>Sort runs<select aria-label="Sort runs" value={draft.sort} onChange={e=>setDraft({...draft,sort:e.target.value})}>{["run_id",...(page?.metric_names||[])].map(n=><option key={n}>{n}</option>)}</select></label><label><input type="checkbox" checked={draft.descending} onChange={e=>setDraft({...draft,descending:e.target.checked})}/>Descending</label><label>Page size<select aria-label="Page size" value={draft.page_size} onChange={e=>setDraft({...draft,page_size:Number(e.target.value)})}>{[2,32,100].map(n=><option key={n}>{n}</option>)}</select></label><button>Apply query</button></form>
 {owner.analysis?.plot&&<ResultPlot plot={owner.analysis.plot}/>}
 {page&&<>{page.reset&&<p role="status">{page.reset_reason}</p>}<div className="table-scroll"><table id="runs"><caption tabIndex={-1} ref={caption}>{page.caption}</caption><thead><tr><th scope="col">Select</th><th scope="col">Run ID / inspect</th><th scope="col">Status</th><th scope="col">Feasible</th><th scope="col">Pareto</th>{page.metric_names?.map(n=><th key={n} scope="col">{n}</th>)}</tr></thead><tbody>{page.rows?.map(row=><tr key={row.run_id}><td><input type="checkbox" aria-label={"Select run "+row.run_id} checked={owner.selected.includes(row.run_id)} onChange={e=>{try{owner.select(row.run_id,e.target.checked);setSelectionStatus("");}catch(error){setSelectionStatus((error as Error).message);}}}/></td><th scope="row"><button className="run-id" onClick={()=>void act(async()=>{setDetail(await(await owner.connect()).inspect(owner.config!.id,row.run_id));})}>{row.run_id}</button></th><td>{row.status}{row.unusable_reason&&<> — {row.unusable_reason}</>}</td><td>{row.feasible?"Yes":"No"}</td><td>{row.pareto?"Yes":"No"}</td>{row.metrics?.map(cell=><td key={cell.name}><Metric cell={cell}/></td>)}</tr>)}</tbody></table></div><button disabled={page.page<=1} onClick={()=>load({...owner.query,page:page.page-1})}>Previous runs</button><button disabled={page.page>=page.page_count} onClick={()=>load({...owner.query,page:page.page+1})}>Next runs</button></>}
 {owner.job&&["running","cancelling","queued"].includes(owner.job.state)&&<p>Previous collected result snapshot. The active execution has not settled; prior immutable result revisions are retained.</p>}
 <section aria-label="Selected runs"><h2>Selected runs</h2><p>Select two to four distinct runs. Selection survives sorting, filters and pages; a fifth selection is refused.</p><p data-testid="selection-status" role="status">{selectionStatus}</p><ul id="selected-runs">{owner.selected.map(id=><li key={id}><code>{id}</code> <button onClick={()=>owner.select(id,false)}>Remove {id}</button></li>)}</ul><button onClick={()=>void act(async()=>{if(await owner.results(owner.query,owner.selected)&&alive.current){setComparison(true);}})}>Compare metrics</button></section>
 {comparison&&owner.analysis?.comparison&&<section><h2 tabIndex={-1} ref={heading}>Compare metrics</h2><div className="table-scroll"><table id="comparison"><caption>Stored metrics by immutable run identity. No models rerun.</caption><thead><tr><th scope="col">Metric and unit</th>{owner.analysis.comparison.run_ids?.map(id=><th className="run-id" key={id} scope="col">{id}</th>)}</tr></thead><tbody>{owner.analysis.comparison.metrics?.map(m=><tr key={m.name}><th scope="row">{m.name} {m.values?.[0]?.metric?.unit}{m.reason&&<> — {m.reason}</>}</th>{m.values?.map((c,i)=><td key={i}><Metric cell={c}/></td>)}</tr>)}</tbody></table></div></section>}
 {detail&&<section><h2>Run details and provenance</h2><pre>{detail}</pre><button onClick={()=>setDetail("")}>Close details</button></section>}
 <section><h2>Explicit local operations</h2>{!local&&<p>Stored-run verification, report generation, export and downloads require the local application. No localhost request is made.</p>}
 <button onClick={()=>void act(async()=>{await(await owner.connect()).verify(owner.config!.id);setAction("Verified stored runs using Go integrity review; not device validation");})}>Verify stored runs</button><button onClick={()=>void act(async()=>{const files=await(await owner.connect()).reports(owner.config!.id);setReports(files.files||[]);setAction("Reports generated from stored results; no simulations rerun");})}>Generate reports</button><button onClick={()=>void act(async()=>{if(owner.selected.length!==1)throw new Error("Select exactly one eligible run for export");setArtifacts(await(await owner.connect()).export(owner.config!.id,owner.selected[0]));setAction("Export generated from the verified selected run; no simulations rerun");})}>Export selected run</button><p data-testid="result-action" role="status">{action}</p>
 {local&&<><ul>{reports.map(file=><li key={file}><a href={`/api/v1/projects/${owner.config!.id}/reports/${encodeURIComponent(file)}`} download>Download {file}</a></li>)}</ul><ul>{artifacts?.files?.map(file=><li key={file}><a href={`/api/v1/projects/${owner.config!.id}/exports/${artifacts.run_id}/${encodeURIComponent(file)}`} download>Download {file}</a></li>)}</ul></>}
 </section></>;
}

function ResultPlot({plot}:{plot:Schema<"ResultsPlot">}) {
 return <figure id="results-plot" aria-label="Result plot"><figcaption>{plot.caption}</figcaption>
 {plot.availability==="available"?<svg role="img" aria-label={plot.caption} viewBox={`0 0 ${plot.width} ${plot.height}`}>
 {plot.points?.map(p=><circle key={p.run_id} cx={p.x} cy={p.y} r={owner.selected.includes(p.run_id)?7:4} fill={owner.selected.includes(p.run_id)?"#155eef":"currentColor"} data-selected={owner.selected.includes(p.run_id)?"true":"false"}><title>{p.run_id}</title></circle>)}
 </svg>:<p>{plot.reason}</p>}<p>{plot.x_axis.label} / {plot.y_axis.label}. {plot.scope}</p></figure>;
}
