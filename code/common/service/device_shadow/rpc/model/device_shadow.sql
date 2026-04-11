CREATE TABLE IF NOT EXISTS `device_shadow` (
  `device_id` bigint(20) NOT NULL,
  `device_name` varchar(255) NOT NULL,
  `desired_state` text,
  `reported_state` text,
  `delta_state` text,
  `version` bigint(20) DEFAULT '0',
  `created_at` bigint(20) NOT NULL,
  `updated_at` bigint(20) NOT NULL,
  PRIMARY KEY (`device_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
