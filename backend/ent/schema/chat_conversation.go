package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ChatConversation holds the schema definition for the ChatConversation entity.
type ChatConversation struct {
	ent.Schema
}

func (ChatConversation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "chat_conversations"},
	}
}

func (ChatConversation) Fields() []ent.Field {
	return []ent.Field{
		field.String("guest_token").
			MaxLen(64).
			Optional().
			Nillable().
			Comment("匿名访客令牌"),
		field.Int64("user_id").
			Optional().
			Nillable().
			Comment("登录用户ID"),
		field.String("visitor_name").
			MaxLen(100).
			Default("").
			Comment("访客显示名"),
		field.String("status").
			MaxLen(20).
			Default("open").
			Comment("会话状态: open, closed"),
		field.Int("admin_unread_count").
			Default(0).
			Comment("管理员未读消息数"),
		field.Time("last_message_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("最后消息时间"),
		field.String("last_message_preview").
			MaxLen(200).
			Default("").
			Comment("最后消息预览"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ChatConversation) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("messages", ChatMessage.Type),
	}
}

func (ChatConversation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("guest_token"),
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("last_message_at"),
	}
}
