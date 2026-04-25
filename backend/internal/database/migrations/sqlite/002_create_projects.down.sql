DROP INDEX IF EXISTS idx_projects_owner;
DROP INDEX IF EXISTS idx_projects_is_active;
DROP INDEX IF EXISTS idx_projects_deleted_at;
DROP INDEX IF EXISTS idx_projects_id;

DROP INDEX IF EXISTS idx_project_members_project;
DROP INDEX IF EXISTS idx_project_members_user;
DROP INDEX IF EXISTS idx_project_members_role;
DROP INDEX IF EXISTS idx_project_members_id;

DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;