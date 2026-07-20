CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    created_by INT NOT NULL,
    title VARCHAR(100) NOT NULL CHECK (btrim(title) <> ''),
    goal VARCHAR(200) NOT NULL CHECK (btrim(goal) <> ''),
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (created_by) REFERENCES users(id)
);

CREATE TABLE project_members(
    project_id INT NOT NULL,
    user_id INT NOT NULL,
    role TEXT NOT NULL CHECK (btrim(role) <> ''),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (project_id, user_id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);