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

// ChatMessage holds the schema definition for the ChatMessage entity.
type ChatMessage struct {
	ent.Schema
}

func (ChatMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "chat_messages"},
	}
}

func (ChatMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("conversation_id").
			Comment("所属会话ID"),
		field.String("sender_type").
			MaxLen(20).
			NotEmpty().
			Comment("发送者类型: visitor, admin"),
		field.String("content").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			NotEmpty().
			Comment("消息内容"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ChatMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("conversation", ChatConversation.Type).
			Ref("messages").
			Field("conversation_id").
			Unique().
			Required(),
	}
}

func (ChatMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("conversation_id"),
		index.Fields("conversation_id", "created_at"),
	}
}
