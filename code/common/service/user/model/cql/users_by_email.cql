CREATE TABLE IF NOT EXISTS users_by_email (
    email text,
    user_id bigint,
    created_at timestamp,
    PRIMARY KEY (email)
);
