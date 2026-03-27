/**
 * 原生桥接模块
 * 用于与安卓 WebView 通信
 */

// 一次性检测，结果缓存到模块级变量
// 无论用户在哪个页面，第一次调用时确定，后续直接返回缓存结果
let _isNative = null

// 登录完成回调（内部使用）
let _onLoginCompleteCallback = null
let _onLoginFailedCallback = null

/**
 * 检测是否在安卓原生 WebView 中
 * 优先检测 window.Android 对象，其次通过 User-Agent 辅助判断
 * 检测结果缓存，整个 App 生命周期只检测一次
 */
export const isNativeAvailable = () => {
  if (_isNative !== null) return _isNative

  // 方式1：安卓注入的 JS Bridge 对象
  if (window.Android) {
    _isNative = true
    return true
  }

  // 方式2：通过 User-Agent 辅助判断（wv 表示 WebView）
  const ua = navigator.userAgent || ''
  if (/wv/i.test(ua) && /Android/i.test(ua)) {
    _isNative = true
    return true
  }

  _isNative = false
  return false
}

/**
 * 设置登录回调（安卓端登录完成后会调用这些回调）
 * @param {Function} onComplete - 登录完成回调
 * @param {Function} onFailed - 登录失败回调，参数为错误信息
 */
export const setLoginCallbacks = (onComplete, onFailed) => {
  _onLoginCompleteCallback = onComplete
  _onLoginFailedCallback = onFailed
  
  // 注册到 window 对象，供安卓调用
  window.onLoginComplete = () => {
    console.log('[NativeBridge] 登录完成回调')
    if (_onLoginCompleteCallback) {
      _onLoginCompleteCallback()
    }
  }
  
  window.onLoginFailed = (errorMessage) => {
    console.log('[NativeBridge] 登录失败回调:', errorMessage)
    if (_onLoginFailedCallback) {
      _onLoginFailedCallback(errorMessage)
    }
  }
}

/**
 * 从教务系统获取课表
 * @returns {Promise<{success: boolean, needLogin?: boolean, base64?: string, fileName?: string, mimeType?: string, message?: string}>}
 * 
 * 返回情况：
 * - 成功: { success: true, base64: "...", fileName: "...", mimeType: "..." }
 * - 需要登录: { success: false, needLogin: true, message: "..." }
 * - 其他错误: { success: false, message: "..." }
 */
export const fetchScheduleFromSchool = () => {
  if (window.Android?.fetchScheduleFromSchool) {
    return new Promise((resolve, reject) => {
      try {
        const resultStr = window.Android.fetchScheduleFromSchool()
        const result = JSON.parse(resultStr)
        
        console.log('[NativeBridge] fetchScheduleFromSchool 结果:', result)
        
        // 情况1：需要登录
        if (result.needLogin) {
          resolve({ success: false, needLogin: true, message: result.message || '请先登录教务系统' })
          return
        }
        
        // 情况2：成功获取课表
        if (result.success && result.base64) {
          resolve(result)
          return
        }
        
        // 情况3：其他错误
        reject(new Error(result.message || '获取课表失败'))
      } catch (e) {
        reject(new Error(e.message || '解析原生返回数据失败'))
      }
    })
  }
  return Promise.reject(new Error('当前环境不支持获取课表'))
}

/**
 * 分享/保存文件（导出排班表时使用）
 * @param {Object} data
 * @param {number} data.week - 周次
 * @param {string} data.fileName - 文件名，如 "排班表_第5周.xlsx"
 * @param {string} data.fileData - 文件内容的 base64 编码
 * @returns {Promise<boolean>}
 */
export const shareSchedule = (data) => {
  if (window.Android?.shareSchedule) {
    return new Promise((resolve, reject) => {
      try {
        const result = window.Android.shareSchedule(JSON.stringify(data))
        if (result === 'true' || result === true) {
          resolve(true)
        } else {
          reject(new Error('分享失败'))
        }
      } catch (e) {
        reject(new Error('调用分享功能失败'))
      }
    })
  }
  return Promise.reject(new Error('当前环境不支持分享功能'))
}

/**
 * 下载文件并分享（安卓端执行下载）
 * @param {Object} options
 * @param {string} options.url - 完整的下载URL
 * @param {string} options.fileName - 保存的文件名
 * @param {string} [options.method='GET'] - 请求方法 GET/POST
 * @param {Object} [options.headers] - 请求头
 * @param {string} [options.body] - POST请求体（JSON字符串）
 * @returns {Promise<{success: boolean, message?: string}>}
 */
export const downloadAndShare = (options) => {
  if (window.Android?.downloadAndShare) {
    return new Promise((resolve, reject) => {
      try {
        const result = JSON.parse(window.Android.downloadAndShare(JSON.stringify(options)))
        if (result.success) {
          resolve(result)
        } else {
          reject(new Error(result.message || '下载失败'))
        }
      } catch (e) {
        reject(new Error('调用下载功能失败: ' + e.message))
      }
    })
  }
  return Promise.reject(new Error('当前环境不支持下载功能'))
}

export default {
  fetchScheduleFromSchool,
  shareSchedule,
  downloadAndShare,
  isNativeAvailable,
  setLoginCallbacks
}
