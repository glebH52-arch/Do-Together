// Use the same public gateway for API calls and backend-hosted media.
export const apiBase = (import.meta.env?.VITE_API_BASE_URL || "/api").replace(/\/$/, "");
export function resolveAvatarURL(value: string | null | undefined, base = apiBase): string | null {
  if (!value) return null;
  if (/^https?:\/\//i.test(value)) return value;
  if (value.startsWith("//") || /^[a-z][a-z\d+.-]*:/i.test(value)) return null;
  const normalizedBase = base.replace(/\/$/, "");
  if (value === normalizedBase || value.startsWith(normalizedBase + "/")) return value;
  return normalizedBase + "/" + value.replace(/^\/+/, "");
}

export type UUID = string;
export type Tokens = {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
};
export type User = {
  id: UUID;
  username: string;
  display_name: string;
  avatar_url: string | null;
  email: string;
  status: string;
};
export type Project = {
  id: UUID;
  created_by: UUID;
  title: string;
  goal: string;
  status: "active" | "archived";
  created_at: string;
  updated_at?: string;
};
export type MemberRole = "creator" | "admin" | "member";
export type ProjectMember = {
  user_id: UUID;
  username: string;
  display_name: string;
  avatar_url: string | null;
  email: string;
  role: MemberRole;
  joined_at: string;
};
export type Invite = {
  id: UUID;
  project_id: UUID;
  inviter_id: UUID;
  inviter_name: string;
  project_title: string;
  invitee_id: UUID;
  role: "admin" | "member";
  status: "pending" | "accepted" | "declined" | "expired";
  expires_at: string;
  created_at: string;
  updated_at?: string;
};
export type TaskStatus = "todo" | "in_progress" | "review" | "done";
export type TaskType = "task" | "bug" | "feature" | "research";
export type TaskPriority = "low" | "medium" | "high" | "critical";
export type Task = {
  id: UUID;
  project_id?: UUID;
  title: string;
  description: string;
  type: TaskType;
  priority: TaskPriority;
  assignee_id: UUID | null;
  assignee_name: string | null;
  assignee_avatar_url: string | null;
  created_by: UUID;
  creator_name: string;
  creator_avatar_url: string | null;
  status: TaskStatus;
  created_at: string;
  update_at?: string;
  due_at: string | null;
  completed_at?: string | null;
  allowed_transitions: TaskStatus[];
};
export type Board = Record<TaskStatus, Task[]>;
export type Paged<T> = {
  items: T[];
  total: number;
  limit: number;
  offset: number;
};
export type Comment = {
  id: UUID;
  task_id: UUID;
  author_id: UUID;
  author_name: string;
  body: string;
  created_at: string;
  updated_at?: string;
};
export type History = {
  id: UUID;
  task_id: UUID;
  actor_id: UUID;
  actor_name: string;
  action: string;
  old_values?: Record<string, unknown>;
  new_values?: Record<string, unknown>;
  created_at: string;
};
export type Notification = {
  id: UUID;
  project_id: UUID;
  task_id?: UUID;
  invite_id?: UUID;
  kind: string;
  payload: Record<string, unknown>;
  created_at: string;
  read_at?: string;
};
export type Activity = {
  id: UUID;
  type: string;
  occurred_at: string;
  actor: { id: UUID; name: string };
  object: { id: UUID; type: string; title: string; deleted: boolean };
  data: { old: Record<string, unknown>; new: Record<string, unknown> };
};
export type ActivityPage = { items: Activity[]; next_cursor: string | null };
export const refreshKey = "do-together-refresh-v2";
let session: Tokens | null = null,
  version = 0,
  refreshing: Promise<Tokens> | null = null;
const listeners = new Set<(v: Tokens | null) => void>();
export function setSession(v: Tokens | null) {
  version++;
  session = v;
  if (v) sessionStorage.setItem(refreshKey, v.refresh_token);
  else sessionStorage.removeItem(refreshKey);
  listeners.forEach((fn) => fn(v));
}
export function subscribeSession(fn: (v: Tokens | null) => void) {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}
export class ApiError extends Error {
  status: number;
  details?: unknown;
  constructor(
    message: string,
    status: number,
    details?: unknown,
  ) {
    super(message);
    this.status = status;
    this.details = details;
  }
}
const fallback: Record<number, string> = {
  400: "Проверьте введённые данные.",
  401: "Сессия завершена. Войдите снова.",
  403: "Нет доступа к этому действию.",
  404: "Данные не найдены или нет доступа.",
  409: "Данные уже изменились. Получите актуальную версию.",
  422: "Действие нарушает правила рабочего процесса.",
  429: "Слишком много попыток. Попробуйте позже.",
  503: "Сервис временно недоступен.",
};
async function responseError(r: Response) {
  let d: any;
  try {
    d = await r.json();
  } catch {}
  return new ApiError(
    typeof d?.message === "string"
      ? d.message
      : (fallback[r.status] ?? `Запрос не выполнен (${r.status}).`),
    r.status,
    d,
  );
}
async function refresh() {
  if (!refreshing) {
    const rt = session?.refresh_token ?? sessionStorage.getItem(refreshKey);
    if (!rt) throw new ApiError(fallback[401], 401);
    const own = version;
    refreshing = fetch(apiBase + "/auth/refresh", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: rt }),
    })
      .then(async (r) => {
        if (!r.ok) throw await responseError(r);
        const v = (await r.json()) as Tokens;
        if (version !== own && session) throw new ApiError(fallback[401], 401);
        setSession(v);
        return v;
      })
      .catch((e) => {
        if (version === own) setSession(null);
        throw e;
      })
      .finally(() => {
        refreshing = null;
      });
  }
  return refreshing;
}
export async function request<T>(
  path: string,
  o: {
    method?: string;
    body?: unknown;
    token?: string;
    signal?: AbortSignal;
    query?: Record<string, string | number | undefined | null>;
  } = {},
): Promise<T> {
  const q = new URLSearchParams();
  Object.entries(o.query ?? {}).forEach(([k, v]) => {
    if (v !== undefined && v !== null && v !== "") q.set(k, String(v));
  });
  const url = apiBase + path + (q.size ? "?" + q : "");
  const guarded = !!o.token && !path.startsWith("/auth/");
  let token = guarded ? (session?.access_token ?? o.token) : o.token;
  const send = (t?: string) =>
    fetch(url, {
      method: o.method ?? (o.body === undefined ? "GET" : "POST"),
      headers: {
        ...(o.body === undefined ? {} : { "Content-Type": "application/json" }),
        ...(t ? { Authorization: `Bearer ${t}` } : {}),
      },
      body: o.body === undefined ? undefined : JSON.stringify(o.body),
      signal: o.signal,
    });
  let r = await send(token);
  if (r.status === 401 && guarded) {
    try {
      token = (await refresh()).access_token;
      r = await send(token);
    } catch (e) {
      setSession(null);
      throw e;
    }
  }
  if (!r.ok) throw await responseError(r);
  if (r.status === 204) return undefined as T;
  const data = await r.json();
  if (Array.isArray(data) && (path.includes("/search") || path.includes("/comments") || path.includes("/history") || path === "/notifications")) {
    return { items: data, total: data.length, limit: Number(o.query?.limit ?? data.length), offset: Number(o.query?.offset ?? 0) } as T;
  }
  return data as T;
}
export async function upload<T>(path: string, token: string, form: FormData) {
  let r = await fetch(apiBase + path, { method: "PUT", headers: { Authorization: `Bearer ${session?.access_token ?? token}` }, body: form });
  if (!r.ok) throw await responseError(r);
  return (await r.json()) as T;
}
export const api = {
  register: (body: { username: string; email: string; password: string }) =>
    request<User>("/users", { body }),
  verify: (body: { email: string; code: string }) =>
    request<Tokens>("/auth/verify-email", { body }),
  login: (body: { email: string; password: string }) =>
    request<Tokens>("/auth/login", { body }),
  profile: (t: string) => request<User>("/users/me", { token: t }),
  updateProfile: (t: string, display_name: string) => request<User>("/users/me", { token: t, method: "PATCH", body: { display_name } }),
  uploadAvatar: (t: string, file: File) => { const f = new FormData(); f.append("avatar", file); return upload<User>("/users/me/avatar", t, f); },
  deleteAvatar: (t: string) => request<User>("/users/me/avatar", { token: t, method: "DELETE" }),
  projects: (t: string, status?: "active"|"archived", s?: AbortSignal) =>
    request<Project[]>("/projects", { token: t, signal: s, query: { status } }),
  createProject: (t: string, body: { title: string; goal: string }) =>
    request<Project>("/projects", { token: t, body }),
  updateProject: (
    t: string,
    id: UUID,
    body: { title?: string; goal?: string },
  ) => request<Project>(`/projects/${id}`, { token: t, method: "PATCH", body }),
  archiveProject: (t: string, id: UUID) =>
    request<void>(`/projects/${id}`, { token: t, method: "DELETE" }),
  restoreProject: (t: string, id: UUID) => request<Project>(`/projects/${id}/restore`, { token: t, method: "POST" }),
  project: (t: string, id: UUID, s?: AbortSignal) => request<Project>(`/projects/${id}`, { token: t, signal: s }),
  members: (t: string, id: UUID, s?: AbortSignal) =>
    request<ProjectMember[]>(`/projects/${id}/members`, {
      token: t,
      signal: s,
    }),
  invite: (
    t: string,
    id: UUID,
    body: { email: string; role: "admin" | "member" },
  ) => request<Invite>(`/projects/${id}/invites`, { token: t, body }),
  memberRole: (t: string, p: UUID, u: UUID, role: "admin" | "member") =>
    request<void>(`/projects/${p}/members/${u}`, {
      token: t,
      method: "PATCH",
      body: { role },
    }),
  removeMember: (t: string, p: UUID, u: UUID) =>
    request<void>(`/projects/${p}/members/${u}`, {
      token: t,
      method: "DELETE",
    }),
  invites: (t: string, s?: AbortSignal) =>
    request<Invite[]>("/invites", { token: t, signal: s }),
  decideInvite: (t: string, id: UUID, a: "accept" | "decline") =>
    request<void>(`/invites/${id}/${a}`, { token: t, method: "POST" }),
  inviteCount: (t: string) => request<{count:number}>("/invites/count", { token: t }),
  notificationCount: (t: string) => request<{count:number}>("/notifications/unread-count", { token: t }),
  board: (t: string, p: UUID, s?: AbortSignal) => request<Board>(`/projects/${p}/board`, { token: t, signal: s }),
  tasks: (
    t: string,
    p: UUID,
    q: Record<string, string | number | undefined | null> = {},
    s?: AbortSignal,
  ) =>
    request<Paged<Task>>(`/projects/${p}/tasks/search`, {
      token: t,
      query: q,
      signal: s,
    }),
  task: (t: string, p: UUID, id: UUID) =>
    request<Task>(`/projects/${p}/tasks/${id}`, { token: t }),
  createTask: (t: string, p: UUID, body: unknown) =>
    request<Task>(`/projects/${p}/tasks`, { token: t, body }),
  updateTask: (t: string, p: UUID, id: UUID, body: unknown) =>
    request<Task>(`/projects/${p}/tasks/${id}`, {
      token: t,
      method: "PATCH",
      body,
    }),
  deleteTask: (t: string, p: UUID, id: UUID) =>
    request<void>(`/projects/${p}/tasks/${id}`, { token: t, method: "DELETE" }),
  comments: (t: string, p: UUID, id: UUID, l = 20, o = 0) =>
    request<Paged<Comment>>(`/projects/${p}/tasks/${id}/comments`, {
      token: t,
      query: { limit: l, offset: o },
    }),
  addComment: (t: string, p: UUID, id: UUID, body: string) =>
    request<Comment>(`/projects/${p}/tasks/${id}/comments`, {
      token: t,
      body: { body },
    }),
  editComment: (t: string, p: UUID, id: UUID, c: UUID, body: string) =>
    request<void>(`/projects/${p}/tasks/${id}/comments/${c}`, {
      token: t,
      method: "PATCH",
      body: { body },
    }),
  deleteComment: (t: string, p: UUID, id: UUID, c: UUID) =>
    request<void>(`/projects/${p}/tasks/${id}/comments/${c}`, {
      token: t,
      method: "DELETE",
    }),
  history: (t: string, p: UUID, id: UUID, l = 50, o = 0) =>
    request<Paged<History>>(`/projects/${p}/tasks/${id}/history`, {
      token: t,
      query: { limit: l, offset: o },
    }),
  notifications: (t: string, l = 30, o = 0) =>
    request<Paged<Notification>>("/notifications", {
      token: t,
      query: { limit: l, offset: o },
    }),
  readNotification: (t: string, id: UUID) =>
    request<void>(`/notifications/${id}/read`, { token: t, method: "PATCH" }),
  activity: (t: string, p: UUID, cursor?: string | null) => request<ActivityPage>(`/projects/${p}/activity`, { token: t, query: { limit: 50, cursor } }),
};
