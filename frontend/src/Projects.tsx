import { useEffect, useState, type FormEvent } from "react";
import { api, type Project, type User, type UUID } from "./api";
import { chromeCopy, localizedRequestError, type Locale } from "./i18n";
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
  const t = chromeCopy[locale];
  const [items, setItems] = useState<Project[] | null>(null),
    [error, setError] = useState<unknown>(null),
    [dialog, setDialog] = useState<false | "create" | Project>(false),
    [title, setTitle] = useState(""),
    [goal, setGoal] = useState(""),
    [busy, setBusy] = useState(false);
  const load = () =>
    Promise.all([api.projects(accessToken,"active"),api.projects(accessToken,"archived")])
      .then(([active,archived])=>setItems([...active,...archived]))
      .catch((e) => setError(e));
  useEffect(() => {
    void load();
  }, [accessToken,revision]);
  function open(v: "create" | Project) {
    setDialog(v);
    setTitle(v === "create" ? "" : v.title);
    setGoal(v === "create" ? "" : v.goal);
    setError(null);
  }
  async function submit(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
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
      setError(e);
    } finally {
      setBusy(false);
    }
  }
  async function archive(p: Project) {
    if (!confirm(t.archiveConfirm.replace("{name}",p.title))) return;
    try {
      await api.archiveProject(accessToken, p.id);
      await load();
    } catch (e) {
      setError(e);
    }
  }
  async function restore(p: Project) { try { await api.restoreProject(accessToken,p.id); await load(); } catch(e){ setError(e) } }
  const active=(items??[]).filter(p=>p.status==="active"), archived=(items??[]).filter(p=>p.status==="archived");
  return (
    <main className="workspace-main">
      <div className="section-heading">
        <div>
          <h1>{t.myProjects}</h1>
          <p className="muted">
            {t.projectsIntro}
          </p>
        </div>
        <button onClick={() => open("create")}>+ {t.createProject}</button>
      </div>
      {!!error && <p className="error">{localizedRequestError(error, locale)}</p>}
      {items === null ? (
        <p>{t.loadingProjects}</p>
      ) : items.length === 0 ? (
        <div className="state-card">
          <svg className="empty-illustration" viewBox="0 0 180 120" role="img" aria-label={t.emptyBoard}><rect x="20" y="18" width="140" height="84" rx="10" fill="none" stroke="currentColor" strokeWidth="4"/><path d="M40 39h35M40 54h24M91 39h48M91 54h34M40 77h54" fill="none" stroke="currentColor" strokeWidth="7" strokeLinecap="round"/><circle cx="130" cy="78" r="12" fill="var(--accent)"/><path d="M130 72v12m-6-6h12" stroke="white" strokeWidth="3" strokeLinecap="round"/></svg>
          <h2>{t.noProjects}</h2>
          <p>{t.firstProject}</p>
        </div>
      ) : (
        <><h2 className="list-heading">{t.active}</h2><div className="project-grid">
          {active.map((p) => (
            <article className="project-card" key={p.id}>
              <button
                className="project-link linklike"
                onClick={() => onOpen(p.id)}
              >
                <div className="project-card-top">
                  <h2>{p.title}</h2>
                  <span className="status">
                    {p.status === "archived" ? t.archived : t.activeStatus}
                  </span>
                </div>
                <p>{p.goal}</p>
              </button>
              {p.created_by === user.id && p.status === "active" && (
                <div className="inline-actions">
                  <button className="secondary" onClick={() => open(p)}>
                    {t.edit}
                  </button>
                  <button className="danger" onClick={() => void archive(p)}>
                    {t.archiveAction}
                  </button>
                </div>
              )}
            </article>
          ))}
        </div><h2 className="list-heading">{t.archive}</h2>{archived.length===0?<p className="muted">{t.archiveEmpty}</p>:<div className="project-grid">{archived.map(p=><article className="project-card" key={p.id}><button className="project-link linklike" onClick={()=>onOpen(p.id)}><div className="project-card-top"><h2>{p.title}</h2><span className="status">{t.archived}</span></div><p>{p.goal}</p></button>{p.created_by===user.id&&<button className="secondary" onClick={()=>void restore(p)}>{t.restore}</button>}</article>)}</div>}</>
      )}
      {dialog && (
        <div className="dialog-backdrop">
          <div className="dialog-card">
            <h2>{dialog === "create" ? t.newProject : t.editProject}</h2>
            <form onSubmit={submit}>
              <label>
                {t.title}
                <input
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  maxLength={100}
                  required
                />
              </label>
              <label>
                {t.goal}
                <textarea
                  value={goal}
                  onChange={(e) => setGoal(e.target.value)}
                  maxLength={500}
                  required
                />
              </label>
              {!!error && <p className="error">{localizedRequestError(error, locale)}</p>}
              <div className="dialog-actions">
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setDialog(false)}
                >
                  {t.cancel}
                </button>
                <button disabled={busy}>
                  {busy ? t.saving : t.save}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </main>
  );
}
