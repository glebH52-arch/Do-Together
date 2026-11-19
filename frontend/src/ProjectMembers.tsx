import { useEffect, useState, type FormEvent } from "react";
import { ApiError, api, type ProjectMember, type User, type UUID } from "./api";
import { chromeCopy, localizedRequestError, roleLabel, type Locale } from "./i18n";
import { Avatar } from "./Avatar";
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
  const t = chromeCopy[locale];
  const [items, setItems] = useState<ProjectMember[] | null>(null),
    [error, setError] = useState<unknown>(null),
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
      .catch((e) => setError(e));
  useEffect(() => {
    void load();
  }, [projectID, accessToken, user.avatar_url, user.display_name]);
  const mine = items?.find((v) => v.user_id === user.id)?.role,
    can = mine === "creator" || mine === "admin";
  async function send(e: FormEvent) {
    e.preventDefault();
    setBusy(true);
    setError(null);
    try {
      await api.invite(accessToken, projectID, { email: email.trim(), role });
      setInvite(false);
      setEmail("");
      await load();
    } catch (e) {
      setError(e);
    } finally {
      setBusy(false);
    }
  }
  async function change(m: ProjectMember, next: "admin" | "member") {
    setError(null);
    try {
      await api.memberRole(accessToken, projectID, m.user_id, next);
      await load();
    } catch (e) {
      setError(e);
    }
  }
  async function remove(m: ProjectMember) {
    if (!confirm(t.removeMember.replace("{name}",m.username))) return;
    setError(null);
    try {
      await api.removeMember(accessToken, projectID, m.user_id);
      await load();
    } catch (e) {
      setError(e);
    }
  }
  return (
    <section className="members-section">
      <div className="section-heading">
        <div>
          <h2>{t.members}</h2>
          <p className="muted">{t.membersIntro}</p>
        </div>
        {can && <button onClick={() => {setError(null);setInvite(true)}}>{t.invite}</button>}
      </div>
      {!!error && !invite && <p className="error" role="alert">{localizedRequestError(error,locale)}</p>}
      {items === null ? (
        <p>{t.loading}</p>
      ) : (
        <div className="members-list">
          {items.map((m) => (
            <article className="member-card" key={m.user_id}>
              <Avatar user={m} className="member-avatar"/>
              <div className="member-info">
                <h3>
                  {m.display_name||m.username}
                  {m.user_id === user.id ? ` · ${t.you}` : ""}
                </h3>
                <p>{m.email}</p>
              </div>
              <div className="member-controls">
                <b>{roleLabel(m.role,locale)}</b>
                {can && m.role !== "creator" && m.user_id !== user.id && (
                  <>
                    <select
                      aria-label={`${t.role} ${m.username}`}
                      value={m.role}
                      onChange={(e) =>
                        void change(m, e.target.value as "admin" | "member")
                      }
                    >
                      <option value="member">{roleLabel("member",locale)}</option>
                      <option value="admin">{roleLabel("admin",locale)}</option>
                    </select>
                    <button
                      className="delete-link"
                      onClick={() => void remove(m)}
                    >
                      {t.remove}
                    </button>
                  </>
                )}
              </div>
            </article>
          ))}
        </div>
      )}
      {invite && (
        <div className="dialog-backdrop" onClick={e=>{if(e.target===e.currentTarget)setInvite(false)}} onKeyDown={e=>{if(e.key==="Escape"){e.stopPropagation();setInvite(false)}}}>
          <div className="dialog-card" role="dialog" aria-modal="true" aria-label={t.inviteMember}>
            <h2>{t.inviteMember}</h2>
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
                {t.role}
                <select
                  value={role}
                  onChange={(e) => setRole(e.target.value as any)}
                >
                  <option value="member">{roleLabel("member",locale)}</option>
                  <option value="admin">{roleLabel("admin",locale)}</option>
                </select>
              </label>
              <p className="muted">
                {t.inviteNote}
              </p>
              {!!error && <p className="error" role="alert">{error instanceof ApiError && error.status===404 ? t.inviteMissing : error instanceof ApiError && error.status===409 ? t.inviteExists : error instanceof ApiError && error.status===403 ? t.inviteForbidden : localizedRequestError(error,locale)}</p>}
              <div className="dialog-actions">
                <button
                  type="button"
                  className="secondary"
                  onClick={() => setInvite(false)}
                >
                  {t.cancel}
                </button>
                <button disabled={busy}>{busy?t.sending:t.send}</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </section>
  );
}
