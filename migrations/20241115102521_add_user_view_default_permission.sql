-- +goose Up
update roles
set permissions = permissions || '["user_view"]'
where name = 'admin' or name = 'read_only_admin';


-- +goose Down

