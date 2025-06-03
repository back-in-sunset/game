CREATE TABLE `comment_index` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `obj_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '评论对象ID使用唯一id的话不用type联合主键',
  `obj_type` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '评论对象类型',
  `member_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '作者用户ID',
  `root_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '根评论ID 不为0表示是回复评论',
  `reply_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '被回复的评论ID',
  `floor` bigint unsigned NOT NULL DEFAULT '0' COMMENT '评论楼层',
  `count` int(11) NOT NULL DEFAULT '0' COMMENT '挂载子评论总数 可见',
  `root_count` int(11) NOT NULL DEFAULT '0' COMMENT '挂载子评论总数 所以',
  `like_count` int(11) NOT NULL DEFAULT '0' COMMENT '点赞数',
  `hate_count` int(11) NOT NULL DEFAULT '0' COMMENT '点踩数',
  `state` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '0-正常, 1-隐藏',
  `attrs` int(11) NOT NULL DEFAULT '0' COMMENT '属性(bit 0-运营置顶, 1-owner置顶 2-大数据)',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_state_attrs_obj_type_unique` (`state`, `attrs`, `obj_id`, `obj_type`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='评论表[0-255] obj_id bitmod sharding';



-- -- 创建存储过程生成256张分表
-- DELIMITER $$
-- CREATE PROCEDURE CreateCommentIndexTables()
-- BEGIN
--     DECLARE i INT DEFAULT 0;
--     DECLARE table_name VARCHAR(64);
    
--     WHILE i < 256 DO
--         SET table_name = CONCAT('comment_index_', i);
        
--         -- 动态生成建表语句
--         SET @create_sql = CONCAT(
--             'CREATE TABLE IF NOT EXISTS `', table_name, '` (',
--             '  `id` bigint unsigned NOT NULL AUTO_INCREMENT,',
--             '  `obj_id` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''评论对象ID'',',
--             '  `obj_type` tinyint(3) unsigned NOT NULL DEFAULT ''0'' COMMENT ''评论对象类型'',',
--             '  `member_id` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''作者用户ID'',',
--             '  `root_id` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''根评论ID'',',
--             '  `reply_id` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''被回复的评论ID'',',
--             '  `floor` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''评论楼层'',',
--             '  `count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''可见子评论数'',',
--             '  `root_count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''所有子评论数'',',
--             '  `like_count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''点赞数'',',
--             '  `hate_count` int(11) NOT NULL DEFAULT ''0'' COMMENT ''点踩数'',',
--             '  `state` tinyint(3) unsigned NOT NULL DEFAULT ''0'' COMMENT ''0-正常,1-隐藏'',',
--             '  `attrs` int(11) NOT NULL DEFAULT ''0'' COMMENT ''属性(bit:0-运营置顶,1-owner置顶,2-大数据)'',',
--             '  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,',
--             '  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,',
--             '  PRIMARY KEY (`id`),',
--             '  UNIQUE KEY `idx_state_attrs_obj_type_unique` (`state`, `attrs`, `obj_id`, `obj_type`)',
--             ') ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT=''评论索引表[', i, ']'';'
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
-- CALL CreateCommentIndexTables();

-- -- 删除存储过程（可选）
-- DROP PROCEDURE IF EXISTS CreateCommentIndexTables;