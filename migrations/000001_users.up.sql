CREATE TYPE roles AS ENUM ('user', 'admin', 'manager');
CREATE TABLE users (
    national_id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    username TEXT NOT NULL,
    mobile_number TEXT NOT NULL,
    gender SMALLINT NOT NULL,
    email TEXT NOT NULL,
    role  roles DEFAULT 'user' NOT null,
    hashed_password TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT (TRUE),
    created_at TIMESTAMPTZ NOT NULL DEFAULT (NOW()),
    updated_at TIMESTAMPTZ
);