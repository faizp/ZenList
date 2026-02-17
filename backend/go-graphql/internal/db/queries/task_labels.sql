-- name: DeleteTaskLabelsForTask :exec
DELETE FROM task_labels
WHERE task_id = $1;

-- name: InsertTaskLabel :exec
INSERT INTO task_labels (task_id, label_id)
VALUES ($1, $2)
ON CONFLICT (task_id, label_id) DO NOTHING;
