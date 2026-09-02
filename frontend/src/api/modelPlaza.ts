/**
 * Model Plaza API（公开端点，可匿名访问）
 * 以分组为中心的模型价目：分组信息 + 模型渠道定价 + LiteLLM 官方参考价。
 * 带 token 请求时后端会额外返回专属分组与用户专属倍率。
 */

import { apiClient } from './client'
import type { UserPricingInterval, UserSupportedModelPricing } from './channels'

/** 官方参考价（USD per token，与计费目录同源；字段缺失 = 目录未覆盖）。 */
export interface PlazaOfficialPricing {
  input_price: number | null
  output_price: number | null
  /** 5m 缓存写入（= LiteLLM cache_creation）。 */
  cache_write_price: number | null
  /** 1h 缓存写入（LiteLLM cache_creation_above_1hr），多数模型缺失。 */
  cache_write_1h_price?: number | null
  cache_read_price: number | null
  /** 官方长上下文阶梯（多档模型才有），不受分组开关影响。 */
  intervals?: UserPricingInterval[]
}

/**
 * 多档时的计价基准：
 * - whole_request：整单按所在档单价计价（目录阶梯、渠道区间）；
 * - marginal：仅超出阈值的部分按该档单价计价（平台旧规则）。
 */
export type PlazaLongContextBasis = 'whole_request' | 'marginal'

/** 分时倍率时段：配置时区当天 [start_time, end_time) 内整单实付乘 multiplier。 */
export interface PlazaTimePricingPeriod {
  start_time: string
  end_time: string
  multiplier: number
}

/** 计费会生效的分时倍率（仅倍率 ≠ 1 的时段，已按开始时间升序）。 */
export interface PlazaTimePricing {
  /** IANA 时区名，如 Asia/Shanghai。 */
  timezone: string
  /** true 时时段仅周一至周五生效，周末整天按标准价计费。 */
  weekdays_only?: boolean
  periods: PlazaTimePricingPeriod[]
}

/**
 * 上游官方的时段分档（如 DeepSeek 高峰价 + 空闲时段半价）。
 * ⚠️ 与分组的 peak_rate_*（本站订阅高峰倍率）是两个正交概念，不要混用。
 * current_band 由后端按官方时区判定 —— 前端**不得**按浏览器本地时区自己算，
 * 否则不同时区的用户会看到不同档位。
 */
export interface ModelTimeTier {
  /** 高峰窗口，如 ['09:00-12:00', '14:00-18:00']。 */
  peak_windows: string[]
  /** 窗口所用时区标签，如 'UTC+08:00'。 */
  timezone: string
  /** 空闲档系数（如 0.5 表示半价）。 */
  off_peak_factor: number
  /** 当前所处档位：'peak' | 'offpeak'。 */
  current_band: string
  /** 高峰窗口是否仅工作日生效（周末全天空闲）。缺省视为 false。 */
  peak_weekdays_only?: boolean
}

export interface PlazaModel {
  name: string
  platform: string
  /**
   * 价格数值的币种（'CNY' | 'USD'）。国产按人民币官方计价的模型（deepseek /
   * kimi / glm 等）为 'CNY'，前端据此显示 ¥；缺省按 '$' 处理。
   */
  price_currency?: string
  /** 实收口径的展示定价：多档时 intervals 为各档绝对单价（已由计费服务折算）；均为标准时段价。 */
  pricing: UserSupportedModelPricing | null
  official_pricing: PlazaOfficialPricing | null
  /** 仅多档模型返回。 */
  long_context_basis?: PlazaLongContextBasis
  /** 仅配置了分时倍率的模型返回（本站渠道售价倍率）。 */
  time_pricing?: PlazaTimePricing
  /** 官方时段分档（上游厂商峰谷价，与 time_pricing 正交）；无分档的模型不下发该字段。 */
  time_tier?: ModelTimeTier | null
}

export interface ModelPlazaGroup {
  id: number
  name: string
  description: string
  platform: string
  /** 'standard' | 'subscription' */
  subscription_type: string
  rate_multiplier: number
  /** 登录且管理员为该用户配了专属倍率时返回；生效倍率 = user_rate ?? rate_multiplier。 */
  user_rate_multiplier?: number
  peak_rate_enabled: boolean
  peak_start: string
  peak_end: string
  peak_rate_multiplier: number
  is_exclusive: boolean
  /** 生图独立倍率：true 时图片计费模型的实付倍率取 image_rate_multiplier，不取分组/专属倍率。 */
  image_rate_independent: boolean
  image_rate_multiplier: number
  /** 分组是否启用长上下文阶梯计费；false 时实付列只展示最低档，官方阶梯仅供参考。 */
  long_context_pricing_enabled: boolean
  models: PlazaModel[]
}

export interface ModelPlazaResponse {
  /** 管理员配置的全局价格说明（Markdown）。 */
  description: string
  groups: ModelPlazaGroup[]
}

/** 获取模型广场数据。开关未启用时后端返回 404。 */
export async function getModelPlaza(options?: { signal?: AbortSignal }): Promise<ModelPlazaResponse> {
  const { data } = await apiClient.get<ModelPlazaResponse>('/model-plaza', {
    signal: options?.signal
  })
  return data
}

export const modelPlazaAPI = { getModelPlaza }

export default modelPlazaAPI
