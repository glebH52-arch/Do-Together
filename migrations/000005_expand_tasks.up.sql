ALTER TABLE projects ADD COLUMN next_task_number BIGINT NOT NULL DEFAULT 1;

ALTER TABLE tasks
    ADD COLUMN uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN task_number BIGINT,
    ADD COLUMN task_key TEXT,
    ADD COLUMN task_type TEXT NOT NULL DEFAULT 'task',
    ADD COLUMN priority TEXT NOT NULL DEFAULT 'medium',
    ADD COLUMN assignee_id INT REFERENCES users(id),
    ADD COLUMN position BIGINT,
    ADD COLUMN completed_at TIMESTAMPTZ;

WITH numbered AS (
    SELECT id, project_id,
           ROW_NUMBER() OVER (PARTITION BY project_id ORDER BY id) AS number,
           ROW_NUMBER() OVER (PARTITION BY project_id, status ORDER BY id) * 1024 AS board_position
    FROM tasks
)
UPDATE tasks AS t
SET task_number = numbered.number,
    task_key = 'DOT-' || numbered.project_id || '-' || numbered.number,
    position = numbered.board_position,
    completed_at = CASE WHEN t.status = 'done' THEN COALESCE(t.updated_at, t.created_at) ELSE NULL END
FROM numbered
WHERE t.id = numbered.id;

UPDATE projects AS p
SET next_task_number = COALESCE(
    (SELECT MAX(t.task_number) + 1 FROM tasks AS t WHERE t.project_id = p.id),
    1
);

ALTER TABLE tasks
    ALTER COLUMN task_number SET NOT NULL,
    ALTER COLUMN task_key SET NOT NULL,
    ALTER COLUMN position SET NOT NULL,
    ADD CONSTRAINT tasks_uuid_unique UNIQUE (uuid),
    ADD CONSTRAINT tasks_key_unique UNIQUE (task_key),
    ADD CONSTRAINT tasks_project_number_unique UNIQUE (project_id, task_number),
    ADD CONSTRAINT tasks_type_check CHECK (task_type IN ('task', 'bug', 'feature')),
    ADD CONSTRAINT tasks_priority_check CHECK (priority IN ('low', 'medium', 'high', 'critical'));

ALTER TABLE tasks DROP CONSTRAINT tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check
    CHECK (status IN ('todo', 'in_progress', 'review', 'done'));

CREATE INDEX tasks_board_idx ON tasks (project_id, status, position, id);
CREATE INDEX tasks_assignee_idx ON tasks (project_id, assignee_id) WHERE assignee_id IS NOT NULL;
