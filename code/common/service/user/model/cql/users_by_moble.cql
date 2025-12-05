CREATE TABLE IF NOT EXISTS users_by_mobile (
    mobile text,
    user_id bigint,
    created_at timestamp,
    PRIMARY KEY (mobile)
);
