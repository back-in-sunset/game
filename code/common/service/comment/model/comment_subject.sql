CREATE TABLE `comment_subject_0` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `obj_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '评论对象ID 使用唯一id的话不用type联合主键',
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
  UNIQUE KEY  `idx_obj_type_unique` (`state`, `attrs`, `obj_id`, `obj_type`),
  UNIQUE KEY `idx_member_unique` (`state`, `attrs`, `member_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='评论主题表[0-63] obj_id bitmod sharding';


-- -- 创建存储过程生成64张分表
-- DELIMITER $$
-- CREATE PROCEDURE CreateCommentSubjectTables()
-- BEGIN
--     DECLARE i INT DEFAULT 0;
--     DECLARE table_name VARCHAR(64);
    
--     WHILE i < 64 DO
--         SET table_name = CONCAT('comment_subject_', i);
        
--         -- 动态生成建表语句（修正 obj_id 默认值为数字 0）
--         SET @create_sql = CONCAT(
--             'CREATE TABLE IF NOT EXISTS `', table_name, '` (',
--             '  `id` bigint unsigned NOT NULL AUTO_INCREMENT,',
--             '  `obj_id` bigint unsigned NOT NULL DEFAULT 0 COMMENT ''评论对象ID'',',
--             '  `obj_type` tinyint(3) unsigned NOT NULL DEFAULT ''0'' COMMENT ''评论对象类型'',',
--             '  `member_id` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''作者用户ID'',',
--             '  `count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''评论总数'',',
--             '  `root_count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''根评论总数'',',
--             '  `all_count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''所有评论+回复总数'',',
--             '  `state` tinyint(3) unsigned NOT NULL DEFAULT ''0'' COMMENT ''0-正常,1-隐藏'',',
--             '  `attrs` int(11) NOT NULL DEFAULT ''0'' COMMENT ''属性(bit:0-运营置顶,1-owner置顶,2-大数据)'',',
--             '  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,',
--             '  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,',
--             '  PRIMARY KEY (`id`),',
--             '  UNIQUE KEY `idx_obj_type_unique` (`state`, `attrs`, `obj_id`, `obj_type`),',
--             '  UNIQUE KEY `idx_member_unique` (`state`, `attrs`, `member_id`)',
--             ') ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT=''评论主题表[', i, ']'';'
--         );
        
--         -- 执行建表语句
--         PREPARE stmt FROM @create_sql;
--         EXECUTE stmt;
--         DEALLOCATE PREPARE stmt;
        
--         SET i = i + 1;
--     END WHILE;
-- END$$
-- DELIMITER ;

-- -- 执行存储过程创建所有表
-- CALL CreateCommentSubjectTables();

-- -- 删除存储过程（可选）
-- DROP PROCEDURE IF EXISTS CreateCommentSubjectTables;