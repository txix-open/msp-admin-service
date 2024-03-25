-- +goose Up
-- user_roles
CREATE TABLE user_roles (
    user_id integer,
    role_id integer,

    PRIMARY KEY (user_id, role_id),

    CONSTRAINT FK_roles_id__role_id FOREIGN KEY (role_id) REFERENCES roles (id),
    CONSTRAINT FK_users_id__user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
INSERT INTO user_roles SELECT id as user_id, role_id FROM users;

-- alter users
ALTER TABLE users DROP COLUMN role_id;
ALTER TABLE users ADD COLUMN description TEXT NOT NULL DEFAULT ('');

-- alter roles
ALTER TABLE roles DROP COLUMN description;
ALTER TABLE roles DROP COLUMN rights;
ALTER TABLE roles ADD COLUMN permissions JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE roles ADD COLUMN external_group TEXT NOT NULL DEFAULT '';

-- +goose Down
