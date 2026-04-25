CREATE TABLE IF NOT EXISTS processes (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    is_default BOOLEAN DEFAULT FALSE,
    project_id TEXT,
    created_by TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS statuses (
    id TEXT PRIMARY KEY,
    process_id TEXT NOT NULL,
    name TEXT NOT NULL,
    color TEXT DEFAULT '#6b7280',
    "order" INTEGER DEFAULT 0,
    is_final BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (process_id) REFERENCES processes(id) ON DELETE CASCADE,
    UNIQUE(process_id, name)
);

CREATE TABLE IF NOT EXISTS transitions (
    id TEXT PRIMARY KEY,
    process_id TEXT NOT NULL,
    from_status_id TEXT,
    to_status_id TEXT NOT NULL,
    name TEXT NOT NULL,
    "order" INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (process_id) REFERENCES processes(id) ON DELETE CASCADE,
    FOREIGN KEY (from_status_id) REFERENCES statuses(id) ON DELETE SET NULL,
    FOREIGN KEY (to_status_id) REFERENCES statuses(id) ON DELETE CASCADE,
    UNIQUE(process_id, from_status_id, to_status_id)
);

CREATE INDEX IF NOT EXISTS idx_processes_id ON processes(id);
CREATE INDEX IF NOT EXISTS idx_processes_project ON processes(project_id);
CREATE INDEX IF NOT EXISTS idx_processes_is_default ON processes(is_default);

CREATE INDEX IF NOT EXISTS idx_statuses_id ON statuses(id);
CREATE INDEX IF NOT EXISTS idx_statuses_process ON statuses(process_id);
CREATE INDEX IF NOT EXISTS idx_statuses_order ON statuses("order");

CREATE INDEX IF NOT EXISTS idx_transitions_id ON transitions(id);
CREATE INDEX IF NOT EXISTS idx_transitions_process ON transitions(process_id);
CREATE INDEX IF NOT EXISTS idx_transitions_from_status ON transitions(from_status_id);
CREATE INDEX IF NOT EXISTS idx_transitions_to_status ON transitions(to_status_id);
CREATE INDEX IF NOT EXISTS idx_transitions_process_from ON transitions(process_id, from_status_id);