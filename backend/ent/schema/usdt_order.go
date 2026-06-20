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

// UsdtOrder holds the schema definition for the USDT (TRC20) recharge order entity.
//
// 与 AliMPay 的 Order 并列：业务都是「CNY 计价 → 平台余额入账」，差异在支付通道——
// USDT 走链上收款（无 webhook，由 UsdtMonitor 轮询 TronGrid 按唯一金额匹配）。
// 价钱仍以 CNY 计（amount/credit_amount，1:1 入账），usdt_amount 是按下单时
// 冻结汇率换算出的、带唯一尾数的链上应付金额（6 位小数，独立列避免 decimal(10,2) 精度坑）。
type UsdtOrder struct {
	ent.Schema
}

func (UsdtOrder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "usdt_orders"},
	}
}

func (UsdtOrder) Fields() []ent.Field {
	return []ent.Field{
		field.String("order_no").
			MaxLen(64).
			NotEmpty().
			Unique().
			Comment("商户订单号（U 前缀）"),
		field.String("trade_no").
			MaxLen(80).
			Optional().
			Nillable().
			Comment("链上交易哈希（确认后回填）"),
		field.Int64("user_id").
			Comment("用户ID"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("订单金额（CNY 计价）"),
		field.Float("credit_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Default(0).
			Comment("实际到账余额（= amount，1:1）"),
		field.Float("multiplier").
			Default(1.0).
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("倍率快照"),
		field.String("chain").
			MaxLen(20).
			Default("trc20").
			Comment("链/网络：trc20"),
		field.String("receiving_address").
			MaxLen(64).
			Comment("收款地址快照（下单时冻结）"),
		field.Float("usdt_rate").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0).
			Comment("下单时冻结的汇率：1 USDT = ? CNY"),
		field.Float("usdt_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,6)"}).
			Comment("应付 USDT 金额（含唯一尾数，用于链上匹配）"),
		field.Float("paid_usdt_amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,6)"}).
			Optional().
			Nillable().
			Comment("实际收到的 USDT 金额"),
		field.String("from_address").
			MaxLen(64).
			Optional().
			Nillable().
			Comment("付款方地址"),
		field.Int64("block_number").
			Optional().
			Nillable().
			Comment("到账交易所在区块高度"),
		field.String("status").
			MaxLen(20).
			Default("pending").
			Comment("订单状态: pending/paid/expired/refunded"),
		field.String("pay_type").
			MaxLen(20).
			Default("usdt").
			Optional().
			Nillable().
			Comment("支付方式: usdt"),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}).
			Comment("支付时间"),
		field.String("source_domain").
			MaxLen(255).
			Optional().
			Nillable().
			Comment("下单时的来源域名（审计用）"),
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

func (UsdtOrder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("status"),
		index.Fields("order_no"),
		index.Fields("created_at"),
		// 唯一应付金额：仅在 (chain, usdt_amount) 上对 pending 订单去重，
		// 保证同一收款地址下并发订单的链上金额各不相同，可据此唯一归属。
		index.Fields("chain", "usdt_amount").
			Unique().
			Annotations(entsql.IndexWhere("status = 'pending'")),
		// tx_hash 幂等：同一笔链上转账只能确认一个订单。
		index.Fields("trade_no").
			Unique().
			Annotations(entsql.IndexWhere("trade_no IS NOT NULL")),
	}
}
