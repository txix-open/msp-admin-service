-- +goose Up
INSERT INTO roles (id, name) VALUES (2, 'read_only_admin');
SELECT setval(pg_get_serial_sequence('roles', 'id'), 3, FALSE);

-- +goose Down
DELETE FROM roles WHERE id = 2;
SELECT setval(pg_get_serial_sequence('roles', 'id'), 2, FALSE);
