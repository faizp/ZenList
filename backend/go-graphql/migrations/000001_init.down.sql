DROP INDEX IF EXISTS task_labels_label_id_idx;
DROP INDEX IF EXISTS tasks_user_project_status_priority_created_idx;
DROP INDEX IF EXISTS tasks_project_parent_created_idx;
DROP INDEX IF EXISTS labels_user_created_idx;
DROP INDEX IF EXISTS projects_user_created_idx;

DROP TABLE IF EXISTS task_labels;
DROP TABLE IF EXISTS tasks;
DROP INDEX IF EXISTS labels_user_name_active_idx;
DROP TABLE IF EXISTS labels;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
