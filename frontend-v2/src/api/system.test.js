import { describe, it, expect, vi, beforeEach } from 'vitest'
import request from './request'
import {
  getVersion,
  checkUpdate,
  startDownload,
  getDownloadProgress,
  cancelDownload,
  installUpdate,
  getUpdateStatus,
  getUpdateLog,
  getMirrors,
  saveMirrors,
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

  it('startDownload 请求 POST /admin/update/download', async () => {
    await startDownload()
    expect(request.post).toHaveBeenCalledWith('/admin/update/download')
  })

  it('getDownloadProgress 请求 GET /admin/update/progress', async () => {
    await getDownloadProgress()
    expect(request.get).toHaveBeenCalledWith('/admin/update/progress')
  })

  it('cancelDownload 请求 POST /admin/update/cancel-download', async () => {
    await cancelDownload()
    expect(request.post).toHaveBeenCalledWith('/admin/update/cancel-download')
  })

  it('installUpdate 请求 POST /admin/update/install', async () => {
    await installUpdate()
    expect(request.post).toHaveBeenCalledWith('/admin/update/install')
  })

  it('getUpdateStatus 请求 GET /admin/update/status', async () => {
    await getUpdateStatus()
    expect(request.get).toHaveBeenCalledWith('/admin/update/status')
  })

  it('getUpdateLog 请求 GET /admin/update/log', async () => {
    await getUpdateLog()
    expect(request.get).toHaveBeenCalledWith('/admin/update/log')
  })

  it('getMirrors 请求 GET /admin/update/mirrors', async () => {
    await getMirrors()
    expect(request.get).toHaveBeenCalledWith('/admin/update/mirrors')
  })

  it('saveMirrors 请求 PUT /admin/update/mirrors 并携带 mirrors 字段', async () => {
    await saveMirrors(['https://mirror.example.com'])
    expect(request.put).toHaveBeenCalledWith('/admin/update/mirrors', {
      mirrors: ['https://mirror.example.com'],
    })
  })

  it('更新接口路径与后端路由表一致（/api/v1 前缀由 axios baseURL 添加）', async () => {
    // 这些路径必须与 backend/internal/router/routes.go 中注册的路由对应，
    // 后端契约测试（TestFrontendBackendAPIContract）会做双向校验。
    const expected = [
      [getVersion, 'get', '/system/version'],
      [checkUpdate, 'get', '/admin/update/check'],
      [startDownload, 'post', '/admin/update/download'],
      [getDownloadProgress, 'get', '/admin/update/progress'],
      [cancelDownload, 'post', '/admin/update/cancel-download'],
      [installUpdate, 'post', '/admin/update/install'],
      [getUpdateStatus, 'get', '/admin/update/status'],
      [getUpdateLog, 'get', '/admin/update/log'],
      [getMirrors, 'get', '/admin/update/mirrors'],
    ]
    for (const [fn, method, path] of expected) {
      vi.clearAllMocks()
      await fn()
      expect(request[method]).toHaveBeenCalledWith(path)
    }
  })
})
