CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    timezone TEXT NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT,
    color TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT projects_color_hex CHECK (
        color IS NULL OR color ~ '^#[0-9A-Fa-f]{6}$'
    )
);

CREATE TABLE labels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX labels_user_name_active_idx
ON labels (user_id, LOWER(name))
WHERE deleted_at IS NULL;

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    parent_task_id UUID REFERENCES tasks(id),
    title TEXT NOT NULL,
    description TEXT,
    status TEXT NOT NULL,
    priority TEXT NOT NULL,
    start_at TIMESTAMPTZ,
    due_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT tasks_status_check CHECK (status IN ('TODO', 'IN_PROGRESS', 'BLOCKED', 'DONE')),
    CONSTRAINT tasks_priority_check CHECK (priority IN ('P1', 'P2', 'P3', 'P4', 'P5')),
    CONSTRAINT tasks_due_after_start CHECK (due_at IS NULL OR start_at IS NULL OR due_at >= start_at)
);

CREATE TABLE task_labels (
    task_id UUID NOT NULL REFERENCES tasks(id),
    label_id UUID NOT NULL REFERENCES labels(id),
    PRIMARY KEY (task_id, label_id)
);

CREATE INDEX projects_user_created_idx
ON projects (user_id, created_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX labels_user_created_idx
ON labels (user_id, created_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX tasks_project_parent_created_idx
ON tasks (project_id, parent_task_id, created_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX tasks_user_project_status_priority_created_idx
ON tasks (user_id, project_id, status, priority, created_at DESC, id DESC)
WHERE deleted_at IS NULL;

CREATE INDEX task_labels_label_id_idx ON task_labels (label_id);
