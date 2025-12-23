-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- USERS TABLE
-- =============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email_verified_at TIMESTAMPTZ,
    deletion_scheduled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_users_email UNIQUE (email)
);

-- Index for users scheduled for deletion (cleanup job)
CREATE INDEX idx_users_deletion_scheduled_at
    ON users(deletion_scheduled_at)
    WHERE deletion_scheduled_at IS NOT NULL;

-- =============================================================================
-- AUTH TOKENS TABLE
-- =============================================================================
CREATE TABLE auth_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_auth_tokens_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_auth_tokens_token_hash UNIQUE (token_hash)
);

-- Index for token lookup
CREATE INDEX idx_auth_tokens_user_id ON auth_tokens(user_id);

-- Composite index for RevokeAllForUser query: WHERE user_id = ? AND revoked_at IS NULL
CREATE INDEX idx_auth_tokens_user_active
    ON auth_tokens(user_id)
    WHERE revoked_at IS NULL;

-- Partial index for cleanup of expired tokens
CREATE INDEX idx_auth_tokens_expired
    ON auth_tokens(expires_at)
    WHERE revoked_at IS NULL;

-- =============================================================================
-- PASSWORD RESETS TABLE
-- =============================================================================
CREATE TABLE password_resets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_password_resets_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_password_resets_token_hash UNIQUE (token_hash)
);

-- Index for user's password reset requests
CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);

-- Partial index for cleanup of expired/unused tokens
CREATE INDEX idx_password_resets_expired
    ON password_resets(expires_at)
    WHERE used_at IS NULL;

-- =============================================================================
-- EMAIL VERIFICATIONS TABLE
-- =============================================================================
CREATE TABLE email_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_email_verifications_user FOREIGN KEY (user_id)
        REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT uq_email_verifications_token_hash UNIQUE (token_hash)
);

-- Index for user's email verification requests
CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);

-- Partial index for cleanup of expired/unused tokens
CREATE INDEX idx_email_verifications_expired
    ON email_verifications(expires_at)
    WHERE used_at IS NULL;
