import { getConfiguredTableDefaultPageSize, normalizeTablePageSize } from '@/utils/tablePreferences'

/**
 * 读取当前系统配置的表格默认每页条数。
 * 不再使用本地持久化缓存，所有页面统一以通用表格设置为准。
 */
export function getPersistedPageSize(fallback = getConfiguredTableDefaultPageSize()): number {
  return normalizeTablePageSize(getConfiguredTableDefaultPageSize() || fallback)
}

/**
 * 本 fork 不做每页条数的本地持久化（见上），保留此导出仅为兼容上游调用方。
 */
export function setPersistedPageSize(_size: number): void {}
