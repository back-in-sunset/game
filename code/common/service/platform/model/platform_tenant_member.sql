CREATE TABLE `platform_tenant_member` (
  `tenant_member_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `member_id` varchar(64) NOT NULL,
  `role` varchar(32) NOT NULL DEFAULT 'developer',
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`tenant_member_id`),
  UNIQUE KEY `uk_platform_tenant_member` (`tenant_id`, `member_id`),
  KEY `idx_platform_member_status` (`member_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
