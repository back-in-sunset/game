CREATE TABLE friend_relationship (
  id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  friend_id BIGINT NOT NULL,
  domain VARCHAR(16) NOT NULL DEFAULT 'platform',
  tenant_id VARCHAR(64) NOT NULL DEFAULT '',
  created_at BIGINT NOT NULL,
  PRIMARY KEY (user_id, friend_id, domain, tenant_id),
  KEY idx_id (id)
);

CREATE TABLE friend_request (
  id BIGINT NOT NULL,
  from_user_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  domain VARCHAR(16) NOT NULL DEFAULT 'platform',
  tenant_id VARCHAR(64) NOT NULL DEFAULT '',
  message VARCHAR(256) NOT NULL DEFAULT '',
  status TINYINT NOT NULL DEFAULT 0,
  created_at BIGINT NOT NULL,
  updated_at BIGINT NOT NULL,
  PRIMARY KEY (id),
  KEY idx_to_user_status (to_user_id, status, created_at),
  KEY idx_from_user (from_user_id, created_at)
);

CREATE TABLE friend_block (
  id BIGINT NOT NULL,
  user_id BIGINT NOT NULL,
  blocked_user_id BIGINT NOT NULL,
  domain VARCHAR(16) NOT NULL DEFAULT 'platform',
  tenant_id VARCHAR(64) NOT NULL DEFAULT '',
  created_at BIGINT NOT NULL,
  PRIMARY KEY (user_id, blocked_user_id, domain, tenant_id),
  KEY idx_id (id)
);
