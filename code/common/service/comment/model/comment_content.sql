CREATE TABLE `comment_content` (
  `comment_id` bigint unsigned NOT NULL COMMENT '同评论indx_id',
  `at_member_ids` text NOT NULL DEFAULT '' COMMENT 'at用户ID列表',
  `ip` varchar(255) NOT NULL DEFAULT '' COMMENT '评论IP',
  `platform` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '评论平台',
  `device` varchar(255) NOT NULL DEFAULT '' COMMENT '评论设备',
  `massage` text NOT NULL DEFAULT '' COMMENT '评论内容',
  `meta` text NOT NULL DEFAULT '0' COMMENT '评论元数据 背景 字体',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`comment_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='评论点赞表[0-255]';