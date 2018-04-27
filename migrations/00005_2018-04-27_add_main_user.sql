-- +goose Up
INSERT INTO roles (id, name) VALUES (1, 'admin');
-- password: Xnp5kRZh
INSERT INTO users (id, role_id, first_name, last_name, email, password)
    VALUES (1, 1, 'admin', 'admin', 'admin@ex.com', '$2y$12$6KXNZ8VXwiOu91.ZixqYpeYnRhGxmKuhEI0YB44.z8v0rhBVUywdu');

-- +goose Down
DELETE FROM users WHERE id = 1;
DELETE FROM roles WHERE id = 1;