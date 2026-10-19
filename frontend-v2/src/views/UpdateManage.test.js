import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import UpdateManage from './UpdateManage.vue'

// mock 系统 API 层：组件测试只关心交互与状态，不关心 HTTP
vi.mock('../api/system', () => ({
  getVersion: vi.fn(() =>
    Promise.resolve({ version: 'v3.0.1', platform: 'windows/amd64', updaterPresent: true })
  ),
  checkUpdate: vi.fn(),
  startDownload: vi.fn(() => Promise.resolve({ target: 'v3.0.2' })),
  getDownloadProgress: vi.fn(() => Promise.resolve({ phase: 'idle' })),
  cancelDownload: vi.fn(() => Promise.resolve({})),
  installUpdate: vi.fn(() => Promise.resolve({ target: 'v3.0.2' })),
  getUpdateStatus: vi.fn(() => Promise.resolve({ hasResult: false })),
  getUpdateLog: vi.fn(() => Promise.resolve({ log: '' })),
  getMirrors: vi.fn(() => Promise.resolve({ mirrors: [] })),
  saveMirrors: vi.fn(() => Promise.resolve({ mirrors: [] })),
}))

import {
  checkUpdate,
  startDownload,
  getDownloadProgress,
  cancelDownload,
  installUpdate,
  getUpdateStatus,
  getMirrors,
  saveMirrors,
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
  // jsdom 不支持导航，替换 location 以便断言跳转
  vi.stubGlobal('location', { href: '', reload: vi.fn() })
  vi.stubGlobal('open', vi.fn())
  vi.stubGlobal(
    'fetch',
    vi.fn(() => Promise.resolve({ json: () => Promise.resolve({ phase: 'downloading' }) }))
  )
})

afterEach(() => {
  vi.restoreAllMocks()
  vi.unstubAllGlobals()
  vi.useRealTimers()
})

function mountPage() {
  return mount(UpdateManage, {
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

// ==================== 版本与检查更新 ====================

describe('UpdateManage 版本与检查更新', () => {
  it('挂载后展示当前版本与平台信息', async () => {
    const wrapper = mountPage()
    await flushPromises()

    const verText = wrapper.findAll('.ver-text')
    expect(verText[0].text()).toBe('v3.0.1')
    expect(wrapper.text()).toContain('windows/amd64')
    wrapper.unmount()
  })

  it('挂载后拉取下载进度（server 重启后状态可恢复）', async () => {
    const wrapper = mountPage()
    await flushPromises()

    expect(getDownloadProgress).toHaveBeenCalled()
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
    // 有可热更新版本时才展示下载面板与下载按钮
    expect(findButton(wrapper, '下载更新包')).toBeTruthy()
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
    expect(findButton(wrapper, '下载更新包')).toBeFalsy()
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
    expect(findButton(wrapper, '下载更新包')).toBeFalsy()
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
    const wrapper = mountPage()
    await flushPromises()

    await findButton(wrapper, '检查更新').trigger('click')
    await flushPromises()

    expect(findButton(wrapper, '下载更新包')).toBeFalsy()
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

// ==================== 下载与安装 ====================

describe('UpdateManage 下载与安装', () => {
  // 让页面进入「有可热更新版本」状态
  async function mountWithUpdate() {
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
    return wrapper
  }

  it('下载更新包：点击后调用接口并展示下载进度', async () => {
    // 挂载时空闲（展示下载按钮），点击后进入下载中
    getDownloadProgress
      .mockResolvedValueOnce({ phase: 'idle' })
      .mockResolvedValue({
        phase: 'downloading',
        message: '正在下载更新包… 42%',
        received: 1024,
        total: 2048,
        progress: 50,
        mirror: 'https://ghfast.top',
        running: true,
      })
    const wrapper = await mountWithUpdate()

    await findButton(wrapper, '下载更新包').trigger('click')
    await flushPromises()

    expect(startDownload).toHaveBeenCalled()
    expect(wrapper.text()).toContain('正在下载更新包… 42%')
    expect(wrapper.text()).toContain('镜像 https://ghfast.top')
    expect(wrapper.text()).toContain('1.0 KB / 2.0 KB')
    // 下载中展示取消按钮，不展示安装按钮
    expect(findButton(wrapper, '取消下载')).toBeTruthy()
    expect(findButton(wrapper, '安装更新')).toBeFalsy()
    wrapper.unmount()
  })

  it('下载中：取消下载需确认，确认后调用取消接口', async () => {
    getDownloadProgress
      .mockResolvedValueOnce({ phase: 'idle' })
      .mockResolvedValue({
        phase: 'downloading',
        message: '正在下载更新包… 42%',
        progress: 50,
        running: true,
      })
    const wrapper = await mountWithUpdate()
    await findButton(wrapper, '下载更新包').trigger('click')
    await flushPromises()

    await findButton(wrapper, '取消下载').trigger('click')
    await flushPromises()
    expect(ElMessageBox.confirm).toHaveBeenCalled()

    await findButton(wrapper, '取消下载').trigger('click')
    await flushPromises()
    expect(cancelDownload).toHaveBeenCalled()
    wrapper.unmount()
  })

  it('就绪：展示安装按钮与重新下载入口', async () => {
    getDownloadProgress.mockResolvedValue({
      phase: 'ready',
      message: '更新包已就绪',
      target: 'v3.0.2',
      mirror: 'https://ghfast.top',
    })
    const wrapper = await mountWithUpdate()

    expect(findButton(wrapper, '安装更新')).toBeTruthy()
    expect(findButton(wrapper, '重新下载')).toBeTruthy()
    expect(wrapper.text()).toContain('更新包已就绪')
    wrapper.unmount()
  })

  it('安装更新：确认后调用接口并轮询进度页，上线后跳转', async () => {
    vi.useFakeTimers()
    getDownloadProgress.mockResolvedValue({ phase: 'ready', target: 'v3.0.2' })
    const wrapper = await mountWithUpdate()

    await findButton(wrapper, '安装更新').trigger('click')
    await flushPromises()

    // 确认弹窗已确认 → 触发安装接口
    expect(ElMessageBox.confirm).toHaveBeenCalled()
    expect(installUpdate).toHaveBeenCalled()

    // 轮询到安装进度页（phase 存在）→ 跳转到根路径
    await vi.advanceTimersByTimeAsync(800)
    expect(fetch).toHaveBeenCalledWith('/api/state', { cache: 'no-store' })
    expect(location.href).toBe('/')
    wrapper.unmount()
  })

  it('安装更新：取消确认时不触发安装', async () => {
    vi.spyOn(ElMessageBox, 'confirm').mockRejectedValue(new Error('cancel'))
    getDownloadProgress.mockResolvedValue({ phase: 'ready', target: 'v3.0.2' })
    const wrapper = await mountWithUpdate()

    await findButton(wrapper, '安装更新').trigger('click')
    await flushPromises()

    expect(installUpdate).not.toHaveBeenCalled()
    wrapper.unmount()
  })

  it('安装更新：服务停止后又恢复（安装失败回滚）时提示刷新', async () => {
    vi.useFakeTimers()
    getDownloadProgress.mockResolvedValue({ phase: 'ready', target: 'v3.0.2' })
    const wrapper = await mountWithUpdate()

    // 第一次轮询：服务已停止（网络错误）；第二次：服务已恢复（无 phase）
    fetch
      .mockRejectedValueOnce(new Error('connection refused'))
      .mockResolvedValueOnce({ json: () => Promise.resolve(null) })

    await findButton(wrapper, '安装更新').trigger('click')
    await flushPromises()

    await vi.advanceTimersByTimeAsync(800) // 第一次：连接失败 → sawDown
    await vi.advanceTimersByTimeAsync(800) // 第二次：服务恢复 → 提示刷新
    expect(ElMessage.info).toHaveBeenCalledWith('服务已恢复，正在刷新页面')

    await vi.advanceTimersByTimeAsync(1500)
    expect(location.reload).toHaveBeenCalled()
    wrapper.unmount()
  })

  it('下载失败：展示错误原因与重新下载按钮', async () => {
    getDownloadProgress.mockResolvedValue({
      phase: 'failed',
      message: '下载失败',
      error: '镜像下载要求 sha256 校验，获取校验文件失败',
      running: false,
    })
    const wrapper = await mountWithUpdate()

    expect(wrapper.text()).toContain('镜像下载要求 sha256 校验，获取校验文件失败')
    expect(findButton(wrapper, '重新下载')).toBeTruthy()
    expect(findButton(wrapper, '安装更新')).toBeFalsy()
    wrapper.unmount()
  })
})

// ==================== 镜像管理 ====================

describe('UpdateManage 镜像管理', () => {
  it('挂载后加载镜像列表', async () => {
    getMirrors.mockResolvedValue({ mirrors: ['https://ghfast.top'] })
    const wrapper = mountPage()
    await flushPromises()

    expect(getMirrors).toHaveBeenCalled()
    expect(wrapper.findAll('.mirror-row').length).toBe(1)
    wrapper.unmount()
  })

  it('添加镜像：非法协议被拒绝，合法地址进入草稿', async () => {
    const wrapper = mountPage()
    await flushPromises()

    const input = wrapper.find('.mirror-add input')
    await input.setValue('ftp://mirror.example.com')
    await findButton(wrapper, '添加').trigger('click')
    expect(ElMessage.warning).toHaveBeenCalled()
    expect(wrapper.findAll('.mirror-row').length).toBe(0)

    await input.setValue('https://ghfast.top')
    await findButton(wrapper, '添加').trigger('click')
    expect(wrapper.findAll('.mirror-row').length).toBe(1)
    wrapper.unmount()
  })

  it('填入常用镜像：去重追加', async () => {
    getMirrors.mockResolvedValue({ mirrors: ['https://ghfast.top/'] })
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.findAll('.mirror-row').length).toBe(1)

    await findButton(wrapper, '填入常用镜像').trigger('click')
    // 已有 1 个 + 其余 3 个常用镜像
    expect(wrapper.findAll('.mirror-row').length).toBe(4)
    wrapper.unmount()
  })

  it('保存镜像配置：调用接口并回显', async () => {
    saveMirrors.mockResolvedValue({ mirrors: ['https://ghfast.top'] })
    const wrapper = mountPage()
    await flushPromises()

    const input = wrapper.find('.mirror-add input')
    await input.setValue('https://ghfast.top')
    await findButton(wrapper, '添加').trigger('click')
    await findButton(wrapper, '保存镜像配置').trigger('click')
    await flushPromises()

    expect(saveMirrors).toHaveBeenCalledWith(['https://ghfast.top'])
    expect(ElMessage.success).toHaveBeenCalledWith('镜像配置已保存')
    wrapper.unmount()
  })

  it('移除镜像：从草稿中删除', async () => {
    getMirrors.mockResolvedValue({ mirrors: ['https://ghfast.top', 'https://ghproxy.net/'] })
    const wrapper = mountPage()
    await flushPromises()
    expect(wrapper.findAll('.mirror-row').length).toBe(2)

    await wrapper.findAll('.mirror-row button').at(0).trigger('click')
    expect(wrapper.findAll('.mirror-row').length).toBe(1)
    wrapper.unmount()
  })
})
