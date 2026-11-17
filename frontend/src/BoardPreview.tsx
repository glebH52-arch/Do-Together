import { useMemo, useState } from "react";
import { Clock3, Filter, History, MoreHorizontal, Plus, Search, Users, X } from "lucide-react";
import type { TaskStatus } from "./api";

const statuses: TaskStatus[] = ["todo", "in_progress", "review", "done"];
const names: Record<TaskStatus, string> = { todo: "К выполнению", in_progress: "В работе", review: "На проверке", done: "Готово" };
const demo = [
  ["Подготовить короткий релиз", "todo", "Анна", "Сегодня", "high"],
  ["Согласовать очень длинное название задачи для проверки сетки", "todo", "Илья", "—", "medium"],
  ["Проверить адаптивную версию", "in_progress", "Мария", "Завтра", "high"],
  ["Обновить тексты приглашений", "in_progress", "Анна", "—", "low"],
  ["Проверить сценарий отзыва доступа", "review", "Илья", "18 ноя", "critical"],
  ["Сверить роли владельца и модератора", "review", "Мария", "—", "medium"],
  ["Записать короткую заметку о решении", "done", "Анна", "16 ноя", "low"],
  ["Закрыть задачу с длинным русским описанием", "done", "Илья", "15 ноя", "medium"],
  ["Проверить пустую колонку", "done", "Мария", "—", "low"],
  ["Подготовить демонстрационные данные", "todo", "Анна", "20 ноя", "medium"],
] as const;
const people = ["Анна", "Илья", "Мария"];

export function BoardPreview({ variant }: { variant: "a" | "b" }) {
  const [query, setQuery] = useState("");
  const [filtersOpen, setFiltersOpen] = useState(false);
  const [teamOpen, setTeamOpen] = useState(false);
  const [selectedStatus, setSelectedStatus] = useState("");
  const filtered = useMemo(() => demo.filter(([title, status]) => !query || title.toLowerCase().includes(query.toLowerCase())).filter(([, status]) => !selectedStatus || status === selectedStatus), [query, selectedStatus]);
  const chip = selectedStatus ? names[selectedStatus as TaskStatus] : null;
  return <main className={`preview-shell preview-${variant}`}>
    {variant === "b" && <aside className="preview-rail"><strong>DT</strong><button aria-label="Проекты">⌂</button><button aria-label="Команда"><Users/></button><button aria-label="Настройки">⚙</button></aside>}
    <section className="preview-page">
      <header className="preview-header"><div><div className="crumbs">Проекты <span>/</span> Запуск продукта</div><h1>Запуск продукта</h1></div><div className="preview-actions"><div className="avatar-stack">{people.map((p, i)=><span key={p} title={p}>{p[0]}</span>)}</div>{variant === "b" && <button className="secondary" onClick={() => setTeamOpen(true)}><Users/> Команда</button>}<button className="secondary" aria-label="История проекта"><History/></button><button><Plus/> Задача</button></div></header>
      <div className="preview-toolbar"><label className="preview-search"><Search/><input aria-label="Поиск задач" placeholder="Поиск задач" value={query} onChange={e=>setQuery(e.target.value)}/></label><button className="secondary" onClick={()=>setFiltersOpen(v=>!v)}><Filter/> Фильтры</button>{chip&&<button className="filter-chip" onClick={()=>setSelectedStatus("")}>Статус: {chip}<X/></button>}</div>
      {filtersOpen&&<div className="preview-filter-panel"><label>Статус<select value={selectedStatus} onChange={e=>setSelectedStatus(e.target.value)}><option value="">Все статусы</option>{statuses.map(s=><option value={s} key={s}>{names[s]}</option>)}</select></label><label>Исполнитель<select><option>Все исполнители</option>{people.map(p=><option key={p}>{p}</option>)}</select></label><label>Приоритет<select><option>Все приоритеты</option><option>Высокий</option></select></label><label>С <input type="date"/></label><label>По <input type="date"/></label></div>}
      <div className="preview-board-layout"><div className="preview-board">{statuses.map(status=><section className="preview-column" key={status}><header><strong>{names[status]}</strong><span>{filtered.filter(x=>x[1]===status).length}</span></header><div className="preview-column-body">{filtered.filter(x=>x[1]===status).map(([title,,person,due,priority])=><article className="preview-task" key={title}><div className="preview-task-title"><strong>{title}</strong><button className="icon-control" aria-label={`Меню задачи ${title}`}><MoreHorizontal/></button></div><div className="preview-meta"><span>{person}</span><span>{due}</span></div><div className={`preview-priority ${priority}`}>{priority === "critical" ? "Критический" : priority === "high" ? "Высокий" : priority === "medium" ? "Средний" : "Низкий"}</div></article>)}{filtered.filter(x=>x[1]===status).length===0&&<div className="preview-empty">Свободное место</div>}</div></section>)}</div>{variant === "a"&&<aside className="preview-team"><h2>Команда</h2>{people.map(p=><div className="preview-person" key={p}><span className="avatar avatar-fallback">{p[0]}</span><div><strong>{p}</strong><small>{p === "Анна" ? "Владелец" : "Участник"}</small></div></div>)}</aside>}</div>
      {teamOpen&&<div className="preview-overlay" onClick={()=>setTeamOpen(false)}><aside className="preview-team preview-team-drawer" onClick={e=>e.stopPropagation()}><div className="panel-heading"><h2>Команда</h2><button className="icon-control" onClick={()=>setTeamOpen(false)}><X/></button></div>{people.map(p=><div className="preview-person" key={p}><span className="avatar avatar-fallback">{p[0]}</span><div><strong>{p}</strong><small>Участник</small></div></div>)}</aside></div>}
    </section>
  </main>;
}
