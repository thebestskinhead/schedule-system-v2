/**
 * 延迟工具
 * 模拟网络请求延迟
 */

/**
 * 模拟延迟
 * @param {number} ms - 基础延迟毫秒数
 * @param {number} variance - 随机波动范围（默认±30%）
 */
export function delay(ms = 300, variance = 0.3) {
  const varianceMs = ms * variance * (Math.random() * 2 - 1)
  const actualDelay = Math.max(50, ms + varianceMs)
  
  return new Promise(resolve => setTimeout(resolve, actualDelay))
}

/**
 * 快速延迟（用于简单查询）
 */
export function fastDelay() {
  return delay(100, 0.2)
}

/**
 * 中等延迟（用于普通操作）
 */
export function mediumDelay() {
  return delay(300, 0.3)
}

/**
 * 慢延迟（用于复杂操作）
 */
export function slowDelay() {
  return delay(600, 0.4)
}

export default {
  delay,
  fastDelay,
  mediumDelay,
  slowDelay
}
