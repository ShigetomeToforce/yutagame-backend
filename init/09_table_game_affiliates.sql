SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS `game_affiliates` (
    `id` BIGINT(19) NOT NULL AUTO_INCREMENT COMMENT 'ID',
    `game_id` BIGINT(19) NOT NULL COMMENT 'ゲームID',
    `category` ENUM('AMAZON','RAKUTEN','YAHOO','SURUGAYA','PLAYSTATION_STORE','NINTENDO_STORE','STEAM') NOT NULL COMMENT 'カテゴリ' COLLATE 'utf8mb4_0900_ai_ci',
    `url` VARCHAR(1024) NOT NULL COMMENT '購入先URL' COLLATE 'utf8mb4_0900_ai_ci',
    `created_at` DATETIME NOT NULL COMMENT '登録日',
    `updated_at` DATETIME NOT NULL COMMENT '更新日',
    PRIMARY KEY (`id`) USING BTREE,
    UNIQUE INDEX `uniq_game_affiliate_category` (`game_id`, `category`) USING BTREE,
    CONSTRAINT `fk_game_affiliates_games` FOREIGN KEY (`game_id`) REFERENCES `games` (`id`) ON DELETE CASCADE
)
COMMENT='ゲーム購入導線'
COLLATE='utf8mb4_0900_ai_ci'
ENGINE=InnoDB;
