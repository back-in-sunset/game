CREATE TABLE IF NOT EXISTS user_scan (
    bucket int,
    user_id bigint,
    created_at timestamp,
    PRIMARY KEY ((bucket), user_id)
) WITH CLUSTERING ORDER BY (user_id ASC);
