package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Channel holds the schema definition for the Channel entity.
// 渠道实体，用于配置多个渠道并查询余额。
type Channel struct {
	ent.Schema
}

func (Channel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "channels"},
	}
}

func (Channel) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Channel) Fields() []ent.Field {
	return []ent.Field{
		// 基本信息
		field.String("name").
			MaxLen(100).
			NotEmpty().
			Comment("渠道名称"),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("描述"),
		field.String("platform").
			MaxLen(50).
			Optional().
			Nillable().
			Comment("所属平台（anthropic/openai/gemini/other）"),
		field.String("status").
			MaxLen(20).
			Default("active").
			Comment("状态：active/inactive/error"),
		field.String("icon_url").
			MaxLen(500).
			Optional().
			Nillable().
			Comment("渠道图标URL"),
		field.String("website_url").
			MaxLen(500).
			Optional().
			Nillable().
			Comment("渠道网站链接"),

		// 余额查询配置
		field.String("balance_url").
			MaxLen(500).
			Optional().
			Nillable().
			Comment("余额查询API URL"),
		field.String("balance_method").
			MaxLen(10).
			Default("GET").
			Comment("请求方式：GET/POST"),
		field.JSON("balance_headers", map[string]string{}).
			Optional().
			Comment("请求头（包含认证信息）"),
		field.String("balance_body").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("请求体（POST时使用）"),
		field.String("balance_path").
			MaxLen(200).
			Optional().
			Nillable().
			Comment("响应中余额字段路径（如 data.balance）"),
		field.String("balance_unit").
			MaxLen(20).
			Default("$").
			Comment("余额单位"),

		// 缓存的余额信息
		field.Float("cached_balance").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,6)"}).
			Comment("缓存的余额值"),
		field.Time("last_check_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("上次查询时间"),
		field.String("last_error").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("上次查询错误"),
	}
}

func (Channel) Edges() []ent.Edge {
	return nil
}

func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("platform"),
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}
