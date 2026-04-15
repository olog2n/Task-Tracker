CREATE TABLE IF NOT EXISTS tasks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    description TEXT,
    executor VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'backlog',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Индексы на уровне таблицы (синтаксис MySQL)
    INDEX idx_tasks_status (status),
    INDEX idx_tasks_author (author),
    INDEX idx_tasks_created_at (created_at),
    INDEX idx_tasks_status_created (status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;