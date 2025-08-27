CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    last_login TIMESTAMPTZ,
    login_attempts INT DEFAULT 0,
    is_locked BOOLEAN DEFAULT FALSE,
    locked_until TIMESTAMPTZ,
    password_reset_token TEXT, 
    token_expiry TIMESTAMPTZ,
    must_change_password BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,
    mfa_enabled BOOLEAN DEFAULT FALSE,
    last_password_change TIMESTAMPTZ,
    user_type TEXT NOT NULL CHECK (
        user_type IN (
            'owner',      -- the main account holder
            'employee',   -- internal staff
            'contractor', -- temporary/internal but not staff
            'vendor',     -- external partner/supplier
            'guest',      -- external temporary access
            'customer'    -- end-client of the business
        )
    ),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    full_name TEXT,
    phone TEXT,
    avatar_url TEXT,
    timezone TEXT,
    permission JSONB DEFAULT '{
        "resources": {
            "system": { "view": true, "restart": true, "backup": true, "shutdown": true },
            "users": { "view": true, "create": true, "edit": true, "delete": true },
            "hr": { "view": true, "create": true, "export": true },
            "finance": { "view": true, "create": true, "edit": true, "delete": true }
        },
        "routes": {
            "/*": true,
        }
    }'::jsonb
);

-- CREATE INDEX idx_users_reset_token ON users(password_reset_token);
-- CREATE INDEX idx_users_locked_status ON users(is_locked, locked_until);

CREATE TABLE user_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL,
    type TEXT NOT NULL,   -- 'reset', 'verify_email', 'mfa'
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ
);

-- Refresh tokens table for JWT or session tokens
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permission_template JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);