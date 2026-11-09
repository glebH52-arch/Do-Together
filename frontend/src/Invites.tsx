import { useState } from 'react';
import { acceptInvite, declineInvite, type Invite } from './api';

export type InviteListState = { kind: 'loading' } | { kind: 'error' } | { kind: 'ready'; invites: Invite[] };

const roleLabels = { admin: 'Администратор', member: 'Участник' } as const;
const statusLabels = { pending: 'Ожидает ответа', accepted: 'Принято', declined: 'Отклонено', expired: 'Истекло' } as const;
const date = (value: string) => new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));

export function Invites({ accessToken, state, onReload, onOpenProject, onBack }: {
  accessToken: string;
  state: InviteListState;
  onReload: () => void;
  onOpenProject: (projectID: number) => void;
  onBack: () => void;
}) {
  const [busyID, setBusyID] = useState<number | null>(null);
  const [error, setError] = useState<{ id: number; message: string } | null>(null);
  const [notice, setNotice] = useState<{ message: string; projectID?: number } | null>(null);

  async function decide(invite: Invite, action: 'accept' | 'decline') {
    if (busyID !== null) return;
    setBusyID(invite.id);
    setError(null);
    setNotice(null);
    try {
      if (action === 'accept') await acceptInvite(invite.id, accessToken);
      else await declineInvite(invite.id, accessToken);
      setNotice({ message: action === 'accept' ? 'Приглашение принято' : 'Приглашение отклонено', projectID: action === 'accept' ? invite.project_id : undefined });
      onReload();
    } catch (cause) {
      setError({ id: invite.id, message: cause instanceof Error ? cause.message : 'Не удалось обработать приглашение.' });
    } finally { setBusyID(null); }
  }

  const pending = state.kind === 'ready' ? state.invites.filter(invite => invite.status === 'pending') : [];
  const history = state.kind === 'ready' ? state.invites.filter(invite => invite.status !== 'pending') : [];

  return <main className="workspace-main">
    <button className="back-link" onClick={onBack}>Все проекты ←</button>
    <div className="section-heading"><div><h1>Приглашения</h1><p className="muted">Проекты, в которые вас пригласили</p></div></div>
    {notice && <div className="notice" role="status">{notice.message}{notice.projectID && <button className="notice-action" onClick={() => onOpenProject(notice.projectID!)}>Открыть проект</button>}</div>}
    {state.kind === 'loading' && <p role="status">Загружаем приглашения…</p>}
    {state.kind === 'error' && <div className="state-card" role="alert"><h2>Не удалось загрузить приглашения</h2><button onClick={onReload}>Повторить</button></div>}
    {state.kind === 'ready' && <>
      {pending.length === 0 && <div className="state-card"><h2>Новых приглашений нет</h2></div>}
      {pending.length > 0 && <div className="invite-list">{pending.map(invite => {
        const expired = new Date(invite.expires_at).getTime() <= Date.now();
        return <article className="invite-card" key={invite.id}>
          <div><h2>Проект #{invite.project_id}</h2><p>От: Пользователь #{invite.inviter_id}</p><p>Роль: {roleLabels[invite.role]}</p><p>Приглашение: {date(invite.created_at)}</p><p className={expired ? 'overdue' : ''}>Действует до: {date(invite.expires_at)}{expired ? ' · срок истёк' : ''}</p></div>
          {error?.id === invite.id && <p className="error" role="alert">{error.message}</p>}
          {!expired && <div className="invite-actions"><button disabled={busyID !== null} onClick={() => void decide(invite, 'accept')}>{busyID === invite.id ? 'Подождите…' : 'Принять'}</button><button className="secondary" disabled={busyID !== null} onClick={() => void decide(invite, 'decline')}>Отклонить</button></div>}
        </article>;
      })}</div>}
      {history.length > 0 && <section className="invite-history"><h2>История приглашений</h2><div className="invite-list">{history.map(invite => <article className="invite-card" key={invite.id}><h3>Проект #{invite.project_id}</h3><p>От: Пользователь #{invite.inviter_id}</p><p>Статус: {statusLabels[invite.status]}</p></article>)}</div></section>}
    </>}
  </main>;
}
