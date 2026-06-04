CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY,
    sender TEXT NOT NULL,
    recipient TEXT NOT NULL,
    message_type TEXT NOT NULL,
    content TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_thread_created
    ON messages (LEAST(sender, recipient), GREATEST(sender, recipient), created_at DESC);
