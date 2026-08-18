CREATE TABLE IF NOT EXISTS team_members (
    team_id   CHAR(36)        NOT NULL,
    user_id   CHAR(36)        NOT NULL,
    role      ENUM('owner', 'admin', 'member') NOT NULL DEFAULT 'member',
    PRIMARY KEY (team_id, user_id),
    CONSTRAINT fk_team_members_team FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE,
    CONSTRAINT fk_team_members_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
