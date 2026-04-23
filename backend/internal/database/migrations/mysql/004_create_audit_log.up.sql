CREATE TABLE IF NOT EXISTS audit_log (
    id INT AUTO_INCREMENT PRIMARY KEY,
    actor_id INT NOT NULL,
    action VARCHAR(100) NOT NULL,
    target_type VARCHAR(50) NOT NULL,      -- "user", "task", "project", "project_member"
    target_id INT,
    old_value JSON,                        -- JSON
    new_value JSON,
    ip_address VARCHAR(45),                -- IPv4/IPv6
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_audit_log_actor 
        FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_audit_log_actor (actor_id),
    INDEX idx_audit_log_action (action),
    INDEX idx_audit_log_target (target_type, target_id),
    INDEX idx_audit_log_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;