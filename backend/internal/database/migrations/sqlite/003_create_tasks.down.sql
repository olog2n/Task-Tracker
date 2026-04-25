DROP INDEX IF EXISTS idx_tasks_project;
DROP INDEX IF EXISTS idx_tasks_assignee;
DROP INDEX IF EXISTS idx_tasks_created_by;

DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_priority;

DROP INDEX IF EXISTS idx_tasks_deleted_at;

DROP INDEX IF EXISTS idx_tasks_status_created;
DROP INDEX IF EXISTS idx_tasks_project_status;

DROP INDEX IF EXISTS idx_tasks_id;

DROP TABLE IF EXISTS tasks;