CREATE TABLE `tag` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tag_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '标签id',
  `tag_name` varchar(32) NOT NULL DEFAULT '' COMMENT '标签名称',
  `biz_type` int NOT NULL DEFAULT '0' COMMENT '业务类型',
  `biz_id` bigint NOT NULL DEFAULT '0' COMMENT '业务id',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `idx_tag_unique` (`tag_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='标签';
