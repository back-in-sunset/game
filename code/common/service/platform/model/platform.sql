CREATE TABLE `platform_tenant` (
  `tenant_id` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  `slug` varchar(64) NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`tenant_id`),
  UNIQUE KEY `uk_platform_tenant_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_project` (
  `project_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `name` varchar(128) NOT NULL,
  `project_key` varchar(64) NOT NULL,
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`project_id`),
  UNIQUE KEY `uk_platform_project_tenant_key` (`tenant_id`, `project_key`),
  CONSTRAINT `fk_platform_project_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_environment` (
  `environment_id` varchar(64) NOT NULL,
  `project_id` varchar(64) NOT NULL,
  `name` varchar(64) NOT NULL,
  `display_name` varchar(128) NOT NULL DEFAULT '',
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`environment_id`),
  UNIQUE KEY `uk_platform_environment_project_name` (`project_id`, `name`),
  CONSTRAINT `fk_platform_environment_project` FOREIGN KEY (`project_id`) REFERENCES `platform_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

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
  KEY `idx_platform_member_status` (`member_id`, `status`),
  CONSTRAINT `fk_platform_tenant_member_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_tenant_invitation` (
  `invitation_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `inviter_member_id` varchar(64) NOT NULL,
  `invitee_member_id` varchar(64) NOT NULL,
  `role` varchar(32) NOT NULL DEFAULT 'developer',
  `status` varchar(32) NOT NULL DEFAULT 'pending',
  `expired_at` timestamp NULL DEFAULT NULL,
  `accepted_at` timestamp NULL DEFAULT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`invitation_id`),
  UNIQUE KEY `uk_platform_tenant_invitee_pending` (`tenant_id`, `invitee_member_id`, `status`),
  KEY `idx_platform_invitee_status` (`invitee_member_id`, `status`),
  CONSTRAINT `fk_platform_tenant_invitation_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_project_member` (
  `project_member_id` varchar(64) NOT NULL,
  `project_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `member_id` varchar(64) NOT NULL,
  `role` varchar(32) NOT NULL DEFAULT 'developer',
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `joined_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`project_member_id`),
  UNIQUE KEY `uk_platform_project_member` (`project_id`, `member_id`),
  KEY `idx_platform_project_member_tenant` (`tenant_id`, `member_id`, `status`),
  CONSTRAINT `fk_platform_project_member_project` FOREIGN KEY (`project_id`) REFERENCES `platform_project` (`project_id`),
  CONSTRAINT `fk_platform_project_member_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_member_role_history` (
  `history_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `member_id` varchar(64) NOT NULL,
  `scope_type` varchar(32) NOT NULL DEFAULT 'tenant',
  `old_role` varchar(32) DEFAULT NULL,
  `new_role` varchar(32) NOT NULL,
  `operator_member_id` varchar(64) NOT NULL,
  `reason` varchar(255) NOT NULL DEFAULT '',
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`history_id`),
  KEY `idx_platform_member_role_history_member` (`member_id`, `created_at`),
  KEY `idx_platform_member_role_history_scope` (`tenant_id`, `project_id`, `scope_type`),
  CONSTRAINT `fk_platform_member_role_history_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`),
  CONSTRAINT `fk_platform_member_role_history_project` FOREIGN KEY (`project_id`) REFERENCES `platform_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_access_token` (
  `token_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) NOT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `name` varchar(128) NOT NULL,
  `token_hash` varchar(128) NOT NULL,
  `scopes` varchar(255) NOT NULL DEFAULT '',
  `status` varchar(32) NOT NULL DEFAULT 'active',
  `expired_at` timestamp NULL DEFAULT NULL,
  `last_used_at` timestamp NULL DEFAULT NULL,
  `created_by` varchar(64) NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`token_id`),
  UNIQUE KEY `uk_platform_access_token_hash` (`token_hash`),
  KEY `idx_platform_access_token_scope` (`tenant_id`, `project_id`, `status`),
  CONSTRAINT `fk_platform_access_token_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`),
  CONSTRAINT `fk_platform_access_token_project` FOREIGN KEY (`project_id`) REFERENCES `platform_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `platform_audit_log` (
  `audit_id` varchar(64) NOT NULL,
  `tenant_id` varchar(64) DEFAULT NULL,
  `project_id` varchar(64) DEFAULT NULL,
  `operator_member_id` varchar(64) NOT NULL,
  `action` varchar(64) NOT NULL,
  `resource_type` varchar(64) NOT NULL,
  `resource_id` varchar(64) NOT NULL,
  `result` varchar(32) NOT NULL DEFAULT 'success',
  `ip` varchar(64) NOT NULL DEFAULT '',
  `user_agent` varchar(255) NOT NULL DEFAULT '',
  `meta_json` text,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`audit_id`),
  KEY `idx_platform_audit_operator` (`operator_member_id`, `created_at`),
  KEY `idx_platform_audit_scope` (`tenant_id`, `project_id`, `created_at`),
  KEY `idx_platform_audit_resource` (`resource_type`, `resource_id`, `created_at`),
  CONSTRAINT `fk_platform_audit_log_tenant` FOREIGN KEY (`tenant_id`) REFERENCES `platform_tenant` (`tenant_id`),
  CONSTRAINT `fk_platform_audit_log_project` FOREIGN KEY (`project_id`) REFERENCES `platform_project` (`project_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
