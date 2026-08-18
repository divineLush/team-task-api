CREATE TABLE IF NOT EXISTS teams (
    id         CHAR(36)        NOT NULL,
    name       VARCHAR(255)    NOT NULL,
    created_by CHAR(36)        NOT NULL,
    created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    CONSTRAINT fk_teams_created_by FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
