CREATE TABLE `comment_subject` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `obj_id` bigint unsigned NOT NULL DEFAULT '' COMMENT '评论对象ID 使用唯一id的话不用type联合主键',
  `obj_type` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '评论对象类型',
  `member_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '作者用户ID',
  `count`  int(11) NOT NULL DEFAULT '0' COMMENT '评论总数',
  `root_count` int(11) NOT NULL DEFAULT '0' COMMENT '根评论总数',
  `all_count` int(11) NOT NULL DEFAULT '0' COMMENT '所有评论+回复总数',
  `state` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '0-正常, 1-隐藏',
  `attrs` int(11) NOT NULL DEFAULT '0' COMMENT '属性(bit 0-运营置顶, 1-owner置顶 2-大数据)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY  `idx_obj_type_unique` (`state`, `obj_id`, `obj_type`),
  UNIQUE KEY `idx_member_unique` (`state`, `member_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='评论主题表[0-49]';
