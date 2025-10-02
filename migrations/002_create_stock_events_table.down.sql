-- Drop stock_events table
DROP INDEX IF EXISTS idx_stock_events_article_date;
DROP INDEX IF EXISTS idx_stock_events_event_type;
DROP INDEX IF EXISTS idx_stock_events_created_at;
DROP INDEX IF EXISTS idx_stock_events_order_id;
DROP INDEX IF EXISTS idx_stock_events_article_id;
DROP TABLE IF EXISTS stock_events;