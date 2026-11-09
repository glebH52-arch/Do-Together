CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX tasks_title_trgm_idx ON tasks USING gin (title gin_trgm_ops);
CREATE INDEX tasks_description_trgm_idx ON tasks USING gin (description gin_trgm_ops);
CREATE INDEX tasks_filter_idx ON tasks (project_id, status, assignee_id, priority, task_type, due_at, id);
