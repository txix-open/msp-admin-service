-- +goose Up
CREATE TABLE IF NOT EXISTS users (
  id serial8 PRIMARY KEY NOT NULL,
  role_id    integer,
  image      varchar(255),
  first_name varchar(255),
  last_name  varchar(255),
  email      varchar(255) UNIQUE NOT NULL,
  password   varchar(255),
  phone      varchar(255) UNIQUE,
  created_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL,
  updated_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE TRIGGER update_create_update_time BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE update_created_modified_column_date();

ALTER TABLE users
    ADD CONSTRAINT "FK_users_roleId_roles_id"
    FOREIGN KEY ("role_id") REFERENCES roles ("id")
    ON DELETE SET NULL ON UPDATE SET NULL;

CREATE INDEX IX_users_roleId ON users USING hash (role_id);

-- +goose Down
DROP TABLE users;