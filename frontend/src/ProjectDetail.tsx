import { useEffect, useState, type DragEvent, type FormEvent, type KeyboardEvent } from 'react';
import { ApiError, request, type Project, type Task, type User } from './api';
import { ProjectMembers } from './ProjectMembers';

type LoadState = { kind: 'loading' } | { kind: 'denied' } | { kind: 'error' } | { kind: 'ready'; project: Project; tasks: Task[] };
const labels: Record<Task['status'], string> = { todo: 'К выполнению', in_progress: 'В работе', review: 'На проверке', done: 'Готово' };
const columns: Task['status'][] = ['todo', 'in_progress', 'review', 'done'];

function taskCount(count: number) {
  const lastTwo = count % 100;
  const last = count % 10;
  return `${count} ${lastTwo >= 11 && lastTwo <= 14 ? 'задач' : last === 1 ? 'задача' : last >= 2 && last <= 4 ? 'задачи' : 'задач'}`;
}

function errorText(cause: unknown) {
  return cause instanceof Error ? cause.message : 'Запрос не выполнен.';
}

function formatDue(value: string) {
  return new Intl.DateTimeFormat('ru-RU', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

export function ProjectDetail({ projectID, user, accessToken, onBack }: {
  projectID: number;
  user: User;
  accessToken: string;
  onBack: () => void;
}) {
  const base = `/projects/${projectID}`;
  const [state, setState] = useState<LoadState>({ kind: 'loading' });
  const [reload, setReload] = useState(0);
  const [dialog, setDialog] = useState<'create' | Task | null>(null);
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dueAt, setDueAt] = useState('');
  const [status, setStatus] = useState<Task['status']>('todo');
  const [busy, setBusy] = useState(false);
  const [busyTask, setBusyTask] = useState<number | null>(null);
  const [dropTarget, setDropTarget] = useState<Task['status'] | null>(null);
  const [confirmDelete, setConfirmDelete] = useState(false);
  const [error, setError] = useState('');
  const [taskError, setTaskError] = useState('');

  useEffect(() => {
    const controller = new AbortController();
    setState({ kind: 'loading' });
    void Promise.all([
      request<Project>(base, undefined, accessToken, controller.signal),
      request<Task[]>(`${base}/tasks`, undefined, accessToken, controller.signal),
    ]).then(([project, tasks]) => setState({ kind: 'ready', project, tasks }))
      .catch(cause => {
        if (controller.signal.aborted) return;
        setState(cause instanceof ApiError && (cause.status === 403 || cause.status === 404) ? { kind: 'denied' } : { kind: 'error' });
      });
    return () => controller.abort();
  }, [base, accessToken, reload]);

  async function refreshTasks() {
    const tasks = await request<Task[]>(`${base}/tasks`, undefined, accessToken);
    setState(current => current.kind === 'ready' ? { ...current, tasks } : current);
  }

  function openCreate() {
    setTitle(''); setDescription(''); setDueAt(''); setStatus('todo'); setError(''); setConfirmDelete(false); setDialog('create');
  }

  function openEdit(task: Task) {
    setTitle(task.title); setDescription(task.description); setStatus(task.status); setError(''); setConfirmDelete(false); setDialog(task);
  }

  async function moveTask(task: Task, nextStatus: Task['status']) {
    if (busy || busyTask !== null || task.status === nextStatus) return;
    setBusyTask(task.id);
    setTaskError('');
    try {
      await request<void>(`${base}/tasks/${task.id}`, { status: nextStatus }, accessToken, undefined, 'PATCH');
      setState(current => current.kind === 'ready' ? {
        ...current,
        tasks: current.tasks.map(item => item.id === task.id ? { ...item, status: nextStatus } : item),
      } : current);
      try { await refreshTasks(); }
      catch { setTaskError('Статус сохранён, но не удалось обновить список. Обновите страницу.'); }
    } catch (cause) {
      setTaskError(`Не удалось перенести «${task.title}»: ${errorText(cause)}`);
    } finally { setBusyTask(null); }
  }

  function dropTask(event: DragEvent<HTMLElement>, nextStatus: Task['status']) {
    event.preventDefault();
    setDropTarget(null);
    if (state.kind !== 'ready') return;
    const id = Number(event.dataTransfer.getData('text/plain'));
    const task = state.tasks.find(item => item.id === id);
    if (task) void moveTask(task, nextStatus);
  }

  function openOnKeyboard(event: KeyboardEvent<HTMLElement>, task: Task) {
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      openEdit(task);
    }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (busy || busyTask !== null || !dialog) return;
    const cleanTitle = title.trim();
    const cleanDescription = description.trim();
    if (!cleanTitle || !cleanDescription) { setError('Заполните название и описание.'); return; }
    if (Array.from(cleanTitle).length > 100 || Array.from(cleanDescription).length > 1000) {
      setError('Название: до 100 символов, описание: до 1000 символов.'); return;
    }
    const isCreate = dialog === 'create';
    const changes: { title?: string; description?: string; status?: Task['status'] } = {};
    if (!isCreate) {
      if (cleanTitle !== dialog.title) changes.title = cleanTitle;
      if (cleanDescription !== dialog.description) changes.description = cleanDescription;
      if (status !== dialog.status) changes.status = status;
      if (!Object.keys(changes).length) { setDialog(null); return; }
    }
    setBusy(true); setError(''); setTaskError('');
    try {
      if (isCreate) {
        await request<Task>(`${base}/tasks`, {
          title: cleanTitle, description: cleanDescription,
          ...(dueAt ? { due_at: new Date(dueAt).toISOString() } : {}),
        }, accessToken);
      } else {
        await request<void>(`${base}/tasks/${dialog.id}`, changes, accessToken, undefined, 'PATCH');
      }
      await refreshTasks();
      setDialog(null);
    } catch (cause) { setError(errorText(cause)); }
    finally { setBusy(false); }
  }

  async function removeTask(id: number) {
    if (busy || busyTask !== null) return;
    setBusyTask(id); setTaskError('');
    try {
      await request<void>(`${base}/tasks/${id}`, undefined, accessToken, undefined, 'DELETE');
      setState(current => current.kind === 'ready' ? { ...current, tasks: current.tasks.filter(item => item.id !== id) } : current);
      setDialog(null);
      setConfirmDelete(false);
      try { await refreshTasks(); }
      catch { setTaskError('Задача удалена, но не удалось обновить список. Обновите страницу.'); }
    } catch (cause) { setError(errorText(cause)); }
    finally { setBusyTask(null); }
  }

  return <>
    <main className="workspace-main">
      <button className="back-link" onClick={onBack}>Все проекты ←</button>
      {state.kind === 'loading' && <p role="status">Загружаем проект и задачи…</p>}
      {state.kind === 'denied' && <div className="state-card" role="alert"><h1>Нет доступа к проекту</h1><p>Проект не найден или у вас нет доступа к нему.</p></div>}
      {state.kind === 'error' && <div className="state-card" role="alert"><h1>Не удалось загрузить проект</h1><button onClick={() => setReload(value => value + 1)}>Повторить</button></div>}
      {state.kind === 'ready' && <>
        <section className="project-intro"><h1>{state.project.title}</h1><p>{state.project.goal}</p></section>
        <div className="section-heading task-heading"><h2>Задачи</h2><button onClick={openCreate}>+ Добавить задачу</button></div>
        {taskError && <p className="error" role="alert">{taskError}</p>}
        {state.tasks.length === 0 && <p className="board-empty" role="status">Пока нет задач. Добавьте первую задачу для этого проекта.</p>}
        <div className="kanban-board" aria-label="Доска задач">
          {columns.map(column => {
            const tasks = state.tasks.filter(task => task.status === column);
            return <section key={column} className={`kanban-column ${dropTarget === column ? 'drop-target' : ''}`}
              aria-label={`${labels[column]}: ${taskCount(tasks.length)}`}
              onDragOver={event => { if (busy || busyTask !== null) return; event.preventDefault(); event.dataTransfer.dropEffect = 'move'; setDropTarget(column); }}
              onDragLeave={event => { if (!event.currentTarget.contains(event.relatedTarget as Node)) setDropTarget(null); }}
              onDrop={event => dropTask(event, column)}>
              <h3 className="column-heading"><span>{labels[column]}</span><span className="column-count">{tasks.length}</span></h3>
              <div className="column-tasks">
                {tasks.length === 0 && <p className="empty-column">Нет задач</p>}
                {tasks.map(task => {
                  const overdue = Boolean(task.due_at && task.status !== 'done' && new Date(task.due_at).getTime() < Date.now());
                  return <article key={task.id} className={`task-card ${busyTask === task.id ? 'task-updating' : ''}`}
                    role="button" tabIndex={0} aria-label={`Открыть задачу ${task.title}`}
                    draggable={!busy && busyTask === null}
                    onDragStart={event => { event.dataTransfer.setData('text/plain', String(task.id)); event.dataTransfer.effectAllowed = 'move'; }}
                    onDragEnd={() => setDropTarget(null)}
                    onClick={() => openEdit(task)} onKeyDown={event => openOnKeyboard(event, task)}>
                    <span className="drag-handle" aria-hidden="true">⠿</span>
                    <span className="open-indicator" aria-hidden="true">↗</span>
                    <h4>{task.title}</h4><p>{task.description}</p>
                    <div className="task-meta"><span className={overdue ? 'overdue' : ''}>{task.due_at ? `${overdue ? 'Просрочено · ' : 'Срок · '}${formatDue(task.due_at)}` : 'Без срока'}</span><span>Создал: Пользователь #{task.created_by}</span></div>
                    {busyTask === task.id && <span className="updating-label" role="status">Обновляем…</span>}
                  </article>;
                })}
              </div>
            </section>;
          })}
        </div>
        <ProjectMembers projectID={projectID} user={user} accessToken={accessToken} />
      </>}
    </main>
    {dialog && <div className="dialog-backdrop"><div className="dialog-card" role="dialog" aria-modal="true" aria-labelledby="task-dialog-title">
      <h2 id="task-dialog-title">{dialog === 'create' ? 'Новая задача' : 'Задача'}</h2>
      {confirmDelete && dialog !== 'create' ? <div className="delete-dialog">
        <h3>Удалить эту задачу?</h3><p>Действие нельзя отменить.</p>
        {error && <p className="error" role="alert">{error}</p>}
        <div className="dialog-actions"><button className="secondary" disabled={busyTask !== null} onClick={() => { setConfirmDelete(false); setError(''); }}>Отмена</button><button className="danger" disabled={busyTask !== null} onClick={() => void removeTask(dialog.id)}>{busyTask === dialog.id ? 'Удаляем…' : 'Удалить'}</button></div>
      </div> : <form onSubmit={event => void submit(event)}>
        <label>Название<input autoFocus value={title} onChange={event => setTitle(event.target.value)} maxLength={100} required /></label>
        <label>Описание<textarea value={description} onChange={event => setDescription(event.target.value)} maxLength={1000} rows={4} required /></label>
        {dialog === 'create' ? <label>Срок (необязательно)<input type="datetime-local" value={dueAt} onChange={event => setDueAt(event.target.value)} /></label>
          : <label>Статус<select value={status} onChange={event => setStatus(event.target.value as Task['status'])}>{Object.entries(labels).map(([value, label]) => <option key={value} value={value}>{label}</option>)}</select></label>}
        {dialog !== 'create' && <p className="readonly-due">Срок: {dialog.due_at ? formatDue(dialog.due_at) : 'Без срока'} · редактирование срока пока не добавлено в этот интерфейс.</p>}
        {error && <p className="error" role="alert">{error}</p>}
        <div className="dialog-actions"><button className="secondary" type="button" disabled={busy} onClick={() => setDialog(null)}>Отмена</button><button type="submit" disabled={busy}>{busy ? 'Сохраняем…' : dialog === 'create' ? 'Создать задачу' : 'Сохранить'}</button></div>
        {dialog !== 'create' && dialog.created_by === user.id && <button className="delete-link" type="button" disabled={busy} onClick={() => { setError(''); setConfirmDelete(true); }}>Удалить задачу</button>}
      </form>}
    </div></div>}
  </>;
}
