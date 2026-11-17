ALTER TABLE task_history
    ADD COLUMN project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    ADD COLUMN object_type TEXT,
    ADD COLUMN object_id UUID,
    ADD COLUMN object_title TEXT,
    ADD COLUMN actor_name TEXT;

UPDATE task_history AS history
SET
    project_id = task.project_id,
    object_type = CASE
        WHEN history.action LIKE 'comment_%'
          OR history.old_values ? 'comment_id'
          OR history.new_values ? 'comment_id'
            THEN 'comment'
        ELSE 'task'
    END,
    object_id = CASE
        WHEN history.action LIKE 'comment_%'
          OR history.old_values ? 'comment_id'
          OR history.new_values ? 'comment_id'
            THEN COALESCE(
                (history.new_values ->> 'comment_id')::UUID,
                (history.old_values ->> 'comment_id')::UUID
            )
        ELSE task.id
    END,
    object_title = CASE
        WHEN history.action LIKE 'comment_%' THEN task.title
        ELSE task.title
    END,
    actor_name = actor.display_name
FROM tasks AS task, users AS actor
WHERE task.id = history.task_id
  AND actor.id = history.actor_id;

UPDATE task_history
SET
    old_values = old_values - 'body',
    new_values = new_values - 'body'
WHERE object_type = 'comment';

ALTER TABLE task_history
    ALTER COLUMN project_id SET NOT NULL,
    ALTER COLUMN object_type SET NOT NULL,
    ALTER COLUMN object_id SET NOT NULL,
    ALTER COLUMN object_title SET NOT NULL,
    ALTER COLUMN actor_name SET NOT NULL,
    ALTER COLUMN task_id DROP NOT NULL,
    DROP CONSTRAINT task_history_task_id_fkey,
    ADD CONSTRAINT task_history_task_id_fkey
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE SET NULL,
    ADD CONSTRAINT task_history_object_type_check
        CHECK (object_type IN ('project', 'invite', 'member', 'task', 'comment'));

CREATE INDEX task_history_project_activity_idx
    ON task_history (project_id, occurred_at DESC, id DESC);
