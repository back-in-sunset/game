Create keyspace IF NOT EXISTS tag with replication={'class':'SimpleStrategy', 'replication_factor':3}；

CREATE TABLE IF NOT EXISTS tags (
    id bigint,
    tag_id bigint,
    tag_name varchar,
    PRIMARY KEY (id, tag_id)
);

CREATE TABLE IF NOT EXISTS object_tag_relations (
    obj_id uuid,
    obj_type int,
    obj_name varchar,
    tag_id int,
    tag_name varchar,
    PRIMARY KEY ((obj_id, obj_type), tag_id)
) WITH CLUSTERING ORDER BY (tag_id ASC);

CREATE TABLE IF NOT EXISTS organizations (
    id int,
    org_id int,
    org_name varchar,
    parent_org_id int,
    org_path varchar,
    org_name_path varchar,
    is_leaf tinyint,
    PRIMARY KEY (id, org_id)
) WITH CLUSTERING ORDER BY (org_id ASC);

CREATE INDEX IF NOT EXISTS parent_org_idx ON organizations(parent_org_id);

CREATE TABLE IF NOT EXISTS categories (
    category_id int PRIMARY KEY,
    category_name varchar
) WITH compaction = {'class': 'LeveledCompactionStrategy'};

CREATE TABLE IF NOT EXISTS object_categories (
    obj_id uuid,
    obj_type int,
    obj_name text,
    category_id int,
    category_name text,
    PRIMARY KEY ((obj_id, obj_type), category_id)
) WITH CLUSTERING ORDER BY (category_id ASC)
   AND compaction = {'class': 'TimeWindowCompactionStrategy'}
   AND compression = {'sstable_compression': 'ZstdCompressor'};
