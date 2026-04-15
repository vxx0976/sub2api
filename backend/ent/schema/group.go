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

// Group holds the schema definition for the Group entity.
type Group struct {
	ent.Schema
}

func (Group) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "groups"},
	}
}

func (Group) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (Group) Fields() []ent.Field {
	return []ent.Field{
		// 唯一约束通过部分索引实现（WHERE deleted_at IS NULL），支持软删除后重用
		// 见迁移文件 016_soft_delete_partial_unique_indexes.sql
		field.String("name").
			MaxLen(100).
			NotEmpty(),
		field.String("description").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}),
		field.JSON("name_i18n", map[string]string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("多语言名称，如 {\"en\": \"English Name\", \"ru\": \"Русское имя\"}"),
		field.JSON("description_i18n", map[string]string{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("多语言描述，如 {\"en\": \"English desc\", \"ru\": \"Русское описание\"}"),
		field.Float("rate_multiplier").
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,4)"}).
			Default(1.0),
		field.Bool("is_exclusive").
			Default(false),
		field.String("status").
			MaxLen(20).
			Default(domain.StatusActive),

		// Subscription-related fields (added by migration 003)
		field.String("platform").
			MaxLen(50).
			Default(domain.PlatformAnthropic),
		field.String("subscription_type").
			MaxLen(20).
			Default(domain.SubscriptionTypeStandard),
		field.Float("daily_limit_usd").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("weekly_limit_usd").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("monthly_limit_usd").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Int("default_validity_days").
			Default(30),

		// 图片生成计费配置（antigravity 和 gemini 平台使用）
		field.Float("image_price_1k").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("image_price_2k").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),
		field.Float("image_price_4k").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(20,8)"}),

		// Claude Code 客户端限制 (added by migration 029)
		field.Bool("claude_code_only").
			Default(false).
			Comment("是否仅允许 Claude Code 客户端"),
		field.Int64("fallback_group_id").
			Optional().
			Nillable().
			Comment("非 Claude Code 请求降级使用的分组 ID"),
		field.Int64("fallback_group_id_on_invalid_request").
			Optional().
			Nillable().
			Comment("无效请求兜底使用的分组 ID"),

		// 模型路由配置 (added by migration 040)
		field.JSON("model_routing", map[string][]int64{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("模型路由配置：模型模式 -> 优先账号ID列表"),

		// 模型路由开关 (added by migration 041)
		field.Bool("model_routing_enabled").
			Default(false).
			Comment("是否启用模型路由配置"),

		// MCP XML 协议注入开关 (added by migration 042)
		field.Bool("mcp_xml_inject").
			Default(true).
			Comment("是否注入 MCP XML 调用协议提示词（仅 antigravity 平台）"),

		// 支持的模型系列 (added by migration 046)
		field.JSON("supported_model_scopes", []string{}).
			Default([]string{"claude", "gemini_text", "gemini_image"}).
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("支持的模型系列：claude, gemini_text, gemini_image"),

		// 支付相关字段 (added by migration 044)
		field.Float("price").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "decimal(10,2)"}).
			Comment("套餐售价（元）"),
		field.Bool("is_purchasable").
			Default(false).
			Comment("是否允许用户购买"),
		field.Int("sort_order").
			Default(0).
			Comment("分组显示排序，数值越小越靠前"),
		field.Bool("is_recommended").
			Default(false).
			Comment("是否推荐套餐"),
		field.String("external_buy_url").
			Optional().
			Nillable().
			SchemaType(map[string]string{dialect.Postgres: "text"}).
			Comment("外部购买链接（如淘宝）"),

		// 分销商所有权
		field.Int64("owner_id").
			Optional().
			Nillable().
			Comment("分销商用户 ID（NULL=管理员分组）"),
		field.Int64("source_group_id").
			Optional().
			Nillable().
			Comment("克隆来源的管理员模板分组 ID"),
		field.Bool("reseller_template").
			Default(false).
			Comment("是否标记为分销商可用模板"),

		// 账号过滤 (added by migration 074)
		field.Bool("require_oauth_only").
			Default(false).
			Comment("仅允许非 apikey 类型账号关联到此分组"),
		field.Bool("require_privacy_set").
			Default(false).
			Comment("调度时仅允许 privacy 已成功设置的账号"),

		// OpenAI Messages 调度配置 (added by migration 069)
		field.Bool("allow_messages_dispatch").
			Default(false).
			Comment("是否允许 /v1/messages 调度到此 OpenAI 分组"),
		field.String("default_mapped_model").
			MaxLen(100).
			Default("").
			Comment("默认映射模型 ID，当账号级映射找不到时使用此值"),

		// 定时上线时间窗口（格式 "HH:MM"，两者都设置时生效）
		field.String("active_start_time").
			Optional().
			Nillable().
			MaxLen(5).
			Comment("每日可用开始时间（HH:MM），与 active_end_time 配合使用"),
		field.String("active_end_time").
			Optional().
			Nillable().
			MaxLen(5).
			Comment("每日可用结束时间（HH:MM），与 active_start_time 配合使用"),

		// 健康检查间隔（分钟），默认 30
		field.Int("health_check_interval_min").
			Default(30).
			Comment("健康检查间隔（分钟），默认 30"),

		// 健康检查自定义测试模型（空=使用平台默认）
		field.String("health_check_test_model").
			Default("").
			MaxLen(128).
			Comment("健康检查使用的测试模型 ID，空表示按平台使用默认"),

		// 分组健康检查状态
		field.String("health_status").
			MaxLen(20).
			Default("").
			Comment("分组可用性状态：available/unavailable/空=未检查"),
		field.Int("healthy_accounts").
			Default(0).
			Comment("最近一次健康检查中可用的账号数"),
		field.Int("total_checked_accounts").
			Default(0).
			Comment("最近一次健康检查中检查的账号总数"),
		field.Time("last_health_check_at").
			Optional().
			Nillable().
			Comment("最近一次健康检查时间"),

		// 智能路由（故障转移）分组字段 (added by migration 095)
		field.Bool("is_failover_group").
			Default(false).
			Comment("是否为智能路由（虚拟故障转移分组）"),
		field.JSON("failover_member_ids", []int64{}).
			Optional().
			SchemaType(map[string]string{dialect.Postgres: "jsonb"}).
			Comment("故障转移成员分组 ID 有序列表（仅 is_failover_group=true 时生效）"),
		field.Int64("failover_active_member_id").
			Optional().
			Nillable().
			Comment("当前激活成员 ID；为空表示尚未 reconcile"),
		field.Int64("failover_active_version").
			Default(0).
			Comment("active_member_id 的乐观锁版本号（CAS）"),
		field.Int64("failover_pin_member_id").
			Optional().
			Nillable().
			Comment("手动锁定的成员 ID；非空时自动 reconcile 被短路"),
		field.Time("failover_pin_expires_at").
			Optional().
			Nillable().
			Comment("手动锁定过期时间；到期后自动清除"),
	}
}

func (Group) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("api_keys", APIKey.Type),
		edge.To("redeem_codes", RedeemCode.Type),
		edge.To("subscriptions", UserSubscription.Type),
		edge.To("usage_logs", UsageLog.Type),
		edge.To("orders", Order.Type),
		edge.From("accounts", Account.Type).
			Ref("groups").
			Through("account_groups", AccountGroup.Type),
		edge.From("allowed_users", User.Type).
			Ref("allowed_groups").
			Through("user_allowed_groups", UserAllowedGroup.Type),
		// 注意：fallback_group_id 直接作为字段使用，不定义 edge
		// 这样允许多个分组指向同一个降级分组（M2O 关系）
	}
}

func (Group) Indexes() []ent.Index {
	return []ent.Index{
		// name 字段已在 Fields() 中声明 Unique()，无需重复索引
		index.Fields("status"),
		index.Fields("platform"),
		index.Fields("subscription_type"),
		index.Fields("is_exclusive"),
		index.Fields("deleted_at"),
		index.Fields("sort_order"),
		index.Fields("owner_id"),
	}
}
