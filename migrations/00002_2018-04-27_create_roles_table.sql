-- +goose Up
CREATE TABLE IF NOT EXISTS roles (
  id serial4 PRIMARY KEY NOT NULL,
  name varchar(64) NOT NULL,
  rights jsonb NOT NULL DEFAULT '{}',
  description text,
  created_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL,
  updated_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE TRIGGER update_create_update_time BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE PROCEDURE update_created_modified_column_date();

-- +goose Down
DROP TABLE roles;