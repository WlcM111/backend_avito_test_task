-- Создание таблицы команд
CREATE TABLE IF NOT EXISTS teams (
    team_name  TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Пользователи
CREATE TABLE IF NOT EXISTS users (
    user_id    TEXT PRIMARY KEY,
    username   TEXT NOT NULL,
    team_name  TEXT NOT NULL REFERENCES teams(team_name) ON DELETE RESTRICT,
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_team_active
    ON users (team_name, is_active);

-- Pull Requests
CREATE TABLE IF NOT EXISTS pull_requests (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL,
    author_id  TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    status     TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMPTZ,
    merged_at  TIMESTAMPTZ
);

-- Назначенные ревьюверы (0..2 на PR)
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id       TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    reviewer_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
    PRIMARY KEY (pr_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer
    ON pr_reviewers (reviewer_id);
