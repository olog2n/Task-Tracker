CREATE TABLE IF NOT EXISTS projects (
    -- ID is TEXT (UUID format)
    id TEXT PRIMARY KEY,
    
    -- Project data
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    
    -- Owner (references users.id - UUID)
    owner_id TEXT NOT NULL,
    
    -- Project status
    is_active BOOLEAN DEFAULT 1 NOT NULL,
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    -- FOREIGN KEY with TEXT (UUID)
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS project_members (
    -- ID is TEXT (UUID format)
    id TEXT PRIMARY KEY,
    
    -- References (все UUID)
    project_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    
    -- Role (соответствует model.Role)
    role TEXT NOT NULL DEFAULT 'member',  -- viewer, member, project_admin
    
    -- Timestamps
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Who added this member (UUID)
    added_by TEXT,
    
    -- FOREIGN KEY with TEXT (UUID)
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (added_by) REFERENCES users(id) ON DELETE SET NULL,
    
    -- UNIQUE constraint (one user = one table)
    UNIQUE(project_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_projects_owner ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_is_active ON projects(is_active);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX IF NOT EXISTS idx_projects_id ON projects(id);

CREATE INDEX IF NOT EXISTS idx_project_members_project ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_project_members_user ON project_members(user_id);
CREATE INDEX IF NOT EXISTS idx_project_members_role ON project_members(role);
CREATE INDEX IF NOT EXISTS idx_project_members_id ON project_members(id);