-- Migration up: create authorized_devices and device_bootstrap_tokens tables

CREATE TABLE IF NOT EXISTS authorized_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hardware_uuid_hash VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used_at TIMESTAMPTZ,
    registered_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_authorized_devices_hw ON authorized_devices (hardware_uuid_hash);
CREATE INDEX IF NOT EXISTS idx_authorized_devices_teacher ON authorized_devices (teacher_id);

CREATE TABLE IF NOT EXISTS device_bootstrap_tokens (
    token VARCHAR(64) PRIMARY KEY,
    device_id UUID NOT NULL REFERENCES authorized_devices(id) ON DELETE CASCADE,
    teacher_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_device_bootstrap_lookup ON device_bootstrap_tokens (token, expires_at) WHERE used_at IS NULL;
