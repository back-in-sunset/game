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
