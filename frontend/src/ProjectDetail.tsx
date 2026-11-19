import {
  useEffect,
  useMemo,
  useRef,
  useState,
  type DragEvent,
  type FormEvent,
  type ReactNode,
} from "react";
import { Filter, History as HistoryIcon, Plus, Search, Users, X } from "lucide-react";
import {
  ApiError,
  api,
  request,
  type Comment,
  type History,
  type Paged,
  type Project,
  type ProjectMember,
  type Task,
  type TaskPriority,
  type TaskStatus,
  type TaskType,
  type User,
  type Activity,
  type UUID,
} from "./api";
import { ProjectMembers } from "./ProjectMembers";
import { boardCopy, boardErrorText, activityCopy, roleLabel, fmt as formatDate, statusLabel, type Locale, type BoardErrorKey } from "./i18n";
import { ActionMenu, useModalSurface } from "./TaskSurface";
import { canMoveTask, mergeUniqueByID, moveTaskLocally } from "./project-detail-model";
const statuses: TaskStatus[] = ["todo", "in_progress", "review", "done"];
const types: TaskType[] = ["task", "bug", "feature", "research"];
const priorities: TaskPriority[] = ["low", "medium", "high", "critical"];
type UIError = {cause:unknown;fallback:BoardErrorKey}|null;
const names = (locale:Locale) => Object.fromEntries(statuses.map(status=>[status,statusLabel(status,locale)])) as Record<TaskStatus,string>;
const page = <T,>(v: Paged<T> | T[]): Paged<T> =>
  Array.isArray(v)
    ? { items: v, total: v.length, limit: v.length, offset: 0 }
    : v;
function Markdown({ text }: { text: string }) {
  const inline = (s: string) => {
    const parts = s.split(
      /(`[^`]+`|\*\*[^*]+\*\*|\*[^*]+\*|\[[^\]]+\]\(https?:\/\/[^)]+\))/g,
    );
    return parts.map((p, i) => {
      if (p.startsWith("`")) return <code key={i}>{p.slice(1, -1)}</code>;
      if (p.startsWith("**")) return <strong key={i}>{p.slice(2, -2)}</strong>;
      if (p.startsWith("*")) return <em key={i}>{p.slice(1, -1)}</em>;
      const m = /^\[([^\]]+)\]\((https?:\/\/[^)]+)\)$/.exec(p);
      return m ? (
        <a key={i} href={m[2]} target="_blank" rel="noreferrer">
          {m[1]}
        </a>
      ) : (
        p
      );
    });
  };
  return (
    <div className="markdown">
      {text
        .split("\n")
        .map((l, i) =>
          l.startsWith("- ") ? (
            <div key={i}>• {inline(l.slice(2))}</div>
          ) : (
            <p key={i}>{inline(l)}</p>
          ),
        )}
    </div>
  );
}
export function ProjectDetail({
  projectID,
  initialTaskID,
  user,
  accessToken,
  onBack,
  onTask,
  locale,
}: {
  projectID: UUID;
  initialTaskID?: UUID;
  user: User;
  accessToken: string;
  onBack: () => void;
  onTask: (id: UUID) => void;
  locale: Locale;
}) {
  const copy=boardCopy[locale],statusName=names(locale),typeName=copy.types,priorityName=copy.priorities;
  const fmt=(value?:string|null)=>formatDate(value,locale);
  const [project, setProject] = useState<Project | null>(null),
    [members, setMembers] = useState<ProjectMember[]>([]),
    [tasks, setTasks] = useState<Task[]>([]),
    [total, setTotal] = useState(0),
    [selected, setSelected] = useState<Task | null>(null),
    [edit, setEdit] = useState(false),
    [create, setCreate] = useState(false),
    [error, setError] = useState<UIError>(null),
    [loading,setLoading]=useState(true),
    [activityError,setActivityError]=useState<UIError>(null),
    [activityLoading,setActivityLoading]=useState(false),
    [query, setQuery] = useState(""),
    [status, setStatus] = useState(""),
    [type, setType] = useState(""),
    [priority, setPriority] = useState(""),
    [assignee, setAssignee] = useState(""),
    [dueFrom, setDueFrom] = useState(""),
    [dueTo, setDueTo] = useState(""),
    [sort, setSort] = useState("created_at"),
    [direction, setDirection] = useState("desc"),
    [offset, setOffset] = useState(0),
    [comments, setComments] = useState<Comment[]>([]),
    [history, setHistory] = useState<History[]>([]),
    [tab, setTab] = useState<"details" | "comments" | "history">("details"),
    [drag, setDrag] = useState<UUID | null>(null),
    [moving,setMoving]=useState<Set<UUID>>(new Set()),
    [moveErrors,setMoveErrors]=useState<Record<UUID,'moveDenied'|'moveFailed'|'moveConflict'>>({}),
    [activityOpen,setActivityOpen]=useState(false),
    [teamOpen,setTeamOpen]=useState(false),
    [filtersOpen,setFiltersOpen]=useState(false),
    [activity,setActivity]=useState<Activity[]>([]),
    [activityCursor,setActivityCursor]=useState<string|null>(null);
  const movingRef=useRef(new Set<UUID>());
  const requestSequence=useRef(0);
  const taskSequence=useRef(0);
  const loadController=useRef<AbortController|null>(null);
  const taskTrigger=useRef<HTMLElement|null>(null);
  const teamTrigger=useRef<HTMLButtonElement|null>(null);
  const activityTrigger=useRef<HTMLButtonElement|null>(null);
  const limit = 20;
  const filters = useMemo(
    () => ({
      q: query,
      status,
      type,
      priority,
      assignee_id: assignee,
      due_after: dueFrom ? new Date(dueFrom + "T00:00:00").toISOString() : undefined,
      due_before: dueTo ? new Date(new Date(dueTo + "T00:00:00").getTime()+86400000).toISOString() : undefined,
      sort,
      direction,
      limit,
      offset,
    }),
    [query, status, type, priority, assignee, dueFrom, dueTo, sort, direction, offset],
  );
  async function load() {
    const sequence=++requestSequence.current;
    loadController.current?.abort();
    const controller=new AbortController();
    loadController.current=controller;
    setError(null);setLoading(true);
    try {
      const [p, r] = await Promise.all([
        request<Project>(`/projects/${projectID}`, { token: accessToken, signal: controller.signal }),
        api.tasks(accessToken, projectID, filters, controller.signal),
      ]);
      const pg = page(r);
      if(sequence!==requestSequence.current)return;
      setProject(p);
      setTasks(pg.items);
      setTotal(pg.total);
    } catch (e) {
      if(e instanceof DOMException&&e.name==="AbortError")return;
      if(sequence===requestSequence.current)setError({cause:e,fallback:'loadFailed'});
    } finally {if(sequence===requestSequence.current)setLoading(false)}
  }
  useEffect(() => {
    const t = setTimeout(() => void load(), query ? 250 : 0);
    return () => clearTimeout(t);
  }, [projectID, accessToken, filters]);
  useEffect(()=>()=>loadController.current?.abort(),[]);
  useEffect(()=>{
    const controller=new AbortController();
    api.members(accessToken,projectID,controller.signal).then(setMembers).catch((e)=>{if(!(e instanceof DOMException&&e.name==="AbortError"))setError({cause:e,fallback:'membersFailed'})});
    return()=>controller.abort();
  },[projectID,accessToken,user.avatar_url,user.display_name]);
  useEffect(() => {
    if (initialTaskID && selected?.id!==initialTaskID)
      void api
        .task(accessToken, projectID, initialTaskID)
        .then(openTask)
        .catch((e) => setError({cause:e,fallback:'loadFailed'}));
  }, [initialTaskID]);
  function openTask(t: Task) {
    const sequence=++taskSequence.current;
    setTeamOpen(false);setActivityOpen(false);
    setSelected(t);
    setEdit(false);
    setTab("details");
    setComments([]);setHistory([]);
    void Promise.all([
      api.comments(accessToken, projectID, t.id),
      api.history(accessToken, projectID, t.id),
    ])
      .then(([c, h]) => {
        if(sequence!==taskSequence.current)return;
        setComments(page(c).items);
        setHistory(page(h).items);
      })
      .catch(cause=>{if(sequence===taskSequence.current)setError({cause,fallback:'loadFailed'})});
  }
  async function move(task: Task, next: TaskStatus) {
    if (task.status === next || movingRef.current.has(task.id)) return;
    if (!canMoveTask(task,next)) {setMoveErrors(v=>({...v,[task.id]:'moveDenied'}));return;}
    movingRef.current.add(task.id);setMoving(new Set(movingRef.current));
    setTasks(current=>moveTaskLocally(current,task.id,next));
    try {
      await api.updateTask(accessToken,projectID,task.id,{status:next});
      setMoveErrors(current=>{const nextErrors={...current};delete nextErrors[task.id];return nextErrors;});
      await load();
    } catch(cause) {
      setTasks(current=>current.map(item=>item.id===task.id?task:item));
      setMoveErrors(current=>({...current,[task.id]:cause instanceof ApiError&&cause.status===409?'moveConflict':'moveFailed'}));
    } finally {movingRef.current.delete(task.id);setMoving(new Set(movingRef.current));}
  }
  function drop(e: DragEvent, next: TaskStatus) {
    e.preventDefault();
    const id = e.dataTransfer.getData("text/plain"),
      task = tasks.find((v) => v.id === id);
    setDrag(null);
    if (task) void move(task, next);
  }
  async function openActivity(){setTeamOpen(false);setActivityOpen(true);setActivityLoading(true);setActivityError(null);try{const p=await api.activity(accessToken,projectID);setActivity(p.items);setActivityCursor(p.next_cursor)}catch(cause){setActivityError({cause,fallback:'loadFailed'})}finally{setActivityLoading(false)}}
  async function moreActivity(){if(!activityCursor||activityLoading)return;setActivityLoading(true);setActivityError(null);try{const p=await api.activity(accessToken,projectID,activityCursor);setActivity(v=>mergeUniqueByID(v,p.items));setActivityCursor(p.next_cursor)}catch(cause){setActivityError({cause,fallback:'loadFailed'})}finally{setActivityLoading(false)}}
  const closeTask=()=>{taskSequence.current++;setCreate(false);setSelected(null);window.history.replaceState(null,"",`/projects/${projectID}`);requestAnimationFrame(()=>taskTrigger.current?.focus())};
  const closeTeam=()=>{setTeamOpen(false);requestAnimationFrame(()=>teamTrigger.current?.focus())};
  const closeActivity=()=>{setActivityOpen(false);requestAnimationFrame(()=>activityTrigger.current?.focus())};

  return (
    <>
      <main className="workspace-main">
        <button className="back-link" onClick={onBack}>
          ← {copy.allProjects}
        </button>
        {error && (
          <p className="error" role="alert">
            {boardErrorText(error.cause,error.fallback,locale)}
          </p>
        )}
        {project && <div className="project-heading"><section className="project-intro"><div className="project-crumbs">{copy.projects} <span>/</span> {project.title}</div><h1>{project.title}</h1><p>{project.goal}</p></section><div className="project-actions"><button ref={teamTrigger} className="secondary" onClick={()=>{setActivityOpen(false);setTeamOpen(true)}}><Users/> {copy.team}</button><button ref={activityTrigger} className="secondary" aria-label={copy.projectHistory} title={copy.projectHistory} onClick={()=>void openActivity()}><HistoryIcon/></button><button onClick={(e) => {taskTrigger.current=e.currentTarget;setCreate(true)}}><Plus/> {copy.task}</button></div></div>}
        <div className="board-toolbar"><label className="board-search"><Search/><input aria-label={copy.search} placeholder={copy.searchTasks} value={query} onChange={e=>{setQuery(e.target.value);setOffset(0)}}/></label><button className="secondary" aria-expanded={filtersOpen} onClick={()=>setFiltersOpen(v=>!v)}><Filter/> {copy.filters}</button></div>
        {filtersOpen&&<section className="filters" aria-label={copy.searchFilters}>
          <select
            aria-label={copy.status}
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="">{copy.allStatuses}</option>
            {statuses.map((v) => (
              <option key={v} value={v}>
                {statusName[v]}
              </option>
            ))}
          </select>
          <select
            aria-label={copy.type}
            value={type}
            onChange={(e) => setType(e.target.value)}
          >
            <option value="">{copy.allTypes}</option>
            {types.map((v) => (
              <option key={v} value={v}>{typeName[v]}</option>
            ))}
          </select>
          <select
            aria-label={copy.priority}
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
          >
            <option value="">{copy.allPriorities}</option>
            {priorities.map((v) => (
              <option key={v} value={v}>{priorityName[v]}</option>
            ))}
          </select>
          <select
            aria-label={copy.assignee}
            value={assignee}
            onChange={(e) => setAssignee(e.target.value)}
          >
            <option value="">{copy.allAssignees}</option>
            {members.map((v) => (
              <option key={v.user_id} value={v.user_id}>
                {v.username}
              </option>
            ))}
          </select>
          <input
            aria-label={copy.dueFrom}
            type="date"
            value={dueFrom}
            onChange={(e) => {setDueFrom(e.target.value);setOffset(0)}}
          />
          <input aria-label={copy.dueTo} type="date" value={dueTo} onChange={e=>{setDueTo(e.target.value);setOffset(0)}}/>
          <button className="secondary" onClick={()=>{setQuery("");setStatus("");setType("");setPriority("");setAssignee("");setDueFrom("");setDueTo("");setOffset(0)}}>{copy.reset}</button>
          <select
            aria-label={copy.sorting}
            value={sort}
            onChange={(e) => setSort(e.target.value)}
          >
            <option value="created_at">{copy.createdSort}</option>
            <option value="updated_at">{copy.updatedSort}</option>
            <option value="due_at">{copy.dueSort}</option>
            <option value="priority">{copy.prioritySort}</option>
          </select>
          <button
            className="secondary"
            onClick={() => setDirection((v) => (v === "asc" ? "desc" : "asc"))}
          >
            {direction === "asc" ? "↑" : "↓"}
          </button>
        </section>}
        {(query||status||type||priority||assignee||dueFrom||dueTo)&&<div className="active-filter-chips" aria-label={copy.activeFilters}>
          {query&&<button onClick={()=>setQuery("")}>{copy.search}: {query} ×</button>}
          {status&&<button onClick={()=>setStatus("")}>{copy.status}: {statusName[status as TaskStatus]} ×</button>}
          {type&&<button onClick={()=>setType("")}>{copy.type}: {typeName[type as TaskType]} ×</button>}
          {priority&&<button onClick={()=>setPriority("")}>{copy.priority}: {priorityName[priority as TaskPriority]} ×</button>}
          {assignee&&<button onClick={()=>setAssignee("")}>{copy.assignee}: {members.find(m=>m.user_id===assignee)?.display_name||members.find(m=>m.user_id===assignee)?.username||"—"} ×</button>}
          {dueFrom&&<button onClick={()=>setDueFrom("")}>{copy.dueFrom}: {dueFrom} ×</button>}
          {dueTo&&<button onClick={()=>setDueTo("")}>{copy.dueTo}: {dueTo} ×</button>}
        </div>}
        {total===0&&(query||status||type||priority||assignee||dueFrom||dueTo)&&<p className="empty-filter-state">{copy.emptyFilters}</p>}
        {loading&&<p role="status">{copy.loading}</p>}
        <div className="kanban-board" aria-busy={loading}>
          {statuses.map((col) => (
            <section
              className={`kanban-column ${drag ? "dragging" : ""} ${drag&&tasks.find(t=>t.id===drag)?.allowed_transitions.includes(col)?"drop-allowed":""}`}
              key={col}
              onDragOver={(e) => e.preventDefault()}
              onDrop={(e) => drop(e, col)}
            >
              <h3 className="column-heading">
                <span>{statusName[col]}</span>
                <b>{tasks.filter((v) => v.status === col).length}</b>
              </h3>
              <div className="column-tasks">
                {tasks
                  .filter((v) => v.status === col)
                  .map((t) => (
                    <article data-task-id={t.id}
                      draggable={!moving.has(t.id)}
                      onDragStart={e=>{setDrag(t.id);e.dataTransfer.effectAllowed="move";e.dataTransfer.setData("text/plain",t.id)}}
                      onDragEnd={()=>setDrag(null)}
                      className={`task-card ${moving.has(t.id)?"task-moving":""}`}
                      key={t.id}
                    >
                      <button
                        className="task-open linklike"
                        onClick={(e) => {
                          taskTrigger.current=e.currentTarget;
                          openTask(t);
                          onTask(t.id);
                        }}
                      >
                        <h4>{t.title}</h4>
                        <p>{t.description}</p>
                        <div className="chips">
                          <span>{typeName[t.type]}</span>
                          <span className={`priority ${t.priority}`}>
                            {priorityName[t.priority]}
                          </span>
                        </div>
                        <small>
                          {t.assignee_name ?? copy.unassigned} ·{" "}
                          {t.due_at ? fmt(t.due_at) : copy.noDue}
                        </small>
                      </button>
                      <small className="task-move-error" role="alert" aria-live="assertive">{moveErrors[t.id]?boardCopy[locale][moveErrors[t.id]]:null}</small>{moving.has(t.id)&&<span className="task-saving" role="status">{boardCopy[locale].saving}</span>}
                      <div className="status-menu"><ActionMenu label={`${copy.changeStatus} ${t.title}`}>{close=><>{t.allowed_transitions.length===0?<span>{copy.noTransitions}</span>:t.allowed_transitions.map(next=><button role="menuitem" disabled={moving.has(t.id)} key={next} onClick={()=>{close();void move(t,next)}}>{copy.moveTo} {statusName[next]}</button>)}</>}</ActionMenu></div>
                    </article>
                  ))}
              </div>
            </section>
          ))}
        </div>
        <div className="pager">
          <button
            className="secondary"
            disabled={offset === 0}
            onClick={() => setOffset(Math.max(0, offset - limit))}
          >
            {copy.previous}
          </button>
          <span>
            {total
              ? `${offset + 1}–${Math.min(offset + limit, total)} ${copy.of} ${total}`
              : copy.noResults}
          </span>
          <button
            className="secondary"
            disabled={offset + limit >= total}
            onClick={() => setOffset(offset + limit)}
          >
            {copy.next}
          </button>
        </div>
      </main>
      {activityOpen&&<Drawer title={copy.projectHistory} closeLabel={copy.close} onClose={closeActivity} className="team-drawer activity-panel"><div className="activity-list">{activityError&&<p className="error" role="alert">{boardErrorText(activityError.cause,activityError.fallback,locale)} <button onClick={()=>void openActivity()}>{copy.retry}</button></p>}{activity.map(a=><article key={a.id}><time>{fmt(a.occurred_at)}</time><p><strong>{a.actor.name}</strong> {activityText(a,locale)}</p>{Object.keys(a.data.new??{}).length>1&&<details><summary>{copy.changes}</summary><p>{historyText({id:a.id,task_id:a.object.id,actor_id:a.actor.id,actor_name:a.actor.name,action:a.type,created_at:a.occurred_at,old_values:a.data.old,new_values:a.data.new},locale)}</p></details>}</article>)}{activityLoading&&<p role="status">{copy.loading}</p>}{!activityLoading&&!activityError&&activity.length===0&&<p>{copy.emptyActivity}</p>}{activityCursor&&<button className="secondary" disabled={activityLoading} onClick={()=>void moreActivity()}>{copy.loadMore}</button>}</div></Drawer>}
      {teamOpen&&<Drawer title={copy.team} closeLabel={copy.close} onClose={closeTeam} className="team-drawer"><div className="team-drawer-content"><ProjectMembers projectID={projectID} user={user} accessToken={accessToken} onMembers={setMembers} locale={locale}/></div></Drawer>}
      {(create || selected) && (
        <TaskDialog
          locale={locale}
          key={create?"new":selected?.id}
          task={create ? null : selected}
          members={members}
          user={user}
          accessToken={accessToken}
          projectID={projectID}
          edit={create || edit}
          setEdit={setEdit}
          tab={tab}
          setTab={setTab}
          comments={comments}
          setComments={setComments}
          history={history}
          onClose={closeTask}
          onSaved={async () => {
            closeTask();
            await load();
          }}
        />
      )}
    </>
  );
}

export function TaskDialog({
  locale,
  task,
  members,
  user,
  accessToken,
  projectID,
  edit,
  setEdit,
  tab,
  setTab,
  comments,
  setComments,
  history,
  onClose,
  onSaved,
}: {
  locale: Locale;
  task: Task | null;
  members: ProjectMember[];
  user: User;
  accessToken: string;
  projectID: UUID;
  edit: boolean;
  setEdit: (v: boolean) => void;
  tab: "details" | "comments" | "history";
  setTab: (v: any) => void;
  comments: Comment[];
  setComments: (v: Comment[]) => void;
  history: History[];
  onClose: () => void;
  onSaved: () => void;
}) {
  const copy=boardCopy[locale],statusName=names(locale),typeName=copy.types,priorityName=copy.priorities;
  const fmt=(value?:string|null)=>formatDate(value,locale);
  const dialogRef=useRef<HTMLDivElement>(null);useModalSurface(dialogRef,onClose);
  const [title, setTitle] = useState(task?.title ?? ""),
    [description, setDescription] = useState(task?.description ?? ""),
    [type, setType] = useState<TaskType>(task?.type ?? "task"),
    [priority, setPriority] = useState<TaskPriority>(
      task?.priority ?? "medium",
    ),
    [status, setStatus] = useState<TaskStatus>(task?.status ?? "todo"),
    [assignee, setAssignee] = useState(task?.assignee_id ?? ""),
    [due, setDue] = useState(localDateTime(task?.due_at)),
    [body, setBody] = useState(""),
    [editingComment, setEditingComment] = useState<UUID | null>(null),
    [editingBody, setEditingBody] = useState(""),
    [error, setError] = useState<UIError>(null),
    [busy, setBusy] = useState(false),
    [commentBusy,setCommentBusy]=useState(false);
  const submitting=useRef(false),commentSubmitting=useRef(false);
  async function save(e: FormEvent) {
    e.preventDefault();
    if(submitting.current)return;
    submitting.current=true;
    setBusy(true);
    setError(null);
    try {
      const base: any = {
        title: title.trim(),
        description: description.trim(),
        type,
        priority,
      };
      if (task) {
        if (status !== task.status) base.status = status;
        if (assignee !== (task.assignee_id ?? ""))
          base.assignee_id = assignee || null;
        const old = localDateTime(task.due_at);
        if (due !== old) base.due_at = due ? new Date(due).toISOString() : null;
        await api.updateTask(accessToken, projectID, task.id, base);
      } else
        await api.createTask(accessToken, projectID, {
          ...base,
          assignee_id: assignee || null,
          due_at: due ? new Date(due).toISOString() : null,
        });
      await onSaved();
    } catch (e) {
      setError({cause:e,fallback:'saveFailed'});
    } finally {
      submitting.current=false;
      setBusy(false);
    }
  }
  async function remove(){if(!task||submitting.current||!confirm(`${copy.deleteTask} «${task.title}»?`))return;submitting.current=true;setBusy(true);setError(null);try{await api.deleteTask(accessToken,projectID,task.id);await onSaved();}catch(cause){setError({cause,fallback:'removeFailed'})}finally{submitting.current=false;setBusy(false)}}
  async function add(e: FormEvent) {
    e.preventDefault();
    if (!task||commentSubmitting.current||!body.trim()) return;
    commentSubmitting.current=true;
    setCommentBusy(true);setError(null);
    try {
      const c = await api.addComment(accessToken,projectID,task.id,body.trim());
      setComments([...comments, c]);
      setBody("");
    } catch(e) { setError({cause:e,fallback:'commentFailed'}); }
    finally { commentSubmitting.current=false;setCommentBusy(false); }
  }
  async function saveComment(comment:Comment){
    if(!task||commentSubmitting.current||!editingBody.trim())return;commentSubmitting.current=true;setCommentBusy(true);setError(null);
    const text=editingBody.trim();try{await api.editComment(accessToken,projectID,task.id,comment.id,text);setComments(comments.map(item=>item.id===comment.id?{...item,body:text}:item));setEditingComment(null);}catch(cause){setError({cause,fallback:'commentEditFailed'});}finally{commentSubmitting.current=false;setCommentBusy(false);}
  }
  async function removeComment(comment:Comment){
    if(!task||commentSubmitting.current||!confirm(copy.deleteComment))return;commentSubmitting.current=true;setCommentBusy(true);setError(null);
    try{await api.deleteComment(accessToken,projectID,task.id,comment.id);setComments(comments.filter(item=>item.id!==comment.id));}catch(cause){setError({cause,fallback:'commentDeleteFailed'});}finally{commentSubmitting.current=false;setCommentBusy(false);}
  }
  return (
    <div className="dialog-backdrop" onMouseDown={(e)=>{if(e.target===e.currentTarget)onClose()}}>
      <div ref={dialogRef} tabIndex={-1} className="dialog-card task-dialog" role="dialog" aria-modal="true" aria-label={task?`${copy.task} ${task.title}`:copy.newTask}>
        <div className="dialog-title">
          <h2>{task ? task.title : copy.newTask}</h2>
          <button type="button" className="icon-control" aria-label={copy.close} onClick={onClose}><X/></button>
        </div>
        <div className={`task-dialog-body ${task?"":"task-dialog-body-single"}`}><section className="task-details">
        {error&&<p className="error" role="alert">{boardErrorText(error.cause,error.fallback,locale)}</p>}
        {tab === "details" &&
          (!edit && task ? (
            <>
              <Markdown text={task.description} />
              <dl className="task-fields">
                <dt>{copy.type}</dt>
                <dd>{typeName[task.type]}</dd>
                <dt>{copy.priority}</dt>
                <dd>{priorityName[task.priority]}</dd>
                <dt>{copy.status}</dt>
                <dd>{statusName[task.status]}</dd>
                <dt>{copy.author}</dt>
                <dd>{task.creator_name}</dd>
                <dt>{copy.assignee}</dt>
                <dd>{task.assignee_name ?? copy.unassigned}</dd>
                <dt>{copy.due}</dt>
                <dd>{fmt(task.due_at)}</dd>
              </dl>
              <div className="inline-actions">
                <button onClick={() => setEdit(true)}>{copy.edit}</button>
                {task.created_by === user.id && (
                  <button className="danger" onClick={() => void remove()}>
                    {copy.remove}
                  </button>
                )}
              </div>
            </>
          ) : (
            <form onSubmit={save}>
              <label>
                {copy.title}
                <input
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  required
                  maxLength={100}
                />
              </label>
              <label>
                {copy.markdownDescription}
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  required
                  rows={6}
                />
              </label>
              <div className="form-grid">
                <label>
                  {copy.type}
                  <select
                    value={type}
                    onChange={(e) => setType(e.target.value as TaskType)}
                  >
                    {types.map((v) => (
                      <option key={v} value={v}>
                        {typeName[v]}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  {copy.priority}
                  <select
                    value={priority}
                    onChange={(e) =>
                      setPriority(e.target.value as TaskPriority)
                    }
                  >
                    {priorities.map((v) => (
                      <option key={v} value={v}>
                        {priorityName[v]}
                      </option>
                    ))}
                  </select>
                </label>
                {task&&<label>
                  {copy.status}
                  <select value={status} onChange={(e) => setStatus(e.target.value as TaskStatus)}>
                    <option value={task.status}>{statusName[task.status]}</option>
                    {task.allowed_transitions.map((v) => <option key={v} value={v}>{statusName[v]}</option>)}
                  </select>
                </label>}
                <label>
                  {copy.assignee}
                  <select
                    value={assignee}
                    onChange={(e) => setAssignee(e.target.value)}
                  >
                    <option value="">{copy.unassigned}</option>
                    {members.map((v) => (
                      <option key={v.user_id} value={v.user_id}>
                        {v.username}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  {copy.due}
                  <input
                    type="datetime-local"
                    value={due}
                    onChange={(e) => setDue(e.target.value)}
                    onInput={(e) => setDue(e.currentTarget.value)}
                  />
                </label>
              </div>
              <div className="dialog-actions">
                {task && (
                  <button
                    type="button"
                    className="secondary"
                    onClick={() => setEdit(false)}
                  >
                    {copy.cancel}
                  </button>
                )}
                <button disabled={busy}>{busy?copy.saving:copy.save}</button>
              </div>
            </form>
          ))}
        {task && !edit && (
          <details className="task-history"><summary>{copy.history}</summary><div className="timeline">
            {history.map((h) => (
              <article key={h.id}>
                <b>{historyText(h,locale)}</b>
                <small>
                  {h.actor_name} · {fmt(h.created_at)}
                </small>
              </article>
            ))}
          </div></details>
        )}
        </section>
        {task && <aside className="task-comments-side">
          <div className="task-comments-heading"><h3>{copy.comments}</h3><span>{comments.length}</span></div>
          <div className="task-comments-list">{comments.length===0&&<p className="muted">{copy.noComments}</p>}{comments.map(c=><article className="comment" key={c.id}>
            <div className="comment-header"><strong>{c.author_name}</strong><div className="comment-meta"><time dateTime={c.created_at}>{fmt(c.created_at)}</time>{canManageComment(c,user.id)&&<ActionMenu label={`${copy.commentActions} ${c.author_name}`}>{close=><><button role="menuitem" disabled={commentBusy} onClick={()=>{close();setEditingComment(c.id);setEditingBody(c.body)}}>{copy.editComment}</button><button role="menuitem" className="delete-link" disabled={commentBusy} onClick={()=>{close();void removeComment(c)}}>{copy.remove}</button></>}</ActionMenu>}</div></div>
            {editingComment===c.id?<div className="comment-editor"><textarea aria-label={copy.editComment} value={editingBody} onChange={e=>setEditingBody(e.target.value)} rows={3}/><div className="inline-actions"><button disabled={commentBusy||!editingBody.trim()} onClick={()=>void saveComment(c)}>{copy.save}</button><button className="secondary" disabled={commentBusy} onClick={()=>setEditingComment(null)}>{copy.cancel}</button></div></div>:<p className="comment-body">{c.body}</p>}
          </article>)}</div>
          <form className="task-comment-compose" onSubmit={add}><textarea aria-label={copy.writeComment} value={body} onChange={e=>setBody(e.target.value)} placeholder={copy.writeComment} required rows={3}/><button disabled={commentBusy||!body.trim()}>{commentBusy?copy.sending:copy.send}</button></form>
        </aside>}

        </div>
      </div>
    </div>
  );
}
export function canManageComment(comment:Pick<Comment,'author_id'>,userID:UUID){return comment.author_id===userID;}
export function historyText(h: History,locale:Locale) {
  const copy=boardCopy[locale],activity=activityCopy[locale];
  const labels:Record<string,string>=activity.fields;
  const value=(key:string,input:unknown):string=>{
    if(input==null)return '—';
    if(typeof input==='boolean')return input?activity.yes:activity.no;
    if(key==='status')return (activity.statuses as Record<string,string>)[String(input)]??statusLabel(String(input),locale);
    if(key==='priority')return copy.priorities[input as TaskPriority]??String(input);
    if(key==='type')return copy.types[input as TaskType]??String(input);
    if(key==='role')return roleLabel(String(input),locale);
    if(['due_at','created_at','updated_at','completed_at','expires_at'].includes(key))return formatDate(String(input),locale);
    if(typeof input==='object')return Object.entries(input).map(([field,item])=>`${labels[field]??activity.details}: ${value(field,item)}`).join('; ');
    return String(input);
  };
  // Task edits carry a nested changes object; project/member events carry flat snapshots.
  const nested=h.new_values?.changes;
  if(nested&&typeof nested==='object')return Object.entries(nested).map(([key,item])=>{
    const change=item as {old?:unknown;new?:unknown;changed?:boolean};
    return `${labels[key]??activity.details}: ${change.changed?activity.changed:`${value(key,change.old)} → ${value(key,change.new)}`}`;
  }).join('; ');
  const keys=[...new Set([...Object.keys(h.old_values??{}),...Object.keys(h.new_values??{})])];
  if(!keys.length)return (copy.actions as Record<string,string>)[h.action]??copy.changed;
  return keys.map(key=>`${labels[key]??activity.details}: ${value(key,h.old_values?.[key])} → ${value(key,h.new_values?.[key])}`).join('; ');
}
function activityText(a:Activity,locale:Locale){
  const copy=boardCopy[locale],object=a.object.title?`«${a.object.title}»`:copy.object;
  const oldStatus=a.data.old?.status,newStatus=a.data.new?.status;
  if(a.type.startsWith('task.')&&oldStatus&&newStatus)return `${copy.moved} ${object}: ${statusLabel(String(oldStatus),locale)} → ${statusLabel(String(newStatus),locale)}`;
  return `${(copy.actions as Record<string,string>)[a.type]??copy.changed} ${object}`;
}

function Drawer({title,closeLabel,onClose,className,children}:{title:string;closeLabel:string;onClose:()=>void;className:string;children:ReactNode}){
  const ref=useRef<HTMLElement>(null);useModalSurface(ref,onClose);
  return <div className="team-overlay" onMouseDown={event=>{if(event.target===event.currentTarget)onClose()}}><aside ref={ref} tabIndex={-1} className={className} role="dialog" aria-modal="true" aria-label={title}><div className="panel-heading"><h2>{title}</h2><button className="icon-control" aria-label={closeLabel} onClick={onClose}><X/></button></div>{children}</aside></div>;
}

function localDateTime(value:string|null|undefined){if(!value)return '';const date=new Date(value);return new Date(date.getTime()-date.getTimezoneOffset()*60000).toISOString().slice(0,16);}
