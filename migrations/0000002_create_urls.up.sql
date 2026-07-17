CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE urls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    user_id UUID NOT NULL,
    slug VARCHAR(64) NOT NULL,
    original_url TEXT NOT NULL,
    expires_at TIMESTAMPTZ,
    click_cap BIGINT CHECK (click_cap > 0),
    click_count BIGINT DEFAULT 0,
    password_hash TEXT,
    state VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_tenant_slug ON urls (tenant_id, slug);

CREATE INDEX idx_tenant_slug_active ON urls (tenant_id, slug) WHERE state = 'active';