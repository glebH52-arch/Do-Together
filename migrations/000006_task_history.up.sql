CREATE TABLE task_history (
    id BIGSERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    actor_id INT NOT NULL REFERENCES users(id),
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    action TEXT NOT NULL,
    old_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    new_values JSONB NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT task_history_action_check CHECK (action IN ('created', 'edited', 'status_changed', 'assignee_changed', 'priority_changed', 'due_at_changed', 'comment_created', 'comment_deleted'))
);

CREATE INDEX task_history_task_time_idx ON task_history (task_id, occurred_at, id);
