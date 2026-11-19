import type { Task, TaskStatus } from "./api";

export function canMoveTask(task: Task, next: TaskStatus) {
  return task.status !== next && task.allowed_transitions.includes(next);
}

export function moveTaskLocally(tasks: Task[], id: string, next: TaskStatus) {
  return tasks.map((task) => task.id === id ? { ...task, status: next } : task);
}

export function mergeUniqueByID<T extends { id: string }>(current: T[], incoming: T[]) {
  const seen = new Set(current.map((item) => item.id));
  return [...current, ...incoming.filter((item) => !seen.has(item.id))];
}

export function friendlyActionError(error: unknown, fallback: string) {
  if (typeof error === "object" && error && "status" in error) {
    const status = Number((error as { status: unknown }).status);
    if (status === 400 || status === 422) return "Проверьте обязательные поля и введённые данные.";
    if (status === 403) return "У вас нет доступа к этому действию.";
    if (status === 409) return "Данные уже изменились. Обновите состояние и повторите действие.";
    if (status >= 500) return "Сервис временно недоступен. Введённые данные сохранены — попробуйте ещё раз.";
  }
  return fallback;
}

export function escapeTarget(state: { task: boolean; team: boolean; activity: boolean }) {
  if (state.task) return "task";
  if (state.team) return "team";
  if (state.activity) return "activity";
  return null;
}
