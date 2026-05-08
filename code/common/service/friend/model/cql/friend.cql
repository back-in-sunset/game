-- Friend service ScyllaDB schema (reserved for future migration)

CREATE TABLE friend_relationship (
    user_id bigint,
    friend_id bigint,
    domain text,
    tenant_id text,
    created_at timestamp,
    PRIMARY KEY ((user_id), friend_id)
);

CREATE TABLE friend_relationship_by_friend (
    user_id bigint,
    friend_id bigint,
    domain text,
    tenant_id text,
    created_at timestamp,
    PRIMARY KEY ((friend_id), user_id)
);

CREATE TABLE friend_request_inbox (
    to_user_id bigint,
    status tinyint,
    created_at timestamp,
    request_id bigint,
    from_user_id bigint,
    message text,
    PRIMARY KEY ((to_user_id), status, created_at, request_id)
) WITH CLUSTERING ORDER BY (status ASC, created_at DESC);

CREATE TABLE friend_request_outbox (
    from_user_id bigint,
    created_at timestamp,
    request_id bigint,
    to_user_id bigint,
    status tinyint,
    message text,
    PRIMARY KEY ((from_user_id), created_at, request_id)
) WITH CLUSTERING ORDER BY (created_at DESC);

CREATE TABLE friend_request_by_id (
    request_id bigint PRIMARY KEY,
    from_user_id bigint,
    to_user_id bigint,
    domain text,
    tenant_id text,
    status tinyint,
    message text,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE friend_block (
    user_id bigint,
    blocked_user_id bigint,
    domain text,
    tenant_id text,
    created_at timestamp,
    PRIMARY KEY ((user_id), blocked_user_id)
);

CREATE TABLE friend_block_by_target (
    blocked_user_id bigint,
    user_id bigint,
    domain text,
    tenant_id text,
    created_at timestamp,
    PRIMARY KEY ((blocked_user_id), user_id)
);
