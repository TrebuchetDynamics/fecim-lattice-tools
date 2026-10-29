import { useEffect, useRef } from "react";
import { RequestError, type Field, type Schema } from "../api/types";
export function Value({ value }: { value:Schema<"Value"> }) {
 return <><span aria-hidden="true">{value.Display}</span><span className="sr-only">{value.Speakable || value.Reason || "Not available"}</span>{value.Availability!=="available" && value.Reason && <small> — {value.Reason}</small>}</>;
}
export function Metric({ cell }: { cell:Schema<"MetricValue"> }) {
 return <><Value value={cell.presentation}/><small> Evidence: {cell.metric?.evidence || "Not available"}</small>{cell.metric?.assumptions?.length ? <details><summary>Assumptions</summary><ul>{cell.metric.assumptions.map((a,i)=><li key={i}>{a}</li>)}</ul></details>:null}</>;
}
export const fieldID = (field:Pick<Field,"document"|"path">) => "field-"+encodeURIComponent(field.document+field.path);
export function FieldInput({ field, input, readOnly, change }: { field:Field;input:string;readOnly:boolean;change:(value:string)=>void }) {
 const id=fieldID(field);
 return <div className="field"><label htmlFor={id}>{field.label}</label><input id={id} value={input} readOnly={readOnly} onChange={e=>change(e.target.value)} aria-describedby={id+"-value"}/><small id={id+"-value"}>Go value: <Value value={field.presentation}/>{field.unit && <> · Unit: {field.unit}</>}</small></div>;
}
export function ErrorSummary({ error, clear, reapply, reload, document }: {error:Error;clear:()=>void;reapply?:(revision:string)=>void;reload?:()=>void;document?:string}) {
 const ref=useRef<HTMLDivElement>(null);
 useEffect(()=>{ref.current?.focus();},[error]);
 const problem=error instanceof RequestError ? error.problem : undefined;
 return <div role="alert" tabIndex={-1} ref={ref} className="error error-summary"><h2>{problem?.status===412?"Configuration conflict":"Action failed"}</h2><p>{error.message}</p>{problem?.issues?.length ? <><p>{problem.issues.length} {problem.issues.length===1?"error":"errors"}</p><ul>{problem.issues.map((issue,i)=><li key={i}><a href={"#"+fieldID(issue)} onClick={e=>{e.preventDefault();window.document.getElementById(fieldID(issue))?.focus();}}>{issue.label}</a>: {issue.reason}</li>)}</ul></>:null}{problem?.fields?.length ? <ul>{problem.fields.map((f,i)=><li key={i}><a href={"#"+fieldID({document:document||"design.yaml",path:f.path})} onClick={e=>{e.preventDefault();window.document.getElementById(fieldID({document:document||"design.yaml",path:f.path}))?.focus();}}>{f.label}</a>: yours <Value value={f.yours_presentation}/>; on disk {f.on_disk_available?<Value value={f.on_disk_presentation}/>:f.reason || "Not available"}</li>)}</ul>:null}
 {problem?.status===412 && <><p>Your inputs are retained. Review these current disk values before reapplying; another edit may conflict again.</p><button onClick={reload}>Reload and discard my edits</button>{problem.revision && <button onClick={()=>reapply?.(problem.revision!)}>Reapply reviewed edits</button>}</>}
 <button onClick={clear}>Dismiss error</button></div>;
}
