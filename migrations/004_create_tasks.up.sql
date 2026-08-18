CREATE TABLE IF NOT EXISTS tasks (
    id          CHAR(36)        NOT NULL,
    team_id     CHAR(36)        NOT NULL,
    title       VARCHAR(255)    NOT NULL,
    description TEXT            NOT NULL,
    status      ENUM('pending', 'in_progress', 'done') NOT NULL DEFAULT 'pending',
    created_by  CHAR(36)        NOT NULL,
    assignee_id CHAR(36)        DEFAULT NULL,
    created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    closed_at   TIMESTAMP       NULL DEFAULT NULL,
    version     INT UNSIGNED    NOT NULL DEFAULT 1,
    PRIMARY KEY (id),
    CONSTRAINT fk_tasks_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    CONSTRAINT fk_tasks_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT fk_tasks_assignee FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
