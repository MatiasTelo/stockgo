-- Recreate stock_reservations table (rollback)
CREATE TABLE IF NOT EXISTS stock_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id VARCHAR(100) NOT NULL,
    order_id VARCHAR(100) NOT NULL,
    quantity INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_reservation_quantity_positive CHECK (quantity > 0),
    CONSTRAINT chk_reservation_status CHECK (status IN ('ACTIVE', 'CONFIRMED', 'CANCELLED', 'EXPIRED')),
    CONSTRAINT uk_order_article_active UNIQUE (order_id, article_id, status) DEFERRABLE INITIALLY DEFERRED
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_reservations_article_id ON stock_reservations(article_id);
CREATE INDEX IF NOT EXISTS idx_reservations_order_id ON stock_reservations(order_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON stock_reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires_at ON stock_reservations(expires_at) WHERE status = 'ACTIVE';

-- Composite index for active reservations
CREATE INDEX IF NOT EXISTS idx_reservations_active ON stock_reservations(article_id, order_id) WHERE status = 'ACTIVE';