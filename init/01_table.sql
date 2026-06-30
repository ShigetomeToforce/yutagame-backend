SET NAMES utf8mb4;

-- ==========================================
-- 1. 独立したマスタテーブル（親）から作成
-- ==========================================

CREATE TABLE `manufacturers` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `name` VARCHAR(256) NOT NULL COMMENT 'メーカー名称' COLLATE 'utf8mb4_0900_ai_ci',
    `kana` VARCHAR(256) NOT NULL COMMENT 'メーカー名称（かな）' COLLATE 'utf8mb4_0900_ai_ci',
    `overview` VARCHAR(2000) NOT NULL COMMENT '概要' COLLATE 'utf8mb4_0900_ai_ci',
    `code` VARCHAR(30) NOT NULL COMMENT 'コード' COLLATE 'utf8mb4_0900_ai_ci',
    `created_at` DATETIME NOT NULL COMMENT '登録日',
    `updated_at` DATETIME NOT NULL COMMENT '更新日',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `code` (`code`) USING BTREE
)
COMMENT='メーカー情報'
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
AUTO_INCREMENT=81
;

CREATE TABLE `genres` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `name` VARCHAR(256) NOT NULL COMMENT 'ジャンル名称' COLLATE 'utf8mb4_0900_ai_ci',
    `kana` VARCHAR(256) NOT NULL COMMENT 'ジャンル名称（かな）' COLLATE 'utf8mb4_0900_ai_ci',
    `overview` VARCHAR(2000) NOT NULL COMMENT '概要' COLLATE 'utf8mb4_0900_ai_ci',
    `code` VARCHAR(30) NOT NULL COMMENT 'コード' COLLATE 'utf8mb4_0900_ai_ci',
    `created_at` DATETIME NOT NULL COMMENT '登録日',
    `updated_at` DATETIME NOT NULL COMMENT '更新日',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `code` (`code`) USING BTREE
)
COMMENT='ジャンル情報'
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
AUTO_INCREMENT=14
;

CREATE TABLE `keywords` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `name` VARCHAR(256) NOT NULL COMMENT 'キーワード名称' COLLATE 'utf8mb4_0900_ai_ci',
    `kana` VARCHAR(256) NOT NULL COMMENT 'キーワード名称（かな）' COLLATE 'utf8mb4_0900_ai_ci',
    `overview` VARCHAR(2000) NOT NULL COMMENT '概要' COLLATE 'utf8mb4_0900_ai_ci',
    `code` VARCHAR(30) NOT NULL COMMENT 'コード' COLLATE 'utf8mb4_0900_ai_ci',
    `keyword_type` ENUM('SERIES','SYSTEM','MACHINE','OTHER') NOT NULL COMMENT 'タイプ' COLLATE 'utf8mb4_0900_ai_ci',
    `sort_order` INT(10) NOT NULL COMMENT '並び順',
    `created_at` DATETIME NOT NULL COMMENT '登録日',
    `updated_at` DATETIME NOT NULL COMMENT '更新日',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `code` (`code`) USING BTREE
)
COMMENT='キーワード情報'
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
AUTO_INCREMENT=266
;

-- ==========================================
-- 2. 依存関係のあるテーブル（子）の作成
-- ==========================================

CREATE TABLE `machines` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `name` VARCHAR(256) NOT NULL COMMENT '機種名称' COLLATE 'utf8mb4_0900_ai_ci',
    `kana` VARCHAR(256) NOT NULL COMMENT '機種名称（かな）' COLLATE 'utf8mb4_0900_ai_ci',
    `overview` VARCHAR(2000) NOT NULL COMMENT '概要' COLLATE 'utf8mb4_0900_ai_ci',
    `code` VARCHAR(30) NOT NULL COMMENT 'コード' COLLATE 'utf8mb4_0900_ai_ci',
    `abbreviation` VARCHAR(10) NOT NULL COMMENT '略称' COLLATE 'utf8mb4_0900_ai_ci',
    `manufacturer_id` BIGINT(19) NOT NULL COMMENT 'メーカーID',
    `machine_type` ENUM('STATIONARY','PORTABLE','BOTH') NOT NULL COMMENT '種別' COLLATE 'utf8mb4_0900_ai_ci',
    `release_date` DATE NOT NULL COMMENT '発売日',
    `sort_order` INT(10) NOT NULL COMMENT '並び順',
    `created_at` DATETIME NOT NULL COMMENT '登録日',
    `updated_at` DATETIME NOT NULL COMMENT '更新日',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `code` (`code`) USING BTREE,
    -- 💡 外部キー制約の追加：manufacturersテーブルとの紐付け
    CONSTRAINT `fk_machine_manufacturers` FOREIGN KEY (`manufacturer_id`) REFERENCES `manufacturers` (`id`) ON DELETE RESTRICT
)
COMMENT='機種'
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
AUTO_INCREMENT=36
;

CREATE TABLE `games` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `name` VARCHAR(256) NOT NULL COMMENT 'ゲーム名称' COLLATE 'utf8mb4_0900_ai_ci',
    `kana` VARCHAR(256) NOT NULL COMMENT 'ゲーム名称（かな）' COLLATE 'utf8mb4_0900_ai_ci',
    `overview` VARCHAR(2000) NOT NULL COMMENT '概要' COLLATE 'utf8mb4_0900_ai_ci',
    `code` VARCHAR(30) NOT NULL COMMENT 'コード' COLLATE 'utf8mb4_0900_ai_ci',
    `manufacturer_id` BIGINT(19) NOT NULL COMMENT 'メーカーID',
    `machine_id` BIGINT(19) NOT NULL COMMENT '機種ID',
    `genre_id` BIGINT(19) NOT NULL COMMENT 'ジャンルID',
    `sub_genre` VARCHAR(256) NOT NULL COMMENT 'サブジャンル' COLLATE 'utf8mb4_0900_ai_ci',
    `catch_copy` VARCHAR(256) NOT NULL COMMENT 'キャッチコピー' COLLATE 'utf8mb4_0900_ai_ci',
    `sub_catch` VARCHAR(512) NOT NULL COMMENT 'サブキャッチ' COLLATE 'utf8mb4_0900_ai_ci',
    `list_price` INT(10) NOT NULL COMMENT '定価',
    `release_date` DATE NOT NULL COMMENT '発売日',
    `official_site_url` VARCHAR(256) NOT NULL COMMENT '公式サイトURL' COLLATE 'utf8mb4_0900_ai_ci',
    `youtube_url` VARCHAR(256) NOT NULL COMMENT 'YoutubeURL' COLLATE 'utf8mb4_0900_ai_ci',
    `is_play` TINYINT(1) NOT NULL COMMENT 'プレイ済みフラグ',
    `is_clear` TINYINT(1) NOT NULL COMMENT 'クリア済みフラグ',
    `is_favourite` TINYINT(1) NOT NULL COMMENT 'お気に入りフラグ',
    `created_at` DATETIME NOT NULL COMMENT '登録日',
    `updated_at` DATETIME NOT NULL COMMENT '更新日',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `code` (`code`) USING BTREE,
    INDEX `name` (`name`) USING BTREE,
    -- 💡 外部キー制約の追加：各種マスタテーブルとの紐付け
    CONSTRAINT `fk_game_manufacturers` FOREIGN KEY (`manufacturer_id`) REFERENCES `manufacturers` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_game_machines` FOREIGN KEY (`machine_id`) REFERENCES `machines` (`id`) ON DELETE RESTRICT,
    CONSTRAINT `fk_game_genres` FOREIGN KEY (`genre_id`) REFERENCES `genres` (`id`) ON DELETE RESTRICT
)
COMMENT='ゲーム'
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
AUTO_INCREMENT=1011
;

-- ==========================================
-- 3. 認証・ユーザー関連（独立）
-- ==========================================

CREATE TABLE `admins` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT,
    `email` VARCHAR(256) NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `password` VARCHAR(128) NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `name` VARCHAR(32) NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `role_type` ENUM('ADMIN','USER') NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `email` (`email`) USING BTREE
)
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
AUTO_INCREMENT=4
;

CREATE TABLE `users` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT,
    `email` VARCHAR(256) NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `password` VARCHAR(128) NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `name` VARCHAR(32) NOT NULL COLLATE 'utf8mb4_0900_ai_ci',
    `created_at` DATETIME NOT NULL,
    `updated_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `email` (`email`) USING BTREE
)
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB
;

-- ==========================================
-- 4. 中間テーブル
-- ==========================================
CREATE TABLE `game_keywords` (
  `game_id` bigint NOT NULL,
  `keyword_id` bigint NOT NULL,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`game_id`, `keyword_id`),
  INDEX `idx_game_keywords_keyword_id` (`keyword_id`),
  CONSTRAINT `fk_game_keywords_games` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON DELETE CASCADE, -- ゲームが削除されたら、中間テーブルの紐付けも自動消滅    
  CONSTRAINT `fk_game_keywords_keywords` FOREIGN KEY (`keyword_id`) REFERENCES `keywords` (`id`) ON DELETE CASCADE  -- キーワードが削除されたら、中間テーブルの紐付けも自動消滅
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
