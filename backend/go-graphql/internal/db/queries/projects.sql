-- name: CreateProject :one
INSERT INTO projects (user_id, title, description, color)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, title, description, color, created_at, updated_at, deleted_at;

-- name: GetProjectByID :one
SELECT id, user_id, title, description, color, created_at, updated_at, deleted_at
FROM projects
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: ListProjects :many
SELECT id, user_id, title, description, color, created_at, updated_at, deleted_at
FROM projects
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    NOT $2::boolean
    OR (created_at, id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $5;

-- name: UpdateProject :one
UPDATE projects
SET
  title = $3,
  description = $4,
  color = $5,
  updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING id, user_id, title, description, color, created_at, updated_at, deleted_at;

-- name: SoftDeleteProject :one
UPDATE projects
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING id, deleted_at;
