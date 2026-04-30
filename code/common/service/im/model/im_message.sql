CREATE TABLE IF NOT EXISTS `im_message` (
  `id` BIGINT NOT NULL AUTO_INCREMENT,
  `domain` VARCHAR(16) NOT NULL,
  `tenant_id` VARCHAR(64) NOT NULL DEFAULT '',
  `project_id` VARCHAR(64) NOT NULL DEFAULT '',
  `environment` VARCHAR(64) NOT NULL DEFAULT '',
  `conversation_key` VARCHAR(255) NOT NULL,
  `sender` BIGINT NOT NULL,
  `receiver` BIGINT NOT NULL,
  `msg_type` VARCHAR(32) NOT NULL,
  `seq` BIGINT NOT NULL,
  `payload_json` JSON NOT NULL,
  `sent_at` DATETIME(3) NOT NULL,
  `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  PRIMARY KEY (`id`),
  KEY `idx_im_message_conversation_sent` (`conversation_key`, `sent_at`, `id`),
  KEY `idx_im_message_receiver_sent` (`receiver`, `sent_at`, `id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
