-- +goose Up
alter table roles
    add unique (name);

alter table tokens
    add column id serial8 not null;

alter table tokens
    add constraint fk_user_id__user_id
        foreign key (user_id) references users (id)
            on update cascade on delete cascade;

create index ix_tokens__user_id on tokens (user_id);

create table audit
(
    id         serial8 primary key,
    user_id    int8      not null,
    message    text      not null,
    created_at timestamp not null
);

create index ix_audit__created_at on audit (created_at DESC);

alter table users
    add column blocked boolean not null default false;

alter table users
    add unique (sudir_user_id);

-- +goose Down
