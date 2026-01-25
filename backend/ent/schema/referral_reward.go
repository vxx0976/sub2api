package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ReferralReward holds the schema definition for the ReferralReward entity.
type ReferralReward struct {
	ent.Schema
}

func (ReferralReward) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "referral_rewards"},
	}
}

func (ReferralReward) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
	}
}

func (ReferralReward) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("referrer_id").
			Comment("邀请人ID"),
		field.Int64("invitee_id").
			Unique().
			Comment("被邀请人ID（每个被邀请人只能触发一次奖励）"),
		field.Int64("trigger_order_id").
			Comment("触发奖励的订单ID"),
		field.Float("referrer_reward").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Default(0).
			Comment("邀请人获得的奖励金额"),
		field.Float("invitee_reward").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Default(0).
			Comment("被邀请人获得的奖励金额"),
		field.String("skip_referrer_reason").
			MaxLen(100).
			Optional().
			Nillable().
			Comment("跳过邀请人奖励的原因（如：邀请人已被封禁）"),
	}
}

func (ReferralReward) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("referrer", User.Type).
			Ref("referral_rewards_given").
			Field("referrer_id").
			Unique().
			Required(),
		edge.From("invitee", User.Type).
			Ref("referral_reward_received").
			Field("invitee_id").
			Unique().
			Required(),
		edge.From("trigger_order", Order.Type).
			Ref("referral_reward").
			Field("trigger_order_id").
			Unique().
			Required(),
	}
}

func (ReferralReward) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("referrer_id"),
		index.Fields("invitee_id"),
		index.Fields("trigger_order_id"),
		index.Fields("created_at"),
	}
}
