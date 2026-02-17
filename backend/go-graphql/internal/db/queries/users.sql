-- name: GetUserByID :one
SELECT id, name, email, timezone, avatar_url, created_at, updated_at, deleted_at
FROM users
WHERE id = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, name, email, timezone, avatar_url, created_at, updated_at, deleted_at
FROM users
WHERE email = $1
  AND deleted_at IS NULL
LIMIT 1;

-- name: UpsertUserByEmail :one
INSERT INTO users (name, email, timezone, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (email) DO UPDATE
SET
  name = EXCLUDED.name,
  timezone = EXCLUDED.timezone,
  avatar_url = EXCLUDED.avatar_url,
  deleted_at = NULL,
  updated_at = NOW()
RETURNING id, name, email, timezone, avatar_url, created_at, updated_at, deleted_at;
