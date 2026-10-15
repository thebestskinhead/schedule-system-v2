import { describe, it, expect, vi, beforeEach } from 'vitest'
import request from './request'
import {
  getVersion,
  checkUpdate,
  startUpdate,
  getUpdateStatus,
  getUpdateLog,
} from './system'

// mock axios 实例，只关心「调了哪个方法、传了什么路径」
vi.mock('./request', () => ({
  default: {
    get: vi.fn(() => Promise.resolve({})),
    post: vi.fn(() => Promise.resolve({})),
    put: vi.fn(() => Promise.resolve({})),
    delete: vi.fn(() => Promise.resolve({})),
  },
}))

describe('系统更新 API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getVersion 请求 GET /system/version', async () => {
    await getVersion()
    expect(request.get).toHaveBeenCalledWith('/system/version')
  })

  it('checkUpdate 请求 GET /admin/update/check', async () => {
    await checkUpdate()
    expect(request.get).toHaveBeenCalledWith('/admin/update/check')
  })

  it('startUpdate 请求 POST /admin/update/start', async () => {
    await startUpdate()
    expect(request.post).toHaveBeenCalledWith('/admin/update/start')
  })

  it('getUpdateStatus 请求 GET /admin/update/status', async () => {
    await getUpdateStatus()
    expect(request.get).toHaveBeenCalledWith('/admin/update/status')
  })

  it('getUpdateLog 请求 GET /admin/update/log', async () => {
    await getUpdateLog()
    expect(request.get).toHaveBeenCalledWith('/admin/update/log')
  })

  it('更新接口路径与后端路由表一致（/api/v1 前缀由 axios baseURL 添加）', async () => {
    // 这些路径必须与 backend/internal/router/routes.go 中注册的路由对应，
    // 后端契约测试（TestFrontendBackendAPIContract）会做双向校验。
    const expected = [
      [getVersion, 'get', '/system/version'],
      [checkUpdate, 'get', '/admin/update/check'],
      [startUpdate, 'post', '/admin/update/start'],
      [getUpdateStatus, 'get', '/admin/update/status'],
      [getUpdateLog, 'get', '/admin/update/log'],
    ]
    for (const [fn, method, path] of expected) {
      vi.clearAllMocks()
      await fn()
      expect(request[method]).toHaveBeenCalledWith(path)
    }
  })
})
