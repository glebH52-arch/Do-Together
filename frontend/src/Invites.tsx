import { useEffect, useState } from "react";
import { api, type Invite, type UUID } from "./api";
import { chromeCopy, localizedRequestError, roleLabel, useCopy, type Locale } from "./i18n";
export function Invites({
  accessToken,
  onChanged,
  onOpenProject,
  onBack,
  revision,
  locale = "ru",
}: {
  accessToken: string;
  onChanged: () => void;
  onOpenProject: (id: UUID) => void;
  onBack: () => void;
  revision: number;
  locale?: Locale;
}) {
  const t=chromeCopy[locale], common=useCopy(locale);
  const [items, setItems] = useState<Invite[] | null>(null),
    [error, setError] = useState<unknown>(null),
    [busy, setBusy] = useState<UUID | null>(null);
  const load = () =>
    api
      .invites(accessToken)
      .then(setItems)
      .catch((e) => setError(e));
  useEffect(() => {
    void load();
  }, [accessToken, revision]);
  async function decide(i: Invite, a: "accept" | "decline") {
    setBusy(i.id);
    setError(null);
    try {
      await api.decideInvite(accessToken, i.id, a);
      await load();
      onChanged();
      if (a === "accept") onOpenProject(i.project_id);
    } catch (e) {
      setError(e);
    } finally {
      setBusy(null);
    }
  }
  return (
    <main className="workspace-main">
      <button className="back-link" onClick={onBack}>
        ← {t.allProjects}
      </button>
      <h1>{common.invites}</h1>
      <p className="muted">{t.invitesIntro}</p>
      {!!error && <p className="error" role="alert">{localizedRequestError(error,locale)}</p>}
      {items === null ? (
        <p>{t.loading}</p>
      ) : items.length === 0 ? (
        <div className="state-card">{t.noInvites}</div>
      ) : (
        <div className="invite-list">
          {items.map((i) => (
            <article className="invite-card" key={i.id}>
              <h2>{i.project_title}</h2>
              <p>{t.sender}: {i.inviter_name}</p>
              <p>{t.role}: {roleLabel(i.role,locale)}</p>
              <p>{t.status}: {t[i.status]}</p>
              {i.status === "pending" &&
                new Date(i.expires_at) > new Date() && (
                  <div className="inline-actions">
                    <button
                      disabled={!!busy}
                      onClick={() => void decide(i, "accept")}
                    >
                      {busy === i.id ? t.wait : common.accept}
                    </button>
                    <button
                      className="secondary"
                      disabled={!!busy}
                      onClick={() => void decide(i, "decline")}
                    >
                      {common.decline}
                    </button>
                  </div>
                )}
            </article>
          ))}
        </div>
      )}
    </main>
  );
}
