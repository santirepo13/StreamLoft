-- StreamLoft Database Schema
-- PostgreSQL
-- Per SRS.MD requirements

-- Users table (per SRS 14.2 Data Dictionary - User Table)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    numeric_id VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    stream_key VARCHAR(32) NOT NULL,  -- SRS: Required, generated on first login
    bitrate INTEGER,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS destinations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    rtmp_url VARCHAR(500) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS user_destinations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    destination_id INTEGER NOT NULL REFERENCES destinations(id) ON DELETE CASCADE,
    stream_key TEXT,  -- SRS BR-004: Stored encrypted
    enabled INTEGER NOT NULL DEFAULT 1 CHECK (enabled IN (0, 1)),  -- 1=true, 0=false — runtime toggle
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, destination_id)
);

CREATE TABLE IF NOT EXISTS broadcast_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_destination_id INTEGER NOT NULL REFERENCES user_destinations(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    duration_minutes INTEGER NOT NULL DEFAULT 0 CHECK (duration_minutes >= 0),
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS user_machines (
    id SERIAL PRIMARY KEY,
    machine_id VARCHAR(255) NOT NULL UNIQUE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- UserSessions table (per SRS 14.2 Data Dictionary - supports multiple devices per user)
-- Note: access_token and refresh_token stored as RAW (not encrypted) - verified via HMAC signature
-- Only stream_key in user_destinations is encrypted (per SRS DR-001)
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    machine_id VARCHAR(255) NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    token_expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, machine_id)
);

CREATE INDEX IF NOT EXISTS idx_users_numeric_id ON users(numeric_id);
CREATE INDEX IF NOT EXISTS idx_users_stream_key ON users(stream_key);
CREATE INDEX IF NOT EXISTS idx_user_destinations_user_id ON user_destinations(user_id);
CREATE INDEX IF NOT EXISTS idx_user_destinations_destination_id ON user_destinations(destination_id);
CREATE INDEX IF NOT EXISTS idx_broadcast_sessions_user_id ON broadcast_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_broadcast_sessions_user_destination_id ON broadcast_sessions(user_destination_id);
CREATE INDEX IF NOT EXISTS idx_user_machines_machine_id ON user_machines(machine_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_access_token ON user_sessions(access_token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_token ON user_sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_destinations_name ON destinations(name);