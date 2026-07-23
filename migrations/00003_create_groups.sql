-- +goose Up
CREATE TABLE groups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT groups_name_not_blank CHECK (char_length(btrim(name)) > 0)
);

CREATE INDEX groups_created_by_idx ON groups(created_by);

CREATE TABLE group_members (
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL DEFAULT 'member' CHECK (role IN ('admin', 'member')),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX group_members_user_idx ON group_members(user_id);
CREATE INDEX group_members_group_idx ON group_members(group_id);

CREATE TABLE group_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    sender_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    client_message_id VARCHAR(128) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT group_messages_body_not_blank CHECK (char_length(btrim(body)) > 0),
    CONSTRAINT group_messages_sender_client_message_id_unique UNIQUE (sender_id, client_message_id)
);

CREATE INDEX group_messages_group_created_at_idx ON group_messages(group_id, created_at DESC);
CREATE INDEX group_messages_sender_idx ON group_messages(sender_id);

-- +goose Down
DROP TABLE group_messages;
DROP TABLE group_members;
DROP TABLE groups;
