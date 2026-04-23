DROP INDEX IF EXISTS idx_projects_owner;
DROP INDEX IF EXISTS idx_projects_is_active;
DROP INDEX IF EXISTS idx_project_members_project;
DROP INDEX IF EXISTS idx_project_members_user;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;