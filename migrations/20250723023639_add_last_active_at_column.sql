-- +goose Up
ALTER TABLE users
    ADD COLUMN last_active_at TIMESTAMP DEFAULT (now() at time zone 'utc') NOT NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN last_active_at;
