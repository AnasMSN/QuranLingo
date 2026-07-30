CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email             text NOT NULL UNIQUE,
    password_hash     text NOT NULL,
    display_name      text NOT NULL,
    total_xp          integer NOT NULL DEFAULT 0,
    hearts            integer NOT NULL DEFAULT 5,
    hearts_refill_at  timestamptz,
    current_streak    integer NOT NULL DEFAULT 0,
    longest_streak    integer NOT NULL DEFAULT 0,
    last_activity_date date,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  text NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
