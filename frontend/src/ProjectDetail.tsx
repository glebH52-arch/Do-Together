import {
  Fragment,
  useEffect,
  useMemo,
  useRef,
  useState,
  type DragEvent,
  type FormEvent,
} from "react";
import { History as HistoryIcon, Plus, X } from "lucide-react";
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
import type { Locale } from "./i18n";
const statuses: TaskStatus[] = ["todo", "in_progress", "review", "done"],
  statusName = {
    todo: "К выполнению",
    in_progress: "В работе",
    review: "На проверке",
    done: "Готово",
  },
  types: TaskType[] = ["task", "bug", "feature", "research"],
  typeName = {
    task: "Задача",
    bug: "Ошибка",
    feature: "Функция",
    research: "Исследование",
  },
  priorities: TaskPriority[] = ["low", "medium", "high", "critical"],
  priorityName = {
    low: "Низкий",
    medium: "Средний",
    high: "Высокий",
    critical: "Критический",
  };
const fmt = (v?: string | null) =>
  v ? new Date(v).toLocaleString("ru-RU") : "—";
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
  const [project, setProject] = useState<Project | null>(null),
    [members, setMembers] = useState<ProjectMember[]>([]),
    [tasks, setTasks] = useState<Task[]>([]),
    [total, setTotal] = useState(0),
    [selected, setSelected] = useState<Task | null>(null),
    [edit, setEdit] = useState(false),
    [create, setCreate] = useState(false),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false),
    [conflict, setConflict] = useState(false),
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
    [activityOpen,setActivityOpen]=useState(false),
    [activity,setActivity]=useState<Activity[]>([]),
    [activityCursor,setActivityCursor]=useState<string|null>(null);
  const requestSequence=useRef(0);
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
    setError("");
    try {
      const [p, r] = await Promise.all([
        request<Project>(`/projects/${projectID}`, { token: accessToken }),
        api.tasks(accessToken, projectID, filters),
      ]);
      const pg = page(r);
      if(sequence!==requestSequence.current)return;
      setProject(p);
      setTasks(pg.items);
      setTotal(pg.total);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка загрузки");
    }
  }
  useEffect(() => {
    const t = setTimeout(() => void load(), query ? 250 : 0);
    return () => clearTimeout(t);
  }, [projectID, accessToken, filters]);
  useEffect(() => {
    if (initialTaskID)
      void api
        .task(accessToken, projectID, initialTaskID)
        .then(openTask)
        .catch((e) => setError(e.message));
  }, [initialTaskID]);
  function openTask(t: Task) {
    setSelected(t);
    setEdit(false);
    setTab("details");
    setConflict(false);
    void Promise.all([
      api.comments(accessToken, projectID, t.id),
      api.history(accessToken, projectID, t.id),
    ])
      .then(([c, h]) => {
        setComments(page(c).items);
        setHistory(page(h).items);
      })
      .catch(() => {});
  }
  async function move(task: Task, next: TaskStatus) {
    if (task.status === next || moving.has(task.id)) return;
    if (!task.allowed_transitions.includes(next)) { setError("Этот переход недоступен для задачи."); return; }
    const before = tasks;
    setTasks((v) =>
      v.map((x) => (x.id === task.id ? { ...x, status: next } : x)),
    );
    setMoving(v=>new Set(v).add(task.id));
    try {
      await api.updateTask(accessToken, projectID, task.id, { status: next });
      await load();
    } catch (e) {
      setTasks(before);
      if (e instanceof ApiError && e.status === 409) setConflict(true);
      setError(e instanceof Error ? e.message : "Не удалось переместить");
    } finally {
      setMoving(v=>{const n=new Set(v);n.delete(task.id);return n});
    }
  }
  function drop(e: DragEvent, next: TaskStatus) {
    e.preventDefault();
    const id = e.dataTransfer.getData("text/plain"),
      task = tasks.find((v) => v.id === id);
    setDrag(null);
    if (task) void move(task, next);
  }
  async function openActivity(){setActivityOpen(true);if(activity.length)return;const p=await api.activity(accessToken,projectID);setActivity(p.items);setActivityCursor(p.next_cursor)}
  async function moreActivity(){if(!activityCursor)return;const p=await api.activity(accessToken,projectID,activityCursor);setActivity(v=>[...v,...p.items.filter(x=>!v.some(y=>y.id===x.id))]);setActivityCursor(p.next_cursor)}
  return (
    <>
      <main className="workspace-main">
        <button className="back-link" onClick={onBack}>
          ← Все проекты
        </button>
        {error && (
          <p className="error" role="alert">
            {error}
          </p>
        )}
        {conflict && (
          <div className="notice">
            Доска изменилась в другой сессии.{" "}
            <button
              className="linklike"
              onClick={() => {
                setConflict(false);
                void load();
              }}
            >
              Получить актуальное состояние
            </button>
          </div>
        )}
        {project && <div className="project-heading"><section className="project-intro"><h1>{project.title}</h1><p>{project.goal}</p></section><div className="row"><button className="secondary" onClick={()=>void openActivity()}><HistoryIcon/> {locale==="ru"?"История":"History"}</button><button onClick={() => setCreate(true)}><Plus/> {locale==="ru"?"Добавить задачу":"Add task"}</button></div></div>}
        <section className="filters" aria-label="Поиск и фильтры">
          <input
            aria-label="Поиск"
            placeholder="Поиск по названию и описанию"
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setOffset(0);
            }}
          />
          <select
            aria-label="Статус"
            value={status}
            onChange={(e) => setStatus(e.target.value)}
          >
            <option value="">Все статусы</option>
            {statuses.map((v) => (
              <option key={v} value={v}>
                {statusName[v]}
              </option>
            ))}
          </select>
          <select
            aria-label="Тип"
            value={type}
            onChange={(e) => setType(e.target.value)}
          >
            <option value="">Все типы</option>
            {types.map((v) => (
              <option key={v}>{v}</option>
            ))}
          </select>
          <select
            aria-label="Приоритет"
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
          >
            <option value="">Все приоритеты</option>
            {priorities.map((v) => (
              <option key={v}>{priorityName[v]}</option>
            ))}
          </select>
          <select
            aria-label="Исполнитель"
            value={assignee}
            onChange={(e) => setAssignee(e.target.value)}
          >
            <option value="">Все исполнители</option>
            {members.map((v) => (
              <option key={v.user_id} value={v.user_id}>
                {v.username}
              </option>
            ))}
          </select>
          <input
            aria-label="Срок с"
            type="date"
            value={dueFrom}
            onChange={(e) => {setDueFrom(e.target.value);setOffset(0)}}
          />
          <input aria-label="Срок по" type="date" value={dueTo} onChange={e=>{setDueTo(e.target.value);setOffset(0)}}/>
          <button className="secondary" onClick={()=>{setQuery("");setStatus("");setType("");setPriority("");setAssignee("");setDueFrom("");setDueTo("");setOffset(0)}}>{locale==="ru"?"Сбросить":"Reset"}</button>
          <select
            aria-label="Сортировка"
            value={sort}
            onChange={(e) => setSort(e.target.value)}
          >
            <option value="created_at">По созданию</option>
            <option value="updated_at">По обновлению</option>
            <option value="due_at">По сроку</option>
            <option value="priority">По приоритету</option>
          </select>
          <button
            className="secondary"
            onClick={() => setDirection((v) => (v === "asc" ? "desc" : "asc"))}
          >
            {direction === "asc" ? "↑" : "↓"}
          </button>
        </section>
        <div className="kanban-board">
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
                    <article
                      draggable={!moving.has(t.id)}
                      onDragStart={(e) => {
                        setDrag(t.id);
                        e.dataTransfer.setData("text/plain", t.id);
                      }}
                      onDragEnd={() => setDrag(null)}
                      className="task-card"
                      key={t.id}
                    >
                      <button
                        className="task-open linklike"
                        onClick={() => {
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
                          {t.assignee_name ?? "Не назначена"} ·{" "}
                          {t.due_at ? fmt(t.due_at) : "Без срока"}
                        </small>
                      </button>
                      <label className="move-select"><span className="sr-only">Переместить {t.title}</span><select value="" onChange={e=>e.target.value&&void move(t,e.target.value as TaskStatus)}><option value="">Переместить в…</option>{t.allowed_transitions.map(s=><option key={s} value={s}>{statusName[s]}</option>)}</select></label>
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
            Назад
          </button>
          <span>
            {total
              ? `${offset + 1}–${Math.min(offset + limit, total)} из ${total}`
              : "0 результатов"}
          </span>
          <button
            className="secondary"
            disabled={offset + limit >= total}
            onClick={() => setOffset(offset + limit)}
          >
            Далее
          </button>
        </div>
        <ProjectMembers
          projectID={projectID}
          user={user}
          accessToken={accessToken}
          onMembers={setMembers}
          locale={locale}
        />
      </main>
      {activityOpen&&<aside className="activity-panel" role="dialog" aria-modal="true" aria-label="История проекта"><div className="panel-heading"><h2>{locale==="ru"?"История проекта":"Project history"}</h2><button className="icon-control" aria-label="Закрыть" onClick={()=>setActivityOpen(false)}><X/></button></div><div className="activity-list">{activity.map(a=><article key={a.id}><time>{new Date(a.occurred_at).toLocaleString(locale==="ru"?"ru-RU":"en-US")}</time><p><strong>{a.actor.name}</strong> {activityText(a,locale)}</p>{Object.keys(a.data.new??{}).length>1&&<details><summary>{locale==="ru"?"Изменения":"Changes"}</summary><pre>{JSON.stringify({old:a.data.old,new:a.data.new},null,2)}</pre></details>}</article>)}{activityCursor&&<button className="secondary" onClick={()=>void moreActivity()}>{locale==="ru"?"Загрузить ещё":"Load more"}</button>}</div></aside>}
      {(create || selected) && (
        <TaskDialog
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
          onClose={() => {
            setCreate(false);
            setSelected(null);
              window.history.replaceState(null, "", `/projects/${projectID}`);
          }}
          onSaved={async () => {
            setCreate(false);
            setSelected(null);
            await load();
          }}
        />
      )}
    </>
  );
}

function TaskDialog({
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
  const [title, setTitle] = useState(task?.title ?? ""),
    [description, setDescription] = useState(task?.description ?? ""),
    [type, setType] = useState<TaskType>(task?.type ?? "task"),
    [priority, setPriority] = useState<TaskPriority>(
      task?.priority ?? "medium",
    ),
    [status, setStatus] = useState<TaskStatus>(task?.status ?? "todo"),
    [assignee, setAssignee] = useState(task?.assignee_id ?? ""),
    [due, setDue] = useState(task?.due_at ? task.due_at.slice(0, 16) : ""),
    [body, setBody] = useState(""),
    [error, setError] = useState(""),
    [busy, setBusy] = useState(false);
  async function save(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
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
        const old = task.due_at ? task.due_at.slice(0, 16) : "";
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
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(false);
    }
  }
  async function remove() {
    if (!task || !confirm(`Удалить задачу «${task.title}»?`)) return;
    await api.deleteTask(accessToken, projectID, task.id);
    await onSaved();
  }
  async function add(e: FormEvent) {
    e.preventDefault();
    if (!task) return;
    const c = await api.addComment(
      accessToken,
      projectID,
      task.id,
      body.trim(),
    );
    setComments([...comments, c]);
    setBody("");
  }
  return (
    <div className="dialog-backdrop">
      <div className="dialog-card task-dialog">
        <div className="dialog-title">
          <h2>{task ? task.title : "Новая задача"}</h2>
          <button className="secondary" onClick={onClose}>
            Закрыть
          </button>
        </div>
        {task && (
          <nav className="tabs">
            <button onClick={() => setTab("details")}>Детали</button>
            <button onClick={() => setTab("comments")}>Комментарии</button>
            <button onClick={() => setTab("history")}>История</button>
          </nav>
        )}
        {tab === "details" &&
          (!edit && task ? (
            <>
              <Markdown text={task.description} />
              <dl className="task-fields">
                <dt>Тип</dt>
                <dd>{typeName[task.type]}</dd>
                <dt>Приоритет</dt>
                <dd>{priorityName[task.priority]}</dd>
                <dt>Статус</dt>
                <dd>{statusName[task.status]}</dd>
                <dt>Автор</dt>
                <dd>{task.creator_name}</dd>
                <dt>Исполнитель</dt>
                <dd>{task.assignee_name ?? "Не назначен"}</dd>
                <dt>Срок</dt>
                <dd>{fmt(task.due_at)}</dd>
              </dl>
              <div className="inline-actions">
                <button onClick={() => setEdit(true)}>Редактировать</button>
                {task.created_by === user.id && (
                  <button className="danger" onClick={() => void remove()}>
                    Удалить
                  </button>
                )}
              </div>
            </>
          ) : (
            <form onSubmit={save}>
              <label>
                Название
                <input
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  required
                  maxLength={100}
                />
              </label>
              <label>
                Описание (Markdown)
                <textarea
                  value={description}
                  onChange={(e) => setDescription(e.target.value)}
                  required
                  rows={6}
                />
              </label>
              <div className="form-grid">
                <label>
                  Тип
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
                  Приоритет
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
                <label>
                  Статус
                  <select
                    value={status}
                    onChange={(e) => setStatus(e.target.value as TaskStatus)}
                  >
                    {statuses.map((v) => (
                      <option key={v} value={v}>
                        {statusName[v]}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Исполнитель
                  <select
                    value={assignee}
                    onChange={(e) => setAssignee(e.target.value)}
                  >
                    <option value="">Не назначен</option>
                    {members.map((v) => (
                      <option key={v.user_id} value={v.user_id}>
                        {v.username}
                      </option>
                    ))}
                  </select>
                </label>
                <label>
                  Срок
                  <input
                    type="datetime-local"
                    value={due}
                    onChange={(e) => setDue(e.target.value)}
                  />
                </label>
              </div>
              {error && <p className="error">{error}</p>}
              <div className="dialog-actions">
                {task && (
                  <button
                    type="button"
                    className="secondary"
                    onClick={() => setEdit(false)}
                  >
                    Отмена
                  </button>
                )}
                <button disabled={busy}>Сохранить</button>
              </div>
            </form>
          ))}
        {task && tab === "comments" && (
          <>
            <form className="comment-form" onSubmit={add}>
              <textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder="Написать комментарий"
                required
              />
              <button>Отправить</button>
            </form>
            <div className="timeline">
              {comments.map((c) => (
                <article key={c.id}>
                  <b>{c.author_name}</b>
                  <small>{fmt(c.created_at)}</small>
                  <p>{c.body}</p>
                  {c.author_id === user.id && (
                    <div className="inline-actions">
                      <button
                        className="secondary"
                        onClick={() => {
                          const v = prompt("Изменить комментарий", c.body);
                          if (v && task)
                            void api
                              .editComment(
                                accessToken,
                                projectID,
                                task.id,
                                c.id,
                                v,
                              )
                              .then(() =>
                                setComments(
                                  comments.map((x) =>
                                    x.id === c.id ? { ...x, body: v } : x,
                                  ),
                                ),
                              );
                        }}
                      >
                        Изменить
                      </button>
                      <button
                        className="delete-link"
                        onClick={() =>
                          void api
                            .deleteComment(
                              accessToken,
                              projectID,
                              task.id,
                              c.id,
                            )
                            .then(() =>
                              setComments(
                                comments.filter((x) => x.id !== c.id),
                              ),
                            )
                        }
                      >
                        Удалить
                      </button>
                    </div>
                  )}
                </article>
              ))}
            </div>
          </>
        )}
        {task && tab === "history" && (
          <div className="timeline">
            {history.map((h) => (
              <article key={h.id}>
                <b>{historyText(h)}</b>
                <small>
                  {h.actor_name} · {fmt(h.created_at)}
                </small>
              </article>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
function historyText(h: History) {
  const labels: Record<string, string> = {
    priority: "Приоритет",
    status: "Статус",
    assignee_id: "Исполнитель",
    due_at: "Срок",
    title: "Название",
    description: "Описание",
  };
  const keys = Object.keys(h.new_values ?? h.old_values ?? {});
  if (!keys.length) return h.action;
  return keys
    .map(
      (k) =>
        `${labels[k] ?? k}: ${String(h.old_values?.[k] ?? "—")} → ${String(h.new_values?.[k] ?? "—")}`,
    )
    .join("; ");
}
function activityText(a: Activity, locale: Locale) {
  const object=a.object.title?`«${a.object.title}»`:locale==="ru"?"объект":"object";
  const verbs:Record<string,[string,string]>={"task.created":["создал(а)","created"],"task.updated":["изменил(а)","updated"],"task.deleted":["удалил(а)","deleted"],"member.joined":["добавил(а) участника","added member"],"member.role_changed":["изменил(а) роль","changed role"],"project.updated":["изменил(а) проект","updated project"]};
  const verb=(verbs[a.type]??[a.type,a.type])[locale==="ru"?0:1];
  const oldStatus=a.data.old?.status, newStatus=a.data.new?.status;
  if(a.type==="task.updated"&&oldStatus&&newStatus)return locale==="ru"?`перенёс(ла) ${object}: ${statusName[String(oldStatus) as TaskStatus]} → ${statusName[String(newStatus) as TaskStatus]}`:`moved ${object}: ${String(oldStatus)} → ${String(newStatus)}`;
  return `${verb} ${object}`;
}
