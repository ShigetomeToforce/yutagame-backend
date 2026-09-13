-- +goose Up
ALTER TABLE `games`
  ADD COLUMN `surugaya_id` varchar(64) NULL AFTER `list_price`,
  ADD COLUMN `reference_used_price` int NULL AFTER `surugaya_id`,
  ADD COLUMN `reference_buy_price` int NULL AFTER `reference_used_price`;

-- +goose Down
ALTER TABLE `games`
  DROP COLUMN `reference_buy_price`,
  DROP COLUMN `reference_used_price`,
  DROP COLUMN `surugaya_id`;