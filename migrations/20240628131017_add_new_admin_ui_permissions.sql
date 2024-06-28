-- +goose Up
update roles
set permissions = permissions || '[
  "read", "write"
]'::jsonb
where name = 'admin';

update roles
set permissions = permissions || '[
  "read"
]'::jsonb
where name = 'read_only_admin';

-- +goose Down
