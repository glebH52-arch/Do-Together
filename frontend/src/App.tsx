import { useEffect, useRef, useState, type FormEvent } from "react";
import { Bell, Inbox, Settings, X } from "lucide-react";
import { api, refreshKey, request, setSession, subscribeSession, type Invite, type Notification, type Tokens, type User } from "./api";
import { Projects } from "./Projects";
import { ProjectDetail } from "./ProjectDetail";
import { type Locale, chromeCopy, localizedRequestError, notificationText, fmt, useCopy } from "./i18n";
import { BoardPreview } from "./BoardPreview";
import { Avatar } from "./Avatar";

type Mode = "login" | "register" | "verify";
type Theme = "light" | "dark" | "system";
const uuid = "[0-9a-fA-F-]{36}";

export function App() {
  const [mode,setMode]=useState<Mode>("login"), [email,setEmail]=useState(""), [username,setUsername]=useState(""), [password,setPassword]=useState(""), [code,setCode]=useState("");
  const [user,setUser]=useState<User|null>(null), [tokens,setTokens]=useState<Tokens|null>(null), [loading,setLoading]=useState(true), [busy,setBusy]=useState(false), [error,setError]=useState<unknown>(null), [notice,setNotice]=useState(false);
  const [path,setPath]=useState(location.pathname), [popover,setPopover]=useState<"invites"|"notifications"|"settings"|null>(null);
  const [invites,setInvites]=useState<Invite[]>([]), [notifications,setNotifications]=useState<Notification[]>([]), [inviteCount,setInviteCount]=useState(0), [notificationCount,setNotificationCount]=useState(0), [projectsRevision,setProjectsRevision]=useState(0);
  const [locale,setLocale]=useState<Locale>(()=>localStorage.getItem("dt-locale")==="en"?"en":"ru"), [theme,setTheme]=useState<Theme>(()=>(localStorage.getItem("dt-theme") as Theme)||"system");
  const popRef=useRef<HTMLDivElement>(null), triggerRef=useRef<HTMLButtonElement|null>(null); const t=useCopy(locale), c=chromeCopy[locale];
  const nav=(p:string)=>{history.pushState(null,"",p);setPath(p);setPopover(null)};
  useEffect(()=>subscribeSession(v=>{setTokens(v);if(!v)setUser(null)}),[]);
  useEffect(()=>{const f=()=>setPath(location.pathname);addEventListener("popstate",f);return()=>removeEventListener("popstate",f)},[]);
  useEffect(()=>{localStorage.setItem("dt-locale",locale);document.documentElement.lang=locale},[locale]);
  useEffect(()=>{localStorage.setItem("dt-theme",theme);document.documentElement.dataset.theme=theme},[theme]);
  useEffect(()=>{const rt=sessionStorage.getItem(refreshKey);if(!rt){setLoading(false);return} void request<Tokens>("/auth/refresh",{body:{refresh_token:rt}}).then(async x=>{setSession(x);setUser(await api.profile(x.access_token))}).catch(()=>setSession(null)).finally(()=>setLoading(false))},[]);
  async function refreshChrome(){if(!tokens)return; const [i,n,ic,nc]=await Promise.all([api.invites(tokens.access_token),api.notifications(tokens.access_token),api.inviteCount(tokens.access_token),api.notificationCount(tokens.access_token)]);setInvites(i);setNotifications(n.items);setInviteCount(ic.count);setNotificationCount(nc.count)}
  useEffect(()=>{if(!user||!tokens)return;void refreshChrome().catch(setError); const focus=()=>void refreshChrome().catch(setError);addEventListener("focus",focus);const timer=window.setInterval(()=>{if(!document.hidden)void refreshChrome().catch(setError)},60000);return()=>{removeEventListener("focus",focus);clearInterval(timer)}},[user?.id,tokens?.access_token]);
  useEffect(()=>{if(!popover)return;const down=(e:MouseEvent)=>{if(!popRef.current?.contains(e.target as Node))setPopover(null)};const key=(e:KeyboardEvent)=>{if(e.key==="Escape"){setPopover(null);triggerRef.current?.focus()}};addEventListener("mousedown",down);addEventListener("keydown",key);return()=>{removeEventListener("mousedown",down);removeEventListener("keydown",key)}},[popover]);
  async function submit(e:FormEvent){e.preventDefault();setBusy(true);setError(null);setNotice(false);try{if(mode==="register"){await api.register({username:username.trim(),email:email.trim(),password});setPassword("");setMode("verify");setNotice(true);return}if(mode==="verify"){if(!/^\d{6}$/.test(code))throw new Error("sixDigits");const x=await api.verify({email:email.trim(),code});setSession(x);setUser(await api.profile(x.access_token));nav("/projects");return}const x=await api.login({email:email.trim(),password});setSession(x);setUser(await api.profile(x.access_token));nav("/projects")}catch(e){setError(e)}finally{setBusy(false)}}
  async function decide(id:string,action:"accept"|"decline"){if(!tokens)return;setError(null);try{await api.decideInvite(tokens.access_token,id,action);await refreshChrome();setProjectsRevision(x=>x+1)}catch(e){setError(e)}}
  async function logout(){if(tokens)try{await request("/auth/logout",{body:{refresh_token:tokens.refresh_token}})}catch{}setSession(null);nav("/")}
  function toggle(next:typeof popover,e:React.MouseEvent<HTMLButtonElement>){setError(null);triggerRef.current=e.currentTarget;setPopover(v=>v===next?null:next)}
  if (location.pathname === "/preview/board-a") return <BoardPreview variant="a"/>;
  if (location.pathname === "/preview/board-b") return <BoardPreview variant="b"/>;
  if(loading)return <main className="auth-shell"><div className="auth-card">{c.restoring}</div></main>;
  if(user&&tokens){const taskMatch=new RegExp(`^/projects/(${uuid})/tasks/(${uuid})$`).exec(path),projectMatch=new RegExp(`^/projects/(${uuid})$`).exec(path);return <div className="app-shell">
    <header className="topbar"><button className="brand" onClick={()=>nav("/projects")}>Do Together</button><nav className="top-actions" ref={popRef}>
      <button className="icon-control" aria-label={t.invites} aria-expanded={popover==="invites"} onClick={e=>toggle("invites",e)}><Inbox/>{inviteCount>0&&<b className="badge">{inviteCount}</b>}<span className="tooltip">{t.invites}</span></button>
      <button className="icon-control" aria-label={t.notifications} aria-expanded={popover==="notifications"} onClick={e=>toggle("notifications",e)}><Bell/>{notificationCount>0&&<b className="badge">{notificationCount}</b>}<span className="tooltip">{t.notifications}</span></button>
      <button className="icon-control" aria-label={t.settings} aria-expanded={popover==="settings"} onClick={e=>toggle("settings",e)}><Settings/><span className="tooltip">{t.settings}</span></button>
      <button className="account-button" onClick={e=>toggle("settings",e)}><Avatar user={user}/><span>{user.display_name||user.username}</span></button>
      {popover&&<div className="popover" role="dialog"><div className="popover-title"><strong>{popover==="invites"?t.invites:popover==="notifications"?t.notifications:t.settings}</strong><button className="icon-control small" aria-label={t.close} onClick={()=>setPopover(null)}><X/></button></div>
        {popover==="invites"&&(invites.length?invites.map(i=><article className="popover-item" key={i.id}><strong>{i.project_title}</strong><small>{i.inviter_name}</small><div className="row"><button onClick={()=>void decide(i.id,"accept")}>{t.accept}</button><button className="secondary" onClick={()=>void decide(i.id,"decline")}>{t.decline}</button></div></article>):<p className="empty-copy">{t.emptyInvites}</p>)}
        {popover==="notifications"&&(notifications.length?notifications.map(n=><button className="notification-item" key={n.id} onClick={()=>{if(!n.read_at)void api.readNotification(tokens.access_token,n.id).then(refreshChrome).catch(setError);if(n.task_id)nav(`/projects/${n.project_id}/tasks/${n.task_id}`)}}><span>{notificationText(n.kind,n.payload,locale)}</span><small>{fmt(n.created_at,locale)}</small></button>):<p className="empty-copy">{t.emptyNotifications}</p>)}
        {popover==="settings"&&<SettingsPanel user={user} token={tokens.access_token} locale={locale} setLocale={setLocale} theme={theme} setTheme={setTheme} onUser={setUser} onLogout={()=>void logout()}/>}
        {!!error&&<p className="error" role="alert">{localizedRequestError(error,locale)}</p>}
      </div>}
    </nav></header>
    {(projectMatch||taskMatch)?<ProjectDetail projectID={(taskMatch??projectMatch)![1]} initialTaskID={taskMatch?.[2]} user={user} accessToken={tokens.access_token} locale={locale} onBack={()=>nav("/projects")} onTask={id=>nav(`/projects/${(taskMatch??projectMatch)![1]}/tasks/${id}`)}/>:<Projects accessToken={tokens.access_token} user={user} locale={locale} revision={projectsRevision} onOpen={id=>nav(`/projects/${id}`)}/>}
  </div>}
  return <main className="auth-shell"><section className="auth-card"><span className="eyebrow">DO TOGETHER</span><h1>{mode==="login"?(c.welcome):mode==="register"?(c.createAccount):(c.enterCode)}</h1>{mode==="verify"&&<p>{email}</p>}<form onSubmit={submit}>{mode==="register"&&<label>{c.username}<input value={username} onChange={e=>setUsername(e.target.value)} required autoComplete="username"/></label>}{mode!=="verify"&&<label>Email<input type="email" value={email} onChange={e=>setEmail(e.target.value)} required autoComplete="email"/></label>}{mode!=="verify"&&<label>{c.password}<input type="password" value={password} onChange={e=>setPassword(e.target.value)} required autoComplete={mode==="login"?"current-password":"new-password"}/></label>}{mode==="verify"&&<label className="otp-label"><span className="sr-only">{c.codeLabel}</span><div className="otp-slots" aria-hidden>{Array.from({length:6},(_,i)=><span key={i}>{code[i]||""}</span>)}</div><input className="otp-input" inputMode="numeric" autoComplete="one-time-code" value={code} onChange={e=>setCode(e.target.value.replace(/\D/g,"").slice(0,6))} autoFocus/></label>}{!!error&&<p className="error" role="alert">{error instanceof Error && error.message==="sixDigits" ? c.sixDigits : localizedRequestError(error,locale)}</p>}{notice&&<p className="notice">{c.codeSent}</p>}<button className="primary wide" disabled={busy}>{busy?"…":mode==="login"?(c.signIn):mode==="register"?(c.continue):(c.verify)}</button></form><button className="text-button" onClick={()=>setMode(mode==="login"?"register":"login")}>{mode==="login"?(c.noAccount):(c.haveAccount)}</button></section></main>;
}

export function SettingsPanel({user,token,locale,setLocale,theme,setTheme,onUser,onLogout}:{user:User;token:string;locale:Locale;setLocale:(x:Locale)=>void;theme:Theme;setTheme:(x:Theme)=>void;onUser:(x:User)=>void;onLogout:()=>void}) {
  const t=useCopy(locale), c=chromeCopy[locale];
  const [name,setName]=useState(user.display_name), [busy,setBusy]=useState(false), [error,setError]=useState<unknown>(null);
  const pending=useRef(false);
  async function update(action:()=>Promise<User>) {
    if(pending.current)return;
    pending.current=true;setBusy(true);setError(null);
    try { onUser(await action()); } catch(e) { setError(e); }
    finally { pending.current=false;setBusy(false); }
  }
  async function upload(file:File) {
    if(file.size>5*1024*1024||!/^image\/(jpeg|png)$/.test(file.type)){setError("avatarFormat");return;}
    await update(()=>api.uploadAvatar(token,file));
  }
  return <div className="settings-panel">
    <label>{t.language}<select value={locale} onChange={e=>setLocale(e.target.value as Locale)}><option value="ru">Русский</option><option value="en">English</option></select></label>
    <label>{t.theme}<select value={theme} onChange={e=>setTheme(e.target.value as Theme)}><option value="system">{t.system}</option><option value="light">{t.light}</option><option value="dark">{t.dark}</option></select></label>
    <hr/><strong>{t.profile}</strong>
    <div className="profile-avatar"><Avatar user={user} className="avatar large"/><label className="button secondary">{busy?c.saving:c.choosePhoto}<input disabled={busy} hidden type="file" accept="image/jpeg,image/png" onChange={e=>{const file=e.target.files?.[0];e.target.value="";if(file)void upload(file)}}/></label>{user.avatar_url&&<button disabled={busy} className="text-button danger-text" onClick={()=>void update(()=>api.deleteAvatar(token))}>{t.removeAvatar}</button>}</div>
    <label>{c.username}<input value={user.username} disabled/></label>
    <label>{c.displayName}<input value={name} onChange={e=>setName(e.target.value)} maxLength={80}/></label>
    {!!error&&<p className="error" role="alert">{error==="avatarFormat"?c.avatarFormat:localizedRequestError(error,locale)}</p>}
    <button disabled={busy} onClick={()=>void update(()=>api.updateProfile(token,name))}>{busy?c.saving:t.save}</button><button className="secondary" onClick={onLogout}>{t.logout}</button>
  </div>;
}
