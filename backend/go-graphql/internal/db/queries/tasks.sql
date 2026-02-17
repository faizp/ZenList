-- name: CreateTask :one
INSERT INTO tasks (
  user_id,
  project_id,
  parent_task_id,
  title,
  description,
  status,
  priority,
  start_at,
  due_at,
  completed_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, user_id, project_id, parent_task_id, title, description, status, priority, start_at, due_at, completed_at, created_at, updated_at, deleted_at;

-- name: GetTaskByID :one
SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, start_at, due_at, completed_at, created_at, updated_at, deleted_at
FROM tasks
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: ListRootTasks :many
SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, start_at, due_at, completed_at, created_at, updated_at, deleted_at
FROM tasks
WHERE user_id = $1
  AND project_id = $2
  AND parent_task_id IS NULL
  AND deleted_at IS NULL
  AND (cardinality($3::text[]) = 0 OR status = ANY($3::text[]))
  AND (cardinality($4::text[]) = 0 OR priority = ANY($4::text[]))
  AND (
    NOT $5::boolean
    OR (created_at, id) < ($6::timestamptz, $7::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $8;

-- name: ListSubtasks :many
SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, start_at, due_at, completed_at, created_at, updated_at, deleted_at
FROM tasks
WHERE user_id = $1
  AND project_id = $2
  AND parent_task_id = $3
  AND deleted_at IS NULL
  AND (cardinality($4::text[]) = 0 OR status = ANY($4::text[]))
  AND (cardinality($5::text[]) = 0 OR priority = ANY($5::text[]))
  AND (
    NOT $6::boolean
    OR (created_at, id) < ($7::timestamptz, $8::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $9;

-- name: ListSubtasksByParentID :many
SELECT id, user_id, project_id, parent_task_id, title, description, status, priority, start_at, due_at, completed_at, created_at, updated_at, deleted_at
FROM tasks
WHERE user_id = $1
  AND parent_task_id = $2
  AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: UpdateTask :one
UPDATE tasks
SET
  title = $3,
  description = $4,
  status = $5,
  priority = $6,
  start_at = $7,
  due_at = $8,
  completed_at = $9,
  updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING id, user_id, project_id, parent_task_id, title, description, status, priority, start_at, due_at, completed_at, created_at, updated_at, deleted_at;

-- name: SoftDeleteTask :one
UPDATE tasks
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING id, deleted_at;

-- name: SoftDeleteDirectSubtasks :execrows
UPDATE tasks
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE parent_task_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;

-- name: SoftDeleteTasksByProject :execrows
UPDATE tasks
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE project_id = $1
  AND user_id = $2
  AND deleted_at IS NULL;
