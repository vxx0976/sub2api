package schema

import (
	"github.com/Wei-Shaw/sub2api/ent/schema/mixins"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ResellerDomain holds the schema definition for the ResellerDomain entity.
type ResellerDomain struct {
	ent.Schema
}

func (ResellerDomain) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "reseller_domains"},
	}
}

func (ResellerDomain) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.TimeMixin{},
		mixins.SoftDeleteMixin{},
	}
}

func (ResellerDomain) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("reseller_id").
			Comment("分销商用户 ID"),
		field.String("domain").
			MaxLen(255).
			NotEmpty().
			Comment("分销商自定义域名"),
		field.String("site_name").
			MaxLen(100).
			Default("").
			Comment("站点名称"),
		field.String("site_logo").
			Default("").
			Comment("站点 Logo（base64 或 URL）"),
		field.String("brand_color").
			MaxLen(20).
			Default("").
			Comment("品牌主色调（如 #6366f1）"),
		field.String("custom_css").
			Default("").
			Comment("自定义 CSS"),
		field.String("subtitle").
			MaxLen(255).
			Default("").
			Comment("站点副标题"),
		field.String("home_content").
			Default("").
			Comment("首页自定义内容（HTML 或外部 URL）"),
		field.String("home_template").
			MaxLen(50).
			Default("").
			Comment("首页模板类型：default/minimal/custom_html/external_url"),
		field.String("doc_url").
			MaxLen(500).
			Default("").
			Comment("文档链接"),
		field.Bool("purchase_enabled").
			Default(false).
			Comment("是否启用购买页"),
		field.String("purchase_url").
			MaxLen(500).
			Default("").
			Comment("自定义购买链接"),
		field.String("default_locale").
			MaxLen(20).
			Default("").
			Comment("站点默认语言"),
		field.String("seo_title").
			MaxLen(255).
			Default("").
			Comment("SEO 标题"),
		field.String("seo_description").
			Default("").
			Comment("SEO 描述"),
		field.String("seo_keywords").
			MaxLen(500).
			Default("").
			Comment("SEO 关键词"),
		field.String("login_redirect").
			MaxLen(255).
			Default("").
			Comment("登录后跳转路径"),
		field.String("verify_token").
			MaxLen(64).
			Default("").
			Comment("DNS TXT 验证 token"),
		field.Bool("verified").
			Default(false).
			Comment("域名是否已验证"),
		field.Time("verified_at").
			Optional().
			Nillable().
			Comment("域名验证通过时间"),
	}
}

func (ResellerDomain) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("reseller", User.Type).
			Ref("reseller_domains").
			Field("reseller_id").
			Unique().
			Required(),
	}
}

func (ResellerDomain) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("domain").Unique(),
		index.Fields("reseller_id"),
		index.Fields("verified"),
	}
}
