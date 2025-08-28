
-- Migration: 001_initial_schema.sql
CREATE TABLE IF NOT EXISTS urls (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    created_by VARCHAR(255) NULL
);

CREATE TABLE IF NOT EXISTS url_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) NOT NULL,
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    user_agent TEXT,
    ip_address INET,
    referer TEXT,
    country VARCHAR(10)
);

-- Basic indexes
CREATE INDEX IF NOT EXISTS idx_urls_short_code ON urls(short_code);
CREATE INDEX IF NOT EXISTS idx_urls_original_url ON urls(original_url);
CREATE INDEX IF NOT EXISTS idx_urls_expires_at ON urls(expires_at);
CREATE INDEX IF NOT EXISTS idx_analytics_short_code ON url_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_analytics_clicked_at ON url_analytics(clicked_at);
