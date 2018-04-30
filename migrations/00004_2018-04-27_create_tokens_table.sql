-- +goose Up
CREATE TABLE IF NOT EXISTS tokens (
  id serial8 NOT NULL PRIMARY KEY,
  user_id int8 NOT NULL UNIQUE,
  token text NOT NULL,
  expired_at timestamp DEFAULT (now() at time zone 'utc') + '30 days'::interval NOT NULL,
  created_at timestamp DEFAULT (now() at time zone 'utc') NOT NULL
);

CREATE TRIGGER update_create_update_time BEFORE UPDATE ON tokens
    FOR EACH ROW EXECUTE PROCEDURE update_created_column_date();

ALTER TABLE tokens
    ADD CONSTRAINT "FK_users_userId_users_id"
    FOREIGN KEY ("user_id") REFERENCES users ("id")
    ON DELETE CASCADE ON UPDATE CASCADE;

CREATE INDEX IX_tokens_userId ON tokens USING hash (user_id);
CREATE INDEX IX_tokens_token ON tokens USING btree (token);

-- +goose Down
DROP TABLE tokens;
