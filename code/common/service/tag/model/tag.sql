CREATE TABLE `tag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tag_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '标签id',
  `tag_name` varchar(32) NOT NULL DEFAULT '' COMMENT '标签名称',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tag_unique` (`tag_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='标签';
