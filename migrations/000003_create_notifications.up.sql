-- Create notifications schema
CREATE SCHEMA IF NOT EXISTS notifications;

-- Create notifications table
CREATE TABLE notifications.notifications (
    id              UUID PRIMARY KEY,
    tenant_id       VARCHAR(255),
    channel         VARCHAR(50) NOT NULL,
    recipient       VARCHAR(255) NOT NULL,
    subject         VARCHAR(500) NOT NULL,
    body            TEXT NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending',
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    sent_at         TIMESTAMPTZ,
    
    -- Context
    user_id         VARCHAR(255),
    correlation_id  VARCHAR(255),
    event_type      VARCHAR(255),
    
    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common queries
CREATE INDEX idx_notifications_tenant_id ON notifications.notifications(tenant_id) WHERE tenant_id IS NOT NULL;
CREATE INDEX idx_notifications_user_id ON notifications.notifications(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_notifications_channel ON notifications.notifications(channel);
CREATE INDEX idx_notifications_status ON notifications.notifications(status);
CREATE INDEX idx_notifications_recipient ON notifications.notifications(recipient);
CREATE INDEX idx_notifications_created_at ON notifications.notifications(created_at DESC);

-- Composite index for pending retry queries
CREATE INDEX idx_notifications_pending ON notifications.notifications(status, created_at ASC) WHERE status = 'pending';

-- Comment
COMMENT ON TABLE notifications.notifications IS 'Outbound notifications (email, SMS, push)';