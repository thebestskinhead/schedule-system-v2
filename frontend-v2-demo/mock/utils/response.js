/**
 * 响应工具函数
 */

/**
 * 成功响应
 */
export function success(data = null, message = 'success') {
  return {
    code: 200,
    message,
    data
  }
}

/**
 * 错误响应
 */
export function error(message = '请求失败', code = 400) {
  return {
    code,
    message,
    data: null
  }
}

/**
 * 未授权响应
 */
export function unauthorized(message = '未登录或Token已过期') {
  return {
    code: 401,
    message,
    data: null
  }
}

/**
 * 禁止访问响应
 */
export function forbidden(message = '无权限执行此操作') {
  return {
    code: 403,
    message,
    data: null
  }
}

/**
 * 未找到响应
 */
export function notFound(message = '资源不存在') {
  return {
    code: 404,
    message,
    data: null
  }
}

/**
 * 服务器错误响应
 */
export function serverError(message = '服务器内部错误') {
  return {
    code: 500,
    message,
    data: null
  }
}

export default {
  success,
  error,
  unauthorized,
  forbidden,
  notFound,
  serverError
}
