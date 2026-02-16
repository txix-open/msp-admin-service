-- +goose Up
ALTER TABLE users ADD COLUMN full_name VARCHAR(255) DEFAULT '' NOT NULL;

UPDATE users u SET full_name=u.first_name
    WHERE (u.last_name = '') AND (u.first_name <> '');

UPDATE users u SET full_name=u.last_name || ' ' || u.first_name
    WHERE (u.last_name <> '') AND (u.first_name <> '');

-- +goose Down
ALTER TABLE users DROP COLUMN full_name;
