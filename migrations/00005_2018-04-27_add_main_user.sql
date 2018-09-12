-- +goose Up
INSERT INTO roles (id, name) VALUES (1, 'admin');
-- password: Xnp5kRZh
INSERT INTO users (id, role_id, first_name, last_name, email, password)
    VALUES (1, 1, 'admin', 'admin', 'admin@ex.com', '$2y$12$6KXNZ8VXwiOu91.ZixqYpeYnRhGxmKuhEI0YB44.z8v0rhBVUywdu');
INSERT INTO users (id, role_id, first_name, last_name, email, password)
    VALUES (2, 1, 'admin_second', 'admin_second', 'admin2@ex.com', '$2y$12$6KXNZ8VXwiOu91.ZixqYpeYnRhGxmKuhEI0YB44.z8v0rhBVUywdu');
SELECT setval(pg_get_serial_sequence('roles', 'id'), 2, FALSE);
SELECT setval(pg_get_serial_sequence('users', 'id'), 3, FALSE);

-- +goose Down
DELETE FROM users WHERE id IN (1, 2);
DELETE FROM roles WHERE id = 1;