import { useEffect, useState, type FormEvent } from 'react';
import { ApiError, createInvite, listProjectMembers, removeMember, updateMemberRole, type ProjectMember, type User } from './api';

type MembersState = { kind: 'loading' } | { kind: 'error' } | { kind: 'denied' } | { kind: 'ready'; members: ProjectMember[] };
const roleLabels = { creator: 'Создатель', admin: 'Администратор', member: 'Участник' } as const;
const date = (value: string) => new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium' }).format(new Date(value));

export function ProjectMembers({ projectID, user, accessToken }: { projectID: number; user: User; accessToken: string }) {
  const [state, setState] = useState<MembersState>({ kind: 'loading' });
  const [reload, setReload] = useState(0);
  const [busyID, setBusyID] = useState<number | null>(null);
  const [rowError, setRowError] = useState<{ id: number; message: string } | null>(null);
  const [confirmID, setConfirmID] = useState<number | null>(null);
  const [inviteOpen, setInviteOpen] = useState(false);
  const [inviteeID, setInviteeID] = useState('');
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviteRole, setInviteRole] = useState<'admin' | 'member'>('member');
  const [inviteBusy, setInviteBusy] = useState(false);
  const [inviteError, setInviteError] = useState('');
  const [notice, setNotice] = useState('');

  useEffect(() => {
    const controller = new AbortController();
    setState({ kind: 'loading' });
    void listProjectMembers(projectID, accessToken, controller.signal)
      .then(members => setState({ kind: 'ready', members }))
      .catch(cause => {
        if (controller.signal.aborted) return;
        setState(cause instanceof ApiError && (cause.status === 403 || cause.status === 404) ? { kind: 'denied' } : { kind: 'error' });
      });
    return () => controller.abort();
  }, [projectID, accessToken, reload]);

  const ownRole = state.kind === 'ready' ? state.members.find(member => member.user_id === user.id)?.role : undefined;
  const canManage = ownRole === 'creator' || ownRole === 'admin';

  async function changeRole(member: ProjectMember, role: 'admin' | 'member') {
    if (!canManage || busyID !== null || inviteBusy || member.user_id === user.id || member.role === 'creator') return;
    setBusyID(member.user_id); setRowError(null);
    try {
      await updateMemberRole(projectID, member.user_id, role, accessToken);
      setState(current => current.kind === 'ready' ? { kind: 'ready', members: current.members.map(item => item.user_id === member.user_id ? { ...item, role } : item) } : current);
      try {
        const members = await listProjectMembers(projectID, accessToken);
        setState({ kind: 'ready', members });
      } catch { setRowError({ id: member.user_id, message: 'Роль сохранена, но список не обновился. Обновите страницу.' }); }
    } catch (cause) { setRowError({ id: member.user_id, message: cause instanceof Error ? cause.message : 'Не удалось изменить роль.' }); }
    finally { setBusyID(null); }
  }

  async function deleteMember(member: ProjectMember) {
    if (!canManage || busyID !== null || member.user_id === user.id || member.role === 'creator') return;
    setBusyID(member.user_id); setRowError(null);
    try {
      await removeMember(projectID, member.user_id, accessToken);
      setState(current => current.kind === 'ready' ? { kind: 'ready', members: current.members.filter(item => item.user_id !== member.user_id) } : current);
      setConfirmID(null);
      try {
        const members = await listProjectMembers(projectID, accessToken);
        setState({ kind: 'ready', members });
      } catch { setNotice('Участник удалён, но список не обновился. Обновите страницу.'); }
    } catch (cause) { setRowError({ id: member.user_id, message: cause instanceof Error ? cause.message : 'Не удалось удалить участника.' }); }
    finally { setBusyID(null); }
  }

  async function sendInvite(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (inviteBusy) return;
    const email = inviteEmail.trim();
    const id = Number(inviteeID);
    if ((email !== '' && inviteeID !== '') || (email === '' && inviteeID === '')) { setInviteError('Укажите email или ID пользователя.'); return; }
    if (email === '' && (!Number.isSafeInteger(id) || id <= 0)) { setInviteError('Введите положительный ID пользователя.'); return; }
    setInviteBusy(true); setInviteError(''); setNotice('');
    try {
      await createInvite(projectID, email !== '' ? { email, role: inviteRole } : { invitee_id: id, role: inviteRole }, accessToken);
      setInviteOpen(false); setInviteeID(''); setInviteEmail(''); setInviteRole('member');
      setNotice('Приглашение отправлено');
      setReload(value => value + 1);
    } catch (cause) { setInviteError(cause instanceof Error ? cause.message : 'Не удалось отправить приглашение.'); }
    finally { setInviteBusy(false); }
  }

  return <section className="members-section" aria-labelledby="members-title">
    <div className="section-heading"><div><h2 id="members-title">Участники проекта</h2><p className="muted">Люди, у которых есть доступ к проекту</p></div>
      {canManage && <button onClick={() => { setInviteOpen(true); setInviteError(''); }}>Пригласить участника</button>}
    </div>
    {notice && <p className="notice" role="status">{notice}</p>}
    {state.kind === 'loading' && <p role="status">Загружаем участников…</p>}
    {state.kind === 'denied' && <div className="state-card" role="alert"><h3>Нет доступа к участникам проекта</h3></div>}
    {state.kind === 'error' && <div className="state-card" role="alert"><h3>Не удалось загрузить участников</h3><button onClick={() => setReload(value => value + 1)}>Повторить</button></div>}
    {state.kind === 'ready' && (state.members.length === 0 ? <div className="state-card">Пока нет участников</div> : <div className="members-list">{state.members.map(member => <article className={`member-card ${member.role === 'creator' ? 'member-creator' : ''}`} key={member.user_id}>
      <span className="member-avatar" aria-hidden="true">{member.username.slice(0, 1).toUpperCase()}</span>
      <div className="member-info"><h3>{member.username}{member.user_id === user.id ? ' · Вы' : ''}</h3><p>{member.email}</p><p>Присоединился: {date(member.joined_at)}</p></div>
      <div className="member-controls"><span className="member-role">{roleLabels[member.role]}</span>
        {canManage && member.role !== 'creator' && member.user_id !== user.id && <>
          <label className="role-select">Изменить роль<select value={member.role} disabled={busyID !== null || inviteBusy} onChange={event => void changeRole(member, event.target.value as 'admin' | 'member')}><option value="member">Участник</option><option value="admin">Администратор</option></select></label>
          {confirmID === member.user_id ? <div className="member-confirm"><span>Удалить участника из проекта?</span><button className="secondary" disabled={busyID !== null} onClick={() => setConfirmID(null)}>Отмена</button><button className="danger" disabled={busyID !== null} onClick={() => void deleteMember(member)}>Удалить</button></div>
            : <button className="delete-link" disabled={busyID !== null} onClick={() => { setConfirmID(member.user_id); setRowError(null); }}>Удалить участника</button>}
        </>}
        {rowError?.id === member.user_id && <p className="error" role="alert">{rowError.message}</p>}
      </div>
    </article>)}</div>)}
    {inviteOpen && <div className="dialog-backdrop"><div className="dialog-card" role="dialog" aria-modal="true" aria-labelledby="invite-dialog-title"><h2 id="invite-dialog-title">Пригласить участника</h2><form onSubmit={event => void sendInvite(event)}>
      <label>Email зарегистрированного пользователя<input autoFocus type="email" value={inviteEmail} onChange={event => setInviteEmail(event.target.value)} disabled={inviteeID !== ''} /></label>
      <label>Или ID пользователя<input type="number" min="1" step="1" value={inviteeID} onChange={event => setInviteeID(event.target.value)} disabled={inviteEmail !== ''} /></label>
      <label>Роль<select value={inviteRole} onChange={event => setInviteRole(event.target.value as 'admin' | 'member')}><option value="member">Участник</option><option value="admin">Администратор</option></select></label>
      <p className="muted">Пригласить можно зарегистрированного пользователя. Подтверждение email пока не требуется.</p>
      {inviteError && <p className="error" role="alert">{inviteError}</p>}
      <div className="dialog-actions"><button type="button" className="secondary" disabled={inviteBusy} onClick={() => setInviteOpen(false)}>Отмена</button><button type="submit" disabled={inviteBusy}>{inviteBusy ? 'Отправляем…' : 'Отправить приглашение'}</button></div>
    </form></div></div>}
  </section>;
}
