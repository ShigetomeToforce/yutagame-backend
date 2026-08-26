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
    '$2a$10$SwIs8SXB.Ji8h7Y5EY3EnOPIKuDVlOpSab2rt2Rb7zcE6chDntE/6', 
    'デフォルト管理者アカウント', 
    'ADMIN', 
    NOW(), 
    NOW()
) ON DUPLICATE KEY UPDATE 
    `name` = VALUES(`name`),
    `role_type` = VALUES(`role_type`),
    `updated_at` = NOW();