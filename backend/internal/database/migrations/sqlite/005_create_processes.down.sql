DROP INDEX IF EXISTS idx_processes_id;
DROP INDEX IF EXISTS idx_processes_project;
DROP INDEX IF EXISTS idx_processes_is_default;

DROP INDEX IF EXISTS idx_statuses_id;
DROP INDEX IF EXISTS idx_statuses_process;
DROP INDEX IF EXISTS idx_statuses_order;

DROP INDEX IF EXISTS idx_transitions_id
DROP INDEX IF EXISTS idx_transitions_process
DROP INDEX IF EXISTS idx_transitions_from_status
DROP INDEX IF EXISTS idx_transitions_to_status
DROP INDEX IF EXISTS idx_transitions_process_from

DROP TABLE IF EXISTS processes
DROP TABLE IF EXISTS statuses
DROP TABLE IF EXISTS transitions