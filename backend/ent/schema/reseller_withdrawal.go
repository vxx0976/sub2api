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

// ResellerWithdrawal 提现申请表
type ResellerWithdrawal struct {
	ent.Schema
}

func (ResellerWithdrawal) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reseller_withdrawals"},
	}
}

func (ResellerWithdrawal) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("reseller_id"),
		field.Float("amount").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,10)"}),
		field.String("status").
			MaxLen(20).
			Default("pending"),
		field.String("payment_method").
			MaxLen(50),
		field.String("payment_account").
			MaxLen(512),
		field.String("payment_name").
			MaxLen(100),
		field.String("admin_notes").
			MaxLen(2000).
			Default(""),
		field.Int64("admin_id").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("paid_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
		field.Time("rejected_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "timestamptz"}),
	}
}

func (ResellerWithdrawal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("reseller_id"),
		index.Fields("status"),
	}
}
