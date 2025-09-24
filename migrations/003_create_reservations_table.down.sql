-- Drop stock_reservations table
DROP INDEX IF EXISTS idx_reservations_active;
DROP INDEX IF EXISTS idx_reservations_expires_at;
DROP INDEX IF EXISTS idx_reservations_status;
DROP INDEX IF EXISTS idx_reservations_order_id;
DROP INDEX IF EXISTS idx_reservations_article_id;
DROP TABLE IF EXISTS stock_reservations;