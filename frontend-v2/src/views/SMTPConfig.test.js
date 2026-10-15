import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ElMessage, ElMessageBox, ElButton } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import SMTPConfig from './SMTPConfig.vue'

// mock 系统 API 层：组件测试只关心交互与状态，不关心 HTTP
vi.mock('../api/system', () => ({
  getSMTPConfig: vi.fn(() => Promise.resolve({})),
  saveSMTPConfig: vi.fn(() => Promise.resolve({})),
  testSMTPConfig: vi.fn(() => Promise.resolve({})),
  getSiteConfig: vi.fn(() => Promise.resolve({ domain: '' })),
  saveSiteConfig: vi.fn(() => Promise.resolve({})),
  getVersion: vi.fn(() =>
    Promise.resolve({ version: 'v3.0.1', platform: 'windows/amd64', updaterPresent: true })
  ),
  checkUpdate: vi.fn(),
  startUpdate: vi.fn(() => Promise.resolve({ target: 'v3.0.2' })),
  getUpdateStatus: vi.fn(() => Promise.resolve({ hasResult: false })),
  getUpdateLog: vi.fn(() => Promise.resolve({ log: '' })),
}))

import {
  getVersion,
  checkUpdate,
  startUpdate,
  getUpdateStatus,
} from '../api/system'

// 图标全局注册（与 main.js 保持一致）
const iconComponents = {}
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  iconComponents[key] = component
}

// element-plus 默认导出（插件安装用）
let ElementPlus

beforeEach(async () => {
  ElementPlus = (await import('element-plus')).default
  vi.spyOn(ElMessage, 'success').mockImplementation(() => {})
  vi.spyOn(ElMessage, 'info').mockImplementation(() => {})
  vi.spyOn(ElMessage, 'warning').mockImplementation(() => {})
  vi.spyOn(ElMessage, 'error').mockImplementation(() => {})
  vi.spyOn(ElMessageBox, 'confirm').mockResolvedValue()
  vi.stubGlobal(
    'fetch',
    vi.fn(() => Promise.resolve({ json: () => Promise.resolve({ phase: 'downloading' }) }))
  )
  // jsdom 不支持导航，替换 location 以便断言跳转
  vi.stubGlobal('location', { href: '', reload: vi.fn() })
  vi.stubGlobal('open', vi.fn())
})

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

function mountPage() {
  return mount(SMTPConfig, {
    global: {
      plugins: [ElementPlus],
      components: iconComponents,
    },
  })
}

// 按按钮文本查找
function findButton(wrapper, text) {
  return wrapper.findAll('button').find((b) => b.text() === text)
}

describe('SMTPConfig 系统更新卡片', () => {
  it('挂载后展示当前版本与平台信息', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(getVersion).toHaveBeenCalled()
    const verText = wrapper.findAll('.ver-text')
    expect(verText[0].text()).toBe('v3.0.1')
    expect(wrapper.text()).toContain('windows/amd64')
    wrapper.unmount()
  })

  it('检查更新：存在同大版本新版本时展示可热更新标签与体积', async () => {
    checkUpdate.mockResolvedValue({
      current: 'v3.0.1',
      target: 'v3.0.2',
      hasUpdate: true,
      sameMajor: true,
      assetSize: 17112839,
      releaseNotes: '修复若干问题',
    })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    expect(checkUpdate).toHaveBeenCalled()
    const tags = wrapper.findAll('.el-tag').map((t) => t.text())
    expect(tags.some((t) => t.includes('可热更新'))).toBe(true)
    expect(tags.some((t) => t.includes('16.32 MB'))).toBe(true)
    expect(wrapper.text()).toContain('修复若干问题')
    // 同大版本：出现「立即更新」按钮，不出现「前往下载整包」
    expect(findButton(wrapper, '立即更新')).toBeTruthy()
    expect(findButton(wrapper, '前往下载整包')).toBeFalsy()
    wrapper.unmount()
  })

  it('检查更新：已是最新版本时只展示状态标签', async () => {
    checkUpdate.mockResolvedValue({
      current: 'v3.0.1',
      target: 'v3.0.1',
      hasUpdate: false,
      sameMajor: true,
    })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    const tags = wrapper.findAll('.el-tag').map((t) => t.text())
    expect(tags.some((t) => t.includes('已是最新'))).toBe(true)
    expect(findButton(wrapper, '立即更新')).toBeFalsy()
    wrapper.unmount()
  })

  it('检查更新：无可热更新版本但存在更高大版本时同样提示手动下载', async () => {
    // 后端在"找不到可用同大版本更新、但存在更高大版本"时置 crossMajor=true
    checkUpdate.mockResolvedValue({
      current: 'v3.1.0',
      target: 'v4.0.0',
      hasUpdate: false,
      sameMajor: false,
      crossMajor: true,
    })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    const tags = wrapper.findAll('.el-tag').map((t) => t.text())
    expect(tags.some((t) => t.includes('跨大版本'))).toBe(true)
    expect(tags.some((t) => t.includes('已是最新'))).toBe(false)
    expect(findButton(wrapper, '立即更新')).toBeFalsy()
    expect(findButton(wrapper, '前往下载整包')).toBeTruthy()
    expect(wrapper.text()).toContain('v4.0.0')
    wrapper.unmount()
  })

  it('检查更新：无更新时最新版本展示为当前版本而非零值', async () => {
    // 后端保证无更新时 target = current（回归：曾返回 v0.0.0）
    checkUpdate.mockResolvedValue({
      current: 'v3.1.0-dev+e40455c',
      target: 'v3.1.0-dev+e40455c',
      hasUpdate: false,
      sameMajor: true,
      crossMajor: false,
    })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    expect(wrapper.text()).not.toContain('v0.0.0')
    const verTexts = wrapper.findAll('.ver-text').map((v) => v.text())
    expect(verTexts).toContain('v3.1.0-dev+e40455c')
    wrapper.unmount()
  })

  it('检查更新：跨大版本时提示手动下载整包', async () => {
    checkUpdate.mockResolvedValue({
      current: 'v2.2.0',
      target: 'v3.0.1',
      hasUpdate: true,
      sameMajor: false,
    })
    getVersion.mockResolvedValue({ version: 'v2.2.0', platform: 'windows/amd64', updaterPresent: true })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    expect(findButton(wrapper, '立即更新')).toBeFalsy()
    const downloadBtn = findButton(wrapper, '前往下载整包')
    expect(downloadBtn).toBeTruthy()
    await downloadBtn.trigger('click')
    expect(open).toHaveBeenCalledWith(
      'https://github.com/thebestskinhead/schedule-system-v2/releases',
      '_blank'
    )
    expect(wrapper.text()).toContain('跨大版本')
    wrapper.unmount()
  })

  it('立即更新：确认后调用接口并轮询进度页，上线后跳转', async () => {
    vi.useFakeTimers()
    checkUpdate.mockResolvedValue({
      current: 'v3.0.1',
      target: 'v3.0.2',
      hasUpdate: true,
      sameMajor: true,
    })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    await findButton(wrapper, '立即更新').trigger('click')
    await flushPromises()

    // 确认弹窗已确认 → 触发更新接口
    expect(ElMessageBox.confirm).toHaveBeenCalled()
    expect(startUpdate).toHaveBeenCalled()
    expect(ElMessage.success).toHaveBeenCalled()

    // 轮询到进度页（phase 存在）→ 跳转到 8080 根路径
    await vi.advanceTimersByTimeAsync(800)
    expect(fetch).toHaveBeenCalledWith('/api/state', { cache: 'no-store' })
    expect(location.href).toBe('/')
    wrapper.unmount()
  })

  it('立即更新：取消确认时不触发更新', async () => {
    vi.spyOn(ElMessageBox, 'confirm').mockRejectedValue(new Error('cancel'))
    checkUpdate.mockResolvedValue({
      current: 'v3.0.1',
      target: 'v3.0.2',
      hasUpdate: true,
      sameMajor: true,
    })
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    await findButton(wrapper, '立即更新').trigger('click')
    await flushPromises()

    expect(startUpdate).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('立即更新：服务停止后又恢复（更新失败回滚）时提示刷新', async () => {
    vi.useFakeTimers()
    checkUpdate.mockResolvedValue({
      current: 'v3.0.1',
      target: 'v3.0.2',
      hasUpdate: true,
      sameMajor: true,
    })
    // 第一次轮询：服务已停止（网络错误）；第二次：服务已恢复（无 phase）
    fetch
      .mockRejectedValueOnce(new Error('connection refused'))
      .mockResolvedValueOnce({ json: () => Promise.resolve(null) })

    const wrapper = mountPage()
    await flushPromises()
    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()
    await findButton(wrapper, '立即更新').trigger('click')
    await flushPromises()

    await vi.advanceTimersByTimeAsync(800) // 第一次：连接失败 → sawDown
    await vi.advanceTimersByTimeAsync(800) // 第二次：服务恢复 → 提示刷新
    expect(ElMessage.info).toHaveBeenCalledWith('服务已恢复，正在刷新页面')

    await vi.advanceTimersByTimeAsync(1500)
    expect(location.reload).toHaveBeenCalled()
    wrapper.unmount()
  })

  it('展示上次更新结果', async () => {
    getUpdateStatus.mockResolvedValue({
      hasResult: true,
      phase: 'done',
      message: '更新成功，已升级到 v3.0.2',
      at: '2026-10-15 09:29:21',
    })
    const wrapper = mountPage()
    await flushPromises()

    expect(getUpdateStatus).toHaveBeenCalled()
    expect(wrapper.text()).toContain('上次更新')
    expect(wrapper.text()).toContain('更新成功，已升级到 v3.0.2')
    expect(wrapper.text()).toContain('2026-10-15 09:29:21')
    wrapper.unmount()
  })
})
