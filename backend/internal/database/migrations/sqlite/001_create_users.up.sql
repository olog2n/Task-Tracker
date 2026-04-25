CREATE TABLE IF NOT EXISTS users (
    -- ID is TEXT (UUID format)
    id TEXT PRIMARY KEY,
    
    -- User data
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    
    -- Account status
    is_active BOOLEAN DEFAULT 1 NOT NULL,
    require_password_reset BOOLEAN DEFAULT 0 NOT NULL,
    
    -- Soft delete
    deleted_at DATETIME,
    deleted_by TEXT,  -- UUID reference
    
    -- Token management (for JWT invalidation)
    token_version INTEGER DEFAULT 1 NOT NULL,
    
    -- Account reactivation tracking
    reactivated_at DATETIME,
    reactivated_by TEXT,  -- UUID reference
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME,
    
    -- FOREIGN KEY with TEXT (UUID)
    FOREIGN KEY (deleted_by) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (reactivated_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_id ON users(id);