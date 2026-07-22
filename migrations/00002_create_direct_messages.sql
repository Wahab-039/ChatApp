-- +goose Up
CREATE TABLE direct_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    recipient_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    client_message_id VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT direct_messages_body_not_blank CHECK (char_length(btrim(body)) > 0),
    CONSTRAINT direct_messages_not_self CHECK (sender_id <> recipient_id),
    CONSTRAINT direct_messages_sender_client_message_id_unique UNIQUE (sender_id, client_message_id)
);

CREATE INDEX direct_messages_conversation_created_at_idx
    ON direct_messages (sender_id, recipient_id, created_at);

CREATE INDEX direct_messages_recipient_created_at_idx
    ON direct_messages (recipient_id, created_at);

-- +goose Down
DROP TABLE direct_messages;
