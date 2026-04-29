CREATE TABLE `user` (
  `user_id` bigint NOT NULL,
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '用户姓名',
  `gender` tinyint(3) NOT NULL DEFAULT '0' COMMENT '用户性别',
  `mobile` varchar(255) NOT NULL DEFAULT '' COMMENT '用户电话',
  `email` varchar(255) NOT NULL DEFAULT '' COMMENT '用户邮箱',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '用户密码',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`),
  UNIQUE KEY `idx_mobile_unique` (`mobile`),
  UNIQUE KEY `idx_email_unique_non_empty` ((NULLIF(`email`, '')))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_mobile_index` (
  `mobile` varchar(255) NOT NULL COMMENT '手机号',
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`mobile`),
  UNIQUE KEY `idx_user_id_unique` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `user_profile` (
  `user_id` bigint NOT NULL COMMENT '用户ID',
  `avatar` varchar(1024) NOT NULL DEFAULT '' COMMENT '头像URL',
  `bio` varchar(1024) NOT NULL DEFAULT '' COMMENT '个人简介',
  `birthday` date DEFAULT NULL COMMENT '生日',
  `location` varchar(255) NOT NULL DEFAULT '' COMMENT '地区',
  `extra` json DEFAULT NULL COMMENT '扩展信息',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
