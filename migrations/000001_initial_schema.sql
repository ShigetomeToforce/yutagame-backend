-- +goose Up
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
CREATE TABLE `admins` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `password` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `role_type` enum('ADMIN','USER') CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_admins_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `announcements` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` longtext NOT NULL,
  `excerpt` longtext NOT NULL,
  `body_html` longtext NOT NULL,
  `status` longtext NOT NULL,
  `published_at` datetime(3) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `publish_start_at` datetime(3) DEFAULT NULL,
  `publish_end_at` datetime(3) DEFAULT NULL,
  `display_order` bigint NOT NULL DEFAULT '0',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `banners` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `title` longtext NOT NULL,
  `placement` varchar(191) NOT NULL,
  `image_key` longtext NOT NULL,
  `link_url` longtext,
  `open_in_new_tab` tinyint(1) NOT NULL,
  `starts_at` datetime(3) DEFAULT NULL,
  `ends_at` datetime(3) DEFAULT NULL,
  `display_order` bigint NOT NULL,
  `click_count` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_banners_placement` (`placement`),
  KEY `idx_banners_display_order` (`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `contact_inquiries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` longtext NOT NULL,
  `email` longtext NOT NULL,
  `subject` longtext NOT NULL,
  `message` longtext NOT NULL,
  `status` longtext NOT NULL,
  `admin_note` longtext NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `content_access_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `content_type` varchar(32) NOT NULL,
  `content_key` varchar(191) NOT NULL,
  `visitor_id` varchar(128) NOT NULL DEFAULT '',
  `ip_hash` char(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_content_access_target_day` (`content_type`,`content_key`,`created_at`),
  KEY `idx_content_access_logs_visitor_id` (`visitor_id`),
  KEY `idx_content_access_logs_ip_hash` (`ip_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `feature_games` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `feature_id` bigint NOT NULL,
  `game_id` bigint NOT NULL,
  `display_order` bigint NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_feature_games_feature_order` (`feature_id`,`display_order`),
  KEY `idx_feature_games_game_id` (`game_id`),
  CONSTRAINT `fk_features_feature_games` FOREIGN KEY (`feature_id`) REFERENCES `features` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `features` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `code` varchar(191) NOT NULL,
  `title` longtext NOT NULL,
  `excerpt` longtext NOT NULL,
  `body_html` longtext NOT NULL,
  `thumbnail_image_key` longtext,
  `status` longtext NOT NULL,
  `published_at` datetime(3) DEFAULT NULL,
  `publish_start_at` datetime(3) DEFAULT NULL,
  `display_order` bigint NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  `publish_end_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_features_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_affiliates` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `game_id` bigint NOT NULL,
  `category` enum('AMAZON','RAKUTEN','YAHOO','SURUGAYA','PLAYSTATION_STORE','NINTENDO_STORE','STEAM') NOT NULL,
  `url` varchar(1024) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uniq_game_affiliate_category` (`game_id`,`category`) USING BTREE,
  CONSTRAINT `fk_game_affiliates_games` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_games_affiliates` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ゲーム購入導線';
CREATE TABLE `game_favorites` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `game_id` bigint NOT NULL,
  `visitor_id` varchar(64) NOT NULL,
  `favorite_date` date NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `ux_game_favorites_game_visitor_date` (`game_id`,`visitor_id`,`favorite_date`),
  KEY `idx_game_favorites_game_id` (`game_id`),
  KEY `idx_game_favorites_visitor_id` (`visitor_id`),
  KEY `idx_game_favorites_favorite_date` (`favorite_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_keywords` (
  `game_id` bigint NOT NULL,
  `keyword_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`game_id`,`keyword_id`),
  KEY `idx_game_keywords_keyword_id` (`keyword_id`),
  CONSTRAINT `fk_game_keywords_game` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`),
  CONSTRAINT `fk_game_keywords_games` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_game_keywords_keyword` FOREIGN KEY (`keyword_id`) REFERENCES `keywords` (`id`),
  CONSTRAINT `fk_game_keywords_keywords` FOREIGN KEY (`keyword_id`) REFERENCES `keywords` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_ranking_active_entries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `game_id` bigint NOT NULL,
  `display_rank` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_game_ranking_active_game_id` (`game_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_ranking_draft_entries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `game_id` bigint NOT NULL,
  `display_rank` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_game_ranking_draft_game_id` (`game_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_ranking_previous_entries` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `game_id` bigint NOT NULL,
  `display_rank` int NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_game_ranking_previous_game_id` (`game_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_recommendations` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `game_name` longtext NOT NULL,
  `reason` longtext NOT NULL,
  `status` longtext NOT NULL,
  `admin_note` longtext NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `game_view_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `visitor_id` varchar(128) NOT NULL DEFAULT '',
  `ip_hash` char(64) NOT NULL DEFAULT '',
  `game_code` varchar(64) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_game_view_logs_visitor_id` (`visitor_id`),
  KEY `idx_game_view_logs_ip_hash` (`ip_hash`),
  KEY `idx_game_view_logs_game_code` (`game_code`),
  KEY `idx_game_view_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `games` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(256) NOT NULL,
  `kana` varchar(256) NOT NULL,
  `overview` varchar(2000) NOT NULL,
  `code` varchar(30) NOT NULL,
  `image_key` varchar(255) DEFAULT NULL,
  `manufacturer_id` bigint NOT NULL,
  `machine_id` bigint NOT NULL,
  `genre_id` bigint NOT NULL,
  `sub_genre` varchar(256) NOT NULL,
  `catch_copy` varchar(256) NOT NULL,
  `sub_catch` varchar(512) NOT NULL,
  `list_price` int NOT NULL,
  `release_date` datetime(3) NOT NULL,
  `official_site_url` varchar(256) NOT NULL,
  `youtube_url` varchar(256) NOT NULL,
  `is_play` tinyint(1) NOT NULL,
  `is_clear` tinyint(1) NOT NULL,
  `is_favourite` tinyint(1) NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_games_code` (`code`),
  KEY `name` (`name`) USING BTREE,
  KEY `fk_game_manufacturers` (`manufacturer_id`),
  KEY `fk_game_machines` (`machine_id`),
  KEY `fk_game_genres` (`genre_id`),
  KEY `idx_games_name` (`name`),
  CONSTRAINT `fk_game_genres` FOREIGN KEY (`genre_id`) REFERENCES `genres` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_game_machines` FOREIGN KEY (`machine_id`) REFERENCES `machines` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_game_manufacturers` FOREIGN KEY (`manufacturer_id`) REFERENCES `manufacturers` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ゲーム';
CREATE TABLE `genres` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(256) NOT NULL,
  `kana` varchar(256) NOT NULL,
  `overview` varchar(2000) NOT NULL,
  `code` varchar(30) NOT NULL,
  `image_key` varchar(255) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_genres_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='ジャンル情報';
CREATE TABLE `keywords` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(256) NOT NULL,
  `kana` varchar(256) NOT NULL,
  `overview` varchar(2000) NOT NULL,
  `code` varchar(30) NOT NULL,
  `keyword_type` enum('SERIES','SYSTEM','MACHINE','OTHER') NOT NULL,
  `sort_order` int NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_keywords_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='キーワード情報';
CREATE TABLE `machines` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(256) NOT NULL,
  `kana` varchar(256) NOT NULL,
  `overview` varchar(2000) NOT NULL,
  `code` varchar(30) NOT NULL,
  `image_key` varchar(255) DEFAULT NULL,
  `abbreviation` varchar(10) NOT NULL,
  `manufacturer_id` bigint NOT NULL,
  `machine_type` enum('STATIONARY','PORTABLE','BOTH') NOT NULL,
  `release_date` datetime(3) NOT NULL,
  `sort_order` int NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_machines_code` (`code`),
  KEY `fk_machines_manufacturer` (`manufacturer_id`),
  CONSTRAINT `fk_machine_manufacturers` FOREIGN KEY (`manufacturer_id`) REFERENCES `manufacturers` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_machines_manufacturer` FOREIGN KEY (`manufacturer_id`) REFERENCES `manufacturers` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='機種';
CREATE TABLE `manufacturers` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(256) NOT NULL,
  `kana` varchar(256) NOT NULL,
  `overview` varchar(2000) NOT NULL,
  `code` varchar(30) NOT NULL,
  `image_key` varchar(255) DEFAULT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_manufacturers_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='メーカー情報';
CREATE TABLE `purchase_candidates` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` longtext NOT NULL,
  `kana` longtext NOT NULL,
  `image_key` longtext,
  `code` varchar(191) NOT NULL,
  `list_price` int DEFAULT NULL,
  `official_site_url` varchar(191) NOT NULL DEFAULT '',
  `youtube_url` varchar(191) NOT NULL DEFAULT '',
  `release_date_text` varchar(191) NOT NULL DEFAULT '',
  `manufacturer_id` bigint NOT NULL,
  `machine_id` bigint NOT NULL,
  `genre_id` bigint DEFAULT NULL,
  `is_purchased` tinyint(1) NOT NULL DEFAULT '0',
  `display_order` bigint NOT NULL DEFAULT '0',
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_purchase_candidates_code` (`code`),
  KEY `idx_purchase_candidates_is_purchased` (`is_purchased`),
  KEY `idx_purchase_candidates_display_order` (`display_order`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `search_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `visitor_id` varchar(128) NOT NULL DEFAULT '',
  `ip_hash` char(64) NOT NULL DEFAULT '',
  `machine_code` varchar(64) NOT NULL DEFAULT '',
  `manufacturer_code` varchar(64) NOT NULL DEFAULT '',
  `genre_code` varchar(64) NOT NULL DEFAULT '',
  `keyword_code` varchar(64) NOT NULL DEFAULT '',
  `search_word` varchar(255) NOT NULL DEFAULT '',
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_search_logs_visitor_id` (`visitor_id`),
  KEY `idx_search_logs_ip_hash` (`ip_hash`),
  KEY `idx_search_logs_machine_code` (`machine_code`),
  KEY `idx_search_logs_manufacturer_code` (`manufacturer_code`),
  KEY `idx_search_logs_genre_code` (`genre_code`),
  KEY `idx_search_logs_keyword_code` (`keyword_code`),
  KEY `idx_search_logs_search_word` (`search_word`),
  KEY `idx_search_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
CREATE TABLE `users` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `email` varchar(256) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `password` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `created_at` datetime(3) NOT NULL,
  `updated_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE KEY `uni_users_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
SET FOREIGN_KEY_CHECKS = 1;

-- +goose Down
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS purchase_candidates;
DROP TABLE IF EXISTS content_access_logs;
DROP TABLE IF EXISTS game_view_logs;
DROP TABLE IF EXISTS search_logs;
DROP TABLE IF EXISTS game_favorites;
DROP TABLE IF EXISTS game_ranking_previous_entries;
DROP TABLE IF EXISTS game_ranking_draft_entries;
DROP TABLE IF EXISTS game_ranking_active_entries;
DROP TABLE IF EXISTS feature_games;
DROP TABLE IF EXISTS features;
DROP TABLE IF EXISTS banners;
DROP TABLE IF EXISTS game_recommendations;
DROP TABLE IF EXISTS contact_inquiries;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS game_affiliates;
DROP TABLE IF EXISTS game_keywords;
DROP TABLE IF EXISTS games;
DROP TABLE IF EXISTS machines;
DROP TABLE IF EXISTS keywords;
DROP TABLE IF EXISTS genres;
DROP TABLE IF EXISTS manufacturers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS admins;
SET FOREIGN_KEY_CHECKS = 1;
