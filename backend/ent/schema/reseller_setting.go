package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ResellerSetting holds the schema definition for the ResellerSetting entity.
type ResellerSetting struct {
	ent.Schema
}

func (ResellerSetting) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reseller_settings"},
	}
}

func (ResellerSetting) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("reseller_id").
			Comment("分销商用户 ID"),
		field.String("key").
			MaxLen(100).
			NotEmpty().
			Comment("设置键名"),
		field.String("value").
			Default("").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("设置值"),
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

func (ResellerSetting) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("reseller_id", "key").Unique(),
	}
}
