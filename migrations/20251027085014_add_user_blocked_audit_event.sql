-- +goose Up
INSERT INTO audit_event (event, enable)
VALUES
       ('user_blocked', true);

-- +goose Down
DELETE FROM audit_event WHERE event = 'user_blocked';
