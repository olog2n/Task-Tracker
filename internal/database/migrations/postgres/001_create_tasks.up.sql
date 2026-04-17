CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    author VARCHAR(255) NOT NULL,
    author_id INTEGER NOT NULL,
    description TEXT,
    executor VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'backlog',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP

    CONSTRAINT fk_tasks_author 
        FOREIGN KEY (author_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
CREATE INDEX IF NOT EXISTS idx_tasks_author ON tasks(author);
CREATE INDEX IF NOT EXISTS idx_tasks_created_at ON tasks(created_at);
CREATE INDEX IF NOT EXISTS idx_tasks_status_created ON tasks(status, created_at);

COMMENT ON TABLE tasks IS 'Task management';
COMMENT ON COLUMN tasks.status IS 'Task status: backlog, in_progress, done, cancelled';