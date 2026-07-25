/**
 * 格式化缓存 token 数量（1K/1M 缩写）
 */
export function formatCacheTokens(tokens: number): string {
  if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`
  if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}K`
  return tokens.toLocaleString()
}

/**
 * 去掉定点小数字符串末尾多余的 0："1.500000"→"1.5"，"100.000000"→"100"。
 * 仅在含小数点时裁剪，避免整数字符串被误删末尾 0（如 "100"→"1"）。
 * 全站成本/价格/倍率展示统一用它裁掉无意义的尾随 0（后端仍下发精确定点串，前端按需裁剪）。
 */
export function trimTrailingZeros(s: string): string {
  // 无小数点、或科学计数法（含 e/E）时原样返回：避免把整数末尾的 0（"100"→"1"）
  // 或指数里的 0（"1.0e+10"→"1.0e+1"）也误删。
  if (!s.includes('.') || s.includes('e') || s.includes('E')) return s
  return s.replace(/0+$/, '').replace(/\.$/, '')
}

/**
 * 自适应精度格式化倍率：保留至多 4 位有效小数，避免 0.035 被 toFixed(2) 四舍五入成 0.04
 * （上游 fix(usage): keep significant decimals in cost tooltip rate multiplier），
 * 再按本仓库的展示约定裁掉无意义的尾随 0（1.0000→1，2.5000→2.5，0.0350→0.035）。
 */
export function formatMultiplier(val: number): string {
  if (val >= 0.0001) return trimTrailingZeros(val.toFixed(4))
  return trimTrailingZeros(val.toPrecision(2))
}
