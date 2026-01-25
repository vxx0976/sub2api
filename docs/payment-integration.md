# Sub2API 支付系统集成方案

## 一、概述

### 1.1 目标

集成暮色聚合支付系统，实现用户在线购买套餐订阅功能。

### 1.2 支付系统信息

- **支付平台**: 暮色聚合支付
- **文档地址**: https://ms.1212z.cn/doc_old.html
- **支付接口**: https://ms.1212z.cn/submit.php (页面跳转)

### 1.3 整体流程

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            用户购买套餐流程                               │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐          │
│  │  首页    │───▶│ 控制台   │───▶│ 套餐页面 │───▶│ 选择分组 │          │
│  │ 套餐展示 │    │ (需登录) │    │          │    │ 点击购买 │          │
│  └──────────┘    └──────────┘    └──────────┘    └────┬─────┘          │
│                                                       │                 │
│                                                       ▼                 │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐    ┌──────────┐          │
│  │ 订阅管理 │◀───│ 分配订阅 │◀───│ 验证签名 │◀───│ 创建订单 │          │
│  │ 页面     │    │          │    │ 更新订单 │    │ 跳转支付 │          │
│  └──────────┘    └──────────┘    └──────────┘    └────┬─────┘          │
│       ▲                              ▲                │                 │
│       │                              │                ▼                 │
│       │                         ┌────┴─────┐    ┌──────────┐          │
│       └─────────────────────────│ 支付回调 │◀───│ 支付系统 │          │
│              (重定向)           │ notify   │    │ 支付完成 │          │
│                                 └──────────┘    └──────────┘          │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## 二、数据库设计

### 2.1 扩展 groups 表（添加套餐字段）

```sql
ALTER TABLE groups ADD COLUMN IF NOT EXISTS price DECIMAL(10,2);
ALTER TABLE groups ADD COLUMN IF NOT EXISTS validity_days INT DEFAULT 30;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_purchasable BOOLEAN DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS sort_order INT DEFAULT 0;
```

### 2.2 新增订单表 `orders`

```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,     -- 商户订单号
    trade_no VARCHAR(64),                      -- 支付平台订单号
    user_id BIGINT NOT NULL REFERENCES users(id),
    group_id BIGINT NOT NULL REFERENCES groups(id),
    amount DECIMAL(10,2) NOT NULL,            -- 订单金额
    status VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/paid/expired/refunded
    pay_type VARCHAR(20),                     -- alipay/wxpay
    paid_at TIMESTAMPTZ,                      -- 支付时间
    subscription_id BIGINT REFERENCES user_subscriptions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMPTZ                    -- 订单过期时间
);

CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_order_no ON orders(order_no);
```

### 2.3 迁移文件

创建 `backend/migrations/XXX_add_payment.sql`:

```sql
-- +migrate Up

-- 扩展 groups 表
ALTER TABLE groups ADD COLUMN IF NOT EXISTS price DECIMAL(10,2);
ALTER TABLE groups ADD COLUMN IF NOT EXISTS validity_days INT DEFAULT 30;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS is_purchasable BOOLEAN DEFAULT FALSE;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE groups ADD COLUMN IF NOT EXISTS sort_order INT DEFAULT 0;

-- 订单表
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    trade_no VARCHAR(64),
    user_id BIGINT NOT NULL REFERENCES users(id),
    group_id BIGINT NOT NULL REFERENCES groups(id),
    amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    pay_type VARCHAR(20),
    paid_at TIMESTAMPTZ,
    subscription_id BIGINT REFERENCES user_subscriptions(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_order_no ON orders(order_no);

-- +migrate Down
DROP TABLE IF EXISTS orders;
ALTER TABLE groups DROP COLUMN IF EXISTS price;
ALTER TABLE groups DROP COLUMN IF EXISTS validity_days;
ALTER TABLE groups DROP COLUMN IF EXISTS is_purchasable;
ALTER TABLE groups DROP COLUMN IF EXISTS description;
ALTER TABLE groups DROP COLUMN IF EXISTS sort_order;
```

---

## 三、后端开发

### 3.1 配置项

`config.yaml` 添加：

```yaml
payment:
  enabled: true
  muse:
    pid: "YOUR_MERCHANT_ID"
    key: "YOUR_MERCHANT_KEY"
    submit_url: "https://ms.1212z.cn/submit.php"
    notify_url: "https://your-domain.com/api/v1/payment/notify"
    return_url: "https://your-domain.com/payment/result"
```

### 3.2 支付签名工具

页面跳转方式只需要拼接 URL 参数并签名，不需要调用 API。

`backend/internal/pkg/payment/muse.go`:

```go
package payment

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type MuseConfig struct {
	PID       string
	Key       string
	SubmitURL string
	NotifyURL string
	ReturnURL string
}

type MusePayment struct {
	config MuseConfig
}

func NewMusePayment(config MuseConfig) *MusePayment {
	return &MusePayment{config: config}
}

type CreatePayParams struct {
	OrderNo string
	Money   float64
	Name    string
	Type    string // alipay/wxpay，空则跳转收银台
}

// PaymentURL 生成支付跳转URL
func (m *MusePayment) PaymentURL(params CreatePayParams) string {
	data := map[string]string{
		"pid":          m.config.PID,
		"out_trade_no": params.OrderNo,
		"notify_url":   m.config.NotifyURL,
		"return_url":   m.config.ReturnURL,
		"name":         params.Name,
		"money":        fmt.Sprintf("%.2f", params.Money),
		"sign_type":    "MD5",
	}
	if params.Type != "" {
		data["type"] = params.Type
	}
	data["sign"] = m.Sign(data)

	values := url.Values{}
	for k, v := range data {
		values.Set(k, v)
	}
	return m.config.SubmitURL + "?" + values.Encode()
}

// Sign 生成签名
func (m *MusePayment) Sign(params map[string]string) string {
	// 过滤空值和签名字段
	filtered := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "sign_type" && v != "" {
			filtered[k] = v
		}
	}

	// 按key排序
	keys := make([]string, 0, len(filtered))
	for k := range filtered {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接字符串
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, filtered[k]))
	}
	str := strings.Join(parts, "&") + m.config.Key

	// MD5
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}

// VerifySign 验证签名
func (m *MusePayment) VerifySign(params map[string]string) bool {
	sign := params["sign"]
	if sign == "" {
		return false
	}
	return m.Sign(params) == sign
}

type NotifyParams struct {
	PID         string `form:"pid"`
	TradeNo     string `form:"trade_no"`
	OutTradeNo  string `form:"out_trade_no"`
	Type        string `form:"type"`
	Name        string `form:"name"`
	Money       string `form:"money"`
	TradeStatus string `form:"trade_status"`
	Sign        string `form:"sign"`
	SignType    string `form:"sign_type"`
}

func (n *NotifyParams) IsSuccess() bool {
	return n.TradeStatus == "TRADE_SUCCESS"
}

func (n *NotifyParams) ToMap() map[string]string {
	return map[string]string{
		"pid":          n.PID,
		"trade_no":     n.TradeNo,
		"out_trade_no": n.OutTradeNo,
		"type":         n.Type,
		"name":         n.Name,
		"money":        n.Money,
		"trade_status": n.TradeStatus,
		"sign":         n.Sign,
		"sign_type":    n.SignType,
	}
}
```

### 3.3 订单服务

`backend/internal/service/order_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
)

type OrderService struct {
	orderRepo       OrderRepository
	groupRepo       GroupRepository
	subscriptionSvc *SubscriptionService
	payment         *payment.MusePayment
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, userID, groupID int64) (*Order, string, error) {
	// 获取分组信息
	group, err := s.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return nil, "", fmt.Errorf("分组不存在")
	}
	if !group.IsPurchasable || group.Price == nil {
		return nil, "", fmt.Errorf("该套餐不可购买")
	}

	// 生成订单号
	orderNo := fmt.Sprintf("%s%s", time.Now().Format("20060102150405"), uuid.New().String()[:8])

	// 创建订单
	order := &Order{
		OrderNo:   orderNo,
		UserID:    userID,
		GroupID:   groupID,
		Amount:    *group.Price,
		Status:    "pending",
		CreatedAt: time.Now(),
		ExpiredAt: time.Now().Add(30 * time.Minute),
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, "", err
	}

	// 生成支付URL
	payURL := s.payment.PaymentURL(payment.CreatePayParams{
		OrderNo: orderNo,
		Money:   *group.Price,
		Name:    group.Name,
	})

	return order, payURL, nil
}

// HandleNotify 处理支付回调
func (s *OrderService) HandleNotify(ctx context.Context, params *payment.NotifyParams) error {
	// 验证签名
	if !s.payment.VerifySign(params.ToMap()) {
		return fmt.Errorf("签名验证失败")
	}

	// 查询订单
	order, err := s.orderRepo.GetByOrderNo(ctx, params.OutTradeNo)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}

	// 幂等检查
	if order.Status == "paid" {
		return nil
	}

	// 检查支付状态
	if !params.IsSuccess() {
		return fmt.Errorf("支付未成功")
	}

	// 检查金额
	if params.Money != fmt.Sprintf("%.2f", order.Amount) {
		return fmt.Errorf("金额不匹配")
	}

	// 事务处理
	return s.orderRepo.Transaction(ctx, func(tx context.Context) error {
		// 更新订单
		now := time.Now()
		order.Status = "paid"
		order.TradeNo = params.TradeNo
		order.PayType = params.Type
		order.PaidAt = &now
		if err := s.orderRepo.Update(tx, order); err != nil {
			return err
		}

		// 获取分组
		group, err := s.groupRepo.GetByID(tx, order.GroupID)
		if err != nil {
			return err
		}

		// 分配订阅
		subscription, err := s.subscriptionSvc.CreateSubscription(tx, &CreateSubscriptionParams{
			UserID:          order.UserID,
			GroupID:         order.GroupID,
			ValidityDays:    group.ValidityDays,
			DailyLimitUSD:   group.DailyLimitUSD,
			WeeklyLimitUSD:  group.WeeklyLimitUSD,
			MonthlyLimitUSD: group.MonthlyLimitUSD,
		})
		if err != nil {
			return err
		}

		// 关联订阅
		order.SubscriptionID = &subscription.ID
		return s.orderRepo.Update(tx, order)
	})
}
```

### 3.4 API接口

`backend/internal/handler/payment_handler.go`:

```go
package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	orderSvc *service.OrderService
	groupSvc *service.GroupService
}

// GetPurchasablePlans 获取可购买套餐
// GET /api/v1/plans
func (h *PaymentHandler) GetPurchasablePlans(c *gin.Context) {
	groups, err := h.groupSvc.GetPurchasableGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "获取失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": groups})
}

// CreateOrder 创建订单
// POST /api/v1/orders
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req struct {
		GroupID int64 `json:"group_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "参数错误"})
		return
	}

	userID := c.GetInt64("user_id")
	order, payURL, err := h.orderSvc.CreateOrder(c.Request.Context(), userID, req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"order_no": order.OrderNo,
			"pay_url":  payURL,
		},
	})
}

// PaymentNotify 支付回调
// GET /api/v1/payment/notify
func (h *PaymentHandler) PaymentNotify(c *gin.Context) {
	var params payment.NotifyParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	if err := h.orderSvc.HandleNotify(c.Request.Context(), &params); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}

	c.String(http.StatusOK, "success")
}

// PaymentReturn 支付返回重定向
// GET /api/v1/payment/return
func (h *PaymentHandler) PaymentReturn(c *gin.Context) {
	orderNo := c.Query("out_trade_no")
	status := c.Query("trade_status")

	if status == "TRADE_SUCCESS" {
		c.Redirect(http.StatusFound, "/subscriptions?order="+orderNo+"&status=success")
	} else {
		c.Redirect(http.StatusFound, "/subscriptions?order="+orderNo+"&status=failed")
	}
}
```

### 3.5 路由配置

```go
// 公开
public.GET("/plans", paymentHandler.GetPurchasablePlans)
public.GET("/payment/notify", paymentHandler.PaymentNotify)
public.GET("/payment/return", paymentHandler.PaymentReturn)

// 需登录
user.POST("/orders", paymentHandler.CreateOrder)
user.GET("/orders", paymentHandler.GetMyOrders)
```

---

## 四、前端开发

### 4.1 API

`frontend/src/api/plans.ts`:

```typescript
import request from '@/utils/request'

export interface PurchasablePlan {
  id: number
  name: string
  description: string
  price: number
  validity_days: number
  monthly_limit_usd: number
  is_recommended: boolean
}

export const plansApi = {
  getPlans: () => request.get<PurchasablePlan[]>('/plans'),
  createOrder: (groupId: number) => request.post<{ order_no: string; pay_url: string }>('/orders', { group_id: groupId })
}
```

### 4.2 套餐页面

`frontend/src/views/user/PlansView.vue`:

```vue
<template>
  <div class="mx-auto max-w-7xl px-4 py-8">
    <div class="mb-8 text-center">
      <h1 class="text-2xl font-bold text-gray-900 dark:text-white">选择套餐</h1>
      <p class="mt-2 text-gray-600 dark:text-dark-400">选择适合您的套餐</p>
    </div>

    <div class="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
      <div
        v-for="plan in plans"
        :key="plan.id"
        class="relative rounded-2xl border p-6"
        :class="plan.is_recommended ? 'border-primary-500 ring-2 ring-primary-500/20' : 'border-gray-200 dark:border-dark-700'"
      >
        <span v-if="plan.is_recommended" class="absolute -top-3 left-1/2 -translate-x-1/2 rounded-full bg-primary-500 px-3 py-1 text-xs text-white">推荐</span>

        <h3 class="text-lg font-semibold">{{ plan.name }}</h3>
        <div class="mt-4">
          <span class="text-3xl font-bold">¥{{ plan.price }}</span>
          <span class="text-gray-500">/{{ plan.validity_days }}天</span>
        </div>
        <p class="mt-2 text-sm text-gray-600">{{ plan.description }}</p>
        <ul class="mt-4 space-y-2 text-sm">
          <li>月限额 ${{ plan.monthly_limit_usd }}</li>
          <li>有效期 {{ plan.validity_days }} 天</li>
        </ul>

        <button
          @click="handleBuy(plan.id)"
          :disabled="buying === plan.id"
          class="mt-6 w-full rounded-xl bg-primary-500 py-2 text-white hover:bg-primary-600 disabled:opacity-50"
        >
          {{ buying === plan.id ? '处理中...' : '立即购买' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { plansApi, type PurchasablePlan } from '@/api/plans'

const plans = ref<PurchasablePlan[]>([])
const buying = ref<number | null>(null)

onMounted(async () => {
  const { data } = await plansApi.getPlans()
  plans.value = data
})

async function handleBuy(groupId: number) {
  buying.value = groupId
  try {
    const { data } = await plansApi.createOrder(groupId)
    window.location.href = data.pay_url
  } finally {
    buying.value = null
  }
}
</script>
```

### 4.3 路由

```typescript
{
  path: '/plans',
  name: 'Plans',
  component: () => import('@/views/user/PlansView.vue'),
  meta: { requiresAuth: true }
}
```

---

## 五、安全要点

1. **签名验证** - 所有回调必须验签
2. **幂等处理** - 防止重复分配订阅
3. **金额校验** - 回调金额必须与订单一致
4. **订单过期** - 30分钟未支付自动过期

---

## 六、上线检查

- [ ] 配置支付密钥（pid/key）
- [ ] 配置回调地址（HTTPS）
- [ ] 执行数据库迁移
- [ ] 设置分组的 price/is_purchasable
- [ ] 测试完整支付流程
