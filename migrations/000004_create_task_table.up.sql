CREATE TABLE tasks(
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL,
    title VARCHAR(100) NOT NULL CHECK (btrim(title) <> ''),
    description VARCHAR(1000) NOT NULL CHECK (btrim(description) <> ''),
    created_by INT NOT NULL,
   status TEXT NOT NULL DEFAULT 'todo'
        CHECK (status IN ('todo', 'in_progress', 'done')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    due_at TIMESTAMPTZ,
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);