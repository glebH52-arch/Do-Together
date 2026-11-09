DROP INDEX IF EXISTS tasks_assignee_idx;
DROP INDEX IF EXISTS tasks_board_idx;

ALTER TABLE tasks DROP CONSTRAINT tasks_status_check;
ALTER TABLE tasks ADD CONSTRAINT tasks_status_check
    CHECK (status IN ('todo', 'in_progress', 'done'));

ALTER TABLE tasks
    DROP COLUMN completed_at,
    DROP COLUMN position,
    DROP COLUMN assignee_id,
    DROP COLUMN priority,
    DROP COLUMN task_type,
    DROP COLUMN task_key,
    DROP COLUMN task_number,
    DROP COLUMN uuid;

ALTER TABLE projects DROP COLUMN next_task_number;
