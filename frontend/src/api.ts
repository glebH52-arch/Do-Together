export type Tokens = {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
};

export type User = {
  id: number;
  username: string;
  email: string;
};

export type Project = {
  id: number;
  created_by: number;
  title: string;
  goal: string;
  status: 'active' | 'archived';
  created_at: string;
  updated_at?: string;
};

export type Task = {
  id: number;
  title: string;
  description: string;
  created_by: number;
  status: 'todo' | 'in_progress' | 'review' | 'done';
  created_at: string;
  update_at?: string;
  due_at?: string | null;
};

export type Invite = {
  id: number;
  project_id: number;
  inviter_id: number;
  inviter_name?: string;
  project_title?: string;
  invitee_id: number;
  role: 'admin' | 'member';
  status: 'pending' | 'accepted' | 'declined' | 'expired';
  expires_at: string;
  created_at: string;
  updated_at?: string;
};

export type ProjectMember = {
  username: string;
  email: string;
  user_id: number;
  role: 'creator' | 'admin' | 'member';
  joined_at: string;
};

export type CreateInviteRequest = { invitee_id: number; role: 'admin' | 'member' } | { email: string; role: 'admin' | 'member' };

export const refreshKey = 'do-together-demo-refresh';
let session: Tokens | null = null;
let sessionVersion = 0;
let refreshPromise: Promise<Tokens> | null = null;
const listeners = new Set<(tokens: Tokens | null) => void>();

export function setSession(tokens: Tokens | null) {
  sessionVersion += 1;
  session = tokens;
  if (tokens) sessionStorage.setItem(refreshKey, tokens.refresh_token);
  else sessionStorage.removeItem(refreshKey);
  listeners.forEach(listener => listener(tokens));
}

export function subscribeSession(listener: (tokens: Tokens | null) => void) {
  listeners.add(listener);
  return () => { listeners.delete(listener); };
}

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

function responseError(status: number) {
  if (status === 400) return new ApiError('Проверьте введённые данные.', status);
  if (status === 401) return new ApiError('Сессия завершена. Войдите снова.', status);
  if (status === 403) return new ApiError('Нет доступа к этому действию.', status);
  if (status === 404) return new ApiError('Данные не найдены или нет доступа.', status);
  if (status === 409) return new ApiError('Приглашение уже обработано, истекло или есть конфликт данных.', status);
  if (status === 503) return new ApiError('Сервис временно недоступен.', status);
  return new ApiError('Запрос не выполнен. Код ' + status + '.', status);
}

async function refreshSession(): Promise<Tokens> {
  if (!refreshPromise) {
    const refreshToken = session?.refresh_token;
    if (!refreshToken) throw responseError(401);
    const version = sessionVersion;
    refreshPromise = (async () => {
      const response = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      if (!response.ok) {
        if (response.status === 401 && sessionVersion === version) setSession(null);
        throw responseError(response.status);
      }
      const next = await response.json() as Tokens;
      if (sessionVersion !== version) throw responseError(401);
      setSession(next);
      return next;
    })().finally(() => { refreshPromise = null; });
  }
  return refreshPromise;
}

export async function request<T>(
  path: string,
  body?: object,
  accessToken?: string,
  signal?: AbortSignal,
  method?: 'GET' | 'POST' | 'PATCH' | 'DELETE',
): Promise<T> {
  const send = (token?: string) => fetch('/api' + path, {
    method: method ?? (body ? 'POST' : 'GET'),
    headers: {
      ...(body ? { 'Content-Type': 'application/json' } : {}),
      ...(token ? { Authorization: 'Bearer ' + token } : {}),
    },
    body: body ? JSON.stringify(body) : undefined,
    signal,
  });
  const protectedRequest = Boolean(accessToken) && !path.startsWith('/auth/');
  const token = protectedRequest ? session?.access_token ?? accessToken : accessToken;
  let response = await send(token);
  if (response.status === 401 && protectedRequest) {
    const nextToken = session?.access_token !== token
      ? session?.access_token
      : (await refreshSession()).access_token;
    response = await send(nextToken);
    if (response.status === 401 && session?.access_token === nextToken) setSession(null);
  }
  if (!response.ok) {
    if (path === '/auth/login' && response.status === 401) throw new ApiError('Неверный email или пароль.', 401);
    throw responseError(response.status);
  }
  return response.status === 204 ? (undefined as T) : response.json() as Promise<T>;
}

export const listInvites = (accessToken: string, signal?: AbortSignal) => request<Invite[]>('/invites', undefined, accessToken, signal);
export const acceptInvite = (id: number, accessToken: string) => request<void>(`/invites/${id}/accept`, undefined, accessToken, undefined, 'POST');
export const declineInvite = (id: number, accessToken: string) => request<void>(`/invites/${id}/decline`, undefined, accessToken, undefined, 'POST');
export const listProjectMembers = (projectID: number, accessToken: string, signal?: AbortSignal) => request<ProjectMember[]>(`/projects/${projectID}/members`, undefined, accessToken, signal);
export const createInvite = (projectID: number, data: CreateInviteRequest, accessToken: string) => request<Invite>(`/projects/${projectID}/invites`, data, accessToken);
export const updateMemberRole = (projectID: number, userID: number, role: 'admin' | 'member', accessToken: string) => request<void>(`/projects/${projectID}/members/${userID}`, { role }, accessToken, undefined, 'PATCH');
export const removeMember = (projectID: number, userID: number, accessToken: string) => request<void>(`/projects/${projectID}/members/${userID}`, undefined, accessToken, undefined, 'DELETE');
