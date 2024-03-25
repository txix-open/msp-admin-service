-- +goose Up

CREATE TABLE audit_event
(
    event  text NOT NULL PRIMARY KEY,
    enable bool NOT NULL DEFAULT false
);

INSERT INTO audit_event (event, enable)
VALUES ('success_login', true),
       ('error_login', true),
       ('success_logout', true),
       ('role_changed', true),
       ('user_changed', true);

ALTER TABLE audit
    ADD COLUMN event text not null default '';

-- +goose Down
