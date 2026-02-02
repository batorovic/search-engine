-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Contents table
CREATE TABLE IF NOT EXISTS contents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    title VARCHAR(500) NOT NULL,
    type VARCHAR(20) NOT NULL CHECK (type IN ('video', 'text')),
    published_at TIMESTAMP NOT NULL,
    raw_data JSONB NOT NULL,

    -- Denormalized fields for search performance
    views INTEGER DEFAULT 0,
    likes INTEGER DEFAULT 0,
    reactions INTEGER DEFAULT 0,
    reading_time INTEGER DEFAULT 0,

    -- Computed score
    score DECIMAL(10,2) DEFAULT 0,

    tags TEXT[] DEFAULT '{}',

    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),

    UNIQUE(provider, external_id)
);

-- Full-text search index
CREATE INDEX IF NOT EXISTS idx_contents_search ON contents
    USING GIN (to_tsvector('english', title));

-- Tags search index
CREATE INDEX IF NOT EXISTS idx_contents_tags ON contents USING GIN (tags);

-- Type filter index
CREATE INDEX IF NOT EXISTS idx_contents_type ON contents(type);

-- Score sorting index
CREATE INDEX IF NOT EXISTS idx_contents_score ON contents(score DESC);

-- Provider + published_at index for time-based queries
CREATE INDEX IF NOT EXISTS idx_contents_provider_published ON contents(provider, published_at DESC);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for contents table
DROP TRIGGER IF EXISTS update_contents_updated_at ON contents;
CREATE TRIGGER update_contents_updated_at
    BEFORE UPDATE ON contents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
