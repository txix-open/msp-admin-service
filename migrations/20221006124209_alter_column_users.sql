-- +goose Up
ALTER TABLE users
    DROP CONSTRAINT users_phone_key,
    DROP COLUMN phone,
    DROP COLUMN image;

ALTER TABLE users ALTER COLUMN first_name SET DEFAULT '';
UPDATE users SET first_name=DEFAULT WHERE first_name IS NULL;
ALTER TABLE users ALTER COLUMN first_name SET NOT NULL;

ALTER TABLE users ALTER COLUMN last_name SET DEFAULT '';
UPDATE users SET last_name=DEFAULT WHERE last_name IS NULL;
ALTER TABLE users ALTER COLUMN last_name SET NOT NULL;

ALTER TABLE users ALTER COLUMN password SET DEFAULT '';
UPDATE users SET password=DEFAULT WHERE password IS NULL;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;

-- +goose Down
ALTER TABLE users
    ADD COLUMN phone      varchar(255) UNIQUE,
    ADD COLUMN image      varchar(255),
    ALTER COLUMN first_name DROP NOT NULL,
    ALTER COLUMN last_name DROP NOT NULL,
    ALTER COLUMN password DROP NOT NULL;


