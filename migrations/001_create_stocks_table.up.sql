-- Create stocks table
CREATE TABLE IF NOT EXISTS stocks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    article_id VARCHAR(100) NOT NULL UNIQUE,
    quantity INTEGER NOT NULL DEFAULT 0,
    reserved INTEGER NOT NULL DEFAULT 0,
    min_stock INTEGER NOT NULL DEFAULT 0,
    max_stock INTEGER NOT NULL DEFAULT 0,
    location VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    CONSTRAINT chk_quantity_positive CHECK (quantity >= 0),
    CONSTRAINT chk_reserved_positive CHECK (reserved >= 0),
    CONSTRAINT chk_min_stock_positive CHECK (min_stock >= 0),
    CONSTRAINT chk_max_stock_positive CHECK (max_stock >= 0),
    CONSTRAINT chk_reserved_not_greater_than_quantity CHECK (reserved <= quantity),
    CONSTRAINT chk_max_stock_greater_than_min CHECK (max_stock = 0 OR max_stock >= min_stock)
);

-- Create index on article_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_stocks_article_id ON stocks(article_id);

-- Create index for low stock queries
CREATE INDEX IF NOT EXISTS idx_stocks_low_stock ON stocks(quantity, min_stock) WHERE quantity <= min_stock;