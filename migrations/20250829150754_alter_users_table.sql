-- +goose Up
ALTER TABLE users DROP CONSTRAINT users_email_key;
ALTER TABLE users ADD CONSTRAINT users_email_sudir_user_id_key UNIQUE (email, sudir_user_id);

-- +goose Down
ALTER TABLE users DROP CONSTRAINT users_email_sudir_user_id_key;
ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email);
