DROP INDEX IF EXISTS task_history_project_activity_idx;

ALTER TABLE task_history
    DROP CONSTRAINT IF EXISTS task_history_object_type_check,
    DROP CONSTRAINT task_history_task_id_fkey,
    ADD CONSTRAINT task_history_task_id_fkey
        FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
    ALTER COLUMN task_id SET NOT NULL,
    DROP COLUMN actor_name,
    DROP COLUMN object_title,
    DROP COLUMN object_id,
    DROP COLUMN object_type,
    DROP COLUMN project_id;
