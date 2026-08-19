CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    plan TEXT NOT NULL DEFAULT 'free',
    status TEXT NOT NULL DEFAULT 'active',
    settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,  -- Now NOT NULL for all users
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    last_login TIMESTAMPTZ,
    login_attempts INT NOT NULL DEFAULT 0,
    is_locked BOOLEAN NOT NULL DEFAULT FALSE,
    locked_until TIMESTAMPTZ,
    password_reset_token TEXT, 
    token_expiry TIMESTAMPTZ,
    must_change_password BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    mfa_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    last_password_change TIMESTAMPTZ,
    user_type TEXT NOT NULL CHECK (
        user_type IN (
            'tenant',     -- the main account holder
            'staff',      -- internal staff
            'guest'       -- external temporary access for contractors, clients etc.
        )
    ),
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Foreign key to tenants table
    CONSTRAINT fk_users_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(id) ON DELETE RESTRICT
    -- Removed the check_tenant_id constraint
);

-- Create indexes for better performance
CREATE INDEX idx_users_tenant_id ON users(tenant_id);
CREATE INDEX idx_users_tenant_type ON users(tenant_id, user_type);
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- User profiles table
CREATE TABLE user_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL, 
    full_name TEXT,
    phone TEXT,
    avatar_url TEXT,
    timezone TEXT,
    permissions JSONB NOT NULL DEFAULT '{
            "system": { "index": true, "restart": true, "backup": true, "shutdown": true },
            "dashboard" : {"index" : true},
            "users": { "index": true, "view": true, "add": true, "edit": true, "delete": true },
            "employees": { "index": true, "view": true, "add": true, "archive": true },
            "clients": { "index": true, "view": true, "add": true, "edit": true, "delete": true }
    }'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user_profiles_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_profiles_tenant ON user_profiles(tenant_id);

-- User tokens table
CREATE TABLE user_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    token TEXT NOT NULL,
    type TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    used_at TIMESTAMPTZ,
    
    CONSTRAINT fk_user_tokens_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_user_tokens_tenant ON user_tokens(tenant_id);
CREATE INDEX idx_user_tokens_token ON user_tokens(token) WHERE used_at IS NULL;

-- Refresh tokens table
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_refresh_tokens_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(id) ON DELETE CASCADE
);

CREATE INDEX idx_refresh_tokens_tenant ON refresh_tokens(tenant_id);

-- Roles table - tenant-specific
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    permission_template JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_roles_tenant FOREIGN KEY (tenant_id) 
        REFERENCES tenants(id) ON DELETE CASCADE,
    CONSTRAINT unique_role_per_tenant UNIQUE (tenant_id, name)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id);
