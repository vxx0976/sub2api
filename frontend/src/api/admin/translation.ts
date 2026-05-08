/**
 * Admin one-click translation API
 * 通过后端代理调用 OpenAI 兼容大模型，给多语言字段一键填充翻译。
 */

import { apiClient } from '../client'

export interface TranslationConfigView {
  base_url: string
  model: string
  timeout_ms: number
  api_key_configured: boolean
  api_key_masked: string
}

export interface UpdateTranslationConfigRequest {
  base_url?: string
  model?: string
  api_key?: string
  clear_api_key?: boolean
  timeout_ms?: number
}

export interface TranslateRequest {
  texts: string[]
  target_langs: string[]
  source_lang?: string
}

export interface TranslateResponse {
  translations: Array<Record<string, string>>
}

export async function getConfig(): Promise<TranslationConfigView> {
  const { data } = await apiClient.get<TranslationConfigView>('/admin/translation/config')
  return data
}

export async function updateConfig(
  payload: UpdateTranslationConfigRequest
): Promise<TranslationConfigView> {
  const { data } = await apiClient.put<TranslationConfigView>(
    '/admin/translation/config',
    payload
  )
  return data
}

export async function translate(payload: TranslateRequest): Promise<TranslateResponse> {
  const { data } = await apiClient.post<TranslateResponse>(
    '/admin/translation/translate',
    payload
  )
  return data
}
