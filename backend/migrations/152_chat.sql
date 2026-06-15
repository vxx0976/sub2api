-- 客服聊天系统: 会话 + 消息表

CREATE TABLE IF NOT EXISTS chat_conversations (
    id                   BIGSERIAL PRIMARY KEY,
    guest_token          VARCHAR(64),
    user_id              BIGINT,
    visitor_name         VARCHAR(100)  NOT NULL DEFAULT '',
    status               VARCHAR(20)   NOT NULL DEFAULT 'open',
    admin_unread_count   INT           NOT NULL DEFAULT 0,
    last_message_at      TIMESTAMPTZ,
    last_message_preview VARCHAR(200)  NOT NULL DEFAULT '',
    created_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_conversations_guest_token_open
    ON chat_conversations (guest_token) WHERE guest_token IS NOT NULL AND status = 'open';
CREATE INDEX IF NOT EXISTS idx_chat_conversations_user_id
    ON chat_conversations (user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_chat_conversations_status
    ON chat_conversations (status);
CREATE INDEX IF NOT EXISTS idx_chat_conversations_last_message_at
    ON chat_conversations (last_message_at DESC NULLS LAST);

CREATE TABLE IF NOT EXISTS chat_messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT       NOT NULL REFERENCES chat_conversations(id) ON DELETE CASCADE,
    sender_type     VARCHAR(20)  NOT NULL,
    content         TEXT         NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation_id
    ON chat_messages (conversation_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_conv_created
    ON chat_messages (conversation_id, created_at);
