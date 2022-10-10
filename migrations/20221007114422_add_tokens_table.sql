-- +goose Up
CREATE TABLE tokens
(
    token      TEXT PRIMARY KEY,
    user_id    INT8      NOT NULL,
    status     TEXT      NOT NULL,
    expired_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);


-- +goose Down
DROP TABLE tokens;
