-- +goose Up
update roles
set permissions = '[
  "user_create",
  "user_update",
  "user_delete",
  "user_block",
  "role_view",
  "role_add",
  "role_update",
  "role_delete",
  "session_view",
  "session_revoke",
  "security_log_view",
  "audit_management_view"
]'
where name = 'admin';

update roles
set permissions = '[
  "role_view",
  "session_view",
  "security_log_view"
]'
where name = 'read_only_admin';

-- +goose Down
