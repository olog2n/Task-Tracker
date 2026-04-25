CREATE TABLE IF NOT EXISTS tasks (
    -- ID is TEXT (UUID format)
    id TEXT PRIMARY KEY,
    
    -- Task data
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    
    -- Status (для v2 будет status_id TEXT references statuses(id))
    status TEXT DEFAULT 'backlog' NOT NULL,
    
    -- Priority (low, medium, high)
    priority TEXT DEFAULT 'medium' NOT NULL,
    
    -- References (все UUID)
    project_id TEXT,
    assignee_id TEXT,
    created_by TEXT NOT NULL,
    updated_by TEXT,
    
    -- Soft delete
    deleted_at DATETIME,
    deleted_by TEXT,
    
    -- Data classification (audit)
    classification TEXT DEFAULT 'internal',
    is_sensitive BOOLEAN DEFAULT 0,
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- FOREIGN KEY with TEXT (UUID)
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (updated_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (deleted_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_created_by ON tasks(created_by);

-- Status & priority filtering
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_priority ON tasks(priority);

-- Soft delete
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at ON tasks(deleted_at);

-- Composite indexes for common queries
CREATE INDEX IF NOT EXISTS idx_tasks_status_created ON tasks(status, created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks(project_id, status);

-- UUID primary key index
CREATE INDEX IF NOT EXISTS idx_tasks_id ON tasks(id); 