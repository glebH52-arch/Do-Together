import { useEffect, useState, type FormEvent } from "react";
import { api, type ProjectMember, type User, type UUID } from "./api";
import { roleLabel, type Locale } from "./i18n";
export function ProjectMembers({
  projectID,
  user,
  accessToken,
  onMembers,
  locale,
}: {
  projectID: UUID;
  user: User;
  accessToken: string;
  onMembers?: (v: ProjectMember[]) => void;
  locale: Locale;
}) {
  const [items, setItems] = useState<ProjectMember[] | null>(null),
    [error, setError] = useState(""),
    [invite, setInvite] = useState(false),
    [email, setEmail] = useState(""),
    [role, setRole] = useState<"admin" | "member">("member"),
    [busy, setBusy] = useState(false);
  const load = () =>
    api
      .members(accessToken, projectID)
      .then((v) => {
        setItems(v);
        onMembers?.(v);
      })
      .catch((e) => setError(e.message));
  useEffect(() => {
    void load();
  }, [projectID, accessToken]);
  const mine = items?.find((v) => v.user_id === user.id)?.role,
    can = mine === "creator" || mine === "admin";
  async function send(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    try {
      await api.invite(accessToken, projectID, { email: email.trim(), role });
      setInvite(false);
      setEmail("");
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(false);
    }
  }
  async function change(m: ProjectMember, next: "admin" | "member") {
    try {
      await api.memberRole(accessToken, projectID, m.user_id, next);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    }
  }
  async function remove(m: ProjectMember) {
    if (!confirm(`Удалить ${m.username} из проекта?`)) return;
    try {
      await api.removeMember(accessToken, projectID, m.user_id);
      await load();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    }
  }
  return (
    <section className="members-section">
      <div className="section-heading">
        <div>
          <h2>Участники</h2>
          <p className="muted">Имена, роли и доступные действия</p>
        </div>
        {can && <button onClick={() => setInvite(true)}>Пригласить</button>}
      </div>
      {error && <p className="error">{error}</p>}
      {items === null ? (
        <p>Загружаем…</p>
      ) : (
        <div className="members-list">
          {items.map((m) => (
            <article className="member-card" key={m.user_id}>
              {m.avatar_url?<img className="member-avatar" src={m.avatar_url} alt=""/>:<span className="member-avatar">{(m.display_name||m.username)[0]?.toUpperCase()}</span>}
              <div className="member-info">
                <h3>
                  {m.display_name||m.username}
                  {m.user_id === user.id ? " · Вы" : ""}
                </h3>
                <p>{m.email}</p>
              </div>
              <div className="member-controls">
                <b>{roleLabel(m.role,locale)}</b>
                {can && m.role !== "creator" && m.user_id !== user.id && (
                  <>
                    <select
                      aria-label={`Роль ${m.username}`}
                      value={m.role}
                      onChange={(e) =>
                        void change(m, e.target.value as "admin" | "member")
                      }
                    >
                      <option value="member">Участник</option>
                      <option value="admin">Модератор</option>
                    </select>
                    <button
                      className="delete-link"
                      onClick={() => void remove(m)}
                    >
                      Удалить
                    </button>
                  </>
                )}
              </div>
            </article>
          ))}
        </div>
      )}
      {invite && (
        <div className="dialog-backdrop">
          <div className="dialog-card">
            <h2>Пригласить участника</h2>
            <form onSubmit={send}>
              <label>
                Email
                <input
                  autoFocus
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </label>
              <label>
                Роль
                <select
                  value={role}
                  onChange={(e) => setRole(e.target.value as any)}
                >
                  <option value="member">Участник</option>
                  <option value="admin">Администратор</option>
                </select>
              </label>
              <p className="muted">
                Пользователь должен зарегистрироваться и подтвердить email.
              </p>
              <div className="dialog-actions">
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setInvite(false)}
                >
                  Отмена
                </button>
                <button disabled={busy}>Отправить</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </section>
  );
}
