-- +goose Up
ALTER TABLE users
    ADD COLUMN sudir_user_id TEXT UNIQUE NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN sudir_user_id;
