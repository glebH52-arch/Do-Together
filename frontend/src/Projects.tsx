import { useEffect, useState, type FormEvent } from "react";
import { api, type Project, type User, type UUID } from "./api";
import type { Locale } from "./i18n";
export function Projects({
  accessToken,
  user,
  onOpen,
  locale,
  revision,
}: {
  accessToken: string;
  user: User;
  onOpen: (id: UUID) => void;
  locale: Locale;
  revision: number;
}) {
  const [items, setItems] = useState<Project[] | null>(null),
    [error, setError] = useState(""),
    [dialog, setDialog] = useState<false | "create" | Project>(false),
    [title, setTitle] = useState(""),
    [goal, setGoal] = useState(""),
    [busy, setBusy] = useState(false);
  const load = () =>
    Promise.all([api.projects(accessToken,"active"),api.projects(accessToken,"archived")])
      .then(([active,archived])=>setItems([...active,...archived]))
      .catch((e) => setError(e.message));
  useEffect(() => {
    void load();
  }, [accessToken,revision]);
  function open(v: "create" | Project) {
    setDialog(v);
    setTitle(v === "create" ? "" : v.title);
    setGoal(v === "create" ? "" : v.goal);
    setError("");
  }
  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError("");
    try {
      if (dialog === "create")
        await api.createProject(accessToken, {
          title: title.trim(),
          goal: goal.trim(),
        });
      else if (dialog)
        await api.updateProject(accessToken, dialog.id, {
          title: title.trim(),
          goal: goal.trim(),
        });
      setDialog(false);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(false);
    }
  }
  async function archive(p: Project) {
    if (!confirm(`Архивировать «${p.title}»?`)) return;
    try {
      await api.archiveProject(accessToken, p.id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    }
  }
  async function restore(p: Project) { try { await api.restoreProject(accessToken,p.id); await load(); } catch(e){ setError(e instanceof Error?e.message:"Ошибка") } }
  const active=(items??[]).filter(p=>p.status==="active"), archived=(items??[]).filter(p=>p.status==="archived");
  return (
    <main className="workspace-main">
      <div className="section-heading">
        <div>
          <h1>Мои проекты</h1>
          <p className="muted">
            Рабочие пространства, к которым у вас есть доступ
          </p>
        </div>
        <button onClick={() => open("create")}>+ Создать проект</button>
      </div>
      {error && <p className="error">{error}</p>}
      {items === null ? (
        <p>Загружаем проекты…</p>
      ) : items.length === 0 ? (
        <div className="state-card">
          <svg className="empty-illustration" viewBox="0 0 180 120" role="img" aria-label="Пустая доска"><rect x="20" y="18" width="140" height="84" rx="10" fill="none" stroke="currentColor" strokeWidth="4"/><path d="M40 39h35M40 54h24M91 39h48M91 54h34M40 77h54" fill="none" stroke="currentColor" strokeWidth="7" strokeLinecap="round"/><circle cx="130" cy="78" r="12" fill="var(--accent)"/><path d="M130 72v12m-6-6h12" stroke="white" strokeWidth="3" strokeLinecap="round"/></svg>
          <h2>Пока нет проектов</h2>
          <p>Создайте первый проект для совместной работы.</p>
        </div>
      ) : (
        <><h2 className="list-heading">{locale==="ru"?"Активные":"Active"}</h2><div className="project-grid">
          {active.map((p) => (
            <article className="project-card" key={p.id}>
              <button
                className="project-link linklike"
                onClick={() => onOpen(p.id)}
              >
                <div className="project-card-top">
                  <h2>{p.title}</h2>
                  <span className="status">
                    {p.status === "archived" ? "Архив" : "Активен"}
                  </span>
                </div>
                <p>{p.goal}</p>
              </button>
              {p.created_by === user.id && p.status === "active" && (
                <div className="inline-actions">
                  <button className="secondary" onClick={() => open(p)}>
                    Изменить
                  </button>
                  <button className="danger" onClick={() => void archive(p)}>
                    В архив
                  </button>
                </div>
              )}
            </article>
          ))}
        </div><h2 className="list-heading">{locale==="ru"?"Архив":"Archive"}</h2>{archived.length===0?<p className="muted">{locale==="ru"?"Архив пуст":"Archive is empty"}</p>:<div className="project-grid">{archived.map(p=><article className="project-card" key={p.id}><div className="project-card-top"><h2>{p.title}</h2><span className="status">{locale==="ru"?"Архив":"Archived"}</span></div><p>{p.goal}</p>{p.created_by===user.id&&<button className="secondary" onClick={()=>void restore(p)}>{locale==="ru"?"Восстановить":"Restore"}</button>}</article>)}</div>}</>
      )}
      {dialog && (
        <div className="dialog-backdrop">
          <div className="dialog-card">
            <h2>{dialog === "create" ? "Новый проект" : "Изменить проект"}</h2>
            <form onSubmit={submit}>
              <label>
                Название
                <input
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  maxLength={100}
                  required
                />
              </label>
              <label>
                Цель
                <textarea
                  value={goal}
                  onChange={(e) => setGoal(e.target.value)}
                  maxLength={500}
                  required
                />
              </label>
              {error && <p className="error">{error}</p>}
              <div className="dialog-actions">
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setDialog(false)}
                >
                  Отмена
                </button>
                <button disabled={busy}>
                  {busy ? "Сохраняем…" : "Сохранить"}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </main>
  );
}
