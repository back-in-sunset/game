CREATE TABLE IF NOT EXISTS users (
    user_id bigint,
    name text,
    gender int,
    mobile text,
    email text,
    password text,
    created_at timestamp,
    PRIMARY KEY (user_id)
);

CREATE KEYSPACE user 
WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 3
};
