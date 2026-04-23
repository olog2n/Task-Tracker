DROP INDEX IF EXISTS idx_tasks_project;
DROP INDEX IF EXISTS idx_tasks_author_id;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_created_at;
DROP INDEX IF EXISTS idx_tasks_status_created;
DROP TABLE IF EXISTS tasks CASCADE;