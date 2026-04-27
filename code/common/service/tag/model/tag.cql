Create keyspace IF NOT EXISTS tag with replication={'class':'SimpleStrategy', 'replication_factor':3};


CREATE TABLE IF NOT EXISTS tags (
    biz_type int
    biz_id bigint
    tag_id bigint,
    tag_name text,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY (biz_type, tag_id)
) WITH comment = '标签表';

CREATE TABLE IF NOT EXISTS object_tag_relations (
    obj_id uuid,
    obj_type int,
    obj_name varchar,
    tag_id int,
    tag_name varchar,
    PRIMARY KEY ((obj_id, obj_type), tag_id)
) WITH CLUSTERING ORDER BY (tag_id DESC);

CREATE MATERIALIZED VIEW tag_search AS
SELECT * FROM object_tag_relations
WHERE tag_name IS NOT NULL
PRIMARY KEY (tag_name, obj_id);

