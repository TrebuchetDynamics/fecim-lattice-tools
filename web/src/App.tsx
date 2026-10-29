import { lazy, Suspense, useEffect, useRef, useState, useSyncExternalStore, type ComponentType } from "react";
import { Home } from "./Home";
import { researchOwner } from "./api/owner";
import { manifest } from "./api/hosted";
import type { Schema } from "./api/types";
const Workbench=lazy(()=>import("./research/Workbench"));
// Fixed, build-time-only discovery. Neither URLs nor Go responses can name an
// arbitrary import. Each module owns its Screen entry, not this route owner.
const built=import.meta.glob<{default:ComponentType}>("./modules/*/Screen.tsx");
const screens=Object.fromEntries(Object.entries(built).map(([path,load])=>[path.split("/")[2],lazy(load)]));
function route() {return location.pathname.replace(/\/$/,"")||"/";}
const researchViews=["projects","configuration","experiment","results"] as const;
export function App() {
 useSyncExternalStore(researchOwner.subscribe,researchOwner.getSnapshot);
 const [navigation,setNavigation]=useState(0);
 const [path,setPath]=useState(route),[visited,setVisited]=useState<Set<string>>(new Set()),[modules,setModules]=useState<Schema<"ModuleDescriptor">[]>([]),[moduleError,setModuleError]=useState("");
 const main=useRef<HTMLElement>(null);
 const id=path==="/hysteresis"?"hysteresis":path.startsWith("/modules/")?path.slice(9):path==="/docs"?"docs":undefined;
 const view=researchViews.find(v=>path==="/"+v);
 const descriptor=modules.find(m=>m.id===id);
 const canMount=!!id && !!screens[id] && (id==="hysteresis"&&path==="/hysteresis" || !!descriptor?.available&&!!descriptor.capabilities?.some(c=>c.profile===__FECIM_PROFILE__&&c.available));
 useEffect(()=>{if(canMount&&id)setVisited(old=>new Set([...old,id]));},[id,canMount]);
 useEffect(()=>{
  const pop=()=>{setPath(route());setNavigation(n=>n+1);};window.addEventListener("popstate",pop);
  const click=(event:MouseEvent)=>{
   if(event.defaultPrevented||event.button!==0||event.metaKey||event.ctrlKey||event.shiftKey||event.altKey)return;
   const anchor=(event.target as Element).closest("a");if(!anchor||anchor.download||anchor.target)return;
   const url=new URL(anchor.href);if(url.origin!==location.origin||url.search||anchor.getAttribute("href")?.startsWith("#"))return;
   if(!/^\/(?:$|hysteresis$|projects$|configuration$|experiment$|results$|docs$|modules\/[a-z][a-z0-9-]*$)/.test(url.pathname))return;
   event.preventDefault();if(url.href!==location.href){history.pushState(null,"",url.pathname+url.hash);setPath(route());setNavigation(n=>n+1);}
  };
  document.addEventListener("click",click);return()=>{window.removeEventListener("popstate",pop);document.removeEventListener("click",click);};
 },[]);
 useEffect(()=>{
  // Focus once, after the matching lazy route has mounted; later data updates
  // cannot steal focus. Hidden retained module headings are never candidates.
  let done=false;
  const focus=()=>{if(done)return;const fragment=location.hash ? document.getElementById(decodeURIComponent(location.hash.slice(1))) : null;const h=fragment || Array.from(main.current?.querySelectorAll<HTMLElement>("h1")||[]).find(h=>!h.closest("[hidden]")&&h.getClientRects().length>0);if(h){h.tabIndex=-1;h.focus();done=true;}};
  const observer=new MutationObserver(focus);if(main.current)observer.observe(main.current,{childList:true,subtree:true});focus();
  return()=>observer.disconnect();
 },[path,canMount,navigation]);
 useEffect(()=>{
  if(!id||path==="/hysteresis")return;
  let current=true;
  void (__FECIM_PROFILE__==="local"?researchOwner.connect().then(a=>a.modules()):manifest().then(m=>m.Modules)).then(value=>{if(current){setModules(value);setModuleError("");}}).catch(error=>{if(current)setModuleError((error as Error).message);});
  return()=>{current=false;};
 },[id,path]);
 const title=descriptor?.title || (id==="docs"?"Documentation":id? id[0].toUpperCase()+id.slice(1):"Page not found");
 return <><a className="skip-link" href="#main">Skip to content</a><header className="nav"><a className="brand" href="/"><b className="brand-mark" aria-hidden="true">Fe</b> FeCIM Lattice Tools</a><nav className="nav__links" aria-label="Primary navigation"><a href="/hysteresis">Hysteresis lab</a><a href="/projects">Projects</a><a href="/docs">Documentation</a><a href="/#modules">Modules</a><a href="/#science">Science</a><a href="/#start">Local setup</a></nav></header>
 <p className="sr-only" role="status" aria-live="polite" aria-atomic="true">{researchOwner.announcement}</p>
 <main id="main" ref={main}><Suspense fallback={<p>Loading interface…</p>}>
 {path==="/"&&<Home/>}
 {view&&<Workbench view={view}/>}
 {[...new Set([...visited,...(canMount&&id?[id]:[])])].map(key=>{const Screen=screens[key];return <div key={key} hidden={key!==id||!canMount}><Screen/></div>;})}
 {id&&!canMount&&<section className="section"><h1>{title}</h1><p>{descriptor?.reason || "This browser module is not available in this build. No scientific port or screen is advertised as complete."}</p>{moduleError&&<p>{moduleError}</p>}<p>Choose an available module from the home page. Documentation is available as a read-only browser candidate; reading it never stops active science.</p><a href="/projects">Projects</a></section>}
 {!view&&!id&&path!=="/"&&<section className="section"><h1>Page not found</h1><a href="/">Home</a></section>}
 </Suspense></main><footer>FeCIM / Simulation only · MIT · Go scientific core</footer></>;
}
