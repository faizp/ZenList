-- name: CreateLabel :one
INSERT INTO labels (user_id, name)
VALUES ($1, $2)
RETURNING id, user_id, name, created_at, updated_at, deleted_at;

-- name: GetLabelByID :one
SELECT id, user_id, name, created_at, updated_at, deleted_at
FROM labels
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetLabelsByIDs :many
SELECT id, user_id, name, created_at, updated_at, deleted_at
FROM labels
WHERE user_id = $1
  AND id = ANY($2::uuid[])
  AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC;

-- name: ListLabels :many
SELECT id, user_id, name, created_at, updated_at, deleted_at
FROM labels
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    NOT $2::boolean
    OR (created_at, id) < ($3::timestamptz, $4::uuid)
  )
ORDER BY created_at DESC, id DESC
LIMIT $5;

-- name: UpdateLabel :one
UPDATE labels
SET
  name = $3,
  updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING id, user_id, name, created_at, updated_at, deleted_at;

-- name: SoftDeleteLabel :one
UPDATE labels
SET
  deleted_at = NOW(),
  updated_at = NOW()
WHERE id = $1
  AND user_id = $2
  AND deleted_at IS NULL
RETURNING id, deleted_at;

-- name: DeleteTaskLabelsByLabelID :execrows
DELETE FROM task_labels
WHERE label_id = $1;

-- name: ListLabelsByTaskID :many
SELECT l.id, l.user_id, l.name, l.created_at, l.updated_at, l.deleted_at
FROM labels l
JOIN task_labels tl ON tl.label_id = l.id
WHERE tl.task_id = $1
  AND l.user_id = $2
  AND l.deleted_at IS NULL
ORDER BY l.created_at DESC, l.id DESC;
