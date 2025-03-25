-- +goose Up
update roles
set permissions = permissions || '[
    "profile_view",
    "profile_change_password",
    "application_group_view",
    "application_group_token_view",
    "app_access_view",
    "variable_view",
    "module_view"
]'
where name = 'admin' or name = 'read_only_admin';

update roles
set permissions = permissions || '[
    "application_group_add",
    "application_group_edit",
    "application_group_delete",
    "application_group_app_add",
    "application_group_app_edit",
    "application_group_app_delete",
    "application_group_token_add",
    "application_group_token_delete",
    "app_access_edit",
    "variable_add",
    "variable_edit",
    "variable_delete",
    "module_delete",
    "module_configuration_save_unsafe",
    "module_configuration_set_active",
    "module_history_set",
    "module_history_delete_version",
    "module_configuration_edit",
    "module_configuration_add"
]'
where name = 'admin';

-- +goose Down
