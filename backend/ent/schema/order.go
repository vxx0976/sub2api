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

// Order holds the schema definition for the Order entity.
type Order struct {
	ent.Schema
}

func (Order) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "orders"},
	}
}

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("商户订单号"),
		field.String("trade_no").
			MaxLen(64).
			Optional().
			Nillable().
			Comment("支付平台订单号"),
		field.Int64("user_id").
			Comment("用户ID"),
		field.Int64("group_id").
			Optional().
			Nillable().
			Comment("分组ID（历史字段，AliMPay 充值订单不使用）"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("订单金额"),
		field.Float("payment_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Optional().
			Comment("实际支付金额（唯一金额，用于匹配账单）"),
		field.Float("credit_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Default(0).
			Comment("实际到账余额"),
		field.Float("multiplier").
			Default(1.0).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("倍率快照"),
		field.String("status").
			MaxLen(20).
			Default("pending").
			Comment("订单状态: pending/paid/expired/refunded"),
		field.String("pay_type").
			MaxLen(20).
			Optional().
			Nillable().
			Comment("支付方式: alipay/wxpay"),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("支付时间"),
		field.String("source_domain").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("下单时的来源域名（审计用，归属走 User.register_domain）"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("创建时间"),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("更新时间"),
		field.Time("expired_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("订单过期时间"),
	}
}

func (Order) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("orders").
			Field("user_id").
			Unique().
			Required(),
		edge.From("group", Group.Type).
			Ref("orders").
			Field("group_id").
			Unique(),
		// Referral reward triggered by this order (if any)
		edge.To("referral_reward", ReferralReward.Type).
			Unique(),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("group_id"),
		index.Fields("status"),
		index.Fields("payment_amount").
			Unique().
			Annotations(entsql.IndexWhere("status = 'pending' AND payment_amount IS NOT NULL")),
		index.Fields("order_no"),
		index.Fields("created_at"),
	}
}
