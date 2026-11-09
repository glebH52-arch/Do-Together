CREATE TABLE notifications (
    id BIGSERIAL PRIMARY KEY,
    recipient_id INT NOT NULL REFERENCES users(id),
    project_id INT NOT NULL REFERENCES projects(id),
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    invite_id INT REFERENCES invites(id) ON DELETE CASCADE,
    source_history_id BIGINT REFERENCES task_history(id) ON DELETE CASCADE,
    kind TEXT NOT NULL CHECK (kind IN ('assignment_changed', 'status_changed', 'due_at_changed', 'invitation')),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    read_at TIMESTAMPTZ,
    CONSTRAINT notification_source_check CHECK (
        (kind = 'invitation' AND invite_id IS NOT NULL AND task_id IS NULL)
        OR (kind <> 'invitation' AND task_id IS NOT NULL AND source_history_id IS NOT NULL)
    )
);

CREATE UNIQUE INDEX notifications_task_event_unique ON notifications (recipient_id, source_history_id) WHERE source_history_id IS NOT NULL;
CREATE UNIQUE INDEX notifications_invite_unique ON notifications (recipient_id, invite_id) WHERE invite_id IS NOT NULL;
CREATE INDEX notifications_inbox_idx ON notifications (recipient_id, created_at DESC, id DESC);
CREATE INDEX notifications_unread_idx ON notifications (recipient_id, created_at DESC, id DESC) WHERE read_at IS NULL;
