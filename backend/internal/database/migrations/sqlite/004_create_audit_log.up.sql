CREATE TABLE IF NOT EXISTS audit_log (
    -- ID is TEXT (UUID format)
    id TEXT PRIMARY KEY,
    
    -- UUID reference to users(id)
    actor_id TEXT,
    
    -- User info (denormalized for audit integrity)
    user_email TEXT,
    user_name TEXT,
    
    -- Action details
    action TEXT NOT NULL,              -- create, update, delete, select, status_change, etc.
    target_type TEXT NOT NULL,         -- user, task, project, system
    target_id TEXT,                    -- UUID (nullable for massive op)
    
    -- Old/New values (JSON)
    old_value TEXT,
    new_value TEXT,
    
    -- Additional context (JSON)
    metadata TEXT,
    
    -- Data classification
    classification TEXT DEFAULT 'internal',  -- public, internal, confidential, restricted
    
    -- Request context
    ip_address TEXT,
    user_agent TEXT,
    
    -- Timestamp
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- FOREIGN KEY with TEXT (UUID)
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Actor lookup
CREATE INDEX IF NOT EXISTS idx_audit_log_actor ON audit_log(actor_id);

-- Action type filtering 
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

-- Target lookup 
CREATE INDEX IF NOT EXISTS idx_audit_log_target ON audit_log(target_type, target_id);

-- Time-based queries
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log(created_at);

-- Classification filtering 
CREATE INDEX IF NOT EXISTS idx_audit_log_classification ON audit_log(classification);

-- UUID primary key index
CREATE INDEX IF NOT EXISTS idx_audit_log_id ON audit_log(id);

-- Composite index for common admin queries
CREATE INDEX IF NOT EXISTS idx_audit_log_actor_created ON audit_log(actor_id, created_at);