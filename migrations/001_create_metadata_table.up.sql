create table metadata
(
    client_id    uuid                    not null,
    created_at   timestamp not null default now(),
    updated_at   timestamp not null default now(),
    num_examples integer   not null,
    object_key   text                    not null
);

comment on column metadata.client_id is 'uuid пользака';

comment on column metadata.created_at is 'время создания записи';

comment on column metadata.updated_at is 'время обновления записи';

comment on column metadata.num_examples is 'сколько примеров было в обучении для этого пользака';

comment on column metadata.object_key is 'айдишник, по которому можно достать веса из minio';

create index metadata__idx_by_client_id
    on metadata (client_id);

comment on index metadata__idx_by_client_id is 'по клиенту';

create index metadata__idx_by_updated_at
    on metadata (updated_at);

comment on index metadata__idx_by_updated_at is 'по времени обновления';

