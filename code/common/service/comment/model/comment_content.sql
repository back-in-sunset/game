CREATE TABLE `comment_content_0` (
  `comment_id` bigint unsigned NOT NULL COMMENT '同评论indx_id',
  `obj_id` bigint unsigned NOT NULL DEFAULT '0' COMMENT '评论对象ID使用唯一id的话不用type联合主键',
  `at_member_ids` text NOT NULL COMMENT 'at用户ID列表',
  `ip` varchar(255) NOT NULL COMMENT '评论IP',
  `platform` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '评论平台',
  `device` varchar(255) NOT NULL DEFAULT '' COMMENT '评论设备',
  `message` text NOT NULL COMMENT '评论内容',
  `meta` text NOT NULL COMMENT '评论元数据 背景 字体',
  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`comment_id`),
  KEY `idx_comment_obj_unique` (`comment_id`, `obj_id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COMMENT='评论点赞表[0-255]';


-- 创建存储过程生成256张分表
DELIMITER $$
CREATE PROCEDURE CreateCommentTables()
BEGIN
    DECLARE i INT DEFAULT 0;
    DECLARE table_name VARCHAR(64);
    
    WHILE i < 256 DO
        SET table_name = CONCAT('comment_content_', i);
        
        -- 动态生成建表语句
        SET @create_sql = CONCAT(
            'CREATE TABLE IF NOT EXISTS `', table_name, '` (',
            '  `comment_id` bigint unsigned NOT NULL COMMENT ''同评论indx_id'',',
            '  `obj_id` bigint unsigned NOT NULL DEFAULT ''0'' COMMENT ''评论对象ID使用唯一id的话不用type联合主键'',',
            '  `at_member_ids` text NOT NULL COMMENT ''at用户ID列表'',',
            '  `ip` varchar(255) NOT NULL DEFAULT '''' COMMENT ''评论IP'',',
            '  `platform` tinyint(3) unsigned NOT NULL DEFAULT ''0'' COMMENT ''评论平台'',',
            '  `device` varchar(255) NOT NULL DEFAULT '''' COMMENT ''评论设备'',',
            '  `message` text NOT NULL COMMENT ''评论内容'',',   -- 修正字段名：massage -> message
            '  `meta` text NOT NULL COMMENT ''评论元数据 背景 字体'',',
            '  `created_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP,',
            '  `updated_at` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,',
            '  PRIMARY KEY (`comment_id`),',
            '  KEY `idx_comment_obj_unique` (`comment_id`, `obj_id`)',
            ') ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT=''评论内容表[', i, ']'';'
        );
        
        -- 执行建表语句
        PREPARE stmt FROM @create_sql;
        EXECUTE stmt;
        DEALLOCATE PREPARE stmt;
        
        SET i = i + 1;
    END WHILE;
END$$
DELIMITER ;

-- 执行存储过程创建所有表
CALL CreateCommentTables();

-- 删除存储过程（可选）
DROP PROCEDURE IF EXISTS CreateCommentTables;