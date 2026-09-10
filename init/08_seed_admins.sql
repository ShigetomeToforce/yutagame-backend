SET NAMES utf8mb4;

DELETE FROM `admins`;
-- ==========================================
-- 初期Adminユーザーのシードデータ
-- ==========================================

INSERT INTO `admins` (
    `id`, 
    `email`, 
    `password`, 
    `name`, 
    `role_type`, 
    `created_at`, 
    `updated_at`
) VALUES (
    1, 
    'admin@example.com', 
    '$2a$10$53xmr3o2m1Tuxo0IQFfko.Suq4Z426P16q4eStnZ9u4acD0WlyeAq', 
    'デフォルト管理者アカウント', 
    'ADMIN', 
    NOW(), 
    NOW()
) ON DUPLICATE KEY UPDATE 
    `password` = VALUES(`password`),
    `name` = VALUES(`name`),
    `role_type` = VALUES(`role_type`),
    `updated_at` = NOW();