CREATE TABLE user_lesson_completions (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id        uuid NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    score            integer NOT NULL,
    xp_earned        integer NOT NULL,
    idempotency_key  text,
    completed_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, idempotency_key)
);

CREATE TABLE xp_transactions (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id               uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount                integer NOT NULL,
    reason                text NOT NULL,
    lesson_completion_id  uuid REFERENCES user_lesson_completions(id) ON DELETE SET NULL,
    created_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_lesson_completions_user_id ON user_lesson_completions(user_id);
CREATE INDEX idx_user_lesson_completions_lesson_id ON user_lesson_completions(lesson_id);
CREATE INDEX idx_xp_transactions_user_id_created_at ON xp_transactions(user_id, created_at);
