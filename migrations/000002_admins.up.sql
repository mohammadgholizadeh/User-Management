CREATE TABLE IF NOT EXISTS admins (
    user_national_id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW())
);
