/**
 * User Channels API endpoints (non-admin)
 * 用户侧「可用渠道」聚合查询：渠道 + 用户可访问的分组 + 支持模型（含定价）。
 */

import { apiClient } from './client'
import type { BillingMode } from '@/constants/channel'

export interface UserAvailableGroup {
  id: number
  name: string
  platform: string
  /** 'standard' | 'subscription' — 订阅分组视觉加深，和 API 密钥页保持一致。 */
  subscription_type: string
  /** 分组默认倍率。用户专属倍率（若有）通过 /groups/rates 获取后在前端 join。 */
  rate_multiplier: number
  peak_rate_enabled: boolean
  peak_start: string
  peak_end: string
  peak_rate_multiplier: number
  /** true = 专属分组（小范围授权）；false = 公开分组。 */
  is_exclusive: boolean
}

export interface UserPricingInterval {
  min_tokens: number
  max_tokens: number | null
  tier_label?: string
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  per_request_price: number | null
}

export interface UserSupportedModelPricing {
  billing_mode: BillingMode
  input_price: number | null
  output_price: number | null
  cache_write_price: number | null
  cache_read_price: number | null
  image_input_price: number | null
  image_output_price: number | null
  per_request_price: number | null
  intervals: UserPricingInterval[]
}

export interface UserSupportedModel {
  name: string
  platform: string
  pricing: UserSupportedModelPricing | null
}

/**
 * 渠道下单个平台的子视图：用户可访问的分组 + 该平台支持的模型。
 * 后端把一个渠道按平台聚合成 sections，前端可以把渠道名作为 row-group
 * 一次渲染，后面按 sections 顺序用 rowspan 铺开。
 */
export interface UserChannelPlatformSection {
  platform: string
  groups: UserAvailableGroup[]
  supported_models: UserSupportedModel[]
}

export interface UserAvailableChannel {
  name: string
  description: string
  platforms: UserChannelPlatformSection[]
}

/** 列出当前用户可见的「可用渠道」（与 /groups/available 保持一致，返回平数组）。 */
export async function getAvailable(options?: { signal?: AbortSignal }): Promise<UserAvailableChannel[]> {
  const { data } = await apiClient.get<UserAvailableChannel[]>('/channels/available', {
    signal: options?.signal
  })
  return data
}

/**
 * 「模型广场」端点（group）下的单条模型定价。
 * input_price 等为渠道显式配置的基础单价（USD / per token），未配置时为 null，
 * 前端回退到对应 official_*（LiteLLM 官方价，USD / per token）。site 模式按
 * group.rate_multiplier / cny_per_usd 换算本站价，与计费链路一致。
 */
export interface UserPricingModel {
  name: string
  /** 'token'（默认）按 token 计费；'per_request' / 'image' 渲染 intervals 子表。 */
  billing_mode?: BillingMode | string
  input_price?: number | null
  output_price?: number | null
  cache_write_price?: number | null
  cache_read_price?: number | null
  /** per_request / image 模式：每次请求/每图片的基础单价（USD）。 */
  per_request_price?: number | null
  /** per_request / image 模式：tier 分层定价。复用 UserPricingInterval。 */
  intervals?: UserPricingInterval[]
  /** LiteLLM 官方价（USD / per token）。模型不在 LiteLLM 表里或为 0 时缺失。 */
  official_input_price?: number | null
  official_output_price?: number | null
  official_cache_write_price?: number | null
  official_cache_read_price?: number | null
  /** 价格币种口径：'CNY' = 国产人民币计价模型（显示 ¥），其余 'USD'。与用量页同源。 */
  price_currency?: string
}

/** 「模型广场」展示页的端点 = 一个 group。models 由后端按账号映射交集/LiteLLM 兜底解析。 */
export interface UserPricingGroup {
  id: number
  name: string
  platform: string
  rate_multiplier: number
  is_exclusive: boolean
  models: UserPricingModel[]
}

/** GET /pricing/public/groups — 模型广场公开端点：只返回非专属、非订阅的活跃分组。 */
export async function getPublicPricingGroups(options?: { signal?: AbortSignal }): Promise<UserPricingGroup[]> {
  const { data } = await apiClient.get<UserPricingGroup[]>('/pricing/public/groups', {
    signal: options?.signal,
  })
  return data
}

/** GET /pricing/groups — 登录后「模型定价」页：当前用户可访问的分组（含专属/订阅组）。 */
export async function getPricingGroups(options?: { signal?: AbortSignal }): Promise<UserPricingGroup[]> {
  const { data } = await apiClient.get<UserPricingGroup[]>('/pricing/groups', {
    signal: options?.signal,
  })
  return data
}

/** 展示用汇率：cny_per_usd 即 pricing.cny_to_usd_rate（1 USD ≈ 几 CNY；1¥=1$ 余额模型下默认为 1）。 */
export interface FXRate {
  cny_per_usd: number
  last_updated: string | null
}

/** GET /pricing/public/fx-rate — 模型广场本站价换算用汇率（公开，无需认证）。 */
export async function getPublicFXRate(options?: { signal?: AbortSignal }): Promise<FXRate> {
  const { data } = await apiClient.get<FXRate>('/pricing/public/fx-rate', {
    signal: options?.signal,
  })
  return data
}

export const userChannelsAPI = { getAvailable, getPublicPricingGroups, getPricingGroups, getPublicFXRate }

export default userChannelsAPI
