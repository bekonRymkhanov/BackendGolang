CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    favoriteBooks text[],
    name text NOT NULL,
    activated boolean NOT NULL DEFAULT false,
    is_admin boolean NOT NULL DEFAULT false,
    email citext UNIQUE NOT NULL,
    password_hash bytea NOT NULL,
    version integer NOT NULL DEFAULT 1
);
