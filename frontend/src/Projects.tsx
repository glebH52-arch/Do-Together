import { useEffect, useState, type FormEvent } from 'react';
import { request, type Project } from './api';

type ListState =
  | { kind: 'loading' }
  | { kind: 'error' }
  | { kind: 'ready'; projects: Project[] };

function statusLabel(status: Project['status']) {
  return status === 'archived' ? 'В архиве' : 'Активный';
}

export function Projects({ accessToken, onOpen }: {
  accessToken: string;
  onOpen: (id: number) => void;
}) {
  const [list, setList] = useState<ListState>({ kind: 'loading' });
  const [reload, setReload] = useState(0);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [title, setTitle] = useState('');
  const [goal, setGoal] = useState('');
  const [createBusy, setCreateBusy] = useState(false);
  const [createError, setCreateError] = useState('');

  useEffect(() => {
    const controller = new AbortController();
    setList({ kind: 'loading' });
    void request<Project[]>('/projects', undefined, accessToken, controller.signal)
      .then(projects => setList({ kind: 'ready', projects }))
      .catch(() => {
        if (!controller.signal.aborted) setList({ kind: 'error' });
      });
    return () => controller.abort();
  }, [accessToken, reload]);

  function closeDialog() {
    if (createBusy) return;
    setDialogOpen(false);
    setTitle('');
    setGoal('');
    setCreateError('');
  }

  async function createProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (createBusy) return;
    const cleanTitle = title.trim();
    const cleanGoal = goal.trim();
    if (!cleanTitle || !cleanGoal) {
      setCreateError('Заполните название и цель проекта.');
      return;
    }
    if (Array.from(cleanTitle).length > 100 || Array.from(cleanGoal).length > 200) {
      setCreateError('Название: до 100 символов, цель: до 200 символов.');
      return;
    }
    setCreateBusy(true);
    setCreateError('');
    try {
      const project = await request<Project>(
        '/projects',
        { title: cleanTitle, goal: cleanGoal },
        accessToken,
      );
      setList(current => current.kind === 'ready'
        ? { kind: 'ready', projects: [project, ...current.projects.filter(item => item.id !== project.id)] }
        : { kind: 'ready', projects: [project] });
      setDialogOpen(false);
      setTitle('');
      setGoal('');
    } catch (cause) {
      setCreateError(cause instanceof Error ? cause.message : 'Не удалось создать проект.');
    } finally {
      setCreateBusy(false);
    }
  }

  return (
    <>
      <main className="workspace-main">
        <div className="section-heading">
          <div>
            <h1>Мои проекты</h1>
            <p className="muted">Проекты, к которым у вас есть доступ</p>
          </div>
          {list.kind === 'ready' && list.projects.length > 0 && (
            <button onClick={() => setDialogOpen(true)}>+ Создать проект</button>
          )}
        </div>

        {list.kind === 'loading' && <p role="status">Загружаем проекты…</p>}
        {list.kind === 'error' && (
          <div className="state-card" role="alert">
            <h2>Не удалось загрузить проекты</h2>
            <button onClick={() => setReload(value => value + 1)}>Повторить</button>
          </div>
        )}
        {list.kind === 'ready' && list.projects.length === 0 && (
          <div className="state-card">
            <h2>Пока нет проектов</h2>
            <p className="muted">Создайте проект и опишите, чего хотите достичь</p>
            <button onClick={() => setDialogOpen(true)}>Создать первый проект</button>
          </div>
        )}
        {list.kind === 'ready' && list.projects.length > 0 && (
          <div className="project-grid">
            {list.projects.map(project => (
              <button type="button" className="project-card project-link" key={project.id} onClick={() => onOpen(project.id)} aria-label={`Открыть проект ${project.title}`}>
                <div className="project-card-top">
                  <h2>{project.title}</h2>
                  <span className="status">{statusLabel(project.status)}</span>
                </div>
                <p>{project.goal}</p>
              </button>
            ))}
          </div>
        )}
      </main>

      {dialogOpen && (
        <div className="dialog-backdrop">
          <div className="dialog-card" role="dialog" aria-modal="true" aria-labelledby="dialog-title">
            <h2 id="dialog-title">Новый проект</h2>
            <form onSubmit={event => void createProject(event)}>
              <label>Название
                <input
                  autoFocus
                  value={title}
                  onChange={event => setTitle(event.target.value)}
                  placeholder="Запуск сайта"
                  maxLength={100}
                  required
                />
              </label>
              <label>Цель проекта
                <textarea
                  value={goal}
                  onChange={event => setGoal(event.target.value)}
                  placeholder="Какого результата вы хотите достичь?"
                  maxLength={200}
                  rows={4}
                  required
                />
              </label>
              {createError && <p className="error" role="alert">{createError}</p>}
              <div className="dialog-actions">
                <button className="secondary" type="button" disabled={createBusy} onClick={closeDialog}>Отмена</button>
                <button type="submit" disabled={createBusy}>{createBusy ? 'Создаём…' : 'Создать проект'}</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  );
}
