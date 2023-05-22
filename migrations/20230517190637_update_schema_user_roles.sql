-- +goose Up
-- user_roles
CREATE TABLE user_roles (
    user_id integer REFERENCES users(id),
    role_id integer REFERENCES roles(id),

    PRIMARY KEY (user_id, role_id)
);
INSERT INTO user_roles SELECT id as user_id, role_id FROM users;

-- alter users
ALTER TABLE users DROP COLUMN role_id;
ALTER TABLE users ADD COLUMN description TEXT NOT NULL DEFAULT ('');
ALTER TABLE users ADD COLUMN last_session_created_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL;

-- alter roles
ALTER TABLE roles DROP COLUMN description;
ALTER TABLE roles ADD COLUMN change_message TEXT NOT NULL DEFAULT ('');
ALTER TABLE roles DROP COLUMN rights;
ALTER TABLE roles ADD COLUMN permissions JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE roles ADD COLUMN external_group TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users ADD COLUMN role_id REFERENCES roles(id);
UPDATE users u SET role_id = (SELECT role_id FROM user_roles WHERE user_id = u.id LIMIT 1);
DROP TABLE user_roles;
