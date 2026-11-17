import { useEffect, useState } from "react";
import { api, type Invite, type UUID } from "./api";
export function Invites({
  accessToken,
  onChanged,
  onOpenProject,
  onBack,
  revision,
}: {
  accessToken: string;
  onChanged: () => void;
  onOpenProject: (id: UUID) => void;
  onBack: () => void;
  revision: number;
}) {
  const [items, setItems] = useState<Invite[] | null>(null),
    [error, setError] = useState(""),
    [busy, setBusy] = useState<UUID | null>(null);
  const load = () =>
    api
      .invites(accessToken)
      .then(setItems)
      .catch((e) => setError(e.message));
  useEffect(() => {
    void load();
  }, [accessToken, revision]);
  async function decide(i: Invite, a: "accept" | "decline") {
    setBusy(i.id);
    setError("");
    try {
      await api.decideInvite(accessToken, i.id, a);
      await load();
      onChanged();
      if (a === "accept") onOpenProject(i.project_id);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Ошибка");
    } finally {
      setBusy(null);
    }
  }
  return (
    <main className="workspace-main">
      <button className="back-link" onClick={onBack}>
        ← Все проекты
      </button>
      <h1>Приглашения</h1>
      <p className="muted">Входящие приглашения и история решений</p>
      {error && <p className="error">{error}</p>}
      {items === null ? (
        <p>Загружаем…</p>
      ) : items.length === 0 ? (
        <div className="state-card">Приглашений нет</div>
      ) : (
        <div className="invite-list">
          {items.map((i) => (
            <article className="invite-card" key={i.id}>
              <h2>{i.project_title}</h2>
              <p>Отправитель: {i.inviter_name}</p>
              <p>Роль: {i.role === "admin" ? "Администратор" : "Участник"}</p>
              <p>Статус: {i.status}</p>
              {i.status === "pending" &&
                new Date(i.expires_at) > new Date() && (
                  <div className="inline-actions">
                    <button
                      disabled={!!busy}
                      onClick={() => void decide(i, "accept")}
                    >
                      {busy === i.id ? "Подождите…" : "Принять"}
                    </button>
                    <button
                      className="secondary"
                      disabled={!!busy}
                      onClick={() => void decide(i, "decline")}
                    >
                      Отклонить
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
