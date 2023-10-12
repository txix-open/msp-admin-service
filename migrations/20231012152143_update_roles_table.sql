-- +goose Up
ALTER TABLE roles
    ADD COLUMN immutable BOOL NOT NULL DEFAULT false,
    ADD COLUMN exclusive BOOL NOT NULL DEFAULT false;

-- +goose Down
