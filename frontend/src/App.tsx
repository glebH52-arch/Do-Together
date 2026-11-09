import { useEffect, useState, type FormEvent } from 'react';
import { listInvites, refreshKey, request, setSession, subscribeSession, type Invite, type Tokens, type User } from './api';
import { Projects } from './Projects';
import { ProjectDetail } from './ProjectDetail';
import { Invites, type InviteListState } from './Invites';

export function App() {
  const [mode, setMode] = useState<'login' | 'register'>('login');
  const [email, setEmail] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [user, setUser] = useState<User | null>(null);
  const [tokens, setTokens] = useState<Tokens | null>(null);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [pathname, setPathname] = useState(window.location.pathname);
  const [inviteList, setInviteList] = useState<InviteListState>({ kind: 'loading' });
  const [inviteReload, setInviteReload] = useState(0);

  useEffect(() => subscribeSession(next => {
    setTokens(next);
    if (!next) {
      setUser(null);
      setInviteList({ kind: 'loading' });
    }
  }), []);

  useEffect(() => {
    if (!user || !tokens) return;
    const controller = new AbortController();
    setInviteList({ kind: 'loading' });
    void listInvites(tokens.access_token, controller.signal)
      .then(invites => setInviteList({ kind: 'ready', invites }))
      .catch(() => { if (!controller.signal.aborted) setInviteList({ kind: 'error' }); });
    return () => controller.abort();
  }, [user?.id, tokens?.access_token, inviteReload]);

  useEffect(() => {
    const onPopState = () => setPathname(window.location.pathname);
    window.addEventListener('popstate', onPopState);
    return () => window.removeEventListener('popstate', onPopState);
  }, []);

  function navigate(path: string) {
    if (window.location.pathname !== path) window.history.pushState(null, '', path);
    setPathname(path);
  }

  useEffect(() => {
    const refreshToken = sessionStorage.getItem(refreshKey);
    if (!refreshToken) {
      setLoading(false);
      return;
    }
    void (async () => {
      try {
        const next = await request<Tokens>('/auth/refresh', { refresh_token: refreshToken });
        setSession(next);
        const profile = await request<User>('/users/me', undefined, next.access_token);
        setUser(profile);
      } catch (cause) {
        if (cause instanceof Error && cause.message.includes('Сессия завершена')) {
          setSession(null);
        } else {
          setError('Не удалось восстановить сессию. Попробуйте обновить страницу.');
        }
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setBusy(true);
    setError('');
    setNotice('');
    try {
      if (mode === 'register') {
        await request<User>('/users', { username: username.trim(), email: email.trim(), password });
        setPassword('');
        setMode('login');
        setNotice('Аккаунт создан. Теперь войдите.');
        return;
      }
      const next = await request<Tokens>('/auth/login', { email: email.trim(), password });
      const profile = await request<User>('/users/me', undefined, next.access_token);
      setSession(next);
      setUser(profile);
      setPassword('');
      if (!/^\/projects\/\d+$/.test(pathname)) navigate('/projects');
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Не удалось выполнить запрос.');
    } finally {
      setBusy(false);
    }
  }

  async function logout() {
    setBusy(true);
    setError('');
    try {
      if (tokens) await request<void>('/auth/logout', { refresh_token: tokens.refresh_token });
    } catch {
      setNotice('Локальная сессия закрыта, но сервер не подтвердил выход.');
    } finally {
      setSession(null);
      setUser(null);
      setInviteList({ kind: 'loading' });
      setBusy(false);
      navigate('/');
    }
  }

  if (loading) return <main className="card"><p>Восстанавливаем сессию…</p></main>;

  if (user && tokens) {
    const match = /^\/projects\/([1-9]\d*)$/.exec(pathname);
    const pendingCount = inviteList.kind === 'ready' ? inviteList.invites.filter((invite: Invite) => invite.status === 'pending').length : null;
    return <div className="workspace">
      <header className="workspace-header">
        <span className="brand">Do Together</span>
        <div className="account">
          <button className="invite-nav" aria-label="Приглашения" title="Приглашения" onClick={() => navigate('/invites')}>
            <span aria-hidden="true">✉</span><span className="invite-count">{pendingCount === null ? '…' : pendingCount}</span>
          </button>
          <span className="account-name">{user.username}</span>
          <button className="secondary" disabled={busy} onClick={() => void logout()}>Выйти</button>
        </div>
      </header>
      {match ? <ProjectDetail key={`${user.id}:${match[1]}`} projectID={Number(match[1])} user={user} accessToken={tokens.access_token} onBack={() => navigate('/projects')} />
        : pathname === '/invites' ? <Invites accessToken={tokens.access_token} state={inviteList} onReload={() => setInviteReload(value => value + 1)} onOpenProject={id => navigate(`/projects/${id}`)} onBack={() => navigate('/projects')} />
          : <Projects key={user.id} accessToken={tokens.access_token} onOpen={id => navigate(`/projects/${id}`)} />}
    </div>;
  }

  return (
    <main className="card">
      <span className="eyebrow">DO TOGETHER</span>
        <>
          <h1>{mode === 'login' ? 'Войти в проект' : 'Создать аккаунт'}</h1>
          <p className="muted">Совместная работа над проектами и задачами.</p>
          <form onSubmit={event => void submit(event)}>
            {mode === 'register' && (
              <label>Имя пользователя
                <input value={username} onChange={event => setUsername(event.target.value)} autoComplete="username" required />
              </label>
            )}
            <label>Email
              <input type="email" value={email} onChange={event => setEmail(event.target.value)} autoComplete="email" required />
            </label>
            <label>Пароль
              <input type="password" value={password} onChange={event => setPassword(event.target.value)} autoComplete={mode === 'login' ? 'current-password' : 'new-password'} required />
            </label>
            <button disabled={busy} type="submit">{busy ? 'Подождите…' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}</button>
          </form>
          <button className="link" disabled={busy} onClick={() => { setMode(mode === 'login' ? 'register' : 'login'); setError(''); setNotice(''); }}>
            {mode === 'login' ? 'Нет аккаунта? Зарегистрироваться' : 'Уже есть аккаунт? Войти'}
          </button>
        </>
      {notice && <p className="notice" role="status">{notice}</p>}
      {error && <p className="error" role="alert">{error}</p>}
    </main>
  );
}
