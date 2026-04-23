CREATE TABLE IF NOT EXISTS tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    project_id INT,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,           -- Email автора (денормализация)
    author_id INT NOT NULL,                 -- author`s id for validation role
    description TEXT,
    executor VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'backlog',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_tasks_project 
        FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL,
    CONSTRAINT fk_tasks_author 
        FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL,
    
    INDEX idx_tasks_project (project_id),
    INDEX idx_tasks_author_id (author_id),
    INDEX idx_tasks_status (status),
    INDEX idx_tasks_created_at (created_at),
    INDEX idx_tasks_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;