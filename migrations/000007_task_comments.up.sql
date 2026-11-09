CREATE TABLE task_comments (
    id BIGSERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    author_id INT NOT NULL REFERENCES users(id),
    body TEXT NOT NULL CHECK (btrim(body) <> '' AND char_length(body) <= 5000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX task_comments_page_idx ON task_comments (task_id, created_at, id);
