package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"
	"github.com/Wei-Shaw/sub2api/internal/domain"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		// 唯一约束通过部分索引实现（WHERE deleted_at IS NULL），支持软删除后重用
		// 见迁移文件 016_soft_delete_partial_unique_indexes.sql
		field.String("email").
			MaxLen(255).
			NotEmpty(),
		field.String("password_hash").
			MaxLen(255).
			NotEmpty(),
		field.String("role").
			MaxLen(20).
			Default(domain.RoleUser),
		field.Float("balance").
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}).
			Default(0),
		field.Int("concurrency").
			Default(5),
		field.String("status").
			MaxLen(20).
			Default(domain.StatusActive),

		// Optional profile fields (added later; default '' in DB migration)
		field.String("username").
			MaxLen(100).
			Default(""),
		// wechat field migrated to user_attribute_values (see migration 019)
		field.String("notes").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Default(""),

		// Referral system fields
		field.String("referral_code").
			MaxLen(8).
			Optional().
			Unique().
			Nillable().
			Comment("用户专属邀请码（格式：R + 7位字符）"),
		field.Int64("referred_by").
			Optional().
			Nillable().
			Comment("邀请人用户ID"),
		field.Bool("referral_rewarded").
			Default(false).
			Comment("是否已发放邀请奖励"),

		// TOTP 双因素认证字段
		field.String("totp_secret_encrypted").
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Optional().
			Nillable(),
		field.Bool("totp_enabled").
			Default(false),
		field.Time("totp_enabled_at").
			Optional().
			Nillable(),

		// 分销商层级关系
		field.Int64("parent_id").
			Optional().
			Nillable().
			Comment("上级分销商用户 ID"),

		// Token 版本控制（密码修改时递增，使旧 token 失效）
		field.Int64("token_version").
			Default(0).
			Comment("Token 版本号，密码修改时递增"),

		// 角色版本控制（角色变更时递增，使旧 token 失效）
		field.Int64("role_version").
			Default(0).
			Comment("角色版本号，角色变更时递增"),
	}
}

func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("api_keys", APIKey.Type),
		edge.To("redeem_codes", RedeemCode.Type),
		edge.To("subscriptions", UserSubscription.Type),
		edge.To("assigned_subscriptions", UserSubscription.Type),
		edge.To("announcement_reads", AnnouncementRead.Type),
		edge.To("allowed_groups", Group.Type).
			Through("user_allowed_groups", UserAllowedGroup.Type),
		edge.To("usage_logs", UsageLog.Type),
		edge.To("attribute_values", UserAttributeValue.Type),
		edge.To("promo_code_usages", PromoCodeUsage.Type),
		edge.To("orders", Order.Type),
		edge.To("recharge_orders", RechargeOrder.Type),
		// Referral system edges
		edge.To("referral_rewards_given", ReferralReward.Type),
		edge.To("referral_reward_received", ReferralReward.Type),
		// 分销商层级：一个用户有多个下级用户
		edge.To("sub_users", User.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
		// 分销商层级：一个用户有一个上级（反向边）
		edge.From("parent", User.Type).
			Ref("sub_users").
			Field("parent_id").
			Unique(),
		// 分销商自定义域名
		edge.To("reseller_domains", ResellerDomain.Type),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		// email 字段已在 Fields() 中声明 Unique()，无需重复索引
		index.Fields("status"),
		index.Fields("deleted_at"),
		index.Fields("parent_id"),
	}
}
