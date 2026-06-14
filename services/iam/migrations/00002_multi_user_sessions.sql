-- +goose Up
ALTER TABLE user_sessions
    DROP CONSTRAINT IF EXISTS user_sessions_user_id_key;

ALTER TABLE user_sessions
    ADD CONSTRAINT user_sessions_refresh_token_hash_key UNIQUE (refresh_token_hash);

CREATE INDEX IF NOT EXISTS user_sessions_user_id_idx ON user_sessions(user_id);

-- +goose Down
DROP INDEX IF EXISTS user_sessions_user_id_idx;

ALTER TABLE user_sessions
    DROP CONSTRAINT IF EXISTS user_sessions_refresh_token_hash_key;

ALTER TABLE user_sessions
    ADD CONSTRAINT user_sessions_user_id_key UNIQUE (user_id);
