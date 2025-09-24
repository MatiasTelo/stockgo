-- Create stock_events table
CREATE TABLE IF NOT EXISTS stock_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    quantity INTEGER NOT NULL,
    order_id VARCHAR(100),
    reason TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_event_type CHECK (event_type IN ('ADD', 'REPLENISH', 'DEDUCT', 'RESERVE', 'CANCEL_RESERVE', 'LOW_STOCK')),
    CONSTRAINT chk_event_quantity_positive CHECK (quantity >= 0)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_stock_events_article_id ON stock_events(article_id);
CREATE INDEX IF NOT EXISTS idx_stock_events_order_id ON stock_events(order_id) WHERE order_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_stock_events_created_at ON stock_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_stock_events_event_type ON stock_events(event_type);

-- Composite index for article events ordered by date
CREATE INDEX IF NOT EXISTS idx_stock_events_article_date ON stock_events(article_id, created_at DESC);