-- +goose Up
CREATE TABLE IF NOT EXISTS `page_view_logs` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `visitor_hash` char(64) NOT NULL,
  `page_path` varchar(191) NOT NULL,
  `viewed_on` date NOT NULL,
  `created_at` datetime(3) NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_page_view_daily` (`visitor_hash`,`page_path`,`viewed_on`),
  KEY `idx_page_view_logs_page_path` (`page_path`),
  KEY `idx_page_view_logs_viewed_on` (`viewed_on`),
  KEY `idx_page_view_logs_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- +goose Down
SELECT 1;