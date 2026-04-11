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

// FailoverGroupEvent 记录智能路由（虚拟故障转移分组）每次 active_id 切换的事件。
//
// 这是只追加的表，不支持更新和删除（除了后台保留期清理）。每条事件都包含切换前后
// 的成员 id、触发原因、可选的触发人（管理员 user id；系统触发为 NULL）。
type FailoverGroupEvent struct {
	ent.Schema
}

func (FailoverGroupEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "failover_group_events"},
	}
}

func (FailoverGroupEvent) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("virtual_group_id").
			Comment("虚拟分组 ID"),
		field.Int64("from_member_id").
			Optional().
			Nillable().
			Comment("切换前的 active 成员 ID"),
		field.Int64("to_member_id").
			Optional().
			Nillable().
			Comment("切换后的 active 成员 ID"),
		field.String("reason").
			MaxLen(40).
			Comment("切换原因：soft_demote/probe_recovered/health_reconcile/manual_pin/manual_unpin/manual_unpin_expired/bootstrap/no_members_available/config_change"),
		field.Int64("triggered_by").
			Optional().
			Nillable().
			Comment("触发人：系统触发为 NULL，否则为管理员 user id"),
		field.String("note").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("可选备注"),
		field.Time("occurred_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (FailoverGroupEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("virtual_group_id", "occurred_at"),
		index.Fields("occurred_at"),
	}
}
