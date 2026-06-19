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

// ResellerAPIToken holds the schema definition for the ResellerAPIToken entity.
// 这是分销商用于"机器对机器(M2M)"调用 key 管理接口的服务令牌。
// 明文只在创建时返回一次，库里仅保存 SHA-256 哈希。
type ResellerAPIToken struct {
	ent.Schema
}

func (ResellerAPIToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reseller_api_tokens"},
	}
}

func (ResellerAPIToken) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (ResellerAPIToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("reseller_id").
			Comment("分销商用户 ID"),
		field.String("name").
			MaxLen(100).
			Default("").
			Comment("令牌用途备注"),
		field.String("token_prefix").
			MaxLen(20).
			Comment("明文前缀，用于 UI 展示，如 rst-1a2b"),
		field.String("token_hash").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("SHA-256(token) 十六进制；明文不落库"),
		field.String("status").
			MaxLen(20).
			Default("active").
			Comment("active / revoked"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("最近一次使用时间"),
		field.Time("expires_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("过期时间，NULL 表示永不过期"),
	}
}

func (ResellerAPIToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("token_hash").Unique(),
		index.Fields("reseller_id"),
	}
}
