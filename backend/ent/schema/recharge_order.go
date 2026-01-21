package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type RechargeOrder struct {
	ent.Schema
}

func (RechargeOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			Unique().
			MaxLen(50).
			Comment("充值订单号"),
		field.String("trade_no").
			Optional().
			Nillable().
			MaxLen(100).
			Comment("支付平台订单号"),
		field.Int64("user_id").
			Comment("用户ID"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("用户实际支付金额"),
		field.Float("credit_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("实际到账余额"),
		field.Float("multiplier").
			Default(1.0).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("倍率快照"),
		field.String("status").
			Default("pending").
			MaxLen(20).
			Comment("订单状态: pending|paid|expired|refunded"),
		field.String("pay_type").
			Optional().
			Nillable().
			MaxLen(20).
			Comment("支付方式: alipay|wxpay"),
		field.Time("paid_at").
			Optional().
			Nillable().
			Comment("支付时间"),
		field.Time("created_at").
			Immutable().
			Default(time.Now).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			Comment("更新时间"),
		field.Time("expired_at").
			Optional().
			Nillable().
			Comment("过期时间"),
	}
}

func (RechargeOrder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("recharge_orders").
			Field("user_id").
			Unique().
			Required(),
	}
}

func (RechargeOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("order_no").Unique(),
	}
}
